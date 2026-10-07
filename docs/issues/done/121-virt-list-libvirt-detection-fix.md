# Issue #121: Fix libvirt Detection in `portunix virt list`

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Bug
**Labels**: bug, ptx-virt, libvirt, detection
**Created**: 2026-01-02
**Assigned to**: Development Team

## Summary

`portunix virt list` incorrectly reports that libvirt is not installed even when it is properly installed and running.

## Problem Statement

**Current behavior:**
```bash
$ ./portunix virt list
Listing virtual machines...
Error listing VMs: VM listing requires libvirt (use: portunix virt install-qemu)
```

**Expected behavior:**
```bash
$ ./portunix virt list
NAME       STATE    IP
myvm       running  192.168.122.100
```

## Environment Verification

Libvirt is correctly installed and functional:

```bash
$ which virsh
/usr/bin/virsh

$ virsh version
Compiled against library: libvirt 11.0.0
Using library: libvirt 11.0.0
Using API: QEMU 11.0.0
Running hypervisor: QEMU 9.2.1

$ systemctl status libvirtd
● libvirtd.service - libvirt legacy monolithic daemon
     Active: active (running)

$ ls -la /var/run/libvirt/libvirt-sock
srw-rw---- 1 root libvirt 0 ... /var/run/libvirt/libvirt-sock
```

## Root Cause Analysis

**CONFIRMED**: Investigation completed 2026-01-02.

### The Problem

The issue is **NOT** with detection - libvirt IS detected correctly. The problem is that `QEMUAdapter` has **stub implementations** that always return errors.

### Technical Details

1. **`DetectLibvirtStatus()`** in `src/app/virt/libvirt_detector.go` works correctly:
   - Detects `virsh` binary via `exec.LookPath()`
   - Checks daemon status (libvirtd or virtqemud)
   - Checks socket activation
   - Returns complete status with version, recommendations, etc.

2. **`QEMUAdapter.GetDiagnosticInfo()`** correctly uses `DetectLibvirtStatus()` for diagnostic output in `virt check` command.

3. **`QEMUAdapter.List()`** and other VM management methods are **STUBS**:
   ```go
   func (q *QEMUAdapter) List() ([]*types.VMInfo, error) {
       // Basic QEMU adapter doesn't track VMs
       // This would require libvirt integration
       return nil, fmt.Errorf("VM listing requires libvirt (use: portunix virt install-qemu)")
   }
   ```

4. **Provider Selection Flow**:
   - `manager.go:selectProvider()` detects QEMU is available
   - Creates `QEMUAdapter` instance
   - `QEMUAdapter.IsAvailable()` returns `true` (QEMU binary found)
   - But all VM operations are stubs that return errors

### Code Locations

| File | Function | Status |
|------|----------|--------|
| `src/app/virt/libvirt_detector.go` | `DetectLibvirtStatus()` | ✅ Works correctly |
| `src/helpers/ptx-virt/qemu_adapter.go:54-106` | `GetDiagnosticInfo()` | ✅ Uses detection |
| `src/helpers/ptx-virt/qemu_adapter.go:186-190` | `List()` | ❌ **Stub - always fails** |
| `src/helpers/ptx-virt/qemu_adapter.go:149-181` | `Create/Start/Stop/...` | ❌ **All stubs** |
| `src/helpers/ptx-virt/manager.go:80-87` | `createProvider()` | Creates QEMUAdapter |

### Why This Happened

The `QEMUAdapter` was designed as a "detection-only" adapter to check if QEMU/KVM is installed and provide diagnostic info. The actual VM management was intended to use libvirt, but `LibvirtAdapter` was never implemented.

## Proposed Solutions

### Option A: Enhance QEMUAdapter with libvirt support (Recommended)

Modify `QEMUAdapter` to use `virsh` commands when libvirt is available:

```go
func (q *QEMUAdapter) List() ([]*types.VMInfo, error) {
    // Check if libvirt is available
    status, err := virt.DetectLibvirtStatus()
    if err != nil || !status.Installed {
        return nil, fmt.Errorf("VM listing requires libvirt (use: portunix install libvirt)")
    }

    if !status.Running && !status.SocketActivated {
        return nil, fmt.Errorf("libvirt daemon not running (use: portunix virt check --fix-libvirt)")
    }

    // Use virsh to list VMs
    cmd := exec.Command("virsh", "list", "--all")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to list VMs: %w", err)
    }

    return parseVirshListOutput(output), nil
}
```

### Option B: Create dedicated LibvirtAdapter

Create new `libvirt_adapter.go` that wraps `virsh` commands:

```go
type LibvirtAdapter struct {
    status *virt.LibvirtStatus
}

func (l *LibvirtAdapter) IsAvailable() bool {
    status, err := virt.DetectLibvirtStatus()
    if err != nil {
        return false
    }
    l.status = status
    return status.Installed && (status.Running || status.SocketActivated)
}

func (l *LibvirtAdapter) List() ([]*types.VMInfo, error) {
    // Use virsh list --all
}
```

Update `manager.go:createProvider()` to return `LibvirtAdapter` when libvirt is detected.

### Option C: Improve error messaging (Minimal fix)

At minimum, improve error messages to explain the actual situation:

```go
func (q *QEMUAdapter) List() ([]*types.VMInfo, error) {
    status, _ := virt.DetectLibvirtStatus()
    if status != nil && status.Installed {
        if !status.Running && !status.SocketActivated {
            return nil, fmt.Errorf("libvirt detected but not running - start with: sudo systemctl start libvirtd")
        }
        // Libvirt is ready but operations not implemented
        return nil, fmt.Errorf("libvirt VM management not yet implemented in portunix (use virsh directly)")
    }
    return nil, fmt.Errorf("VM listing requires libvirt (use: portunix install libvirt)")
}
```

## Implementation Effort

| Option | Effort | Coverage |
|--------|--------|----------|
| Option A | Medium | `list` command only |
| Option B | High | Full VM lifecycle management |
| Option C | Low | Error messages only |

**Recommendation**: Start with Option A for `list` command, then gradually extend to other operations.

## Acceptance Criteria

- [ ] `portunix virt list` works when libvirt is installed and running
- [ ] `portunix virt list` shows clear error when libvirt is installed but not running
- [ ] `portunix virt list` shows clear error when libvirt is NOT installed
- [ ] Clear error message for permission issues (user not in libvirt group)
- [ ] Works on Ubuntu, Fedora, Debian
- [ ] Handles both monolithic (libvirtd) and modular (virtqemud) daemon types

## Related Files

- `src/helpers/ptx-virt/qemu_adapter.go` - **MAIN TARGET** - Stub implementations to fix
- `src/helpers/ptx-virt/manager.go` - Provider selection logic
- `src/helpers/ptx-virt/provider.go` - VirtualizationProvider interface
- `src/app/virt/libvirt_detector.go` - Libvirt detection (already works)

## Testing

```bash
# Test with libvirt installed
./portunix virt list

# Test detection
./portunix virt check

# Verbose mode if available
./portunix virt list --verbose
```

---

**Created**: 2026-01-02
**Last Updated**: 2026-01-02
