package product

import (
	"context"

	"github.com/google/uuid"
)

// Repository 商品仓储接口
type Repository interface {
	// Create 创建商品
	Create(ctx context.Context, product *Product) error

	// GetByID 根据ID获取商品
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)

	// GetBySKU 根据SKU获取商品
	GetBySKU(ctx context.Context, sku string) (*Product, error)

	// List 获取商品列表
	List(ctx context.Context, filter ListFilter) ([]*Product, int, error)

	// Update 更新商品
	Update(ctx context.Context, product *Product) error

	// Delete 删除商品
	Delete(ctx context.Context, id uuid.UUID) error

	// GetActiveDiscounts 获取商品的有效折扣规则
	GetActiveDiscounts(ctx context.Context, productID uuid.UUID) ([]*DiscountRule, error)

	// CreateDiscount 创建折扣规则
	CreateDiscount(ctx context.Context, rule *DiscountRule) error
}

// ListFilter 商品列表过滤器
type ListFilter struct {
	Page        int
	PageSize    int
	ProductType *ProductType
	IsActive    *bool
	Search      string
}
