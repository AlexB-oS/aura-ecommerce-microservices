// Package domain holds the order service's entities and errors. An order is a
// snapshot of a checkout: the line items and prices are copied in at placement
// time so the order is immutable even as carts and catalog prices change.
package domain

import (
	"errors"
	"time"
)

// Status is the order lifecycle. An order starts pending_payment and is moved to
// paid or payment_failed asynchronously when the payment service reports back.
type Status string

const (
	StatusPendingPayment Status = "pending_payment"
	StatusPaid           Status = "paid"
	StatusPaymentFailed  Status = "payment_failed"
	StatusCancelled      Status = "cancelled"
)

// OrderItem is one line of an order (a price/name snapshot from the cart).
type OrderItem struct {
	ProductID  string
	Slug       string
	Name       string
	PriceCents int
	Qty        int
}

// Order is the aggregate root.
type Order struct {
	ID         string
	UserID     string
	Status     Status
	TotalCents int
	Currency   string
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var (
	ErrEmptyCart    = errors.New("cannot place an order with an empty cart")
	ErrOrderNotFound = errors.New("order not found")
)
