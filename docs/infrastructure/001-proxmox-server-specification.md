# ProxMox Server Specification for Portunix Development Infrastructure

## Overview
This document specifies the hardware requirements for a ProxMox VE server dedicated to Portunix development and testing infrastructure. The server will provide virtual machines for testing Docker/Podman installations, cross-platform compatibility, and automated testing scenarios.

**IMPORTANT NOTE**: Until the ProxMox server is acquired, QEMU and potentially VirtualBox will be used as temporary virtualization solutions for Portunix development. Therefore, tasks related to QEMU and VirtualBox support in the Portunix project must be given higher priority.

## Infrastructure Evolution Path

### Current State: Local Testing (⚠️ Problematic)
```
┌─────────────────────────────────────┐
│     Development Workstation         │
├─────────────────────────────────────┤
│  IDE & Development Tools            │
│  ├── Portunix Source Code           │
│  ├── Testing (DIRECT on HOST)       │
│  │   ├── Package Installation       │
│  │   ├── Docker/Podman Changes      │
│  │   └── System Modifications       │
│  └── Result: HOST CONTAMINATION ⚠️  │
└─────────────────────────────────────┘
Problems:
- Testing affects development environment
- Irreversible system changes
- Package conflicts
- No rollback capability
```

### Intermediate State: Local Virtualization (🔄 Current Solution)
```
┌─────────────────────────────────────┐
│     Development Workstation         │
├─────────────────────────────────────┤
│  IDE & Development Tools            │
│  └── QEMU/VirtualBox                │
│      ├── Ubuntu VM (4GB RAM)        │
│      ├── Debian VM (4GB RAM)        │
│      └── Windows VM (8GB RAM)       │
├─────────────────────────────────────┤
│  Resource Competition:              │
│  - Development: 50% resources       │
│  - VMs: 50% resources               │
│  - Result: SLOW PERFORMANCE ⚠️      │
└─────────────────────────────────────┘
Problems:
- High resource consumption
- Slow test execution
- Limited concurrent VMs
- Development workstation stress
```

### Target State: ProxMox Server (✅ Optimal)
```
┌─────────────────────────┐     ┌─────────────────────────┐
│  Development Workstation│     │    ProxMox Server       │
├─────────────────────────┤     ├─────────────────────────┤
│  IDE & Development      │────▶│  Testing Infrastructure │
│  100% resources for dev │ SSH │  ├── Ubuntu VMs (5x)    │
│  - Fast compilation     │     │  ├── Debian VMs (3x)    │
│  - Smooth IDE operation │     │  ├── Windows VMs (2x)   │
│  - No VM overhead       │     │  ├── Gitea Server       │
└─────────────────────────┘     │  └── Jenkins CI/CD      │
                                └─────────────────────────┘
Benefits:
✅ Full workstation performance for development
✅ Unlimited test iterations
✅ Parallel test execution
✅ Clean test environments
✅ Snapshot/rollback capability
```

### Future Vision: Hybrid Cloud Infrastructure (🚀 Next Evolution)

