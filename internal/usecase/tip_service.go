package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/infra/eventbus"
	"github.com/studio/platform/internal/pkg/apperr"
)

// TipService handles tip (donation) functionality
type TipService struct {
	orderRepo        order.Repository
	idempotencyStore TipIdempotencyStore
	idempotencyTTL   time.Duration
}

type TipServiceOption func(*TipService)

type TipIdempotencyStore interface {
	TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error)
	Unlock(ctx context.Context, key, token string) error
}

func NewTipService(orderRepo order.Repository, opts ...TipServiceOption) *TipService {
	s := &TipService{
		orderRepo:      orderRepo,
		idempotencyTTL: 5 * time.Second,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithTipPublisher keeps the tip service signature aligned with the shared event bus,
// even though tip events are emitted by PaymentService after payment completion.
func WithTipPublisher(p eventbus.Publisher) TipServiceOption {
	return func(_ *TipService) {
		_ = p
	}
}

func WithTipIdempotency(store TipIdempotencyStore, ttl time.Duration) TipServiceOption {
	return func(s *TipService) {
		s.idempotencyStore = store
		if ttl > 0 {
			s.idempotencyTTL = ttl
		}
	}
}

// CreateTipInput represents input for creating a tip
type CreateTipInput struct {
	FromUserID     uuid.UUID
	ToUserID       uuid.UUID
	AmountCNY      float64 // Amount in yuan (will be converted to cents)
	Message        string
	IdempotencyKey string
}

// CreateTip creates a tip order from one user to another
func (s *TipService) CreateTip(ctx context.Context, input CreateTipInput) (*order.Order, error) {
	if input.FromUserID == input.ToUserID {
		return nil, apperr.BadRequest("不能给自己打赏")
	}
	if input.AmountCNY <= 0 {
		return nil, apperr.BadRequest("打赏金额必须大于0")
	}
	if input.AmountCNY < 0.01 {
		return nil, apperr.BadRequest("打赏金额最少0.01元")
	}

	amountCents := int(input.AmountCNY * 100)
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if len(idempotencyKey) > 64 {
		return nil, apperr.BadRequest("幂等键长度不能超过64")
	}

	if idempotencyKey != "" {
		existing, err := s.orderRepo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return s.matchExistingTip(existing, input, amountCents)
		}
		if !isOrderNotFound(err) {
			return nil, apperr.Wrap(apperr.CodeInternalError, "查询幂等订单失败", err)
		}
	}

	lockKey := s.tipLockKey(input, amountCents, idempotencyKey)
	lockToken, lockAcquired, err := s.tryAcquireTipLock(ctx, lockKey)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeDependencyFailed, "获取幂等锁失败", err)
	}
	if !lockAcquired && idempotencyKey != "" {
		if existing, lookupErr := s.waitForIdempotentTip(ctx, idempotencyKey, input, amountCents); lookupErr == nil {
			return existing, nil
		}
		return nil, apperr.New(apperr.CodeRateLimited, "请求正在处理中，请勿重复提交")
	}
	if !lockAcquired {
		return nil, apperr.New(apperr.CodeRateLimited, "请求正在处理中，请勿重复提交")
	}
	if s.idempotencyStore != nil {
		defer func() {
			_ = s.idempotencyStore.Unlock(context.Background(), lockKey, lockToken)
		}()
	}

	o := &order.Order{
		ID:             uuid.New(),
		OrderNo:        generateOrderNo(),
		UserID:         input.FromUserID,
		Status:         order.OrderStatusPendingPayment,
		TotalCents:     amountCents,
		Currency:       "CNY",
		ExpiresAt:      &expiresAt,
		IdempotencyKey: idempotencyKey,
		Metadata: map[string]any{
			"type":       "tip",
			"to_user_id": input.ToUserID.String(),
			"message":    strings.TrimSpace(input.Message),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		if idempotencyKey != "" {
			existing, lookupErr := s.orderRepo.GetByIdempotencyKey(ctx, idempotencyKey)
			if lookupErr == nil {
				return s.matchExistingTip(existing, input, amountCents)
			}
		}
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建打赏订单失败", err)
	}

	return o, nil
}

func (s *TipService) matchExistingTip(existing *order.Order, input CreateTipInput, amountCents int) (*order.Order, error) {
	if existing.UserID != input.FromUserID || existing.TotalCents != amountCents {
		return nil, apperr.ErrIdempotencyConflict
	}
	if existing.Metadata["type"] != "tip" {
		return nil, apperr.ErrIdempotencyConflict
	}
	toUserID, _ := existing.Metadata["to_user_id"].(string)
	if toUserID != input.ToUserID.String() {
		return nil, apperr.ErrIdempotencyConflict
	}
	message, _ := existing.Metadata["message"].(string)
	if strings.TrimSpace(message) != strings.TrimSpace(input.Message) {
		return nil, apperr.ErrIdempotencyConflict
	}
	return existing, nil
}

func (s *TipService) tipLockKey(input CreateTipInput, amountCents int, idempotencyKey string) string {
	if idempotencyKey != "" {
		return fmt.Sprintf("tip:idempotency:%s:%s", input.FromUserID.String(), idempotencyKey)
	}
	signature := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d:%s",
		input.FromUserID.String(),
		input.ToUserID.String(),
		amountCents,
		strings.TrimSpace(input.Message),
	)))
	return fmt.Sprintf("tip:debounce:%s:%s", input.FromUserID.String(), hex.EncodeToString(signature[:8]))
}

func (s *TipService) tryAcquireTipLock(ctx context.Context, key string) (string, bool, error) {
	if s.idempotencyStore == nil {
		return "", true, nil
	}
	return s.idempotencyStore.TryLock(ctx, key, s.idempotencyTTL)
}

func (s *TipService) waitForIdempotentTip(ctx context.Context, key string, input CreateTipInput, amountCents int) (*order.Order, error) {
	var lastErr error
	for i := 0; i < 4; i++ {
		existing, err := s.orderRepo.GetByIdempotencyKey(ctx, key)
		if err == nil {
			return s.matchExistingTip(existing, input, amountCents)
		}
		lastErr = err
		if !isOrderNotFound(err) {
			return nil, err
		}
		time.Sleep(150 * time.Millisecond)
	}
	return nil, lastErr
}

func isOrderNotFound(err error) bool {
	var appErr *apperr.AppError
	return errors.As(err, &appErr) && appErr.Code == apperr.CodeNotFound
}

// ListReceivedTips lists tip orders received by a creator
func (s *TipService) ListReceivedTips(ctx context.Context, creatorID uuid.UUID, page, pageSize int) ([]*order.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.ListTipsReceivedByUser(ctx, creatorID, page, pageSize)
}

// FormatAmount formats cents to yuan string
func FormatAmount(cents int) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

// GetTipStats returns summary stats for received tips
func (s *TipService) GetMyTipStats(ctx context.Context, userID uuid.UUID) (int64, int, error) {
	totalAmountCents, tipCount, err := s.orderRepo.GetTipStatsByUser(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return totalAmountCents, int(tipCount), nil
}
