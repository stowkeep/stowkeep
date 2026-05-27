# Stage 1 Phase Gate — Sign-off Checklist

Use this document when opening the **Phase gate: Stage 1 complete** GitHub issue after merging [PR #14](https://github.com/stowkeep/stowkeep/pull/14) to `main`.

Copy the checklist from [planning/phase-gates.md](../planning/phase-gates.md#stage-1--swarm-read-only-dashboard) and attach evidence links.

**Blocker:** Do not start Stage 2 until this gate is signed off and the issue is closed.

---

## Stage PRs

| PR | Title | Status |
|----|-------|--------|
| [#14](https://github.com/stowkeep/stowkeep/pull/14) | Stage 1: Swarm read-only dashboard | Merge to `main` first |

---

## Pre-merge verification (local)

Run on the Stage 1 branch before merging PR #14:

```bash
make lint && make test
go test ./cmd/... ./pkg/... -short -race
cd web && npm run lint && npm run typecheck && npm run test
make build
```

Optional Docker integration (requires socket + Swarm):

```bash
export STOWKEEP_INTEGRATION_DOCKER=1
go test -tags=integration ./pkg/docker/... -count=1
```

---

## Post-merge verification (maintainer)

After PR #14 merges and CI on `main` is green:

```bash
docker pull ghcr.io/stowkeep/stowkeep:main
docker run -d --name stowkeep-stage1 \
  -p 8080:8080 \
  -v stowkeep-data:/data \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e STOWKEEP_DATABASE_PATH=/data/stowkeep.db \
  -e STOWKEEP_FEATURES=swarm_readonly \
  ghcr.io/stowkeep/stowkeep:main
curl -sf http://localhost:8080/healthz
```

Manual checklist ([testing-strategy.md §8](../planning/testing-strategy.md#stage-1)):

- [ ] Connect via socket on Linux manager — lists match `docker node ls` / `docker service ls`
- [ ] Node drain/cordon reflects in UI within refresh interval
- [ ] Logout clears session

First-run flow: [docs/install.md](./install.md)

---

## Gate checklist (three pillars)

### Testing

| Item | Status | Evidence |
|------|--------|----------|
| Docker integration tests (list nodes/services/stacks) — skip with `-short` when no Docker | ✅ | `pkg/docker/integration_test.go`; CI job `test-go-docker` |
| Auth bootstrap tests (first admin, login, session) | ✅ | `pkg/auth/*_test.go`, `pkg/server/auth_test.go` |
| HTTP middleware tests (`request_id`, access log shape) | ✅ | `pkg/http/middleware/middleware_test.go` |
| Manual: connect socket, lists match CLI | ⬜ | Maintainer — post-merge on Swarm host |

**CI (PR #14):** https://github.com/stowkeep/stowkeep/actions/runs/26526043277  
**CI on `main` after merge:** _fill in URL_

**Known deferred tests (non-blocking for gate):** [testing-strategy.md §14](../planning/testing-strategy.md#14-stage-1-follow-up-test-backlog) — Swarm HTTP handler matrix and frontend route guards. Track in first Stage 2 hardening PR or before Stage 3.

### Documentation

| Item | Status | Evidence |
|------|--------|----------|
| `docs/install.md` or README section: first-run + socket mount | ✅ | [docs/install.md](./install.md), [README.md](../README.md) |
| OpenAPI spec covers all Stage 1 endpoints | ✅ | [openapi/openapi.yaml](../openapi/openapi.yaml) — `/api/v1/swarm/*`, auth routes |
| Threat model updated: Docker socket trust boundary | ✅ | [docs/security/threat-model.md](./security/threat-model.md) §4.7 |

### Code quality

| Item | Status | Evidence |
|------|--------|----------|
| All new API routes behind auth (except `/healthz`, static assets) | ✅ | `pkg/server/server.go`, `pkg/server/swarm_test.go`, `auth_test.go` |
| Feature flag `swarm_readonly` documented | ✅ | [docs/install.md](./install.md), `.env.example`, [CHANGELOG.md](../CHANGELOG.md) |
| No Docker API response bodies logged at info level | ✅ | `pkg/docker/swarm.go` logs counts only (`count` field), not payloads |

---

## Open gate issue

The phase gate template (`.github/ISSUE_TEMPLATE/phase_gate.yml`) is a **YAML form** — `gh issue create --template` only works with markdown templates. Use one of:

**Browser form (recommended):**

```bash
gh issue create -w
# pick "Phase gate sign-off", or open:
# https://github.com/stowkeep/stowkeep/issues/new?template=phase_gate.yml
```

**CLI with body:**

```bash
gh issue create \
  --title "Phase gate: Stage 1 complete" \
  --label "phase-gate,triage" \
  --body-file docs/stage-1-gate-issue-body.md
```

Or copy the checklist from this file into the issue body.

---

## On sign-off

1. Close the gate issue with maintainer sign-off on all three pillars.
2. Move `[Unreleased]` Stage 1 entries in [CHANGELOG.md](../CHANGELOG.md) to a `v0.1.0` (or `v0.2.0-stage1`) release section.
3. Note any deferred items in [decisions-todo.md](../planning/decisions-todo.md) that affect Stage 2.
4. **Stage 2 work may begin** — stack deploy and manage ([PRD § Stage 2](../planning/PRD.md)).
