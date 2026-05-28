# Product Requirements Document (PRD)

**Product:** Stowkeep  
**Version:** 0.1 (Draft)  
**Last updated:** 2026-05-25  
**Status:** Planning

---

## 1. Executive summary

Stowkeep is a self-hosted web UI and control plane for **Docker Swarm**. It connects to the Docker Engine (typically via socket on a manager node) and provides:

- Visual management of nodes, services, stacks, tasks, networks, and volumes
- One-click deployment of Compose files and Swarm stacks
- **Pull-based GitOps** — repositories as source of truth, synced on webhook or poll (Doco-CD / Argo-style)
- **Secrets management** with encryption, versioning, RBAC, and optional approval workflows (Infisical-inspired)
- **Preview / ephemeral environments** per pull request (Dokploy-inspired)
- **Deep RBAC** — users, groups, roles, and fine-grained permissions with scoped conditions

The product targets teams who run production workloads on Swarm and want Portainer-level usability with modern GitOps, secrets, and preview workflows — without migrating to Kubernetes.

---

## 2. Problem statement

| Pain | Today | With Stowkeep |
|------|-------|---------------------|
| Swarm UX is CLI-heavy | `docker stack deploy`, manual secret creation | Guided UI + validated Compose editor |
| No native GitOps for Swarm | Custom scripts, Portainer Git stacks (push-ish), or Doco-CD as separate tool | Unified pull-based sync with drift visibility |
| Secrets scattered | `.env` files, manual `docker secret create` | Central store, version history, RBAC, audit |
| PR previews on Swarm | Manual stacks or external PaaS | Automated ephemeral stacks with cleanup |
| Access control | Shared admin on Portainer or SSH keys | Granular RBAC per environment, stack, secret path |

---

## 3. Goals and non-goals

### Goals

1. **Swarm-native** — First-class support for stacks, services, configs, and secrets; Compose v2 spec.
2. **GitOps by default** — Declarative deployments from Git with observable sync status.
3. **Security-first secrets** — Encrypted at rest, never logged, permission-gated read, full version history.
4. **Operable RBAC** — Admins can express policies like "deploy to `staging` but not `production`" or "read secret names in `/db` but not values."
5. **Preview environments** — Isolated PR stacks with automatic teardown.
6. **Self-hosted single binary/container** — Low resource footprint suitable for homelab and production Swarm.
7. **Zero-dependency quick start** — SQLite on a Docker volume for small installs; PostgreSQL recommended for production.

### Non-goals (v1 / initial releases)

