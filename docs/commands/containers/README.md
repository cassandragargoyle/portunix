# Container Commands

Container management and orchestration commands for Docker and Podman with enhanced Portunix integration.

## Commands in this Category

### [`container`](container.md) - Universal Container Interface **⭐ Main Command**
**The primary container management command** that automatically detects and uses the best available container runtime (Docker or Podman). This is the **recommended command** for all container operations.

**Quick Examples:**
```bash
portunix container run ubuntu               # Auto-detect and use best runtime
portunix container list --all-runtimes     # List containers across Docker/Podman
portunix container run ubuntu --runtime podman  # Force specific runtime
```

**Key Features:**
- **Automatic runtime detection** (Docker/Podman priority)
- **Universal API** across all container platforms
- **Intelligent fallback** when primary runtime unavailable
- **Cross-runtime management** and migration
- **Performance optimization** per runtime
- **Unified monitoring** across all runtimes

**Why Use `container` Instead of `docker`:**
- **Future-proof**: Works with Docker, Podman, and future runtimes
- **Intelligent**: Automatically selects best runtime for your system
- **Consistent**: Same commands work across different environments
- **Flexible**: Easy runtime switching without command changes

---

### [`docker`](docker.md) - Docker-Specific Operations
Enhanced Docker integration with SSH access, pre-installed software, and Docker-specific features. **Use this when you specifically need Docker features or compatibility.**

**Quick Examples:**
```bash
portunix docker run ubuntu               # Specifically use Docker
portunix docker run-in-container nodejs  # Container with pre-installed Node.js
portunix docker ssh my-container         # SSH into Docker container
```

**Key Features:**
- Built-in SSH access to containers
- Pre-installed software packages via Portunix
- Container templates for common environments
- Multi-platform container support (AMD64, ARM64)
- Integrated monitoring and resource management
- Docker-specific optimizations and features

**When to Use `docker` Instead of `container`:**
- Docker-specific features (BuildKit, Docker Desktop integration)
- Existing Docker workflows that need specific compatibility
- Advanced Docker networking or volume configurations
- Docker Swarm or Docker Compose integration

---

### `podman` - Podman-Specific Operations *(Coming Soon)*
Podman-specific operations with rootless container support and enhanced security features.

**Planned Features:**
- Rootless container operations
- Systemd integration
- Podman-specific security features
- Pod management capabilities

## Category Overview

The **Containers** category provides powerful container management capabilities that enhance standard Docker and Podman functionality with Portunix-specific features.

### Enhanced Container Features

#### SSH Integration
Unlike standard containers, Portunix containers come with built-in SSH access:
```bash
# Standard Docker - no direct SSH
docker run -d ubuntu sleep infinity
docker exec -it container_id bash

# Portunix Docker - SSH ready
portunix docker run ubuntu --name dev-env
portunix docker ssh dev-env
```

#### Pre-installed Software
Install software during container creation:
```bash
# Create container with development tools pre-installed
portunix docker run-in-container "nodejs python git vscode" \
  --image ubuntu:22.04 \
  --name full-dev-env
```

#### Container Templates
Reusable container configurations:
```bash
# Create template from existing container
portunix docker template create node-dev --from my-container

# Use template for new containers
portunix docker run --template node-dev --name new-project
```

## Architecture Integration

### With Core Commands
```bash
# Install Portunix in container
portunix docker run ubuntu --name portunix-container
portunix docker ssh portunix-container
# Inside container:
portunix install nodejs python
```

### With Plugin System
```bash
# Container with plugin support
portunix docker run ubuntu --enable-plugins --name plugin-dev
portunix docker exec plugin-dev portunix plugin install agile-software-development
```

### With MCP Integration
```bash
# MCP-enabled development container
portunix docker run-mcp-dev --image ubuntu:22.04 --name ai-dev
# Provides MCP server access for AI assistants
```

### With Virtualization
```bash
# Run containers inside VMs for extra isolation
portunix virt create container-vm --iso ubuntu.iso
portunix virt ssh container-vm "portunix docker run nginx"
```

## Development Workflows

### Full-Stack Development Environment
```bash
# 1. Create development container with full stack
portunix docker run-in-container "nodejs python postgresql redis" \
  --name fullstack-dev \
  --ports 3000:3000,5432:5432,6379:6379 \
  --volumes $(pwd):/workspace

# 2. SSH into development environment
portunix docker ssh fullstack-dev

# 3. Start services and develop
# Inside container: start database, Redis, and dev server
```

### Microservices Development
```bash
# Create network for microservices
portunix docker network create microservices

# Service 1: API Gateway
portunix docker run-in-container nodejs \
  --name api-gateway \
  --network microservices \
  --ports 8080:8080

# Service 2: User Service
portunix docker run-in-container "nodejs postgresql" \
  --name user-service \
  --network microservices \
  --ports 8081:8081

# Service 3: Product Service
portunix docker run-in-container "nodejs mongodb" \
  --name product-service \
  --network microservices \
  --ports 8082:8082
```

