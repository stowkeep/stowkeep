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
cosign verify ghcr.io/stowkeep/stowkeep:main
cosign verify-attestation --type slsaprovenance ghcr.io/stowkeep/stowkeep:main
docker pull ghcr.io/stowkeep/stowkeep:main
docker run --rm -p 8080:8080 \
  -v stowkeep-data:/data \
  -e STOWKEEP_DATABASE_PATH=/data/stowkeep.db \
  ghcr.io/stowkeep/stowkeep:main
curl -sf http://localhost:8080/healthz
```

## Open gate issue

```bash
gh issue create \
  --title "Phase gate: Stage 0 complete" \
  --template phase_gate.yml
```

Fill in CI run URL, merged PR link, and manual test notes. Maintainer sign-off on all three pillars (testing, documentation, code quality) unblocks Stage 1.