- Kubernetes or Nomad support
- Replacing full CI/CD (we integrate with GitHub Actions; we don't replace build pipelines)
- Built-in container registry
- Multi-tenant SaaS hosting (self-hosted only initially)
- Windows containers
- Replacing HashiCorp Vault — Vault is a great fit for Kubernetes/Nomad teams. Stowkeep targets the gap *below* that orchestration tier. A `SecretBackend` interface with a Vault implementation is on the Stage 7+ roadmap for teams that mature into Vault without changing orchestration.

---

## 4. Personas

| Persona | Needs |
|---------|-------|
| **Platform engineer** | Connect Swarm, configure Git repos, define RBAC, manage secrets, set preview policies |
| **Application developer** | Deploy stacks, view logs, trigger redeploys, open PR previews, rotate secrets (with approval) |
| **Security / compliance** | Audit logs, approval workflows, least-privilege roles, no plaintext secrets in DB |
| **Homelab operator** | Simple install, socket mount, deploy Compose from Git with minimal setup |

---

## 5. Core concepts

| Concept | Definition |
|---------|------------|
| **Environment** | Logical scope (e.g. `production`, `staging`, `preview`) mapping to a Swarm context and config |
| **Stack** | Swarm stack deployed from Compose file(s) |
| **Application** | Git-backed deployable unit with sync config (`.stowkeep.yml`) |
| **Secret path** | Hierarchical namespace (`/app/database`, `/shared/registry`) within an environment |
| **Sync** | GitOps reconciliation run — fetch repo, resolve secrets, deploy, record status |
| **Preview** | Ephemeral stack tied to a PR/branch with unique URL and TTL |
| **Role** | Collection of permissions; assignable to users and groups |
| **Permission** | Subject + action + optional conditions (environment, stack, secret path, tags) |

---

## 6. Functional requirements

### 6.1 Docker Swarm connection

| ID | Requirement | Priority |
|----|-------------|----------|
| SW-01 | Connect via local Unix socket (default for homelab quick-start), a [Tecnativa-style socket proxy](https://github.com/Tecnativa/docker-socket-proxy) (recommended for any production install), or remote TLS (`DOCKER_HOST`). First-party agent mode is targeted for Stage 7. | P0 |
| SW-02 | Detect Swarm mode; show manager/worker role | P0 |
| SW-03 | List nodes with status, availability, labels, resources | P0 |
| SW-04 | List services with replicas, image, update status | P0 |
| SW-05 | List tasks with state, node, exit codes | P0 |
| SW-06 | List stacks with services and deploy time | P0 |
| SW-07 | View service/task logs (streaming) | P1 |
| SW-08 | Scale services, rolling updates, rollback | P1 |
| SW-09 | Create/update/remove stacks from Compose (validated against the Compose spec; **Swarm-compatibility validation is intentionally not built in v1** — target users are technical enough to know that `docker stack deploy` silently ignores some Compose v2 keys) | P0 |
| SW-10 | Manage networks, configs, volumes (read + CRUD where API allows) | P2 |
| SW-11 | Multi-endpoint registry (multiple Swarm clusters) | P2 |
| SW-12 | Prominent UI banner whenever the Docker daemon is unreachable; UI degrades to read-only where feasible (decision D-013) | P0 |

### 6.2 Compose and stack deployment

| ID | Requirement | Priority |
|----|-------------|----------|
| CD-01 | Upload or paste Compose YAML; validate against the Compose spec (Swarm-compat validation deferred — see SW-09 and D-017) | P0 |
| CD-02 | Deploy as Swarm stack with name + env override | P0 |
| CD-03 | Show deploy diff (services added/changed/removed) before apply | P1 |
| CD-04 | Support `docker-compose.yml` + override files | P1 |
| CD-05 | Inject secrets/env from Stowkeep secret store at deploy | P1 |
| CD-06 | Deploy history per stack (who, when, git SHA if applicable) | P1 |

### 6.3 GitOps (pull-based sync)

| ID | Requirement | Priority |
|----|-------------|----------|
| GO-01 | Register **any Git remote** via HTTPS+token or SSH+deploy key. Provider-agnostic by design — works against GitHub, GitLab (cloud or self-hosted), Gitea, Forgejo, Bitbucket, AWS CodeCommit, Azure DevOps, or self-hosted `git+ssh`. Provider-specific features (PR comments, native webhook signature validation) are scheduled per-provider when the dependent feature lands (D-029). | P0 |
| GO-02 | Configure via `.stowkeep.yml` in repo | P0 |
| GO-03 | Poll on interval (e.g. 1–60 min) | P0 |
| GO-04 | Inbound webhooks for push/PR events — **deferred to a Stage 5.5 bolt-on** (D-014). Stage 5 ships polling + manual "Sync now" only. | P1 (Stage 5.5) |
| GO-05 | Show sync status: synced, out of sync, error, last commit SHA | P0 |
| GO-06 | Manual "Sync now" with audit | P0 |
| GO-07 | Auto-discover compose files in subdirectories (optional) | P2 |
| GO-08 | Branch/tag pinning per environment | P1 |
| GO-09 | SOPS-encrypted secrets in repo (decrypt at deploy) — **promoted into Stage 4 alongside native secrets** (D-028) | P1 |
| GO-10 | PR comment with deploy status / preview URL — per-provider feature; scheduled with webhook ingestion (Stage 5.5+) | P2 |

**Example `.stowkeep.yml`**

```yaml
name: api
working_dir: ./
compose_files:
  - docker-compose.yml
environment: production
branch: main
poll_interval: 5m
secrets:
  - path: /api/database
    inject: file
```

### 6.4 Secrets management

| ID | Requirement | Priority |
|----|-------------|----------|
| SEC-01 | Create/read/update/delete secrets within environment + path | P0 |
| SEC-02 | Full version history per secret (git-style log) | P0 |
| SEC-03 | Diff between versions (metadata always; value only if permitted) | P1 |
| SEC-04 | Rollback to previous version | P1 |
| SEC-05 | Envelope encryption at rest via a `MasterKeyProvider` interface (`EnvKey` impl in v1; cloud KMS implementations addable without rewrite). Each DEK row stores a `key_id` so future MEK rotation does not require big-bang re-encryption (D-008). | P0 |
| SEC-06 | Never write secret values to logs, audit payloads, or error traces; CI sentinel-value tests enforce on every release (D-021) | P0 |
| SEC-07 | Separate permissions: `describe` vs `read_value` vs `write` | P0 |
| SEC-08 | Tags on secrets for RBAC conditions | P2 |
| SEC-09 | Approval workflow for protected paths (change request → approve → merge) | P2 |
| SEC-10 | Materialize secrets into Swarm via the **versioned-secret pattern**: each SO secret version maps to a uniquely-named Swarm secret (e.g. `appname_dbpassword_v7`); services are updated to reference the new name via rolling update; old Swarm secrets are garbage-collected once unreferenced. Rotation = rolling update; new container with new secret comes up healthy before traffic shifts (D-009). | P1 |
| SEC-11 | Point-in-time view ("as of version N") | P2 |
| SEC-12 | Export audit of secret access (who read value, when) | P1 |
| SEC-13 | Backups of secret ciphertext travel through the product backup pipeline (DB-07) and are encrypted with a `BACKUP_KEY` distinct from the MEK | P0 |

### 6.5 RBAC and identity

| ID | Requirement | Priority |
|----|-------------|----------|
| AUTH-01 | Local email/password accounts (bcrypt/argon2) | P0 |
| AUTH-02 | OIDC SSO (GitHub, GitLab, Google) | P1 |
| AUTH-03 | Users and groups | P0 |
| AUTH-04 | Built-in roles: `owner`, `admin`, `developer`, `viewer` | P0 |
| AUTH-05 | Custom roles with fine-grained permission builder UI | P1 |
| AUTH-06 | Permission conditions: environment, stack name, secret path prefix, tags | P1 |
| AUTH-07 | Service accounts / API tokens for automation | P1 |
| AUTH-08 | Session management + optional MFA (TOTP) | P2 |
| AUTH-09 | **Hash-chained** audit log for all mutating actions (each row stores `prev_hash` + `row_hash = sha256(prev_hash \|\| canonical(row))`); chain integrity verified on startup. Tamper-evident from day 1 — cannot be retrofitted later (D-011). | P0 |

**Permission model (subject → actions → conditions)**

| Subject | Example actions |
|---------|-----------------|
| `swarm.nodes` | list, inspect |
| `swarm.stacks` | list, deploy, remove, inspect |
| `swarm.services` | list, scale, update, logs |
| `gitops.apps` | list, sync, configure |
| `secrets` | describe, read_value, create, edit, delete |
| `previews` | create, delete, configure |
| `rbac` | manage_users, manage_roles |
| `audit` | read |

### 6.6 Preview / ephemeral environments

> **Priority note (D-015):** Preview environments stay in the roadmap (Stage 6) but are explicitly the **lowest-priority stage**. Routing helper is part of the same stage (see PRE-09) — previews are useless without it. Detailed design is deferred (DEF-004) until Stage 5.5 webhook ingestion lands.

| ID | Requirement | Priority |
|----|-------------|----------|
| PRE-01 | Enable previews per Git-backed application | P1 |
| PRE-02 | Auto-create stack on PR open; update on push; delete on close/merge | P1 |
| PRE-03 | Templated preview URL: `{pr_number}.{app}.{base_domain}` | P1 |
| PRE-04 | Preview-specific env vars and secret paths (scoping design deferred to Stage 6 kickoff) | P1 |
| PRE-05 | Configurable max concurrent previews per app | P1 |
| PRE-06 | Deploy from CI-built image (webhook/API with image tag) — Dokploy pattern | P2 |
| PRE-07 | TTL auto-cleanup for stale previews | P2 |
| PRE-08 | Require collaborator permission on repo | P2 |
| PRE-09 | Traefik / nginx labels helper for routing — **promoted to P1**; previews ship with at least one opinionated routing path | P1 |

### 6.7 Observability and operations

| ID | Requirement | Priority |
|----|-------------|----------|
| OPS-01 | Health endpoints (`/healthz`, `/readyz`) | P0 |
| OPS-02 | Structured JSON logs to stdout (`slog`); configurable level and format | P0 |
| OPS-02a | HTTP request logging (method, path, status, duration) — no bodies or auth headers | P0 |
| OPS-02b | `request_id` correlation on all HTTP and derived background work | P1 |
| OPS-03 | Prometheus metrics (sync duration, deploy count, errors) | P2 |
| OPS-04 | Email/Slack/webhook notifications on sync failure | P2 |
| OPS-05 | **Scheduled, encrypted backups** for both SQLite (via the `sqlite3` Online Backup API — never `cp`) and PostgreSQL (`pg_dump` streaming), with local-volume and S3-compatible destinations, configurable from the UI. One-click restore. Backup files encrypted with a `BACKUP_KEY` distinct from the MEK. First-class product feature — not Stage 7 polish (D-016). | P0 |

### 6.8 Data persistence

| ID | Requirement | Priority |
|----|-------------|----------|
| DB-01 | SQLite embedded mode — single file on configurable path (default `/data/stowkeep.db`) | P0 |
| DB-02 | PostgreSQL support via connection URL (recommended for production) | P0 |
| DB-03 | Same schema and features on both backends; CI validates migrations on both | P0 |
| DB-04 | Auto-detect driver from URL or explicit `DATABASE_DRIVER` | P0 |
| DB-05 | UI indicates active driver; warn when SQLite used in production multi-user context | P1 |
| DB-06 | Documented path to migrate SQLite → PostgreSQL | P2 |
| DB-07 | Backups (see OPS-05) cover the full DB plus encrypted secret ciphertext; integrity-checked at restore (HMAC with `BACKUP_KEY`); restore requires `system:restore` permission (D-016) | P0 |

---

## 7. Non-functional requirements

| Category | Target |
|----------|--------|
| **Availability** | **Single-replica control plane in v1** (D-007); recoverable via restart. Swarm workloads run independently if the control plane is down. Multi-replica + leader election is a Stage 5+ design decision (DEF-012). |
| **Performance** | Dashboard load < 2s for clusters with ≤ 50 services |
| **Security** | OWASP ASVS L2 aspirational; CSRF protection; rate limiting on auth |
| **Compatibility** | Docker Engine 24+; Swarm mode; Compose specification v2 |
| **Resource** | Idle < 256MB RAM (excluding external PostgreSQL server) |
| **Browser support** | Last 2 versions Chrome, Firefox, Safari, Edge |

---

## 8. Phased delivery roadmap

Each stage is independently shippable and adds user-visible value. Stages assume **trunk-based development** — each merges to `main` behind feature flags where needed.

### Phase gates (mandatory)

**No stage may begin until the previous stage's phase gate passes.**

Every stage must satisfy three pillars before the next stage starts:

1. **Testing** — CI green, stage test plan complete, manual smoke done
2. **Documentation** — user docs, in-code docs, CHANGELOG, ADRs/threat model as needed
3. **Code quality** — lint clean, reviews complete, migrations dual-DB, no critical CVEs

Full checklists, sign-off process, and LLM stop rules: **[planning/phase-gates.md](./phase-gates.md)**.

Per-PR requirements (every commit during a stage): **[planning/open-source-standards.md](./open-source-standards.md)**.

Agents and junior contributors should read **[AGENTS.md](../AGENTS.md)** before implementing.

---

### Stage 0 — Foundation (2–3 weeks)

**Outcome:** Empty repo becomes a buildable, testable, open-source-ready product skeleton with foundation hardening that cannot cheaply be retrofitted later.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Monorepo scaffold (Go API + React web) | `make dev` runs API and UI locally (SQLite default); Vite dev server proxies `/api` to Go |
| Frontend embedded via `embed.FS` (D-026) | Production binary serves SPA from `//go:embed web/dist/*`; no nginx sidecar |
| Dockerfile multi-stage build | `docker build` produces runnable image; SQLite quick-start works with one volume; multi-arch `linux/amd64` + `linux/arm64` required, `linux/arm/v7` best-effort (D-018) |
| GitHub Actions: lint, test, build image on PR | PR shows green checks; CI matrix runs tests on SQLite **and** PostgreSQL; all actions pinned to SHAs |
| Supply-chain posture in CI from day 1 (D-022) | Cosign image signing, SLSA provenance attestation, Syft SBOM publication on every release-image build |
| Database migrations baseline | goose migrations apply cleanly on both backends; hash-chained `audit_events` schema landed (D-011); MEK envelope-canary row landed |
| Health check endpoint | `GET /healthz` returns 200 |
| Open source foundation | LICENSE, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, CHANGELOG, issue/PR templates |
| `CODEOWNERS` and `GOVERNANCE.md` (D-023, D-024) | Security-sensitive paths (`pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/`) require maintainer review; governance documents single-maintainer status and hand-off plan |
| Threat model in repo (already present) | `docs/security/threat-model.md` — updated when any new ingress/trust boundary lands |
| Project decisions tracker (already present) | `planning/decisions-todo.md` — walked at every release |
| ADRs 0003–0009 drafted | HA stance, MasterKeyProvider, audit hash chain, Swarm secrets abstraction, RBAC commitment, frontend embed, telemetry posture |
| Contributor docs | `docs/development-guide.md`, `docs/code-standards.md`, `docs/database.md`, `docs/logging.md` |
| Planning docs in repo | PRD, tech stack, testing strategy, open-source standards, decisions tracker |

**Feature flag:** N/A  
**User-visible:** None (internal)  
**Phase gate:** [Stage 0 checklist](./phase-gates.md#stage-0--foundation)

---

### Stage 1 — Swarm read-only dashboard (2–3 weeks)

**Outcome:** Connect to Docker and explore the cluster.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Docker socket connection config | Admin sets socket path or `DOCKER_HOST`; connection tested in UI |
| Nodes, services, tasks, stacks list views | Data matches `docker node ls`, `docker service ls`, etc. |
| Stack detail view | Shows compose services, replicas, published ports |
| Basic local auth (single admin bootstrap) | First-run creates admin user; login required |
| Live task state refresh | UI updates task states without full page reload |

**Depends on:** Stage 0  
**Feature flag:** `swarm_readonly`  
**Phase gate:** [Stage 1 checklist](./phase-gates.md#stage-1--swarm-read-only-dashboard)

---

### Stage 2 — Deploy and manage stacks (2–3 weeks)

**Outcome:** Users can deploy and lifecycle-manage Swarm stacks from the UI.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Compose editor/uploader with validation | Invalid compose shows actionable errors |
| Deploy stack wizard | Stack appears in Swarm; services reach running state |
| Remove stack | Stack and services removed (with confirm) |
| Scale service | Replica count changes reflected in Swarm |
| Service logs viewer | Stream logs for selected task/service |
| Deploy audit entries | Who deployed what, when (chain-verified on read) |
| **Permission-builder UI prototype** (D-010) | Static HTML or Figma mockup of how an admin authors a custom role with environment + stack-name conditions. Sole purpose: validate that Casbin can power a usable admin UX *before* Stage 3 commits to the engine. If the prototype is clearly unworkable, that is the trigger to revisit the RBAC engine choice. |

**Depends on:** Stage 1  
**Feature flag:** `stack_deploy`  
**Phase gate:** [Stage 2 checklist](./phase-gates.md#stage-2--deploy-and-manage-stacks)

---

### Stage 3 — RBAC and audit (3–4 weeks)

**Outcome:** Multi-user access with least privilege.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Users, groups, roles CRUD | Admin can create developer group with limited role |
| Casbin policy enforcement on all API routes | Unauthorized requests return 403 |
| Permission builder UI | Admin creates custom role scoped to `staging` environment |
| Audit log UI | Filter by user, action, resource, time range |
| API tokens for automation | Token can deploy to allowed stacks only |

**Depends on:** Stage 2  
**Feature flag:** `rbac`  
**Phase gate:** [Stage 3 checklist](./phase-gates.md#stage-3--rbac-and-audit)

---

### Stage 4 — Secrets (4–6 weeks)

**Outcome:** Encrypted, versioned secrets integrated with deploys.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Secret CRUD with path hierarchy | Secrets organized by environment + path |
| Version history + rollback | Each edit creates version; rollback restores prior value |
| Envelope encryption via `MasterKeyProvider` interface (D-008) | DB inspection shows ciphertext only; `EnvKey` impl shipped; `KMSProvider` stub interface in place; each DEK row stores `key_id` |
| Bootstrap config via `STOWKEEP_*` and `STOWKEEP_*_FILE` | Operators set MEK via env (homelab) or Swarm secret file mounts (production); no bootstrap MEK wizard |
| MEK input normalization (`pkg/secrets/keyinput`) | Accept passphrase, hex, or base64 in env; raw bytes from `_FILE`; no manual base64 encoding required |
| MEK rotation procedure documented + tested | Rotation path re-wraps DEKs in the background without re-encrypting values; documented in `docs/secrets-rotation.md` |
| Swarm secret materialization (D-009) | Versioned Swarm secret names; rolling-update path; old-version garbage collection after no service references remain |
| Secret inject modes for deployed stacks | Default: Swarm secret file mounts; optional `inject: env` for legacy images (documented downgrade); `inject: file` is native Swarm behavior |
| `describe` vs `read_value` permissions | User with describe-only sees names, not values |
| Inject secrets into stack deploy | Deploy resolves `${SO_SECRET_*}` or configured mapping; supports `inject: file` (default) and opt-in `inject: env` for legacy containers (no plaintext in logs — sentinel-tested) |
| Secret access audit | Read-value events logged (hash-chained, no value) |
| **SOPS support** (D-028) | Encrypted YAML files in Git can be decrypted at deploy via configured age/GPG keys |
| Scheduled backups with secret ciphertext (D-016) | UI configurable; local + S3; HMAC-integrity-checked on restore |

**Depends on:** Stage 3  
**Feature flag:** `secrets`  
**Phase gate:** [Stage 4 checklist](./phase-gates.md#stage-4--secrets)

---

### Stage 5 — GitOps sync (4–6 weeks)

**Outcome:** Pull-based deployment from any Git remote.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Generic Git remote support (D-029) | HTTPS+token or SSH+deploy key; works against GitHub, GitLab, Gitea, Bitbucket, AWS CodeCommit, Azure DevOps, self-hosted — no provider-specific code |
| Application registration + `.stowkeep.yml` parsing | Invalid config rejected with clear errors |
| Poll-based sync | Change on `main` deploys within poll interval |
| Sync status dashboard | Shows last SHA, status, error messages |
| Manual "Sync now" trigger | Button triggers immediate reconciliation; audited |
| Drift indicator | UI shows when live stack differs from Git |
| Single-flight per-application reconcile | One concurrent reconcile per app; queue depth bounded |
| Repo size cap + shallow clone (DEF-003) | Documented max repo size; reject `.git` > N MB cleanly |

**Depends on:** Stage 4 (for secret resolution)  
**Feature flag:** `gitops`  
**Phase gate:** [Stage 5 checklist](./phase-gates.md#stage-5--gitops-sync)

> **Webhook receivers are NOT in this stage.** Stage 5 ships polling + manual sync only. Webhook ingestion is the Stage 5.5 bolt-on below (D-014).

---

### Stage 5.5 — Webhook ingestion (2–3 weeks, optional bolt-on)

**Outcome:** Push-triggered sync for users who don't want to wait for the poll interval.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Webhook receiver per provider | At minimum GitHub (`X-Hub-Signature-256`), GitLab (`X-Gitlab-Token`), Gitea (`X-Gitea-Signature`) — constant-time signature compare |
| Replay protection | Reject deliveries older than 5 minutes; idempotency key per delivery |
| Per-IP rate limiting | Documented limits; configurable |
| Audit of every webhook delivery | Source, delivery id, accepted/rejected, reason |

**Depends on:** Stage 5  
**Feature flag:** `webhooks`  
**Phase gate:** [Stage 5.5 checklist](./phase-gates.md#stage-55--webhook-ingestion)

---

### Stage 6 — Preview environments (lowest priority — explore after Stage 5.5)

**Outcome:** PR-based ephemeral stacks.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| Preview config per application | Enable/disable, domain template, limits |
| PR open → deploy preview stack | Unique stack name per PR |
| PR sync → redeploy | New commits update preview |
| PR close → teardown | Stack removed; secrets cleaned up |
| Preview env var overrides | `PREVIEW_*` vars injected |
| PR comment with preview URL (GitHub) | Comment posted on open/sync |

**Depends on:** Stage 5  
**Feature flag:** `previews`  
**Phase gate:** [Stage 6 checklist](./phase-gates.md#stage-6--preview-environments)

---

### Stage 7 — Hardening and enterprise (ongoing)

**Outcome:** Production-grade operations.

| Deliverable | Acceptance criteria |
|-------------|---------------------|
| OIDC SSO | Login via GitHub/Google/GitLab |
| MFA (TOTP / WebAuthn) | Optional per user; WebAuthn recommended for admins given socket-mount blast radius |
| Secret approval workflows | Protected paths require approver |
| Multi-cluster endpoints | Manage 2+ Swarm clusters from one UI |
| **Agent mode** (D-027) | mTLS-based agent on worker/manager nodes; replaces direct socket mount as recommended production deployment |
| **Multi-replica HA** (DEF-012) | PG advisory-lock leader election for the reconciler; shared session store; documented active-passive (and later active-active) modes |
| `SecretBackend` interface + Vault implementation (D-025, DEF-015) | Teams that mature into Vault can plug it in without changing orchestration |
| **Anonymous opt-in telemetry** (D-030) | Off by default; heartbeat-only payload; configurable endpoint URL; transparent docs |
| KMS implementation of `MasterKeyProvider` (DEF-001 / D-008) | At least one cloud KMS (AWS KMS or Vault Transit) with documented rotation runbook |
| Prometheus metrics + alerting hooks | `/metrics` scraped |
| HA documentation | Postgres HA, backup restore runbook, lost-MEK disaster recovery |

**Depends on:** Stage 6 (some items independent and may land earlier)  
**Phase gate:** [Stage 7 checklist](./phase-gates.md#stage-7--hardening-and-enterprise)

---

## 9. User journeys (MVP → full)

### Journey A: First deploy (Stage 1–2)

1. Operator installs Stowkeep container with Docker socket mount + data volume (SQLite — no Postgres required).
2. Opens UI, completes first-run admin setup.
3. Confirms Swarm connection; sees 3 nodes healthy.
4. Pastes `docker-compose.yml`, names stack `web`, deploys.
5. Views service logs; scales `web_api` from 2 → 4 replicas.

### Journey B: GitOps production (Stage 5)

1. Platform engineer connects GitHub org repo.
2. Adds `.stowkeep.yml` pointing at `compose.prod.yml`, branch `main`.
3. Sets poll 5m + webhook.
4. Developer merges PR; sync runs; UI shows green synced at commit `abc123`.
5. Failed sync surfaces compose error with link to logs.

### Journey C: Secret rotation (Stage 4)

1. Developer with `secrets.edit` on `/api/database` updates `DB_PASSWORD`.
2. Change creates version 7; diff shows metadata change.
3. Admin with `secrets.read_value` reviews value in UI.
4. Platform engineer redeploys stack; new Swarm secret version mounted.

### Journey D: PR preview (Stage 6)

1. App has previews enabled with domain `{pr}.preview.example.com`.
2. Developer opens PR #42; preview stack deploys automatically.
3. Bot comments preview URL on PR.
4. Developer pushes fix; preview redeploys.
5. PR merged; preview stack deleted within 5 minutes.

---

## 10. Security requirements

Full threat model: [docs/security/threat-model.md](../docs/security/threat-model.md).

1. **Transport:** TLS required in production; HSTS recommended.
2. **Authentication:** Secure session cookies (`HttpOnly`, `SameSite`, `Secure`); API tokens hashed at rest; per-account and per-IP login rate limits.
3. **Authorization:** Deny-by-default; every API route checked against Casbin policy; integration tests assert deny on cross-tenant IDs (IDOR regression suite).
4. **Secrets:** Envelope encryption via `MasterKeyProvider` interface (`EnvKey` v1; KMS impl in Stage 7); each DEK row tagged with `key_id` for non-breaking rotation; sentinel-value tests enforce no plaintext in logs (D-008, D-021).
5. **Audit:** **Hash-chained `audit_events` table from day 1** (D-011); chain integrity verified on startup; never contains secret values.
6. **Input validation:** Compose and YAML parsed with limits (size, depth, anchor expansion) to prevent DoS.
7. **Docker socket:** Quick-start documents the risk of a raw socket mount explicitly; **socket-proxy is the recommended production interim path** (D-012); first-party agent mode replaces it in Stage 7 (D-027).
8. **Dependencies:** Dependabot/Renovate; CI runs `govulncheck`, `gosec`, Trivy on the image, `npm audit --omit=dev`; container images are cosign-signed with SLSA provenance and a Syft SBOM from Stage 0 (D-022).
9. **Backups:** Encrypted with a `BACKUP_KEY` distinct from the MEK (D-016); restore requires `system:restore` permission; HMAC-integrity-checked.

---

## 11. Success metrics

| Metric | Target (6 months post Stage 5) |
|--------|--------------------------------|
| Time to first stack deploy (new install) | < 15 minutes |
| Git sync reliability | > 99% success excluding user compose errors |
| Secret value leakage incidents | 0 |
| Mean time to preview URL (PR open) | < 5 minutes |
| Active self-hosted installs (if telemetry opt-in) | Track via anonymous heartbeat (D-030 — off by default; payload limited to `install_id`/`version`/`db_driver`/`os`/`arch`/`enabled_features`; configurable endpoint URL) |

---

## 12. Decisions (was: Open questions)

All five original open questions have been resolved. Live decisions and deferred questions are tracked in [planning/decisions-todo.md](./decisions-todo.md); the entries below are kept for traceability.

| # | Original question | Resolution |
|---|-------------------|------------|
| 1 | Agent mode in MVP or Stage 7 only? | **Stage 7** (D-027). Interim production path is the Tecnativa-style socket proxy; quick-start documents the raw socket-mount risk. |
| 2 | Embed frontend in Go binary vs nginx sidecar? | **`embed.FS`** (D-026). One binary, one container, atomic FE/API releases — same pattern as Portainer/Grafana/Vault/Argo. |
| 3 | Native secrets only first, or SOPS in Stage 5? | **SOPS in Stage 4** (D-028), alongside native secrets. |
| 4 | GitLab as second provider vs Bitbucket? | **Neither — generic Git remote support in v1** (D-029). Any HTTPS+token or SSH+deploy-key endpoint works. Provider-specific features (PR comments, native webhook signature schemes) are scheduled per-provider when the dependent feature lands. |
| 5 | Optional telemetry opt-in? | **Anonymous opt-in heartbeat, off by default** (D-030). Implementation in Stage 7; posture locked now to prevent accidental collection earlier. |

---

## 13. Appendix: competitive positioning

| Capability | Portainer | Doco-CD | Dokploy | Infisical | Stowkeep |
|------------|-----------|---------|---------|-----------|----------------|
| Swarm UI | Strong | None | Partial | N/A | **Core** |
| Pull GitOps | Limited | **Core** | Partial | N/A | **Core** |
| PR previews | No | No | **Core** | N/A | **Core** |
| Secret versioning + RBAC | Basic | External | Basic | **Core** | **Core** |
| Unified RBAC across features | Medium | N/A | Medium | Strong | **Core** |
| Self-hosted | Yes | Yes | Yes | Yes | **Yes** |

**Positioning statement:** *Stowkeep is the open-source control plane for Docker Swarm teams who want Portainer's cluster visibility, Doco-CD's GitOps, Dokploy's preview environments, and Infisical-grade secrets — in one RBAC-aware platform.*
