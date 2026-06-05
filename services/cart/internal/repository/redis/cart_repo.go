// Package redis is the cart service's persistence adapter. Carts are stored as
// a single JSON blob per user under the key "cart:{userID}", with a TTL so
// abandoned carts expire on their own. Redis is ideal here: carts are small,
// hot, and disposable.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aura-commerce/platform/services/cart/internal/domain"
	"github.com/redis/go-redis/v9"
)

// NewClient parses a redis:// URL and returns a client (not yet connected;
// the first command connects).
func NewClient(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opt), nil
}

// CartRepository persists carts in Redis.
type CartRepository struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewCartRepository(rdb *redis.Client, ttl time.Duration) *CartRepository {
	return &CartRepository{rdb: rdb, ttl: ttl}
}

func key(userID string) string { return "cart:" + userID }

// Get returns the user's cart, or a fresh empty cart if none is stored (callers
// never have to distinguish "no cart" from "empty cart").
func (r *CartRepository) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	val, err := r.rdb.Get(ctx, key(userID)).Result()
	if errors.Is(err, redis.Nil) {
		return &domain.Cart{UserID: userID, Items: []domain.CartItem{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cart domain.Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return nil, err
	}
	if cart.Items == nil {
		cart.Items = []domain.CartItem{}
	}
	return &cart, nil
}

// Save writes the cart and (re)sets its TTL.
func (r *CartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	cart.UpdatedAt = time.Now().UTC()
	raw, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key(cart.UserID), raw, r.ttl).Err()
}

// Delete removes the user's cart entirely.
func (r *CartRepository) Delete(ctx context.Context, userID string) error {
	return r.rdb.Del(ctx, key(userID)).Err()
}

// Ping reports Redis reachability for readiness checks.
func (r *CartRepository) Ping(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}
