// Package log configures structured logging for Stowkeep using log/slog.
package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey struct{}

// New builds a slog.Logger with the given level, format, and service metadata.
func New(level, format, version string, addSource bool) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(level),
		AddSource: addSource,
	}
	var handler slog.Handler
	switch strings.ToLower(format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler).With(
		slog.String("service", "stowkeep"),
		slog.String("version", version),
	)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID returns a context carrying request_id for log correlation.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestID)
}

// RequestIDFromContext returns the request_id from ctx, or empty string.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey{}).(string); ok {
		return v
	}
	return ""
}

// LoggerFromContext returns a logger with request_id attached when present.
func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}
	if id := RequestIDFromContext(ctx); id != "" {
		return base.With(slog.String("request_id", id))
	}
	return base
}
