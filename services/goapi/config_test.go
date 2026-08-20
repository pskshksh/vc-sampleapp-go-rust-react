package main

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"WARN":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		" error ": slog.LevelError,
	}
	for in, want := range cases {
		got, err := parseLogLevel(in)
		if err != nil {
			t.Errorf("parseLogLevel(%q) returned error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", in, got, want)
		}
	}

	if _, err := parseLogLevel("nonsense"); err == nil {
		t.Error("parseLogLevel(\"nonsense\") should return an error")
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("GOAPI_TEST_KEY", "set-value")
	if got := envOr("GOAPI_TEST_KEY", "fallback"); got != "set-value" {
		t.Errorf("envOr with set var = %q, want %q", got, "set-value")
	}
	if got := envOr("GOAPI_TEST_MISSING", "fallback"); got != "fallback" {
		t.Errorf("envOr with unset var = %q, want %q", got, "fallback")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Ensure a clean environment so the documented defaults apply.
	for _, k := range []string{"GOAPI_ADDR", "RSCOUNTER_URL", "RSCOUNTER_TIMEOUT", "LOG_LEVEL"} {
		t.Setenv(k, "")
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.RscounterURL != "http://localhost:8081" {
		t.Errorf("RscounterURL = %q, want %q", cfg.RscounterURL, "http://localhost:8081")
	}
	if cfg.RscounterTimeout.String() != "5s" {
		t.Errorf("RscounterTimeout = %v, want 5s", cfg.RscounterTimeout)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
}

func TestLoadConfigInvalidTimeout(t *testing.T) {
	t.Setenv("RSCOUNTER_TIMEOUT", "not-a-duration")
	if _, err := LoadConfig(); err == nil {
		t.Error("LoadConfig() should error on an invalid RSCOUNTER_TIMEOUT")
	}
}
