# Acceptance Protocol - Issue #193 (Detect pre-existing `ollama` container in aiops)

**Issue**: `portunix aiops ollama container status` — Detect Pre-existing `ollama` Container
**Branch**: `feature/issue-193-aiops-ollama-detect-existing-container` (commit `4e89d3b`, already
merged to `main` in `d51b44d`; tested on `main` at `58b30e6`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-26
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, static checks, read-only `status`
  against a real pre-existing `ollama` container; Go 1.26.8
- Alpine Linux (`alpine:latest`) containers in rootless Podman 5.7.0 with an isolated storage
  (`CONTAINERS_STORAGE_CONF` / `CONTAINERS_CONF_OVERRIDE` pointing to a temporary graphroot,
  runroot and tmp dir) — integration test and lifecycle scenarios

No software was installed. The host already has a real external `ollama` container
(`docker.io/ollama/ollama`, stopped), so the integration test would skip on the default storage
(safety guard). All destructive scenarios therefore ran in the isolated Podman storage, where
neither `ollama` nor `portunix-ollama` existed beforehand. After the test the real host
`ollama` container was verified untouched (`Exited (0) 3 months ago`).

## Scope Note

`src/helpers/ptx-aiops/ollama.go` resolves the container in precedence order
`portunix-ollama` → `ollama`, marks the result with `Managed`, lets `start` / `stop` / `remove`
fall back to the external container (with a note; `remove` asks `[y/N]`) and keeps `create`
on the managed name only (`getManagedOllamaContainerStatus`).

**Independence note**: the code was written by Claude Code in the developer role in an earlier
session; the test file was written together with the code.

**Not covered**:

- Docker — not installed on the host; ptx-aiops uses the same `inspect` / `start` / `stop` /
  `rm` CLI calls for both runtimes
- Windows and macOS — the package cross-compiles and vets, nothing was executed
- `create` with only an external `ollama` present — would pull the real `ollama/ollama` image
  and bind port 11434; verified by code review only (R2)

## Test Summary

- Total test scenarios: 18
- Passed: 18
- Failed: 0
- Skipped: 0
- Defects found: 0
- Observations: 2 (O1, O2)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make build`, `gofmt`, `go vet` (also `GOOS=windows` / `darwin`, `-tags integration`) | host |
| Integration | `TestIssue193_OllamaStatusExternalContainer` | isolated Podman |
| Manual | Lifecycle edge cases not in the integration test | isolated Podman |
| Real-world | `status` against the real external `ollama` container | host Podman, read-only |
| Review | `create` path, precedence logic | repository |

## Test Cases

### TC01: External `ollama` only (AC1, AC5, AC6)

- **Given** only a container named `ollama`
- **When** `aiops ollama container status` / `stop` / `start` / `remove` run
- **Then** the container is reported with `(external, not managed by Portunix)`, lifecycle
  commands act on it with a note, `remove` asks for confirmation, the footer offers only
  working commands

### TC02: Managed `portunix-ollama` present (AC2, AC4)

- **Given** `portunix-ollama` alone, or together with `ollama`
- **When** `status` runs
- **Then** `portunix-ollama` is reported without the external marker and with the original
  footer

### TC03: Neither present (AC3)

- **Given** no `ollama` and no `portunix-ollama`
- **When** `status` / `start` / `stop` / `remove` run
- **Then** `status` prints `Not found` with the create hint, lifecycle commands report the
  managed name as missing

## Test Results

### Static Checks

| ID | Check | Result |
| -- | ----- | ------ |
| S1 | `make build` | PASS |
| S2 | `gofmt -l` of `src/helpers/ptx-aiops` and the test file | PASS — no output |
| S3 | `go vet` of ptx-aiops, also `GOOS=windows` / `darwin` | PASS |
| S4 | `go vet -tags integration ./test/integration/` | PASS |

### Integration Test (isolated Podman)

`go test -tags integration ./test/integration/ -run TestIssue193 -v -count=1` —
**PASS** (8 steps, 49.1 s)

| ID | Step | Result |
| -- | ---- | ------ |
| I1 | TC003: neither present → `Not found` + create hint | PASS |
| I2 | TC001: external `ollama` running → external marker, `🟢 Running`, footer with `container stop` and `model list --container ollama` | PASS |
| I3 | TC001b: `stop` falls back to external, prints note, container stopped | PASS |
| I4 | Stopped external footer → `container start` | PASS |
| I5 | TC001c: `start` falls back to external, prints note, container running | PASS |
| I6 | TC001d: `remove` external, answer `n` → `Cancelled.`, container preserved | PASS |
| I7 | TC002: both present → `portunix-ollama` wins, no external marker | PASS |
| I8 | TC002b: managed only → unchanged managed output and footer | PASS |

### Manual Scenarios (isolated Podman)

| ID | Scenario | Result |
| -- | -------- | ------ |
| M1 | `remove` external, answer `y` | PASS — stopped and removed, no data-directory note |
| M2 | `remove` managed (stdin closed) | PASS — no prompt, removed, data-directory note printed |
| M3 | Neither present: `start` / `stop` / `remove` | PASS — `Container 'portunix-ollama' does not exist` (+ create hint for `start`) |
| M4 | `remove` external, empty input | PASS — default `N`, `Cancelled.`, container preserved |
| M5 | `create` with both present | PASS — refuses on the managed `portunix-ollama`, does not mention `ollama` |
| M6 | No leftovers after the integration test | PASS — isolated storage empty |

### Real-world Check (host Podman, read-only)

| ID | Scenario | Result |
| -- | -------- | ------ |
| W1 | `aiops ollama container status` with the real stopped `ollama` container | PASS — `ollama  (external, not managed by Portunix)`, `🔴 Stopped`, image `docker.io/ollama/ollama:latest`, footer `container start`; before #193 this printed `Not found` |

### Review

| ID | Item | Result |
| -- | ---- | ------ |
| R1 | Precedence | PASS — `ollamaContainerCandidates` is ordered managed first; `getOllamaContainerStatus` returns the first existing one |
| R2 | `create` | PASS — uses `getManagedOllamaContainerStatus` and the `OllamaContainerName` constant only; never adopts `ollama` |
| R3 | Data-directory note | PASS — printed only when `status.Managed` |

## Observations

- **O1 — GPU shown as `Disabled` for Podman containers with a GPU.** The real external
  `ollama` was created with `--device nvidia.com/gpu=all`, but `status` prints
  `GPU: Disabled`: detection reads `{{.HostConfig.DeviceRequests}}`, which is the Docker
  `--gpus` mechanism. The same applies to a managed container that `create` makes under Podman
  (it also uses `--device nvidia.com/gpu=all`). Pre-existing, not introduced by #193 —
  tracked in #242.
- **O2 — Port and API endpoint are fixed to 11434.** `status.Port` is the constant
  `OllamaPort` and the API check uses `http://localhost:11434`, also for an external container
  that may publish a different host port. Correct for the conventional setup (and for the real
  container on this host); tracked in #242.

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | External `ollama` only → real state, marked external | PASS (I2, I4, W1) |
| 2 | `portunix-ollama` present → output unchanged | PASS (I8) |
| 3 | Neither present → `Not found` + create hint | PASS (I1, M3) |
| 4 | Both present → `portunix-ollama` wins | PASS (I7, R1) |
| 5 | `start` / `stop` / `remove` fallback with note, `remove` confirmation, `create` unchanged | PASS (I3, I5, I6, M1, M2, M4, M5, R2) |
| 6 | External footer offers only valid commands | PASS (I2, I4, W1) |
| 7 | Verified with Docker and Podman | PARTIAL — Podman only, Docker not installed (see Not covered) |
| 8 | Integration test covers the three states | PASS — `test/integration/issue_193_aiops_ollama_status_test.go` |

## Final Decision

**STATUS**: PASS

AC7 is met for Podman only; the Docker code path issues the same CLI calls and is accepted
without execution. O1 and O2 are pre-existing limitations outside the scope of #193.

**Approval for merge**: YES (already merged in `d51b44d`)
**Date**: 2026-10-26
**Tester signature**: Claude Code (tester role, Linux)
