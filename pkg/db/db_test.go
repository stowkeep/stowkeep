package db_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func TestMigrationsSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := &config.Config{
		DatabaseDriver: "sqlite",
		DatabasePath:   dbPath,
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	migrationsRoot := filepath.Join("..", "..", "migrations")
	if err := db.Up(database, "sqlite", migrationsRoot); err != nil {
		t.Fatalf("Up: %v", err)
	}

	assertTableExists(t, database, "audit_events")
	assertTableExists(t, database, "envelope_canary")
	assertTableExists(t, database, "users")
	assertTableExists(t, database, "sessions")
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var count int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&count)
	if err != nil {
		t.Fatalf("query table %s: %v", name, err)
	}
	if count != 1 {
		t.Fatalf("table %s not found", name)
	}
}

func TestOpenCreatesSQLiteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dev.db")
	cfg := &config.Config{DatabasePath: path}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file: %v", err)
	}
}

func TestOpenSQLiteURLCaseInsensitiveScheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "case.db")
	cfg := &config.Config{DatabaseURL: "SQLite://" + path}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file at %s: %v", path, err)
	}
}

func TestMigrationsPostgresFromEnv(t *testing.T) {
	url := os.Getenv("STOWKEEP_DATABASE_URL")
	if url == "" {
		t.Skip("STOWKEEP_DATABASE_URL not set")
	}

	cfg := &config.Config{DatabaseURL: url}
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	migrationsRoot := filepath.Join("..", "..", "migrations")
	if err := db.Up(database, "postgres", migrationsRoot); err != nil {
		t.Fatalf("Up: %v", err)
	}

	assertTableExistsPostgres(t, database, "audit_events")
	assertTableExistsPostgres(t, database, "envelope_canary")
	assertTableExistsPostgres(t, database, "users")
	assertTableExistsPostgres(t, database, "sessions")
}

func assertTableExistsPostgres(t *testing.T, database *sql.DB, name string) {
	t.Helper()
	var exists bool
	err := database.QueryRowContext(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`, name).Scan(&exists)
	if err != nil {
		t.Fatalf("query table %s: %v", name, err)
	}
	if !exists {
		t.Fatalf("table %s not found", name)
	}
}
