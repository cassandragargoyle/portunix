# Issue #88: VirtualBox/KVM Conflict Detection and Auto-Resolution

## Overview
**Type**: Enhancement
**Priority**: High
**Status**: 📋 Open
**Component**: Virtualization, VirtualBox, KVM
**Platform**: Linux
**Created**: 2025-10-04
**Target Version**: 1.6.x

## Problem Description

VirtualBox cannot start VMs when KVM kernel modules are loaded, resulting in critical error:

```
VirtualBox can't enable the AMD-V extension. Please disable the KVM kernel extension,
recompile your kernel and reboot (VERR_SVM_IN_USE).

Result Code: NS_ERROR_FAILURE (0X80004005)
Component: ConsoleWrap
Interface: IConsole {6ac83d89-6ee7-4e33-8ae6-b257b2e81be8}
```

**Root Cause**: Both VirtualBox and KVM try to use hardware virtualization extensions (AMD-V/Intel VT-x), causing a conflict when KVM modules are loaded.

## Current Behavior

When user runs `portunix virt check` or tries to start VirtualBox VM:
- ❌ VirtualBox fails with cryptic error message
- ❌ No detection of KVM conflict
- ❌ No suggested solution
- ❌ User must manually research and resolve

## Expected Behavior

`portunix virt check` should:
1. ✅ Detect VirtualBox installation
2. ✅ Detect loaded KVM modules
3. ✅ Identify conflict between VirtualBox and KVM
4. ✅ Provide clear explanation and resolution options

`portunix virt check --fix` should:
5. ✅ Offer guided fix with user confirmation

## Technical Analysis

### Conflict Detection

**KVM Modules to Check:**
- `kvm` (base module)
- `kvm_intel` (Intel CPUs)
- `kvm_amd` (AMD CPUs)

**Detection Method:**
```bash
# Check if KVM modules are loaded
lsmod | grep kvm

# Expected output if loaded:
kvm_amd               159744  0
kvm                  1355776  1 kvm_amd
```

**VirtualBox Check:**
```bash
# Check if VirtualBox kernel modules are loaded
lsmod | grep vbox

# Check VirtualBox installation
VBoxManage --version
```

### Resolution Options

#### Option 1: Unload KVM Modules (Temporary)
```bash
sudo rmmod kvm_amd    # or kvm_intel
sudo rmmod kvm
```
**Pros:** Quick, no reboot required
**Cons:** KVM-based VMs (QEMU/KVM, libvirt) won't work

#### Option 2: Blacklist KVM Modules (Permanent)
```bash
# Add to /etc/modprobe.d/blacklist-kvm.conf
blacklist kvm
blacklist kvm_amd     # or kvm_intel
blacklist kvm_intel   # or kvm_amd
```
**Pros:** Permanent solution, survives reboot
**Cons:** Disables KVM completely

#### Option 3: Switch Between VirtualBox and KVM
Create helper scripts to switch contexts:
```bash
# Use VirtualBox
sudo rmmod kvm_amd kvm
sudo modprobe vboxdrv vboxnetflt vboxnetadp

# Use KVM
sudo rmmod vboxnetadp vboxnetflt vboxdrv
sudo modprobe kvm kvm_amd
```

## Implementation Plan

### Phase 1: Detection (virt check enhancement)

**File:** `src/app/virtualization/conflict_detector.go` (NEW)

```go
type VirtualizationConflict struct {
    VirtualBox   bool
    KVM          bool
    Conflict     bool
    LoadedModules []string
    Recommendation string
}

func DetectVirtualizationConflict() (*VirtualizationConflict, error) {
    // 1. Check if VirtualBox is installed
    // 2. Check if KVM modules are loaded
    // 3. Determine if both are active (conflict)
    // 4. Provide recommendation
}
```

**Enhanced `virt check` output:**
```
🔍 Virtualization System Check
════════════════════════════════════════════════════════════

✅ VirtualBox: Installed (v7.0.12)
⚠️  KVM: Active (modules loaded)

⚠️  CONFLICT DETECTED:
   VirtualBox and KVM cannot run simultaneously.
   Both require exclusive access to AMD-V/Intel VT-x.

📊 Loaded KVM modules:
   - kvm_amd (AMD virtualization)
   - kvm (base module)

💡 RECOMMENDED ACTIONS:

To fix this conflict, run:
   sudo portunix virt check --fix

Or manually choose an option:
1. Unload KVM modules (temporary):
   sudo rmmod kvm_amd kvm

2. Blacklist KVM permanently:
   Create /etc/modprobe.d/blacklist-kvm.conf

3. Switch to KVM (unload VirtualBox):
   sudo rmmod vboxnetadp vboxnetflt vboxdrv
```

