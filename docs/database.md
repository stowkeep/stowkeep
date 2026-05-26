# Database Guide

Stowkeep supports two database backends. Choose based on your deployment size and operational requirements.

---

## Quick comparison

| | **SQLite (embedded)** | **PostgreSQL (recommended production)** |
|---|----------------------|----------------------------------------|
| **Best for** | Homelab, single-node, try-it-out installs | Production, teams, multi-user, heavy GitOps |
| **Dependencies** | None — single container + volume | PostgreSQL server (managed or self-hosted) |
| **Setup complexity** | Minimal | Moderate |
| **Concurrent writes** | Single writer (WAL mode helps reads) | Full MVCC concurrency |
| **Backup** | Copy the `.db` file (stop app or use SQLite backup API) | `pg_dump`, managed snapshots |
| **HA / replication** | Not supported | Standard Postgres HA |
| **Resource use** | Very low | Higher (separate service) |

Both backends run the **same application code** and **equivalent schema**. CI tests every migration against both.

---

## SQLite (embedded mode)

### When to use

- First install / evaluation
- Homelab Swarm with one manager
- Single administrator or small team with light concurrent use
- You want one container and one Docker volume — no external database

### Configuration

```bash
# Explicit SQLite path (default for quick-start if no DATABASE_URL set)
STOWKEEP_DATABASE_DRIVER=sqlite
STOWKEEP_DATABASE_PATH=/data/stowkeep.db
```

Or use a URL-style DSN:

```bash
STOWKEEP_DATABASE_URL=sqlite:///data/stowkeep.db
```

### Docker example

```yaml
services:
  stowkeep:
    image: ghcr.io/stowkeep/stowkeep:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - stowkeep-data:/data
    environment:
      STOWKEEP_DATABASE_DRIVER: sqlite
      STOWKEEP_DATABASE_PATH: /data/stowkeep.db

volumes:
  stowkeep-data:
```

### Operational notes

- **WAL mode** is enabled automatically for better read concurrency.
- **Foreign keys** are enforced (`PRAGMA foreign_keys = ON`).
- **Backup:** Stop the container or use the built-in backup endpoint (Stage 7); copy `/data/stowkeep.db` and `-wal`/`-shm` files if present.
- **Permissions:** Ensure the volume is writable by the non-root container user.

### Limitations

- Not suitable for high write throughput (many simultaneous GitOps syncs + audit logging).
- No built-in replication — plan migration to PostgreSQL before scaling users or sync frequency.
- Some advanced analytics queries may be slower at very large audit volume (100k+ rows).

---

## PostgreSQL (production mode)

### When to use

- Production deployments
- Multiple concurrent users and GitOps applications
- Large audit history and secret version tables
- Existing Postgres infrastructure or managed DB (RDS, Cloud SQL, etc.)
- You need standard backup/restore and HA tooling

### Configuration

```bash
STOWKEEP_DATABASE_DRIVER=postgres
STOWKEEP_DATABASE_URL=postgres://user:pass@postgres:5432/stowkeep?sslmode=disable
```

For production, always use `sslmode=require` (or verify-full) when connecting over a network.

### Docker Compose example

See `deploy/docker-compose.postgres.yml` (Stage 0+) for a reference stack with Postgres + Stowkeep.

---

## Driver selection logic

At startup, Stowkeep resolves the database in this order. The first rule that matches wins; later rules are not consulted.

1. **`STOWKEEP_DATABASE_URL` set** — driver is inferred from the URL scheme:
   - `postgres://` or `postgresql://` → PostgreSQL
   - `sqlite:///absolute/path` → SQLite
   - Unknown scheme → startup fails with a clear error.
2. **`STOWKEEP_DATABASE_DRIVER` set** (no URL set):
   - `postgres` → requires a separate `STOWKEEP_DATABASE_URL` (which would have been caught by rule 1; this combination is the error case and startup fails with a clear message)
   - `sqlite` → uses `STOWKEEP_DATABASE_PATH` (default `/data/stowkeep.db` if unset)
