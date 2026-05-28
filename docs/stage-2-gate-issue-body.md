## Stage 2 — Deploy and manage stacks

**Gate checklist doc:** [docs/stage-2-gate.md](docs/stage-2-gate.md)  
**Stage PR:** _update when PR opens_  
**CI (PR branch):** _update when CI runs_  
**CI on main (post-merge):** _update after merge_

### Manual verification (maintainer)

- [ ] Enable `STOWKEEP_FEATURES=swarm_readonly,stack_deploy`
- [ ] Deploy multi-service stack with published port from UI
- [ ] Invalid compose shows field-level validation errors
- [ ] Remove stack requires confirmation dialog
- [ ] Scale service replicas from stack detail
- [ ] View service logs stream in UI
- [ ] Permission-builder prototype reviewed ([docs/prototypes/permission-builder.html](docs/prototypes/permission-builder.html))

### Testing pillar

- [ ] Compose validation unit tests — `pkg/compose/*_test.go`, `testdata/compose/`
- [ ] Deploy + remove integration — `pkg/docker/integration_test.go`; CI `test-go-docker`
- [ ] Scale service integration test
- [ ] Audit chain write + verify — `pkg/audit/*_test.go` (sqlite + postgres)
- [ ] Permission-builder prototype outcome recorded (D-010 / DEF-002)
- [ ] CI green on PR branch
- [ ] Stage manual checklist complete (testing-strategy.md §8 Stage 2)

### Documentation pillar

- [ ] [docs/stacks.md](docs/stacks.md) — deploy, remove, scale, logs, feature flag
- [ ] Compose validation errors documented
- [ ] [docs/audit.md](docs/audit.md) — audit event format, hash chain
- [ ] OpenAPI — Stage 2 stack/deploy endpoints
- [ ] Threat model §4.3 — stack deploy mitigations
- [ ] CHANGELOG `[Unreleased]` Stage 2 section

### Code quality pillar

- [ ] Deploy paths use context timeouts
- [ ] Confirm dialogs for destructive stack remove
- [ ] RBAC stub on all mutating routes (`pkg/rbac`)
- [ ] `stack_deploy` documented
- [ ] All stage PRs reviewed and merged

### Maintainer sign-off

- [ ] I confirm all three pillars pass and **Stage 3 may begin**
