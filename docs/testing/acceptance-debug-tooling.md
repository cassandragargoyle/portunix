# Acceptance Protocol — Debug tooling & help_registry comments

**Scope**: Branch `docs/help-registry-comments` (no formal issue — maintenance changes)
**Branch**: `docs/help-registry-comments`
**Commits under test**:

- `a9ef428` docs(help): comment GenerateBasicHelp and add help_registry code doc
- `dfad061` chore(debug): add ensure-delve preLaunchTask, fix launch.json help arg
- `e49ef44` chore(docs-site): bump hugo-book theme submodule

**Tester**: claude (role: Tester / generic)
**Date**: 2026-07-03
**Testing OS**:

- Host: Linux 6.17.0-35-generic
- Shell: bash
- Toolchain: Go 1.25.0 (`go env GOVERSION`), Delve (`dlv`) built with Go 1.25.11
- ShellCheck: v0.10.0

**Build**: `make build` clean from branch tip; main `portunix` binary + 19 helper binaries rebuilt successfully.

## Scope

Low-risk maintenance changes:

| Change | Risk | Focus |
| ------ | ---- | ----- |
| Comments added to `GenerateBasicHelp` in `src/cmd/help_registry.go` | Build regression / accidental output change | Verify build + unchanged help output |
| New `src/cmd/help_registry.go.md` (code documentation) | None (doc only) | N/A |
| New `scripts/ensure-delve.sh` (VS Code preLaunchTask) | Wrong version comparison / broken guards | Version-compare logic + guards |
| `.vscode/launch.json` debug arg `help` → `--help` | Editor config only | Smoke |
| `docs-site/themes/hugo-book` submodule bump | None (pointer) | Smoke |

**Out of scope** (intentionally): the rebuild branch of `ensure-delve.sh`
(`GOTOOLCHAIN=… go install github.com/go-delve/delve/cmd/dlv@latest`) was not
executed — it touches the network and would reinstall the user's `dlv`. Its
decision logic is covered isolated in TC-08.

## Test Plan

| TC | Area | Description |
| -- | ---- | ----------- |
| TC-01 | Build | `make build` rebuilds all binaries |
| TC-02 | Basic help | `portunix --help` shows commands + installed-plugins section |
| TC-03 | Alignment | Column width spans commands **and** plugins (commented logic) |
| TC-04 | Expert help | `portunix --help-expert` renders without panic |
| TC-05 | AI help | `portunix --help-ai` is valid JSON (23 commands) |
| TC-06 | Script syntax | `bash -n scripts/ensure-delve.sh` |
| TC-07 | No-op path | dlv (1.25.11) ≥ toolchain (1.25.0) → no rebuild, exit 0 |
| TC-08 | Version compare | 4 dlv/toolchain pairs → correct OK/REBUILD verdict |
| TC-09 | Missing-`go` guard | `go` absent from PATH → error message, exit 1 |
| TC-10 | ShellCheck | `shellcheck scripts/ensure-delve.sh` clean |

## Test Results

### Functional

- [x] **TC-01** `make build` → `portunix` + 19 helpers, "All binaries built successfully"
- [x] **TC-02** `portunix --help` lists core commands and an "Installed plugins:" section (8 plugins)
- [x] **TC-03** Column width driven by longest name across both groups — plugin
  names `text-extractor` / `table-detector` (14 chars) widen the command column,
  confirming the commented alignment logic behaves as documented
- [x] **TC-04** `portunix --help-expert` renders header + categorized reference, no panic
- [x] **TC-05** `portunix --help-ai` parses as JSON via `json.load`; `commands` length = 23
- [x] **TC-06** `bash -n scripts/ensure-delve.sh` → syntax OK
- [x] **TC-07** Real run: `ensure-delve: dlv (go1.25.11) matches toolchain (go1.25.0) — OK`, exit 0 (no rebuild)
- [x] **TC-08** Isolated `sort -V` comparison, 4/4 PASS:

  | dlv | toolchain | expected | got |
  | --- | --------- | -------- | --- |
  | go1.25.11 | go1.25.0 | OK | OK |
  | go1.24.0 | go1.25.0 | REBUILD | REBUILD |
  | go1.25.0 | go1.25.0 | OK | OK |
  | go1.25.2 | go1.25.10 | REBUILD | REBUILD |

- [x] **TC-09** PATH stubbed with coreutils but no `go` → `ensure-delve: 'go' not on PATH`, exit 1

### Static analysis

- [x] **TC-10** `shellcheck v0.10.0 scripts/ensure-delve.sh` → 0 findings (clean).
  Survey of other `scripts/*.sh` showed pre-existing findings, so the new CI
  lint step is intentionally scoped to `ensure-delve.sh` only (see below).

## Coverage

- **Total: 10 | Passed: 10 | Failed: 0**
- **Not executed (by design):** `ensure-delve.sh` rebuild branch (network + reinstalls `dlv`); decision logic verified in TC-08.
- Comment-only changes in `help_registry.go` have no runtime effect — TC-02…TC-05 confirm identical behaviour.

## CI notes

- Added lint step to `.github/workflows/test.yml` (`lint` job):
  `shellcheck scripts/ensure-delve.sh` (shellcheck preinstalled on
  `ubuntu-latest`). Scoped to this script because a repo-wide sweep would fail
  on pre-existing findings in other scripts.
- **Follow-up recommendation**: incrementally clean the remaining
  `scripts/*.sh` findings, then widen the lint to `scripts/` (or adopt a
  shellcheck config with an agreed severity threshold).

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-07-03
**Tester signature**: claude (role=tester)
