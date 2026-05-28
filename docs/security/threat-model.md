# Threat Model

**Project:** Stowkeep
**Status:** Initial draft (v0.1) — living document
**Last updated:** 2026-05-25
**Maintainer responsibility:** This document is updated whenever a new trust boundary, ingress, persistence, or external dependency is introduced. PRs that change those things must update this file in the same PR.

This is a STRIDE-based threat model for the Stowkeep control plane. Its purpose is to make security trade-offs visible and to drive concrete mitigations into the design — not to be a compliance artifact.

> **Important caveat (pre-1.0):** Stowkeep is pre-alpha, currently developed by a single maintainer. While the design favors conservative defaults, **pre-1.0 releases should not be used to host third-party workloads or hold production secrets you cannot afford to lose**. This caveat will be lifted only after at least one external security review of `pkg/secrets`, `pkg/auth`, and `pkg/rbac`.

---

## 1. What we are protecting (assets)

| Asset | Sensitivity | Where it lives | Notes |
|-------|-------------|----------------|-------|
| Secret values (plaintext) | **Critical** | In-memory at decrypt time only | Never written to disk, logs, audit payloads, or HTTP responses outside `read_value` flow |
| Secret ciphertext + DEK | High | Database (SQLite/Postgres) | Encrypted at rest with envelope (DEK per secret version, MEK wraps DEKs) |
| MEK (master encryption key) | **Critical** | `STOWKEEP_MASTER_KEY` env or `STOWKEEP_MASTER_KEY_FILE` Swarm mount (later: KMS) | Loss = unrecoverable secrets; theft = full secret disclosure |
| User credentials | High | Database — argon2id / bcrypt hashed | Never logged, never returned in API responses |
| Session tokens | High | HTTP-only cookies + DB-backed session table | `HttpOnly`, `Secure`, `SameSite=Lax`/`Strict`; short idle timeout |
| API tokens | High | Database — hashed at rest | Bearer pattern; revocable; scoped to actions |
| Audit log | High (integrity) | Database, append-only with hash chain (D-011) | Tamper-evident; never contains secret values |
| Docker socket / Engine API access | **Critical** | Host filesystem mount or agent TLS | Effective root on the host; abuse = container escape, data exfil, ransomware |
| GitOps repo credentials (deploy keys / PATs) | High | Database — encrypted with MEK | Git read access = potential code injection vector |
| Stack deployment desired state | Medium | Database | Tampering = wrong workload runs in production |
| Compose files in flight | Medium | Memory + transient DB | Can reference secrets and env |
| RBAC policy state | High (integrity) | Database | Tampering = privilege escalation |
| Backups (SQLite/Postgres dumps) | **Mirrors all of above** | Local volume + S3 (when configured) | Backups must be encrypted; backup destination credentials are themselves secrets |

---

## 2. Trust boundaries

```
                      ┌────────────────────────────────────────────────┐
                      │                  Untrusted                     │
                      │  (internet, anonymous attackers, scanners)     │
                      └────────────────────────────────────────────────┘
                                          │ HTTPS
   ── TRUST BOUNDARY 1: ingress ─────────│──────────────────────────────
                                          ▼
                      ┌────────────────────────────────────────────────┐
                      │           Authenticated user (low-priv)         │
                      │   (developer/viewer roles via UI or API token)  │
                      └────────────────────────────────────────────────┘
   ── TRUST BOUNDARY 2: authz (Casbin) ──│──────────────────────────────
                                          ▼
                      ┌────────────────────────────────────────────────┐
                      │       Stowkeep process (Go server)       │
                      │  Handles: HTTP, reconciler, secret decrypt,    │
                      │  audit write, Docker calls                     │
                      └────────────────────────────────────────────────┘
                              │                  │                │
                              │ socket/TLS       │ TCP/TLS        │ git+https/ssh
   ── TRUST BOUNDARY 3 ──────│──────────────────│────────────────│────────
                              ▼                  ▼                ▼
                       ┌────────────┐    ┌────────────┐   ┌────────────┐
                       │ Docker     │    │ Database   │   │ Git        │
                       │ Engine     │    │ (SQLite/PG)│   │ provider   │
                       │ (root!)    │    │            │   │ (external) │
                       └────────────┘    └────────────┘   └────────────┘
                                                  │
   ── TRUST BOUNDARY 4: ops ─────────────────────│────────────────────────
                                                  ▼
                                            ┌──────────────┐
                                            │ Host operator│
                                            │ (root, MEK,  │
                                            │  backups)    │
                                            └──────────────┘
```

