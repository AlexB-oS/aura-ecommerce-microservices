// Package service holds the order service's business logic. Placing an order
// orchestrates two synchronous calls (read the cart, clear the cart) and one
// asynchronous event (order.placed) that kicks off the rest of the saga.
package service

import (
	"context"
	"log/slog"

	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/services/order/internal/domain"
	"github.com/aura-commerce/platform/services/order/internal/events"
	"github.com/google/uuid"
)

// Repository is the persistence port.
type Repository interface {
	Create(ctx context.Context, o *domain.Order) error
	GetByID(ctx context.Context, userID, id string) (*domain.Order, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.Status) error
}

// Cart is the port to the cart service (REST). bearer is the customer's token.
type Cart interface {
	Get(ctx context.Context, bearer string) ([]domain.OrderItem, error)
	Clear(ctx context.Context, bearer string) error
}

// Publisher is the event-bus port.
type Publisher interface {
	Publish(subject, correlationID string, payload any) error
}

type Service struct {
	repo Repository
	cart Cart
	bus  Publisher
	log  *slog.Logger
}

func New(repo Repository, cart Cart, bus Publisher, log *slog.Logger) *Service {
	return &Service{repo: repo, cart: cart, bus: bus, log: log}
}

// PlaceOrder turns the customer's current cart into an immutable order. It then
// publishes order.placed; stock decrement, payment and notifications all happen
// asynchronously in response to that event.
func (s *Service) PlaceOrder(ctx context.Context, userID, bearer string) (*domain.Order, error) {
	items, err := s.cart.Get(ctx, bearer)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	total := 0
	for _, it := range items {
		total += it.PriceCents * it.Qty
	}

	order := &domain.Order{
		ID:         uuid.NewString(),
		UserID:     userID,
		Status:     domain.StatusPendingPayment,
		TotalCents: total,
		Currency:   "USD",
		Items:      items,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Clear the cart now that it's been captured. Best-effort: the order is
	// already durable, so a clear failure shouldn't fail the checkout.
	if err := s.cart.Clear(ctx, bearer); err != nil {
		s.log.WarnContext(ctx, "failed to clear cart after order", slog.Any("error", err))
	}

	s.publish(ctx, events.SubjectOrderPlaced, buildOrderPlaced(order))
	return order, nil
}

// ListOrders returns a user's order history.
func (s *Service) ListOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetOrder returns one of the user's orders with its line items.
func (s *Service) GetOrder(ctx context.Context, userID, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// MarkPaid is called by the saga subscriber when payment succeeds.
func (s *Service) MarkPaid(ctx context.Context, orderID, userID string) error {
	if err := s.repo.UpdateStatus(ctx, orderID, domain.StatusPaid); err != nil {
		return err
	}
	s.publish(ctx, events.SubjectOrderPaid, events.OrderPaid{OrderID: orderID, UserID: userID})
	return nil
}

// MarkPaymentFailed is called by the saga subscriber when payment fails.
func (s *Service) MarkPaymentFailed(ctx context.Context, orderID string) error {
	return s.repo.UpdateStatus(ctx, orderID, domain.StatusPaymentFailed)
}

func (s *Service) publish(ctx context.Context, subject string, payload any) {
	if s.bus == nil {
		return
	}
	if err := s.bus.Publish(subject, logger.RequestID(ctx), payload); err != nil {
		s.log.WarnContext(ctx, "failed to publish event", slog.String("subject", subject), slog.Any("error", err))
	}
}

func buildOrderPlaced(o *domain.Order) events.OrderPlaced {
	items := make([]events.OrderPlacedItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, events.OrderPlacedItem{
			ProductID:  it.ProductID,
			Slug:       it.Slug,
			Qty:        it.Qty,
			PriceCents: it.PriceCents,
		})
	}
	return events.OrderPlaced{
		OrderID:    o.ID,
		UserID:     o.UserID,
		TotalCents: o.TotalCents,
		Currency:   o.Currency,
		Items:      items,
	}
}
