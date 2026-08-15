package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"rate-limiter/internal/platform/clock"
	"rate-limiter/internal/ratelimit/entity"
	"rate-limiter/internal/ratelimit/repository"
)

type tokenBucketService struct {
	bucketRepo repository.BucketRepository
	capacity   float64
	refillRate float64
	clock      clock.Clock
}

func NewTokenBucketService(bucketRepo repository.BucketRepository, cfg TokenBucketConfig) (*tokenBucketService, error) {
	if bucketRepo == nil {
		return nil, errors.New("bucket repository is required")
	}
	if cfg.Capacity <= 0 {
		return nil, errors.New("capacity must be greater than zero")
	}
	if cfg.RefillRate <= 0 {
		return nil, errors.New("refill rate must be greater than zero")
	}

	clockRef := cfg.Clock
	if clockRef == nil {
		clockRef = clock.SystemClock{}
	}

	return &tokenBucketService{
		bucketRepo: bucketRepo,
		capacity:   float64(cfg.Capacity),
		refillRate: cfg.RefillRate,
		clock:      clockRef,
	}, nil
}

func (s *tokenBucketService) Allow(ctx context.Context, key string) (entity.Decision, error) {
	if err := ctx.Err(); err != nil {
		return entity.Decision{}, err
	}
	if strings.TrimSpace(key) == "" {
		return entity.Decision{}, entity.ErrEmptyKey
	}

	now := s.clock.Now()

	var decision entity.Decision
	initialBucket := entity.Bucket{
		Tokens:     s.capacity,
		LastRefill: now,
		LastSeen:   now,
	}

	err := s.bucketRepo.UpdateBucket(ctx, key, initialBucket, func(bucket *entity.Bucket) {
		s.refill(bucket, now)
		bucket.LastSeen = now

		decision = entity.Decision{
			Allowed:   false,
			Limit:     int(s.capacity),
			Remaining: int(math.Floor(bucket.Tokens)),
		}

		if bucket.Tokens >= 1 {
			bucket.Tokens--
			decision.Allowed = true
			decision.Remaining = int(math.Floor(bucket.Tokens))
		} else {
			decision.RetryAfter = durationFromSeconds((1 - bucket.Tokens) / s.refillRate)
		}

		decision.ResetAfter = durationFromSeconds((s.capacity - bucket.Tokens) / s.refillRate)
	})
	if err != nil {
		return entity.Decision{}, err
	}

	return decision, nil
}

func (s *tokenBucketService) refill(bucket *entity.Bucket, now time.Time) {
	elapsed := now.Sub(bucket.LastRefill)
	if elapsed <= 0 {
		return
	}

	bucket.Tokens = math.Min(s.capacity, bucket.Tokens+elapsed.Seconds()*s.refillRate)
	bucket.LastRefill = now
}

func durationFromSeconds(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(math.Ceil(seconds * float64(time.Second)))
}
