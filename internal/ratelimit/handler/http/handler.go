package http

import (
	"log/slog"
	nethttp "net/http"

	"rate-limiter/internal/ratelimit/service"
)

type KeyFunc func(*nethttp.Request) string

type Handler struct {
	limiter service.Limiter
	logger  *slog.Logger
	keyFunc KeyFunc
}

type Option func(*Handler)

func WithKeyFunc(keyFunc KeyFunc) Option {
	return func(h *Handler) {
		if keyFunc != nil {
			h.keyFunc = keyFunc
		}
	}
}

func NewHandler(limiter service.Limiter, logger *slog.Logger, opts ...Option) *Handler {
	if logger == nil {
		logger = slog.Default()
	}

	h := &Handler{
		limiter: limiter,
		logger:  logger,
		keyFunc: DefaultKeyFunc,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
