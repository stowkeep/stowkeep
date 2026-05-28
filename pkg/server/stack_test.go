package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
