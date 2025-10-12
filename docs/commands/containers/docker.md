# Portunix Docker Command

## Quick Start

The `docker` command provides complete Docker container management through Portunix, offering simplified interfaces for common Docker operations with enhanced functionality.

### Simplest Usage
```bash
# Run a Ubuntu container
portunix docker run ubuntu

# List running containers
portunix docker list

# SSH into a container
portunix docker ssh my-container
```

### Basic Syntax
```bash
portunix docker [subcommand] [options]
```

### Common Subcommands
- `run` - Create and run a container
- `list` - List containers
- `ssh` - SSH into a container
- `exec` - Execute command in container
- `stop` - Stop container
- `remove` - Remove container
- `logs` - View container logs
- `install` - Install Docker if not present

## Intermediate Usage

### Running Containers

Create and run containers with various configurations:

```bash
# Basic container
portunix docker run ubuntu

# Named container
portunix docker run ubuntu --name my-dev-env

# With port mapping
portunix docker run nginx --ports 8080:80

# With volume mounting
portunix docker run ubuntu --volumes /host/path:/container/path

# With environment variables
portunix docker run node --env NODE_ENV=development

# Interactive container
portunix docker run ubuntu --interactive

# Background container
portunix docker run ubuntu --detach
```

### Container with Pre-installed Software

Portunix can create containers with pre-installed software packages:

```bash
# Container with Node.js
portunix docker run-in-container nodejs --image ubuntu:22.04

# Container with Python development environment
portunix docker run-in-container python --variant full --image debian:bookworm

# Container with multiple packages
portunix docker run-in-container "nodejs python java" --image ubuntu:22.04
```

### SSH Access

Built-in SSH access to containers:

```bash
# SSH into running container
portunix docker ssh my-container

# SSH with specific user
portunix docker ssh my-container --user developer

# SSH with custom key
portunix docker ssh my-container --key ~/.ssh/my-key

# SSH tunnel with port forwarding
portunix docker ssh my-container --tunnel 5432:localhost:5432
```

### Container Management

```bash
# List all containers
portunix docker list

# List only running containers
portunix docker list --running

# List with detailed information
portunix docker list --detailed

# Stop container
portunix docker stop my-container

# Start existing container
portunix docker start my-container

# Restart container
portunix docker restart my-container

# Remove container
portunix docker remove my-container

# Remove with volumes
portunix docker remove my-container --volumes
```

### Log Management

```bash
# View logs
portunix docker logs my-container

# Follow logs in real-time
portunix docker logs my-container --follow

# Last N lines
portunix docker logs my-container --tail 100

# Logs with timestamps
portunix docker logs my-container --timestamps

# Logs since specific time
portunix docker logs my-container --since 2024-01-15T10:00:00
```

## Advanced Usage

### Docker Installation and Setup

Automatic Docker installation:

```bash
# Install Docker if not present
portunix docker install

# Install specific Docker version
portunix docker install --version 24.0.0

# Install Docker with specific configuration
portunix docker install --config docker-daemon.json

# Verify Docker installation
portunix docker verify
```

Configuration example (`docker-daemon.json`):
```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "insecure-registries": ["registry.company.com:5000"]
}
```

### Container Templates

Pre-configured container templates:

```bash
# Development environment template
portunix docker template create dev-env \
  --base ubuntu:22.04 \
  --packages "nodejs python git" \
  --user developer \
  --shell zsh

# Use template
portunix docker run --template dev-env --name my-dev

# List available templates
portunix docker template list

# Export template
portunix docker template export dev-env > dev-env-template.yaml
```

### Multi-Platform Support

Cross-platform container support:

```bash
# Run ARM container on x86
portunix docker run ubuntu:22.04 --platform linux/arm64

# List supported platforms
portunix docker platforms

# Build multi-platform images
portunix docker build --platforms linux/amd64,linux/arm64 .
```

### Container Networking

Advanced networking configurations:

```bash
# Create custom network
portunix docker network create my-network

# Run container in custom network
portunix docker run ubuntu --network my-network

# Connect container to additional network
portunix docker network connect my-network my-container

# List networks
portunix docker network list

# Inspect network
portunix docker network inspect my-network
```

### Volume Management

Persistent storage management:

```bash
# Create named volume
portunix docker volume create my-data

# Run container with named volume
portunix docker run ubuntu --volume my-data:/data

# List volumes
portunix docker volume list

# Backup volume
portunix docker volume backup my-data --output backup.tar.gz

# Restore volume
portunix docker volume restore my-data --input backup.tar.gz

# Clean unused volumes
portunix docker volume prune
```

### Container Monitoring

Real-time container monitoring:

