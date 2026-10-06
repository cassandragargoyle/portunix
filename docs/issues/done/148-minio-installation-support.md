# Issue #148: MinIO Installation Support

**Type**: Enhancement
**Priority**: High
**Status**: ✅ Implemented
**Labels**: enhancement, package-management, minio, object-storage, s3-compatible, cross-platform, ptx-installer, go-install
**Author**: Architect
**Created**: 2026-02-10

---

## Overview

Add MinIO (S3-compatible object storage server) and MinIO Client (mc) as installable packages in ptx-installer. MinIO community edition is **source-only** since October 2025 — no pre-compiled binaries are provided. Installation requires building from source via `go install` or deploying as a container.

## Motivation

- MinIO is the leading S3-compatible object storage platform (49k+ GitHub stars, 1B+ Docker pulls)
- Essential for AI/ML data lakes, backup storage, and cloud-native development environments
- Provides local S3-compatible storage for development and testing (replaces dependency on AWS S3)
- Commonly used as backend for fulltext search (Elasticsearch), CI/CD artifacts, and data pipelines
- MinIO Client (mc) is a powerful CLI tool for object and file management across S3-compatible services
- Container-based deployment is the most common production pattern
- Complements existing Portunix packages: Elasticsearch (#142), Docker, Podman

## Requirements

### Functional Requirements

- [ ] FR-1: Install MinIO server via `portunix install minio` command
- [ ] FR-2: Install MinIO Client via `portunix install minio-client` command
- [ ] FR-3: Support `go install` build-from-source method (primary, requires Go prerequisite)
- [ ] FR-4: Support container-based deployment variant (Docker/Podman)
- [ ] FR-5: Cross-platform support (Linux x64/ARM64, Windows x64)
- [ ] FR-6: Post-installation verification via `minio --version` and `mc --version`
- [ ] FR-7: Go prerequisite auto-detection and guidance

### Non-Functional Requirements

- [ ] NFR-1: Follow existing ptx-installer package definition pattern
- [ ] NFR-2: Auto-discovery by registry (no manual index updates)
- [ ] NFR-3: Clearly communicate source-only nature and Go build requirement to user

## Technical Design

### Package Definitions

Create two files:

1. `src/helpers/ptx-installer/assets/packages/minio.json` — MinIO server
2. `src/helpers/ptx-installer/assets/packages/minio-client.json` — MinIO Client (mc)

Pattern: Go-based build from source (similar to `go install` pattern) + container variant

### Installation Methods — MinIO Server

| Platform | Method | Command/URL |
| -------- | ------ | ----------- |
| Linux/macOS | go install (recommended) | `go install github.com/minio/minio@latest` |
| Linux | Container (alternative) | `portunix container run --image quay.io/minio/minio:latest` |
| Windows | go install (recommended) | `go install github.com/minio/minio@latest` |
| Linux | Build from source | `git clone https://github.com/minio/minio.git && cd minio && go build -o minio .` |

### Installation Methods — MinIO Client (mc)

| Platform | Method | Command/URL |
| -------- | ------ | ----------- |
| Linux/macOS | go install (recommended) | `go install github.com/minio/mc@latest` |
| Windows | go install (recommended) | `go install github.com/minio/mc@latest` |
| Linux | Build from source | `git clone https://github.com/minio/mc.git && cd mc && go build -o mc .` |

### Download URLs / Source Repositories

```text
# MinIO Server (source-only since October 2025)
https://github.com/minio/minio

# MinIO Client
https://github.com/minio/mc

# AIStor container images (commercial, but freely pullable)
quay.io/minio/minio:latest

# go install commands
go install github.com/minio/minio@latest
go install github.com/minio/mc@latest
```

### Key Technical Details

- **Current version**: Rolling releases (date-based versioning, e.g. RELEASE.2025-10-15T17-29-55Z)
- **License**: GNU AGPLv3 (community edition)
- **Binary size**: ~80-100MB (server), ~30MB (mc client)
- **Build time**: 2-5 minutes depending on hardware
- **Dependencies**:
  - **Go 1.24+** (REQUIRED — build prerequisite)
  - **Git** (for `go install` module fetching)
  - Linux: libc (standard)
  - Windows: standard Go build chain
- **Default ports**: 9000 (S3 API), 9001 (Console — AIStor only)
- **Default credentials**: minioadmin / minioadmin
- **Install path (go install)**: `$GOPATH/bin/minio`, `$GOPATH/bin/mc`

### Variant Support

#### MinIO Server (`minio.json`)

| Variant | Description | Use Case |
| ------- | ----------- | -------- |
| `default` | `go install` from source | Development, local S3 storage |
| `source` | git clone + go build (manual) | Offline environments, custom builds |
| `container` | Container image deployment | Production, isolated environment |

#### MinIO Client (`minio-client.json`)

| Variant | Description | Use Case |
| ------- | ----------- | -------- |
| `default` | `go install` from source | CLI management of any S3-compatible storage |
| `source` | git clone + go build (manual) | Offline environments, custom builds |

### Prerequisites

```json
"dependencies": [
  {
    "name": "go",
    "minVersion": "1.24",
    "purpose": "Build from source (MinIO community edition is source-only)"
  }
]
```

### Environment Variables

```bash
# MinIO server configuration (runtime)
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

# Go install path (where binaries land)
GOPATH=$HOME/go
PATH=$GOPATH/bin:$PATH
```

### Verification

```bash
# Server
minio --version
# Expected output: minio version RELEASE.2025-xx-xxTxx-xx-xxZ (...)

# Client
mc --version
# Expected output: mc version RELEASE.2025-xx-xxTxx-xx-xxZ (...)

# Quick functional test
minio server /tmp/minio-data &
mc alias set local http://localhost:9000 minioadmin minioadmin
mc ls local
```

## Architecture Considerations

### Source-Only Distribution Model

Since October 2025, MinIO community edition provides **no pre-compiled binaries**. This has important implications:

1. **Go prerequisite is mandatory** — unlike most packages in ptx-installer, MinIO cannot be installed as a simple binary download
2. **Build time overhead** — `go install` takes 2-5 minutes (first build downloads all dependencies)
3. **Disk space** — Go module cache adds ~200-300MB during build
4. **Version pinning** — `@latest` gets the latest tag; specific versions possible via `@RELEASE.2025-xx-xx...`

### Two Separate Packages

MinIO server and MinIO Client are separate projects with independent release cycles. They should be two separate package definitions:

1. **`minio`** — The S3-compatible object storage server
2. **`minio-client`** — The `mc` CLI tool for managing S3-compatible storage

Rationale: Users may want only `mc` client (to manage remote S3/MinIO) without running a local server.

### Container Installation (variant `container`)

The container variant follows the established pattern from Elasticsearch (#142) — `type: "container"` with full lifecycle management. MinIO is one of the most commonly container-deployed services (1B+ Docker Hub pulls).

**Container image**: `quay.io/minio/minio` (AIStor image — freely pullable, includes console UI)

#### Container Configuration

```yaml
image: quay.io/minio/minio:latest
command: server /data --console-address ":9001"
name: portunix-minio
ports:
  - "9000:9000"    # S3 API
  - "9001:9001"    # Web Console (AIStor)
environment:
  MINIO_ROOT_USER: minioadmin
  MINIO_ROOT_PASSWORD: minioadmin
volumes:
  - minio-data:/data
detach: true
healthCheck:
  endpoint: http://localhost:9000/minio/health/live
  timeout: 60
  interval: 5
  retries: 12
```

#### Container Package Definition (in `minio.json`)

The `container` variant should be defined for all platforms (linux, darwin, windows) following the Elasticsearch pattern:

```json
"container": {
  "version": "latest",
  "description": "MinIO S3-compatible object storage as container service",
  "container": {
    "image": "quay.io/minio/minio",
    "tag": "latest",
    "name": "portunix-minio",
    "command": ["server", "/data", "--console-address", ":9001"],
    "ports": ["9000:9000", "9001:9001"],
    "environment": {
      "MINIO_ROOT_USER": "minioadmin",
      "MINIO_ROOT_PASSWORD": "minioadmin"
    },
    "volumes": ["minio-data:/data"],
    "detach": true,
    "healthCheck": {
      "endpoint": "http://localhost:9000/minio/health/live",
      "timeout": 60,
      "interval": 5,
      "retries": 12
    }
  }
}
```

#### Container Variant Features

- **Persistent data volume** (`minio-data:/data`) — data survives container restarts
- **Web Console** on port 9001 — browser-based management UI (AIStor feature)
- **S3 API** on port 9000 — standard S3-compatible endpoint
- **Health check** via MinIO's built-in `/minio/health/live` endpoint
- **Default credentials** — minioadmin/minioadmin (user should change for production)
- **Compatible with Docker and Podman** — uses Portunix container abstraction

#### Container Verification

```json
"verification": {
  "command": "curl -s http://localhost:9000/minio/health/live",
  "expectedExitCode": 0
}
```

#### Container CLI Usage

```bash
# Install MinIO as container service
portunix install minio --variant container

# Manage via portunix container commands
portunix container list          # should show portunix-minio
portunix container stop portunix-minio
portunix container start portunix-minio
portunix container rm portunix-minio

# Access
# S3 API:      http://localhost:9000
# Web Console: http://localhost:9001 (login: minioadmin / minioadmin)

# Use mc client to interact
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/my-bucket
mc cp file.txt local/my-bucket/
```

#### Container vs go install — Decision Matrix

| Aspect | `default` (go install) | `container` |
| ------ | ---------------------- | ----------- |
| Prerequisites | Go 1.24+ | Docker or Podman |
| Build time | 2-5 min | ~30s (pull) |
| Disk usage | ~300MB (build cache) | ~200MB (image) |
| Web Console | No (community) | Yes (AIStor) |
| Data isolation | Local filesystem | Named volume |
| Best for | CLI tools, scripts | Services, dev environments |
| Production | Not recommended | Recommended |

### mc Binary Name Conflict

The MinIO Client binary is named `mc` which may conflict with Midnight Commander on some Linux systems. The package definition should note this in aiPrompts.

## Implementation Scope

### Files to Create

| File | Description |
| ---- | ----------- |
| `src/helpers/ptx-installer/assets/packages/minio.json` | MinIO server package definition |
| `src/helpers/ptx-installer/assets/packages/minio-client.json` | MinIO Client (mc) package definition |

### Files to Modify

None required — registry auto-discovers packages from `assets/packages/` directory.

## Acceptance Criteria

### go install variant (default)

1. `portunix install minio` successfully builds and installs MinIO server (via `go install`)
2. `portunix install minio-client` successfully builds and installs MinIO Client mc (via `go install`)
3. Go prerequisite is detected and user is informed if Go is missing or version is insufficient
4. `minio --version` returns expected version after installation
5. `mc --version` returns expected version after installation
6. Both packages appear in `portunix install --list` output
7. Installation tested in clean container environment with Go pre-installed

### Container variant

8. `portunix install minio --variant container` pulls MinIO image and starts container service
9. Container is accessible on port 9000 (S3 API) and 9001 (Web Console)
10. Health check passes before returning success (`/minio/health/live`)
11. Data persists across container restarts (named volume `minio-data`)
12. Works with both Docker and Podman (via Portunix container abstraction)
13. Container appears in `portunix container list` as `portunix-minio`
14. `mc` client can connect to container instance and perform basic operations (create bucket, upload, list)

## Testing

All installation testing MUST be performed in containers per project methodology:

```bash
# Linux testing (Ubuntu) — prerequisite: Go must be installed first
portunix container run --image ubuntu:22.04
# inside container:
apt update && apt install -y curl git
./portunix install go          # install Go prerequisite
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
./portunix install minio
minio --version

# MinIO Client testing
./portunix install minio-client
mc --version

# Functional test (server + client)
minio server /tmp/minio-test &
sleep 2
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/test-bucket
mc ls local
# cleanup
mc rb local/test-bucket --force
kill %1

# ──────────────────────────────────────────────
# Container variant testing
# ──────────────────────────────────────────────

# Install MinIO as container service
./portunix install minio --variant container
# Expected: pulls quay.io/minio/minio:latest, starts portunix-minio container

# Verify container is running
portunix container list
# Expected: portunix-minio in list, status running

# Verify health check
curl -s http://localhost:9000/minio/health/live
# Expected: HTTP 200

# Verify Web Console is accessible
curl -s -o /dev/null -w "%{http_code}" http://localhost:9001
# Expected: 200 or 302 (redirect to login)

# Verify S3 API with mc client
./portunix install minio-client    # install mc first (go install)
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/test-bucket
mc cp /etc/hostname local/test-bucket/
mc ls local/test-bucket
mc rb local/test-bucket --force

# Verify data persistence
portunix container stop portunix-minio
portunix container start portunix-minio
# data in minio-data volume should survive restart

# Cleanup
portunix container rm portunix-minio

# ──────────────────────────────────────────────
# Container variant — Podman compatibility
# ──────────────────────────────────────────────

# On system with Podman instead of Docker
./portunix install minio --variant container
# Expected: same behavior, uses Podman transparently

# ──────────────────────────────────────────────
# Negative tests
# ──────────────────────────────────────────────

# Without Go installed (go install variant)
portunix container run --image ubuntu:22.04
./portunix install minio
# Expected: clear error message about missing Go prerequisite

# Without container runtime (container variant)
# On system without Docker/Podman
./portunix install minio --variant container
# Expected: clear error about missing container runtime
```

## Reference

- Manifest: `portunix-architecture/docs/manifests/manifest-minio.md`
- MinIO website: <https://min.io/>
- Source code: <https://github.com/minio/minio>
- MinIO Client: <https://github.com/minio/mc>
- Documentation: <https://docs.min.io/>
- Container image: `quay.io/minio/minio:latest`
- Related issues: #142 (Elasticsearch — complementary infrastructure)

## Complexity

**Medium-High** — Combines two installation patterns: (1) `go install` build-from-source, which is new for ptx-installer and requires Go prerequisite handling, and (2) container service deployment following the Elasticsearch pattern (#142) with health checks, persistent volumes, and lifecycle management. Two separate package definitions (minio + minio-client), cross-platform support, and the `mc` binary name conflict with Midnight Commander add to scope.

---

*Created: 2026-02-10*
