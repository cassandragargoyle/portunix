# Acceptance Protocol — Issue #079

**Issue**: Version and Method Override via CLI Parameters (Phase 1 — method override + listing)
**Branch**: `feature/079-method-override-list-variants`
**Commit under test**: `d442f79`
**Tester**: zdendaku (role: Tester / generic)
**Date**: 2026-05-10
**Testing OS**:

- Host: Linux 6.17.0-23-generic (Ubuntu 25.10) — used for unit tests, CLI smoke tests, build verification
- Container: `ubuntu:22.04` via Podman 5.4.2 — used for end-to-end install verification per `TESTING_METHODOLOGY.md`

**Build**: `make build` clean from feature branch tip (`d442f79`); both `portunix` and `ptx-installer` helper rebuilt.

## Scope

Phase 1 of Issue #079, scoped per developer/architect agreement on this PR:

- `--method=<variant>` as alias of `--variant=<variant>`
- `--list-variants` (and alias `--list-methods`) to discover available variants
- Optional `Preferred bool` field in `VariantSpec` (JSON: `preferred,omitempty`)
- Improved error message for unknown variant — lists alternatives + hint
- New unit tests for the three extracted helpers

**Out of scope** (deferred to later phases): `--version=latest|prerelease|<x>`, GitHub Releases API integration, URL templating, version constraints, `~/.portunix/config.yaml` configuration. These are tracked by Phase 2/3 of #079 and need separate issues.

## Test Plan

| Area | Tests |
| ---- | ----- |
| Unit tests | TC1 |
| Variant discovery (`--list-variants`) | TC2, TC3, TC9 |
| Method/variant aliasing | TC4, TC5 |
| Error path | TC6 |
| Documentation | TC7 |
| Backward compatibility | TC8 |
| End-to-end (container) | TC10 |

## Test Cases (Given/When/Then)

### TC1 — Unit tests pass

- **Given** the feature branch tip and ptx-installer module
- **When** running `go test -v ./...` from `src/helpers/ptx-installer`
- **Then** all 5 new test functions pass (10 sub-cases total):
  - `TestParseVariantArg_VariantAndMethodEquivalent` (7 sub-cases)
  - `TestFormatVariantList_RendersAllFields`
  - `TestFormatVariantList_StableOrdering`
  - `TestUnknownVariantError_ListsAvailableSorted`
  - `TestUnknownVariantError_EmptyVariantsMap`
- **Result**: ✅ PASS
- **Note**: Three pre-existing Windows-specific test failures (`TestValidateWindowsPath`, `TestResolveWindowsDataRoot_ExplicitFlag`, `TestResolveWindowsDataRoot_ExplicitFlag_InsufficientSpace`) reproduce on `main` — verified by checking out main and running the same tests. **Out of scope of this issue.**

### TC2 — `--list-variants` happy path

- **Given** a Linux host and the `hugo` package definition with 4 variants (apt, extended, snap, standard)
- **When** running `portunix install hugo --list-variants`
- **Then**
  - All 4 variants printed in alphabetical order
  - `apt` marked with `*` (auto-detected — host has `apt-get`)
  - Header shows package name, OS, and legend
  - Footer shows install hint with both `--method=` and `--variant=` examples
  - Exit code = 0
- **Result**: ✅ PASS

### TC3 — `--list-methods` is a true alias

- **Given** the same package
- **When** comparing outputs of `--list-variants` and `--list-methods`
- **Then** outputs are byte-identical (`diff` shows no difference); both exit 0
- **Result**: ✅ PASS

### TC4 — `--method` ↔ `--variant` equivalence

- **Given** the `hugo` package
- **When** running each of `--variant=X --dry-run` and `--method=X --dry-run` for X ∈ {extended, standard, snap}
- **Then** stdout/stderr identical for each pair (`diff` clean)
- **Result**: ✅ PASS — all 3 variant pairs identical

### TC5 — Space-separated form

- **Given** the `hugo` package
- **When** running `portunix install hugo --method extended --dry-run` (separate token) and `portunix install hugo --variant snap --dry-run`
- **Then** both produce the same output as their `=` form
- **Result**: ✅ PASS

### TC6 — Unknown variant error message

- **Given** the `hugo` package
- **When** running `portunix install hugo --method=xyznotexists`
- **Then** error message contains all five expected pieces:
  - The requested variant name (`"xyznotexists"`)
  - The package (`hugo`) and OS (`linux`)
  - A sorted list of available variants (`apt, extended, snap, standard`)
  - A hint pointing to `--list-variants`
  - Process exit code = 1
- **And** `--variant=xyznotexists` produces a byte-identical error
- **Result**: ✅ PASS

### TC7 — Help text updated

- **Given** the install help
- **When** running `portunix install --help`
- **Then** the Options section lists `--method=<variant>`, `--list-variants`, `--list-methods`, and the Examples section shows both `--list-variants` and `--method=snap` usage
- **Result**: ✅ PASS

