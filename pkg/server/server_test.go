package server_test

import (
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
	path := filepath.Join(dir, "test.db")
	cfg := &config.Config{
		DatabasePath: path,
		Version:      "test",
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return server.New(cfg, logger, database)
}
