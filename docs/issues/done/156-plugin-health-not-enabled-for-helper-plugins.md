# Issue #156: Plugin Health Reports "Not Enabled" for All Helper Plugins

## Summary

`portunix plugin health <name>` returns `Error: failed to get plugin health: plugin <name> is not enabled` for all helper-type plugins, even though they are correctly installed, enabled, and functional. The health check appears to only work for gRPC/service-type plugins.

This is inconsistent — `plugin list` shows helper plugins as `ready`, `plugin enable` reports success, and the plugins work correctly via the dispatcher (`portunix <plugin> <command>`), yet `plugin health` treats them as not enabled.

## Reproduction

```bash
# Install and enable a helper plugin
portunix plugin install /path/to/dist/youtube/
portunix plugin enable youtube

# Plugin works correctly
portunix plugin list
# youtube   1.0.0   ready   YouTube transcript download...

portunix youtube --help
# (works, shows help)

# But health check fails
portunix plugin health youtube
# Error: failed to get plugin health: plugin youtube is not enabled
```

## Affected Plugins

All helper-type plugins exhibit this behavior:

| Plugin | Type | Status in `plugin list` | `plugin health` result |
| ------ | ---- | ----------------------- | ---------------------- |
| youtube | helper/native | ready | "not enabled" |
| vox | helper/native | ready | "not enabled" |
| fulltext | helper/java | ready | "not enabled" |
| docgen | grpc/native | stopped | "not enabled" |

## Expected Behavior

`plugin health` should either:

1. **Support helper plugins** — verify binary exists and is executable, check version output, report health status
2. **Return a clear message** — e.g. "health check is not available for helper-type plugins" instead of the misleading "not enabled" error

## Root Cause Hypothesis

The `plugin health` command likely checks health via gRPC health endpoint (standard gRPC health checking protocol). Helper plugins don't run a gRPC server — they are stateless CLI tools that execute and exit. The health check implementation probably:

1. Looks up the plugin in the registry
2. Tries to connect to a gRPC port
3. When no port is found (helper plugins don't have one), reports "not enabled"

## Suggested Fix

For helper-type plugins, `plugin health` should perform an alternative health check:

```text
1. Verify binary exists at expected path (~/.portunix/plugins/<name>/<binary>)
2. Verify binary is executable
3. Run `<binary> --version` and verify it returns successfully
4. Report: "Plugin <name> v<version> — healthy (helper mode)"
```

For gRPC plugins, keep the existing gRPC health check.

## Impact

- Users cannot verify helper plugin health through the standard command
- Misleading "not enabled" error message causes confusion
- Inconsistent behavior between `plugin list` (shows ready) and `plugin health` (says not enabled)

---

**Status**: ✅ Implemented (2026-03-15)
**Priority**: Low
**Type**: Bug
**Labels**: bug, plugin-system, cli, ux
**Milestone**: Backlog
