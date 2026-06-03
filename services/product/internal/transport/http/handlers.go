package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/validate"
	"github.com/aura-commerce/platform/services/product/internal/domain"
	"github.com/aura-commerce/platform/services/product/internal/service"
	"github.com/go-chi/chi/v5"
)

// Handlers adapts HTTP onto the product service. Thin by design: parse, call
// the service, map the result/error to JSON.
type Handlers struct {
	svc      *service.Service
	validate *validate.Validator
}

func NewHandlers(svc *service.Service, v *validate.Validator) *Handlers {
	return &Handlers{svc: svc, validate: v}
}

// List handles GET /api/v1/products?category=&q=&inStock=&limit=&offset=
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.Filter{
		Category:    q.Get("category"),
		Search:      q.Get("q"),
		InStockOnly: q.Get("inStock") == "true",
		Limit:       atoiDefault(q.Get("limit"), 0),
		Offset:      atoiDefault(q.Get("offset"), 0),
	}

	products, err := h.svc.List(r.Context(), filter)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toProductResponses(products))
}

// GetOne handles GET /api/v1/products/{slug}
func (h *Handlers) GetOne(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.GetBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toProductResponse(product))
}

// Create handles POST /api/v1/products (admin only).
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	product, err := h.svc.Create(r.Context(), service.CreateInput{
		Name:        req.Name,
		Tagline:     req.Tagline,
		Description: req.Description,
		Category:    req.Category,
		PriceCents:  req.PriceCents,
		Stock:       req.Stock,
		Badge:       req.Badge,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toProductResponse(product))
}

// AdjustStock handles POST /api/v1/products/{id}/stock (admin only).
func (h *Handlers) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var req adjustStockRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	product, err := h.svc.AdjustStock(r.Context(), chi.URLParam(r, "id"), req.Delta)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toProductResponse(product))
}

func (h *Handlers) bind(r *http.Request, dst any) error {
	if err := httpx.Decode(r, dst); err != nil {
		return err
	}
	if fields := h.validate.Struct(dst); fields != nil {
		return httpx.ErrValidation("some fields are invalid").WithFields(fields)
	}
	return nil
}

func (h *Handlers) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		httpx.Error(w, httpx.ErrNotFound("product not found"))
	case errors.Is(err, domain.ErrSlugTaken):
		httpx.Error(w, httpx.ErrConflict("slug_taken", "a product with that name already exists"))
	case errors.Is(err, domain.ErrInsufficientStock):
		httpx.Error(w, httpx.ErrConflict("insufficient_stock", "not enough stock"))
	case errors.Is(err, domain.ErrInvalidPrice):
		httpx.Error(w, httpx.ErrValidation("price must be greater than zero").
			WithFields(map[string]string{"priceCents": "must be greater than zero"}))
	default:
		httpx.Error(w, err)
	}
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
