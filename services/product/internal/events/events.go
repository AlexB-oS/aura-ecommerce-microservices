// Package events defines the product service's published event payloads — the
// contract other services (orders, search indexers, notifications) consume off
// the NATS bus.
package events

import pkgevents "github.com/aura-commerce/platform/pkg/events"

const (
	SubjectProductCreated  = pkgevents.SubjectProductCreated
	SubjectProductStockLow = pkgevents.SubjectProductStockLow
)

// ProductCreated is published when a new product is added to the catalog.
type ProductCreated struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
}

// ProductStockLow is published when a stock adjustment drops a product at or
// below the low-stock threshold — e.g. a re-ordering service could subscribe.
type ProductStockLow struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Stock int    `json:"stock"`
}
