# Docker vs Podman - Container Runtime Comparison

## Overview

Portunix supports both Docker and Podman as container runtimes. This document explains the differences, how compose tools work with each, and how Portunix abstracts these differences.

## What is a Container?

**Container** is a lightweight, isolated environment for running applications. Unlike virtual machines, containers share the host operating system's kernel, making them faster to start and more resource-efficient.

```
┌─────────────────────────────────────────────────────┐
│                    Host OS                          │
├─────────────────────────────────────────────────────┤
│              Container Runtime                       │
│         (Docker, Podman, containerd)                │
├─────────────┬─────────────┬─────────────────────────┤
│ Container 1 │ Container 2 │ Container 3             │
│ ┌─────────┐ │ ┌─────────┐ │ ┌─────────┐             │
│ │  App    │ │ │  App    │ │ │  App    │             │
│ │  Libs   │ │ │  Libs   │ │ │  Libs   │             │
│ └─────────┘ │ └─────────┘ │ └─────────┘             │
└─────────────┴─────────────┴─────────────────────────┘
```

**Key benefits:**
- **Isolation**: Each container has its own filesystem, network, and processes
- **Portability**: "Works on my machine" → "Works everywhere"
- **Reproducibility**: Same container = same environment every time
- **Lightweight**: Shares kernel with host, starts in seconds

## What is Compose?

**Compose** is a tool for defining and running **multi-container applications**. Instead of starting each container manually with long command lines, you define your entire application stack in a single YAML file.

### Why Compose Exists

Without Compose (manual approach):
```bash
# Start database
docker run -d --name db \
  -e POSTGRES_PASSWORD=secret \
  -v db-data:/var/lib/postgresql/data \
  postgres:15

# Start backend (needs database)
docker run -d --name backend \
  --link db:db \
  -e DATABASE_URL=postgres://db:5432/app \
  my-backend:latest

# Start frontend (needs backend)
docker run -d --name frontend \
  -p 8080:80 \
  --link backend:backend \
  my-frontend:latest
```

**Problems:**
- Long, error-prone commands
- Must remember correct order
- Hard to share with team
- Manual cleanup needed

With Compose (declarative approach):
```yaml
# docker-compose.yml
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret
    volumes:
      - db-data:/var/lib/postgresql/data

  backend:
    image: my-backend:latest
    environment:
      DATABASE_URL: postgres://db:5432/app
    depends_on:
      - db

  frontend:
    image: my-frontend:latest
    ports:
      - "8080:80"
    depends_on:
      - backend

volumes:
  db-data:
```

Then just run:
```bash
portunix container compose up -d    # Start everything
portunix container compose down     # Stop and cleanup
```

### Compose Features

| Feature | Description |
|---------|-------------|
| **Services** | Define multiple containers as services |
| **Networks** | Automatic networking between services |
| **Volumes** | Persistent data storage |
| **Dependencies** | `depends_on` ensures correct startup order |
| **Environment** | Easy environment variable management |
| **Ports** | Expose services to host |
| **Scaling** | Run multiple instances of a service |
| **Health checks** | Ensure services are ready before dependents start |

### Real-World Example: Web Application with Database

```yaml
# Typical web application stack
services:
  # PostgreSQL database
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: myapp
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "myapp"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis cache
  cache:
    image: redis:7-alpine
    volumes:
      - redis-data:/data

  # Application server
  app:
    build: .
    environment:
      DATABASE_URL: postgres://myapp:${DB_PASSWORD}@db:5432/myapp
      REDIS_URL: redis://cache:6379
    depends_on:
      db:
        condition: service_healthy
      cache:
        condition: service_started
    ports:
      - "3000:3000"

volumes:
  postgres-data:
  redis-data:
```

## Quick Comparison

| Feature | Docker | Podman |
|---------|--------|--------|
| Architecture | Client-server (daemon) | Daemonless |
| Root access | Requires root daemon | Rootless by default |
| Compose | Docker Compose (built-in V2) | podman compose / podman-compose |
| Images | OCI-compatible | OCI-compatible |
| Kubernetes | docker-compose only | Native pod support |
| Resource usage | Higher (daemon always running) | Lower (on-demand) |
| Security | Daemon runs as root | Rootless containers |
| Systemd | Separate service | User-level socket |

## Architecture Differences

### Docker Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────>│   Daemon    │────>│  Container  │
│ (docker CLI)│     │ (dockerd)   │     │  Runtime    │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    Runs as root
                    Always running
```

- **Daemon-based**: Docker requires `dockerd` daemon running continuously
- **Root privileges**: Daemon runs as root (security consideration)
- **Centralized**: All containers managed by single daemon
- **Service**: `systemctl start docker`

### Podman Architecture

```
┌─────────────┐     ┌─────────────┐
│   Client    │────>│  Container  │
│ (podman CLI)│     │  Runtime    │
└─────────────┘     └─────────────┘
        │
   Direct fork
   No daemon
```

- **Daemonless**: Each container is a child process of podman command
- **Rootless**: Runs as regular user by default
- **Decentralized**: No single point of failure
- **Socket**: `systemctl --user start podman.socket` (for API compatibility)

## Compose Tools

### Docker Compose

Docker Compose manages multi-container applications using YAML configuration.

**Versions:**
- **V1** (deprecated): Standalone Python tool (`docker-compose`)
- **V2** (current): Built into Docker CLI (`docker compose`)

```bash
# Docker Compose V2 (recommended)
docker compose up -d
docker compose ps
docker compose down

