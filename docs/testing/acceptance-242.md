# Acceptance Protocol - Issue #242 (aiops ollama status: GPU under Podman and published port)

**Issue**: `aiops ollama container status` misreports GPU under Podman and assumes port 11434
**Branch**: `fix/242-aiops-ollama-status-gpu-port` (commit `c3e1451`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-26
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, static checks, unit tests, read-only
  `status` against a real pre-existing `ollama` container; Go 1.26.8
- Alpine Linux (`alpine:latest`) containers in rootless Podman 5.7.0 with an isolated storage
  (`CONTAINERS_STORAGE_CONF` / `CONTAINERS_CONF_OVERRIDE` pointing to a temporary graphroot,
  runroot and tmp dir) — integration tests and manual scenarios

No software was installed and no host configuration was changed. All containers ran in the
isolated Podman storage, which was empty before and after the test. The real host `ollama`
container was only inspected and was verified untouched afterwards (`Exited (0) 3 months ago`).

## Scope Note

`inspectOllamaContainer()` in `src/helpers/ptx-aiops/ollama.go` now reads the full `inspect`
JSON and passes it to `parseGPUEnabled()` and `parseHostPort()`. `status` and `start` use the
resolved `APIEndpoint`; `create` and the `model` / `stack` / `webui` commands keep the fixed
`OllamaAPIEndpoint` (out of scope of the issue).

**Independence note**: the code and the tests were written by Claude Code in the developer role
in the same session. The manual scenarios below were designed separately in the tester role.

**Not covered**:

- Docker — not installed on the host; Docker `inspect` samples (`--gpus all`, CDI, no GPU,
  `/dev/fuse`, port mappings) are covered by unit tests only
- A *running* container with a GPU — the host CDI spec is stale (see O1), so a container with
  `--device nvidia.com/gpu=all` or `--gpus all` is created but cannot start. Detection was
  verified on the created containers (via `Config.CreateCommand`); the `/dev/nvidia*` device
  path signal of a running container is covered by unit tests only
- Windows and macOS — the package cross-compiles and vets, nothing was executed

## Test Summary

- Total test scenarios: 25
- Passed: 25
- Failed: 0
- Skipped: 0
- Defects found: 0
- Observations: 2 (O1, O2)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make build`, `gofmt`, `go vet` (also `GOOS=windows` / `darwin`, `-tags integration`) | host |
| Unit | `TestParseGPUEnabled`, `TestParseHostPort` (Docker and Podman `inspect` samples) | host |
| Integration | `TestIssue242_OllamaStatusPublishedPort`, regression `TestIssue193_OllamaStatusExternalContainer` | isolated Podman |
| Manual | GPU flags, non-GPU devices, port mappings, real API stub, `start`, foreign service on 11434 | isolated Podman |
| Real-world | `status` against the real external `ollama` container | host Podman, read-only |

## Test Cases

### TC01: GPU detection (AC1, AC2, AC3)

- **Given** a container started with `--device nvidia.com/gpu=all`, `--gpus all`, a non-GPU
  device (`/dev/fuse`, `/dev/dri`) or no device
- **When** `aiops ollama container status` runs
- **Then** `GPU: Enabled` is shown only for the GPU flags, for both the managed
  `portunix-ollama` and an external `ollama`

### TC02: Published API port (AC4, AC6)

- **Given** a container publishing `11434/tcp` as `11434:11434`, `21434:11434`, a random host
  port, `127.0.0.1:21434:11434`, or not at all; running or stopped
- **When** `status` runs
- **Then** `API Port` is the published host port, or `not published`

### TC03: API check uses the resolved endpoint (AC5)

- **Given** a container whose API is served on the published host port, or a foreign service
  listening on host port 11434
- **When** `status` / `start` run
- **Then** the check targets the resolved endpoint, is skipped when no port is published, and
  never reports the foreign service on 11434 as the Ollama API

## Test Results

### Static Checks

| ID | Check | Result |
| -- | ----- | ------ |
| S1 | `make build` | PASS |
| S2 | `gofmt -l` of `src/helpers/ptx-aiops` and the integration test | PASS — no output |
| S3 | `go vet` of ptx-aiops, also `GOOS=windows` / `darwin` | PASS |
| S4 | `go vet -tags integration ./test/integration/` | PASS |

### Unit Tests

`go test ./src/helpers/ptx-aiops/ -count=1 -v` — **PASS**

| ID | Test | Result |
| -- | ---- | ------ |
| U1 | `TestParseGPUEnabled` — 12 cases: Docker `--gpus`, Docker CDI, no GPU, `/dev/fuse`; Podman CDI (`--device X` and `--device=X`), `--gpus`, `/dev/nvidia*` nodes, `/dev/dri` only, no GPU; invalid JSON, empty array | PASS |
| U2 | `TestParseHostPort` — 10 cases: `11434:11434` (Docker, Podman), `21434:11434`, not published, no port data, stopped with random port, stopped bound to `127.0.0.1`, IPv6 `::1`, other port only, invalid JSON | PASS |

### Integration Tests (isolated Podman)

`go test -tags integration ./test/integration/ -run 'TestIssue242|TestIssue193' -v -count=1` —
**PASS** (54.2 s)

| ID | Step | Result |
| -- | ---- | ------ |
| I1 | TC001: external `ollama` with `-p 21434:11434` → `API Port: 21434`, no `localhost:11434` | PASS |
| I2 | TC002: no published port → `API Port: not published`, API check skipped | PASS |
| I3 | TC003: container without GPU → `GPU: Disabled` | PASS |
| I4 | Regression: `TestIssue193_OllamaStatusExternalContainer` (8 steps) | PASS |

### Manual Scenarios (isolated Podman)

The API was simulated by an `nc` loop in the `alpine` container answering on port 11434.

| ID | Scenario | Result |
| -- | -------- | ------ |
| M1 | External `ollama`, `--device nvidia.com/gpu=all` (created, see O1) | PASS — `GPU: Enabled` |
| M2 | Managed `portunix-ollama`, `--device nvidia.com/gpu=all` (created, see O1) | PASS — `GPU: Enabled` |
| M3 | External `ollama`, `--gpus all` (created, see O1) | PASS — `GPU: Enabled` |
| M4 | External `ollama`, `--device /dev/fuse`, running | PASS — `GPU: Disabled` |
| M5 | External `ollama`, `--device /dev/dri`, running | PASS — `GPU: Disabled` |
| M6 | `-p 11434:11434`, API served | PASS — `API Port: 11434`, `✓ Available at http://localhost:11434` (same output as before) |
| M7 | `-p 11434` (random host port), API served | PASS — `API Port: 33865`, `✓ Available at http://localhost:33865`, matches `podman port` |
| M8 | `-p 127.0.0.1:21434:11434`, API served | PASS — `API Port: 21434`, `✓ Available at http://127.0.0.1:21434` |
| M9 | Stopped container, `-p 21434:11434` | PASS — `API Port: 21434` (from `HostConfig.PortBindings`) |
| M10 | `container start` of M9, API served | PASS — `Waiting for Ollama API.... Ready!` on port 21434 |
| M11 | `container start`, no published port | PASS — `⚠ Container port 11434 is not published to the host, skipping API check`, no wait loop |
| M12 | Foreign HTTP service on host `127.0.0.1:11434`, container `-p 21434:11434` without API | PASS — `⚠ Not responding` (before the fix the foreign service would be reported as available) |
| M13 | Foreign service on 11434, container without published port | PASS — `API Port: not published`, `⚠ Not checked …` |
| M14 | No leftovers after the scenarios | PASS — isolated storage empty |

### Real-world Check (host Podman, read-only)

| ID | Scenario | Result |
| -- | -------- | ------ |
| W1 | `status` with the real stopped `ollama` (`--device nvidia.com/gpu=all`, `-p 11434:11434`) | PASS — `GPU: Enabled`, `API Port: 11434`; before #242 this printed `GPU: Disabled` (#193 observation O1) |

## Observations

- **O1 — Stale CDI spec on the test host.** `podman run --device nvidia.com/gpu=all` fails with
  `failed to stat CDI host device "/dev/nvidia-modeset"`: the generated CDI spec references a
  device node that no longer exists (regenerating it with `nvidia-ctk cdi generate` would fix
  it). Host environment issue, unrelated to #242; the host configuration was not changed. It
  limits the GPU scenarios to created, not running, containers.
- **O2 — Non-GPU signal in a container command could match.** `parseGPUEnabled()` treats every
  `Config.CreateCommand` argument of the form `<vendor>/gpu=<id>` as a CDI GPU device, including
  arguments of the container command itself (for example `podman run img foo/gpu=1`). Such
  arguments are unrealistic for an Ollama container; accepted as is.

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | Podman `--device nvidia.com/gpu=all` → `GPU: Enabled`, managed and external | PASS (M1, M2, W1, U1) |
| 2 | Docker detection unchanged: `--gpus all` → `Enabled`, no GPU → `Disabled` | PASS (U1) — Docker not installed, unit tests with Docker samples only |
| 3 | No GPU → `Disabled` under both runtimes; non-GPU devices do not count | PASS (M4, M5, I3, U1) |
| 4 | `API Port` is the published host port for `11434/tcp` | PASS (M6–M9, I1, U2) |
| 5 | API check in `status` and `start` uses the resolved port; `not published` when absent | PASS (M6–M8, M10–M13, I1, I2) |
| 6 | Conventional `-p 11434:11434` output unchanged | PASS (M6, W1) |
| 7 | Parsing covered by unit tests with Docker and Podman samples | PASS (U1, U2) |

## Final Decision

**STATUS**: PASS

AC2 is verified by unit tests only (Docker not installed). Running GPU containers could not be
started on this host because of O1; detection was verified on created containers and by unit
tests. Neither limitation blocks the merge.

**Approval for merge**: YES
**Date**: 2026-10-26
**Tester signature**: Claude Code (tester role, Linux)
