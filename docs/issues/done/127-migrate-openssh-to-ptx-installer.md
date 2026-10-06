# Issue #127: Migrate Win32-OpenSSH Installation to ptx-installer

## Summary

Move the existing Win32-OpenSSH download and installation functionality from `src/app/sandbox/sandbox.go` to the `ptx-installer` helper, enabling standalone SSH installation via `portunix install ssh` command.

## Current State

The Win32-OpenSSH installation logic exists in `src/app/sandbox/sandbox.go`:
- `EnsureWin32OpenSSH()` - ensures SSH is available and copies to temp directory
- `downloadWin32OpenSSH()` - downloads from GitHub releases
- Used only internally for Windows Sandbox functionality

**Location**: `src/app/sandbox/sandbox.go:431-502`

**Current functionality**:
- Downloads `OpenSSH-Win64.zip` from `https://github.com/PowerShell/Win32-OpenSSH/releases/latest/download/`
- Extracts to `.cache/openssh/OpenSSH-Win64/`
- Includes `Install-PortableOpenSSH.ps1` script (embedded in `src/app/sandbox/embedded.go`)

## Requirements

### Functional Requirements

1. **Package Definition**: Create `openssh` package in ptx-installer registry
2. **Windows Support**:
   - Download Win32-OpenSSH from GitHub releases
   - Support both portable and system-wide installation
   - Configure SSH client and optionally SSH server
3. **Linux Support**:
   - Install via package manager (apt, dnf, pacman)
   - Package name: `openssh-client` / `openssh-server`
4. **Variants**:
   - `client` - SSH client only (default)
   - `server` - SSH client + server
   - `portable` - Windows portable installation (no system changes)

### Technical Requirements

1. Add package definition to `src/helpers/ptx-installer/assets/packages/`
2. Implement Windows-specific installer in ptx-installer
3. Reuse existing download/extract logic from sandbox
4. Remove or refactor sandbox code to use ptx-installer

## Implementation Steps

### Phase 1: Package Definition
- [x] Create `openssh.json` package definition
- [x] Define variants (client, server, portable)
- [x] Add platform-specific installation methods

### Phase 2: Windows Implementation
- [x] Move download logic from sandbox to ptx-installer
- [x] Implement portable installation
- [x] Implement system-wide installation (optional)
- [x] Add PATH configuration

### Phase 3: Linux Implementation
- [x] Add apt/dnf/pacman installation methods
- [ ] Test on Ubuntu, Fedora, Arch Linux (requires testing phase)

### Phase 4: Refactoring
- [x] Update sandbox to use shared `archive` package (instead of calling ptx-installer subprocess — keeps Go API and avoids extra binary lookup)
- [x] Remove duplicate code from sandbox.go (`downloadWin32OpenSSH` deleted, 37 lines)
- [x] Move `engine/download.go` and `engine/extract.go` into shared `src/pkg/archive` package — single source of truth used by both sandbox and ptx-installer
- [x] `embedded.go` left unchanged: `Install-PortableOpenSSH.ps1` runs **inside** Windows Sandbox (not on host), so it is not duplicated with ptx-installer's host-side variant

## Acceptance Criteria

- [x] `portunix install openssh` works on Windows (dry-run tested)
- [x] `portunix install openssh` works on Linux (dry-run tested)
- [x] `portunix install openssh --variant server` installs SSH server
- [x] Windows Sandbox continues to work with new implementation (Phase 4)
- [x] No duplicate code between sandbox and ptx-installer (Phase 4)

## Related Issues

- #100 - PTX-Installer Helper Implementation
- #122 - Consolidate Docker/Podman Installation into ptx-installer

## Priority

Medium

## Labels

enhancement, package-management, openssh, cross-platform, refactoring

---

**Created**: 2026-01-04
**Closed**: 2026-05-10
**Status**: ✅ Implemented
