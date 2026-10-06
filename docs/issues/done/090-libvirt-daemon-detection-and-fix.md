# Issue #90: Libvirt Daemon Detection and Auto-Fix

## Overview
**Type**: Bug Fix / Feature Enhancement
**Priority**: High
**Status**: 📋 Open
**Component**: Virtualization, QEMU/KVM, Libvirt
**Platform**: Linux
**Created**: 2025-10-04
**Target Version**: 1.6.x

## Problem Description

Virtual Machine Manager (virt-manager) hlásí **"QEMU/KVM - není připojeno"** i když:
- ✅ QEMU je nainstalovaný (`/usr/bin/qemu-system-x86_64`)
- ✅ KVM moduly jsou načtené (`kvm_amd`, `kvm`)
- ✅ `portunix virt check` detekuje QEMU/KVM jako dostupný

**Root Cause**:
```bash
$ systemctl status libvirtd
○ libvirtd.service
   Loaded: masked (Reason: Unit libvirtd.service is masked.)
   Active: inactive (dead)
```

Libvirt daemon je **masked** (zamaskovaný) a neběží. Virtual Machine Manager potřebuje běžící libvirt daemon pro správu QEMU/KVM virtuálních strojů.

## Current Behavior

**`portunix virt check` výstup:**
```
✅ kvm (v9.2.1) at /usr/bin/qemu-system-x86_64
  └─ KVM acceleration: enabled
  └─ Loaded modules: kvm_amd, kvm, irqbypass, ccp
```

**Virtual Machine Manager:**
```
QEMU/KVM - není připojeno
[Nelze se připojit k libvirt daemonu]
```

**Problém**: `virt check` nedetekuje, že libvirt daemon není spuštěný nebo je zamaskovaný.

## Expected Behavior

`portunix virt check` by měl:
1. ✅ Detekovat instalaci libvirt
2. ✅ Zjistit stav libvirt daemonu (běžící/zastavený/masked)
3. ✅ Varovat pokud daemon není dostupný
4. ✅ Nabídnout řešení přes `virt check --fix-libvirt`

## Technical Analysis

### Libvirt Daemon Variants

**Starší systémy (monolitický daemon):**
- `libvirtd.service` - hlavní daemon pro všechny hypervisory

**Novější systémy (modulární daemony):**
- `virtqemud.service` - QEMU/KVM specifický daemon
- `virtnetworkd.service` - síťový daemon
- `virtstoraged.service` - storage daemon

### Detekce Stavu

```bash
# Check if libvirtd exists and its status
systemctl list-unit-files | grep libvirtd
systemctl status libvirtd
systemctl is-enabled libvirtd
systemctl is-active libvirtd

# Check modular daemons (newer systems)
systemctl status virtqemud
systemctl status virtnetworkd

# Check if masked
systemctl show libvirtd -p LoadState,ActiveState,UnitFileState
```

### Možné Stavy

1. **Running** - daemon běží, vše OK
2. **Stopped** - daemon existuje, ale není spuštěn
3. **Masked** - daemon je zamaskovaný (nelze spustit)
4. **Not installed** - libvirt není nainstalovaný
5. **Socket-activated** - daemon je spuštěn přes socket aktivaci

## Solution Design

### Phase 1: Libvirt Detection (virt check enhancement)

**New File:** `src/app/virtualization/libvirt_detector.go`

