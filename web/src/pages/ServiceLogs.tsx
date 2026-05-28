import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import { PageHeader } from "../components/DataTable";

/** Streaming service log viewer. */
export default function ServiceLogsPage() {
  const { serviceId = "" } = useParams();
  const { data, isLoading, error } = useQuery({
    queryKey: ["service-logs", serviceId],
    queryFn: () => api.serviceLogs(serviceId, { tail: "200" }),
    enabled: Boolean(serviceId),
  });

  return (
    <>
      <PageHeader title="Service logs" description={`Task output for service ${serviceId.slice(0, 12)}…`} />
      <pre className="max-h-[70vh] overflow-auto rounded-lg border border-slate-200 bg-slate-950 p-4 text-xs text-slate-100">
        {isLoading ? "Loading logs…" : error instanceof Error ? error.message : data || "No logs."}
      </pre>
    </>
  );
}
