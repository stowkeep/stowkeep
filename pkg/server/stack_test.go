package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStackValidateRequiresAuthAndFeature(t *testing.T) {
	srv := testServerWithFeatures(t, "stack_deploy")
	cookie := bootstrapSession(t, srv)

	body, _ := json.Marshal(map[string]string{
		"name":    "web",
		"compose": "services:\n  web:\n    image: nginx:alpine\n",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks/validate", bytes.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("validate = %d body=%s", rec.Code, rec.Body.String())
	}
	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["valid"] != true {
		t.Fatalf("expected valid compose: %+v", result)
	}
}

func TestStackDeployFeatureDisabled(t *testing.T) {
	srv := testServerWithFeatures(t, "swarm_readonly")
	cookie := bootstrapSession(t, srv)
	body, _ := json.Marshal(map[string]string{"name": "web", "compose": "services:\n  web:\n    image: nginx:alpine\n"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks", bytes.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("feature disabled = %d", rec.Code)
	}
}

func TestStackValidateInvalidCompose(t *testing.T) {
	srv := testServerWithFeatures(t, "stack_deploy")
	cookie := bootstrapSession(t, srv)
	body, _ := json.Marshal(map[string]string{
		"name":    "web",
		"compose": "services:\n  web:\n    image: nginx:alpine\n    ports:\n      - not-a-port\n",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks/validate", bytes.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("validate = %d", rec.Code)
	}
	var result map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if result["valid"] == true {
		t.Fatal("expected invalid compose")
	}
}

func TestStackEndpointsRequireAuth(t *testing.T) {
	srv := testServerWithFeatures(t, "stack_deploy")
	paths := []struct {
		method string
		path   string
		body   []byte
	}{
		{http.MethodPost, "/api/v1/stacks/validate", mustJSON(t, map[string]string{"name": "web", "compose": "services:\n  a:\n    image: nginx\n"})},
		{http.MethodDelete, "/api/v1/stacks/web", nil},
		{http.MethodPatch, "/api/v1/stacks/services/svc1/scale", mustJSON(t, map[string]uint64{"replicas": 1})},
		{http.MethodGet, "/api/v1/stacks/services/svc1/logs", nil},
	}
	for _, tc := range paths {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(tc.body))
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("unauthenticated = %d", rec.Code)
			}
		})
	}
}

func TestStackRemoveInvalidName(t *testing.T) {
	srv := testServerWithFeatures(t, "stack_deploy")
	cookie := bootstrapSession(t, srv)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/stacks/INVALID", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid name = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestStackValidateRequestBodyTooLarge(t *testing.T) {
	srv := testServerWithFeatures(t, "stack_deploy")
	cookie := bootstrapSession(t, srv)
	oversize := strings.Repeat("x", 1<<20+2048)
	body, _ := json.Marshal(map[string]string{"name": "web", "compose": oversize})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks/validate", bytes.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize body = %d", rec.Code)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
