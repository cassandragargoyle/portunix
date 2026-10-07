# Issue #49: Universal Virtualization Support with QEMU/KVM and VirtualBox

## Overview
Implement universal virtualization support in Portunix through a unified `virt` command that automatically selects between QEMU/KVM (Linux) and VirtualBox (Windows), similar to the universal container command. This complements Docker/Podman containerization by providing full OS virtualization for scenarios where containers are insufficient. Includes installation management, VM lifecycle, snapshots, SSH integration, and nested Portunix deployment within VMs.

## Problem Statement
Currently, Portunix has partial virtualization support with limitations:
- Separate commands for QEMU and VirtualBox (no unified interface)
- **WRONG**: Installation via `portunix vm install-qemu` should be `portunix install qemu`
- Basic VM creation exists but lacks full lifecycle management
- No configuration-based virtualization backend selection
- No integrated SSH access to VMs
- No automated Portunix deployment inside VMs
- No ISO download/management capabilities
- Missing VM enhancement features (clipboard, shared folders)
- Limited snapshot management implementation

## Proposed Solution

### Phase 0: Universal Virtualization Interface

#### Universal `virt` Command
Implement a universal virtualization interface similar to the container system:

```bash
# Universal commands that auto-select backend
portunix virt list                    # Lists VMs using configured backend
portunix virt create ubuntu-test      # Creates VM using configured backend
portunix virt start ubuntu-test       # Starts VM using configured backend
portunix virt ssh ubuntu-test         # SSH into VM

# Configuration determines backend:
# - Linux: defaults to QEMU/KVM
# - Windows: defaults to VirtualBox
# - Configurable via ~/.portunix/config.yaml
```

#### Configuration System
```yaml
# ~/.portunix/config.yaml
virtualization_backend: auto  # Options: auto, qemu, virtualbox
# auto = QEMU on Linux, VirtualBox on Windows

container_runtime: podman     # Existing container configuration
virt_defaults:
  ram: 4G
  cpus: 2
  disk: 40G
```

#### Backend Selection Logic
1. Check configuration file for `virtualization_backend`
2. If not configured or "auto":
   - Linux: Use QEMU/KVM if available, fallback to VirtualBox
   - Windows: Use VirtualBox if available, fallback to WSL2+QEMU
3. If specific backend configured, use it or error if unavailable
4. Translate universal commands to backend-specific implementations

### Phase 1: Core Virtualization Installation & Management

#### 1.1 Enhanced Virtualization Installation
```bash
# Install virtualization stack via portunix install command
portunix install virt              # Auto-selects: QEMU on Linux, VirtualBox on Windows
portunix install qemu              # Explicitly install QEMU/KVM (Linux only)
portunix install virtualbox        # Explicitly install VirtualBox (cross-platform)

# Variants for virt command
portunix install virt --variant full     # Include GUI tools
portunix install virt --variant minimal  # CLI tools only
```

#### 1.2 Post-Installation Verification
```bash
portunix virt check                # Verify virtualization installation
portunix vm check                  # Legacy: Verify QEMU/KVM installation
portunix system info               # Show virtualization backend in capabilities
```

### Phase 2: ISO Management

#### 2.1 ISO Download & Caching
```bash
# Download official ISOs (part of virt commands)
portunix virt iso download ubuntu-24.04
portunix virt iso download ubuntu-22.04-server
portunix virt iso download debian-12
portunix virt iso download windows11-latest
portunix virt iso download windows10-22h2

# Preview download without actually downloading
portunix virt iso download debian-12 --dry-run      # Show what would be downloaded
portunix virt iso download ubuntu-24.04 --dry-run   # Preview download details

# ISO management
portunix virt iso list              # List downloaded ISOs
portunix virt iso info ubuntu-24.04  # ISO details (size, checksum)
portunix virt iso rm ubuntu-22.04   # Remove ISO
portunix virt iso clean             # Remove old/unused ISOs

# Verify downloaded ISOs
portunix virt iso verify ubuntu-24.04  # Checksum verification
```

#### 2.2 ISO Storage Structure
```
# Konzistentní s ostatními cache adresáři v projektu
.cache/
├── isos/                           # ISO soubory (konzistentní s existujícím systémem)
│   ├── ubuntu-24.04-desktop-amd64.iso
│   ├── ubuntu-22.04.3-server-amd64.iso
│   ├── debian-12.4.0-amd64-netinst.iso
│   ├── Win11_24H2_English_x64v2.iso
│   └── checksums.json
├── PodmanDesktopInstaller.exe      # Existující podman cache
├── Microsoft.DesktopAppInstaller.msixbundle  # Winget cache
└── další instalátory...
```

