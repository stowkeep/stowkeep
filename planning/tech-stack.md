# Tech Stack Decision Record

**Project:** Stowkeep  
**Status:** Proposed (v0.1)  
**Last updated:** 2026-05-25

## Summary

| Layer | Choice | Primary rationale |
|-------|--------|-----------------|
| Backend API | **Go 1.23+** | Docker ecosystem alignment (Portainer, Doco-CD); single static binary; strong concurrency for reconcilers |
| Frontend | **React 19 + TypeScript + Vite** | Mature admin UI ecosystem; fast HMR; easy hiring |
| UI components | **shadcn/ui + Tailwind CSS** | Accessible primitives; full ownership of components; consistent design system |
| API style | **REST + OpenAPI 3.1** | Simple to debug; good tooling; SSE/WebSocket for live Swarm events |
| Embedded database | **SQLite 3** (`modernc.org/sqlite`) | Zero-dependency small installs; single file on Docker volume |
| Production database | **PostgreSQL 16** (recommended) | Multi-user concurrency, HA, standard ops tooling |
| Migrations | **goose** | Lightweight SQL migrations; no ORM lock-in |
| Data access | **sqlc** | Type-safe SQL; explicit queries for security-sensitive paths |
| Job queue / cache | **Redis 7** (Stage 5+) | Git sync workers, webhook debounce, session cache |
| Docker integration | **moby/moby client + compose-go** | Official Engine API; Compose spec parsing |
| Git operations | **go-git** | Pure Go; no shelling out to git binary in production |
| Auth | **OAuth2/OIDC + local accounts** | GitHub/GitLab/Google SSO; service accounts for automation |
| Authorization | **Casbin** (RBAC + ABAC conditions) | Infisical-style subject-action-object with scoped conditions |
| Secrets encryption | **Envelope encryption (AES-256-GCM + age/X25519)** | DEK per secret version; MEK in env/KMS; inspired by Infisical V2 |
| Git-encrypted secrets | **Mozilla SOPS** (optional path) | Encrypted files in Git repos; decrypt at deploy time |
| Observability | **slog (JSON stdout) + OpenTelemetry traces + Prometheus metrics** | 12-factor logs; no in-container log files; see [docs/logging.md](../docs/logging.md) |
| Container runtime | **Distroless or Alpine multi-stage** | Minimal attack surface for production image |

## Architecture overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Browser (React SPA)                       │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS / WSS
┌────────────────────────────▼────────────────────────────────────┐
│                     Go API Server (chi router)                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────┐ │
│  │ Auth/RBAC│ │ Swarm    │ │ Secrets  │ │ GitOps Reconciler    │ │
│  │          │ │ Proxy    │ │ Service  │ │ (poll + webhooks)    │ │
│  └──────────┘ └────┬─────┘ └──────────┘ └──────────┬───────────┘ │
└────────────────────┼───────────────────────────────┼─────────────┘
                     │ Docker Engine API              │ git clone / webhook
              ┌──────▼──────┐                  ┌─────▼─────┐
              │ Docker      │                  │ Git       │
              │ Socket/TLS  │                  │ Providers │
              └─────────────┘                  └───────────┘
                     │
              ┌──────▼──────┐
              │ SQLite or   │
              │ PostgreSQL  │
              │ Redis (opt) │
              └─────────────┘
```

## Backend: Go

**Why Go over Node/Python/Rust**

- **Portainer** and **Doco-CD** validate Go for Docker management tools.
- Static binary simplifies Swarm deployment (one service, minimal deps).
- `testcontainers-go` enables realistic integration tests against Docker.
- Goroutines suit long-running GitOps reconcilers and webhook handlers.

**Key libraries**

| Concern | Library |
|---------|---------|
| HTTP router | `go-chi/chi/v5` |
| Docker API | `github.com/moby/moby/client` |
| Compose | `github.com/compose-spec/compose-go/v2` |
| Validation | `github.com/go-playground/validator/v10` |
| Config | `github.com/kelseyhightower/envconfig` |
| JWT / OIDC | `github.com/coreos/go-oidc/v3`, `golang.org/x/oauth2` |
| RBAC | `github.com/casbin/casbin/v2` |
| Crypto | `crypto/aes`, `filippo.io/age`, `golang.org/x/crypto` |
| Git | `github.com/go-git/go-git/v5` |
| Migrations | `github.com/pressly/goose/v3` |
| SQL codegen | `github.com/sqlc-dev/sqlc` |
| SQLite driver | `modernc.org/sqlite` (pure Go, static binary friendly) |
| Postgres driver | `github.com/jackc/pgx/v5` |

## Frontend: React + TypeScript + Vite

**Why not Next.js**

- App is a self-hosted admin SPA behind the API; no SSR/SEO requirement.
- Vite keeps the frontend build tooling independent from Go.

**Packaging: `embed.FS` (D-026)**

Built static assets are baked into the Go binary via `//go:embed web/dist/*` and served by the same Go HTTP server. One binary, one container, one image, atomic FE/API releases. Same pattern as Portainer, Grafana, Vault, ArgoCD, Gitea, Caddy. The Dockerfile multi-stage build produces `web/dist/` in a Node stage and copies it into the Go build context. Local dev is unchanged: Vite dev server serves the UI with HMR and proxies `/api` to the Go process.

