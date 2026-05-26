# 0001. Dual database: SQLite and PostgreSQL

- **Status:** Accepted
- **Date:** 2026-05-25

## Context

Stowkeep targets both homelab operators (minimal dependencies) and production teams (concurrency, HA, standard backup tooling). Requiring PostgreSQL for every install raises the barrier for try-it-out and single-node deployments.

## Decision

Support two database backends with identical application features:

1. **SQLite** (`modernc.org/sqlite`) — default for quick-start and small installs; single file on a Docker volume at `/data/stowkeep.db`.
2. **PostgreSQL 16** — recommended for production; required documentation for backup, HA, and migration from SQLite.

- Migrations use goose with portable SQL in `migrations/shared/` where possible.
- CI runs migration and integration tests on **both** backends on every PR.

## Consequences

**Easier**

- One-container install with no external database for evaluation and homelab.
- Contributors can run `make dev` without Docker Compose services.

**Harder**

- Migration authors must consider dialect differences (JSON types, index syntax).
- Performance limits of SQLite must be documented; no silent degradation.

## References

- [docs/database.md](../database.md)
- [planning/tech-stack.md](../../planning/tech-stack.md)
