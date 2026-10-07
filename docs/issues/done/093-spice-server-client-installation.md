# Issue #93: Spice Server and Client Installation Support

**Status:** ✅ Implemented
**Priority:** High
**Type:** Enhancement
**Created:** 2025-10-04
**Closed:** 2026-05-16
**Labels:** enhancement, package-management, virtualization, spice, qemu, kvm, clipboard

## Problem Statement

When working with QEMU/KVM virtual machines, clipboard sharing between host and guest requires Spice protocol support. Currently, users must manually install:

1. **Spice Server** (on host) - for serving the VM display and clipboard
2. **Spice Client** (on host) - for viewing and interacting with VMs (virt-viewer)
3. **Spice Guest Agent** (inside VM) - for clipboard integration and guest features

This manual installation is error-prone and requires knowledge of correct package names across different Linux distributions and Windows.

## Current Situation

Users wanting clipboard integration with QEMU/KVM VMs must:
- Research correct package names for their distribution
- Manually install spice-server, virt-viewer, and related tools
- Install spice-vdagent or spice-guest-tools inside each VM
- Configure libvirt XML to include Spice channels

## Proposed Solution

Implement Spice installation support in Portunix package registry with the following packages:

### 1. Spice Server (Host)
**Package Name:** `spice-server`

**Purpose:** Backend for serving VM display and clipboard

**Installation:**
- **Linux (Debian/Ubuntu):** `apt install qemu-system spice-server`
- **Linux (Fedora/RHEL):** `dnf install qemu-kvm spice-server`
- **Linux (Arch):** `pacman -S qemu spice`
- **Windows:** Included with QEMU Windows build

### 2. Spice Client/Viewer (Host)
**Package Name:** `spice-client` or `virt-viewer`

**Purpose:** Client application for connecting to Spice-enabled VMs

**Installation:**
- **Linux (Debian/Ubuntu):** `apt install virt-viewer`
- **Linux (Fedora/RHEL):** `dnf install virt-viewer`
- **Linux (Arch):** `pacman -S virt-viewer`
- **Windows:** Download virt-viewer MSI installer from https://virt-manager.org/download/

### 3. Spice Guest Agent (Inside VM)
**Package Name:** `spice-guest-agent`

**Purpose:** Guest-side agent for clipboard sharing and VM integration

**Installation:**
- **Linux (Debian/Ubuntu):** `apt install spice-vdagent`
- **Linux (Fedora/RHEL):** `dnf install spice-vdagent`
- **Linux (Arch):** `pacman -S spice-vdagent`
- **Windows:** Download spice-guest-tools.exe from https://www.spice-space.org/download.html

## Implementation Details

### Package Registry Entries

Create three new package entries in `assets/packages/`:

1. **spice-server.json** - Spice server component
2. **spice-client.json** or **virt-viewer.json** - Client viewer application
3. **spice-guest-agent.json** - Guest-side agent for VMs

### Commands

```bash
# Install Spice server (on host)
portunix install spice-server

# Install Spice viewer client (on host)
portunix install virt-viewer
# or
portunix install spice-client

# Install guest agent (inside VM)
portunix install spice-guest-agent
```

### Integration with Existing Features

This enhancement complements:
- **Issue #092:** Libvirt package installation
- **Issue #090:** Libvirt daemon detection and auto-fix
- **Issue #089:** QEMU/KVM adapter implementation
- **Issue #055:** VM Management Requirements

### Expected Workflow

```bash
# On host machine
portunix install libvirt        # Issue #092
portunix install spice-server   # This issue
portunix install virt-viewer    # This issue

# Inside VM (after creation)
portunix install spice-guest-agent  # This issue
```

## Technical Requirements

### Package Detection
- Detect if QEMU/KVM is already installed
- Check for existing Spice components
- Verify libvirt compatibility

