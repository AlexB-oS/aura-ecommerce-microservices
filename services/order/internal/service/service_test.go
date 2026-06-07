package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/aura-commerce/platform/services/order/internal/domain"
	"github.com/aura-commerce/platform/services/order/internal/service"
)

type fakeRepo struct {
	orders  map[string]*domain.Order
	created int
}

func newFakeRepo() *fakeRepo { return &fakeRepo{orders: map[string]*domain.Order{}} }

func (f *fakeRepo) Create(_ context.Context, o *domain.Order) error {
	cp := *o
	f.orders[o.ID] = &cp
	f.created++
	return nil
}
func (f *fakeRepo) GetByID(_ context.Context, userID, id string) (*domain.Order, error) {
	if o, ok := f.orders[id]; ok && o.UserID == userID {
		return o, nil
	}
	return nil, domain.ErrOrderNotFound
}
func (f *fakeRepo) ListByUser(_ context.Context, userID string) ([]domain.Order, error) {
	var out []domain.Order
	for _, o := range f.orders {
		if o.UserID == userID {
			out = append(out, *o)
		}
	}
	return out, nil
}
func (f *fakeRepo) UpdateStatus(_ context.Context, id string, status domain.Status) error {
	o, ok := f.orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	o.Status = status
	return nil
}

type fakeCart struct {
	items   []domain.OrderItem
	cleared bool
}

func (f *fakeCart) Get(context.Context, string) ([]domain.OrderItem, error) { return f.items, nil }
func (f *fakeCart) Clear(context.Context, string) error                     { f.cleared = true; return nil }

type recordingPublisher struct{ subjects []string }

func (r *recordingPublisher) Publish(subject, _ string, _ any) error {
	r.subjects = append(r.subjects, subject)
	return nil
}

func newService(repo service.Repository, cart service.Cart, pub service.Publisher) *service.Service {
	return service.New(repo, cart, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestPlaceOrder_EmptyCart(t *testing.T) {
	svc := newService(newFakeRepo(), &fakeCart{items: nil}, &recordingPublisher{})
	if _, err := svc.PlaceOrder(context.Background(), "u1", "tok"); !errors.Is(err, domain.ErrEmptyCart) {
		t.Errorf("expected ErrEmptyCart, got %v", err)
	}
}

func TestPlaceOrder_Success(t *testing.T) {
	repo := newFakeRepo()
	cart := &fakeCart{items: []domain.OrderItem{
		{ProductID: "p1", Slug: "a", Name: "A", PriceCents: 1000, Qty: 2},
		{ProductID: "p2", Slug: "b", Name: "B", PriceCents: 500, Qty: 1},
	}}
	pub := &recordingPublisher{}
	svc := newService(repo, cart, pub)

	order, err := svc.PlaceOrder(context.Background(), "u1", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.TotalCents != 2500 {
		t.Errorf("expected total 2500, got %d", order.TotalCents)
	}
	if order.Status != domain.StatusPendingPayment {
		t.Errorf("new order should be pending_payment, got %s", order.Status)
	}
	if !cart.cleared {
		t.Error("cart should be cleared after a successful order")
	}
	if repo.created != 1 {
		t.Errorf("expected 1 order persisted, got %d", repo.created)
	}
	if len(pub.subjects) != 1 || pub.subjects[0] != "order.placed" {
		t.Errorf("expected an order.placed event, got %v", pub.subjects)
	}
}

func TestMarkPaid(t *testing.T) {
	repo := newFakeRepo()
	pub := &recordingPublisher{}
	svc := newService(repo, &fakeCart{}, pub)
	repo.orders["o1"] = &domain.Order{ID: "o1", UserID: "u1", Status: domain.StatusPendingPayment}

	if err := svc.MarkPaid(context.Background(), "o1", "u1"); err != nil {
		t.Fatalf("mark paid failed: %v", err)
	}
	if repo.orders["o1"].Status != domain.StatusPaid {
		t.Errorf("expected status paid, got %s", repo.orders["o1"].Status)
	}
}
