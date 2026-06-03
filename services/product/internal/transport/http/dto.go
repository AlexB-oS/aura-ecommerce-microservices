package http

import (
	"time"

	"github.com/aura-commerce/platform/services/product/internal/domain"
)

// --- Requests --------------------------------------------------------------

type createProductRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=120"`
	Tagline     string `json:"tagline"     validate:"max=160"`
	Description string `json:"description" validate:"max=2000"`
	Category    string `json:"category"    validate:"required,min=2,max=60"`
	PriceCents  int    `json:"priceCents"  validate:"required,gt=0"`
	Stock       int    `json:"stock"       validate:"gte=0"`
	Badge       string `json:"badge"       validate:"max=40"`
}

type adjustStockRequest struct {
	// Positive to restock, negative to reduce. Zero is rejected as a no-op.
	Delta int `json:"delta" validate:"required"`
}

// --- Responses -------------------------------------------------------------

type productResponse struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Tagline     string    `json:"tagline"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	PriceCents  int        `json:"priceCents"`
	Currency    string    `json:"currency"`
	Stock       int        `json:"stock"`
	Badge       string    `json:"badge,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toProductResponse(p *domain.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		Slug:        p.Slug,
		Name:        p.Name,
		Tagline:     p.Tagline,
		Description: p.Description,
		Category:    p.Category,
		PriceCents:  p.PriceCents,
		Currency:    p.Currency,
		Stock:       p.Stock,
		Badge:       p.Badge,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProductResponses(products []domain.Product) []productResponse {
	out := make([]productResponse, len(products))
	for i := range products {
		out[i] = toProductResponse(&products[i])
	}
	return out
}
