# Issue #2: Docker Management Command - Similar to Sandbox Command

**Status:** ✅ Implemented  
**Priority:** High  
**Labels:** enhancement, docker, cross-platform  
**Milestone:** v1.1.0  

## Feature Request: Docker Management Command

### Overview
Add a `docker` command similar to the existing `sandbox` command to manage Docker containers. This should provide comprehensive Docker container management capabilities for Portunix, allowing users to run Portunix installations and commands inside Docker containers across Windows and Linux.

**This issue also includes extending `portunix install docker`** to intelligently detect the operating system and install the appropriate Docker variant:
- **Windows**: Install Docker Desktop for Windows
- **Linux**: Install Docker Engine (docker.io or docker-ce based on distribution)
- **macOS**: Install Docker Desktop for macOS (future support)

### Current Docker Support
Currently Portunix has:
- ✅ Docker environment detection (`portunix system check docker`)
- ✅ Basic Dockerfile and build script (`docker.portunix.bat`)  
- ✅ PowerShell Docker detection functions (`Test-IsDocker`)

### Required Docker Command Structure
```
portunix docker
├── run-in-container [installation-type]    # Similar to sandbox run-in-sandbox
├── build                                   # Build Portunix Docker image
├── start                                   # Start existing container
├── stop                                    # Stop running container
├── list                                    # List Portunix containers
├── remove                                  # Remove containers
├── logs                                    # View container logs
└── exec                                    # Execute commands in container
```

**Enhanced Installation Command:**
```
portunix install docker    # Intelligent OS-based Docker installation
```

### Detailed Requirements

#### 1. Core Commands

**`portunix docker run-in-container [installation-type] [--image base-image]`**
- **Flexible Base Image Selection:** Unlike Windows Sandbox where the OS is fixed, Docker containers can run on different base images
- Support for multiple Linux distributions and versions:
  - `ubuntu:22.04`, `ubuntu:20.04`, `ubuntu:18.04`
  - `debian:bullseye`, `debian:buster`
  - `alpine:3.18`, `alpine:3.17`, `alpine:latest`
  - `centos:7`, `centos:8`
  - `fedora:38`, `fedora:37`
  - `rockylinux:8`, `rockylinux:9`
  - Custom images: `myregistry/custom-image:tag`
- Automatic adaptation of installation scripts based on detected package manager (apt, yum, apk, etc.)
- Image validation and automatic pulling if not present locally
- **Similar workflow to `portunix sandbox run-in-sandbox`:**
  - Automatic SSH server installation and configuration in container
  - Generate and display SSH connection credentials (username/password/IP/port)
  - Verify SSH connectivity before completing setup
  - Provide user with connection instructions and commands
- Supports same installation types: `default`, `empty`, `python`, `java`, `vscode`
- Cross-platform (Windows and Linux hosts)
- Automatic image building if not present
- Volume mounting for file sharing
- Port forwarding for SSH/HTTP services

**`portunix docker build`**
- Build optimized Portunix Docker images
- Support multiple base images: `ubuntu:22.04`, `alpine`, `debian`
- Include Portunix binary and common tools
- Automatic tagging with version

**`portunix docker start/stop/remove`**
- Manage existing Portunix containers
- Support container naming conventions
- Cleanup temporary containers

**`portunix install docker [-y]`** 
- **Intelligent OS Detection**: Use existing system detection to identify platform
- **Smart Storage Location Detection**:
  - Analyze available disk space on all drives/partitions
  - Identify drive with most available space for Docker data
  - **Windows**: Suggest optimal drive for Docker Desktop data-root (default: C:\)
  - **Linux**: Suggest optimal mount point for `/var/lib/docker` (default: root partition)
  - Present user with storage options and recommendations
  - **`-y` flag**: Auto-accept recommended storage location without prompting
- **Windows Installation**: 
  - Download and install Docker Desktop for Windows
  - Configure custom data-root location if user selects non-default drive
  - Handle Windows version requirements (Windows 10/11 Pro/Enterprise)
  - Enable WSL2 backend if available
  - Configure Windows features (Hyper-V or WSL2)
- **Linux Installation**:
  - **Ubuntu/Debian**: Install `docker.io` or `docker-ce` via apt
  - **CentOS/RHEL/Rocky**: Install `docker-ce` via yum/dnf  
  - **Alpine**: Install `docker` via apk
  - **Generic Linux**: Download Docker binaries directly
  - Configure custom docker data directory if user selects non-default location
  - Configure docker daemon and add user to docker group
  - Enable and start docker service
