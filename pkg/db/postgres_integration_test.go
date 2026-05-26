//go:build integration

package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func TestMigrationsPostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("stowkeep"),
		postgres.WithUsername("stowkeep"),
		postgres.WithPassword("stowkeep"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	cfg := &config.Config{DatabaseURL: connStr}
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	migrationsRoot := filepath.Join("..", "..", "migrations")
	if err := db.Up(database, "postgres", migrationsRoot); err != nil {
		t.Fatalf("Up: %v", err)
	}

	assertPostgresTableExists(t, database, "audit_events")
	assertPostgresTableExists(t, database, "envelope_canary")
}

func assertPostgresTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var exists bool
	err := db.QueryRowContext(context.Background(),
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
