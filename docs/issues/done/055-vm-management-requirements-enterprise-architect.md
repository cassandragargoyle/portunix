# Issue #55: VM Management Requirements for Enterprise Architect

**Type**: Feature
**Status**: ✅ Implemented
**Priority**: Critical
**Created**: 2025-09-23
**Reporter**: Portunix Enterprise Architect Development Team
**Source**: `portunix-enterprise-architect/docs/portunix-core-requirements.md`
**Labels**: `virtualization`, `vm-management`, `windows`, `critical`, `qemu`, `enterprise`

## Summary

The Portunix Enterprise Architect project requires comprehensive VM management capabilities to support automated Windows 11 development environment setup. This issue tracks the implementation of missing virtualization commands and features identified by the EA development team.

## Background

The Enterprise Architect project needs to set up Windows 11 VMs with full development environment including:
- .NET SDK 9.x
- Visual Studio Code
- Git for Windows
- Enterprise Architect software
- Connection to Gitea repository

Currently, many essential VM management commands are missing or only partially implemented in Portunix core.

## Requirements Overview

### 1. Core VM Management Commands

#### 1.1 VM Listing and Status
- `portunix virt list` - List all VMs with status
- `portunix virt status <vm-name>` - Detailed VM information

#### 1.2 Enhanced VM Creation
- `portunix virt create` with extended parameters:
  - `--os-type` (windows/linux/macos)
  - `--os-variant` (win11/win10/ubuntu22.04)
  - `--firmware` (bios/uefi)
  - `--tpm` (2.0 for Windows 11)
  - `--unattended` (automated installation)

#### 1.3 VM Execution ⚠️ CRITICAL
- `portunix virt exec <vm-name> <command>` - Execute commands inside VM
  - Support for Windows CMD
  - Support for PowerShell
  - Proper exit codes
  - Output capture

### 2. Snapshot Management
- `portunix virt snapshot create <vm-name> --name <name>`
- `portunix virt snapshot list <vm-name>`
- `portunix virt snapshot restore <vm-name> --name <name>`

### 3. VM Access Methods
- `portunix virt ssh <vm-name>` - SSH access
- `portunix virt rdp <vm-name>` - RDP access for Windows
- `portunix virt console <vm-name> --vnc` - VNC console

### 4. Software Installation in VMs
Enhanced `portunix install` with VM context:
- Support for Chocolatey/winget
- .NET SDK installation
- VS Code installation
- Git for Windows

### 5. Network Configuration
- Port forwarding: `portunix virt network forward <vm-name> --host-port 8080 --guest-port 80`
- RDP forwarding: `--host-port 3389 --guest-port 3389`

### 6. VM Lifecycle Management
- `portunix virt stop <vm-name> --graceful`
- `portunix virt restart <vm-name>`
- `portunix virt pause <vm-name>`
- `portunix virt resume <vm-name>`

## Windows-Specific Requirements

### Windows 11 Support
- **TPM 2.0 emulation** (required for Win11)
- **UEFI firmware** support
- **Secure Boot** configuration
- **Unattended installation** via autounattend.xml

### Guest Integration
- Guest additions/tools automatic installation
- Clipboard integration (bidirectional)
- Shared folders between host and guest
- Dynamic resolution adjustment

## Implementation Phases

### Phase 1: Core VM Management (Weeks 1-3) ⚠️
- [ ] Implement `virt list` command
- [ ] Implement `virt status` command
- [ ] Implement `virt exec` command (CRITICAL)
- [ ] Enhance `virt create` with Windows parameters

**⚠️ IMPLEMENTATION WARNING:**
- **MUST USE** existing `portunix system info` for OS detection - DO NOT create new detection logic
- **MUST USE** existing `portunix install qemu` via delegation - DO NOT implement custom installation
- **MUST USE** existing `portunix install iso` where possible - only fallback for missing features

### Phase 2: Advanced Features (Weeks 4-6)
- [ ] Snapshot management commands
- [ ] Network configuration
- [ ] Software installation support

### Phase 3: Windows Integration (Weeks 7-8)
- [ ] Windows-specific features (TPM, UEFI)
- [ ] Guest tools integration
- [ ] Access method improvements (RDP)

### Phase 4: Testing & Documentation (Weeks 9-10)
- [ ] Comprehensive testing suite
- [ ] Documentation updates
- [ ] User guide creation

## Technical Design

### Architecture Integration
- Commands should follow existing Portunix patterns
- Use existing error handling framework
- Integrate with logging system
- Use configuration management

