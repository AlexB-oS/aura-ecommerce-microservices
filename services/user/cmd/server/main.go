// Command server is the entrypoint for the user service. It wires the layers
// together (config → logger → db → events → service → http), starts the HTTP
// server, and shuts everything down gracefully on SIGINT/SIGTERM.
//
// It also supports a "healthcheck" subcommand used by the Docker HEALTHCHECK:
// because the runtime image is distroless (no shell, no curl), the binary
// probes itself.
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
	"github.com/aura-commerce/platform/services/user/internal/config"
	"github.com/aura-commerce/platform/services/user/internal/repository/postgres"
	"github.com/aura-commerce/platform/services/user/internal/service"
	httptransport "github.com/aura-commerce/platform/services/user/internal/transport/http"
	"github.com/aura-commerce/platform/services/user/migrations"
)

func main() {
	// Docker HEALTHCHECK calls `user healthcheck`; handle it before any of the
	// heavyweight startup so the probe is cheap and dependency-free.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}

	if err := run(); err != nil {
		// Use stderr directly: the structured logger may not exist yet if config
		// loading panicked.
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	log := logger.New(cfg.ServiceName, cfg.LogLevel)
	log.Info("starting", "addr", cfg.HTTPAddr)

	// Root context cancelled on shutdown signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database -----------------------------------------------------------
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("migrations applied")

	// --- Event bus ----------------------------------------------------------
	bus, err := events.Connect(cfg.NATSURL, cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}
	defer bus.Close()
	log.Info("connected to event bus")

	// --- Wire the application ----------------------------------------------
	tokens := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.AccessTTL, cfg.RefreshTTL)
	repo := postgres.NewUserRepository(pool)
	svc := service.New(repo, tokens, bus, log)
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

	// --- Serve until signalled, then drain ---------------------------------
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

// healthcheck performs an HTTP GET against the local /healthz endpoint and maps
// the result to a process exit code (0 healthy, 1 unhealthy).
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
