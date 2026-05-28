package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	mobyclient "github.com/moby/moby/client"

	"github.com/stowkeep/stowkeep/pkg/compose"
)

// DeployStack deploys a Compose file as a Swarm stack.
func (c *Client) DeployStack(ctx context.Context, name string, content []byte, env map[string]string) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	project, err := compose.LoadProject(ctx, content, name)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}
	if len(env) > 0 {
		for k, v := range env {
			project.Environment[k] = v
		}
	}

	networkID, err := c.ensureStackNetwork(ctx, name)
	if err != nil {
		return err
	}

	for _, svc := range project.Services {
		if err := c.deployService(ctx, name, networkID, svc); err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "deployed stack",
		slog.String("component", "swarm"),
		slog.String("stack", name),
		slog.Int("services", len(project.Services)),
	)
	return nil
}

func (c *Client) ensureStackNetwork(ctx context.Context, stack string) (string, error) {
	netName := stack + "_default"
	filters := mobyclient.Filters{}.Add("name", netName)
	result, err := c.cli.NetworkList(ctx, mobyclient.NetworkListOptions{Filters: filters})
	if err != nil {
		return "", fmt.Errorf("list networks: %w", err)
	}
	for _, n := range result.Items {
		if n.Name == netName {
			return n.ID, nil
		}
	}

	create, err := c.cli.NetworkCreate(ctx, netName, mobyclient.NetworkCreateOptions{
		Labels: map[string]string{
			stackNamespaceLabel: stack,
		},
		Driver: "overlay",
		Attachable: true,
	})
	if err != nil {
		return "", fmt.Errorf("create network: %w", err)
	}
	return create.ID, nil
}

func (c *Client) deployService(ctx context.Context, stack, networkID string, svc types.ServiceConfig) error {
	serviceName := stack + "_" + svc.Name
	spec, err := serviceSpecFromCompose(stack, networkID, svc)
	if err != nil {
		return err
	}

	existing, err := c.findServiceByName(ctx, serviceName)
	if err != nil {
		return err
	}
	if existing != nil {
		_, err = c.cli.ServiceUpdate(ctx, existing.ID, mobyclient.ServiceUpdateOptions{
			Version: existing.Version,
			Spec:    spec,
		})
		if err != nil {
			return fmt.Errorf("update service %s: %w", serviceName, err)
		}
		return nil
	}

	_, err = c.cli.ServiceCreate(ctx, mobyclient.ServiceCreateOptions{Spec: spec})
	if err != nil {
		return fmt.Errorf("create service %s: %w", serviceName, err)
	}
	return nil
}

func (c *Client) findServiceByName(ctx context.Context, name string) (*swarm.Service, error) {
	filters := mobyclient.Filters{}.Add("name", name)
	result, err := c.cli.ServiceList(ctx, mobyclient.ServiceListOptions{Filters: filters})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	for i := range result.Items {
		if result.Items[i].Spec.Name == name {
			svc := result.Items[i]
			return &svc, nil
		}
	}
	return nil, nil
}

func serviceSpecFromCompose(stack, networkID string, svc types.ServiceConfig) (swarm.ServiceSpec, error) {
	if svc.Image == "" {
		return swarm.ServiceSpec{}, fmt.Errorf("service %q requires an image", svc.Name)
	}
	scale := svc.GetScale()
	if scale < 0 {
		return swarm.ServiceSpec{}, fmt.Errorf("service %q has negative replicas", svc.Name)
	}
	replicas := uint64(scale)
	env := environmentList(svc.Environment)

	spec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: stack + "_" + svc.Name,
			Labels: map[string]string{
				stackNamespaceLabel: stack,
				"com.docker.stack.image": svc.Image,
			},
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image: svc.Image,
				Env:   env,
			},
			RestartPolicy: &swarm.RestartPolicy{Condition: swarm.RestartPolicyConditionAny},
			Networks: []swarm.NetworkAttachmentConfig{
				{Target: networkID},
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{Replicas: &replicas},
		},
	}

	if ports, err := publishedPorts(svc.Ports); err != nil {
		return swarm.ServiceSpec{}, err
	} else if len(ports) > 0 {
		spec.EndpointSpec = &swarm.EndpointSpec{Ports: ports}
	}
	return spec, nil
}

func environmentList(env types.MappingWithEquals) []string {
	if len(env) == 0 {
		return nil
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		if v == nil {
			out = append(out, k)
		} else {
			out = append(out, k+"="+*v)
		}
	}
	return out
}

func publishedPorts(ports []types.ServicePortConfig) ([]swarm.PortConfig, error) {
	var out []swarm.PortConfig
	for _, p := range ports {
		if p.Target == 0 {
			continue
		}
		pub := parsePublishedPort(p.Published)
		proto := network.TCP
		if strings.EqualFold(p.Protocol, "udp") {
			proto = network.UDP
		}
		mode := swarm.PortConfigPublishModeIngress
		if strings.EqualFold(p.Mode, "host") {
			mode = swarm.PortConfigPublishModeHost
		}
		out = append(out, swarm.PortConfig{
			Protocol:      proto,
			TargetPort:    p.Target,
			PublishedPort: pub,
			PublishMode:   mode,
		})
	}
	return out, nil
}

func parsePublishedPort(s string) uint32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(n)
}

// RemoveStack removes all services and networks for a stack namespace.
func (c *Client) RemoveStack(ctx context.Context, name string) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	filters := mobyclient.Filters{}.Add("label", stackNamespaceLabel+"="+name)
	services, err := c.cli.ServiceList(ctx, mobyclient.ServiceListOptions{Filters: filters})
	if err != nil {
		return fmt.Errorf("list stack services: %w", err)
	}
	for _, svc := range services.Items {
		if _, err := c.cli.ServiceRemove(ctx, svc.ID, mobyclient.ServiceRemoveOptions{}); err != nil {
			return fmt.Errorf("remove service %s: %w", svc.Spec.Name, err)
		}
	}

	networks, err := c.cli.NetworkList(ctx, mobyclient.NetworkListOptions{Filters: filters})
	if err != nil {
		return fmt.Errorf("list stack networks: %w", err)
	}
	for _, net := range networks.Items {
		if _, err := c.cli.NetworkRemove(ctx, net.ID, mobyclient.NetworkRemoveOptions{}); err != nil {
			return fmt.Errorf("remove network %s: %w", net.Name, err)
		}
	}

	slog.InfoContext(ctx, "removed stack",
		slog.String("component", "swarm"),
		slog.String("stack", name),
	)
	return nil
}

// ScaleService updates replica count for a service.
func (c *Client) ScaleService(ctx context.Context, serviceID string, replicas uint64) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()

	inspect, err := c.cli.ServiceInspect(ctx, serviceID, mobyclient.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("inspect service: %w", err)
	}
	spec := inspect.Service.Spec
	if spec.Mode.Replicated == nil {
		spec.Mode.Replicated = &swarm.ReplicatedService{}
	}
	spec.Mode.Replicated.Replicas = &replicas

	_, err = c.cli.ServiceUpdate(ctx, serviceID, mobyclient.ServiceUpdateOptions{
		Version: inspect.Service.Version,
		Spec:    spec,
	})
	if err != nil {
		return fmt.Errorf("scale service: %w", err)
	}
	return nil
}