| Boundary | What crosses it | Primary controls |
|----------|-----------------|------------------|
| 1. Ingress | HTTP requests, webhooks (later) | TLS, rate limiting, session/JWT verification, CSRF (cookie auth), webhook HMAC |
| 2. Authorization | Authenticated subject → action on resource | Casbin policy (deny by default), condition evaluator, route-level enforcement |
| 3. Outbound to dependencies | Docker, DB, Git | Least-privilege creds, connection allow-lists, TLS where applicable |
| 4. Operational | Host filesystem, env vars, backups | Document operator responsibilities; no in-product control here |

---

## 3. Adversaries (in scope)

| Adversary | Capability | Goal |
|-----------|------------|------|
| **External anonymous attacker** | Network access to the API and any webhook endpoint | RCE, secret disclosure, account takeover, DoS |
| **External authenticated low-priv user** | Valid `viewer`/`developer` session or API token | Privilege escalation, secret theft outside scope, deploying unauthorized stacks |
| **Compromised user account** | Stolen cookie / token / password | Same as authenticated user; possibly admin |
| **Compromised Git repository** | Write access to a GitOps-connected repo | Push malicious compose → arbitrary container on Swarm |
| **Compromised Git provider** | Full impersonation of provider | Webhook forgery, MITM on `git pull` |
| **Compromised dependency (supply chain)** | Malicious npm/Go module | RCE inside the control-plane process |
| **Curious operator** | Read access to host (DB file, env vars) | Bulk secret disclosure, audit tampering |
| **Curious cloud/storage admin** (S3 backup) | Read access to backup destination | Bulk secret disclosure via backups |

### Out of scope (explicitly)

- Adversary with root on the Stowkeep host — owns everything by definition.
- Physical access to the host.
- Supply chain attacks on the Linux kernel / Docker daemon itself.
- Misconfiguration where the operator publicly exposes `/var/run/docker.sock` over TCP without TLS.

---

## 4. STRIDE per major flow

We focus on the flows that exist at v1 scope: login, API authz, stack deploy, secret read/write, GitOps reconcile (pull-based), and Docker socket interaction. Webhooks are explicitly post-v1 (per D-014) and get a placeholder section.

### 4.1 Login (Stage 1+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S**poofing | Stolen password, credential stuffing | argon2id hashing, per-account rate limit, account lockout on N failed attempts, generic failure messages |
| **T**ampering | Session cookie modification | `HttpOnly`, `Secure`, `SameSite=Lax`, HMAC-signed or DB-backed sessions (no JWT in cookies) |
| **R**epudiation | "I didn't log in" | Audit row on every login attempt (success and failure category, no password contents) |
| **I**nfo disclosure | Username enumeration via timing or message difference | Constant-time compare on lookup miss; uniform "invalid credentials" response |
| **D**oS | Brute-force, login flood | Rate limit per IP and per account; exponential backoff |
| **E**levation | Bootstrap admin race (first-run setup window) | First-run setup token printed to logs (Portainer pattern), single-use, short TTL; refusal to bootstrap if any user exists |

### 4.2 Authenticated API call → authorization (Stage 3+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S** | Token replay across users | Sessions tied to user_id + ip-prefix (optional); API tokens hashed at rest |
| **T** | CSRF on cookie-auth endpoints | `SameSite` + double-submit token (`X-Csrf-Token` header) for mutating verbs |
| **R** | Action without trace | Every mutating route writes an `audit_events` row with actor, action, resource, before/after digest |
| **I** | IDOR — accessing another user's stack/secret/path | Authz check on every route uses resource owner + Casbin condition; integration test suite asserts deny on cross-tenant IDs |
| **D** | Expensive endpoints (large list, log streaming) | Pagination required; per-route concurrency limits; cancellable contexts |
| **E** | Missing authz check on a new route | Lint rule / test that asserts every route is registered through the authz middleware; default-deny if not |

### 4.3 Stack deploy (Stage 2+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S** | Deploy attributed to wrong user | Actor from session; reflected in audit |
| **T** | Compose file tampered between validate and deploy (TOCTOU) | Hash the parsed compose; deploy uses the hashed canonical form |
| **R** | "I didn't deploy that" | Audit + retained compose snapshot (hash + content) per deploy |
| **I** | Compose file env block contains secrets that get logged | Compose parser scrubs `environment:` from any log payload; reviewer test (D-021) enforces |
| **D** | Huge compose / nested YAML bomb | Parser limits: max size (1 MiB default), max depth, max anchors |
| **E** | Deploy reaches services outside user's allowed env/stack scope | Stage 2: `pkg/rbac` admin-only stub on all mutating routes; Stage 3 Casbin conditions enforce env/stack scope |

**Stage 2 implementation (2026-05):** `pkg/compose` enforces 1 MiB / depth / anchor limits; deploy stores compose content hash in audit `after_hash`; mutating routes require `stack_deploy` + session auth.

