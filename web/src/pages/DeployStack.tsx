import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, ApiError } from "../api/client";
import { PageHeader } from "../components/DataTable";
import { Button, Input, Label } from "../components/ui/primitives";
import { useDockerStatus } from "../hooks/useDockerStatus";

/** Deploy a new stack from Compose YAML. */
export default function DeployStackPage() {
  const navigate = useNavigate();
  const { data: dockerStatus } = useDockerStatus();
  const [name, setName] = useState("");
  const [compose, setCompose] = useState("services:\n  web:\n    image: nginx:alpine\n");
  const [errors, setErrors] = useState<Array<{ path: string; message: string }>>([]);
  const disabled = dockerStatus != null && !dockerStatus.connected;

  const validate = useMutation({
    mutationFn: () => api.validateStack(name, compose),
    onSuccess: (result) => {
      setErrors(result.errors ?? []);
    },
    onError: (err: unknown) => {
      const message = err instanceof ApiError ? err.message : "Validation request failed";
      setErrors([{ path: "compose", message }]);
    },
  });

  const deploy = useMutation({
    mutationFn: () => api.deployStack(name, compose),
    onSuccess: (detail) => {
      void navigate(`/stacks/${encodeURIComponent(detail.name)}`);
    },
    onError: (err: unknown) => {
      if (err instanceof ApiError && err.status === 400) {
        setErrors([{ path: "compose", message: err.message }]);
      }
    },
  });

  async function handleValidate() {
    setErrors([]);
    try {
      await validate.mutateAsync();
    } catch {
      // onError sets inline validation errors
    }
  }

  async function handleDeploy() {
    setErrors([]);
    let result;
    try {
      result = await validate.mutateAsync();
    } catch {
      return;
    }
    if (!result.valid) {
      setErrors(result.errors ?? []);
      return;
    }
    await deploy.mutateAsync();
  }

  return (
    <>
      <PageHeader title="Deploy stack" description="Paste or upload a Compose file to deploy to Swarm." />
      <div className="max-w-3xl space-y-4 rounded-lg border border-slate-200 bg-white p-6">
        <div>
          <Label htmlFor="stack-name">Stack name</Label>
          <Input
            id="stack-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="web"
            disabled={disabled}
          />
        </div>
        <div>
          <Label htmlFor="compose">Compose YAML</Label>
          <textarea
            id="compose"
            className="min-h-64 w-full rounded-md border border-slate-300 px-3 py-2 font-mono text-sm shadow-sm focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500"
            value={compose}
            onChange={(e) => setCompose(e.target.value)}
            disabled={disabled}
          />
        </div>
        {errors.length > 0 && (
          <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">
            <div className="font-medium">Validation errors</div>
            <ul className="mt-2 list-disc pl-5">
              {errors.map((e) => (
                <li key={`${e.path}:${e.message}`}>
                  <span className="font-mono">{e.path}</span>: {e.message}
                </li>
              ))}
            </ul>
          </div>
        )}
        <div className="flex gap-2">
          <Button variant="secondary" disabled={disabled || validate.isPending} onClick={() => void handleValidate()}>
            {validate.isPending ? "Validating…" : "Validate"}
          </Button>
          <Button disabled={disabled || deploy.isPending} onClick={() => void handleDeploy()}>
            {deploy.isPending ? "Deploying…" : "Deploy"}
          </Button>
        </div>
        {disabled && <p className="text-sm text-amber-700">Docker is unreachable — deploy is disabled.</p>}
        {deploy.error instanceof ApiError && deploy.error.status !== 400 && (
          <p className="text-sm text-red-700">{deploy.error.message}</p>
        )}
      </div>
    </>
  );
}
