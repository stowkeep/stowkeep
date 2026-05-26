export type User = {
  id: string;
  email: string;
  role: string;
};

export type SwarmStatus = {
  connected: boolean;
  docker_host: string;
  swarm_active: boolean;
  node_role?: string;
  local_node_state?: string;
  error?: string;
};

export type Node = {
  id: string;
  hostname: string;
  status: string;
  availability: string;
  role: string;
  manager_lead: boolean;
  address?: string;
};

export type Service = {
  id: string;
  name: string;
  image: string;
  mode: string;
  replicas: string;
  update_status?: string;
  stack?: string;
  published_ports?: Array<{
    protocol: string;
    publish_mode: string;
    published_port: number;
    target_port: number;
  }>;
};

export type Task = {
  id: string;
  service_id: string;
  node_id: string;
  state: string;
  desired_state: string;
  slot?: number;
  exit_code?: number;
  error?: string;
};

export type StackSummary = {
  name: string;
  service_count: number;
};

export type StackDetail = {
  name: string;
  services: Service[];
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const api = {
  setupStatus: () => request<{ needs_bootstrap: boolean }>("/api/v1/setup/status"),
  setupAdmin: (email: string, password: string) =>
    request<User>("/api/v1/setup/admin", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  login: (email: string, password: string) =>
    request<User>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<{ status: string }>("/api/v1/auth/logout", { method: "POST" }),
  me: () => request<User>("/api/v1/auth/me"),
  swarmStatus: () => request<SwarmStatus>("/api/v1/swarm/status"),
  nodes: () => request<{ items: Node[] }>("/api/v1/swarm/nodes"),
  services: () => request<{ items: Service[] }>("/api/v1/swarm/services"),
  tasks: (serviceId?: string) => {
    const q = serviceId ? `?service_id=${encodeURIComponent(serviceId)}` : "";
    return request<{ items: Task[] }>(`/api/v1/swarm/tasks${q}`);
  },
  stacks: () => request<{ items: StackSummary[] }>("/api/v1/swarm/stacks"),
  stack: (name: string) => request<StackDetail>(`/api/v1/swarm/stacks/${encodeURIComponent(name)}`),
};
