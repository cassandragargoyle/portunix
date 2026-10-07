# Issue #3: Podman Management Command - Similar to Docker Command

**Status:** ✅ Implemented  
**Priority:** High  
**Labels:** enhancement, podman, cross-platform  
**Milestone:** v1.2.0  

## Feature Request: Podman Management Command

### Overview

Add a `podman` command similar to the existing `docker` command to manage Podman containers. This should provide comprehensive Podman container management capabilities for Portunix, allowing users to run Portunix installations and commands inside Podman containers across Windows and Linux.

**This issue extends `portunix install podman`** to intelligently detect the operating system and install the appropriate Podman variant:

- **Windows**: Install Podman Desktop for Windows
- **Linux**: Install Podman (podman package based on distribution)
- **macOS**: Install Podman via Homebrew (future support)

### Current Podman Support Status

Currently Portunix has:

- ❌ No Podman support
- ✅ Docker implementation as reference (`app/docker/`, `cmd/docker_*.go`)

### Required Podman Command Structure

```text
portunix podman
├── run-in-container [installation-type]    # Similar to docker run-in-container
├── build                                   # Build Portunix Podman image
├── start                                   # Start existing container
├── stop                                    # Stop running container
├── list                                    # List Portunix containers
├── remove                                  # Remove containers
├── logs                                    # View container logs
└── exec                                    # Execute commands in container
```

**Enhanced Installation Command:**

```bash
portunix install podman    # Intelligent OS-based Podman installation
```

### Key Differences from Docker

**Podman Advantages:**

- **Rootless containers** - Run containers without root privileges
- **Daemonless** - No background daemon required
- **OCI compliance** - Full compatibility with Docker images
- **Pod support** - Native Kubernetes-style pod management
- **Security focused** - Better default security posture

**Podman-Specific Features:**

- `podman pod create/start/stop` - Pod management
- `podman generate systemd` - SystemD service generation
- `podman machine` - VM management for rootless containers
- Rootless operation by default

### Detailed Requirements

#### 1. Core Commands

**`portunix podman run-in-container [installation-type] [--image base-image]`**

- **Identical functionality to Docker version** but using Podman
- Support same base images: `ubuntu:22.04`, `alpine:3.18`, `debian:bullseye`, etc.
- Rootless container execution by default
- Automatic SSH server setup (same as Docker implementation)
- Same volume mounting and cache directory support
- Same installation types: `default`, `empty`, `python`, `java`, `vscode`

**`portunix podman build`**

- Build optimized Portunix Podman images
- Use `podman build` instead of `docker build`
- Support same base images as Docker version
- Rootless image building

**`portunix install podman [-y]`**

- **Linux Installation**:
  - **Ubuntu/Debian**: Install `podman` via apt
  - **CentOS/RHEL/Rocky/Fedora**: Install `podman` via yum/dnf
  - **Alpine**: Install `podman` via apk
- **Windows Installation**:
  - Download and install Podman Desktop for Windows
  - Configure WSL2 backend if needed
- **Storage optimization** same as Docker (optimal drive/partition selection)

#### 2. Implementation Structure

**Files to Create (based on Docker structure):**

```text
cmd/
├── podman.go                    # Main podman command
├── podman_install.go           # Install podman command
├── podman_management.go        # Management subcommands
└── podman_run_in_container.go  # Run-in-container command

app/
└── podman/
    ├── podman.go               # Core Podman functionality
    └── podman_test.go          # Tests
```

**Key Functions to Implement:**

```go
// app/podman/podman.go
func InstallPodman(autoAccept bool) error
func CheckPodmanAvailableWithInstall(autoInstall bool) error
func RunInContainer(config PodmanConfig) error
func BuildImage(baseImage string) error
func ListContainers() ([]ContainerInfo, error)
// ... other functions similar to Docker
```

#### 3. Configuration Structure

**PodmanConfig (based on DockerConfig):**

```go
type PodmanConfig struct {
    Image             string
    ContainerName     string
    Ports             []string
    Volumes           []string
    Environment       []string
    Command           []string
    EnableSSH         bool
    KeepRunning       bool
    Disposable        bool
    Privileged        bool
    Network           string
    CacheShared       bool
    CachePath         string
    InstallationType  string
    DryRun            bool
    AutoInstallPodman bool
    Rootless          bool  // Podman-specific
    Pod               string // Podman-specific
}
```

#### 4. Command Examples

**Podman Installation:**

```bash
# Interactive installation
portunix install podman

# Automated installation
portunix install podman -y

# Expected output similar to Docker but for Podman
```

**Container Management:**

```bash
# Run development container (rootless by default)
portunix podman run-in-container python

# Run with specific image
portunix podman run-in-container java --image ubuntu:22.04

# List containers
portunix podman list

# Stop/remove containers
portunix podman stop abc123
portunix podman remove abc123
```

**Podman-Specific Features:**

```bash
# Run in privileged mode (if needed)
portunix podman run-in-container default --privileged

# Run in specific pod
portunix podman run-in-container python --pod my-pod
```

#### 5. Expected Output

**Installation Output:**

