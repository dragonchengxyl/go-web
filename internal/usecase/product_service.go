package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/studio/platform/internal/domain/product"
)

type ProductService struct {
	productRepo product.Repository
}

func NewProductService(productRepo product.Repository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, p *product.Product) error {
	// 检查SKU是否已存在
	existing, err := s.productRepo.GetBySKU(ctx, p.SKU)
	if err == nil && existing != nil {
		return fmt.Errorf("SKU already exists: %s", p.SKU)
	}

	return s.productRepo.Create(ctx, p)
}

// GetProduct 获取商品详情
func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU 根据SKU获取商品
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*product.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

// ListProducts 获取商品列表
func (s *ProductService) ListProducts(ctx context.Context, filter product.ListFilter) ([]*product.Product, int, error) {
	return s.productRepo.List(ctx, filter)
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, p *product.Product) error {
	// 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, p.ID)
	if err != nil {
		return err
	}

	return s.productRepo.Update(ctx, p)
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.productRepo.Delete(ctx, id)
}

// GetActiveDiscounts 获取商品的有效折扣
func (s *ProductService) GetActiveDiscounts(ctx context.Context, productID uuid.UUID) ([]*product.DiscountRule, error) {
	return s.productRepo.GetActiveDiscounts(ctx, productID)
}

// CreateDiscount 创建折扣规则
func (s *ProductService) CreateDiscount(ctx context.Context, rule *product.DiscountRule) error {
	// 检查商品是否存在
	_, err := s.productRepo.GetByID(ctx, rule.ProductID)
	if err != nil {
		return err
	}

	return s.productRepo.CreateDiscount(ctx, rule)
}

// CalculatePrice 计算商品实际价格（应用折扣）
func (s *ProductService) CalculatePrice(ctx context.Context, productID uuid.UUID) (int, error) {
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return 0, err
	}

	// 获取有效折扣
	discounts, err := s.productRepo.GetActiveDiscounts(ctx, productID)
	if err != nil {
		return 0, err
	}

	// 如果没有折扣，返回原价
	if len(discounts) == 0 {
		return p.PriceCents, nil
	}

	// 应用第一个折扣（已按折扣力度排序）
	discount := discounts[0]
	finalPrice := p.PriceCents

	switch discount.DiscountType {
	case product.DiscountTypePercentage:
		// 百分比折扣
		finalPrice = int(float64(p.PriceCents) * (1 - discount.DiscountValue/100))
	case product.DiscountTypeFixed:
		// 固定金额折扣
		finalPrice = p.PriceCents - int(discount.DiscountValue*100)
		if finalPrice < 0 {
			finalPrice = 0
		}
	}

	return finalPrice, nil
}
