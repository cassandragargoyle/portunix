# Acceptance Protocol - Issue #238 (container exec: "--" Separator Passed to the Runtime)

**Issue**: container exec: "--" separator is passed to the runtime as the command
**Branch**: `fix/238-container-exec-double-dash-separator` (commit `0c58df7`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, unit tests, static checks, E2E scenarios
  with `ubuntu:22.04` containers started through `portunix container run`; Go 1.26.8,
  Podman 5.7.0

No software was installed on the host. Every test container was removed after the run.

## Scope Note

The branch removes the first standalone `--` from the `container exec` arguments (before or
after the container name), limits the help flag detection of `container run` to the arguments
before `--` and of `container exec` to the first argument, documents `--` in the exec help and
extracts `parseExecArgs`, `buildExecArgs` and `hasHelpFlag` for unit tests.

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Docker — the daemon is not available on the test host; `execDockerCommand` uses the same
  `buildExecArgs` as the Podman path and is covered only by unit tests
- Windows execution

## Test Summary

- Total test scenarios: 26
- Passed: 26
- Failed: 0
- Skipped: 2 (Docker, Windows)
- Defects found in the scope of #238: 0
- Observations: 3 (O1–O3)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | New tests in `run_args_test.go` and the existing `ptx-container` tests | host |
| Static | `go vet`, `gofmt -l` of `src/helpers/ptx-container`, `make build` | host |
| E2E | Acceptance criteria 1–5 through the main `portunix` binary (dispatcher) | Podman, `ubuntu:22.04` |
| Failure injection | Missing name/command, lone and repeated `--`, `--` inside the command, unknown container, failing command, help flag positions, stdin pass-through | Podman, `ubuntu:22.04` |
| Regression | `container run` separator handling from #235, `container list`, help texts | host + Podman |

## Test Cases

Setup for the exec cases: `portunix container run -d --name t238a ubuntu:22.04 sleep 600`.

### TC01: Exec with separator

- **When** `portunix container exec t238a -- echo x`
- **Then** the output is `x`, exit code 0

### TC02: Exec without separator

- **When** `portunix container exec t238a echo x`
- **Then** the output is `x`, exit code 0

### TC03: Only the first separator is removed

- **When** `portunix container exec t238a -- sh -c 'echo -- x'`
- **Then** the output is `-- x`

### TC04: Separator before the container name

- **When** `portunix container exec -- t238a echo before`
- **Then** the output is `before`

### TC05: Help flag inside the exec command

- **When** `portunix container exec t238a ls -h /etc/hostname`
- **Then** the command runs, the output is `/etc/hostname`

### TC06: Help flag inside the exec command after the separator

- **When** `portunix container exec t238a -- ls -h /etc/hostname`
- **Then** the command runs, the output is `/etc/hostname`

### TC07: Help flag inside the run command after the separator

- **When** `portunix container run ubuntu:22.04 -- ls -h /etc/hostname`
- **Then** the container runs the command, the output is `/etc/hostname`

### TC08: Help is still shown

- **When** `container run --help`, `container run -h`, `container exec --help`, `container exec -h`
- **Then** the matching help text is shown

### TC09: Unit tests

- **When** `go test -count=1 -v .` in `src/helpers/ptx-container`
- **Then** all tests pass, including `TestHasHelpFlag`, `TestParseExecArgs`, `TestBuildExecArgs`

### TC10: Static checks and build

- **When** `go vet ./...`, `gofmt -l .`, `make build`
- **Then** no findings, all binaries built

### FI01: Exit code of a failing command

- **When** `portunix container exec t238a -- sh -c 'exit 3'`
- **Then** Portunix exits non-zero

### FI02–FI05: Missing container name or command

- **When** `container exec`, `container exec t238a`, `container exec t238a --`, `container exec --`
- **Then** the exec help is shown, no panic

### FI06: Double separator

- **When** `portunix container exec t238a -- -- echo x`
- **Then** the second `--` is the command, the runtime reports `exec: "--"`, Portunix exits
  non-zero

### FI07: `--` in the command without a leading separator

- **When** `portunix container exec t238a echo a -- b`
- **Then** documented behaviour, the same as `container run` after #235: the first `--` is
  removed, output `a b`

### FI08: Unknown container

- **When** `portunix container exec t238-missing -- echo x`
- **Then** the runtime error `no such container` is shown, Portunix exits non-zero

### FI09: Help flag after the container name

- **When** `portunix container exec t238a --help`
- **Then** `--help` is the command; the runtime reports `exec: "--help"`, Portunix exits
  non-zero (see O1)

### FI10: Help flag before the container name

- **When** `portunix container exec -h t238a ls`
- **Then** the exec help is shown

### FI11: Stdin pass-through

- **When** `echo piped | portunix container exec t238a -- cat`
- **Then** the output is `piped` (non-TTY path, `exec -i`)

### FI12: Run help flag without separator

- **When** `portunix container run ubuntu:22.04 ls -h /etc/hostname`
- **Then** the run help is shown (known limitation, see O2)

### FI13: Run help flag before the separator

- **When** `portunix container run --rm -h ubuntu:22.04 -- ls`
- **Then** the run help is shown

### FI14: `--` as part of a value

- **When** `portunix container exec t238a -- sh -c 'X=--; echo "X=$X"'`
- **Then** the output is `X=--`, only the standalone `--` is removed

### REG1: `container run` separator from #235

- **When** `portunix container run ubuntu:22.04 -- sh -c 'echo -- y'`
- **Then** the output is `-- y`

### REG2: Container list

- **When** `portunix container list` with `t238a` running
- **Then** `t238a` is listed

## Test Results

### Unit Tests and Static Checks

| Check | Result |
| ----- | ------ |
| `go test -count=1 -v .` (`ptx-container`) | ✅ ok, 26 tests passed, 3 of them new |
| `go vet ./...` | ✅ no findings |
| `gofmt -l .` | ✅ clean |
| `make build` | ✅ main binary and all helpers |

### E2E — Podman

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `x`, exit 0 |
| TC02 | ✅ PASS | `x`, exit 0 |
| TC03 | ✅ PASS | `-- x` |
| TC04 | ✅ PASS | `before` |
| TC05 | ✅ PASS | `/etc/hostname` |
| TC06 | ✅ PASS | `/etc/hostname` |
| TC07 | ✅ PASS | `/etc/hostname` |
| TC08 | ✅ PASS | run help for `--help` / `-h`, exec help for `--help` / `-h` |

### Failure Injection — Podman

| Case | Result | Evidence |
| ---- | ------ | -------- |
| FI01 | ✅ PASS | `exit status 3`, Portunix exit 1 |
| FI02–FI05 | ✅ PASS | exec help, exit code 0 (see O3) |
| FI06 | ✅ PASS | `exec: "--": executable file not found`, exit 1 |
| FI07 | ✅ PASS | `a b` |
| FI08 | ✅ PASS | `no such container`, exit 1 |
| FI09 | ✅ PASS | `exec: "--help": executable file not found`, exit 1 |
| FI10 | ✅ PASS | exec help |
| FI11 | ✅ PASS | `piped` |
| FI12 | ✅ PASS | run help |
| FI13 | ✅ PASS | run help |
| FI14 | ✅ PASS | `X=--` |

### Regression

| Check | Result |
| ----- | ------ |
| REG1 `container run -- sh -c 'echo -- y'` | ✅ `-- y` |
| REG2 `container list` | ✅ `t238a` listed |
| #235 unit tests (`TestStripCommandSeparator`, `TestBuildRunArgs_Separator`, `TestBuildManagedRunArgs_Separator`, `TestExtractLifecycleFlags_StopsAtSeparator`) | ✅ pass |

The `test/e2e` container run suite was not run; it is known to be red on `main` as well
(`acceptance-235.md`, O3) and has no exec cases.

### Skipped

- Docker execution (daemon not available)
- Windows execution

## Acceptance Criteria

| Criterion | Result |
| --------- | ------ |
| 1. `container exec <name> -- echo x` prints `x` with Podman and with Docker | ✅ Podman (TC01); Docker not executed, covered by unit tests of the shared `buildExecArgs` |
| 2. `container exec <name> echo x` keeps working | ✅ TC02 |
| 3. `container exec <name> -- sh -c 'echo -- x'` prints `-- x` | ✅ TC03 |
| 4. `container run ubuntu:22.04 -- ls -h /` and `container exec <name> ls -h /` run the command | ✅ TC05, TC07 |
| 5. `container run --help` and `container exec --help` still show the help | ✅ TC08 |
| 6. Unit tests cover exec argument handling with and without `--` and help detection for `run` and `exec` | ✅ TC09 — `TestParseExecArgs`, `TestHasHelpFlag`, `TestBuildExecArgs` |

## Coverage

- **Separator handling in exec**: none, after the name, before the name, trailing, lone,
  repeated, inside the command, inside a value
- **Help detection**: run — before and after `--`, without `--`; exec — first argument, before
  the name, after the name, after `--`
- **Exec argument building**: TTY and non-TTY variants by unit tests, non-TTY path by E2E
  (stdin pipe)
- **Gaps**: Docker runtime execution, Windows, interactive TTY path in E2E; `handleContainerExec`
  itself is not unit tested (it calls the runtime and `os.Exit`)

## CI Notes

- The new unit tests need no container runtime and no network
- The E2E cases above can be scripted against any runtime once a container named `t238a` is
  started; they are not part of `test/e2e` yet

## Observations

- **O1**: behaviour change — `container exec <name> --help` used to show the Portunix help, now
  `--help` is executed in the container. This is the intended behaviour from the issue
  ("only when it comes before the container name"), and the help text now says so
- **O2**: `container run ubuntu:22.04 ls -h /` without `--` still shows the help. Accepted
  limitation: run accepts runtime flags before the image, so the image position cannot be found
  without parsing every runtime flag; the documented form `run <image> -- ls -h /` works
- **O3**: exec with a missing container name or command shows the help with exit code 0 (also on
  `main`); the same applies to `Image name required` in run (`acceptance-235.md`, O2)

## Recommendations

1. Run TC01–TC06 and FI01 once with Docker
2. Return a non-zero exit code when exec is called without a container name or command (O3)

## Final Decision

**STATUS**: PASS

Linux with Podman: all executed scenarios passed, acceptance criteria 2–6 are fully met and
criterion 1 is met for Podman, no regression in the `container run` separator handling from
issue #235. The Docker path shares the argument builder with the Podman path and is covered by unit
tests.

**Approval for merge**: YES
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
