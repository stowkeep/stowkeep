# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Stage 2 — Deploy and manage stacks**
  - Compose validation via `pkg/compose` (Compose spec; 1 MiB / depth limits)
  - Stack deploy, remove, service scale, and log streaming (`/api/v1/stacks/*`)
  - Hash-chained deploy audit events (`pkg/audit`, ADR-0005)
  - Feature flag `stack_deploy` gates mutating stack routes
  - RBAC stub (`pkg/rbac`) — admin-only until Stage 3 Casbin enforcement
  - UI: deploy wizard, stack remove confirmation, scale controls, log viewer
  - Permission-builder UI prototype ([docs/prototypes/permission-builder.html](docs/prototypes/permission-builder.html)) per D-010
  - Operator docs: [docs/stacks.md](docs/stacks.md), [docs/audit.md](docs/audit.md)

- **Stage 1 — Swarm read-only dashboard**
  - Local admin auth: bootstrap setup, login/logout, DB-backed sessions (`pkg/auth`)
  - Docker Swarm read API: nodes, services, tasks, stacks (`pkg/docker`, `/api/v1/swarm/*`)
  - Feature flag `swarm_readonly` gates Swarm routes
  - React dashboard: nodes/services/tasks/stacks views, stack detail, task polling (8s), Docker unreachable banner
  - OpenAPI spec updated for Stage 1 endpoints
  - CI: Docker integration test job with Swarm init
  - User docs: [docs/install.md](docs/install.md)

- **Stage 0 application scaffold:** Go API server (`cmd/stowkeep`), React/Vite frontend (`web/`), embedded SPA via `pkg/web` + `embed.FS`
- HTTP endpoints: `GET /healthz`, `GET /readyz`, `GET /api/v1/version`
- Structured logging (`pkg/observability/log`) with request ID correlation and access-log scrubbing tests
- Dual-database layer (`pkg/db`) with SQLite default and PostgreSQL support; goose migrations for hash-chained `audit_events` and `envelope_canary` tables
- `MasterKeyProvider` interface stub (`pkg/secrets`) with `EnvKey` and `KMSProvider` skeleton
- Backup foundation (`pkg/backup`): SQLite `VACUUM INTO` wrapper and PostgreSQL `pg_dump` streaming interface
- Makefile, Dockerfile (multi-stage, distroless), `docker-compose.dev.yml`, OpenAPI stub
- Migrations bundled in container image via `STOWKEEP_MIGRATIONS_DIR` (default `/migrations` in Docker)
- React frontend uses Tailwind CSS; `web/components.json` prepared for shadcn/ui
- CI workflows: `ci.yml`, `release-image.yml` (cosign + SLSA + Syft SBOM), `codeql.yml`
- `GOVERNANCE.md`, `.github/CODEOWNERS`, pre-alpha contributor policy in CONTRIBUTING (closes OPEN-005)

- Planning documents: PRD, tech stack, engineering playbook, testing strategy
- Open source foundation: LICENSE (Apache 2.0), CONTRIBUTING, CODE_OF_CONDUCT, SECURITY
- Documentation: development guide, database guide (SQLite + PostgreSQL), code standards
- Dual database strategy: SQLite for embedded/small installs; PostgreSQL recommended for production
- Production logging strategy: structured JSON via Go `slog` to stdout ([docs/logging.md](docs/logging.md))
- Project decisions tracker ([planning/decisions-todo.md](planning/decisions-todo.md)) covering decided, deferred, and open foundational choices
- Initial STRIDE threat model ([docs/security/threat-model.md](docs/security/threat-model.md)) covering login, authz, deploy, secrets, MEK lifecycle, GitOps, Docker socket, database, backups, and webhooks
- Phase gate framework: mandatory testing, documentation, and code quality bar between stages ([planning/phase-gates.md](planning/phase-gates.md), [AGENTS.md](AGENTS.md))

### Changed

- Tech stack updated to support SQLite embedded mode with Docker volume persistence
- Frontend packaging: assets will be embedded into the Go binary via `embed.FS` (no nginx sidecar)
- Backups elevated to a first-class product feature: scheduled SQLite and PostgreSQL backups to local volume and S3-compatible storage, configurable in the UI
- GitOps Stage 5 narrowed to polling + manual sync; webhook ingestion deferred to a Stage 5.5/6 bolt-on
- Audit log committed to a hash-chained schema from day 1 (cannot be retrofitted later)
- Master key handling committed to a `MasterKeyProvider` interface from Stage 4 day 1 (env-var impl + KMS stub) to keep KMS swap non-breaking
- Coverage gates on `pkg/secrets`, `pkg/rbac`, `pkg/auth` will be enforced in CI from first commit; supply-chain posture (cosign signing, SLSA provenance, Syft SBOM) enabled in Stage 0 CI
- Agent mode confirmed for Stage 7; interim production path uses Tecnativa-style Docker socket proxy with a strong README warning on raw socket mounts
- SOPS support promoted into Stage 4 alongside native secrets
- Git provider strategy: v1 supports any Git endpoint generically (HTTPS+token or SSH+deploy key); provider-specific features (PR comments, native webhook signature validation) deferred and scheduled per-provider when dependent features land
- Telemetry posture locked in (anonymous opt-in heartbeat, off by default, no PII or workload metadata, configurable collector endpoint); implementation deferred to Stage 7
- PRD + key-doc consistency pass: PRD §6/§7/§8/§10/§11/§12, README quick-start socket-mount warning + honest pre-alpha status callout, engineering-and-devops Stage 0 checklist (hardening + supply chain), tech-stack secrets + embed + decision log, testing-strategy coverage hard gates + fuzz + log-leak sentinels + supply-chain assertions, and docs/database driver-selection fix + Backups section all updated to reflect today's locked decisions
- Stage 5.5 added to the roadmap: webhook ingestion (deferred from Stage 5)
- ADRs 0003–0009 drafted and accepted: single-replica control plane, `MasterKeyProvider` interface, hash-chained audit log, Swarm secrets abstraction, RBAC engine + UI-prototype gate, frontend `embed.FS`, telemetry posture
- **Project renamed from "Swarm Operator" to "Stowkeep"** (D-031, [ADR 0010](docs/adr/0010-project-name-stowkeep.md)). Closes OPEN-001. Display name, repo slug (`stowkeep`), Go module path, container image (`ghcr.io/stowkeep/stowkeep`), config manifest (`.stowkeep.yml`), CLI binary, env prefix (`STOWKEEP_*`), default DB filename, and security/conduct contact base domain (`stowkeep.dev`) all updated across PRD, README, all `planning/*` and `docs/*` docs, ADRs 0001–0008, LICENSE copyright, `.env.example`, and `.github/ISSUE_TEMPLATE/*`.
- **`stowkeep` namespace reserved** (OPEN-006 partially closed): `stowkeep.dev` domain registered; GitHub organization and Docker Hub organization created. npm package reservation deferred per D-033 (Go-first project, frontend embedded via `embed.FS`, no JS package planned).
- **Contact inbox consolidated** (D-032, closes OPEN-003): `contact@stowkeep.dev` is the single canonical address for security reports and code-of-conduct concerns pre-`v0.1.0`. Dedicated `security@` and `conduct@` inboxes are a post-`v0.1.0` follow-up. `SECURITY.md` and `CODE_OF_CONDUCT.md` updated.

## [0.0.0] — Planning

Initial repository planning phase. No application release yet.

[Unreleased]: https://github.com/stowkeep/stowkeep/compare/v0.0.0...HEAD
[0.0.0]: https://github.com/stowkeep/stowkeep/releases/tag/v0.0.0
