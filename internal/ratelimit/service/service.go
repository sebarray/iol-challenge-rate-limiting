package service

import (
	"context"

	"rate-limiter/internal/platform/clock"
	"rate-limiter/internal/ratelimit/entity"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (entity.Decision, error)
}

type TokenBucketConfig struct {
	Capacity   int
	RefillRate float64
	Clock      clock.Clock
}