**Key libraries**

| Concern | Library |
|---------|---------|
| Routing | `react-router-dom` |
| Server state | `@tanstack/react-query` |
| Forms | `react-hook-form` + `zod` |
| Tables | `@tanstack/react-table` |
| Charts (Stage 2+) | `recharts` |
| Real-time | native `EventSource` / WebSocket hooks |
| API client | `openapi-typescript` generated types from OpenAPI spec |
| E2E | Playwright |
| Unit/component | Vitest + React Testing Library |

## Database: SQLite + PostgreSQL (dual backend)

Both backends store the same logical schema:

- Users, groups, roles, Casbin policy rows
- Audit events (append-only)
- Registered Swarm endpoints / agents
- Git repository connections and sync state (commit SHA, drift status)
- Secret **metadata** and **encrypted ciphertext** per version (never plaintext at rest)
- Preview environment records and lifecycle

### SQLite — embedded / small installs

| Attribute | Detail |
|-----------|--------|
| **Driver** | `modernc.org/sqlite` (pure Go, no CGO — works in distroless images) |
| **Default path** | `/data/stowkeep.db` in container; `./.data/dev.db` locally |
| **Persistence** | Single Docker named volume |
| **Pragmas** | WAL mode, foreign keys ON |
| **Best for** | Homelab, evaluation, single-node, minimal dependencies |

**Quick-start default:** If no `DATABASE_URL` is set, Stowkeep uses SQLite — one container, one volume, no Postgres sidecar.

### PostgreSQL — recommended production

| Attribute | Detail |
|-----------|--------|
| **Driver** | `pgx/v5` |
| **Best for** | Production, teams, concurrent GitOps syncs, large audit tables |
| **Ops** | Standard backup (`pg_dump`), replication, managed services |

### Implementation rules

1. **Portable SQL first** — migrations in `migrations/shared/`; dialect splits only when required.
2. **CI matrix** — every migration and integration test runs on **both** SQLite and PostgreSQL.
3. **No feature gaps** — same application features on both backends; document performance limits for SQLite only.
4. **UI hint** — setup wizard warns when SQLite is used in multi-user production contexts.

Full operator guide: [docs/database.md](../docs/database.md).

**Why not BoltDB (Portainer)**

- Relational queries, migrations tooling (goose/sqlc), and dual-backend testing are simpler with SQL.
- BoltDB lacks the cross-dialect portability story for contributors familiar with Postgres.

## Secrets architecture (high level)

Inspired by Infisical, adapted for Stowkeep scope:

1. **`MasterKeyProvider` interface (D-008):** Stowkeep never calls AES directly — it asks a `MasterKeyProvider` to wrap/unwrap DEKs. Ships with `EnvKey` (uses `STOWKEEP_MASTER_KEY`); a `KMSProvider` interface stub lands in Stage 4 so a cloud KMS (AWS KMS, Vault Transit) implementation in Stage 7 is non-breaking.
2. **Envelope encryption:** Random DEK encrypts secret value (AES-256-GCM); DEK wrapped with the active MEK; each DEK row stores a `key_id` so rotation can re-wrap DEKs in the background without re-encrypting payloads.
3. **Version table:** Every change creates a `secret_versions` row (git-style history); current pointer on `secrets`.
4. **Blind metadata:** Name, path, environment, tags stored in plaintext for RBAC condition evaluation; values always encrypted.
5. **Deploy-time materialization to Swarm (D-009):** Each Stowkeep secret version maps to a uniquely-named Docker Swarm secret (e.g. `appname_dbpassword_v7`). Rotation is a rolling update — new containers spin up referencing the new Swarm-secret name and only take traffic once healthy. Old Swarm secrets are GC'd once no service references them. Plaintext never touches disk on the operator host; values never appear in logs (sentinel-tested).
6. **SOPS path (D-028):** Stage 4 ships SOPS-encrypted YAML support alongside native secrets; reconciler decrypts at deploy with configured age/GPG keys.
7. **Backups (D-016):** Secret ciphertext travels through the product backup pipeline (SQLite Online Backup API / `pg_dump` streaming), encrypted with a `BACKUP_KEY` distinct from the MEK, HMAC-integrity-checked on restore.
8. **Approval workflows (Stage 7):** Change requests for protected environments before merge.

### Audit log

Append-only `audit_events` table is **hash-chained from day 1** (D-011): each row stores `prev_hash` and `row_hash = sha256(prev_hash || canonical(row))`. Startup verifies the chain and warns on break. Tamper-evidence cannot be retrofitted after data exists.

## GitOps reconciler

Custom in-process reconciler (not embedding Doco-CD) for unified RBAC/audit:

- **Pull model:** Poll interval + inbound webhooks (GitHub, GitLab, Gitea, Bitbucket).
- **Manifest:** `.stowkeep.yml` in repo root (name inspired by `.doco-cd.yml`).
- **Reconciliation loop:** clone/fetch → validate compose → resolve secrets → deploy stack → record status.
- **Drift detection:** Compare desired Git SHA vs deployed SHA; surface in UI.

