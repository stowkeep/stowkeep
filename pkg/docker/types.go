package docker

// Status describes Docker Engine connectivity and Swarm state.
type Status struct {
	Connected      bool   `json:"connected"`
	DockerHost     string `json:"docker_host"`
	SwarmActive    bool   `json:"swarm_active"`
	NodeRole       string `json:"node_role,omitempty"`
	LocalNodeState string `json:"local_node_state,omitempty"`
	Error          string `json:"error,omitempty"`
}

// Node is a Swarm node summary for the API.
type Node struct {
	ID           string            `json:"id"`
	Hostname     string            `json:"hostname"`
	Status       string            `json:"status"`
	Availability string            `json:"availability"`
	Role         string            `json:"role"`
	ManagerLead  bool              `json:"manager_lead"`
	Address      string            `json:"address,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// Service is a Swarm service summary.
type Service struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Image          string `json:"image"`
	Mode           string `json:"mode"`
	Replicas       string `json:"replicas"`
	UpdateStatus   string `json:"update_status,omitempty"`
	Stack          string `json:"stack,omitempty"`
	PublishedPorts []Port `json:"published_ports,omitempty"`
}

// Port is a published port mapping.
type Port struct {
	Protocol      string `json:"protocol"`
	PublishMode   string `json:"publish_mode"`
	PublishedPort uint32 `json:"published_port"`
	TargetPort    uint32 `json:"target_port"`
}

// Task is a Swarm task summary.
type Task struct {
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	NodeID    string `json:"node_id"`
	State     string `json:"state"`
	Desired   string `json:"desired_state"`
	Slot      int    `json:"slot,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

// StackSummary is a deployed stack name with service count.
type StackSummary struct {
	Name         string `json:"name"`
	ServiceCount int    `json:"service_count"`
}

// StackDetail is a stack with its services.
type StackDetail struct {
	Name     string    `json:"name"`
	Services []Service `json:"services"`
}
