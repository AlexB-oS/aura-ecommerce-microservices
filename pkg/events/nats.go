// Package events is a thin wrapper over NATS that gives every service the same
// publish/subscribe API and a single, versioned envelope for messages. NATS is
// the asynchronous backbone of the platform: when the user service finishes a
// registration it publishes `user.registered`, and any number of other
// services (notification, analytics, ...) can react without the user service
// knowing or caring who is listening. This is what decouples the services.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Subject constants are declared centrally so producers and consumers cannot
// drift apart over a typo'd string.
const (
	SubjectUserRegistered = "user.registered"
	SubjectUserUpdated    = "user.updated"

	SubjectProductCreated  = "product.created"
	SubjectProductStockLow = "product.stock_low"

	SubjectOrderPlaced        = "order.placed"
	SubjectOrderPaid          = "order.paid"
	SubjectOrderPaymentFailed = "order.payment_failed"

	SubjectPaymentSucceeded = "payment.succeeded"
	SubjectPaymentFailed    = "payment.failed"
)

// Envelope is the standard shape of every event on the bus. The metadata makes
// events self-describing, replayable, and easy to trace.
type Envelope struct {
	ID            string          `json:"id"`              // unique event id
	Subject       string          `json:"subject"`         // e.g. "user.registered"
	Source        string          `json:"source"`          // emitting service name
	Time          time.Time       `json:"time"`            // when it was produced
	CorrelationID string          `json:"correlation_id"`  // request id, for tracing
	Data          json.RawMessage `json:"data"`            // service-specific payload
}

// Decode unmarshals the event's service-specific Data payload into v.
func (e Envelope) Decode(v any) error {
	return json.Unmarshal(e.Data, v)
}

// Bus is a connected NATS client scoped to a single service.
type Bus struct {
	conn   *nats.Conn
	source string
}

// Connect dials NATS with sensible reconnect behaviour. `source` is the name of
// the calling service and is stamped onto every event it publishes.
func Connect(url, source string) (*Bus, error) {
	conn, err := nats.Connect(url,
		nats.Name(source),
		nats.MaxReconnects(-1), // retry forever; services may start before NATS
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return &Bus{conn: conn, source: source}, nil
}

// Publish marshals payload into a standard envelope and sends it on subject.
// correlationID should be the inbound request id so the event can be traced
// back to the HTTP request that caused it.
func (b *Bus) Publish(subject, correlationID string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := Envelope{
		ID:            uuid.NewString(),
		Subject:       subject,
		Source:        b.source,
		Time:          time.Now().UTC(),
		CorrelationID: correlationID,
		Data:          data,
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return b.conn.Publish(subject, body)
}

// Subscribe registers a handler for a subject. The handler receives the decoded
// envelope. This is used by consumer services (e.g. notification) — included
// here so the pattern is ready the moment the next service is added.
func (b *Bus) Subscribe(subject string, handler func(Envelope)) (*nats.Subscription, error) {
	return b.conn.Subscribe(subject, func(msg *nats.Msg) {
		var env Envelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			return // malformed message; a real consumer would log / dead-letter it
		}
		handler(env)
	})
}

// Healthy reports whether the connection to NATS is currently usable.
func (b *Bus) Healthy() bool {
	return b.conn != nil && b.conn.IsConnected()
}

// Close flushes pending messages and drains the connection.
func (b *Bus) Close() {
	if b.conn != nil {
		_ = b.conn.Drain()
	}
}
