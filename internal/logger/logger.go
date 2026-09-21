// package logger contains logger config for dropzone.
package logger

import (
	"io"
	"log/slog"
)

type Format string

const FormatJSON Format = "json"
const FormatText Format = "text"

// New creates a new logger object.
func New(output io.Writer, format Format) *slog.Logger {

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	var handler slog.Handler
	switch format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	case FormatText:
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}
