# Issue #100: PTX-Installer Helper - Phase 2 Test Plan

**Issue**: #100 - PTX-Installer Helper Implementation
**Phase**: Phase 2 - Core Installation Migration (Testing)
**Date**: 2025-10-29
**Tester**: Claude (AI Assistant)
**Testing OS**: Container-based (Ubuntu 22.04)

---

## Test Objectives

Validate that PTX-Installer helper can successfully:
1. Load package registry (34+ packages)
2. Perform dry-run installations
3. Execute real package installations (archive, APT)
4. Resolve dependencies
5. Handle platform detection
6. Manage downloads and extractions

---

## Test Environment

**Container Requirements:**
- Base Image: `ubuntu:22.04`
- Container Type: Portunix Container integration
- Network: Enabled (for package downloads)
- Privileges: Regular user + sudo

**Setup Command:**
```bash
./portunix container run ubuntu
# Inside container:
# - Copy portunix binary
# - Copy ptx-installer helper
# - Copy assets directory
```

---

## Test Cases

### TC001: Package Registry Loading ✅
**Description**: Verify package registry loads successfully
**Prerequisites**: None
**Steps**:
1. Run `./ptx-installer package list`
2. Verify 34+ packages displayed
3. Check package metadata (name, description, category)

**Expected Result**:
- Registry loads without errors
- All 34 packages displayed with correct metadata
- No validation errors

---

### TC002: Package Listing via Dispatcher ✅
**Description**: Verify dispatcher routes package commands correctly
**Prerequisites**: None
**Steps**:
1. Run `./portunix package list`
2. Verify routing to ptx-installer
3. Check output format

**Expected Result**:
- Dispatcher successfully routes to ptx-installer
- Package list displayed correctly
- No dispatcher overhead issues

---

### TC003: Dry-Run Installation (Archive) ✅
**Description**: Verify dry-run mode for tar.gz packages
**Prerequisites**: None
**Steps**:
1. Run `./portunix install hugo --dry-run`
2. Verify package detection
3. Check variant resolution
4. Verify no actual installation occurs

**Expected Result**:
- Package: Hugo Static Site Generator detected
- Platform: linux, Variant: standard, Type: tar.gz
- Dry-run message displayed
- No files downloaded or extracted

---

### TC004: Dry-Run Installation (APT) ✅
**Description**: Verify dry-run mode for APT packages
**Prerequisites**: None
**Steps**:
1. Run `./portunix install gh --dry-run`
2. Verify package detection
3. Check variant resolution for APT

**Expected Result**:
- Package: GitHub CLI detected
- Platform: linux, Variant: apt, Type: apt
- Dry-run message displayed
- No apt-get commands executed

---

### TC005: Real Archive Installation 🧪
**Description**: Test actual tar.gz package installation
**Prerequisites**: Clean container environment
**Package**: hugo (Hugo Static Site Generator)
**Steps**:
1. Run `./portunix install hugo` (no --dry-run)
2. Verify download to cache
3. Verify extraction to ~/.local/share/portunix/packages/
4. Verify binary symlink in ~/.local/bin/
5. Test binary execution: `~/.local/bin/hugo version`

**Expected Result**:
- Archive downloaded to ~/.portunix/cache/
- Files extracted successfully
- Symlink created: ~/.local/bin/hugo
- Hugo binary executes successfully
- Version command returns Hugo version

---

### TC006: Real APT Installation 🧪
**Description**: Test actual APT package installation
**Prerequisites**: Container with apt-get available
**Package**: curl (or lightweight utility)
**Steps**:
1. Run `./portunix install curl` (if package exists in registry)
2. Verify apt-get update executed
3. Verify apt-get install executed
4. Test installed package: `curl --version`

**Expected Result**:
- APT update successful
- Package installed via apt-get
- Binary available in PATH
- Version command works

---

### TC007: Dependency Resolution 🧪
**Description**: Test dependency resolution for packages with prerequisites
**Prerequisites**: Clean container
**Package**: Package with dependencies (if available)
**Steps**:
1. Identify package with dependencies in registry
2. Run installation
3. Verify dependency order resolution
4. Check dependency installation sequence

