package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/pkg/apperr"
)

type orderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) order.Repository {
	return &orderRepo{db: db}
}

const createOrderSQL = `
	INSERT INTO orders (id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code, payment_method, idempotency_key, metadata, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING created_at, updated_at
`

func (r *orderRepo) Create(ctx context.Context, o *order.Order) error {
	o.ID = uuid.New()

	var metadataJSON []byte
	var err error
	if o.Metadata != nil {
		metadataJSON, err = json.Marshal(o.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	err = r.db.QueryRow(ctx, createOrderSQL,
		o.ID, o.OrderNo, o.UserID, o.Status, o.TotalCents, o.Currency,
		o.DiscountCents, o.CouponCode, o.PaymentMethod, o.IdempotencyKey,
		metadataJSON, o.ExpiresAt,
	).Scan(&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

const getOrderByIDSQL = `
	SELECT id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code, payment_method, paid_at, idempotency_key, metadata, created_at, expires_at, updated_at
	FROM orders WHERE id = $1
`

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	var o order.Order
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, getOrderByIDSQL, id).Scan(
		&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalCents, &o.Currency,
		&o.DiscountCents, &o.CouponCode, &o.PaymentMethod, &o.PaidAt,
		&o.IdempotencyKey, &metadataJSON, &o.CreatedAt, &o.ExpiresAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("order", "id", id.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &o.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &o, nil
}

const getOrderByOrderNoSQL = `
	SELECT id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code, payment_method, paid_at, idempotency_key, metadata, created_at, expires_at, updated_at
	FROM orders WHERE order_no = $1
`

func (r *orderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*order.Order, error) {
	var o order.Order
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, getOrderByOrderNoSQL, orderNo).Scan(
		&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalCents, &o.Currency,
		&o.DiscountCents, &o.CouponCode, &o.PaymentMethod, &o.PaidAt,
		&o.IdempotencyKey, &metadataJSON, &o.CreatedAt, &o.ExpiresAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("order", "order_no", orderNo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by order_no: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &o.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &o, nil
}

const getOrderByIdempotencyKeySQL = `
	SELECT id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code, payment_method, paid_at, idempotency_key, metadata, created_at, expires_at, updated_at
	FROM orders WHERE idempotency_key = $1
`

func (r *orderRepo) GetByIdempotencyKey(ctx context.Context, key string) (*order.Order, error) {
	var o order.Order
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, getOrderByIdempotencyKeySQL, key).Scan(
		&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalCents, &o.Currency,
		&o.DiscountCents, &o.CouponCode, &o.PaymentMethod, &o.PaidAt,
		&o.IdempotencyKey, &metadataJSON, &o.CreatedAt, &o.ExpiresAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // 幂等性键不存在时返回nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by idempotency_key: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &o.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &o, nil
}

const listOrdersByUserIDSQL = `
	SELECT id, order_no, user_id, status, total_cents, currency, discount_cents, coupon_code, payment_method, paid_at, idempotency_key, metadata, created_at, expires_at, updated_at
	FROM orders WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
`

const countOrdersByUserIDSQL = `SELECT COUNT(*) FROM orders WHERE user_id = $1`

func (r *orderRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*order.Order, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, countOrdersByUserIDSQL, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, listOrdersByUserIDSQL, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	orders := []*order.Order{}
	for rows.Next() {
		var o order.Order
		var metadataJSON []byte
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Status, &o.TotalCents, &o.Currency,
			&o.DiscountCents, &o.CouponCode, &o.PaymentMethod, &o.PaidAt,
			&o.IdempotencyKey, &metadataJSON, &o.CreatedAt, &o.ExpiresAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &o.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		orders = append(orders, &o)
	}

	return orders, total, nil
}

const updateOrderSQL = `
	UPDATE orders SET status = $2, total_cents = $3, discount_cents = $4, coupon_code = $5, payment_method = $6, paid_at = $7, metadata = $8, updated_at = NOW()
	WHERE id = $1
`

func (r *orderRepo) Update(ctx context.Context, o *order.Order) error {
	var metadataJSON []byte
	var err error
	if o.Metadata != nil {
		metadataJSON, err = json.Marshal(o.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	_, err = r.db.Exec(ctx, updateOrderSQL,
		o.ID, o.Status, o.TotalCents, o.DiscountCents, o.CouponCode,
		o.PaymentMethod, o.PaidAt, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

const updateOrderStatusSQL = `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status order.OrderStatus) error {
	_, err := r.db.Exec(ctx, updateOrderStatusSQL, id, status)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

const createOrderItemSQL = `
	INSERT INTO order_items (id, order_id, product_id, price_cents, quantity)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING created_at
`

func (r *orderRepo) CreateItem(ctx context.Context, item *order.OrderItem) error {
	item.ID = uuid.New()
	err := r.db.QueryRow(ctx, createOrderItemSQL,
		item.ID, item.OrderID, item.ProductID, item.PriceCents, item.Quantity,
	).Scan(&item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	return nil
}

const getOrderItemsByOrderIDSQL = `
	SELECT id, order_id, product_id, price_cents, quantity, created_at
	FROM order_items WHERE order_id = $1
`

func (r *orderRepo) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*order.OrderItem, error) {
	rows, err := r.db.Query(ctx, getOrderItemsByOrderIDSQL, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	items := []*order.OrderItem{}
	for rows.Next() {
		var item order.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.PriceCents, &item.Quantity, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

const cancelExpiredOrdersSQL = `
	UPDATE orders SET status = 'cancelled', updated_at = NOW()
	WHERE status = 'pending_payment' AND expires_at < NOW()
`

func (r *orderRepo) CancelExpiredOrders(ctx context.Context) (int, error) {
	result, err := r.db.Exec(ctx, cancelExpiredOrdersSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel expired orders: %w", err)
	}
	return int(result.RowsAffected()), nil
}
