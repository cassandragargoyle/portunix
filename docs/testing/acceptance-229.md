# Acceptance Protocol - Issue #229 (Plugin Caller Working Directory)

**Issue**: Plugin dispatcher does not pass the caller's working directory to plugins
**Branch**: `fix/229-plugin-caller-working-directory` (commit `515c482`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-10-10
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-34, amd64 (host) — all scenarios; Go 1.26.0, Temurin JDK 21,
  Python 3, bash

No software was installed during testing, so no container was needed. Stub plugins ran
against an isolated `HOME` with its own plugin registry; the real registry was only read.

## Scope Note

The branch makes the CLI plugin dispatcher (`portunix <plugin> ...`) export
`PORTUNIX_CWD` — the absolute working directory of the caller — into the plugin process
environment. The plugin working directory (install path) and all other environment
variables stay unchanged. Long-running plugin services (`grpc_client.go`) are
intentionally untouched; the decision is documented in
`docs/plugin-development/architecture.md`.

**Not covered**: Windows. The code and the test were only compiled for Windows
(`GOOS=windows go vet`, `GOOS=windows go test -c`), not executed. Acceptance criterion 5
(test passes on Windows) and the manual `cabinet` verification on Windows are open — see
Final Decision.

## Test Summary

- Total test scenarios: 38
- Passed: 36
- Failed: 0
- Skipped: 2 (both Windows executions)
- Open defects in the scope of #229: 0
- Observations: 3 (O1–O3)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `gofmt -s`, `go vet` (linux, windows), test compile (windows), `go build` (darwin), `make build` | host |
| Unit | `pluginEnvironment`, `buildPluginCommand` for all runtimes, launched stub plugin (`go test -race`) | host |
| Mutation | remove the `cmd.Env` assignment and confirm the new tests fail | host |
| Baseline | binary built from `main` (`fb530b6`): variable absent, `cabinet` relative path fails | host |
| E2E | real `portunix` binary + stub plugins for native, Python script, Python wrapper, Java | host, isolated `HOME` |
| Failure injection | stale inherited value, spaces and non-ASCII in path, symlinked cwd, deleted cwd, exit code, other variables, stdin | host, isolated `HOME` |
| Regression | installed plugins that ignore the variable: output and exit code `main` vs branch | host |
| Acceptance | `cabinet pdf prepare` per criteria 1–3 | host |

Each stub plugin prints only the value of `PORTUNIX_CWD`, its own working directory, and
its arguments. The environment was never dumped (issue caution about
`PORTUNIX_PLUGINS_GITHUB_TOKEN`).

## Test Cases

### TC01: Plugin receives the caller directory

- **Given** an enabled plugin and a shell standing in directory `D` different from the
  plugin install path
- **When** `portunix <plugin> a b` is run
- **Then** the plugin sees `PORTUNIX_CWD=D`, its working directory is the install path,
  and its arguments are `a b`

### TC02: Stale inherited value is replaced

- **Given** `PORTUNIX_CWD=/stale` already present in the environment of `portunix`
- **When** a plugin is dispatched from directory `D`
- **Then** the plugin sees exactly one `PORTUNIX_CWD` and its value is `D`

### TC03: Caller directory cannot be determined

- **Given** a shell standing in a directory that has been deleted
- **When** a plugin is dispatched
- **Then** the plugin starts, `PORTUNIX_CWD` is unset, exit code is unchanged

### TC04: Relative input path (criterion 1)

- **Given** `book.pdf` in the caller's directory
- **When** `portunix cabinet pdf prepare --dry-run book.pdf` is run there
- **Then** the command succeeds

### TC05: Relative output path (criterion 2)

- **Given** `book.pdf` in the caller's directory
- **When** `portunix cabinet pdf prepare -o out.pdf book.pdf` is run there
- **Then** `out.pdf` is created in the caller's directory and nothing is written into the
  plugin directory

### TC06: Absolute paths (criterion 3)

- **Given** an absolute path to `book.pdf`
- **When** the command is run from an unrelated directory
- **Then** the result is the same as on `main`

### TC07: Plugin ignoring the variable (criterion 4)

- **Given** a plugin that does not read `PORTUNIX_CWD`
- **When** the same command is run with the `main` binary and the branch binary
- **Then** output and exit code are identical

## Test Results

### Static Checks and Unit Tests

| Check | Result |
| ----- | ------ |
| `gofmt -s -l` on changed files | PASS (no output) |
| `go vet portunix.ai/cmd` | PASS |
| `GOOS=windows go vet portunix.ai/cmd` | PASS |
| `GOOS=windows go test -c portunix.ai/cmd` (compile only) | PASS |
| `GOOS=darwin go build .` | PASS |
| `make build` (main binary + all helpers) | PASS |
| `TestPluginEnvironment` (4 subtests, `-race`) | PASS |
| `TestBuildPluginCommand_AllRuntimes` (5 subtests, `-race`) | PASS |
| `TestBuildPluginCommand_StubPluginReceivesCallerDir` (`-race`) | PASS |
| Mutation: without `cmd.Env = pluginEnvironment(...)` both `TestBuildPluginCommand*` tests fail | PASS (tests detect the defect) |

### Baseline on `main`

| Scenario | Result |
| -------- | ------ |
| Stub plugins (4 runtimes) on `main`: `PORTUNIX_CWD` unset | PASS (defect reproduced) |
| `cabinet pdf prepare --dry-run book.pdf` on `main`: `input file not found: ~/.portunix/plugins/cabinet/book.pdf` | PASS (defect reproduced) |

### E2E — All Runtimes (TC01)

Run from `<T>/caller`; `<T>` is the temporary test root.

| Runtime | `PORTUNIX_CWD` | Working directory | Arguments | Result |
| ------- | -------------- | ----------------- | --------- | ------ |
| native | `<T>/caller` | install path | `a b` | PASS |
| Python script (`.py`) | `<T>/caller` | install path | `a b` | PASS |
| Python wrapper (`.sh`) | `<T>/caller` | install path | `a b` | PASS |
| Java (`.jar`) | `<T>/caller` | install path | `a b` | PASS |

The Python wheel runtime is covered by the unit test and by the `cabinet` plugin below.

### Failure Injection

| ID | Scenario | Observed | Result |
| -- | -------- | -------- | ------ |
| F1 | `PORTUNIX_CWD=/stale` inherited (TC02) | branch: `<T>/caller`; exactly one occurrence | PASS |
| F2 | caller directory `dir with space/žluťoučký` (native, Java, Python) | value passed intact | PASS |
| F3 | caller directory reached through a symlink | logical path `<T>/link` (see O2) | PASS |
| F4 | caller directory deleted (TC03) | plugin runs, variable unset, exit code 0 — same as `main` | PASS |
| F5 | plugin exits with code 7 | `portunix` exits with 7 on branch and on `main` | PASS |
| F6 | unrelated variable with spaces and `=` in the value | passed unchanged | PASS |
| F7 | data piped to stdin | reaches the plugin | PASS |

### Regression — Plugins Ignoring the Variable (TC07)

| Command | `main` vs branch | Result |
| ------- | ---------------- | ------ |
| `query --help` (Java) | identical output, exit code 0 | PASS |
| `code-info --help` | identical output, exit code 2 | PASS |
| `text-extractor --help` | identical output, exit code 0 | PASS |
| `scraper --help` | identical output, exit code 0 | PASS |
| `cabinet --help` | identical output, exit code 0 | PASS |
| `cabinet pdf prepare --dry-run <absolute path>` | identical output | PASS |

### Acceptance — `cabinet` Plugin (TC04–TC06)

Run from `<T>/my books` (path with a space), one-page PDF.

| ID | Command | Observed | Result |
| -- | ------- | -------- | ------ |
| A1 | `prepare --dry-run book.pdf` | `1 pages: 0 whitened, 1 skipped`, exit code 0 | PASS |
| A2 | `prepare --dry-run sub/inner.pdf` | success | PASS |
| A3 | from `sub/`: `prepare --dry-run ../book.pdf` | success | PASS |
| A4 | `prepare -o out.pdf book.pdf` | `<T>/my books/out.pdf` created, valid PDF header | PASS |
| A5 | `prepare -o sub/o2.pdf book.pdf` | `<T>/my books/sub/o2.pdf` created | PASS |
| A6 | from `/`: `prepare --dry-run <absolute path>` | success, same as `main` | PASS |
| A7 | `prepare -o <absolute path> <absolute path>` | file created at the absolute path | PASS |

The listing of the `cabinet` plugin directory was identical before and after the run.

### Skipped

| Scenario | Reason |
| -------- | ------ |
| `go test portunix.ai/cmd -run 'TestPluginEnvironment\|TestBuildPluginCommand'` on Windows | no Windows environment in this session |
| Manual `cabinet` verification (criteria 1–3) on Windows | no Windows environment in this session |

## Acceptance Criteria

| # | Criterion | Status |
| - | --------- | ------ |
| 1 | Relative input path works from an arbitrary directory | PASS (Linux) |
| 2 | `-o out.pdf` writes into the caller's directory | PASS (Linux) |
| 3 | Absolute paths work as before | PASS (Linux) |
| 4 | Plugin not reading the variable behaves as before | PASS (Linux) |
| 5 | Test passes on Linux and on Windows | Linux PASS, Windows NOT EXECUTED |
| 6 | `PORTUNIX_CWD` documented, including the service / MCP decision | PASS |

## Coverage

- Changed code: `buildPluginCommand` and `pluginEnvironment` are covered by unit tests for
  every runtime branch; the Windows-only case-insensitive key comparison is not exercised
  on Linux
- `executePlugin` (registry lookup, `os.Getwd`, exit code propagation) is covered by the
  E2E runs with the real binary, not by an automated test

## CI Notes

- The package must be tested by import path from the repository root
  (`go test portunix.ai/cmd/...`); `go test ./src/cmd/` does not resolve the satellite module
- The stub-plugin test needs `bash` on Linux and `powershell.exe` on Windows
- `go test portunix.ai/cmd/...` as a whole still reports FAIL because of O1

## Observations

- **O1 (out of scope)**: six `TestContainer*` tests in `portunix.ai/cmd` fail
  (`TestContainerExecCommand`, `TestContainerCommandErrorMessages`,
  `TestContainerRunCommand_ErrorHandling`, `_ArgumentParsing`, `_FlagValidation`,
  `_OriginalIssueRepro`). The same six fail on `main` (`fb530b6`); tracked as #225
- **O2**: when the caller stands in a symlinked directory, `PORTUNIX_CWD` holds the logical
  path (as shown by the shell), not the resolved one. This matches what the user sees and
  needs no change
- **O3**: when the caller's directory cannot be determined the variable is left unset and
  the plugin falls back to its own directory — the behaviour of `main`. Documented in the
  plugin developer documentation

## Recommendations

1. Run on Windows before or right after merge:
   `go test portunix.ai/cmd -run "TestPluginEnvironment|TestBuildPluginCommand" -v`
   and the `cabinet` checks A1, A4, A6
2. Consider an automated E2E test of `executePlugin` with an isolated `HOME`, as used
   manually in this protocol

## Final Decision

**STATUS**: CONDITIONAL

Linux: all executed scenarios passed, no defects found. The only open item is the
execution on Windows required by acceptance criterion 5.

**Approval for merge**: YES for the Linux scope; full acceptance requires either a
passing Windows run (Recommendation 1) or an explicit decision of the Product Owner to
merge with Windows verification outstanding
**Date**: 2026-10-10
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
