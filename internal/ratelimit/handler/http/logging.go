package http

import (
	nethttp "net/http"
	"time"
)

func (h *Handler) LoggingMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: nethttp.StatusOK}

		next.ServeHTTP(recorder, r)

		h.logger.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	nethttp.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
