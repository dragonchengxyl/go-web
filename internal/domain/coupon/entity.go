package coupon

import (
	"time"

	"github.com/google/uuid"
)

// CouponType 优惠券类型
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage" // 折扣（如8折）
	CouponTypeFixed      CouponType = "fixed"      // 满减（如满50减10）
	CouponTypeAmount     CouponType = "amount"     // 定额优惠（如减5元）
)

// Coupon 优惠券
type Coupon struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`              // 优惠券代码
	Name            string     `json:"name"`              // 优惠券名称
	CouponType      CouponType `json:"coupon_type"`       // 优惠券类型
	DiscountValue   float64    `json:"discount_value"`    // 折扣值
	MinPurchaseCents *int      `json:"min_purchase_cents"` // 最低消费金额（分）
	MaxDiscountCents *int      `json:"max_discount_cents"` // 最大折扣金额（分）
	StartAt         time.Time  `json:"start_at"`          // 开始时间
	EndAt           time.Time  `json:"end_at"`            // 结束时间
	MaxUses         *int       `json:"max_uses"`          // 最大使用次数
	UsedCount       int        `json:"used_count"`        // 已使用次数
	MaxUsesPerUser  *int       `json:"max_uses_per_user"` // 每用户最大使用次数
	IsActive        bool       `json:"is_active"`         // 是否激活
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RedeemCode 兑换码
type RedeemCode struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`         // 兑换码
	ProductID   *uuid.UUID `json:"product_id"`   // 关联的商品ID
	Description string     `json:"description"`  // 描述
	UsedBy      *uuid.UUID `json:"used_by"`      // 使用者ID
	UsedAt      *time.Time `json:"used_at"`      // 使用时间
	ExpiresAt   *time.Time `json:"expires_at"`   // 过期时间
	IsActive    bool       `json:"is_active"`    // 是否激活
	CreatedAt   time.Time  `json:"created_at"`
}
