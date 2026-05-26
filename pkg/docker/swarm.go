package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	mobyclient "github.com/moby/moby/client"
)

// Status checks Engine connectivity and Swarm mode.
func (c *Client) Status(ctx context.Context) Status {
	out := Status{DockerHost: c.host}
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	if _, err := c.cli.Ping(ctx, mobyclient.PingOptions{}); err != nil {
		out.Error = "docker engine unreachable"
		slog.WarnContext(ctx, "docker ping failed",
			slog.String("component", "swarm"),
			slog.String("error", err.Error()),
		)
		return out
	}
	out.Connected = true

	info, err := c.cli.Info(ctx, mobyclient.InfoOptions{})
	if err != nil {
		out.Error = "failed to read docker info"
		slog.WarnContext(ctx, "docker info failed",
			slog.String("component", "swarm"),
			slog.String("error", err.Error()),
		)
		return out
	}

	swarmInfo := info.Info.Swarm
	out.LocalNodeState = string(swarmInfo.LocalNodeState)
	out.SwarmActive = swarmInfo.NodeID != "" && swarmInfo.LocalNodeState == swarm.LocalNodeStateActive
	if swarmInfo.ControlAvailable {
		out.NodeRole = "manager"
	} else if out.SwarmActive {
		out.NodeRole = "worker"
	}
	return out
}

// ListNodes returns Swarm nodes.
func (c *Client) ListNodes(ctx context.Context) ([]Node, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	result, err := c.cli.NodeList(ctx, mobyclient.NodeListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	slog.InfoContext(ctx, "listed swarm nodes",
		slog.String("component", "swarm"),
		slog.Int("count", len(result.Items)),
	)

	out := make([]Node, 0, len(result.Items))
	for _, n := range result.Items {
		out = append(out, mapNode(n))
	}
	return out, nil
}

// ListServices returns Swarm services.
func (c *Client) ListServices(ctx context.Context) ([]Service, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	result, err := c.cli.ServiceList(ctx, mobyclient.ServiceListOptions{Status: true})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	slog.InfoContext(ctx, "listed swarm services",
		slog.String("component", "swarm"),
		slog.Int("count", len(result.Items)),
	)

	out := make([]Service, 0, len(result.Items))
	for _, svc := range result.Items {
		out = append(out, mapService(svc))
	}
	return out, nil
}

// ListTasks returns Swarm tasks, optionally filtered by service ID.
func (c *Client) ListTasks(ctx context.Context, serviceID string) ([]Task, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	opts := mobyclient.TaskListOptions{}
	if serviceID != "" {
		opts.Filters = mobyclient.Filters{}.Add("service", serviceID)
	}
	result, err := c.cli.TaskList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	slog.InfoContext(ctx, "listed swarm tasks",
		slog.String("component", "swarm"),
		slog.Int("count", len(result.Items)),
	)

	out := make([]Task, 0, len(result.Items))
	for _, task := range result.Items {
		out = append(out, mapTask(task))
	}
	return out, nil
}

// ListStacks returns stack names derived from service labels.
func (c *Client) ListStacks(ctx context.Context) ([]StackSummary, error) {
	services, err := c.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, svc := range services {
		if svc.Stack == "" {
			continue
		}
		counts[svc.Stack]++
	}
	out := make([]StackSummary, 0, len(counts))
	for name, count := range counts {
		out = append(out, StackSummary{Name: name, ServiceCount: count})
	}
	return out, nil
}

// GetStack returns services belonging to a stack namespace.
func (c *Client) GetStack(ctx context.Context, name string) (*StackDetail, error) {
	services, err := c.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	var matched []Service
	for _, svc := range services {
		if svc.Stack == name {
			matched = append(matched, svc)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("stack %q not found", name)
	}
	return &StackDetail{Name: name, Services: matched}, nil
}

func mapNode(n swarm.Node) Node {
	status := "unknown"
	if n.Status.State != "" {
		status = string(n.Status.State)
	}
	availability := string(n.Spec.Availability)
	role := "worker"
	if n.Spec.Role == swarm.NodeRoleManager {
		role = "manager"
	}
	addr := ""
	if len(n.Status.Addr) > 0 {
		addr = n.Status.Addr
	}
	return Node{
		ID:           n.ID,
		Hostname:     n.Description.Hostname,
		Status:       status,
		Availability: availability,
		Role:         role,
		ManagerLead:  n.ManagerStatus != nil && n.ManagerStatus.Leader,
		Address:      addr,
		Labels:       copyLabels(n.Spec.Labels),
	}
}

func mapService(svc swarm.Service) Service {
	image := ""
	if len(svc.Spec.TaskTemplate.ContainerSpec.Image) > 0 {
		image = svc.Spec.TaskTemplate.ContainerSpec.Image
	}
	mode := "replicated"
	replicas := ""
	if svc.Spec.Mode.Replicated != nil {
		replicas = formatReplicas(svc.ServiceStatus)
	} else if svc.Spec.Mode.Global != nil {
		mode = "global"
		replicas = formatReplicas(svc.ServiceStatus)
	}
	updateStatus := ""
	if svc.UpdateStatus != nil {
		updateStatus = string(svc.UpdateStatus.State)
	}
	return Service{
		ID:             svc.ID,
		Name:           strings.TrimPrefix(svc.Spec.Name, "/"),
		Image:          image,
		Mode:           mode,
		Replicas:       replicas,
		UpdateStatus:   updateStatus,
		Stack:          svc.Spec.Labels[stackNamespaceLabel],
		PublishedPorts: mapPorts(svc.Endpoint.Ports),
	}
}

func formatReplicas(status *swarm.ServiceStatus) string {
	if status == nil {
		return "0/0"
	}
	return strconv.FormatUint(status.RunningTasks, 10) + "/" + strconv.FormatUint(status.DesiredTasks, 10)
}

func mapPorts(ports []swarm.PortConfig) []Port {
	out := make([]Port, 0, len(ports))
	for _, p := range ports {
		out = append(out, Port{
			Protocol:      string(p.Protocol),
			PublishMode:   string(p.PublishMode),
			PublishedPort: p.PublishedPort,
			TargetPort:    p.TargetPort,
		})
	}
	return out
}

func mapTask(task swarm.Task) Task {
	exitCode := 0
	if task.Status.ContainerStatus != nil {
		exitCode = int(task.Status.ContainerStatus.ExitCode)
	}
	errMsg := task.Status.Err
	return Task{
		ID:        task.ID,
		ServiceID: task.ServiceID,
		NodeID:    task.NodeID,
		State:     string(task.Status.State),
		Desired:   string(task.DesiredState),
		Slot:      task.Slot,
		ExitCode:  exitCode,
		Error:     errMsg,
	}
}

func copyLabels(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
