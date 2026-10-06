# Issue #134: PFT Config Cross-Platform Path Support

## Summary

The PTX-PFT helper fails to list feedback items when used across different operating systems because the `.pft-config.json` stores absolute paths that are OS-specific.

## Problem Description

### Reproduction Steps

1. Create PFT project on Linux:
   ```bash
   portunix pft configure
   # Creates config with path: "/home/user/project/docs/pft-project"
   ```

2. Clone/copy project to Windows and try to list items:
   ```bash
   portunix pft list --path "E:\Dev\project\docs\pft-project"
   # Output: Feedback Items - Project Name
   #         Total: 0 items
   ```

### Root Cause

In `loadOrCreateConfig()` (main.go:559-584):
- Config is loaded from the provided `--path` directory
- But `config.Path` is used from the JSON file (contains Linux path)
- The Linux path `/home/user/...` doesn't exist on Windows
- `scanLocalDirectory()` searches non-existent directories → 0 items

### Affected Code

```go
// main.go:1514-1523
config, err := loadOrCreateConfig(configPath)
// ...
projectDir := config.Path  // Uses path from JSON, not --path argument
if projectDir == "" {
    projectDir, _ = os.Getwd()
}
```

## Proposed Solution

See [ADR-032](../../adr/032-pft-cross-platform-config-path-resolution.md) for detailed design.

### Summary

1. **Support relative/empty paths in config**
   - Empty `path` → use config file directory
   - Relative `path` → resolve from config file location
   - Absolute `path` → use as-is (backwards compatible)

2. **Runtime --path override**
   - When `--path` is explicitly provided, always use it
   - Ignore stored path from config (may be from different OS)

3. **Migration helper**
   - Add `pft configure --fix-paths` to convert absolute to relative

## Implementation Tasks

- [x] Implement `resolveProjectPath()` function
- [x] Update `loadOrCreateConfig()` to return config file path
- [x] Modify all commands to use resolved path
- [x] Update `handleListCommand()`, `handleShowCommand()`, etc.
- [x] Add `--fix-paths` subcommand to configure
- [x] Update documentation
- [x] Add integration tests for cross-platform scenarios

## Affected Commands

All PFT commands that use `config.Path`:
- `pft list`
- `pft show`
- `pft push`
- `pft pull`
- `pft sync`
- `pft user *`
- `pft role *`
- `pft category *`

## Priority

**High** - Blocks cross-platform team collaboration on PFT projects

## Type

Enhancement / Bug Fix

## Labels

bug, enhancement, helper-binary, ptx-pft, cross-platform, configuration

## Status

✅ Implemented (2026-05-10)

## Related

- ADR: [032-pft-cross-platform-config-path-resolution](../../adr/032-pft-cross-platform-config-path-resolution.md)
- Parent: [107-ptx-pft-product-feedback-tool-helper](107-ptx-pft-product-feedback-tool-helper.md)
