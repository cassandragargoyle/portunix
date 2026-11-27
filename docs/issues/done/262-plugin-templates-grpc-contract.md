# Issue #262: Go, Java and Rust plugin templates do not implement the Portunix plugin contract

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Bug
**Component**: portunix
**Created**: 2026-11-23
**Closed**: 2026-11-24
**GitHub**: #262
**Related**: #7, #257, #260

## Summary

The language templates in `docs/plugin-development/languages/{go,java,rust}/template`
define their own gRPC service instead of `portunix.plugins.PluginService` from
`src/app/plugins/proto/plugin.proto`. Portunix starts a service plugin and calls
`Initialize`, which the templates do not implement, so `portunix plugin start` fails.
The same defect in the JavaScript template was found and fixed in #260.

## Problem Description

`portunix plugin start` launches the plugin with `--port <port>`, connects to
`localhost:<port>` and calls `Initialize`, then `Health`, `Execute` and `Shutdown`
(`src/app/plugins/grpc_client.go`). The templates declare a different package and a
different set of RPCs:

| Template | Proto package | RPCs |
| -------- | ------------- | ---- |
| Portunix contract | `portunix.plugins` | `Initialize`, `Execute`, `GetInfo`, `Health`, `Shutdown`, `ListCommands` |
| Go | `go_plugin_template` | `GetInfo`, `HealthCheck`, `Shutdown`, `Execute`, `ListTools`, `CallTool` |
| Java | `ai.portunix.plugin.proto` | same as Go |
| Rust | `rust_plugin_template` | same as Go |

A plugin built from these templates fails on start with:

```text
Error: failed to start plugin: failed to initialize plugin: rpc error: code = Unimplemented
desc = The server does not implement the method /portunix.plugins.PluginService/Initialize
```

The message types differ as well (`HealthCheckResponse.status` enum vs.
`HealthResponse.healthy`, `ExecuteResponse.result` vs. `output`, plugin info without
`commands`), so renaming the service is not enough.

CLI dispatch (`portunix <plugin> <command>`) does not use gRPC and is not affected.

## Proposed Solution

1. Replace each template proto with a copy of the `PluginService` part of
   `src/app/plugins/proto/plugin.proto` (package `portunix.plugins`), as done for the
   JavaScript template in #260; keep the language-specific options
   (`go_package`, `java_package`, ...) where code generation needs them
2. Rewrite the handlers to implement `Initialize`, `GetInfo`, `Health`, `Execute`,
   `ListCommands` and `Shutdown`; take plugin info and commands from `plugin.json`
3. `Shutdown` only cleans up and answers; Portunix ends the process right after it
   (SIGTERM), and exiting inside `Shutdown` drops the response (`plugin stop` then
   fails with `EOF`)
4. Bind the server to `127.0.0.1` only
5. Extend the guard test `TestJavaScriptTemplateProtoMatchesContract`
   (`src/app/plugins/manifest_node_test.go`) to all templates, so a template proto
   cannot drift from the contract again
6. Update the language guides (`getting-started.md`) where they show the old RPCs
   (`HealthCheck`, `ListTools`, `CallTool`)

## Acceptance Criteria

1. The proto of every template declares `portunix.plugins.PluginService` with the
   same messages as the Portunix contract; the guard test covers all templates
2. For each template: build, `portunix plugin install`, `plugin enable`,
   `plugin start`, `plugin health` (healthy), `plugin stop` succeed and no plugin
   process is left
3. `portunix <plugin> hello` keeps working
4. Template unit tests pass (`go test`, `mvn test`, `cargo test`)

## Testing

- Unit: guard test for all template protos; template unit tests per language
- Container (`ubuntu:24.04`): install the toolchain with `portunix install`
  (`go`, `java-21` + `maven`, Rust), build each template, run the service lifecycle
  from acceptance criterion 2
- Regression: JavaScript template service lifecycle (#260)
