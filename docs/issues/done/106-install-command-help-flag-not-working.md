# Issue #106: Install Command --help Flag Not Working

## Summary

The `portunix install --help` command does not display help information. Instead, it interprets `--help` as a package name and attempts to install a package called `--help`.

## Problem Description

When running `portunix install --help`, the expected behavior is to display help information for the install command. However, the actual behavior is:

```
$ portunix install --help

🔧 Installing package: --help

❌ Installation failed: package not found: package '--help' not found
```

## Expected Behavior

```
$ portunix install --help
Install packages and tools

Usage:
  portunix install <package-name> [flags]

Flags:
  --variant string    Package variant to install
  --dry-run          Show what would be installed
  -h, --help         Show help for install command

Available Packages:
  python, nodejs, java, go, hugo, ...
```

## Affected Commands

- `portunix install --help`
- Potentially other commands that accept arguments

## Impact

- **Documentation Generation**: The `post-release-docs.py` script relies on `portunix <cmd> --help` to generate command documentation. This bug causes incorrect documentation to be generated for the `install` command.
- **User Experience**: Users cannot easily discover install command options and available packages.

## Root Cause

The `install` command parser processes all arguments as package names before checking for help flags. The `--help` flag should be intercepted before argument processing.

## Proposed Solution

Modify the install command handler to:
1. Check for `--help` or `-h` flags first
2. Display help information when help flag is detected
3. Only process remaining arguments as package names

## Files to Modify

- `src/helpers/ptx-installer/main.go` - Add help flag detection in `handleInstall()` function
- `src/helpers/ptx-installer/registry/registry.go` - Fix embedded assets path for cross-platform compatibility

## Additional Fix: Embedded Assets Path Bug

During implementation, a related bug was discovered and fixed:

**Problem**: The `loadPackageFromEmbedded()` function used `filepath.Join()` which creates platform-specific paths (backslashes on Windows). The embedded filesystem requires forward slashes regardless of platform.

**Solution**: Changed from `filepath.Join("assets/packages", filename)` to `"assets/packages/" + filename` for cross-platform compatibility.

This fix enables `portunix package list` and `portunix package info <package>` commands to work correctly.

## Test Cases

1. `portunix install --help` - should show help
2. `portunix install -h` - should show help
3. `portunix install python --help` - should show help (not install python)
4. `portunix install python` - should install python (unchanged)
5. `portunix package list` - should list all available packages
6. `portunix package info python` - should show package details

## Priority

**High** - Affects documentation generation and user experience.

## Type

Bug Fix

## Labels

bug, cli, help-system, install, user-experience, documentation

## Related Issues

- #076 - Container Run Help Command Not Working (similar pattern)
- #077 - Container Run-in-Container Help Flag Parsing
- #096 - Container Start/Stop Commands Misinterpret --help Flag

## Created

2025-12-03

## Status

✅ Implemented

## Implementation Notes (2025-12-03)

- Fixed `--help` flag handling in `src/helpers/ptx-installer/main.go`
- Fixed embedded assets path bug in `src/helpers/ptx-installer/registry/registry.go`
- Added `make` and `package` commands to `src/cmd/help_registry.go` for automatic discovery
- Updated `scripts/post-release-docs.py` for Windows binary path detection
- Documentation now generates correctly with all commands and rich package metadata
