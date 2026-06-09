// Package service holds the payment service's logic: process the charge for an
// order (triggered by the order.placed event) and list a user's payments.
package service

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/services/payment/internal/domain"
	"github.com/aura-commerce/platform/services/payment/internal/events"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, p *domain.Payment) error
	ExistsForOrder(ctx context.Context, orderID string) (bool, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Payment, error)
}

type Publisher interface {
	Publish(subject, correlationID string, payload any) error
}

type Service struct {
	repo Repository
	bus  Publisher
	log  *slog.Logger
}

func New(repo Repository, bus Publisher, log *slog.Logger) *Service {
	return &Service{repo: repo, bus: bus, log: log}
}

// Charge is the gateway call. Here it's a mock that always approves; a real
// provider integration (Stripe/PayPal) would replace just this method.
func (s *Service) charge(amountCents int) (domain.Status, string) {
	if amountCents <= 0 {
		return domain.StatusFailed, "invalid amount"
	}
	return domain.StatusSucceeded, ""
}

// ProcessOrder is invoked from the order.placed subscriber. It is idempotent:
// if a payment already exists for the order (e.g. the event was redelivered) it
// does nothing, so an order is never charged twice.
func (s *Service) ProcessOrder(ctx context.Context, orderID, userID string, amountCents int, currency string) error {
	exists, err := s.repo.ExistsForOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if exists {
		s.log.InfoContext(ctx, "payment already processed, skipping", slog.String("order_id", orderID))
		return nil
	}

	status, reason := s.charge(amountCents)

	payment := &domain.Payment{
		ID:          uuid.NewString(),
		OrderID:     orderID,
		UserID:      userID,
		AmountCents: amountCents,
		Currency:    currency,
		Status:      status,
		Provider:    "mock",
	}
	if err := s.repo.Create(ctx, payment); err != nil {
		return err
	}

	if status == domain.StatusSucceeded {
		s.publish(ctx, events.SubjectPaymentSucceeded, events.PaymentSucceeded{
			PaymentID: payment.ID, OrderID: orderID, UserID: userID, AmountCents: amountCents,
		})
	} else {
		s.publish(ctx, events.SubjectPaymentFailed, events.PaymentFailed{
			OrderID: orderID, UserID: userID, Reason: reason,
		})
	}
	return nil
}

// ListByUser returns a customer's payments.
func (s *Service) ListByUser(ctx context.Context, userID string) ([]domain.Payment, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) publish(ctx context.Context, subject string, payload any) {
	if s.bus == nil {
		return
	}
	if err := s.bus.Publish(subject, logger.RequestID(ctx), payload); err != nil {
		s.log.WarnContext(ctx, "failed to publish event", slog.String("subject", subject), slog.Any("error", err))
	}
}
