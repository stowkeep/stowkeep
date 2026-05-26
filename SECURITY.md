# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| latest release | Yes |
| main branch | Best-effort (pre-1.0) |
| older releases | Security fixes at maintainer discretion |

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Report privately to **contact@stowkeep.dev**. (A dedicated `security@stowkeep.dev` inbox is planned post-`v0.1.0` so that vulnerability reports can be routed to a tighter audience; tracked in [planning/decisions-todo.md](planning/decisions-todo.md).) Include:

- Description of the vulnerability
- Steps to reproduce
- Impact assessment (e.g. secret disclosure, RBAC bypass, RCE)
- Affected version or commit

We aim to acknowledge reports within **3 business days** and provide an initial assessment within **7 business days**.

## What we consider in scope

- Authentication and session handling bypass
- RBAC / authorization bypass
- Secret plaintext disclosure (logs, API, audit, database)
- SQL injection or unsafe deserialization
- Remote code execution via Docker socket abuse or unsafe compose handling
- Cross-site scripting or CSRF in the web UI
- Supply chain issues in official release images

## Out of scope

- Denial of service without demonstrated security impact
- Issues requiring physical access to the host
- Misconfiguration by operators (e.g. exposing Docker socket publicly without TLS)
- Vulnerabilities in third-party Swarm workloads deployed through the platform

## Security practices (project)

- Dependencies scanned in CI (Dependabot, Trivy/gosec)
- Secrets never logged or stored in audit payloads
- Envelope encryption for secret values at rest
- Dual database support tested in CI; no dialect-specific security shortcuts
- Coordinated disclosure preferred; credit given in CHANGELOG unless you prefer anonymity

## Safe harbor

We support good-faith security research on your own installations. Do not access systems or data you do not own or lack permission to test.

Thank you for helping keep Stowkeep and its users safe.