### Phase 2: Fix Flag for virt check

**Enhanced Command:** `portunix virt check --fix`

**Flags:**
- `--fix` - Interactive guided fix for detected conflicts
- `--unload-kvm` - Temporarily unload KVM modules (with --fix)
- `--blacklist-kvm` - Permanently disable KVM (with --fix)
- `--use-kvm` - Switch to KVM, unload VirtualBox (with --fix)

**Implementation:**

```go
// File: src/cmd/virt_check.go

var virtCheckCmd = &cobra.Command{
    Use:   "check",
    Short: "Check virtualization system configuration",
    Long: `Check virtualization system configuration and detect conflicts.

Use --fix flag to interactively resolve detected conflicts.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Existing check logic
        // + conflict detection

        if fixFlag {
            // Interactive conflict resolution
        }
    },
}

func init() {
    virtCheckCmd.Flags().Bool("fix", false, "Interactively fix detected conflicts")
    virtCheckCmd.Flags().Bool("unload-kvm", false, "Unload KVM modules (use with --fix)")
    virtCheckCmd.Flags().Bool("blacklist-kvm", false, "Blacklist KVM permanently (use with --fix)")
    virtCheckCmd.Flags().Bool("use-kvm", false, "Switch to KVM (use with --fix)")
}
```

### Phase 3: Automatic Error Detection

**Enhanced Error Handling:**

When VirtualBox VM fails to start with `VERR_SVM_IN_USE`:
1. Automatically run conflict detection
2. Display clear error message
3. Suggest fix command
4. Offer to run fix automatically

**Example:**
```
❌ VirtualBox VM Start Failed

Error: VirtualBox can't enable AMD-V extension (VERR_SVM_IN_USE)
Cause: KVM kernel modules are loaded and conflicting

🔧 Auto-detected conflict: VirtualBox ↔ KVM

Run this command to fix:
  sudo portunix virt check --fix

This will guide you through resolving the conflict.
```

### Phase 4: Helper Script Integration

**Create:** `helpers/ptx-virt` enhancements

Add functions:
- `detect_virt_conflict()` - Shell-based detection
- `unload_kvm()` - Safely unload KVM modules
- `blacklist_kvm()` - Create blacklist configuration
- `switch_to_kvm()` - Unload VirtualBox, load KVM

## Files to Modify/Create

### New Files
1. `src/app/virtualization/conflict_detector.go` - Conflict detection logic
2. `docs/virtualization/kvm-virtualbox-conflict.md` - User documentation

### Modified Files
1. `src/cmd/virt_check.go` - Add --fix flag and conflict resolution
2. `helpers/ptx-virt/main.go` - Add conflict resolution helpers
3. `src/app/virtualization/virtualbox.go` - Enhanced error detection

## Test Cases

### TC001: Conflict Detection
```bash
# Setup: Load KVM modules, install VirtualBox
sudo modprobe kvm kvm_amd
portunix virt check

# Expected:
# ⚠️ CONFLICT DETECTED: VirtualBox ↔ KVM
```

### TC002: Unload KVM (Temporary Fix)
```bash
sudo portunix virt check --fix --unload-kvm

# Expected:
# ✅ KVM modules unloaded
# ✅ VirtualBox ready to use
```

### TC003: Blacklist KVM (Permanent Fix)
```bash
sudo portunix virt check --fix --blacklist-kvm

# Expected:
# ✅ Created /etc/modprobe.d/blacklist-kvm.conf
# ✅ KVM will not load on next boot
```

### TC004: Automatic Detection on VM Start
```bash
# Setup: KVM loaded, try to start VirtualBox VM
portunix virt start archlinux

# Expected:
# ❌ Error detected
# 🔧 Auto-detected KVM conflict
# 💡 Suggested fix command
```

### TC005: Interactive Fix
```bash
sudo portunix virt check --fix

# Expected:
# 1. Detect conflict
# 2. Show interactive options menu
# 3. Execute selected fix based on user choice
# 4. Verify resolution
```

## Success Criteria

- [ ] `virt check` detects VirtualBox/KVM conflicts
- [ ] Clear error messages explain the issue
- [ ] `virt check --fix` flag implemented
- [ ] Temporary fix (unload KVM) works with --fix --unload-kvm
- [ ] Permanent fix (blacklist KVM) works with --fix --blacklist-kvm
- [ ] Auto-detection on VM start failure
- [ ] Interactive mode guides users through fix options
- [ ] Documentation for all resolution methods

## User Experience Flow

### Scenario 1: User tries to start VM

```bash
$ portunix virt start archlinux

