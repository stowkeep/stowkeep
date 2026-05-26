# 0009. Anonymous opt-in telemetry posture

- **Status:** Accepted
- **Date:** 2026-05-25
- **Tracker entry:** [D-030](../../planning/decisions-todo.md)

## Context

Self-hosted software has a fundamental visibility problem: maintainers cannot tell how many people are using it, on what platforms, with which features. Without any telemetry we are making product decisions in the dark.

But telemetry is also the most reliable way to destroy trust. Defaulting on, collecting PII, being unclear about what's sent, exposing data to third parties, or making it hard to turn off — any of these will rightly cost us the audience that picked us for self-hostability in the first place.

The default-off / opt-in / minimal-payload / configurable-endpoint pattern has been adopted by Grafana, GoToSocial, Forgejo, Plausible, and others, and is the only stance that survives community scrutiny in 2026.

This posture is locked **now**, before any code is written, so that no contributor or maintainer can later quietly broaden collection without an explicit ADR superseding this one.

## Decision

**Default:** off. No telemetry of any kind unless an admin opts in during first-run setup.

**Payload (when opted in), sent once per week:**

| Field | Example | Why |
|-------|---------|-----|
| `install_id` | `7f3c2a1b-4e5d-6f7a-8b9c-0d1e2f3a4b5c` | Locally-generated UUID at first run; never reset, never tied to any user identity |
| `version` | `v0.4.2` | What version is in the field |
| `db_driver` | `sqlite` \| `postgres` | Helps prioritize backend work |
| `os` | `linux` | Platform mix |
| `arch` | `amd64` \| `arm64` \| `arm` | Confirms arm/v7 demand (D-018) |
| `enabled_features` | `["gitops","secrets"]` | Which feature flags are on — guides stage prioritization |
| `swarm_version` | `24.0.7` (optional) | The Docker Engine version we're talking to — affects compatibility |

**Never collected, ever:**

- Project names, repository URLs, branch names, stack names.
- User emails, usernames, IDs.
- Node counts, service counts, secret counts, container counts.
- IP addresses (the receiving collector strips the source IP from logs).
- Anything from compose file contents.
- Any free-text the user has entered.

**Transport:** HTTPS POST to a configurable endpoint URL (default: a maintainer-operated collector; clearly named so it can be allow-listed or blocked at the firewall). Setting the endpoint to an empty string disables telemetry even if the opt-in flag is on.

**Collector source code is open** in the same repository (`tools/telemetry-collector/`). The collector logs only the fields above; aggregate dashboards are published.

**Transparency surfaces:**

- A dedicated `docs/telemetry.md` documenting exactly what is sent, when, and how to turn it off.
- A line in the first-run setup wizard with a link to `docs/telemetry.md`.
- A line in `README.md` linking to `docs/telemetry.md`.
- A line in `--help` output of the binary.

**Implementation timing:** posture is locked now. Code lands in Stage 7 — there is no need for telemetry during alpha/beta when we know the operators by name.

## Consequences

**Easier**

- Trust posture is defensible from day 1; the conversation with privacy-conscious users is "here is the full payload, here is the source, here is the off switch."
- No GDPR/CCPA/PIPL exposure under any reasonable reading (no PII, no behavioral data).
- Users can self-host the collector or block it entirely.

**Harder**

- We get less data than products that default on or collect richer payloads. If we later want crash reports, feature-usage counts, or error summaries, those need separate opt-in consent flows with their own ADRs — we cannot quietly broaden this one.
- The maintainer-operated collector is operational work we now own (a small Go service, a database, dashboards). Acceptable, scoped, and shippable in Stage 7.
- Once published, the payload list above is effectively a contract with the community. Adding a new field requires an ADR and a release-notes call-out.

## References

- [planning/decisions-todo.md](../../planning/decisions-todo.md) — `D-030`
- [planning/PRD.md §11, §12](../../planning/PRD.md)
- [docs/security/threat-model.md §6](../security/threat-model.md) — pre-1.0 caveats and accepted risks
