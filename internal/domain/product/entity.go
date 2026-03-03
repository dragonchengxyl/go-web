package product

import (
	"time"

	"github.com/google/uuid"
)

// ProductType 商品类型
type ProductType string

const (
	ProductTypeGame       ProductType = "game"
	ProductTypeDLC        ProductType = "dlc"
	ProductTypeOST        ProductType = "ost"
	ProductTypeBundle     ProductType = "bundle"
	ProductTypeMembership ProductType = "membership"
)

// Product 商品实体
type Product struct {
	ID                 uuid.UUID   `json:"id"`
	SKU                string      `json:"sku"`
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	ProductType        ProductType `json:"product_type"`
	EntityID           *uuid.UUID  `json:"entity_id"`           // 关联的 game_id / dlc_id / album_id
	PriceCents         int         `json:"price_cents"`         // 定价（分）
	Currency           string      `json:"currency"`            // 货币代码
	OriginalPriceCents *int        `json:"original_price_cents"` // 原价（用于划线价展示）
	IsActive           bool        `json:"is_active"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// DiscountType 折扣类型
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypeBuyXGetY   DiscountType = "buy_x_get_y"
)

// DiscountRule 折扣规则
type DiscountRule struct {
	ID            uuid.UUID    `json:"id"`
	ProductID     uuid.UUID    `json:"product_id"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue float64      `json:"discount_value"`
	StartAt       time.Time    `json:"start_at"`
	EndAt         time.Time    `json:"end_at"`
	MaxUses       *int         `json:"max_uses"`
	UsedCount     int          `json:"used_count"`
	CreatedAt     time.Time    `json:"created_at"`
}
