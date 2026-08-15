package memory

import (
	"context"
	"testing"
	"time"

	"rate-limiter/internal/ratelimit/entity"
)

func TestBucketRepositoryCreatesAndUpdatesBucket(t *testing.T) {
	repo := NewBucketRepository()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	err := repo.UpdateBucket(context.Background(), "client-a", entity.Bucket{
		Tokens:     10,
		LastRefill: now,
		LastSeen:   now,
	}, func(bucket *entity.Bucket) {
		bucket.Tokens--
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateBucket(context.Background(), "client-a", entity.Bucket{}, func(bucket *entity.Bucket) {
		if bucket.Tokens != 9 {
			t.Fatalf("tokens = %f, want 9", bucket.Tokens)
		}
		bucket.Tokens--
	})
	if err != nil {
		t.Fatal(err)
	}

	count, err := repo.CountBuckets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestBucketRepositoryDeletesIdleBuckets(t *testing.T) {
	repo := NewBucketRepository()
	oldSeen := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	recentSeen := oldSeen.Add(10 * time.Minute)

	err := repo.UpdateBucket(context.Background(), "old", entity.Bucket{LastSeen: oldSeen}, func(_ *entity.Bucket) {})
	if err != nil {
		t.Fatal(err)
	}
	err = repo.UpdateBucket(context.Background(), "recent", entity.Bucket{LastSeen: recentSeen}, func(_ *entity.Bucket) {})
	if err != nil {
		t.Fatal(err)
	}

	removed, err := repo.DeleteIdleBefore(context.Background(), oldSeen.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	count, err := repo.CountBuckets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}
