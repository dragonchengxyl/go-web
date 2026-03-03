package order

import (
	"context"

	"github.com/google/uuid"
)

// Repository 订单仓储接口
type Repository interface {
	// Create 创建订单
	Create(ctx context.Context, order *Order) error

	// GetByID 根据ID获取订单
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)

	// GetByOrderNo 根据订单号获取订单
	GetByOrderNo(ctx context.Context, orderNo string) (*Order, error)

	// GetByIdempotencyKey 根据幂等性键获取订单
	GetByIdempotencyKey(ctx context.Context, key string) (*Order, error)

	// ListByUserID 获取用户的订单列表
	ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Order, int, error)

	// Update 更新订单
	Update(ctx context.Context, order *Order) error

	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error

	// CreateItem 创建订单项
	CreateItem(ctx context.Context, item *OrderItem) error

	// GetItemsByOrderID 获取订单的所有订单项
	GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error)

	// CancelExpiredOrders 取消过期未支付的订单
	CancelExpiredOrders(ctx context.Context) (int, error)
}
