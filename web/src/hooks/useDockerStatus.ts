import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

/** Polls Docker/Swarm connectivity for layout and banner components. */
export function useDockerStatus() {
  return useQuery({
    queryKey: ["swarm", "status"],
    queryFn: api.swarmStatus,
    refetchInterval: 15000,
  });
}