🚀 Starting VirtualBox VM: archlinux...
❌ Error: VirtualBox can't enable AMD-V extension (VERR_SVM_IN_USE)

🔍 Analyzing issue...
⚠️  Detected conflict: KVM modules are loaded

💡 Quick fix available via:
   sudo portunix virt check --fix

Would you like detailed instructions? [Y/n]: y

═══════════════════════════════════════════════════════════
CONFLICT RESOLUTION OPTIONS:

1. Unload KVM (temporary - recommended for quick fix):
   sudo portunix virt check --fix --unload-kvm

2. Blacklist KVM (permanent - prevents KVM loading):
   sudo portunix virt check --fix --blacklist-kvm

3. Interactive wizard:
   sudo portunix virt check --fix
═══════════════════════════════════════════════════════════
```

### Scenario 2: User runs virt check

```bash
$ portunix virt check

🔍 Virtualization System Check
════════════════════════════════════════════════════════════

Hardware Virtualization:
✅ CPU supports AMD-V (SVM)
✅ Virtualization enabled in BIOS

Hypervisors:
✅ VirtualBox: Installed (v7.0.12)
   └─ Kernel modules: vboxdrv, vboxnetflt, vboxnetadp
⚠️  KVM: Active
   └─ Loaded modules: kvm, kvm_amd

⚠️  CONFLICT DETECTED:
   VirtualBox and KVM both require exclusive access to AMD-V.
   Only one can be active at a time.

📊 Current Status:
   - KVM is currently active (loaded)
   - VirtualBox will fail to start VMs

💡 Resolution Options:

1. Temporarily unload KVM (quick fix):
   sudo portunix virt check --fix --unload-kvm

2. Permanently disable KVM (blacklist):
   sudo portunix virt check --fix --blacklist-kvm

3. Switch to KVM (unload VirtualBox):
   sudo portunix virt check --fix --use-kvm

4. Interactive guided fix:
   sudo portunix virt check --fix
```

## Dependencies

- Existing: `virt check` command
- Existing: VirtualBox integration
- New: Kernel module detection utilities
- New: `/etc/modprobe.d/` configuration management

## Security Considerations

- ⚠️ Module loading/unloading requires `sudo`
- ✅ Confirm before making permanent changes (blacklist)
- ✅ Backup existing modprobe configurations
- ✅ Validate module names before operations
- ✅ Check if user has necessary permissions

## Documentation

### User Guide Section
Create: `docs/virtualization/resolving-kvm-virtualbox-conflicts.md`

Topics:
1. Understanding the conflict
2. Detection with `virt check`
3. Resolution options comparison
4. Step-by-step guides
5. Troubleshooting
6. FAQ

### Command Help
```bash
portunix virt check --help

Check virtualization system configuration

Usage:
  portunix virt check [flags]

Flags:
  --fix             Interactively fix detected conflicts
  --unload-kvm      Temporarily unload KVM modules (use with --fix)
  --blacklist-kvm   Permanently disable KVM via blacklist (use with --fix)
  --use-kvm         Switch to KVM, unload VirtualBox (use with --fix)
  --dry-run         Show what would be done without making changes

Examples:
  # Check virtualization status
  portunix virt check

  # Detect and fix conflicts interactively
  sudo portunix virt check --fix

  # Quickly unload KVM to use VirtualBox
  sudo portunix virt check --fix --unload-kvm

  # Permanently disable KVM
  sudo portunix virt check --fix --blacklist-kvm

  # Preview changes without applying
  sudo portunix virt check --fix --dry-run
```

## Future Enhancements

1. **Smart Switching**: Automatically switch between VirtualBox and KVM based on VM type
2. **Module Auto-reload**: Reload appropriate modules when switching contexts
3. **Configuration Profiles**: Save user preferences for automatic conflict resolution
4. **Nested Virtualization**: Detect and handle nested virtualization scenarios
5. **Integration with libvirt**: Manage KVM VMs through libvirt with automatic switching

## Related Issues

- #049 - Full QEMU/KVM Support Implementation
- #055 - VM Management Requirements for Enterprise Architect
- #057 - VirtualBox Detection False Negative on Windows
- #061 - Virtual Machine Snapshot List Shows Empty Names

## Notes

- This is a common issue on Linux systems with both VirtualBox and KVM installed
- Many users are confused by the cryptic VirtualBox error message
- Auto-detection and guided resolution will significantly improve UX
- Consider making this a general "hypervisor conflict detector" for future use

---

**Reporter**: User (via CLI error)
**Assigned**: Development Team
**Labels**: enhancement, virtualization, virtualbox, kvm, user-experience, auto-fix
