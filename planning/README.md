# Stowkeep — Planning

This directory contains product and engineering planning for **Stowkeep**, a self-hosted control plane for Docker Swarm that combines cluster management, GitOps, secrets, RBAC, and preview environments.

## Documents

| Document | Purpose |
|----------|---------|
| [PRD.md](./PRD.md) | Product vision, requirements, personas, and phased delivery roadmap |
| [tech-stack.md](./tech-stack.md) | Technology choices with rationale and alternatives considered |
| [engineering-and-devops.md](./engineering-and-devops.md) | Naming, repository setup, CI/CD, trunk-based development, release process |
| [testing-strategy.md](./testing-strategy.md) | Testing pyramid, frameworks, environments, and quality gates |
| [open-source-standards.md](./open-source-standards.md) | Documentation, comments, testing, and review expectations |
| [phase-gates.md](./phase-gates.md) | **Mandatory** testing + docs + code quality bar between stages |
| [decisions-todo.md](./decisions-todo.md) | Decisions tracker — decided, deferred, open foundational choices |

## Contributor & user docs

See [../docs/README.md](../docs/README.md) for development guide, code standards, and database guide.

## Inspirations

| Product | What we borrow |
|---------|----------------|
| [Doco-CD](https://doco.cd/) | Pull-based GitOps for Compose/Swarm; polling + webhooks; SOPS/external secrets |
| [Portainer](https://www.portainer.io/) | Swarm-first UX; stack/service management; multi-environment model |
| [Dokploy](https://dokploy.com/) | Preview/ephemeral environments per PR; templated domains; CI-built image deploy |
| [Infisical](https://infisical.com/) | Secret versioning; subject-action RBAC with conditions; approval workflows; audit logs |

## Recommended build order

1. **Stage 0** — Repo bootstrap, CI, container skeleton *(engineering-and-devops.md)*
2. **Stage 1** — Docker socket connection + Swarm read-only dashboard
3. **Stage 2** — Stack/service deploy + basic auth
4. **Stage 3** — RBAC + audit logging
5. **Stage 4** — Secrets with versioning and encryption
6. **Stage 5** — GitOps pull-based sync
7. **Stage 6** — Preview/ephemeral environments
8. **Stage 7** — Hardening, HA, and enterprise features

See [PRD.md](./PRD.md) for full scope, acceptance criteria, and out-of-scope items.

## Quality enforcement

- **Every PR:** [open-source-standards.md](./open-source-standards.md)
- **Every stage boundary:** [phase-gates.md](./phase-gates.md) — do not start Stage N+1 until Stage N gate passes
- **AI agents:** [AGENTS.md](../AGENTS.md)