#### Cloud Cost Comparison (Wedos VPS)
For equivalent resources to Price-Optimized Configuration:
- **CPU**: 6 vCPU (similar to Ryzen 5 7600)
- **RAM**: 32 GB
- **Storage**: 200 GB SSD (sufficient for VM testing)
- **Wedos VPS Estimate**: ~1,800 CZK/month = **~21,600 CZK/year**
- **Comparison**: 1.2x the cost of local ProxMox server per year
- **Benefits**: No maintenance, 24/7 operation, instant scaling, no electricity costs
```
┌─────────────────────────┐
│  Development Workstation│
├─────────────────────────┤
│  Portunix CLI           │
└───────────┬─────────────┘
            │
      ┌─────▼─────┐
      │ Portunix  │
      │ Orchestr. │
      └─────┬─────┘
            │
    ┌───────┴───────────────────────┐
    │                               │
┌───▼──────────────┐    ┌──────────▼──────────┐
│  ProxMox Server  │    │   Cloud Providers   │
├──────────────────┤    ├─────────────────────┤
│ Local VMs        │    │ AWS EC2 Instances   │
│ Fast iterations  │    │ Azure VMs           │
│ Private testing  │    │ GCP Compute         │
└──────────────────┘    └─────────────────────┘
            │                      │
    ┌───────┴──────────────────────┴───────┐
    │                                       │
┌───▼──────────────┐    ┌──────────────────▼──┐
│ Local Kubernetes │    │  Cloud Kubernetes   │
├──────────────────┤    ├─────────────────────┤
│ K3s/MicroK8s     │    │ EKS/AKS/GKE         │
│ Container tests  │    │ Scale testing       │
│ Orchestration    │    │ Production simul.   │
└──────────────────┘    └─────────────────────┘

Capabilities:
🚀 Hybrid local/cloud testing
🚀 Kubernetes-native testing
🚀 Multi-cloud support
🚀 Cost-optimized resource usage
🚀 Global test distribution
🚀 24/7 continuous testing at same cost as local server
🚀 Automated regression testing for external dependencies
   Example: Daily tests of Java/Python/NodeJS installations
   to detect when upstream changes break Portunix
🚀 Early warning system before users encounter issues
🚀 Notification pipeline for detected failures
```

## Purpose
- **Development Testing**: Create VMs for testing Portunix across different operating systems
- **Container Runtime Testing**: Test Docker and Podman installation/uninstallation scenarios
- **CI/CD Pipeline**: Automated testing in clean, reproducible environments
- **Cross-platform Validation**: Windows, Linux (Ubuntu, Debian, CentOS, Fedora) testing
- **Isolation**: Test dangerous operations without affecting development workstations
- **Gitea Server Migration**: Host existing Gitea development server in dedicated VM
- **Consolidated Infrastructure**: Single server for both development repository and testing VMs

## Hardware Specification

### Form Factor
- **Mini-ITX (mITX)** motherboard for compact size
- Suitable for home lab/office environment
- Low power consumption
- Quiet operation

## a) Price-Optimized Configuration

### CPU
- **AMD Ryzen 5 7600** (Socket AM5, 6 cores, 12 threads, 3.8-5.1 GHz, 65W TDP)
- **Estimation: 4,290 CZK** (Alza.cz)

### Motherboard
- **ASRock B650M Pro RS WiFi** (AM5, Micro-ATX)
- **Estimation: 3,200 CZK** (Alza.cz)
- **Features**: AM5 socket, 4x DDR5 slots, WiFi 6E, Bluetooth, 2.5GbE LAN

### Memory (RAM)
- **32GB DDR5-5600** (2x 16GB kit)
- **Estimation: 6,000 CZK** (Alza.cz)
- **Reasoning**: Sufficient for 4-5 concurrent VMs

### Storage
- **Primary**: NVMe SSD 500GB (budget model)
- **Estimation: 1,500 CZK** (Alza.cz)
- **VM Storage**: NVMe SSD 1TB (budget model)
- **Estimation: 3,000 CZK** (Alza.cz)

### Case & Cooling
- **CHIEFTEC CI-02B-OP Mini Tower** (Micro-ATX/Mini-ITX, already owned)
- **Estimation: 0 CZK** (existing hardware)
- **Features**: Supports mATX/mITX, ATX PSU, GPU up to 32cm, CPU cooler up to 16cm
- **Standard AM5 cooler** (included with Ryzen 5 7600)
- **Estimation: 0 CZK** (included with CPU)

### Power Supply
- **Existing ATX PSU** (already owned with case)
- **Estimation: 0 CZK** (existing hardware)
- **Note**: CHIEFTEC CI-02B-OP supports ATX PSU up to 16cm length

