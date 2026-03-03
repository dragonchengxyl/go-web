package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/studio/platform/internal/domain/coupon"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/domain/product"
	"github.com/studio/platform/internal/pkg/apperr"
)

type OrderService struct {
	orderRepo   order.Repository
	productRepo product.Repository
	couponRepo  coupon.Repository
}

func NewOrderService(
	orderRepo order.Repository,
	productRepo product.Repository,
	couponRepo coupon.Repository,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		couponRepo:  couponRepo,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID         uuid.UUID
	Items          []OrderItemRequest
	CouponCode     *string
	IdempotencyKey string
}

// OrderItemRequest 订单项请求
type OrderItemRequest struct {
	ProductID uuid.UUID
	Quantity  int
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*order.Order, error) {
	// 检查幂等性
	if req.IdempotencyKey != "" {
		existingOrder, err := s.orderRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existingOrder != nil {
			return existingOrder, nil // 返回已存在的订单
		}
	}

	// 验证订单项
	if len(req.Items) == 0 {
		return nil, apperr.BadRequest("订单至少需要一个商品")
	}

	// 计算订单总金额
	var totalCents int
	orderItems := make([]*order.OrderItem, 0, len(req.Items))

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, apperr.BadRequest("商品数量必须大于0")
		}

		// 获取商品信息
		p, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %s: %w", item.ProductID, err)
		}

		if !p.IsActive {
			return nil, apperr.BadRequest(fmt.Sprintf("商品 %s 已下架", p.Name))
		}

		// 获取商品的有效折扣
		discounts, err := s.productRepo.GetActiveDiscounts(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		// 计算商品价格（应用折扣）
		itemPrice := p.PriceCents
		if len(discounts) > 0 {
			discount := discounts[0]
			switch discount.DiscountType {
			case product.DiscountTypePercentage:
				itemPrice = int(float64(p.PriceCents) * (1 - discount.DiscountValue/100))
			case product.DiscountTypeFixed:
				itemPrice = p.PriceCents - int(discount.DiscountValue*100)
				if itemPrice < 0 {
					itemPrice = 0
				}
			}
		}

		totalCents += itemPrice * item.Quantity

		orderItems = append(orderItems, &order.OrderItem{
			ProductID:  item.ProductID,
			PriceCents: itemPrice,
			Quantity:   item.Quantity,
		})
	}

	// 应用优惠券
	var discountCents int
	if req.CouponCode != nil && *req.CouponCode != "" {
		couponEntity, err := s.couponRepo.ValidateCoupon(ctx, *req.CouponCode, req.UserID)
		if err != nil {
			return nil, err
		}

		// 检查最低消费
		if couponEntity.MinPurchaseCents != nil && totalCents < *couponEntity.MinPurchaseCents {
			return nil, apperr.BadRequest(fmt.Sprintf("订单金额未达到优惠券最低消费要求（%d分）", *couponEntity.MinPurchaseCents))
		}

		// 计算折扣金额
		switch couponEntity.CouponType {
		case coupon.CouponTypePercentage:
			discountCents = int(float64(totalCents) * couponEntity.DiscountValue / 100)
		case coupon.CouponTypeFixed:
			// 满减
			discountCents = int(couponEntity.DiscountValue * 100)
		case coupon.CouponTypeAmount:
			// 定额优惠
			discountCents = int(couponEntity.DiscountValue * 100)
		}

		// 应用最大折扣限制
		if couponEntity.MaxDiscountCents != nil && discountCents > *couponEntity.MaxDiscountCents {
			discountCents = *couponEntity.MaxDiscountCents
		}

		// 折扣不能超过订单总额
		if discountCents > totalCents {
			discountCents = totalCents
		}
	}

	// 创建订单
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute) // 30分钟后过期

	o := &order.Order{
		OrderNo:        generateOrderNo(),
		UserID:         req.UserID,
		Status:         order.OrderStatusPendingPayment,
		TotalCents:     totalCents - discountCents,
		Currency:       "CNY",
		DiscountCents:  discountCents,
		CouponCode:     req.CouponCode,
		IdempotencyKey: req.IdempotencyKey,
		ExpiresAt:      &expiresAt,
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		return nil, err
	}

	// 创建订单项
	for _, item := range orderItems {
		item.OrderID = o.ID
		if err := s.orderRepo.CreateItem(ctx, item); err != nil {
			return nil, err
		}
	}

	// 加载订单项
	o.Items = orderItems

	return o, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (*order.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 加载订单项
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

	// 加载订单项
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

	// 检查订单状态
	if o.Status != order.OrderStatusPendingPayment {
		return apperr.BadRequest("订单状态不正确")
	}

	// 检查订单是否过期
	if o.ExpiresAt != nil && time.Now().After(*o.ExpiresAt) {
		return apperr.BadRequest("订单已过期")
	}

	// 更新订单状态
	now := time.Now()
	o.Status = order.OrderStatusPaid
	o.PaymentMethod = paymentMethod
	o.PaidAt = &now

	if err := s.orderRepo.Update(ctx, o); err != nil {
		return err
	}

	// 使用优惠券
	if o.CouponCode != nil && *o.CouponCode != "" {
		if err := s.couponRepo.UseCoupon(ctx, *o.CouponCode); err != nil {
			// 记录错误但不影响订单支付
			fmt.Printf("failed to use coupon: %v\n", err)
		}
	}

	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 只有待支付的订单可以取消
	if o.Status != order.OrderStatusPendingPayment {
		return apperr.BadRequest("只有待支付的订单可以取消")
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, order.OrderStatusCancelled)
}

// FulfillOrder 完成订单（发货）
func (s *OrderService) FulfillOrder(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 只有已支付的订单可以完成
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
	// 格式：yyyyMMddHHmmss + 8位随机数
	now := time.Now()
	return fmt.Sprintf("%s%08d", now.Format("20060102150405"), now.UnixNano()%100000000)
}
