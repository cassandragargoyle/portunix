# Acceptance Protocol - Issue #246 (CI test debt: Windows, Coverage, Performance)

**Issue**: CI: Windows, Coverage Report and Performance jobs fail on pre-existing test debt
**Branch**: `fix/246-ci-test-debt` (commit `ed8207c`; retest 1 on `76a9a7d`; retest 2 on `6b0f2b1`; retest 3 on `67d5fdc`; retest 4 on `79c5dde`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The change touches tests, the CI workflow and adds `Factory.Close()` to `pkg/logging`. No
software was installed and no host configuration was changed, so the tests ran on the host.

## Scope Note

**Independence note**: the code was written by Claude Code in the developer role in the same
session. The scenarios below were designed separately in the tester role, including an ad hoc
file-descriptor leak test that is not part of the branch.

**Not covered locally**: Windows and macOS execution, and the GitHub jobs themselves (AC1–AC3).
Windows cannot be run on this host; the changed packages only vet for `GOOS=windows` and
`GOOS=darwin`. These criteria are verified by the CI run of the release PR #245.

## Test Summary

Initial run (`ed8207c`):

- Total test scenarios: 10
- Passed: 10 (locally)
- Failed: 0
- Pending: AC1–AC3 (GitHub CI)
- Observations: 2 (O1, O2), not blocking

Retest 1 (`76a9a7d`) — see [Retest 1](#retest-1):

- Total test scenarios: 18 (TC01–TC10, R1–R8)
- Passed: 18 (locally)
- Failed: 0
- O1 and O2 resolved, new observation O3 (not blocking)

Retest 2 (`6b0f2b1`) — see [Retest 2](#retest-2):

- Total test scenarios: 27 (TC01–TC10, R1–R16, R17)
- Passed: 27 (locally)
- Failed: 0
- O3 resolved, no new findings

Retest 3 (`67d5fdc`, documentation only) — see [Retest 3](#retest-3):

- Total test scenarios: 6 (D1–D6)
- Passed: 6
- Failed: 0
- Observations: 2 (O4, O5), not blocking

Retest 4 (`79c5dde`, documentation only) — see [Retest 4](#retest-4):

- Total test scenarios: 5 (L1–L5)
- Passed: 5
- Failed: 0
- O4 and O5 resolved, no new findings

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make build`, `gofmt`, `go vet` (`GOOS=windows` / `darwin`), `golangci-lint`, YAML | host |
| Unit | CI commands of Unit Tests, Coverage Report, Performance Tests | host |
| Failure injection | helper binaries removed; fd count around `Factory.Close()` | host |
| CI | all jobs of PR #245, Windows and macOS included | GitHub |

## Test Cases

### TC01: Unit Tests job command

- **Given** binaries built by `make build`
- **When** `go test -tags=unit -race ./...` runs
- **Then** every package passes

Result: **PASS** — exit 0.

### TC02: Coverage Report job command (AC2)

- **When** `go test -tags=unit -race -coverprofile=coverage.out -covermode=atomic ./...` runs
  after `make build`
- **Then** it passes and produces a profile

Result: **PASS** — exit 0, total coverage 11.2 %.

### TC03: Performance Tests job command (AC2, AC6)

- **When** `go test -tags=unit -run='^$' -bench=. -benchmem ./...` runs (`-benchtime=1x`
  locally to save time)
- **Then** only benchmarks run and the output goes to stdout

Result: **PASS** — exit 0, 11 benchmarks, no `=== RUN` of regular tests in the output.

### TC04: Factory.Close releases log files (AC4)

- **Given** a factory with file output that creates two loggers, an MCP logger and a test logger
- **When** `Close()` is called
- **Then** the number of open descriptors in `/proc/self/fd` returns to the value before the
  loggers were created; a second `Close()` and `Close()` on a factory without files return `nil`

Result: **PASS** — 4 files opened (including those of `CreateMCPLogger` / `CreateTestLogger`),
all 4 released; both extra calls return `nil`. Ad hoc test, removed after the run.

### TC05: logging tests that failed on Windows

`TestCreateMCPLogger`, `TestCreateFileWriter`, `TestFactoryWithMultipleOutputs`,
`TestMCPLoggerFileOutput` with `-race` — **PASS**, every one closes its factory.

### TC06: dispatcher tests skip without helpers (AC5)

- **Given** `ptx-mcp` and `ptx-installer` moved out of the repository root
- **When** the root tests and `TestInstallCommand(Short)HelpFlag` run
- **Then** they skip with a clear message, the tests that need no helper still pass

| Test | Result |
| ---- | ------ |
| `TestMainWithoutArguments` | PASS |
| `TestMCPServerCommandRemoved` | PASS |
| `TestMCPServeCommand` | SKIP — `Helper ptx-mcp not built (run make build first)` |
| `TestInstallCommandHelpFlag` | SKIP — `ptx-installer not built (run make build first)` |
| `TestInstallCommandShortHelpFlag` | SKIP — same message |

Result: **PASS**.

### TC07: dispatcher tests run with helpers

After restoring the helpers, `TestMCPServeCommand` and both install help tests run and pass
(not skipped). The temporary `portunix_test` binary is removed after the run.

Result: **PASS**.

### TC08: platform-neutral paths

- `TestNewSyncCache` (`ptx-pft`) expects `filepath.Join(projectDir, ".pft-cache.json")` under
  `t.TempDir()` — **PASS**
- `TestValidateVolumePath_ValidPaths_Success`: all four subtests pass; `validateVolumePath` no
  longer special-cases `/tmp`, the non-existent path lives under the suite temp dir — **PASS**

Result: **PASS**.

### TC09: Windows build assumptions

- `Makefile` adds `$(EXE_EXT)` (`.exe` on `Windows_NT`) to the main and helper binaries, so
  `requireHelper` / `requireInstallBinaries` look for the right file names
- `buildTestBinary` adds `.exe` on Windows and uses an absolute path
- `GOOS=windows go vet -tags=unit` and `GOOS=darwin go vet -tags=unit` pass for `.`,
  `pkg/logging`, `src/helpers/ptx-pft`, `test/unit`

Result: **PASS** (static only).

### TC10: workflow and static checks (AC7)

- `.github/workflows/test.yml` parses as YAML; no `actions/cache` step remains, `setup-go@v4`
  keeps its default module cache
- `gofmt -l` clean, `golangci-lint run ./...` exit 0

Result: **PASS**.

## Findings

### O1: Factory.Close is not called by production code (observation, not blocking)

`ConfigManager` (`pkg/logging/config.go`) creates factories and never closes them, so a
long-running process keeps its log files open as before. This is the existing behaviour and
outside the scope of the issue (test cleanup on Windows); `Close()` now makes a fix possible.

### O2: no dedicated unit test for Factory.Close (recommendation)

The existing tests only call `Close()` in `defer`; nothing asserts that the files are released.
TC04 could be added as a test (Linux-only fd count, or on every platform by removing the log
file after `Close()`, which fails on Windows while the file is open).

## Coverage

- Linux: all three previously failing job commands, skip paths and `Factory.Close()` covered
- Windows / macOS: covered only by static checks until the CI run of PR #245

## CI Notes

- Cross Platform Tests run `make build` and `go test -tags=unit -race ./...` on all three OSes,
  so the Windows fixes are exercised there
- Coverage Report and Performance Tests run only on pull requests (Performance) / on the
  workflow trigger, PR #245 is the first place to see them green

## Retest 1

**Commit**: `76a9a7d` — `fix(246): release log files replaced by ConfigManager.UpdateConfig`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The fix makes `UpdateConfig` hand the open log files of the replaced factory over to the new
one, adds `ConfigManager.Close()` and `CloseGlobalLoggers()`, and makes
`Config.SetModuleLevel` initialize a nil `Modules` map (a panic found by the new test). New
tests: `TestFactoryCloseReleasesFiles`, `TestConfigManagerCloseReleasesReplacedFiles`,
`TestSetModuleLevelWithoutModules`.

Scenarios R1–R6 were run as an ad hoc test file, removed after the run.

| ID | Scenario | Result |
| -- | -------- | ------ |
| R1 | 2 loggers, 50 × `SetLevel`, then `Close()` — descriptors back to start, file removable | PASS |
| R2 | global path: `UpdateGlobalConfig` (file output), `GetLogger`, `SetGlobalLogLevel`, both messages in the file, `CloseGlobalLoggers()` releases everything | PASS |
| R3 | logger used after `Close()` (unsupported) — no panic | PASS |
| R4 | factory reused after `Close()` — newly opened file tracked, second `Close()` releases it | PASS |
| R5 | 20 goroutines create loggers and MCP loggers concurrently, then `Close()`, `-race` | PASS |
| R6 | console-only manager with `SetModuleLevel` + `Close()`; file path in a non-existent directory | PASS |
| R7 | new developer tests, `-race -count=5` | PASS |
| R8 | mutation: `adoptFiles(previous)` removed from `UpdateConfig` → `TestConfigManagerCloseReleasesReplacedFiles` fails (`Open descriptors after Close: 9, before: 7`); code restored | PASS (test detects the leak) |

### Regression — PASS

- `make build`, `go test -tags=unit -race ./...`, benchmarks (`-run='^$'`), `gofmt`,
  `go vet` (`GOOS=windows` / `darwin`), `golangci-lint run ./...`
- `test/integration` Issue #052 logging tests (`-tags=integration`, local-only, not run in CI):
  5 fail on this branch, the same 5 also fail on the baseline `dd23b37` (plus
  `TestIssue052LoggingE2EErrorScenarios` there) — pre-existing (binary build in the test,
  5 % performance threshold, empty high-frequency log file), not a regression of this issue

### O3: open log files grow with every configuration change (observation, not blocking)

Each `UpdateConfig` opens the log file again for every cached logger and keeps the previous
descriptors until `Close()` (R1: +102 descriptors after 50 level changes with 2 loggers).
Before the fix they were never released; they are now releasable, but a long-running process
that changes the level often still accumulates descriptors. A shared writer per file path would
remove the growth. No production binary uses `pkg/logging` today, so this is a follow-up
candidate, not part of this issue.

### Findings status

- **O1**: resolved — replaced files are owned by the manager and released by `Close()`
- **O2**: resolved — dedicated tests, platform-neutral check by removing the file
- **O3**: new, not blocking (resolved in Retest 2)

## Retest 2

**Commits**: `6fb671f` — `fix(246): share one log file per path across loggers and
configuration changes`; `353176d` — `chore(logging): add SPDX license headers`; `6b0f2b1` —
`docs(pkg): add README describing the pkg directory`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The fix replaces the per-factory file list with one registry of open files keyed by the absolute
log path. It is shared by a factory and the factories derived from it (`CreateMCPLogger`,
`CreateTestLogger`, `ConfigManager.UpdateConfig`), so a path is opened once. New developer
tests: `TestFactorySharesFilePerPath`, `TestConfigManagerUpdateConfigReusesFiles`.

Scenarios R9–R16 ran as an ad hoc test file with `-race -count=3`, removed after the run.

| ID | Scenario | Result |
| -- | -------- | ------ |
| R9 | `UpdateConfig` changes the path: the stale logger writes only to the old file, the new logger only to the new one, 2 files tracked, `Close()` releases both and both files are removable | PASS |
| R10 | `rel.log`, its absolute form and `./sub/../rel.log` map to 1 open file | PASS |
| R11 | 20 goroutines create loggers and MCP loggers concurrently — 1 open file, no race | PASS |
| R12 | factory reused after `Close()`: a stale logger does not panic, the new logger reopens the file, `Close()` returns descriptors to the start | PASS |
| R13 | 3 loggers, 30 × (`SetLevel` + `SetModuleLevel`) — +1 descriptor | PASS |
| R14 | 20 × 100 concurrent writes to the shared file — 2000 lines, every line valid JSON | PASS |
| R15 | global manager, two `UpdateGlobalConfig` calls with the same path — +1 descriptor, `CloseGlobalLoggers()` releases it | PASS |
| R16 | log path is a directory (open fails) — nothing tracked, `Close()` returns `nil` | PASS |
| R17 | mutation: `filepath.Abs` normalization disabled → R10 fails (`tracked files: 2, want 1`); code restored | PASS (test detects it) |

O3 measurement from Retest 1 repeated: 2 loggers and 50 level changes previously left
+102 descriptors open, now +1 (developer test and R13).

### License headers and README

- `gofmt -l` clean; a blank line separates each header from the `package` clause and from
  `//go:build`
- `go doc ./pkg/logging` shows `Package logging provides structured logging for Portunix`
- build constraints still select `syslog.go` on Linux and `syslog_windows.go` on Windows
  (`go list` with `GOOS`)
- `pkg/README.md`: markdownlint clean, no sensitive patterns; statements checked against the
  code (no production importer of `pkg/`, `Factory` does not use `handlers`, the Windows syslog
  handler returns an error, `pkg/docker` imports `portunix.ai/app/system`)

### Regression (Retest 2) — PASS

- `make build`, `go test -tags=unit -race ./...`, coverage run (`Close` 88.9 %, `derive`
  100 %, `createFileWriter` 81.0 %), benchmarks (`-run='^$'`)
- `go vet -tags=unit` for linux, windows, darwin; `golangci-lint run ./...`

### Findings status (Retest 2)

- **O1**: resolved
- **O2**: resolved
- **O3**: resolved — descriptors grow with distinct log paths only, not with loggers or
  configuration changes

## Retest 3

**Commits**: `f7692af` — `docs(246): fix markdownlint findings in acceptance protocol and
README`; `67d5fdc` — `docs: fix markdownlint findings in contributing and setup guides`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); markdownlint-cli2 with the
repository configuration; markdown-it-py 3.0.0 (CommonMark) for rendering

Both commits change Markdown only, in 14 files. Each file was rendered before (`2c68e06`) and
after the change, and the rendered structure and text were compared.

| ID | Scenario | Result |
| -- | -------- | ------ |
| D1 | markdownlint findings per changed file, before → after: no file gains a finding; 11 files go to 0; `007`, `181`, `206` keep only findings that existed before (total 88 → 53) | PASS |
| D2 | ordered lists render as one list each: `WINDOWS-SETUP` 1–4 (and 1–2, 1–2, 1–3), `AUR-WORKFLOW` 1–3, `HELPER-BINARY-DEVELOPMENT` 1–7, `getting-started` 1–6; before, every step was a separate list | PASS |
| D3 | code blocks: same count and same content (after removing the list indentation) in every file | PASS |
| D4 | rendered text, word by word: only the intended changes — `**Windows Users:**` (README), heading `Dispatcher Integration` removed, `Issue` added in `241`, `(Retest 2)` in the protocol, and the user's date change in `WINDOWS-SETUP` | PASS |
| D5 | issue references re-wrapped in 7 files: identical rendered structure and text, no line starts with `#NNN` any more (no MD018) | PASS |
| D6 | sensitive patterns in added lines: only the public GitHub release URL in `AUR-WORKFLOW` (re-indented, existed before) | PASS |

Nested bullets in steps 2 and 3 of `getting-started.md` are now part of their list item
(nesting level 2), not separate lists between the steps.

### O4: date in WINDOWS-SETUP (observation)

`**Last Updated:**` was changed from `2025-01-04` to `2025-10-27`; the change was made on
2026-10-27, so the year is likely a typo.

### O5: broken relative links (observation, pre-existing)

`README.md` links to `TEST_GUIDE.md` (3 ×), `TESTING.md` (2 ×) and `TODO:.md`, and
`docs/plugin-development/getting-started.md` to `testing/unit-testing.md` and
`deployment/packaging.md`; none of the targets exists. The links are the same on `2c68e06`, so
they are not caused by these commits.

## Retest 4

**Commit**: `79c5dde` — `docs: fix broken links in READMEs and plugin guide, WINDOWS-SETUP
date`
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); markdownlint-cli2 with the
repository configuration; markdown-it-py 3.0.0 (CommonMark)

Changed files: `README.md`, `README.cs.md`, `README.github.md`, `README.github.cs.md`,
`docs/WINDOWS-SETUP.md`, `docs/plugin-development/getting-started.md`.

| ID | Scenario | Result |
| -- | -------- | ------ |
| L1 | every relative link in the 6 files (64) resolves to an existing file; every `#anchor` matches a heading of the target, slug computed the GitHub way | PASS |
| L2 | link targets of `README.github*.md` (`test/TEST_GUIDE.md`, `test/TESTING.md`) are not in the private-file list of the GitHub publication | PASS |
| L3 | rendered structure (headings, lists, code blocks) identical to `21388d4`; text differs only in the link targets, the `TODO:` label → `Go Code Style`, the Testing / Deployment resource entries and the date | PASS |
| L4 | markdownlint: no finding in the 6 files; `**Last Updated:** 2026-10-27` | PASS |
| L5 | failure injection on the link checker: a missing anchor, a missing file and a link to `docs/adr/` from a GitHub README are all reported, a valid anchor passes | PASS (checker detects them) |

### Findings status (Retest 4)

- **O4**: resolved
- **O5**: resolved — all links named in O5 now point to existing files or sections

## Final Decision

**STATUS**: CONDITIONAL

Initial run, Retest 1 and Retest 2: all locally verifiable acceptance criteria (AC4–AC7) are
met, O1, O2 and O3 resolved, no regressions found. AC1–AC3 (Cross Platform Tests on windows-latest and
macos-latest, Coverage Report, Performance Tests and Quality Gate) must be confirmed by the CI
of PR #245.

**Approval for merge**: YES, once PR #245 shows these jobs green
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)