**Poznámka**: Projekt už používá `.cache/isos/` v `install-isos.json`, takže je to konzistentní!

### Phase 3: VM Lifecycle Management

#### 3.1 VM Creation with Templates
```bash
# Universal VM creation (auto-selects backend)
portunix virt create ubuntu-test \
  --template ubuntu-24.04 \
  --ram 4G \
  --disk 40G \
  --cpus 4

# Create Windows 11 VM with TPM/UEFI
portunix virt create win11-test \
  --template windows11 \
  --ram 8G \
  --disk 100G \
  --tpm \              # QEMU: swtpm, VirtualBox: TPM passthrough
  --secure-boot

# Use downloaded ISOs in VM creation
portunix virt create ubuntu-test --iso ubuntu-24.04  # Uses downloaded ISO
portunix virt create win11-test --iso windows11-latest

# Create VM from configuration file
portunix virt create vm-config.json     # Create VMs from JSON config
portunix virt create cluster-config.json --cluster  # Create VM cluster

# Backend-specific creation when needed
portunix qemu create custom-vm --iso ~/custom.iso
portunix virtualbox create test-vm --iso ~/test.iso
```

#### 3.2 VM Operations
```bash
# Universal lifecycle commands
portunix virt start ubuntu-test
portunix virt stop ubuntu-test [--force]
portunix virt restart ubuntu-test
portunix virt suspend ubuntu-test
portunix virt resume ubuntu-test
portunix virt delete ubuntu-test [--keep-disk]
portunix virt rm ubuntu-test       # Alias for delete

# Status & monitoring
portunix virt list                 # List all VMs
portunix virt status ubuntu-test   # Detailed VM status
portunix virt info ubuntu-test     # VM configuration

# Backend-specific when needed
portunix qemu list                 # QEMU VMs only
portunix virtualbox list           # VirtualBox VMs only
```

#### 3.3 State Management & Error Handling
```bash
# Smart state management
portunix virt start ubuntu-test    # Auto-detects if already running
# Output: "✓ VM 'ubuntu-test' is already running"

portunix virt stop ubuntu-test     # Auto-detects if already stopped
# Output: "✓ VM 'ubuntu-test' is already stopped"

portunix virt restart ubuntu-test  # Works regardless of current state
# If running: stops then starts
# If stopped: just starts

# State checking before operations
portunix virt status ubuntu-test --format simple
# Output: "running" | "stopped" | "suspended" | "error" | "not-found"

# Force operations when needed
portunix virt start ubuntu-test --force     # Force start even if running
portunix virt stop ubuntu-test --force      # Force kill (no graceful shutdown)
```

### Phase 4: Snapshot Management

#### 4.1 Snapshot Operations
```bash
# Universal snapshot commands
portunix virt snapshot create ubuntu-test clean-install
portunix virt snapshot create ubuntu-test "before testing" --description "Clean state before integration tests"

# Manage snapshots
portunix virt snapshot list ubuntu-test
portunix virt snapshot info ubuntu-test clean-install
portunix virt snapshot revert ubuntu-test clean-install
portunix virt snapshot delete ubuntu-test old-snapshot

# Automated snapshots
portunix virt snapshot auto ubuntu-test --interval daily --keep 7
```

### Phase 5: SSH Integration & Portunix Deployment

#### 5.1 SSH Setup & Access
```bash
# Universal SSH commands
portunix virt create ubuntu-test --enable-ssh --ssh-key ~/.ssh/id_rsa.pub
portunix virt ssh ubuntu-test
portunix virt ssh ubuntu-test --command "uname -a"

# Copy files
portunix virt copy ubuntu-test:/etc/hosts ./hosts.backup
portunix virt copy ./portunix ubuntu-test:/tmp/
```

