# 0006. Swarm secrets abstraction: versioned naming + rolling-update materialization

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-009](../../planning/decisions-todo.md)

## Context

Docker Swarm secrets are **immutable**. There is no `docker secret update` — only `create` and `rm`. To rotate a secret used by a service:

1. Create a new secret with a new name.
2. Update every service that referenced the old secret to reference the new one (which triggers a rolling update).
3. Wait for the rolling update to complete and the old tasks to drain.
4. `docker secret rm` the old secret once nothing references it.

Our user-facing model is the opposite: "edit secret `db_password`, click save, it's now version 7 everywhere it's used." We have to bridge the gap without leaking the awkwardness to the user — and without losing the safety guarantees Swarm gives us (atomic per-secret rotation via rolling update).

A naive implementation ("delete and recreate the same-named secret") is impossible — Swarm refuses to delete a secret referenced by a running service. A second naive implementation ("manage all secrets out-of-band, inject as env vars at deploy") loses the security benefits of Swarm secrets (tmpfs mounts, not in `docker inspect`, not in environment dumps).

We need a two-layer abstraction.

## Decision

**App layer (Stowkeep):** The user-facing model. Stowkeep maintains its own `secret_versions` table with sequential integer version numbers and metadata (path, env, tags, who changed it, when, ciphertext). Rollback = mark an older version as `active`. The user only ever sees "secret X, version 7".

**Swarm layer:** Each active SO secret version materializes as a uniquely-named Swarm secret following the convention:

```
{stack}__{secret_name}__v{version}
```

(Double-underscore separators chosen to avoid collisions with user-chosen names that contain single underscores.)

**Reconcile flow** on any deploy:

1. For each SO secret referenced by the deploy, ensure the Swarm secret for the active version exists (create if missing, using the SO ciphertext decrypted via the `MasterKeyProvider` — see [ADR 0004](./0004-master-key-provider.md)).
2. Render the service spec with `secrets:` entries pointing at the active-version Swarm-secret names.
3. Submit the service update. Swarm performs a rolling update: new tasks start with the new secret tmpfs-mounted, take traffic only when their healthcheck passes, then old tasks drain.
4. After the update completes and reconciles, mark older Swarm-secret versions as candidates for garbage collection.

**Garbage collection** is a periodic job. A Swarm secret is GC-eligible when no service spec (across all stacks SO is responsible for) references its name. GC has a configurable grace period (default 24h) before the actual `docker secret rm`, to allow rollback without a re-encrypt.

### Secret inject modes (Stage 4+)

| Mode | Behavior | Default |
|------|----------|---------|
| `file` | Materialize Swarm secret; mount under `/run/secrets/...` | **Yes** |
| `env` | Inject decrypted value into `ContainerSpec.Env` | Opt-in; less secure |

Native Swarm file mounts do not require Stowkeep-specific `*_FILE` env pointers unless the target application expects them. Legacy images that cannot read file mounts may use opt-in `inject: env` or a future `inject: wrapper` (flavor wall — see [planning/todo.md](../../planning/todo.md)).

**Rollback** is a redeploy with a different `active` version chosen in the SO database. Garbage collection is paused for the now-unreferenced "newer" versions during the grace window.

**Naming constraints:** Swarm secret names are limited to 64 characters. The reconciler validates `len(stack)+len(name)+len("__"+"__v")+len(version) ≤ 64` and rejects deploys that would exceed the limit with a clear error.

## Consequences

**Easier**

- App UX is clean: versioning, rollback, audit all work like the user expects.
- Rotation = standard Swarm rolling update; no novel orchestration to debug.
- Failures are local and recoverable: a failed rolling update leaves the previous version still running, fully referenced, in the DB.

**Harder**

- Garbage collection must be conservative. Deleting an in-use Swarm secret breaks the cluster. The GC job is single-flight, runs against the live cluster inventory (not a cache), and has the grace window above.
- Old Swarm secrets accumulate if GC is paused or broken. The UI surfaces "N orphaned Swarm secrets pending GC" as a status signal.
- Cross-stack secret sharing is a future feature (DEF-tracked) — the `{stack}__` prefix means sharing a single secret across two stacks isn't possible without an explicit "shared" namespace. Acceptable for v1.
- Documentation needs to explain to operators why their cluster has many `appname__db_password__v*` secrets — answer: that's the version history; GC handles it.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-009`
- [docs/security/threat-model.md §4.4](../security/threat-model.md) — secret lifecycle threats
- [planning/PRD.md §6.4 SEC-10](../../planning/PRD.md)
- [planning/tech-stack.md](../../planning/tech-stack.md) — secrets architecture
- [ADR 0004 — `MasterKeyProvider` interface](./0004-master-key-provider.md)
