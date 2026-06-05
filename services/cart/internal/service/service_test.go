package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/aura-commerce/platform/services/cart/internal/domain"
	"github.com/aura-commerce/platform/services/cart/internal/service"
)

// fakeRepo is an in-memory cart store.
type fakeRepo struct{ carts map[string]*domain.Cart }

func newFakeRepo() *fakeRepo { return &fakeRepo{carts: map[string]*domain.Cart{}} }

func (f *fakeRepo) Get(_ context.Context, userID string) (*domain.Cart, error) {
	if c, ok := f.carts[userID]; ok {
		cp := *c
		cp.Items = append([]domain.CartItem(nil), c.Items...)
		return &cp, nil
	}
	return &domain.Cart{UserID: userID, Items: []domain.CartItem{}}, nil
}
func (f *fakeRepo) Save(_ context.Context, c *domain.Cart) error {
	cp := *c
	cp.Items = append([]domain.CartItem(nil), c.Items...)
	f.carts[c.UserID] = &cp
	return nil
}
func (f *fakeRepo) Delete(_ context.Context, userID string) error {
	delete(f.carts, userID)
	return nil
}

// fakeCatalog returns canned products and records misses.
type fakeCatalog struct{ products map[string]*domain.ProductInfo }

func (f *fakeCatalog) Lookup(_ context.Context, slug string) (*domain.ProductInfo, error) {
	if p, ok := f.products[slug]; ok {
		return p, nil
	}
	return nil, domain.ErrProductUnavailable
}

func newService() *service.Service {
	cat := &fakeCatalog{products: map[string]*domain.ProductInfo{
		"drift-headphones": {ID: "p1", Slug: "drift-headphones", Name: "Drift", PriceCents: 21900, Stock: 5},
	}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.New(newFakeRepo(), cat, log)
}

func TestAddItem_NewThenMerge(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	cart, err := svc.AddItem(ctx, "u1", "drift-headphones", 1)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if len(cart.Items) != 1 || cart.Items[0].Qty != 1 {
		t.Fatalf("expected 1 item qty 1, got %+v", cart.Items)
	}
	if cart.Items[0].PriceCents != 21900 {
		t.Errorf("price should be snapshotted from the catalog, got %d", cart.Items[0].PriceCents)
	}

	cart, _ = svc.AddItem(ctx, "u1", "drift-headphones", 2)
	if len(cart.Items) != 1 || cart.Items[0].Qty != 3 {
		t.Errorf("adding the same product should merge to qty 3, got %+v", cart.Items)
	}
	if cart.TotalCents() != 3*21900 {
		t.Errorf("unexpected total %d", cart.TotalCents())
	}
}

func TestAddItem_Unavailable(t *testing.T) {
	svc := newService()
	if _, err := svc.AddItem(context.Background(), "u1", "ghost", 1); !errors.Is(err, domain.ErrProductUnavailable) {
		t.Errorf("expected ErrProductUnavailable, got %v", err)
	}
}

func TestAddItem_InvalidQty(t *testing.T) {
	svc := newService()
	if _, err := svc.AddItem(context.Background(), "u1", "drift-headphones", 0); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestSetItemQty_RemovesAtZero(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	if _, err := svc.AddItem(ctx, "u1", "drift-headphones", 2); err != nil {
		t.Fatalf("setup add failed: %v", err)
	}

	cart, err := svc.SetItemQty(ctx, "u1", "p1", 0)
	if err != nil {
		t.Fatalf("set qty failed: %v", err)
	}
	if len(cart.Items) != 0 {
		t.Errorf("qty 0 should remove the item, got %+v", cart.Items)
	}

	if _, err := svc.SetItemQty(ctx, "u1", "missing", 1); !errors.Is(err, domain.ErrItemNotInCart) {
		t.Errorf("expected ErrItemNotInCart, got %v", err)
	}
}
