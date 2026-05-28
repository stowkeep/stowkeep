# Project Decisions Tracker

**Project:** Stowkeep
**Last updated:** 2026-05-25
**Status:** Living document — update whenever a decision is made, deferred, or reopened

This tracks every meaningful project decision: what was decided, what was deliberately deferred, and what is still open. The goal is that no decision is ever "lost" or quietly drifted from in code without a paper trail.

Conventions:

- `D-NNN` — **Decided** (locked in; change requires a new entry that supersedes it)
- `DEF-NNN` — **Deferred** with an explicit revisit trigger
- `OPEN-NNN` — **Open** and needs an answer (with target-by date or stage)
- `SUP-NNN` — **Superseded** (kept for history; points to the new decision)

ADRs in `docs/adr/` capture the deeper reasoning for the most significant decisions; this file is the index and the lightweight ones.

---

## Decided

| ID | Decision | Date | Source / ADR |
|----|----------|------|--------------|
| D-001 | **Backend:** Go 1.23+ | 2026-05-25 | [tech-stack.md](./tech-stack.md) |
| D-002 | **Frontend:** React 19 + TypeScript + Vite + shadcn/ui + Tailwind | 2026-05-25 | [tech-stack.md](./tech-stack.md) |
| D-003 | **API style:** REST + OpenAPI 3.1; SSE/WebSocket for live events | 2026-05-25 | [tech-stack.md](./tech-stack.md) |
| D-004 | **Dual database:** SQLite default + PostgreSQL recommended for production; CI runs both | 2026-05-25 | [ADR 0001](../docs/adr/0001-dual-database-sqlite-postgres.md) |
| D-005 | **Logging:** stdlib `log/slog`, JSON to stdout in prod, text in dev | 2026-05-25 | [ADR 0002](../docs/adr/0002-structured-logging-slog.md) |
| D-006 | **License:** Apache 2.0 (pure OSS, no monetization plan) | 2026-05-25 | [LICENSE](../LICENSE) |
| D-007 | **HA stance (v1):** single-replica control plane; multi-replica + leader election is a Stage 5+ design decision. Document this so future work doesn't silently assume HA. | 2026-05-25 | ADR 0003 to be written |
| D-008 | **Master key handling:** ship `MasterKeyProvider` interface in Stage 4 day 1 with `EnvKey` impl and stub `KMSProvider`. Each DEK row carries a `key_id` so future rotation/KMS swap is non-breaking. | 2026-05-25 | ADR 0004 to be written |
| D-009 | **Swarm secrets abstraction:** Stowkeep stores its own versioned secret records (with a Stowkeep version ID). On deploy we materialize a uniquely-named Swarm secret (e.g. `appname_dbpassword_v7`) and update services to reference it. Old Swarm secrets are garbage-collected once no service references them. Rotation = rolling update; new container with new secret comes up healthy before traffic shifts. | 2026-05-25 | [ADR 0006](../docs/adr/0006-swarm-secrets-abstraction.md) |
| D-010 | **RBAC engine:** stick with Casbin. **Mitigation:** Stage 2 must ship a paper/HTML mockup of the permission-builder UI as part of the PR that introduces auth. If the only realistic admin interface ends up being the raw Casbin CSV, treat that as an abort signal and revisit. | 2026-05-25 | ADR 0007 to be written |
| D-011 | **Audit log:** hash-chained from day 1 (each `audit_events` row stores `prev_hash` and `row_hash = sha256(prev_hash \|\| canonical(row))`). Cannot be retrofitted after data exists. | 2026-05-25 | ADR 0005 to be written |
| D-012 | **Docker connectivity:** direct socket mount in quick-start (with strong warning), Tecnativa-style socket proxy as documented production path, agent mode (Portainer pattern) targeted for Stage 7. | 2026-05-25 | (note in README + future agent ADR) |
| D-013 | **Docker connectivity UX:** prominent UI banner whenever the Docker socket is unreachable; degrade to read-only mode where feasible. | 2026-05-25 | Update PRD §6.1 |
| D-014 | **Webhooks (Stage 5):** deferred. Stage 5 ships polling + manual "Sync now" button only. Webhooks come as a Stage 5.5/6 bolt-on with HMAC validation, replay protection, and idempotency from day 1. | 2026-05-25 | Update PRD §6.3 |
| D-015 | **Preview environments:** stay in Stage 6 but **lower priority overall**; not a v1 must-ship. Routing helper (Traefik labels) is part of the same stage when it lands. | 2026-05-25 | Update PRD §6.6 / §8 |
| D-016 | **Backups (NEW P0/P1 product requirement):** built-in scheduled backups for **both** SQLite (via `sqlite3` backup API, not file copy) **and** PostgreSQL (via `pg_dump`), configurable via UI, targeting local volume **and** S3-compatible object store. One-click restore. This is a first-class user-facing feature, not Stage 7 polish. | 2026-05-25 | New PRD section under §6.8 |
| D-017 | **Compose-vs-Swarm linter:** deliberately **not** building one for v1. Target users are technical and accept that `docker stack deploy` silently ignores some Compose v2 keys. Revisit when first non-technical user complaint arrives. | 2026-05-25 | DEF-006 |
| D-018 | **Multi-arch images:** `linux/amd64` + `linux/arm64` required in CI. `linux/arm/v7` added as best-effort if BuildKit `platforms` line is the only change needed. | 2026-05-25 | Update engineering-and-devops.md §4 |
| D-019 | **Coverage gates:** hard CI gate on `pkg/secrets`, `pkg/rbac`, `pkg/auth` (≥90% line cov) from the first commit in those packages. No "aspirational" coverage targets. | 2026-05-25 | Update testing-strategy.md §5 |
| D-020 | **Fuzz tests:** required for compose parser, RBAC evaluator, secret path resolver, and (when added) webhook payload parser. Land alongside the parser in the same PR. | 2026-05-25 | Update testing-strategy.md |
| D-021 | **Log-leak detection:** every `slog` invocation in a security-sensitive path covered by a sentinel-value test that runs the workload and greps captured log output. CI-enforced. | 2026-05-25 | Update testing-strategy.md |
| D-022 | **CI supply-chain posture:** cosign image signing + SLSA provenance attestation + Syft SBOM publication from Stage 0 (not Stage 7). All Actions pinned to SHAs. | 2026-05-25 | Update engineering-and-devops.md §8 |
| D-023 | **CODEOWNERS:** ship in Stage 0 even with one maintainer; security-sensitive paths (`pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/`) all owned by the maintainer with explicit "two-eyes-on-merge" comment. | 2026-05-25 | New file `.github/CODEOWNERS` |
| D-024 | **Governance:** single maintainer (BDFL) until v0.5.0; documented in `GOVERNANCE.md`. README should reflect "currently single-maintainer; not actively seeking contributors yet" until then. | 2026-05-25 | New file `GOVERNANCE.md` + README update |
| D-025 | **Vault scope:** **not** a Vault replacement. Vault integration is **not** a v1 goal. Future option: `SecretBackend` interface in Stage 7+ with a Vault implementation for teams that mature into Vault. Today's user is on Swarm precisely because they're not running Vault. | 2026-05-25 | Update PRD §3 non-goals |
| D-026 | **Frontend packaging:** embed built static assets into the Go binary via `embed.FS`. One-binary, one-container deploy story stays intact. No nginx sidecar. Rationale: matches the single-binary pitch; FE/API release atomicity; standard pattern in Portainer/Grafana/Vault/Argo. Single nuance: every UI change requires a Go rebuild in CI (transparent), dev unaffected (Vite dev server proxies `/api`). | 2026-05-25 | Closes PRD §12 open Q2 |
| D-027 | **Agent mode timing:** Stage 7. MVP uses direct socket mount with warnings; documented socket-proxy is the production-recommended interim path. Rationale: agent mode is a 4–6 week effort (mTLS PKI, enrollment, separate `cmd/agent`, long-lived bidi connection, capability scoping) and adds zero value for the v1 single-node homelab user. Only matters once multi-node / multi-cluster lands (already Stage 7). | 2026-05-25 | Closes PRD §12 open Q1 |
| D-028 | **SOPS support:** ship in Stage 4 (alongside native secrets), not Stage 5+. Lets users with existing SOPS-encrypted YAML in Git adopt without a parallel secrets store. | 2026-05-25 | Closes DEF-009 / PRD §12 open Q3 |
| D-029 | **Git provider strategy:** v1 supports **any Git endpoint** generically — HTTPS+token or SSH+deploy key. Works for GitHub, GitLab (cloud or self-hosted), Gitea, Forgejo, Bitbucket, AWS CodeCommit, Azure DevOps, self-hosted `git+ssh`. **Provider-specific features** (PR comments, native webhook signature validation, API-driven deploy-key provisioning, status checks) are deferred and scheduled per-provider when the dependent feature (webhooks, PR comments, previews) lands. | 2026-05-25 | Closes DEF-014 / PRD §12 open Q4 |
| D-030 | **Telemetry posture (locked now, implemented Stage 7):** opt-in only, off by default. Anonymous heartbeat once per week if enabled. **Collected:** `install_id` (locally-generated UUID, never reset, never user-tied), `version`, `db_driver`, `os`, `arch`, `enabled_features`. **Never collected:** project names, repo URLs, user emails or IDs, stack names, secret names, node counts, IP addresses (stripped server-side), compose file content, Docker/Swarm metadata. Configurable endpoint URL so users can self-host or `/dev/null`. README documents exactly what is sent. | 2026-05-25 | Closes OPEN-002 / DEF-013 / PRD §12 open Q5 |
| D-033 | **npm package reservation: deferred.** Stowkeep is Go-first; the frontend is embedded into the Go binary via `embed.FS` (D-026 / ADR 0008), so there is no public JS package to publish. The squat risk on an unknown project is small, and if a JS SDK or CLI is ever needed, a scoped name (`@stowkeep/sdk`, `@stowkeep/cli`) avoids the bare-slot question entirely. Revisit only if (a) a JS package is genuinely planned, or (b) Stowkeep reaches noticeable public visibility and squat risk rises. | 2026-05-25 | OPEN-006 |
| D-032 | **Single contact inbox for pre-`v0.1.0`.** `contact@stowkeep.dev` is the canonical address for both security reports and code-of-conduct concerns until after `v0.1.0`, at which point dedicated `security@stowkeep.dev` and `conduct@stowkeep.dev` inboxes will be provisioned. Rationale: avoids multi-inbox overhead during single-maintainer pre-alpha while still providing a live address. Closes OPEN-003. | 2026-05-25 | SECURITY.md, CODE_OF_CONDUCT.md |
| D-031 | **Project name: Stowkeep.** Locks in the display name, repo slug (`stowkeep`), Go module path (`github.com/stowkeep/stowkeep`), container image (`ghcr.io/stowkeep/stowkeep`), config manifest (`.stowkeep.yml`), CLI binary (`stowkeep`), env prefix (`STOWKEEP_*`), default DB filename (`stowkeep.db`), and security/conduct contact base domain (`stowkeep.dev`). Selection criteria: (1) all four key surfaces clean — `github.com/stowkeep`, `hub.docker.com/u/stowkeep`, `npmjs.com/package/stowkeep`, no software-namespace search hits; (2) no trademark conflict in software classes (Class 9/42); (3) no phonetic collision with existing tools (unlike "Sailwind"/Tailwind); (4) etymologically fits the product — "stow" = "keep safely", "keep" = both maintain and a castle's fortified inner tower. | 2026-05-25 | [ADR 0010](../docs/adr/0010-project-name-stowkeep.md); closes OPEN-001 |
| D-034 | **Bootstrap config:** optional `.env` via godotenv (dev); `STOWKEEP_*_FILE` for production secret mounts; MEK input normalization in `pkg/secrets/keyinput` (passphrase/hex/base64/file). No bootstrap MEK wizard. UI reads feature flags from `/api/v1/version` only. | 2026-05-27 | [ADR 0011](../docs/adr/0011-bootstrap-config-and-mek-input.md) |
| D-035 | **Secret inject modes:** default Swarm file mount; opt-in `inject: env` for legacy images; `inject: wrapper` deferred to flavor wall. | 2026-05-27 | [ADR 0006](../docs/adr/0006-swarm-secrets-abstraction.md) § inject modes |

