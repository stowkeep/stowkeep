import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { api, ApiError } from "../api/client";
import { AuthShell } from "../components/AuthShell";
import { Button, Input, Label } from "../components/ui/primitives";

/** First-run admin account creation. */
export default function SetupPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await api.setupAdmin(email, password);
      window.location.href = "/";
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Setup failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthShell title="Create admin account" subtitle="First-run setup for Stowkeep.">
      <form className="space-y-4" onSubmit={(e) => void onSubmit(e)}>
        <div>
          <Label htmlFor="email">Email</Label>
          <Input id="email" type="email" autoComplete="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div>
          <Label htmlFor="password">Password</Label>
          <Input id="password" type="password" autoComplete="new-password" minLength={8} required value={password} onChange={(e) => setPassword(e.target.value)} />
        </div>
        {error ? <p className="text-sm text-red-600">{error}</p> : null}
        <Button type="submit" disabled={loading} className="w-full">
          {loading ? "Creating…" : "Create admin"}
        </Button>
      </form>
      <p className="mt-4 text-center text-sm text-slate-500">
        Already set up? <Link className="text-slate-900 underline" to="/login">Sign in</Link>
      </p>
    </AuthShell>
  );
}
