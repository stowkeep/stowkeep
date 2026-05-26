# Stowkeep

Self-hosted control plane for **Docker Swarm** — cluster UI, GitOps, secrets, RBAC, and preview environments.

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

> _Stow_ — to put something in a place where it can be kept safely.
> _Keep_ — to maintain; also, a castle's fortified inner tower.
> **Stowkeep** is the keeper of what you have stowed — your stacks, secrets, GitOps state, audit log, and backups.

## What it does

Stowkeep connects to the Docker Engine (via socket or TLS) and gives teams a modern way to run Swarm without giving up GitOps or secrets best practices:

- **Swarm dashboard** — nodes, services, stacks, tasks, logs
- **Stack deploy** — validated Compose → Swarm stacks
- **GitOps** — pull-based sync from Git (webhooks + polling), inspired by [Doco-CD](https://doco.cd/)
- **Secrets** — encrypted, versioned, RBAC-gated, inspired by [Infisical](https://infisical.com/)
- **Preview environments** — ephemeral PR stacks, inspired by [Dokploy](https://dokploy.com/)
- **RBAC** — users, groups, roles, fine-grained permissions

## Quick start (SQLite — minimal dependencies)

Single container, SQLite on a Docker volume. Ideal for homelab and evaluation.

```bash
docker run -d \
  --name stowkeep \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v stowkeep-data:/data \
  -e STOWKEEP_DATABASE_DRIVER=sqlite \
  -e STOWKEEP_DATABASE_PATH=/data/stowkeep.db \
  -e STOWKEEP_FEATURES=swarm_readonly \
  ghcr.io/stowkeep/stowkeep:latest
```

Open `http://localhost:8080` and complete first-run setup.

> ⚠️ **About the Docker socket mount.** Mounting `/var/run/docker.sock` — even read-only — grants the container the ability to control the Docker daemon, which is effectively root on the host. The `:ro` flag restricts the filesystem mount, not the Docker API actions. This is acceptable for a single-tenant homelab where you trust the workload, but **for any production install you should put a [Docker socket proxy](https://github.com/Tecnativa/docker-socket-proxy) between Stowkeep and the socket** to restrict the allowed API verbs. First-party agent mode (no socket exposure on the host) is targeted for Stage 7 — see [planning/decisions-todo.md](planning/decisions-todo.md) D-012 / D-027.

> **Production:** Use [PostgreSQL](docs/database.md) for multi-user and high-throughput deployments.

## Documentation

| Topic                     | Link                                                           |
| ------------------------- | -------------------------------------------------------------- |
| Install & first-run       | [docs/install.md](docs/install.md)                             |
| Development setup         | [docs/development-guide.md](docs/development-guide.md)         |
| SQLite vs PostgreSQL      | [docs/database.md](docs/database.md)                           |
| Production logging        | [docs/logging.md](docs/logging.md)                             |
| Security threat model     | [docs/security/threat-model.md](docs/security/threat-model.md) |
| Contributing              | [CONTRIBUTING.md](CONTRIBUTING.md)                             |
| Code standards            | [docs/code-standards.md](docs/code-standards.md)               |
| Product roadmap           | [planning/PRD.md](planning/PRD.md)                             |
| Phase gates (quality bar) | [planning/phase-gates.md](planning/phase-gates.md)             |
| Stage 0 gate checklist  | [docs/stage-0-gate.md](docs/stage-0-gate.md)                   |
| Governance              | [GOVERNANCE.md](GOVERNANCE.md)                                 |
| AI agent instructions     | [AGENTS.md](AGENTS.md)                                         |
| Project decisions tracker | [planning/decisions-todo.md](planning/decisions-todo.md)       |
| Security                  | [SECURITY.md](SECURITY.md)                                     |

## Project status

> **Pre-alpha — not for production hosting of workloads or secrets you cannot afford to lose.** Stage 1 ships a read-only Swarm dashboard with local admin auth. See [docs/install.md](docs/install.md) for setup.

## Contributing

Ideas, issues, and feedback are welcome via [GitHub Issues](https://github.com/stowkeep/stowkeep/issues). PR-based contributions are not actively accepted during pre-alpha; once we reach `v0.1.0` the workflow in [CONTRIBUTING.md](CONTRIBUTING.md), [phase gates](planning/phase-gates.md), and [open source standards](planning/open-source-standards.md) take effect. **Every stage requires testing, documentation, and code quality sign-off before the next stage begins.**

## License

Apache License 2.0 — see [LICENSE](LICENSE).
