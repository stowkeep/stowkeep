# 0004. `MasterKeyProvider` interface from Stage 4 day 1

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-008](../../planning/decisions-todo.md)

## Context

Secret values are envelope-encrypted: a per-version Data Encryption Key (DEK) encrypts the value, and the DEK itself is wrapped by a Master Encryption Key (MEK). The MEK is the linchpin of the entire secret store — losing it loses everything, leaking it leaks everything.

Where the MEK lives matters:

| Source | Pros | Cons |
|--------|------|------|
| Env var (`STOWKEEP_MASTER_KEY`) | Trivial to deploy; works in any environment | Operator must protect the env reliably; rotation is manual; theft = total disclosure |
| Cloud KMS (AWS KMS, GCP KMS, Vault Transit) | Hardware-rooted protection; audit-logged Unwrap calls; rotation supported by the platform | Adds external dependency; harder for homelab |
| HSM | Strongest | Niche; very few self-hosted operators have one |

If we hard-wire the env-var source into the crypto path now, swapping to KMS later is a rewrite of `pkg/secrets`. If we never plan for rotation, we paint ourselves into a corner the moment a key is suspected of compromise.

The plan from D-008: ship the interface and a single env-var implementation in Stage 4, design the DB schema to support rotation, leave the KMS implementation as a clearly-typed stub. The "extra work" today is small. The "extra work" of a rewrite later is large.

## Decision

Stage 4 ships a `MasterKeyProvider` interface as the only call site for MEK-level operations. `EnvKey` is the v1 implementation. `KMSProvider` is a documented stub interface (no implementation; tests cover the interface contract via a fake).

Interface (informative — exact shape may evolve in code):

```go
// pkg/secrets/keyprovider
type MasterKeyProvider interface {
    // Wrap encrypts a fresh DEK and returns the wrapped bytes plus the
    // key_id of the active MEK at wrap time.
    Wrap(ctx context.Context, dek []byte) (wrapped []byte, keyID string, err error)

    // Unwrap decrypts a wrapped DEK using the MEK identified by key_id.
    // Implementations must support all key_ids still present in the DB
    // (current + previous, until a rewrap migration retires them).
    Unwrap(ctx context.Context, wrapped []byte, keyID string) (dek []byte, err error)

    // ActiveKeyID returns the key_id used for new Wrap calls.
    ActiveKeyID() string
}
```

DB schema rule: every row that stores a wrapped DEK also stores `key_id`. Rotation re-wraps DEKs in the background against the new active MEK and updates `key_id` per row, without re-encrypting the underlying secret values.

Startup invariant: a single `envelope_canary` row (random plaintext encrypted at install time) is unwrapped on startup; mismatch fails the boot with a clear "wrong MEK" error rather than silently producing garbage.

Stage 7 plans at least one cloud KMS implementation (likely AWS KMS or Vault Transit) plus the documented rotation runbook in `docs/secrets-rotation.md`.

## Consequences

**Easier**

- KMS adoption is a Stage 7 add of one file, not a rewrite of `pkg/secrets`.
- MEK rotation is a background DEK-rewrap, not a stop-the-world re-encryption of values.
- Tests are simpler: a `FakeKeyProvider` in tests exercises the same interface as production.

**Harder**

- Stage 4 has slightly more code than "just `aesgcm.Seal` with env-var MEK".
- The `envelope_canary` discipline must be respected on every install and rotation.
- Operators have a new "do not lose" asset to manage (the MEK), with clear documentation.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-008`, `DEF-001`
- [docs/security/threat-model.md §4.5](../security/threat-model.md) — MEK lifecycle threats
- [planning/tech-stack.md](../../planning/tech-stack.md) — secrets architecture
- [planning/PRD.md §6.4 SEC-05](../../planning/PRD.md)
