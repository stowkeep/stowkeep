# Open Source Standards

**Project:** Stowkeep  
**Last updated:** 2026-05-25

This document defines how we build Stowkeep as a welcoming, high-quality open source project from day one. Every contributor — human or agent — should follow these standards.

---

## 1. Principles

1. **Defaults should work** — A single `docker run` with SQLite gets a usable install; PostgreSQL is documented for production.
2. **Tests are not optional** — No feature merges without appropriate test coverage.
3. **Document for strangers** — Assume the reader has never seen the codebase.
4. **Security is everyone's job** — Report vulnerabilities privately; never commit secrets.
5. **Small, reviewable changes** — Trunk-based development with focused PRs.
6. **Stage gates are mandatory** — No next stage until testing, documentation, and code quality pillars pass ([phase-gates.md](./phase-gates.md)).

---

## 1.1 Who this applies to

These standards apply to **every contributor**, including junior developers and **LLM coding agents**. Agents must read [AGENTS.md](../AGENTS.md) and obey phase gate stop rules.

---

## 2. Repository documentation map

| Document | Audience | Purpose |
|----------|----------|---------|
| [README.md](../README.md) | Users + contributors | Project overview, quick start, links |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contributors | How to contribute, PR process, CLA-free workflow |
| [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) | Everyone | Community behavior expectations |
| [SECURITY.md](../SECURITY.md) | Security researchers | Vulnerability reporting |
| [CHANGELOG.md](../CHANGELOG.md) | Users | Release history ([Keep a Changelog](https://keepachangelog.com/)) |
| [docs/development-guide.md](../docs/development-guide.md) | Contributors | Local setup, commands, debugging |
| [docs/code-standards.md](../docs/code-standards.md) | Contributors | Comments, naming, API docs in code |
| [docs/database.md](../docs/database.md) | Operators + contributors | SQLite vs PostgreSQL |
| [docs/logging.md](../docs/logging.md) | Operators + contributors | Production logging (slog, JSON stdout) |
| [planning/phase-gates.md](./phase-gates.md) | Maintainers + agents | Stage completion checklists |
| [AGENTS.md](../AGENTS.md) | LLM agents | Agent entrypoint and hard rules |
| [planning/](../planning/) | Maintainers | PRD, architecture, testing strategy |

---

## 3. Code documentation (in-repo)

### Go (`api/`, `pkg/`)

- **Package comments:** Every package has a `doc.go` or leading comment in the primary file explaining purpose and boundaries.
- **Exported symbols:** All exported types, functions, and constants have godoc comments starting with the symbol name.
- **Comments explain why, not what** — Prefer clear names over narrating obvious code.
- **Complex logic:** Non-obvious algorithms, security decisions, and Docker/Swarm edge cases get brief comments with links to issues or Docker docs when helpful.
- **TODOs:** Format `// TODO(username): description — issue #123` and link a GitHub issue when possible.
- **Examples:** Use `Example*` tests in godoc for non-trivial public APIs (especially `pkg/secrets`, `pkg/rbac`).

```go
// Package secrets provides envelope-encrypted secret storage and version history.
// Secret plaintext must never be logged or written to audit payloads.
package secrets
```

### TypeScript / React (`web/`)

- **TSDoc** on exported hooks, utilities, and complex component props.
- **Components:** Prop interfaces named `{ComponentName}Props`; document non-obvious props.
- **No commented-out code** in merged PRs — delete it; git has history.
- **README per package** only when `web/` grows sub-packages (not required at Stage 0).

### SQL (`migrations/`)

- Each migration file starts with a one-line comment describing the change.
- Destructive migrations require a note in CHANGELOG and upgrade docs.

### OpenAPI (`openapi/`)

- Every public REST endpoint documented with summary, parameters, response schemas, and error codes.
- Generated TypeScript client must stay in sync (CI check).

---

## 4. Standalone documentation

- **User docs** live in `docs/` (install, configuration, upgrade paths).
- **Planning docs** live in `planning/` (PRD, not end-user facing).
- **ADRs** (Architecture Decision Records) go in `docs/adr/NNNN-title.md` for significant decisions (database dual-support, auth model changes).
- Update docs in the **same PR** as behavior changes — docs drift is a bug.

---

## 5. Testing requirements

See [testing-strategy.md](./testing-strategy.md). Summary for contributors:

| Change type | Minimum tests |
|-------------|---------------|
| Bug fix | Regression test reproducing the bug |
| New API endpoint | Handler test + RBAC allow/deny |
| Secrets / auth | Security tests (no leakage, deny bypass) |
| DB migration | Applies cleanly on **both** SQLite and PostgreSQL in CI |
| UI component | Vitest + RTL for interactive logic |
| Critical user journey | Playwright E2E (or issue to add in follow-up if large) |

**CI must pass** before merge. No `--no-verify`.

---

## 6. Code review standards

Reviewers check:

- [ ] Tests included and meaningful
- [ ] Godoc/TSDoc on new exported APIs
- [ ] No secrets, tokens, or real credentials in diff
- [ ] Logging follows [docs/logging.md](../docs/logging.md) — no secret values or auth headers in log statements
- [ ] Migrations work on SQLite and PostgreSQL (or documented exception)
- [ ] CHANGELOG updated for user-visible changes
- [ ] Security-sensitive paths reviewed by someone familiar with RBAC/secrets rules

---

## 7. Licensing and compliance

- **License:** Apache License 2.0 ([LICENSE](../LICENSE))
- **Dependencies:** Compatible licenses only (Apache-2.0, MIT, BSD, ISC). GPL dependencies require maintainer approval.
- **Headers:** Optional per-file Apache headers on Go source; not required if LICENSE is clear (project policy: LICENSE at root is sufficient).
- **No CLA** — Apache 2.0 + DCO sign-off via commit message (`Signed-off-by`) or GitHub co-author metadata.

---

## 8. Community files (GitHub)

Required from Stage 0:

- [x] LICENSE
- [x] README.md
- [x] CONTRIBUTING.md
- [x] CODE_OF_CONDUCT.md
- [x] SECURITY.md
- [x] CHANGELOG.md
- [x] Issue templates (bug, feature)
- [x] Pull request template
- [ ] Dependabot config (Stage 0 implementation)
- [ ] GitHub Actions CI (Stage 0 implementation)

---

## 9. Commit and PR hygiene

- [Conventional Commits](https://www.conventionalcommits.org/) — `feat:`, `fix:`, `docs:`, `test:`, `chore:`
- PR title matches commit style
- PR description: what, why, how to test
- Link issues: `Fixes #123` or `Refs #123`
- Keep PRs under ~400 lines when possible; split large features

---

## 10. What “done” means

### Per PR (during a stage)

A merged PR satisfies open source standards when:

1. Code is readable without oral tradition
2. Tests prove the behavior and guard regressions
3. User-facing changes appear in CHANGELOG
4. Contributor-facing changes appear in docs/ or CONTRIBUTING if process changed
5. CI is green on SQLite and PostgreSQL job matrix

### Per stage (before next stage)

A stage is **not complete** until the [phase gate checklist](./phase-gates.md) for that stage is signed off — all three pillars: **testing**, **documentation**, **code quality**.

Do not begin Stage N+1 work until the Stage N gate issue is closed.
