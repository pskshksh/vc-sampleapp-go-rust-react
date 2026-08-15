package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config holds all runtime settings. Every value is sourced from an environment
// variable so nothing is hardcoded in the request path.
type Config struct {
	Addr             string        // GOAPI_ADDR — address goapi listens on
	RscounterURL     string        // RSCOUNTER_URL — base URL of the rscounter service
	RscounterTimeout time.Duration // RSCOUNTER_TIMEOUT — client timeout for rscounter calls
	LogLevel         slog.Level    // LOG_LEVEL — debug|info|warn|error
}

// LoadConfig reads configuration from the environment, falling back to the
// documented defaults when a variable is unset.
func LoadConfig() (Config, error) {
	timeout, err := time.ParseDuration(envOr("RSCOUNTER_TIMEOUT", "5s"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid RSCOUNTER_TIMEOUT: %w", err)
	}

	level, err := parseLogLevel(envOr("LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Addr:             envOr("GOAPI_ADDR", ":8080"),
		RscounterURL:     envOr("RSCOUNTER_URL", "http://localhost:8081"),
		RscounterTimeout: timeout,
		LogLevel:         level,
	}, nil
}

func parseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid LOG_LEVEL: %q", s)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
