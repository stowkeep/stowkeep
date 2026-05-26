package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthBootstrapAndMe(t *testing.T) {
	srv := testServer(t)

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	statusRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("setup status = %d", statusRec.Code)
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "password123",
	})
	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/setup/admin", bytes.NewReader(body))
	setupRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(setupRec, setupReq)
	if setupRec.Code != http.StatusCreated {
		t.Fatalf("setup admin = %d body=%s", setupRec.Code, setupRec.Body.String())
	}
	cookie := setupRec.Result().Cookies()[0]

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.AddCookie(cookie)
	meRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me = %d", meRec.Code)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(cookie)
	logoutRec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout = %d", logoutRec.Code)
	}

	meReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq2.AddCookie(cookie)
	meRec2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(meRec2, meReq2)
	if meRec2.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout = %d", meRec2.Code)
	}
}

func TestSwarmRoutesRequireAuth(t *testing.T) {
	srv := testServerWithFeatures(t, "swarm_readonly")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/swarm/nodes", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated swarm = %d", rec.Code)
	}
}
