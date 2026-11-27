# Issue #252: container exec masks the command exit code and re-runs failed commands via Docker

**Status**: Open
**Priority**: High
**Type**: Bug
**Component**: ptx-container
**Created**: 2026-11-08
**GitHub**: #252

## Summary

`portunix container exec` does not propagate the exit code of the executed
command. A failing command in a Podman container is additionally executed a
second time through Docker when Docker is installed.

## Problem Description

Reproduction (Podman 5.7.0, `ubuntu:24.04`):

```bash
portunix container exec <name> -- bash -c 'exit 7'
echo $?   # prints 1, expected 7
```

Scripts can only distinguish success from failure, not the specific exit code.

Root cause in `handleContainerExec` (`src/helpers/ptx-container/main.go`):

1. Every error returned by `execPodmanCommand` / `execDockerCommand` is mapped
   to `os.Exit(1)`; the `*exec.ExitError` status is discarded.
2. When `execPodmanCommand` returns an error, the handler falls back to
   `execDockerCommand` if Docker is available. A non-zero exit of the *command*
   is indistinguishable from a failure of *Podman itself*, so the command is
   run again through Docker (usually failing with "no such container", but a
   non-idempotent command may have already had side effects).

The main dispatcher already forwards helper exit codes (`main.go`), so the fix
is limited to `ptx-container`.

## Acceptance Criteria

1. `portunix container exec <name> -- bash -c 'exit 7'` exits with code 7 on
   Podman and on Docker
2. A non-zero exit code of the executed command never triggers the Docker
   fallback; the fallback only applies when the Podman runtime cannot run at
   all (binary missing / not startable)
3. Runtime-level failures (container not found, runtime error) keep a clear
   error message on stderr; command output is not altered
4. Unit tests cover exit code propagation and the fallback decision

## Testing

- Unit test for the fallback decision (`*exec.ExitError` vs. other errors)
- Integration test with a Podman container: `exit 0`, `exit 7`, `exit 255`
- Verify that a command writing a marker file runs exactly once on failure
