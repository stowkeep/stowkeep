# Phase Gates — Quality Bar Between Stages

**Project:** Stowkeep  
**Last updated:** 2026-05-25  
**Audience:** Maintainers, junior developers, and automated agents (LLMs) writing code

---

## Hard rule

> **Do not start Stage N+1 until Stage N's phase gate is signed off.**

A phase gate is a mandatory checkpoint covering **testing**, **documentation**, and **code quality**. Feature work that lands on `main` during a stage must still satisfy per-PR rules in [open-source-standards.md](./open-source-standards.md). The phase gate is the **stage-level** proof that the cumulative work is shippable, documented, and safe to build on.

**Agents and contributors:** If you are asked to implement the next stage but the previous gate checklist is incomplete, **stop** and complete or report blockers on the gate — do not skip ahead.

---

## Three pillars (every stage)

Every phase gate evaluates the same three pillars. All must pass.

### 1. Testing

| Requirement | Evidence |
|-------------|----------|
| CI green on `main` | Latest `main` workflow run passed (lint, test, build, DB matrix) |
| Stage test plan executed | Checklist in [testing-strategy.md](./testing-strategy.md) §8 for this stage — all items checked |
| Coverage not regressed | No decrease in critical packages vs previous stage (secrets ≥90%, rbac ≥90% when those packages exist) |
| Security tests | No known gaps: secret leakage, auth header logging, RBAC bypass for new routes |
| Manual smoke | Maintainer (or designated reviewer) ran stage manual checklist |

### 2. Documentation

| Requirement | Evidence |
|-------------|----------|
| User/docs updated | Every user-visible behavior documented in `docs/` (same PR as feature, or gate PR) |
| In-code docs | New exported Go/TS APIs have godoc/TSDoc; packages have boundary comments |
| CHANGELOG | `[Unreleased]` section lists all user-visible changes for the stage |
| ADRs | New architectural decisions have ADRs in `docs/adr/`; existing ADRs updated if assumptions changed |
| Threat model | `docs/security/threat-model.md` updated if ingress, auth, or trust boundaries changed |
| Development guide | `docs/development-guide.md` reflects new commands, env vars, or workflows |

### 3. Code quality

| Requirement | Evidence |
|-------------|----------|
| Lint clean | `golangci-lint`, `eslint`, `tsc --noEmit` — zero new warnings on `main` |
| Review complete | All stage PRs reviewed; security paths (`pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/`) have CODEOWNER approval where required |
| No TODO debt | No unresolved `TODO` in merged stage code without a linked GitHub issue |
| Dependencies | No critical CVEs in container or direct deps (Trivy/govulncheck/npm audit) |
| Logging & errors | Follows [docs/logging.md](../docs/logging.md); errors wrapped with context at boundaries |
| Migrations | Apply cleanly on **SQLite and PostgreSQL**; down migrations documented or N/A with justification |

---

## Gate sign-off process

1. **Open a gate issue** — Title: `Phase gate: Stage N complete` (or use milestone).
2. **Copy the stage checklist** from §3 below into the issue.
3. **Fill evidence links** — CI run URL, PR list, doc paths, manual test notes.
4. **Reviewer sign-off** — At least one maintainer checks all three pillars.
5. **Record completion:**
   - Update [CHANGELOG.md](../CHANGELOG.md) — move items to `v0.x.0` release section
   - Tag optional pre-release: `v0.2.0-stage1` or semver minor at maintainer discretion
   - Note in [decisions-todo.md](./decisions-todo.md) walk if any deferred items affect next stage
6. **Announce** — Stage N+1 work may begin.

Gate issues are **blockers** for starting the next stage in the same branch of work.

---

## Universal PR checklist (every PR within a stage)

Use this on **every** PR. Same content lives in [.github/pull_request_template.md](../.github/pull_request_template.md).

- [ ] `make lint && make test` passes locally
- [ ] Tests added/updated for behavior changes
- [ ] SQLite + PostgreSQL tested if DB/migrations touched
- [ ] Godoc/TSDoc on new exported APIs
- [ ] User docs + CHANGELOG updated if user-visible
- [ ] Logging follows [docs/logging.md](../docs/logging.md)
- [ ] No secrets/credentials in diff
- [ ] Threat model + ADR updated if architecture/security changed