```bash
$ portunix install podman
Detecting operating system...
✓ Detected: Ubuntu 22.04 LTS
✓ Package manager: apt

Installing Podman...
🔧 Adding repositories...
📦 Installing podman package...
✓ Podman installed successfully
✓ Configuring rootless containers...
✓ Setting up user namespaces...

Verification:
✓ podman --version: podman version 4.3.1
✓ Rootless mode: enabled
✓ Container storage: /home/user/.local/share/containers

Podman is ready to use!
Note: Podman runs rootless by default for better security.
```

**Container Startup:**

```bash
$ portunix podman run-in-container python
✓ Podman is available (rootless mode)
Pulling base image: ubuntu:22.04...
✓ Image pulled successfully
✓ Detected package manager: apt-get
✓ Creating rootless container: portunix-python-abc123

Setting up container environment...
✓ Installing Portunix binary
✓ Mounting cache directory: .cache:/portunix-cache
✓ Setting up Python via apt-get...
✓ Installing OpenSSH server...
✓ Generating SSH credentials...
✓ Starting SSH daemon...
✓ Container startup completed

📡 SSH CONNECTION INFORMATION:
════════════════════════════════════════════════════════
🔗 Container IP:   10.88.0.2
📄 SSH Port:      localhost:2222
👤 Username:      portunix_user_abc123
🔐 Password:      Kx9mP2nQ8vR5wL7s
📄 SSH Command:   ssh portunix_user_abc123@localhost -p 2222

💡 PODMAN FEATURES:
   • Running in rootless mode (enhanced security)
   • No daemon required
   • OCI-compatible with Docker images
════════════════════════════════════════════════════════

Container is running and ready for SSH connections!
```

#### 6. Implementation Steps

1. **Create issue documentation** ✅
2. **Copy Docker implementation structure**
   - Copy `app/docker/` → `app/podman/`
   - Copy `cmd/docker*.go` → `cmd/podman*.go`
3. **Adapt for Podman**
   - Replace `docker` commands with `podman`
   - Add rootless configuration
   - Add pod support
4. **Update main command registration**
5. **Add tests**
6. **Update documentation**

#### 7. Benefits Over Docker

**Security:**

- Rootless containers by default
- No privileged daemon
- Better isolation

**Simplicity:**

- No daemon to manage
- Direct container execution
- Lighter resource usage

**Compatibility:**

- Full Docker image compatibility
- Same CLI patterns as Docker
- Drop-in replacement capability

### Implementation Priority

1. **Phase 1:** Basic installation and container management
2. **Phase 2:** Installation types integration (reuse Docker logic)
3. **Phase 3:** Podman-specific features (rootless, pods)
4. **Phase 4:** Advanced security and performance features

### Integration with Existing Codebase

- Reuse existing Docker installation templates
- Share common container logic where possible
- Maintain consistent CLI patterns with Docker commands
- Support both Docker and Podman simultaneously

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented:** 2025-01-09  
**Version:** v1.5.7+  

### Implementation Summary

All requested Podman management features have been successfully implemented:

#### ✅ Core Podman Management Commands

- **`portunix podman install`** - Intelligent OS-specific Podman CLI and Desktop installation
- **`portunix podman run-in-container`** - Container-based Portunix installations with rootless support
- **`portunix podman build/start/stop/list/remove/logs/exec`** - Complete lifecycle management
- **`portunix podman check-requirements`** - System requirements validation
- **`portunix podman desktop`** - Podman Desktop GUI installation

#### ✅ Podman-Specific Features

- **Rootless containers by default** - Enhanced security with non-root execution
- **Daemonless operation** - No background daemon required, unlike Docker
- **Pod support** - Kubernetes-style container grouping capabilities
- **OCI-compatible** - Full compatibility with Docker images and registries
- **Enhanced security model** - SELinux integration and user namespace isolation

#### ✅ Shared Container Features

- **SSH-enabled containers** with automatic server setup
- **Multi-platform support** - Ubuntu, Alpine, CentOS, Debian images  
- **Package manager detection** with certificate management integration
- **Installation type profiles** (default, python, java, vscode, empty)
- **Cache directory mounting** for persistent storage

#### ✅ Cross-Platform Installation

- **Linux:** Podman CLI installation via distribution package managers
- **Windows/macOS:** Podman Desktop installation with full GUI support
- **System requirements checking** with automatic dependency resolution

### Verification Commands

```bash
# Test Podman management commands
portunix podman --help
portunix podman check-requirements
portunix podman install
portunix podman run-in-container python --rootless
portunix podman list
portunix podman desktop

# Test rootless containers
portunix podman run-in-container java --image alpine:3.18
```

### Resolution Status

**FULLY IMPLEMENTED** ✅ - Complete Podman management command suite with rootless security, daemonless operation, pod support, and full Docker image compatibility. Includes both CLI and Desktop installation with cross-platform support.

---
**Created:** 2025-08-19  
**Last Updated:** 2025-01-09  
**Assigned:** Development Team  
**Related Issues:** [#2](002-docker-management-command.md) (Docker Management Command), [#30](216-container-tls-certificate-verification-failure.md) (Certificate Management), [#28](214-universal-container-parameters-support.md) (Universal Container Parameters)