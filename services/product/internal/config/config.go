// Package config gathers the product service's runtime configuration from the
// environment. It mirrors the user service's config so every service is
// configured the same 12-factor way.
package config

import "github.com/aura-commerce/platform/pkg/config"

// Config is the fully-resolved configuration for the service.
type Config struct {
	ServiceName   string
	HTTPAddr      string
	LogLevel      string
	DatabaseURL   string
	NATSURL       string
	JWTSecret     string // shared across services so this service can verify admin tokens
	JWTIssuer     string
	AllowedOrigin string
	DefaultPage   int // default List() page size
}

// Load reads configuration from the environment with safe defaults.
func Load() Config {
	return Config{
		ServiceName:   config.String("SERVICE_NAME", "product-service"),
		HTTPAddr:      config.String("HTTP_ADDR", ":8080"),
		LogLevel:      config.String("LOG_LEVEL", "info"),
		DatabaseURL:   config.MustString("DATABASE_URL"),
		NATSURL:       config.String("NATS_URL", "nats://nats:4222"),
		JWTSecret:     config.MustString("JWT_SECRET"),
		JWTIssuer:     config.String("JWT_ISSUER", "aura.auth"),
		AllowedOrigin: config.String("ALLOWED_ORIGIN", "*"),
		DefaultPage:   config.Int("DEFAULT_PAGE_SIZE", 24),
	}
}
