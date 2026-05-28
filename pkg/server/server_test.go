package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
	"github.com/stowkeep/stowkeep/pkg/server"
)

func TestHealthz(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q", body["status"])
	}
}

func TestListenAndServeLogsEnabledFeatureFlags(t *testing.T) {
	var logs bytes.Buffer
	cfg := testConfig(t)
	cfg.HTTPAddr = "invalid-address"
	cfg.Features = "swarm_readonly,stack_deploy"
	srv := newServerWithLogger(t, cfg, slog.New(slog.NewJSONHandler(&logs, nil)))

	if err := srv.ListenAndServe(); err == nil {
		t.Fatal("expected ListenAndServe error for invalid address")
	}

	entry := findLogEntry(t, &logs, "feature flags enabled")
	features, ok := entry["features"].([]any)
	if !ok {
		t.Fatalf("features field = %T, want array", entry["features"])
	}
	want := []string{"stack_deploy", "swarm_readonly"}
	if len(features) != len(want) {
		t.Fatalf("features = %v, want %v", features, want)
	}
	for i, name := range want {
		if features[i] != name {
			t.Fatalf("features = %v, want %v", features, want)
		}
	}
}

func testServer(t *testing.T) *server.Server {
	t.Helper()
	return newServer(t, testConfig(t))
}

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		DatabasePath: filepath.Join(dir, "test.db"),
		Version:      "test",
		DockerHost:   "unix:///var/run/docker.sock",
	}
}

func newServer(t *testing.T, cfg *config.Config) *server.Server {
	t.Helper()
	return newServerWithLogger(t, cfg, slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

func newServerWithLogger(t *testing.T, cfg *config.Config, logger *slog.Logger) *server.Server {
	t.Helper()
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	if err := db.Up(database, "sqlite", filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return server.New(cfg, logger, database)
}

func findLogEntry(t *testing.T, r io.Reader, msg string) map[string]any {
	t.Helper()
	dec := json.NewDecoder(r)
	for dec.More() {
		var entry map[string]any
		if err := dec.Decode(&entry); err != nil {
			t.Fatalf("decode log entry: %v", err)
		}
		if entry["msg"] == msg {
			return entry
		}
	}
	t.Fatalf("log entry %q not found", msg)
	return nil
}

func bootstrapSession(t *testing.T, srv *server.Server) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("bootstrap = %d body=%s", rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()[0]
}