**Expected Result**:
- Dependencies identified correctly
- Installation order calculated (topological sort)
- Circular dependency detection works (if applicable)
- Dependencies installed before main package

---

### TC008: Platform Detection 🧪
**Description**: Verify platform and architecture detection
**Prerequisites**: Container environment
**Steps**:
1. Run any installation command
2. Check platform detection output
3. Verify correct variant selection for platform

**Expected Result**:
- Platform: linux detected
- Architecture: amd64 (or container arch)
- Correct variant selected for Linux platform
- No Windows/Darwin variants attempted

---

### TC009: Download Manager 🧪
**Description**: Test download functionality and caching
**Prerequisites**: Clean cache directory
**Steps**:
1. Install package requiring download (hugo)
2. Verify cache directory creation: ~/.portunix/cache/
3. Check downloaded file exists
4. Verify filename detection from headers
5. Attempt re-download (should use cache)

**Expected Result**:
- Cache directory created
- File downloaded to cache
- Proper filename extracted
- File size matches (progress displayed)
- Cache reuse works (if implemented)

---

### TC010: Error Handling 🧪
**Description**: Test error scenarios and messages
**Prerequisites**: Various conditions
**Test Scenarios**:
1. Invalid package name: `./portunix install nonexistent-pkg`
2. Missing variant: `./portunix install hugo --variant=invalid`
3. Network failure simulation (if possible)
4. Permission denied scenarios

**Expected Result**:
- Clear error messages for each scenario
- Graceful failure (no panics)
- Helpful suggestions in error messages
- Proper exit codes

---

## Test Execution Protocol

### Phase 1: Pre-Installation Verification
- [ ] Build all binaries successfully
- [ ] Verify binary sizes (portunix < 50MB, ptx-installer exists)
- [ ] Create clean Ubuntu 22.04 container
- [ ] Copy binaries and assets to container

### Phase 2: Basic Functionality Tests
- [ ] TC001: Package Registry Loading
- [ ] TC002: Package Listing via Dispatcher
- [ ] TC003: Dry-Run Installation (Archive)
- [ ] TC004: Dry-Run Installation (APT)

### Phase 3: Real Installation Tests
- [ ] TC005: Real Archive Installation
- [ ] TC006: Real APT Installation (if applicable)
- [ ] TC007: Dependency Resolution
- [ ] TC008: Platform Detection

### Phase 4: Advanced Tests
- [ ] TC009: Download Manager
- [ ] TC010: Error Handling

### Phase 5: Post-Installation Verification
- [ ] Verify no host system contamination
- [ ] Document all test results
- [ ] Create acceptance protocol

---

## Success Criteria

### Mandatory Requirements:
- [x] All TC001-TC004 (Dry-run tests) PASS
- [ ] TC005 (Archive installation) PASS
- [ ] Package registry loads without errors
- [ ] Dispatcher routing works correctly
- [ ] No crashes or panics during testing

### Optional Requirements:
- [ ] TC006-TC010 PASS
- [ ] Installation speed acceptable (< 2 minutes for small packages)
- [ ] Clear progress indicators displayed
- [ ] Error messages are helpful and actionable

---

## Risk Assessment

**High Risk:**
- Archive extraction might fail on complex archives
- APT installation might require sudo (container testing needed)
- Download failures on network issues

**Medium Risk:**
- Binary symlink creation might fail
- PATH not updated automatically
- Cache directory permissions

**Low Risk:**
- Package listing display format
- Dry-run accuracy
- Version detection

---

## Test Data

**Test Packages:**
1. **hugo** - Archive (tar.gz), ~15MB, no dependencies
2. **gh** - APT package, system package manager
3. **act** - Archive (tar.gz), ~10MB, no dependencies
4. **curl** - APT package, lightweight utility (if in registry)

**Container Environment:**
- OS: Ubuntu 22.04 LTS
- User: non-root with sudo
- Network: Enabled
- Disk: Sufficient for downloads (~100MB)

---

## Notes

- All tests MUST be performed in containers (per ISSUE-DEVELOPMENT-METHODOLOGY.md)
- No testing on host development system
- Document container OS in acceptance protocol
- Clean container for each critical test to ensure reproducibility

---

**Test Plan Version**: 1.0
**Status**: Ready for Execution
**Approval**: Pending test execution
