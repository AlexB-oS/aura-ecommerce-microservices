// Package domain holds the payment service's entities. This is a SIMULATED
// payment provider: it records a payment row and reports success, standing in
// for a real Stripe/PayPal integration (which would slot in behind the same
// service interface).
package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

// Payment records one charge attempt against an order.
type Payment struct {
	ID         string
	OrderID    string
	UserID     string
	AmountCents int
	Currency   string
	Status     Status
	Provider   string // "mock" here; would be "stripe"/"paypal" in production
	CreatedAt  time.Time
}

var ErrPaymentNotFound = errors.New("payment not found")
