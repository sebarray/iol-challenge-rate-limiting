package httpx

import (
	"encoding/json"
	nethttp "net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w nethttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w nethttp.ResponseWriter, status int, code string) {
	WriteJSON(w, status, ErrorResponse{Error: code})
}
