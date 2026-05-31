package http

import (
	"context"
	"net/http"
	"time"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

// Pinger is the slice of *pgxpool.Pool the router needs for readiness checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// BusChecker is the slice of the event bus the router needs for readiness.
type BusChecker interface {
	Healthy() bool
}

// RouterDeps bundles everything NewRouter needs. Passing a struct keeps the
// signature stable as dependencies grow.
type RouterDeps struct {
	Handlers      *Handlers
	Tokens        *auth.Manager
	Log           *slog.Logger
	AllowedOrigin string
	DB            Pinger
	Bus           BusChecker
}

// NewRouter builds the fully-configured HTTP handler for the user service:
// platform middleware, health probes, and the versioned API surface.
func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Order matters: assign a request id first so every later log line and
	// recovered panic can be correlated, then recover, then log, then CORS.
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(d.Log))
	r.Use(middleware.Logger(d.Log))
	r.Use(middleware.CORS(d.AllowedOrigin))

	// Kubernetes-style probes. /healthz is liveness (is the process up?),
	// /readyz is readiness (can it actually serve traffic — DB + bus reachable?).
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()

		checks := map[string]string{"database": "ok", "events": "ok"}
		ready := true
		if err := d.DB.Ping(ctx); err != nil {
			checks["database"] = "unreachable"
			ready = false
		}
		if d.Bus == nil || !d.Bus.Healthy() {
			checks["events"] = "unreachable"
			ready = false
		}
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		httpx.JSON(w, status, checks)
	})

	// Versioned API. The gateway strips/forwards these prefixes verbatim.
	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(authR chi.Router) {
			authR.Post("/register", d.Handlers.Register)
			authR.Post("/login", d.Handlers.Login)
			authR.Post("/refresh", d.Handlers.Refresh)
		})

		api.Route("/users", func(users chi.Router) {
			// Everything under /users requires a valid access token.
			users.Use(middleware.Authenticated(d.Tokens))
			users.Get("/me", d.Handlers.Me)
			users.Patch("/me", d.Handlers.UpdateMe)
		})
	})

	return r
}
