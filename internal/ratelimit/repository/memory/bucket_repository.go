package memory

import (
	"context"
	"sync"
	"time"

	"rate-limiter/internal/ratelimit/entity"
)

type BucketRepository struct {
	mu      sync.Mutex
	buckets map[string]entity.Bucket
}

func NewBucketRepository() *BucketRepository {
	return &BucketRepository{
		buckets: make(map[string]entity.Bucket),
	}
}

func (r *BucketRepository) UpdateBucket(ctx context.Context, key string, initial entity.Bucket, update func(*entity.Bucket)) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, ok := r.buckets[key]
	if !ok {
		bucket = initial
	}

	update(&bucket)
	r.buckets[key] = bucket
	return nil
}

func (r *BucketRepository) CountBuckets(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.buckets), nil
}

func (r *BucketRepository) DeleteIdleBefore(ctx context.Context, cutoff time.Time) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	removed := 0
	for key, bucket := range r.buckets {
		if bucket.LastSeen.Before(cutoff) {
			delete(r.buckets, key)
			removed++
		}
	}

	return removed, nil
}
