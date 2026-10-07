# Issue #83: Hugo Registry Installation Fix

**Status**: 📋 Open
**Priority**: High
**Type**: Bug Fix
**Created**: 2025-09-28
**Labels**: bug, package-management, registry, hugo

## Summary
Fix the error "real installation from registry not yet implemented - use legacy format for now" when installing Hugo package from the new Package Registry system.

## Problem Description
When attempting to install Hugo using the command `portunix install hugo`, the installation fails with an error message indicating that registry-based installation is not yet fully implemented. The system falls back to a placeholder error instead of executing the actual installation.

## Current Behavior
```bash
$ portunix install hugo
Error installing package 'hugo': real installation from registry not yet implemented - use legacy format for now
```

## Expected Behavior
The Hugo package should be installed successfully from the Package Registry using the metadata defined in `assets/package-registry.json`.

## Root Cause Analysis
The issue is located in `app/install/registry.go` where there's a placeholder error instead of actual implementation for registry-based installations. The code currently returns an error with the message "real installation from registry not yet implemented - use legacy format for now".

## Technical Requirements
1. Complete the implementation of `installFromRegistry()` function in `app/install/registry.go`
2. Support all installation methods defined in the registry:
   - Direct download from URL
   - GitHub releases download
   - Package manager installations (apt, yum, dnf, etc.)
   - Platform-specific installers
3. Handle version selection and platform detection
4. Implement proper error handling and rollback
5. Support prerequisite checking and installation

## Affected Components
- `app/install/registry.go` - Main registry installation logic
- `app/install/installer.go` - Installation dispatcher
- `assets/package-registry.json` - Package definitions

## Implementation Plan
1. Analyze the existing legacy installation methods
2. Map registry metadata to appropriate installation handlers
3. Implement download and extraction logic for direct URLs
4. Implement GitHub releases API integration
5. Connect to existing package manager handlers
6. Add proper logging and error handling
7. Test with Hugo and other registry packages

## Test Scenarios
1. Install Hugo on Windows (direct download)
2. Install Hugo on Linux (various distributions)
3. Install Hugo on macOS
4. Test prerequisite handling
5. Test version selection
6. Test error scenarios and rollback

## Acceptance Criteria
- [ ] Hugo installs successfully from registry on all platforms
- [ ] Other registry packages install correctly
- [ ] Proper error messages are displayed
- [ ] Installation progress is shown
- [ ] Rollback works on failure
- [ ] All existing legacy packages continue to work

## Related Issues
- #082 - Package Registry Architecture Implementation (parent issue)
- #075 - Implement Hugo Installation Support (original request)

## Notes
This is a critical fix needed to complete the Package Registry implementation from Issue #082. Without this fix, none of the new registry-based packages can be installed, making the entire registry system non-functional.