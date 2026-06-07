// Package events defines the order service's published event payloads.
package events

import pkgevents "github.com/aura-commerce/platform/pkg/events"

const (
	SubjectOrderPlaced = pkgevents.SubjectOrderPlaced
	SubjectOrderPaid   = pkgevents.SubjectOrderPaid
)

// OrderPlaced is published when a customer checks out. It is the trigger for the
// rest of the saga: the product service decrements stock, the payment service
// charges the order, and the notification service emails a confirmation — none
// of which the order service calls directly.
type OrderPlaced struct {
	OrderID    string             `json:"orderId"`
	UserID     string             `json:"userId"`
	TotalCents int                `json:"totalCents"`
	Currency   string             `json:"currency"`
	Items      []OrderPlacedItem  `json:"items"`
}

type OrderPlacedItem struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug"`
	Qty        int    `json:"qty"`
	PriceCents int    `json:"priceCents"`
}

// OrderPaid is published once payment for an order succeeds.
type OrderPaid struct {
	OrderID string `json:"orderId"`
	UserID  string `json:"userId"`
}
