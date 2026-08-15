package httpx

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareSetsHeadersOnGet(t *testing.T) {
	next := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	CORSMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods should be set")
	}
}

func TestCORSMiddlewareHandlesOptionsPreflight(t *testing.T) {
	nextCalled := false
	next := nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, _ *nethttp.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(nethttp.MethodOptions, "/", nil)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	rec := httptest.NewRecorder()

	CORSMiddleware(next).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("next handler should not be called for OPTIONS")
	}
	if rec.Code != nethttp.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "X-Custom-Header" {
		t.Fatalf("Access-Control-Allow-Headers = %q, want X-Custom-Header", got)
	}
}
