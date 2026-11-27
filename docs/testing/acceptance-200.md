# Acceptance Protocol - Issue #200 (GitHub latest release installs without ptx-installer)

**Issue**: GitHub `latest` Release v2.2.2 Installs Without `ptx-installer` on Windows
**Branch**: `fix/200-update-all-helpers` (commit `8dade71`; retest 1 on `0771345`; retest 2 on `0fcc2fc`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-10-27
**Testing OS**: Ubuntu 24.04 container (Podman 5.7.0) for installation tests;
Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) for unit tests, lint and release-script checks;
Go 1.26.8

## Scope Note

The branch changes three things:

- `src/app/update/github.go` — `extractAllFromZip` / `extractAllFromTarGz` extract
  `portunix` and every `ptx-*` entry of the release archive instead of the hardcoded
  `portunix`, `ptx-container`, `ptx-mcp`
- `scripts/make-release.py` — helpers are discovered from `src/helpers/`;
  `verify_outputs` fails when a release archive lacks `portunix` or a helper
- unit tests in `src/app/update/extract_test.go` and `src/app/selfinstall/install_test.go`

**Independence note**: the code and the unit tests were written by Claude Code in the developer
role in the same session. The scenarios below were designed separately in the tester role.

**Not covered**: Windows. AC2–AC4 are defined on a clean Windows and were not executed;
the changed packages vet for `GOOS=windows`. A `win11` KVM VM exists on the host and is the
intended environment for the Windows retest.

## Test Summary

Initial run (`8dade71`):

- Total test scenarios: 9
- Passed: 8
- Failed: 1 (TC05, pre-existing defect O1, not a regression)
- Skipped: 0
- Not executed: Windows AC2–AC4
- Defects found: 1 pre-existing (O1)
- Observations: 2 (O2, O3)

