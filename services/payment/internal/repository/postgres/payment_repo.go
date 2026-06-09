// Package postgres is the payment service's persistence adapter.
package postgres

import (
	"context"

	"github.com/aura-commerce/platform/services/payment/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// Create records a payment attempt.
func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO payments (id, order_id, user_id, amount_cents, currency, status, provider)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at`,
		p.ID, p.OrderID, p.UserID, p.AmountCents, p.Currency, p.Status, p.Provider,
	).Scan(&p.CreatedAt)
}

// ExistsForOrder reports whether a payment already exists for an order. The
// subscriber uses this for idempotency: NATS may redeliver order.placed, and we
// must not charge an order twice.
func (r *PaymentRepository) ExistsForOrder(ctx context.Context, orderID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM payments WHERE order_id = $1)`, orderID).Scan(&exists)
	return exists, err
}

// ListByUser returns a user's payments newest-first.
func (r *PaymentRepository) ListByUser(ctx context.Context, userID string) ([]domain.Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, user_id, amount_cents, currency, status, provider, created_at
		FROM payments WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]domain.Payment, 0)
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.AmountCents, &p.Currency,
			&p.Status, &p.Provider, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
