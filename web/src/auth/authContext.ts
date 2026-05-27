import { createContext, useContext } from "react";
import type { User } from "../api/client";

/** Auth state and actions exposed to React components. */
export type AuthContextValue = {
  /** Current user, or null when unauthenticated. */
  user: User | null;
  /** True while the initial session check is in progress. */
  loading: boolean;
  /** Re-fetches the current user from the API. */
  refresh: () => Promise<void>;
  /** Authenticates and stores the session. */
  login: (email: string, password: string) => Promise<void>;
  /** Clears the session and local auth state. */
  logout: () => Promise<void>;
};

/** React context holding auth state for the application. */
export const AuthContext = createContext<AuthContextValue | null>(null);

/** Returns the current auth context. */
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
