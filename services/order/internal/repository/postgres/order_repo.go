// Package postgres is the order service's persistence adapter. An order and its
// items are written together in a single transaction so an order is never
// half-persisted.
package postgres

import (
	"context"
	"errors"

	"github.com/aura-commerce/platform/services/order/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// Create inserts the order and all its items atomically.
func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful commit

	err = tx.QueryRow(ctx, `
		INSERT INTO orders (id, user_id, status, total_cents, currency)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING created_at, updated_at`,
		o.ID, o.UserID, o.Status, o.TotalCents, o.Currency,
	).Scan(&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return err
	}

	for _, it := range o.Items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, slug, name, price_cents, qty)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			o.ID, it.ProductID, it.Slug, it.Name, it.PriceCents, it.Qty,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID returns an order with its items, scoped to a user so one customer can
// never read another's order.
func (r *OrderRepository) GetByID(ctx context.Context, userID, id string) (*domain.Order, error) {
	var o domain.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, total_cents, currency, created_at, updated_at
		FROM orders WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&o.ID, &o.UserID, &o.Status, &o.TotalCents, &o.Currency, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT product_id, slug, name, price_cents, qty
		FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ProductID, &it.Slug, &it.Name, &it.PriceCents, &it.Qty); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}

// ListByUser returns a user's orders newest-first, without line items (the list
// view only needs status + total).
func (r *OrderRepository) ListByUser(ctx context.Context, userID string) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status, total_cents, currency, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalCents, &o.Currency, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// UpdateStatus moves an order to a new status (called by the saga subscriber).
func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
