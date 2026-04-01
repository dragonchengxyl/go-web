package order

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	Status        *OrderStatus
	PaymentMethod *PaymentMethod
	Type          string
	Search        string
	Page          int
	PageSize      int
}

// Repository 订单仓储接口
type Repository interface {
	// Create 创建订单
	Create(ctx context.Context, order *Order) error

	// GetByID 根据ID获取订单
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)

	// GetByIdempotencyKey 根据幂等键获取订单
	GetByIdempotencyKey(ctx context.Context, key string) (*Order, error)

	// Update 更新订单
	Update(ctx context.Context, order *Order) error

	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error

	// MarkPaidIfPending 原子地将待支付订单更新为已支付。
	// 返回 transitioned=false 表示订单当前不再是 pending_payment。
	MarkPaidIfPending(ctx context.Context, id uuid.UUID, method PaymentMethod, paidAt time.Time) (*Order, bool, error)

	// List 获取订单列表
	List(ctx context.Context, filter ListFilter) ([]*Order, int64, error)

	// ListTipsReceivedByUser 获取用户收到的打赏订单
	ListTipsReceivedByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Order, int, error)

	// GetTipStatsByUser 汇总用户已收款的打赏数据
	GetTipStatsByUser(ctx context.Context, userID uuid.UUID) (int64, int64, error)

	// CancelExpiredOrders 取消过期未支付的订单
	CancelExpiredOrders(ctx context.Context) (int, error)
}