### Platform Support
- ✅ Linux (Debian/Ubuntu)
- ✅ Linux (Fedora/RHEL/CentOS)
- ✅ Linux (Arch)
- ✅ Windows (guest agent only, server via QEMU)

### Post-Installation
- Enable spice-vdagent service on Linux guests
- Provide configuration examples for libvirt XML
- Display verification commands

## Acceptance Criteria

1. ✅ Spice server can be installed on Linux hosts
2. ✅ Virt-viewer (Spice client) can be installed on Linux and Windows hosts
3. ✅ Spice guest agent can be installed inside VMs (Linux and Windows)
4. ✅ Installation detects platform and uses correct package manager
5. ✅ Post-installation verification confirms Spice components are working
6. ✅ Help text explains clipboard sharing setup
7. ✅ Integration with libvirt VM management (Issue #092)

## Testing Strategy

### Unit Tests
- Package registry entry validation
- Platform detection for Spice packages
- Installation method selection

### Integration Tests
1. **Host Installation Test:**
   - Install spice-server on clean Linux system
   - Verify QEMU Spice support
   - Check virt-viewer installation

2. **Guest Installation Test:**
   - Create test VM with Portunix
   - Install spice-guest-agent inside VM
   - Verify service is enabled and running

3. **Clipboard Integration Test:**
   - Create VM with Spice channel
   - Install guest agent
   - Test clipboard copy/paste between host and guest

### Container Testing
```bash
# Test server installation in container
portunix container run-in-container spice-server --image ubuntu:22.04

# Test guest agent installation
portunix container run-in-container spice-guest-agent --image ubuntu:22.04
```

## Documentation

### User Documentation
- Add Spice section to virtualization guide
- Document clipboard sharing setup
- Provide libvirt XML examples
- Troubleshooting guide for clipboard issues

### Example Configuration

```xml
<!-- Add to VM XML for clipboard sharing -->
<channel type='spicevmc'>
  <target type='virtio' name='com.redhat.spice.0'/>
</channel>
```

## Benefits

1. **Simplified Setup:** One-command installation of Spice components
2. **Cross-Platform:** Works on all major Linux distributions
3. **Integration:** Seamless integration with libvirt and QEMU
4. **Clipboard Sharing:** Easy setup for host-guest clipboard integration
5. **VM Management:** Enhanced VM experience with Spice features

## Dependencies

- **Prerequisite:** QEMU/KVM installed
- **Prerequisite:** Libvirt installed (Issue #092)
- **Related:** VM management features (Issue #055)

## References

- [Spice Project](https://www.spice-space.org/)
- [Virt-Manager Download](https://virt-manager.org/download/)
- [Spice Protocol Documentation](https://www.spice-space.org/docs.html)
- [Libvirt Spice Configuration](https://libvirt.org/formatdomain.html#graphical-framebuffers)

## Implementation Checklist

- [ ] Create `spice-server.json` package definition
- [ ] Create `virt-viewer.json` package definition
- [ ] Create `spice-guest-agent.json` package definition
- [ ] Implement platform-specific installation methods
- [ ] Add post-installation verification
- [ ] Add help text and usage examples
- [ ] Write unit tests
- [ ] Write integration tests (container-based)
- [ ] Update documentation
- [ ] Test on multiple distributions
- [ ] Test Windows guest agent installation
- [ ] Integration testing with libvirt VMs

## Notes

- Spice server is typically bundled with QEMU packages
- Virt-viewer is the recommended Spice client for Linux
- Windows guest tools include VirtIO drivers + Spice agent
- Clipboard sharing requires both server channel and guest agent
- Alternative to Spice: QXL graphics + QEMU guest agent (limited clipboard support)

---

**Related Issues:**
- #092: Libvirt Package Installation Support
- #090: Libvirt Daemon Detection and Auto-Fix
- #089: QEMU/KVM Adapter Implementation
- #055: VM Management Requirements for Enterprise Architect
- #049: Full QEMU/KVM Support Implementation in Portunix
