package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rate-limiter/internal/ratelimit/entity"
)

type stubLimiter struct {
	decision entity.Decision
	err      error
	key      string
}

func (s *stubLimiter) Allow(_ context.Context, key string) (entity.Decision, error) {
	s.key = key
	return s.decision, s.err
}

func TestMiddlewareAllowsRequestAndWritesHeaders(t *testing.T) {
	stub := &stubLimiter{
		decision: entity.Decision{
			Allowed:    true,
			Limit:      10,
			Remaining:  9,
			ResetAfter: 200 * time.Millisecond,
		},
	}
	nextCalled := false
	next := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		nextCalled = true
		w.WriteHeader(nethttp.StatusCreated)
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	rec := httptest.NewRecorder()

	NewHandler(stub, nil).Middleware(next).ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if rec.Code != nethttp.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if stub.key != "192.0.2.10" {
		t.Fatalf("key = %q, want client IP", stub.key)
	}
	if got := rec.Header().Get("RateLimit-Limit"); got != "10" {
		t.Fatalf("RateLimit-Limit = %q, want 10", got)
	}
	if got := rec.Header().Get("RateLimit-Remaining"); got != "9" {
		t.Fatalf("RateLimit-Remaining = %q, want 9", got)
	}
	if got := rec.Header().Get("RateLimit-Reset"); got != "1" {
		t.Fatalf("RateLimit-Reset = %q, want 1", got)
	}
}

func TestMiddlewareRejectsWithRetryAfter(t *testing.T) {
	stub := &stubLimiter{
		decision: entity.Decision{
			Allowed:    false,
			Limit:      2,
			Remaining:  0,
			RetryAfter: 1500 * time.Millisecond,
			ResetAfter: 3 * time.Second,
		},
	}
	next := nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, _ *nethttp.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	rec := httptest.NewRecorder()

	NewHandler(stub, nil).Middleware(next).ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "2" {
		t.Fatalf("Retry-After = %q, want 2", got)
	}
	if got := rec.Body.String(); got != "{\"error\":\"rate_limited\"}\n" {
		t.Fatalf("body = %q, want rate_limited error", got)
	}
}

func TestMiddlewareUsesCustomKeyFunc(t *testing.T) {
	stub := &stubLimiter{
		decision: entity.Decision{Allowed: true, Limit: 1},
	}
	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "api-key-1")
	rec := httptest.NewRecorder()

	handler := NewHandler(stub, nil, WithKeyFunc(func(r *nethttp.Request) string {
		return r.Header.Get("X-API-Key")
	}))
	handler.Middleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if stub.key != "api-key-1" {
		t.Fatalf("key = %q, want custom key", stub.key)
	}
}

func TestMiddlewareReturnsBadRequestForEmptyKey(t *testing.T) {
	stub := &stubLimiter{err: entity.ErrEmptyKey}
	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewHandler(stub, nil).Middleware(nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, _ *nethttp.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"error\":\"invalid_rate_limit_key\"}\n" {
		t.Fatalf("body = %q, want invalid_rate_limit_key error", got)
	}
}

func TestMiddlewareReturnsInternalServerError(t *testing.T) {
	stub := &stubLimiter{err: errors.New("store unavailable")}
	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewHandler(stub, nil).Middleware(nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, _ *nethttp.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"error\":\"internal_error\"}\n" {
		t.Fatalf("body = %q, want internal_error", got)
	}
}
