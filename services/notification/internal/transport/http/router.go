package http

import (
	"log/slog"
	"net/http"

	"github.com/aura-commerce/platform/pkg/httpx"
	"github.com/aura-commerce/platform/pkg/middleware"
	"github.com/aura-commerce/platform/services/notification/internal/service"
	"github.com/go-chi/chi/v5"
)

type BusChecker interface {
	Healthy() bool
}

type RouterDeps struct {
	Notifier      *service.Notifier
	Log           *slog.Logger
	AllowedOrigin string
	Bus           BusChecker
}

// NewRouter builds the notification service's HTTP handler. The recent-feed
// endpoint is public so you can watch notifications fire during a demo (it would
// be locked down or removed in production).
func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recover(d.Log))
	r.Use(middleware.Logger(d.Log))
	r.Use(middleware.CORS(d.AllowedOrigin))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		status := http.StatusOK
		checks := map[string]string{"events": "ok"}
		if d.Bus == nil || !d.Bus.Healthy() {
			checks["events"] = "unreachable"
			status = http.StatusServiceUnavailable
		}
		httpx.JSON(w, status, checks)
	})

	r.Get("/api/v1/notifications", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, d.Notifier.Recent())
	})

	return r
}