### Price-Optimized Total
| Component | Price (CZK) |
|-----------|-------------|
| CPU (Ryzen 5 7600) | 4,290 |
| Motherboard (ASRock B650M) | 3,200 |
| RAM (32GB DDR5) | 6,000 |
| SSD 500GB (ProxMox) | 1,500 |
| SSD 1TB (VMs) | 3,000 |
| Case (CHIEFTEC CI-02B-OP) | 0 |
| CPU Cooler (stock) | 0 |
| PSU (existing) | 0 |
| **Total** | **~17,990 CZK** |

**Performance**: 2-3 concurrent testing VMs + 1 Gitea server VM (6 cores with production workload)

---

## b) Claude's Recommended Configuration

### CPU
- **AMD Ryzen 7 7700** (Socket AM5, 8 cores, 16 threads, 3.8-5.3 GHz)
- **Estimation: 8,500 CZK** (Alza.cz)
- **Reasoning**: Excellent virtualization performance, 8 cores sufficient for 6-8 concurrent VMs

### Motherboard
- **ASUS ROG Strix B650I-E Gaming WiFi** (AM5, Mini-ITX)
- **Estimation: 7,500 CZK** (Alza.cz)
- **Features**: 2.5G Ethernet, WiFi 6E, PCIe 5.0, robust VRM, premium build quality

### Memory (RAM)
- **64GB DDR5-5600** (2x 32GB kit)
- **Estimation: 12,000 CZK** (Alza.cz)
- **Reasoning**: Future-proofing, allows 6-8 concurrent VMs with comfortable headroom

### Storage
- **Primary**: Samsung 980 PRO NVMe SSD 500GB
- **Estimation: 2,500 CZK** (Alza.cz)
- **VM Storage**: Samsung 980 PRO NVMe SSD 2TB
- **Estimation: 6,500 CZK** (Alza.cz)
- **Reasoning**: High IOPS for multiple concurrent VMs, enterprise reliability

### Case & Cooling
- **Fractal Design Node 202** (Mini-ITX)
- **Estimation: 3,000 CZK** (Alza.cz)
- **Noctua NH-L9a-AM5** (Premium low-profile cooler)
- **Estimation: 1,200 CZK** (Alza.cz)

### Power Supply
- **Corsair SF600** (600W SFX, 80+ Gold)
- **Estimation: 4,500 CZK** (Alza.cz)
- **Reasoning**: Modular cables, headroom for expansion, premium efficiency

### Claude's Recommended Total
| Component | Price (CZK) |
|-----------|-------------|
| CPU (Ryzen 7 7700) | 8,500 |
| Motherboard (ASUS B650I-E) | 7,500 |
| RAM (64GB DDR5) | 12,000 |
| SSD 500GB (ProxMox) | 2,500 |
| SSD 2TB (VMs) | 6,500 |
| Case (Node 202) | 3,000 |
| CPU Cooler (Noctua) | 1,200 |
| PSU (SF600) | 4,500 |
| **Total** | **~45,700 CZK** |

**Performance**: 6-8 concurrent VMs with premium reliability and future expansion capability

*Prices estimated based on Alza.cz pricing as of September 2025*

### Configuration Comparison
| Feature | Price-Optimized | Claude's Recommended |
|---------|----------------|---------------------|
| **Budget** | ~17,990 CZK | ~45,700 CZK |
| **VM Capacity** | 2-3 test VMs + Gitea | 4-6 test VMs + Gitea |
| **RAM** | 32GB | 64GB |
| **Storage** | 1.5TB total | 2.5TB total |
| **Reliability** | Good | Premium |
| **Future-proofing** | Basic | Excellent |

## Performance Expectations

### VM Capacity
- **Production VM**: 1x Gitea server (4GB RAM, 2 vCPU, 60GB disk)
- **Testing VMs**: 2-3 concurrent lightweight VMs (4GB RAM each)
- **Heavy Testing VMs**: 1-2 VMs with Docker/Podman (8GB RAM each)
- **Storage**: ~200GB per major OS template + Gitea data
- **Snapshots**: Multiple snapshots per VM for testing states
- **Reserved Resources**: ~4GB RAM + 2 vCPU for ProxMox and Gitea

