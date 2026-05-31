// Package service holds the user service's business logic: registration,
// authentication, token refresh, and profile management. It depends only on
// abstract ports (Repository, Publisher), never on Postgres or HTTP directly,
// which keeps it unit-testable without any infrastructure.
package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/services/user/internal/domain"
	"github.com/aura-commerce/platform/services/user/internal/events"
	"github.com/google/uuid"
)

// Repository is the persistence port. The Postgres adapter implements it; tests
// supply an in-memory fake.
type Repository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
}

// Publisher is the event-bus port. pkg/events.Bus satisfies it.
type Publisher interface {
	Publish(subject, correlationID string, payload any) error
}

// TokenPair is what a successful login/refresh hands back to the client.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int // access-token lifetime in seconds
}

// Service is the use-case layer. Construct it with New.
type Service struct {
	repo   Repository
	tokens *auth.Manager
	bus    Publisher
	log    *slog.Logger
}

// New wires the service's dependencies.
func New(repo Repository, tokens *auth.Manager, bus Publisher, log *slog.Logger) *Service {
	return &Service{repo: repo, tokens: tokens, bus: bus, log: log}
}

// Register creates a new customer account and returns it together with a fresh
// token pair so the client is logged in immediately.
func (s *Service) Register(ctx context.Context, email, name, password string) (*domain.User, TokenPair, error) {
	email = normalizeEmail(email)

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, TokenPair{}, err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		Name:         strings.TrimSpace(name),
		PasswordHash: hash,
		Role:         domain.RoleCustomer,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, TokenPair{}, err // includes domain.ErrEmailTaken
	}

	// Fire-and-forget: a failure to publish must not fail the registration the
	// user already completed, so we log and move on. The event is best-effort.
	s.publish(ctx, events.SubjectUserRegistered, events.UserRegistered{
		ID: user.ID, Email: user.Email, Name: user.Name,
	})

	pair, err := s.issueTokens(user)
	return user, pair, err
}

// Authenticate verifies credentials and returns the user with a token pair.
// It always returns ErrInvalidCredentials for both "no such email" and "wrong
// password" so the response never reveals whether an email is registered.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*domain.User, TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Spend roughly the same time as a real bcrypt check to blunt
			// user-enumeration via response timing.
			auth.CheckPassword("$2a$12$invalidinvalidinvalidinvalidinvalidinvalidinvalidinv", password)
			return nil, TokenPair{}, domain.ErrInvalidCredentials
		}
		return nil, TokenPair{}, err
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return nil, TokenPair{}, domain.ErrInvalidCredentials
	}

	pair, err := s.issueTokens(user)
	return user, pair, err
}

// Refresh exchanges a valid refresh token for a brand-new token pair (rotation).
func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := s.tokens.Verify(refreshToken, auth.RefreshToken)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidCredentials
	}
	// Confirm the account still exists before minting new tokens.
	user, err := s.repo.GetByID(ctx, claims.UserID())
	if err != nil {
		return TokenPair{}, domain.ErrInvalidCredentials
	}
	return s.issueTokens(user)
}

// Profile returns a user by id.
func (s *Service) Profile(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// UpdateProfile changes mutable fields (currently just the display name).
func (s *Service) UpdateProfile(ctx context.Context, userID, name string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Name = strings.TrimSpace(name)
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	s.publish(ctx, events.SubjectUserUpdated, events.UserUpdated{ID: user.ID, Name: user.Name})
	return user, nil
}

func (s *Service) issueTokens(u *domain.User) (TokenPair, error) {
	access, _, err := s.tokens.Issue(u.ID, u.Email, string(u.Role), auth.AccessToken)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, _, err := s.tokens.Issue(u.ID, u.Email, string(u.Role), auth.RefreshToken)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.tokens.AccessTTL() / time.Second),
	}, nil
}

func (s *Service) publish(ctx context.Context, subject string, payload any) {
	if s.bus == nil {
		return
	}
	if err := s.bus.Publish(subject, logger.RequestID(ctx), payload); err != nil {
		s.log.WarnContext(ctx, "failed to publish event", slog.String("subject", subject), slog.Any("error", err))
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
