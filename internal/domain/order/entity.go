package order

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusFulfilled      OrderStatus = "fulfilled"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusFailed         OrderStatus = "failed"
	OrderStatusRefunded       OrderStatus = "refunded"
)

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodAlipay PaymentMethod = "alipay"
	PaymentMethodWechat PaymentMethod = "wechat"
	PaymentMethodStripe PaymentMethod = "stripe"
	PaymentMethodPayPal PaymentMethod = "paypal"
)

// Order 订单实体
type Order struct {
	ID             uuid.UUID      `json:"id"`
	OrderNo        string         `json:"order_no"` // 业务订单号
	UserID         uuid.UUID      `json:"user_id"`
	Status         OrderStatus    `json:"status"`
	TotalCents     int            `json:"total_cents"`     // 总金额（分）
	Currency       string         `json:"currency"`        // 货币代码
	DiscountCents  int            `json:"discount_cents"`  // 折扣金额（分）
	CouponCode     *string        `json:"coupon_code"`     // 优惠券代码
	PaymentMethod  PaymentMethod  `json:"payment_method"`  // 支付方式
	PaidAt         *time.Time     `json:"paid_at"`         // 支付时间
	IdempotencyKey string         `json:"idempotency_key"` // 幂等性键
	Metadata       map[string]any `json:"metadata"`        // 元数据
	CreatedAt      time.Time      `json:"created_at"`
	ExpiresAt      *time.Time     `json:"expires_at"` // 订单过期时间
	UpdatedAt      time.Time      `json:"updated_at"`
}
