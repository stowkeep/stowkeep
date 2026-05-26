package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionCookies(t *testing.T) {
	rec := httptest.NewRecorder()
	exp := time.Now().UTC().Add(time.Hour)
	SetSessionCookie(rec, "token-value", exp, true)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != CookieName || !cookies[0].Secure {
		t.Fatalf("cookie = %+v", cookies)
	}

	rec = httptest.NewRecorder()
	ClearSessionCookie(rec, false)
	cleared := rec.Result().Cookies()[0]
	if cleared.MaxAge != -1 {
		t.Fatalf("expected cleared cookie, got %+v", cleared)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "abc"})
	token, ok := SessionTokenFromRequest(req)
	if !ok || token != "abc" {
		t.Fatalf("token = %q ok=%v", token, ok)
	}
}

func TestMeWithoutUserInContext(t *testing.T) {
	handler := NewHandler(&Store{}, HandlerConfig{})
	rec := httptest.NewRecorder()
	handler.Me(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSetupAdminMethodNotAllowed(t *testing.T) {
	handler := NewHandler(&Store{}, HandlerConfig{})
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestLogoutWithSessionCookie(t *testing.T) {
	store, handler := openTestHandler(t)
	body := []byte(`{"email":"admin@example.com","password":"password123"}`)
	rec := httptest.NewRecorder()
	handler.SetupAdmin(rec, httptest.NewRequest(http.MethodPost, "/setup/admin", bytes.NewReader(body)))
	cookie := rec.Result().Cookies()[0]

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	handler.Logout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("logout = %d", rec.Code)
	}
	_, err := store.GetUserBySessionToken(req.Context(), HashToken(cookie.Value))
	if err != ErrSessionNotFound {
		t.Fatalf("session still valid: %v", err)
	}
}
