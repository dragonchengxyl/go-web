package coupon

import (
	"context"

	"github.com/google/uuid"
)

// Repository 优惠券仓储接口
type Repository interface {
	// CreateCoupon 创建优惠券
	CreateCoupon(ctx context.Context, coupon *Coupon) error

	// GetCouponByCode 根据代码获取优惠券
	GetCouponByCode(ctx context.Context, code string) (*Coupon, error)

	// ValidateCoupon 验证优惠券是否可用
	ValidateCoupon(ctx context.Context, code string, userID uuid.UUID) (*Coupon, error)

	// UseCoupon 使用优惠券（增加使用次数）
	UseCoupon(ctx context.Context, code string) error

	// CreateRedeemCode 创建兑换码
	CreateRedeemCode(ctx context.Context, code *RedeemCode) error

	// GetRedeemCodeByCode 根据代码获取兑换码
	GetRedeemCodeByCode(ctx context.Context, code string) (*RedeemCode, error)

	// UseRedeemCode 使用兑换码
	UseRedeemCode(ctx context.Context, code string, userID uuid.UUID) error

	// BatchCreateRedeemCodes 批量创建兑换码
	BatchCreateRedeemCodes(ctx context.Context, codes []*RedeemCode) error
}
