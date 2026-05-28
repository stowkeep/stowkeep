# 0011. Bootstrap config: `.env`, `STOWKEEP_*_FILE`, and MEK input normalization

- **Status:** Accepted
- **Date:** 2026-05-27
- **Stage:** Dev ergonomics now; `STOWKEEP_*_FILE` + MEK input in Stage 4

## Context

Stowkeep is configured through `STOWKEEP_*` environment variables. Local development docs already reference `.env`, but the binary did not load it. Production Swarm deployments need Docker secret file mounts for sensitive bootstrap values (MEK, backup key) without operators manually base64-encoding key material.

Bootstrap MEK generation in the setup wizard was considered and rejected — operators generate and protect the MEK once.

## Decision

### Local development (Stage 2+)

- `config.Load()` loads an optional `.env` file (or `STOWKEEP_ENV_FILE`) via `godotenv` **before** `envconfig.Process`.
- Existing process environment variables take precedence over `.env`.
- Production containers do **not** rely on committed `.env` files.

### Production bootstrap secrets (Stage 4)

- Sensitive bootstrap variables support a `_FILE` suffix, e.g. `STOWKEEP_MASTER_KEY_FILE=/run/secrets/stowkeep_master_key`.
- Resolution order: direct env value wins; else read file (trim trailing newline); else unset.
- Initial targets: `STOWKEEP_MASTER_KEY`, `STOWKEEP_BACKUP_KEY` (when backups ship).

### MEK input (`pkg/secrets/keyinput`, Stage 4)

- **Env (`STOWKEEP_MASTER_KEY`):** accept user-provided passphrase, hex (64 chars), or base64-encoded 32 bytes. Normalize internally to 32-byte key material (passphrase via KDF).
- **File (`STOWKEEP_MASTER_KEY_FILE`):** read raw bytes from Swarm secret mount; must decode to 32 bytes after trim.
- Operators are **not** required to base64-encode manually for homelab use.
- No bootstrap wizard generates or rotates the MEK.

### UI feature flags

- Feature flags remain environment-only (`STOWKEEP_FEATURES`).
- `GET /api/v1/version` exposes enabled flags; UI hides disabled capabilities. Server middleware remains authoritative.

## Consequences

**Easier**

- `cp .env.example .env && make dev` works without Makefile env hacks.
- Swarm stack deploy can mount MEK via native Docker secrets.
- Homelab operators paste a passphrase or `openssl rand -hex 32` without encoding steps.

**Harder**

- `pkg/secrets/keyinput` needs thorough unit tests for every accepted format and rejection case.
- Docs must clearly distinguish dev `.env` from production secret mounts.

## References

- [ADR 0004 — MasterKeyProvider](./0004-master-key-provider.md)
- [docs/configuration.md](../configuration.md)
- [planning/PRD.md § Stage 4](../../planning/PRD.md)
