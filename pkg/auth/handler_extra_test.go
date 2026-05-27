package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/auth"
)

func TestBootstrapConflict(t *testing.T) {
	handler, _ := testHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first bootstrap = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("second bootstrap = %d", rec.Code)
	}
}

func TestMeRequiresAuth(t *testing.T) {
	handler, store := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(handler.Me)).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("me without auth = %d", rec.Code)
	}
}

func TestSetupStatusMethodNotAllowed(t *testing.T) {
	handler, _ := testHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/status", nil)
	rec := httptest.NewRecorder()
	handler.SetupStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestMeMethodNotAllowed(t *testing.T) {
	handler, _ := testHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.Me(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("me method = %d", rec.Code)
	}
}

func TestRequireAuthInvalidCookie(t *testing.T) {
	store := testStore(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "not-a-valid-session"})
	rec := httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLogoutMethodNotAllowed(t *testing.T) {
	handler, _ := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	handler.Logout(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("logout method = %d", rec.Code)
	}
}

func TestLoginInvalidBody(t *testing.T) {
	handler, _ := testHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("login bad body = %d", rec.Code)
	}
}

func TestMeWithAuthenticatedUser(t *testing.T) {
	handler, store := testHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	cookie := rec.Result().Cookies()[0]

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(handler.Me)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me = %d", rec.Code)
	}
}

func TestLogoutClearsSession(t *testing.T) {
	handler, store := testHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	cookie := rec.Result().Cookies()[0]

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	auth.RequireAuth(store)(http.HandlerFunc(handler.Logout)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("logout = %d", rec.Code)
	}
}

func TestAuthenticateUnknownUser(t *testing.T) {
	store := testStore(t)
	_, err := store.Authenticate(context.Background(), "nobody@example.com", "password123")
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("err = %v", err)
	}
}

func TestOptionalAuthWithValidSession(t *testing.T) {
	handler, store := testHandler(t)
	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, req)
	cookie := rec.Result().Cookies()[0]

	called := false
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	auth.OptionalAuth(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Email != "admin@example.com" {
			t.Fatal("expected user in context")
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("optional auth failed called=%v code=%d", called, rec.Code)
	}
}
