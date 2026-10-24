# Acceptance Protocol - Issue #234 (Plugin Install/Repair UX Follow-ups)

**Issue**: Plugin install/repair: false Python runtime error with uv, unpinned uv interpreter, usage text after errors
**Branch**: `fix/234-plugin-install-repair-ux-followups` (tested as the working tree on top of `1bb8c30`, committed unchanged as `901e300`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, unit tests, `go vet`, `gofmt`,
  regression of `portunix.ai/cmd` against `main`, read-only check of the installed plugins
- `ubuntu:24.04` container started through `portunix container` — every scenario that installs
  uv or Python, creates a venv or removes an interpreter

No software was installed on the host. The test wheel was built in a session-local scratch
directory, the container was removed after the run.

## Scope Note

The branch changes three things in the plugin manager:

1. For wheel plugins the runtime check of `install`, `install --force` and `repair` asks uv
   (`uv python find --managed-python`) instead of the host `python3` when uv is available;
   `plugin check` checks the plugin venv interpreter
2. A wheel plugin without `python_version` gets the newest CPython minor released at least six
   months ago (3.14 on the test date), overridable with `PORTUNIX_PLUGIN_PYTHON_VERSION`, and a
   pin recommendation is printed
3. Runtime errors of the `plugin` command tree and of the plugin dispatcher are printed without
   the cobra usage text; argument errors keep it

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Windows — `Scripts\python.exe` of the venv check and uv on Windows were not executed
- `plugin install-github` — shares the install path, not executed end to end
- Offline host, where `uv python find` fails and uv cannot download — not executed

## Test Summary

- Total test scenarios: 14
- Passed: 14
- Failed: 0
- Skipped: 3 (Windows, `install-github`, offline)
- Defects found: 0
- Observations: 4 (O1–O4)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | New tests in `src/app/plugins`, `src/app/plugins/manager` and `src/cmd` | host |
| Static | `go vet`, `gofmt` of `src/app/plugins` and `src/cmd` | host |
| E2E | Install, force install, repair and check of a wheel plugin with uv and no system Python | `ubuntu:24.04` |
| E2E | Interpreter choice: unpinned, env override, `python_version`, `python_min_version` | `ubuntu:24.04` |
| E2E | System Python and uv together; system Python fallback without uv | `ubuntu:24.04` |
| Failure injection | No Python and no uv; removed uv-managed interpreter | `ubuntu:24.04` |
| CLI | Runtime errors vs. argument errors (usage text) | `ubuntu:24.04` |
| Regression | `portunix.ai/cmd` on the branch and on `main` (separate git worktree) | host |
| Read-only | `plugin check --all` against the real installed plugins | host |

## Test Cases

The test plugin `demo` is a helper wheel plugin (`ptx_demo-1.0.0-py3-none-any.whl`, console
script `ptx-demo` printing the interpreter version) in three manifest variants: no Python
field (unpinned), `"python_version": "3.13"` (pinned) and `"python_min_version": "3.15"`.

### TC01: No uv and no Python

- **Given** a container without Python and without uv
- **When** `portunix plugin install /root/demo-unpinned`
- **Then** the runtime is still reported as an error (correct — nothing can provide it), the
  install fails with the hint to install uv or Python, no usage text is printed, no plugin
  directory remains

### TC02: Install with uv, no system Python, no managed Python yet (AC1, AC3)

- **Given** uv installed with `portunix install uv`, no system Python
- **When** `portunix plugin install /root/demo-unpinned`
- **Then** the runtime is a warning (`uv will download Python 3.14 during setup`), not an error;
  the chosen version 3.14 and the pin recommendation are printed; the venv uses the uv-managed
  CPython 3.14; the plugin runs

### TC03: `plugin check` of a healthy uv-backed plugin (AC2)

- **When** `portunix plugin check demo` and `--json`
- **Then** exit code 0, `Runtime: ✅ Found: 3.14.9 (plugin venv)`, JSON `status: ok`

### TC04: `install --force` and `repair` with uv (AC1)

- **When** `portunix plugin install --force /root/demo-unpinned`, then `portunix plugin repair demo`
- **Then** no prerequisite problem is printed (uv already has 3.14), the pin recommendation is
  printed, both succeed, `plugin list` shows `ready`

### TC05: Environment override

- **When** `PORTUNIX_PLUGIN_PYTHON_VERSION=3.12 portunix plugin repair demo`
- **Then** the venv uses Python 3.12 and the warning names 3.12

### TC06: Pinned `python_version`

- **When** `portunix plugin install --force /root/demo-pinned`
- **Then** Python 3.13 is used and no pin recommendation is printed

### TC07: `python_min_version` above the default

- **When** `portunix plugin install --force /root/demo-minver` (`>=3.15`, default is 3.14)
- **Then** the request is `>=3.15`, Python 3.15 is used, no pin recommendation; `check` exits 0

### TC08: Removed uv-managed interpreter (AC4)

- **Given** the plugin from TC07 and the `cpython-3.15*` directories moved away
- **When** `portunix demo`, `portunix plugin run demo`, `portunix plugin check demo`
- **Then** dispatcher and `plugin run` print the error and `Fix: portunix plugin repair demo`
  without usage text and exit 1; `check` reports runtime and installation as errors with the fix
  and exits 1; `repair` recreates the venv and the plugin runs again

### TC09: Runtime errors vs. argument errors (AC4)

- **When** `plugin repair nope`, `plugin install /root/nope` (runtime errors) and `plugin repair`,
  `plugin install`, `plugin check` without argument, `plugin check a b` (argument errors)
- **Then** runtime errors print only the error; argument errors print the error and the usage

### TC10: System Python and uv together

- **Given** `python3` 3.12.3 installed from apt, uv present
- **When** `portunix plugin install --force /root/demo-unpinned`
- **Then** the uv-managed 3.14 is used (not the host 3.12), the pin recommendation is printed,
  `check` reports `3.14.9 (plugin venv)`

### TC11: System Python fallback without uv

- **Given** uv hidden from `PATH`
- **When** `portunix plugin repair demo`, then `check`
- **Then** the venv uses `/usr/bin` Python 3.12.3, `check` reports `3.12.3 (plugin venv)`

### TC12: Unit tests (AC5)

- **When** `go test -count=1 portunix.ai/app/plugins/...` and `go test -count=1 -v portunix.ai/cmd`
- **Then** all plugin tests pass, including the 11 new ones

### TC13: Static checks and regression

- **When** `go vet`, `gofmt -l`, and `portunix.ai/cmd` on the branch and on `main`
- **Then** no findings; the same set of tests fails on both

### TC14: Real installed plugins on the host (read-only)

- **When** `portunix plugin check --all`
- **Then** wheel plugins report the venv interpreter, no false runtime error

## Test Results

### Unit Tests and Static Checks

| Check | Result |
| ----- | ------ |
| `go test -count=1 portunix.ai/app/plugins/...` | ✅ `plugins`, `manager`, `server` ok |
| New tests | ✅ `TestRecomputeOverall`, `TestCheckVenvPythonRuntime`, `TestCheckUVPythonRuntime`, `TestSatisfiesVersionConstraint`, `TestMatureCPythonMinor`, `TestUVPythonRequest`, `TestDefaultPluginPythonVersionWithoutOverride`, `TestApplyInstalledRuntimeCheckUsesVenv`, `TestManifestPrerequisitesWithUV`, `TestPluginCommandSilencesUsage`, `TestPluginCheckArgs` |
| `go vet portunix.ai/app/plugins/... portunix.ai/cmd` | ✅ no findings |
| `gofmt -l src/app/plugins src/cmd` | ✅ clean |
| `make build` | ✅ main binary and all helpers |

### E2E — `ubuntu:24.04` Container

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `❌ Python runtime not found`; `Error: … Python not found. Install uv (portunix install uv) or Python (portunix install python)`; exit 1; no usage; plugins dir holds only `registry.json` |
| TC02 | ✅ PASS | `⚠️ Found: uv uv will download Python 3.14 during setup`; `Plugin demo does not set plugin.python_version; using Python 3.14 …`; `home = …/cpython-3.14-linux-x86_64-gnu/bin`; `ptx-demo OK python 3.14.9` |
| TC03 | ✅ PASS | `✅ Found: 3.14.9 (plugin venv)`, `Installation: ok`; exit 0; JSON `"status": "ok"` |
| TC04 | ✅ PASS | no prerequisite table; `Plugin reinstalled successfully!`, `Plugin repaired successfully!`; `ready` |
| TC05 | ✅ PASS | `using Python 3.12`; `ptx-demo OK python 3.12.15` |
| TC06 | ✅ PASS | `Creating … with uv (Python 3.13)`, no pin warning; `ptx-demo OK python 3.13.17` |
| TC07 | ✅ PASS | `(Python >=3.15)`, no pin warning; `ptx-demo OK python 3.15.1`; check exit 0 |
| TC08 | ✅ PASS | `Error: plugin 'demo' is broken: venv interpreter is missing: … -> …/cpython-3.15-…/bin/python3.15` + `Fix: portunix plugin repair demo`, no usage, exit 1 (dispatcher and `plugin run`); check: `❌ venv interpreter not usable` and `Installation: broken`, exit 1; repair → `ptx-demo OK python 3.15.1` |
| TC09 | ✅ PASS | `Error: failed to repair plugin: plugin nope not found` and `Error: plugin.json not found in /root/nope` alone; `accepts 1 arg(s), received 0`, `plugin name required …`, `accepts at most 1 arg(s), received 2` each followed by `Usage:` |
| TC10 | ✅ PASS | host `Python 3.12.3`; `home = …/cpython-3.14-linux-x86_64-gnu/bin`; `ptx-demo OK python 3.14.9`; check `3.14.9 (plugin venv)` |
| TC11 | ✅ PASS | `Creating Python virtual environment...` (system path), `home = /usr/bin`; `ptx-demo OK python 3.12.3`; check `3.12.3 (plugin venv)` |

### Regression — Against `main`

| Package | Branch | `main` |
| ------- | ------ | ------ |
| `portunix.ai/cmd` | 37 top-level tests, 6 fail | 35 top-level tests, 6 fail — the same tests |
| `portunix.ai/app/plugins/...` | pass | pass |

The two additional tests on the branch are the new `TestPluginCommandSilencesUsage` and
`TestPluginCheckArgs`. The six failures (`TestContainer*`) exist on `main` and are part of #225.

### Read-only Check on the Host (TC14)

✅ PASS — `Summary: 23 plugins checked, 0 issues found, 1 warnings`; seven wheel plugins report
`Found: <version> (plugin venv)`. Exit code 2 comes from a missing optional tool (`exiftool`),
as before the change.

### Skipped

- Windows execution
- `plugin install-github` end to end
- Offline host

## Acceptance Criteria

| Criterion | Result |
| --------- | ------ |
| `install`, `install --force` and `repair` of a wheel plugin print no runtime error when uv is available and no system Python is installed | ✅ TC02, TC04, TC05–TC07 |
| `plugin check <name>` exits 0 for a healthy uv-backed wheel plugin on a host without system Python, and reports the venv interpreter version | ✅ TC03, TC07 |
| A wheel plugin without `python_version` does not silently get the newest CPython; the chosen version and a recommendation to pin it are printed | ✅ TC02, TC04, TC10 (3.14 while 3.15 is available, TC07) |
| Runtime errors of `plugin repair`, `plugin install` and `portunix <plugin> …` print the error and the fix hint without the usage text; argument errors still show usage | ✅ TC01, TC08, TC09 |
| Unit tests cover the runtime check for wheel plugins and the interpreter choice | ✅ TC12 |

## Coverage

- **Runtime check**: uv with and without the managed interpreter, venv healthy, venv broken,
  no Python at all — unit and E2E; version constraint against the venv — unit
- **Interpreter choice**: default, env override, `python_version`, `python_min_version` below and
  above the default — unit and E2E; date boundaries of the default — unit
- **Usage silencing**: `plugin` subcommands and dispatcher — unit and E2E
- **Gaps**: Windows, offline uv, `install-github`

## CI Notes

- The new tests that use shell-script interpreters or a fake `uv` skip on Windows
- No new test needs network access; the E2E scenarios download uv-managed CPython builds
- `TestMatureCPythonMinor` uses fixed dates; `TestDefaultPluginPythonVersionWithoutOverride`
  compares against the same function, so neither depends on the current date

## Observations

- **O1**: the uv warning reads `⚠️ Found: uv uv will download Python 3.14 during setup` — the
  word `uv` is printed twice because `Found` is `uv` and the message also starts with it
  (cosmetic)
- **O2**: the pin recommendation follows the prerequisite table without an empty line
  (cosmetic)
- **O3**: a runtime warning now raises the overall result to `warning` (before it was ignored),
  so the prerequisite table is printed during install and `plugin check` exits 2 for e.g.
  `Could not parse Python version`; intended by the change, worth a CHANGELOG line
- **O4**: without uv the venv uses the host `python3` and no pin recommendation is printed; this
  is documented in the plugin development guide

## Recommendations

1. Fix O1 — e.g. leave `Found` empty for the download case or reword the message
2. ~~Commit the branch before merge~~ — done, `901e300`
3. Run TC02–TC08 once on Windows (tester-windows role)
4. Mention O3 in the CHANGELOG
5. Pin `python_version` in the manifests of published Python plugins

## Final Decision

**STATUS**: PASS (Linux)

All executed scenarios passed, every acceptance criterion is met, no regression against `main`.
Observations O1 and O2 are cosmetic and do not block the merge.

**Approval for merge**: YES for the Linux scope; Windows to be confirmed by the tester-windows role
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
