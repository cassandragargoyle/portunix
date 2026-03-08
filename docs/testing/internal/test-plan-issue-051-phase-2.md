# Test Plan - Issue #051 Phase 2: Helper Binaries Implementation

**Branch**: `feature/v1.6-issue-051-git-dispatcher-architecture`
**Version**: 1.6.0 (Phase 2)
**Date**: 2025-09-19

## Test Objectives
Verify Phase 2 implementation of Git-like dispatcher architecture with helper binaries extraction (ptx-container, ptx-mcp).

## Prerequisites
- [x] Phase 1 completed and tested (dispatcher infrastructure)
- [ ] Build Portunix with helper binaries from feature branch
- [ ] Test on both Windows and Linux platforms

## Test Cases

### TC-01: Build System Verification
**Description**: Verify helper binaries build system
**Steps**:
1. Run `make build-helpers`
2. Run `make build-all`
3. Check binaries exist: `ls -la ptx-*`
4. Verify file sizes are reasonable (> 5MB each)

**Expected Result**:
- ptx-container and ptx-mcp binaries created
- Both binaries executable and properly sized

### TC-02: Helper Binary Versions
**Description**: Verify helper binaries report correct version
**Steps**:
1. Run `./ptx-container --version`
2. Run `./ptx-mcp --version`
3. Compare with main binary: `./portunix --version`

**Expected Result**: All binaries show same version (dev during development)

### TC-03: Dispatcher Discovery
**Description**: Test dispatcher helper discovery
**Steps**:
1. Run `./portunix system dispatcher`
2. Run `./portunix system dispatcher --json`
3. Verify helper binaries are detected and available

**Expected Result**:
- Shows ptx-container and ptx-mcp as available
- Lists correct commands for each helper
- JSON output is valid

### TC-04: Container Command Delegation
**Description**: Test container command delegation to ptx-container
**Steps**:
1. Run `./portunix container`
2. Run `./portunix container run ubuntu`
3. Run `./portunix docker`
4. Run `./portunix podman`

**Expected Result**:
- Commands delegated to ptx-container helper
- Shows appropriate help and command handling
- No error messages or fallback to main binary

### TC-05: MCP Command Delegation
**Description**: Test MCP command delegation to ptx-mcp
**Steps**:
1. Run `./portunix mcp`
2. Run `./portunix mcp start`
3. Run `./portunix mcp status`
4. Run `./portunix mcp serve --help`

**Expected Result**:
- Commands delegated to ptx-mcp helper
- Shows appropriate help and command handling
- No error messages or fallback to main binary

### TC-06: Version Compatibility Validation
**Description**: Test version validation between main and helpers
**Steps**:
1. Rename ptx-container to ptx-container.bak
2. Create fake ptx-container with different version
3. Run `./portunix container`
4. Restore original ptx-container

**Expected Result**:
- Version mismatch detected (if implemented)
- Graceful handling of missing/incompatible helpers

### TC-07: Fallback Behavior
**Description**: Test behavior when helpers are missing
**Steps**:
1. Rename ptx-container to ptx-container.bak
2. Run `./portunix container`
3. Check if command falls back to main binary
4. Restore ptx-container

**Expected Result**: Appropriate fallback or error handling

### TC-08: Direct Helper Execution
**Description**: Test direct helper binary execution
**Steps**:
1. Run `./ptx-container container run ubuntu`
2. Run `./ptx-mcp mcp status`
3. Test helper help: `./ptx-container --help`

**Expected Result**:
- Helpers work when called directly
- Proper command parsing and execution
- Help system functions correctly

### TC-09: Performance Validation
**Description**: Verify no performance degradation with dispatcher
**Steps**:
1. Time commands: `time ./portunix container`
2. Time commands: `time ./portunix mcp`
3. Compare with Phase 1 baseline

**Expected Result**:
- Commands execute quickly (< 200ms)
- No significant performance degradation vs Phase 1

### TC-10: Backward Compatibility
**Description**: Ensure non-helper commands still work
**Steps**:
1. Run `./portunix --help`
2. Run `./portunix install --help`
3. Run `./portunix plugin --help`
4. Run `./portunix system info`

**Expected Result**: All non-helper commands work unchanged

### TC-11: Error Handling
**Description**: Test error handling in delegation
**Steps**:
1. Run invalid helper command: `./portunix container invalid-subcommand`
2. Run helper with wrong args: `./portunix mcp --invalid-flag`
3. Test permission issues (if applicable)

**Expected Result**: Appropriate error messages, no crashes

### TC-12: Cross-Platform Helper Execution
**Description**: Test helpers work on target platform
**Steps**:
1. Build for current platform
2. Test all helper commands
3. Verify file paths and execution work correctly

**Expected Result**: Helpers work correctly on test platform

## Phase 2 Specific Validations

### Helper Binary Architecture
- [x] Helper binaries created under `src/helpers/`
- [x] Separate go.mod for each helper
- [x] Independent build targets in Makefile
- [x] Version synchronization with main binary

### Dispatcher Integration
- [x] Commands correctly routed to helpers
- [x] Arguments passed correctly to helpers
- [x] Helper discovery working
- [x] Version validation implemented

### Build System Integration
- [x] `make build-helpers` target added
- [x] `make build-all` builds main + helpers
- [x] Helper binaries placed in root directory
- [x] Cross-platform build support

## Testing Status

### Platform Testing Results
- **Linux**: ⏳ PENDING (Assign to Linux tester)
- **Windows**: ⏳ PENDING (Assign to Windows tester)

### Performance Baseline
- **Phase 1 Baseline**: ~20ms (Linux), ~500ms (Windows)
- **Phase 2 Target**: No significant degradation (< 50ms increase)

## Test Environment
- **OS**: Linux Ubuntu 22.04 / Windows 11
- **Go Version**: 1.23.3
- **Test Type**: Integration
- **Required Tools**: make, go build tools

## Notes
- Phase 2 introduces helper binaries but keeps fallback compatibility
- All existing functionality must remain working
- Helper binaries should NOT be called directly by users
- Dispatcher handles all routing transparently

## Regression Testing
Include all Phase 1 test cases to ensure no regression:
- [ ] All Phase 1 tests pass
- [ ] Dispatcher info shows helper binaries
- [ ] No breaking changes to existing commands

## Sign-off Template

**Linux Tester**: _______________
**Date**: _______________
**Result**: [ ] PASS [ ] FAIL [ ] CONDITIONAL
**Notes**: _______________

**Windows Tester**: _______________
**Date**: _______________
**Result**: [ ] PASS [ ] FAIL [ ] CONDITIONAL
**Notes**: _______________

**Cross-Platform Result**: [ ] PASS [ ] FAIL [ ] CONDITIONAL
**Ready for Phase 3**: [ ] YES [ ] NO

---

**Test Plan Status**: Ready for execution
**Created by**: Developer (Phase 2 implementation)
**Review Required**: QA Team