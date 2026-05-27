package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stowkeep/stowkeep/pkg/auth"
	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
)

func testHandler(t *testing.T) (*auth.Handler, *auth.Store) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "handler.db")
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
	if err := db.Up(database, "sqlite", filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := auth.NewStore(database, "sqlite")
	handler := auth.NewHandler(store, auth.HandlerConfig{
		SessionIdleTTL: time.Hour,
		CookieSecure:   false,
	})
	return handler, store
}

func TestHandlerSetupAndLoginFlow(t *testing.T) {
	handler, store := testHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	handler.SetupStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup status = %d", rec.Code)
	}
	var status struct {
		NeedsBootstrap bool `json:"needs_bootstrap"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !status.NeedsBootstrap {
		t.Fatal("expected needs bootstrap")
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup admin = %d body=%s", rec.Code, rec.Body.String())
	}

	cookie := rec.Result().Cookies()[0]
	if cookie.Name != auth.CookieName {
		t.Fatalf("cookie name = %q", cookie.Name)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(handler.Me)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	handler.Logout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("logout = %d", rec.Code)
	}

	body, _ = json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthRejectsMissingSession(t *testing.T) {
	store := testStore(t)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	handler, store := testHandler(t)
	_, err := store.CreateBootstrapAdmin(context.Background(), "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"email": "admin@example.com", "password": "wrong"})
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()
		handler.Login(rec, req)
		if i < 10 && rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d", i, rec.Code)
		}
		if i == 10 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected rate limit, got %d", rec.Code)
		}
	}
}

func TestHandlerValidationErrors(t *testing.T) {
	handler, _ := testHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("setup bad body = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec = httptest.NewRecorder()
	handler.Login(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("login method = %d", rec.Code)
	}

	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "short"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("short password = %d", rec.Code)
	}
}

func TestCreateBootstrapValidation(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	_, err := store.CreateBootstrapAdmin(ctx, "", "password123")
	if err == nil {
		t.Fatal("expected email required error")
	}
	_, err = store.CreateBootstrapAdmin(ctx, "a@b.com", "short")
	if err == nil {
		t.Fatal("expected password length error")
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	store := testStore(t)
	user, hash, err := store.GetUserByEmail(context.Background(), "missing@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user != nil || hash != "" {
		t.Fatal("expected nil user")
	}
}

func TestNeedsBootstrapFalseAfterUser(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	if _, err := store.CreateBootstrapAdmin(ctx, "a@b.com", "password123"); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	needs, err := store.NeedsBootstrap(ctx)
	if err != nil {
		t.Fatalf("NeedsBootstrap: %v", err)
	}
	if needs {
		t.Fatal("expected bootstrap complete")
	}
}

func TestExpiredSessionRejected(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()
	user, err := store.CreateBootstrapAdmin(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	expired := time.Now().UTC().Add(-time.Hour)
	if err := store.CreateSession(ctx, user.ID, hash, expired); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_, err = store.GetUserBySessionToken(ctx, hash)
	if err != auth.ErrSessionNotFound {
		t.Fatalf("expected expired session error, got %v", err)
	}
	_ = plain
}
