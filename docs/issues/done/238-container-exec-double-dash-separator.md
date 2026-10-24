# Issue #238: container exec: "--" separator is passed to the runtime as the command

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Bug
**Component**: ptx-container
**Created**: 2026-10-24
**Closed**: 2026-10-24
**GitHub**: #238
**Related**: `docs/testing/acceptance-238.md`, #235 (same defect in `container run`, `docs/testing/acceptance-235.md` observation O4)

## Summary

`portunix container exec <name> -- <command>` fails, because the `--` separator is forwarded to
`podman exec` / `docker exec` and the runtime runs `--` as the program. In addition,
`container run` and `container exec` treat `-h` / `--help` anywhere in the arguments as a help
request, so a flag that belongs to the container command shows the Portunix help instead.

## Problem Description

### Separator in `container exec`

```text
$ portunix container exec t235e -- echo x
Error: runc: exec failed: unable to start container process: exec: "--": executable file not found in $PATH:
OCI runtime attempted to invoke a command that was not found
❌ Error: Failed to execute command in container 't235e': exit status 127
```

`handleContainerExec` in `src/helpers/ptx-container/main.go` takes `args[0]` as the container
name and passes `args[1:]` unchanged to `execPodmanCommand` / `execDockerCommand`, which build
`<runtime> exec -i[t] <name> <args…>`. Podman and Docker treat everything after the container
name as the command, so `--` is the program name.

`container exec --help` does not document `--`, but users expect it to work the same way as in
`container run` after #235, and `--` is the usual way to separate a command that starts with
`-` from Portunix flags.

### Help flag inside the container command

Both `handleContainerRun` and `handleContainerExec` scan all arguments for `-h` / `--help`
before doing anything else:

```text
$ portunix container run ubuntu:22.04 -- ls -h /
Usage: portunix container run [flags] <image> [command...]
...
```

The same happens for `portunix container exec <name> ls -h /`, with or without `--`.

## Proposed Solution

1. In `handleContainerExec`, remove the first standalone `--` (reuse `stripCommandSeparator`
   from #235); accept it both before the container name and between the name and the command
2. Stop the help flag scan at the first `--` in `handleContainerRun` and `handleContainerExec`;
   for `exec`, treat `-h` / `--help` as a help request only when it comes before the container
   name
3. Document `--` in `showExecHelp` with an example
4. Extract the exec argument building into a helper and cover it with unit tests

## Acceptance Criteria

1. `portunix container exec <name> -- echo x` prints `x` with Podman and with Docker
2. `portunix container exec <name> echo x` keeps working
3. `portunix container exec <name> -- sh -c 'echo -- x'` prints `-- x` (only the first
   separator is removed)
4. `portunix container run ubuntu:22.04 -- ls -h /` and `portunix container exec <name> ls -h /`
   run the command instead of showing the help
5. `portunix container run --help` and `portunix container exec --help` still show the help
6. Unit tests cover the exec argument handling with and without `--` and the help flag
   detection for `run` and `exec`

## Testing

- Start a test container with `portunix container run -d --name <name> ubuntu:22.04 sleep 300`
  and run the variants above with Podman; repeat with Docker where available
- Check the exit code of a failing command (`-- sh -c 'exit 3'`) is non-zero
- Remove the test container afterwards
