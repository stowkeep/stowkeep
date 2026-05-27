# Testing Strategy

**Project:** Stowkeep  
**Last updated:** 2026-05-26

This document defines testing frameworks, the test pyramid, environments, quality gates, and stage-specific test plans.

---

## 1. Testing philosophy

1. **Test behavior, not implementation** — especially for RBAC and GitOps reconciliation.
2. **Integration over mock for Docker paths** — Docker API behavior is the risk; use real Engine in CI where feasible.
3. **Security tests are first-class** — secret leakage, permission bypass, and audit completeness are release blockers.
4. **Fast feedback on PR** — unit tests complete in < 2 minutes; heavy integration runs in parallel or post-merge nightly.
5. **No untested permission paths** — every API route has at least one allow and one deny test.
6. **Dual database parity** — migrations and persistence tests run on SQLite and PostgreSQL in CI.
7. **Stage gates block progress** — see [phase-gates.md](./phase-gates.md); no next stage until the gate passes.

---

## 2. Test pyramid

```
                    ┌─────────────┐
                    │  E2E (few)  │  Playwright — critical user journeys
                   ┌┴─────────────┴┐
                   │ Integration    │  testcontainers-go; SQLite + Postgres + Docker
                  ┌┴───────────────┴┐
                  │  Unit (many)    │  Go table tests, Vitest component tests
                  └─────────────────┘
```

| Layer | % of tests (target) | Max duration (PR) |
|-------|---------------------|-------------------|
| Unit | ~70% | 90s |
| Integration | ~25% | 5 min (parallel jobs) |
| E2E | ~5% | 10 min (nightly or main-only initially) |

---

## 3. Frameworks and tools

### Backend (Go)

| Tool | Purpose |
|------|---------|
| **`testing` + `testify`** | Unit tests, assertions, suites |
| **`testcontainers-go`** | Ephemeral PostgreSQL, Redis, Docker-in-Docker |
| **`httptest` + chi** | HTTP handler tests without full server |
| **`go.uber.org/mock` (mockgen)** | Mock interfaces for Git providers, external only |
| **`golangci-lint`** | Static analysis (includes `govulncheck`, `staticcheck`) |
| **`gotestsum`** | CI-friendly test output |

**Avoid:** Over-mocking Docker client — prefer testcontainers with real Docker socket in CI runners that support it (GitHub `ubuntu-latest` with Docker).

### Frontend (React)

| Tool | Purpose |
|------|---------|
| **Vitest** | Unit and component tests (Vite-native, fast) |
| **React Testing Library** | User-centric component tests |
| **MSW (Mock Service Worker)** | Mock API in component tests |
| **Playwright** | E2E browser tests |
| **ESLint + typescript-eslint** | Lint |
| **`tsc --noEmit`** | Type checking in CI |

### API contract

| Tool | Purpose |
|------|---------|
| **OpenAPI 3.1 spec** | Single source for API documentation |
| **`openapi-typescript`** | Generate TypeScript client for web |
| **`schemathesis` or `dredd`** (Stage 3+) | Contract tests — responses match spec |

### Security

| Tool | Purpose |
|------|---------|
| **gosec** | Go security scanner in CI |
| **Trivy / Grype** | Container and dependency CVE scan |
| **Custom tests** | Assert secret values never appear in logs/responses/audit |
| **Log handler tests** | Capture slog output in unit tests; verify HTTP middleware omits auth headers |

---

## 4. Test environments

| Environment | Where | Docker | Database | Use |
|-------------|-------|--------|----------|-----|
| **Local** | Developer machine | Host socket | SQLite (default) or Postgres | Dev + manual QA |
| **CI unit** | GitHub Actions | Not required | Temp SQLite file | PR gate (fast) |
| **CI integration** | GitHub Actions | DinD or socket mount | Matrix: SQLite + PostgreSQL | PR gate |
| **CI E2E** | GitHub Actions | Swarm init optional | SQLite or Postgres | Nightly / pre-release |
| **Staging** | Self-hosted Swarm | Real cluster | PostgreSQL (recommended) | Pre-release manual QA |

### Swarm in CI

GitHub Actions runners support Docker. For Swarm-specific tests:

