# Install Stowkeep

Stowkeep is a self-hosted control plane for Docker Swarm. Stage 1 ships a **read-only Swarm dashboard** with local admin authentication.

---

## Quick start (SQLite)

**Warning:** Mounting `/var/run/docker.sock` gives Stowkeep effective root on the host. For production, place a [Docker socket proxy](https://github.com/Tecnativa/docker-socket-proxy) between Stowkeep and the engine, or wait for Stage 7 agent mode.

```bash
docker run -d --name stowkeep \
  -p 8080:8080 \
  -v stowkeep-data:/data \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e STOWKEEP_DATABASE_PATH=/data/stowkeep.db \
  -e STOWKEEP_FEATURES=swarm_readonly \
  ghcr.io/stowkeep/stowkeep:latest
```

Open `http://localhost:8080` and complete first-run admin setup.

---

## First-run bootstrap

1. Visit the UI — if no users exist, you are redirected to **Setup**.
2. Create the admin email and password (minimum 8 characters).
3. Sign in and confirm **Settings → Test connection** reports Docker reachable.
4. Browse **Nodes**, **Services**, **Tasks**, and **Stacks**.

Logout clears the session cookie.

---

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `STOWKEEP_HTTP_ADDR` | `:8080` | HTTP listen address |
| `STOWKEEP_DATABASE_PATH` | `/data/stowkeep.db` | SQLite file (quick start) |
| `STOWKEEP_DATABASE_URL` | — | PostgreSQL DSN (production) |
| `STOWKEEP_DOCKER_HOST` | `unix:///var/run/docker.sock` | Docker Engine endpoint |
| `STOWKEEP_DOCKER_TIMEOUT` | `30s` | Engine API call timeout |
| `STOWKEEP_FEATURES` | — | Comma-separated capability flags (see below) |
| `STOWKEEP_SESSION_IDLE_TTL` | `24h` | Session lifetime |
| `STOWKEEP_COOKIE_SECURE` | `false` | Set `true` behind HTTPS |

See [`.env.example`](../.env.example) for local development and [development-guide.md](./development-guide.md) for `.env` loading during `make dev`.

---

## Feature flags

Feature flags are **operator configuration** set via `STOWKEEP_FEATURES` at container/process startup. They are not toggled from the UI.

| Flag | Stage | Enables |
|------|-------|---------|
| `swarm_readonly` | 1 | Swarm dashboard API and navigation |
| `stack_deploy` | 2 | Stack deploy, remove, scale, and logs API |
| `rbac` | 3 | Multi-user RBAC (future) |
| `secrets` | 4 | Secret store (future) |

When a flag is disabled, the API returns `404 feature_disabled` on gated routes and the UI hides the related navigation and actions.

Example (read-only dashboard only):

```bash
-e STOWKEEP_FEATURES=swarm_readonly
```

Example (Stage 2 — deploy stacks):

```bash
-e STOWKEEP_FEATURES=swarm_readonly,stack_deploy
```

---

## Docker socket options

| Mode | `STOWKEEP_DOCKER_HOST` | Notes |
|------|------------------------|-------|
| Local socket (homelab) | `unix:///var/run/docker.sock` | Simplest; highest risk |
| Remote TLS | `tcp://host:2376` | Requires TLS client config on the engine |
| Socket proxy (recommended interim) | Proxy container URL | Restricts allowed Docker API verbs |

Stage 1 does **not** expose runtime socket reconfiguration in the UI — change the environment variable and restart.

---

## Verifying the install

```bash
curl -sf http://localhost:8080/healthz
curl -sf http://localhost:8080/api/v1/setup/status
```

After bootstrap, Swarm endpoints require a session cookie:

```bash
# Example after logging in via browser — use cookie from DevTools
curl -sf -b stowkeep_session=... http://localhost:8080/api/v1/swarm/status
```

Compare node and service counts with:

```bash
docker node ls
docker service ls
docker stack ls
```

---

## Related docs

- [Development guide](./development-guide.md)
- [Database guide](./database.md)
- [Security threat model](./security/threat-model.md)
