# Issue #235: container run: "--" separator from the help example is passed to the runtime as the command

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Bug
**Component**: ptx-container
**Created**: 2026-10-22
**Closed**: 2026-10-24
**GitHub**: #235
**Related**: `docs/testing/acceptance-235.md`, #233 (found during testing, `docs/testing/acceptance-233.md` defect D1)

## Summary

`portunix container run --help` documents the example
`portunix container run -d --name test ubuntu:22.04 -- bash -c "echo test"`. Running it fails,
because the `--` separator is forwarded to podman/docker after the image and the runtime tries
to execute `--` as the container command.

## Problem Description

```text
$ portunix container run -d --name t233 ubuntu:24.04 -- sleep infinity
💡 For software installation testing, use 'run-in-container' instead
Error: runc: runc create failed: unable to start container process: error during container init:
exec: "--": executable file not found in $PATH: OCI runtime attempted to invoke a command that was not found
❌ Podman run failed: exit status 127
```

The same command without `--` works.

`handleContainerRun` in `src/helpers/ptx-container/main.go` passes the arguments to
`runPodmanContainer` / `runDockerContainer` unchanged, which build
`podman run <args…>`. Podman and Docker treat everything after the image as the command, so a
`--` placed after the image is not a separator for them but the program name. The help text in
`showRunHelp` suggests the opposite.

The managed path (lifecycle flags, `runManagedContainer`) may have the same problem and must be
checked.

## Proposed Solution

1. Strip the first standalone `--` that follows the image before the arguments are handed to the
   runtime (both the direct and the managed run path); a `--` inside the command itself
   (after the first one) is kept
2. Keep the help example, since `--` is the documented way to separate a command that starts
   with `-` from Portunix flags
3. Add a unit test for the argument builder

## Acceptance Criteria

1. `portunix container run -d --name test ubuntu:22.04 -- bash -c "echo test"` starts the
   container and runs `bash -c "echo test"` with Podman and with Docker
2. The same command without `--` keeps working
3. `portunix container run ubuntu:22.04 -- sh -c 'echo -- x'` prints `-- x` (only the first
   separator is removed)
4. Unit tests cover the argument handling with and without `--`, for detached and
   interactive runs and for the managed (lifecycle flags) path

## Testing

- Run the help example and the variants above with Podman; repeat with Docker where available
- `--debug` shows the final runtime arguments: no `--` directly after the image
