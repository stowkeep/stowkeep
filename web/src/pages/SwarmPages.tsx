import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { DataTable, PageHeader } from "../components/DataTable";
import { useFeatures } from "../hooks/useFeatures";

/** Swarm nodes list. */
export default function NodesPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["swarm", "nodes"],
    queryFn: async () => (await api.nodes()).items,
  });

  return (
    <>
      <PageHeader title="Nodes" description="Cluster members and their roles." />
      <DataTable
        loading={isLoading}
        error={error instanceof Error ? error.message : null}
        emptyMessage="No nodes found."
        columns={[
          { header: "Hostname", cell: (n) => n.hostname },
          { header: "Status", cell: (n) => n.status },
          { header: "Availability", cell: (n) => n.availability },
          { header: "Role", cell: (n) => (n.manager_lead ? `${n.role} (leader)` : n.role) },
          { header: "Address", cell: (n) => n.address || "—" },
        ]}
        rows={data ?? []}
        rowKey={(n) => n.id}
      />
    </>
  );
}

/** Swarm services list. */
export function ServicesPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["swarm", "services"],
    queryFn: async () => (await api.services()).items,
  });

  return (
    <>
      <PageHeader title="Services" description="Running Swarm services and replica counts." />
      <DataTable
        loading={isLoading}
        error={error instanceof Error ? error.message : null}
        emptyMessage="No services found."
        columns={[
          { header: "Name", cell: (s) => s.name },
          { header: "Image", cell: (s) => s.image },
          { header: "Mode", cell: (s) => s.mode },
          { header: "Replicas", cell: (s) => s.replicas },
          { header: "Stack", cell: (s) => s.stack || "—" },
          { header: "Update", cell: (s) => s.update_status || "—" },
        ]}
        rows={data ?? []}
        rowKey={(s) => s.id}
      />
    </>
  );
}

/** Swarm tasks list with live polling. */
export function TasksPage() {
  const { data, isLoading, error, dataUpdatedAt } = useQuery({
    queryKey: ["swarm", "tasks"],
    queryFn: async () => (await api.tasks()).items,
    refetchInterval: 8000,
  });

  return (
    <>
      <PageHeader
        title="Tasks"
        description="Task states refresh every 8 seconds."
        meta={dataUpdatedAt ? `Updated ${new Date(dataUpdatedAt).toLocaleTimeString()}` : undefined}
      />
      <DataTable
        loading={isLoading}
        error={error instanceof Error ? error.message : null}
        emptyMessage="No tasks found."
        columns={[
          { header: "Task ID", cell: (t) => t.id.slice(0, 12) },
          { header: "Service", cell: (t) => t.service_id.slice(0, 12) },
          { header: "Node", cell: (t) => t.node_id.slice(0, 12) },
          { header: "State", cell: (t) => t.state },
          { header: "Desired", cell: (t) => t.desired_state },
          { header: "Exit", cell: (t) => (t.exit_code != null ? String(t.exit_code) : "—") },
        ]}
        rows={data ?? []}
        rowKey={(t) => t.id}
      />
    </>
  );
}

/** Deployed stacks list. */
export function StacksPage() {
  const { stackDeployEnabled } = useFeatures();
  const { data, isLoading, error } = useQuery({
    queryKey: ["swarm", "stacks"],
    queryFn: async () => (await api.stacks()).items,
  });

  return (
    <>
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <PageHeader title="Stacks" description="Compose stacks deployed to Swarm." />
        {stackDeployEnabled && (
          <Link
            to="/stacks/deploy"
            className="inline-flex items-center rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
          >
            Deploy stack
          </Link>
        )}
      </div>
      <DataTable
        loading={isLoading}
        error={error instanceof Error ? error.message : null}
        emptyMessage="No stacks found."
        columns={[
          {
            header: "Name",
            cell: (s) => (
              <Link className="font-medium text-slate-900 underline" to={`/stacks/${encodeURIComponent(s.name)}`}>
                {s.name}
              </Link>
            ),
          },
          { header: "Services", cell: (s) => String(s.service_count) },
        ]}
        rows={data ?? []}
        rowKey={(s) => s.name}
      />
    </>
  );
}
