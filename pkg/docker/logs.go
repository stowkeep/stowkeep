package docker

import (
	"context"
	"fmt"
	"io"

	mobyclient "github.com/moby/moby/client"
)

// ServiceLogs streams logs for a Swarm service.
func (c *Client) ServiceLogs(ctx context.Context, serviceID string, follow bool, tail string) (io.ReadCloser, error) {
	if tail == "" {
		tail = "100"
	}
	result, err := c.cli.ServiceLogs(ctx, serviceID, mobyclient.ServiceLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	})
	if err != nil {
		return nil, fmt.Errorf("service logs: %w", err)
	}
	return result, nil
}
