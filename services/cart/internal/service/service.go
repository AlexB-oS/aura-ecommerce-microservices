// Package service holds the cart's business logic. It depends on two ports: a
// Repository (Redis) and a Catalog (the product service). Both are interfaces,
// so the logic is unit-testable with fakes.
package service

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/services/cart/internal/domain"
)

// Repository is the persistence port (Redis adapter implements it).
type Repository interface {
	Get(ctx context.Context, userID string) (*domain.Cart, error)
	Save(ctx context.Context, cart *domain.Cart) error
	Delete(ctx context.Context, userID string) error
}

// Catalog is the product-lookup port (the catalog REST client implements it).
type Catalog interface {
	Lookup(ctx context.Context, slug string) (*domain.ProductInfo, error)
}

type Service struct {
	repo    Repository
	catalog Catalog
	log     *slog.Logger
}

func New(repo Repository, catalog Catalog, log *slog.Logger) *Service {
	return &Service{repo: repo, catalog: catalog, log: log}
}

// Get returns the user's cart (empty if they have none).
func (s *Service) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	return s.repo.Get(ctx, userID)
}

// AddItem adds qty of a product (by slug) to the cart. The name and price are
// taken from the product service, never trusted from the client. Adding a
// product already in the cart increases its quantity.
func (s *Service) AddItem(ctx context.Context, userID, slug string, qty int) (*domain.Cart, error) {
	if qty <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	info, err := s.catalog.Lookup(ctx, slug)
	if err != nil {
		return nil, err // includes domain.ErrProductUnavailable
	}

	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range cart.Items {
		if cart.Items[i].ProductID == info.ID {
			cart.Items[i].Qty += qty
			// refresh the price snapshot to the current catalog price
			cart.Items[i].PriceCents = info.PriceCents
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, domain.CartItem{
			ProductID:  info.ID,
			Slug:       info.Slug,
			Name:       info.Name,
			PriceCents: info.PriceCents,
			Qty:        qty,
		})
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// SetItemQty sets an item's quantity. A quantity of zero (or less) removes it.
func (s *Service) SetItemQty(ctx context.Context, userID, productID string, qty int) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	idx := -1
	for i := range cart.Items {
		if cart.Items[i].ProductID == productID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, domain.ErrItemNotInCart
	}

	if qty <= 0 {
		cart.Items = append(cart.Items[:idx], cart.Items[idx+1:]...)
	} else {
		cart.Items[idx].Qty = qty
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// RemoveItem deletes an item from the cart (idempotent: removing something that
// isn't there is not an error).
func (s *Service) RemoveItem(ctx context.Context, userID, productID string) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	kept := cart.Items[:0]
	for _, it := range cart.Items {
		if it.ProductID != productID {
			kept = append(kept, it)
		}
	}
	cart.Items = kept
	if err := s.repo.Save(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// Clear empties the cart entirely (used after a successful checkout).
func (s *Service) Clear(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}
