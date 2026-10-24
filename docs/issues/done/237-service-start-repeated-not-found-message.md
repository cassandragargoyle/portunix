# Issue #237: service start: "plugin X not found" repeated in the error message

**Status**: Implemented
**Priority**: Low
**Type**: Bug
**Component**: service command (`src/cmd/service.go`)
**Created**: 2026-10-24
**Closed**: 2026-10-24
**GitHub**: #237
**Related**: `docs/testing/acceptance-236.md`, #159 (service orchestration command group), portunix-vscode #19 (acceptance
defect X2)

## Summary

For an unknown plugin the error repeats its reason:

```text
$ portunix service start nonexistent-plugin-xyz --output json
Error: plugin nonexistent-plugin-xyz not found: plugin nonexistent-plugin-xyz not found
```

Clients such as the VS Code extension and Portunix Pilot show this line to the user.

## Problem Description

`resolvePluginBinary` in `src/cmd/service.go` wraps the registry error:

```go
registryData, regErr := pluginManager.GetPluginRegistryData(pluginName)
if regErr != nil {
    return "", "", nil, fmt.Errorf("plugin %s not found: %w", pluginName, regErr)
}
```

`GetPluginRegistryData` in `src/app/plugins/manager/registry.go` already returns
`plugin %s not found`, so the text appears twice.

## Proposed Solution

Return the registry error as it is, or wrap it without repeating the text, for example
`fmt.Errorf("cannot start service: %w", regErr)`. Check the other wrappers of the same
error in `src/app/plugins/manager/manager.go` (`plugin not found: %w`).

## Acceptance Criteria

- [x] `portunix service start <unknown>` prints the reason once
- [x] `--output json` and text output use the same message

## Testing

- Unit test for the message of `resolvePluginBinary` with an unknown plugin
