package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/pkg/apperr"
)

type OrderService struct {
	orderRepo order.Repository
}

func NewOrderService(
	orderRepo order.Repository,
	_ any, // kept for API compatibility (was productRepo)
	_ any, // kept for API compatibility (was couponRepo)
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (*order.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(ctx context.Context, orderNo string) (*order.Order, error) {
	o, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	items, err := s.orderRepo.GetItemsByOrderID(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

// ListUserOrders 获取用户的订单列表
func (s *OrderService) ListUserOrders(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*order.Order, int, error) {
	return s.orderRepo.ListByUserID(ctx, userID, page, pageSize)
}

// PayOrder 支付订单
func (s *OrderService) PayOrder(ctx context.Context, orderID uuid.UUID, paymentMethod order.PaymentMethod) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if o.Status != order.OrderStatusPendingPayment {
		return apperr.BadRequest("订单状态不正确")
	}

	if o.ExpiresAt != nil && time.Now().After(*o.ExpiresAt) {
		return apperr.BadRequest("订单已过期")
	}

	now := time.Now()
	o.Status = order.OrderStatusPaid
	o.PaymentMethod = paymentMethod
	o.PaidAt = &now

	return s.orderRepo.Update(ctx, o)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if o.Status != order.OrderStatusPendingPayment {
		return apperr.BadRequest("只有待支付的订单可以取消")
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, order.OrderStatusCancelled)
}

// FulfillOrder 完成订单
func (s *OrderService) FulfillOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if o.Status != order.OrderStatusPaid {
		return apperr.BadRequest("只有已支付的订单可以完成")
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, order.OrderStatusFulfilled)
}

// CancelExpiredOrders 取消过期订单（定时任务）
func (s *OrderService) CancelExpiredOrders(ctx context.Context) (int, error) {
	return s.orderRepo.CancelExpiredOrders(ctx)
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("%s%08d", now.Format("20060102150405"), now.UnixNano()%100000000)
}