```go
package virtualization

import (
    "os/exec"
    "strings"
)

// LibvirtStatus represents libvirt daemon status
type LibvirtStatus struct {
    Installed      bool
    DaemonType     string // "monolithic" (libvirtd) or "modular" (virtqemud)
    Running        bool
    Enabled        bool
    Masked         bool
    SocketActivated bool
    Version        string
    Issues         []string
    Recommendations []string
}

// DetectLibvirtStatus detects libvirt daemon status
func DetectLibvirtStatus() (*LibvirtStatus, error) {
    status := &LibvirtStatus{
        Issues:         []string{},
        Recommendations: []string{},
    }

    // Check libvirt-daemon package
    if _, err := exec.LookPath("virsh"); err == nil {
        status.Installed = true
    }

    if !status.Installed {
        status.Issues = append(status.Issues, "Libvirt is not installed")
        status.Recommendations = append(status.Recommendations, "Install libvirt: portunix install libvirt")
        return status, nil
    }

    // Get version
    if version := getLibvirtVersion(); version != "" {
        status.Version = version
    }

    // Try monolithic daemon first (older systems)
    if checkService("libvirtd") {
        status.DaemonType = "monolithic"
        status.Running = isServiceActive("libvirtd")
        status.Enabled = isServiceEnabled("libvirtd")
        status.Masked = isServiceMasked("libvirtd")
    } else if checkService("virtqemud") {
        // Try modular daemon (newer systems)
        status.DaemonType = "modular"
        status.Running = isServiceActive("virtqemud")
        status.Enabled = isServiceEnabled("virtqemud")
        status.Masked = isServiceMasked("virtqemud")
    }

    // Check socket activation
    if checkService("libvirtd.socket") {
        status.SocketActivated = isServiceActive("libvirtd.socket")
    }

    // Analyze issues
    if status.Masked {
        status.Issues = append(status.Issues, "Libvirt daemon is masked (cannot be started)")
        status.Recommendations = append(status.Recommendations, "Unmask and start: sudo portunix virt check --fix-libvirt")
    } else if !status.Running && !status.SocketActivated {
        status.Issues = append(status.Issues, "Libvirt daemon is not running")
        status.Recommendations = append(status.Recommendations, "Start daemon: sudo portunix virt check --fix-libvirt")
    } else if !status.Enabled {
        status.Issues = append(status.Issues, "Libvirt daemon not enabled on boot")
        status.Recommendations = append(status.Recommendations, "Enable daemon: sudo systemctl enable "+getDaemonName(status))
    }

    return status, nil
}

// Helper functions
func checkService(name string) bool {
    cmd := exec.Command("systemctl", "list-unit-files", name)
    output, err := cmd.Output()
    return err == nil && strings.Contains(string(output), name)
}

func isServiceActive(name string) bool {
    cmd := exec.Command("systemctl", "is-active", name)
    output, _ := cmd.Output()
    return strings.TrimSpace(string(output)) == "active"
}

func isServiceEnabled(name string) bool {
    cmd := exec.Command("systemctl", "is-enabled", name)
    output, _ := cmd.Output()
    enabled := strings.TrimSpace(string(output))
    return enabled == "enabled" || enabled == "static"
}

func isServiceMasked(name string) bool {
    cmd := exec.Command("systemctl", "show", name, "-p", "UnitFileState")
    output, err := cmd.Output()
    if err != nil {
        return false
    }
    return strings.Contains(string(output), "masked")
}

func getLibvirtVersion() string {
    cmd := exec.Command("virsh", "--version")
    output, err := cmd.Output()
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(output))
}

func getDaemonName(status *LibvirtStatus) string {
    if status.DaemonType == "modular" {
        return "virtqemud"
    }
    return "libvirtd"
}
```

### Phase 2: Enhanced QEMU Adapter with Libvirt Check

**Update:** `src/helpers/ptx-virt/qemu_adapter.go`

Add libvirt status to diagnostic info:

```go
func (q *QEMUAdapter) GetDiagnosticInfo() *DiagnosticInfo {
    diag := &DiagnosticInfo{
        Suggestions: []string{},
        Details:     []string{},
    }

    // ... existing QEMU/KVM checks ...

    // Check libvirt status
    if libvirtStatus, err := virtualization.DetectLibvirtStatus(); err == nil {
        if libvirtStatus.Installed {
            if libvirtStatus.Running || libvirtStatus.SocketActivated {
                diag.Details = append(diag.Details,
                    fmt.Sprintf("Libvirt %s: running (%s)", libvirtStatus.Version, libvirtStatus.DaemonType))
            } else {
                diag.Details = append(diag.Details,
                    fmt.Sprintf("Libvirt %s: NOT running", libvirtStatus.Version))
                diag.Suggestions = append(diag.Suggestions, libvirtStatus.Recommendations...)
            }
        } else {
            diag.Suggestions = append(diag.Suggestions, "Install libvirt: portunix install libvirt")
        }
    }

    return diag
}
```

### Phase 3: Fix Command Implementation

**Update:** `src/cmd/virt_exec.go`

Add `--fix-libvirt` flag:

