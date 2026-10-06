# Acceptance Protocol - Issue #186d

**Issue**: Migrate `install apt source / iso / chocolatey` admin subcommands
into `ptx-installer`
**Branch**: `refactor/issue-186d-install-admin-helpers` (commit `c5484f0`)
**Tester**: Tester (generic role)
**Date**: 2026-05-19
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — source/build-level
verification. Cross-platform behaviour (apt on Linux, choco on Windows)
verified by source review since the manager binaries themselves
(`apt-get`, `choco`) are not present in the build environment.

## Test Summary

- Total test scenarios: 7
- Passed: 6
- Conditional: 1 (TC07 — container/Linux runtime test deferred to a
  later platform-specific tester; see "CI Notes" below)
- Failed: 0

## Test Results

### TC01 — `cmd/install_apt.go`, `install_iso.go`, `install_chocolatey.go` deleted

- Acceptance criterion: these three files must be deleted; dispatch
  handles everything via helper.
- Result: **PASS** — `ls src/cmd/install_{apt,iso,chocolatey}.go`
  returns "No such file or directory" for all three.

### TC02 — Main binary `cmd/` no longer imports `app/install/apt`, `app/install/chocolatey`, `install.ISOInstaller`

- `grep portunix.ai/app/install/apt src/cmd/` → no matches
- `grep portunix.ai/app/install/chocolatey src/cmd/` → no matches
- `grep install.ISOInstaller src/cmd/` → no matches
- Result: **PASS**

### TC03 — Full `make build` (main + helpers)

- Command: `make build`
- Result: **PASS** — main binary and 19 helpers built successfully.

### TC04 — Shared packages exist with correct package names

- `src/pkg/installapt/installapt.go` exists, `package installapt`.
- `src/pkg/installchocolatey/installchocolatey.go` exists,
  `package installchocolatey`.
- `src/pkg/installiso/installiso.go` exists, `package installiso`.
- Each is a pure source move of the previous implementation (verified
  by checking method signatures match `app/install/apt/apt.go`,
  `chocolatey/chocolatey.go`, `install_iso.go` originals).
- Result: **PASS**

### TC05 — `app/install/{apt,chocolatey,install_iso.go}` reduced to aliases

- `src/app/install/apt/apt.go` reduced to type-alias wrapper
  (`AptManager`, `Repository`, `PackageInfo`) + `var NewAptManager`.
- `src/app/install/chocolatey/chocolatey.go` reduced to type-alias
  wrapper (`ChocolateyManager`, `PackageInfo`) + `var NewChocolateyManager`.
- `src/app/install/install_iso.go` reduced to type-alias wrapper
  (`ISOInstaller`, `ISOConfig`, `ISOPackage`, `ISOPlatform`,
  `ISOVariant`, `HashInfo`, `DownloadSettings`).
- Existing call site in `src/app/install/installer.go:672`
  (`apt.NewAptManager()`) still compiles — verified by TC03 (full build).
- Result: **PASS**

### TC06 — `ptx-installer` handles `install apt|iso|chocolatey`

- New file `src/helpers/ptx-installer/cmd_admin.go` (~530 LOC):
  - `handleInstallApt(args)` — subcommands: install, remove, purge,
    search, list-installed, upgrade, clean, repo (add / remove).
    Flags: `--dry-run`, `--gpg-url`, `--gpg-key`, `--purge`.
  - `handleInstallIso(args)` — positional os-type +
    `--variant=NAME`, `--output=DIR` / `-o`. Defaults match the
    previous cmd: osType=windows11, variant=latest.
  - `handleInstallChocolatey(args)` — subcommands: install, uninstall,
    search, list, upgrade, info, self-install. Flag: `--dry-run`.
- `src/helpers/ptx-installer/main.go` dispatches lowercase packageName
  values `"apt"`, `"iso"`, `"chocolatey"`, `"choco"` to the new
  handlers before the generic package-install path.
- Result: **PASS** — handler surface covers all subcommands previously
  exposed via cobra in cmd/install_apt.go (8 subs), install_iso.go
  (1 sub), install_chocolatey.go (7 subs). All call into the shared
  packages directly.

### TC07 — Runtime behavioural test (CONDITIONAL — host: Windows)

- `apt-get` is not present on the Windows test host; APT subcommands
  cannot be exercised end-to-end here.
- `choco` is not installed on the test host; Chocolatey subcommands
  cannot be exercised end-to-end here.
- ISO download is gated on elevation (`ptx-installer.exe` requires
  Windows UAC).
- Result: **CONDITIONAL PASS** — source-level equivalence with the
  pre-refactor command logic was reviewed line-by-line (manual diff
  vs. `cmd/install_apt.go` etc. as in commit `c5484f0`). For full
  container-based testing (per ISSUE-DEVELOPMENT-METHODOLOGY) a
  follow-up tester-linux acceptance run is recommended before this
  becomes part of a release.

## Functional Tests Checklist

- [x] `portunix install apt source add/remove/list`,
      `portunix install iso ...`, `portunix install chocolatey ...`
      reach a real handler (no longer dead-code-path in main binary)
- [x] `src/cmd/install_apt.go`, `install_iso.go`,
      `install_chocolatey.go` deleted; dispatch handles everything via
      helper
- [x] No regression in main binary surface

## Regression Tests Checklist

- [x] Existing functionality in `app/install/installer.go` unaffected
      (TC05 + TC03 confirm aliases keep `apt.NewAptManager()` call site
      compiling)
- [x] Cross-platform build (TC03)

## CI Notes

- The dead-code-path observation: before this refactor,
  `portunix install apt …`, `portunix install iso …`, and
  `portunix install chocolatey …` in the main binary were
  **unreachable** because the dispatcher routes `install` to
  ptx-installer before cobra processes subcommands. The refactor moves
  the implementation to where the dispatcher actually directs traffic,
  so users gain functionality that was previously broken.
- A Linux-host follow-up acceptance run is recommended for full apt
  subcommand coverage (currently CONDITIONAL on the Windows tester).
- The pre-existing `go vet` issue with
  `test/integration/issue_037_mcp_serve_test.go` (Linux syscall on
  Windows) is unrelated to this refactor.

## Final Decision

**STATUS**: PASS (with one CONDITIONAL on TC07 — see CI Notes)

**Approval for merge**: YES — the conditional item (Linux runtime apt
test) is recommended as a follow-up but does not block the refactor:
the source-level equivalence is verified and pre-refactor behaviour
was a dead-code path anyway.

**Date**: 2026-05-19

**Tester signature**: Tester (generic role). Source-level acceptance
PASS. Cross-platform runtime acceptance deferred to tester-linux.
