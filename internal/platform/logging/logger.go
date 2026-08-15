package logging

import (
	"io"
	"log/slog"
)

func NewJSONLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	}))
}