#### 5.2 Smart SSH with Boot Waiting
```bash
# SSH with automatic boot waiting
portunix virt ssh ubuntu-test                    # Waits for boot if needed
# Output: "🔄 VM is starting, waiting for SSH availability..."
# Output: "⏳ Waiting for SSH (30s timeout)..."
# Output: "✅ SSH connection established"

# SSH with custom timeouts
portunix virt ssh ubuntu-test --wait-timeout 60s  # Wait up to 60 seconds
portunix virt ssh ubuntu-test --no-wait          # Fail immediately if not ready

# Check SSH availability
portunix virt ssh ubuntu-test --check            # Just check if SSH is ready
# Exit codes: 0 = ready, 1 = not ready, 2 = VM not running

# Auto-start and SSH
portunix virt ssh ubuntu-test --start            # Start VM if stopped, then SSH
# Output: "🚀 Starting VM 'ubuntu-test'..."
# Output: "⏳ Waiting for boot and SSH availability..."
# Output: "✅ Connected to ubuntu-test"
```

#### 5.3 Portunix Installation in VM
```bash
# Install Portunix inside VM
portunix virt install-portunix ubuntu-test

# Or during creation
portunix virt create ubuntu-dev --template ubuntu --install-portunix

# Execute Portunix commands in VM
portunix virt exec ubuntu-test "portunix system info"
portunix virt exec ubuntu-test "portunix install python"
```

### Phase 6: Advanced Features

#### 6.1 Network Configuration
```bash
# Universal networking commands
portunix virt network ubuntu-test --mode nat      # Default NAT
portunix virt network ubuntu-test --mode bridge   # Bridge networking
portunix virt network ubuntu-test --forward 8080:80  # Port forwarding
```

#### 6.2 Shared Folders & Clipboard
```bash
# Share folders between host and VM
portunix virt share ubuntu-test ~/projects:/mnt/projects
portunix virt share ubuntu-test --remove /mnt/projects

# Enable clipboard (QEMU: SPICE, VirtualBox: Guest Additions)
portunix virt enhance ubuntu-test --clipboard bidirectional
portunix virt enhance ubuntu-test --drag-drop enable
```

#### 6.3 Resource Management
```bash
# Modify VM resources
portunix virt resize ubuntu-test --ram 8G
portunix virt resize ubuntu-test --cpus 6
portunix virt resize ubuntu-test --disk +20G  # Expand disk
```

### Phase 7: Advanced Configuration & Testing Infrastructure

#### 7.1 Configuration-Based VM Creation
```bash
# Single VM from config file
portunix virt create vm-config.json
portunix virt create dev-server.json

# Cluster deployment from config
portunix virt create cluster-config.json --cluster
portunix virt create k8s-cluster.json --cluster

# Development environment setup
portunix virt create dev-env.json --environment
```

#### 7.2 VM-Based Testing
```bash
# Create test VM from template
portunix virt create test-env --template ubuntu --snapshot-after-create

# Create test environment from config
portunix virt create test-config.json

# Run tests in VM
portunix virt test test-env --script ./integration-tests.sh

# Auto-cleanup after tests
portunix virt test test-env --cleanup-after
```

#### 7.3 Multi-VM Scenarios
```bash
# Create VM cluster from config
portunix virt create k8s-cluster.json --cluster

# Manage cluster
portunix virt cluster start k8s-cluster
portunix virt cluster stop k8s-cluster
portunix virt cluster status k8s-cluster

# Individual VM management in cluster
portunix virt start k8s-master
portunix virt ssh k8s-worker-1
```

## Technical Implementation Details

### State Management Logic

#### VM State Detection
```go
// VM States
type VMState string

const (
    VMStateRunning   VMState = "running"
    VMStateStopped   VMState = "stopped"
    VMStateSuspended VMState = "suspended"
    VMStateError     VMState = "error"
    VMStateNotFound  VMState = "not-found"
    VMStateStarting  VMState = "starting"
    VMStateStopping  VMState = "stopping"
)

// State handling for commands
func (v *VirtManager) Start(vmName string, force bool) error {
    state := v.GetState(vmName)

    switch state {
    case VMStateRunning:
        if force {
            v.logger.Info("VM is running, force restarting...")
            return v.Restart(vmName)
        }
        v.logger.Success("✓ VM '%s' is already running", vmName)
        return nil

    case VMStateStopped:
        v.logger.Info("🚀 Starting VM '%s'...", vmName)
        return v.backend.Start(vmName)

    case VMStateSuspended:
        v.logger.Info("▶️ Resuming VM '%s'...", vmName)
        return v.backend.Resume(vmName)

    case VMStateNotFound:
        return fmt.Errorf("VM '%s' not found", vmName)

    default:
        return fmt.Errorf("VM '%s' is in invalid state: %s", vmName, state)
    }
}
```

