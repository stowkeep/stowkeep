package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/server"
)

func TestSwarmFeatureDisabledWhenAuthenticated(t *testing.T) {
	srv := testServerWithFeatures(t, "")
	cookie := bootstrapSession(t, srv)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/swarm/status", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("feature disabled = %d", rec.Code)
	}
}

func TestSwarmStatusRequiresAuthAndFeature(t *testing.T) {
	srv := testServerWithFeatures(t, "swarm_readonly")
	cookie := bootstrapSession(t, srv)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/swarm/status", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["docker_host"]; !ok {
		t.Fatalf("missing docker_host: %+v", body)
	}
}

func testServerWithFeatures(t *testing.T, features string) *server.Server {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		DatabasePath: dir + "/test.db",
		Version:      "test",
		Features:     features,
		DockerHost:   "unix:///var/run/docker.sock",
	}
	return newServer(t, cfg)
}
