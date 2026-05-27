package docker_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stowkeep/stowkeep/pkg/docker"
)

func TestListNodesDoesNotLogResponseBodies(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	host := "unix:///var/run/docker.sock"
	cli, err := docker.New(host, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })

	_, _ = cli.ListNodes(context.Background())
	out := buf.String()
	if len(out) > 0 && (contains(out, "ContainerSpec") || contains(out, `"ID"`)) {
		t.Fatalf("unexpected verbose docker payload in logs: %s", out)
	}
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
