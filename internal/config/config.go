// Package config loads runtime configuration for the DSL tools.
package config

import (
	"os"
	"strings"
	"time"
)

const (
	EnvLSPDebounce     = "VALUE_DSL_LSP_DEBOUNCE"
	DefaultLSPDebounce = 200 * time.Millisecond
)

type Config struct {
	LSPDebounce time.Duration
}

func Load() Config {
	return Config{
		LSPDebounce: durationFromEnv(EnvLSPDebounce, DefaultLSPDebounce),
	}
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	if value == "0" {
		return 0
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}
