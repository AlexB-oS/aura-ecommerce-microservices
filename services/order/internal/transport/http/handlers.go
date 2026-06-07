package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/aura-commerce/platform/services/order/internal/domain"
	"github.com/aura-commerce/platform/services/order/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	svc *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

// Place handles POST /api/v1/orders — checkout from the current cart.
func (h *Handlers) Place(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}

	// Forward the caller's own token so the order service can read their cart.
	order, err := h.svc.PlaceOrder(r.Context(), claims.UserID(), bearer(r))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toOrderResponse(order))
}

// List handles GET /api/v1/orders — the user's order history.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	orders, err := h.svc.ListOrders(r.Context(), claims.UserID())
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toOrderResponses(orders))
}

// Get handles GET /api/v1/orders/{id}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	order, err := h.svc.GetOrder(r.Context(), claims.UserID(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toOrderResponse(order))
}

func (h *Handlers) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmptyCart):
		httpx.Error(w, httpx.ErrConflict("empty_cart", "your cart is empty"))
	case errors.Is(err, domain.ErrOrderNotFound):
		httpx.Error(w, httpx.ErrNotFound("order not found"))
	default:
		httpx.Error(w, err)
	}
}

// bearer extracts the raw access token from the Authorization header so it can
// be forwarded to the cart service.
func bearer(r *http.Request) string {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token
}
