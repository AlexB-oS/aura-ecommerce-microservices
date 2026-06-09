package http

import (
	"net/http"
	"time"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/aura-commerce/platform/services/payment/internal/domain"
	"github.com/aura-commerce/platform/services/payment/internal/service"
)

type Handlers struct {
	svc *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

type paymentResponse struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"orderId"`
	AmountCents int       `json:"amountCents"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Provider    string    `json:"provider"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toPaymentResponses(payments []domain.Payment) []paymentResponse {
	out := make([]paymentResponse, len(payments))
	for i, p := range payments {
		out[i] = paymentResponse{
			ID:          p.ID,
			OrderID:     p.OrderID,
			AmountCents: p.AmountCents,
			Currency:    p.Currency,
			Status:      string(p.Status),
			Provider:    p.Provider,
			CreatedAt:   p.CreatedAt,
		}
	}
	return out
}

// List handles GET /api/v1/payments — the authenticated user's payments.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	payments, err := h.svc.ListByUser(r.Context(), claims.UserID())
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toPaymentResponses(payments))
}
