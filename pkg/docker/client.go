package docker

import (
	"context"
	"fmt"
	"time"

	mobyclient "github.com/moby/moby/client"
)

const stackNamespaceLabel = "com.docker.stack.namespace"

// Client provides read-only Swarm operations against Docker Engine.
type Client struct {
	cli     *mobyclient.Client
	host    string
	timeout time.Duration
}

// New creates a Docker client for host (e.g. unix:///var/run/docker.sock).
// Connection errors are reported by Status and list methods, not at construction.
func New(host string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cli, err := mobyclient.New(mobyclient.WithHost(host))
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &Client{cli: cli, host: host, timeout: timeout}, nil
}

// Host returns the configured Docker host URL.
func (c *Client) Host() string {
	return c.host
}

// Close releases client resources.
func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < c.timeout {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.timeout)
}
