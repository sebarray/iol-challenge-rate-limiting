package config

import (
	"log/slog"
	"testing"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("RATE_LIMIT_CAPACITY", "")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "")
	t.Setenv("RATE_LIMIT_KEY_HEADER", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTP.Addr != DefaultHTTPAddr {
		t.Fatalf("HTTP addr = %q, want %q", cfg.HTTP.Addr, DefaultHTTPAddr)
	}
	if cfg.Logging.Level != DefaultLogLevel {
		t.Fatalf("log level = %s, want %s", cfg.Logging.Level, DefaultLogLevel)
	}
	if cfg.RateLimit.Capacity != DefaultRateLimitCapacity {
		t.Fatalf("capacity = %d, want %d", cfg.RateLimit.Capacity, DefaultRateLimitCapacity)
	}
	if cfg.RateLimit.RefillRate != DefaultRateLimitRefillRate {
		t.Fatalf("refill rate = %f, want %f", cfg.RateLimit.RefillRate, DefaultRateLimitRefillRate)
	}
	if cfg.RateLimit.KeyHeader != "" {
		t.Fatalf("key header = %q, want empty", cfg.RateLimit.KeyHeader)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("RATE_LIMIT_CAPACITY", "25")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "12.5")
	t.Setenv("RATE_LIMIT_KEY_HEADER", " X-Real-IP ")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("HTTP addr = %q, want :9090", cfg.HTTP.Addr)
	}
	if cfg.Logging.Level != slog.LevelDebug {
		t.Fatalf("log level = %s, want debug", cfg.Logging.Level)
	}
	if cfg.RateLimit.Capacity != 25 {
		t.Fatalf("capacity = %d, want 25", cfg.RateLimit.Capacity)
	}
	if cfg.RateLimit.RefillRate != 12.5 {
		t.Fatalf("refill rate = %f, want 12.5", cfg.RateLimit.RefillRate)
	}
	if cfg.RateLimit.KeyHeader != "X-Real-IP" {
		t.Fatalf("key header = %q, want X-Real-IP", cfg.RateLimit.KeyHeader)
	}
}

func TestLoadRejectsInvalidCapacity(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("RATE_LIMIT_CAPACITY", "0")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsInvalidRefillRate(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("RATE_LIMIT_CAPACITY", "")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "nope")

	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("RATE_LIMIT_CAPACITY", "")
	t.Setenv("RATE_LIMIT_REFILL_RATE", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}
