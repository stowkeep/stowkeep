/** Authenticated user returned by the API. */
export type User = {
  id: string;
  email: string;
  role: string;
};

/** Docker Engine and Swarm connectivity status. */
export type SwarmStatus = {
  connected: boolean;
  docker_host: string;
  swarm_active: boolean;
  node_role?: string;
  local_node_state?: string;
  error?: string;
};

/** Swarm node summary. */
export type Node = {
  id: string;
  hostname: string;
  status: string;
  availability: string;
  role: string;
  manager_lead: boolean;
  address?: string;
};

/** Swarm service summary. */
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

/** Swarm task summary. */
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

/** Deployed stack name with service count. */
export type StackSummary = {
  name: string;
  service_count: number;
};

/** Stack detail with member services. */
export type StackDetail = {
  name: string;
  services: Service[];
};

/** Compose validation error from the API. */
export type ComposeValidationError = {
  path: string;
  message: string;
};

/** Compose validation result. */
export type ComposeValidationResult = {
  valid: boolean;
  errors?: ComposeValidationError[];
  hash?: string;
};

/** HTTP error thrown by the API client. */
export class ApiError extends Error {
  /** HTTP status code from the failed response. */
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

/** Typed HTTP client for Stowkeep API endpoints. */
export const api = {
  /** Returns whether first-run admin bootstrap is required. */
  setupStatus: () => request<{ needs_bootstrap: boolean }>("/api/v1/setup/status"),
  /** Creates the initial admin account during bootstrap. */
  setupAdmin: (email: string, password: string) =>
    request<User>("/api/v1/setup/admin", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  /** Authenticates and establishes a session cookie. */
  login: (email: string, password: string) =>
    request<User>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  /** Ends the current session. */
  logout: () => request<{ status: string }>("/api/v1/auth/logout", { method: "POST" }),
  /** Returns the currently authenticated user. */
  me: () => request<User>("/api/v1/auth/me"),
  /** Returns Docker/Swarm connectivity status. */
  swarmStatus: () => request<SwarmStatus>("/api/v1/swarm/status"),
  /** Lists Swarm nodes. */
  nodes: () => request<{ items: Node[] }>("/api/v1/swarm/nodes"),
  /** Lists Swarm services. */
  services: () => request<{ items: Service[] }>("/api/v1/swarm/services"),
  /** Lists Swarm tasks, optionally filtered by service ID. */
  tasks: (serviceId?: string) => {
    const q = serviceId ? `?service_id=${encodeURIComponent(serviceId)}` : "";
    return request<{ items: Task[] }>(`/api/v1/swarm/tasks${q}`);
  },
  /** Lists deployed stacks. */
  stacks: () => request<{ items: StackSummary[] }>("/api/v1/swarm/stacks"),
  /** Returns services belonging to a stack. */
  stack: (name: string) => request<StackDetail>(`/api/v1/swarm/stacks/${encodeURIComponent(name)}`),
  /** Validates compose YAML for a stack name. */
  validateStack: (name: string, compose: string) =>
    request<ComposeValidationResult>("/api/v1/stacks/validate", {
      method: "POST",
      body: JSON.stringify({ name, compose }),
    }),
  /** Deploys a stack from compose YAML. */
  deployStack: (name: string, compose: string, env?: Record<string, string>) =>
    request<StackDetail>("/api/v1/stacks", {
      method: "POST",
      body: JSON.stringify({ name, compose, env }),
    }),
  /** Removes a deployed stack. */
  removeStack: (name: string) =>
    request<{ status: string }>(`/api/v1/stacks/${encodeURIComponent(name)}`, { method: "DELETE" }),
  /** Scales a service replica count. */
  scaleService: (serviceId: string, replicas: number) =>
    request<{ service_id: string; replicas: number }>(
      `/api/v1/stacks/services/${encodeURIComponent(serviceId)}/scale`,
      { method: "PATCH", body: JSON.stringify({ replicas }) },
    ),
  /** Fetches recent service logs as plain text. */
  serviceLogs: async (serviceId: string, opts?: { tail?: string; follow?: boolean }) => {
    const params = new URLSearchParams();
    if (opts?.tail) params.set("tail", opts.tail);
    if (opts?.follow) params.set("follow", "true");
    const q = params.toString();
    const res = await fetch(`/api/v1/stacks/services/${encodeURIComponent(serviceId)}/logs${q ? `?${q}` : ""}`, {
      credentials: "include",
    });
    if (!res.ok) {
      throw new ApiError(res.status, res.statusText);
    }
    return res.text();
  },
};
