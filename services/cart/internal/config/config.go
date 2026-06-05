// Package config gathers the cart service's runtime configuration.
package config

import (
	"time"

	"github.com/aura-commerce/platform/pkg/config"
)

type Config struct {
	ServiceName       string
	HTTPAddr          string
	LogLevel          string
	RedisURL          string        // redis connection URL, e.g. redis://redis:6379/0
	CartTTL           time.Duration // how long an idle cart lives in Redis
	ProductServiceURL string        // base URL for synchronous catalog lookups
	JWTSecret         string        // to authenticate the cart owner
	JWTIssuer         string
	AllowedOrigin     string
}

func Load() Config {
	return Config{
		ServiceName:       config.String("SERVICE_NAME", "cart-service"),
		HTTPAddr:          config.String("HTTP_ADDR", ":8080"),
		LogLevel:          config.String("LOG_LEVEL", "info"),
		RedisURL:          config.String("REDIS_URL", "redis://redis:6379/0"),
		CartTTL:           config.Duration("CART_TTL", 30*24*time.Hour),
		ProductServiceURL: config.String("PRODUCT_SERVICE_URL", "http://product-service:8080"),
		JWTSecret:         config.MustString("JWT_SECRET"),
		JWTIssuer:         config.String("JWT_ISSUER", "aura.auth"),
		AllowedOrigin:     config.String("ALLOWED_ORIGIN", "*"),
	}
}
