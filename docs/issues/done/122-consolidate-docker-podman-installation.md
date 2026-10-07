# Issue #122: Consolidate Docker/Podman Installation into ptx-installer

## Summary

Current Docker/Podman installation logic is scattered across multiple locations (`app/docker/`, `cmd/docker.go`). This needs to be consolidated into `ptx-installer` helper binary while preserving existing functionality like storage selection logic.

## Problem Statement

1. **Fragmented architecture**: Docker installation code is in `app/docker/docker.go` (main binary)
2. **Inconsistent commands**: `portunix docker install` vs expected `portunix install docker`
3. **Code duplication risk**: Storage selection, OS detection duplicated between helpers
4. **Maintenance burden**: Fixes need to be applied in multiple places

## Current State

```text
app/docker/docker.go          - Docker installation logic, storage selection
cmd/docker.go                 - Docker command handlers
src/helpers/ptx-installer/    - Package installer (doesn't handle Docker/Podman)
```

## Target State

```text
src/helpers/ptx-installer/
├── docker.go                 - Docker installation (moved from app/docker)
├── podman.go                 - Podman installation
├── storage_windows.go        - Windows storage selection logic
└── storage_linux.go          - Linux storage selection logic

Commands:
- portunix install docker     - Install Docker (delegates to ptx-installer)
- portunix install podman     - Install Podman (delegates to ptx-installer)
```

## Requirements

### Functional Requirements

1. `portunix install docker` - Install Docker with intelligent storage selection
2. `portunix install podman` - Install Podman
3. Preserve existing storage selection logic (non-C drives with sufficient space first)
4. Preserve minimum 10 GB space requirement for Docker
5. Support `--dry-run` flag
6. Support Windows and Linux platforms

### Technical Requirements

1. Move `analyzeWindowsStorage()` from `app/docker/docker.go` to `ptx-installer`
2. Move `parseSpaceString()` utility to shared location
3. Update `portunix docker install` to delegate to `portunix install docker`
4. Deprecate direct Docker installation via `portunix docker install`
5. Add Docker/Podman package definitions to `assets/install-packages.json`

## Migration Plan

### Phase 1: Add to ptx-installer

- [ ] Create `docker.go` in ptx-installer with installation logic
- [ ] Create `podman.go` in ptx-installer
- [ ] Move storage selection logic to ptx-installer
- [ ] Add Docker/Podman to install-packages.json

### Phase 2: Update Commands

- [ ] Add `docker` and `podman` to `portunix install` command
- [ ] Make `portunix docker install` delegate to `portunix install docker`
- [ ] Add deprecation warning to old command

### Phase 3: Cleanup

- [ ] Remove duplicated code from `app/docker/`
- [ ] Update documentation
- [ ] Update tests

## Acceptance Criteria

- [ ] `portunix install docker` works on Windows with correct storage selection
- [ ] `portunix install docker` works on Linux
- [ ] `portunix install podman` works on both platforms
- [ ] Storage selection ignores drives with < 10 GB
- [ ] `--dry-run` shows what would be installed
- [ ] Old `portunix docker install` still works (with deprecation warning)

## Related Issues

- #119 - PTX-Ansible (discovered storage selection bug during testing)

## Priority

High

## Labels

refactoring, architecture, ptx-installer, docker, podman

---
**Created**: 2026-01-03
**Status**: Implemented
**Implemented**: 2026-01-03
**Branch**: feature/122-consolidate-docker-podman-installer