### 4.4 Secret create / update / read (Stage 4+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S** | Service account impersonation | API tokens scoped to specific actions, hashed at rest, revocable |
| **T** | Direct DB edit to swap ciphertext | Hash-chained audit log (D-011) makes inconsistency detectable; row-level integrity checks at startup (optional) |
| **R** | "I didn't read that secret" | Every `read_value` writes an `audit_events` row; never contains the value |
| **I** | Value leak via logs, error wrapping, panic stack, prom metric labels | Sentinel-value tests (D-021) on all secret code paths; banned `%v`/`%+v` on `Secret` types; `String()` and `Error()` return `[REDACTED]` |
| **I** | Value leak via backups | Backups encrypted; backup destination credentials are themselves secrets (chicken-and-egg → bootstrap backup key separate from MEK, documented) |
| **D** | Mass enumeration of secret names | List endpoints paginate and audit; rate limit on `secrets:describe`; alert threshold on read_value spike |
| **E** | `describe` permission accidentally exposes value | Separate API methods for `describe` vs `read_value`; serializer omits ciphertext fields in `describe` response; contract test |

### 4.5 MEK lifecycle (Stage 4+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S** | Fake MEK substituted at startup | Each DEK row stores `key_id` (D-008); decrypt fails closed on mismatch instead of garbage output |
| **T** | MEK truncated / replaced silently | Startup verifies MEK against an "envelope canary" row (encrypted constant); refuse to start if mismatch |
| **R** | MEK rotation without record | `mek_id`, rotated_at, rotated_by stored in `mek_history` table; audit row on rotation |
| **I** | MEK in process memory dumped (core dump, swap) | `GOMEMLIMIT` + disable core dumps in container; `mlock` where supported (best-effort); zeroize key buffers after use |
| **D** | KMS provider unreachable | Cache wrapped DEK material in memory for service lifetime; degrade to read-only secret access if MEK unwrap fails for new DEKs |
| **E** | Anyone with DB access can decrypt with leaked MEK | MEK never lives in the DB; document operator separation |

### 4.6 GitOps reconcile — pull (Stage 5+)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S** | DNS hijack on Git provider host | TLS verification mandatory; SSH host key pinning for SSH remotes; document SSH known_hosts setup |
| **T** | Repo author pushes malicious compose | Out of scope to detect arbitrary intent; mitigate via RBAC (only allowed envs), required branch protection on the repo side, and operator-side max-resource limits |
| **R** | Reconcile attributed to no one | Reconcile rows carry `triggered_by: poll/webhook/manual:user_id` + git SHA |
| **I** | Logging compose env values during reconcile | Same as 4.3 |
| **I** | Storing repo creds in plaintext in DB | All repo credentials encrypted with MEK; never returned in API |
| **D** | Repo pull storm (large repo, frequent poll) | Per-app concurrency = 1 (single-flight); shallow clone where possible; per-app poll interval lower bound enforced |
| **D** | OOM via giant repo | Repo size cap (DEF-003); shallow clone; reject `.git` > N MB |
| **E** | Reconciler runs as root in container | Run as non-root uid 65532 (distroless); reconciler subprocess (if any) inherits the same uid |

### 4.7 Docker socket interaction (all stages)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **S/E** | Anyone who compromises the API gains effective host root | **Cannot be mitigated structurally without agent mode (DEF-008)**. Interim controls: socket-proxy (Tecnativa) restricts Docker API verbs; strong warning in README quick-start; recommended deployment behind a separate non-internet-facing reverse proxy |
| **T** | API call modified mid-flight | Local socket = local trust; remote TLS = `tls.Config` with verify-full |
| **R** | Docker action without app-side record | Every Docker call from Stowkeep originates from a handled request; audit row created before the call |
| **I** | Docker API responses leak unrelated container info to low-priv users | RBAC filters list responses to allowed scope; integration tests assert filtering |
| **D** | Slow / blocked Docker daemon hangs reconcilers | All Docker calls use `context.WithTimeout` (default 30s via `STOWKEEP_DOCKER_TIMEOUT`); UI banner when unreachable (D-013, Stage 1) |
| **E** | Compose declares privileged: true / host network | Validator rejects (or warns) on privileged primitives; admin-only override; this list maintained per Stowkeep release |

**Stage 1 shipped mitigations:** Session auth on all `/api/v1/swarm/*` routes; Docker list handlers log counts only (never full API bodies at info level); prominent UI banner when `/api/v1/swarm/status` reports disconnected; install guide documents raw socket-mount risk and socket-proxy interim path ([docs/install.md](../install.md)).

### 4.8 Database access (all stages)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **T** | Audit row deletion / edit by operator | Hash chain (D-011) makes tampering detectable on next startup verification |
| **I** | Plaintext leak via SQL logging | Driver-level query logging disabled in production; explicit allowlist of loggable queries |
| **I** | Backup file readable by anyone with volume access | Backup files written with `0600`; S3 backups uploaded with SSE; backup keys distinct from MEK |
| **D** | SQLite locked under contention | Per [3a in critique]: bounded write pool, BUSY retry, documented load limits |
| **E** | SQL injection | sqlc-generated typed queries only; no string-concatenated SQL in handlers; lint rule |

