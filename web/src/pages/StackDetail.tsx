import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { DataTable, PageHeader } from "../components/DataTable";
import { Button } from "../components/ui/primitives";

/** Stack detail with services, replicas, and published ports. */
export default function StackDetailPage() {
  const { name = "" } = useParams();
  const { data, isLoading, error } = useQuery({
    queryKey: ["swarm", "stack", name],
    queryFn: () => api.stack(name),
    enabled: Boolean(name),
  });

  return (
    <>
      <div className="mb-4">
        <Link className="text-sm text-slate-600 hover:text-slate-900" to="/stacks">
          ← Back to stacks
        </Link>
      </div>
      <PageHeader title={name} description="Services in this stack." />
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
