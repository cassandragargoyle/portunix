# Portunix Virt Command

## Quick Start

The `virt` command provides universal virtualization management supporting multiple backends including QEMU/KVM, VirtualBox, VMware, and Hyper-V for cross-platform virtual machine operations.

### Simplest Usage
```bash
# Create a new VM
portunix virt create myvm --iso ubuntu.iso

# Start VM
portunix virt start myvm

# SSH into VM
portunix virt ssh myvm
```

### Basic Syntax
```bash
portunix virt [subcommand] [vm-name] [options]
```

### Common Subcommands
- `create` - Create new virtual machine
- `start` - Start virtual machine
- `stop` - Stop virtual machine
- `ssh` - SSH into virtual machine
- `list` - List virtual machines
- `remove` - Remove virtual machine
- `status` - Check VM status
- `console` - Access VM console

## Intermediate Usage

### Virtual Machine Creation

Create VMs with various configurations:

```bash
# Basic VM with ISO
portunix virt create dev-vm --iso ubuntu-22.04.iso

# VM with specific resources
portunix virt create dev-vm \
  --iso ubuntu-22.04.iso \
  --ram 4G \
  --disk 50G \
  --cpu 2

# VM with network configuration
portunix virt create dev-vm \
  --iso ubuntu-22.04.iso \
  --network bridge \
  --mac 52:54:00:12:34:56

# VM from template
portunix virt create dev-vm --template ubuntu-dev

# VM with cloud-init
portunix virt create cloud-vm \
  --image ubuntu-cloud.img \
  --cloud-init user-data.yaml
```

### Virtual Machine Management

Manage VM lifecycle:

```bash
# List all VMs
portunix virt list

# List running VMs only
portunix virt list --running

# VM detailed status
portunix virt status myvm --detailed

# Start VM with specific options
portunix virt start myvm --headless

# Stop VM gracefully
portunix virt stop myvm --graceful

# Force stop VM
portunix virt stop myvm --force

# Restart VM
portunix virt restart myvm

# Pause VM
portunix virt pause myvm

# Resume VM
portunix virt resume myvm
```

### VM Configuration

Modify VM settings:

```bash
# Change VM resources
portunix virt config myvm --ram 8G --cpu 4

# Add storage
portunix virt storage add myvm --size 100G --type ssd

# Add network interface
portunix virt network add myvm --type nat

# Set boot order
portunix virt boot myvm --order cdrom,hd

# Enable/disable features
portunix virt feature myvm --enable kvm --disable usb
```

### Snapshot Management

VM snapshot operations:

```bash
# Create snapshot
portunix virt snapshot create myvm --name "clean-install"

# List snapshots
portunix virt snapshot list myvm

# Restore snapshot
portunix virt snapshot restore myvm --name "clean-install"

# Delete snapshot
portunix virt snapshot delete myvm --name "old-snapshot"

# Export snapshot
portunix virt snapshot export myvm --name "backup" --output vm-backup.qcow2
```

### Remote Access

Access VMs remotely:

```bash
# SSH access (requires guest agent)
portunix virt ssh myvm

# SSH with specific user
portunix virt ssh myvm --user developer

# SSH with port forwarding
portunix virt ssh myvm --tunnel 8080:localhost:80

# VNC console access
portunix virt console myvm --vnc

# SPICE console access
portunix virt console myvm --spice

# Serial console
portunix virt console myvm --serial
```

## Advanced Usage

### Virtualization Backend Support

Portunix supports multiple virtualization backends:

#### QEMU/KVM (Linux)
```bash
# Create KVM-accelerated VM
portunix virt create myvm --backend qemu --accel kvm --iso linux.iso

# Check KVM support
portunix virt backend qemu --check-kvm

# QEMU-specific options
portunix virt create myvm \
  --backend qemu \
  --machine q35 \
  --firmware uefi \
  --tpm enable
```

#### VirtualBox (Cross-platform)
```bash
# Create VirtualBox VM
portunix virt create myvm --backend virtualbox --iso windows.iso

# VirtualBox-specific features
portunix virt create myvm \
  --backend virtualbox \
  --guest-additions \
  --shared-folders /host/path:/guest/path
```

#### VMware (Commercial)
```bash
# Create VMware VM
portunix virt create myvm --backend vmware --iso vmware-tools.iso

# VMware Workstation integration
portunix virt create myvm \
  --backend vmware-workstation \
  --vmx-version 19 \
  --tools-install
```

#### Hyper-V (Windows)
```bash
# Create Hyper-V VM
portunix virt create myvm --backend hyperv --iso server.iso

# Hyper-V specific options
portunix virt create myvm \
  --backend hyperv \
  --generation 2 \
  --secure-boot \
  --checkpoints enable
```

### VM Templates

Pre-configured VM templates:

```bash
# List available templates
portunix virt template list

# Create template from VM
portunix virt template create ubuntu-dev --from myvm

# Use template
portunix virt create newvm --template ubuntu-dev

# Import template
portunix virt template import template.json

# Export template
portunix virt template export ubuntu-dev > template.json
```

