package http

import (
	"errors"
	"net/http"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/aura-commerce/platform/pkg/validate"
	"github.com/aura-commerce/platform/services/user/internal/domain"
	"github.com/aura-commerce/platform/services/user/internal/service"
)

// Handlers adapts HTTP requests onto the service layer. It is deliberately thin:
// decode, validate, call the service, map the result (or error) to JSON.
type Handlers struct {
	svc      *service.Service
	validate *validate.Validator
}

// NewHandlers builds the HTTP handler set.
func NewHandlers(svc *service.Service, v *validate.Validator) *Handlers {
	return &Handlers{svc: svc, validate: v}
}

// Register handles POST /api/v1/auth/register.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	user, pair, err := h.svc.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toAuthResponse(user, pair))
}

// Login handles POST /api/v1/auth/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	user, pair, err := h.svc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAuthResponse(user, pair))
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	// No user object on refresh — just the new tokens.
	httpx.JSON(w, http.StatusOK, authResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    pair.TokenType,
		ExpiresIn:    pair.ExpiresIn,
	})
}

// Me handles GET /api/v1/users/me (requires authentication).
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}

	user, err := h.svc.Profile(r.Context(), claims.UserID())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateMe handles PATCH /api/v1/users/me (requires authentication).
func (h *Handlers) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}

	var req updateProfileRequest
	if err := h.bind(r, &req); err != nil {
		httpx.Error(w, err)
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), claims.UserID(), req.Name)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(user))
}

// bind decodes the JSON body and runs struct validation, returning a ready-made
// API error on failure.
func (h *Handlers) bind(r *http.Request, dst any) error {
	if err := httpx.Decode(r, dst); err != nil {
		return err
	}
	if fields := h.validate.Struct(dst); fields != nil {
		return httpx.ErrValidation("some fields are invalid").WithFields(fields)
	}
	return nil
}

// writeServiceError maps domain errors onto HTTP responses. Anything unexpected
// becomes a generic 500 (httpx.Error handles that fallback), so internal
// details never leak.
func (h *Handlers) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		httpx.Error(w, httpx.ErrConflict("email_taken", "that email is already registered").
			WithFields(map[string]string{"email": "already registered"}))
	case errors.Is(err, domain.ErrInvalidCredentials):
		httpx.Error(w, httpx.ErrUnauthorized("invalid email or password"))
	case errors.Is(err, domain.ErrUserNotFound):
		httpx.Error(w, httpx.ErrNotFound("user not found"))
	default:
		httpx.Error(w, err)
	}
}