- **Prerequisites Check**: Verify system requirements before installation
- **Post-Installation Verification**: Test Docker installation with `docker --version`

#### 2. Container Features

**Volume Mounting:**
- Mount current directory to `/workspace` (Linux) or `C:\workspace` (Windows containers)
- **Cache Directory Integration:**
  - Mount host `.cache` directory to container `/portunix-cache`
  - Share downloaded packages, tools, and artifacts between host and container
  - Persistent storage for frequently used downloads (Python packages, Java JDK, VSCode extensions)
  - Avoid re-downloading same packages in multiple container instances
  - Cache structure preserved: `.cache/python-embeddable/`, `.cache/openssh/`, `.cache/notepadplusplus/`
- Support custom volume mappings

**Network Configuration:**
- Automatic port forwarding for common services (SSH: 22, HTTP: 8080, etc.)
- Network isolation options
- Host networking when needed

**Environment Setup (preserving sandbox workflow):**
- Automatic PATH configuration for Portunix tools
- Environment variable forwarding
- **SSH Server Setup** (identical to sandbox logic):
  - Install OpenSSH server in container during startup
  - Generate random username and password
  - Configure SSH daemon with proper security settings
  - Start SSH service and expose port (default: 22 → random host port)
- **Connection Information Display:**
  - Display container IP address and SSH port mapping
  - Show generated SSH credentials (username/password)
  - Provide ready-to-use SSH connection command
  - Test SSH connectivity before showing success message
- System detection module availability

#### 3. Installation Type Support

**Available installation types (same as sandbox):**
```bash
portunix docker run-in-container default    # Python + Java + VSCode
portunix docker run-in-container empty      # Clean container
portunix docker run-in-container python     # Python only
portunix docker run-in-container java       # Java only  
portunix docker run-in-container vscode     # VSCode only
```

**Container-specific optimizations:**
- Lightweight base images for faster startup
- Multi-stage builds for smaller images
- Package caching for repeated builds
- Automatic cleanup of build artifacts

#### 4. Cross-Platform Support

**Linux Containers (primary):**
- Ubuntu 22.04, 20.04 base images
- Alpine Linux for minimal footprint
- Debian stable for compatibility
- CentOS/RHEL for enterprise environments

**Windows Containers (if supported):**
- Windows Server Core base images
- PowerShell and system detection integration
- Windows-specific tool installations

#### 5. Use Cases

**Docker Installation:**
```bash
# Interactive installation with storage selection
portunix install docker

# Automated installation (accepts recommended storage)
portunix install docker -y

# Windows: Installs Docker Desktop with optimal drive selection
# Ubuntu: Installs docker.io via apt with optimal /var/lib/docker location
# CentOS: Installs docker-ce via yum with optimal storage
# Alpine: Installs docker via apk with optimal storage
```

**Development Environment:**
```bash
# Quick development container with all tools
portunix docker run-in-container default --port 8080:8080 --volume $(pwd):/workspace

# Python development environment
portunix docker run-in-container python --keep-running
```

**CI/CD Integration:**
```bash
# Automated testing in clean environment  
portunix docker run-in-container empty
portunix docker exec test-container "portunix install python && python test.py"
```

**Isolated Installations:**
```bash
# Test installations without affecting host
portunix docker run-in-container java --disposable
portunix docker run-in-container vscode --ssh-enabled
```

#### 6. Configuration Options

**Base Image Selection:**
```bash
# Use specific Ubuntu version
portunix docker run-in-container python --image ubuntu:22.04

# Use Alpine for minimal footprint
portunix docker run-in-container default --image alpine:3.18

# Use CentOS for enterprise compatibility  
portunix docker run-in-container java --image centos:8

# Use custom registry image
portunix docker run-in-container empty --image myregistry.com/myimage:v1.0

# Default behavior (if no --image specified)
portunix docker run-in-container python  # Uses ubuntu:22.04 by default
```

