# ADR-002: PowerShell Linux Installation Architecture

**Date:** 2025-08-23  
**Status:** Accepted  
**Architect:** ZK  
**Related Issue:** [#011](../issues/012-powershell-linux-installation.md)  

## Context

IT support workers frequently need to execute PowerShell scripts across diverse Linux environments, creating operational complexity due to distribution variety and manual installation requirements. Current Portunix package management capabilities need extension to support PowerShell installation across multiple Linux distributions.

The request originated from support teams requiring consistent PowerShell availability across Ubuntu, Kubuntu, Fedora, Debian, Mint, Elementary OS, and Rocky Linux distributions.

## Decision

### Core Implementation Strategy
**Decision:** Implement PowerShell Linux installation functionality entirely within Portunix core engine, leveraging existing package management infrastructure.

**Rationale:** 
- Maintains consistency with existing Portunix architecture
- Ensures tight integration with OS detection and package management systems
- Provides unified interface for all package operations

### Target Platform Coverage
**Decision:** Support Top 10 Linux distributions for 2025:
1. Ubuntu (22.04, 24.04)
2. Debian (11, 12)
3. Fedora (39, 40, 41)
4. Rocky Linux (8, 9)
5. Mint (21, 22)
6. Elementary OS (7, 8)
7. Kubuntu (22.04, 24.04)
8. openSUSE (Leap 15.5, Tumbleweed)
9. Arch Linux
10. CentOS Stream (9)

**Rationale:**
- Covers 90%+ of enterprise Linux deployments
- Balances maintenance overhead with market coverage
- Aligns with Microsoft PowerShell official support matrix

### Extended Virtualization Support
**Decision:** Implement automatic installation of supported Linux distributions into virtual environments:

#### VirtualBox Integration
- Automatic VM creation with supported Linux distributions
- Dual template approach:
  - Templates with PowerShell pre-installed for immediate use
  - Clean templates without PowerShell for testing Portunix installation process
- Template-based approach for rapid environment provisioning

#### Container Integration (Podman)
- Dual container image approach:
  - Images with PowerShell pre-installed for immediate deployment
  - Base images without PowerShell for testing Portunix installation functionality
- Multi-architecture support (x64, ARM64)
- Integration with existing Portunix container management

**Rationale:**
- Enables rapid testing environment creation
- Supports development and CI/CD workflows
- Reduces environment setup time for support teams

## Technical Implementation

### Installation Flow Sequence

```
User Request → Distribution Detection → Repository Setup → GPG Key Import → Package Installation → Verification → Success
     ↓
VM/Container Request → Base Image Selection → PowerShell Integration → Image Creation → Testing → Storage
```

### Package Definition Integration
Extend `assets/install-packages.json` with comprehensive PowerShell definitions supporting:
- Multiple installation methods per distribution
- Version-specific repositories
- Fallback mechanisms (Snap, manual installation)
- Container and VM integration metadata

## Consequences

### Positive Consequences
- **Unified Experience**: Single command installs PowerShell across all supported Linux distributions
- **Rapid Environment Creation**: Automated VM/container creation reduces setup time from hours to minutes
- **Enterprise Ready**: Supports top enterprise Linux distributions with consistent tooling
- **CI/CD Integration**: Container support enables PowerShell in automated workflows
- **Scalability**: Architecture supports future distribution additions

### Negative Consequences
- **Increased Complexity**: VM and container integration adds significant codebase complexity
- **Resource Requirements**: VirtualBox integration requires additional system resources
- **Maintenance Overhead**: Supporting 10+ distributions increases testing and maintenance burden
- **Dependency Growth**: Additional dependencies on VirtualBox and advanced Podman features

### Mitigation Strategies
- **Modular Design**: VM and container features as optional modules
- **Automated Testing**: CI pipeline testing across all supported distributions
- **Documentation**: Comprehensive troubleshooting guides for each distribution
- **Fallback Mechanisms**: Multiple installation methods per distribution

## Implementation Phases

### Phase 1: Core PowerShell Installation (Weeks 1-2)
- Ubuntu, Debian, Fedora, Rocky Linux support
- Basic package management integration
- Installation verification

### Phase 2: Extended Distribution Support (Weeks 3-4)
- Remaining 6 distributions
- Snap package fallback implementation
- Error handling and recovery

### Phase 3: Virtualization Integration (Weeks 5-8)
- VirtualBox template creation
- Automated VM provisioning
- Container image building with Podman

### Phase 4: Testing & Hardening (Weeks 9-10)
- Cross-distribution testing automation
- Performance optimization
- Documentation and user guides

## Risk Assessment

**High Risk:**
- Repository changes by Microsoft affecting installation methods
- Distribution-specific packaging conflicts

**Medium Risk:**
- VirtualBox API compatibility across versions
- Container registry limitations

**Low Risk:**
- Performance impact on existing Portunix operations
- User adoption of virtualization features

## Success Metrics
- PowerShell installation success rate >95% across all distributions
- VM creation time <5 minutes for standard configurations
- Container build time <2 minutes for PowerShell-enabled images
- User adoption rate >60% within 6 months

---
*Architect: ZK*  
*Review Status: Draft - Pending Additional Input*