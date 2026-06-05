package http

import (
	"errors"
	"net/http"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/aura-commerce/platform/pkg/validate"
	"github.com/aura-commerce/platform/services/cart/internal/domain"
	"github.com/aura-commerce/platform/services/cart/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	svc      *service.Service
	validate *validate.Validator
}

func NewHandlers(svc *service.Service, v *validate.Validator) *Handlers {
	return &Handlers{svc: svc, validate: v}
}

// Get handles GET /api/v1/cart
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	cart, err := h.svc.Get(r.Context(), userID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
}

// AddItem handles POST /api/v1/cart/items
func (h *Handlers) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	var req addItemRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}
	qty := req.Qty
	if qty == 0 {
		qty = 1
	}

	cart, err := h.svc.AddItem(r.Context(), userID, req.Slug, qty)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
}

// SetQty handles PATCH /api/v1/cart/items/{productId}
func (h *Handlers) SetQty(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	var req setQtyRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	cart, err := h.svc.SetItemQty(r.Context(), userID, chi.URLParam(r, "productId"), *req.Qty)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
}

// RemoveItem handles DELETE /api/v1/cart/items/{productId}
func (h *Handlers) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	cart, err := h.svc.RemoveItem(r.Context(), userID, chi.URLParam(r, "productId"))
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toCartResponse(cart))
}

// Clear handles DELETE /api/v1/cart
func (h *Handlers) Clear(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	if err := h.svc.Clear(r.Context(), userID); err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, cartResponse{UserID: userID, Items: []domain.CartItem{}})
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
	case errors.Is(err, domain.ErrProductUnavailable):
		httpx.Error(w, httpx.ErrNotFound("that product is unavailable"))
	case errors.Is(err, domain.ErrItemNotInCart):
		httpx.Error(w, httpx.ErrNotFound("item is not in your cart"))
	case errors.Is(err, domain.ErrInvalidQuantity):
		httpx.Error(w, httpx.ErrValidation("quantity must be greater than zero").
			WithFields(map[string]string{"qty": "must be greater than zero"}))
	default:
		httpx.Error(w, err)
	}
}

// userID pulls the authenticated user's id off the request context.
func userID(r *http.Request) (string, bool) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		return "", false
	}
	return claims.UserID(), true
}
