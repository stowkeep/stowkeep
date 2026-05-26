package backup_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stowkeep/stowkeep/pkg/backup"
)

func TestSQLiteBackupCreatesFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "src.db")
	destPath := filepath.Join(dir, "backup.db")

	src, err := sql.Open("sqlite", "file:"+srcPath)
	if err != nil {
		t.Fatalf("open src: %v", err)
	}
	defer src.Close()

	if _, err := src.Exec(`CREATE TABLE IF NOT EXISTS health_check (id INTEGER PRIMARY KEY, ok TEXT NOT NULL)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := src.Exec(`INSERT INTO health_check (ok) VALUES ('yes')`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := backup.SQLiteBackup(context.Background(), src, destPath); err != nil {
		t.Fatalf("SQLiteBackup: %v", err)
	}

	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("backup file is empty")
	}
}

func TestPostgresDumperRequiresURL(t *testing.T) {
	dumper := backup.PostgresDumper{}
	err := dumper.Stream(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty database URL")
	}
}
