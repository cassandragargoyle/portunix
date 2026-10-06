# Issue #194: `make test` Does Not Cover Non-root Go Modules (Helpers, src/app, src/cmd)

## 🎯 Priority

**HIGH** — CI/local quality gate gap. `make test` silently skips most of the
codebase: 15 of 19 helper binaries and the `src/app`, `src/cmd`, `src/parser`
modules are never tested, even though many of them contain `_test.go` files.
A developer running `make test` (as required by the issue development
methodology) gets a green result without those tests ever executing.

## 📋 Status

- **Created**: 2026-07-07
- **Status**: ✅ Implemented
- **Closed**: 2026-07-08
- **Assignee**: -
- **Branch**: feature/issue-194-make-test-multi-module-coverage
- **Related**:
  - `Makefile` — `test`, `test-cli`, `test-coverage`, `test-coverage-ci`,
    `test-unit`, `test-report` targets
  - `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` — mandates
    `go test ./...` as pre-testing validation

## 📝 Problem Description

### Current Situation

The repository is multi-module (no `go.work`). The root module
`portunix.ai/portunix` references `src/app` only via a `replace` directive,
and most helper binaries have their own `go.mod`:

```text
go.mod                        # root module — the ONLY one make test runs
src/app/go.mod                # + submodules virt, github, wizard, sandbox
src/cmd/go.mod
src/parser/go.mod
src/helpers/ptx-<name>/go.mod # 15 of 19 helpers are separate modules
```

`make test` runs `go test -v ./...` from the repository root. In Go, `./...`
stops at nested `go.mod` boundaries, so only the 26 packages of the root
module are tested. Verified with `go list ./...`:

- **Helpers covered** (part of root module): `ptx-aiops`, `ptx-make`,
  `ptx-pft`, `ptx-virt` — only 4 of 19
- **Helpers NOT covered** (own `go.mod`): `ptx-ansible`, `ptx-container`,
  `ptx-credential`, `ptx-database`, `ptx-github`, `ptx-installer`,
  `ptx-mcp`, `ptx-plugin-registry`, `ptx-prompting`, `ptx-proxmox`,
  `ptx-python`, `ptx-specpm`, `ptx-ssh`, `ptx-trace`, `ptx-wizard`
- **Core modules NOT covered**: `src/app` (+ `virt`, `github`, `wizard`,
  `sandbox` submodules), `src/cmd`, `src/parser`

### Silently Skipped Tests

Existing `_test.go` files that `make test` never runs:

| Module | Test files |
| ------ | ---------- |
| `src/helpers/ptx-installer` | 12 |
| `src/helpers/ptx-github` | 6 |
| `src/helpers/ptx-proxmox` | 6 |
| `src/helpers/ptx-ssh` | 4 |
| `src/helpers/ptx-container` | 2 |
| `src/helpers/ptx-credential` | 2 |
| `src/helpers/ptx-database` | 2 |
| `src/helpers/ptx-specpm` | 2 |
| `src/app` | 19 |
| `src/cmd` | 7 |
| `src/app/wizard` | 3 |
| `src/app/github` | 1 |

In total **66 test files** exist outside the root module and are skipped by
`make test`, `test-coverage`, `test-coverage-ci`, `test-unit`, and
`test-report` (all use `./...` from root).

### Additional Defect

`test-cli` runs `go test -v ./cmd/...`, but no `cmd/` directory exists at the
repository root (CLI commands live in the separate `src/cmd` module) — the
target matches no packages.

## ✅ Acceptance Criteria

1. `make test` runs unit tests of ALL Go modules in the repository:
   root, `src/app` (incl. submodules), `src/cmd`, `src/parser`, and every
   `src/helpers/ptx-*` module with its own `go.mod`
2. Module discovery is dynamic (e.g., `find`/`go work` based) so newly added
   helpers are picked up without editing the Makefile — consistent with the
   helper checklist in `docs/contributing/HELPER-BINARY-DEVELOPMENT.md`
3. A failing test in any module fails `make test` (non-zero exit code)
4. `test-coverage` / `test-coverage-ci` aggregate coverage across modules,
   or explicitly document root-module-only scope
5. `test-cli` is fixed to point at `src/cmd` (or removed if redundant)
6. `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` checklist mentions test
   coverage expectations for new helper modules (if not already covered by
   dynamic discovery)
7. Cross-platform: works on Linux and Windows (Git Bash/WSL) build
   environments

## 💡 Implementation Notes

Two candidate approaches:

1. **Makefile loop over modules** (no repo-layout change):

   ```make
   GO_MODULES := $(shell find . -name go.mod -not -path './docs/*' \
       -not -path './docs-site/*' -not -path './.git/*' \
       -exec dirname {} \;)

   test:
   	@set -e; for mod in $(GO_MODULES); do \
   		echo "==> go test $$mod"; \
   		(cd $$mod && go test ./...); \
   	done
   ```

2. **Introduce `go.work`** listing all modules, then a single
   `go test ./...` covers everything. More invasive: affects builds,
   GoReleaser, and release scripts — needs an architect decision (ADR).

Approach 1 is the minimal, low-risk fix; approach 2 can be a follow-up.

Exclusions to keep: `docs-site/themes/hugo-book` (vendored theme) and
`docs/plugin-development/.../template` (template with placeholder module).

## 🧪 Test Instructions

1. Run `make test` and verify output contains test runs for `src/app`,
   `src/cmd`, and helper modules (e.g., `ptx-installer` tests execute)
2. Introduce a deliberately failing test in `src/helpers/ptx-installer`,
   run `make test`, verify non-zero exit code; revert
3. Run `make test-cli` and verify it tests `src/cmd` packages