#### SSH Availability Checking
```go
// SSH readiness with timeout
func (v *VirtManager) WaitForSSH(vmName string, timeout time.Duration) error {
    start := time.Now()
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    v.logger.Info("⏳ Waiting for SSH availability (timeout: %v)...", timeout)

    for {
        select {
        case <-ticker.C:
            if v.isSSHReady(vmName) {
                elapsed := time.Since(start)
                v.logger.Success("✅ SSH connection established (%.1fs)", elapsed.Seconds())
                return nil
            }

            elapsed := time.Since(start)
            if elapsed > timeout {
                return fmt.Errorf("SSH timeout after %v", timeout)
            }

            v.logger.Info("🔄 Waiting for SSH... (%.0fs elapsed)", elapsed.Seconds())

        case <-time.After(timeout):
            return fmt.Errorf("SSH timeout after %v", timeout)
        }
    }
}

func (v *VirtManager) isSSHReady(vmName string) bool {
    ip, err := v.backend.GetIP(vmName)
    if err != nil {
        return false
    }

    conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), 3*time.Second)
    if err != nil {
        return false
    }
    defer conn.Close()

    return true
}
```

#### Smart SSH Command
```go
func (v *VirtManager) SSH(vmName string, opts SSHOptions) error {
    // Check VM state first
    state := v.GetState(vmName)

    switch state {
    case VMStateNotFound:
        return fmt.Errorf("VM '%s' not found", vmName)

    case VMStateStopped:
        if opts.AutoStart {
            v.logger.Info("🚀 VM is stopped, starting...")
            if err := v.Start(vmName, false); err != nil {
                return fmt.Errorf("failed to start VM: %w", err)
            }
            // Continue to wait for SSH
        } else {
            return fmt.Errorf("VM '%s' is not running (use --start to auto-start)", vmName)
        }

    case VMStateSuspended:
        if opts.AutoStart {
            v.logger.Info("▶️ VM is suspended, resuming...")
            if err := v.Resume(vmName); err != nil {
                return fmt.Errorf("failed to resume VM: %w", err)
            }
            // Continue to wait for SSH
        } else {
            return fmt.Errorf("VM '%s' is suspended", vmName)
        }

    case VMStateStarting:
        v.logger.Info("🔄 VM is starting, waiting for boot...")
        // Continue to wait for SSH

    case VMStateRunning:
        // Check if SSH is immediately available
        if v.isSSHReady(vmName) {
            return v.connectSSH(vmName, opts)
        }
        // If not ready, fall through to wait
    }

    // Wait for SSH if needed
    if !opts.NoWait {
        timeout := opts.WaitTimeout
        if timeout == 0 {
            timeout = 30 * time.Second
        }

        if err := v.WaitForSSH(vmName, timeout); err != nil {
            return err
        }
    }

    return v.connectSSH(vmName, opts)
}
```

### Directory Structure
```
app/
├── virt/
│   ├── interface.go      # Universal virtualization interface
│   ├── config.go         # Configuration management
│   ├── backend.go        # Backend selection logic
│   ├── qemu/
│   │   ├── qemu.go       # QEMU backend implementation
│   │   ├── libvirt.go    # Libvirt integration
│   │   └── kvm.go        # KVM-specific features
│   ├── virtualbox/
│   │   ├── vbox.go       # VirtualBox backend
│   │   ├── vboxmanage.go # VBoxManage wrapper
│   │   └── guest.go      # Guest additions
│   ├── common/
│   │   ├── lifecycle.go  # Common VM lifecycle
│   │   ├── snapshot.go   # Common snapshot logic
│   │   ├── ssh.go        # SSH integration
│   │   ├── network.go    # Network configuration
│   │   ├── storage.go    # Disk & ISO management
│   │   └── templates.go  # VM templates
│   └── install.go        # Portunix installation in VMs
cmd/
├── virt.go              # Universal virt command
├── qemu.go              # QEMU-specific commands (if needed)
├── virtualbox.go        # VirtualBox-specific commands (if needed)
├── vm.go                # Legacy VM command (deprecate vm install-qemu)
├── vm_check.go          # VM check command (keep for compatibility)
├── virt_create.go       # Universal creation
├── virt_lifecycle.go    # Universal lifecycle
├── virt_snapshot.go     # Universal snapshots
├── virt_ssh.go          # Universal SSH
└── virt_iso.go          # ISO management (download, list, verify)
```