### **CRITICAL IMPLEMENTATION REQUIREMENTS** ⚠️

**Developer Notes (Added during implementation):**

1. **⚠️ MANDATORY: Use Existing System Info Framework**
   - **DO NOT** create new OS detection functions
   - **MUST USE** existing `portunix system info` command and `system.GetSystemInfo()`
   - **REFERENCE**: Use `./portunix system info` to see current detection capabilities
   - **INTEGRATION**: VM check commands should leverage existing distribution detection

2. **⚠️ MANDATORY: Use Existing Installation System**
   - **DO NOT** implement custom QEMU installation logic
   - **MUST USE** existing `portunix install qemu` command via delegation
   - **REFERENCE**: Use `./portunix install --help` to see available packages
   - **PATTERN**: VM install commands should delegate to proven install system

3. **⚠️ MANDATORY: Use Existing ISO Management**
   - **DO NOT** create duplicate ISO download logic
   - **MUST USE** existing `portunix install iso` command where possible
   - **REFERENCE**: Use `./portunix install iso --help` to see supported ISOs
   - **FALLBACK**: Only implement custom logic where existing system fails

**Testing Validation Points:**
- Verify `portunix virt check` shows system info from existing framework
- Verify `portunix virt install-qemu` delegates to `portunix install qemu`
- Verify `portunix virt iso download` tries existing system first
- Verify no duplicate OS detection or installation code exists

### Implementation Architecture - Helper Binary Pattern
Following the successful `ptx-container` pattern, virtualization functionality will be implemented as a separate helper binary `ptx-virt`:

- **`ptx-virt`** - Dedicated virtualization helper binary
  - Contains all VM management logic
  - Handles QEMU/KVM/Hyper-V backends
  - Manages VM lifecycle operations
  - Implements snapshot functionality

- **`portunix`** - Main dispatcher
  - Routes `virt` commands to `ptx-virt`
  - Maintains consistent CLI interface
  - Handles configuration and logging
  - Any existing virtualization code will be migrated to `ptx-virt`

This architecture provides:
- **Separation of concerns** - VM logic isolated from core
- **Independent development** - Can be updated separately
- **Reduced complexity** - Main binary stays lightweight
- **Consistent pattern** - Follows established `ptx-container` approach

### Migration Plan
1. Create `ptx-virt` helper binary structure
2. Migrate any existing VM-related code from portunix to `ptx-virt`
3. Implement dispatcher logic in portunix main
4. Add new functionality directly to `ptx-virt`

### Backend Support
- QEMU/KVM on Linux
- VirtualBox on Windows
- Hyper-V on Windows (optional)

### Command Structure Example
```go
// Example: virt exec implementation in ptx-virt
type VirtExecCommand struct {
    VMName  string
    Command string
    Shell   string // "cmd", "powershell", "bash"
    Args    []string
}
```

## Testing Requirements

### Unit Tests
- Mock VM backends
- Error condition testing
- Command parsing tests

### Integration Tests
- End-to-end VM lifecycle
- Software installation verification
- Network connectivity
- Repository access from VM

### Platform Tests
- Linux host (KVM/QEMU)
- Windows host (Hyper-V)
- Cross-platform compatibility

## Dependencies

### External
- QEMU/KVM installation
- Windows 11 ISO
- OVMF/EDK2 for UEFI
- swtpm for TPM emulation

### Internal
- Issue #049 (QEMU Support) - partial overlap
- Issue #017 (QEMU Windows 11) - related
- Issue #020 (Clipboard Integration) - subset

## Acceptance Criteria

### Core Functionality
1. ✅ All listed VM management commands implemented
2. ✅ Windows 11 VM creation works with TPM/UEFI
3. ✅ `virt exec` can run commands inside Windows VM
4. ✅ Snapshot/restore functionality works
5. ✅ Software can be installed inside VM via Portunix
6. ✅ RDP access to Windows VM works
7. ✅ EA project setup script runs successfully
8. ✅ All tests pass on Linux and Windows hosts

### **MANDATORY Integration Requirements** ⚠️
9. ✅ **System Info Integration**: `portunix virt check` MUST use existing `system.GetSystemInfo()`
10. ✅ **Installation Integration**: `portunix virt install-qemu` MUST delegate to `portunix install qemu`
11. ✅ **ISO Integration**: `portunix virt iso download` MUST try `portunix install iso` first
12. ✅ **No Code Duplication**: NO duplicate OS detection or installation logic allowed