```bash
docker swarm init --advertise-addr 127.0.0.1
# run integration tests
docker swarm leave --force
```

Run Swarm init in a **dedicated integration job** — not every unit test job.

---

## 5. CI quality gates

### Required on every PR (blocking)

| Gate | Command / check |
|------|-----------------|
| Go lint | `golangci-lint run ./...` |
| Go unit tests | `go test ./... -short -race` |
| Migration smoke (SQLite) | `make test-migrations-sqlite` |
| Migration smoke (PostgreSQL) | `make test-migrations-postgres` |
| Web lint | `npm run lint` |
| Web typecheck | `npm run typecheck` |
| Web unit tests | `npm run test -- --run` |
| Go build | `go build -o /dev/null ./...` |
| Web build | `npm run build` |
| Docker build | `docker build .` |

### Required on merge to `main` (blocking)

All PR gates plus:

| Gate | Command / check |
|------|-----------------|
| Integration tests | `go test ./... -tags=integration` |
| Image smoke test | Start container, curl `/healthz` |

### Nightly (non-blocking initially → blocking pre-1.0)

| Gate | Command / check |
|------|-----------------|
| E2E Playwright | Full journey against docker-compose stack |
| Security scan | Trivy image + gosec |
| Coverage report | Upload to Codecov; no hard threshold until Stage 3 |

### Coverage gates (D-019)

| Package | Line coverage | Enforcement |
|---------|---------------|-------------|
| `pkg/secrets` | ≥ 90% | **Hard CI gate** from first commit in package |
| `pkg/rbac` | ≥ 90% | **Hard CI gate** from first commit in package |
| `pkg/auth` | ≥ 90% | **Hard CI gate** from first commit in package |
| `pkg/gitops` | ≥ 80% | Reported; soft gate until Stage 5 GA, then hard |
| `pkg/docker` | ≥ 75% (integration-heavy) | Reported |
| `web/` components | ≥ 60% (focus on critical paths) | Reported |

The three security-sensitive packages are hard-gated from day 1 because retrofitting coverage discipline after the code exists never works. Use `go-test-coverage` action with `--threshold-file 90` against the relevant package paths.

### Fuzz tests (D-020)

Every parser that ingests untrusted input gets a Go fuzz target (`testing.F`) landing in the same PR as the parser:

| Target | Inputs |
|--------|--------|
| Compose YAML parser | Random YAML, including malicious anchors and depth |
| `.stowkeep.yml` parser | Random YAML, including unknown keys |
| RBAC policy/condition evaluator | Random subject/action/resource/condition tuples |
| Secret path resolver | Random path strings, including traversal attempts |
| Webhook payload parsers (Stage 5.5) | Random JSON + signature header combinations |

Run fuzz targets short (`-fuzz -fuzztime=30s`) on every PR; long fuzz runs nightly.

### Log-leak sentinel tests (D-021)

Every code path that touches secret values, tokens, or session cookies has a test that:

1. Captures all `slog` output during a known workload using a buffered `slog.Handler` test helper.
2. Asserts that no sentinel value (`__SENTINEL_SECRET_VALUE__`) used as the secret/token appears in the captured output.
3. Repeats for both `json` and `text` log formats.

A small helper (`pkg/observability/log/logtest`) exposes `CaptureSlog(t)` returning a `*bytes.Buffer` and a teardown. Reviewers reject PRs that touch secret code paths without a corresponding sentinel test.

### Supply-chain CI checks (D-022)

Every `release-image` build asserts (and fails on missing):

- `cosign verify` succeeds against the just-pushed image
- `cosign verify-attestation --type slsaprovenance` succeeds
- `cosign verify-attestation --type cyclonedx` (or `spdx`) succeeds against the published SBOM
- `govulncheck`, `gosec`, Trivy image scan, `npm audit --omit=dev` all green (or documented allowlist)

---

## 6. Test categories by domain

### 6.1 Docker / Swarm (`pkg/docker`)

**Unit**

- Parse and normalize Docker API responses
- Compose file validation (invalid YAML, unsupported keys)
- Stack name sanitization

**Integration**

