# Issue #89: QEMU/KVM Adapter Implementation for virt check

## Overview
**Type**: Bug Fix / Feature Implementation
**Priority**: High
**Status**: 📋 Open
**Component**: Virtualization, QEMU/KVM, ptx-virt Helper
**Platform**: Linux
**Created**: 2025-10-04
**Target Version**: 1.6.x

## Problem Description

`portunix virt check` does **NOT** detect QEMU/KVM installation despite:
- ✅ QEMU being installed (`/usr/bin/qemu-system-x86_64`)
- ✅ KVM modules being loaded (`kvm_amd`, `kvm`)
- ✅ Helper function `system.GetQEMUVersion()` exists and works

**Current Output:**
```
$ portunix virt check
Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: virtualbox

Available Providers:
  ✅ virtualbox (v7.0.20_Ubuntu) at /usr/bin/VBoxManage
```

**Expected Output:**
```
$ portunix virt check
Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: qemu/kvm

Available Providers:
  ✅ qemu/kvm (v8.0.0) at /usr/bin/qemu-system-x86_64
    └─ KVM acceleration: enabled (kvm_amd loaded)
  ✅ virtualbox (v7.0.20_Ubuntu) at /usr/bin/VBoxManage
    ⚠️  Conflict: KVM modules loaded (see: portunix virt check --fix)
```

## Root Cause Analysis

**File:** `src/helpers/ptx-virt/manager.go:86-88`

```go
case "qemu", "kvm":
    // TODO: Implement QEMU/KVM provider adapter
    return nil  // ❌ Returns nil, causing QEMU/KVM to be skipped
```

**Flow:**
1. `CheckCapabilities()` iterates through providers: `["virtualbox", "qemu", "kvm", "vmware", "hyperv"]`
2. Calls `m.createProvider("qemu")` → returns `nil`
3. Skips QEMU (line 231-232): `if provider == nil { continue }`
4. Result: QEMU/KVM never appears in output

**Existing Infrastructure:**
- ✅ `system.GetQEMUVersion()` - already implemented (system/virtualization.go:12)
- ✅ `system.GetLibvirtVersion()` - already implemented (system/virtualization.go:88)
- ✅ `VirtualizationProvider` interface - defined in provider.go
- ✅ `VirtualBoxAdapter` - working reference implementation

## Solution Design

### Phase 1: Create QEMU/KVM Adapter

**New File:** `src/helpers/ptx-virt/qemu_adapter.go`

