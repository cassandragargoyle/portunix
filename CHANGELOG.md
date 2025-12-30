# Changelog

All notable changes to Portunix will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.6] - 2025-12-02

### Added
- **PTX-Make Helper** - New helper binary for cross-platform Makefile utilities (Issue #102)
  - File operations: `copy`, `mkdir`, `rm`, `exists`
  - Build metadata: `version`, `commit`, `timestamp`
  - Utilities: `checksum`, `chmod`, `json`, `env`
  - Dispatcher integration via `portunix make <command>`
  - `chmod` is no-op on Windows for portability

## [1.7.5] - 2025-12-01

### Added
- **PTX-AIOps Helper** - AI Operations helper for GPU management and Ollama integration (Issue #101)
  - GPU status monitoring with NVIDIA support
  - Ollama container management
  - Model installation and management
  - Open WebUI deployment

## [1.7.4] - 2025-11-30

### Added
- PTX-Virt helper binary for virtualization management
- PTX-Prompting helper for template-based prompt generation

## [1.7.3] - 2025-11-28

### Added
- Clipboard support for interactive prompting

## [1.7.2] - 2025-11-25

### Fixed
- Version embedding in build process

## [1.7.1] - 2025-11-24

### Fixed
- Build script version updates

## [1.7.0] - 2025-11-20

### Added
- PTX-Installer Helper for package management (Issue #100)
- Package Registry Architecture with AI integration (Issue #082)
- Hugo installation support (Issue #075)
- Container list command (Issue #084)

### Fixed
- Hugo installation permission issues (Issue #085)
- Container exec command malfunction (Issue #095)
- Container rm subcommand recognition (Issue #094)

## [1.6.4] - 2025-11-15

### Added
- Ansible Infrastructure as Code integration (Issue #056)
- VirtualBox/KVM conflict detection (Issue #088)
- QEMU/KVM adapter for virt check (Issue #089)
- Libvirt daemon detection and auto-fix (Issue #090, #091)

### Fixed
- Virtual machine snapshot list empty names (Issue #061)
- VS Code installation filename resolution (Issue #064)

## [1.6.3] - 2025-11-10

### Added
- GitHub CLI installation support (Issue #078)
- Ansible Galaxy Collections support (Issue #063)
- Universal virtualization support (Issue #049)

## [1.6.0] - 2025-11-01

### Added
- Multi-level help system (Issue #050)
- Git-like dispatcher architecture (Issue #051)
- Container runtime capability detection (Issue #039)
- Node.js/npm installation support (Issue #041)

### Fixed
- Container run command flag parsing (Issue #038)
- Module path naming inconsistencies (Issue #053)
