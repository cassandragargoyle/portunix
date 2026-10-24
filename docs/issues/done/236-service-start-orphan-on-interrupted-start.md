# Issue #236: service start: plugin process left running outside the state file when the CLI is killed during start

**Status**: Implemented
**Priority**: Medium
**Type**: Bug
**Component**: service orchestration (`src/app/service/orchestrator.go`)
**Created**: 2026-10-24
**Closed**: 2026-10-24
**GitHub**: #236
**Related**: `docs/testing/acceptance-236.md`, #159 (service orchestration command group), portunix-vscode #19 (acceptance
defect X1)

## Summary

When `portunix service start <plugin>` is killed after it spawned the plugin but before the
health check finished, the plugin process keeps running and has no entry in
`~/.portunix/processes.json`. `portunix service list` does not show it, and
`portunix service stop --all --force` cannot stop it. The port stays taken until the user
kills the process by hand.

## Problem Description

Clients call `portunix service start` with a timeout. The VS Code extension and Portunix
Pilot use `execFile` with a 30 s timeout and send SIGTERM when it expires. Found in the
acceptance test of portunix-vscode issue 019 with Portunix v2.6.2.

### Steps to reproduce

```text
$ timeout 0.05 portunix service start reco-studio --output json
$ portunix service list
No running service instances.
$ pgrep -fa "ptx-reco-studio --grpc"
1125488 …/ptx-reco-studio --grpc-port 50100
$ portunix service start reco-studio --output json   # picks 50101, 50100 is taken
```

The process on port 50100 stayed alive until it was killed by hand.

### Cause

`Orchestrator.StartService` runs, inside the state lock:

1. `spawnPlugin` — starts the plugin detached (`platformSysProcAttr`, `Process.Release`),
   so it survives the CLI
2. `waitForHealthy` — polls the health check for up to 10 s
3. writes the instance to the state only after the health check

If the CLI gets SIGTERM or SIGINT during step 2, nothing kills the spawned process and
nothing records it. `killProcess(pid)` runs only when the health check fails, not when the
CLI itself stops.

## Proposed Solution

Either of:

1. **Record before waiting** — write the instance with a `starting` state right after the
   spawn; on health failure remove it and kill the process. `cleanStaleEntries` already
   prunes entries whose process is gone, so a stopped CLI leaves a visible, stoppable entry
2. **Clean up on signal** — while `StartService` waits, catch SIGTERM/SIGINT, kill the
   spawned process and exit

Option 1 also covers SIGKILL of the CLI.

## Acceptance Criteria

- [x] After `timeout 0.05 portunix service start <plugin>` no plugin process runs outside
      `service list`
- [x] `portunix service stop --all --force` stops every process that `service start` spawned
- [x] A normal start and a failed health check behave as before

## Testing

- Unit test: interrupted start leaves no unrecorded process (fake spawner)
- Manual: the steps above with a Python plugin (slow start) and with SIGKILL of the CLI
