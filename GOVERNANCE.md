# Governance

**Project:** Stowkeep  
**Last updated:** 2026-05-25

This document describes how Stowkeep is governed, how decisions are made, and how maintainership may evolve over time.

---

## Current status

Stowkeep is **pre-alpha** and maintained by a **single maintainer** (BDFL model) until **v0.5.0**. We are not actively seeking contributors yet, but feedback and issue reports are welcome. See [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md) for the current contributor policy.

---

## Decision-making

| Type | Process |
|------|---------|
| **Product scope** | [planning/PRD.md](planning/PRD.md) and stage phase gates |
| **Technical architecture** | ADRs in [docs/adr/](docs/adr/) |
| **Day-to-day decisions** | Maintainer discretion; recorded in [planning/decisions-todo.md](planning/decisions-todo.md) |
| **Breaking changes** | ADR + CHANGELOG + migration notes before merge |

Significant decisions get a `D-NNN` entry in [decisions-todo.md](planning/decisions-todo.md). Architectural decisions also get an ADR.

---

## Maintainer responsibilities

The maintainer is responsible for:

- Keeping `main` deployable and CI green
- Enforcing [phase gates](planning/phase-gates.md) between stages
- Reviewing changes to security-sensitive paths (see [.github/CODEOWNERS](.github/CODEOWNERS))
- Triaging issues and security reports per [SECURITY.md](SECURITY.md)
- Walking [decisions-todo.md](planning/decisions-todo.md) before each release

---

## Hand-off plan

When the project reaches **v0.5.0** or the maintainer can no longer serve:

1. Publish intent to add co-maintainers via GitHub Discussions
2. Update this document with named maintainers and review expectations
3. Require two approvals on security-sensitive paths (already noted in CODEOWNERS)
4. Transition from BDFL to consensus-based merge policy for non-security changes

Until then, the maintainer may delegate review to trusted contributors on a case-by-case basis without changing this document.

---

## Related documents

| Document | Role |
|----------|------|
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |
| [planning/decisions-todo.md](planning/decisions-todo.md) | Decision index |
| [docs/adr/README.md](docs/adr/README.md) | Architecture decisions |
| [planning/phase-gates.md](planning/phase-gates.md) | Stage completion bar |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
