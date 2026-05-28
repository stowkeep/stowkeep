import { Navigate, Outlet } from "react-router-dom";
import { useFeatures } from "../hooks/useFeatures";

/** Redirects when a required server feature flag is disabled. */
export function RequireFeature({ feature }: { feature: string }) {
  const { hasFeature, isLoading, isError } = useFeatures();

  if (isLoading) {
    return (
      <div className="flex min-h-[12rem] items-center justify-center text-slate-600">Loading…</div>
    );
  }

  if (isError || !hasFeature(feature)) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