**Container Configuration:**
```bash
--image ubuntu:22.04          # Specify base image
--name my-portunix            # Container name
--port 8080:8080              # Port forwarding  
--volume /host:/container     # Volume mounting
--cache-shared                # Mount host .cache to /portunix-cache (default: true)
--cache-path /custom/cache    # Custom cache directory path  
--no-cache                    # Disable cache mounting
--env VAR=value               # Environment variables
--ssh-enabled                 # Enable SSH server
--keep-running               # Don't stop after execution
--disposable                 # Auto-remove when stopped
--privileged                 # Run with privileges
--network host               # Use host networking
```

**Cache Usage Examples:**
```bash
# Default: automatic .cache mounting
portunix docker run-in-container python
# Mounts: .cache:/portunix-cache

# Custom cache directory
portunix docker run-in-container java --cache-path /my/custom/cache
# Mounts: /my/custom/cache:/portunix-cache

# Disable cache for clean environment
portunix docker run-in-container empty --no-cache
# No cache mounting

# Share cache with existing container
portunix docker run-in-container default --cache-shared
# Leverages previously downloaded packages
```

#### 7. Intelligent Package Manager Detection

**Automatic Package Manager Recognition:**
- **Ubuntu/Debian:** Use `apt-get` for package installations
- **CentOS/RHEL/Rocky/Fedora:** Use `yum` or `dnf` for package installations  
- **Alpine:** Use `apk` for package installations
- **Custom Images:** Attempt to detect available package managers

**Installation Script Adaptation:**
```bash
# Same command, different execution based on base image
portunix docker run-in-container python --image ubuntu:22.04
# → Uses: apt-get update && apt-get install python3

portunix docker run-in-container python --image alpine:3.18  
# → Uses: apk add --no-cache python3

portunix docker run-in-container python --image centos:8
# → Uses: yum install -y python3
```

**Image Compatibility Matrix:**
| Installation Type | Ubuntu | Debian | Alpine | CentOS | Fedora |
|------------------|--------|--------|--------|--------|--------|
| `python`         | ✅      | ✅      | ✅      | ✅      | ✅      |
| `java`           | ✅      | ✅      | ✅      | ✅      | ✅      |
| `vscode`         | ✅      | ✅      | ⚠️*     | ✅      | ✅      |
| `default`        | ✅      | ✅      | ⚠️*     | ✅      | ✅      |

*Alpine may require additional setup for some GUI applications

#### 8. Expected Output Examples

**Docker Installation (Interactive):**
```bash
$ portunix install docker
Detecting operating system...
✓ Detected: Windows 11 Pro (Build 22631)
✓ Docker Desktop compatible
✓ WSL2 available

Analyzing available storage...
📊 Drive Space Analysis:
   C:\ - 125 GB available (system drive)
   D:\ - 850 GB available (recommended)
   E:\ - 2.1 TB available (external)

💡 Storage Recommendation: Drive D:\ (850 GB available)
   Docker images and containers can consume significant space.
   Using D:\ will provide better performance and prevent system drive filling up.

Select Docker data storage location:
1. C:\ (default) - 125 GB available
2. D:\ (recommended) - 850 GB available  
3. E:\ - 2.1 TB available
4. Custom path

Choice [2]: 2

Configuring Docker Desktop for D:\ storage...
Installing Docker Desktop for Windows...
📦 Downloading Docker Desktop installer...
✓ Download completed (450MB)
🔧 Running installer with admin privileges...
✓ Docker Desktop installed successfully
🔧 Configuring data-root: D:\docker-data
✓ WSL2 backend configured
✓ Docker daemon started

Verification:
✓ docker --version: Docker version 24.0.6
✓ docker compose --version: Docker Compose version v2.21.0
✓ Docker data location: D:\docker-data

Docker is ready to use!
Storage optimization: Images and containers will be stored on D:\ drive.
```

**Docker Installation (Automated with -y flag):**
```bash
$ portunix install docker -y
Detecting operating system...
✓ Detected: Windows 11 Pro (Build 22631)
✓ Docker Desktop compatible
✓ WSL2 available

Analyzing available storage...
✓ Automatically selected optimal storage: D:\ (850 GB available)

Installing Docker Desktop for Windows...
📦 Downloading Docker Desktop installer...
✓ Download completed (450MB)
🔧 Running installer with admin privileges...
✓ Docker Desktop installed successfully
🔧 Configuring data-root: D:\docker-data
✓ WSL2 backend configured
✓ Docker daemon started

Verification:
✓ docker --version: Docker version 24.0.6
✓ Docker data location: D:\docker-data

Docker is ready to use with optimal storage configuration!
```

