package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/aura-commerce/platform/services/payment/internal/domain"
	"github.com/aura-commerce/platform/services/payment/internal/service"
)

type fakeRepo struct {
	byOrder map[string]*domain.Payment
}

func newFakeRepo() *fakeRepo { return &fakeRepo{byOrder: map[string]*domain.Payment{}} }

func (f *fakeRepo) Create(_ context.Context, p *domain.Payment) error {
	f.byOrder[p.OrderID] = p
	return nil
}
func (f *fakeRepo) ExistsForOrder(_ context.Context, orderID string) (bool, error) {
	_, ok := f.byOrder[orderID]
	return ok, nil
}
func (f *fakeRepo) ListByUser(_ context.Context, userID string) ([]domain.Payment, error) {
	var out []domain.Payment
	for _, p := range f.byOrder {
		if p.UserID == userID {
			out = append(out, *p)
		}
	}
	return out, nil
}

type recordingPublisher struct{ subjects []string }

func (r *recordingPublisher) Publish(subject, _ string, _ any) error {
	r.subjects = append(r.subjects, subject)
	return nil
}

func newService(repo service.Repository, pub service.Publisher) *service.Service {
	return service.New(repo, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestProcessOrder_SucceedsAndPublishes(t *testing.T) {
	repo := newFakeRepo()
	pub := &recordingPublisher{}
	svc := newService(repo, pub)

	if err := svc.ProcessOrder(context.Background(), "o1", "u1", 2500, "USD"); err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if p := repo.byOrder["o1"]; p == nil || p.Status != domain.StatusSucceeded {
		t.Fatalf("expected a succeeded payment, got %+v", p)
	}
	if len(pub.subjects) != 1 || pub.subjects[0] != "payment.succeeded" {
		t.Errorf("expected payment.succeeded event, got %v", pub.subjects)
	}
}

func TestProcessOrder_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	pub := &recordingPublisher{}
	svc := newService(repo, pub)
	ctx := context.Background()

	_ = svc.ProcessOrder(ctx, "o1", "u1", 2500, "USD")
	// Redelivery of the same order must not charge or publish again.
	_ = svc.ProcessOrder(ctx, "o1", "u1", 2500, "USD")

	if len(pub.subjects) != 1 {
		t.Errorf("expected exactly one event despite redelivery, got %v", pub.subjects)
	}
}
