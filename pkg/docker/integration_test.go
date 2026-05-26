//go:build integration

package docker_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stowkeep/stowkeep/pkg/docker"
)

func TestDockerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration")
	}
	if os.Getenv("STOWKEEP_INTEGRATION_DOCKER") == "" {
		t.Skip("STOWKEEP_INTEGRATION_DOCKER not set")
	}

	host := os.Getenv("STOWKEEP_DOCKER_HOST")
	if host == "" {
		host = "unix:///var/run/docker.sock"
	}

	cli, err := docker.New(host, 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })

	ctx := context.Background()
	status := cli.Status(ctx)
	if !status.Connected {
		t.Fatalf("docker not connected: %+v", status)
	}

	if _, err := cli.ListNodes(ctx); err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if _, err := cli.ListServices(ctx); err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if _, err := cli.ListTasks(ctx, ""); err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if _, err := cli.ListStacks(ctx); err != nil {
		t.Fatalf("ListStacks: %v", err)
	}
}