```go
func init() {
    virtCheckCmd.Flags().Bool("fix", false, "Interactively fix detected conflicts")
    virtCheckCmd.Flags().Bool("fix-libvirt", false, "Fix libvirt daemon issues")
    virtCheckCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
    // ... existing flags ...
}

func handleVirtCheck(cmd *cobra.Command, args []string) {
    fixLibvirt, _ := cmd.Flags().GetBool("fix-libvirt")

    if fixLibvirt {
        handleLibvirtFix(cmd)
        return
    }

    // ... existing logic ...
}

func handleLibvirtFix(cmd *cobra.Command) {
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    status, err := virtualization.DetectLibvirtStatus()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error detecting libvirt: %v\n", err)
        os.Exit(1)
    }

    if !status.Installed {
        fmt.Println("❌ Libvirt is not installed")
        fmt.Println("   Install: portunix install libvirt")
        return
    }

    fmt.Printf("📊 Libvirt Status:\n")
    fmt.Printf("   Version: %s\n", status.Version)
    fmt.Printf("   Daemon type: %s\n", status.DaemonType)
    fmt.Printf("   Running: %v\n", status.Running)
    fmt.Printf("   Enabled: %v\n", status.Enabled)
    fmt.Printf("   Masked: %v\n", status.Masked)
    fmt.Println()

    if len(status.Issues) == 0 {
        fmt.Println("✅ Libvirt daemon is running correctly")
        return
    }

    // Show issues
    fmt.Println("⚠️  Libvirt Issues Detected:")
    for _, issue := range status.Issues {
        fmt.Printf("   • %s\n", issue)
    }
    fmt.Println()

    // Fix based on issue type
    daemonName := getDaemonName(status)

    if status.Masked {
        executeUnmaskLibvirt(daemonName, dryRun)
    } else if !status.Running {
        executeStartLibvirt(daemonName, dryRun)
    }

    if !status.Enabled {
        executeEnableLibvirt(daemonName, dryRun)
    }
}

func executeUnmaskLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Unmasking libvirt daemon...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute:\n")
        fmt.Printf("  sudo systemctl unmask %s\n", daemonName)
        fmt.Printf("  sudo systemctl start %s\n", daemonName)
        return
    }

    // Unmask
    cmd := exec.Command("sudo", "systemctl", "unmask", daemonName)
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to unmask %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    // Start
    cmd = exec.Command("sudo", "systemctl", "start", daemonName)
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to start %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon unmasked and started")
}

func executeStartLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Starting libvirt daemon...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute: sudo systemctl start %s\n", daemonName)
        return
    }

    cmd := exec.Command("sudo", "systemctl", "start", daemonName)
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to start %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon started")
}

func executeEnableLibvirt(daemonName string, dryRun bool) {
    fmt.Println("🔧 Enabling libvirt daemon on boot...")

    if dryRun {
        fmt.Printf("[DRY RUN] Would execute: sudo systemctl enable %s\n", daemonName)
        return
    }

    cmd := exec.Command("sudo", "systemctl", "enable", daemonName)
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to enable %s: %v\n", daemonName, err)
        os.Exit(1)
    }

    fmt.Println("✅ Libvirt daemon enabled on boot")
}
```

### Phase 4: Enhanced virt check Output

**Expected Output with Libvirt Issues:**

```bash
$ portunix virt check

Checking virtualization support...
Platform: linux
Hardware Virtualization: true
Recommended Provider: kvm

Available Providers:
  ✅ kvm (v9.2.1) at /usr/bin/qemu-system-x86_64
    └─ KVM acceleration: enabled
    └─ Loaded modules: kvm_amd, kvm, irqbypass, ccp
    └─ Libvirt 8.0.0: NOT running
    ⚠️  Libvirt daemon is masked. Run: sudo portunix virt check --fix-libvirt
```

**Expected Output After Fix:**

```bash
$ sudo portunix virt check --fix-libvirt

📊 Libvirt Status:
   Version: 8.0.0
   Daemon type: monolithic
   Running: false
   Enabled: false
   Masked: true

⚠️  Libvirt Issues Detected:
   • Libvirt daemon is masked (cannot be started)

🔧 Unmasking libvirt daemon...
✅ Libvirt daemon unmasked and started

🔧 Enabling libvirt daemon on boot...
✅ Libvirt daemon enabled on boot

$ portunix virt check

Available Providers:
  ✅ kvm (v9.2.1) at /usr/bin/qemu-system-x86_64
    └─ KVM acceleration: enabled
    └─ Loaded modules: kvm_amd, kvm, irqbypass, ccp
    └─ Libvirt 8.0.0: running (monolithic)
```

## Implementation Checklist

### Phase 1: Libvirt Detection
- [ ] Create `src/app/virtualization/libvirt_detector.go`
- [ ] Implement `DetectLibvirtStatus()` function
- [ ] Detect monolithic daemon (libvirtd)
- [ ] Detect modular daemons (virtqemud, virtnetworkd)
- [ ] Check daemon status (running/stopped/masked/enabled)
- [ ] Check socket activation
- [ ] Get libvirt version via `virsh --version`

### Phase 2: QEMU Adapter Integration
- [ ] Update `qemu_adapter.go` GetDiagnosticInfo()
- [ ] Add libvirt status to Details
- [ ] Add libvirt recommendations to Suggestions
- [ ] Test display in `virt check` output

### Phase 3: Fix Command
- [ ] Add `--fix-libvirt` flag to virtCheckCmd
- [ ] Implement `handleLibvirtFix()` function
- [ ] Implement `executeUnmaskLibvirt()` for masked daemons
- [ ] Implement `executeStartLibvirt()` for stopped daemons
- [ ] Implement `executeEnableLibvirt()` for boot persistence
- [ ] Add dry-run support

