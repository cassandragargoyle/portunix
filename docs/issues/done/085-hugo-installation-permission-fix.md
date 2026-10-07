# Issue #85: Hugo Installation Permission Fix

## Overview
During Hugo installation testing on development machine, encountered permission errors when installing to `/usr/local/bin`.

## Problem Description
When running `./portunix install hugo --variant extended`, the installation fails with:
```
tar: hugo: Funkce open selhala: Operace zamítnuta
tar: README.md: Funkce open selhala: Operace zamítnuta
tar: LICENSE: Funkce open selhala: Operace zamítnuta
tar: Končí se chybovým kódem, protože byly zaznamenány chyby
Error installing package 'hugo': failed to install hugo: exit status 2
```

## Root Cause
The installation process attempts to extract files to `/usr/local/bin` without proper sudo privileges. The tar extraction fails due to permission restrictions.

## Special Testing Permission
**APPROVED FOR THIS ISSUE ONLY**: Testing Hugo installation on development machine is explicitly approved as part of this issue resolution process, since this issue originated from testing on this specific machine.

## Proposed Solution
1. **Immediate Fix**: Modify Hugo installation to use proper sudo elevation for Linux
2. **Long-term Fix**: Implement user-space installation option (e.g., `~/.local/bin`)

### Technical Implementation
- Check if target directory requires elevated permissions
- Automatically prompt for sudo when needed
- Provide fallback installation to user directory if sudo not available
- Update package verification to check both system and user paths

## Acceptance Criteria
- [ ] Hugo installs successfully with appropriate permissions
- [ ] No permission errors during extraction
- [ ] Hugo binary is accessible via PATH
- [ ] `hugo version` command works after installation
- [ ] Both standard and extended variants work

## Test Cases
1. Install Hugo on clean system with sudo
2. Install Hugo on system without sudo (should use user directory)
3. Verify Hugo functionality after installation
4. Test both standard and extended variants

## Priority
**Medium** - Affects package installation functionality

## Labels
- bug
- installation
- permissions
- hugo
- linux

## CRITICAL UPDATE: Architectural Dependency Discovered

### Issue #086 Dependency
**BLOCKING ISSUE**: During acceptance testing, a critical architectural flaw was discovered in the package registry system that prevents testing of this Hugo installation fix.

**Problem**: Package discovery system fails because:
- Hugo package definition exists in `assets/packages/hugo.json`
- But Hugo is not registered in `assets/registry/index.json`
- Current system requires manual registration of every package
- This prevents discovery and testing of the permission fix

**Required Before Implementation**:
- ✅ **Issue #086** must be completed first: Package Registry Automatic Discovery System
- ✅ Package discovery must work for Hugo before permission fix can be tested
- ✅ Architecture must be changed from static index to directory scanning

**Implementation Order**:
1. **First**: Complete Issue #086 (Package Registry Automatic Discovery)
2. **Second**: Test that Hugo package is discoverable
3. **Third**: Implement and test permission fix functionality

### Updated Acceptance Criteria
- [ ] **PREREQUISITE**: Hugo package must be discoverable (Issue #086 dependency)
- [ ] Hugo installs successfully with appropriate permissions
- [ ] No permission errors during extraction
- [ ] Hugo binary is accessible via PATH
- [ ] `hugo version` command works after installation
- [ ] Both standard and extended variants work

### Updated Test Cases
**Phase 1: Verify Package Discovery (Issue #086 dependency)**
1. Verify Hugo package is discoverable via new directory scanning system
2. Verify Hugo package definition loads correctly

**Phase 2: Permission Fix Testing (this issue)**
1. Install Hugo on clean system with sudo
2. Install Hugo on system without sudo (should use user directory)
3. Verify Hugo functionality after installation
4. Test both standard and extended variants

## Related Issues
- **Issue #086**: Package Registry Automatic Discovery System (BLOCKING DEPENDENCY)
- Related to package installation system
- May affect other packages requiring system directory installation

## Testing Status
**BLOCKED**: Cannot proceed with acceptance testing until Issue #086 is resolved.
**Reference**: See `docs/testing/acceptance-085.md` for detailed analysis of blocking issue.

---
**Created**: 2025-09-28
**Updated**: 2025-09-28 (Added Issue #086 dependency)
**Status**: Blocked (waiting for Issue #086)
**Assignee**: Developer
**Dependencies**: Issue #086 (Package Registry Automatic Discovery)
**Testing Machine**: This development machine (approved for this issue)