```go
package main

import (
    "fmt"
    "os/exec"
    "strings"

    "portunix.ai/app/system"
)

// QEMUAdapter implements VirtualizationProvider interface for QEMU/KVM
type QEMUAdapter struct {
    hasKVM bool
}

// IsAvailable checks if QEMU is installed
func (q *QEMUAdapter) IsAvailable() bool {
    // Check if qemu-system-x86_64 exists
    _, err := exec.LookPath("qemu-system-x86_64")
    if err != nil {
        // Try alternative 'kvm' command
        _, err = exec.LookPath("kvm")
    }

    // Check KVM availability
    q.hasKVM = q.checkKVMModules()

    return err == nil
}

// GetVersion returns QEMU version
func (q *QEMUAdapter) GetVersion() (string, error) {
    version := system.GetQEMUVersion()
    if version == "" {
        return "", fmt.Errorf("failed to get QEMU version")
    }
    // Remove 'v' prefix if present
    return strings.TrimPrefix(version, "v"), nil
}

// GetDiagnosticInfo returns diagnostic information
func (q *QEMUAdapter) GetDiagnosticInfo() *DiagnosticInfo {
    diag := &DiagnosticInfo{
        Suggestions: []string{},
    }

    // Find QEMU path
    if path, err := exec.LookPath("qemu-system-x86_64"); err == nil {
        diag.PathEnvironment = path
    } else if path, err := exec.LookPath("kvm"); err == nil {
        diag.PathEnvironment = path
    }

    // Check KVM acceleration
    if q.hasKVM {
        diag.Details = append(diag.Details, "KVM acceleration: enabled")

        // Check which KVM module is loaded
        if kmods := q.getLoadedKVMModules(); len(kmods) > 0 {
            diag.Details = append(diag.Details, fmt.Sprintf("Loaded modules: %s", strings.Join(kmods, ", ")))
        }
    } else {
        diag.Details = append(diag.Details, "KVM acceleration: disabled (software emulation only)")
        diag.Suggestions = append(diag.Suggestions, "Load KVM kernel module: sudo modprobe kvm kvm_amd (or kvm_intel)")
    }

    // Check for VirtualBox conflict
    if q.hasKVM && q.isVirtualBoxConflict() {
        diag.Suggestions = append(diag.Suggestions, "⚠️  VirtualBox/KVM conflict detected. Run: portunix virt check --fix")
    }

    return diag
}

// checkKVMModules checks if KVM kernel modules are loaded
func (q *QEMUAdapter) checkKVMModules() bool {
    cmd := exec.Command("lsmod")
    output, err := cmd.Output()
    if err != nil {
        return false
    }

    return strings.Contains(string(output), "kvm")
}

// getLoadedKVMModules returns list of loaded KVM modules
func (q *QEMUAdapter) getLoadedKVMModules() []string {
    cmd := exec.Command("lsmod")
    output, err := cmd.Output()
    if err != nil {
        return nil
    }

    modules := []string{}
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, "kvm") {
            parts := strings.Fields(line)
            if len(parts) > 0 {
                modules = append(modules, parts[0])
            }
        }
    }
    return modules
}

// isVirtualBoxConflict checks if VirtualBox is installed (potential conflict)
func (q *QEMUAdapter) isVirtualBoxConflict() bool {
    _, err := exec.LookPath("VBoxManage")
    return err == nil
}

// List returns list of VMs (not implemented for basic adapter)
func (q *QEMUAdapter) List() ([]*VirtualMachine, error) {
    // Basic QEMU adapter doesn't track VMs
    // This would require libvirt integration
    return nil, fmt.Errorf("VM listing requires libvirt (use: portunix virt install-qemu)")
}

// Additional VirtualizationProvider interface methods...
// (Implement stubs for other required methods)
```

### Phase 2: Update Manager to Use QEMU Adapter

**File:** `src/helpers/ptx-virt/manager.go`

```go
func (m *VirtManager) createProvider(name string) VirtualizationProvider {
    switch name {
    case "virtualbox":
        return &VirtualBoxAdapter{
            backend: virtualbox.NewBackend(),
        }
    case "qemu", "kvm":
        // ✅ NEW: Create QEMU adapter instead of returning nil
        return &QEMUAdapter{}
    case "vmware":
        // TODO: Implement VMware provider adapter
        return nil
    case "hyperv":
        // TODO: Implement Hyper-V provider adapter
        return nil
    default:
        return nil
    }
}
```

### Phase 3: Enhanced Output with KVM Details

**Update:** `src/helpers/ptx-virt/main.go`

Enhance display to show KVM-specific information:

```go
func handleCheckCommand() {
    fmt.Println("Checking virtualization support...")

    // ... existing code ...

    fmt.Println("\nAvailable Providers:")
    for _, provider := range capabilities.AvailableProviders {
        status := "❌"
        if provider.Available {
            status = "✅"
        }
        fmt.Printf("  %s %s", status, provider.Name)
        if provider.Version != "" {
            fmt.Printf(" (v%s)", provider.Version)
        }
        if provider.InstallationPath != "" {
            fmt.Printf(" at %s", provider.InstallationPath)
        }
        fmt.Println()

        // Show additional details for QEMU/KVM
        if provider.Name == "qemu" || provider.Name == "kvm" {
            if len(provider.Details) > 0 {
                for _, detail := range provider.Details {
                    fmt.Printf("    └─ %s\n", detail)
                }
            }
        }

        // Show warnings/suggestions
        if len(provider.Recommendations) > 0 {
            for _, rec := range provider.Recommendations {
                fmt.Printf("    %s\n", rec)
            }
        }
    }
}
```

