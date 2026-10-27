# Issue #242: aiops ollama container status misreports GPU under Podman and assumes port 11434

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Bug
**Component**: ptx-aiops
**Created**: 2026-10-26
**Completed**: 2026-10-26
**GitHub**: #242
**Related**: #193, #101

## Summary

`portunix aiops ollama container status` reports container details that do not come from the
container itself:

1. **GPU** — a Podman container started with `--device nvidia.com/gpu=all` is shown as
   `GPU: Disabled`. This affects external containers and also the managed `portunix-ollama`,
   which `container create` starts with exactly this flag under Podman.
2. **Port / API endpoint** — `API Port` is always `11434` and the API check always calls
   `http://localhost:11434`, even when the container publishes a different host port.

Found while testing #193 (`docs/testing/acceptance-193.md`, observations O1 and O2). Both
behaviours were there before #193, but #193 makes them more visible because `status` now also
reports external `ollama` containers, which Portunix did not create.

## Problem Description

### GPU detection

`inspectOllamaContainer()` in `src/helpers/ptx-aiops/ollama.go`:

```go
cmd = exec.Command(runtime, "inspect", name,
    "--format", "{{.HostConfig.DeviceRequests}}")
output, err = cmd.Output()
if err == nil {
    status.GPUEnabled = strings.Contains(string(output), "gpu")
}
```

- Docker: `--gpus all` creates a device request with the `gpu` capability, so the substring
  check works.
- Podman: `--device nvidia.com/gpu=all` is a CDI device. Podman resolves it into host device
  paths, so `HostConfig.DeviceRequests` holds entries such as
  `{"PathOnHost":"/dev/dri/renderD129", ...}`. Nothing contains `gpu`, so the result is
  `Disabled`.

Reproduction (Podman, NVIDIA GPU with CDI configured):

```bash
podman run -d --name ollama --device nvidia.com/gpu=all -p 11434:11434 docker.io/ollama/ollama
portunix aiops ollama container status
# GPU:         Disabled     <- expected: Enabled
```

### Port and API endpoint

- `inspectOllamaContainer()` sets `status.Port = OllamaPort` (constant `"11434"`) and does
  not read the port binding.
- `handleOllamaContainerStatus()`, `handleOllamaContainerStart()` and the create wait loop
  check `OllamaAPIEndpoint` (`http://localhost:11434`).

For an external container started with, for example, `-p 21434:11434`, `status` prints
`API Port: 11434` and `API Status: ⚠ Not responding`, although the API works on
`localhost:21434`. If another service listens on 11434, the check may even report the wrong
service as available.

## Acceptance Criteria

1. `status` shows `GPU: Enabled` for a Podman container started with
   `--device nvidia.com/gpu=all`, for both the managed `portunix-ollama` and an external
   `ollama` container.
2. Docker detection is unchanged: `--gpus all` → `Enabled`; no GPU → `Disabled`.
3. A container without a GPU is reported as `Disabled` under both runtimes. Non-GPU devices
   (for example a plain `--device /dev/fuse`) do not count as a GPU.
4. `API Port` is the host port published for container port `11434/tcp`, read from the
   container (`NetworkSettings.Ports` or `HostConfig.PortBindings`).
5. The API check in `status` and `start` uses the resolved host port. If no host port is
   published, `status` says so (for example `API Port: not published`) instead of checking
   `localhost:11434`.
6. For the conventional setup (`-p 11434:11434`) the output is unchanged.
7. The GPU and port parsing is covered by unit tests with sample `inspect` output from Docker
   and Podman, so no real GPU is needed in CI.

## Out of Scope

- The fixed `OllamaAPIEndpoint` in `model` commands, `stack.go` and `webui.go`. Those commands
  talk to the managed container, which `create` always publishes on 11434. A configurable
  endpoint for them is a separate topic.
- Changing the port that `container create` publishes.

## Implementation Notes

- GPU under Podman can be detected from the CDI device name in `Config.CreateCommand`
  (`nvidia.com/gpu=`) or from NVIDIA device paths (`/dev/nvidia*`) in
  `HostConfig.Devices` / `DeviceRequests`. `/dev/dri/*` alone also matches integrated GPUs,
  so it should not be the only signal.
- Separate the parsing from the `exec` call (for example `parseGPUEnabled(runtime, output)` and
  `parseHostPort(output)`), so it can be tested with sample output.

## Testing

- Unit tests for the parsers with sample Docker and Podman `inspect` JSON (GPU via `--gpus`,
  GPU via CDI, no GPU, non-GPU device, port `11434:11434`, `21434:11434`, no port published).
- Integration test (`test/integration/`, `testframework`, isolated Podman storage as in
  `docs/testing/acceptance-193.md`): an `alpine` container named `ollama` with
  `-p 21434:11434` → `status` prints `API Port: 21434`.
- Manual GPU check on a host with an NVIDIA GPU under Podman, and under Docker where available.
