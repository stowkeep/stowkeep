import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

/** Banner shown when Docker Engine is unreachable (D-013). */
export function DockerBanner() {
  const { data, isLoading } = useQuery({
    queryKey: ["swarm", "status"],
    queryFn: api.swarmStatus,
    refetchInterval: 15000,
  });

  if (isLoading || !data || data.connected) {
    return null;
  }

  return (
    <div
      className="border-b border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-950"
      role="alert"
    >
      <strong>Docker unreachable.</strong> Swarm views are read-only until the engine at{" "}
      <code className="rounded bg-amber-100 px-1">{data.docker_host || "unknown"}</code> responds.
      {data.error ? ` ${data.error}` : ""}
    </div>
  );
}
