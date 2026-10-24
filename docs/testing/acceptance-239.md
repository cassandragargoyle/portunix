# Acceptance Protocol - Issue #239 (Duplicated "uv" in the uv Runtime Warning)

**Issue**: Plugin prerequisites: duplicated "uv" in the uv runtime warning
**Branch**: `fix/239-uv-warning-duplicated-word` (commit `163bd18`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, unit tests, `go vet`, `gofmt`, a mutation
  check of the new test, a JSON check of `CheckUVPythonRuntime` against a fake `uv` script
- `ubuntu:24.04` container started through `portunix container` — every scenario that installs
  uv or Python or creates a venv

No software was installed on the host. The test wheel was built in a session-local scratch
directory, and the container was removed after the run.

## Scope Note

The branch changes two things:

1. `CheckUVPythonRuntime` leaves `Found` empty when uv has to download Python, so the warning
   names uv once
2. `install`, `install --force` and `repair` print an empty line after the prerequisite table
   when it shows only warnings (the error branch already had one)

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Windows — the change is platform-independent output formatting, not executed
- `plugin install-github` — shares `InstallPlugin`, not executed end to end

## Test Summary

- Total test scenarios: 9
- Passed: 9
- Failed: 0
- Skipped: 2 (Windows, `install-github`)
- Defects found: 0

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `TestCheckUVPythonRuntime` and the rest of `portunix.ai/app/plugins/...` | host |
| Unit | Mutation check: restore `Found: "uv"`, the test must fail | host |
| API | JSON of the download-case runtime check | host, fake `uv` |
| Static | `go vet`, `gofmt` of `src/app/plugins` | host |
| E2E | `install`, `install --force`, `repair` with uv lacking the requested interpreter | `ubuntu:24.04` |
| Regression | Output without warnings, `plugin check`, the error branch without uv and Python | `ubuntu:24.04` |

## Test Cases

The test plugin `demo` is a helper wheel plugin (`ptx_demo-1.0.0-py3-none-any.whl`, console
script `ptx-demo` printing the interpreter version) in two manifest variants: no Python field
(unpinned) and `"python_version": "3.13"` (pinned). uv 0.13.5 was installed in the container
with `portunix install uv`, and no system Python was present.

### TC01: `install` with uv and no managed Python (AC1, AC3)

- **Given** uv without any managed interpreter
- **When** `portunix plugin install /root/demo-unpinned`
- **Then** the runtime line reads `⚠️ uv will download Python 3.14 during setup`, and an empty
  line separates the table from the pin recommendation

### TC02: JSON of the download case (AC2)

- **Given** a fake `uv` that exits 2 on `uv python find`
- **When** the result of `CheckUVPythonRuntime("python", uv, "3.14")` is marshalled to JSON
- **Then** the JSON is valid and has no `found` key

### TC03: `install --force` with a missing pinned interpreter (AC1, AC3)

- **Given** uv holds 3.14 only
- **When** `portunix plugin install --force /root/demo-pinned` (3.13)
- **Then** `⚠️ uv will download Python 3.13 during setup`, empty line, then the setup output

### TC04: `repair` with a missing interpreter (AC1, AC3)

- **Given** the unpinned plugin installed, uv holds 3.14 and 3.13
- **When** `PORTUNIX_PLUGIN_PYTHON_VERSION=3.12 portunix plugin repair demo`
- **Then** `⚠️ uv will download Python 3.12 during setup`, empty line, then the pin
  recommendation; the plugin runs on 3.12

### TC05: No warning, no table

- **When** `portunix plugin install --force /root/demo-unpinned` (uv already holds 3.14)
- **Then** no prerequisite table and no extra empty line; the setup output follows the
  `Force installing …` line

### TC06: `plugin check` of a healthy plugin

- **When** `portunix plugin check demo` and `--json`
- **Then** unchanged: `✅ Found: 3.14.9 (plugin venv)`, exit 0, JSON `found: "3.14.9"`

### TC07: Error branch without uv and without Python

- **Given** uv hidden from `PATH`, no system Python
- **When** `portunix plugin install --force /root/demo-unpinned`
- **Then** unchanged: `❌ Python runtime not found` with the fix, one empty line, the
  prerequisite issue notice, the install error, exit 1

### TC08: Unit tests and the mutation check (AC4)

- **When** `go test -count=1 portunix.ai/app/plugins/...`, then the same with `Found: "uv"`
  temporarily restored in the download case
- **Then** all tests pass on the branch; with the mutation, `TestCheckUVPythonRuntime` fails

### TC09: Static checks

- **When** `go vet portunix.ai/app/plugins/...`, `gofmt -l src/app/plugins`, `make build`
- **Then** no findings, and the build succeeds

## Test Results

### E2E — `ubuntu:24.04` Container

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | Line 4 `Runtime: python ⚠️ uv will download Python 3.14 during setup`; line 7 empty; line 8 `⚠️  Plugin demo does not set plugin.python_version; using Python 3.14 …`; `grep -c "uv uv"` = 0 |
| TC03 | ✅ PASS | `⚠️ uv will download Python 3.13 during setup`; empty line; `Creating Python virtual environment with uv (Python 3.13)...`; `Plugin reinstalled successfully!` |
| TC04 | ✅ PASS | `⚠️ uv will download Python 3.12 during setup`; empty line; `using Python 3.12 …`; `Plugin repaired successfully!`; `ptx-demo OK python 3.12.15` |
| TC05 | ✅ PASS | `Force installing plugin from /root/demo-unpinned...` followed directly by the pin recommendation, no table |
| TC06 | ✅ PASS | `✅ Found: 3.14.9 (plugin venv)`, `Installation: ok`, exit 0; JSON `"found": "3.14.9"`, `"message": "(plugin venv)"` |
| TC07 | ✅ PASS | `❌ Python runtime not found` + `Fix: portunix install python`; one empty line; `⚠️  Prerequisites check found issues …`; `Error: … Python not found. Install uv … or Python …`; exit 1 |

TC04 first ran against the pinned installation from TC03. uv already held 3.13, so no table
was printed, which is correct. The case was repeated with the unpinned plugin as described
above.

### Host

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC02 | ✅ PASS | `{"status": "warning", "required": "python", "message": "uv will download Python 3.14 during setup"}`; valid JSON; `found` absent |
| TC08 | ✅ PASS | `plugins`, `manager`, `server` ok; mutation: `--- FAIL: TestCheckUVPythonRuntime … want empty Found and "uv" once in the message`; file restored with `git checkout` |
| TC09 | ✅ PASS | `go vet` no findings; `gofmt -l` clean; `make build`: main binary and all helpers |

### Skipped

- Windows execution
- `plugin install-github` end to end

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role, Linux) on behalf of zdendaku