- List nodes/services when Swarm initialized
- Deploy minimal stack (`nginx:alpine`, 1 service)
- Remove stack; verify services gone
- Scale service replicas

**Fixtures:** `testdata/compose/valid-stack.yml`, `invalid-stack.yml`

**Stage 1 status (PR #14):** `pkg/docker` has unit tests for mappers (`mapService`, `mapNode`, `mapTask`) and `stackFromServices` (not-found/found). CI job `test-go-docker` runs Swarm integration with `STOWKEEP_INTEGRATION_DOCKER=1`. HTTP handler behavior tests for list endpoints are tracked in [§14.1](#141-swarm-http-handler-behavior-pkgserver).

### 6.2 Authentication (`pkg/auth`)

**Unit**

- Password hash verify
- JWT/session expiry
- API token generation and validation

**Integration**

- Login → cookie → authorized request
- Expired session → 401
- Bootstrap admin only on empty DB

**Stage 1 status (PR #14):** Handler and store tests cover bootstrap, login, session middleware (401 vs 500), rate limiting, and Argon2 verification. Frontend auth routing follow-ups are in [§14.2](#142-app-routing-and-route-guards-websrc).

### 6.3 RBAC (`pkg/rbac`)

**Unit**

- Casbin policy evaluation table tests:
  - User in `developers` role → `swarm.stacks.deploy` on `staging` → **allow**
  - Same user → deploy on `production` → **deny**
  - `secrets.describe` without `read_value` → list shows names, values redacted

**Integration**

- Every HTTP route: test with role fixtures from `testdata/rbac/policies.csv`

**Critical:** Permission bypass regression suite — attempt IDOR on stack IDs, secret paths, other users' resources.

### 6.4 Secrets (`pkg/secrets`)

**Unit**

- Encrypt/decrypt roundtrip (envelope)
- Version increment on update
- Rollback restores ciphertext at version N

**Integration**

- Create secret → DB contains no plaintext (grep ciphertext column)
- User without `read_value` → API returns 403 or redacted
- Deploy injects secret into stack without logging value

**Security tests**

```go
func TestSecretNeverInLogs(t *testing.T) {
    // capture log buffer; perform CRUD; assert plaintext absent
}

func TestHTTPMiddleware_NeverLogsAuthorizationHeader(t *testing.T) {
    // send request with Authorization; assert header value absent from log output
}
```

### 6.5 GitOps (`pkg/gitops`)

**Unit**

- Parse `.stowkeep.yml` (valid, missing fields, invalid types)
- Compute desired state from compose + env

**Integration**

- Clone public test repo (or embedded git fixture via `go-git` memory storage)
- Detect commit change → trigger reconcile
- Failed compose → sync status `error`, no partial deploy

**Mock:** GitHub webhook payloads from `testdata/webhooks/github-push.json`

### 6.6 Previews (`pkg/previews`)

**Integration**

- Simulate PR open event → preview stack name `app-pr-42`
- PR close → stack removed
- Max preview limit → oldest evicted or new blocked (per config)

### 6.7 Audit (`pkg/audit`)

**Integration**

- Mutating action → audit row with actor, action, resource, timestamp
- Secret read_value → audit without value payload

### 6.8 Database (`pkg/db`)

**Unit**

- Driver detection from env / URL
- DSN parsing for SQLite paths and Postgres URLs

**Integration**

- goose `up` / `down` on fresh SQLite file
- goose `up` / `down` on testcontainers PostgreSQL
- Parameterized store tests: `TestUsersStore/sqlite` and `TestUsersStore/postgres`

**Rule:** New migrations must pass both backends before merge.

---

## 7. End-to-end test plan (Playwright)

Run against `docker-compose.test.yml` stack: API + web + Docker. Test both SQLite-only and Postgres profiles in nightly matrix.

### Critical journeys (automate in nightly)

| ID | Journey | Stages |
|----|---------|--------|
| E2E-01 | Login → view nodes list | 1 |
| E2E-02 | Deploy stack from compose paste → see running service | 2 |
| E2E-03 | Non-admin denied deploy on production environment | 3 |
| E2E-04 | Create secret → version history shows 2 entries after edit | 4 |
| E2E-05 | Register git app → manual sync → status synced | 5 |
| E2E-06 | Simulate webhook → preview URL visible | 6 |

### Playwright structure

```
web/e2e/
├── fixtures/auth.ts      # login helpers
├── stacks.spec.ts
├── secrets.spec.ts
├── gitops.spec.ts
└── playwright.config.ts    # baseURL, trace on failure
```

---

## 8. Manual test checklist (release)

Before tagging `v0.x.0` at each stage:

### Stage 1

- [ ] Connect via socket on Linux manager
- [ ] Node drain/cordon reflects in UI within refresh interval
- [ ] Logout clears session

### Stage 2

- [ ] Deploy multi-service stack with published port
- [ ] Invalid compose shows field-level error
- [ ] Remove stack confirmation prevents accidents

### Stage 3

- [ ] Create user with viewer role — no deploy buttons
- [ ] Audit log shows deploy from Stage 2

### Stage 4

- [ ] Rotate secret, redeploy, app picks up new value
- [ ] DB dump contains no plaintext secrets

### Stage 5

- [ ] Push to connected repo triggers deploy within webhook/poll window
- [ ] Disconnect git — sync shows error gracefully

### Stage 6

- [ ] Preview stack isolated (different network/name)
- [ ] Close PR removes preview within cleanup window

---

## 9. Test data and fixtures

| Path | Contents |
|------|----------|
| `testdata/compose/` | Valid/invalid compose files |
| `testdata/rbac/` | Casbin policy CSV fixtures |
| `testdata/webhooks/` | GitHub/GitLab webhook JSON |
| `testdata/secrets/` | Encrypted test vectors (never real secrets) |
| `web/src/test/mocks/` | MSW handlers for component tests |

**Rule:** No real credentials in repo. Use `STOWKEEP_TEST_*` env vars in CI secrets for live GitHub integration tests (optional, scheduled).

---

## 10. Performance and load (Stage 7)

Not required for MVP. When needed:

- **`k6`** or **`vegeta`** against API list endpoints
- Target: 50 services, 200 tasks — list APIs < 500ms p95
- GitOps: 20 concurrent syncs without deadlock

---

## 11. Implementation order for test infrastructure

Align with Stage 0 engineering checklist:

1. **Stage 0:** `go test ./...`, vitest scaffold, CI workflow, testcontainers postgres smoke test
2. **Stage 1:** Docker integration test with skip if no docker (`testing.Short()`)
3. **Stage 2:** Compose deploy integration test
4. **Stage 3:** RBAC table-driven suite for all routes
5. **Stage 4:** Secret encryption + leakage tests
6. **Stage 5:** Git fixture + webhook parser tests
7. **Stage 6:** Playwright E2E first journey; expand nightly

---

## 12. Definition of done (testing)

A feature is not done unless:

- [ ] Unit tests cover happy path and primary error paths
- [ ] RBAC allow/deny tests exist for new endpoints
- [ ] Integration test exists if feature touches Docker, Git, or DB transactions
- [ ] No secret values in test assertions logged on failure
- [ ] CI passes on PR
- [ ] Manual checklist item added to release doc if user-visible

---

## 13. Phase gates and testing

Stage completion requires the **testing pillar** of [phase-gates.md](./phase-gates.md):

| When | Action |
|------|--------|
| **During stage** | Every PR meets §12 definition of done |
| **End of stage** | Run §8 manual checklist for that stage |
| **Before next stage** | Gate issue opened; CI green on `main`; maintainer sign-off |

Stage-specific automated test requirements are listed in [phase-gates.md §3](./phase-gates.md#stage-gate-checklists).

---

## 14. Stage 1 follow-up test backlog

Tracked during [PR #14](https://github.com/stowkeep/stowkeep/pull/14) review (2026-05-26). These items were intentionally deferred from the Stage 1 PR to keep review scope focused. Revisit before signing the Stage 1 phase gate or in the first hardening PR after merge.

### 14.1 Swarm HTTP handler behavior (`pkg/server`)

**Shipped in PR #14**

| Area | Tests | Location |
|------|-------|----------|
| Feature flag | `swarm_readonly` off → 404 when authenticated | `pkg/server/swarm_test.go` |
| Auth gate | Unauthenticated → 401 on status + list endpoints | `pkg/server/swarm_test.go`, `auth_test.go` |
| Status shape | Authenticated `/swarm/status` returns `docker_host` | `pkg/server/swarm_test.go` |
| Stack lookup | `stackFromServices` 404/200 without Docker | `pkg/docker/swarm_test.go` |

**Deferred — full endpoint behavior matrix**

| Endpoint | Happy path | Error branches | Why deferred |
|----------|------------|----------------|--------------|
| `GET /api/v1/swarm/nodes` | 200 + `items[]` | 502 when Engine unreachable | `test-go` job has no Docker socket; real client returns transport errors, not fixture data |
| `GET /api/v1/swarm/services` | same | same | same |
| `GET /api/v1/swarm/tasks` | same; `?service_id=` filter | same | same |
| `GET /api/v1/swarm/stacks` | same | same | same |
| `GET /api/v1/swarm/stacks/{name}` | 200 detail | 404 not found; 502 backend failure | 404 logic tested at `stackFromServices` layer; HTTP 404 vs 502 distinction needs injectable backend |

**Planned approach**

1. Introduce a narrow interface (e.g. `SwarmReader`) consumed by `SwarmHandler`, implemented by `pkg/docker.Client` in production.
2. Add table-driven tests in `pkg/server/swarm_test.go` with a stub returning fixed nodes/services/tasks/stacks or injected errors.
3. Assert status codes and JSON bodies: 200 + `items`, 404 `stack not found`, 502 `docker request failed`, plus existing 401/404 feature-flag cases.
4. Optionally add one happy-path case in the existing `test-go-docker` CI job (Swarm already initialized there) as a thin integration smoke test.

**Done when:** Every Stage 1 Swarm route has at least one happy-path and one error-path automated test without flaking on laptops without Docker.

### 14.2 App routing and route guards (`web/src`)

**Shipped in PR #14**

| Scenario | Coverage | Location |
|----------|----------|----------|
| Bootstrap complete → login | Heading “Sign in” on `/login` | `web/src/App.test.tsx` |
| Bootstrap required → setup | Heading “Create admin account” on `/setup` | `web/src/App.test.tsx` |
| Fetch stub cleanup | `afterEach(() => vi.unstubAllGlobals())` | `web/src/App.test.tsx` |
| `setupStatus()` failure | Guards default to `needsBootstrap: false` | `web/src/auth/RouteGuards.tsx` |

**Deferred — route/guard matrix**

| Scenario | Entry route | Expected outcome |
|----------|-------------|------------------|
| Authenticated home | `/` | Redirect to dashboard (e.g. `/nodes`) |
| Guest on login | `/login` while session valid | Redirect to `/` or dashboard |
| Unknown path | `/non-existent` | Redirect per router policy (login or dashboard) |
| Protected dashboard | `/nodes` without session | Redirect to `/login` |
| Setup after bootstrap | `/setup` when `needs_bootstrap: false` | Redirect to `/login` |

**Planned approach**

1. Prefer **MSW** handlers (see `web/src/test/mocks/` in §9) over ad-hoc `fetch` stubs for repeatable API responses.
2. Extend `App.test.tsx` (or split `App.routes.test.tsx`) with one test per row above.
3. Align with Playwright **E2E-01** (login → nodes list) for nightly confidence once `web/e2e/` exists (§7).

**Done when:** Each guard component (`RequireAuth`, `GuestOnly`, `SetupOnly`) has a dedicated behavior test; wildcard routing is asserted once.

### 14.3 Checklist (loop back)

Use this when picking up the backlog:

- [ ] §14.1 — `SwarmReader` stub + handler table tests for all five list/detail routes
- [ ] §14.1 — optional `test-go-docker` happy-path smoke for one list endpoint
- [ ] §14.2 — MSW fixtures for auth/setup/me
- [ ] §14.2 — five route/guard Vitest cases (matrix above)
- [ ] §7 — wire Playwright E2E-01 when e2e scaffold lands
- [ ] Update this section: move completed rows to “Shipped” and delete checklist items
