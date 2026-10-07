# Issue #182: Python Wheel Plugins Fail on Windows — Missing `.exe` Suffix in venv Path

## 🎯 Priority

**HIGH** — Python wheel plugins (introduced in #158, extended in #160 and #161)
are unusable on Windows because the plugin manager builds `.venv` executable
paths without the `.exe` suffix, breaking both wheel installation and health
checks.

## 📋 Status

- **Created**: 2026-04-27
- **Closed**: 2026-04-27
- **Status**: ✅ Implemented
- **Assignee**: -
- **Branch**: fix/windows-python-wheel-venv-exec (merged into main)
- **Related**: #158 (Python wheel plugin support), #160 (Python plugin gRPC),
  #161 (extra wheels), #138 (project-local venv)

## 📝 Problem Description

### Current Situation

On Windows, Python `venv` places executables in `.venv\Scripts\` with a `.exe`
suffix (e.g. `pip.exe`, `<plugin-binary>.exe`). The plugin manager already
selected the correct directory via `venvBinDir()` (`Scripts` on Windows,
`bin` elsewhere), but built the binary path without the `.exe` suffix, e.g.:

```text
C:\...\plugin\.venv\Scripts\my-plugin     # built path (wrong)
C:\...\plugin\.venv\Scripts\my-plugin.exe # actual file
```

Two consequences:

1. **Wheel installation fails** — `setupPythonWheelPlugin` invokes
   `filepath.Join(venvPath, venvBinDir(), "pip")` to call pip, but on Windows
   that path doesn't exist; only `pip.exe` does. `exec.Command` fails to spawn
   pip, so `pip install <wheel>` is never executed.
2. **Health check fails** — `checkHelperPluginHealth` `os.Stat`s the path
   without `.exe`, returning `binary_missing` even when the plugin is correctly
   installed. Worse, when the file *is* found, the executable-bit check
   (`info.Mode()&0111 == 0`) always fails on Windows because NTFS does not
   expose Unix permission bits, so the plugin is reported as `not_executable`.

### Expected Behavior

- `portunix plugin install <wheel-plugin>` succeeds on Windows: pip is
  invoked from `.venv\Scripts\pip.exe` and installs the wheel.
- `portunix plugin health <wheel-plugin>` reports `ready` on Windows when the
  binary is correctly installed in `.venv\Scripts\<name>.exe`.
- Linux/macOS behavior unchanged (no `.exe` suffix, executable-bit check still
  enforced).

## 🎯 Root Cause

Two locations hard-coded `filepath.Join(plugin.InstallPath, ".venv",
venvBinDir(), plugin.BinaryName)` and a third did the same for `pip`:

- `src/app/plugins/manager/manager.go:468` — `checkHelperPluginHealth`
- `src/app/plugins/manager/manager.go:617` — `setupPythonWheelPlugin` (pip path)
- `src/app/plugins/manager/registry.go:553` — `RegistryPlugin.BinaryPath`

None appended `.exe` on Windows. Additionally, the executable-bit check at
`manager.go:484` was unconditional, so it produced false negatives on Windows
even when the right path was used.

## ✅ Acceptance Criteria

1. A Python wheel plugin installs successfully on Windows via
   `portunix plugin install <plugin-dir>`, with pip invoked from
   `.venv\Scripts\pip.exe`.
2. `portunix plugin health <plugin>` returns healthy on Windows when the
   plugin's `.venv\Scripts\<name>.exe` exists.
3. `RegistryPlugin.BinaryPath()` returns a path ending in `.exe` on Windows
   for Python wheel plugins.
4. Linux/macOS behavior is unchanged (paths have no `.exe` suffix; executable
   bit is still enforced).
5. The fix is centralized in a single helper (`venvExecPath`) so future call
   sites cannot regress on the `.exe` suffix.

## 🛠️ Fix Summary

Introduced `venvExecPath(venvDir, name string) string` in
`src/app/plugins/manager/manager.go`:

```go
// venvExecPath returns the absolute path to a venv executable, appending
// the .exe suffix on Windows when missing
func venvExecPath(venvDir, name string) string {
    if goruntime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
        name += ".exe"
    }
    return filepath.Join(venvDir, venvBinDir(), name)
}
```

Replaced the three hard-coded joins with calls to `venvExecPath` and gated
the executable-bit check on `goruntime.GOOS != "windows"`.

## 🧪 Testing Strategy

- **Manual**: install a Python wheel plugin (e.g. one of the fixtures used by
  #158/#160) on Windows; verify both install and `plugin health` succeed.
- **Manual (Linux)**: regression — install the same plugin on Linux and verify
  paths still resolve to `.venv/bin/<name>` (no `.exe`) and the executable-bit
  check still rejects a `chmod -x`'d binary.
- **Container-based** (per `ISSUE-DEVELOPMENT-METHODOLOGY.md`): no good
  Windows container option; rely on local Windows + Linux container for the
  regression case.

## 📎 References

- `src/app/plugins/manager/manager.go:464-492` — `checkHelperPluginHealth`
- `src/app/plugins/manager/manager.go:608-648` — `setupPythonWheelPlugin`
- `src/app/plugins/manager/manager.go:670-677` — new `venvExecPath` helper
- `src/app/plugins/manager/registry.go:550-556` — `RegistryPlugin.BinaryPath`
- #158 (Python wheel plugin installation support — original implementation)
- #160 (Python plugin gRPC service support — same venv layout)
- #161 (Plugin install extra wheels support — also calls pip from venv)
