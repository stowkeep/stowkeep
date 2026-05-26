import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "./authContext";

/** Redirects unauthenticated users to login or setup. */
export function RequireAuth() {
  const { user, loading } = useAuth();
  const location = useLocation();
  const [needsBootstrap, setNeedsBootstrap] = useState<boolean | null>(null);

  useEffect(() => {
    void api.setupStatus().then((s) => setNeedsBootstrap(s.needs_bootstrap));
  }, []);

  if (loading || needsBootstrap === null) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-600">Loading…</div>
    );
  }

  if (needsBootstrap) {
    return <Navigate to="/setup" replace />;
  }

  if (!user) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }

  return <Outlet />;
}

/** Redirects authenticated users away from guest pages. */
export function GuestOnly() {
  const { user, loading } = useAuth();
  const location = useLocation();
  const [needsBootstrap, setNeedsBootstrap] = useState<boolean | null>(null);

  useEffect(() => {
    void api.setupStatus().then((s) => setNeedsBootstrap(s.needs_bootstrap));
  }, []);

  if (loading || needsBootstrap === null) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-600">Loading…</div>
    );
  }

  if (needsBootstrap && location.pathname !== "/setup") {
    return <Navigate to="/setup" replace />;
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}

/** Setup page only when bootstrap is required. */
export function SetupOnly() {
  const { loading } = useAuth();
  const [needsBootstrap, setNeedsBootstrap] = useState<boolean | null>(null);

  useEffect(() => {
    void api.setupStatus().then((s) => setNeedsBootstrap(s.needs_bootstrap));
  }, []);

  if (loading || needsBootstrap === null) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-600">Loading…</div>
    );
  }

  if (!needsBootstrap) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