### Phase 4: Testing
- [ ] Test detection on system with masked libvirtd
- [ ] Test detection on system with stopped libvirtd
- [ ] Test detection on system with running libvirtd
- [ ] Test modular daemon detection (virtqemud)
- [ ] Test --fix-libvirt with masked daemon
- [ ] Test --fix-libvirt with stopped daemon
- [ ] Test --dry-run mode
- [ ] Verify Virtual Machine Manager connects after fix

## Test Cases

### TC001: Detect Masked Libvirtd
```bash
# Setup: Mask libvirtd
sudo systemctl mask libvirtd

# Test
portunix virt check

# Expected:
# ✅ kvm detected
# └─ Libvirt: NOT running
# ⚠️  Libvirt daemon is masked. Run: sudo portunix virt check --fix-libvirt
```

### TC002: Fix Masked Daemon
```bash
# Setup: libvirtd masked
sudo portunix virt check --fix-libvirt

# Expected:
# 🔧 Unmasking libvirt daemon...
# ✅ Libvirt daemon unmasked and started
# ✅ Libvirt daemon enabled on boot
```

### TC003: Detect Stopped Daemon
```bash
# Setup: Unmask but stop daemon
sudo systemctl unmask libvirtd
sudo systemctl stop libvirtd

# Test
portunix virt check

# Expected:
# └─ Libvirt: NOT running
# ⚠️  Start daemon: sudo portunix virt check --fix-libvirt
```

### TC004: Modular Daemon Detection
```bash
# Setup: System with virtqemud instead of libvirtd
# Test
portunix virt check

# Expected:
# └─ Libvirt 9.0.0: running (modular)
```

### TC005: Dry Run Mode
```bash
sudo portunix virt check --fix-libvirt --dry-run

# Expected:
# [DRY RUN] Would execute:
#   sudo systemctl unmask libvirtd
#   sudo systemctl start libvirtd
```

## Success Criteria

- [ ] `virt check` detects libvirt installation
- [ ] `virt check` shows libvirt daemon status
- [ ] Correctly identifies masked/stopped/running state
- [ ] Supports both monolithic and modular daemons
- [ ] `--fix-libvirt` unmasks and starts daemon
- [ ] `--fix-libvirt` enables daemon on boot
- [ ] Dry-run mode works correctly
- [ ] Virtual Machine Manager connects after fix
- [ ] Clear error messages and recommendations

## Integration with Related Issues

This issue extends:
- **#088** - VirtualBox/KVM Conflict Detection
  - Libvirt detection adds another layer to virtualization health checks

- **#089** - QEMU/KVM Adapter Implementation
  - QEMU adapter now includes libvirt status in diagnostics
  - Completes the QEMU/KVM detection story

## User Experience Flow

### Scenario: User cannot connect virt-manager

```bash
$ virt-manager
# Shows: QEMU/KVM - není připojeno

$ portunix virt check
✅ kvm (v9.2.1) at /usr/bin/qemu-system-x86_64
  └─ KVM acceleration: enabled
  └─ Libvirt 8.0.0: NOT running
  ⚠️  Libvirt daemon is masked. Run: sudo portunix virt check --fix-libvirt

$ sudo portunix virt check --fix-libvirt
📊 Libvirt Status:
   Daemon type: monolithic
   Running: false
   Masked: true

🔧 Unmasking libvirt daemon...
✅ Libvirt daemon unmasked and started
✅ Libvirt daemon enabled on boot

$ virt-manager
# Now connects successfully to QEMU/KVM
```

## Files to Create/Modify

### New Files
1. `src/app/virtualization/libvirt_detector.go` - Libvirt detection logic

### Modified Files
1. `src/helpers/ptx-virt/qemu_adapter.go` - Add libvirt status to diagnostics
2. `src/cmd/virt_exec.go` - Add --fix-libvirt flag and handlers

## Future Enhancements

1. **Socket Activation Support**: Properly handle socket-activated daemons
2. **Libvirt Installation**: Add `portunix install libvirt` command
3. **Connection URI Management**: Manage qemu:///system vs qemu:///session
4. **Permission Checks**: Verify user is in libvirt group
5. **Network Detection**: Check libvirt default network status

## Notes

- Libvirt can be running but not accessible due to permissions
- Socket activation means daemon starts on-demand
- Modular daemons are default on Fedora 35+ and Ubuntu 22.04+
- User must be in `libvirt` group for qemu:///system access

## References

- Existing: `src/app/system/virtualization.go:88` - `GetLibvirtVersion()`
- Related: `src/app/virtualization/conflict_detector.go` - VirtualBox/KVM conflicts
- Related: `src/helpers/ptx-virt/qemu_adapter.go` - QEMU adapter

---

**Reporter**: User (virt-manager connection issue)
**Assigned**: Development Team
**Labels**: bug, enhancement, virtualization, qemu, kvm, libvirt, virt-manager, daemon-management
