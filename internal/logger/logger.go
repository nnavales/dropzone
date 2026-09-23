// package logger contains logger config for dropzone.
package logger

import (
	"io"
	"log/slog"
)

// New creates a new logger.
func New(output io.Writer, level slog.Level, addSource bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	return slog.New(slog.NewTextHandler(output, opts))
}
