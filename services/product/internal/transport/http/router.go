package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

// Pinger / BusChecker are the slices of the pool and event bus the router needs
// for readiness checks.
type Pinger interface {
	Ping(ctx context.Context) error
}
type BusChecker interface {
	Healthy() bool
}

// RouterDeps bundles everything NewRouter needs.
type RouterDeps struct {
	Handlers      *Handlers
	Tokens        *auth.Manager
	Log           *slog.Logger
	AllowedOrigin string
	DB            Pinger
	Bus           BusChecker
}

// NewRouter builds the product service's HTTP handler: platform middleware,
// health probes, public reads and admin-gated writes.
func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(d.Log))
	r.Use(middleware.Logger(d.Log))
	r.Use(middleware.CORS(d.AllowedOrigin))

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

	r.Route("/api/v1/products", func(p chi.Router) {
		// Public catalog reads.
		p.Get("/", d.Handlers.List)
		p.Get("/{slug}", d.Handlers.GetOne)

		// Admin-only writes: must be authenticated AND have the admin role.
		p.Group(func(admin chi.Router) {
			admin.Use(middleware.Authenticated(d.Tokens))
			admin.Use(middleware.RequireRole(auth.RoleAdmin))
			admin.Post("/", d.Handlers.Create)
			admin.Post("/{id}/stock", d.Handlers.AdjustStock)
		})
	})

	return r
}