### Changes Required

1. **Remove**: `cmd/vm_install.go` - move functionality to install system
2. **Update**: `assets/install-packages.json` - add `virt` package with redirects
3. **Keep**: `cmd/vm_check.go` - rename to `virt check` but keep `vm check` alias
4. **Create**: `app/virt/` - new universal virtualization system

### Implementation Plan for Fixing Installation

#### Step 1: Remove Wrong Installation Command
- **File**: `cmd/vm_install.go`
- **Action**: Delete the entire file
- **Reason**: `portunix vm install-qemu` is wrong architecture

#### Step 2: Update install-packages.json
```json
{
  "virt": {
    "name": "Universal Virtualization",
    "description": "Auto-selects QEMU on Linux, VirtualBox on Windows",
    "platforms": {
      "linux": {
        "type": "redirect",
        "redirect_to": "qemu"
      },
      "windows": {
        "type": "redirect",
        "redirect_to": "virtualbox"
      }
    }
  },
  "virtualbox": {
    "name": "VirtualBox",
    "description": "Cross-platform virtualization",
    "platforms": {
      "windows": {
        "type": "chocolatey",
        "packages": ["virtualbox"]
      },
      "linux": {
        "type": "apt",
        "distributions": {
          "apt": {
            "packages": ["virtualbox", "virtualbox-ext-pack"]
          }
        }
      }
    }
  }
}
```

#### Step 3: Update vm.go to remove install subcommand
- **File**: `cmd/vm_install.go` init() function
- **Action**: Remove `vmCmd.AddCommand(vmInstallCmd)`
- **Keep**: `vmCmd.AddCommand(vmCheckCmd)` - check is still useful

#### Step 4: Create virt command structure
- **New file**: `cmd/virt.go`
- **Implementation**: Universal command that reads config and delegates
- **Commands**: create, start, stop, restart, suspend, resume, delete, rm, list, status, info, iso
- **Config support**: JSON file-based VM creation for single VMs, clusters, and environments

#### Step 5: Update documentation
- Change all references from `portunix vm install-qemu` to `portunix install qemu`
- Add examples for `portunix install virt`

### Configuration Files

#### VM Templates (`~/.portunix/vm-templates.json`)
```json
{
  "templates": {
    "ubuntu-24.04": {
      "iso": "ubuntu-24.04-desktop-amd64.iso",
      "os_variant": "ubuntu24.04",
      "min_ram": "2G",
      "recommended_ram": "4G",
      "min_disk": "25G",
      "recommended_disk": "40G",
      "features": ["uefi", "virtio"],
      "post_install": ["enable-ssh", "install-guest-tools"]
    },
    "windows11": {
      "iso": "Win11_24H2_English_x64v2.iso",
      "os_variant": "win11",
      "min_ram": "4G",
      "recommended_ram": "8G",
      "min_disk": "64G",
      "recommended_disk": "100G",
      "features": ["uefi", "tpm2.0", "secure-boot", "virtio"],
      "drivers": ["virtio-win.iso"]
    }
  }
}
```

#### VM Configuration Files (Custom Deployments)

**Single VM Configuration** (`vm-config.json`)
```json
{
  "vm": {
    "name": "dev-server",
    "template": "ubuntu-24.04",
    "resources": {
      "ram": "8G",
      "cpus": 4,
      "disk": "80G"
    },
    "network": {
      "mode": "bridge",
      "forwards": [
        {"host": 8080, "guest": 80},
        {"host": 2222, "guest": 22}
      ]
    },
    "features": {
      "ssh": true,
      "install_portunix": true,
      "snapshot_after_create": "fresh-install"
    },
    "post_create": [
      "portunix install docker",
      "portunix install nodejs"
    ]
  }
}
```

