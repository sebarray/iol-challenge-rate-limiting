package repository

import (
	"context"
	"time"

	"rate-limiter/internal/ratelimit/entity"
)

type BucketRepository interface {
	UpdateBucket(ctx context.Context, key string, initial entity.Bucket, update func(*entity.Bucket)) error
	CountBuckets(ctx context.Context) (int, error)
	DeleteIdleBefore(ctx context.Context, cutoff time.Time) (int, error)
}
