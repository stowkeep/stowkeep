// Package backup provides database backup primitives for SQLite and PostgreSQL.
package backup

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// SQLiteBackup copies a live SQLite database to destPath using VACUUM INTO.
func SQLiteBackup(ctx context.Context, src *sql.DB, destPath string) error {
	query := fmt.Sprintf("VACUUM INTO %q", destPath)
	if _, err := src.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("sqlite backup: %w", err)
	}
	return nil
}

// PostgresDumper streams a logical backup from PostgreSQL using pg_dump.
type PostgresDumper struct {
	PgDumpPath string
}

// Stream runs pg_dump and writes output to w.
func (d PostgresDumper) Stream(ctx context.Context, databaseURL string, w io.Writer) error {
	if databaseURL == "" {
		return fmt.Errorf("database URL is required")
	}
	bin := d.PgDumpPath
	if bin == "" {
		bin = "pg_dump"
	}
	cmd := exec.CommandContext(ctx, bin, "--no-owner", "--no-acl", databaseURL)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump: %w", err)
	}
	return nil
}
