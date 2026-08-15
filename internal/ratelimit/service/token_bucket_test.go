package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"rate-limiter/internal/ratelimit/entity"
	"rate-limiter/internal/ratelimit/repository/memory"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock(now time.Time) *fakeClock {
	return &fakeClock{now: now}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestTokenBucketAllowsCapacityThenRejects(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{
		Capacity:   2,
		RefillRate: 1,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}

	first, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	if !first.Allowed || first.Remaining != 1 {
		t.Fatalf("first request = %+v, want allowed with 1 remaining", first)
	}

	second, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	if !second.Allowed || second.Remaining != 0 {
		t.Fatalf("second request = %+v, want allowed with 0 remaining", second)
	}

	third, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	if third.Allowed {
		t.Fatalf("third request = %+v, want rejected", third)
	}
	if third.RetryAfter != time.Second {
		t.Fatalf("retry after = %s, want 1s", third.RetryAfter)
	}
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{
		Capacity:   2,
		RefillRate: 2,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		decision, err := service.Allow(context.Background(), "client-a")
		if err != nil {
			t.Fatal(err)
		}
		if !decision.Allowed {
			t.Fatalf("request %d unexpectedly rejected", i+1)
		}
	}

	clock.Advance(500 * time.Millisecond)

	decision, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed {
		t.Fatalf("request after refill = %+v, want allowed", decision)
	}
	if decision.Remaining != 0 {
		t.Fatalf("remaining = %d, want 0", decision.Remaining)
	}
}

func TestTokenBucketDoesNotRefillPastCapacity(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{
		Capacity:   3,
		RefillRate: 10,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}

	clock.Advance(10 * time.Second)

	decision, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed || decision.Remaining != 2 {
		t.Fatalf("decision = %+v, want allowed with 2 remaining", decision)
	}
}

func TestTokenBucketUsesIndependentKeys(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{
		Capacity:   1,
		RefillRate: 1,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}

	allowedA, err := service.Allow(context.Background(), "client-a")
	if err != nil {
		t.Fatal(err)
	}
	allowedB, err := service.Allow(context.Background(), "client-b")
	if err != nil {
		t.Fatal(err)
	}

	if !allowedA.Allowed || !allowedB.Allowed {
		t.Fatalf("decisions = %+v, %+v, want both clients allowed", allowedA, allowedB)
	}
}

func TestTokenBucketRejectsInvalidInputs(t *testing.T) {
	if _, err := NewTokenBucketService(nil, TokenBucketConfig{Capacity: 1, RefillRate: 1}); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{Capacity: 0, RefillRate: 1}); err == nil {
		t.Fatal("expected capacity error")
	}
	if _, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{Capacity: 1, RefillRate: 0}); err == nil {
		t.Fatal("expected refill rate error")
	}

	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{Capacity: 1, RefillRate: 1})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.Allow(context.Background(), " "); err != entity.ErrEmptyKey {
		t.Fatalf("empty key error = %v, want ErrEmptyKey", err)
	}
}

func TestTokenBucketHonorsCanceledContext(t *testing.T) {
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{Capacity: 1, RefillRate: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Allow(ctx, "client-a"); err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestTokenBucketIsSafeForConcurrentUse(t *testing.T) {
	clock := newFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service, err := NewTokenBucketService(memory.NewBucketRepository(), TokenBucketConfig{
		Capacity:   50,
		RefillRate: 1,
		Clock:      clock,
	})
	if err != nil {
		t.Fatal(err)
	}

	var allowed int64
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decision, err := service.Allow(context.Background(), "client-a")
			if err != nil {
				t.Error(err)
				return
			}
			if decision.Allowed {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	if allowed != 50 {
		t.Fatalf("allowed = %d, want 50", allowed)
	}
}