---

## Deferred (with explicit revisit triggers)

| ID | Question | Revisit when | Notes |
|----|----------|--------------|-------|
| DEF-001 | MEK rotation procedure (re-wrap DEKs without big-bang re-encryption) | Stage 4 kickoff | Per-DEK `key_id` from D-008 makes this possible without rewrite. Design + runbook needed before any real secrets exist. |
| DEF-002 | Casbin policy authoring UX prototype | **Shipped Stage 2** — [docs/prototypes/permission-builder.html](../docs/prototypes/permission-builder.html). Outcome: Casbin CSV mappable to guided UI; proceed with Casbin in Stage 3 unless maintainer review rejects mockup. |
| DEF-003 | `go-git` limits & repo size policy | Stage 5 kickoff, or first repo > ~100 MB | Document max repo size, no LFS, treeless clone strategy. Switch to shelling out to `git` if `go-git` proves insufficient for the typical infra-repo use case. |
| DEF-004 | Preview environment design (routing strategy, secret scoping, TTL semantics) | Post-Stage 5, before Stage 6 begins | Per D-015. |
| DEF-005 | Webhook ingestion design (per-provider HMAC, replay window, idempotency key, debounce) | When webhooks are scheduled (post initial Stage 5) | Per D-014. |
| DEF-006 | Compose-vs-Swarm validation linter | First user complaint, or pre-1.0 | Per D-017. |
| DEF-007 | Multi-cluster endpoint registry (manage 2+ Swarms from one UI) | Stage 7 / first user request | PRD SW-11 P2. |
| DEF-008 | Agent mode design (TLS bootstrap, agent enrollment, capability scoping) | Stage 7 kickoff | Per D-027. |
| ~~DEF-009~~ | ~~SOPS support~~ | ~~Stage 4 kickoff~~ | **Closed by D-028 (ship in Stage 4).** |
| DEF-010 | OIDC SSO (GitHub/Google/GitLab) | Stage 7 | Currently AUTH-02 / Stage 7. Re-evaluate when first multi-user team adopter appears. |
| DEF-011 | 2FA / TOTP / WebAuthn | Stage 7 | Docker socket access is essentially root; WebAuthn worth considering early for admin accounts. |
| DEF-012 | Multi-replica HA + leader election (PG advisory lock or Raft-lite) | Stage 5 | Per D-007. |
| ~~DEF-013~~ | ~~Telemetry / anonymous heartbeat opt-in~~ | ~~Before first public release~~ | **Closed by D-030 (opt-in, off by default, implementation in Stage 7).** |
| ~~DEF-014~~ | ~~Git provider order beyond GitHub~~ | ~~Stage 5 mid~~ | **Closed by D-029 (generic-Git v1; provider-specific features scheduled per dependent feature).** |
| DEF-015 | Vault `SecretBackend` integration | Stage 7+ / first request | Per D-025. |
| DEF-016 | i18n / localization | Post-1.0 | Out of scope for v1; document the decision in README. |
| DEF-017 | Accessibility (a11y) testing pass | Stage 2 | shadcn/ui gives a head start; explicit Playwright a11y assertions before v1.0. |

