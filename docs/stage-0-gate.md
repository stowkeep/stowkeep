# Stage 0 Phase Gate — Sign-off Checklist

Use this document when opening the **Phase gate: Stage 0 complete** GitHub issue after merging the Stage 0 PR to `main`.

Copy the checklist from [planning/phase-gates.md](../planning/phase-gates.md#stage-0--foundation) and attach evidence links.

## Pre-merge verification (local)

```bash
make lint && make test
make test-migrations-sqlite
# optional: docker compose -f docker-compose.dev.yml up -d postgres && make test-migrations-postgres
make docker-build
docker run --rm -p 8080:8080 \
  -e STOWKEEP_DATABASE_PATH=/tmp/stowkeep.db \
  stowkeep:local &
curl -sf http://localhost:8080/healthz
```

## Post-merge verification (maintainer)

After CI on `main` is green and `release-image.yml` has pushed to GHCR:

```bash
docker pull ghcr.io/stowkeep/stowkeep:main
docker run --rm -p 8080:8080 \
  -v stowkeep-data:/data \
  -e STOWKEEP_DATABASE_PATH=/data/stowkeep.db \
  ghcr.io/stowkeep/stowkeep:main
curl -sf http://localhost:8080/healthz
```

Release images are signed with **keyless cosign** from GitHub Actions. Verification must pin the workflow identity:

```bash
COSIGN_IDENTITY='^https://github.com/stowkeep/stowkeep/\.github/workflows/release-image\.yml@refs/heads/main$'
COSIGN_ISSUER='https://token.actions.githubusercontent.com'
IMAGE='ghcr.io/stowkeep/stowkeep:main'

# Image signature (cosign sign step)
cosign verify \
  --certificate-identity-regexp="${COSIGN_IDENTITY}" \
  --certificate-oidc-issuer="${COSIGN_ISSUER}" \
  "${IMAGE}"

# SLSA provenance attestation (actions/attest-build-provenance)
cosign verify-attestation \
  --type 'https://slsa.dev/provenance/v1' \
  --certificate-identity-regexp="${COSIGN_IDENTITY}" \
  --certificate-oidc-issuer="${COSIGN_ISSUER}" \
  "${IMAGE}"
```

## Open gate issue

The phase gate template (`.github/ISSUE_TEMPLATE/phase_gate.yml`) is a **YAML form** — `gh issue create --template` only works with markdown templates. Use one of:

**Browser form (recommended):**

```bash
gh issue create -w
# pick "Phase gate sign-off", or open:
# https://github.com/stowkeep/stowkeep/issues/new?template=phase_gate.yml
```

**CLI with body** (labels must exist first — one-time bootstrap):

```bash
gh label create "phase-gate" \
  --description "Stage completion sign-off — blocks next stage until closed" \
  --color "5319E7" 2>/dev/null || true
gh label create "triage" \
  --description "Needs maintainer review or prioritization" \
  --color "FBCA04" 2>/dev/null || true

gh issue create \
  --title "Phase gate: Stage 0 complete" \
  --label "phase-gate,triage" \
  --body "See planning/phase-gates.md § Stage 0 — attach CI run URL, merged PR link, and manual test notes."
```

Fill in CI run URL, merged PR link, and manual test notes. Maintainer sign-off on all three pillars (testing, documentation, code quality) unblocks Stage 1.