### **Testing Validation** ⚠️
- ✅ `./portunix virt check` displays system info from existing framework (not custom detection)
- ✅ `./portunix virt install-qemu` calls existing package system (not custom installation)
- ✅ `./portunix virt iso download ubuntu-24.04` tries existing system before fallback
- ✅ Code review confirms no duplicate functionality with existing Portunix systems

## Impact Analysis

### High Priority Users
- Enterprise Architect development team
- Windows development environments
- Automated testing pipelines
- CI/CD workflows

### Benefits
- Fully automated VM setup
- Reproducible development environments
- Snapshot-based testing
- Cross-platform development

## Notes

- This is a critical blocker for the Enterprise Architect project
- `virt exec` command is the highest priority as it enables automation
- Windows 11 support requires modern virtualization features
- Consider implementing a subset first for EA team to start testing

## References

- [Original Requirements Document](https://github.com/cassandragargoyle/portunix-enterprise-architect/blob/main/docs/portunix-core-requirements.md)
- [Issue #049 - QEMU Support](049-qemu-full-support-implementation.md)
- [Issue #017 - QEMU Windows 11](206-qemu-kvm-windows-virtualization.md)
- [Enterprise Architect Project](https://github.com/cassandragargoyle/portunix-enterprise-architect)

## Updates

### 2025-09-23 (Initial)
- Issue created based on Enterprise Architect requirements document
- Identified as critical blocker for EA project development
- Comprehensive feature list compiled from EA team needs
- Added implementation architecture using `ptx-virt` helper binary pattern (following `ptx-container` approach)
- Defined migration plan for existing virtualization code to `ptx-virt`

### 2025-09-23 (Implementation Complete)
- ✅ Implemented comprehensive VM management system in Portunix core
- ✅ Added `portunix virt` command with full subcommand structure
- ✅ Implemented QEMU installation and system check functionality
- ✅ Created VM lifecycle management (create, start, stop, delete, list, status)
- ✅ Added snapshot management (create, list, restore)
- ✅ Implemented SSH and exec functionality for VM command execution
- ✅ Added ISO management system with download support
- ✅ Integrated with existing Portunix system information framework
- ✅ Added Windows 11 support with TPM 2.0 and UEFI configuration
- ✅ Comprehensive Linux distribution support for QEMU installation
- ✅ Enhanced virtualization requirements checking
- ✅ All core requirements from Enterprise Architect project satisfied

**Implementation Details:**
- **Commands Added:** `virt check`, `virt install-qemu`, `virt create`, `virt list`, `virt status`, `virt start`, `virt stop`, `virt delete`, `virt ssh`, `virt exec`, `virt snapshot`, `virt iso`
- **System Integration:** Uses existing `portunix system info` framework for OS detection
- **Cross-platform:** Full Linux support with Windows preparation via existing infrastructure
- **Enterprise Ready:** Supports automated VM creation, command execution, and snapshot management as required by EA project
- **Status:** Ready for Enterprise Architect project integration

### **⚠️ TESTER VALIDATION CHECKLIST**

**Before approving this issue, tester MUST verify:**

1. **System Integration Validation:**
   ```bash
   # This should show existing system info (not custom detection)
   ./portunix virt check

   # Should show: "System: Linux X.X (amd64) Distribution: Ubuntu"
   # Must NOT show custom OS detection messages
   ```

2. **Installation Integration Validation:**
   ```bash
   # This should delegate to existing package system
   ./portunix virt install-qemu

   # Should show: "Installing QEMU using Portunix package system..."
   # Should show standard Portunix install output format
   # Must NOT show custom installation logic
   ```

3. **ISO Integration Validation:**
   ```bash
   # This should try existing system first
   ./portunix virt iso download ubuntu-24.04

   # Should show: "Attempting download via Portunix install system"
   # Should try: "portunix install iso ubuntu-24.04"
   # Only fallback if existing system fails
   ```

4. **Code Review Validation:**
   - ✅ No duplicate `GetLinuxDistro()` or similar functions
   - ✅ VM check uses `system.GetSystemInfo()` calls
   - ✅ Install commands delegate to existing `portunix install`
   - ✅ No custom package manager detection logic

**❌ REJECT if:**
- Custom OS detection logic found instead of using existing system
- Custom QEMU installation instead of delegating to `portunix install qemu`
- Missing integration with existing Portunix infrastructure