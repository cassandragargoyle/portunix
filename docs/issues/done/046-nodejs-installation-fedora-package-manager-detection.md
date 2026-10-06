# Issue #46: Node.js Installation Fails on Fedora Due to Incorrect Package Manager Detection

**Issue ID**: #046  
**Type**: Bug Fix  
**Priority**: High  
**Status**: 📋 Open  
**Created**: 2025-09-12  
**Labels**: bug, nodejs, fedora, package-manager  

## Problem Description

Node.js installation fails on Fedora 39 because the package installer incorrectly selects APT as the installation method instead of DNF (Fedora's native package manager).

### Error Output
```
═══════════════════════════════════════════════
📦 INSTALLING: Node.js JavaScript Runtime
════════════════════════════════════════════════
📄 Description: JavaScript runtime built on Chrome's V8 engine with npm package manager
🔧 Variant:  (vlatest)
💻 Platform: linux
🏗️  Installation type: apt
📋 Packages: nodejs, npm
════════════════════════════════════════════════
🔍 Checking if nodejs is already installed...
📋 nodejs is not installed, proceeding with installation...
🚀 Starting installation...

❌ Installation FAILED!
Package: Node.js JavaScript Runtime
Variant: 
Error: APT is not available on this system
Error installing package 'nodejs': APT is not available on this system
[root@a1e1d4d8c634 /]# 
```

## Root Cause Analysis

1. **Incorrect Package Manager Detection**: The system detects `installation type: apt` for Fedora 39, but Fedora uses DNF
2. **Platform vs Package Manager Mismatch**: While the system correctly identifies `Platform: linux`, it fails to detect the correct package manager for Fedora-based distributions
3. **Missing DNF Support**: Node.js package definition may not include proper DNF installation instructions

## Expected Behavior

- **Fedora 39**: Should detect `installation type: dnf` and use `dnf install nodejs npm`
- **Package Manager Detection**: Should correctly identify DNF as the package manager for Fedora distributions
- **Successful Installation**: Node.js should install successfully using native Fedora packages

## Affected Systems

- **Primary**: Fedora 39 (`fedora:39` container image)
- **Likely Affected**: Other Fedora versions (Fedora 40, Fedora 38, etc.)
- **Potentially Affected**: Other RPM-based distributions using DNF (RHEL 8+, CentOS Stream, Rocky Linux, AlmaLinux)

## Technical Analysis

### Package Manager Detection Logic
The issue appears to be in the package manager detection logic that runs before installation:

1. **Container Detection**: System correctly identifies Fedora container
2. **Package Manager Detection**: Incorrectly selects APT instead of DNF
3. **Installation Attempt**: Tries to run APT commands on Fedora system
4. **Failure**: APT is not available on Fedora, causing installation failure

### Node.js Package Definition
Location: `assets/install-packages.json`

The Node.js package definition may be missing proper DNF instructions or have incorrect platform detection rules.

## Acceptance Criteria

### Must Have
- [ ] Fedora 39 correctly detects DNF as package manager
- [ ] Node.js installs successfully on Fedora 39 using `dnf install nodejs npm`
- [ ] Installation type shows `dnf` instead of `apt` for Fedora systems
- [ ] Package manager detection works for all Fedora versions

### Should Have  
- [ ] Other RPM-based distributions (Rocky Linux, AlmaLinux) work correctly
- [ ] Proper error handling for unsupported package managers
- [ ] Clear error messages when package manager detection fails

### Nice to Have
- [ ] Verbose logging for package manager detection process
- [ ] Fallback mechanisms for package manager detection failures
- [ ] Support for multiple package manager options per distribution

## Implementation Strategy

### Phase 1: Diagnosis
1. **Review Package Manager Detection Logic**
   - Check `app/install/` package manager detection code
   - Identify where Fedora → APT mapping occurs
   
2. **Review Node.js Package Definition**
   - Check `assets/install-packages.json` for Node.js entry
   - Verify DNF installation instructions exist

### Phase 2: Fix Package Manager Detection
1. **Update Detection Logic**
   - Ensure Fedora distributions map to DNF
   - Add proper distribution family detection (Debian→APT, RHEL→DNF, etc.)
   
2. **Test Detection**
   - Verify detection works in Fedora 39 container
   - Test other Fedora versions and RPM-based distributions

### Phase 3: Fix Node.js Package Definition
1. **Add DNF Support**
   - Ensure Node.js package has DNF installation variant
   - Update package definition if needed
   
2. **Test Installation**
   - Verify Node.js installs successfully on Fedora
   - Test npm functionality after installation

## Testing Strategy

### Container Testing (Mandatory)
Following `TESTING_METHODOLOGY.md`:

```bash
# Test Fedora 39
./portunix container run-in-container nodejs --image fedora:39

# Test other Fedora versions  
./portunix container run-in-container nodejs --image fedora:40
./portunix container run-in-container nodejs --image fedora:38

# Test other RPM-based distributions
./portunix container run-in-container nodejs --image rockylinux:9
./portunix container run-in-container nodejs --image almalinux:9
```

### Verification Tests
```bash
# Verify package manager detection
./portunix install nodejs --dry-run  # Should show 'dnf' for Fedora

# Verify successful installation
./portunix container exec <fedora-container> node --version
./portunix container exec <fedora-container> npm --version
```

## Related Issues

- **#045**: Node.js Installation Critical Fixes - This issue was discovered during comprehensive distribution testing
- **#041**: Node.js/npm Installation Support - Original Node.js installation implementation

## Security Considerations

- **Container Isolation**: All testing must be done in containers, not on host system
- **Package Verification**: Ensure DNF packages are from official Fedora repositories
- **Permission Handling**: Verify sudo/root permissions work correctly with DNF

## Success Metrics

1. **Package Manager Detection**: 100% accuracy for Fedora distributions
2. **Installation Success Rate**: 100% success on Fedora 39, 40 (CRITICAL priority distributions in ADR-009)
3. **Node.js Functionality**: Both `node --version` and `npm --version` work after installation
4. **No Regression**: APT-based distributions (Ubuntu, Debian) continue working

## Time Estimate

- **Investigation**: 1-2 hours
- **Implementation**: 2-3 hours  
- **Testing**: 2-3 hours
- **Total**: 5-8 hours

## Notes

- This issue blocks comprehensive Node.js testing across all officially supported distributions (ADR-009)
- High priority due to Fedora being a HIGH priority distribution in testing matrix
- Should be fixed before Issue #045 is considered fully resolved

---

**Created by**: Claude Code QA/Test Engineer  
**Discovery Context**: Issue #045 comprehensive distribution testing  
**Related Test**: `test/integration/issue_045_nodejs_critical_fixes_test.go`