package server_test

import (
	"bytes"
	"encoding/json"
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

func testServer(t *testing.T) *server.Server {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		DatabasePath: filepath.Join(dir, "test.db"),
		Version:      "test",
		DockerHost:   "unix:///var/run/docker.sock",
	}
	return newServer(t, cfg)
}

func newServer(t *testing.T, cfg *config.Config) *server.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
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
