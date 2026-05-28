import { Navigate } from "react-router-dom";
import { useFeatures } from "../hooks/useFeatures";

/** Sends authenticated users to the first available dashboard view. */
export function HomeRedirect() {
  const { isLoading, swarmEnabled } = useFeatures();

  if (isLoading) {
    return (
      <div className="flex min-h-[12rem] items-center justify-center text-slate-600">Loading…</div>
    );
  }

  if (swarmEnabled) {
    return <Navigate to="/nodes" replace />;
  }

  return <Navigate to="/settings" replace />;
}
