# Architecture Decision Records

We document significant technical decisions as ADRs. Format: [MADR](https://adr.github.io/madr/) simplified.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [0001](./0001-dual-database-sqlite-postgres.md) | Dual database: SQLite + PostgreSQL | Accepted |
| [0002](./0002-structured-logging-slog.md) | Structured logging with log/slog | Accepted |
| [0003](./0003-single-replica-control-plane.md) | Single-replica control plane in v1 (HA deferred) | Accepted |
| [0004](./0004-master-key-provider.md) | `MasterKeyProvider` interface from Stage 4 day 1 | Accepted |
| [0005](./0005-hash-chained-audit-log.md) | Hash-chained `audit_events` from day 1 | Accepted |
| [0006](./0006-swarm-secrets-abstraction.md) | Swarm secrets abstraction (versioned + rolling update) | Accepted |
| [0007](./0007-rbac-engine-casbin-with-prototype-gate.md) | RBAC engine = Casbin, with mandatory UI-prototype gate | Accepted |
| [0008](./0008-frontend-embed-fs.md) | Frontend embedded via `embed.FS` | Accepted |
| [0009](./0009-telemetry-posture.md) | Anonymous opt-in telemetry posture | Accepted |
| [0010](./0010-project-name-stowkeep.md) | Project name: Stowkeep | Accepted |

## Creating an ADR

1. Copy the template below to `NNNN-short-title.md`
2. Open a PR with the ADR; discuss before implementing large changes
3. Status: Proposed → Accepted | Rejected | Superseded

### Template

```markdown
# NNNN. Title

- **Status:** Proposed
- **Date:** YYYY-MM-DD

## Context

What is the issue?

## Decision

What did we decide?

## Consequences

What becomes easier or harder?
```
