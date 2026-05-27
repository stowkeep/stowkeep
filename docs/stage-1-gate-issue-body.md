## Stage 1 — Swarm read-only dashboard

**Stage PR:** https://github.com/stowkeep/stowkeep/pull/14  
**Gate checklist doc:** https://github.com/stowkeep/stowkeep/blob/main/docs/stage-1-gate.md (after gate doc PR merges)  
**CI (PR branch):** https://github.com/stowkeep/stowkeep/actions/runs/26526043277  
**CI on main (post-merge):** _update after PR #14 merges_

### Manual verification (maintainer)

- [ ] Bootstrap admin on fresh install
- [ ] Login, browse nodes/services/tasks/stacks — counts match `docker node ls` / `docker service ls`
- [ ] Settings → Test connection reports Docker reachable
- [ ] Logout clears session
- [ ] Node drain/cordon visible after UI refresh (if Swarm cluster available)

### Testing pillar

- [x] Docker integration tests — `pkg/docker/integration_test.go`; CI `test-go-docker`
- [x] Auth bootstrap / login / session tests — `pkg/auth/*_test.go`
- [x] HTTP middleware tests — `pkg/http/middleware/middleware_test.go`
- [ ] CI green on **main** after PR #14 merge
- [ ] Stage manual checklist complete (testing-strategy.md §8 Stage 1)

**Deferred (documented, non-blocking):** testing-strategy.md §14 — full Swarm handler matrix + frontend route guards

### Documentation pillar

- [x] [docs/install.md](docs/install.md) — first-run, socket mount, `swarm_readonly`
- [x] OpenAPI — `/api/v1/swarm/*` and auth endpoints
- [x] Threat model §4.7 — Docker socket trust boundary
- [x] CHANGELOG `[Unreleased]` Stage 1 section

### Code quality pillar

- [x] Swarm routes behind auth; `/healthz` and static assets public
- [x] `swarm_readonly` documented in install guide and `.env.example`
- [x] Docker list calls log counts only, not API response bodies
- [ ] All stage PRs reviewed and merged
- [ ] Migrations `00002_auth` apply on SQLite + PostgreSQL in CI

### Maintainer sign-off

- [ ] I confirm all three pillars pass and **Stage 2 may begin**