---

## Stage gate checklists

### Stage 0 — Foundation

**Testing**

- [ ] `go test ./... -short -race` passes
- [ ] Vitest scaffold passes (`web/`, when present)
- [ ] Migrations apply on SQLite and PostgreSQL in CI
- [ ] Docker image builds; `curl /healthz` smoke test in CI
- [ ] Log handler test: no secrets in default middleware output

**Documentation**

- [ ] All planning docs present and cross-linked
- [ ] `docs/development-guide.md`, `code-standards.md`, `database.md`, `logging.md` accurate
- [ ] ADRs 0001–0010 indexed in `docs/adr/README.md`
- [ ] `AGENTS.md` and this file (`phase-gates.md`) present

**Code quality**

- [ ] `.golangci.yml` and ESLint config enforced in CI
- [ ] GitHub Actions pinned to SHAs
- [ ] Cosign + SBOM + SLSA provenance on release-image workflow (per D-022)
- [ ] `CODEOWNERS` covers security-sensitive paths
- [ ] `make dev` documented and verified on clean clone

---

### Stage 1 — Swarm read-only dashboard

**Testing**

- [ ] Docker integration tests (list nodes/services/stacks) — skip with `-short` when no Docker
- [ ] Auth bootstrap tests (first admin, login, session)
- [ ] HTTP middleware tests (`request_id`, access log shape)
- [ ] Manual: connect socket, lists match CLI

**Documentation**

- [ ] `docs/install.md` or README section: first-run + socket mount
- [ ] OpenAPI spec covers all Stage 1 endpoints
- [ ] Threat model updated: Docker socket trust boundary

**Code quality**

- [ ] All new API routes behind auth (except `/healthz`, static assets)
- [ ] Feature flag `swarm_readonly` documented
- [ ] No Docker API response bodies logged at info level

---

### Stage 2 — Deploy and manage stacks

**Testing**

- [ ] Compose validation unit tests (valid/invalid fixtures)
- [ ] Integration: deploy + remove minimal stack on Swarm
- [ ] Scale service integration test
- [ ] Audit chain: deploy event written and verifies on read
- [ ] Permission-builder **prototype** reviewed (D-010) — outcome recorded in ADR or decisions-todo
- [ ] Manual checklist §8 Stage 2 complete

**Documentation**

- [ ] Stack deploy documented for operators
- [ ] Compose validation errors documented (common fixes)
- [ ] Audit log format documented

**Code quality**

- [ ] Deploy paths use context timeouts
- [ ] Confirm dialogs for destructive actions in UI
- [ ] RBAC hooks stubbed or enforced consistently (prep for Stage 3)

---

### Stage 3 — RBAC and audit

**Testing**

- [ ] Casbin table tests for every API route (allow + deny)
- [ ] IDOR regression suite (stack IDs, user resources)
- [ ] Audit hash-chain integrity tests
- [ ] API token scope tests
- [ ] Manual: viewer role cannot deploy

**Documentation**

- [ ] `docs/rbac.md` — roles, permissions, condition keys
- [ ] Permission builder UI help text
- [ ] ADR-0007 assumptions validated or updated

**Code quality**

- [ ] Deny-by-default on all routes
- [ ] No permission check only in UI (server enforced)
- [ ] CODEOWNER review on all `pkg/rbac` changes

---

### Stage 4 — Secrets

**Testing**

- [ ] Envelope encrypt/decrypt roundtrip tests
- [ ] `TestSecretNeverInLogs` and audit-no-value tests
- [ ] describe vs read_value permission tests
- [ ] Swarm secret materialization integration test
- [ ] MEK rotation procedure tested (documented path)
- [ ] `STOWKEEP_*_FILE` resolution tests (env vs file precedence, trim newline, missing file)
- [ ] MEK input normalization tests (`pkg/secrets/keyinput`: passphrase, hex, base64, raw file bytes)
- [ ] Startup fails clearly on invalid MEK material
- [ ] SOPS decrypt test fixture (if D-028 in stage)
- [ ] Manual: DB dump has no plaintext secrets

**Documentation**

