package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/pkg/apperr"
)

// TipService handles tip (donation) functionality
type TipService struct {
	orderRepo order.Repository
	publisher *streams.Publisher // may be nil
}

func NewTipService(orderRepo order.Repository, opts ...func(*TipService)) *TipService {
	s := &TipService{orderRepo: orderRepo}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithTipPublisher injects an event publisher into TipService.
func WithTipPublisher(p *streams.Publisher) func(*TipService) {
	return func(s *TipService) { s.publisher = p }
}

// CreateTipInput represents input for creating a tip
type CreateTipInput struct {
	FromUserID uuid.UUID
	ToUserID   uuid.UUID
	AmountCNY  float64 // Amount in yuan (will be converted to cents)
	Message    string
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

	o := &order.Order{
		ID:       uuid.New(),
		OrderNo:  generateOrderNo(),
		UserID:   input.FromUserID,
		Status:   order.OrderStatusPendingPayment,
		TotalCents: amountCents,
		Currency: "CNY",
		ExpiresAt: &expiresAt,
		Metadata: map[string]any{
			"type":        "tip",
			"to_user_id": input.ToUserID.String(),
			"message":    input.Message,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建打赏订单失败", err)
	}

	if s.publisher != nil {
		go func() {
			_ = s.publisher.Publish(context.Background(), streams.EventTipSent, streams.TipSentPayload{
				TipID:       o.ID.String(),
				SenderID:    input.FromUserID.String(),
				ReceiverID:  input.ToUserID.String(),
				AmountCents: amountCents,
			})
		}()
	}

	return o, nil
}

// ListReceivedTips lists tip orders received by a creator
func (s *TipService) ListReceivedTips(ctx context.Context, creatorID uuid.UUID, page, pageSize int) ([]*order.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	// We query orders where metadata->>'to_user_id' = creatorID
	// This is done via the order repo
	return s.orderRepo.ListByUserID(ctx, creatorID, page, pageSize)
}

// FormatAmount formats cents to yuan string
func FormatAmount(cents int) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

// GetTipStats returns summary stats for received tips
func (s *TipService) GetMyTipStats(ctx context.Context, userID uuid.UUID) (int64, int, error) {
	orders, total, err := s.orderRepo.ListByUserID(ctx, userID, 1, 1000)
	if err != nil {
		return 0, 0, err
	}
	var totalAmount int
	for _, o := range orders {
		if o.Status == order.OrderStatusPaid {
			totalAmount += o.TotalCents
		}
	}
	return int64(total), totalAmount, nil
}
