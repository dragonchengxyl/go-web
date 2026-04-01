package usecase_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/usecase"
)

func TestTipServiceCreateTipReturnsExistingOrderForSameIdempotencyKey(t *testing.T) {
	t.Parallel()

	repo := newFakeOrderRepo()
	svc := usecase.NewTipService(repo)

	input := usecase.CreateTipInput{
		FromUserID:     uuid.New(),
		ToUserID:       uuid.New(),
		AmountCNY:      12.34,
		Message:        "谢谢",
		IdempotencyKey: "tip-001",
	}

	first, err := svc.CreateTip(context.Background(), input)
	require.NoError(t, err)

	second, err := svc.CreateTip(context.Background(), input)
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, 1, repo.createCalls)
}

func TestTipServiceCreateTipRejectsMismatchedPayloadForSameIdempotencyKey(t *testing.T) {
	t.Parallel()

	repo := newFakeOrderRepo()
	svc := usecase.NewTipService(repo)

	fromUserID := uuid.New()
	_, err := svc.CreateTip(context.Background(), usecase.CreateTipInput{
		FromUserID:     fromUserID,
		ToUserID:       uuid.New(),
		AmountCNY:      10,
		Message:        "第一次请求",
		IdempotencyKey: "tip-002",
	})
	require.NoError(t, err)

	_, err = svc.CreateTip(context.Background(), usecase.CreateTipInput{
		FromUserID:     fromUserID,
		ToUserID:       uuid.New(),
		AmountCNY:      20,
		Message:        "第二次请求",
		IdempotencyKey: "tip-002",
	})
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeIdempotencyConflict, appErr.Code)
}

func TestTipServiceCreateTipUsesDebounceLockWithoutIdempotencyKey(t *testing.T) {
	t.Parallel()

	repo := newFakeOrderRepo()
	lockStore := &fakeTipIdempotencyStore{locks: map[string]string{}}
	svc := usecase.NewTipService(repo, usecase.WithTipIdempotency(lockStore, time.Second))

	input := usecase.CreateTipInput{
		FromUserID: uuid.New(),
		ToUserID:   uuid.New(),
		AmountCNY:  5,
		Message:    "别重复点",
	}

	lockKey := fmt.Sprintf("tip:debounce:%s:%s", input.FromUserID.String(), "busy")
	lockStore.forceBusyPrefix = fmt.Sprintf("tip:debounce:%s:", input.FromUserID.String())
	lockStore.forceBusyToken = "busy"
	lockStore.forceBusyKey = lockKey

	_, err := svc.CreateTip(context.Background(), input)
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeRateLimited, appErr.Code)
}

type fakeOrderRepo struct {
	mu          sync.Mutex
	createCalls int
	byID        map[uuid.UUID]*order.Order
	byKey       map[string]uuid.UUID
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{
		byID:  make(map[uuid.UUID]*order.Order),
		byKey: make(map[string]uuid.UUID),
	}
}

func (r *fakeOrderRepo) Create(_ context.Context, o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o.IdempotencyKey != "" {
		if _, exists := r.byKey[o.IdempotencyKey]; exists {
			return fmt.Errorf("duplicate idempotency key")
		}
	}
	r.createCalls++
	stored := cloneOrder(o)
	if stored.ID == uuid.Nil {
		stored.ID = uuid.New()
	}
	r.byID[stored.ID] = stored
	if stored.IdempotencyKey != "" {
		r.byKey[stored.IdempotencyKey] = stored.ID
	}
	*o = *cloneOrder(stored)
	return nil
}

func (r *fakeOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.byID[id]
	if !ok {
		return nil, apperr.NotFound("order", "id", id.String())
	}
	return cloneOrder(item), nil
}

func (r *fakeOrderRepo) GetByIdempotencyKey(_ context.Context, key string) (*order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byKey[key]
	if !ok {
		return nil, apperr.NotFound("order", "idempotency_key", key)
	}
	return cloneOrder(r.byID[id]), nil
}

func (r *fakeOrderRepo) Update(_ context.Context, o *order.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[o.ID] = cloneOrder(o)
	if o.IdempotencyKey != "" {
		r.byKey[o.IdempotencyKey] = o.ID
	}
	return nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, id uuid.UUID, status order.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.byID[id]
	if !ok {
		return apperr.NotFound("order", "id", id.String())
	}
	item.Status = status
	return nil
}

func (r *fakeOrderRepo) MarkPaidIfPending(_ context.Context, id uuid.UUID, method order.PaymentMethod, paidAt time.Time) (*order.Order, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.byID[id]
	if !ok {
		return nil, false, apperr.NotFound("order", "id", id.String())
	}
	if item.Status != order.OrderStatusPendingPayment {
		return nil, false, nil
	}
	item.Status = order.OrderStatusPaid
	item.PaymentMethod = method
	item.PaidAt = &paidAt
	item.UpdatedAt = paidAt
	return cloneOrder(item), true, nil
}

func (r *fakeOrderRepo) List(_ context.Context, _ order.ListFilter) ([]*order.Order, int64, error) {
	return nil, 0, nil
}

func (r *fakeOrderRepo) ListTipsReceivedByUser(_ context.Context, _ uuid.UUID, _, _ int) ([]*order.Order, int, error) {
	return nil, 0, nil
}

func (r *fakeOrderRepo) GetTipStatsByUser(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return 0, 0, nil
}

func (r *fakeOrderRepo) CancelExpiredOrders(_ context.Context) (int, error) {
	return 0, nil
}

type fakeTipIdempotencyStore struct {
	mu              sync.Mutex
	locks           map[string]string
	forceBusyPrefix string
	forceBusyKey    string
	forceBusyToken  string
}

func (s *fakeTipIdempotencyStore) TryLock(_ context.Context, key string, _ time.Duration) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.forceBusyPrefix != "" && len(key) >= len(s.forceBusyPrefix) && key[:len(s.forceBusyPrefix)] == s.forceBusyPrefix {
		s.locks[s.forceBusyKey] = s.forceBusyToken
		return "", false, nil
	}
	if _, exists := s.locks[key]; exists {
		return "", false, nil
	}
	token := uuid.NewString()
	s.locks[key] = token
	return token, true, nil
}

func (s *fakeTipIdempotencyStore) Unlock(_ context.Context, key, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.locks[key] == token {
		delete(s.locks, key)
	}
	return nil
}

func cloneOrder(o *order.Order) *order.Order {
	if o == nil {
		return nil
	}
	copyOrder := *o
	if o.Metadata != nil {
		copyOrder.Metadata = make(map[string]any, len(o.Metadata))
		for k, v := range o.Metadata {
			copyOrder.Metadata[k] = v
		}
	}
	return &copyOrder
}
