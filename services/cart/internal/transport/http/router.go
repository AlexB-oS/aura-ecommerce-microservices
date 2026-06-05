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

type RouterDeps struct {
	Handlers      *Handlers
	Tokens        *auth.Manager
	Log           *slog.Logger
	AllowedOrigin string
	Redis         Pinger
}

// NewRouter builds the cart service's HTTP handler. Every cart route requires a
// valid access token — a cart belongs to exactly one authenticated user.
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
		checks := map[string]string{"redis": "ok"}
		status := http.StatusOK
		if err := d.Redis.Ping(ctx); err != nil {
			checks["redis"] = "unreachable"
			status = http.StatusServiceUnavailable
		}
		httpx.JSON(w, status, checks)
	})

	r.Route("/api/v1/cart", func(c chi.Router) {
		c.Use(middleware.Authenticated(d.Tokens))
		c.Get("/", d.Handlers.Get)
		c.Delete("/", d.Handlers.Clear)
		c.Post("/items", d.Handlers.AddItem)
		c.Patch("/items/{productId}", d.Handlers.SetQty)
		c.Delete("/items/{productId}", d.Handlers.RemoveItem)
	})

	return r
}
