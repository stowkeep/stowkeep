package log_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	applog "github.com/stowkeep/stowkeep/pkg/observability/log"
)

func TestScrubQuery(t *testing.T) {
	got := applog.ScrubQuery("foo=bar&token=secret&key=abc")
	if strings.Contains(got, "secret") {
		t.Fatalf("token value leaked: %q", got)
	}
	if !strings.Contains(got, "token=[REDACTED]") {
		t.Fatalf("expected redacted token, got %q", got)
	}
}

func TestScrubHeaderValue(t *testing.T) {
	if got := applog.ScrubHeaderValue("Authorization", "Bearer secret"); got != "[REDACTED]" {
		t.Fatalf("Authorization not redacted: %q", got)
	}
	if got := applog.ScrubHeaderValue("Content-Type", "application/json"); got != "application/json" {
		t.Fatalf("Content-Type incorrectly redacted: %q", got)
	}
}

func TestSentinelValueNeverInLogs(t *testing.T) {
	const sentinel = "SENTINEL_SECRET_VALUE_XYZ_12345"
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("test message",
		slog.String("component", "http"),
		slog.String("safe_field", "ok"),
	)

	// Simulate middleware scrubbing: never log Authorization with raw value.
	auth := applog.ScrubHeaderValue("Authorization", "Bearer "+sentinel)
	logger.Info("request completed",
		slog.String("component", "http"),
		slog.String("authorization", auth),
	)

	output := buf.String()
	if strings.Contains(output, sentinel) {
		t.Fatalf("sentinel value appeared in log output: %s", output)
	}
}