# Docker Compose V1 (legacy)
docker-compose up -d
```

### Podman Compose

Podman offers two compose solutions:

**1. Built-in `podman compose` (Podman 3.0+)**
- Native integration with Podman
- Uses podman-compose under the hood or custom implementation
- Recommended for most users

```bash
podman compose up -d
podman compose ps
podman compose down
```

**2. Standalone `podman-compose`**
- Python-based tool
- Works with older Podman versions
- Install: `pip install podman-compose`

```bash
podman-compose up -d
podman-compose ps
podman-compose down
```

### Compose Compatibility

Both Docker Compose and Podman Compose use the same YAML format (docker-compose.yml):

```yaml
version: "3.8"
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret
```

## Portunix Abstraction

### Why Abstraction Matters

Portunix provides a unified interface for container operations, abstracting the differences between Docker and Podman:

```bash
# Always use portunix container commands
portunix container compose up -d
portunix container compose ps
portunix container compose down

# NEVER use docker/podman directly in scripts
# docker compose up -d      # WRONG
# podman compose up -d      # WRONG
```

### How It Works

The `ptx-container` helper automatically:

1. **Detects available runtime** (Docker or Podman)
2. **Checks if daemon/socket is running**
3. **Selects appropriate compose tool**
4. **Executes command with correct binary**

```
portunix container compose up
        │
        ▼
┌───────────────────────────────────────────┐
│            ptx-container                   │
│                                           │
│  1. Check Docker (docker info)            │
│     └─> If OK: use docker compose         │
│                                           │
│  2. Check Podman (podman info)            │
│     └─> If OK: try podman compose         │
│         └─> Fallback: podman-compose      │
│                                           │
└───────────────────────────────────────────┘
```

### Detection Priority

1. **Docker with running daemon** → Docker Compose V2
2. **Docker Compose V1** (if V2 not available)
3. **Podman with running socket** → podman compose
4. **podman-compose** (standalone)

## Daemon/Socket Management

### Docker Daemon

```bash
# Start Docker daemon
sudo systemctl start docker

# Enable on boot
sudo systemctl enable docker

# Check status
sudo systemctl status docker
```

### Podman Socket

Podman doesn't require a daemon, but compose tools need the API socket:

```bash
# Start Podman socket (user-level)
systemctl --user start podman.socket

# Enable on login
systemctl --user enable podman.socket

# Check status
systemctl --user status podman.socket
```

## Detection via Portunix

Use `portunix system info` to see container runtime status:

```bash
$ portunix system info

Container Runtimes:
Docker:       not installed
Podman:       v5.4.1 (running)

Compose Support:
Type:         Podman Compose v2.39.2 (ready)
```

Possible states:
- `(running)` / `(ready)` - Runtime and compose ready to use
- `(daemon not running)` - Docker installed but daemon not started
- `(socket not running)` - Podman installed but socket not started

## Troubleshooting

### Docker: Daemon Not Running

```bash
# Check daemon status
sudo systemctl status docker

# Start daemon
sudo systemctl start docker

# Common issue: Permission denied
# Solution: Add user to docker group
sudo usermod -aG docker $USER
# Then logout and login again
```

### Podman: Socket Not Running

```bash
# Check socket status
systemctl --user status podman.socket

# Start socket
systemctl --user start podman.socket

# Enable for future sessions
systemctl --user enable podman.socket

# If socket still not working, check XDG_RUNTIME_DIR
echo $XDG_RUNTIME_DIR
# Should be something like /run/user/1000
```

### Compose Not Found

```bash
# Docker: Install Docker Desktop or docker-compose-plugin
sudo apt install docker-compose-plugin

# Podman: Install podman-compose
pip install podman-compose
# OR use built-in (Podman 3.0+)
podman compose version
```

## Migration: Docker to Podman

### Image Compatibility

Both use OCI-compatible images. Docker images work with Podman:

```bash
# Pull Docker Hub image with Podman
podman pull docker.io/nginx:latest

# Run same image
podman run -d -p 8080:80 nginx:latest
```

### Compose File Compatibility

Most docker-compose.yml files work with Podman Compose without changes:

```bash
# Replace docker-compose with podman-compose
podman-compose up -d

# Or use Portunix abstraction (recommended)
portunix container compose up -d
```

### Known Differences

1. **Networking**: Podman uses different default network configuration
2. **Volumes**: Host path volumes may need `:Z` suffix for SELinux
3. **Build context**: Some build features differ slightly
4. **Named pipes**: Windows-specific features may not work

## Best Practices

### For Developers

1. **Use Portunix abstraction** for all container operations in scripts
2. **Test with both runtimes** when possible
3. **Document runtime requirements** if using specific features

### For Scripts

```bash
# CORRECT: Use portunix container
portunix container compose up -d

# WRONG: Direct docker/podman calls
docker compose up -d
podman compose up -d
```

### For CI/CD

```yaml
# GitHub Actions example
- name: Run containers
  run: |
    portunix container compose up -d
    portunix container compose ps
```

## References

- [Docker Documentation](https://docs.docker.com/)
- [Podman Documentation](https://podman.io/docs/)
- [Docker Compose](https://docs.docker.com/compose/)
- [podman-compose](https://github.com/containers/podman-compose)

## See Also

- [PTX-Container Helper](../src/helpers/ptx-container/README.md)
- [Issue #107 - PTX-PFT](issues/internal/107-ptx-pft-product-feedback-tool-helper.md)
