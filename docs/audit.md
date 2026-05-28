# Audit log format

Stowkeep records security-relevant actions in a hash-chained `audit_events` table (ADR-0005). Stage 2 writes deploy lifecycle events; the audit UI arrives in Stage 3.

## Event fields

| Column | Description |
|--------|-------------|
| `occurred_at` | UTC timestamp (RFC3339Nano) |
| `actor_id` | User ID from session |
| `action` | e.g. `stack.deploy`, `stack.remove`, `service.scale` |
| `resource_type` | `stack` or `service` |
| `resource_id` | Stack name or service ID |
| `request_id` | HTTP request correlation ID |
| `before_hash` / `after_hash` | SHA-256 of canonical payload metadata (never secret plaintext) |

## Hash chain

Each row stores:

- `prev_hash` — `row_hash` of the previous row (32 zero bytes for the genesis row)
- `row_hash` — `sha256(prev_hash || canonical_json(fields))`

On startup, Stowkeep verifies the chain in the background. A break is logged at ERROR, recorded in `audit_integrity_events`, and surfaced in the UI (Stage 3+) — the app continues running.

## Stage 2 actions

| Action | When |
|--------|------|
| `stack.deploy` | Successful stack deploy; `after_hash` is compose content hash |
| `stack.remove` | Stack removed |
| `service.scale` | Service replica count changed |

Secret values and compose `environment:` entries are never stored in audit payloads or application logs.
