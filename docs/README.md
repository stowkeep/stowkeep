# Documentation

User and contributor documentation for [Stowkeep](../README.md).

## Getting started

| Document | Description |
|----------|-------------|
| [development-guide.md](./development-guide.md) | Local setup, commands, debugging |
| [database.md](./database.md) | SQLite (embedded) vs PostgreSQL (production) |
| [logging.md](./logging.md) | Production logging — levels, JSON format, shipping to Loki/ELK |

## Security

| Document | Description |
|----------|-------------|
| [security/threat-model.md](./security/threat-model.md) | STRIDE threat model for login, deploy, secrets, GitOps, Docker socket, and backups |
| [../SECURITY.md](../SECURITY.md) | Vulnerability reporting policy |

## Contributing

| Document | Description |
|----------|-------------|
| [code-standards.md](./code-standards.md) | Go, TypeScript, SQL, and API conventions |
| [../CONTRIBUTING.md](../CONTRIBUTING.md) | Contribution workflow |
| [../AGENTS.md](../AGENTS.md) | Instructions for AI coding agents |
| [../planning/phase-gates.md](../planning/phase-gates.md) | Mandatory testing + docs + quality bar between stages |
| [../planning/open-source-standards.md](../planning/open-source-standards.md) | Per-PR testing, docs, and review expectations |

## Planning (maintainers)

Product and architecture planning lives in [../planning/](../planning/).

## Future docs (as features ship)

- `docs/install.md` — installation paths (SQLite quick-start, Postgres production)
- `docs/configuration.md` — environment variables reference
- `docs/stowkeep-yml.md` — GitOps manifest reference
- `docs/rbac.md` — roles and permissions cookbook
- `docs/secrets.md` — secrets encryption and rotation
- `docs/upgrade/` — version upgrade guides