3. **Neither set (quick-start default):** SQLite at `/data/stowkeep.db` — zero external dependencies.

The UI setup wizard (Stage 1+) shows which driver is active and warns when SQLite is used in a multi-user production context.

---

## Migrations

Migrations use [goose](https://github.com/pressly/goose). Layout:

```
migrations/
├── postgres/     # PostgreSQL-specific when dialect differs
├── sqlite/       # SQLite-specific when dialect differs
└── shared/       # Portable SQL run on both (preferred)
```

**Rule:** Prefer portable SQL in `migrations/shared/`. Split into dialect-specific folders only when necessary (e.g. `JSONB` vs `TEXT`, index syntax).

CI runs:

```bash
goose -dir migrations/sqlite sqlite /tmp/test.db up
goose -dir migrations/postgres postgres $TEST_DATABASE_URL up
```

---

## Backups (first-class product feature — D-016)

Stowkeep ships built-in scheduled backups for **both** database backends. Backups are a first-class feature (not Stage 7 polish), available from the UI, with a one-click restore path.

### What gets backed up

- Full database (users, RBAC, audit chain, secret ciphertext, GitOps state, etc.).
- Backup files are encrypted with a `BACKUP_KEY` that is **distinct from the MEK**, so a compromised backup destination does not leak secrets, and a compromised MEK does not invalidate backup integrity.
- Each backup file is HMAC-signed (key = `BACKUP_KEY`); restore refuses on mismatch.

### How it works

| Backend | Mechanism |
|---------|-----------|
| **SQLite** | [`sqlite3` Online Backup API](https://www.sqlite.org/backup.html) — constant memory, no app downtime. **Never** copy the `.db` file directly while the app is running — the WAL state can produce inconsistent copies. |
| **PostgreSQL** | Streaming `pg_dump` — no downtime; honours WAL consistency. Documented support for managed-DB snapshot APIs as an alternative (RDS, Cloud SQL). |

### Destinations

| Destination | Notes |
|-------------|-------|
| Local volume | Default; same volume as the DB (configurable separate path recommended) |
| S3-compatible object store | AWS S3, Cloudflare R2, MinIO, Backblaze B2 — configured via UI with bucket + access key + endpoint URL |

### Schedule and retention

- Configurable cron-style schedule (default: daily 03:00 UTC).
- Configurable retention policy (e.g. "keep 7 daily + 4 weekly + 12 monthly").
- One-click on-demand backup from the UI.

### Restore

- One-click from the UI; requires the `system:restore` permission (default = owner only).
- Restore is an audit event written **after** the new chain is verified.
- Restore CLI subcommand also available for disaster recovery (`stowkeep restore --from s3://...`).

> **Why `BACKUP_KEY` is separate from the MEK.** If the MEK is the only key, then anyone who can read both the backup file and the MEK can decrypt everything. By splitting keys, an attacker needs both the backup destination credentials *and* the `BACKUP_KEY` — typically held by different humans / systems. Document both in your operator runbook as "do not lose" assets.

---

## Migrating SQLite → PostgreSQL

1. Deploy PostgreSQL and set `STOWKEEP_DATABASE_URL`.
2. Run the migration tool (Stage 2+ CLI or documented `stowkeep migrate export`):
   - Export data from SQLite
   - Import into PostgreSQL
3. Restart Stowkeep against Postgres.
4. Verify audit count and secret version counts match.

Detailed runbook: `docs/upgrade/sqlite-to-postgres.md` (added when export tooling ships).

---

## Development

| Task | SQLite | PostgreSQL |
|------|--------|------------|
| Local default | `make dev` uses SQLite in `./.data/dev.db` | `make dev-postgres` |
| Tests | `-tags=sqlite` or default fast path | testcontainers PostgreSQL |
| Reset DB | Delete file | `docker compose down -v` |

See [development-guide.md](./development-guide.md) for full setup.
