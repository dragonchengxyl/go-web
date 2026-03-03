package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-web/internal/domain/product"
	"go-web/pkg/apperr"
)

type productRepo struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) product.Repository {
	return &productRepo{db: db}
}

const createProductSQL = `
	INSERT INTO products (id, sku, name, description, product_type, entity_id, price_cents, currency, original_price_cents, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING created_at, updated_at
`

func (r *productRepo) Create(ctx context.Context, p *product.Product) error {
	p.ID = uuid.New()
	err := r.db.QueryRow(ctx, createProductSQL,
		p.ID, p.SKU, p.Name, p.Description, p.ProductType, p.EntityID,
		p.PriceCents, p.Currency, p.OriginalPriceCents, p.IsActive,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

const getProductByIDSQL = `
	SELECT id, sku, name, description, product_type, entity_id, price_cents, currency, original_price_cents, is_active, created_at, updated_at
	FROM products WHERE id = $1
`

func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	var p product.Product
	err := r.db.QueryRow(ctx, getProductByIDSQL, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.ProductType, &p.EntityID,
		&p.PriceCents, &p.Currency, &p.OriginalPriceCents, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("product", "id", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &p, nil
}

const getProductBySKUSQL = `
	SELECT id, sku, name, description, product_type, entity_id, price_cents, currency, original_price_cents, is_active, created_at, updated_at
	FROM products WHERE sku = $1
`

func (r *productRepo) GetBySKU(ctx context.Context, sku string) (*product.Product, error) {
	var p product.Product
	err := r.db.QueryRow(ctx, getProductBySKUSQL, sku).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.ProductType, &p.EntityID,
		&p.PriceCents, &p.Currency, &p.OriginalPriceCents, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("product", "sku", sku)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product by sku: %w", err)
	}
	return &p, nil
}

func (r *productRepo) List(ctx context.Context, filter product.ListFilter) ([]*product.Product, int, error) {
	query := `SELECT id, sku, name, description, product_type, entity_id, price_cents, currency, original_price_cents, is_active, created_at, updated_at FROM products WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM products WHERE 1=1`
	args := []any{}
	argPos := 1

	if filter.ProductType != nil {
		query += fmt.Sprintf(" AND product_type = $%d", argPos)
		countQuery += fmt.Sprintf(" AND product_type = $%d", argPos)
		args = append(args, *filter.ProductType)
		argPos++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		countQuery += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filter.IsActive)
		argPos++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos)
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := []*product.Product{}
	for rows.Next() {
		var p product.Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.ProductType, &p.EntityID,
			&p.PriceCents, &p.Currency, &p.OriginalPriceCents, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &p)
	}

	return products, total, nil
}

const updateProductSQL = `
	UPDATE products SET name = $2, description = $3, product_type = $4, entity_id = $5,
		price_cents = $6, currency = $7, original_price_cents = $8, is_active = $9, updated_at = NOW()
	WHERE id = $1
`

func (r *productRepo) Update(ctx context.Context, p *product.Product) error {
	_, err := r.db.Exec(ctx, updateProductSQL,
		p.ID, p.Name, p.Description, p.ProductType, p.EntityID,
		p.PriceCents, p.Currency, p.OriginalPriceCents, p.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

const deleteProductSQL = `DELETE FROM products WHERE id = $1`

func (r *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, deleteProductSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

const getActiveDiscountsSQL = `
	SELECT id, product_id, discount_type, discount_value, start_at, end_at, max_uses, used_count, created_at
	FROM discount_rules
	WHERE product_id = $1 AND start_at <= NOW() AND end_at >= NOW()
		AND (max_uses IS NULL OR used_count < max_uses)
	ORDER BY discount_value DESC
`

func (r *productRepo) GetActiveDiscounts(ctx context.Context, productID uuid.UUID) ([]*product.DiscountRule, error) {
	rows, err := r.db.Query(ctx, getActiveDiscountsSQL, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active discounts: %w", err)
	}
	defer rows.Close()

	rules := []*product.DiscountRule{}
	for rows.Next() {
		var rule product.DiscountRule
		if err := rows.Scan(&rule.ID, &rule.ProductID, &rule.DiscountType, &rule.DiscountValue,
			&rule.StartAt, &rule.EndAt, &rule.MaxUses, &rule.UsedCount, &rule.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan discount rule: %w", err)
		}
		rules = append(rules, &rule)
	}

	return rules, nil
}

const createDiscountSQL = `
	INSERT INTO discount_rules (id, product_id, discount_type, discount_value, start_at, end_at, max_uses, used_count)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING created_at
`

func (r *productRepo) CreateDiscount(ctx context.Context, rule *product.DiscountRule) error {
	rule.ID = uuid.New()
	err := r.db.QueryRow(ctx, createDiscountSQL,
		rule.ID, rule.ProductID, rule.DiscountType, rule.DiscountValue,
		rule.StartAt, rule.EndAt, rule.MaxUses, rule.UsedCount,
	).Scan(&rule.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create discount rule: %w", err)
	}
	return nil
}