**Docker Installation (Linux with storage optimization):**
```bash
$ portunix install docker
Detecting operating system...
✓ Detected: Ubuntu 22.04 LTS
✓ Package manager: apt

Analyzing available storage...
📊 Partition Space Analysis:
   / - 45 GB available (root partition)
   /home - 180 GB available (recommended)
   /data - 500 GB available

💡 Storage Recommendation: /data (500 GB available)
   Docker images and containers can consume significant space.
   Using /data will prevent root partition from filling up.

Select Docker data storage location:
1. /var/lib/docker (default) - 45 GB available
2. /home/docker-data (recommended) - 180 GB available
3. /data/docker-data - 500 GB available
4. Custom path

Choice [3]: 3

Installing Docker Engine...
🔧 Adding Docker GPG key...
🔧 Adding Docker repository...
📦 Installing docker.io package...
✓ Docker installed successfully
🔧 Configuring data-root: /data/docker-data
✓ Adding user to docker group...
✓ Enabling docker service...
✓ Docker daemon started

Verification:
✓ docker --version: Docker version 24.0.7
✓ Docker data location: /data/docker-data
✓ Docker daemon is running

Docker is ready to use with optimal storage configuration!
Note: You may need to log out and back in for group changes to take effect.
```

**Starting Container (with SSH workflow):**
```
$ portunix docker run-in-container python
Pulling base image: ubuntu:22.04...
✓ Image pulled successfully
✓ Detected package manager: apt-get
✓ Creating container: portunix-python-abc123

Setting up container environment...
✓ Installing Portunix binary
✓ Mounting cache directory: .cache:/portunix-cache
✓ Found cached packages: python-embeddable, openssh
✓ Setting up Python via apt-get...
✓ Leveraging cached Python packages from host
✓ Configuring PATH environment
✓ Installing OpenSSH server (using cached binaries)...
✓ Generating SSH credentials...
✓ Starting SSH daemon...
✓ Container startup completed

Testing SSH connectivity...
✓ SSH server is responding on port 2222
✓ Authentication test successful

📡 SSH CONNECTION INFORMATION:
════════════════════════════════════════════════════════
🔗 Container IP:   172.17.0.2
📄 SSH Port:      localhost:2222 
👤 Username:      portunix_user_abc123
🔐 Password:      Kx9mP2nQ8vR5wL7s
📄 SSH Command:   ssh portunix_user_abc123@localhost -p 2222

💡 CONNECTION TIPS:
   • Open new terminal window
   • Run: ssh portunix_user_abc123@localhost -p 2222
   • Enter password: Kx9mP2nQ8vR5wL7s
   • Files are shared at: /workspace
   • Cache directory: /portunix-cache
   • Portunix tools available in PATH
════════════════════════════════════════════════════════

Container is running and ready for SSH connections!

Available management commands:
  portunix docker exec abc123 "command"     # Execute command
  portunix docker logs abc123               # View container logs  
  portunix docker stop abc123               # Stop container
  portunix docker remove abc123             # Remove container
```

**Image Selection Examples:**
```bash
$ portunix docker run-in-container python --image alpine:3.18
Pulling image: alpine:3.18...
✓ Image pulled successfully
✓ Detected package manager: apk
✓ Base OS: Alpine Linux 3.18
✓ Installing Portunix binary
✓ Setting up Python via apk...
✓ Container ready: portunix-python-alpine-abc123

Container Information:
===================  
Base Image:     alpine:3.18
Package Mgr:    apk
Container ID:   abc123456789
Size:          45MB (lightweight)
```

```bash
$ portunix docker run-in-container java --image ubuntu:20.04
Using cached image: ubuntu:20.04
✓ Detected package manager: apt-get  
✓ Base OS: Ubuntu 20.04 LTS
✓ Installing Portunix binary
✓ Setting up Java via apt-get...
✓ Container ready: portunix-java-ubuntu2004-def456

Container Information:
===================
Base Image:     ubuntu:20.04  
Package Mgr:    apt-get
Container ID:   def456789012
Size:          180MB (full-featured)
```

