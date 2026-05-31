// Package config gathers all runtime configuration for the user service from
// the environment into one typed struct, loaded once at startup.
package config

import (
	"time"

	"github.com/aura-commerce/platform/pkg/config"
)

// Config is the fully-resolved configuration for the service.
type Config struct {
	ServiceName   string
	HTTPAddr      string        // address the HTTP server listens on, e.g. ":8080"
	LogLevel      string        // debug|info|warn|error
	DatabaseURL   string        // postgres connection string
	NATSURL       string        // nats connection string
	JWTSecret     string        // HMAC signing secret shared across services
	JWTIssuer     string        // "iss" claim
	AccessTTL     time.Duration // access token lifetime
	RefreshTTL    time.Duration // refresh token lifetime
	AllowedOrigin string        // CORS origin for the storefront
}

// Load reads configuration from the environment, applying safe defaults for
// everything except the two secrets that must never be defaulted.
func Load() Config {
	return Config{
		ServiceName:   config.String("SERVICE_NAME", "user-service"),
		HTTPAddr:      config.String("HTTP_ADDR", ":8080"),
		LogLevel:      config.String("LOG_LEVEL", "info"),
		DatabaseURL:   config.MustString("DATABASE_URL"),
		NATSURL:       config.String("NATS_URL", "nats://nats:4222"),
		JWTSecret:     config.MustString("JWT_SECRET"),
		JWTIssuer:     config.String("JWT_ISSUER", "aura.auth"),
		AccessTTL:     config.Duration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    config.Duration("JWT_REFRESH_TTL", 7*24*time.Hour),
		AllowedOrigin: config.String("ALLOWED_ORIGIN", "*"),
	}
}