```bash
# Monitor all containers
portunix docker monitor

# Monitor specific container
portunix docker monitor my-container

# Resource usage statistics
portunix docker stats

# Container health check
portunix docker health my-container

# Performance metrics
portunix docker metrics my-container --duration 60s
```

### Image Management

Container image operations:

```bash
# List images
portunix docker images

# Pull image
portunix docker pull ubuntu:22.04

# Remove image
portunix docker rmi ubuntu:22.04

# Build image from Dockerfile
portunix docker build . --tag my-app:latest

# Export image
portunix docker save my-app:latest --output my-app.tar

# Import image
portunix docker load --input my-app.tar

# Clean unused images
portunix docker image prune
```

### Registry Operations

Container registry management:

```bash
# Login to registry
portunix docker login registry.company.com

# Push image to registry
portunix docker push my-app:latest

# Tag image for registry
portunix docker tag my-app:latest registry.company.com/my-app:v1.0.0

# Search registry
portunix docker search nodejs

# Registry configuration
portunix docker registry config --url https://registry.company.com
```

## Expert Tips & Tricks

### 1. Development Workflow

```bash
# Development container with live reload
portunix docker run-dev nodejs \
  --volumes $(pwd):/workspace \
  --working-dir /workspace \
  --ports 3000:3000 \
  --auto-restart

# Debug container
portunix docker debug my-container \
  --attach-debugger \
  --expose-debug-port 9229
```

### 2. Container Orchestration

```bash
# Docker Compose integration
portunix docker compose up --file docker-compose.yml

# Swarm mode
portunix docker swarm init
portunix docker service create --name web nginx

# Stack deployment
portunix docker stack deploy --compose-file docker-compose.yml myapp
```

### 3. Security Features

```bash
# Run with security profiles
portunix docker run ubuntu --security-profile restricted

# Scan for vulnerabilities
portunix docker scan my-image:latest

# Run as non-root user
portunix docker run ubuntu --user 1000:1000 --no-root

# Read-only filesystem
portunix docker run ubuntu --read-only --tmpfs /tmp
```

### 4. Performance Optimization

```bash
# Resource constraints
portunix docker run ubuntu \
  --memory 512m \
  --cpus 0.5 \
  --swap 256m

# Container startup optimization
portunix docker run ubuntu --fast-start --preload-libs

# Image layer optimization
portunix docker optimize-image my-app:latest
```

### 5. Backup and Migration

```bash
# Backup complete container
portunix docker backup my-container --include-volumes

# Export container configuration
portunix docker export-config my-container > container.json

# Migrate container to another host
portunix docker migrate my-container --to user@remote-host

# Clone container
portunix docker clone my-container --name my-container-copy
```

## Container Lifecycle Management

### Automated Lifecycle

```bash
# Auto-remove after exit
portunix docker run ubuntu --rm --command "echo 'Hello World'"

# Auto-restart policies
portunix docker run nginx --restart always
portunix docker run app --restart on-failure:3

# Health checks
portunix docker run app \
  --health-cmd "curl -f http://localhost:8080/health" \
  --health-interval 30s \
  --health-timeout 10s
```

### Container Updates

```bash
# Update container image
portunix docker update my-container --image ubuntu:24.04

# Rolling update
portunix docker rolling-update my-app:v1.0.0 my-app:v2.0.0

# Blue-green deployment
portunix docker blue-green-deploy my-app:v2.0.0
```

## Integration with Portunix Features

### Plugin Integration

```bash
# Run container with plugin support
portunix docker run ubuntu --enable-plugins

# Install packages in container via plugin
portunix docker exec my-container portunix install nodejs
```

### MCP Server Integration

```bash
# Container with MCP server access
portunix docker run ubuntu --mcp-server-access

# Run MCP-enabled development environment
portunix docker run-mcp-dev --image ubuntu:22.04
```

### Sandbox Integration

```bash
# Container in sandbox mode
portunix docker run ubuntu --sandbox

# Isolated development environment
portunix docker run-isolated dev-env
```

## Troubleshooting

### Common Issues

#### 1. Docker Not Running
```bash
# Start Docker service
portunix docker service start

# Check Docker status
portunix docker service status

# Install Docker if missing
portunix docker install
```

#### 2. Permission Denied
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER

# Use rootless Docker
portunix docker install --rootless

# Run with sudo (temporary)
sudo portunix docker run ubuntu
```

#### 3. Container Won't Start
```bash
# Check container logs
portunix docker logs my-container

# Inspect container configuration
portunix docker inspect my-container

# Debug container startup
portunix docker debug my-container --startup
```

#### 4. Network Issues
```bash
# Check container networking
portunix docker network diagnose

