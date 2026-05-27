package docker

import (
	"errors"
	"testing"

	"github.com/moby/moby/api/types/swarm"
)

func TestMapServiceReplicated(t *testing.T) {
	running := uint64(2)
	desired := uint64(3)
	svc := swarm.Service{
		ID: "svc1",
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: "web_api",
				Labels: map[string]string{
					stackNamespaceLabel: "web",
				},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{Image: "nginx:alpine"},
			},
			Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &desired}},
		},
		ServiceStatus: &swarm.ServiceStatus{RunningTasks: running, DesiredTasks: desired},
		Endpoint: swarm.Endpoint{
			Ports: []swarm.PortConfig{{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: 8080,
				TargetPort:    80,
			}},
		},
	}
	got := mapService(svc)
	if got.Name != "web_api" || got.Stack != "web" || got.Replicas != "2/3" {
		t.Fatalf("service = %+v", got)
	}
	if len(got.PublishedPorts) != 1 || got.PublishedPorts[0].PublishedPort != 8080 {
		t.Fatalf("ports = %+v", got.PublishedPorts)
	}
}

func TestMapServiceNilContainerSpec(t *testing.T) {
	running := uint64(2)
	desired := uint64(3)
	svc := swarm.Service{
		ID: "svc1",
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name: "web_api",
				Labels: map[string]string{
					stackNamespaceLabel: "web",
				},
			},
			TaskTemplate: swarm.TaskSpec{},
			Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &desired}},
		},
		ServiceStatus: &swarm.ServiceStatus{RunningTasks: running, DesiredTasks: desired},
	}
	got := mapService(svc)
	if got.Name != "web_api" || got.Stack != "web" || got.Image != "" || got.Replicas != "2/3" {
		t.Fatalf("service = %+v", got)
	}
}

func TestGetStackNotFound(t *testing.T) {
	services := []Service{{ID: "svc1", Name: "web_api", Stack: "web"}}
	_, err := stackFromServices(services, "missing")
	if !errors.Is(err, ErrStackNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetStackFound(t *testing.T) {
	services := []Service{
		{ID: "svc1", Name: "web_api", Stack: "web"},
		{ID: "svc2", Name: "web_db", Stack: "web"},
		{ID: "svc3", Name: "other_api", Stack: "other"},
	}
	detail, err := stackFromServices(services, "web")
	if err != nil {
		t.Fatalf("stackFromServices: %v", err)
	}
	if detail.Name != "web" || len(detail.Services) != 2 {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestMapNodeManager(t *testing.T) {
	node := swarm.Node{
		ID: "node1",
		Description: swarm.NodeDescription{Hostname: "mgr1"},
		Spec: swarm.NodeSpec{
			Role:         swarm.NodeRoleManager,
			Availability: swarm.NodeAvailabilityActive,
		},
		Status: swarm.NodeStatus{State: swarm.NodeStateReady, Addr: "10.0.0.1"},
		ManagerStatus: &swarm.ManagerStatus{Leader: true},
	}
	got := mapNode(node)
	if got.Role != "manager" || !got.ManagerLead || got.Hostname != "mgr1" {
		t.Fatalf("node = %+v", got)
	}
}

func TestMapTask(t *testing.T) {
	task := swarm.Task{
		ID:        "task1",
		ServiceID: "svc1",
		NodeID:    "node1",
		Slot:      1,
		Status:    swarm.TaskStatus{State: swarm.TaskStateRunning},
		DesiredState: swarm.TaskStateRunning,
	}
	got := mapTask(task)
	if got.State != "running" || got.ServiceID != "svc1" {
		t.Fatalf("task = %+v", got)
	}
}