Template definition example:
```json
{
  "name": "ubuntu-dev",
  "description": "Ubuntu development environment",
  "base": {
    "os": "ubuntu",
    "version": "22.04",
    "arch": "amd64"
  },
  "resources": {
    "ram": "4G",
    "cpu": 2,
    "disk": "50G"
  },
  "software": [
    "build-essential",
    "git",
    "nodejs",
    "python3"
  ],
  "users": [
    {
      "name": "developer",
      "sudo": true,
      "ssh_key": "~/.ssh/id_rsa.pub"
    }
  ],
  "network": {
    "type": "bridge",
    "dhcp": true
  }
}
```

### Cloud Integration

Integration with cloud platforms:

```bash
# Import cloud image
portunix virt cloud import \
  --provider aws \
  --image ami-12345678 \
  --name aws-ubuntu

# Convert cloud image
portunix virt cloud convert \
  --input cloud.vmdk \
  --output local.qcow2 \
  --format qcow2

# Export to cloud
portunix virt cloud export myvm \
  --provider azure \
  --resource-group mygroup
```

### Automated Provisioning

Automated VM setup and configuration:

```bash
# Cloud-init provisioning
portunix virt create auto-vm \
  --image ubuntu-cloud.img \
  --cloud-init provision.yaml \
  --user-data user.yaml \
  --meta-data meta.yaml

# Ansible provisioning
portunix virt provision myvm \
  --ansible playbook.yml \
  --inventory inventory.ini

# Script-based provisioning
portunix virt provision myvm \
  --script setup.sh \
  --wait-for-ssh
```

Cloud-init example (`provision.yaml`):
```yaml
#cloud-config
users:
  - name: developer
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2E...

packages:
  - git
  - curl
  - nodejs
  - npm

runcmd:
  - npm install -g yarn
  - git clone https://github.com/user/project.git /home/developer/project
  - chown -R developer:developer /home/developer/project
```

### Storage Management

Advanced storage operations:

```bash
# Create storage pool
portunix virt storage pool create --name default --path /var/lib/vms

# List storage pools
portunix virt storage pool list

# Create volume
portunix virt storage volume create --pool default --name vm-disk --size 100G

# Attach storage to VM
portunix virt storage attach myvm --volume vm-disk --device vdb

# Storage migration
portunix virt storage migrate myvm --target /new/storage/path

# Backup VM storage
portunix virt storage backup myvm --output backup.qcow2
```

### Network Configuration

Advanced networking features:

```bash
# Create virtual network
portunix virt network create dev-network \
  --subnet 192.168.100.0/24 \
  --dhcp-range 192.168.100.100,192.168.100.200

# List networks
portunix virt network list

# Connect VM to network
portunix virt network attach myvm --network dev-network

# Port forwarding
portunix virt network forward myvm \
  --host-port 8080 \
  --guest-port 80 \
  --protocol tcp

# Network isolation
portunix virt network isolate dev-network --from production-network
```

### Performance Tuning

Optimize VM performance:

```bash
# CPU optimization
portunix virt tune myvm \
  --cpu-model host \
  --cpu-features +vmx,+svm \
  --numa auto

# Memory optimization
portunix virt tune myvm \
  --memory-balloon enable \
  --memory-huge-pages \
  --memory-shared

# Storage optimization
portunix virt tune myvm \
  --disk-cache writeback \
  --disk-io native \
  --disk-discard unmap

# Network optimization
portunix virt tune myvm \
  --network-model virtio \
  --network-queues 4
```

## Expert Tips & Tricks

### 1. High Availability Setup

```bash
# Create clustered VMs
portunix virt cluster create ha-cluster \
  --nodes node1,node2,node3 \
  --shared-storage /shared/vms

# Live migration
portunix virt migrate myvm --to node2 --live

# Failover configuration
portunix virt ha configure myvm \
  --failover-node node2 \
  --health-check http://vm:8080/health
```

### 2. GPU Passthrough

```bash
# Enable GPU passthrough
portunix virt gpu passthrough myvm \
  --device 0000:01:00.0 \
  --driver vfio-pci

# SR-IOV configuration
portunix virt sriov configure \
  --device eth0 \
  --vfs 4

# IOMMU setup
portunix virt iommu enable --intel
```

### 3. Development Workflows

```bash
# Development VM with synced folders
portunix virt create dev-vm \
  --template ubuntu-dev \
  --sync-folder $(pwd):/workspace \
  --auto-reload

# Test environment provisioning
portunix virt test-env create \
  --from Vagrantfile \
  --provider portunix
```

### 4. Backup and Disaster Recovery

```bash
# Full VM backup
portunix virt backup myvm \
  --include-config \
  --include-snapshots \
  --compress \
  --output vm-full-backup.tar.gz

# Incremental backup
portunix virt backup myvm \
  --incremental \
  --base-backup previous-backup.tar.gz

# Disaster recovery
portunix virt restore \
  --backup vm-full-backup.tar.gz \
  --target new-host
```

