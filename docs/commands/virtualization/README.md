# Virtualization Commands

Cross-platform virtual machine management supporting multiple backends including QEMU/KVM, VirtualBox, VMware, and Hyper-V.

## Commands in this Category

### [`virt`](virt.md) - Virtual Machine Management
Universal VM management with multi-backend support, automated provisioning, and cloud integration.

**Quick Examples:**
```bash
portunix virt create myvm --iso ubuntu.iso    # Create VM from ISO
portunix virt start myvm                      # Start virtual machine
portunix virt ssh myvm                        # SSH into VM
portunix virt snapshot create myvm            # Create VM snapshot
```

**Key Features:**
- Multi-backend support (QEMU/KVM, VirtualBox, VMware, Hyper-V)
- Automated provisioning with cloud-init and Ansible
- VM templates for rapid deployment
- Built-in SSH access and management
- Snapshot and backup management
- Cross-platform compatibility

**Common Use Cases:**
- Development environment isolation
- Cross-platform testing and validation
- Legacy application support
- Security testing and sandboxing
- Training and education environments

---

### `sandbox` - Windows Sandbox Integration *(Coming Soon)*
Windows Sandbox integration for lightweight, disposable Windows environments.

**Planned Features:**
- Quick Windows environment creation
- Application testing in isolation
- Malware analysis sandbox
- Registry and filesystem isolation

## Category Overview

The **Virtualization** category provides comprehensive virtual machine management that abstracts the complexity of different virtualization platforms while maintaining their unique capabilities.

### Multi-Backend Architecture

Portunix automatically detects and utilizes available virtualization backends:

```
┌─────────────────┐
│   Portunix VM   │
│    Management   │
└─────────┬───────┘
          │
    ┌─────┴─────┐
    │  Backend  │
    │ Detection │
    └─────┬─────┘
          │
┌─────────┼─────────┐─────────┐─────────┐
│         │         │         │         │
▼         ▼         ▼         ▼         ▼
QEMU/KVM  VirtualBox VMware   Hyper-V   Cloud
Linux     Cross     Commercial Windows  Providers
```

### Backend-Specific Features

#### QEMU/KVM (Linux)
```bash
# High-performance virtualization on Linux
portunix virt create linux-vm \
  --backend qemu \
  --accel kvm \
  --cpu host-passthrough \
  --iso ubuntu-server.iso
```

#### VirtualBox (Cross-platform)
```bash
# Cross-platform virtualization
portunix virt create cross-vm \
  --backend virtualbox \
  --guest-additions \
  --shared-folders /host/path:/guest/path \
  --iso windows10.iso
```

#### VMware (Commercial)
```bash
# Enterprise virtualization
portunix virt create enterprise-vm \
  --backend vmware \
  --vmx-version 19 \
  --tools-install \
  --iso rhel8.iso
```

#### Hyper-V (Windows)
```bash
# Windows native virtualization
portunix virt create windows-vm \
  --backend hyperv \
  --generation 2 \
  --secure-boot \
  --iso server2022.iso
```

## VM Lifecycle Management

### Complete VM Lifecycle
```bash
# 1. Create VM from template or ISO
portunix virt create development-vm \
  --template ubuntu-dev \
  --ram 8G \
  --cpu 4 \
  --disk 100G

# 2. Provision with software
portunix virt provision development-vm \
  --cloud-init cloud-config.yaml \
  --ansible playbook.yml

# 3. Create baseline snapshot
portunix virt snapshot create development-vm --name "baseline"

# 4. Use for development
portunix virt ssh development-vm

# 5. Backup before major changes
portunix virt backup development-vm --output vm-backup.tar.gz

# 6. Clone for different projects
portunix virt clone development-vm --name project-a-vm

# 7. Archive when no longer needed
portunix virt archive development-vm --keep-snapshots
```

