package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHTTPAddr            = ":8085"
	DefaultLogLevel            = slog.LevelInfo
	DefaultRateLimitCapacity   = 10
	DefaultRateLimitRefillRate = 5.0
)

type Config struct {
	HTTP      HTTPConfig
	Logging   LoggingConfig
	RateLimit RateLimitConfig
}

type HTTPConfig struct {
	Addr string
}

type LoggingConfig struct {
	Level slog.Level
}

type RateLimitConfig struct {
	Capacity   int
	RefillRate float64
	KeyHeader  string
}

func Load() (Config, error) {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = DefaultHTTPAddr
	}

	logLevel := DefaultLogLevel
	if value := os.Getenv("LOG_LEVEL"); value != "" {
		switch strings.ToLower(value) {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			return Config{}, fmt.Errorf("invalid LOG_LEVEL: must be debug, info, warn or error")
		}
	}

	capacity := DefaultRateLimitCapacity
	if value := os.Getenv("RATE_LIMIT_CAPACITY"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_CAPACITY: %w", err)
		}
		if parsed <= 0 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_CAPACITY: must be greater than zero")
		}
		capacity = parsed
	}

	refillRate := DefaultRateLimitRefillRate
	if value := os.Getenv("RATE_LIMIT_REFILL_RATE"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_REFILL_RATE: %w", err)
		}
		if parsed <= 0 {
			return Config{}, fmt.Errorf("invalid RATE_LIMIT_REFILL_RATE: must be greater than zero")
		}
		refillRate = parsed
	}

	keyHeader := strings.TrimSpace(os.Getenv("RATE_LIMIT_KEY_HEADER"))

	return Config{
		HTTP: HTTPConfig{
			Addr: httpAddr,
		},
		Logging: LoggingConfig{
			Level: logLevel,
		},
		RateLimit: RateLimitConfig{
			Capacity:   capacity,
			RefillRate: refillRate,
			KeyHeader:  keyHeader,
		},
	}, nil
}
