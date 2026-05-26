# 0005. Hash-chained `audit_events` table from day 1

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-011](../../planning/decisions-todo.md)

## Context

Stowkeep promises an "immutable audit log" for compliance-aware users. A plain `INSERT`-only table is only enforced by application code: anyone with database write access (which the operator has, by definition — they hold the SQLite file or the Postgres credentials) can edit or delete rows after the fact. That makes the "immutable" claim hollow.

Tamper-evidence cannot be retrofitted. Once historical rows exist without a hash chain, you can never prove the older rows are unmodified. The cost of adding a chain on day 1 is one extra column, one extra `sha256` call per insert, and one startup verification job. The cost of adding it later is "we have to start a new chain from scratch and the old data is forever unverifiable."

This kind of decision is exactly what foundation hardening exists for — cheap now, impossible later.

## Decision

From the first migration that creates `audit_events`, the schema includes:

- `prev_hash BLOB NOT NULL` — the `row_hash` of the immediately preceding row, or `\x00...00` (32 zero bytes) for the first row.
- `row_hash BLOB NOT NULL` — `sha256(prev_hash || canonical_serialize(row_fields))`.

`canonical_serialize` is JCS-style ([RFC 8785](https://www.rfc-editor.org/rfc/rfc8785)) deterministic JSON of the audit row's stable fields: `{time, actor, action, resource, request_id, before_hash, after_hash}`. We hash `before`/`after` payloads rather than embedding them so chain integrity does not depend on payload encoding stability.

Writes happen under a serializable transaction that:

1. Reads the latest `row_hash` (using `FOR UPDATE` on Postgres; `BEGIN IMMEDIATE` on SQLite).
2. Computes the new `row_hash`.
3. Inserts the new row.

On every startup, a background goroutine verifies the chain from a checkpoint (or from row 1 on small DBs). A break is logged at `ERROR`, surfaced in the UI under a permanent "audit integrity" banner, and recorded in a `audit_integrity_events` table — but does **not** halt the application (would create a denial of service if an operator wanted to recover from a real incident).

The audit row never contains plaintext secret values; payloads are recorded as content hashes plus opaque metadata. This protects the log from accidental secret leakage and keeps the chain verifiable independent of payload size.

## Consequences

**Easier**

- Tamper-detection is automatic; operators get a verifiable claim, not a promise.
- The chain is portable across SQLite and Postgres because it only uses standard SHA-256 + ordered reads.
- Post-incident forensics: a missing-row or modified-row attack is detectable.

**Harder**

- Every insert pays a serializable-transaction cost. At our expected scale (≤ few hundred audit events per minute peak) this is invisible; document the limit anyway.
- Bulk inserts and reorderings (e.g. migration tooling) must preserve hash semantics — covered by a SQL-style invariant check in tests.
- Future audit-payload schema changes must extend the canonical serializer carefully; a versioned canonical-serialization scheme is reserved (`{ "v": 1, ... }`) so we can evolve without breaking older chain segments.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-011`
- [docs/security/threat-model.md §4.8](../security/threat-model.md) — audit tampering
- [planning/PRD.md §6.5 AUTH-09, §10 #5](../../planning/PRD.md)
- [RFC 8785 — JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785)
