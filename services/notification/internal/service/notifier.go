// Package service contains the notification "sender". In this learning build,
// sending is mocked: each notification is logged and kept in a small in-memory
// ring buffer that the HTTP endpoint exposes, so you can watch the event-driven
// flow without wiring up SendGrid/Twilio. Swapping in a real provider is a
// change to Send() only.
package service

import (
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notification is one message the platform decided to send.
type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`      // welcome | order_confirmation | receipt
	Channel   string    `json:"channel"`   // email | sms
	Recipient string    `json:"recipient"` // email address or "user:<id>"
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	SentAt    time.Time `json:"sentAt"`
}

// Notifier records and "sends" notifications.
type Notifier struct {
	mu     sync.Mutex
	recent []Notification
	max    int
	log    *slog.Logger
}

func New(log *slog.Logger) *Notifier {
	return &Notifier{max: 100, log: log}
}

// Send dispatches a notification. Here that means: log it and remember it. A
// real implementation would call an email/SMS provider; if that failed it would
// retry or dead-letter, none of which the publishers need to know about.
func (n *Notifier) Send(typ, channel, recipient, subject, message string) {
	notif := Notification{
		ID:        uuid.NewString(),
		Type:      typ,
		Channel:   channel,
		Recipient: recipient,
		Subject:   subject,
		Message:   message,
		SentAt:    time.Now().UTC(),
	}

	n.log.Info("notification sent",
		slog.String("type", typ),
		slog.String("channel", channel),
		slog.String("recipient", recipient),
		slog.String("subject", subject),
	)

	n.mu.Lock()
	defer n.mu.Unlock()
	n.recent = append(n.recent, notif)
	if len(n.recent) > n.max {
		n.recent = n.recent[len(n.recent)-n.max:]
	}
}

// Recent returns the most-recent notifications, newest first.
func (n *Notifier) Recent() []Notification {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]Notification, len(n.recent))
	for i, item := range n.recent {
		out[len(n.recent)-1-i] = item
	}
	return out
}
