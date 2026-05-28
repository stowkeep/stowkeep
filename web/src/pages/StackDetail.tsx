import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, ApiError } from "../api/client";
import { DataTable, PageHeader } from "../components/DataTable";
import { Button } from "../components/ui/primitives";
import { useDockerStatus } from "../hooks/useDockerStatus";

/** Stack detail with services, replicas, and published ports. */
export default function StackDetailPage() {
  const { name = "" } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: dockerStatus } = useDockerStatus();
  const disabled = dockerStatus != null && !dockerStatus.connected;

  const { data, isLoading, error } = useQuery({
    queryKey: ["swarm", "stack", name],
    queryFn: () => api.stack(name),
    enabled: Boolean(name),
    refetchInterval: 8000,
  });

  const removeStack = useMutation({
    mutationFn: () => api.removeStack(name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["swarm", "stacks"] });
      void navigate("/stacks");
    },
  });

  const scaleService = useMutation({
    mutationFn: ({ serviceId, replicas }: { serviceId: string; replicas: number }) =>
      api.scaleService(serviceId, replicas),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["swarm", "stack", name] });
    },
  });

  function handleRemove() {
    if (!window.confirm(`Remove stack "${name}" and all its services? This cannot be undone.`)) {
      return;
    }
    void removeStack.mutateAsync();
  }

  function handleScale(serviceId: string, current: string) {
    const next = window.prompt("New replica count:", current.split("/")[0] || "1");
    if (next == null) return;
    const replicas = Number.parseInt(next, 10);
    if (Number.isNaN(replicas) || replicas < 0) return;
    void scaleService.mutateAsync({ serviceId, replicas });
  }

  return (
    <>
      <div className="mb-4 flex items-center justify-between gap-4">
        <Link className="text-sm text-slate-600 hover:text-slate-900" to="/stacks">
          ← Back to stacks
        </Link>
        <Button variant="secondary" disabled={disabled || removeStack.isPending} onClick={handleRemove}>
          {removeStack.isPending ? "Removing…" : "Remove stack"}
        </Button>
      </div>
      <PageHeader title={name} description="Services in this stack." />
      {removeStack.error instanceof ApiError && (
        <p className="mb-4 text-sm text-red-700">{removeStack.error.message}</p>
      )}
      <DataTable
        loading={isLoading}
        error={error instanceof Error ? error.message : null}
        emptyMessage="No services in this stack."
        columns={[
          { header: "Service", cell: (s) => s.name },
          { header: "Image", cell: (s) => s.image },
          { header: "Replicas", cell: (s) => s.replicas },
          {
            header: "Ports",
            cell: (s) =>
              s.published_ports?.length
                ? s.published_ports.map((p) => `${p.published_port}:${p.target_port}/${p.protocol}`).join(", ")
                : "—",
          },
          {
            header: "Actions",
            cell: (s) => (
              <div className="flex gap-2">
                <Button variant="ghost" disabled={disabled} onClick={() => handleScale(s.id, s.replicas)}>
                  Scale
                </Button>
                <Link
                  className="inline-flex items-center rounded-md px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                  to={`/services/${encodeURIComponent(s.id)}/logs`}
                >
                  Logs
                </Link>
              </div>
            ),
          },
        ]}
        rows={data?.services ?? []}
        rowKey={(s) => s.id}
      />
    </>
  );
}

/** Connection settings (read-only for Stage 1). */
export function SettingsPage() {
  const { data, isLoading, refetch, isFetching } = useQuery({
    queryKey: ["swarm", "status"],
    queryFn: api.swarmStatus,
  });

  return (
    <>
      <PageHeader title="Settings" description="Docker connection (configured via environment)." />
      <div className="max-w-xl space-y-4 rounded-lg border border-slate-200 bg-white p-6">
        <div>
          <div className="text-sm font-medium text-slate-700">Docker host</div>
          <div className="mt-1 font-mono text-sm text-slate-900">{data?.docker_host ?? "—"}</div>
        </div>
        <div>
          <div className="text-sm font-medium text-slate-700">Connection</div>
          <div className="mt-1 text-sm text-slate-900">
            {isLoading ? "Checking…" : data?.connected ? "Connected" : "Unreachable"}
          </div>
        </div>
        <div>
          <div className="text-sm font-medium text-slate-700">Swarm</div>
          <div className="mt-1 text-sm text-slate-900">
            {data?.swarm_active ? `Active (${data.node_role || "unknown role"})` : "Not active"}
          </div>
        </div>
        <Button variant="secondary" disabled={isFetching} onClick={() => void refetch()}>
          {isFetching ? "Testing…" : "Test connection"}
        </Button>
      </div>
    </>
  );
}
