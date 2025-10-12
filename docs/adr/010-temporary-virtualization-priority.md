# ADR-010: Temporary Virtualization Priority

## Status
Active

## Context
The Portunix project has planned infrastructure based on a ProxMox VE server for development and testing (see `/docs/infrastructure/001-proxmox-server-specification.md`). However, this server has not yet been acquired, creating a gap in our virtualization capabilities for development and testing.

## Decision
Until the ProxMox server is acquired, QEMU and potentially VirtualBox will be used as temporary virtualization solutions for Portunix development. Therefore, tasks related to QEMU and VirtualBox support in the Portunix project must be given higher priority.

## Consequences

### Positive
- Development can continue without waiting for hardware acquisition
- QEMU/VirtualBox support will benefit users who don't have ProxMox infrastructure
- These features will remain useful even after ProxMox server acquisition
- Broader virtualization platform support increases Portunix flexibility

### Negative
- Additional development effort for temporary solutions
- Potential technical debt if implementations need refactoring for ProxMox
- Less optimized performance compared to dedicated ProxMox server
- More complex testing matrix with multiple virtualization platforms

## Implementation Priority
1. **QEMU support** - Primary focus due to cross-platform availability
2. **VirtualBox support** - Secondary option for broader compatibility
3. **VM management commands** - Unified interface for both platforms
4. **Testing infrastructure** - Adapt current tests to work with QEMU/VirtualBox

## Transition Plan
Once ProxMox server is acquired:
1. Maintain QEMU/VirtualBox support for user environments
2. Migrate development testing to ProxMox
3. Use ProxMox for CI/CD pipeline
4. Keep QEMU/VirtualBox as fallback options

## References
- `/docs/infrastructure/001-proxmox-server-specification.md`
- Issue tracking for VM support features

---

**Document Version**: 1.0
**Created**: 2025-01-16
**Author**: Zdenek
**Purpose**: Define temporary virtualization strategy priority