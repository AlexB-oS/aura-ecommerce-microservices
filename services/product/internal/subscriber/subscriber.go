// Package subscriber lets the product service react to events. When an order is
// placed, the product service decrements stock for each line item — it learns
// about the order purely from the order.placed event, never via a direct call
// from the order service. (A robust production version would do this inside a
// JetStream durable consumer for at-least-once delivery + retries.)
package subscriber

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/events"
)

// Inventory is the slice of the product service the subscriber needs.
type Inventory interface {
	ReduceStock(ctx context.Context, productID string, qty int) error
}

// orderPlaced is the consumer-defined view of the order.placed event.
type orderPlaced struct {
	OrderID string `json:"orderId"`
	Items   []struct {
		ProductID string `json:"productId"`
		Qty       int    `json:"qty"`
	} `json:"items"`
}

// Register subscribes to order.placed and adjusts stock.
func Register(bus *events.Bus, inv Inventory, log *slog.Logger) error {
	_, err := bus.Subscribe(events.SubjectOrderPlaced, func(env events.Envelope) {
		var o orderPlaced
		if err := env.Decode(&o); err != nil {
			log.Error("bad order.placed payload", slog.Any("error", err))
			return
		}
		for _, item := range o.Items {
			if err := inv.ReduceStock(context.Background(), item.ProductID, item.Qty); err != nil {
				// Log and continue; one bad line shouldn't abort the rest. A real
				// system would compensate (e.g. flag the order for review).
				log.Error("failed to reduce stock",
					slog.String("order_id", o.OrderID),
					slog.String("product_id", item.ProductID),
					slog.Any("error", err))
			}
		}
		log.Info("adjusted stock for order", slog.String("order_id", o.OrderID))
	})
	return err
}
