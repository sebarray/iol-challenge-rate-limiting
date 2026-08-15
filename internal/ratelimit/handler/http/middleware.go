package http

import (
	"errors"
	"math"
	"net"
	nethttp "net/http"
	"strconv"
	"time"

	"rate-limiter/internal/platform/httpx"
	"rate-limiter/internal/ratelimit/entity"
)

func (h *Handler) Middleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		decision, err := h.limiter.Allow(r.Context(), h.keyFunc(r))
		if err != nil {
			status := nethttp.StatusInternalServerError
			code := "internal_error"
			if errors.Is(err, entity.ErrEmptyKey) {
				status = nethttp.StatusBadRequest
				code = "invalid_rate_limit_key"
			}

			httpx.WriteError(w, status, code)
			return
		}

		writeRateLimitHeaders(w.Header(), decision)
		if !decision.Allowed {
			httpx.WriteError(w, nethttp.StatusTooManyRequests, "rate_limited")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func DefaultKeyFunc(r *nethttp.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func writeRateLimitHeaders(header nethttp.Header, decision entity.Decision) {
	header.Set("RateLimit-Limit", strconv.Itoa(decision.Limit))
	header.Set("RateLimit-Remaining", strconv.Itoa(decision.Remaining))
	header.Set("RateLimit-Reset", durationSeconds(decision.ResetAfter))
	if !decision.Allowed {
		header.Set("Retry-After", durationSeconds(decision.RetryAfter))
	}
}

func durationSeconds(d time.Duration) string {
	if d <= 0 {
		return "0"
	}
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}
