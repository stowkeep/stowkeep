// Package backup provides database backup primitives for SQLite and PostgreSQL.
package backup

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// SQLiteBackup copies a live SQLite database to destPath using VACUUM INTO.
func SQLiteBackup(ctx context.Context, src *sql.DB, destPath string) error {
	if src == nil {
		return fmt.Errorf("database connection is required")
	}
	// destPath is an operator-controlled backup destination, not user SQL input.
	query := fmt.Sprintf("VACUUM INTO %q", destPath) // #nosec G201
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
	if w == nil {
		return fmt.Errorf("output writer is required")
	}
	if databaseURL == "" {
		return fmt.Errorf("database URL is required")
	}
	bin := d.PgDumpPath
	if bin == "" {
		bin = "pg_dump"
	}
	// pg_dump path and database URL come from server config, not request input.
	cmd := exec.CommandContext(ctx, bin, "--no-owner", "--no-acl", databaseURL) // #nosec G204
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("pg_dump: %w: %s", err, msg)
		}
		return fmt.Errorf("pg_dump: %w", err)
	}
	return nil
}