## Deployment model

**Single container (MVP):** Go binary serves API + embedded SPA via `embed.FS` (D-026); connects to Docker socket via bind mount.

**Production (interim — D-012):** Place a [Tecnativa Docker socket proxy](https://github.com/Tecnativa/docker-socket-proxy) between Stowkeep and `/var/run/docker.sock` to restrict the allowed Docker API verbs. README quick-start documents the raw socket-mount risk explicitly.

**Production (long term — D-027):** First-party agent mode lands in Stage 7. Lightweight agent on worker/manager nodes; control plane connects via mTLS (Portainer-agent pattern) for multi-host without exposing the socket on the operator host.

## Logging and observability

Full guide: [docs/logging.md](../docs/logging.md).

| Concern | Approach |
|---------|----------|
| **Application logs** | `log/slog` → stdout, JSON in production |
| **Log shipping** | External (Docker log driver, Promtail, Datadog agent) — not embedded |
| **Request tracing** | `request_id` in context; OpenTelemetry traces in Stage 7 |
| **Metrics** | Prometheus `/metrics` endpoint (Stage 7) |
| **Security audit** | `audit_events` table — complementary to stdout logs |
| **Secret values** | Never logged; CI enforces |

Environment: `STOWKEEP_LOG_LEVEL`, `STOWKEEP_LOG_FORMAT` (`json`|`text`).

## Alternatives considered

| Option | Verdict |
|--------|---------|
| **Rust backend** | Excellent safety; slower iteration; smaller Docker-lib ecosystem |
| **Node/Bun backend** | Fast UI sharing; weaker fit for socket-heavy reconciler workloads |
| **SvelteKit full-stack** | Simpler monolith; smaller admin UI component ecosystem |
| **Kubernetes operator pattern** | Out of scope — target is Swarm, not K8s |
| **Embed Doco-CD** | Would split RBAC/audit; better to implement unified reconciler |
| **HashiCorp Vault** | Powerful but heavy; offer as optional external backend later |
| **MongoDB** | Poor fit for audit/version relational queries |
| **Zap / Zerolog** | Mature; extra dependency — slog sufficient for this scale |

## Monorepo layout (proposed)

```
stowkeep/
├── api/                 # Go backend
├── web/                 # React frontend
├── pkg/                 # Shared Go packages (docker, gitops, secrets, observability/log)
├── migrations/          # goose: shared/, postgres/, sqlite/
├── deploy/              # Compose/Swarm stack for self-host
├── planning/            # This directory
├── .github/workflows/   # CI/CD
├── Dockerfile
├── docker-compose.yml   # Local dev
├── Makefile
└── openapi/             # Generated + hand-written API spec
```

## Local development requirements

- Docker Desktop or Docker Engine with Swarm init optional
- Go 1.23+, Node 22+
- **Default dev:** SQLite (no external DB)
- **Production-like dev:** PostgreSQL 16 via `make dev-postgres`
- Redis 7 optional until Stage 5
- `make dev` orchestrates API + web

## Decision log

The full live decision tracker is at [planning/decisions-todo.md](./decisions-todo.md). Highlights:

| Date | Decision | Notes |
|------|----------|-------|
| 2026-05-25 | Go + React monorepo | Initial stack proposal |
| 2026-05-25 | SQLite + PostgreSQL dual backend | SQLite default for small installs; Postgres recommended production ([ADR 0001](../docs/adr/0001-dual-database-sqlite-postgres.md)) |
| 2026-05-25 | Structured logging with slog | JSON to stdout ([ADR 0002](../docs/adr/0002-structured-logging-slog.md)) |
| 2026-05-25 | Single-replica control plane in v1 | Multi-replica + leader election is a Stage 5 design decision (D-007) |
| 2026-05-25 | `MasterKeyProvider` interface in Stage 4 day 1 | Env-var impl + KMS stub; per-DEK `key_id` (D-008) |
| 2026-05-25 | Swarm secrets abstraction via versioned naming + rolling updates | D-009 |
| 2026-05-25 | RBAC engine = Casbin, with mandatory Stage 2 UI prototype gate | D-010 |
| 2026-05-25 | Hash-chained audit log from day 1 | D-011 |
| 2026-05-25 | Docker socket = direct mount + socket-proxy interim + agent mode Stage 7 | D-012, D-027 |
| 2026-05-25 | Webhooks deferred to Stage 5.5; polling + manual sync only in Stage 5 | D-014 |
| 2026-05-25 | Backups elevated to first-class P0 product feature (SQLite + Postgres → local + S3) | D-016 |
| 2026-05-25 | Frontend embedded via `embed.FS` | D-026 |
| 2026-05-25 | SOPS support in Stage 4 alongside native secrets | D-028 |
| 2026-05-25 | Generic Git remote support in v1; provider-specific features per-provider when scheduled | D-029 |
| 2026-05-25 | Anonymous opt-in telemetry, off by default, implementation in Stage 7 | D-030 |
