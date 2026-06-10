// Command server is the entrypoint for the notification service. It has no HTTP
// API of its own beyond a demo feed — its job is to subscribe to events and
// "send" messages.
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

	"github.com/aura-commerce/platform/pkg/events"
	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/aura-commerce/platform/services/notification/internal/config"
	"github.com/aura-commerce/platform/services/notification/internal/service"
	"github.com/aura-commerce/platform/services/notification/internal/subscriber"
	httptransport "github.com/aura-commerce/platform/services/notification/internal/transport/http"
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

	bus, err := events.Connect(cfg.NATSURL, cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}
	defer bus.Close()
	log.Info("connected to event bus")

	notifier := service.New(log)
	if err := subscriber.Register(bus, notifier, log); err != nil {
		return fmt.Errorf("register subscribers: %w", err)
	}
	log.Info("subscribed to platform events")

	router := httptransport.NewRouter(httptransport.RouterDeps{
		Notifier:      notifier,
		Log:           log,
		AllowedOrigin: cfg.AllowedOrigin,
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
