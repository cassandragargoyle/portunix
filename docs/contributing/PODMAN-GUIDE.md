# Podman Guide

## Purpose
This document provides comprehensive guidance for using Podman as a container engine in CassandraGargoyle projects. Podman is a daemonless, rootless container engine that provides Docker-compatible commands while offering enhanced security features.

## What is Podman?

**Podman** (Pod Manager) is an open-source container management tool that allows you to manage OCI containers and pods. Unlike Docker, Podman doesn't require a daemon and can run containers as a non-root user, providing better security isolation.

### Key Features
- **Daemonless Architecture** - No background service required
- **Rootless Containers** - Run containers without root privileges
- **Docker Compatible** - Supports Docker CLI commands and Dockerfiles
- **Pod Support** - Native Kubernetes pod management
- **Systemd Integration** - Generate systemd units for containers
- **Security Enhanced** - Better isolation and security by default

## Installation

### Windows
```powershell
# Using Chocolatey
choco install podman-desktop

# Using Winget
winget install RedHat.Podman-Desktop

# Using Portunix
portunix install podman
```

### Linux

#### Ubuntu/Debian
```bash
# Add repository
sudo apt update
sudo apt install -y podman

# Using Portunix
portunix install podman
```

#### Fedora/RHEL/Rocky Linux
```bash
# Install from default repository
sudo dnf install -y podman

# Using Portunix
portunix install podman
```

#### Arch Linux
```bash
# Install from community repository
sudo pacman -S podman

# Using Portunix
portunix install podman
```

### macOS
```bash
# Using Homebrew
brew install podman

# Initialize and start Podman machine
podman machine init
podman machine start

# Using Portunix
portunix install podman
```

## Basic Usage

### Container Management

#### Pull an Image
```bash
# Pull from Docker Hub (default)
podman pull ubuntu:22.04

# Pull from specific registry
podman pull registry.fedoraproject.org/fedora:39

# Pull from Quay.io
podman pull quay.io/podman/hello
```

#### Run a Container
```bash
# Run interactive container
podman run -it ubuntu:22.04 /bin/bash

# Run container in background
podman run -d --name webserver nginx

# Run with port mapping
podman run -d -p 8080:80 --name web nginx

# Run with volume mount
podman run -v /host/path:/container/path:Z ubuntu

# Run rootless container
podman run --userns=keep-id ubuntu
```

#### List Containers
```bash
# List running containers
podman ps

# List all containers
podman ps -a

# List with custom format
podman ps --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"
```

#### Stop and Remove
```bash
# Stop container
podman stop container_name

# Remove container
podman rm container_name

# Stop and remove all containers
podman stop -a
podman rm -a

# Remove unused images
podman image prune
```

### Image Management

#### Build Images
```bash
# Build from Dockerfile
podman build -t myapp:latest .

# Build with specific Dockerfile
podman build -f Dockerfile.prod -t myapp:prod .

# Build with build arguments
podman build --build-arg VERSION=1.0 -t myapp:1.0 .
```

#### Manage Images
```bash
# List images
podman images

# Tag image
podman tag myapp:latest myregistry.io/myapp:latest

# Push image
podman push myregistry.io/myapp:latest

# Remove image
podman rmi myapp:latest

# Save image to tar
podman save -o myapp.tar myapp:latest

# Load image from tar
podman load -i myapp.tar
```

### Pod Management

#### Create and Manage Pods
```bash
# Create a pod
podman pod create --name mypod -p 8080:80

# Add container to pod
podman run -d --pod mypod nginx

# List pods
podman pod list

# Stop pod
podman pod stop mypod

# Remove pod
podman pod rm mypod

# Generate Kubernetes YAML from pod
podman generate kube mypod > pod.yaml
```

## Rootless Configuration

### Setup Rootless Podman
```bash
# Enable lingering for user (allows containers to run after logout)
loginctl enable-linger $USER

# Check subuid/subgid ranges
cat /etc/subuid
cat /etc/subgid

# Configure user namespaces
podman unshare cat /proc/self/uid_map
```

### Storage Configuration
```bash
# Configure storage location
mkdir -p ~/.config/containers
cat > ~/.config/containers/storage.conf << EOF
[storage]
driver = "overlay"
runroot = "/run/user/$(id -u)/containers"
graphroot = "$HOME/.local/share/containers/storage"
EOF
```

### Network Configuration
```bash
# List networks
podman network ls

# Create custom network
podman network create mynetwork

# Inspect network
podman network inspect mynetwork

# Connect container to network
podman run -d --network mynetwork nginx
```

## Docker Compatibility

### Alias Setup
```bash
# Add to ~/.bashrc or ~/.zshrc
alias docker=podman

# For complete compatibility
sudo dnf install podman-docker  # Fedora/RHEL
sudo apt install podman-docker   # Debian/Ubuntu
```

### Docker Compose Alternative
```bash
# Install podman-compose
pip install podman-compose

# Or use docker-compose with Podman socket
systemctl --user start podman.socket
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
docker-compose up
```

### Migration from Docker
```bash
# Export from Docker
docker save myapp:latest -o myapp.tar

# Import to Podman
podman load -i myapp.tar

# Convert Docker Compose
podman-compose -f docker-compose.yml up
```

## Advanced Features

### Systemd Integration

#### Generate Systemd Units
```bash
# Generate systemd unit for container
podman generate systemd --name mycontainer > ~/.config/systemd/user/mycontainer.service

# Generate with auto-update
podman generate systemd --new --name mycontainer

# Enable and start service
systemctl --user daemon-reload
systemctl --user enable mycontainer.service
systemctl --user start mycontainer.service
```