**VM Cluster Configuration** (`cluster-config.json`)
```json
{
  "cluster": {
    "name": "k8s-cluster",
    "description": "Kubernetes development cluster",
    "vms": [
      {
        "name": "k8s-master",
        "template": "ubuntu-24.04",
        "role": "master",
        "resources": {"ram": "4G", "cpus": 2, "disk": "40G"},
        "network": {"ip": "192.168.100.10"}
      },
      {
        "name": "k8s-worker-1",
        "template": "ubuntu-24.04",
        "role": "worker",
        "resources": {"ram": "4G", "cpus": 2, "disk": "40G"},
        "network": {"ip": "192.168.100.11"}
      },
      {
        "name": "k8s-worker-2",
        "template": "ubuntu-24.04",
        "role": "worker",
        "resources": {"ram": "4G", "cpus": 2, "disk": "40G"},
        "network": {"ip": "192.168.100.12"}
      }
    ],
    "shared_config": {
      "features": {
        "ssh": true,
        "install_portunix": true
      },
      "post_create": [
        "portunix install docker",
        "curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -",
        "echo 'deb https://apt.kubernetes.io/ kubernetes-xenial main' | sudo tee /etc/apt/sources.list.d/kubernetes.list",
        "sudo apt update && sudo apt install -y kubelet kubeadm kubectl"
      ]
    }
  }
}
```

**Development Environment Configuration** (`dev-env.json`)
```json
{
  "environment": {
    "name": "full-stack-dev",
    "description": "Complete development environment with database",
    "vms": [
      {
        "name": "web-server",
        "template": "ubuntu-24.04",
        "resources": {"ram": "4G", "cpus": 2, "disk": "60G"},
        "post_create": [
          "portunix install nodejs",
          "portunix install python",
          "portunix install docker"
        ]
      },
      {
        "name": "database",
        "template": "ubuntu-24.04",
        "resources": {"ram": "2G", "cpus": 1, "disk": "40G"},
        "post_create": [
          "portunix install docker",
          "docker run -d --name postgres -e POSTGRES_DB=appdb -e POSTGRES_USER=dev -e POSTGRES_PASSWORD=devpass -p 5432:5432 postgres:16"
        ]
      },
      {
        "name": "testing",
        "template": "windows11",
        "resources": {"ram": "8G", "cpus": 4, "disk": "100G"},
        "features": {
          "tpm": true,
          "secure_boot": true,
          "clipboard": "bidirectional"
        }
      }
    ]
  }
}
```

#### Package Configuration Enhancement
```json
// assets/install-packages.json addition
{
  "virt": {
    "name": "Universal Virtualization Stack",
    "description": "Auto-selects virtualization backend based on platform",
    "platforms": {
      "linux": {
        "type": "redirect",
        "redirect_to": "qemu",
        "default_variant": "default"
      },
      "windows": {
        "type": "redirect",
        "redirect_to": "virtualbox",
        "default_variant": "default"
      }
    }
  },
  "qemu": {
    "name": "QEMU/KVM Virtualization Stack",
    "platforms": {
      "linux": {
        "variants": {
      "default": {
        "description": "Standard QEMU/KVM with libvirt",
        "packages": ["qemu-system-x86", "libvirt-daemon-system", "libvirt-clients", "bridge-utils", "ovmf", "swtpm", "swtpm-tools"]
      },
      "full": {
        "description": "Full QEMU stack with GUI tools",
        "packages": ["@default", "virt-manager", "virt-viewer", "spice-client-gtk"]
      },
      "minimal": {
        "description": "Minimal QEMU for CLI-only usage",
        "packages": ["qemu-system-x86", "qemu-utils", "libvirt-daemon-system", "libvirt-clients"]
      }
    },
    "post_install_script": "scripts/setup-qemu.sh",
    "verification": {
      "commands": ["qemu-system-x86_64 --version", "virsh version"],
      "check_kvm": true
    }
  }
}
```

## Implementation Phases

### Phase 1: Universal Interface (Week 1)
- [ ] **FIX**: Remove `portunix vm install-qemu`, use `portunix install qemu` instead
- [ ] Add `virt` package to install-packages.json with backend auto-selection
- [ ] Implement configuration system for virtualization backend
- [ ] Create universal `virt` command structure
- [ ] Add backend detection and selection logic
- [ ] Map `virt` commands to backend implementations

### Phase 2: Backend Integration (Week 2)
- [ ] Complete QEMU backend implementation for `virt` commands
- [ ] Add VirtualBox backend support
- [ ] Implement backend-specific command translations
- [ ] Test cross-platform functionality
- [ ] Update `portunix system info` to show virtualization backend

### Phase 3: Core VM Management (Week 3)
- [ ] Enhance VM creation with templates
- [ ] Implement full lifecycle commands via `virt`
- [ ] Add VM listing and status commands
- [ ] Basic network configuration for both backends