### Testing and CI/CD
```bash
# 1. Create test environment
portunix docker run-in-container "nodejs python java" \
  --name test-env \
  --volumes $(pwd):/app \
  --working-dir /app

# 2. Run tests in isolated environment
portunix docker exec test-env npm test
portunix docker exec test-env python -m pytest
portunix docker exec test-env mvn test

# 3. Clean up after tests
portunix docker remove test-env --volumes
```

## Container Security and Best Practices

### Security Features
```bash
# Run with security profiles
portunix docker run ubuntu --security-profile restricted

# Non-root execution
portunix docker run ubuntu --user 1000:1000 --no-root

# Read-only filesystem with writable tmp
portunix docker run ubuntu --read-only --tmpfs /tmp

# Network isolation
portunix docker run ubuntu --network none
```

### Resource Management
```bash
# Set resource limits
portunix docker run ubuntu \
  --memory 512m \
  --cpus 0.5 \
  --swap 256m

# Monitor resource usage
portunix docker stats --live
portunix docker monitor my-container
```

### Backup and Recovery
```bash
# Backup container with data
portunix docker backup my-container --include-volumes

# Export container configuration
portunix docker export-config my-container > container.json

# Restore container on different host
portunix docker restore --backup container-backup.tar.gz
```

## Multi-Platform Support

### Cross-Platform Development
```bash
# Build for ARM64 on x86 machine
portunix docker run ubuntu:22.04 --platform linux/arm64

# Test application on different architectures
portunix docker run --platform linux/arm64 my-app:latest
portunix docker run --platform linux/amd64 my-app:latest
```

### Platform-Specific Features
```bash
# Windows containers (Windows only)
portunix docker run mcr.microsoft.com/windows/servercore:ltsc2022

# Linux containers on Windows with WSL2
portunix docker run ubuntu --linux --wsl2

# GPU acceleration (Linux/Windows)
portunix docker run tensorflow/tensorflow:latest-gpu --gpu all
```

## Performance Optimization

### Container Optimization
```bash
# Optimize startup time
portunix docker run ubuntu --fast-start --preload-libs

# Layer optimization
portunix docker optimize-image my-app:latest

# Cache optimization
portunix docker build --cache-from my-app:cache .
```

### Monitoring and Metrics
```bash
# Performance monitoring
portunix docker metrics my-container --export prometheus

# Resource recommendations
portunix docker recommend-resources my-container

# Performance analysis
portunix docker analyze-performance my-container
```

## Integration Examples

### With IDEs and Development Tools
```bash
# VS Code development container
portunix docker run-in-container "nodejs python git" \
  --name vscode-dev \
  --volumes $(pwd):/workspace \
  --ports 8080:8080 \
  --env DISPLAY=$DISPLAY

# IntelliJ IDEA remote development
portunix docker run-in-container java \
  --name idea-remote \
  --volumes ~/.m2:/root/.m2 \
  --ports 22:22
```

### With External Services
```bash
# Database development
portunix docker run postgresql \
  --name dev-db \
  --env POSTGRES_DB=myapp \
  --env POSTGRES_USER=developer \
  --env POSTGRES_PASSWORD=secret \
  --volumes db-data:/var/lib/postgresql/data

# Redis cache
portunix docker run redis \
  --name dev-cache \
  --ports 6379:6379 \
  --volumes redis-data:/data
```

## Troubleshooting Common Issues

### Container Connectivity
```bash
# Diagnose network issues
portunix docker network diagnose

# Test container connectivity
portunix docker test my-container --network --connectivity

# Reset networking
portunix docker network reset
```

### Performance Issues
```bash
# Resource analysis
portunix docker analyze my-container --performance --bottlenecks

# Memory debugging
portunix docker debug my-container --memory-profile

# CPU optimization
portunix docker tune my-container --cpu-optimization
```

### SSH Access Problems
```bash
# Regenerate SSH keys
portunix docker ssh-setup my-container --regenerate-keys

# Debug SSH connection
portunix docker ssh my-container --debug

# Manual SSH configuration
portunix docker ssh-config my-container --show
```

## Future Roadmap

### Planned Features
- **Universal Container Interface** - Automatic Docker/Podman detection
- **Container Orchestration** - Built-in container clustering
- **Advanced Networking** - Service mesh integration
- **Security Scanning** - Vulnerability assessment
- **Performance Analytics** - ML-based optimization

### Integration Improvements
- **Kubernetes Integration** - Native K8s deployment
- **Cloud Platform Support** - AWS ECS, Azure Container Instances
- **Registry Management** - Private registry support
- **Backup Automation** - Scheduled container backups

## Related Categories

- **[Core](../core/)** - Install software in containers
- **[Virtualization](../virtualization/)** - Containers in VMs
- **[Plugins](../plugins/)** - Container-aware plugins
- **[Integration](../integration/)** - MCP in containers

---

*Container commands provide isolated, reproducible development environments with enhanced Portunix integration.*