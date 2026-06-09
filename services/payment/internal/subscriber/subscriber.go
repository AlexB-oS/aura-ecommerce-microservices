// Package subscriber wires the payment service's event consumer: it listens for
// order.placed and charges the order. This is the heart of the choreography —
// the order service never calls payment directly; payment simply reacts.
package subscriber

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/events"
)

// Payments is the slice of the payment service the subscriber needs.
type Payments interface {
	ProcessOrder(ctx context.Context, orderID, userID string, amountCents int, currency string) error
}

// orderPlaced is the consumer-defined view of the order.placed event.
type orderPlaced struct {
	OrderID    string `json:"orderId"`
	UserID     string `json:"userId"`
	TotalCents int    `json:"totalCents"`
	Currency   string `json:"currency"`
}

// Register subscribes to order.placed.
func Register(bus *events.Bus, payments Payments, log *slog.Logger) error {
	_, err := bus.Subscribe(events.SubjectOrderPlaced, func(env events.Envelope) {
		var o orderPlaced
		if err := env.Decode(&o); err != nil {
			log.Error("bad order.placed payload", slog.Any("error", err))
			return
		}
		if err := payments.ProcessOrder(context.Background(), o.OrderID, o.UserID, o.TotalCents, o.Currency); err != nil {
			log.Error("failed to process payment", slog.String("order_id", o.OrderID), slog.Any("error", err))
			return
		}
		log.Info("processed payment for order", slog.String("order_id", o.OrderID))
	})
	return err
}
