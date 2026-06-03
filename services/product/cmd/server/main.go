// Command server is the entrypoint for the product catalog service. It mirrors
// the user service: wire layers, serve HTTP, shut down gracefully, and support
// a dependency-free `healthcheck` subcommand for the Docker HEALTHCHECK.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aura-commerce/platform/pkg/auth"
	"github.com/aura-commerce/platform/pkg/db"
	"github.com/aura-commerce/platform/pkg/events"
	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/pkg/validate"
	"github.com/aura-commerce/platform/services/product/internal/config"
	"github.com/aura-commerce/platform/services/product/internal/repository/postgres"
	"github.com/aura-commerce/platform/services/product/internal/service"
	"github.com/aura-commerce/platform/services/product/internal/subscriber"
	httptransport "github.com/aura-commerce/platform/services/product/internal/transport/http"
	"github.com/aura-commerce/platform/services/product/migrations"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting", "addr", cfg.HTTPAddr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("migrations applied")

	bus, err := events.Connect(cfg.NATSURL, cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}
	defer bus.Close()
	log.Info("connected to event bus")

	// The product service only VERIFIES tokens (for admin endpoints); it never
	// issues them, so the TTLs here are unused — they matter only to Issue.
	tokens := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, 15*time.Minute, 168*time.Hour)
	repo := postgres.NewProductRepository(pool)
	svc := service.New(repo, bus, log, cfg.DefaultPage)

	// React to order.placed by decrementing stock for each line item.
	if err := subscriber.Register(bus, svc, log); err != nil {
		return fmt.Errorf("register subscribers: %w", err)
	}
	log.Info("subscribed to order events")

	handlers := httptransport.NewHandlers(svc, validate.New())

	router := httptransport.NewRouter(httptransport.RouterDeps{
		Handlers:      handlers,
		Tokens:        tokens,
		Log:           log,
		AllowedOrigin: cfg.AllowedOrigin,
		DB:            pool,
		Bus:           bus,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("stopped cleanly")
	return nil
}

func healthcheck() int {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1" + addr + "/healthz")
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
