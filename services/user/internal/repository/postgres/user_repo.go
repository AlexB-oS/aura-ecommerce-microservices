package postgres

import (
	"context"
	"errors"

	"github.com/aura-commerce/platform/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolation is Postgres' SQLSTATE for a unique-constraint breach. We map
// it onto a domain error so the service layer never needs to know SQL codes.
const uniqueViolation = "23505"

// UserRepository is the Postgres-backed implementation of the repository port
// declared by the service layer.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository wires the repository to a connection pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user. A duplicate email surfaces as domain.ErrEmailTaken.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (id, email, name, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, u.ID, u.Email, u.Name, u.PasswordHash, u.Role).
		Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return domain.ErrEmailTaken
		}
		return err
	}
	return nil
}

// GetByID returns a user by id, or domain.ErrUserNotFound.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1`
	return r.scanOne(ctx, q, id)
}

// GetByEmail returns a user by email (case-insensitive), or ErrUserNotFound.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users WHERE email = lower($1)`
	return r.scanOne(ctx, q, email)
}

// Update persists mutable profile fields and bumps updated_at.
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE users SET name = $2, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q, u.ID, u.Name).Scan(&u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	}
	return err
}

func (r *UserRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