### Phase 4: Provider Priority Update

**File:** `src/helpers/ptx-virt/provider.go`

Update provider priorities for Linux to prefer QEMU/KVM when available:

```go
func GetProviderPriority(platform string) []string {
    switch platform {
    case "linux":
        return []string{
            "kvm",        // ✅ Prefer KVM on Linux (native, best performance)
            "qemu",       // QEMU with KVM acceleration
            "virtualbox", // VirtualBox (conflicts with KVM)
            "vmware",
        }
    case "windows":
        return []string{
            "hyperv",
            "virtualbox",
            "vmware",
        }
    case "darwin":
        return []string{
            "vmware",
            "virtualbox",
        }
    default:
        return []string{"virtualbox", "vmware", "hyperv", "qemu"}
    }
}
```

## Implementation Checklist

### Phase 1: QEMU Adapter
- [ ] Create `qemu_adapter.go` in `src/helpers/ptx-virt/`
- [ ] Implement `QEMUAdapter` struct
- [ ] Implement `IsAvailable()` method
- [ ] Implement `GetVersion()` method using `system.GetQEMUVersion()`
- [ ] Implement `GetDiagnosticInfo()` method
- [ ] Implement KVM module detection (`checkKVMModules()`)
- [ ] Implement VirtualBox conflict detection
- [ ] Implement stub methods for full interface compliance

### Phase 2: Integration
- [ ] Update `manager.go` `createProvider()` to return `QEMUAdapter`
- [ ] Update provider priority to prefer KVM on Linux
- [ ] Test QEMU detection on Linux with KVM
- [ ] Test QEMU detection on Linux without KVM

### Phase 3: Enhanced Display
- [ ] Update `handleCheckCommand()` to show KVM details
- [ ] Add KVM module information display
- [ ] Add conflict warnings when both VirtualBox and KVM detected
- [ ] Add suggestions for enabling KVM if not loaded

### Phase 4: Testing
- [ ] Test on system with QEMU + KVM
- [ ] Test on system with QEMU only (no KVM)
- [ ] Test on system with both VirtualBox and QEMU
- [ ] Verify conflict detection works
- [ ] Test version detection
- [ ] Verify diagnostic information accuracy

## Expected Behavior After Fix

### Scenario 1: QEMU with KVM (optimal)
```bash
$ portunix virt check
Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: kvm

Available Providers:
  ✅ kvm (v8.0.0) at /usr/bin/qemu-system-x86_64
    └─ KVM acceleration: enabled
    └─ Loaded modules: kvm_amd, kvm
  ✅ virtualbox (v7.0.20_Ubuntu) at /usr/bin/VBoxManage
    ⚠️  Conflict: KVM modules loaded (see: portunix virt check --fix)
```

### Scenario 2: QEMU without KVM
```bash
$ portunix virt check
Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: virtualbox

Available Providers:
  ✅ qemu (v8.0.0) at /usr/bin/qemu-system-x86_64
    └─ KVM acceleration: disabled (software emulation only)
    └─ Load KVM kernel module: sudo modprobe kvm kvm_amd
  ✅ virtualbox (v7.0.20_Ubuntu) at /usr/bin/VBoxManage
```

### Scenario 3: No QEMU
```bash
$ portunix virt check
Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: virtualbox

Available Providers:
  ❌ qemu (not installed)
    └─ Install QEMU: portunix virt install-qemu
  ✅ virtualbox (v7.0.20_Ubuntu) at /usr/bin/VBoxManage
```

## Files to Create/Modify

