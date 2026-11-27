# Acceptance Protocol - Issue #203 (Windows `installScript` Written in PowerShell Is Run by `cmd.exe`)

**Issue**: Windows `installScript` Written in PowerShell Is Run by `cmd.exe` (`portunix install claude-desktop`)
**Branch**: `fix/203-windows-powershell-install-scripts` (commits `bfb57fa2`, `5035f758`;
retest 1 on `40acd909`, retest 2 on `e80be13d`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-07 (retest 2: 2026-11-08)
**Testing OS**:

- Windows 11 Pro 10.0.26300 (host) — unit tests, `go vet`, `make build`, dry-runs only,
  nothing installed
- Windows 11 Enterprise 10.0.26100 (Windows Sandbox, clean instance, user
  `WDAGUtilityAccount`, Windows PowerShell 5.1.26100.9711) — all installation scenarios

## Scope Note

The fix adds the variant field `scriptShell` (`powershell` | `cmd` | `sh`). For
`powershell`, `ptx-installer` writes the inline script to a temporary `.ps1` and runs it
with `powershell.exe -File`, stopping on the first error and passing the exit code of
the last native command through. Fifteen Windows variants in twelve packages declare
`scriptShell: powershell`. A registry test fails when a Windows PowerShell script
lacks the field.

All installation scenarios ran in Windows Sandbox, never on the host. No baseline run
from `main` was made: the issue already records the `cmd.exe` failure, and
`TestInstallScript_PowerShellPackage` reproduces it (same script without `scriptShell`
fails with `'$x' is not recognized`).

## Test Summary

Initial run (`bfb57fa2`, `5035f758`):

- Total test scenarios: 28 (15 unit + 6 dry-run + 7 sandbox)
- Passed: 25
- Failed: 2 (T1 claude-desktop, T3 claude-code `curl` verification)
- Skipped: 0
- Unexpected, outside the issue scope: 1 (T5 qemu false success, pre-existing)

Retest 1 (`40acd909`) — see [Retest 1](#retest-1):

- Total test scenarios: 19 (9 host CLI + 10 Windows Sandbox) plus 3 diagnostic runs
- Passed: 9
- Failed: 6 (D1: R4, W4; D2: W2, W3; D3: R1; D4: `docker install`)
- Partially met: 2 (D4: `install`, `container` / `wizard`)
- Blocked: 2 (R2, R5 by D3)
- Resolved: the `curl` verification (initial Required Change 2), the qemu false success,
  items 8 and 10 of the scope extension

Retest 2 (`e80be13d`) — see [Retest 2](#retest-2):

- Total test scenarios: 16 (10 host checks + 6 Windows Sandbox)
- Passed: 13
- Failed: 0
- Skipped: 3 (T4 qemu by decision — winget crash in Windows Sandbox; T5, T6 not run)
- **Final decision: PASS**

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `scriptCommand` for every shell, PowerShell quoting, error stop, native exit code, end-to-end `Install` with and without `scriptShell` | host, `go test` |
| Registry | all 69 manifests: Windows PowerShell scripts declare `scriptShell`, invalid values rejected | host, `go test` |
| Dry-run | `portunix install <pkg> --dry-run` for affected packages | host |
| E2E | real installs of affected packages on clean Windows | Windows Sandbox |
| Regression | `ptx-installer` test suite, `go vet`, `make build` | host |

The sandbox runner (mapped folder, `LogonCommand`) records stdout, stderr and duration
per scenario.

## Test Cases

### Unit Tests (host)

- [x] `TestScriptCommand_PowerShellRunsVariablesAndQuotes` — `$var`, `Join-Path` and
  a literal `"quoted"` reach PowerShell unchanged
- [x] `TestScriptCommand_PowerShellCmdletErrorFails` — a failing cmdlet stops the
  script with a non-zero exit code
- [x] `TestScriptCommand_PowerShellNativeExitCodePropagates` — `cmd /c exit 4` inside
  the script gives exit code 4
- [x] `TestScriptCommand_DefaultShell`, `TestScriptCommand_CmdOnlyOnWindows`,
  `TestScriptCommand_UnknownShellFails`
- [x] `TestInstallScript_PowerShellPackage` — the PowerShell package installs. The same
  script without `scriptShell` fails under `cmd.exe`
- [x] `TestUsesPowerShellSyntax` — the detector separates PowerShell, cmd, sh and
  `powershell -c "..."` wrapper lines (uv)
- [x] `TestPackages_WindowsPowerShellScriptsDeclareShell` — passes on the branch. Before
  the manifest change it reported all 15 variants, including `chocolatey`, which the
  issue did not list
- [x] `TestValidateVariant_ScriptShell`
- [x] #201 regression: `TestShellCommand_*`, `TestInstallScript_Verification*`

### Dry-Run (host)

- [x] `claude-desktop`, `claude-code --variant npm`, `claude-code --variant curl`,
  `powershell`, `wireguard`, `qemu`: the registry loads 69 packages with 0 errors and the
  variant is resolved

### E2E — Windows Sandbox

| ID | Command | Result | Verdict |
| -- | ------- | ------ | ------- |
| T1 | `portunix install claude-desktop` | Script runs in PowerShell (no `'$url' is not recognized`). `Invoke-WebRequest` to `https://claude.ai/download/win` gets the Cloudflare challenge page (HTTP 403). Install fails with `script failed: exit status 1`, nothing installed, no false success | ❌ FAIL (AC1) |
| T2a | `portunix install claude-code --variant npm` without Node.js | `Write-Host` runs. `npm` not found ends the script, install fails | ✅ PASS (correct failure) |
| T2b | `portunix install nodejs` | Node.js 18.20.6 MSI installed | ✅ PASS (setup) |
| T2c | `portunix install claude-code --variant npm` | npm install runs, verification `claude --version` passes, `2.1.197 (Claude Code)` | ✅ PASS (AC2) |
| T3 | `portunix install claude-code --variant curl` | `irm https://claude.ai/install.ps1 \| iex` installs Claude Code 2.1.319 into `%USERPROFILE%\.local\bin`. The installer says the directory is not in `PATH`, and verification `claude --version` fails | ❌ FAIL (false negative) |
| T4 | `portunix install powershell` | PowerShell 7.4.6 MSI installed. The double-quoted `ArgumentList` with `$output` works and `pwsh --version` verifies | ✅ PASS |
| T5 | `portunix install qemu` | `$url` now expands (`Please download QEMU from https://qemu.weilnetz.de/w64/`). Reports `Installation completed successfully!` although nothing is installed. The Windows platform has no `verification` | ⚠️ pre-existing |

### Regression (host)

- [x] `go vet ./...` in `src/helpers/ptx-installer`
- [x] `go test ./...`: `ptx-installer`, `engine` and `registry` pass
- [x] `make build` builds the main binary and all helpers
- [ ] `engine/skills` `TestWriteFailureReportedAsFailed` fails on Windows. It depends on
  a read-only directory made with `chmod`, which Windows ignores. The package is not
  touched by this branch, so the failure is unrelated

## Acceptance Criteria

- [ ] On a clean Windows, `portunix install claude-desktop` installs Claude Desktop
  without a window waiting for input, and it starts afterwards. **Not met**: the
  download URL is behind a Cloudflare challenge (T1)
- [x] `portunix install claude-code` runs its npm install on Windows (T2c)
- [x] A package can choose `powershell`, `cmd` or `sh` for its inline script (unit, T2c, T4)
- [x] The registry test catches a PowerShell script declared without
  `scriptShell: powershell`

## Coverage

| Area | Covered by |
| ---- | ---------- |
| `scriptShell` = powershell | unit + sandbox T1–T4 |
| `scriptShell` = default / cmd / sh / invalid | unit |
| Error stop (`$ErrorActionPreference = 'Stop'`) | unit + sandbox T1, T2a |
| Native exit code passthrough | unit |
| UTF-8 BOM (`winget` `✓`) | not exercised end-to-end — `winget` not installed in this run |
| `chocolatey`, `wireguard`, `virtualbox`, `virt-viewer`, `spice-*`, `winget` | registry test + dry-run only |
| Linux / `pwsh` path | not tested (no PowerShell manifests for Linux) |

## Failure Injection

- Blocked download (Cloudflare challenge on `claude.ai/download/win`, T1): the failure
  stops the script and the exit code is non-zero
- Missing prerequisite (`npm` absent, T2a): CommandNotFound stops the script

## CI Notes

- PowerShell unit tests skip when `powershell.exe` / `pwsh` is not on `PATH`
- `TestInstallScript_PowerShellPackage` is Windows-only
- E2E needs Windows Sandbox and network access to `claude.ai`, `nodejs.org`, `npmjs.org`
  and `github.com`

## Required Changes (blocking)

1. **claude-desktop download source** — `https://claude.ai/download/win` cannot be
   fetched by a script. Use a source that is reachable without a browser challenge.
   Options: `winget install Anthropic.Claude` as the preferred variant, or the direct
   installer URL behind the redirect. Then retest AC1, including that the installer
   runs unattended with `--silent` and the app starts
2. **claude-code `curl` variant** — add `%USERPROFILE%\.local\bin` to `PATH`
   (`environment.PATH_APPEND`) or verify with the full path, so a working install is not
   reported as failed (T3)

## Observations / Recommendations (non-blocking)

1. `qemu` on Windows is a placeholder that only prints a link, yet it reports success.
   Add a `verification` or mark the variant as manual (pre-existing, related to #201)
2. `claude-code` declares the dependency `nodejs`, but the install only warns
   `Dependency nodejs should be installed first` and does not install it (T2a)
3. `portunix sandbox generate --help-expert` generates a `.wsb` file instead of showing
   help
4. `portunix sandbox generate --map-folder` splits on `:`, so Windows paths with a drive
   letter (`C:\...`) are rejected

## Final Decision

**STATUS**: FAIL

The `scriptShell` mechanism works as specified: PowerShell scripts now run, and
failures are reported. Acceptance criterion 1 (Claude Desktop installs) is not met
because of the download source. The `curl` variant changed on this branch reports a
false failure.

**Approval for merge**: NO — return to developer for Required Changes 1 and 2
**Date**: 2026-11-07
**Tester signature**: Claude Code (tester role) — to be confirmed by zdendaku

## Retest 1

**Commit**: `40acd909` — `fix(203): resolve acceptance findings and scope extension`
**Date**: 2026-11-07
**Testing OS**: Windows 11 Enterprise 10.0.26100 (Windows Sandbox, three clean instances,
Windows PowerShell 5.1.26100.9711); host as above for CLI checks, nothing installed on the host

The developer fix installs `claude-desktop` and `qemu` through winget, adds `PATH_APPEND`
to `claude-code` and exposes it to the `script` verification, installs missing
dependencies, guards `--help-expert` / `--help-ai` on subcommands and parses drive-letter
paths in `--map-folder`. The issue file was extended with items 5–10.

Binaries were rebuilt with `make build` from `40acd909`.

### Run 1 — Windows Sandbox (clean, no winget, no Node.js)

| ID | Command | Result | Verdict |
| -- | ------- | ------ | ------- |
| R1 | `portunix install winget` (default variant `github`) | The PowerShell script runs and prints its non-ASCII messages correctly (BOM). `Add-AppxPackage` fails: winget v1.6.3482 needs the framework `Microsoft.UI.Xaml.2.7`, which the variant does not install. Exit 1, no false success | ❌ D3 |
| R2 | `portunix install claude-desktop` | `Winget is not installed`, exit 1 | ⛔ blocked by D3 |
| R3 | `portunix install claude-code` | `📦 Installing dependency nodejs...` → Node.js 18.20.6 MSI, then npm install, verification `claude --version` → `2.1.197`, exit 0 | ✅ PASS (item 8, AC2) |
| R4 | `portunix install claude-code --variant curl` | Exit 0, but not conclusive: `claude` from R3 was already on `PATH`. Prints `Adding to User PATH: …\.local\bin`, yet the user `PATH` in the registry does not contain it | ❌ D1 |
| R5 | `portunix install qemu` | `Winget is not installed`, exit 1 — no false success any more | ⛔ blocked by D3 |
| R6 | `portunix install powershell` | PowerShell 7.4.6, verification passes, exit 0 | ✅ PASS (regression) |

### Run 2 — Windows Sandbox (clean)

| ID | Command | Result | Verdict |
| -- | ------- | ------ | ------- |
| W4 | `portunix install claude-code --variant curl` on a system without Claude Code | Verification `claude --version` passes through `PATH_APPEND`, exit 0. The user `PATH` afterwards is `…\WindowsApps;…\Roaming\npm;C:\Program Files/nodejs`, no `.local\bin`. A new shell: `'claude' is not recognized` | ❌ D1 |
| W1 | `portunix install winget --variant powershell-module` | `Repair-WinGetPackageManager`, `winget --version` → `v1.29.380`, exit 0 | ✅ PASS |
| W2 | `portunix install claude-desktop` | `winget install failed for Anthropic.Claude: exit status 0x8a15003b`, exit 1 | ❌ D2 |
| W3 | `portunix install qemu` | `winget install failed for SoftwareFreedomConservancy.QEMU: exit status 0x8a15003b`, exit 1 | ❌ D2 |

### Run 3 — Windows Sandbox, Diagnosis of D2 (winget Called Directly)

| ID | Command | Result |
| -- | ------- | ------ |
| X1 | `winget install --id Anthropic.Claude -e --accept-package-agreements --accept-source-agreements` (what `InstallViaWinget` runs) | `Failed when searching source: msstore`, `0x8a15003b : Rest API internal error`, then `Please specify one of them using the --source option` |
| X2 | same with `--source winget` | `Successfully installed` Claude 2.42304.2 (installer `.exe` from `downloads.claude.ai`), exit 0 |
| X3 | `winget install --id SoftwareFreedomConservancy.QEMU -e --source winget …` | `Successfully installed` QEMU 11.1.0, `C:\Program Files\qemu\qemu-system-x86_64.exe` exists, exit 0 |

### Host CLI Checks (Items 9 and 10)

| Check | Result | Verdict |
| ----- | ------ | ------- |
| `sandbox generate --help-expert` / `--help-ai` | Usage of `sandbox generate`, no `.wsb` file created | ✅ |
| `system info --help-expert` | Usage of `system info` | ✅ |
| `portunix --help-expert`, `portunix --help-ai` | Root expert help / JSON help as before | ✅ |
| `mcp`, `virt`, `github`, `credential`, `plugin` with `--help-expert` | Help of the command (helper prints its own help) | ✅ |
| `container --help-expert`, `wizard --help-expert` | Error (`Unknown container subcommand`, `unknown flag`), nothing runs | ⚠️ D4 |
| `install --help-expert` | `Installing package: --help-expert` → `package not found` | ⚠️ D4 |
| `docker install --help-expert` | **Runs the Docker installation** (`Starting Docker installation…`); stopped only because Docker was already installed on the host | ❌ D4 |
| `--map-folder` with `C:\A:C:\B:true` | `.wsb` with `HostFolder C:\A`, `SandboxFolder C:\B`, `ReadOnly true` | ✅ |
| `--map-folder` with `C:\A:C:\B:maybe` or `C:\A:true` | Clear error message | ✅ |

### D1: `PATH_APPEND` Is No Longer Written to the User PATH for `script` Installs

`installScript` now prepends the `PATH_APPEND` directory to the process `PATH` before the
verification (`prependProcessPath`). Afterwards `Install` calls `AddToUserPath`, whose
`IsInUserPath` checks the **process** `PATH`, finds the directory and returns without
touching the registry, while `Install` still prints `🔧 Adding to User PATH`. Every
`script` package with `PATH_APPEND` is affected: the install and its verification pass,
but the tool is missing in every new terminal (W4). **Blocking** (item 6, and a
regression for all `script` packages with `PATH_APPEND`).

### D2: `InstallViaWinget` Fails When the `msstore` Source Is Unavailable

`InstallViaWinget` runs `winget install --id <id> -e` without `--source`. winget searches
all sources, the `msstore` source fails (`0x8a15003b`, here in Windows Sandbox) and winget
stops instead of using the `winget` source that has the package (X1). With
`--source winget` both packages install (X2, X3). `InstallViaWinget` also discards the
winget output (`cmd.Stdout = nil`), so the user sees only the exit code. **Blocking** for
AC1 and item 7.

### D3: Default winget Variant `github` Cannot Install winget

The first Windows variant of `winget.json`, `github`, downloads
`Microsoft.DesktopAppInstaller` v1.6.3482 without its framework dependency
`Microsoft.UI.Xaml.2.7`, so `Add-AppxPackage` fails on a system without winget. The
`powershell-module` variant works (W1). `claude-desktop` and `qemu` now depend on winget,
so on a system without winget they cannot be installed with the default variant.
**Blocking** for AC1 and item 7 on systems without winget (Windows Sandbox, Windows
Server, LTSC); a normal Windows 11 ships winget.

### D4: `--help-expert` / `--help-ai` on Pass-Through Commands

The guard skips commands with `DisableFlagParsing` and leaves the flags to the helper
binary. Not every helper handles them: `docker install --help-expert` runs the Docker
installation, `install --help-expert` tries to install a package named `--help-expert`,
`container` and `wizard` fail with an error. Item 9 requires that a subcommand shows its
help **without executing**. **Blocking** for `docker install` (it executes); the others
do not execute anything but do not show help either.

### Further Observations (Non-Blocking)

1. `claude-code --variant curl` installs `nodejs` too: dependencies are declared for the
   whole package, not per variant, and the native installer does not need Node.js (W4)
2. winget installs are not verified afterwards (`type: winget` does not run the platform
   `verification`)
3. `Anthropic.Claude` from the `winget` source installs an `.exe` installer (X2), not the
   MSIX listed in the current `winget-pkgs` manifest; the package verification checks
   both the Squirrel path and an Appx package. Whether the app starts after a Portunix
   install could not be checked because of D2

### Acceptance Criteria Status (Retest 1)

| Criterion | Status |
| --------- | ------ |
| AC1 claude-desktop installs and starts | ❌ D2, D3 (winget with `--source winget` installs it, X2) |
| AC2 claude-code npm | ✅ R3 |
| AC3 `scriptShell` choice | ✅ initial run |
| AC4 registry test | ✅ initial run |
| claude-code `curl` verification | ⚠️ verification passes, but D1 |
| qemu never reports false success | ✅ R5, W3; the installation itself ❌ D2 |
| Dependencies installed first | ✅ R3 |
| `--help-expert` / `--help-ai` show help without executing | ❌ D4 |
| `--map-folder` with drive letters | ✅ |

### Final Decision (Retest 1)

**STATUS**: FAIL — return to developer

1. Fix D1: write `PATH_APPEND` to the user `PATH` even when the process `PATH` already
   contains it (check the registry, not the process)
2. Fix D2: pass `--source winget` in `InstallViaWinget` and show the winget output
3. Fix D3: make `powershell-module` the default Windows variant of `winget`, or install
   the missing frameworks in the `github` variant
4. Fix D4: handle `--help-expert` / `--help-ai` for pass-through commands that do not
   support them, at least `docker install`

**Approval for merge**: NO
**Date**: 2026-11-07
**Tester signature**: Claude Code (tester role) — to be confirmed by zdendaku

## Retest 2

**Commit**: `e80be13d` (fixes of D1–D4)
**Date**: 2026-11-08
**Testing OS**: Windows 11 Pro 10.0.26300 (host, nothing installed); Windows 11 Enterprise
10.0.26100 (Windows Sandbox, clean instance, Windows PowerShell 5.1.26100.9711)

### Host Checks (Nothing Installed)

| Check | Result | Verdict |
| ----- | ------ | ------- |
| `go test -count=1` for `src/dispatcher`, `portunix.ai/cmd`, `ptx-installer` `engine` / `registry` / main | All pass | ✅ |
| `docker install --help-expert`, `docker install --help-ai` | Help of `install`, Docker installation not started | ✅ D4 |
| `install --help-expert`, `install claude-code --help-ai` | Help of `install`, no package installed | ✅ D4 |
| `container --help-expert`, `container run --help-ai`, `wizard --help-expert`, `prompt --help-ai` | Help of the command | ✅ D4 |
| `mcp --help-expert`, `virt --help-expert`, `github --help-ai` | Expert / JSON help of the helper (top level form kept) | ✅ |
| `github clone --help-expert` | Help of `github clone` | ✅ |
| `sandbox generate --help-ai`, `system info --help-expert`, `portunix --help-ai` | Help, no `.wsb` file created | ✅ |
| `portunix.exe` alone without helpers: `docker install --help-expert`, `install --help-ai`, `container --help-expert`, `proxmox --help-expert` | Cobra help, nothing runs | ✅ D4 |
| `install claude-desktop --dry-run` | Variant `winget`, type `winget` | ✅ |
| `install winget --dry-run` | Variant `powershell-module` | ✅ D3 |

### Windows Sandbox (Clean, Without winget, Node.js and Claude Code)

| ID | Command | Result | Verdict |
| -- | ------- | ------ | ------- |
| T1 | `portunix install claude-code --variant curl` | Exit 0 (202 s). The user `PATH` in the registry now contains `C:\Users\WDAGUtilityAccount\.local\bin`; a new terminal (PATH rebuilt from the registry) runs `claude --version` → `2.1.319 (Claude Code)` | ✅ D1 |
| T2 | `portunix install winget` (no variant) | Default variant `powershell-module`, `winget --version` → `v1.29.380`, exit 0 (76 s) | ✅ D3 |
| T3 | `portunix install claude-desktop` | winget output visible (`Found Claude [Anthropic.Claude] Version 2.42304.2` … `Successfully installed`), exit 0 (64 s); `%LOCALAPPDATA%\AnthropicClaude\claude.exe` exists; started: 7 Claude processes, window `Claude` | ✅ D2, AC1 |
| T4 | `portunix install qemu` | winget downloads the installer and verifies its hash, then winget itself crashes during `Starting package install...` with `0xc0000005` (access violation); Portunix reports `❌ Installation failed: winget install failed for SoftwareFreedomConservancy.QEMU: exit status 0xc0000005` — no false success. The installer process kept the runner waiting, so the run stopped here | ⏭️ Skipped (see below) |
| T5 | repeated `install claude-desktop` | Not run (runner stopped at T4) | ⏭️ Skipped |
| T6 | `docker install --help-expert` in the sandbox | Not run (runner stopped at T4); covered by the host checks above | ⏭️ Covered on host |

### T4 Decision

The qemu installation cannot be tested in Windows Sandbox for now: winget crashes
(`0xc0000005`) while running the QEMU installer, after Portunix has handed over to it with
the correct arguments (`--source winget`). In the Retest 1 diagnosis (X3) the same winget
command installed QEMU 11.1.0, so the crash is not reproducible and lies outside Portunix.
Portunix itself behaves correctly: it shows the winget output and reports the failure
with the winget exit code instead of a false success. By decision of zdendaku (2026-11-08)
T4 is recorded as **skipped** and does not block the acceptance. The qemu installation
should be verified on a regular Windows installation later.

### Test Summary (Retest 2)

- Total test scenarios: 16 (10 host checks + 6 Windows Sandbox)
- Passed: 13 (10 host checks, T1, T2, T3)
- Failed: 0
- Skipped: 3 (T4 by decision, T5 not run; T6 covered by host checks)

### Acceptance Criteria Status (Retest 2)

| Criterion | Status |
| --------- | ------ |
| AC1 claude-desktop installs and starts | ✅ T3 |
| AC2 claude-code npm | ✅ Retest 1 R3 |
| AC3 `scriptShell` choice | ✅ initial run |
| AC4 registry test | ✅ initial run |
| claude-code `curl` verification and persistent PATH | ✅ T1 |
| qemu never reports false success | ✅ T4 (failure reported); installation ⏭️ skipped |
| Dependencies installed first | ✅ Retest 1 R3 |
| `--help-expert` / `--help-ai` show help without executing | ✅ host checks |
| `--map-folder` with drive letters | ✅ Retest 1 |

### Non-Blocking Observations

1. `claude-code --variant curl` still installs `nodejs` first (dependencies are per
   package, not per variant)
2. Installs of type `winget` are not verified with the platform `verification` afterwards
3. qemu via winget in Windows Sandbox: winget crash `0xc0000005` (T4) — verify on a
   regular Windows installation

### Final Decision (Retest 2)

**STATUS**: PASS (T4 skipped by decision)

**Approval for merge**: YES
**Date**: 2026-11-08
**Tester signature**: Claude Code (tester role) — confirmed by zdendaku
