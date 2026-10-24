# Acceptance Protocol - Issue #233 (Python Plugin Venv Health and Atomic Force Reinstall)

**Issue**: Python plugins break silently after a Python upgrade, and force reinstall is not atomic
**Branch**: `fix/233-python-plugin-venv-stale-interpreter` (commit `97fe6a3`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-22
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, unit tests, `go vet`, regression of
  `portunix.ai/cmd`, read-only checks against the installed plugins; Go 1.26.8, Podman 5.7.0
- `ubuntu:24.04` container (Python 3.12.3) started through `portunix container` — every
  scenario that creates a venv, installs uv or removes an interpreter

No software was installed on the host. The test wheel was built in a session-local scratch
directory, the container was removed after the run.

## Scope Note

The branch adds an install-state check (`ok` / `missing` / `broken`) used by `plugin list`,
`info`, `health`, `check` and the plugin dispatcher, a new `portunix plugin repair <name>`
command, an atomic `plugin install --force` (staging, pre-check, backup, restore), venv
creation with uv-managed Python (`--managed-python`), detection of the wheel's Python version
and an `ensurepip` pre-check for the system Python fallback.

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Windows — `Scripts\python.exe` and the `pyvenv.cfg` `home` check were not executed
- A real OS upgrade — simulated by moving `/usr/bin/python3.12` away
- Fallback when `uv venv --managed-python` fails (offline, uv older than 0.8) — not executed
  and not covered by a unit test
- Version detection from a `cpXY` tag or `.pyc` magic end to end — unit tests only

## Test Summary

- Total test scenarios: 12
- Passed: 12
- Failed: 0
- Skipped: 4 (Windows, real OS upgrade, `--managed-python` fallback, E2E version detection)
- Defects found in the scope of #233: 0
- Defects found outside the scope: 1 (D1)
- Observations: 3 (O1–O3)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | New tests in `src/app/plugins/manager` (install state, rollback, repair, wheel version) | host |
| Static | `go vet`, `gofmt` of `src/app/plugins` and `src/cmd` | host |
| E2E | Install, break, check, repair and force install of a wheel plugin, system Python | `ubuntu:24.04` |
| E2E | Repair and force install with uv installed through `portunix install uv` | `ubuntu:24.04` |
| Failure injection | Removed interpreter, missing `ensurepip`, no Python at all, missing install directory | `ubuntu:24.04` |
| Regression | `portunix.ai/cmd` and `portunix.ai/app/plugins/...` compared against `main` | host |
| Read-only | `plugin list`, `check`, `health`, dispatcher against the real installed plugins | host |

## Test Cases

The test plugin `demo` is a helper wheel plugin (`ptx_demo-1.0.0-py3-none-any.whl`, console
script `ptx-demo`) without `python_version` in the manifest — the case from the issue.

### TC01: Install with the system Python

- **Given** `python3` and `python3-venv`, no uv
- **When** `portunix plugin install /root/demo-plugin`
- **Then** the venv uses `/usr/bin/python3.12`, `plugin list` shows `ready`, `portunix demo` runs

### TC02: Removed base interpreter is reported

- **Given** the installed plugin from TC01
- **When** `/usr/bin/python3.12` is removed (simulated OS upgrade)
- **Then** `list` and `info` show `broken`, `check` reports `Installation: broken` with
  `Fix: portunix plugin repair demo` and exits 1, `health` is unhealthy, the dispatcher refuses
  to run the plugin and prints the fix

### TC03: Repair without any usable Python

- **Given** the broken plugin, no Python and no uv
- **When** `portunix plugin repair demo`
- **Then** it fails with a hint for uv and Python, says that the installation was left
  unchanged, the registry entry stays, no staging or backup directory remains

### TC04: Failed force install keeps the old installation

- **Given** a working installation and `python3` without `ensurepip` (`python3.12-venv` removed;
  `python3 -m venv --help` still succeeds), no uv
- **When** `portunix plugin install --force /root/demo-plugin`
- **Then** it fails on the `ensurepip` pre-check and suggests `portunix install uv`, the old
  files (including a marker file) and the registry entry (`enabled: true`) are untouched, the
  plugin still runs, no staging or backup directory remains

### TC05: Repair with uv

- **Given** uv installed by `portunix install uv` and a broken venv
- **When** `portunix plugin repair demo`
- **Then** the venv is recreated with a uv-managed interpreter
  (`home = /root/.local/share/uv/python/cpython-3.15-…`), status is `ready`, the plugin runs,
  `enabled` is kept

### TC06: Repaired venv does not depend on the system Python

- **Given** the repaired plugin from TC05 and `/usr/bin/python3.12` still removed
- **When** the plugin is run, then again after the system Python is restored
- **Then** both runs succeed

### TC07: Successful force install with uv

- **When** `portunix plugin install --force /root/demo-plugin`
- **Then** it succeeds, `enabled: true` is kept, no leftovers, the plugin runs

### TC08: Missing install directory

- **Given** the plugin directory moved away
- **When** `list`, `check`, `repair` and the dispatcher are used
- **Then** status is `failed` (never `ready`), `check` reports `Installation: missing` with a
  fix, `repair` fails with a hint to reinstall from source, the dispatcher refuses to run;
  after the directory is moved back the status is `ready` again

### TC09: Unit tests

- **When** `go test -count=1 portunix.ai/app/plugins/...`
- **Then** all tests pass

### TC10: Static checks

- **When** `go vet` and `gofmt -l` on `src/app/plugins` and `src/cmd`
- **Then** no findings

### TC11: Regression against `main`

- **When** `go test -count=1 -v portunix.ai/cmd` on the branch and on `main`
- **Then** the same set of tests passes and fails

### TC12: Real installed plugins on the host (read-only)

- **When** `portunix plugin list`, `check vox`, `health vox` and `portunix vox --help`
- **Then** exactly the four plugins bound to the removed `/usr/bin/python3.13` (graphlens,
  reco-watcher, reco-worker, vox) are `broken`, all others keep their previous status

## Test Results

### Unit Tests and Static Checks

| Check | Result |
| ----- | ------ |
| `go test -count=1 portunix.ai/app/plugins/...` | ✅ `manager` and `server` ok, 52 test functions and subtests passed |
| `go vet portunix.ai/app/plugins/... portunix.ai/cmd` | ✅ no findings |
| `gofmt -l src/app/plugins src/cmd` | ✅ clean |
| `make build` | ✅ main binary and all helpers |

### E2E — `ubuntu:24.04` Container

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `home = /usr/bin`, `ready`, `ptx-demo OK python 3.12` |
| TC02 | ✅ PASS | `broken` in list and info; `venv interpreter is missing: …/.venv/bin/python -> python3`; check exit 1; dispatcher exit 1 with fix |
| TC03 | ✅ PASS | `Python not found. Install uv (portunix install uv) or Python … (existing installation left unchanged)`; plugins dir holds only `demo` and `registry.json` |
| TC04 | ✅ PASS | `Python ensurepip module not available … Install uv (portunix install uv) … (existing installation left unchanged)`; `old-marker` present, plugin runs, registry `install_path` and `enabled: True` unchanged |
| TC05 | ✅ PASS | `Creating Python virtual environment with uv (Python latest)`; uv-managed `home`; `ready`; `ptx-demo OK python 3.15` |
| TC06 | ✅ PASS | runs with the system Python removed and after it is restored |
| TC07 | ✅ PASS | `Plugin reinstalled successfully`, `enabled: true`, no leftovers |
| TC08 | ✅ PASS | `failed`; `Installation: missing`; repair hint `reinstall from source with: portunix plugin install --force <plugin-path>`; back to `ready` |

### Regression — Against `main`

| Package | Branch | `main` |
| ------- | ------ | ------ |
| `portunix.ai/cmd` | 34 top-level tests, 6 fail | 34 top-level tests, 6 fail — the same tests |
| `portunix.ai/app/plugins/...` | pass | — |

The six failures in `portunix.ai/cmd` (among them `TestContainerRunCommand_OriginalIssueRepro`)
exist on `main` and are part of #225.

### Read-only Check on the Host (TC12)

✅ PASS — graphlens, reco-watcher, reco-worker and vox show `broken`; `plugin check vox`
reports `venv interpreter is missing: …/vox/.venv/bin/python -> /usr/bin/python3.13` with
`Fix: portunix plugin repair vox`. `repair` was not run on the host.

### Skipped

- Windows execution
- Real OS upgrade
- `--managed-python` fallback path
- Version detection from a wheel tag or `.pyc` magic end to end

## Defects

### D1: `container run` example with `--` fails — outside #233, tracked as #235

`portunix container run --help` shows `portunix container run -d --name test ubuntu:22.04 -- bash -c "echo test"`.
The `--` is passed to the runtime as the command:
`exec: "--": executable file not found in $PATH`. Without `--` the command works.

## Acceptance Criteria

| Criterion | Result |
| --------- | ------ |
| `plugin list` does not show `ready` for a plugin whose venv interpreter is missing | ✅ TC02, TC12 |
| A broken venv can be repaired with one Portunix command, without the plugin source | ✅ TC05 (`plugin repair`) |
| A failed `install --force` leaves the previous installation working and registered | ✅ TC04, TC03; unit test for a failure after the swap |
| The registry never points to a plugin directory that does not exist | ✅ TC08 — the entry is kept but reported as `failed` / `missing`, never `ready`; a failed install does not remove the old directory |
| Unit tests cover the broken-venv detection and the rollback of a failed force install | ✅ TC09 |

## Coverage

- **Install state**: all three states and the status mapping are covered by unit tests and E2E
- **Force install**: pre-check failure, failure after the swap (unit), success with system
  Python (unit, native plugin) and with uv (E2E)
- **Repair**: success with uv (E2E) and with a fake Python (unit), no Python, no installed
  manifest
- **Gaps**: `--managed-python` fallback, Windows paths, rename failure of the backup or staging
  directory, failure of the registry write

## CI Notes

- The new tests that use a fake `python3` shell script skip on Windows
- No new test needs network access; the E2E scenarios with uv download a managed CPython

## Observations

- **O1**: `repair` and `install --force` print the prerequisite error
  `Python runtime not found … Plugin will be installed but may not work correctly` even when uv
  is about to provide the interpreter (TC03, TC05); `plugin check` also reports the runtime as
  an error for a uv-backed venv when the system Python is missing
- **O2**: without `python_version` in the manifest uv picks the newest managed CPython (3.15 in
  TC05); plugins with native dependencies may not have wheels for it yet
- **O3**: runtime errors of `plugin repair`, `plugin install` and the dispatcher are followed by
  the full cobra usage text, which pushes the fix hint up (also the case on `main`)

## Recommendations

1. Run TC01–TC08 once on Windows (tester-windows role)
2. Add a unit test for the `--managed-python` fallback (fake `uv` failing the first call)
3. Skip or soften the system-Python runtime check when uv is available (O1, tracked as #234)
4. Pin `python_version` in the manifests of Python plugins (O2, tracked as #234)
5. Set `SilenceUsage` on the plugin commands for runtime errors (O3, tracked as #234)
6. Fix D1 (#235)
7. After the merge, run `portunix plugin repair` for graphlens, reco-watcher, reco-worker and vox

## Final Decision

**STATUS**: CONDITIONAL

Linux: all executed scenarios passed, every acceptance criterion is met, no regression against
`main`. The open item is execution on Windows.

**Approval for merge**: YES for the Linux scope; Windows to be confirmed by the tester-windows role
**Date**: 2026-10-22
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
