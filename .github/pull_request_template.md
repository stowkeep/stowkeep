## Summary

<!-- What does this PR do and why? Link issues: Fixes #123 -->

**Stage:** <!-- e.g. Stage 1 — Swarm read-only. Do not implement Stage N+1 until phase gate N passes. -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation only
- [ ] Refactor / chore
- [ ] Phase gate sign-off (Stage N complete)

## Testing

<!-- How was this tested? -->

- [ ] `make lint && make test` passes locally
- [ ] Added/updated unit tests
- [ ] Added/updated integration tests (if applicable)
- [ ] Tested on SQLite
- [ ] Tested on PostgreSQL (if DB-related)
- [ ] Manual testing steps documented below

**Manual test steps:**

1.
2.

## Documentation

- [ ] Godoc / TSDoc on new exported APIs
- [ ] User docs updated (`docs/`)
- [ ] [CHANGELOG.md](../CHANGELOG.md) updated (user-visible changes)
- [ ] ADR and/or [threat model](../docs/security/threat-model.md) updated (if architecture/security changed)
- [ ] No secrets or credentials in diff
- [ ] Logging follows [docs/logging.md](../docs/logging.md) if adding log statements

## Code quality

- [ ] No new lint warnings (`golangci-lint`, eslint, `tsc`)
- [ ] No unresolved `TODO` without linked issue
- [ ] Errors wrapped with context at handling boundaries
- [ ] Security-sensitive paths ready for CODEOWNER review (`pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/`)

## Screenshots (if UI)

<!-- Optional -->

## Checklist

- [ ] I have read [AGENTS.md](../AGENTS.md) (agents), [CONTRIBUTING.md](../CONTRIBUTING.md), [code standards](../docs/code-standards.md), and [phase gates](../planning/phase-gates.md)
- [ ] My commits follow [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] This PR does not skip an incomplete phase gate to start the next stage
