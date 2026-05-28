import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

/** Polls enabled server feature flags from GET /api/v1/version. */
export function useFeatures() {
  const query = useQuery({
    queryKey: ["version"],
    queryFn: api.version,
    staleTime: 60_000,
  });

  const features = new Set(query.data?.features ?? []);

  return {
    ...query,
    features,
    hasFeature: (name: string) => features.has(name),
    swarmEnabled: features.has("swarm_readonly"),
    stackDeployEnabled: features.has("stack_deploy"),
  };
}
