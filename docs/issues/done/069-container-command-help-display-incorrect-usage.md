# Issue #69: Container Command Help Display Shows Incorrect Usage

**Created**: 2025-09-24
**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Bug Fix

## Summary
The `portunix container --help` command displays incorrect usage information showing `ptx-container [flags]` instead of `portunix container [command]`. This creates confusion for users and inconsistency with other commands.

## Current Behavior
```bash
$ ./portunix container --help
Usage:
  ptx-container [flags]
```

## Expected Behavior
```bash
$ ./portunix container --help
Usage:
  portunix container [command]
```

## Problem Analysis
This issue appears to be similar to other helper integration problems where the helper binary's name is being displayed in the usage instead of the main `portunix` command structure.

The issue is likely located in:
- Container command integration code that calls the `ptx-container` helper
- Help text generation logic that doesn't properly override the binary name
- Command registration that doesn't set the correct usage template

## Affected Components
- `portunix container --help` command
- Container command integration with `ptx-container` helper
- Command help text generation system

## Technical Investigation Needed
1. **Command Integration Analysis**
   - Review how `ptx-container` helper is integrated into main binary
   - Check if command name override is properly implemented
   - Verify usage template configuration

2. **Help System Review**
   - Examine help text generation for container commands
   - Compare with other commands that have correct usage display
   - Check if cobra command configuration is complete

3. **Helper Integration Pattern**
   - Review pattern used by other successfully integrated helpers
   - Ensure consistent implementation across all helper integrations
   - Identify any missing configuration steps

## Reproduction Steps
1. Build the latest Portunix binary
2. Run `./portunix container --help`
3. Observe incorrect usage showing `ptx-container [flags]`

## Expected Implementation Areas
- Container command registration in `cmd/` directory
- Helper integration code for `ptx-container`
- Command usage template configuration
- Help text override mechanisms

## Related Issues
This is similar to previously resolved helper integration issues where the usage display was corrected from helper binary names to proper `portunix` command structure.

## Acceptance Criteria
- [ ] `portunix container --help` shows `portunix container [command]` usage
- [ ] Help text is consistent with other main commands
- [ ] Subcommand help also shows correct usage pattern
- [ ] No regression in container functionality
- [ ] Help display matches standard Portunix command patterns

## Testing Requirements
- Test `portunix container --help` shows correct usage
- Test subcommands like `portunix container run --help`
- Verify all container functionality remains working
- Cross-platform testing (Windows/Linux)
- Compare help output with other commands for consistency

## Priority
**Medium** - This is a UX issue that affects user experience and command consistency

## Labels
- bug
- container-management
- help-system
- user-experience
- helper-integration

## Related Commands
- `portunix container run`
- `portunix container ssh`
- `portunix container exec`
- `portunix container ls`
- All other container subcommands

## Expected Fix Areas
1. **Command Registration**
   - Ensure proper command name is set in cobra configuration
   - Override usage template to show `portunix` instead of `ptx-container`
   - Set correct command hierarchy in help display

2. **Helper Integration**
   - Review and fix helper integration pattern
   - Ensure usage templates are properly overridden
   - Maintain consistency with other integrated helpers

3. **Help System**
   - Verify help generation logic handles helpers correctly
   - Ensure usage templates propagate properly to subcommands
   - Test help display across all container commands

## Debugging Information Needed
- Current cobra command configuration for container commands
- Helper integration implementation details
- Comparison with working command help displays
- Usage template configuration analysis