# Portunix Container Command

## Quick Start

The `container` command is the **universal container interface** that automatically detects and uses the best available container runtime (Docker or Podman) on your system. It provides a consistent API across all container platforms.

### Simplest Usage
```bash
# Universal container operations (auto-detects Docker/Podman)
portunix container run ubuntu

# List containers across all runtimes
portunix container list

# Execute commands in containers
portunix container exec my-container bash
```

### Basic Syntax
```bash
portunix container [subcommand] [options]
```

### Common Subcommands
- `run` - Create and run a container
- `list` - List containers across all runtimes
- `exec` - Execute command in container
- `stop` - Stop container
- `remove` - Remove container
- `logs` - View container logs
- `status` - Check container runtime status

## Intermediate Usage

### Automatic Runtime Detection

The `container` command intelligently selects the best available runtime:

```bash
# Automatic detection priority:
# 1. Docker (if available and running)
# 2. Podman (if available)
# 3. Alternative runtimes (future support)

portunix container status
```

Output example:
```
Container Runtime Status
========================
Available Runtimes:
  ✅ Docker    - Version 24.0.7 (Primary)
  ✅ Podman    - Version 4.7.2 (Available)

Selected Runtime: Docker
Reason: Docker daemon running, preferred for compatibility

Containers Summary:
  Running: 3 (2 Docker, 1 Podman)
  Stopped: 5 (4 Docker, 1 Podman)
  Total: 8
```

### Universal Container Operations

```bash
# Run containers with automatic runtime selection
portunix container run ubuntu --name test-container

# Runtime can be explicitly specified if needed
portunix container run ubuntu --runtime podman --name podman-container
portunix container run ubuntu --runtime docker --name docker-container

# Cross-runtime container listing
portunix container list --all-runtimes

# Runtime-specific operations
portunix container list --runtime docker
portunix container list --runtime podman
```

### Enhanced Container Features

The `container` command provides enhanced functionality beyond basic Docker/Podman:

```bash
# Pre-installed software in containers
portunix container run-with-software ubuntu nodejs python git

# SSH-enabled containers
portunix container run ubuntu --enable-ssh --name ssh-container
portunix container ssh ssh-container

# Development environment containers
portunix container create-dev-env \
  --name fullstack-dev \
  --stack "nodejs python postgresql redis" \
  --volumes $(pwd):/workspace
```

### Runtime Compatibility Layer

```bash
# Unified commands work across Docker and Podman
portunix container run nginx --ports 8080:80
# Translates to:
# - docker run -p 8080:80 nginx (if Docker selected)
# - podman run -p 8080:80 nginx (if Podman selected)

# Volume mounting with path translation
portunix container run ubuntu --volumes /host/data:/container/data
# Handles platform-specific path formats automatically
```

## Advanced Usage

### Multi-Runtime Management

```bash
# Manage containers across multiple runtimes
portunix container list --format table

# Output shows runtime for each container:
```
| Name | Runtime | Status | Image | Ports |
|------|---------|--------|--------|--------|
| web-app | docker | running | nginx:latest | 80:8080 |
| db-dev | podman | running | postgres:13 | 5432:5432 |
| test-env | docker | stopped | ubuntu:22.04 | - |

```bash
# Cross-runtime operations
portunix container stop --all-runtimes
portunix container cleanup --unused --all-runtimes
```

### Runtime-Specific Optimizations

```bash
# Docker-specific optimizations
portunix container run ubuntu \
  --runtime docker \
  --optimize-for docker-desktop

# Podman-specific optimizations
portunix container run ubuntu \
  --runtime podman \
  --rootless \
  --systemd-integration

# Automatic optimization based on detected runtime
portunix container run ubuntu --auto-optimize
```

### Container Migration Between Runtimes

```bash
# Migrate containers between Docker and Podman
portunix container migrate web-app --from docker --to podman

# Export container from one runtime
portunix container export web-app --runtime docker --output web-app.tar

# Import to different runtime
portunix container import web-app.tar --runtime podman --name web-app-podman
```

### Universal Container Networking

```bash
# Create networks that work across runtimes
portunix container network create dev-network \
  --subnet 172.20.0.0/16 \
  --runtime-agnostic

# Connect containers from different runtimes
portunix container network connect dev-network docker-app
portunix container network connect dev-network podman-app
```

### Unified Container Monitoring

