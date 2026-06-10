// Package config gathers the notification service's runtime configuration. This
// service has no database and issues no tokens — it just listens to the bus.
package config

import "github.com/aura-commerce/platform/pkg/config"

type Config struct {
	ServiceName   string
	HTTPAddr      string
	LogLevel      string
	NATSURL       string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		ServiceName:   config.String("SERVICE_NAME", "notification-service"),
		HTTPAddr:      config.String("HTTP_ADDR", ":8080"),
		LogLevel:      config.String("LOG_LEVEL", "info"),
		NATSURL:       config.String("NATS_URL", "nats://nats:4222"),
		AllowedOrigin: config.String("ALLOWED_ORIGIN", "*"),
	}
}
