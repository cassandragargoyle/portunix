# Issue #160: Python Plugin gRPC Service Support

**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, plugin-system, python, grpc, service-management, venv

## Problem

Portunix plugin system doesn't support Python plugins running as gRPC services. The `plugin.type` field forces a mutually exclusive choice:

| type   | Python venv              | service start             | Use case           |
|--------|--------------------------|---------------------------|--------------------|
| helper | Yes                      | No ("not a gRPC service") | CLI tools          |
| grpc   | No (expects native binary) | Yes                     | Go/native services |

## Affected Plugin

**reco v0.1.4** — implements `TaskPlatformService` (task discovery + execution via gRPC) but cannot be started as a service because it's a Python plugin.

## Resolution

**Already supported.** Code analysis revealed that the combination `type: "grpc"` + `runtime: "python"` + `wheel: "..."` is fully supported by the existing codebase:

1. `InstallPlugin()` calls `setupPythonWheelPlugin()` when `runtime == "python" && wheel != ""` — independent of `type`
2. `RegisterPlugin()` sets `mode = "service"` for `type == "grpc"`
3. `BinaryPath()` returns `.venv/bin/<binary>` for python wheel plugins
4. `spawnPlugin()` handles `runtime == "python"` correctly

The reported errors were caused by using `type: "helper"` (which blocks service start) or changing type without reinstalling (registry retains old mode).

### Improvement

Added contextual hint to the "helper plugin, not a gRPC service" error message. When a Python wheel plugin is detected as helper, the error now suggests setting `type: "grpc"` and reinstalling.

### Correct plugin.json for Python gRPC Service

```json
{
  "plugin": {
    "type": "grpc",
    "binary": "ptx-reco",
    "runtime": "python",
    "python_min_version": "3.11",
    "wheel": "ptx_reco-0.1.4-py3-none-any.whl",
    "port": 9101,
    "health_check_interval": 30000000000
  }
}
```

### Installation Steps

```bash
portunix plugin uninstall reco          # if previously installed as helper
portunix plugin install ./reco          # installs with type: grpc, creates venv
portunix plugin enable reco
portunix service start reco             # starts Python gRPC service
```

## Code Flow (verified)

```
InstallPlugin()
  ├── setupPythonWheelPlugin() → creates .venv, pip installs wheel
  └── RegisterPlugin() → mode = "service" (from type: grpc)

service start <plugin>
  ├── resolvePluginBinary() → mode == "service" ✓
  │   └── BinaryPath() → .venv/bin/ptx-reco
  └── spawnPlugin(binaryPath, "python", ..., port)
      └── exec.Command(binaryPath, "--grpc-port", port)
```

## Related

- portunix-reco INT-019 (TaskPlatformService implementation — done, unblocked by this)
- portunix-reco INT-011 (gRPC service mode — --port flag, health check, reflection — done)
- portunix-vscode #025 (AI Task Platform sidebar — unblocked)
- portunix-vscode #026 (Document Canvas context menu — unblocked)
