package ratelimit

import (
	"log/slog"
	nethttp "net/http"

	ratelimithttp "rate-limiter/internal/ratelimit/handler/http"
	"rate-limiter/internal/ratelimit/repository"
	"rate-limiter/internal/ratelimit/repository/memory"
	"rate-limiter/internal/ratelimit/service"
)

type Module struct {
	Repository repository.BucketRepository
	Service    service.Limiter
	Handler    *ratelimithttp.Handler
}

func NewModule(cfg service.TokenBucketConfig, logger *slog.Logger, handlerOpts ...ratelimithttp.Option) (*Module, error) {
	bucketRepo := memory.NewBucketRepository()

	rateLimitService, err := service.NewTokenBucketService(bucketRepo, cfg)
	if err != nil {
		return nil, err
	}

	handler := ratelimithttp.NewHandler(rateLimitService, logger, handlerOpts...)

	return &Module{
		Repository: bucketRepo,
		Service:    rateLimitService,
		Handler:    handler,
	}, nil
}

func (m *Module) Routes() nethttp.Handler {
	return m.Handler.Routes()
}
