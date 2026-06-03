package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/aura-commerce/platform/services/product/internal/domain"
	"github.com/aura-commerce/platform/services/product/internal/service"
)

// fakeRepo is an in-memory implementation of service.Repository so the business
// logic can be tested with no database.
type fakeRepo struct {
	mu         sync.Mutex
	byID       map[string]*domain.Product
	bySlug     map[string]*domain.Product
	lastFilter domain.Filter
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*domain.Product{}, bySlug: map[string]*domain.Product{}}
}

func (f *fakeRepo) Create(_ context.Context, p *domain.Product) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.bySlug[p.Slug]; ok {
		return domain.ErrSlugTaken
	}
	cp := *p
	f.byID[p.ID] = &cp
	f.bySlug[p.Slug] = &cp
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id string) (*domain.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p, ok := f.byID[id]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, domain.ErrProductNotFound
}

func (f *fakeRepo) GetBySlug(_ context.Context, slug string) (*domain.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p, ok := f.bySlug[slug]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, domain.ErrProductNotFound
}

func (f *fakeRepo) List(_ context.Context, filter domain.Filter) ([]domain.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastFilter = filter
	out := make([]domain.Product, 0, len(f.byID))
	for _, p := range f.byID {
		out = append(out, *p)
	}
	return out, nil
}

func (f *fakeRepo) AdjustStock(_ context.Context, id string, delta int) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byID[id]
	if !ok {
		return 0, domain.ErrProductNotFound
	}
	if p.Stock+delta < 0 {
		return 0, domain.ErrInsufficientStock
	}
	p.Stock += delta
	return p.Stock, nil
}

// recordingPublisher captures published subjects so tests can assert on events.
type recordingPublisher struct {
	mu       sync.Mutex
	subjects []string
}

func (r *recordingPublisher) Publish(subject, _ string, _ any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subjects = append(r.subjects, subject)
	return nil
}

func newService(repo service.Repository, pub service.Publisher) *service.Service {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.New(repo, pub, log, 24)
}

func TestCreate_Success(t *testing.T) {
	pub := &recordingPublisher{}
	svc := newService(newFakeRepo(), pub)

	p, err := svc.Create(context.Background(), service.CreateInput{
		Name: "Drift Wireless Headphones", Category: "Audio", PriceCents: 21900, Stock: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Slug != "drift-wireless-headphones" {
		t.Errorf("slug should be generated from the name, got %q", p.Slug)
	}
	if p.Currency != "USD" || !p.Active {
		t.Errorf("expected sensible defaults (USD, active), got %q active=%v", p.Currency, p.Active)
	}
	if len(pub.subjects) != 1 || pub.subjects[0] != "product.created" {
		t.Errorf("expected a product.created event, got %v", pub.subjects)
	}
}

func TestCreate_InvalidPrice(t *testing.T) {
	svc := newService(newFakeRepo(), &recordingPublisher{})
	_, err := svc.Create(context.Background(), service.CreateInput{
		Name: "Freebie", Category: "Misc", PriceCents: 0,
	})
	if !errors.Is(err, domain.ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestCreate_DuplicateSlug(t *testing.T) {
	svc := newService(newFakeRepo(), &recordingPublisher{})
	ctx := context.Background()
	in := service.CreateInput{Name: "Same Name", Category: "Home", PriceCents: 100}
	if _, err := svc.Create(ctx, in); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if _, err := svc.Create(ctx, in); !errors.Is(err, domain.ErrSlugTaken) {
		t.Errorf("expected ErrSlugTaken on duplicate name, got %v", err)
	}
}

func TestList_AppliesDefaultPageSize(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo, &recordingPublisher{})
	if _, err := svc.List(context.Background(), domain.Filter{Limit: 0}); err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if repo.lastFilter.Limit != 24 {
		t.Errorf("expected default page size 24 to be applied, got %d", repo.lastFilter.Limit)
	}
}

func TestAdjustStock_LowStockEmitsEvent(t *testing.T) {
	repo := newFakeRepo()
	pub := &recordingPublisher{}
	svc := newService(repo, pub)
	ctx := context.Background()

	created, err := svc.Create(ctx, service.CreateInput{
		Name: "Ember Mug", Category: "Home", PriceCents: 9900, Stock: 6,
	})
	if err != nil {
		t.Fatalf("setup create failed: %v", err)
	}

	// 6 - 2 = 4, which is <= lowStockThreshold (5), so a stock_low event fires.
	if _, err := svc.AdjustStock(ctx, created.ID, -2); err != nil {
		t.Fatalf("adjust failed: %v", err)
	}

	var sawLow bool
	for _, s := range pub.subjects {
		if s == "product.stock_low" {
			sawLow = true
		}
	}
	if !sawLow {
		t.Errorf("expected a product.stock_low event, got %v", pub.subjects)
	}

	// Reducing below zero must fail.
	if _, err := svc.AdjustStock(ctx, created.ID, -100); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}
