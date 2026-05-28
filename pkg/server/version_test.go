package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionIncludesFeatures(t *testing.T) {
	srv := testServerWithFeatures(t, "swarm_readonly,stack_deploy")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Version  string   `json:"version"`
		Service  string   `json:"service"`
		Features []string `json:"features"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Service != "stowkeep" {
		t.Fatalf("service = %q", body.Service)
	}
	if len(body.Features) != 2 {
		t.Fatalf("features = %v, want 2 entries", body.Features)
	}
	if body.Features[0] != "stack_deploy" || body.Features[1] != "swarm_readonly" {
		t.Fatalf("features = %v, want sorted [stack_deploy swarm_readonly]", body.Features)
	}
}
