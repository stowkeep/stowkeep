package auth

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func openTestHandler(t *testing.T) (*Store, *Handler) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "internal.db")
	cfg := &config.Config{DatabasePath: path}
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Up(database, "sqlite", filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := NewStore(database, "sqlite")
	handler := NewHandler(store, HandlerConfig{SessionIdleTTL: time.Hour})
	return store, handler
}

// ensure sql import used when building without all tests
var _ *sql.DB
