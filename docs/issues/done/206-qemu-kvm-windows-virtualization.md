# Issue #206: QEMU/KVM Windows 11 Virtualization with Snapshot Support

> **Renumbered:** formerly internal issue #017. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Type:** Feature
**Priority:** High
**Status:** ❌ Closed (Superseded)
**Labels:** virtualization, qemu, kvm, windows, snapshot
**Closed:** 2026-05-09
**Superseded by:** the `portunix virt` command tree and the `ptx-virt` helper —
implemented across #049 (QEMU full support), #089 (QEMU/KVM adapter),
#088 (VirtualBox/KVM conflict detection), #090–#092 (libvirt detection,
dependencies, package installation), #061 (snapshot list fix), and
#121 (libvirt detection fix).

## Closure Note (2026-05-09)

The functional scope of this issue — installing QEMU/KVM, creating Windows VMs,
and managing snapshots — is fully covered today by `portunix virt` (backed by
`src/app/virt/` and the `ptx-virt` helper) and the install-packages system. The
original four-phase plan from this issue maps to features that already shipped:

| #017 phase | Where it ships today |
|------------|----------------------|
| Phase 1 — QEMU/KVM install + VM creation | `portunix install qemu-kvm` (#092), `portunix virt create` (#049, #089) |
| Phase 2 — Snapshot management | `portunix virt snapshot create\|list\|restore\|delete` (#061) |
| Phase 3 — Templates / automation | `portunix virt template …`, cloud-init support (documented in `docs/commands/virtualization/virt.md`) |
| Phase 4 — virt-manager / GUI integration | libvirt backend integration (#090, #091, #121) — virt-manager works against the same libvirt daemon |
| Hardware-virt detection (VT-x/AMD-V) | Conflict detector + libvirt detector (`src/app/virt/conflict_detector.go`, `libvirt_detector.go`, #088) |

Closing as **Superseded** rather than Implemented because the originally
proposed `portunix vm …` command namespace was never introduced; the same
user-facing problem is solved through `portunix virt …`. Specific Windows-VM
subtopics that this issue listed but that are not yet fully covered are tracked
as narrower follow-up issues — notably **#020** (QEMU Windows clipboard / SPICE),
**#093** (SPICE server/client install), **#067** (disk image helper), and
**#068** (main-binary ⇄ ptx-virt integration). Any further Windows-11-specific
gaps (TPM 2.0 / Secure Boot defaults, Windows ISO download automation) should
be filed as new, focused issues rather than reopened here.

The original scope below is preserved as historical context.

---

## Overview
Implement comprehensive QEMU/KVM virtualization support in Portunix for running Windows 11 VMs with snapshot capabilities. This feature enables users to create isolated Windows environments for testing enterprise software trials (e.g., Enterprise Architect) and easily revert to clean states.

## Problem Statement
Users need a reliable way to:
- Run Windows 11 in a virtualized environment on Linux hosts
- Create and manage VM snapshots for testing trial software
- Easily revert VMs to clean states after trial periods expire
- Automate the VM creation and management process through Portunix

## Proposed Solution

### Core Features
1. **QEMU/KVM Installation Management**
   - Detect and install QEMU/KVM components
   - Install required packages: `qemu-kvm`, `libvirt-daemon-system`, `libvirt-clients`, `virt-manager`, `bridge-utils`
   - Configure user permissions (add to `libvirt` and `kvm` groups)
   - Verify hardware virtualization support (VT-x/AMD-V)

2. **Windows VM Creation Wizard**
   - Download Windows 11 ISO from official sources
   - Create qcow2 disk images with configurable sizes
   - Automated VM creation with optimal settings for Windows 11
   - VirtIO driver integration for better performance
   - TPM 2.0 and Secure Boot configuration for Windows 11 requirements

3. **Snapshot Management**
   - Create named snapshots with descriptions
   - List all available snapshots
   - Revert to specific snapshots
   - Delete unwanted snapshots
   - Automatic snapshot before major changes

4. **VM Lifecycle Management**
   - Start/stop/restart VMs
   - Monitor VM resource usage
   - Configure VM resources (CPU, RAM, disk)
   - Network configuration (NAT, bridge)
   - Shared folder setup between host and guest

### Command Interface

#### Method 1: Dedicated VM Commands (Advanced)
```bash
# Installation and setup
portunix vm install-qemu              # Install QEMU/KVM stack
portunix vm check                      # Check hardware support

# VM creation with advanced options
portunix vm create windows11 \
  --iso ~/iso/win11.iso \
  --disk-size 60G \
  --ram 8G \
  --cpus 4 \
  --os windows11

# Snapshot management
portunix vm snapshot create win11-vm clean-install "Fresh Windows 11 installation"
portunix vm snapshot list win11-vm
portunix vm snapshot revert win11-vm clean-install
portunix vm snapshot delete win11-vm old-snapshot

# VM operations
portunix vm start win11-vm
portunix vm stop win11-vm
portunix vm info win11-vm
portunix vm console win11-vm          # Connect to VM console
```

#### Method 2: Unified Create Command (Simple)
```bash
# Simple VM creation via existing create interface
portunix create vm \
  --vmtype qemu \
  --vmname windows11 \
  --iso ~/iso/win11.iso \
  --basefolder ~/VMs

# Also supports VirtualBox for comparison
portunix create vm \
  --vmtype vbox \
  --vmname ubuntu-dev \
  --iso ~/iso/ubuntu.iso \
  --basefolder ~/VMs
```

### Platform Support
- **Primary**: Linux (Ubuntu, Debian, Fedora, Arch)
- **Secondary**: Windows via WSL2 (limited functionality)
- **Future**: macOS with HVF support

### Technical Implementation

#### 1. Package Management Integration
Add to `assets/install-packages.json`:
```json
{
  "qemu-kvm": {
    "linux": {
      "apt": ["qemu-kvm", "libvirt-daemon-system", "libvirt-clients", "virt-manager", "bridge-utils"],
      "dnf": ["qemu-kvm", "libvirt", "virt-manager", "virt-install"],
      "pacman": ["qemu", "libvirt", "virt-manager", "edk2-ovmf"]
    }
  }
}
```

#### 2. VM Configuration Templates
Pre-configured templates for:
- Windows 11 (with TPM 2.0 and Secure Boot)
- Windows 10
- Windows Server 2022
- Linux distributions

#### 3. Snapshot Strategy
- Use qcow2 format for disk images (native snapshot support)
- Implement both internal and external snapshot management
- Automatic cleanup of old snapshots (configurable retention)

## Benefits
1. **Cost Savings**: Avoid purchasing software licenses for testing
2. **Clean Testing Environment**: Always start from known-good state
3. **Enterprise Software Trials**: Test 30-day trials repeatedly
4. **Development Isolation**: Separate environments for different projects
5. **Cross-platform Testing**: Test Windows software on Linux hosts

## Implementation Priority
- **Phase 1**: Basic QEMU/KVM installation and VM creation
- **Phase 2**: Snapshot management
- **Phase 3**: Advanced features (templates, automation)
- **Phase 4**: GUI integration with virt-manager

## Dependencies
- Linux kernel with KVM support
- CPU with virtualization extensions (Intel VT-x or AMD-V)
- Sufficient disk space for VM images
- Adequate RAM for running VMs

## Security Considerations
- Isolated VM environments
- Network isolation options
- Secure boot for Windows 11
- Encrypted disk image support

## Testing Requirements
- Test on multiple Linux distributions
- Verify Windows 11 TPM/Secure Boot requirements
- Performance benchmarking
- Snapshot integrity testing

## Documentation Requirements
- Step-by-step setup guide
- Troubleshooting common issues
- Performance optimization tips
- Network configuration guide

## Related Issues
- Could integrate with plugin system for VM management UI
- May benefit from wizard framework (#014)
- Could use datastore system (#009) for VM metadata

## Labels
`enhancement`, `virtualization`, `qemu`, `kvm`, `windows`, `snapshot`, `cross-platform`, `enterprise`