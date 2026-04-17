# Portunix

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![Coverage](https://img.shields.io/badge/Coverage-0%25-red.svg)](https://codecov.io/gh/cassandragargoyle/Portunix)
[![Build Status](https://github.com/cassandragargoyle/Portunix/workflows/Test%20Suite/badge.svg)](https://github.com/cassandragargoyle/Portunix/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cassandragargoyle/Portunix)](https://goreportcard.com/report/github.com/cassandragargoyle/Portunix)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 🌐 **Language / Jazyk**: **English** | [Čeština](README.cs.md)

**Unified AI plugin and task platform for development environments** — with intelligent OS detection, Docker container support, and automated software installation.

> Simplify your development workflow across Windows, Linux, and macOS with a single, powerful CLI tool.

## Why Portunix?

- **One Tool, All Platforms**: Consistent experience across Windows, Linux, and macOS
- **Smart Automation**: Intelligent OS detection and automatic configuration
- **Container-First**: First-class Docker and Podman support with SSH-enabled containers
- **Developer-Focused**: Built by developers, for developers

## Quick Install

### From Releases (Recommended)

```bash
# Linux (amd64)
wget https://github.com/cassandragargoyle/Portunix/releases/latest/download/portunix_linux_amd64.tar.gz
tar -xzf portunix_linux_amd64.tar.gz
sudo mv portunix /usr/local/bin/

# Verify installation
portunix version
```

### From Source

```bash
git clone https://github.com/cassandragargoyle/Portunix.git
cd Portunix
make build
```

## Features

### Core Capabilities

- **Universal Installation System**: Install development tools across Windows, Linux, and macOS
- **Docker Management**: Complete Docker container lifecycle management with SSH support
- **Certificate Management**: Automatic CA certificate installation and HTTPS connectivity verification
- **Windows Sandbox Integration**: Isolated development environments on Windows
- **Intelligent OS Detection**: Automatic platform detection and optimization with certificate bundle detection
- **Cross-Platform Support**: Native support for Windows, Linux, and macOS
- **VM Management**: QEMU/KVM virtualization with Windows 11 support and snapshots

### Docker Management

- **Intelligent Docker Installation**: OS-specific Docker installation with storage optimization
- **Multi-Platform Containers**: Support for Ubuntu, Alpine, CentOS, Debian, and custom images
- **SSH-Enabled Containers**: Automatic SSH server setup with generated credentials
- **Package Manager Detection**: Automatic detection of apt-get, yum, dnf, apk
- **Cache Directory Mounting**: Persistent storage for downloads and packages
- **Flexible Base Images**: Choose from various Linux distributions

### Certificate Management

- **Automatic CA Certificate Setup**: Installs CA certificates in containers before software installation
- **HTTPS Connectivity Verification**: Tests HTTPS connections after certificate installation
- **Multi-Distribution Support**: Works with apt, yum, dnf, apk, pacman, zypper package managers
- **System Certificate Detection**: Shows certificate bundle status in system information
- **Standalone Certificate Installation**: `portunix install ca-certificates` command

### Documentation Environment

- **Playbook Template System**: Generate documentation sites with a single command
- **Supported Engines**: Docusaurus, Hugo, Docsy (Hugo + Google Docsy theme), Docsify
- **Container-Based**: Run documentation environments in Docker/Podman containers
- **Shared Folder Workflow**: Edit locally, see changes instantly via live-reload
- **Auto-Dependency Resolution**: `portunix install docusaurus` auto-installs Node.js
- **Quickstart Script**: One-liner PowerShell script for Windows users

### Plugin System

- **gRPC-Based Architecture**: High-performance plugin communication
- **Plugin Lifecycle**: Install, enable, disable, start, stop, uninstall
- **Plugin Creation**: Generate new plugin templates with `portunix plugin create`
- **Plugin Registry**: Centralized plugin discovery and management

### Self-Update System

- **Automatic Updates**: `portunix update` checks and installs latest version
- **SHA256 Verification**: Secure checksum verification of downloads
- **Backup & Rollback**: Automatic backup before update with rollback on failure

### Installation Types

- **`default`**: Python + Java + VSCode (recommended)
- **`empty`**: Clean environment without packages
- **`python`**: Python development environment
- **`java`**: Java development environment
- **`vscode`**: Visual Studio Code setup

## Quick Start

### Basic Usage

```bash
# Show help
portunix --help

# Install Docker with intelligent OS detection
portunix install docker

# Install Docker with auto-accept recommended storage
portunix install docker -y

# Install other software
portunix install python java vscode

# Install CA certificates for HTTPS connectivity
portunix install ca-certificates

# Show system information with certificate status
portunix system info
```

### Docker Container Management

```bash
# Run Python environment in Ubuntu container
portunix docker run-in-container python

# Run Java environment in Alpine container
portunix docker run-in-container java --image alpine:3.18

# Run development environment with custom settings
portunix docker run-in-container default \
  --image ubuntu:20.04 \
  --name my-dev-env \
  --port 8080:8080 \
  --keep-running

# Container management
portunix docker list
portunix docker logs <container-id>
portunix docker stop <container-id>
portunix docker remove <container-id>
```

### Documentation Site Setup

```bash
# Create a Docusaurus documentation site in a container
portunix playbook init my-docs --template static-docs --engine docusaurus --target container
portunix playbook run my-docs.ptxbook --script create
portunix playbook run my-docs.ptxbook --script dev
# -> Open http://localhost:3000

# Or use Hugo / Docsy / Docsify
portunix playbook init my-docs --template static-docs --engine hugo --target container

# Direct install (auto-installs dependencies)
portunix install docusaurus
portunix install hugo
```

### Windows Sandbox

```bash
# Run in Windows Sandbox with SSH
portunix sandbox run-in-sandbox python

# Generate custom sandbox configuration
portunix sandbox generate --enable-ssh
```

### Virtual Machines (QEMU/KVM & VirtualBox)

```bash
# Install QEMU/KVM virtualization stack
portunix vm install-qemu

# Check virtualization support
portunix vm check

# Create Windows 11 VM
portunix vm create win11-vm \
  --iso ~/Downloads/Win11.iso \
  --disk-size 80G \
  --ram 8G \
  --cpus 4 \
  --os windows11

# VM lifecycle management
portunix vm start win11-vm
portunix vm list --all
portunix vm info win11-vm
portunix vm console win11-vm
portunix vm stop win11-vm

# Snapshot management for trial software testing
portunix vm snapshot create win11-vm clean-install \
  --description "Fresh Windows 11 after updates"
portunix vm snapshot list win11-vm
portunix vm snapshot revert win11-vm clean-install
```

## VM Management (QEMU/KVM)

### Core Features

- **Dual VM Backend Support**: Both QEMU/KVM and VirtualBox support
- **Cross-Platform Virtualization**: QEMU/KVM support for Linux hosts, VirtualBox for Windows/macOS
- **Windows 11 Support**: Full Windows 11 support with TPM 2.0 and Secure Boot
- **Snapshot Management**: Create, restore, and manage VM snapshots
- **Trial Software Testing**: Perfect for testing 30-day trial software
- **Multiple OS Support**: Windows, Linux, and custom OS installations
- **Resource Configuration**: Flexible CPU, RAM, and disk configuration
- **Unified Interface**: Two ways to create VMs - dedicated commands or unified create interface

### Supported Guest Operating Systems

- **Windows**: Windows 11, Windows 10, Windows Server 2022
- **Linux**: Ubuntu, Debian, CentOS, Fedora, Arch, and more
- **Custom**: Any OS that supports QEMU/KVM

## Docker Features

### Supported Base Images

- **Ubuntu**: `ubuntu:22.04`, `ubuntu:20.04` (default)
- **Alpine**: `alpine:3.18`, `alpine:latest` (lightweight)
- **Debian**: `debian:bullseye`, `debian:buster`
- **CentOS**: `centos:8`, `centos:7`
- **Fedora**: `fedora:38`, `fedora:37`
- **Rocky Linux**: `rockylinux:8`, `rockylinux:9`
- **Custom**: Any Docker image from registries

### Container Workflow

1. **Image Selection**: Choose base image or use default Ubuntu 22.04
2. **Package Manager Detection**: Automatically detect apt-get/yum/dnf/apk
3. **Container Creation**: Create with proper volume and port mappings
4. **SSH Setup**: Install OpenSSH server with generated credentials
5. **Software Installation**: Install requested packages using detected package manager
6. **Ready for Development**: SSH access with shared workspace and cache

## Development and Testing

### Quick Testing

```bash
# Setup development environment
make dev-setup

# Run all tests
make test

# Unit tests only (fast)
make test-unit

# Integration tests (requires Docker)
make test-integration

# Test coverage
make test-coverage

# Linting and quality
make lint
```

### Local Deployment

```bash
# Build and install to local system (auto-detects existing installation)
make deploy-local

# Remove from local system (auto-detects installation path)
make undeploy-local
```

## Documentation

- **[Windows Setup Guide](docs/WINDOWS-SETUP.md)**: Windows-specific setup and UTF-8 configuration
- **[TEST_GUIDE.md](TEST_GUIDE.md)**: Complete testing guide for developers
- **[TESTING.md](TESTING.md)**: Testing architecture and standards
- **[Issues Documentation](docs/issues/README.md)**: GitHub issues mirror and tracking

## Configuration

### Environment Variables

```bash
# Docker configuration
export DOCKER_HOST=unix:///var/run/docker.sock

# Development mode
export PORTUNIX_DEBUG=true

# Custom cache directory
export PORTUNIX_CACHE_DIR=/custom/cache/path
```

### Configuration Files

- **Install packages**: `assets/install-packages.json`
- **User config**: `examples/user-install-config.json`

## Roadmap

### Current Status

- Cross-platform OS detection system
- Docker/Podman management with intelligent installation
- Multi-platform container support with SSH-enabled development containers
- Container orchestration with docker-compose/podman-compose
- MCP server for AI assistant integration
- Package registry system with automatic discovery
- Multi-level help system (basic, expert, AI)
- QEMU/KVM virtualization with Windows 11 support and snapshots
- Product feedback tool (ptx-pft) with Fider/ClearFlask/Eververse providers
- AIOps helper for GPU/AI container workloads
- Make helper for cross-platform builds
- Ansible infrastructure as code integration
- Self-update system with rollback capability
- Comprehensive testing architecture and CI/CD pipeline
- Python development helper with project-local venv support
- Plugin system with gRPC architecture (#7)
- Playbook template system for documentation environments
- Auto-dependency resolution for package installation
- Docusaurus, Hugo, Docsy, Docsify support

### Upcoming Features

- Virtual development disk management (#8)
- Configurable datastore backends (#9)
- Interactive wizard framework (#14)
- AI assistant installation support (#35)
- VSCode development containers integration

## Contributing

```bash
# Clone and setup
git clone https://github.com/cassandragargoyle/Portunix.git
cd Portunix
make dev-setup

# Run tests
make test

# Check status
make status
```

### Guidelines

1. Follow existing code conventions
2. Write tests for new features
3. Update documentation
4. Run quality checks: `make lint`
5. Ensure all tests pass: `make test`

### Pull Request Process

1. Create feature branch: `git checkout -b feature/my-feature`
2. Implement changes with tests
3. Run quality checks: `make ci-test`
4. Submit pull request
5. Automated CI/CD pipeline will validate changes

## Creating a Release

```bash
# Create release (builds all platforms, generates notes, checksums)
python3 scripts/make-release.py v1.10.7

# Upload to GitHub
python3 scripts/upload-release-to-github.py v1.10.7
```

The release script automatically:

- Updates version in source files and `portunix.rc`
- Creates git tag
- Builds cross-platform binaries using GoReleaser
- Creates platform-specific archives (Linux, Windows, macOS)
- Generates release notes and checksums

## External Partnerships

### act - GitHub Actions Local Runner

Portunix integrates with and contributes to the **[nektos/act](https://github.com/nektos/act)** project for local GitHub Actions testing.

- **Project**: [nektos/act](https://github.com/nektos/act) - Run GitHub Actions locally
- **Website**: [nektosact.com](https://nektosact.com/)
- **Integration**: Portunix provides seamless act installation and GitHub Actions workflow testing capabilities

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- **GitHub**: [cassandragargoyle/Portunix](https://github.com/cassandragargoyle/Portunix)
- **Issues**: [GitHub Issues](https://github.com/cassandragargoyle/Portunix/issues)
- **Documentation**: [docs/](docs/)

---

**Unified AI plugin and task platform — orchestrate, automate, extend.**
