// Package events defines the payment service's published event payloads.
package events

import pkgevents "github.com/aura-commerce/platform/pkg/events"

const (
	SubjectPaymentSucceeded = pkgevents.SubjectPaymentSucceeded
	SubjectPaymentFailed    = pkgevents.SubjectPaymentFailed
)

// PaymentSucceeded tells the order service (and anyone else) that an order's
// payment cleared. The order service marks the order paid; notification sends a
// receipt.
type PaymentSucceeded struct {
	PaymentID   string `json:"paymentId"`
	OrderID     string `json:"orderId"`
	UserID      string `json:"userId"`
	AmountCents int    `json:"amountCents"`
}

// PaymentFailed reports a declined/failed charge.
type PaymentFailed struct {
	OrderID string `json:"orderId"`
	UserID  string `json:"userId"`
	Reason  string `json:"reason"`
}
