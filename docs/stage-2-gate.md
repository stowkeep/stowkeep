# Stage 2 Phase Gate — Sign-off Checklist

Use this document when tracking the **Phase gate: Stage 2 complete** GitHub issue.

Copy the checklist from [planning/phase-gates.md](../planning/phase-gates.md#stage-2--deploy-and-manage-stacks) and attach evidence links.

**Blocker:** Do not start Stage 3 until this gate is signed off and the issue is closed.

---

## Stage PRs

| PR | Title | Status |
|----|-------|--------|
| [#18](https://github.com/stowkeep/stowkeep/pull/18) | Stage 2: Deploy and manage stacks | Open — CI green |

**Gate issue:** https://github.com/stowkeep/stowkeep/issues/17

**CI (PR #18):** https://github.com/stowkeep/stowkeep/actions/runs/26547010557

---

## Pre-merge verification (local)

Run on the Stage 2 branch before opening the PR:

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

Local dev with deploy enabled:

```bash
export STOWKEEP_FEATURES=swarm_readonly,stack_deploy
make dev
```

Manual checklist ([testing-strategy.md §8](../planning/testing-strategy.md#stage-2)):

- [ ] Deploy multi-service stack with published port
- [ ] Invalid compose shows field-level error
- [ ] Remove stack confirmation prevents accidents

Operator docs: [docs/stacks.md](./stacks.md)

---

## Gate checklist (three pillars)

### Testing

| Item | Status | Evidence |
|------|--------|----------|
| Compose validation unit tests (valid/invalid fixtures) | ⬜ | `pkg/compose/*_test.go`, `testdata/compose/` |
| Integration: deploy + remove minimal stack on Swarm | ⬜ | `pkg/docker/integration_test.go`; CI `test-go-docker` |
| Scale service integration test | ⬜ | same |
| Audit chain: deploy event written and verifies on read | ⬜ | `pkg/audit/*_test.go` (sqlite + postgres) |
| Permission-builder prototype reviewed (D-010) | ⬜ | `docs/prototypes/permission-builder.html`; `decisions-todo.md` DEF-002 |
| Manual checklist §8 Stage 2 complete | ⬜ | Maintainer |

**CI (PR):** _fill in URL_  
**CI on `main` after merge:** _fill in URL_

### Documentation

| Item | Status | Evidence |
|------|--------|----------|
| Stack deploy documented for operators | ⬜ | [docs/stacks.md](./stacks.md) |
| Compose validation errors documented (common fixes) | ⬜ | [docs/stacks.md](./stacks.md) |
| Audit log format documented | ⬜ | [docs/audit.md](./audit.md) |
| OpenAPI covers Stage 2 endpoints | ⬜ | [openapi/openapi.yaml](../openapi/openapi.yaml) |
| Threat model §4.3 updated | ⬜ | [docs/security/threat-model.md](./security/threat-model.md) |
| CHANGELOG `[Unreleased]` Stage 2 section | ⬜ | [CHANGELOG.md](../CHANGELOG.md) |

### Code quality

| Item | Status | Evidence |
|------|--------|----------|
| Deploy paths use context timeouts | ⬜ | `pkg/docker/deploy.go` |
| Confirm dialogs for destructive actions in UI | ⬜ | `web/src/pages/StackDetail.tsx` |
| RBAC hooks stubbed or enforced consistently | ⬜ | `pkg/rbac/stub.go`, stack handlers |
| Feature flag `stack_deploy` documented | ⬜ | `docs/stacks.md`, `.env.example` |
| All mutating routes behind auth + `stack_deploy` | ⬜ | `pkg/server/stack_handlers.go` |

---

## Open gate issue

```bash
gh issue create \
  --title "Phase gate: Stage 2 complete" \
  --label "phase-gate,triage" \
  --body-file docs/stage-2-gate-issue-body.md
```

---

## On sign-off

1. Close the gate issue with maintainer sign-off on all three pillars.
2. Move `[Unreleased]` Stage 2 entries in [CHANGELOG.md](../CHANGELOG.md) to a release section.
3. Note any deferred items in [decisions-todo.md](../planning/decisions-todo.md) that affect Stage 3.
4. **Stage 3 work may begin** — RBAC and audit UI ([PRD § Stage 3](../planning/PRD.md)).