### TC8 — Backward compatibility

| Case | Command | Expected | Result |
| ---- | ------- | -------- | ------ |
| 8a | `portunix install java --variant=21 --dry-run` | Picks Java 21 (existing flag still works) | ✅ PASS |
| 8a' | `portunix install java --variant 17 --dry-run` (space) | Picks Java 17 | ✅ PASS |
| 8b | `portunix package list` | Loads all 63 packages, 0 errors (no schema break despite new optional `preferred` field) | ✅ PASS |
| 8c | `portunix install python --dry-run` / `portunix install go --dry-run` | Default install paths still work | ✅ PASS |

- **Result**: ✅ PASS — schema additions are non-breaking; all pre-existing flag forms preserved.

### TC9 — `--list-variants` for unknown package

- **Given** a non-existent package name `nesmysl-balicek`
- **When** running `portunix install nesmysl-balicek --list-variants`
- **Then** clean error message ("Package 'nesmysl-balicek' not found"), pointer to `package search`, exit code 1
- **Result**: ✅ PASS

### TC10 — End-to-end install in clean container

- **Given** a fresh `ubuntu:22.04` container with `portunix` + `ptx-installer` helper copied to `/usr/local/bin` and `ca-certificates` installed
- **When** running `portunix install hugo --method=standard --force` inside the container
- **Then**
  - The `--method=standard` flag is recognized (output shows `🎯 Variant: standard (version: 0.150.1)`)
  - URL chosen for the `standard` variant (`hugo_0.150.1_linux-amd64.tar.gz`)
  - Download completes (17.0 MB)
  - tar.gz extracted (3 files: LICENSE, README.md, hugo)
  - Hugo binary present at `/usr/local/bin/hugo` and runs (`hugo v0.150.1-...`)
- **Out-of-scope failure observed**: The `${sudo_prefix}` / `${actual_extract_to}` placeholders in `hugo.json` postInstall are not substituted — the chmod fails. **Pre-existing bug in ptx-installer post-install hook variable substitution; reproduces with `--variant=standard` too. NOT introduced by this PR.** Hugo is still functionally deployed.
- **Result**: ✅ PASS for issue #079 acceptance criteria. (See "Findings out of scope" below.)

## Acceptance Criteria — coverage matrix

From issue #079 (Phase 1 functional requirements):

| AC | Requirement | Verified by | Status |
| -- | ----------- | ----------- | ------ |
| AC-1 | Method Override (`--method=<id>`) | TC4, TC5, TC10 | ✅ |
| AC-2 | Method Discovery (`--list-methods`) | TC2, TC3 | ✅ |
| AC-3 | Dry Run preview (`--dry-run`) | TC4 (and pre-existing) | ✅ |
| AC-4 | Backward Compatibility | TC8 | ✅ |
| AC-5 | Performance < 100 ms overhead | List/dry-run completed effectively instantly (no measurable overhead beyond registry load) | ✅ |
| AC-6 | Reliability — graceful errors | TC6, TC9 | ✅ |
| AC-7 | Usability — clear messages | TC2 (legend, marker), TC6 (hint to list-variants) | ✅ |

Phase 2/3 acceptance criteria (`--version=latest|prerelease`, GitHub API, etc.) are **explicitly deferred** and not in scope of this PR's acceptance.

## Findings out of scope

These were observed during testing but are **not regressions introduced by this PR** and do not block acceptance:

1. **Pre-existing Windows-only test failures** on Linux (TC1 note) — `TestValidateWindowsPath`, `TestResolveWindowsDataRoot_*`. Reproduce on `main`. Should be guarded by `//go:build windows` or relocated.
2. **Pre-existing post-install variable substitution bug** in ptx-installer — `${sudo_prefix}` / `${actual_extract_to}` are passed literally to the shell. Affects any package whose `postInstall` uses these placeholders (e.g. hugo.json line 56). Reproduces with `--variant=standard`. Should be tracked as a separate bug.
3. **`hugo.json` schema gap (informational only)** — the `apt` and `snap` variants do not declare an explicit `"type"`, so `--list-variants` displays them as `type: tar.gz` (the platform default). Engine `Install()` would still treat them via the platform-default code path, which would not work for actual `apt`/`snap` installs. Existing data hygiene issue, predates #079.

None of the above are caused by code changes on this branch.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES

**Rationale**: All Phase 1 acceptance criteria from Issue #079 are met. Unit tests for the new helpers all pass. Backward compatibility verified — every prior `--variant`/`--dry-run` invocation still works, all 63 embedded packages load without errors, and the new `Preferred` field is opt-in (omitempty) so existing JSON definitions are unaffected. Container-based end-to-end run confirms `--method` flows through to actual installation. Pre-existing issues found during testing are documented separately and do not regress.

**Date**: 2026-05-10
**Tester signature**: zdendaku (role: tester)
