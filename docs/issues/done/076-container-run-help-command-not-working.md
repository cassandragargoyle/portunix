# Issue #76: Container Run Help Command Not Working

**Created**: 2025-09-26
**Status**: ✅ Implemented
**Priority**: High
**Type**: Bug Fix
**Labels**: bug, container-management, help-system, user-experience, cli

## Summary

The `./portunix container run --help` command displays the general container command help instead of the specific help for the `run` subcommand. This makes it difficult for users to understand the available options and proper usage of the `container run` command.

## Problem Description

When executing `./portunix container run --help`, the system displays:

```
Usage: portunix container [command]

Available Commands:
  check            Check container runtime capabilities and versions
  cp               Copy files/folders between container and host
  exec             Execute command in container (universal runtime)
  info             Show container runtime information and availability
  list             List containers from all available runtimes
  logs             Show container logs (universal runtime)
  rm               Remove container (universal runtime)
  run              Run new container (universal runtime)
  run-in-container Run installation in container (RECOMMENDED for testing)
  start            Start stopped container (universal runtime)
  stop             Stop container (universal runtime)

Flags:
  -h, --help   help for container

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix container [command] --help" for more information about a command.
```

**Expected Behavior**: Should display specific help for the `run` subcommand, including available flags, options, and usage examples.

**Actual Behavior**: Displays general container command help instead of run-specific help.

## Impact

- **User Experience**: Users cannot discover available options for `container run` command
- **Documentation**: Missing help information makes the command difficult to use correctly
- **CLI Consistency**: Breaks expected CLI behavior where `[command] --help` shows command-specific help

## Steps to Reproduce

1. Navigate to Portunix directory
2. Execute: `./portunix container run --help`
3. Observe that general container help is displayed instead of run-specific help

## Expected Solution

The `./portunix container run --help` command should display:

- Specific usage syntax for the `run` command
- Available flags and options (image, name, volume mounts, etc.)
- Usage examples for common scenarios
- Platform-specific considerations (Docker vs Podman)

## Technical Analysis

This appears to be related to the CLI command parsing and help system routing. The help flag is likely being processed at the parent `container` command level instead of being passed to the `run` subcommand.

## Acceptance Criteria

- [x] `./portunix container run --help` displays run-specific help
- [x] Help includes all available flags and options
- [x] Help includes usage examples
- [x] Help is consistent with other subcommand help formats
- [x] General container help still works with `./portunix container --help`

## Related Issues

- Similar to #059 (Playbook Help Command Not Working) - help system routing issue
- Related to #050 (Multi-Level Help System) - overall help system architecture

## Environment

- **Portunix Version**: Current development branch (feature/issue-075-hugo-installation-support)
- **Platform**: Linux
- **Test Date**: 2025-09-26

## Additional Notes

This bug affects the usability of the container management system and should be prioritized for quick resolution as it impacts the developer experience and documentation accessibility.

## Implementation Summary

**Verified**: 2025-09-27
**Status**: Already implemented via helper binary system

### Current Behavior (Working)
The `container run --help` command correctly displays run-specific help through the ptx-container helper binary. The issue described in the original report appears to have been resolved during helper binary implementation.

### Verification Results
- ✅ `./portunix container run --help` displays detailed run-specific help
- ✅ Help includes all supported flags: `-d`, `-i`, `-t`, `--name`, `-p`, `-v`, `-e`
- ✅ Help includes 5 comprehensive usage examples
- ✅ Help follows consistent format with emoji and structured layout
- ✅ General container help (`./portunix container --help`) still works correctly

### Technical Details
The help system works correctly because:
1. Main binary uses dispatcher to route `container` commands to `ptx-container` helper
2. Helper binary has `handleContainerRun()` with proper `--help` flag detection
3. Helper binary includes `showRunHelp()` function with comprehensive help text
4. Integration between dispatcher and helper preserves argument routing

### Files Involved
- `src/helpers/ptx-container/main.go`: Contains working help implementation
- `src/dispatcher/dispatcher.go`: Routes container commands to helper
- Helper binary: `ptx-container` (compiled and deployed)

---

**Reporter**: Claude Code Assistant
**Verified**: Claude Code Assistant
**Resolution**: Issue was already resolved via helper binary system