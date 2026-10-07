# Issue #162: Plugin Interfaces Field Misplacement

**Date**: 2026-03-29
**Reporter**: internal review
**Status**: ✅ Implemented
**Priority**: High
**Type**: Bug Fix
**Labels**: bug, plugin-system, manifest, schema-mismatch, interfaces
**Affects**: portunix v2.0.0
**Related**: plugin-manifest.schema.json (api/contract)

## Problem

The `interfaces` field in the Go `PluginManifest` struct is at the top level, but the
canonical JSON Schema (`plugin-manifest.schema.json`) defines it inside the `plugin` object.
This means plugins that follow the schema (e.g. reco-studio v0.1.7) place `interfaces` inside
`plugin {}`, and Portunix CLI silently ignores it because Go unmarshals from the wrong path.

Additionally, there are two display-side bugs:

1. `ListEnabledPlugins()` does not copy `Mode` or `Interfaces` into the returned `PluginInfo`.
2. `showPluginInfo()` does not display the `Interfaces` field at all.

## Reproduction

1. Create a plugin manifest following the JSON Schema:

   ```json
   {
     "name": "example",
     "version": "0.1.0",
     "description": "Example plugin",
     "plugin": {
       "type": "helper",
       "binary": "ptx-example",
       "runtime": "native",
       "interfaces": ["cli"]
     }
   }
   ```

2. Install and enable the plugin:

   ```bash
   portunix plugin install ./example/
   portunix plugin enable example
   ```

3. List plugins:

   ```bash
   portunix plugin list
   ```

4. **Result**: INTERFACE column shows `cli` (derived from mode fallback), not from the manifest.
   The actual `interfaces` value from the manifest is lost during JSON unmarshaling.

5. Show plugin info:

   ```bash
   portunix plugin info example
   ```

6. **Result**: No `Interfaces` line in the output.

## Root Cause

- `PluginManifest.Interfaces` is a top-level field (`types.go:173`)
- JSON Schema defines `interfaces` inside `plugin` object (`plugin-manifest.schema.json:79`)
- `RegisterPlugin` reads `manifest.Interfaces` (top-level) which is always empty for
  schema-compliant manifests

## Required Changes

### 1. Move `Interfaces` from `PluginManifest` to `PluginBinaryConfig` (types.go)

- Remove `Interfaces []string` from `PluginManifest` (line 173)
- Add `Interfaces []string` to `PluginBinaryConfig`
- This aligns Go struct with the canonical JSON Schema

### 2. Update registry references (manager/registry.go)

- Change `manifest.Interfaces` to `manifest.Plugin.Interfaces` in:
  - `RegisterPlugin()` (line 149)
  - `ReregisterPlugin()` (line 224)

### 3. Fix `ListEnabledPlugins()` (manager/registry.go)

- Add missing fields to `PluginInfo` construction (lines 360-371):
  - `Mode: registryPlugin.Mode`
  - `Interfaces: registryPlugin.Interfaces`

### 4. Add Interfaces display to `showPluginInfo()` (cmd/plugin.go)

- After the Status/LastSeen block, add:

  ```go
  if len(pluginInfo.Interfaces) > 0 {
      fmt.Printf("Interfaces:  %s\n", strings.Join(pluginInfo.Interfaces, ", "))
  }
  ```

## Affected Files

- `src/app/plugins/types.go` — struct field relocation
- `src/app/plugins/manager/registry.go` — field path update + ListEnabledPlugins fix
- `src/cmd/plugin.go` — display enhancement

## Testing

- Install a plugin with `interfaces` inside `plugin` object
- Verify `portunix plugin list` shows correct INTERFACE column
- Verify `portunix plugin info <name>` shows Interfaces line
- Verify `portunix plugin list --format json` includes interfaces in output
- Verify backward compatibility: plugins without `interfaces` still derive from mode

## Acceptance Criteria

- [ ] `Interfaces` field parsed from `plugin.interfaces` in manifest JSON
- [ ] `PluginManifest` struct matches JSON Schema structure
- [ ] `ListEnabledPlugins()` returns `Mode` and `Interfaces`
- [ ] `showPluginInfo()` displays interfaces when present
- [ ] Backward-compatible fallback in `outputPluginTable()` preserved
- [ ] All existing tests pass
