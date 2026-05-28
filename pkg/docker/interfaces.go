package docker

import (
	"context"
	"io"
)

// SwarmReader provides read-only Swarm operations for HTTP handlers.
type SwarmReader interface {
	Host() string
	Status(ctx context.Context) Status
	ListNodes(ctx context.Context) ([]Node, error)
	ListServices(ctx context.Context) ([]Service, error)
	ListTasks(ctx context.Context, serviceID string) ([]Task, error)
	ListStacks(ctx context.Context) ([]StackSummary, error)
	GetStack(ctx context.Context, name string) (*StackDetail, error)
}

// SwarmWriter provides mutating Swarm operations for stack lifecycle.
type SwarmWriter interface {
	SwarmReader
	DeployStack(ctx context.Context, name string, compose []byte, env map[string]string) error
	RemoveStack(ctx context.Context, name string) error
	ScaleService(ctx context.Context, serviceID string, replicas uint64) error
	ServiceLogs(ctx context.Context, serviceID string, follow bool, tail string) (io.ReadCloser, error)
}

var (
	_ SwarmReader = (*Client)(nil)
	_ SwarmWriter = (*Client)(nil)
)
