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

	// Update 更新订单
	Update(ctx context.Context, order *Order) error

	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error

	// ListTipsReceivedByUser 获取用户收到的打赏订单
	ListTipsReceivedByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Order, int, error)

	// GetTipStatsByUser 汇总用户已收款的打赏数据
	GetTipStatsByUser(ctx context.Context, userID uuid.UUID) (int64, int64, error)

	// CancelExpiredOrders 取消过期未支付的订单
	CancelExpiredOrders(ctx context.Context) (int, error)
}
