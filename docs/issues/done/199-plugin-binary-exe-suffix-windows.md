# Issue #199: Plugin Binary Path Missing `.exe` Suffix on Windows

## Priority

**MEDIUM** — `portunix plugin health <name>` reports every native helper
plugin as unhealthy on Windows, although the plugin is installed correctly and
runs fine. Misleading diagnostics; any future code that relies on `os.Stat` of
the plugin binary path inherits the same bug.

## Status

- **Created**: 2026-09-28
- **Status**: Implemented
- **Closed**: 2026-09-28
- **Assignee**: zdendaku
- **Branch**: fix/199-plugin-binary-exe-suffix (merged)
- **Related**:
  - ADR-026 — Shared Platform Utilities (`src/pkg/platform`)
  - `src/app/plugins/manager/manager.go`
  - `src/app/plugins/manager/registry.go`
  - `src/cmd/plugin_dispatcher.go`
  - Acceptance protocol: `docs/testing/acceptance-199.md` (PASS)

## Problem Description

### Current Situation

The plugin manifest declares the binary name without an OS-specific
extension (`"binary": "ptx-modeler"`), while the installed file on Windows is
`ptx-modeler.exe`. The binary path is built as a plain join:

```go
// src/app/plugins/manager/manager.go — checkHelperPluginHealth
binaryPath = filepath.Join(plugin.InstallPath, plugin.BinaryName)

info, err := os.Stat(binaryPath)   // fails on Windows: file is ptx-modeler.exe
```

Plugin **execution** works only by accident: `exec.Command` on Windows calls
`exec.LookPath`, which tries `PATHEXT` extensions even for absolute paths.
`os.Stat` does not, so the health check fails.

### Reproduction

```powershell
portunix plugin install modeler
Get-ChildItem $HOME\.portunix\plugins\modeler   # ptx-modeler.exe exists
portunix plugin health modeler
# Status:     ❌ Unhealthy
# Message:    Binary not found: C:\Users\<user>\.portunix\plugins\modeler\ptx-modeler
```

### Affected Code

| Location | Code | Effect |
| -------- | ---- | ------ |
| `src/app/plugins/manager/manager.go:480` | `filepath.Join(InstallPath, BinaryName)` | health check `os.Stat` fails |
| `src/app/plugins/manager/manager.go:263` | same join for gRPC `BinaryPath` | relies on `LookPath` fallback |
| `src/app/plugins/manager/registry.go:555` | `RegistryPlugin.BinaryPath()` | used by `cmd/service.go`, same issue |
| `src/cmd/plugin_dispatcher.go:127,143` | own join, own venv path | duplicates logic, no `.exe` |
| `src/app/plugins/manager/manager.go:771` | `venvExecPath` — private `.exe` handling | correct, but ad-hoc `GOOS` check |

## Proposed Solution

Use the existing shared OS detection (`src/pkg/platform`, ADR-026) instead of
ad-hoc `runtime.GOOS` checks, and resolve the plugin binary path in one place.

1. **`src/pkg/platform`** — add a helper:

   ```go
   // ExecutableName appends the OS-specific executable suffix (".exe" on
   // Windows) when the name has no extension yet
   func ExecutableName(name string) string
   ```

   Uses `IsWindows()`; leaves names that already have `.exe` untouched.

2. **`RegistryPlugin.BinaryPath()`** becomes the single source of truth:
   - `native` runtime → `platform.ExecutableName(BinaryName)`
   - `python` + wheel → `venvExecPath(...)` (which itself uses
     `platform.ExecutableName`)
   - `java` (`.jar`) and plain `python` (`.py`) → unchanged, no suffix

3. **Replace duplicated joins** with `BinaryPath()`:
   - `manager.checkHelperPluginHealth`
   - `manager.go:263` (gRPC plugin config)
   - `cmd/plugin_dispatcher.go` (native + wheel branches)

4. **Out of scope / follow-up**: `dispatcher.binSuffix` and
   `selfinstall.mainBinaryName()` can later switch to
   `platform.ExecutableName` for consistency.

## Acceptance Criteria

- [x] `portunix plugin health modeler` on Windows reports `✅ Healthy`
- [x] Health on Linux unchanged (binary without extension)
- [x] Java and non-wheel Python plugins unaffected (no `.exe` appended)
- [x] Manifest with explicit `ptx-foo.exe` does not become `ptx-foo.exe.exe`
- [x] Unit tests for `platform.ExecutableName` and `RegistryPlugin.BinaryPath()`
- [x] Plugin execution via dispatcher still works on Windows and Linux
