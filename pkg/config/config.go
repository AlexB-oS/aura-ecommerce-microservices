// Package config provides tiny, dependency-free helpers for reading typed
// values from environment variables. The 12-factor approach (config in the
// environment) is what lets the same container image run unchanged in local
// Docker Compose, CI, and production.
package config

import (
	"os"
	"strconv"
	"time"
)

// String returns the env var or a fallback default.
func String(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// MustString returns the env var or panics. Use for values with no safe
// default (database URLs, signing secrets) so misconfiguration fails loudly at
// startup instead of silently at request time.
func MustString(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

// Int returns the env var parsed as an int, or the fallback.
func Int(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// Duration returns the env var parsed as a Go duration (e.g. "15m", "168h"), or
// the fallback.
func Duration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