---

## Open (needs an answer soon)

| ID | Question | Target-by | Why it can't wait |
|----|----------|-----------|-------------------|
| ~~OPEN-001~~ | ~~Project name + domain registration + GitHub org reservation~~ | ~~Before any public publishing of the repo~~ | **Closed by D-031 (name = Stowkeep).** Domain + GitHub org + Docker Hub + npm handle reservation still required as a Stage 0 task before any public publishing — tracked under OPEN-006. |
| ~~OPEN-002~~ | ~~Telemetry stance~~ | ~~Before v0.1.0 tag~~ | **Closed by D-030.** |
| ~~OPEN-003~~ | ~~`SECURITY.md` and `CODE_OF_CONDUCT.md` contact addresses~~ | ~~Before publishing repo~~ | **Closed 2026-05-25 by D-032.** Both files now route to `contact@stowkeep.dev`, which is live. Dedicated `security@` and `conduct@` inboxes are a post-`v0.1.0` follow-up tracked in OPEN-006. |
| OPEN-004 | **README "current status" callout.** Match the actual reality: pre-alpha, single maintainer, not actively seeking contributors yet. | Before publishing repo | Today the docs imply an active OSS project. |
| ~~OPEN-005~~ | ~~Source-available vs OSS posture during pre-alpha~~ | ~~Before first external interest~~ | **Closed 2026-05-25:** pre-alpha contributor policy documented in [CONTRIBUTING.md](../CONTRIBUTING.md) and [GOVERNANCE.md](../GOVERNANCE.md). |
| OPEN-006 | **Reserve the `stowkeep` namespace.** Status as of 2026-05-25: ✅ `stowkeep.dev` registered; ✅ GitHub org created; ✅ Docker Hub org created; ⏸️ npm package — **deferred per D-033** (Go-first project; frontend embedded via `embed.FS`; no JS package planned). Revisit when/if Stowkeep ever ships a `@stowkeep/*` JS SDK or gains notable public visibility. ⏸️ Dedicated `security@` and `conduct@` inboxes — deferred post-`v0.1.0`; interim address is `contact@stowkeep.dev`. ⏸️ Defensive `stowkeep.app` / `stowkeep.com` — optional, register only if cheap and convenient. | Reservations done; remaining items optional/deferred | Done. |

---

## Superseded

*(none yet)*

---

## How to use this file

1. **Made a decision in a chat, issue, or PR?** Add a `D-NNN` entry the same day with date and a link. If the decision is significant, also open an ADR.
2. **Hit a question you can't answer today?** Add `DEF-NNN` with an explicit revisit trigger — *never* a vague "later".
3. **Disagreeing with a past decision?** Add a new `D-NNN` that supersedes the old, and mark the old `SUP-NNN`. Do not edit the old entry in place.
4. **Reviewing the project before a release?** Walk all `OPEN-*` and `DEF-*` rows whose revisit trigger has fired.
