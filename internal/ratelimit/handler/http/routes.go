package http

import (
	"fmt"
	nethttp "net/http"

	"rate-limiter/internal/platform/httpx"
)

func (h *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()
	h.RegisterRoutes(mux)
	return h.LoggingMiddleware(httpx.CORSMiddleware(mux))
}

func (h *Handler) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.Handle("GET /", h.Middleware(nethttp.HandlerFunc(h.accepted)))
}

func (h *Handler) healthz(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func (h *Handler) accepted(w nethttp.ResponseWriter, r *nethttp.Request) {
	_, _ = fmt.Fprintf(w, "request accepted from %s\n", r.RemoteAddr)
}
