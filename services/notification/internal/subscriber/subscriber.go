// Package subscriber wires the notification service to the event bus. It is the
// clearest demonstration of event-driven decoupling: this service subscribes to
// events from three other services (user, order, payment) and none of them know
// it exists. Adding notifications required ZERO changes to those services.
package subscriber

import (
	"fmt"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/events"
)

// Sender is the slice of the notifier the subscriber needs.
type Sender interface {
	Send(typ, channel, recipient, subject, message string)
}

// Consumer-defined views of the events this service cares about. Each only
// declares the fields it actually uses.
type userRegistered struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type orderPlaced struct {
	OrderID    string `json:"orderId"`
	UserID     string `json:"userId"`
	TotalCents int    `json:"totalCents"`
}
type paymentSucceeded struct {
	OrderID     string `json:"orderId"`
	UserID      string `json:"userId"`
	AmountCents int    `json:"amountCents"`
}

// Register subscribes to all the events that should trigger a notification.
func Register(bus *events.Bus, notifier Sender, log *slog.Logger) error {
	subs := []struct {
		subject string
		handle  func(events.Envelope)
	}{
		{events.SubjectUserRegistered, func(env events.Envelope) {
			var e userRegistered
			if decode(env, &e, log) {
				notifier.Send("welcome", "email", e.Email,
					"Welcome to AURA",
					fmt.Sprintf("Hi %s, welcome to AURA! Your account is ready.", e.Name))
			}
		}},
		{events.SubjectOrderPlaced, func(env events.Envelope) {
			var e orderPlaced
			if decode(env, &e, log) {
				notifier.Send("order_confirmation", "email", "user:"+e.UserID,
					"Your AURA order is confirmed",
					fmt.Sprintf("Order %s received. Total: %s.", e.OrderID, money(e.TotalCents)))
			}
		}},
		{events.SubjectPaymentSucceeded, func(env events.Envelope) {
			var e paymentSucceeded
			if decode(env, &e, log) {
				notifier.Send("receipt", "email", "user:"+e.UserID,
					"Payment received",
					fmt.Sprintf("We received %s for order %s. Thank you!", money(e.AmountCents), e.OrderID))
			}
		}},
	}

	for _, s := range subs {
		if _, err := bus.Subscribe(s.subject, s.handle); err != nil {
			return err
		}
	}
	return nil
}

func decode(env events.Envelope, v any, log *slog.Logger) bool {
	if err := env.Decode(v); err != nil {
		log.Error("bad event payload", slog.String("subject", env.Subject), slog.Any("error", err))
		return false
	}
	return true
}

func money(cents int) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}