# Reset Docker networking
portunix docker network reset

# Use host networking
portunix docker run ubuntu --network host
```

### Debug Mode

```bash
# Verbose Docker operations
portunix docker run ubuntu --verbose

# Debug container creation
portunix docker debug --create ubuntu

# Trace Docker API calls
portunix docker --trace run ubuntu
```

## Performance Monitoring

### Resource Usage

```bash
# Real-time resource monitoring
portunix docker stats --live

# Historical resource usage
portunix docker stats --history 24h

# Export metrics
portunix docker metrics --export prometheus
```

### Performance Tuning

```bash
# Optimize container performance
portunix docker tune my-container

# Analyze performance bottlenecks
portunix docker analyze-performance my-container

# Container resource recommendations
portunix docker recommend-resources my-container
```

## Security Best Practices

### Container Security

```bash
# Security scan
portunix docker security-scan my-container

# Compliance check
portunix docker compliance-check my-container

# Vulnerability assessment
portunix docker vulnerability-scan my-image:latest
```

### Secret Management

```bash
# Mount secrets securely
portunix docker run app --secret database-password

# Environment variables from file
portunix docker run app --env-file secrets.env

# Use Docker secrets
portunix docker secret create db-password password.txt
portunix docker run app --secret db-password
```

## API Integration

### REST API

```bash
# Container operations via API
curl -X POST http://localhost:8080/api/docker/run \
  -H "Content-Type: application/json" \
  -d '{"image": "ubuntu:22.04", "name": "my-container"}'
```

### gRPC API

```go
// Go client example
client := portunix.NewClient()
container, err := client.Docker.Run(context.Background(), &DockerRunRequest{
    Image: "ubuntu:22.04",
    Name:  "my-container",
})
```

### Event Streaming

```bash
# Subscribe to Docker events
portunix docker events --format json

# Webhook notifications
portunix docker events --webhook https://hooks.slack.com/xxx
```

## Platform-Specific Features

### Windows

```bash
# Windows containers
portunix docker run mcr.microsoft.com/windows/servercore:ltsc2022

# Linux containers on Windows
portunix docker run ubuntu --linux

# Hyper-V isolation
portunix docker run windows --isolation hyperv
```

### Linux

```bash
# Cgroup management
portunix docker run ubuntu --cgroup-driver systemd

# SELinux support
portunix docker run ubuntu --selinux-label container_t

# Systemd integration
portunix docker run ubuntu --systemd
```

### macOS

```bash
# Docker Desktop integration
portunix docker desktop start

# Resource limits on macOS
portunix docker run ubuntu --memory 2g --cpus 2
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORTUNIX_DOCKER_HOST` | Docker daemon host | unix:///var/run/docker.sock |
| `PORTUNIX_DOCKER_API_VERSION` | Docker API version | auto |
| `PORTUNIX_DOCKER_TIMEOUT` | Docker operation timeout | 60s |
| `PORTUNIX_DOCKER_REGISTRY` | Default registry | docker.io |
| `PORTUNIX_DOCKER_NAMESPACE` | Default namespace | portunix |

## Related Commands

- [`container`](container.md) - Universal container interface
- [`podman`](podman.md) - Podman container management
- [`install`](install.md) - Install Docker
- [`system`](system.md) - System information

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--name` | string | auto | Container name |
| `--image` | string | - | Container image |
| `--ports` | string | - | Port mappings (host:container) |
| `--volumes` | string | - | Volume mounts (host:container) |
| `--env` | string | - | Environment variables |
| `--network` | string | `bridge` | Network mode |
| `--user` | string | `root` | User inside container |
| `--working-dir` | string | `/` | Working directory |
| `--detach` | boolean | `false` | Run in background |
| `--interactive` | boolean | `false` | Interactive mode |
| `--rm` | boolean | `false` | Auto-remove after exit |
| `--restart` | string | `no` | Restart policy |
| `--memory` | string | - | Memory limit |
| `--cpus` | string | - | CPU limit |
| `--platform` | string | auto | Target platform |
| `--security-profile` | string | `default` | Security profile |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Docker not found |
| 3 | Permission denied |
| 4 | Image not found |
| 5 | Container creation failed |
| 6 | Container start failed |
| 7 | Network error |
| 8 | Volume error |
| 9 | Resource limit exceeded |
| 10 | Timeout |

## Version History

- **v1.5.0** - Added SSH integration
- **v1.4.0** - Implemented container templates
- **v1.3.0** - Added multi-platform support
- **v1.2.0** - Enhanced security features
- **v1.1.0** - Added monitoring capabilities
- **v1.0.0** - Initial Docker integration