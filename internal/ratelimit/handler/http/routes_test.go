package http

import (
	"context"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"rate-limiter/internal/ratelimit/entity"
)

type countingLimiter struct {
	calls int
}

func (s *countingLimiter) Allow(_ context.Context, _ string) (entity.Decision, error) {
	s.calls++
	return entity.Decision{
		Allowed:   true,
		Limit:     10,
		Remaining: 9,
	}, nil
}

func TestRegisterRoutesExposesHealthzWithoutRateLimit(t *testing.T) {
	stub := &countingLimiter{}
	routes := NewHandler(stub, discardLogger()).Routes()

	req := httptest.NewRequest(nethttp.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	routes.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "ok\n" {
		t.Fatalf("body = %q, want ok", got)
	}
	if stub.calls != 0 {
		t.Fatalf("rate limit calls = %d, want 0", stub.calls)
	}
}

func TestRegisterRoutesProtectsRootWithRateLimit(t *testing.T) {
	stub := &countingLimiter{}
	routes := NewHandler(stub, discardLogger()).Routes()

	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	rec := httptest.NewRecorder()

	routes.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "request accepted from 192.0.2.10:54321\n" {
		t.Fatalf("body = %q, want accepted response", got)
	}
	if stub.calls != 1 {
		t.Fatalf("rate limit calls = %d, want 1", stub.calls)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
