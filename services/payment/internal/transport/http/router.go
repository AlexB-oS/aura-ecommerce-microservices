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

type Pinger interface {
	Ping(ctx context.Context) error
}
type BusChecker interface {
	Healthy() bool
}

type RouterDeps struct {
	Handlers      *Handlers
	Tokens        *auth.Manager
	Log           *slog.Logger
	AllowedOrigin string
	DB            Pinger
	Bus           BusChecker
}

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

	r.Route("/api/v1/payments", func(p chi.Router) {
		p.Use(middleware.Authenticated(d.Tokens))
		p.Get("/", d.Handlers.List)
	})

	return r
}
