# Contributing to Stowkeep

Thank you for your interest in contributing. Stowkeep aims to be **as open source as open source gets** — clear docs, tested code, and a welcoming community.

## Before you start

1. Read [AGENTS.md](AGENTS.md) if you are an automated coding agent
2. Read the [development guide](docs/development-guide.md)
3. Review [code standards](docs/code-standards.md)
4. Read [phase gates](planning/phase-gates.md) — **do not start a new stage until the previous gate passes**
5. Skim [open source standards](planning/open-source-standards.md) for testing and documentation expectations
6. Check [open issues](https://github.com/stowkeep/stowkeep/issues) or open one to discuss large changes

## Ways to contribute

- Bug reports and feature requests (issue templates)
- Documentation improvements
- Code — features, fixes, tests
- Reviews on pull requests
- Sharing deployment experience in Discussions

## Development workflow

We use **trunk-based development** on `main`:

1. Fork the repository
2. Create a short-lived branch: `feat/`, `fix/`, `docs/`, `chore/`
3. Make focused changes with tests
4. Run `make lint && make test` locally
5. Open a PR against `main`
6. Address review feedback
7. Squash or merge per maintainer preference (we prefer merge commits for clear history, or squash for small PRs)

See [planning/engineering-and-devops.md](planning/engineering-and-devops.md) for branching and CI details.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(stacks): add compose validation on deploy
fix(auth): expire sessions after idle timeout
docs(database): document SQLite backup procedure
test(secrets): assert plaintext never appears in logs
```

## Pull request checklist

- [ ] Tests added or updated for behavior changes
- [ ] `make lint && make test` passes locally
- [ ] Godoc / TSDoc on new exported APIs
- [ ] [CHANGELOG.md](CHANGELOG.md) updated for user-visible changes
- [ ] Docs updated if install, config, or contributor workflow changed
- [ ] Migrations tested on **SQLite and PostgreSQL** (if applicable)
- [ ] No secrets, credentials, or `.env` files in the diff

## Code review

All submissions require review. We aim for constructive, timely feedback. Maintainers may request changes to align with [code standards](docs/code-standards.md) or security requirements — especially for `pkg/secrets`, `pkg/rbac`, and `pkg/auth`.

## Testing expectations

| Change | Minimum |
|--------|---------|
| Bug fix | Regression test |
| API endpoint | Handler + RBAC deny test |
| Migration | Applies on SQLite and Postgres in CI |
| UI logic | Vitest component test |

Full policy: [planning/testing-strategy.md](planning/testing-strategy.md).

## Stage completion (phase gates)

Features merge to `main` throughout a stage, but **the stage is not done** until the [phase gate checklist](planning/phase-gates.md) passes:

| Pillar | What must be true |
|--------|-------------------|
| **Testing** | CI green; stage test plan + manual smoke complete |
| **Documentation** | User docs, CHANGELOG, godoc/TSDoc, ADRs/threat model updated |
| **Code quality** | Lint clean; reviews done; dual-DB migrations; no critical CVEs |

Open a gate issue titled `Phase gate: Stage N complete` and get maintainer sign-off before starting the next stage.

## Documentation

- User-facing → `docs/`
- Planning / architecture → `planning/`
- Code comments → godoc (Go), TSDoc (TypeScript)

Update docs in the same PR as code when behavior changes.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Participants are expected to uphold it.

## Security

Do **not** open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).

## Questions

- **How do I…?** — [GitHub Discussions](https://github.com/stowkeep/stowkeep/discussions)
- **Bug / feature** — [GitHub Issues](https://github.com/stowkeep/stowkeep/issues)

Replace `stowkeep` with the actual GitHub organization when published.
