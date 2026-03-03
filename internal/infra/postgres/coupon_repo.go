package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/studio/platform/internal/domain/coupon"
	"github.com/studio/platform/internal/pkg/apperr"
)

type couponRepo struct {
	db *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) coupon.Repository {
	return &couponRepo{db: db}
}

const createCouponSQL = `
	INSERT INTO coupons (id, code, name, coupon_type, discount_value, min_purchase_cents, max_discount_cents, start_at, end_at, max_uses, max_uses_per_user, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING created_at, updated_at
`

func (r *couponRepo) CreateCoupon(ctx context.Context, c *coupon.Coupon) error {
	c.ID = uuid.New()
	err := r.db.QueryRow(ctx, createCouponSQL,
		c.ID, c.Code, c.Name, c.CouponType, c.DiscountValue,
		c.MinPurchaseCents, c.MaxDiscountCents, c.StartAt, c.EndAt,
		c.MaxUses, c.MaxUsesPerUser, c.IsActive,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

const getCouponByCodeSQL = `
	SELECT id, code, name, coupon_type, discount_value, min_purchase_cents, max_discount_cents, start_at, end_at, max_uses, used_count, max_uses_per_user, is_active, created_at, updated_at
	FROM coupons WHERE code = $1
`

func (r *couponRepo) GetCouponByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	var c coupon.Coupon
	err := r.db.QueryRow(ctx, getCouponByCodeSQL, code).Scan(
		&c.ID, &c.Code, &c.Name, &c.CouponType, &c.DiscountValue,
		&c.MinPurchaseCents, &c.MaxDiscountCents, &c.StartAt, &c.EndAt,
		&c.MaxUses, &c.UsedCount, &c.MaxUsesPerUser, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("coupon", "code", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return &c, nil
}

func (r *couponRepo) ValidateCoupon(ctx context.Context, code string, userID uuid.UUID) (*coupon.Coupon, error) {
	c, err := r.GetCouponByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查是否激活
	if !c.IsActive {
		return nil, apperr.BadRequest("优惠券已失效")
	}

	// 检查时间范围
	now := time.Now()
	if now.Before(c.StartAt) || now.After(c.EndAt) {
		return nil, apperr.BadRequest("优惠券不在有效期内")
	}

	// 检查使用次数
	if c.MaxUses != nil && c.UsedCount >= *c.MaxUses {
		return nil, apperr.BadRequest("优惠券已达到最大使用次数")
	}

	// 检查用户使用次数（需要查询订单表）
	if c.MaxUsesPerUser != nil {
		const countUserUsageSQL = `SELECT COUNT(*) FROM orders WHERE user_id = $1 AND coupon_code = $2 AND status IN ('paid', 'fulfilled')`
		var userUsageCount int
		if err := r.db.QueryRow(ctx, countUserUsageSQL, userID, code).Scan(&userUsageCount); err != nil {
			return nil, fmt.Errorf("failed to count user coupon usage: %w", err)
		}
		if userUsageCount >= *c.MaxUsesPerUser {
			return nil, apperr.BadRequest("您已达到该优惠券的使用次数上限")
		}
	}

	return c, nil
}

const useCouponSQL = `UPDATE coupons SET used_count = used_count + 1, updated_at = NOW() WHERE code = $1`

func (r *couponRepo) UseCoupon(ctx context.Context, code string) error {
	_, err := r.db.Exec(ctx, useCouponSQL, code)
	if err != nil {
		return fmt.Errorf("failed to use coupon: %w", err)
	}
	return nil
}

const createRedeemCodeSQL = `
	INSERT INTO redeem_codes (id, code, product_id, description, expires_at, is_active)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING created_at
`

func (r *couponRepo) CreateRedeemCode(ctx context.Context, rc *coupon.RedeemCode) error {
	rc.ID = uuid.New()
	err := r.db.QueryRow(ctx, createRedeemCodeSQL,
		rc.ID, rc.Code, rc.ProductID, rc.Description, rc.ExpiresAt, rc.IsActive,
	).Scan(&rc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create redeem code: %w", err)
	}
	return nil
}

const getRedeemCodeByCodeSQL = `
	SELECT id, code, product_id, description, used_by, used_at, expires_at, is_active, created_at
	FROM redeem_codes WHERE code = $1
`

func (r *couponRepo) GetRedeemCodeByCode(ctx context.Context, code string) (*coupon.RedeemCode, error) {
	var rc coupon.RedeemCode
	err := r.db.QueryRow(ctx, getRedeemCodeByCodeSQL, code).Scan(
		&rc.ID, &rc.Code, &rc.ProductID, &rc.Description,
		&rc.UsedBy, &rc.UsedAt, &rc.ExpiresAt, &rc.IsActive, &rc.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("redeem_code", "code", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get redeem code: %w", err)
	}
	return &rc, nil
}

const useRedeemCodeSQL = `UPDATE redeem_codes SET used_by = $2, used_at = NOW() WHERE code = $1 AND used_by IS NULL`

func (r *couponRepo) UseRedeemCode(ctx context.Context, code string, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx, useRedeemCodeSQL, code, userID)
	if err != nil {
		return fmt.Errorf("failed to use redeem code: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.BadRequest("兑换码已被使用或不存在")
	}
	return nil
}

func (r *couponRepo) BatchCreateRedeemCodes(ctx context.Context, codes []*coupon.RedeemCode) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, rc := range codes {
		rc.ID = uuid.New()
		err := tx.QueryRow(ctx, createRedeemCodeSQL,
			rc.ID, rc.Code, rc.ProductID, rc.Description, rc.ExpiresAt, rc.IsActive,
		).Scan(&rc.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create redeem code in batch: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
