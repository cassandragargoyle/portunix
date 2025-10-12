# Portunix

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org)
[![Coverage](https://img.shields.io/badge/Coverage-0%25-red.svg)](https://codecov.io/gh/cassandragargoyle/Portunix)
[![Build Status](https://github.com/cassandragargoyle/Portunix/workflows/Test%20Suite/badge.svg)](https://github.com/cassandragargoyle/Portunix/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cassandragargoyle/Portunix)](https://goreportcard.com/report/github.com/cassandragargoyle/Portunix)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Universal development environment management tool with intelligent OS detection, Docker container support, and automated software installation.

## 🚀 Features

### Core Capabilities
- **🔧 Universal Installation System**: Install development tools across Windows, Linux, and macOS
- **🐳 Docker Management**: Complete Docker container lifecycle management with SSH support
- **🔐 Certificate Management**: Automatic CA certificate installation and HTTPS connectivity verification
- **🪟 Windows Sandbox Integration**: Isolated development environments on Windows
- **🧠 Intelligent OS Detection**: Automatic platform detection and optimization with certificate bundle detection
- **⚡ Cross-Platform Support**: Native support for Windows, Linux, and macOS
- **💻 VM Management**: QEMU/KVM virtualization with Windows 11 support and snapshots

### Docker Management (Issue #2)
- **Intelligent Docker Installation**: OS-specific Docker installation with storage optimization
- **Multi-Platform Containers**: Support for Ubuntu, Alpine, CentOS, Debian, and custom images
- **SSH-Enabled Containers**: Automatic SSH server setup with generated credentials
- **Package Manager Detection**: Automatic detection of apt-get, yum, dnf, apk
- **Cache Directory Mounting**: Persistent storage for downloads and packages
- **Flexible Base Images**: Choose from various Linux distributions

### Certificate Management (Issue #30) 
- **Automatic CA Certificate Setup**: Installs CA certificates in containers before software installation
- **HTTPS Connectivity Verification**: Tests HTTPS connections after certificate installation
- **Multi-Distribution Support**: Works with apt, yum, dnf, apk, pacman, zypper package managers  
- **System Certificate Detection**: Shows certificate bundle status in system information
- **Standalone Certificate Installation**: `portunix install ca-certificates` command

### Installation Types
- **`default`**: Python + Java + VSCode (recommended)
- **`empty`**: Clean environment without packages
- **`python`**: Python development environment
- **`java`**: Java development environment  
- **`vscode`**: Visual Studio Code setup

## 📦 Installation

### Prerequisites
- **Go 1.21+** for building from source
- **Docker** (optional, for container features)
- **Windows 10/11 Pro/Enterprise** (for Windows Sandbox features)

### Quick Install
```bash
# Clone repository
git clone https://github.com/cassandragargoyle/Portunix.git
cd Portunix

# Build
make build

# Or build manually
go build -o portunix .
```

## 🎯 Quick Start

### Basic Usage
```bash
# Show help
./portunix --help

# Install Docker with intelligent OS detection
./portunix install docker

# Install Docker with auto-accept recommended storage
./portunix install docker -y

# Install other software
./portunix install python java vscode

# Install CA certificates for HTTPS connectivity
./portunix install ca-certificates

# Show system information with certificate status
./portunix system info
```

### Docker Container Management
```bash
# Run Python environment in Ubuntu container
./portunix docker run-in-container python

# Run Java environment in Alpine container  
./portunix docker run-in-container java --image alpine:3.18

# Run development environment with custom settings
./portunix docker run-in-container default \
  --image ubuntu:20.04 \
  --name my-dev-env \
  --port 8080:8080 \
  --keep-running

# Container management
./portunix docker list
./portunix docker logs <container-id>
./portunix docker stop <container-id>
./portunix docker remove <container-id>
```

### Windows Sandbox
```bash
# Run in Windows Sandbox with SSH
./portunix sandbox run-in-sandbox python

# Generate custom sandbox configuration
./portunix sandbox generate --enable-ssh
```

### Virtual Machines (QEMU/KVM & VirtualBox)
```bash
# Install QEMU/KVM virtualization stack
./portunix vm install-qemu

# Check virtualization support
./portunix vm check

# Method 1: Using new VM commands (recommended for QEMU)
./portunix vm create win11-vm \
  --iso ~/Downloads/Win11.iso \
  --disk-size 80G \
  --ram 8G \
  --cpus 4 \
  --os windows11

# Method 2: Using create vm command (supports both QEMU and VirtualBox)
./portunix create vm \
  --vmtype qemu \
  --vmname ubuntu-dev \
  --iso ~/Downloads/ubuntu-22.04.iso \
  --basefolder ~/VMs

# VirtualBox VM (requires VBox preprocessor)
./portunix create vm \
  --vmtype vbox \
  --vmname win10-dev \
  --iso ~/Downloads/Win10.iso \
  --basefolder ~/VMs

# VM lifecycle management
./portunix vm start win11-vm
./portunix vm list --all
./portunix vm info win11-vm
./portunix vm console win11-vm
./portunix vm stop win11-vm

# Snapshot management for trial software testing
./portunix vm snapshot create win11-vm clean-install \
  --description "Fresh Windows 11 after updates"
./portunix vm snapshot list win11-vm
./portunix vm snapshot revert win11-vm clean-install
./portunix vm snapshot delete win11-vm old-snapshot
```

## 💻 VM Management (QEMU/KVM)

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

### VM Workflow

#### QEMU/KVM (Linux)
1. **Prerequisites Check**: Verify hardware virtualization support (Intel VT-x/AMD-V)
2. **Installation**: `portunix vm install-qemu` - Install QEMU, KVM, libvirt, and management tools
3. **VM Creation**: 
   - Method 1: `portunix vm create` (advanced options)
   - Method 2: `portunix create vm --vmtype qemu` (simple creation)
4. **Snapshot Management**: Create restore points for easy rollback
5. **Lifecycle Management**: Start, stop, monitor, and manage VMs

#### VirtualBox (Windows/macOS/Linux)
1. **Prerequisites**: VirtualBox installed and preprocessor configured
2. **VM Creation**: `portunix create vm --vmtype vbox`
3. **Management**: Through VirtualBox GUI or preprocessor commands

### Use Cases
- **Enterprise Software Trials**: Test software like Enterprise Architect repeatedly
- **Development Environments**: Isolated development environments
- **OS Testing**: Test different operating systems safely
- **Security Research**: Isolated environment for security testing
- **Cross-Platform Development**: Test applications on different OS versions

## 🐳 Docker Features

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

### Example Workflows

**Python Development:**
```bash
# Lightweight Alpine Python environment
./portunix docker run-in-container python --image alpine:3.18

# Full Ubuntu Python environment
./portunix docker run-in-container python --image ubuntu:22.04
```

**Java Development:**
```bash
# Java on CentOS (enterprise environment)
./portunix docker run-in-container java --image centos:8

# Java on Ubuntu (standard environment)
./portunix docker run-in-container java
```

**Full Development Environment:**
```bash
# Complete development setup
./portunix docker run-in-container default \
  --name full-dev \
  --port 3000:3000 \
  --port 8080:8080 \
  --env "NODE_ENV=development" \
  --keep-running
```

## 🛠️ Development and Testing

### Modern Testing Architecture
Portunix includes a comprehensive testing strategy with modern Go testing practices:

- **Unit Tests**: Fast, isolated tests with mocking
- **Integration Tests**: Real Docker container testing
- **CI/CD Pipeline**: Automated testing with GitHub Actions
- **Coverage Reporting**: Comprehensive coverage analysis
- **Quality Gates**: Linting, security scanning, cross-platform testing

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

### Testing Documentation
- **[Testing Guide](TEST_GUIDE.md)**: Comprehensive testing documentation
- **[Testing Architecture](TESTING.md)**: Modern testing strategy and standards

## 📚 Documentation

### User Guides
- **[TEST_GUIDE.md](TEST_GUIDE.md)**: Complete testing guide for developers
- **[TESTING.md](TESTING.md)**: Testing architecture and standards

### Issue Tracking
- **[Issues Documentation](docs/issues/README.md)**: GitHub issues mirror and tracking

### Examples
```bash
# View examples
ls examples/
cat examples/user-install-config.json
```

## 🔧 Configuration

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

## 🚀 CI/CD and Quality

### GitHub Actions Pipeline
- **Lint**: Code quality and formatting checks
- **Unit Tests**: Fast isolated testing
- **Integration Tests**: Real Docker container testing
- **Security**: Vulnerability scanning with gosec
- **Cross-Platform**: Windows, Linux, macOS testing
- **Coverage**: Automated coverage reporting

### Quality Metrics
- **Test Coverage**: Target ≥80% for unit tests
- **Code Quality**: golangci-lint with comprehensive rules
- **Security**: gosec security scanning
- **Performance**: Benchmark testing for Docker operations

## 🤝 Contributing

### Development Setup
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
1. Follow existing code conventions (see [TODO:](TODO:.md))
2. Write tests for new features (see [TEST_GUIDE.md](TEST_GUIDE.md))
3. Update documentation
4. Run quality checks: `make lint`
5. Ensure all tests pass: `make test`

### Pull Request Process
1. Create feature branch: `git checkout -b feature/my-feature`
2. Implement changes with tests
3. Run quality checks: `make ci-test`
4. Submit pull request
5. Automated CI/CD pipeline will validate changes

## 📈 Roadmap

### Current Status ✅
- ✅ Cross-platform OS detection system
- ✅ Docker management with intelligent installation
- ✅ Multi-platform container support
- ✅ SSH-enabled development containers
- ✅ Comprehensive testing architecture
- ✅ CI/CD pipeline with quality gates

### Upcoming Features 🚧
- [ ] Container orchestration with docker-compose
- [ ] VSCode development containers integration
- [ ] Package manager plugins
- [ ] Cloud deployment automation
- [ ] Advanced security scanning

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🔗 Links

- **GitHub**: [cassandragargoyle/Portunix](https://github.com/cassandragargoyle/Portunix)
- **Issues**: [GitHub Issues](https://github.com/cassandragargoyle/Portunix/issues)
- **Documentation**: [docs/](docs/)

## 🤝 External Partnerships & Collaboration

### act - GitHub Actions Local Runner
Portunix integrates with and contributes to the **[nektos/act](https://github.com/nektos/act)** project for local GitHub Actions testing.

- **Project**: [nektos/act](https://github.com/nektos/act) - Run GitHub Actions locally 🚀
- **Website**: [nektosact.com](https://nektosact.com/)
- **Partnership**: CassandraGargoyle team collaborates on act development and integration
- **Integration**: Portunix provides seamless act installation and GitHub Actions workflow testing capabilities

The collaboration enhances both projects by providing better local development workflows and testing capabilities for GitHub Actions users.

---

**🎯 Universal development environment management made simple.**

## Command Line Reference

Usage:
  portunix [command]

Available Commands:
  choco       Chocolatey package manager operations (Windows only)
  completion  Generate the autocompletion script for the specified shell
  create      Creates a new resource (including VMs with --vmtype)
  docker      Manages Docker containers and Docker installation.
  help        Help about any command
  install     Installs specified software.
  sandbox     Manages Windows Sandbox instances.
  unzip       Extracts a ZIP file.
  vm          Manage virtual machines with QEMU/KVM (advanced)
  winget      Windows Package Manager operations and information
  wizard      Starts an interactive wizard.

Flags:
  -h, --help      help for portunix
  -v, --version   version for portunix

Use "portunix [command] --help" for more information about a command.