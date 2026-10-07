# Issue #155: Plugin Prerequisites Validation

**Status:** ✅ Implemented
**Priority:** High
**Labels:** enhancement, plugin-system, validation, prerequisites, user-experience
**Milestone:** TBD
**Closed:** 2026-05-03

## Overview

Portunix plugin system declares prerequisites in `plugin.json` manifest (runtime, runtime version, OS support, optional tools, portunix min version), but **none of these are actually validated at runtime**. When a user installs or starts a plugin, the system does not check whether the required prerequisites are satisfied. This leads to confusing runtime failures instead of clear, actionable error messages.

## Problem Statement

### Current State

The `PluginDependencies` and `PluginBinaryConfig` structures in `src/app/plugins/types.go` support:

- `portunix_min_version` — minimum Portunix version (e.g., `">=1.5.0"`)
- `os_support` — supported operating systems (`linux`, `windows`, `darwin`)
- `optional_tools` — external tools with reason (name + reason)
- `runtime` — plugin runtime (`native`, `java`, `python`)
- `runtime_version` — required runtime version (e.g., `">=21"` for Java)

However, `ValidateManifest()` in `src/app/plugins/manifest.go` only validates **manifest structure** (required fields, valid enum values, port ranges). It does **not** check:

1. Whether the declared runtime is actually installed on the system
2. Whether the installed runtime meets the version requirement
3. Whether the current Portunix version satisfies `portunix_min_version`
4. Whether optional tools are available
5. Whether the current OS matches `os_support`

### Failure Scenarios

| Scenario | Current Behavior | Expected Behavior |
| -------- | --------------- | ------------------ |
| Java plugin, no JDK installed | Plugin start fails with cryptic exec error | Clear message: "Java runtime >=21 required. Install with: `portunix install java-21`" |
| Python plugin, no python3 | Plugin start fails | Clear message: "Python 3 runtime required. Install with: `portunix install python`" |
| Plugin requires Portunix >=2.0, user has 1.10 | Plugin installs but may not work | Reject installation with version mismatch message |
| Plugin supports only Linux, user on Windows | Plugin installs, fails at start | Reject installation with OS incompatibility message |
| Plugin needs `protoc` as optional tool | No warning | Warning: "Optional tool 'protoc' not found — some features may be limited" |

## Requirements

### 1. New Command: `portunix plugin check [plugin-name]`

Check prerequisites for a specific installed plugin or all installed plugins.

```bash
# Check single plugin
portunix plugin check text-extractor

# Check all installed plugins
portunix plugin check --all

# JSON output for programmatic use
portunix plugin check --all --json
```

#### Output Format (Human-Readable)

```text
Plugin: text-extractor v1.0.0
  Runtime:            java >=21    ✅ Found: OpenJDK 21.0.3
  Portunix version:   >=1.8.0     ✅ Current: v1.10.7
  OS support:         linux        ✅ Current: linux
  Optional tools:
    - tesseract                    ⚠️  Not found (OCR features will be unavailable)

Plugin: my-python-plugin v0.5.0
  Runtime:            python >=3.10  ❌ Not found
  Portunix version:   >=1.5.0       ✅ Current: v1.10.7
  OS support:         linux,windows  ✅ Current: windows

Summary: 2 plugins checked, 1 issue found, 1 warning
```

#### Output Format (JSON)

```json
{
  "results": [
    {
      "plugin": "text-extractor",
      "version": "1.0.0",
      "checks": {
        "runtime": { "status": "ok", "required": "java >=21", "found": "OpenJDK 21.0.3" },
        "portunix_version": { "status": "ok", "required": ">=1.8.0", "found": "v1.10.7" },
        "os_support": { "status": "ok", "required": ["linux"], "current": "linux" },
        "optional_tools": [
          { "name": "tesseract", "status": "warning", "reason": "OCR features will be unavailable" }
        ]
      },
      "overall": "warning"
    }
  ],
  "summary": { "total": 2, "ok": 1, "warning": 1, "error": 0 }
}
```

### 2. Prerequisites Check Integration Points

Prerequisites validation should run automatically at these lifecycle points:

| Lifecycle Event | Check Type | Behavior on Failure |
|----------------|-----------|-------------------|
| `plugin install` | Full check | Warning + prompt to continue (non-blocking) |
| `plugin start` | Runtime + OS | **Block start** with actionable error message |
| `plugin enable` | OS support | Warning if OS not supported |
| `plugin check` | Full check | Report only (no side effects) |

### 3. Prerequisite Check Categories

#### 3a. Runtime Availability

- **Native**: No runtime check needed (binary is self-contained)
- **Java**: Check `java -version` output, parse version, compare with `runtime_version`
- **Python**: Check `python3 --version` output, parse version, compare with `runtime_version`

#### 3b. Portunix Version

- Compare current Portunix version against `portunix_min_version` using semver comparison
- Already available: Portunix knows its own version at build time

#### 3c. OS Compatibility

- Compare runtime `GOOS` against `os_support` list
- Straightforward check, already have OS detection in `app/system/`

#### 3d. Optional Tools

- For each entry in `optional_tools`, check if the tool binary is available in PATH
- Report as **warning**, not error — these are optional by definition

### 4. Actionable Remediation Messages

When a prerequisite check fails, provide a concrete remediation command using Portunix's own installation system:

```
❌ Java runtime >=21 required but not found.
   Fix: portunix install java-21

❌ Python 3 runtime required but not found.
   Fix: portunix install python

⚠️  Optional tool 'protoc' not found.
   Fix: portunix install protoc
   Reason: Required for gRPC code generation
```

The remediation mapping should leverage `ptx-installer` and the package registry to suggest correct install commands.

## Acceptance Criteria

- [ ] `portunix plugin check <name>` validates all prerequisites for a single plugin
- [ ] `portunix plugin check --all` validates prerequisites for all installed plugins
- [ ] `--json` flag produces machine-readable output
- [ ] Runtime availability checked for Java and Python plugins
- [ ] Runtime version compared using semver when `runtime_version` is specified
- [ ] Portunix minimum version validated against current version
- [ ] OS compatibility validated against current platform
- [ ] Optional tools checked for availability in PATH
- [ ] Actionable remediation messages with `portunix install` commands
- [ ] `plugin start` blocks if runtime prerequisite is not met
- [ ] `plugin install` warns but does not block on missing prerequisites
- [ ] Exit code 0 = all OK, exit code 1 = errors found, exit code 2 = warnings only

## Affected Components

- `src/app/plugins/manifest.go` — extend with runtime validation logic
- `src/app/plugins/manager/manager.go` — integrate checks into install/start lifecycle
- `src/cmd/plugin.go` — add `check` subcommand
- `src/app/plugins/types.go` — potentially add `PrerequisiteCheckResult` type

## Related Issues

- [#007](007-plugin-system-grpc.md) — Plugin System with gRPC Architecture
- [#024](210-plugin-registration-system.md) — Plugin Registration and Discovery System
- [#132](../132-text-extractor-plugin-integration.md) — Text Extractor Plugin (Java runtime dependency)
- [#150](../150-distributed-mcp-server-ecosystem.md) — Distributed MCP Server Ecosystem

## Design Notes

- Reuse existing `app/system/` for OS detection
- Reuse `ptx-installer` knowledge for remediation suggestions
- Version comparison should use standard semver library or Go's `semver` package
- Consider caching runtime detection results to avoid repeated subprocess calls during batch check

---
**Created:** 2025-03-15
**Assigned:** TBD
**Last Updated:** 2025-03-15
