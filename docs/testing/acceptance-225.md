# Acceptance Protocol - Issue #225 (CI Unit Tests Job Runs Integration/E2E Tests)

**Issue**: CI Unit Tests job runs integration/E2E tests and always fails
**Branch**: `fix/225-ci-unit-tests-build-tags` (commit `19867b4`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Debian-based
`docker.io/library/golang:1.26` (container, run with `portunix container run` on Podman)

The CI-equivalent unit run (AC1) ran in a clean container without a container runtime,
started with `portunix container run`. The static checks and the sample integration tests
install no software and ran on the host.

## Scope Note

The branch changes five things:

1. `//go:build integration` in 48 untagged files in `test/integration`, `//go:build e2e`
   in 3 untagged files in `test/e2e`
2. Compile fixes in the files that were already tagged and had never compiled: duplicate
   declarations, testcontainers v0.38 API (`BindMounts`, `ExecResult`), unused variables
3. `pkg/docker/docker_integration_test.go`: local `pullImageIfNeeded` and
   `detectPackageManager` helpers, `isDockerAvailable` returns `false` instead of panicking
   when no Docker host is found
4. `TestMainWithoutArguments` expects the custom help layout
5. The CI Integration Tests job runs every package except `test/integration`

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- AC4 (green Unit Tests job on the next release PR) — can be verified only on GitHub
- `pkg/docker` integration tests against a real Docker daemon — the host has Podman only, so
  the suite skips; the CI Integration Tests job runs it against `docker:24-dind`
- Windows and macOS (Cross Platform job) — not executed; the change touches only test files
  and build tags
- Separate Go modules (`src/cmd`, `src/app`, helpers) — not part of the root `./...` run in CI

## Test Summary

| Round | Commit | Scenarios | Passed | Failed | Skipped | Decision |
| ----- | ------ | --------- | ------ | ------ | ------- | -------- |
| 1 | `19867b4` | 9 | 9 | 0 | 3 | PASS |

- Skipped: AC4 on GitHub, `pkg/docker` against Docker, Windows/macOS
- Defects found: 0 in scope; 1 unrelated observation (O1)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `go test -tags=unit -race -count=1 ./...` after `make build`, as in the CI job | container, no runtime |
| Static | Build tag coverage, `go list` with `-tags=unit` | host |
| Static | `go vet` with `-tags=integration` and `-tags=e2e`, `gofmt -s -l .`, `go mod tidy -diff` | host |
| CI config | Package list of the Integration Tests job, YAML parse of `test.yml` | host |
| Integration | `pkg/docker` without Docker (skip, no panic); sample `test/integration` tests with the tag | host |

## Test Cases

### TC01: Unit run on a clean machine without a container runtime (AC1)

- **Given** a fresh `golang:1.26` container, the branch cloned from a read-only mount,
  `command -v docker podman` finds nothing
- **When** `make build` and `go test -tags=unit -race -count=1 ./...`
- **Then** every package passes, exit 0

### TC02: Integration and E2E files are excluded by a build tag (AC2)

- **When** the test files are checked for `//go:build` and listed with `-tags=unit`
- **Then** no untagged file remains, `go list -tags=unit` excludes all files in
  `test/integration` and `test/e2e`, and no file was deleted

### TC03: Tagged packages compile

- **When** `go vet -tags=integration ./...`, `go vet -tags=e2e ./test/e2e/` and
  `go test -tags=integration -run '^$' ./test/integration/`
- **Then** no findings, the integration package builds

### TC04: CI Integration Tests job package list (AC3)

- **When** the job's `go list -tags=integration ./... | grep -v '/test/integration$'` runs
- **Then** `pkg/docker` is included, `test/integration` is not

### TC05: `pkg/docker` integration suite without Docker

- **When** `go test -tags=integration -v ./pkg/docker/` on a host without a Docker socket
- **Then** the suite is skipped, no panic, exit 0

### TC06: Root package tests

- **When** `go test -count=1 -run 'TestMainWithoutArguments|TestMCPServe' .`
- **Then** `TestMainWithoutArguments`, `TestMCPServeCommand`, `TestMCPServerCommandRemoved`
  pass

### TC07: CI lint parity

- **When** `gofmt -s -l .` (same as the lint job) and `go mod tidy -diff`
- **Then** no unformatted files, `go.mod` is tidy

### TC08: Workflow file

- **When** `.github/workflows/test.yml` is parsed as YAML
- **Then** it parses and the Integration Tests step contains the filtered package list

### TC09: Integration tests still run locally with the tag (regression)

- **When** `go test -tags=integration -run 'TestIssue050|TestIssue054' ./test/integration/`
- **Then** the selected tests run and pass

## Test Results — Round 1 (`19867b4`)

### Container — `golang:1.26`

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `go1.26.10`, `no container runtime`, `BUILD_OK`; `ok` for `portunix.ai/portunix`, `pkg/docker`, `pkg/logging`, `src/helpers/ptx-make/cmd`, `src/helpers/ptx-pft`, `src/pkg/archive`, `src/pkg/installconfig`, `src/pkg/platform`, `test/unit`; `TEST_EXIT=0` |

### Host

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC02 | ✅ PASS | untagged: `0 / 52` in `test/integration`, `0 / 4` in `test/e2e`; `go list -tags=unit`: `build constraints exclude all Go files` for both packages |
| TC03 | ✅ PASS | `VET_INTEGRATION_OK`, `VET_E2E_OK`; `ok portunix.ai/portunix/test/integration [no tests to run]` |
| TC04 | ✅ PASS | 24 packages; `pkg/docker` present, `test/integration` absent |
| TC05 | ✅ PASS | `--- SKIP: TestDockerIntegrationSuite`, `ok portunix.ai/portunix/pkg/docker` |
| TC06 | ✅ PASS | all three tests `--- PASS` |
| TC07 | ✅ PASS | `unformatted: 0`; `TIDY_OK` |
| TC08 | ✅ PASS | YAML parses; step runs `go test -tags=integration … $(go list -tags=integration ./... \| grep -v '/test/integration$')` |
| TC09 | ✅ PASS | `TestIssue050MultiLevelHelp`, `TestIssue050HelpPerformance`, `TestIssue054GUIDGeneration` PASS |

### Baseline (`main`)

On `main`, 48 of 52 files in `test/integration` and 3 of 4 in `test/e2e` have no build tag,
so the unit run picks them up; GitHub PR #223 shows 29 failing tests.
With `-tags=integration` the `test/integration` and `pkg/docker` packages do not compile on
`main`, and with `-tags=e2e` neither does `test/e2e` (errors fixed on the branch, see Scope
Note item 2).

### Skipped

- AC4 — Unit Tests job on the next release PR
- `pkg/docker` against a real Docker daemon (CI Integration Tests job)
- Windows and macOS execution

## Observations (Not Blocking)

- **O1**: the installed `portunix` (v2.6.2) passes `--` to the runtime as the command, so
  `… golang:1.26 -- bash -c …` failed with `exec: "--": executable file not found in $PATH`
  on Podman. This is the known #235, fixed in v2.6.3 (`c6b7453`); the binary built from the
  branch runs `container run alpine:latest -- echo …` correctly. No new issue needed; TC01
  was re-run without `--`
- The `test/integration` tests are now local-only. Many of them need a built binary,
  plugins, Hugo, Node.js or a VM and are expected to fail on machines without those tools
- The failing container tests in the separate `portunix.ai/cmd` module (see
  [acceptance-231.md](acceptance-231.md)) are outside the root `./...` run and outside the
  scope of this issue

## Final Decision

**STATUS**: PASS (AC1–AC3 verified locally; AC4 must be confirmed on the next release PR)

**Approval for merge**: YES
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role, Linux) on behalf of zdendaku
