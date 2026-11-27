# Issue #258: Plugin dispatcher: execute service plugin commands via gRPC

**Status**: Open
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix
**Created**: 2026-11-20
**GitHub**: #258
**Related**: #133, #7, #143, #160

## Summary

Phase 3 of the Plugin Command Dispatcher (#133), split out when #133 was archived.
`portunix <plugin> <command> [args]` should route commands of service plugins
(`mode: "service"`) through the gRPC `Execute` method of the running plugin
instead of spawning a new process.

## Problem Description

`src/cmd/plugin_dispatcher.go` (`executePlugin` → `buildPluginCommand`) always
starts the plugin binary as a child process, regardless of the plugin `mode`.
For helper plugins this is correct. For service plugins it is not:

- a second instance of the service is started instead of talking to the one
  started by `portunix plugin start`
- state held by the running service (open connections, caches, indexes) is not used
- a service binary may not accept CLI arguments at all and may just start serving

The building blocks already exist:

- `rpc Execute(ExecuteRequest) returns (ExecuteResponse)` in
  `src/app/plugins/proto/plugin.proto`
- `GRPCPlugin.Execute` in `src/app/plugins/grpc_client.go`
- `Manager.ExecuteCommand(pluginName, command, args, options)` in
  `src/app/plugins/manager/manager.go` (checks enabled + running state)

The dispatcher just does not use them.

## Proposed Solution

1. In `executePlugin`, branch on the plugin mode:
   - `helper` → current process-based execution (unchanged)
   - `service` (also the default when `mode` is empty, see `getPluginMode`) →
     gRPC dispatch via `Manager.ExecuteCommand`
2. Map CLI arguments to `ExecuteRequest`: first positional argument is
   `Command`, the rest are `Args`; decide whether `--key=value` flags go to
   `Options` or stay in `Args` (document the choice)
3. Print `ExecuteResponse` output to stdout, error text to stderr, and exit with
   the response exit code (non-zero on failure)
4. When the service plugin is not running, fail with a clear message and a hint:
   `plugin '<name>' is not running. Start it with: portunix plugin start <name>`
5. `portunix <plugin> --help` for a service plugin should list commands from
   `ListCommands` (running plugin) or from the registry/`plugin.json` (stopped plugin)
6. Review the default for an empty `mode`: today the help text shows `service`
   but the plugin is executed as a helper. Make the two consistent without
   breaking existing installed plugins

## Acceptance Criteria

1. A running service plugin receives `portunix <plugin> <command> [args]` through
   gRPC `Execute`; no new plugin process is spawned
2. Output and exit code of the command reach the user
3. A stopped service plugin fails with a clear error and the `plugin start` hint
4. A disabled plugin still fails with the existing `plugin enable` hint
5. Helper plugins (native, Java, Python, wheel) keep working unchanged
6. `portunix <plugin> --help` lists the commands of a service plugin
7. Unit tests cover the mode branching and the argument mapping

## Testing

- Unit: mode branching and `ExecuteRequest` mapping in `src/cmd`
- Integration: reuse the lifecycle test plugin from
  `src/app/plugins/manager/lifecycle_test.go` (supports `ping` via `Execute`);
  start it, dispatch a command through the CLI, verify no extra process
- Regression: a helper plugin (e.g. a Java or Python helper) still runs via process