**Container List:**
```
$ portunix docker list
NAME                IMAGE                STATUS      PORTS           CREATED
portunix-default    portunix:ubuntu      Running     22:2222         2 hours ago
portunix-python     portunix:alpine      Running     22:2223         1 hour ago  
portunix-java       portunix:ubuntu      Stopped     -               3 hours ago
```

#### 9. Implementation Notes

**Architecture:**
- Create `cmd/docker.go` with subcommands (similar to `cmd/sandbox.go`)
- Create `app/docker/` package for Docker management
- Reuse existing system detection for container environment
- Support both Docker Engine and Podman

**Smart Installation System:**
- Detect base image OS and package manager at runtime
- Maintain installation templates for each supported package manager
- Automatic fallback strategies for unsupported package managers  
- Pre-validation of image compatibility before container creation

**Image Management:**
- Pull images automatically if not present locally
- Validate image authenticity and accessibility
- Cache frequently used images for faster startup
- Support for private registries with authentication
- Multi-stage Dockerfiles for optimized images
- Automated image building with caching
- Version-tagged images for reproducibility
- Registry push capabilities for sharing

**Integration with Existing Features:**
- Reuse installation system from `app/install/`  
- Integration with system detection (`app/system/`)
- PowerShell module support in containers
- Consistent CLI patterns with sandbox commands

#### 10. Benefits

**Development Workflow:**
- Isolated, reproducible development environments
- Consistent tooling across team members  
- Easy cleanup and reset of environments
- Cross-platform development support
- **Flexibility in OS choice** - Unlike Windows Sandbox, choose any Linux distribution

**Testing & CI/CD:**
- Clean environment for each test run
- Automated installation validation
- Multi-OS testing capabilities  
- Integration with existing CI systems

**Deployment:**
- Containerized Portunix deployments
- Scalable installation services
- Cloud-native development workflows
- Easy distribution of configured environments

### Implementation Priority
1. **Phase 1:** Basic container management (build, run, stop, list)
2. **Phase 2:** Installation types integration (default, python, java)  
3. **Phase 3:** Advanced features (SSH, port forwarding, volumes)
4. **Phase 4:** Cross-platform containers and Windows container support

### Key Differentiator from Windows Sandbox
This feature would complement the existing Windows Sandbox functionality by providing similar capabilities in a **cross-platform, lightweight container environment** with the major advantage of **flexible base image selection** - users can choose from Ubuntu, Alpine, CentOS, or any custom image, unlike Windows Sandbox where the OS is fixed.

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented:** 2025-01-09  
**Version:** v1.5.7+  

### Implementation Summary

All requested Docker management features have been successfully implemented:

#### ✅ Core Docker Management Commands
- **`portunix docker install`** - Intelligent OS-specific Docker installation
- **`portunix docker run-in-container`** - Container-based Portunix installations
- **`portunix docker build/start/stop/list/remove/logs/exec`** - Full lifecycle management
- **Multi-platform container support** - Ubuntu, Alpine, CentOS, Debian images

#### ✅ Advanced Features
- **SSH-enabled containers** with automatic server setup and credential generation
- **Package manager detection** (apt-get, yum, dnf, apk) with automatic configuration
- **Cache directory mounting** for persistent downloads and package caching
- **Flexible base image selection** with validation and automatic pulling
- **Installation type integration** (default, python, java, vscode, empty profiles)

#### ✅ Cross-Platform Installation
- **Windows:** Docker Desktop installation with storage optimization prompts
- **Linux:** Docker Engine (docker.io/docker-ce) installation with distribution detection  
- **Intelligent storage configuration** with disk space optimization recommendations

### Verification Commands

```bash
# Test Docker management commands
portunix docker --help
portunix docker install
portunix docker run-in-container python
portunix docker list
portunix docker logs <container-id>

# Test container installations
portunix docker run-in-container default --image ubuntu:22.04
portunix docker run-in-container java --image alpine:3.18
```

### Resolution Status

**FULLY IMPLEMENTED** ✅ - All requested features are now available including the comprehensive Docker management command suite with cross-platform installation support, container lifecycle management, SSH integration, and flexible development environment setup.

---
**Created:** 2025-01-18  
**Last Updated:** 2025-01-09  
**Assigned:** Development Team  
**Related Issues:** [#1](001-cross-platform-os-detection.md) (Cross-Platform Intelligent OS Detection System), [#30](030-container-tls-certificate-verification-failure.md) (Certificate Management)