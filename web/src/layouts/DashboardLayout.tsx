import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../auth/authContext";
import { DockerBanner } from "../components/DockerBanner";
import { useDockerStatus } from "../hooks/useDockerStatus";
import { useFeatures } from "../hooks/useFeatures";
import { Button } from "../components/ui/primitives";
import { cn } from "../lib/utils";

const nav = [
  { to: "/nodes", label: "Nodes", feature: "swarm_readonly" as const },
  { to: "/services", label: "Services", feature: "swarm_readonly" as const },
  { to: "/tasks", label: "Tasks", feature: "swarm_readonly" as const },
  { to: "/stacks", label: "Stacks", feature: "swarm_readonly" as const },
  { to: "/settings", label: "Settings" },
];

/** Authenticated dashboard shell with sidebar navigation. */
export function DashboardLayout() {
  const { user, logout } = useAuth();
  const { data: dockerStatus } = useDockerStatus();
  const { hasFeature } = useFeatures();
  const swarmDisabled = dockerStatus != null && !dockerStatus.connected;

  const visibleNav = nav.filter((item) => !item.feature || hasFeature(item.feature));

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <DockerBanner />
      <div className="mx-auto flex min-h-screen max-w-7xl">
        <aside className="flex w-56 shrink-0 flex-col border-r border-slate-200 bg-white p-4">
          <div className="mb-8">
            <h1 className="text-lg font-semibold">Stowkeep</h1>
            <p className="text-xs text-slate-500">Swarm dashboard</p>
          </div>
          <nav className="flex flex-1 flex-col gap-1">
            {visibleNav.map((item) => {
              const isDisabled = swarmDisabled && item.to !== "/settings";
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  aria-disabled={isDisabled || undefined}
                  tabIndex={isDisabled ? -1 : undefined}
                  onClick={isDisabled ? (e) => e.preventDefault() : undefined}
                  className={({ isActive }) =>
                    cn(
                      "rounded-md px-3 py-2 text-sm font-medium",
                      isActive ? "bg-slate-900 text-white" : "text-slate-700 hover:bg-slate-100",
                      isDisabled && "pointer-events-none opacity-40",
                    )
                  }
                >
                  {item.label}
                </NavLink>
              );
            })}
          </nav>
          <div className="mt-auto space-y-2 border-t border-slate-200 pt-4 text-xs text-slate-500">
            <div className="truncate">{user?.email}</div>
            <Button variant="secondary" className="w-full" onClick={() => void logout()}>
              Log out
            </Button>
          </div>
        </aside>
        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