### Phase 4: Snapshots & SSH (Week 4)
- [ ] Complete universal snapshot implementation
- [ ] Add SSH key injection during VM creation
- [ ] Implement `portunix virt ssh` command
- [ ] File copy functionality for both backends

### Phase 5: Advanced Features (Week 5)
- [ ] Automated Portunix installation in VMs
- [ ] Remote command execution
- [ ] Shared folder implementation
- [ ] Clipboard integration (SPICE/Guest Additions)

### Phase 6: Testing & Polish (Week 6)
- [ ] Cross-platform testing (Linux with QEMU, Windows with VirtualBox)
- [ ] Windows 11 VM testing on both backends
- [ ] Documentation and examples
- [ ] Error handling improvements

## Success Criteria

1. **Correct Installation**: `portunix install qemu/virt` works (no `vm install-qemu`)
2. **Universal Interface**: `portunix virt` commands work on both Linux and Windows
3. **Auto-detection**: Correct backend selection based on platform or config
4. **Installation**: `portunix install virt` auto-selects appropriate backend
5. **VM Creation**: Can create VMs with single command on any platform
6. **Lifecycle**: Full VM lifecycle management via universal commands
7. **Snapshots**: Reliable snapshot/revert on both backends
8. **SSH Access**: Seamless SSH connectivity to VMs
9. **Nested Portunix**: Can install and run Portunix inside VMs
10. **Configuration**: Easy backend switching via config file
11. **Performance**: VMs start in <30 seconds, snapshots in <5 seconds

## Benefits

1. **Testing Complementarity**: Full OS testing when containers are insufficient
2. **Container + VM Strategy**: Use containers for lightweight testing, VMs for full system tests
3. **Full OS Testing**: Test on complete OS environments with kernel-level operations
4. **Windows Support**: Native Windows testing on Linux hosts (not possible with containers)
5. **Snapshot Testing**: Easy rollback for trial software and experiments
6. **Integration Testing**: Test Portunix in clean environments
7. **Development Flexibility**: Multiple isolated development environments
8. **Hardware Virtualization**: Test hardware-specific features not available in containers

## Security Considerations

- Secure VM isolation using KVM
- SSH key-based authentication only
- Encrypted disk images support
- Network isolation options
- TPM 2.0 for Windows 11 security

## Documentation Requirements

1. **Quick Start Guide**: 5-minute guide to first VM
2. **Template Documentation**: All available templates and options
3. **Troubleshooting Guide**: Common issues and solutions
4. **Network Configuration**: Detailed networking setup
5. **Windows VM Guide**: Special considerations for Windows guests

## Dependencies

- Linux kernel with KVM module
- CPU with virtualization support (Intel VT-x / AMD-V)
- Minimum 8GB RAM for comfortable usage
- 100GB+ free disk space for VMs
- QEMU 6.0+ for Windows 11 support

## Testing Strategy

1. **Unit Tests**: Test individual VM operations
2. **Integration Tests**: Full VM lifecycle tests
3. **Cross-Distribution**: Test on Ubuntu, Debian, Fedora, Arch
4. **Performance Tests**: VM creation/start/snapshot timing
5. **Stress Tests**: Multiple concurrent VMs

## Related Issues

- Similar to #029 (Universal Container Command Implementation)
- Builds upon #017 (QEMU/KVM Windows Virtualization)
- Related to #008 (Virtual Development Disk)
- Enhances #020 (QEMU Clipboard Integration)
- Supports testing for #027 (Container Lifecycle Management)

## Risk Mitigation

- **Risk**: Complexity of QEMU/libvirt integration
  - **Mitigation**: Start with basic features, iterate
- **Risk**: Cross-distribution compatibility
  - **Mitigation**: Extensive testing matrix
- **Risk**: Windows 11 requirements (TPM, Secure Boot)
  - **Mitigation**: Pre-configured templates with required settings

## Success Metrics

- VM creation success rate >95%
- SSH connection success rate >99%
- Snapshot/revert success rate 100%
- User documentation completeness 100%
- Test coverage >80%

## Notes

- Priority on Linux VM support first, Windows second
- Consider future integration with cloud providers (AWS, Azure)
- Potential for VM export/import functionality
- Consider integration with Proxmox/ESXi in future

## Status
📋 Open

## Priority
Critical (Testing Infrastructure Dependency)

## Type
Feature

## Labels
`enhancement`, `virtualization`, `qemu`, `kvm`, `virtualbox`, `universal-interface`, `testing`, `infrastructure`, `critical`