// Package logger provides a small wrapper around Go's standard structured
// logger (log/slog). Every service logs JSON to stdout so that a future log
// collector (Promtail, Fluent Bit, the ELK stack, ...) can scrape the
// container output without any per-service configuration.
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ctxKey is an unexported type so the request-id key cannot collide with keys
// from other packages stored in the same context.
type ctxKey struct{}

var requestIDKey = ctxKey{}

// New returns a JSON slog.Logger writing to stdout. `level` is one of
// debug/info/warn/error (case-insensitive); anything else falls back to info.
// `service` is attached to every line so logs can be filtered per service once
// they are aggregated centrally.
func New(service, level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return slog.New(handler).With(slog.String("service", service))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID stores a request id on the context so handlers and the logging
// middleware can correlate every log line that belongs to the same request.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request id stored on the context, or "" if none.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
