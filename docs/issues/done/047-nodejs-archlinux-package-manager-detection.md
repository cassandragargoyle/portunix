# Issue #47: Node.js Installation Fails on Arch Linux Due to Incorrect Package Manager Detection

## Summary
Node.js installation on Arch Linux containers fails because Portunix incorrectly detects and attempts to use `apt` package manager instead of `pacman`, which is the native package manager for Arch Linux.

## Problem Description

During container testing of Node.js installation on Arch Linux (`archlinux:latest`), Portunix incorrectly identifies the installation method as `apt`, causing installation failure.

**Current Error Behavior:**
```bash
# In archlinux:latest container
/usr/local/bin/portunix install nodejs --dry-run
🏗️  Installation type: apt
📋 Packages: nodejs, npm
```

**Expected Behavior:**
```bash
# Should detect pacman for Arch Linux
🏗️  Installation type: pacman
📋 Packages: nodejs, npm
```

## Root Cause Analysis

**Package Manager Detection Logic Issue:**
- Portunix installation detection logic fails to properly identify Arch Linux
- Falls back to `apt` as default instead of detecting `pacman`
- Arch Linux uses `pacman` as its native package manager, not `apt`

## Impact

**Blocking Issues:**
1. **Node.js Installation Completely Fails** on Arch Linux containers
2. **Cross-Platform Support Broken** - One of the 9 officially supported distributions doesn't work
3. **Container Integration Incomplete** - Arch Linux containers cannot install Node.js
4. **Testing Coverage Gap** - Issue #045 testing shows false positive for Arch Linux

**Affected Components:**
- 🔴 Node.js installation on Arch Linux
- 🔴 Package manager detection logic in `app/install/`
- 🔴 Cross-platform compatibility claims
- 🔴 Container-based installation testing

## Technical Details

### Current Detection Logic Issues
1. **OS Detection**: May not properly identify Arch Linux from container environment
2. **Package Manager Priority**: `apt` detection takes precedence over `pacman`
3. **Fallback Logic**: Defaults to `apt` when detection fails instead of trying `pacman`

### Arch Linux Specifics
- **Distribution**: Arch Linux
- **Package Manager**: `pacman`
- **Node.js Package Name**: `nodejs` and `npm` (same as other distributions)
- **Installation Command**: `pacman -S nodejs npm`
- **Container Image**: `archlinux:latest`

## Expected Solution

### 1. Fix Package Manager Detection
Update detection logic in installation system:
- Properly identify Arch Linux containers/systems
- Add `pacman` detection before falling back to `apt`
- Test for existence of `/usr/bin/pacman` binary

### 2. Add Arch Linux Support to Node.js Package Definition
Update `assets/install-packages.json`:
```json
{
  "nodejs": {
    "linux": {
      "arch": {
        "pacman": ["nodejs", "npm"]
      }
    }
  }
}
```

### 3. Update Installation Logic
Modify installation detection to:
1. Check for Arch Linux identification
2. Verify `pacman` availability
3. Use correct package installation commands

## Acceptance Criteria

- [ ] `portunix install nodejs --dry-run` shows `Installation type: pacman` on Arch Linux
- [ ] `portunix install nodejs` successfully installs Node.js on Arch Linux containers
- [ ] Package manager detection correctly identifies `pacman` on Arch Linux systems
- [ ] `node --version` and `npm --version` work after installation on Arch Linux
- [ ] Container integration test passes for `archlinux:latest`
- [ ] No regression in existing package manager detection for other distributions
- [ ] Issue #045 test suite shows true positive results for Arch Linux

## Files to Modify

1. **`app/install/package_manager_detection.go`** - Fix Arch Linux detection
2. **`assets/install-packages.json`** - Add Arch Linux pacman support for nodejs
3. **`app/install/nodejs.go`** - Add pacman installation logic
4. **`test/integration/issue_045_nodejs_critical_fixes_test.go`** - Verify Arch Linux fix

## Priority

**HIGH** - Affects cross-platform compatibility and official distribution support claims

## Test Case

After implementation, this should work:

```bash
# Create Arch Linux container
portunix container run archlinux:latest

# Copy latest Portunix binary
portunix container cp ./portunix container-name:/usr/local/bin/portunix

# Test detection
portunix container exec container-name /usr/local/bin/portunix install nodejs --dry-run
# Expected output:
# 🏗️  Installation type: pacman
# 📋 Packages: nodejs, npm

# Test actual installation
portunix container exec container-name /usr/local/bin/portunix install nodejs
# Should complete successfully

# Verify installation
portunix container exec container-name node --version  # Should output version
portunix container exec container-name npm --version   # Should output version
```

## Dependencies

This issue is related to:
- **Issue #041**: Node.js/npm Installation Support (parent feature)
- **Issue #045**: Node.js Installation Critical Fixes (testing revealed this issue)
- **ADR-009**: Official Linux distribution support (affects compliance)

## Discovery Context

**Discovered During**: Issue #041 acceptance testing
**Test Scenario**: Cross-platform Node.js installation validation
**Container**: `nodejs-archlinux-latest` (archlinux:latest)
**Evidence**: Dry-run shows `Installation type: apt` instead of `Installation type: pacman`

---

**Reporter:** Claude Code Assistant  
**Date:** 2025-09-12  
**Category:** Bug Fix  
**Priority:** High  
**Labels:** bug, nodejs, arch-linux, package-manager, container, cross-platform  
**Status:** 📋 Open