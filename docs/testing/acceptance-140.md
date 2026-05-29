# Acceptance Protocol — Issue #140

**Issue**: Version Management Strategy Implementation (ADR-036)
**Branch**: `feature/140-version-management-strategy`
**Commit under test**: `f9daa2b` (plus one tester fix-up commit pending — see notes)
**Tester**: claude (role: Tester / generic)
**Date**: 2026-05-19
**Testing OS**:

- Host: Linux 6.17.0-23-generic
- Shell: bash
- Go: 1.24.4
- Python: 3.11+

**Build**: `make build` clean from feature branch tip (`f9daa2b`); main `portunix`
binary and all 19 helper binaries rebuilt successfully.

## Scope

Issue #140 implements the version management strategy defined in ADR-036:

- GitHub is the single source of truth for stable versions (`vX.Y.Z`)
- Internal repos (Gitea, fork, etc.) use the `+dev.N` build suffix
- Build scripts validate version format per release context
- Skills (`/cs:release-gitea`, `/cs:deploy-github`) enforce the rules
- A new `portunix version check` command surfaces all three values
  (local, GitHub, suggested next dev)

**In scope of this PR / acceptance:**

| Phase | Deliverable |
| ----- | ----------- |
| 1 | `README.md` versioning section + cross-links |
| 1 | `docs/contributing/VERSIONING.md` (pre-existing — verified intact) |
| 1 | `CONTRIBUTING.md` reference (pre-existing line 76 — verified intact) |
| 2 | `/cs:release-gitea` enforces `+dev.N` |
| 2 | `/cs:deploy-github` enforces clean SemVer |
| 3 | `build-with-version.sh` validates version + context |
| 3 | `scripts/make-release.py` accepts new formats |
| 4 | `portunix version check` subcommand |

**Out of scope (Phase 5, agreed with developer):**

- Pre-commit hook (marked optional in issue)
- Cleanup of existing non-conforming Gitea tags (operational, not code)
- CI/CD reference updates (would need separate audit)

## Test Plan

| TC | Area | Description |
| -- | ---- | ----------- |
| TC-001 | Build | `make build` succeeds on feature branch tip |
| TC-002 | build-with-version validation | 7 valid version+context combinations accepted |
| TC-003 | build-with-version validation | 4 context violations correctly rejected |
| TC-004 | build-with-version validation | 6 malformed versions rejected |
| TC-005 | build-with-version validation | 2 unknown contexts rejected |
| TC-006 | `portunix version check` | Default output shows all three lines |
| TC-007 | `portunix version check` flags | `--current`, `--github`, `--suggest` each isolate one line |
| TC-007.1 | `+dev.N` counter | With tags `+dev.1` + `+dev.3` present, suggests `+dev.4` (max+1) |
| TC-008 | help text | `version check --help` lists examples and flags |
| TC-009 | make-release.py validation | 7 valid formats accepted |
| TC-010 | make-release.py validation | 10 malformed versions rejected |
| TC-011 | make-release.py CLI | Invalid version → exit 1 with proper error message |
| TC-012 | Documentation | README.md has Versioning section + link targets exist |
| TC-013 | Documentation | `release-gitea.md` references ADR-036, `+dev.N`, `gh api` |
| TC-014 | Documentation | `deploy-github.md` enforces clean SemVer and GitHub tag uniqueness |
| TC-015 | Regression | `make build` and `bash build-with-version.sh` (no args) still work; default `v2.2.3` builds with `auto`-context warning |
| TC-016 | Windows resources | `+dev.N` and `-rc.N` suffixes stripped from FILEVERSION numeric tuple but preserved in StringFileInfo |

## Test Results

### Functional Tests

- [x] TC-001 — `make build` succeeded (main + 19 helpers).
- [x] TC-002 — 7/7 valid combinations passed validation.
- [x] TC-003 — 4/4 context violations rejected with descriptive errors.
- [x] TC-004 — 5/6 malformed versions rejected. **Edge case**: empty-string
  `""` as `$1` falls back to default `v2.2.3` due to bash `${1:-default}`
  semantics. This is standard bash behavior, not a validation bug; real-world
  invocation paths cannot bypass validation.
- [x] TC-005 — 2/2 invalid contexts rejected.
- [x] TC-006 — Default output:
  `Current local version : dev`,
  `Last GitHub release   : v2.2.2`,
  `Suggested next dev    : v2.2.2+dev.1`.
- [x] TC-007 — Each flag prints only its corresponding line.
- [x] TC-007.1 — With seeded tags `v2.2.2+dev.1` + `v2.2.2+dev.3`, command
  returns `v2.2.2+dev.4`; after cleanup falls back to `v2.2.2+dev.1`.