### Use Cases
1. **Production Services**:
   - Gitea server VM: Primary development repository
   - Backup and monitoring services
   - Development database if needed

2. **Container Testing**:
   - Ubuntu 22.04 VM: Test Docker installation
   - Ubuntu 22.04 VM: Test Podman installation
   - Clean VM: Test no container runtime scenario
   - Combined VM: Test both Docker + Podman

3. **Cross-Platform Testing**:
   - Windows 11 VM: Windows Portunix testing (limited by resources)
   - Various Linux distributions
   - Different kernel versions

4. **CI/CD Integration**:
   - Jenkins/GitLab runner VMs
   - Automated test execution
   - Clean environment provisioning

## Network Configuration

### VLAN Setup (Optional)
- **Management VLAN**: ProxMox web interface access
- **Test VLAN**: VM network isolation
- **Production VLAN**: Access to development resources

### IP Addressing
- **ProxMox Host**: Static IP (e.g., 192.168.1.100)
- **VM Pool**: DHCP range (e.g., 192.168.1.150-200)
- **Test Isolation**: Separate subnet for container testing

## Storage Strategy

### ZFS Configuration
- **Boot Pool**: Single NVMe SSD (ProxMox OS)
- **VM Pool**: Single NVMe SSD (VM storage)
- **Backup**: External storage for VM backups

### Backup Strategy
- **Daily snapshots** of critical VMs
- **Weekly full backups** to external storage
- **Template preservation** for rapid VM deployment

## ProxMox Configuration

### Initial Setup
1. Install ProxMox VE 8.x
2. Configure network interfaces
3. Set up storage pools
4. Create VM templates for common OS types

### VM Templates
- **Ubuntu 22.04 LTS**: Primary Linux testing
- **Ubuntu 20.04 LTS**: Legacy compatibility
- **Debian 12**: Package manager testing
- **CentOS Stream 9**: RPM-based testing
- **Windows 11**: Windows Portunix testing

### Resource Allocation Guidelines
- **Development VMs**: 4GB RAM, 2 vCPU, 40GB disk
- **Container Testing VMs**: 8GB RAM, 4 vCPU, 60GB disk
- **Windows VMs**: 8GB RAM, 4 vCPU, 80GB disk

## Integration with Portunix Testing

### Test Automation
- **VM Provisioning**: Automated VM creation for tests
- **Clean State**: Snapshot-based reset between tests
- **Parallel Testing**: Multiple concurrent test environments
- **Result Aggregation**: Centralized test result collection

### API Integration
- Use ProxMox API for VM lifecycle management
- Integrate with Portunix CI/CD pipeline
- Automated test environment provisioning

## Security Considerations

### Access Control
- **Admin Access**: Restricted to development team
- **VM Access**: SSH key-based authentication
- **Network Security**: Firewall rules for test isolation

### Data Protection
- **VM Encryption**: Encrypt sensitive test data
- **Backup Encryption**: Encrypted backup storage
- **Network Isolation**: Separate test networks

## Maintenance

### Regular Tasks
- **Weekly**: Check VM resource usage
- **Monthly**: Update ProxMox VE
- **Quarterly**: Review and clean old snapshots
- **As needed**: Expand storage capacity

### Monitoring
- **Resource Monitoring**: CPU, RAM, storage usage
- **VM Health**: Automated health checks
- **Network Performance**: Bandwidth and latency monitoring

## Future Expansion

### Scalability
- **Additional Storage**: NAS integration for backup
- **Network Upgrade**: 10GbE for high-throughput testing
- **Cluster Setup**: Multiple ProxMox nodes for HA

### Enhanced Testing
- **GPU Passthrough**: For specialized testing
- **Container Orchestration**: Kubernetes testing environment
- **Cloud Integration**: Hybrid cloud testing scenarios

---

**Document Version**: 1.0
**Created**: September 16, 2025
**Author**: Zdeněk
**Purpose**: Portunix Testing Infrastructure Planning