### Template System
```bash
# Create reusable templates
portunix virt template create ubuntu-dev \
  --base ubuntu:22.04 \
  --packages "git nodejs python docker" \
  --users developer:sudo \
  --ssh-key ~/.ssh/id_rsa.pub

# Use templates for rapid deployment
portunix virt create new-project --template ubuntu-dev

# Share templates across team
portunix virt template export ubuntu-dev > ubuntu-dev-template.yaml
portunix virt template import ubuntu-dev-template.yaml
```

## Integration with Portunix Ecosystem

### With Core Commands
```bash
# Install Portunix in VM during creation
portunix virt create vm-with-portunix \
  --iso ubuntu.iso \
  --provision "curl -sSL https://install.portunix.ai | bash"

# Use Portunix inside VM
portunix virt ssh my-vm "portunix install nodejs python"
```

### With Container Integration
```bash
# VM with container support
portunix virt create container-host \
  --template ubuntu-dev \
  --enable-docker \
  --enable-podman

# Run containers in VM
portunix virt ssh container-host "portunix docker run nginx"
```

### With Plugin System
```bash
# VM with plugin ecosystem
portunix virt provision plugin-vm \
  --install-portunix \
  --enable-plugins

# Install plugins in VM
portunix virt exec plugin-vm "portunix plugin install agile-software-development"
```

### With MCP Integration
```bash
# MCP-enabled VM for AI development
portunix virt create ai-dev-vm \
  --template ubuntu-dev \
  --mcp-server-setup \
  --port-forward 3000:3000

# Connect to VM MCP server
portunix mcp connect vm://ai-dev-vm:3000
```

## Advanced Virtualization Workflows

### Development Environment Provisioning
```yaml
# cloud-config.yaml for automated setup
#cloud-config
users:
  - name: developer
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2E...

packages:
  - git
  - curl
  - build-essential

runcmd:
  - curl -sSL https://install.portunix.ai | bash
  - sudo -u developer portunix install nodejs python docker
  - usermod -aG docker developer
```

```bash
# Create development VM with cloud-init
portunix virt create dev-vm \
  --image ubuntu-cloud.img \
  --cloud-init cloud-config.yaml \
  --ram 8G \
  --cpu 4
```

### Cross-Platform Testing
```bash
# Create test matrix
for os in ubuntu-22.04 debian-12 centos-8 windows-server-2022; do
  portunix virt create test-$os \
    --template $os \
    --ram 4G \
    --cpu 2 \
    --snapshot-on-create
done

# Run tests across all platforms
for vm in test-*; do
  portunix virt ssh $vm "portunix install myapp && myapp test"
done
```

### CI/CD Integration
```bash
# Ephemeral test environments
portunix virt create ci-test-$(date +%s) \
  --template test-base \
  --ram 4G \
  --cpu 2 \
  --auto-destroy 2h

# Run CI pipeline in isolated VM
portunix virt exec ci-test-* "
  git clone $REPO_URL
  cd $(basename $REPO_URL .git)
  ./run-tests.sh
"
```

## Performance and Resource Management

### Resource Optimization
```bash
# Optimize VM performance
portunix virt tune my-vm \
  --cpu-model host \
  --numa auto \
  --memory-balloon enable \
  --disk-cache writeback

# Resource monitoring
portunix virt monitor my-vm --real-time
portunix virt metrics my-vm --export prometheus
```

### High Availability Setup
```bash
# Create clustered VMs
portunix virt cluster create ha-cluster \
  --nodes 3 \
  --shared-storage /shared/vms \
  --failover-policy automatic

# Live migration
portunix virt migrate my-vm --to node2 --live --verify
```

### Backup and Disaster Recovery
```bash
# Automated backup strategy
portunix virt backup-schedule my-vm \
  --frequency daily \
  --retention 30d \
  --incremental \
  --verify

# Disaster recovery test
portunix virt restore \
  --backup vm-backup-20241215.tar.gz \
  --target recovery-host \
  --verify-integrity
```

## Security and Isolation

### Security Features
```bash
# Secure VM configuration
portunix virt create secure-vm \
  --iso hardened-linux.iso \
  --secure-boot \
  --tpm 2.0 \
  --encryption luks2 \
  --network isolated

# Security scanning
portunix virt security-scan my-vm \
  --vulnerability-check \
  --compliance-check \
  --report security-report.pdf
```

