import { Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./auth/AuthProvider";
import { GuestOnly, RequireAuth, SetupOnly } from "./auth/RouteGuards";
import { DashboardLayout } from "./layouts/DashboardLayout";
import LoginPage from "./pages/Login";
import SetupPage from "./pages/Setup";
import StackDetailPage, { SettingsPage } from "./pages/StackDetail";
import NodesPage, { ServicesPage, StacksPage, TasksPage } from "./pages/SwarmPages";

/** Application routes. */
export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<SetupOnly />}>
          <Route path="/setup" element={<SetupPage />} />
        </Route>
        <Route element={<GuestOnly />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>
        <Route element={<RequireAuth />}>
          <Route element={<DashboardLayout />}>
            <Route index element={<Navigate to="/nodes" replace />} />
            <Route path="/nodes" element={<NodesPage />} />
            <Route path="/services" element={<ServicesPage />} />
            <Route path="/tasks" element={<TasksPage />} />
            <Route path="/stacks" element={<StacksPage />} />
            <Route path="/stacks/:name" element={<StackDetailPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AuthProvider>
  );
}
