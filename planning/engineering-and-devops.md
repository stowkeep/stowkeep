# Engineering & DevOps Playbook

**Project:** Stowkeep  
**Last updated:** 2026-05-25

This document covers product naming, repository setup, branching strategy, CI/CD, container publishing, and release process.

---

## 1. Product naming

### Name: **Stowkeep** (locked — D-031, [ADR 0010](../docs/adr/0010-project-name-stowkeep.md))

| Attribute | Value |
|-----------|-------|
| Display name | Stowkeep |
| Repository | `stowkeep` |
| Go module | `github.com/stowkeep/stowkeep` |
| Container image | `ghcr.io/stowkeep/stowkeep` |
| Config manifest | `.stowkeep.yml` |
| CLI binary | `stowkeep` |
| Env prefix | `STOWKEEP_*` |
| Default SQLite filename | `stowkeep.db` |
| Default Postgres database | `stowkeep` |
| Default Docker volume | `stowkeep-data` |
| Security contact | `security@stowkeep.dev` (pending OPEN-006) |
| Conduct contact | `conduct@stowkeep.dev` (pending OPEN-006) |

**Rationale:** "Stow" = to put something in a place where it can be kept safely; "keep" = both *maintain* and (medieval) a castle's fortified inner tower. Reads honestly for a control plane that holds compose stacks, secrets (encrypted), GitOps state, audit logs, and backups. All four key technical surfaces (GitHub org, Docker Hub org, npm, `.dev` domain) verified available on 2026-05-25; no software-namespace search hits; no trademark conflict in software classes; no phonetic collision with widely-used dev tools. Full selection criteria and rejected candidates: [ADR 0010](../docs/adr/0010-project-name-stowkeep.md).

**Reservation status (OPEN-006):** GitHub org, Docker Hub org, npm package, and `stowkeep.dev` domain must be held before any public repo push. Name is *selected* but not fully *locked* until those are reserved.

### Naming conventions in code

| Item | Convention | Example |
|------|------------|---------|
| HTTP routes | kebab-case | `/api/v1/gitops/apps` |
| Go packages | short, lowercase | `pkg/gitops`, `pkg/secrets` |
| React components | PascalCase | `StackDeployWizard.tsx` |
| DB tables | snake_case plural | `secret_versions`, `audit_events` |
| Feature flags | snake_case | `gitops`, `previews` |
| Git branches | `{type}/{short-description}` | `feat/stack-deploy` |

---

## 2. GitHub repository setup

### Organization and visibility

- **Recommended:** GitHub organization (e.g. `stowkeep` or your org) with **public** repo for open source.
- **License:** Apache License 2.0 — see [LICENSE](../LICENSE)

### Initial repository checklist

- [ ] Create repo `stowkeep` with description: *Self-hosted control plane for Docker Swarm — GitOps, secrets, RBAC, and preview environments.*
- [ ] Add topics: `docker`, `docker-swarm`, `gitops`, `devops`, `self-hosted`, `secrets-management`
- [ ] Enable **Issues** and **Discussions** (optional)
- [ ] Branch protection on `main` (see §4)
- [ ] Add `SECURITY.md`, `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `CHANGELOG.md`
- [ ] Add `docs/` contributor guides (development, code standards, database)
- [ ] Add issue templates and PR template
- [ ] Add `.env.example`
- [ ] Configure **Dependabot** for Go modules, npm, GitHub Actions
- [ ] Configure **secret scanning** (GitHub Advanced Security if available)

### Required GitHub secrets (Actions)

| Secret | Purpose |
|--------|---------|
| `GHCR_TOKEN` or use `GITHUB_TOKEN` | Push container images to GHCR |
| `CODECOV_TOKEN` (optional) | Coverage uploads |

No Docker Hub required if using GHCR exclusively.

---

## 3. Trunk-based development

### Principles

1. **`main` is always deployable** — protected, CI-green, releasable at any commit.
2. **Short-lived branches** — target merge within 1–3 days.
3. **Small PRs** — one logical change; map to PRD stages/sub-features.
4. **Feature flags** — incomplete features merged dark; enabled per environment.
5. **No long-lived `develop` branch** — avoids merge drift and release complexity.

### Branch workflow

```
main ─────●─────●─────●─────●─────●───── (always releasable)
           \   /       \   /
            ●─●         ●─●
         feat/x      fix/y