```bash
# Monitor containers across all runtimes
portunix container monitor --all-runtimes

# Resource usage across runtimes
portunix container stats --all-runtimes

# Health checks for all containers
portunix container health --all-runtimes
```

## Expert Tips & Tricks

### 1. Runtime Failover Configuration

```bash
# Configure runtime failover preferences
portunix container config set \
  --primary-runtime docker \
  --fallback-runtime podman \
  --auto-failover true

# Test failover behavior
portunix container test-failover
```

### 2. Performance Optimization per Runtime

```bash
# Docker optimizations
portunix container optimize --runtime docker \
  --use-buildkit \
  --enable-experimental

# Podman optimizations
portunix container optimize --runtime podman \
  --use-cgroups-v2 \
  --enable-pasta

# Auto-optimization based on workload
portunix container optimize --workload development
portunix container optimize --workload production
```

### 3. Cross-Runtime Development Workflows

```bash
# Development with multiple runtimes
portunix container create-workspace development \
  --docker-containers "frontend backend" \
  --podman-containers "database cache" \
  --shared-network dev-net

# Production deployment testing
portunix container deploy-test \
  --docker-production \
  --podman-development \
  --validate-compatibility
```

### 4. Runtime-Specific Feature Detection

```bash
# Check runtime capabilities
portunix container capabilities --runtime docker
portunix container capabilities --runtime podman

# Feature compatibility matrix
portunix container features compare docker podman

# Auto-select runtime based on required features
portunix container run myapp \
  --requires "rootless,systemd,gpu" \
  --auto-select-runtime
```

### 5. Universal Container Backup

```bash
# Backup containers from all runtimes
portunix container backup --all-runtimes \
  --output unified-backup.tar.gz \
  --include-volumes \
  --include-networks

# Restore with runtime selection
portunix container restore unified-backup.tar.gz \
  --target-runtime podman \
  --remap-networks
```

## Integration with Portunix Ecosystem

### With Core Commands

```bash
# Install software and create containers
portunix install docker podman
portunix container run ubuntu --with-portunix

# System information includes container runtimes
portunix system info --containers
```

### With Specific Runtime Commands

The `container` command orchestrates but doesn't replace specific runtime commands:

```bash
# Universal interface
portunix container run nginx

# Still available for specific needs:
portunix docker run nginx --docker-specific-option
portunix podman run nginx --podman-specific-option
```

### With Plugin System

```bash
# Container-aware plugins
portunix plugin install container-dev-tools
portunix container run ubuntu --enable-plugins

# Plugin operations across runtimes
portunix plugin run deployment-plugin --target-runtime auto
```

### With MCP Integration

```bash
# MCP tools work with universal container interface
portunix mcp serve --container-integration

# AI can manage containers across runtimes
# AI: "Create a development environment with the best available container runtime"
```

## Platform-Specific Considerations

### Linux

```bash
# Prefer Podman for rootless containers
portunix container config set linux.prefer-rootless true

# Systemd integration detection
portunix container run app --systemd-auto-detect

# Cgroups v2 optimization
portunix container optimize --cgroups-v2
```

### Windows

```bash
# Windows container support
portunix container run mcr.microsoft.com/windows/servercore

# WSL2 integration
portunix container config set windows.wsl2-integration true

# Hyper-V container isolation
portunix container run ubuntu --isolation hyperv
```

### macOS

```bash
# Docker Desktop integration
portunix container config set macos.docker-desktop true

# Performance optimization for macOS
portunix container optimize --platform macos

# Lima/Podman Desktop integration
portunix container config set macos.lima-integration true
```

## Runtime Detection Algorithm

### Detection Priority

1. **Docker Daemon Check**
   - Check if Docker daemon is running
   - Verify Docker CLI accessibility
   - Test basic Docker operations

2. **Podman Availability**
   - Check Podman installation
   - Verify rootless/rootful capability
   - Test Podman socket availability

3. **Feature Requirements**
   - Match required features to runtime capabilities
   - Consider performance characteristics
   - Evaluate security requirements

4. **User Preferences**
   - Respect explicit runtime selection
   - Apply configured preferences
   - Consider workspace-specific settings

### Fallback Behavior

```bash
# Graceful degradation
portunix container run ubuntu
# If Docker fails -> try Podman
# If Podman fails -> error with helpful suggestions

# Force specific runtime
portunix container run ubuntu --runtime docker --no-fallback
```

## Configuration Management

