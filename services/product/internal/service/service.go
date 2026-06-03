// Package service holds the product catalog's business logic: listing &
// fetching products (public), creating them (admin), and adjusting stock. It
// depends only on the Repository and Publisher ports, so it is unit-testable
// without a database.
package service

import (
	"context"
	"log/slog"
	"regexp"
	"strings"

	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/services/product/internal/domain"
	"github.com/aura-commerce/platform/services/product/internal/events"
	"github.com/google/uuid"
)

// lowStockThreshold is the stock level at or below which a product.stock_low
// event is emitted after an adjustment.
const lowStockThreshold = 5

// Repository is the persistence port (implemented by the Postgres adapter).
type Repository interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	List(ctx context.Context, f domain.Filter) ([]domain.Product, error)
	AdjustStock(ctx context.Context, id string, delta int) (int, error)
}

// Publisher is the event-bus port (pkg/events.Bus satisfies it).
type Publisher interface {
	Publish(subject, correlationID string, payload any) error
}

// CreateInput is the data needed to add a product. The slug is derived from the
// name, and currency/active get sensible defaults in the service.
type CreateInput struct {
	Name        string
	Tagline     string
	Description string
	Category    string
	PriceCents  int
	Stock       int
	Badge       string
}

// Service is the use-case layer.
type Service struct {
	repo        Repository
	bus         Publisher
	log         *slog.Logger
	defaultPage int
	maxPage     int
}

// New wires the service. defaultPage is used when a list request omits a limit.
func New(repo Repository, bus Publisher, log *slog.Logger, defaultPage int) *Service {
	if defaultPage <= 0 {
		defaultPage = 24
	}
	return &Service{repo: repo, bus: bus, log: log, defaultPage: defaultPage, maxPage: 100}
}

// List returns active products for the catalog, clamping the page size to keep
// any single query bounded.
func (s *Service) List(ctx context.Context, f domain.Filter) ([]domain.Product, error) {
	if f.Limit <= 0 {
		f.Limit = s.defaultPage
	}
	if f.Limit > s.maxPage {
		f.Limit = s.maxPage
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.List(ctx, f)
}

// GetBySlug returns one product by its slug.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return s.repo.GetBySlug(ctx, strings.TrimSpace(slug))
}

// Create adds a new product (admin-only at the transport layer).
func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Product, error) {
	if in.PriceCents <= 0 {
		return nil, domain.ErrInvalidPrice
	}

	product := &domain.Product{
		ID:          uuid.NewString(),
		Slug:        slugify(in.Name),
		Name:        strings.TrimSpace(in.Name),
		Tagline:     strings.TrimSpace(in.Tagline),
		Description: strings.TrimSpace(in.Description),
		Category:    strings.TrimSpace(in.Category),
		PriceCents:  in.PriceCents,
		Currency:    "USD",
		Stock:       in.Stock,
		Badge:       strings.TrimSpace(in.Badge),
		Active:      true,
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err // includes domain.ErrSlugTaken
	}

	s.publish(ctx, events.SubjectProductCreated, events.ProductCreated{
		ID: product.ID, Slug: product.Slug, Name: product.Name, PriceCents: product.PriceCents,
	})
	return product, nil
}

// AdjustStock changes a product's stock (e.g. -1 when an order is placed) and
// emits a low-stock event when it crosses the threshold.
func (s *Service) AdjustStock(ctx context.Context, id string, delta int) (*domain.Product, error) {
	newStock, err := s.repo.AdjustStock(ctx, id, delta)
	if err != nil {
		return nil, err
	}
	if delta < 0 && newStock <= lowStockThreshold {
		// Re-read to include the slug in the event payload.
		if p, e := s.repo.GetByID(ctx, id); e == nil {
			s.publish(ctx, events.SubjectProductStockLow, events.ProductStockLow{
				ID: p.ID, Slug: p.Slug, Stock: newStock,
			})
		}
	}
	return s.repo.GetByID(ctx, id)
}

// ReduceStock decrements stock for one order line. It's the entry point used by
// the order.placed subscriber, which only needs success/failure (not the
// updated product), so it wraps AdjustStock and discards the result.
func (s *Service) ReduceStock(ctx context.Context, productID string, qty int) error {
	_, err := s.AdjustStock(ctx, productID, -qty)
	return err
}

func (s *Service) publish(ctx context.Context, subject string, payload any) {
	if s.bus == nil {
		return
	}
	if err := s.bus.Publish(subject, logger.RequestID(ctx), payload); err != nil {
		s.log.WarnContext(ctx, "failed to publish event", slog.String("subject", subject), slog.Any("error", err))
	}
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugify turns "Drift Wireless Headphones" into "drift-wireless-headphones".
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonSlug.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
