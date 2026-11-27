# Acceptance Protocol - Issue #243 (build-with-version.sh helpers)

**Issue**: build-with-version.sh does not build ptx-credential and ptx-python
**Branch**: `fix/243-build-with-version-missing-helpers` (commit `8dbf26e`; retest 1 on `2dfe632`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04.1 LTS, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The change is limited to the build script `build-with-version.sh`. No software was installed
and no host configuration was changed, so the tests ran on the host. Version files rewritten
by test builds (`portunix.rc`, `versioninfo.json`, `*.syso`) were restored with `git checkout`
after every run; the working tree was clean at the end.

## Scope Note

**Independence note**: the change was made by Claude Code in the developer role in the same
session. The scenarios below were designed separately in the tester role.

**Not executed**:

- `make deploy-local` into the real installation directory — it would overwrite the installed
  binaries with a test version. TC04 runs the binary selection of `scripts/deploy-local.py`
  into a scratch directory instead
- Run on Windows (Git Bash) — AC4 is verified by review of the `${EXT}` handling (TC06)
- `scripts/make-release.py vX.Y.Z-SNAPSHOT` — GoReleaser builds the helpers itself from
  `.goreleaser.yml`; `make-release.py` only calls `build-with-version.sh` to update the version
  files, which TC01 covers

## Test Summary

- Total test scenarios: 7
- Passed: 7
- Failed: 0
- Observations: O1, O2 (pre-existing, not caused by this change, not blocking)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Build | `build-with-version.sh` with a test version, `make build` | host |
| Fault injection | extra helper, helper that does not compile | host (temporary dir) |
| Deploy | `deploy-local.py` binary selection, copied into a scratch directory | host |
| Static | `${EXT}` handling for Windows | review |

## Test Cases

### TC01: versioned build of all helpers (AC1, AC2)

- **Given** the branch at `8dbf26e`
- **When** `bash ./build-with-version.sh v0.0.0+dev.1` runs
- **Then** it exits 0 and the version check lists the main binary and every helper in
  `src/helpers/` with `v0.0.0+dev.1`

Result: **PASS** — exit 0; the main binary and 19 helpers report `v0.0.0+dev.1`, including
`ptx-credential` and `ptx-python`. `ptx-installer_windows_{amd64,arm64}.syso` was regenerated
as before.

### TC02: helpers report the version from the repository root (AC3)

- **When** `./ptx-credential --version` and `./ptx-python --version` run after TC01
- **Then** both report `v0.0.0+dev.1`

Result: **PASS** — `ptx-credential version v0.0.0+dev.1`, `ptx-python version v0.0.0+dev.1`.

### TC03: new helper is picked up automatically (AC5)

- **Given** a temporary helper `src/helpers/ptx-zzt243` (own `go.mod`, `var version = "dev"`)
- **When** `bash ./build-with-version.sh v0.0.0+dev.2` runs
- **Then** the helper is built and its version check reports `v0.0.0+dev.2`

Result: **PASS** — `Building ptx-zzt243...`, `ptx-zzt243 version v0.0.0+dev.2`; 21 version
lines (main + 20 helpers). The temporary helper was removed.

### TC04: deploy set carries the version (AC3)

- **Given** a build with `v0.0.0+dev.4`
- **When** the binaries selected by `get_binaries()` from `scripts/deploy-local.py` are copied
  into a scratch directory and run with `--version`
- **Then** every Linux helper reports `v0.0.0+dev.4`

Result: **PASS** — 20 Linux binaries copied (main + 19 helpers); all 19 helpers report
`v0.0.0+dev.4`. See O2 for the extra `.exe` files.

### TC05: helper build failure (fault injection)

- **Given** a temporary helper `src/helpers/ptx-zzt243` that does not compile
- **When** `bash ./build-with-version.sh v0.0.0+dev.3` runs
- **Then** the script exits non-zero and names the failed helper

Result: **PASS** — exit 1, `Helper binary build failed: ptx-zzt243`; all 20 helpers were still
attempted. The temporary helper was removed.

### TC06: Windows `.exe` suffix (AC4)

- **When** the script is reviewed for binary names
- **Then** both the build output and the version check use `${helper}${EXT}`, the same `EXT`
  as the main binary

Result: **PASS** (review) — `go build ... -o "../../../${helper}${EXT}"` and
`"./${helper}${EXT}" --version`; `EXT` detection is unchanged.

### TC07: regression — `make build`

Result: **PASS** — main binary and all 19 helpers built; their version stays `dev`, as
expected for `make build`.

## Observations

### O1: `make-release.py` default-version rewrite no longer matches (pre-existing)

`update_version_files()` in `scripts/make-release.py` rewrites `^VERSION=\$\{1:-vX.Y.Z\}` in
`build-with-version.sh`. That line no longer exists — the default version is read from
`portunix.rc` — so the substitution is a no-op while still printing
`✓ Updated build-with-version.sh default version`. It predates this change and has no
functional effect.

### O2: `deploy-local.py` deploys stale `.exe` files on Linux (pre-existing)

`get_binaries()` globs `ptx-*{ext}`; on Linux `ext` is empty, so the glob also matches the
ten `ptx-*.exe` files left in the repository root (git-ignored, dated 2026-01-03). They would
be copied into the installation directory by `make deploy-local`. Not caused by this change.

### Resolution of O1 and O2

Both observations were fixed on the same branch after this protocol:

- O1 — the dead `re.sub` rewrite was removed from `update_version_files()`; the call of
  `build-with-version.sh` is unchanged. Verified: `update_version_files()` no longer contains
  `re.sub`, both scripts compile
- O2 — `get_binaries()` skips `*.exe` helpers unless running on Windows. Verified on the
  repository root: 20 binaries selected on Linux, none of them `.exe`; with `IS_WINDOWS` set,
  the `.exe` helpers are still selected

## Retest 1

**Commit**: `2dfe632` (fixes for O1 and O2)
**Date**: 2026-10-27
**Testing OS**: same host as above

`build-with-version.sh` is unchanged since `8dbf26e`; the retest covers `scripts/deploy-local.py`
and `scripts/make-release.py`.

- Total test scenarios: 3 (S1–S3)
- Passed: 3
- Failed: 0

### S1: deploy into an installation directory (O2, AC3)

- **Given** a build with `v0.0.0+dev.5` and ten stale `ptx-*.exe` files in the repository root
- **When** `deploy()` from `scripts/deploy-local.py` copies into a scratch installation directory
- **Then** only the Linux binaries are deployed and every helper reports the version

Result: **PASS** — `deploy()` returned `True`; 20 files deployed (main + 19 helpers), 0 `.exe`;
all 19 helpers report `v0.0.0+dev.5`.

### S2: Windows selection is kept (O2, regression)

- **When** `get_binaries()` runs with `IS_WINDOWS = True` (simulated)
- **Then** the `.exe` binaries are still selected

Result: **PASS** — 11 binaries selected, all `.exe` (the content of the repository root).

### S3: `update_version_files()` (O1, regression)

- **When** `update_version_files("v0.0.0-SNAPSHOT")` from `scripts/make-release.py` runs
- **Then** `build-with-version.sh` is not modified, the misleading `✓ Updated` message is gone,
  and `portunix.rc` receives the version

Result: **PASS** — SHA-1 of `build-with-version.sh` identical before and after; output only
`Updating portunix.rc with version v0.0.0-SNAPSHOT...`; `FileVersion` set to
`0.0.0-SNAPSHOT`. Version files were restored and `make build` was rerun afterwards.

## Final Decision

**STATUS**: PASS

All acceptance criteria are met (AC1–AC4, optional AC5). The pre-existing observations O1 and
O2 were fixed in `2dfe632` and pass Retest 1.

**Approval for merge**: YES
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)
