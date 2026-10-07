# Issue #60: Backend Version Display Enhancement

**Status**: ✅ Implemented

## Summary
Enhance `portunix virt list` and `portunix system info` commands to display version information for detected backends (Docker, Podman, QEMU, VirtualBox) alongside their availability status.

## Current Behavior
- `portunix virt list` shows backend availability (✅/❌) but not versions
- `portunix system info` displays system information without backend versions

## Expected Behavior
- Display backend versions in format: `✅ Docker (v24.0.7)`
- Show versions for all detected virtualization backends
- Maintain backward compatibility with existing output format

## Affected Commands
1. **`portunix virt list`**
   - Currently: `Backend: virtualbox`
   - Target: `Backend: virtualbox (v24.0.7)` 

2. **`portunix system info`**
   - Add backend versions to system information output
   - Include versions for: Docker, Podman, QEMU, VirtualBox
   - Currently: Docker:       installed
   - Target: Docker:       v24.0.7 or installed (if version unknown)
   The version should be displayed in json (--json parameter) as well.

## Implementation Requirements

### Technical Requirements
- Add version detection functions for each backend
- Parse version output from backend binaries
- Handle cases where backend is available but version cannot be determined
- Maintain performance - version checks should be fast
- Cache version information during single command execution

### Backend Version Commands
- **Docker**: `docker version --format '{{.Server.Version}}'`
- **Podman**: `podman version --format '{{.Version}}'`
- **QEMU**: `qemu-system-x86_64 --version`
- **VirtualBox**: `VBoxManage --version`

### Error Handling
- If version cannot be determined: `✅ Docker (unknown version)`
- If backend not available: `❌ Docker (not installed)`
- If version command fails: `✅ Docker (version check failed)`

### Output Format Examples

#### virt list Output
```
Available Virtualization Backends:
✅ Docker (v24.0.7)
✅ Podman (v4.6.1)
❌ QEMU (not installed)
✅ VirtualBox (v7.0.12)
```

#### system info Output
```
System Information:
OS: Windows 11 Pro
Architecture: amd64
Go Version: go1.21.4

Container Runtimes:
  Docker: v24.0.7
  Podman: Not installed

Virtualization:
  QEMU/KVM: Not installed
  VirtualBox: v7.0.12
```

## Implementation Notes
- Create utility functions in appropriate modules:
  - `app/docker/version.go` - Docker version detection
  - `app/podman/version.go` - Podman version detection
  - `app/system/virtualization.go` - QEMU/VirtualBox version detection
- Use existing backend detection logic as foundation
- Consider adding `--no-version` flag for performance-critical scenarios

## Acceptance Criteria
- [ ] `portunix virt list` displays backend versions when available
- [ ] `portunix system info` includes backend version information
- [ ] Version detection works on Windows and Linux
- [ ] Performance impact is minimal (< 100ms additional execution time)
- [ ] Error handling works correctly for all edge cases
- [ ] Backward compatibility maintained
- [ ] Help text updated to reflect new version display feature

## Testing Requirements
- Test version detection for all supported backends
- Test behavior when backends are not installed
- Test behavior when version commands fail
- Test cross-platform compatibility (Windows/Linux)
- Performance testing to ensure minimal overhead

## Priority
**Medium** - Enhancement that improves user experience and system visibility

## Labels
- enhancement
- system-info
- virtualization
- docker
- podman
- user-experience

## Related Issues
- Builds upon existing backend detection functionality
- Complements #048 (System Info Enhanced Container Detection)

## Estimated Effort
- **Development**: 4-6 hours
- **Testing**: 2-3 hours
- **Documentation**: 1 hour

---

**Created**: 2025-09-24
**Status**: 📋 Open
**Priority**: Medium
**Type**: Enhancement