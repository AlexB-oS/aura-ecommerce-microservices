// Package postgres is the persistence adapter for the product service. The
// generic pool + migration runner live in pkg/db; this file only contains the
// product-specific queries.
package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/aura-commerce/platform/services/product/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

const columns = `id, slug, name, tagline, description, category,
	price_cents, currency, stock, badge, active, created_at, updated_at`

// ProductRepository is the Postgres-backed implementation of the repository
// port declared by the service layer.
type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

// Create inserts a product. A duplicate slug surfaces as domain.ErrSlugTaken.
func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	const q = `
		INSERT INTO products
			(id, slug, name, tagline, description, category, price_cents, currency, stock, badge, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, q,
		p.ID, p.Slug, p.Name, p.Tagline, p.Description, p.Category,
		p.PriceCents, p.Currency, p.Stock, p.Badge, p.Active,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.ErrSlugTaken
		}
		return err
	}
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return r.scanOne(ctx, `SELECT `+columns+` FROM products WHERE id = $1`, id)
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return r.scanOne(ctx, `SELECT `+columns+` FROM products WHERE slug = $1`, slug)
}

// List returns products matching the filter. A `next` closure appends each
// argument and hands back its positional placeholder ($1, $2, …), so the SQL
// and the args slice can never fall out of sync. Go evaluates `+` operands
// left-to-right, which keeps the placeholder numbering correct.
func (r *ProductRepository) List(ctx context.Context, f domain.Filter) ([]domain.Product, error) {
	where := []string{"active = true"}
	args := make([]any, 0, 4)
	next := func(v any) string {
		args = append(args, v)
		return "$" + itoa(len(args))
	}

	if f.Category != "" {
		where = append(where, "category = "+next(f.Category))
	}
	if f.Search != "" {
		term := "%" + f.Search + "%"
		where = append(where, "(name ILIKE "+next(term)+" OR tagline ILIKE "+next(term)+")")
	}
	if f.InStockOnly {
		where = append(where, "stock > 0")
	}

	query := `SELECT ` + columns + ` FROM products WHERE ` + strings.Join(where, " AND ") +
		` ORDER BY created_at DESC LIMIT ` + next(f.Limit) + ` OFFSET ` + next(f.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		var p domain.Product
		if err := scanInto(rows, &p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// AdjustStock atomically changes stock by delta (negative to decrement). It
// refuses to let stock go below zero, returning ErrInsufficientStock, and
// returns ErrProductNotFound for an unknown id.
func (r *ProductRepository) AdjustStock(ctx context.Context, id string, delta int) (int, error) {
	const q = `
		UPDATE products SET stock = stock + $2, updated_at = now()
		WHERE id = $1 AND stock + $2 >= 0
		RETURNING stock`
	var newStock int
	err := r.pool.QueryRow(ctx, q, id, delta).Scan(&newStock)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if e := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, id).Scan(&exists); e != nil {
			return 0, e
		}
		if !exists {
			return 0, domain.ErrProductNotFound
		}
		return 0, domain.ErrInsufficientStock
	}
	return newStock, err
}

func (r *ProductRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Product, error) {
	var p domain.Product
	err := scanInto(r.pool.QueryRow(ctx, query, args...), &p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// row is satisfied by both pgx.Row (QueryRow) and pgx.Rows (Query loop).
type row interface {
	Scan(dest ...any) error
}

func scanInto(r row, p *domain.Product) error {
	return r.Scan(
		&p.ID, &p.Slug, &p.Name, &p.Tagline, &p.Description, &p.Category,
		&p.PriceCents, &p.Currency, &p.Stock, &p.Badge, &p.Active,
		&p.CreatedAt, &p.UpdatedAt,
	)
}

// itoa is a tiny local int->string to keep the query builder readable without
// importing strconv at every call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
