# Issue #263: plugin health reports interpreter-run plugin files as not executable

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Bug
**Component**: portunix
**Created**: 2026-11-23
**Closed**: 2026-11-23
**GitHub**: #263
**Related**: #257, #260

## Summary

On Linux and macOS, `portunix plugin health` checks the executable bit of the plugin
binary for every runtime except `node`. A helper plugin with `runtime: java` (a JAR)
or `runtime: python` without `wheel` (a script) is run through `java -jar` or
`python3`, so it needs no executable bit, but a file without it is reported as
unhealthy. Found while testing #260, which exempted only the `node` runtime.

## Problem Description

`checkHelperPluginHealth` in `src/app/plugins/manager/manager.go`:

```go
// Check binary is executable (Unix only — Windows uses extension-based execution)
if goruntime.GOOS != "windows" && plugin.Runtime != "node" && info.Mode()&0111 == 0 {
```

Reproduction (`ubuntu:24.04`, Python 3.12.3):

```text
$ ls -l ~/.portunix/plugins/fx-python/main.py
-rw-r--r-- ... main.py
$ portunix fx-python hello x
python args=['hello', 'x'] ...            # the plugin runs
$ portunix plugin health fx-python
Status:     ❌ Unhealthy
Message:    Binary is not executable: /root/.portunix/plugins/fx-python/main.py
```

After `chmod +x main.py` the same plugin is reported healthy. Maven and most build
tools produce JARs without the executable bit, so Java helper plugins hit the same
false report.

The check is right where the file itself is executed:

| Runtime | Started as | Executable bit needed |
| ------- | ---------- | --------------------- |
| `native` | `<binary>` | yes |
| `python` + `wheel` | venv entry point | yes |
| `python` (script) | `python3 <binary>` | no |
| `java` | `java -jar <binary>` | no |
| `node` | `node <binary>` | no (exempted in #260) |

## Proposed Solution

1. Check the executable bit only when the binary is executed directly: `native`
   and `python` with `wheel`
2. Keep the existing behavior for those two cases
3. A missing interpreter is already reported as `runtime_missing` (#260), so the
   health check stays meaningful for interpreter-run plugins

## Acceptance Criteria

1. `plugin health` reports a `java` helper plugin with a non-executable JAR as
   healthy when `java` is available
2. `plugin health` reports a `python` script helper plugin with a non-executable
   script as healthy when `python3` is available
3. A `native` plugin or a wheel entry point without the executable bit is still
   reported as `not_executable`
4. `node` plugins are unchanged
5. Unit tests cover the executable-bit rule per runtime

## Testing

- Unit: `checkHelperPluginHealth` with a non-executable file for `native`,
  `python` (script and wheel), `java` and `node`, using stub interpreters in `PATH`
  (POSIX only)
- Container (`ubuntu:24.04`): Python script helper plugin with a `0644` script →
  healthy; native helper plugin with a `0644` binary → `not_executable`
