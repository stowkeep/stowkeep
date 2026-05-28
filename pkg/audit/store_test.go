package audit_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/audit"
	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func TestAuditChainSQLite(t *testing.T) {
	testAuditChain(t, "sqlite", "")
}

func TestAuditChainPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres integration")
	}
	url := os.Getenv("STOWKEEP_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("STOWKEEP_TEST_DATABASE_URL not set")
	}
	testAuditChain(t, "postgres", url)
}

func testAuditChain(t *testing.T, driver, url string) {
	t.Helper()
	ctx := context.Background()
	var cfg *config.Config
	if driver == "postgres" {
		cfg = &config.Config{DatabaseURL: url}
	} else {
		dir := t.TempDir()
		cfg = &config.Config{DatabasePath: filepath.Join(dir, "audit.db")}
	}

	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	migrationsRoot := filepath.Join("..", "..", "migrations")
	if err := db.Up(database, cfg.ResolvedDriver(), migrationsRoot); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	store := audit.NewStore(database, cfg.ResolvedDriver())
	id1, err := store.Append(ctx, audit.Event{
		ActorID: "user-1", Action: "stack.deploy", ResourceType: "stack", ResourceID: "web",
		RequestID: "req-1", AfterHash: "abc123",
	})
	if err != nil {
		t.Fatalf("append 1: %v", err)
	}
	id2, err := store.Append(ctx, audit.Event{
		ActorID: "user-1", Action: "stack.remove", ResourceType: "stack", ResourceID: "web",
		RequestID: "req-2", BeforeHash: "abc123",
	})
	if err != nil {
		t.Fatalf("append 2: %v", err)
	}
	if id2 <= id1 {
		t.Fatalf("expected increasing ids")
	}

	res, err := audit.VerifyChain(ctx, store)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !res.OK {
		t.Fatalf("chain broken: %+v", res)
	}
}
