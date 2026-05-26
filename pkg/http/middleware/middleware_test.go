package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/http/middleware"
)

func TestAccessLogOmitsAuthorizationHeader(t *testing.T) {
	const sentinel = "SENTINEL_AUTH_TOKEN_98765"
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := middleware.RequestID(middleware.AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Middleware must not log raw Authorization; verify sentinel stays out of logs.
		_ = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks?token=secret", nil)
	req.Header.Set("Authorization", "Bearer "+sentinel)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if strings.Contains(output, sentinel) {
		t.Fatalf("authorization sentinel leaked into logs: %s", output)
	}
	if strings.Contains(output, "secret") && strings.Contains(output, "token=secret") {
		t.Fatalf("query token leaked into logs: %s", output)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}
