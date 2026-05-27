package auth_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stowkeep/stowkeep/pkg/auth"
	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func testStore(t *testing.T) *auth.Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.db")
	cfg := &config.Config{DatabasePath: path}
	database, err := db.Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})
	migrationsRoot := filepath.Join("..", "..", "migrations")
	if err := db.Up(database, "sqlite", migrationsRoot); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return auth.NewStore(database, "sqlite")
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct horse battery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	ok, err := auth.VerifyPassword(hash, "correct horse battery")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
	ok, err = auth.VerifyPassword(hash, "wrong")
	if err != nil {
		t.Fatalf("VerifyPassword wrong: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail")
	}
}

func TestBootstrapAndAuthenticate(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	needs, err := store.NeedsBootstrap(ctx)
	if err != nil {
		t.Fatalf("NeedsBootstrap: %v", err)
	}
	if !needs {
		t.Fatal("expected bootstrap required")
	}

	user, err := store.CreateBootstrapAdmin(ctx, "Admin@Example.com", "password123")
	if err != nil {
		t.Fatalf("CreateBootstrapAdmin: %v", err)
	}
	if user.Email != "admin@example.com" {
		t.Fatalf("email = %q", user.Email)
	}

	_, err = store.CreateBootstrapAdmin(ctx, "other@example.com", "password123")
	if err != auth.ErrBootstrapComplete {
		t.Fatalf("expected ErrBootstrapComplete, got %v", err)
	}

	authUser, err := store.Authenticate(ctx, "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if authUser.ID != user.ID {
		t.Fatalf("user id mismatch")
	}

	_, err = store.Authenticate(ctx, "admin@example.com", "bad-password")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	user, err := store.CreateBootstrapAdmin(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("NewSessionToken: %v", err)
	}
	if auth.HashToken(plain) != hash {
		t.Fatal("token hash mismatch")
	}

	expires := user.CreatedAt.Add(24 * time.Hour)
	if err := store.CreateSession(ctx, user.ID, hash, expires); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := store.GetUserBySessionToken(ctx, hash)
	if err != nil {
		t.Fatalf("GetUserBySessionToken: %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("session user id = %q", got.ID)
	}

	if err := store.DeleteSession(ctx, hash); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	_, err = store.GetUserBySessionToken(ctx, hash)
	if err != auth.ErrSessionNotFound {
		t.Fatalf("expected session not found, got %v", err)
	}
}
