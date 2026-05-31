package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/services/user/internal/domain"
	"github.com/aura-commerce/platform/services/user/internal/service"
)

// fakeRepo is an in-memory implementation of service.Repository. Because the
// service depends on the interface (not Postgres), the entire business layer is
// testable with no database, no Docker, and microsecond-fast tests.
type fakeRepo struct {
	mu       sync.Mutex
	byID     map[string]*domain.User
	byEmail  map[string]*domain.User
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*domain.User{}, byEmail: map[string]*domain.User{}}
}

func (f *fakeRepo) Create(_ context.Context, u *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.byEmail[u.Email]; exists {
		return domain.ErrEmailTaken
	}
	u.CreatedAt, u.UpdatedAt = time.Now(), time.Now()
	cp := *u
	f.byID[u.ID] = &cp
	f.byEmail[u.Email] = &cp
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.byID[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, domain.ErrUserNotFound
}

func (f *fakeRepo) Update(_ context.Context, u *domain.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.byID[u.ID]
	if !ok {
		return domain.ErrUserNotFound
	}
	existing.Name = u.Name
	existing.UpdatedAt = time.Now()
	return nil
}

// noopPublisher swallows events so tests don't need a real bus.
type noopPublisher struct{}

func (noopPublisher) Publish(string, string, any) error { return nil }

func newService(repo service.Repository) *service.Service {
	tokens := auth.NewManager("test-secret", "aura.test", 15*time.Minute, time.Hour)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.New(repo, tokens, noopPublisher{}, log)
}

func TestRegister_Success(t *testing.T) {
	svc := newService(newFakeRepo())

	user, pair, err := svc.Register(context.Background(), "Ada@Example.com", "Ada Lovelace", "supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "ada@example.com" {
		t.Errorf("email should be normalized to lowercase, got %q", user.Email)
	}
	if user.Role != domain.RoleCustomer {
		t.Errorf("new users should default to customer role, got %q", user.Role)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("expected a token pair to be issued on registration")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newService(newFakeRepo())
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "dup@example.com", "First", "password1"); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}
	_, _, err := svc.Register(ctx, "dup@example.com", "Second", "password2")
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAuthenticate(t *testing.T) {
	svc := newService(newFakeRepo())
	ctx := context.Background()
	if _, _, err := svc.Register(ctx, "kit@example.com", "Kit", "correct-horse"); err != nil {
		t.Fatalf("setup register failed: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		if _, _, err := svc.Authenticate(ctx, "kit@example.com", "correct-horse"); err != nil {
			t.Errorf("expected success, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, _, err := svc.Authenticate(ctx, "kit@example.com", "wrong")
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("unknown email returns the same error", func(t *testing.T) {
		_, _, err := svc.Authenticate(ctx, "ghost@example.com", "whatever")
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials to avoid user enumeration, got %v", err)
		}
	})
}

func TestRefresh(t *testing.T) {
	svc := newService(newFakeRepo())
	ctx := context.Background()
	_, pair, err := svc.Register(ctx, "rin@example.com", "Rin", "password123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	refreshed, err := svc.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Error("expected a new access token from refresh")
	}

	// An access token must NOT be accepted where a refresh token is required.
	if _, err := svc.Refresh(ctx, pair.AccessToken); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Error("an access token should not be usable as a refresh token")
	}
}
