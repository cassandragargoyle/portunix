# Issue #019: Docker Installation Issues on Windows

## Overview
**Type**: Bug  
**Priority**: High  
**Status**: 🔧 In Progress  
**GitHub Issue**: [#19](https://github.com/cassandragargoyle/portunix/issues/19)  
**Reporter**: @rjankovic  
**Assigned**: Development Team  
**Created**: 2024-11-30  
**Target Version**: 1.5.4

## Problem Description

Docker installation on Windows has multiple critical issues:

1. **Incorrect Disk Path Detection**: System recommends non-existent D:\ drive when only C:\ exists
2. **Wrong Docker Configuration**: Sets `data-root: D:\docker-data` even when C:\ is selected  
3. **Installation Verification Fails**: Docker service verification doesn't work properly
4. **Inconsistent Commands**: `portunix docker install` and `portunix install docker` behave differently
5. **PATH Configuration Issues**: Docker isn't properly added to PATH on Windows

## Environment
- **OS**: Windows 10.0.26100
- **Portunix Version**: 1.5.0
- **Docker**: Desktop installation

## Reproduction Steps

1. Have a Windows system with only C:\ drive
2. Run `portunix docker install`
3. Observe:
   - System suggests D:\ drive (non-existent)
   - Selecting C:\ still configures D:\docker-data
   - Installation verification fails
   - Docker not in PATH

## Root Cause Analysis

### Disk Detection Issue
- Windows disk enumeration logic incorrectly identifies available drives
- Hardcoded D:\ preference in recommendation logic
- No validation of disk existence before recommendation

### Configuration Issue  
- Docker daemon.json configuration doesn't respect user selection
- Path construction logic has bugs in Windows path handling

### Command Inconsistency
- Two different code paths for `docker install` vs `install docker`
- Different verification logic in each path

## Solution Design

### 1. Fix Disk Detection
- Properly enumerate Windows drives using WMI or PowerShell
- Validate drive existence before recommendation
- Default to system drive if no data drives found

### 2. Fix Configuration Path
- Respect user's drive selection in daemon.json
- Properly construct Windows paths with backslashes
- Validate paths before writing configuration

### 3. Unify Command Behavior
- Single code path for Docker installation
- Consistent verification steps
- Clear success/failure messages

### 4. Fix PATH Configuration
- Properly add Docker to Windows PATH
- Handle both user and system PATH
- Verify PATH changes

## Implementation Plan

1. **Create comprehensive test suite** (test_docker_windows.go)
2. **Fix disk detection logic** (app/docker/docker_win.go)
3. **Fix configuration generation** (app/docker/config_win.go)
4. **Unify installation commands** (cmd/docker_install.go)
5. **Improve PATH handling** (app/system/path_win.go)
6. **Add better error messages**

## Test Cases

### Test 1: Single Drive Detection
- Mock system with only C:\ drive
- Verify correct recommendation (C:\)
- Verify no phantom drives

### Test 2: Configuration Generation
- Select C:\ drive
- Verify daemon.json has C:\docker-data
- Verify paths are valid Windows paths

### Test 3: Command Consistency
- Test both command variants
- Verify identical behavior
- Verify consistent output

### Test 4: PATH Configuration
- Install Docker
- Verify PATH updated
- Verify docker.exe accessible

## Success Criteria

- [ ] Correct drive detection on single-drive systems
- [ ] Configuration respects user selection
- [ ] Both command variants work identically
- [ ] Docker properly added to PATH
- [ ] Clear success/failure messages
- [ ] All tests pass on Windows

## Notes

User feedback indicates the tool is "fairly straightforward" despite issues, suggesting good UX potential once bugs are fixed.

## References

- GitHub Issue: https://github.com/cassandragargoyle/portunix/issues/19
- Related: Docker Desktop Windows documentation
- Windows PATH management best practices