### Network Isolation
```bash
# Create isolated network
portunix virt network create isolated-lab \
  --subnet 192.168.100.0/24 \
  --no-internet \
  --firewall-rules strict

# DMZ configuration
portunix virt network create dmz \
  --subnet 10.0.1.0/24 \
  --firewall-rules dmz.rules \
  --intrusion-detection
```

## Cloud Integration

### Hybrid Cloud Workflows
```bash
# Import cloud images
portunix virt cloud import \
  --provider aws \
  --image ami-12345678 \
  --convert qcow2

# Export to cloud
portunix virt cloud export my-vm \
  --provider azure \
  --resource-group production \
  --optimize-for-cloud
```

### Multi-Cloud Development
```bash
# Test across cloud providers
portunix virt create aws-test --cloud-image aws-ubuntu
portunix virt create azure-test --cloud-image azure-ubuntu
portunix virt create gcp-test --cloud-image gcp-ubuntu

# Validate deployment compatibility
for vm in *-test; do
  portunix virt ssh $vm "deploy-app.sh && test-app.sh"
done
```

## Platform-Specific Features

### Linux (QEMU/KVM)
```bash
# Advanced Linux virtualization
portunix virt create linux-vm \
  --backend qemu \
  --kvm \
  --cpu host-passthrough \
  --numa topology \
  --hugepages \
  --vfio-gpu 0000:01:00.0
```

### Windows (Hyper-V)
```bash
# Windows-specific features
portunix virt create windows-vm \
  --backend hyperv \
  --generation 2 \
  --secure-boot \
  --dynamic-memory 2G-16G \
  --integration-services \
  --enhanced-session
```

### macOS (Limited Support)
```bash
# macOS virtualization (where legally permitted)
portunix virt create macos-vm \
  --backend qemu \
  --machine q35 \
  --firmware edk2 \
  --device-model virtio \
  --legal-compliance-check
```

## Troubleshooting and Diagnostics

### Common Issues
```bash
# VM won't start
portunix virt diagnose my-vm --startup-issues

# Performance problems
portunix virt analyze my-vm --performance --bottlenecks

# Network connectivity
portunix virt test my-vm --network --connectivity

# Storage issues
portunix virt storage diagnose my-vm --corruption-check
```

### Debug Mode
```bash
# Verbose VM operations
portunix virt create debug-vm --debug --trace

# Backend-specific debugging
portunix virt --backend qemu --debug start my-vm

# Console access for troubleshooting
portunix virt console my-vm --serial --capture-output
```

## Best Practices

### VM Management
- Use templates for consistent environments
- Regular snapshots before major changes
- Monitor resource usage and optimize
- Implement proper backup strategies

### Security
- Enable encryption for sensitive VMs
- Use network isolation for testing
- Regular security scanning and updates
- Implement proper access controls

### Performance
- Allocate appropriate resources
- Use backend-specific optimizations
- Monitor and tune performance regularly
- Consider hardware requirements

## Future Roadmap

### Planned Features
- **Enhanced Container Integration** - Seamless VM-container workflows
- **Advanced Networking** - SDN and service mesh support
- **GPU Virtualization** - Enhanced GPU passthrough and sharing
- **Edge Computing** - Lightweight edge VM deployment

### Integration Improvements
- **Kubernetes in VMs** - Native K8s cluster deployment
- **Multi-Cloud Orchestration** - Unified cloud VM management
- **Advanced Automation** - Infrastructure as Code integration
- **Enhanced Monitoring** - ML-based performance optimization

## Related Categories

- **[Core](../core/)** - Install software in VMs
- **[Containers](../containers/)** - Containers within VMs
- **[Plugins](../plugins/)** - VM-aware plugins
- **[Integration](../integration/)** - MCP in VMs

---

*Virtualization commands provide complete isolation and cross-platform compatibility for complex development and testing scenarios.*