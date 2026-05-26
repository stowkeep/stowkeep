// Package db provides database connectivity for SQLite and PostgreSQL backends.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/stowkeep/stowkeep/pkg/config"
)

const (
	maxOpenConns    = 5
	maxIdleConns    = 2
	connMaxLifetime = 30 * time.Minute
	sqliteBusyRetry = 5
	sqliteBusyWait  = 50 * time.Millisecond
)

// Open connects to the database configured in cfg and applies driver-specific pragmas.
func Open(cfg *config.Config) (*sql.DB, error) {
	driver, dsn, err := resolveDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pingWithRetry(ctx, db, driver); err != nil {
		_ = db.Close()
		return nil, err
	}

	if driver == "sqlite" {
		if err := configureSQLite(ctx, db); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return db, nil
}

func resolveDSN(cfg *config.Config) (string, string, error) {
	if cfg.DatabaseURL != "" {
		lower := strings.ToLower(cfg.DatabaseURL)
		switch {
		case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"):
			return "pgx", cfg.DatabaseURL, nil
		case strings.HasPrefix(lower, "sqlite://"):
			path := cfg.DatabaseURL[len("sqlite://"):]
			return "sqlite", "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", nil
		default:
			return "", "", fmt.Errorf("unsupported database URL scheme in %q", cfg.DatabaseURL)
		}
	}

	driver := cfg.ResolvedDriver()
	switch driver {
	case "sqlite":
		path := cfg.ResolvedSQLitePath()
		return "sqlite", "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", nil
	case "postgres":
		return "", "", errors.New("postgres driver requires STOWKEEP_DATABASE_URL")
	default:
		return "", "", fmt.Errorf("unsupported database driver %q", driver)
	}
}

func pingWithRetry(ctx context.Context, db *sql.DB, driver string) error {
	var lastErr error
	attempts := 1
	if driver == "sqlite" {
		attempts = sqliteBusyRetry
	}
	for i := 0; i < attempts; i++ {
		if err := db.PingContext(ctx); err != nil {
			lastErr = err
			if driver == "sqlite" && strings.Contains(strings.ToLower(err.Error()), "locked") {
				time.Sleep(sqliteBusyWait)
				continue
			}
			return fmt.Errorf("ping database: %w", err)
		}
		return nil
	}
	return fmt.Errorf("ping database after retries: %w", lastErr)
}

func configureSQLite(ctx context.Context, db *sql.DB) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
	}
	for _, q := range pragmas {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("sqlite pragma %q: %w", q, err)
		}
	}
	return nil
}

// DriverName returns the driver name for goose migrations.
func DriverName(cfg *config.Config) string {
	return cfg.ResolvedDriver()
}

// Up applies all pending migrations for the given driver.
func Up(db *sql.DB, driver, migrationsRoot string) error {
	dir, err := migrationDir(migrationsRoot, driver)
	if err != nil {
		return err
	}
	if err := goose.SetDialect(driver); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func migrationDir(root, driver string) (string, error) {
	switch driver {
	case "sqlite":
		return filepath.Join(root, "sqlite"), nil
	case "postgres":
		return filepath.Join(root, "postgres"), nil
	default:
		return "", fmt.Errorf("unsupported migration driver %q", driver)
	}
}