```

| Branch prefix | Use |
|---------------|-----|
| `feat/` | New functionality |
| `fix/` | Bug fixes |
| `chore/` | Tooling, deps, CI |
| `docs/` | Documentation only |
| `refactor/` | No behavior change |

### Branch protection rules (`main`)

- Require pull request before merge
- Require 1 approval (2 when team grows)
- Require status checks: `lint`, `test`, `build`
- Require branch up to date before merge
- No force push
- Restrict who can push (maintainers only)

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(stacks): add compose validation on deploy
fix(auth): expire sessions after idle timeout
chore(ci): pin golangci-lint version
docs(prd): add stage 5 acceptance criteria
```

### Feature flags

Store in database or config (`STOWKEEP_FEATURES=gitops,previews`). Default off until stage complete. Allows trunk merges without exposing incomplete UX.

---

## 4. CI/CD pipeline (GitHub Actions)

### Workflow overview

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | PR + push to `main` | Lint, test, build verification |
| `release-image.yml` | Push to `main` | Build and push production container to GHCR |
| `release.yml` | Tag `v*` | Create GitHub Release + semver image tags |
| `codeql.yml` | Weekly + PR | Security analysis |

### `ci.yml` — every PR

```yaml
# Pseudocode structure — implement in Stage 0
jobs:
  lint-go:
    - golangci-lint
  lint-web:
    - eslint + tsc --noEmit
  test-go:
    - go test ./... -race -coverprofile=coverage.out
    - services: postgres (testcontainers or service container)
  test-web:
    - vitest run
  build:
    - go build ./...
    - npm run build (web)
  docker-build:
    - docker build --target test .
    - # verify image starts and /healthz responds
```

**PR policy:** All jobs must pass before merge.

### `release-image.yml` — trunk builds

On every merge to `main`:

1. Build multi-arch image — **`linux/amd64` and `linux/arm64` required**; `linux/arm/v7` added as best-effort when BuildKit `platforms` line is the only change needed (D-018)
2. Push to `ghcr.io/stowkeep/stowkeep:main`
3. Push immutable SHA tag: `ghcr.io/stowkeep/stowkeep:sha-<short-sha>`
4. **Sign the image with cosign** using OIDC keyless signing (no long-lived keys) (D-022)
5. **Attach SLSA provenance** via `actions/attest-build-provenance`
6. **Generate and publish a Syft SBOM** as a cosign attestation
7. Cache BuildKit layers for speed

**Consumers:** Teams who track `main` for latest; not semver until tagged. All images verifiable with `cosign verify` + `cosign verify-attestation`.

### `release.yml` — semver releases

On tag push `v1.2.3`:

1. Run full CI suite
2. Build and push:
   - `ghcr.io/stowkeep/stowkeep:1.2.3`
   - `ghcr.io/stowkeep/stowkeep:1.2`
   - `ghcr.io/stowkeep/stowkeep:1`
   - `ghcr.io/stowkeep/stowkeep:latest`
3. Generate GitHub Release notes from conventional commits
4. Same supply-chain attestations as trunk builds: cosign signature + SLSA provenance + Syft SBOM (all required, all keyless-OIDC) (D-022)

### PR container images (optional, Stage 2+)

On PR open/sync:

- Build image tagged `ghcr.io/stowkeep/stowkeep:pr-<number>`
- Comment on PR with pull command
- Delete image on PR close

Useful for dogfooding preview flow before Stage 6.

### Dockerfile strategy

```dockerfile
# Stage: web-build — node builds static assets
# Stage: go-build — compile with embedded or copied static
# Stage: runtime — distroless/static or alpine + non-root user
# Expose: 8080
# USER: non-root (uid 65532)
# ENTRYPOINT: /stowkeep
```

---

## 5. Versioning and releases

### Semantic versioning

- **MAJOR:** Breaking API, migration, or config changes
- **MINOR:** New stage features (backward compatible)
- **PATCH:** Bug fixes

