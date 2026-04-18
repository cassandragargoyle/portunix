# Contributing to Portunix

Thank you for considering a contribution to Portunix. This file is the
entry point — the full set of guidelines lives under
[`docs/contributing/`](docs/contributing/).

## Quick Start

1. **Pick or open an issue** — browse
   [open issues](https://github.com/cassandragargoyle/portunix/issues) or
   open a new one describing the change you want to make. Non-trivial
   changes should have an issue first so the scope can be agreed on before
   code is written.

2. **Fork and clone**

   ```bash
   git clone https://github.com/<your-user>/portunix.git
   cd portunix
   ```

3. **Create a feature branch** following the project convention:

   ```bash
   git checkout -b feature/issue-<number>-<short-name>
   ```

4. **Set up the dev environment**

   ```bash
   ./scripts/dev-setup.sh          # Linux / macOS
   scripts\dev-setup.ps1           # Windows
   ```

5. **Build and test**

   ```bash
   make build                      # main binary + all helpers
   make test                       # full suite
   make test-unit                  # unit tests only (fast)
   make lint                       # golangci-lint + gofmt -s
   ```

6. **Commit and push** using Conventional-Commit-style messages
   (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:` …).

7. **Open a Pull Request** from your feature branch into `main`.
   The CI pipeline will run lint, unit tests, integration tests, a
   security scan, and a cross-platform build before the PR can be merged.

## Ground Rules

- All code, comments, commit messages, and PR descriptions are written in
  English.
- Follow existing conventions — explore the codebase before introducing new
  patterns.
- Write tests for new features and bug fixes.
  See [TEST_GUIDE.md](TEST_GUIDE.md) and
  [`docs/contributing/TESTING_METHODOLOGY.md`](docs/contributing/TESTING_METHODOLOGY.md).
- Software-installation tests **must** run in containers, never directly on
  the host — see the container-based testing policy in
  [`docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md`](docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md).
- Do not add co-authored-by attributions for AI tools in commits.
- `CHANGELOG.md` is kept current — add an entry under the target version when
  landing a user-visible change.

## Detailed Guidelines

The `docs/contributing/` directory contains the full set of project
standards. A few of the most frequently referenced documents:

- [README](docs/contributing/README.md) — index of all contributing docs
- [Issue Development Methodology](docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md) —
  the mandatory issue → branch → test → merge flow
- [Testing Methodology](docs/contributing/TESTING_METHODOLOGY.md) — the
  TestFramework, verbose mode, container-based testing rules
- [Go Code Style](docs/contributing/CODE-STYLE-GO.md),
  [Python Code Style](docs/contributing/CODE-STYLE-PYTHON.md),
  [Markdown Style](docs/contributing/MARKDOWN-STYLE.md)
- [Git Workflow](docs/contributing/GIT-WORKFLOW.md) and
  [GitHub Workflow](docs/contributing/GITHUB-WORKFLOW.md)
- [Bug Reporting](docs/contributing/BUG-REPORTING.md)
- [Helper Binary Development](docs/contributing/HELPER-BINARY-DEVELOPMENT.md) —
  required checklist when adding a new `ptx-*` helper
- [Versioning](docs/contributing/VERSIONING.md)
- [Terminology](docs/contributing/TERMINOLOGY.md)

## Reporting Bugs and Requesting Features

- **Bugs** — open an issue using the bug reporting guidelines in
  [`docs/contributing/BUG-REPORTING.md`](docs/contributing/BUG-REPORTING.md).
  Include Portunix version (`portunix --version`), OS, steps to reproduce,
  expected vs. actual behavior, and relevant logs.
- **Feature requests** — open an issue describing the use case and the
  motivation. For substantial changes, an ADR in `docs/adr/` may be
  proposed alongside the feature request.

## Security

For security-sensitive reports, please do not open a public issue.
Instead, contact the maintainers privately — see the
[CassandraGargoyle team](https://github.com/cassandragargoyle) page for
contact options.

## License

By contributing to Portunix you agree that your contributions will be
licensed under the [MIT License](LICENSE).
