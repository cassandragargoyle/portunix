# Acceptance Protocol - Issue #235 (container run: "--" Separator Passed to the Runtime)

**Issue**: container run: "--" separator from the help example is passed to the runtime as the command
**Branch**: `fix/235-container-run-double-dash-separator` (commit `c6b7453`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, unit tests, static checks, E2E scenarios
  with `ubuntu:22.04` containers started through `portunix container run`; Go 1.26.8,
  Podman 5.7.0

No software was installed on the host. Every test container was removed after the run.

## Scope Note

The branch removes the first standalone `--` from the `container run` arguments before they
are handed to podman/docker (direct and managed path), stops the detach detection and the
lifecycle flag parser at `--`, and extracts `buildRunArgs` / `buildManagedRunArgs` for unit
tests.

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Docker — not installed on the test host; the Docker path uses the same `buildRunArgs` as the
  Podman path and is covered only by unit tests
- Windows execution

## Test Summary

- Total test scenarios: 18
- Passed: 18
- Failed: 0
- Skipped: 2 (Docker, Windows)
- Defects found in the scope of #235: 0
- Observations: 4 (O1–O4)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | New tests in `run_args_test.go` and the existing `ptx-container` tests | host |
| Static | `go vet`, `gofmt -l` of `src/helpers/ptx-container`, `make build` | host |
| E2E | Acceptance criteria 1–3 for the direct and the managed path | Podman, `ubuntu:22.04` |
| Failure injection | `-d` and lifecycle flags after `--`, lone `--`, `--` before the image, `-- --`, failing command, `--` inside a flag value | Podman, `ubuntu:22.04` |
| Regression | `test/e2e` container run suite on the branch and on `main`, `container list` / `exec`, help text | host + Podman |

## Test Cases

### TC01: Help example, detached

- **Given** the binary built from the branch
- **When** `portunix container run -d --name t235a ubuntu:22.04 -- bash -c "echo test"`
- **Then** the runtime arguments contain no `--` after the image, the container logs `test`,
  `--rm` is not used

### TC02: Same command without `--`

- **When** `portunix container run -d --name t235b ubuntu:22.04 bash -c "echo test"`
- **Then** identical runtime arguments as TC01, the container logs `test`

### TC03: Only the first separator is removed

- **When** `portunix container run ubuntu:22.04 -- sh -c 'echo -- x'`
- **Then** the output is `-- x`

### TC04: Interactive run with `--`

- **When** `portunix container run ubuntu:22.04 -- echo hello` (stdin not a terminal)
- **Then** `run -i --rm ubuntu:22.04 echo hello`, output `hello`

### TC05: Managed path, foreground

- **When** `portunix container run --auto-cleanup ubuntu:22.04 -- sh -c 'echo -- y'`
- **Then** lifecycle labels are set, no `--` after the image, output `-- y`

### TC06: Managed path, detached with `--ttl`

- **When** `portunix container run --ttl 1h -d --name t235c ubuntu:22.04 -- sleep 30`
- **Then** `portunix.ttl=1h0m0s` label, `-d --name t235c ubuntu:22.04 sleep 30`, the container
  is running and labelled `portunix.managed=true`

### TC07: Unit tests

- **When** `go test -count=1 ./...` in `src/helpers/ptx-container`
- **Then** all tests pass

### TC08: Static checks and build

- **When** `go vet ./...`, `gofmt -l .`, `make build`
- **Then** no findings, all binaries built

### FI1: `-d` inside the container command

- **When** `portunix container run ubuntu:22.04 -- ls -d /tmp`
- **Then** the run is not detached (`-i --rm`), output `/tmp`

### FI2: Lifecycle flags inside the container command

- **When** `portunix container run --auto-cleanup ubuntu:22.04 -- echo --ttl --auto-cleanup`
- **Then** the output is `--ttl --auto-cleanup`, the flags are not parsed

### FI3: Lone separator

- **When** `portunix container run --`
- **Then** `Image name required` with usage, no panic

### FI4: Separator before the image

- **When** `portunix container run -- ubuntu:22.04 echo before`
- **Then** `run -i --rm ubuntu:22.04 echo before`, output `before`

### FI5: Double separator

- **When** `portunix container run ubuntu:22.04 -- -- echo x`
- **Then** the second `--` is the container command, the runtime reports `exec: "--"` and
  Portunix exits non-zero

### FI6: Exit code of a failing command

- **When** `portunix container run ubuntu:22.04 -- sh -c 'exit 3'`
- **Then** Portunix exits non-zero

### FI7: `--` in the command without a leading separator

- **When** `portunix container run ubuntu:22.04 echo a -- b`
- **Then** documented behaviour: the first `--` is removed, output `a b` (see O1)

### FI8: `--` as part of a flag value

- **When** `portunix container run -e X=-- ubuntu:22.04 -- sh -c 'echo "X=$X"'`
- **Then** output `X=--`, only the standalone `--` is removed

### REG1: `test/e2e` container run suite against `main`

- **When** `go test -count=1 -v -run TestContainerRun ./...` in `test/e2e` on the branch and on
  `main` (temporary worktree)
- **Then** the same set of cases passes and fails

### REG2: Other container commands and help

- **When** `container exec <name> echo exec-ok`, `container list`, `container run --help`
- **Then** exec and list work, the help keeps the `--` example and states that everything after
  the first `--` is passed verbatim

## Test Results

### Unit Tests and Static Checks

| Check | Result |
| ----- | ------ |
| `go test -count=1 ./...` (`ptx-container`) | ✅ ok, 23 tests and subtests passed, 6 of them new |
| `go vet ./...` | ✅ no findings |
| `gofmt -l .` | ✅ clean |
| `make build` | ✅ main binary and all helpers |

### E2E — Podman

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `[run -d --name t235a ubuntu:22.04 bash -c echo test]`, logs `test`, exit code 0, `autoremove=false` |
| TC02 | ✅ PASS | same arguments as TC01, logs `test` |
| TC03 | ✅ PASS | `-- x` |
| TC04 | ✅ PASS | `[run -i --rm ubuntu:22.04 echo hello]`, `hello` |
| TC05 | ✅ PASS | `[run -i --rm --label portunix.managed=true … --label portunix.policy=on-exit ubuntu:22.04 sh -c echo -- y]`, `-- y` |
| TC06 | ✅ PASS | `portunix.ttl=1h0m0s`, `… -d --name t235c ubuntu:22.04 sleep 30`, `running managed=true` |

### Failure Injection — Podman

| Case | Result | Evidence |
| ---- | ------ | -------- |
| FI1 | ✅ PASS | `[run -i --rm ubuntu:22.04 ls -d /tmp]`, `detached: false`, `/tmp` |
| FI2 | ✅ PASS | `--ttl --auto-cleanup` |
| FI3 | ✅ PASS | `❌ Error: Image name required` (exit code 0, see O2) |
| FI4 | ✅ PASS | `[run -i --rm ubuntu:22.04 echo before]`, `before` |
| FI5 | ✅ PASS | `exec: "--"`, exit 1 |
| FI6 | ✅ PASS | exit 1 |
| FI7 | ✅ PASS | `[run -i --rm ubuntu:22.04 echo a b]`, `a b` |
| FI8 | ✅ PASS | `X=--` |

### Regression

| Check | Branch | `main` |
| ----- | ------ | ------ |
| `test/e2e` `TestContainerRunE2ETestSuite` | 1 of 8 groups pass (`ErrorScenarios`), 7 fail | the same 7 groups fail |
| TC-038-E006 (`… ubuntu:22.04 -- bash -c …`) | container starts, only the expected output message is missing | `exec: "--": executable file not found`, exit 127 — the #235 defect |
| `container exec`, `container list` | ✅ work | — |
| `container run --help` | ✅ example kept, separator explained | — |

The `test/e2e` failures are not caused by the branch (see O3). No test containers were left
after either run.

### Skipped

- Docker execution (Docker not installed)
- Windows execution

## Acceptance Criteria

| Criterion | Result |
| --------- | ------ |
| 1. The help example starts the container and runs `bash -c "echo test"` with Podman and with Docker | ✅ Podman (TC01, TC06); Docker not executed, covered by unit tests of the shared `buildRunArgs` |
| 2. The same command without `--` keeps working | ✅ TC02 |
| 3. `container run ubuntu:22.04 -- sh -c 'echo -- x'` prints `-- x` | ✅ TC03 |
| 4. Unit tests cover with/without `--`, detached and interactive runs and the managed path | ✅ TC07 — `TestBuildRunArgs_Separator`, `TestBuildManagedRunArgs_Separator`, `TestExtractLifecycleFlags_StopsAtSeparator` |

## Coverage

- **Separator handling**: no separator, after the image, before the image, trailing, lone,
  repeated, inside a flag value
- **Detach detection**: `-d` / `--detach` before and after `--`
- **Lifecycle parser**: stops at `--`, flags after it are kept for the command
- **Both run paths**: direct (`buildRunArgs`) and managed (`buildManagedRunArgs`) by unit tests
  and E2E
- **Gaps**: Docker runtime execution, Windows; `handleContainerRun` itself is not unit tested
  (it calls the runtime and `os.Exit`), the tests mirror its sequence of helper calls

## CI Notes

- The new unit tests need no container runtime and no network
- `test/e2e` `TestContainerRunE2ETestSuite` is red on `main` and on the branch; it should not be
  used as a gate until it is updated (O3)

## Observations

- **O1**: without a leading separator, a `--` that belongs to the command is removed
  (`ubuntu:22.04 git log -- file`); the workaround is the documented form
  `ubuntu:22.04 -- git log -- file`. Expected by the design: the image position cannot be found
  without parsing every runtime flag that takes a value
- **O2**: `Image name required` exits with code 0 (also on `main`)
- **O3**: the `test/e2e` container run suite is outdated: it expects the message
  `Container started successfully in detached mode`, which `ptx-container` does not print,
  searches `container list` for full container names although the list truncates them, and uses
  the short image name `nginx:alpine`, which Podman does not resolve without
  `unqualified-search-registries`
- **O4**: `portunix container exec <name> -- echo x` has the same defect: the runtime runs `--`
  as the command (`exec: "--": executable file not found`, exit 127). `container exec --help`
  does not document `--`, so it is outside the scope of #235

## Recommendations

1. Run TC01–TC06 once with Docker
2. Open an issue for O4 (`container exec` separator) or extend the fix to `handleContainerExec`
3. Update or remove the outdated `test/e2e` container run suite (O3, related to #225)
4. Return a non-zero exit code for `Image name required` (O2)

## Final Decision

**STATUS**: PASS

Linux with Podman: all executed scenarios passed, acceptance criteria 2–4 are fully met and
criterion 1 is met for Podman, no regression against `main`. The Docker path shares the
argument builder with the Podman path and is covered by unit tests.

**Approval for merge**: YES
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
