package http

import (
	"time"

	"github.com/aura-commerce/platform/services/user/internal/domain"
	"github.com/aura-commerce/platform/services/user/internal/service"
)

// --- Requests --------------------------------------------------------------
// The `validate` tags are enforced by pkg/validate before any business logic
// runs, and failures are reported per-field straight back to the React forms.

type registerRequest struct {
	Email    string `json:"email"    validate:"required,email,max=254"`
	Name     string `json:"name"     validate:"required,min=2,max=80"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type updateProfileRequest struct {
	Name string `json:"name" validate:"required,min=2,max=80"`
}

// --- Responses -------------------------------------------------------------

// userResponse is the public projection of a user. Note there is no password
// field — the hash never crosses the service boundary.
type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// authResponse is returned by register/login/refresh: the user plus tokens.
type authResponse struct {
	User         *userResponse `json:"user,omitempty"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	TokenType    string        `json:"tokenType"`
	ExpiresIn    int           `json:"expiresIn"`
}

func toUserResponse(u *domain.User) *userResponse {
	return &userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toAuthResponse(u *domain.User, pair service.TokenPair) authResponse {
	return authResponse{
		User:         toUserResponse(u),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
	}
}