### 4.9 Backups (D-016)

| Threat | Concern | Mitigation |
|--------|---------|------------|
| **I** | Backup file = full secret store | Backups are encrypted before upload using a `BACKUP_KEY` (separate from MEK); restore requires both `MEK` and `BACKUP_KEY`; both documented as operator-managed assets |
| **T** | Tampered backup at restore time | Backup files signed with HMAC (key = `BACKUP_KEY`); restore refuses on mismatch |
| **R** | Restore attribution missing | Restore is an audit event; row written *after* the chain is verified |
| **D** | Backup OOM on huge DB | Streaming dump for Postgres; SQLite uses `Online Backup API` (constant memory) |
| **E** | UI lets a low-priv user trigger restore (overwrite of state) | Restore requires the `system:restore` permission; default = owner only; one-time confirmation token |

### 4.10 Webhooks (POST v1; placeholder)

When webhooks land (DEF-005), each provider's signing scheme must be implemented:

- GitHub: `X-Hub-Signature-256`, constant-time compare.
- GitLab: `X-Gitlab-Token`.
- Gitea: `X-Gitea-Signature` (HMAC-SHA256).
- Bitbucket: IP allowlist + (optional) basic auth.

Plus: replay window (reject deliveries older than 5 minutes), idempotency key per delivery, rate limit per source IP.

---

## 5. Defense-in-depth checklist (must hold at every release)

- [ ] All exported HTTP handlers go through the authz middleware (test asserted).
- [ ] No secret type can be `fmt.Sprintf` / `slog`-logged without `[REDACTED]` masking.
- [ ] No path returns a `Secret.Value` field outside the explicit `read_value` route.
- [ ] All Docker calls have a context timeout ≤ 30 s by default, configurable.
- [ ] `Authorization` header, `Cookie` header, `?token=`, `?key=`, `?secret=` query params are stripped from log output.
- [ ] Compose parser bounded by size and depth; YAML anchor expansion limited.
- [ ] sqlc is the only path that writes SQL.
- [ ] `audit_events` rows verify chain integrity on startup (warn loudly on break).
- [ ] CI runs `govulncheck`, `gosec`, `golangci-lint`, Trivy on image, `npm audit --omit=dev`.
- [ ] CI signs images with cosign and publishes SLSA provenance + Syft SBOM (D-022).
- [ ] Every PR touching `pkg/secrets`, `pkg/auth`, `pkg/rbac`, `migrations/` triggers CODEOWNERS review (D-023).

---

## 6. Known accepted risks (pre-1.0)

| Risk | Acceptance |
|------|------------|
| Docker socket mount = host root | Documented; mitigated via socket-proxy guidance and Stage 7 agent mode. |
| Single-maintainer release process | Pre-1.0 only. Releases marked "not for production third-party hosting" until external review. |
| `go-git` memory behavior on large repos | Bounded by repo size cap (DEF-003); revisit. |
| Casbin policy authoring UX | UI prototype required in Stage 2 (D-010); revisit if unusable. |
| No HSM/KMS integration in v1 | `MasterKeyProvider` interface in place (D-008); KMS impl is a Stage 7 effort. |
| No multi-replica HA | Stage 5 design effort (D-007). Single-replica restart is the only failover. |

---

## 7. Out of scope (explicitly)

- DDoS protection at the network layer — that's the operator's reverse proxy / CDN job.
- Securing the host the container runs on.
- Securing the Swarm cluster's overlay networks beyond what Docker provides.
- Vulnerabilities in deployed user workloads.
- Side-channel attacks on the host (Spectre/Meltdown class).

---

## 8. Updating this document

A PR that adds or changes any of the following **must** update this file in the same PR (CODEOWNERS will enforce review):

- A new HTTP route or background worker that handles secrets, auth, RBAC, audit, Docker, or Git.
- A new external dependency that handles network, crypto, or persistence.
- A new trust boundary (new external service, new ingress, new auth flow).
- Any change to the secret encryption, key handling, or backup format.

Reviewers check that the STRIDE table for the affected flow has been considered and that any new "known accepted risk" is called out in §6.

---

## References

- [OWASP ASVS L2](https://owasp.org/www-project-application-security-verification-standard/) — aspirational target for v1.
- [SECURITY.md](../../SECURITY.md) — coordinated disclosure policy.
- [ADR 0001](../adr/0001-dual-database-sqlite-postgres.md), [ADR 0002](../adr/0002-structured-logging-slog.md) — current ADRs.
- [planning/decisions-todo.md](../../planning/decisions-todo.md) — decisions tracker.