#### Auto-update Containers
```bash
# Label container for auto-update
podman run -d --name web --label io.containers.autoupdate=registry nginx

# Enable auto-update timer
systemctl --user enable podman-auto-update.timer
systemctl --user start podman-auto-update.timer

# Manually trigger update
podman auto-update
```

### Security Features

#### SELinux Labels
```bash
# Run with SELinux context
podman run -v /host/path:/container/path:Z ubuntu  # Private label
podman run -v /host/path:/container/path:z ubuntu  # Shared label

# Check SELinux context
ls -Z /var/lib/containers
```

#### User Namespace Mapping
```bash
# Run with specific UID/GID mapping
podman run --uidmap 0:100000:65536 --gidmap 0:100000:65536 ubuntu

# Keep host user ID in container
podman run --userns=keep-id ubuntu
```

#### Capabilities Management
```bash
# Drop all capabilities
podman run --cap-drop=all ubuntu

# Add specific capability
podman run --cap-add=NET_ADMIN ubuntu

# Run privileged (use with caution)
podman run --privileged ubuntu
```

### Registry Configuration

#### Configure Registries
```bash
# Edit registries.conf
sudo vi /etc/containers/registries.conf

# Add insecure registry
[registries.insecure]
registries = ['localhost:5000']

# Add registry mirror
[registries.search]
registries = ['docker.io', 'quay.io', 'registry.fedoraproject.org']
```

#### Authentication
```bash
# Login to registry
podman login docker.io

# Login with credentials
podman login -u username -p password registry.example.com

# Logout
podman logout docker.io
```

## Integration with Development Tools

### Visual Studio Code
```json
// .devcontainer/devcontainer.json
{
  "name": "Podman Dev Container",
  "dockerFile": "Dockerfile",
  "runArgs": ["--userns=keep-id"],
  "remoteUser": "developer",
  "customizations": {
    "vscode": {
      "settings": {
        "docker.host": "unix:///run/user/1000/podman/podman.sock"
      }
    }
  }
}
```

### Jenkins Integration
```groovy
pipeline {
    agent {
        docker {
            image 'maven:3.8-openjdk-11'
            args '--userns=keep-id -v $HOME/.m2:/root/.m2'
            reuseNode true
        }
    }
    environment {
        DOCKER_HOST = 'unix:///run/user/1000/podman/podman.sock'
    }
}
```

### GitHub Actions
```yaml
name: Podman Build
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Podman
        run: |
          sudo apt update
          sudo apt install -y podman
      
      - name: Build with Podman
        run: |
          podman build -t myapp:${{ github.sha }} .
          podman save myapp:${{ github.sha }} -o myapp.tar
      
      - uses: actions/upload-artifact@v3
        with:
          name: container-image
          path: myapp.tar
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Fix storage permissions
podman system reset
podman system migrate

# Fix socket permissions
systemctl --user restart podman.socket
```

#### Networking Issues
```bash
# Reset network
podman network reload --all

# Check firewall rules
sudo firewall-cmd --list-all

# Enable IP forwarding
sudo sysctl net.ipv4.ip_forward=1
```

#### Storage Issues
```bash
# Clean up storage
podman system prune -a

# Check disk usage
podman system df

# Reset storage
podman system reset
```

### Debugging

#### Container Logs
```bash
# View logs
podman logs container_name

# Follow logs
podman logs -f container_name

# View last N lines
podman logs --tail 50 container_name
```

#### Inspect Resources
```bash
# Inspect container
podman inspect container_name

# Inspect image
podman image inspect image_name

# Get container stats
podman stats
```

#### Execute Commands
```bash
# Execute command in running container
podman exec -it container_name /bin/bash

# Execute as specific user
podman exec -u 1000 container_name whoami
```

## Best Practices

### Security
1. **Always run rootless** when possible
2. **Use specific image tags** instead of latest
3. **Scan images for vulnerabilities**: `podman scan image_name`
4. **Limit capabilities** with --cap-drop
5. **Use read-only containers** when possible: `--read-only`

### Performance
1. **Use overlay storage driver** for better performance
2. **Limit container resources**: `--memory`, `--cpus`
3. **Use volume mounts** instead of bind mounts
4. **Clean up unused resources** regularly
5. **Use multi-stage builds** to reduce image size

### Development Workflow
1. **Use podman-compose** for multi-container apps
2. **Generate Kubernetes YAML** for production
3. **Automate with systemd** for production services
4. **Use CI/CD integration** for automated builds
5. **Version your images** properly

## Comparison with Docker

| Feature | Podman | Docker |
|---------|--------|--------|
| Architecture | Daemonless | Daemon-based |
| Root requirement | Rootless by default | Requires root/sudo |
| Security | Better isolation | Standard isolation |
| Systemd integration | Native | Limited |
| Kubernetes pods | Native support | No native support |
| Docker compatibility | High | N/A |
| Resource usage | Lower | Higher |
| Windows support | WSL2 required | Native |

## Resources

### Official Documentation
- [Podman Documentation](https://docs.podman.io/)
- [Podman Desktop](https://podman-desktop.io/)
- [Container Registry](https://quay.io/)

### Community Resources
- [Podman GitHub](https://github.com/containers/podman)
- [Podman Blog](https://podman.io/blogs/)
- [Red Hat Developer](https://developers.redhat.com/topics/podman)

### Related Tools
- [Buildah](https://buildah.io/) - Build container images
- [Skopeo](https://github.com/containers/skopeo) - Work with container images
- [CRI-O](https://cri-o.io/) - Kubernetes container runtime

---

**Note**: This guide covers essential Podman features for development use. For production deployments, consult additional security and performance documentation.

*Created: 2025-08-23*
*Last updated: 2025-08-23*