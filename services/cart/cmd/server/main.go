// Command server is the entrypoint for the cart service (Redis-backed).
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
	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/pkg/validate"
	"github.com/aura-commerce/platform/services/cart/internal/catalog"
	"github.com/aura-commerce/platform/services/cart/internal/config"
	cartredis "github.com/aura-commerce/platform/services/cart/internal/repository/redis"
	"github.com/aura-commerce/platform/services/cart/internal/service"
	httptransport "github.com/aura-commerce/platform/services/cart/internal/transport/http"
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

	// --- Redis --------------------------------------------------------------
	rdb, err := cartredis.NewClient(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	defer rdb.Close()
	repo := cartredis.NewCartRepository(rdb, cfg.CartTTL)

	// Verify connectivity with a short retry; Redis may still be booting.
	if err := retryPing(ctx, repo); err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	log.Info("connected to redis")
	catalogClient := catalog.New(cfg.ProductServiceURL)
	svc := service.New(repo, catalogClient, log)

	tokens := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, 15*time.Minute, 168*time.Hour)
	handlers := httptransport.NewHandlers(svc, validate.New())

	router := httptransport.NewRouter(httptransport.RouterDeps{
		Handlers:      handlers,
		Tokens:        tokens,
		Log:           log,
		AllowedOrigin: cfg.AllowedOrigin,
		Redis:         repo,
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

// retryPing retries the initial Redis connection a few times, since the service
// container can start before Redis is accepting connections.
func retryPing(ctx context.Context, p interface{ Ping(context.Context) error }) error {
	var last error
	for attempt := 1; attempt <= 10; attempt++ {
		if last = p.Ping(ctx); last == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
		}
	}
	return last
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