### 5. Security Hardening

```bash
# Enable VM security features
portunix virt security myvm \
  --secure-boot \
  --tpm 2.0 \
  --encryption luks2

# Isolation configuration
portunix virt isolate myvm \
  --no-network \
  --read-only-disk \
  --restricted-io
```

## Monitoring and Diagnostics

### VM Monitoring

```bash
# Real-time VM monitoring
portunix virt monitor myvm

# Resource usage statistics
portunix virt stats myvm --duration 60s

# Performance metrics
portunix virt metrics myvm --export prometheus

# Health check
portunix virt health myvm --comprehensive
```

### Diagnostics and Troubleshooting

```bash
# VM diagnostics
portunix virt diagnose myvm

# Log analysis
portunix virt logs myvm --system --guest

# Performance analysis
portunix virt analyze myvm --performance --bottlenecks

# Network connectivity test
portunix virt test myvm --network --connectivity
```

## Integration with Portunix Ecosystem

### Container Integration

```bash
# VM with container support
portunix virt create container-vm \
  --template ubuntu-dev \
  --enable-docker \
  --enable-podman

# Run containers in VM
portunix virt exec myvm "portunix docker run nginx"
```

### Plugin Integration

```bash
# Install Portunix in VM
portunix virt provision myvm --install-portunix

# Plugin management in VM
portunix virt exec myvm "portunix plugin install agile-software-development"
```

### MCP Server in VM

```bash
# Setup MCP server in VM
portunix virt mcp-setup myvm \
  --port 3000 \
  --external-access

# Connect to VM MCP server
portunix mcp connect vm://myvm:3000
```

## Platform-Specific Features

### Linux (KVM/QEMU)

```bash
# Native KVM features
portunix virt create myvm \
  --backend qemu \
  --kvm \
  --cpu host-passthrough \
  --nested-virtualization

# QEMU guest agent
portunix virt agent install myvm --qemu-guest-agent
```

### Windows (Hyper-V)

```bash
# Hyper-V specific features
portunix virt create winvm \
  --backend hyperv \
  --generation 2 \
  --secure-boot \
  --dynamic-memory 2G-8G

# Integration services
portunix virt integration-services winvm --install
```

### macOS (QEMU)

```bash
# macOS VM (where legally permitted)
portunix virt create macos-vm \
  --backend qemu \
  --machine q35 \
  --firmware edk2 \
  --device-model virtio
```

## API Integration

### REST API

```bash
# VM operations via API
curl -X POST http://localhost:8080/api/virt/create \
  -H "Content-Type: application/json" \
  -d '{"name": "myvm", "template": "ubuntu-dev"}'
```

### gRPC API

```go
// Go client example
client := portunix.NewClient()
vm, err := client.Virt.Create(context.Background(), &VirtCreateRequest{
    Name:     "myvm",
    Template: "ubuntu-dev",
})
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_VIRT_BACKEND` | Default virtualization backend | auto |
| `PORTUNIX_VIRT_STORAGE` | Default storage location | ~/.portunix/vms |
| `PORTUNIX_VIRT_TIMEOUT` | VM operation timeout | 300s |
| `PORTUNIX_VIRT_LOG_LEVEL` | Logging level | info |
| `PORTUNIX_VIRT_SSH_KEY` | Default SSH key | ~/.ssh/id_rsa |

## Related Commands

- [`install`](install.md) - Install virtualization software
- [`system`](system.md) - System information and capabilities
- [`docker`](docker.md) - Container management
- [`sandbox`](sandbox.md) - Windows Sandbox

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--backend` | string | `auto` | Virtualization backend |
| `--iso` | string | - | ISO image file |
| `--image` | string | - | Disk image file |
| `--template` | string | - | VM template |
| `--ram` | string | `2G` | RAM allocation |
| `--cpu` | int | `2` | CPU cores |
| `--disk` | string | `20G` | Disk size |
| `--network` | string | `nat` | Network type |
| `--headless` | boolean | `false` | Headless mode |
| `--ssh-key` | string | `~/.ssh/id_rsa` | SSH key file |
| `--cloud-init` | string | - | Cloud-init configuration |
| `--provision` | string | - | Provisioning script |
| `--snapshot` | string | - | Snapshot name |
| `--force` | boolean | `false` | Force operation |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | VM not found |
| 3 | Backend not available |
| 4 | Resource allocation failed |
| 5 | Network configuration failed |
| 6 | Storage operation failed |
| 7 | SSH connection failed |
| 8 | Provisioning failed |
| 9 | Snapshot operation failed |
| 10 | Permission denied |

## Version History

- **v1.5.0** - Added cloud integration
- **v1.4.0** - Implemented multi-backend support
- **v1.3.0** - Added template system
- **v1.2.0** - Enhanced networking features
- **v1.1.0** - Added snapshot management
- **v1.0.0** - Initial virtualization support