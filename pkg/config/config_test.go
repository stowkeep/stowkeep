package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/config"
)

func TestLoadSQLiteDefaults(t *testing.T) {
	t.Setenv("STOWKEEP_DATABASE_URL", "")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")
	t.Setenv("STOWKEEP_DATABASE_PATH", "./.data/test.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ResolvedDriver() != "sqlite" {
		t.Fatalf("ResolvedDriver() = %q, want sqlite", cfg.ResolvedDriver())
	}
	if cfg.ResolvedSQLitePath() != "./.data/test.db" {
		t.Fatalf("ResolvedSQLitePath() = %q", cfg.ResolvedSQLitePath())
	}
}

func TestLoadPostgresRequiresURL(t *testing.T) {
	t.Setenv("STOWKEEP_DATABASE_URL", "")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "postgres")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when postgres driver set without URL")
	}
}

func TestLoadPostgresFromURL(t *testing.T) {
	t.Setenv("STOWKEEP_DATABASE_URL", "postgres://user:pass@localhost:5432/stowkeep?sslmode=disable")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ResolvedDriver() != "postgres" {
		t.Fatalf("ResolvedDriver() = %q, want postgres", cfg.ResolvedDriver())
	}
}

func TestLoadRejectsUnsupportedDatabaseURLScheme(t *testing.T) {
	t.Setenv("STOWKEEP_DATABASE_URL", "mysql://user:pass@localhost:3306/stowkeep")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for unsupported database URL scheme")
	}
}

func TestHasFeature(t *testing.T) {
	cfg := &config.Config{Features: "swarm_readonly,other"}
	if !cfg.HasFeature("swarm_readonly") {
		t.Fatal("expected swarm_readonly enabled")
	}
	if cfg.HasFeature("stack_deploy") {
		t.Fatal("expected stack_deploy disabled")
	}
}

func TestResolvedSQLitePathCaseInsensitiveScheme(t *testing.T) {
	t.Setenv("STOWKEEP_DATABASE_URL", "SQLite:///tmp/stowkeep.db")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ResolvedSQLitePath() != "/tmp/stowkeep.db" {
		t.Fatalf("ResolvedSQLitePath() = %q, want /tmp/stowkeep.db", cfg.ResolvedSQLitePath())
	}
}

func TestLoadFromDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("STOWKEEP_LOG_FORMAT=text\nSTOWKEEP_DATABASE_PATH=./.data/fromenv.db\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("STOWKEEP_DATABASE_URL", "")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LogFormat != "text" {
		t.Fatalf("LogFormat = %q, want text from .env", cfg.LogFormat)
	}
	if cfg.ResolvedSQLitePath() != "./.data/fromenv.db" {
		t.Fatalf("ResolvedSQLitePath() = %q", cfg.ResolvedSQLitePath())
	}
}

func TestLoadDotEnvDoesNotOverrideProcessEnv(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("STOWKEEP_LOG_FORMAT=text\nSTOWKEEP_DATABASE_PATH=./.data/fromenv.db\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("STOWKEEP_DATABASE_URL", "")
	t.Setenv("STOWKEEP_DATABASE_DRIVER", "")
	t.Setenv("STOWKEEP_LOG_FORMAT", "json")
	t.Setenv("STOWKEEP_DATABASE_PATH", "./.data/process.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LogFormat != "json" {
		t.Fatalf("LogFormat = %q, want json from process env", cfg.LogFormat)
	}
	if cfg.ResolvedSQLitePath() != "./.data/process.db" {
		t.Fatalf("ResolvedSQLitePath() = %q", cfg.ResolvedSQLitePath())
	}
}

func TestEnabledFeaturesSorted(t *testing.T) {
	cfg := &config.Config{Features: "stack_deploy,swarm_readonly"}
	got := cfg.EnabledFeatures()
	want := []string{"stack_deploy", "swarm_readonly"}
	if len(got) != len(want) {
		t.Fatalf("EnabledFeatures() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("EnabledFeatures() = %v, want %v", got, want)
		}
	}
}
