package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"go-web/internal/domain/coupon"
	"go-web/internal/domain/product"
	"go-web/pkg/apperr"
)

type CouponService struct {
	couponRepo  coupon.Repository
	productRepo product.Repository
}

func NewCouponService(couponRepo coupon.Repository, productRepo product.Repository) *CouponService {
	return &CouponService{
		couponRepo:  couponRepo,
		productRepo: productRepo,
	}
}

// CreateCoupon 创建优惠券
func (s *CouponService) CreateCoupon(ctx context.Context, c *coupon.Coupon) error {
	// 检查代码是否已存在
	existing, err := s.couponRepo.GetCouponByCode(ctx, c.Code)
	if err == nil && existing != nil {
		return apperr.BadRequest("优惠券代码已存在")
	}

	return s.couponRepo.CreateCoupon(ctx, c)
}

// GetCouponByCode 根据代码获取优惠券
func (s *CouponService) GetCouponByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	return s.couponRepo.GetCouponByCode(ctx, code)
}

// ValidateCoupon 验证优惠券
func (s *CouponService) ValidateCoupon(ctx context.Context, code string, userID uuid.UUID) (*coupon.Coupon, error) {
	return s.couponRepo.ValidateCoupon(ctx, code, userID)
}

// CreateRedeemCode 创建兑换码
func (s *CouponService) CreateRedeemCode(ctx context.Context, rc *coupon.RedeemCode) error {
	// 如果关联了商品，检查商品是否存在
	if rc.ProductID != nil {
		_, err := s.productRepo.GetByID(ctx, *rc.ProductID)
		if err != nil {
			return err
		}
	}

	// 检查代码是否已存在
	existing, err := s.couponRepo.GetRedeemCodeByCode(ctx, rc.Code)
	if err == nil && existing != nil {
		return apperr.BadRequest("兑换码已存在")
	}

	return s.couponRepo.CreateRedeemCode(ctx, rc)
}

// GetRedeemCodeByCode 根据代码获取兑换码
func (s *CouponService) GetRedeemCodeByCode(ctx context.Context, code string) (*coupon.RedeemCode, error) {
	return s.couponRepo.GetRedeemCodeByCode(ctx, code)
}

// RedeemCode 使用兑换码
func (s *CouponService) RedeemCode(ctx context.Context, code string, userID uuid.UUID) (*coupon.RedeemCode, error) {
	// 获取兑换码
	rc, err := s.couponRepo.GetRedeemCodeByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查是否已使用
	if rc.UsedBy != nil {
		return nil, apperr.BadRequest("兑换码已被使用")
	}

	// 检查是否激活
	if !rc.IsActive {
		return nil, apperr.BadRequest("兑换码已失效")
	}

	// 检查是否过期
	if rc.ExpiresAt != nil && rc.ExpiresAt.Before(rc.CreatedAt) {
		return nil, apperr.BadRequest("兑换码已过期")
	}

	// 使用兑换码
	if err := s.couponRepo.UseRedeemCode(ctx, code, userID); err != nil {
		return nil, err
	}

	// 重新获取兑换码（包含使用信息）
	return s.couponRepo.GetRedeemCodeByCode(ctx, code)
}

// BatchCreateRedeemCodes 批量创建兑换码
func (s *CouponService) BatchCreateRedeemCodes(ctx context.Context, productID *uuid.UUID, description string, count int, expiresAt *string) ([]*coupon.RedeemCode, error) {
	if count <= 0 || count > 1000 {
		return nil, apperr.BadRequest("批量创建数量必须在1-1000之间")
	}

	// 如果关联了商品，检查商品是否存在
	if productID != nil {
		_, err := s.productRepo.GetByID(ctx, *productID)
		if err != nil {
			return nil, err
		}
	}

	codes := make([]*coupon.RedeemCode, count)
	for i := 0; i < count; i++ {
		code := GenerateRedeemCode()
		codes[i] = &coupon.RedeemCode{
			Code:        code,
			ProductID:   productID,
			Description: description,
			IsActive:    true,
		}
	}

	if err := s.couponRepo.BatchCreateRedeemCodes(ctx, codes); err != nil {
		return nil, err
	}

	return codes, nil
}

// GenerateRedeemCode 生成兑换码
// 格式：XXXX-XXXX-XXXX（共12位，Base32编码，去除易混淆字符）
func GenerateRedeemCode() string {
	// 生成8字节随机数
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// Base32编码
	encoded := base32.StdEncoding.EncodeToString(b)

	// 去除填充字符和易混淆字符（0, O, I, L, 1）
	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.ReplaceAll(encoded, "0", "")
	encoded = strings.ReplaceAll(encoded, "O", "")
	encoded = strings.ReplaceAll(encoded, "I", "")
	encoded = strings.ReplaceAll(encoded, "L", "")
	encoded = strings.ReplaceAll(encoded, "1", "")

	// 取前12位
	if len(encoded) < 12 {
		// 如果不够12位，递归重新生成
		return GenerateRedeemCode()
	}
	encoded = encoded[:12]

	// 格式化为 XXXX-XXXX-XXXX
	return fmt.Sprintf("%s-%s-%s", encoded[0:4], encoded[4:8], encoded[8:12])
}