### New Files
1. `src/helpers/ptx-virt/qemu_adapter.go` - QEMU/KVM adapter implementation

### Modified Files
1. `src/helpers/ptx-virt/manager.go` - Update `createProvider()` to return `QEMUAdapter`
2. `src/helpers/ptx-virt/provider.go` - Update provider priorities for Linux
3. `src/helpers/ptx-virt/main.go` - Enhanced display with KVM details

## Test Cases

### TC001: QEMU Detection with KVM
```bash
# Setup: Install QEMU, load KVM modules
sudo apt install qemu-system-x86
sudo modprobe kvm kvm_amd

# Test
portunix virt check

# Expected:
# ✅ kvm detected with version
# ✅ KVM acceleration shown as enabled
# ✅ Loaded modules listed
```

### TC002: QEMU Detection without KVM
```bash
# Setup: Install QEMU, unload KVM
sudo rmmod kvm_amd kvm

# Test
portunix virt check

# Expected:
# ✅ qemu detected with version
# ✅ KVM acceleration shown as disabled
# ✅ Suggestion to load KVM modules
```

### TC003: VirtualBox/KVM Conflict Detection
```bash
# Setup: Both VirtualBox and KVM loaded
portunix virt check

# Expected:
# ✅ Both providers shown
# ⚠️  Conflict warning on VirtualBox
# ✅ Suggestion to run 'virt check --fix'
```

### TC004: Version Detection
```bash
# Verify QEMU version is correctly detected
qemu-system-x86_64 --version  # e.g., "version 8.0.0"
portunix virt check           # Should show: kvm (v8.0.0)
```

### TC005: No QEMU Installed
```bash
# Setup: Uninstall QEMU
sudo apt remove qemu-system-x86

# Test
portunix virt check

# Expected:
# ❌ qemu shown as not available
# ✅ Suggestion to install via portunix
```

## Success Criteria

- [ ] QEMU/KVM appears in `virt check` output when installed
- [ ] KVM module detection works correctly
- [ ] Version information displayed accurately
- [ ] KVM acceleration status shown (enabled/disabled)
- [ ] VirtualBox/KVM conflict detected and warned
- [ ] Helpful suggestions provided for KVM setup
- [ ] Provider priority prefers KVM on Linux when available
- [ ] No regression in VirtualBox detection
- [ ] All test cases pass

## Integration with Related Issues

This issue integrates with:
- **#088** - VirtualBox/KVM Conflict Detection and Resolution
  - QEMU adapter provides conflict detection
  - Warns when both VirtualBox and KVM active
  - Suggests `virt check --fix` for resolution

- **#049** - Full QEMU/KVM Support Implementation
  - This is Phase 1: Detection and Status
  - Phase 2 will be: VM management via libvirt

## Future Enhancements

1. **Libvirt Integration**: List VMs managed by libvirt
2. **VM Creation**: Support creating QEMU/KVM VMs
3. **Performance Monitoring**: Show KVM performance metrics
4. **Network Configuration**: Bridge/NAT setup for KVM VMs
5. **Advanced Features**: GPU passthrough, USB passthrough detection

## Notes

- QEMU adapter is intentionally basic (detection only)
- Full VM management will require libvirt integration (future issue)
- Focus on detection, version info, and conflict awareness
- Keep adapter lightweight and fast

## References

- Existing: `src/app/system/virtualization.go:12` - `GetQEMUVersion()`
- Existing: `src/app/system/virtualization.go:88` - `GetLibvirtVersion()`
- Reference: `src/helpers/ptx-virt/virtualbox_adapter.go` - Implementation pattern
- Interface: `src/helpers/ptx-virt/provider.go` - `VirtualizationProvider`

---

**Reporter**: User (virt check not showing QEMU/KVM)
**Assigned**: Development Team
**Labels**: bug, enhancement, virtualization, qemu, kvm, ptx-virt, detection
