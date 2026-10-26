# Acceptance Protocol - Issue #195 (Stale container command tests in src/cmd)

**Issue**: Fix Stale Container Command Tests in `src/cmd` (Surfaced by #194)
**Branch**: `fix/issue-195-stale-src-cmd-container-tests` (commits `72d292a`, `31ac0f7`, `f11a427`;
retest 1 on `e6c2b1e`, `416108e`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-25, retest 1 2026-10-26
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, static checks, unit tests, CLI
  comparison against `main`; Go 1.26.8
- Alpine Linux (`alpine:latest`) containers in rootless Podman — real `container run` and
  `container exec` through the fallback commands

No software was installed. The fallback commands ran from a directory that contains only the
`portunix` binary, so the dispatcher did not hand `container` to `ptx-container`. All test
containers were named `t195-*` and removed after the test.

## Scope Note

`main.go` hands `container`, `docker` and `podman` to the `ptx-container` helper when it lies
next to the binary. The Cobra commands in `src/cmd/container.go` are the fallback when the
helper is missing. The branch:

- rewrites the 6 failing test functions to test the real commands (`Args` validators, the run
  flags through `addContainerRunFlags`, `splitContainerRunArgs`, `parseContainerExecArgs`)
  instead of synthetic Cobra commands, and removes the placeholder
  `TestContainerRunCommand_ErrorHandling`
- changes the fallback: `container run` parses flags only before the image
  (`SetInterspersed(false)`) and drops a `--` right after it; `container exec` reads
  `-i`/`-it` only before the container name
- updates `src/cmd/README.md` and `README.cs.md`, documents `init()` of
  `docker_install_alias.go` (`31ac0f7`)
- adds `SPDX-License-Identifier: MIT` to 46 file headers (`f11a427`)

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Windows and macOS — the package cross-compiles and vets, nothing was executed
- Docker — only Podman is installed; both CLIs parse the forwarded arguments the same way
- The primary `ptx-container` path — not changed by the branch (no diff in `src/helpers`,
  `src/dispatcher`, `main.go`); its unit tests pass

## Test Summary

- Total test scenarios: 22
- Passed: 22
- Failed: 0
- Skipped: 0
- Defects found: 0
- Observations: 3 (O1–O3)

Retest 1 on `416108e` (O1–O3 resolved, unit-tagged tests repaired):

- Total test scenarios: 20
- Passed: 20
- Failed: 0
- Defects found: 0
- Observations: O1–O3 resolved, no new observations

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | Build, `gofmt`, `go vet`, cross-compile, gosec, Markdown lint | host |
| Unit | `portunix.ai/cmd` once, 3× with `-race`, `make test-cli`, `make test GO_MODULES=.`, isolation from `HOME` | host |
| Regression | Same suite on `main` | host, worktree |
| CLI | Fallback `container run` and `container exec` on `main` and on the branch | host, Podman containers |
| Review | README accuracy, SPDX commit, remaining container tests | repository |

## Test Cases

### TC01: The cmd module passes (AC1, AC4)

- **Given** the branch
- **When** `go test portunix.ai/cmd/...` and `make test-cli` run
- **Then** both exit 0 with no failing test

### TC02: Failing tests test the real commands or are removed with a reason (AC2)

- **Given** the 6 test functions that fail on `main`
- **When** the diff is reviewed
- **Then** each one calls the real command's validator, flags or parsing helper, or is removed
  because it asserted only its own test data

### TC03: Production changes fix real defects (AC3)

- **Given** a binary without `ptx-container` next to it, built from `main` and from the branch
- **When** the same `container run` and `container exec` commands run against Podman
- **Then** every behaviour the branch changes is wrong on `main` and right on the branch, and
  every behaviour it keeps is the same on both

## Test Results

### Static Checks

| ID | Check | Result |
| -- | ----- | ------ |
| S1 | `make build` | PASS |
| S2 | `gofmt -l src/cmd`, `go vet portunix.ai/cmd` | PASS |
| S3 | Cross-compile and vet `GOOS=windows` / `darwin` | PASS |
| S4 | `uv run --script scripts/gosec-scan.py` | PASS — 23 of 23 modules |
| S5 | `markdownlint-cli2` of `src/cmd/README.md` and `README.cs.md` | PASS |

### Unit Tests

| ID | Suite | Result |
| -- | ----- | ------ |
| U1 | `go test portunix.ai/cmd/...` | PASS — exit 0 |
| U2 | `go test -race -count=3 portunix.ai/cmd/...` | PASS |
| U3 | `make test-cli` | PASS — 39 top-level tests passed, 0 failed, 0 skipped |
| U4 | `make test GO_MODULES=.` (root module with satellite packages) | PASS — 10 packages `ok`, 0 `FAIL` |
| U5 | `go test portunix.ai/cmd/...` on `main` | 6 top-level tests fail — the state the issue describes |
| U6 | Container tests with `HOME` set to an empty directory | PASS — no root help printed, nothing under `.portunix`; only the Go toolchain wrote `.cache/go-build` and `.config/go` |
| U7 | Each container test function alone with an empty `HOME` (23 functions) | PASS — none writes outside the Go toolchain directories |
| U8 | `ptx-container` unit tests | PASS |

### CLI — Fallback Commands, `main` against the Branch

| ID | Command | `main` | Branch | Result |
| -- | ------- | ------ | ------ | ------ |
| F1 | `container run --name … alpine:latest sh -c "echo F1-ok"` (original #038 form) | `Error: unknown shorthand flag: 'c' in -c` | runs `sh -c "echo F1-ok"`, prints `F1-ok` | PASS — defect on `main` fixed |
| F2 | `container run --name … alpine:latest -- sh -c "echo F2-ok"` | `F2-ok`, `--` not passed | `F2-ok`, `--` not passed | PASS — unchanged |
| F3 | `container run --name … alpine:latest ls -d /tmp` | runs `podman run -d … ls /tmp`: `-d` of `ls` taken as detach, the container runs `ls /tmp` in the background | runs `ls -d /tmp`, prints `/tmp` | PASS — silent misinterpretation on `main` fixed |
| E1 | `container exec -it t195-exec echo E1-ok` | `E1-ok`, see O1 | `E1-ok` | PASS |
| E2 | `container exec t195-exec grep -i ROOT /etc/passwd` | `exit status 1`: `-i` removed from the `grep` command | prints the `root` line | PASS — defect on `main` fixed |
| E3 | `container exec -i t195-exec echo E3-ok` | `E3-ok` | `E3-ok` | PASS — unchanged |
| E4 | `container exec t195-exec echo E4-ok` | `E4-ok` | `E4-ok` | PASS — unchanged |
| E5 | `container exec -it t195-exec` (no command) | — | `usage: portunix container exec [flags] <container-name> <command>` | PASS |
| E6 | `container exec`, `container exec web`, `container run` without arguments | — | `requires at least 2 arg(s)` / `requires at least 1 arg(s)` | PASS |

### Review

| ID | Item | Result |
| -- | ---- | ------ |
| R1 | Removed `TestContainerRunCommand_ErrorHandling` | PASS — it asserted that its own `scenario` field contains `error` and called no Portunix code; removal is allowed by AC2 |
| R2 | `src/cmd/README.md` / `README.cs.md` | PASS — every listed file exists, every non-test `.go` file is listed, the dispatcher table matches `registerHelpers` in `src/dispatcher/dispatcher.go`, the call order matches `main.go` |
| R3 | `f11a427` (SPDX) | PASS — 46 files, one added header line each, no other change; incomplete, see O3 |
| R4 | No change to `src/helpers`, `src/dispatcher`, `main.go` | PASS |

## Retest 1 — Commits `e6c2b1e`, `416108e`

The developer resolved O1–O3 and found one more problem on the way:

- O2: runtime selection rules extracted from `GetSelectedRuntime` into `chooseRuntime` with
  injectable checks and tested in `src/app/container/runtime_test.go`; the run command calls
  the runtime through replaceable functions, and the former placeholders now test delegation,
  options translation and runtime errors; `TestContainerExecIntegration` (no assertion) removed
- the unit-tagged `container_go_commands_test.go` did not compile (unused import) on `main`
  either, so it never ran; it builds now and keeps only the tests of the real
  `run-in-container` commands
- O3: SPDX line added to the remaining files
- O1: the issue detail file has an Implementation section with the accurate account

### Static Checks and Test Suites

| ID | Check | Result |
| -- | ----- | ------ |
| RS1 | `make build` | PASS |
| RS2 | `gofmt`; `go vet` of `cmd` and `app/container`, also with `-tags=unit` | PASS |
| RS3 | Cross-compile and vet `GOOS=windows` / `darwin` | PASS |
| RS4 | `uv run --script scripts/gosec-scan.py` | PASS — 23 of 23 modules |
| RS5 | `markdownlint-cli2` of the issue file and this protocol | PASS |
| RU1 | `go test -race -count=3` of `portunix.ai/cmd/...` and `portunix.ai/app/container` | PASS |
| RU2 | `go test -tags=unit -v portunix.ai/cmd` | PASS — 130 tests and subtests, 0 failed (did not compile before) |
| RU3 | `make test-cli` | PASS — 38 top-level tests (one fewer: the empty exec integration test was removed) |
| RU4 | `make test GO_MODULES=.` | PASS — 10 packages `ok` |
| RU5 | `TestChooseRuntime` | PASS — 9 cases incl. daemon not running, not functional, unknown runtime, no silent fallback |

### Review

| ID | Item | Result |
| -- | ---- | ------ |
| RR1 | `GetSelectedRuntime` refactor | PASS — mechanical: same rules and messages, the checks are called through `systemRuntimeChecks`, which points at the existing `Is*Available` / `Is*Running` |
| RR2 | Test seams in `container.go` | PASS — only the run command uses them; by default they are `container.GetSelectedRuntime`, `docker.RunContainer`, `podman.RunContainer`; the other commands call the real functions directly |
| RR3 | Mutation of `chooseRuntime`: default `docker` / docker daemon check skipped / silent fallback to Docker | PASS — the tests fail on each (2, 1 and 1 failing subtests); restored afterwards |
| RR4 | `416108e` | PASS — 6 files, one SPDX header line each; all 57 `.go` files in `src/cmd` now carry it |

### CLI — Fallback Binary from `416108e` against Podman

| ID | Command | Result |
| -- | ------- | ------ |
| RF1 | `container run --name … alpine:latest sh -c "echo RF1-ok"` | PASS — `RF1-ok` through the seams |
| RF3 | `container run --name … alpine:latest ls -d /tmp` | PASS — prints `/tmp` |
| RF4 | `container run -d --name … -e A=1 -p 18080:80 --network podman alpine:latest sleep 30` | PASS — `podman inspect` shows `A=1` and `80/tcp->18080` |
| RE2 | `container exec … grep -i ROOT /etc/passwd` | PASS — prints the `root` line |
| RF5 | `container_runtime: docker` in the config, Docker not installed | PASS — `Error: Docker not installed` with the hint, no container started |
| RF6 | `container_runtime: lxc` in the config | PASS — `unknown container runtime: lxc (must be 'docker' or 'podman')` |

All `t195r-*` test containers were removed.

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | `go test portunix.ai/cmd/...` passes (0 failures) | PASS (U1, U2) |
| 2 | Each failing test is updated to assert the current intended behaviour, or removed with justification | PASS (TC02, R1) |
| 3 | No production code changes purely to satisfy stale assertions; genuine defects are fixed | PASS — the run and exec changes fix F1, F3 and E2, which fail on `main`; F2, E3 and E4 are unchanged (TC03) |
| 4 | `make test-cli` exits 0 | PASS (U3) |

## Coverage

- Fallback `container run`: flag parsing before the image, command arguments with flags, `--`
  separator, missing image; executed against Podman
- Fallback `container exec`: `-i`, `--interactive`, `-it`, command flags after the container
  name, missing command; executed against Podman
- Help texts of stop, start, restart, rm, logs and list by assertion
- Not covered: the runtime-error paths of the fallback commands (no dependency injection for
  `container.GetSelectedRuntime`)

## CI Notes

- `portunix.ai/cmd` passes for the first time since #194 made `make test` run it, so the
  Unit Tests job can include it without exceptions
- The container tests no longer execute Cobra commands, so they need no container runtime and
  do not read the user's `~/.portunix`

## Observations

- **O1**: the developer's summary and the commit message say that the `-it` help example of
  `container exec` was broken on `main`. The code did take `-t` as the container name, but the
  runtime call became `podman exec -it -t <name> …` and Podman read `-t` as its own flag, so
  the example worked (E1). The visible exec defect was E2. The new code is still the cleaner
  form; no action needed beyond accurate wording
- **O2**: `TestContainerRunCommand_RuntimeDelegation` and
  `TestContainerRunCommand_OptionsTranslation` compare their own test data (a mock that is never
  called, options structs built inside the test) and `TestContainerExecIntegration` has no
  assertion. They pass but test no Portunix code. They were not failing, so they are outside
  #195; a follow-up issue should rewrite or remove them
- **O3**: after `f11a427`, 9 of 57 `.go` files in `src/cmd` still lack the SPDX line:
  `config.go`, `datastore.go`, `docker_run_in_container.go` and 6 test files
  (`container_exec_test.go`, `container_go_commands_test.go`, `container_management_test.go`,
  `container_run_integration_test.go`, `container_run_test.go`, `docker_install_alias_test.go`)

## Recommendations

1. Merge the branch
2. Run the fallback scenarios on Windows once a Windows tester is available

O2 and O3 are resolved in retest 1; no follow-up issue is needed.

## Final Decision

**STATUS**: PASS (retest 1)

Retest 1 confirms the follow-up commits: the placeholder tests are replaced by tests that fail
when the code is broken (RR3), the unit-tagged tests compile and pass, the refactored runtime
selection behaves as before against Podman, and every file carries the SPDX line.

Linux: all four acceptance criteria are met. The `src/cmd` module passes completely and under the
race detector, the tests no longer touch the user's home directory, and the production changes
fix three behaviours that are wrong on `main` (F1, F3, E2) without changing the ones that were
right. The observations do not block the merge.

**Approval for merge**: YES
**Date**: 2026-10-26
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
