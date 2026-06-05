package http

import "github.com/aura-commerce/platform/services/cart/internal/domain"

// --- Requests --------------------------------------------------------------

type addItemRequest struct {
	Slug string `json:"slug" validate:"required"`
	// Defaults to 1 when omitted; must not be negative.
	Qty int `json:"qty" validate:"gte=0"`
}

type setQtyRequest struct {
	// Pointer so we can tell "qty: 0" (remove) apart from a missing field.
	Qty *int `json:"qty" validate:"required"`
}

// --- Responses -------------------------------------------------------------

type cartResponse struct {
	UserID     string            `json:"userId"`
	Items      []domain.CartItem `json:"items"`
	Count      int               `json:"count"`
	TotalCents int               `json:"totalCents"`
}

func toCartResponse(c *domain.Cart) cartResponse {
	return cartResponse{
		UserID:     c.UserID,
		Items:      c.Items,
		Count:      c.Count(),
		TotalCents: c.TotalCents(),
	}
}
