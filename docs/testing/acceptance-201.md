# Acceptance Protocol - Issue #201 (Windows Inline `installScript` Is Not Executed, Yet Reported as Success)

**Issue**: Windows Inline `installScript` Is Not Executed, Yet Reported as Success (`portunix install uv`)
**Branch**: `fix/201-windows-inline-install-script` (commit `c4bbbe9`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-10-06
**Testing OS**:

- Windows 11 Pro 10.0.26200 (host) — unit tests, `go vet`, build only, nothing installed
- Windows 11 Enterprise 10.0.26100 AMD64 (Windows Sandbox, clean instance per run,
  user `WDAGUtilityAccount`, elevated) — all installation scenarios
- Ubuntu 22.04.5 LTS (container `ubuntu:22.04`, Docker via `portunix container`, root) —
  Linux binaries cross-compiled from the branch (`portunix`, `ptx-installer`)

## Scope Note

The fix replaces how `ptx-installer` hands inline commands to the shell
(`shellCommand`: raw `cmd /d /s /c "<script>"` on Windows, `sh -c` elsewhere) and adds a
verification step after `type: script` installs. All installation scenarios ran in
Windows Sandbox or in a Linux container, never on the host.

A **baseline run** used the same kit with only `ptx-installer.exe` rebuilt from `main`
(`5e17b87`), to show that each scenario fails before the fix.

## Test Summary

- Total test scenarios: 17 (6 unit + 5 sandbox + 3 container + 3 regression)
- Passed: 17
- Failed: 0
- Skipped: 0

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `shellCommand` quoting, exit codes, verification failure / success, target message | host, `go test` |
| E2E | `portunix install uv` on clean Windows, negative network case, new terminal, reinstall | Windows Sandbox |
| Baseline | same E2E scenarios with `ptx-installer` from `main` | Windows Sandbox |
| E2E Linux | `type: script` installs and their verification with `sh -c` and `refreshPath` | `ubuntu:22.04` container |
| Regression | full `ptx-installer` test suite, `go vet` Windows + Linux, `make build` | host |

Sandbox runs are driven by a PowerShell runner (mapped folder, `LogonCommand`) that
records exit code, output and verdict per scenario. A new terminal is simulated by
building `PATH` from the machine and user registry values.

## Test Cases

### Unit Tests (host)

- [x] `TestShellCommand_QuotedArgumentsReachShellUnchanged` — `echo "hello world"`
  prints `"hello world"`. Pre-fix argument handling prints `\"hello world\"`
- [x] `TestShellCommand_NestedQuotesForPowerShell` — `powershell -c "Write-Output ('ptx' + '-ok')"`
  prints `ptx-ok`. Pre-fix argument handling makes PowerShell echo the script text
- [x] `TestShellCommand_ExitCodePropagates` — `exit 3` returns an error
- [x] `TestInstallScript_VerificationFailureFails` — script `exit 0`, verification `exit 1`:
  `Install` returns `verification failed`
- [x] `TestInstallScript_VerificationSuccessPasses`
- [x] `TestUsesInstallPath`

### E2E — Windows Sandbox

| ID | Given / When / Then | Fixed build | Baseline (`main`) |
| -- | ------------------- | ----------- | ----------------- |
| T201-01 | Given a clean sandbox, when `portunix install uv --dry-run`, then exit 0 and no `target: ./site` | PASS, exit 0 | PASS |
| T201-02 | Given `astral.sh` blocked in `hosts`, when `portunix install uv`, then exit non-zero and no success message | PASS, exit 1, `script failed: exit status 1` | FAIL, exit 0, `Installation completed successfully!` |
| T201-03 | Given a clean sandbox, when `portunix install uv`, then exit 0, `uv.exe` in `%USERPROFILE%\.local\bin`, `Verification passed` | PASS, 9 s | FAIL, exit 0, `uv.exe` missing |
| T201-04 | Given T201-03, when `uv --version` in a new terminal, then exit 0 | PASS, `uv 0.12.23` | FAIL, `'uv' is not recognized` |
| T201-05 | Given uv installed, when `portunix install uv` again, then exit 0 | PASS | PASS |

Fixed build, T201-03 (excerpt):

```text
downloading uv 0.12.23 (x86_64-pc-windows-msvc)
installing to C:\Users\WDAGUtilityAccount\.local\bin
everything's installed!
🔍 Verifying installation: uv --version
✅ Verification passed
✅ Installation completed successfully!
```

Baseline, T201-03 — the bug from the issue, reproduced:

```text
📝 Running install script (target: ./site)...
irm https://astral.sh/uv/install.ps1 | iex
✅ Script installation completed
✅ Installation completed successfully!
```

### E2E — Linux Container

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| L1 | Given a clean `ubuntu:22.04` without `curl`, when `portunix install uv`, then exit non-zero | PASS, exit 1: `curl: not found`, the pipeline `curl … \| sh` exits 0, and `verification failed: 'uv --version' did not succeed after the install script ran` |
| L2 | Given `curl` installed and `PATH` without `~/.local/bin`, when `portunix install uv`, then exit 0 and `Verification passed` | PASS, uv 0.12.23 in `/root/.local/bin`, also works in a new login shell |
| L3 | Given a fresh container with `curl` only and no `rustc` on `PATH`, when `portunix install rust --variant stable`, then exit 0 and `Verification passed` | PASS, `rustc` still absent from the plain shell `PATH`; verification found it through `refreshPath` (`~/.cargo/bin`), rustc 1.99.0 |

L1 is the real-world form of the defect: a script that exits 0 without installing
anything. Before the fix it was reported as success.

### Acceptance Criteria

- [x] **AC1** — On a clean Windows, `portunix install uv` installs `uv.exe` into
  `%USERPROFILE%\.local\bin` and `uv --version` works in a new terminal (T201-03, T201-04)
- [x] **AC2** — When an install script does nothing, `portunix install` exits non-zero
  and says verification failed: end to end on Linux (L1), by
  `TestInstallScript_VerificationFailureFails` on Windows (real `cmd` path); in the
  sandbox a failed download now exits 1 instead of reporting success (T201-02)
- [x] **AC3** — Unit test covers an inline Windows script with double-quoted arguments
  reaching the shell unchanged
- [x] **AC4** — Other `type: script` packages were checked: Windows inline scripts
  written in PowerShell syntax are a separate defect, filed as issue #203; the
  `spice-guest-*` verification was changed from `Get-Service` to `sc query`
  (not executed — no SPICE environment)

### Regression

- [x] Full `ptx-installer` test suite passes (root, `engine`, `registry`)
- [x] `go vet` clean for `GOOS=windows` and `GOOS=linux`
- [x] `make build` succeeds (main binary + 19 helpers)

## Coverage

| Area | Covered by |
| ---- | ---------- |
| Windows quoting (`shellCommand`) | unit + sandbox |
| Verification after `type: script` | unit (fail + pass) + sandbox (pass) |
| `refreshPath` from registry | sandbox T201-03 (uv installed into a new user `PATH` entry) |
| `postInstall` loops using `shellCommand` | not exercised by uv — exercised in the #202 run (python embeddable) |
| Linux `sh -c` and verification | container L1, L2, L3 |
| Linux `refreshPath` (`~/.local/bin`, `~/.cargo/bin`) | container L2, L3 |

## Failure Injection

- Name resolution of the download host blocked through `hosts` (T201-02)
- Missing prerequisite (`curl`) so the install script silently does nothing (L1)

## CI Notes

- Unit tests run on any Windows runner; `TestShellCommand_NestedQuotesForPowerShell`
  skips on non-Windows
- E2E needs Windows Sandbox (or a disposable Windows VM) and network access to
  `astral.sh`

## Observations / Recommendations (non-blocking)

1. `uv.json` declares `"packages": ["curl"]` for Linux, but the `script` install type
   does not install them, so uv fails on a minimal system (L1). The failure is now
   reported correctly; installing declared packages first is a separate improvement
2. The spinner writes `\r` frames to stdout; in redirected logs they pile up on one line.
   Consider disabling the spinner when stdout is not a terminal
3. `refreshPath` runs only before verification; an install script that needs a tool
   installed earlier in the same session still sees the old `PATH`

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-10-06
**Tester signature**: Claude Code (tester role) — to be confirmed by zdendaku
