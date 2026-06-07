// Package config gathers the order service's runtime configuration.
package config

import "github.com/aura-commerce/platform/pkg/config"

type Config struct {
	ServiceName    string
	HTTPAddr       string
	LogLevel       string
	DatabaseURL    string
	NATSURL        string
	CartServiceURL string // base URL for reading/clearing the user's cart
	JWTSecret      string
	JWTIssuer      string
	AllowedOrigin  string
}

func Load() Config {
	return Config{
		ServiceName:    config.String("SERVICE_NAME", "order-service"),
		HTTPAddr:       config.String("HTTP_ADDR", ":8080"),
		LogLevel:       config.String("LOG_LEVEL", "info"),
		DatabaseURL:    config.MustString("DATABASE_URL"),
		NATSURL:        config.String("NATS_URL", "nats://nats:4222"),
		CartServiceURL: config.String("CART_SERVICE_URL", "http://cart-service:8080"),
		JWTSecret:      config.MustString("JWT_SECRET"),
		JWTIssuer:      config.String("JWT_ISSUER", "aura.auth"),
		AllowedOrigin:  config.String("ALLOWED_ORIGIN", "*"),
	}
}