- [ ] `docs/secrets.md` — encryption, rotation, Swarm mapping, inject modes
- [ ] `docs/configuration.md` — bootstrap env vars and `STOWKEEP_*_FILE`
- [ ] `docs/secrets-rotation.md` — MEK rotation runbook
- [ ] MasterKeyProvider documented (ADR-0004, ADR-0011)

**Code quality**

- [ ] Plaintext never in logs, audit payloads, or error traces (CI enforced)
- [ ] CODEOWNER review on all `pkg/secrets` changes
- [ ] Threat model updated for secrets at rest and in transit

---

### Stage 5 — GitOps sync

**Testing**

- [ ] `.stowkeep.yml` parser table tests
- [ ] Git fixture integration: clone, detect SHA, reconcile
- [ ] Failed compose → no partial deploy; status error
- [ ] Single-flight reconcile test (no concurrent double deploy)
- [ ] Drift detection unit tests
- [ ] Manual: poll sync deploys on repo change

**Documentation**

- [ ] `docs/stowkeep-yml.md` — manifest reference
- [ ] Git credential setup (HTTPS/SSH)
- [ ] Sync status and troubleshooting

**Code quality**

- [ ] Repo size cap enforced with clear error
- [ ] Shallow clone used where safe
- [ ] All sync actions audited

---

### Stage 5.5 — Webhook ingestion

**Testing**

- [ ] Signature verification tests (GitHub/GitLab/Gitea vectors)
- [ ] Replay rejection (>5 min old)
- [ ] Idempotency tests
- [ ] Rate limit tests
- [ ] Manual: push triggers sync within seconds

**Documentation**

- [ ] Webhook setup per provider
- [ ] Security notes: constant-time compare, IP limits

**Code quality**

- [ ] No webhook secret logged
- [ ] Failed verification audited with reason

---

### Stage 6 — Preview environments

**Testing**

- [ ] PR open → preview stack created (integration or E2E)
- [ ] PR close → teardown within TTL
- [ ] Max preview limit enforced
- [ ] Playwright E2E-06 or equivalent
- [ ] Manual checklist §8 Stage 6 complete

**Documentation**

- [ ] Preview configuration guide
- [ ] Domain template examples

**Code quality**

- [ ] Preview stacks isolated (naming, networks)
- [ ] Cleanup on failure paths (no orphaned stacks)

---

### Stage 7 — Hardening and enterprise

**Testing**

- [ ] OIDC flow integration test (or documented manual matrix)
- [ ] Agent mode mTLS tests (when shipped)
- [ ] Prometheus `/metrics` scrape test
- [ ] Load smoke per testing-strategy §10 (when applicable)

**Documentation**

- [ ] HA runbooks, backup/restore, lost-MEK recovery
- [ ] Agent mode deployment guide
- [ ] Telemetry opt-in documentation (ADR-0009)

**Code quality**

- [ ] Security review of SSO and agent TLS
- [ ] No critical CVEs in release image

---

## For LLMs and automated agents

When implementing Stowkeep:

1. **Read first:** [AGENTS.md](../AGENTS.md) → this file → stage section in [PRD.md](./PRD.md)
2. **Work within one stage** unless the gate issue for the previous stage is closed
3. **Every PR must pass** the universal PR checklist (§2)
4. **Before claiming a stage complete:** run the stage checklist (§3), open the gate issue, do not start next stage features
5. **Never skip tests** to "move faster" — gate failure blocks the project
6. **Document in the same PR** as code unless the change is docs-only
7. **When unsure:** add tests and docs rather than omitting them

**Stop signals — do not proceed, report to user:**

- Previous stage gate checklist has unchecked items
- CI is red on `main`
- Migration fails on one of SQLite or PostgreSQL
- No godoc on new exported Go APIs
- User-visible change without CHANGELOG entry

---

## Related documents

| Document | Role |
|----------|------|
| [open-source-standards.md](./open-source-standards.md) | Per-PR quality bar |
| [testing-strategy.md](./testing-strategy.md) | Test pyramid, CI gates, manual checklists |
| [PRD.md](./PRD.md) | Stage deliverables and acceptance criteria |
| [docs/code-standards.md](../docs/code-standards.md) | Naming, comments, logging |
| [AGENTS.md](../AGENTS.md) | Agent entrypoint |
