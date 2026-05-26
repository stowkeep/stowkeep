# AGENTS.md — Instructions for AI coding agents

This file tells automated agents (Cursor, Copilot, Claude, etc.) how to work on **Stowkeep** without breaking project rules.

## Read before writing code

1. [planning/phase-gates.md](planning/phase-gates.md) — **mandatory** stage checkpoints
2. [planning/open-source-standards.md](planning/open-source-standards.md) — per-PR requirements
3. [docs/code-standards.md](docs/code-standards.md) — Go, TypeScript, SQL conventions
4. [planning/PRD.md](planning/PRD.md) — what stage we are building and acceptance criteria
5. [planning/testing-strategy.md](planning/testing-strategy.md) — what tests to add

## Hard rules

### Stage boundaries

- **Do not start Stage N+1 until Stage N's phase gate is signed off** (see [phase-gates.md](planning/phase-gates.md)).
- If asked to skip ahead, complete the gate checklist first or tell the user what is blocking.

### Every PR must include

| Pillar | Requirement |
|--------|-------------|
| **Tests** | Behavior tests for all changes; RBAC allow/deny for new API routes; SQLite **and** Postgres if migrations/SQL |
| **Documentation** | Godoc/TSDoc on exports; user docs + CHANGELOG if user-visible; ADR/threat model if security/architecture changed |
| **Code quality** | `make lint && make test` passes; no secrets in code/logs; follow [logging.md](docs/logging.md) |

### Never

- Commit secrets, `.env`, or credentials
- Log secret values, tokens, or `Authorization` headers
- Merge features without tests
- Disable lint rules or skip CI checks without maintainer approval
- Store plaintext secrets in the database or audit payloads

## Project layout

```
api/           # Go HTTP server entry
pkg/           # Shared packages (docker, gitops, secrets, rbac, observability)
web/           # React + Vite frontend
migrations/    # goose SQL (shared/, sqlite/, postgres/)
docs/          # User + contributor documentation
planning/      # PRD, phase gates, testing strategy
openapi/       # API spec
```

## Commands (target — use when Makefile exists)

```bash
make dev              # Local dev (SQLite)
make lint && make test
make test-integration
```

## Current focus

Check [planning/README.md](planning/README.md) and open GitHub issues/milestones for the active stage. Default assumption: **Stage 0** until gate issue closes.

## When finishing work

1. Run lint and tests locally
2. Update CHANGELOG for user-visible changes
3. Update docs in the same PR as code
4. If completing a stage, fill the gate checklist in [phase-gates.md](planning/phase-gates.md) — do not start the next stage

## Questions

Prefer existing docs over guessing. If requirements are ambiguous, ask the user rather than inventing scope.
