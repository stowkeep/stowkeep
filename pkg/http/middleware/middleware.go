// Package middleware provides HTTP middleware for the Stowkeep API server.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	applog "github.com/stowkeep/stowkeep/pkg/observability/log"
)

// RequestID assigns a unique request_id to each request and adds it to context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := applog.WithRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// AccessLog logs one line per HTTP request without bodies or sensitive headers.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			status := wrapped.status
			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			ctx := r.Context()
			entry := applog.LoggerFromContext(ctx, logger).With(
				slog.String("component", "http"),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)

			if q := applog.ScrubQuery(r.URL.RawQuery); q != "" {
				entry = entry.With(slog.String("query", q))
			}

			entry.Log(ctx, level, "request completed")
		})
	}
}

// Logger returns the request-scoped logger stored on context, or base.
func Logger(ctx context.Context, base *slog.Logger) *slog.Logger {
	return applog.LoggerFromContext(ctx, base)
}
