# 0003. Single-replica control plane in v1; multi-replica HA is a Stage 5+ decision

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-007](../../planning/decisions-todo.md)

## Context

Stowkeep is the control plane for Docker Swarm. Two questions sit underneath any HA discussion:

1. **What if the control plane is down?** Swarm workloads keep running on their own. Reconciliation, secret reads, audit writes, and the UI become unavailable until the control plane restarts. For a homelab and small-team product this is acceptable; for an enterprise multi-team product it is not.
2. **What if we want to run more than one replica of the control plane?** That requires real engineering: leader election for the GitOps reconciler (otherwise every replica reconciles every app on every tick), a shared session store (otherwise restarts log everyone out and load balancing breaks sticky cookies), webhook ingest ownership (otherwise duplicated processing or lost deliveries), and shared scheduling for background jobs (backups, MEK rotation, secret GC).

The default database is SQLite, which structurally cannot support multi-replica writes. PostgreSQL would, via advisory locks, but the application code would still need to be written for it.

Building this complexity now — before we have any evidence we need it — is premature. Building features now that *quietly assume* a single replica, then later being surprised when they break under replicas, is the failure mode this ADR exists to prevent.

## Decision

**v1 ships as a single-replica control plane.** Document this explicitly so users plan around it (restart = downtime; rolling restart loses sessions; one webhook ingest target).

Until multi-replica is on the roadmap, every new feature is allowed to assume:

- The GitOps reconciler runs in-process on the only replica.
- Sessions and per-process state are tied to that process; loss on restart is acceptable.
- All scheduled jobs (backups, MEK rotation, secret GC) run in-process.
- Webhook ingest (Stage 5.5) terminates on the same replica that handles HTTP.

**Multi-replica HA is a Stage 5+ design decision.** When it becomes a requirement, a follow-up ADR will specify: leader election (likely PostgreSQL advisory locks for the reconciler), shared session storage (DB-backed sessions or a Redis cache), single-flight webhook processing (idempotency keys + delivery deduplication), and the operator-facing upgrade path from single-replica to active/passive to active/active.

This ADR is intentionally **not** prescriptive about *which* HA model lands. It is prescriptive that v1 won't pretend to support one.

## Consequences

**Easier**

- No leader election, no consensus layer, no shared cache infrastructure in v1.
- SQLite default keeps working without async-replication shenanigans.
- The reconciler is a goroutine with locks, not a distributed system.

**Harder**

- Operator restarts cause user-visible downtime — including everyone being logged out.
- A future multi-replica effort must audit every "runs once in this process" assumption (likely 10–30 sites).
- Any contributor adding a background job needs to know about this constraint (covered by `CODEOWNERS` review on reconciler/scheduler paths).
- For users who need true HA today, the answer is "not yet" — which is a real adoption blocker for enterprise. We accept this for v1.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-007`, `DEF-012`
- [docs/security/threat-model.md §6](../security/threat-model.md) — single-maintainer pre-1.0 caveat
- [planning/PRD.md §7](../../planning/PRD.md) — NFR availability row reflects this decision
