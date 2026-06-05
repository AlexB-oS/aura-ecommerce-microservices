// Package domain holds the cart service's entities and errors. A cart is a
// short-lived, user-scoped collection of line items. Each item stores a price
// snapshot taken from the catalog at add-time, so the cart total is stable even
// if the product's price later changes.
package domain

import (
	"errors"
	"time"
)

// CartItem is one line in a cart. ProductID is the catalog's UUID (used later by
// the order service to decrement stock); Slug is kept for display/links.
type CartItem struct {
	ProductID  string `json:"productId"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Qty        int    `json:"qty"`
}

// Cart is the aggregate stored per user in Redis.
type Cart struct {
	UserID    string     `json:"userId"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TotalCents is the sum of line totals.
func (c *Cart) TotalCents() int {
	total := 0
	for _, it := range c.Items {
		total += it.PriceCents * it.Qty
	}
	return total
}

// Count is the total number of units across all items.
func (c *Cart) Count() int {
	n := 0
	for _, it := range c.Items {
		n += it.Qty
	}
	return n
}

// ProductInfo is the catalog snapshot the cart needs when adding an item. It is
// fetched synchronously from the product service so the cart can store the
// authoritative name and price (not whatever the client claims).
type ProductInfo struct {
	ID         string
	Slug       string
	Name       string
	PriceCents int
	Stock      int
}

var (
	ErrItemNotInCart      = errors.New("item is not in the cart")
	ErrProductUnavailable = errors.New("product is unavailable")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
)
