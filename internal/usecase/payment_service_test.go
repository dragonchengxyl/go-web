package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/pkg/apperr"
)

func TestPaymentServiceMarkPaidTransitionsPendingOrder(t *testing.T) {
	t.Parallel()

	repo := &paymentOrderRepoStub{
		order: &order.Order{
			ID:         uuid.New(),
			Status:     order.OrderStatusPendingPayment,
			TotalCents: 100,
			Metadata:   map[string]any{"type": "tip", "to_user_id": uuid.NewString()},
		},
	}
	svc := &PaymentService{orderRepo: repo}

	err := svc.markPaid(context.Background(), repo.order.ID, order.PaymentMethodAlipay)
	require.NoError(t, err)
	assert.Equal(t, order.OrderStatusPaid, repo.order.Status)
	require.NotNil(t, repo.order.PaidAt)
}

func TestPaymentServiceMarkPaidIsIdempotentForAlreadyPaidOrder(t *testing.T) {
	t.Parallel()

	repo := &paymentOrderRepoStub{
		order: &order.Order{
			ID:     uuid.New(),
			Status: order.OrderStatusPaid,
		},
	}
	svc := &PaymentService{orderRepo: repo}

	err := svc.markPaid(context.Background(), repo.order.ID, order.PaymentMethodWechat)
	require.NoError(t, err)
}

func TestPaymentServiceMarkPaidDoesNotOverrideCancelledOrder(t *testing.T) {
	t.Parallel()

	repo := &paymentOrderRepoStub{
		order: &order.Order{
			ID:     uuid.New(),
			Status: order.OrderStatusCancelled,
		},
	}
	svc := &PaymentService{orderRepo: repo}

	err := svc.markPaid(context.Background(), repo.order.ID, order.PaymentMethodWechat)
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidParam, appErr.Code)
}

type paymentOrderRepoStub struct {
	order *order.Order
}

func (r *paymentOrderRepoStub) Create(context.Context, *order.Order) error { return nil }

func (r *paymentOrderRepoStub) GetByID(_ context.Context, id uuid.UUID) (*order.Order, error) {
	if r.order == nil || r.order.ID != id {
		return nil, apperr.NotFound("order", "id", id.String())
	}
	copyOrder := *r.order
	return &copyOrder, nil
}

func (r *paymentOrderRepoStub) GetByIdempotencyKey(context.Context, string) (*order.Order, error) {
	return nil, apperr.NotFound("order", "idempotency_key", "")
}

func (r *paymentOrderRepoStub) Update(context.Context, *order.Order) error { return nil }

func (r *paymentOrderRepoStub) UpdateStatus(context.Context, uuid.UUID, order.OrderStatus) error {
	return nil
}

func (r *paymentOrderRepoStub) MarkPaidIfPending(_ context.Context, id uuid.UUID, method order.PaymentMethod, paidAt time.Time) (*order.Order, bool, error) {
	if r.order == nil || r.order.ID != id {
		return nil, false, apperr.NotFound("order", "id", id.String())
	}
	if r.order.Status != order.OrderStatusPendingPayment {
		return nil, false, nil
	}
	r.order.Status = order.OrderStatusPaid
	r.order.PaymentMethod = method
	r.order.PaidAt = &paidAt
	r.order.UpdatedAt = paidAt
	copyOrder := *r.order
	return &copyOrder, true, nil
}

func (r *paymentOrderRepoStub) List(context.Context, order.ListFilter) ([]*order.Order, int64, error) {
	return nil, 0, nil
}

func (r *paymentOrderRepoStub) ListTipsReceivedByUser(context.Context, uuid.UUID, int, int) ([]*order.Order, int, error) {
	return nil, 0, nil
}

func (r *paymentOrderRepoStub) GetTipStatsByUser(context.Context, uuid.UUID) (int64, int64, error) {
	return 0, 0, nil
}

func (r *paymentOrderRepoStub) CancelExpiredOrders(context.Context) (int, error) {
	return 0, nil
}
