package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"rate-limiter/cmd/config"
	"rate-limiter/internal/platform/logging"
	"rate-limiter/internal/ratelimit"
	ratelimithttp "rate-limiter/internal/ratelimit/handler/http"
	"rate-limiter/internal/ratelimit/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewJSONLogger(os.Stdout, cfg.Logging.Level)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	handlerOpts := make([]ratelimithttp.Option, 0, 1)
	if cfg.RateLimit.KeyHeader != "" {
		keyHeader := cfg.RateLimit.KeyHeader
		handlerOpts = append(handlerOpts, ratelimithttp.WithKeyFunc(func(r *http.Request) string {
			if value := strings.TrimSpace(r.Header.Get(keyHeader)); value != "" {
				return value
			}
			return ratelimithttp.DefaultKeyFunc(r)
		}))
	}

	rateLimitModule, err := ratelimit.NewModule(service.TokenBucketConfig{
		Capacity:   cfg.RateLimit.Capacity,
		RefillRate: cfg.RateLimit.RefillRate,
	}, logger, handlerOpts...)
	if err != nil {
		logger.Error("failed to create rate limit module", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           rateLimitModule.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-rootCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("rate limiter service started",
		"addr", server.Addr,
		"capacity", cfg.RateLimit.Capacity,
		"refill_rate", cfg.RateLimit.RefillRate,
		"key_header", cfg.RateLimit.KeyHeader,
	)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("http server failed", "error", err)
		os.Exit(1)
	}

	logger.Info("rate limiter service stopped")
}
