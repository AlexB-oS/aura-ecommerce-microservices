package http

import (
	"time"

	"github.com/aura-commerce/platform/services/order/internal/domain"
)

type orderItemResponse struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Qty        int    `json:"qty"`
}

type orderResponse struct {
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	TotalCents int                 `json:"totalCents"`
	Currency   string              `json:"currency"`
	Items      []orderItemResponse `json:"items,omitempty"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemResponse{
			ProductID:  it.ProductID,
			Slug:       it.Slug,
			Name:       it.Name,
			PriceCents: it.PriceCents,
			Qty:        it.Qty,
		})
	}
	return orderResponse{
		ID:         o.ID,
		Status:     string(o.Status),
		TotalCents: o.TotalCents,
		Currency:   o.Currency,
		Items:      items,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

func toOrderResponses(orders []domain.Order) []orderResponse {
	out := make([]orderResponse, len(orders))
	for i := range orders {
		out[i] = toOrderResponse(&orders[i])
		out[i].Items = nil // list view omits line items
	}
	return out
}
