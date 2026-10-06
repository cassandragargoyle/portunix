# Issue #58: VirtualBox VM Information Access Denied

**Status**: ✅ Implemented
**Priority**: High
**Component**: Virtualization / ptx-virt
**Affected Version**: 1.6.3
**Reporter**: User observation
**Created**: 2024-09-23
**Closed**: 2026-05-10
**Implemented in**: commit 1194e66 (`fix: resolve VirtualBox VM information display issues (#058)`)

## Problem Description

The `portunix virt list` command fails to retrieve detailed VM information beyond VM names. All VMs show in "error" state with no RAM, CPU, or disk information available.

### Current Behavior

```bash
$ portunix virt list
Provider: virtualbox

NAME                 STATE        RAM      CPUS   DISK       IP
----                 -----        ---      ----   ----       --
wordpress            error        unknown  0      unknown    -
Preprocessor    error        unknown  0      unknown    -
<inaccessible>       error        unknown  0      unknown    -
```

### Expected Behavior

```bash
$ portunix virt list
Provider: virtualbox

NAME                 STATE        RAM      CPUS   DISK       IP
----                 -----        ---      ----   ----       --
wordpress            stopped      2GB      2      20GB       -
Preprocessor    running      4GB      4      50GB       192.168.1.100
```

## Root Cause Analysis

1. **VBoxManage list vms** - Works correctly, returns VM names
2. **VBoxManage showvminfo [vmname]** - Fails with E_ACCESSDENIED (0x80070005)
3. **COM Object Access** - Limited functionality for IMachine interface

### Error Details

```
VBoxManage.exe: error: The object functionality is limited
VBoxManage.exe: error: Details: code E_ACCESSDENIED (0x80070005), component MachineWrap, interface IMachine, callee IUnknown
VBoxManage.exe: error: Context: "COMGETTER(Platform)(platform.asOutParam())" at line 1116 of file VBoxManageInfo.cpp
```

## Technical Investigation

### What Works
- VirtualBox detection via registry ✓
- VBoxManage path resolution ✓
- Basic VM listing (names only) ✓
- Embedded help documentation ✓

### What Fails
- Getting VM state (showvminfo)
- Getting VM configuration (RAM, CPUs, disk)
- Getting VM network info (IP addresses)
- Any operation requiring IMachine COM interface

## Proposed Solutions

### Solution A: Enhanced Permission Handling
1. Detect E_ACCESSDENIED errors specifically
2. Attempt alternative methods to get VM info:
   - Parse VM .vbox configuration files directly
   - Use VBoxManage list runningvms for basic state
   - Read VM configs from VirtualBox XML files

### Solution B: Privilege Elevation
1. Detect when running without proper privileges
2. Prompt user to run as administrator
3. Provide automatic elevation option
4. Cache VM information when accessible

### Solution C: Hybrid Approach (Recommended)
1. Try primary method (VBoxManage showvminfo)
2. On E_ACCESSDENIED, fall back to:
   - Parse .vbox files from VirtualBox VMs directory
   - Use VBoxManage list runningvms/vms for state
   - Show partial information with warning
3. Guide user to fix permissions permanently

## Implementation Plan

### Phase 1: Direct XML Parsing
1. Locate VirtualBox VMs directory
   - Windows: `%USERPROFILE%\VirtualBox VMs\`
   - Linux: `~/VirtualBox VMs/`
2. Parse .vbox XML files for each VM
3. Extract configuration:
   - Memory size
   - CPU count
   - Disk attachments
   - Network adapters

### Phase 2: State Detection
1. Use `VBoxManage list runningvms` for running state
2. Cross-reference with `list vms` for all VMs
3. Deduce states:
   - In runningvms = "running"
   - Not in runningvms = "stopped"
   - Parse error = "error"

### Phase 3: User Guidance
1. Detect permission issues
2. Show actionable error messages
3. Provide fix instructions:
   - Run as administrator
   - Fix VirtualBox installation
   - Reset COM permissions

## Code Changes Required

### 1. src/app/virt/virtualbox/virtualbox.go
- Add `parseVBoxFile(vmPath string) (*types.VMInfo, error)`
- Add `getVMsDirectory() string`
- Modify `GetInfo()` to use fallback methods
- Add `getRunningVMs() ([]string, error)`

### 2. src/helpers/ptx-virt/virtualbox_adapter.go
- Update error handling in List()
- Add XML parsing capability
- Implement fallback strategy

### 3. src/helpers/ptx-virt/main.go
- Enhance error messages
- Add permission check warnings
- Show partial data with indicators

## Testing Requirements

### Test Cases
1. Normal user without admin rights
2. Admin user with full access
3. Corrupted VirtualBox installation
4. Mixed accessible/inaccessible VMs
5. No VirtualBox VMs directory
6. Invalid .vbox XML files

### Expected Results
- Graceful degradation of features
- Clear error messages
- Partial information display
- Actionable fix instructions

## Acceptance Criteria

1. ✅ VM names always displayed (if accessible)
2. ✅ State detection works without showvminfo
3. ✅ Configuration shown from .vbox files
4. ✅ Clear indication of limited access
5. ✅ Help instructions for fixing permissions
6. ✅ No crashes on E_ACCESSDENIED errors

## Priority Justification

**High Priority** because:
- Core functionality (VM listing) is severely limited
- Affects all Windows users without admin rights
- Makes ptx-virt helper less useful
- Poor user experience with "error" states

## Related Issues

- #057: VirtualBox Detection Windows False Negative (Implemented)
- virtualbox-help.md: E_ACCESSDENIED troubleshooting guide

## Notes

- This is a common VirtualBox issue on Windows
- Often caused by VirtualBox installation under different user
- May require VirtualBox service restart
- Could be related to Windows UAC settings

## Resolution Tracking

- [x] Issue created
- [x] Solution approved
- [x] Implementation started
- [x] Testing completed
- [x] Documentation updated
- [x] Issue closed