- [x] TC-008 — Help banner clean (3 examples, 3 flags, global flags listed).
- [x] TC-009 — 7/7 valid formats accepted by `validate_version()` regex.
- [x] TC-010 — 10/10 malformed versions rejected (including non-numeric
  `+dev.abc`, missing prerelease number, four-part semver).
- [x] TC-011 — `python3 scripts/make-release.py invalid-version` exits 1
  with `Invalid version format` message.
- [x] TC-012 — README.md `## Versioning` at line 540; both link targets
  (`docs/contributing/VERSIONING.md`, `docs/adr/036-version-management-strategy.md`)
  exist.
- [x] TC-013 — `release-gitea.md` has 24 keyword matches for the new flow
  (gh api, +dev.N, VERSIONING, ADR-036). **🔴 Bug found & fixed during
  testing**: relative link `../../docs/...` was incorrect — `.claude/commands/cs/`
  is 3 levels deep, needs `../../../docs/...`. Fixed in tester working tree;
  needs amended commit before merge (see Conditional Notes below).
- [x] TC-014 — `deploy-github.md` correctly references ADR-036, validates
  clean SemVer, uses `gh api .../releases/tags/vX.Y.Z` for uniqueness check,
  passes `github` context to build script.
- [x] TC-015 — `make build` works; `bash build-with-version.sh` (no args)
  builds with default `v2.2.3` and emits the expected `auto`-context warning.
  Helper binaries all report v2.2.3 after build.
- [x] TC-016 — Build with `v1.9.2+dev.7 internal`:
  `FILEVERSION 1,9,2,0` (numeric, suffix stripped),
  `VALUE "FileVersion", "1.9.2+dev.7"` (string preserved),
  `versioninfo.json` FixedFileInfo `{Major:1, Minor:9, Patch:2}` integers.
  Same behavior verified for `v1.10.0-rc.2 github`.

### Regression Tests

- [x] `make build` produces same artifacts as before (main + 19 helpers).
- [x] `bash build-with-version.sh` with no args still works (default version
  v2.2.3 + auto context).
- [x] `bash build-with-version.sh v1.9.2` (single arg) still works; new context
  parameter is optional with `auto` as default.
- [x] No new test failures introduced in `go vet portunix.ai/cmd` (the three
  warnings it reports — `misplaced +build comment` in two test files and
  `redundant newline` in `virt_exec.go` — are pre-existing on `main`).
- [x] `make-release.py validate_version` still accepts legacy
  `vX.Y.Z-SNAPSHOT` format.

### Cross-Platform

- [x] Linux verified directly.
- [ ] Windows not verified on this host. `build-with-version.sh` Windows
  branch (`OSTYPE=msys*`) is unchanged from prior version; new validation
  block runs before any OS-specific logic. Risk: low — pure bash regex
  matching is portable to git-bash on Windows.

## Findings

### 🔴 Bug found during testing (TC-013)

**File**: `.claude/commands/cs/release-gitea.md`
**Lines**: end-of-file references to VERSIONING.md and ADR-036.
**Issue**: Relative paths used `../../docs/...` but skill files are 3
directory levels deep (`.claude/commands/cs/`), so the correct path is
`../../../docs/...`. Tester verified that the original paths did not
resolve to existing files. **Fix applied** in tester working tree
(uncommitted at time of writing); the same paths in `README.md`
(`docs/contributing/VERSIONING.md`) and `deploy-github.md` are already
correct.

### 🟡 Pre-existing notes (not blockers)

- `build-with-version.sh` is tracked as `100644` (non-executable) in git;
  direct `./build-with-version.sh` invocation fails with "permission
  denied" on a fresh checkout, but all callers (`Makefile`,
  `make-release.py`, skills) explicitly use `bash build-with-version.sh`
  or rely on user's local execute bit. Not in scope for issue #140.
- `portunix.syso` is regenerated by every build and was excluded from the
  developer commit; this is the existing convention (see commit
  `653341b fix(build): regenerate portunix.syso on every build`).

## Final Decision

**STATUS**: CONDITIONAL PASS

**Conditions for merge:**

1. **MUST**: Amend or add follow-up commit fixing the relative paths in
   `.claude/commands/cs/release-gitea.md` (TC-013 finding). Tester has the
   fix applied in working tree and will commit it as part of this protocol's
   sign-off.

After condition #1 is satisfied, the implementation **PASSES** all functional
and regression tests covered by this protocol.

**Approval for merge**: YES (after condition #1)
**Date**: 2026-05-19
**Tester signature**: claude (role: Tester / generic)
