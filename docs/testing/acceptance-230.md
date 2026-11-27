# Acceptance Protocol - Issue #230 (PORTUNIX_CWD plugin dispatch on Windows)

**Issue**: Verify PORTUNIX_CWD plugin dispatch on Windows
**Branch**: `fix/230-windows-portunix-cwd` (commit `d8eb1dd4`; retest 1 on `8f2694fe`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-11-06
**Testing OS**: Windows 11 Pro 10.0.26300.9711, amd64 (host); Go 1.26.0, PowerShell 7.6.3,
Windows PowerShell 5.1.26100.9711, Temurin JDK 21.0.6

No software was installed and no host configuration was changed, so the tests ran on the
host. Stub plugins ran against an isolated user profile (`USERPROFILE` pointing to a
temporary directory with its own `.portunix\plugins\registry.json`). The SHA-256 of the real
plugin registry was identical before and after the run. Stub plugins read only
`PORTUNIX_CWD`; the environment was never dumped.

## Scope Note

**Independence note**: the branch adds only tests (`src/cmd/plugin_dispatcher_test.go`); it
was written by Claude Code in the developer role in the same session. The E2E scenarios
below were designed separately in the tester role and run against the real binary.

**Deviation — `cabinet` plugin not available**: `cabinet` is not installed on the Windows
host, the local `portunix-plugins` checkout contains only its README, and
`portunix plugin list-available` cannot reach the plugin repository. Criteria 2–4 were
therefore verified with a native stub plugin that emulates the `cabinet pdf prepare` path
handling (relative paths resolved against `PORTUNIX_CWD`, absolute paths used as is). This
verifies the dispatcher contract, which is the subject of #229; the path handling of the
real `cabinet` plugin itself was verified on Linux in `acceptance-229.md`.

**Not executed**:

- `go test -race` — needs cgo and a C toolchain, not present on the host
- Caller directory deleted while in use — Windows does not allow deleting the working
  directory of a running shell; the code path (`os.Getwd` error) is covered by the unit test
  `unknown caller directory leaves variable unset` and by F4 in `acceptance-229.md`
- Python script (`.py`) and Python wheel runtimes as E2E — the host Python installation is
  broken (stale venv); both runtimes are covered by `TestBuildPluginCommand_AllRuntimes`

## Test Summary

- Total test scenarios: 12
- Passed: 12
- Failed: 0
- Skipped: 3 (see Not executed)
- Defects found: 0
- Observations: 2 (O1, O2)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `gofmt -l`, `go vet` (linux), `make build` | host |
| Unit | `TestPluginEnvironment`, `TestBuildPluginCommand_*` incl. `.ps1` stub via `powershell.exe` | host |
| Mutation | remove the Windows case-insensitive branch of `pluginEnvironment` | host |
| Baseline | binary built from `fb530b63` (`main` before #229) | host, isolated profile |
| E2E | real `portunix.exe` + stub plugins: native, Python wrapper (`.ps1`), Java | host, isolated profile |
| Acceptance | `pdf prepare` emulation per criteria 2–4, special characters, inherited variable | host, isolated profile |
| Regression | plugin ignoring the variable: output and exit code, branch vs baseline | host, isolated profile |

Stub plugins (temporary, not committed):

- `stub-native` (Go): prints `PORTUNIX_CWD`, the number of environment entries whose name
  equals `PORTUNIX_CWD` case-insensitively, its working directory and arguments; with
  `pdf prepare [--dry-run] [-o OUT] IN` it emulates `cabinet`
- `stub-ps` (runtime `python`, `.ps1` wrapper run through `powershell.exe`) and `stub-java`
  (`.jar`): print `PORTUNIX_CWD`, working directory and arguments
- `stub-plain` (Go): does not read `PORTUNIX_CWD`; `fail` exits with code 7

`<S>` below is the temporary test root, `<SP>` is `<S>\work\my books ř\Přílohy` (space and
non-ASCII characters).

## Test Cases

### TC01: dispatcher tests pass on Windows (criterion 1)

- **Given** the branch checked out on Windows
- **When** `go test portunix.ai/cmd -run "TestPluginEnvironment|TestBuildPluginCommand" -v`
  is run from the repository root
- **Then** all tests pass, including the `.ps1` stub launched through `powershell.exe`

### TC02: the new tests detect a regression (mutation)

- **Given** the condition in `pluginEnvironment` reduced to an exact `key == PORTUNIX_CWD`
  comparison
- **When** the tests of TC01 are run
- **Then** `differently cased inherited name` fails

### TC03: all runtimes receive the caller directory (E2E)

- **Given** a shell in `<S>\work\caller`
- **When** `portunix <stub> a "b c"` is run for native, Python wrapper and Java
- **Then** `PORTUNIX_CWD` is the caller directory, the working directory is the install path,
  arguments are `a|b c`; the baseline binary leaves the variable unset

### TC04: relative input path (criterion 2)

- **Given** `book.pdf` and `sub\inner.pdf` in `<SP>`
- **When** `portunix stub-native pdf prepare --dry-run book.pdf` is run in `<SP>`
  (and `sub\inner.pdf`, and `..\book.pdf` from `<SP>\sub`)
- **Then** the command succeeds; the baseline fails with `input file not found`

### TC05: relative output path (criterion 3)

- **Given** `book.pdf` in `<SP>`
- **When** `pdf prepare -o out.pdf book.pdf` and `-o sub\o2.pdf book.pdf` are run in `<SP>`
- **Then** the outputs are created in `<SP>`, no file appears in any plugin directory

### TC06: absolute paths and another drive (criterion 4)

- **Given** a shell on drive `D:` and `book.pdf` on drive `C:`
- **When** `pdf prepare --dry-run <C: absolute>` and `-o <C: absolute> <C: absolute>` are run
- **Then** both succeed, the dry run output is identical to the baseline

### TC07: special characters in the caller directory (criterion 5)

- **Given** a shell in `<SP>`
- **When** each stub is dispatched (from PowerShell and from `cmd.exe` with code page 65001)
- **Then** `PORTUNIX_CWD` equals `<SP>` byte for byte

### TC08: inherited variable, also differently cased (criterion 6)

- **Given** `PORTUNIX_CWD=C:\stale` set in PowerShell, or `Portunix_Cwd=C:\stale` set in
  `cmd.exe` (the name is stored in that case, as `set Portunix_Cwd` shows)
- **When** each stub is dispatched from `<SP>`
- **Then** the plugin sees exactly one entry with the value `<SP>`; the baseline passes
  `C:\stale` through

### TC09: plugin ignoring the variable (criterion 7)

- **Given** `stub-plain`
- **When** it is run with `x "y z"`, `--help`, no arguments, and `fail`, with the branch and
  the baseline binary
- **Then** stdout, stderr and exit code are identical

## Test Results

### Static Checks and Unit Tests (TC01, TC02)

| Check | Result |
| ----- | ------ |
| `gofmt -l src/cmd/plugin_dispatcher_test.go` | PASS (no output) |
| `GOOS=linux go vet portunix.ai/cmd` | PASS |
| `make build` (main binary + all helpers) | PASS |
| `TestPluginEnvironment` (6 subtests) | PASS |
| `TestBuildPluginCommand_AllRuntimes` (5 subtests) | PASS |
| `TestBuildPluginCommand_StubPluginReceivesCallerDir` (`.ps1` through `powershell.exe`) | PASS |
| `TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir` (`.ps1`, `knihy ř\Přílohy`, `Portunix_Cwd` inherited) | PASS |
| `go test portunix.ai/cmd` (whole package) | PASS |
| Mutation: exact-name comparison only — `differently cased inherited name` fails with `Portunix_Cwd values = [/stale], want none on Windows` | PASS (test detects the defect) |

### E2E — All Runtimes (TC03)

| Runtime | Branch `PORTUNIX_CWD` | Baseline | Working directory | Arguments | Result |
| ------- | --------------------- | -------- | ----------------- | --------- | ------ |
| native | `<S>\work\caller`, 1 entry | unset | install path | `a\|b c` | PASS |
| Python wrapper (`.ps1`) | `<S>\work\caller` | unset | install path | `a\|b c` | PASS |
| Java (`.jar`) | `<S>\work\caller` | unset | install path | `a\|b c` | PASS |

### Acceptance — `pdf prepare` Emulation (TC04–TC06)

| ID | Caller | Command | Observed | Result |
| -- | ------ | ------- | -------- | ------ |
| A1 | `<SP>` | `--dry-run book.pdf` | branch: success; baseline: `input file not found: book.pdf`, exit 1 | PASS |
| A2 | `<SP>` | `--dry-run sub\inner.pdf` | success | PASS |
| A3 | `<SP>\sub` | `--dry-run ..\book.pdf` | success, resolved to `<SP>\book.pdf` | PASS |
| A4 | `<SP>` | `-o out.pdf book.pdf` | `<SP>\out.pdf` created, not in the plugin directory | PASS |
| A5 | `<SP>` | `-o sub\o2.pdf book.pdf` | `<SP>\sub\o2.pdf` created | PASS |
| A6 | `D:\…\portunix` | `--dry-run <C: absolute>` | success, identical to baseline | PASS |
| A7 | `D:\…\portunix` | `-o <C: absolute> <C: absolute>` | file created at the absolute path | PASS |
| A8 | `D:\…\portunix` | no arguments | `PORTUNIX_CWD=D:\Dev\CassandraGargoyle\portunix\portunix` | PASS |

File listing (name, size, modification time) of all plugin directories was identical before
and after the `-o` runs.

### Special Characters and Inherited Variable (TC07, TC08)

| ID | Scenario | native | `.ps1` | Java | Result |
| -- | -------- | ------ | ------ | ---- | ------ |
| S1 | caller `<SP>` from PowerShell | `<SP>`, 1 entry | `<SP>` | `<SP>` | PASS |
| S2 | caller `<SP>` from `cmd.exe` (chcp 65001) | `<SP>`, 1 entry | — | — | PASS |
| S3 | `PORTUNIX_CWD=C:\stale` inherited | `<SP>`, 1 entry | `<SP>` | `<SP>` | PASS |
| S4 | `Portunix_Cwd=C:\stale` inherited (`cmd.exe`) | `<SP>`, 1 entry | `<SP>` | `<SP>` | PASS |
| S5 | S4 with the baseline binary | `C:\stale` | `C:\stale` | `C:\stale` | PASS (old behaviour reproduced) |

### Regression — Plugin Ignoring the Variable (TC09)

| Arguments | Exit code (branch / baseline) | stdout | stderr | Result |
| --------- | ----------------------------- | ------ | ------ | ------ |
| `x "y z"` | 0 / 0 | identical | identical | PASS |
| `--help` | 0 / 0 | identical | identical | PASS |
| (none) | 0 / 0 | identical | identical | PASS |
| `fail` | 7 / 7 | identical | identical | PASS |

## Acceptance Criteria

| # | Criterion | Status |
| - | --------- | ------ |
| 1 | Dispatcher tests pass on Windows | PASS |
| 2 | Relative input path from an arbitrary directory | PASS (stub emulating `cabinet`, see Deviation) |
| 3 | `-o out.pdf` writes into the caller's directory | PASS (stub emulating `cabinet`, see Deviation) |
| 4 | Absolute paths, including another drive | PASS (stub emulating `cabinet`, see Deviation) |
| 5 | Caller directory with a space and a non-ASCII character passed intact | PASS |
| 6 | Inherited `PORTUNIX_CWD` / `Portunix_Cwd` — plugin sees exactly one, correct value | PASS |
| 7 | Plugin not reading the variable behaves as before | PASS |
| 8 | `acceptance-229.md` updated with Windows results, status PASS | PASS |

## Coverage

- `pluginEnvironment`: all branches covered on Windows, including the case-insensitive key
  comparison; the mutation in TC02 proves the coverage
- `buildPluginCommand`: every runtime branch by unit test; native, `.ps1` wrapper and Java
  also end to end through the real binary
- `executePlugin` (`os.Getwd`, registry lookup, exit code propagation): covered by the E2E
  runs, not by an automated test

## CI Notes

- Run from the repository root by import path: `go test portunix.ai/cmd -run
  "TestPluginEnvironment|TestBuildPluginCommand" -v`
- On Windows the stub tests need `powershell.exe`; they write the result as UTF-8 into a
  file, so they do not depend on the console code page
- In VS Code, gopls reports `BrokenImport` for files in `src/cmd`; it does not affect
  `go test` by import path

## Observations

### O1: Go `os/exec` deduplicates environment names case-insensitively on Windows

With the mutation of TC02 applied, the E2E-style test
`TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir` still passes: on Windows,
`os/exec` removes duplicate names case-insensitively and keeps the last entry, which is the
`PORTUNIX_CWD` appended by the dispatcher. The Windows branch of `pluginEnvironment` is
therefore defence in depth — it keeps `cmd.Env` clean and independent of that runtime
detail. Only the unit test detects its removal. The production code needs no change; the
E2E-style test should inspect `cmd.Env` directly.

### O2: `Get-ChildItem -Recurse` unreliable in the temporary directory

During E2E setup, `Get-ChildItem -Recurse` in PowerShell 7.6.3 resolved child names against
the current location and returned an empty or incomplete listing. The plugin directory check
was repeated with `[System.IO.Directory]::GetFiles`. Affects only the test harness.

### Resolution of O1 and O2

- **O1**: resolved in commit `8f2694fe`. `TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir`
  now counts entries named `PORTUNIX_CWD` case-insensitively in `cmd.Env` before the stub
  is started (new helper `envNameCountFold`): 1 on Windows, 2 elsewhere, where
  `Portunix_Cwd` is an unrelated variable and must be kept. Verified in Retest 1
- **O2**: no change — it concerns only the temporary test script outside the repository

## Retest 1

**Date**: 2026-11-06
**Commit**: `8f2694fe` (`fix(230): resolve acceptance observation O1`, test file only)
**Testing OS**:

- Windows 11 Pro 10.0.26300.9711, amd64 (host); Go 1.27.3 (`C:\Program Files\Go`)
- Ubuntu 24.04.3 LTS in WSL 2 (kernel 6.18.33.2-microsoft-standard-WSL2); Go 1.26.0

**Toolchain note**: the Go installation used in the first run (`D:\Go`, Go 1.26.0 through
`GOTOOLCHAIN=auto`) disappeared from the host during the retest. The first mutation attempt
therefore did not run at all (`go: command not found`) and was discarded; all retest steps
below were repeated with Go 1.27.3 from `C:\Program Files\Go`.

The production binary is unchanged (the commit touches only `plugin_dispatcher_test.go`),
so the E2E scenarios TC03–TC09 were not repeated.

### R1: static checks and tests (Windows)

| Check | Result |
| ----- | ------ |
| `gofmt -l` (CRLF stripped) on `plugin_dispatcher_test.go` | PASS (no output) |
| `go vet portunix.ai/cmd` (windows), `GOOS=linux go vet portunix.ai/cmd` | PASS |
| `TestPluginEnvironment` (6 subtests), `TestBuildPluginCommand_AllRuntimes` (5 subtests) | PASS |
| `TestBuildPluginCommand_StubPluginReceivesCallerDir` | PASS |
| `TestBuildPluginCommand_StubPluginReceivesSpecialCallerDir` | PASS |
| `go test portunix.ai/cmd` (whole package) | PASS |

### R2: the tests detect a regression (mutation, Windows)

| ID | Mutation of `pluginEnvironment` | Failing tests | Result |
| -- | ------------------------------- | ------------- | ------ |
| M1 | exact-name comparison only (Windows branch removed) | `differently_cased_inherited_name`; `StubPluginReceivesSpecialCallerDir` — `cmd.Env has 2 entries named PORTUNIX_CWD case-insensitively, want 1` | PASS (O1 resolved) |
| M2 | inherited value never dropped | `replaces_inherited_value`, `differently_cased_inherited_name`, `unknown_caller_directory_leaves_variable_unset`, `StubPluginReceivesSpecialCallerDir` | PASS |

`plugin_dispatcher.go` was restored after each mutation (`git diff` empty).

Note: under M2, `TestBuildPluginCommand_StubPluginReceivesCallerDir` still passes, because
`os/exec` deduplicates exact duplicate names on every platform and keeps the last entry.
The mutation is detected by four other tests, so no change is needed.

### R3: Linux (WSL)

| Check | Result |
| ----- | ------ |
| `go test portunix.ai/cmd -run "TestPluginEnvironment\|TestBuildPluginCommand" -v` — all 13 tests and subtests, including the `want = 2` branch and the `.sh` stub in `knihy ř/Přílohy` | PASS |

### Retest 1 Decision

**STATUS**: PASS — O1 resolved, no regression on Windows or Linux

## Recommendations

1. When `cabinet` becomes installable on Windows, run `cabinet pdf prepare --dry-run
   book.pdf` and `-o out.pdf book.pdf` from a directory such as `<SP>` once, to cover the
   plugin's own path handling on Windows

## Final Decision

**STATUS**: PASS

All criteria are met on Windows. Criteria 2–4 were verified with a stub emulating `cabinet`,
because the real plugin is not available on the Windows host (see Deviation); the dispatcher
behaviour these criteria depend on is fully verified. No defect was found, no production
code change is needed. Observation O1 was resolved in the tests and verified in Retest 1
(Windows and Linux).

**Approval for merge**: YES (commit `8f2694fe`)
**Date**: 2026-11-06
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
