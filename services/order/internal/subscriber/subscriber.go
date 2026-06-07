// Package subscriber wires the order service's event consumers. This is the
// "downstream" half of the saga: the order service reacts to payment results
// reported asynchronously by the payment service and updates the order status.
package subscriber

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/events"
)

// Orders is the slice of the order service the subscriber needs.
type Orders interface {
	MarkPaid(ctx context.Context, orderID, userID string) error
	MarkPaymentFailed(ctx context.Context, orderID string) error
}

// paymentResult is the consumer-defined schema for payment events: we only read
// the fields we need, ignoring whatever else the payment service includes.
type paymentResult struct {
	OrderID string `json:"orderId"`
	UserID  string `json:"userId"`
}

// Register subscribes to payment.succeeded and payment.failed. It takes the
// concrete *events.Bus so it can call Subscribe; the handlers update orders.
func Register(bus *events.Bus, orders Orders, log *slog.Logger) error {
	if _, err := bus.Subscribe(events.SubjectPaymentSucceeded, func(env events.Envelope) {
		var p paymentResult
		if err := env.Decode(&p); err != nil {
			log.Error("bad payment.succeeded payload", slog.Any("error", err))
			return
		}
		if err := orders.MarkPaid(context.Background(), p.OrderID, p.UserID); err != nil {
			log.Error("failed to mark order paid", slog.String("order_id", p.OrderID), slog.Any("error", err))
			return
		}
		log.Info("order marked paid", slog.String("order_id", p.OrderID))
	}); err != nil {
		return err
	}

	if _, err := bus.Subscribe(events.SubjectPaymentFailed, func(env events.Envelope) {
		var p paymentResult
		if err := env.Decode(&p); err != nil {
			log.Error("bad payment.failed payload", slog.Any("error", err))
			return
		}
		if err := orders.MarkPaymentFailed(context.Background(), p.OrderID); err != nil {
			log.Error("failed to mark order payment_failed", slog.String("order_id", p.OrderID), slog.Any("error", err))
			return
		}
		log.Info("order marked payment_failed", slog.String("order_id", p.OrderID))
	}); err != nil {
		return err
	}

	return nil
}
