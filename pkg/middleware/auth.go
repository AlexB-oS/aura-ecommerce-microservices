package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/pkg/httpx"
)

type ctxKey struct{ name string }

var claimsKey = ctxKey{"claims"}

// Authenticated is HTTP middleware that requires a valid Bearer access token.
// On success the verified claims are stored on the request context; on failure
// it short-circuits with a 401 in the standard error envelope.
//
// In this architecture each service validates the JWT itself using the shared
// secret, so there is no extra network hop to the user service on every
// request. The user service is only involved at login / refresh time.
func Authenticated(tokens *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				httpx.Error(w, httpx.ErrUnauthorized("missing bearer token"))
				return
			}

			claims, err := tokens.Verify(token, auth.AccessToken)
			if err != nil {
				httpx.Error(w, httpx.ErrUnauthorized("invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFrom returns the authenticated user's claims from the context. The
// second return is false if the request did not pass through Authenticated.
func ClaimsFrom(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}

// RequireRole returns middleware that allows only users whose token carries one
// of the given roles. It must be chained AFTER Authenticated. Any other request
// is rejected with 403.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFrom(r.Context())
			if !ok {
				httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
				return
			}
			if _, allow := allowed[claims.Role]; !allow {
				httpx.Error(w, httpx.ErrForbidden("you do not have access to this resource"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
