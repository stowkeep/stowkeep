# Configuration reference

Stowkeep reads settings from `STOWKEEP_*` environment variables. See [install.md](./install.md) for operator setup.

> **Stage 4:** `STOWKEEP_*_FILE` and MEK input normalization are specified here and implemented in Stage 4. See [ADR 0011](./adr/0011-bootstrap-config-and-mek-input.md).

## Local development

Copy [`.env.example`](../.env.example) to `.env`. The server loads `.env` on startup when present; **process environment overrides `.env`**.

Optional override path:

```bash
STOWKEEP_ENV_FILE=/path/to/local.env make dev
```

## Bootstrap variables

| Variable | Required | Default | Notes |
|----------|----------|---------|-------|
| `STOWKEEP_HTTP_ADDR` | No | `:8080` | HTTP listen address |
| `STOWKEEP_DATABASE_DRIVER` | No | sqlite | `sqlite` or `postgres` |
| `STOWKEEP_DATABASE_PATH` | SQLite | `/data/stowkeep.db` | SQLite file path |
| `STOWKEEP_DATABASE_URL` | Postgres | — | PostgreSQL DSN |
| `STOWKEEP_DOCKER_HOST` | No | `unix:///var/run/docker.sock` | Docker Engine endpoint |
| `STOWKEEP_FEATURES` | No | — | Comma-separated capability flags |
| `STOWKEEP_LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error` |
| `STOWKEEP_LOG_FORMAT` | No | `json` | `json` or `text` |
| `STOWKEEP_MASTER_KEY` | Stage 4+ | — | Homelab/testing MEK (passphrase, hex, or base64) |
| `STOWKEEP_MASTER_KEY_FILE` | Stage 4+ | — | Production MEK from Swarm secret mount |

### `STOWKEEP_*_FILE` (Stage 4)

When `STOWKEEP_MASTER_KEY` is unset, Stowkeep reads the master key from the file path in `STOWKEEP_MASTER_KEY_FILE` (typically `/run/secrets/...`). Direct env values take precedence over files.

## Feature flags

Set at deploy time; not controlled from the UI. The UI reads enabled flags from `GET /api/v1/version`.

| Flag | Enables |
|------|---------|
| `swarm_readonly` | Swarm dashboard |
| `stack_deploy` | Stack deploy, remove, scale, logs |
| `rbac` | Multi-user RBAC (Stage 3) |
| `secrets` | Secret store (Stage 4) |
| `gitops` | GitOps sync (Stage 5) |

See [install.md](./install.md) for examples.