### Global Configuration

```bash
# Set global container preferences
portunix container config set \
  --default-runtime auto \
  --prefer-rootless true \
  --enable-cross-runtime-networking true

# View current configuration
portunix container config show

# Reset to defaults
portunix container config reset
```

### Workspace-Specific Configuration

```bash
# Project-specific runtime preferences
echo "container.runtime=podman" > .portunix.toml
echo "container.features=[rootless,systemd]" >> .portunix.toml

# Auto-apply workspace settings
portunix container run ubuntu # Uses workspace preferences
```

## Troubleshooting

### Common Issues

#### 1. No Container Runtime Available
```bash
# Check runtime availability
portunix container diagnose

# Install recommended runtime
portunix install docker
# or
portunix install podman

# Verify installation
portunix container status
```

#### 2. Runtime Detection Issues
```bash
# Force runtime detection refresh
portunix container refresh-detection

# Debug detection process
portunix container diagnose --runtime-detection --verbose

# Manual runtime specification
portunix container run ubuntu --runtime docker --force
```

#### 3. Cross-Runtime Compatibility
```bash
# Check compatibility between runtimes
portunix container compatibility-check

# Migrate problematic containers
portunix container migrate --auto-fix-issues

# Runtime-specific troubleshooting
portunix container diagnose --runtime-specific
```

#### 4. Performance Issues
```bash
# Compare runtime performance
portunix container benchmark --all-runtimes

# Optimize for detected setup
portunix container optimize --auto

# Resource usage analysis
portunix container analyze --performance --all-runtimes
```

### Debug Mode

```bash
# Verbose container operations
portunix container run ubuntu --debug --trace-runtime-selection

# Runtime detection debugging
portunix container --debug status

# Cross-runtime operation tracing
portunix container --trace list --all-runtimes
```

## Best Practices

### Runtime Selection
- Use `auto` for most cases unless specific features needed
- Prefer rootless containers for security (Podman advantage)
- Use Docker for maximum compatibility with ecosystem
- Consider Podman for server environments and security-focused setups

### Performance Optimization
- Enable auto-optimization for detected platform
- Use runtime-specific optimizations when needed
- Monitor cross-runtime performance differences
- Choose runtime based on workload characteristics

### Security Considerations
- Prefer rootless containers when possible
- Use runtime-specific security features
- Enable security scanning across all runtimes
- Regular security updates for all container runtimes

## Future Roadmap

### Planned Features
- **Additional Runtime Support** - containerd, CRI-O integration
- **Advanced Migration Tools** - Seamless container migration between runtimes
- **Performance Intelligence** - ML-based runtime selection
- **Security Orchestration** - Unified security policies across runtimes

### Integration Improvements
- **Kubernetes Integration** - Multi-runtime cluster support
- **Cloud Platform Support** - Hybrid cloud container management
- **Development Tools** - IDE integration with runtime awareness
- **CI/CD Enhancement** - Pipeline optimization per runtime

## Related Commands

- [`docker`](docker.md) - Docker-specific operations
- `podman` - Podman-specific operations *(Coming Soon)*
- [`install`](../core/install.md) - Install container runtimes
- [`system`](../core/system.md) - System container information

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--runtime` | string | `auto` | Container runtime (auto/docker/podman) |
| `--all-runtimes` | boolean | `false` | Include all runtimes in operation |
| `--no-fallback` | boolean | `false` | Disable runtime fallback |
| `--auto-optimize` | boolean | `false` | Auto-optimize for detected runtime |
| `--rootless` | boolean | `auto` | Prefer rootless containers |
| `--enable-ssh` | boolean | `false` | Enable SSH access |
| `--with-software` | string | - | Pre-install software packages |
| `--debug` | boolean | `false` | Debug runtime selection |
| `--format` | string | `table` | Output format (table/json/yaml) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | No container runtime available |
| 3 | Runtime detection failed |
| 4 | Cross-runtime operation failed |
| 5 | Container creation failed |
| 6 | Network configuration failed |
| 7 | Migration failed |
| 8 | Permission denied |
| 9 | Resource limit exceeded |
| 10 | Compatibility issue |

## Version History

- **v1.6.0** - Initial universal container interface
- **v1.6.1** - Added cross-runtime networking
- **v1.6.2** - Enhanced migration capabilities
- **v1.6.3** - Performance optimization features
- **v1.7.0** - Advanced runtime detection (planned)