Retest 1 (`0771345`) — see [Retest 1](#retest-1):

- Total test scenarios: 13 (TC01–TC09, R1–R4)
- Passed: 12
- Failed: 1 (R4, defect D1)
- O1 resolved on Linux; O2 still to be verified on Windows

Retest 2 (`0fcc2fc`) — see [Retest 2](#retest-2):

- Total test scenarios: 12 (R1–R6, TC03, TC04, TC07–TC09 regression)
- Passed: 12
- Failed: 0
- D1 and D2 resolved; Windows and AC4 wording still open

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make build`, `gofmt`, `go vet` (also `GOOS=windows`), `golangci-lint` | host |
| Unit | `update` and `selfinstall` packages | host |
| Real data | updater extraction and apply against GitHub `v2.6.3` archives | host, temp dir |
| Installation | `install-self`, `install --dry-run`, `update --force` | Ubuntu 24.04 container |
| Gate | `make-release.py` archive check, incl. failure injection | host |

## Test Cases

### TC01: GitHub latest release contains the fix (AC1)

- **Given** the GitHub releases of `cassandragargoyle/portunix`
- **When** `releases/latest` is resolved and checked against `40d1375`
- **Then** latest is `v2.6.3` and `git merge-base --is-ancestor 40d1375 v2.6.3` succeeds
- **Result**: PASS

### TC02: Release archives ship every helper

- **Given** `portunix_2.6.3_windows_amd64.zip` and `portunix_2.6.3_linux_amd64.tar.gz`
- **When** their contents are listed
- **Then** both contain `portunix` and the 19 helpers of `src/helpers/`, flat, no subdirectory
- **Result**: PASS

### TC03: install-self installs every helper (AC2, Linux analogue)

- **Given** a clean `ubuntu:24.04` container with the `make build` binaries mounted read-only
- **When** `portunix install-self --silent --path /opt/ptx/portunix` runs
- **Then** `/opt/ptx` contains `portunix` and all 19 `ptx-*` helpers
- **Result**: PASS — 19 helpers reported and present

### TC04: Package install works right after install-self (AC3, Linux analogue)

- **Given** the installation from TC03
- **When** `/opt/ptx/portunix install python --dry-run` runs
- **Then** `ptx-installer` is found and the dry run resolves the `apt` variant, exit 0
- **Result**: PASS

### TC05: Fixed updater fills missing helpers end to end (AC4, Linux)

- **Given** a container directory with only the fixed `portunix` binary and network access
- **When** `portunix update --force` runs against GitHub `v2.6.3`
- **Then** all helpers of the archive appear next to `portunix`
- **Result**: FAIL — the update aborts before extraction with
  `permission denied … Cannot write to /opt/upd/portunix` and the restore with
  `text file busy`, although the container runs as root. The same failure reproduces with the
  unmodified `v2.6.3` binary, so it is pre-existing — see O1

### TC06: Fixed updater logic on real archives (AC4, logic level)

- **Given** the GitHub `v2.6.3` Windows zip and Linux tar.gz
- **When** a temporary in-package test (not committed) runs `extractAllFromZip`,
  `extractAllFromTarGz` and `applyUpdateAllBinaries` into a directory holding only `portunix`
- **Then** each archive yields 20 binaries and the directory ends with `portunix` + 19 helpers
- **Result**: PASS

### TC07: Archive entry filter

- **Given** `TestReleaseBinaryName`, `TestExtractAllFromZip`, `TestExtractAllFromTarGz`,
  `TestExtractAllRequiresMainBinary`, `TestInstallAllBinariesCopiesEveryHelper`
- **When** `go test ./update/ ./selfinstall/ -count=1 -v` runs in `src/app`
- **Then** helpers are extracted, `LICENSE` / `README.md` / `install.*` and foreign-platform
  binaries are skipped, a missing main binary is an error; all existing tests still pass
- **Result**: PASS — 22 tests passed

### TC08: Release gate detects a missing helper (AC5)

- **Given** `discover_helpers` and `verify_archive_helpers` from `scripts/make-release.py`
- **When** they run against the `v2.6.3` archives, then again with an injected `ptx-missing`
  helper
- **Then** the first run passes with 19 helpers; the second fails and names
  `ptx-missing` / `ptx-missing.exe` for each archive
- **Result**: PASS

### TC09: Static checks and regression

- **Given** the branch
- **When** `make build`, `gofmt -l`, `go vet` (native and `GOOS=windows`), `py_compile` and
  `golangci-lint run ./update/ ./selfinstall/` run
- **Then** build, format, vet and compile are clean; lint reports no finding in the added code
- **Result**: PASS — the 18 lint findings are all in unchanged code (unused `extractBinary`,
  `extractFromZip`, `extractFromTarGz`, `applyUpdateUnix`, `applyUpdateWindows`; errcheck;
  misspell)

## Findings

### O1: `portunix update` cannot replace the running binary on Linux (pre-existing)

`ApplyUpdate` calls `checkWritePermission(execPath)`, which opens the running executable with
`O_WRONLY`; Linux refuses that with `ETXTBSY`, the error is reported as `permission denied`.
`applyUpdateAllBinaries` would then `copyFile` over the running binary with the same result.
Reproduced with `v2.6.3` and with the branch. Self-update on Linux is broken independently of
this issue. Not a regression of the branch, but it blocks the AC4 path on Linux and the
documented `portunix update` + `portunix update --force` workaround. Recommendation: separate
bug issue — replace binaries via write-to-temp + `rename` (unused `applyUpdateUnix` already
exists).

### O2: Windows self-update likely has the same problem (to verify)

On Windows `applyUpdateAllBinaries` also overwrites the running `portunix.exe` with
`copyFile`; Windows does not allow writing to a running executable, and the rename-based
`applyUpdateWindows` is unused. AC4 on Windows is therefore expected to fail the same way.
Must be checked in the Windows retest before the workaround in the issue is relied on.

### O3: Unused legacy extraction functions

`extractBinary`, `extractFromZip` and `extractFromTarGz` in `src/app/update/github.go` are
unused and still use the loose `strings.Contains(name, "portunix")` matching. Cleanup
candidate, not blocking.

## Acceptance Criteria Status

| AC | Status |
| -- | ------ |
| AC1 latest release contains `40d1375` | PASS (TC01) |
| AC2 install-self installs all helpers on clean Windows | Linux analogue PASS (TC03); Windows not executed |
| AC3 `install python --variant=full` after install | Linux analogue PASS (TC04); Windows not executed |
| AC4 `update` from v2.2.2 leaves all helpers | Logic PASS (TC06); end to end blocked by O1 / O2 |
| AC5 release workflow checks helpers | PASS (TC08) |

## Final Decision

**STATUS**: CONDITIONAL

The branch fixes what it changes: the updater and the release gate now handle every helper,
verified on real release archives, with no regression. Merge conditions:

1. Windows retest of AC2 and AC3 on the `win11` VM
2. Decision on AC4: either fix O1 / O2 (self-update replacing the running binary) or move AC4
   with the double-update workaround to a separate issue

**Approval for merge**: NO (pending conditions above)
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)

## Retest 1

**Commit**: `0771345` — `fix(200): self-update replaces running binaries and keeps the archive`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 24.04 container (Podman 5.7.0); host as above

The developer fix replaces binaries by rename (`replaceBinary`), checks write permission on
the install directory and keeps the downloaded archive in `DownloadUpdate`, which was deleted
by a `defer` before extraction (found while fixing O1).

### R1: End-to-end update fills missing helpers (TC05 rerun, AC4 logic)

- **Given** a container directory with only the fixed `portunix` binary
- **When** `portunix update --force` runs against GitHub `v2.6.3`
- **Then** the update completes, `portunix` reports `2.6.3`, all 19 helpers are present,
  no `.old` / `.new` / `.backup` file and no `/tmp/portunix-update-*` is left
- **Result**: PASS

### R2: Update while a helper is running

- **Given** `/opt/r2/ptx-mcp` is a copy of `sleep` running `300`; writing to it fails with
  `Text file busy` (precondition confirmed)
- **When** `portunix update --force` runs in that directory
- **Then** the update completes, `ptx-mcp --version` reports `2.6.3`, the running process
  keeps running, no leftovers
- **Result**: PASS

### R3: Install directory not writable (failure injection)

- **Given** `/opt/r3` owned by root, mode 755, update run as `nobody`
- **When** `portunix update --force` runs
- **Then** it fails cleanly, exit 1, files are byte-identical to before, no leftovers
- **Result**: PASS — `Error: Failed to create backup … portunix.backup: permission denied`

### R4: Rollback when one helper cannot be replaced (failure injection)

- **Given** a full `dev` install in `/opt/r4` and a non-empty directory `ptx-wizard.new`
  blocking the replacement of `ptx-wizard` (it comes after `portunix` in the archive)
- **When** `portunix update --force` runs
- **Then** the update fails, every binary is back to the `dev` version, the user gets a
  correct status and no stale files remain
- **Result**: FAIL — the binary rollback itself works (all 20 binaries report `dev`), but:

```text
Error: Failed to apply update
  failed to apply update: failed to update ptx-wizard: open /opt/r4/ptx-wizard.new: is a directory (restoring backups failed: ptx-wizard from /opt/r4/ptx-wizard.backup: open /opt/r4/ptx-wizard.new: is a directory)
Attempting to restore backup...
Error: Failed to restore backup
  failed to resolve symlinks: lstat /opt/r4/portunix.old: no such file or directory
```

and 20 `*.backup` files stay in the install directory — see D1, D2.

### Regression

- TC03 / TC04 rerun: `install-self` installs 19 helpers, `install python --dry-run` resolves — PASS
- TC07: `go test ./update/ ./selfinstall/` incl. new `replace_test.go` — PASS
- TC08: release gate passes on `v2.6.3`, fails with injected `ptx-missing` — PASS
- TC09: `make build`, `GOOS=windows go vet` clean; lint 16 findings, all in unchanged code
  (18 before, the removed `applyUpdateUnix` / `applyUpdateWindows` are gone) — PASS

### D1: Second restore resolves the wrong path after the running binary was renamed

`performUpdate` in `src/cmd/update.go` calls `update.RestoreBackup(backupPath)` after a failed
`ApplyUpdate`. `RestoreBackup` resolves the target with `os.Executable()`; once
`replaceBinary` has renamed the running `portunix` to `portunix.old`, that returns
`portunix.old` (on Linux the file is already removed, so `EvalSymlinks` fails). The user is
told `Failed to restore backup … Manual intervention may be required` although the install
is intact. On Windows `portunix.exe.old` still exists while running, so the restore would be
written next to it instead of to `portunix.exe`.

`ApplyUpdate` already rolls back all binaries including `portunix`, so the second restore is
redundant. Suggested fix: resolve the install path once before the update and pass it on, or
drop the second restore. **Blocking.**

### D2: Backups stay behind after a partly failed rollback

When `restoreBackups` fails for one binary, `applyUpdateAllBinaries` returns without removing
the backups of the binaries that were restored successfully, and `performUpdate` keeps
`portunix.backup`. R4 left 20 `*.backup` files. Keeping the backup of the failed binary is
intended; the others are noise. Minor, not blocking.

### Acceptance Criteria Status (Retest 1)

| AC | Status |
| -- | ------ |
| AC1 | PASS |
| AC2 | Linux analogue PASS; Windows not executed |
| AC3 | Linux analogue PASS; Windows not executed |
| AC4 | `update` from a fixed version PASS on Linux (R1, R2); from `v2.2.2` not possible by design (its updater cannot replace itself) — AC needs rewording by architect / PO |
| AC5 | PASS |

### Final Decision (Retest 1)

**STATUS**: FAIL — return to developer

1. Fix D1 (blocking); D2 recommended in the same change
2. Then: Windows retest on the `win11` VM (AC2, AC3, update incl. `portunix.exe.old`)
3. AC4 rewording by architect / PO

**Approval for merge**: NO
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)

## Retest 2

**Commit**: `0fcc2fc` — `fix(200): single rollback in update, clean up backups and temp files`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 24.04 container (Podman 5.7.0); host as above

The fix drops the second restore in `performUpdate`, rolls back only binaries already
replaced, removes restored backups and extracted temp files, marks incomplete rollbacks with
`ErrRestoreFailed`, and removes the downloaded archive before `os.Exit`.

### R4: Rollback when one helper cannot be replaced (D1, D2)

- **Given** a full `dev` install in `/opt/r4` and a non-empty directory `ptx-wizard.new`
- **When** `portunix update --force` runs
- **Then** exit 1, message `failed to update ptx-wizard … ✓ Previous version kept`, no
  `Failed to restore backup`; all 20 binaries report `dev`; no `*.backup`, `*.old` or
  `/tmp/portunix-update-*` left (only the injected `ptx-wizard.new` directory)
- **Result**: PASS

### R5: Rollback that cannot restore a binary

- **Given** `restoreBackups` with one restorable and one missing backup
  (`TestRestoreBackupsRemovesRestoredBackups`)
- **Then** it returns `ErrRestoreFailed`, restores the first binary and removes its backup,
  keeps the failing one; `performUpdate` maps `ErrRestoreFailed` to the
  `*.backup files were kept` message (code review)
- **Result**: PASS at unit level — a system-level injection is not possible as root without
  a fault hook, because restore uses the same `.new` path that the successful apply used

### R6: Downloaded archive removed on failure

- **Given** a temporary in-package test (not committed) with a local HTTP server and
  `TMPDIR` set to an empty directory
- **When** `DownloadUpdate` runs with a wrong checksum, with a 404, and with a valid checksum
- **Then** the first two return an error and leave `TMPDIR` empty; the valid one returns the
  archive, readable, for the caller to extract
- **Result**: PASS

### Regression

- R1: update from a directory with only `portunix` installs 19 helpers, `2.6.3`, no
  leftovers — PASS
- R2: update while `ptx-mcp` runs; new `ptx-mcp` is `2.6.3`, old process keeps running — PASS
- R3: non-writable directory as `nobody` — clean `permission denied`, file unchanged,
  no leftovers — PASS
- TC03 / TC04: `install-self` installs 19 helpers, `install python --dry-run` resolves — PASS
- TC07: 27 unit tests in `update` and `selfinstall` — PASS
- TC09: `make build`, `GOOS=windows go vet` clean; lint 16 findings, none in changed files — PASS

### Observation

- O4: on Linux the permission hint still prints the Windows-only lines
  (`Run as Administrator (Right-click cmd.exe …)`) after `Try running with sudo`.
  Pre-existing, cosmetic, not blocking.

### Acceptance Criteria Status (Retest 2)

| AC | Status |
| -- | ------ |
| AC1 | PASS |
| AC2 | Linux analogue PASS; Windows not executed |
| AC3 | Linux analogue PASS; Windows not executed |
| AC4 | Update from a fixed version PASS on Linux incl. running binaries and rollback; from `v2.2.2` not possible by design — rewording by architect / PO pending |
| AC5 | PASS |

### Final Decision (Retest 2)

**STATUS**: CONDITIONAL — Linux PASS, all defects resolved

Remaining merge conditions:

1. Windows retest on the `win11` VM: AC2, AC3, `portunix update --force` (incl. the running
   `portunix.exe` left as `portunix.exe.old` and cleaned by the next update)
2. AC4 rewording by architect / PO

**Approval for merge**: NO (pending conditions above)
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)

## Merge Note

**Date**: 2026-10-27

`fix/200-update-all-helpers` was merged into `main` at `b36cccd` on the decision of the
repository owner (zdendaku), as an exception to the "no merge without PASS" rule. The
Windows retest (AC2, AC3, AC4a on the `win11` VM) is still outstanding. Issue #200 stays
open until it passes; any defect found there is fixed on a new branch.
