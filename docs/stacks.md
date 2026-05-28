# Stack deploy and lifecycle

Stage 2 enables deploying and managing Swarm stacks from the Stowkeep UI when the `stack_deploy` feature flag is enabled.

## Enable stack deploy

Set both read-only and deploy flags (read-only is required for dashboard views):

```bash
STOWKEEP_FEATURES=swarm_readonly,stack_deploy
```

See [.env.example](../.env.example) and [install.md](./install.md).

## Deploy a stack

1. Sign in as the bootstrap admin.
2. Open **Stacks → Deploy stack**.
3. Enter a stack name (lowercase DNS-safe name, max 63 characters).
4. Paste Compose YAML and click **Validate** to see field-level errors.
5. Click **Deploy**. On success you are redirected to the stack detail page.

Compose is validated against the [Compose specification](https://compose-spec.io/). Stowkeep does **not** lint Swarm compatibility in v1 — some Compose keys are ignored by `docker stack deploy` (see PRD SW-09 / D-017).

## Remove a stack

From the stack detail page, click **Remove stack** and confirm. All services and stack-scoped overlay networks are removed.

## Scale a service

On the stack detail page, click **Scale** on a service row and enter the desired replica count.

## View logs

Click **Logs** on a service row to fetch recent task output (plain text).

## Common validation errors

| Error | Fix |
|-------|-----|
| `stack name must start with a lowercase letter` | Use names like `web`, `api_staging` |
| `compose file exceeds … byte limit` | Keep files under 1 MiB |
| `service "…" requires an image` | Add an `image:` key under each service |
| Invalid port syntax | Use `"8080:80"` form under `ports:` |

## API

Mutating routes live under `/api/v1/stacks` and require authentication plus `stack_deploy`. See [openapi/openapi.yaml](../openapi/openapi.yaml).

Deploy and remove actions write hash-chained audit events — see [audit.md](./audit.md).

## Permission builder prototype (D-010)

Before Stage 3 RBAC ships, review the static UX mockup: [prototypes/permission-builder.html](./prototypes/permission-builder.html).
