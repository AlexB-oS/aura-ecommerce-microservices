// Package middleware contains HTTP middleware shared by every service: request
// IDs, panic recovery, structured access logging, and CORS. They are plain
// net/http middleware so they compose with chi, the stdlib, or anything else.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/aura-commerce/platform/pkg/logger"
	"github.com/google/uuid"
)

// RequestID ensures every request carries an X-Request-ID. If the gateway
// already set one it is reused, otherwise a new UUID is generated. The id is
// stored on the context and echoed back on the response so it can be traced
// end-to-end across services.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := logger.WithRequestID(r.Context(), id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusRecorder captures the status code so the logger can report it; the
// stdlib ResponseWriter does not expose the written status.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Logger emits one structured line per request with method, path, status,
// duration and the request id, so logs line up with the rest of the platform.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.LogAttrs(r.Context(), slog.LevelInfo, "http_request",
				slog.String("request_id", logger.RequestID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote", r.RemoteAddr),
			)
		})
	}
}

// Recover turns a panic in any handler into a clean 500 instead of crashing the
// whole server, logging the panic with its request id for debugging.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.LogAttrs(r.Context(), slog.LevelError, "panic_recovered",
						slog.Any("panic", rec),
						slog.String("request_id", logger.RequestID(r.Context())),
						slog.String("path", r.URL.Path),
					)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":{"code":"internal_error","message":"something went wrong on our side"}}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS is a permissive development CORS policy. In production the allowed origin
// should be restricted to the storefront domain via the ALLOWED_ORIGIN env var,
// which the caller passes in here.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "300")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