Pre-1.0 (`v0.x.y`): API may change; document in CHANGELOG.

### Release cadence

| Phase | Cadence |
|-------|---------|
| Stage 0–2 (alpha) | Tag `v0.1.0`, `v0.2.0` at stage completion |
| Stage 3–5 (beta) | Bi-weekly minor if changes warrant |
| Post 1.0 | Monthly minors; patches as needed |

### CHANGELOG

Maintain `CHANGELOG.md` following [Keep a Changelog](https://keepachangelog.com/). Auto-generate release notes from conventional commits.

---

## 6. Local development

### Makefile targets (proposed)

```makefile
dev              # API + web hot reload (SQLite default)
dev-postgres     # API + web with PostgreSQL
test             # all unit tests (Go + web)
test-integration # integration tests including DB matrix
lint             # golangci-lint + eslint
migrate          # goose up (respects DATABASE_DRIVER)
build            # production binary + web bundle
docker-build     # local image
```

### `docker-compose.dev.yml`

Services:

- `postgres:16` (optional — use `make dev-postgres`)
- `redis:7` (when needed, Stage 5+)
- `api` — mount socket read-only for Swarm dev; SQLite volume or Postgres URL
- `web` — Vite dev server with proxy to API

Default local dev uses **SQLite** at `./.data/dev.db` — no compose services required for `make dev`.

---

## 7. Deployment of Stowkeep itself

### Profile A: Quick start (SQLite — recommended for homelab)

Single container, one volume, no external database.

```yaml
# deploy/docker-compose.sqlite.yml (future)
services:
  stowkeep:
    image: ghcr.io/stowkeep/stowkeep:latest
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - stowkeep-data:/data
    environment:
      STOWKEEP_DATABASE_DRIVER: sqlite
      STOWKEEP_DATABASE_PATH: /data/stowkeep.db
      STOWKEEP_LOG_LEVEL: info
      STOWKEEP_LOG_FORMAT: json
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  stowkeep-data:
```

### Profile B: Production (PostgreSQL — recommended)

```yaml
# deploy/stowkeep.stack.yml (future)
services:
  stowkeep:
    image: ghcr.io/stowkeep/stowkeep:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      STOWKEEP_DATABASE_DRIVER: postgres
      STOWKEEP_DATABASE_URL: postgres://user:pass@postgres:5432/stowkeep?sslmode=require
      STOWKEEP_MASTER_KEY: ${MASTER_KEY}
    deploy:
      placement:
        constraints: [node.role == manager]
```

Run PostgreSQL as a separate stack, managed service, or external host. See [docs/database.md](../docs/database.md).

### Environment variables (bootstrap set)

| Variable | Required | Description |
|----------|----------|-------------|
| `STOWKEEP_DATABASE_DRIVER` | No | `sqlite` (default) or `postgres` |
| `STOWKEEP_DATABASE_PATH` | SQLite | Path to DB file (default `/data/stowkeep.db`) |
| `STOWKEEP_DATABASE_URL` | Postgres | PostgreSQL connection string; also accepts `sqlite:///path` |
| `STOWKEEP_MASTER_KEY` | Yes (Stage 4+) | 32-byte base64 MEK for secrets |
| `STOWKEEP_DOCKER_HOST` | No | Default `unix:///var/run/docker.sock` |
| `STOWKEEP_HTTP_ADDR` | No | Default `:8080` |
| `STOWKEEP_FEATURES` | No | Comma-separated feature flags |
| `STOWKEEP_LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` — default `info` in production |
| `STOWKEEP_LOG_FORMAT` | No | `json` (production default) or `text` (development) |

---

## 8. Security in CI/CD

- Pin GitHub Actions to SHA hashes (not `@v4` floating)
- `permissions: contents: read` default; elevate only in release job
- OIDC to GHCR where possible instead of long-lived PAT
- Scan images with Trivy/Grype in CI; fail on critical CVEs in base image
- Run `govulncheck`, `gosec`, `npm audit --omit=dev` on every PR
- Never echo secrets in workflow logs
- **Cosign-sign every released image** (keyless OIDC) — Stage 0, not Stage 7 (D-022)
- **Publish SLSA build provenance** via `actions/attest-build-provenance` — Stage 0
- **Publish Syft SBOM** as a cosign attestation — Stage 0
- For PRs from third-party forks: use `pull_request` (not `pull_request_target`) and never expose secrets to forked-PR workflows

---

## 9. Documentation

- **User docs:** `docs/` — install, configuration, database, upgrade paths
- **Contributor docs:** [docs/development-guide.md](../docs/development-guide.md), [docs/code-standards.md](../docs/code-standards.md)
- **Open source standards:** [open-source-standards.md](./open-source-standards.md)
- **Planning:** `planning/` — PRD, not end-user facing
- **Stage 2+:** Optional VitePress site on GitHub Pages

---

## 10. Stage 0 implementation checklist

Use this as the first sprint after PRD approval. Items are grouped: **code scaffold** must build and pass CI; **foundation hardening** are docs/policies that are cheap now and expensive to retrofit (see [planning/decisions-todo.md](./decisions-todo.md)).

### Code scaffold

1. Initialize Go module + React/Vite app
2. Wire frontend embed via `//go:embed web/dist/*` (D-026); SPA fallback to `index.html`; immutable-cache headers for hashed assets
3. Add Dockerfile (multi-stage: node build → go build → distroless runtime, non-root uid 65532) + `docker-compose.dev.yml`
4. Implement database abstraction (SQLite default + PostgreSQL); bounded write pool + `SQLITE_BUSY` retry for SQLite (decisions §3a)
5. Implement structured logging (`pkg/observability/log` with slog; JSON/text via env)
6. Implement HTTP middleware (`request_id` + access log + log-scrubbing for `Authorization`/`Cookie`/token query params)
7. Implement `GET /healthz`
8. Add goose migrations with CI matrix (SQLite + PostgreSQL); include hash-chained `audit_events` schema (D-011) and MEK envelope canary row (D-008)
9. Implement `MasterKeyProvider` interface with `EnvKey` impl and `KMSProvider` stub (D-008)
10. Add SQLite Online Backup API + `pg_dump` streaming as the basis for D-016 (UI lands later)
11. Add `.github/workflows/ci.yml` and `release-image.yml` — supply-chain attestations from day 1 (cosign + SLSA + SBOM, D-022); multi-arch amd64+arm64 required, arm/v7 best-effort (D-018)
12. Add `Makefile`, `.gitignore`, `.env.example`

### Foundation hardening (docs / policy)

13. `docs/security/threat-model.md` — already in repo; keep current as ingress/trust-boundary set grows
14. `planning/decisions-todo.md` — already in repo; walk before any release
15. Draft ADRs 0003 (HA stance), 0004 (`MasterKeyProvider`), 0005 (hash-chained audit), 0006 (Swarm secrets abstraction), 0007 (RBAC engine commitment + UI prototype trigger), 0008 (frontend embed), 0009 (telemetry posture)
16. `.github/CODEOWNERS` requiring maintainer review on `pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/`, `docs/security/`, `pkg/observability/log` (D-023)
17. `GOVERNANCE.md` documenting single-maintainer status, decision process, and hand-off plan (D-024)
18. Verify open-source files present (LICENSE, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, CHANGELOG, PR/issue templates)
19. Add `AGENTS.md` and [planning/phase-gates.md](./phase-gates.md) — mandatory quality bar between stages
20. CI coverage hard gate on `pkg/secrets`, `pkg/rbac`, `pkg/auth` (≥90%) from first commit in those packages (D-019)

### Verification

21. PR → CI green on both DB backends → merge → image in GHCR, signed, with provenance + SBOM
22. `cosign verify` and `cosign verify-attestation` succeed against a freshly pulled image
23. **Stage 0 phase gate** signed off per [phase-gates.md](./phase-gates.md#stage-0--foundation) before Stage 1 work begins

**Definition of done:** Maintainer can `docker pull ghcr.io/stowkeep/stowkeep:main`, verify signature and SBOM with cosign, run with SQLite volume only, get a healthy container, and **Stage 0 gate issue is closed**.
