# Issue #125: Cross-Platform Binary Distribution

**Status**: 📋 Open
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, architecture, distribution, cross-platform, container, vm

## Summary

Implement cross-platform binary distribution strategy as defined in ADR-031. Each OS release package will contain native binaries plus archived binaries for other platforms, enabling seamless cross-platform provisioning (e.g., Windows host → Linux container).

## Related ADR

- [ADR-031: Cross-Platform Binary Distribution Strategy](../adr/031-cross-platform-binary-distribution.md)

## Problem Statement

When Portunix runs on Windows and needs to provision a Linux container or VM, it currently fails because:
1. Only Windows binaries are available in the Windows package
2. Linux binaries must be manually downloaded and managed
3. Playbook execution fails with "exec /usr/local/bin/portunix: no such file or directory"

## Solution

Include archived binaries for all supported platforms in each release package:

```
portunix_v1.x.x_windows_amd64.zip
├── portunix.exe              # Native Windows
├── ptx-*.exe                 # Native Windows helpers
├── install.ps1
├── install.sh
├── README.md
├── LICENSE
└── platforms/                # All platform archives
    ├── linux-amd64.tar.gz
    ├── linux-arm64.tar.gz
    ├── windows-amd64.zip
    └── darwin-amd64.tar.gz
```

## Implementation Tasks

### Phase 1: Build System Updates ✅

- [x] Update `.goreleaser.yml` to build binaries for all target platforms
- [x] Create script to generate platform archives after build
- [x] Update `scripts/make-release.sh` to include `platforms/` directory
- [x] Modify release artifacts structure

### Phase 2: Runtime Integration ✅

- [x] Add `getPlatformBinaries(platform string)` function in ptx-ansible
- [x] Implement platform detection from container image (e.g., "ubuntu:22.04" → "linux-amd64")
- [x] Add archive extraction logic with caching
- [x] Update `setupContainerEnvironment()` to use platform binaries
- [ ] Update VM provisioning to use platform binaries

### Phase 3: Installation Updates ✅

- [x] Update `scripts/install.sh` to handle `platforms/` directory
- [x] Update `scripts/install.ps1` to handle `platforms/` directory
- [ ] Document new directory structure

### Phase 4: Testing

- [ ] Test Windows → Linux container provisioning
- [ ] Test Linux → Windows VM provisioning (if applicable)
- [ ] Test binary caching mechanism
- [ ] Performance testing of archive extraction

### Phase 5: Script Migration to Python ✅

- [x] Convert `scripts/create-platform-archives.sh` to Python
- [x] Convert `scripts/make-release.sh` to Python
- [x] Implement virtual environment (venv) awareness
- [x] Ensure scripts work on Windows, Linux, and macOS
- [x] Use project's `.venv` directory if available

## Acceptance Criteria

1. Windows package contains `platforms/linux-amd64.tar.gz`
2. Linux package contains `platforms/windows-amd64.zip`
3. `portunix playbook run` with container target works transparently on Windows
4. Platform binaries are cached after first extraction
5. No manual binary management required for cross-platform scenarios

## Size Impact

- Current package: ~60MB
- New package: ~180MB (3× due to 3 platform archives)
- Trade-off accepted for seamless cross-platform experience

## Files to Modify

- `.goreleaser.yml`
- `scripts/make-release.sh` → `scripts/make-release.py` (Phase 5)
- `scripts/create-platform-archives.sh` → `scripts/create-platform-archives.py` (Phase 5)
- `scripts/install.sh`
- `scripts/install.ps1`
- `src/helpers/ptx-ansible/executor.go`
- `Makefile`

### Related Updates (discovered during implementation)

- `aur-package/PKGBUILD` - updated to include all helper binaries, fixed directory name
- `aur-repo/PKGBUILD` - updated to include all helper binaries
- `scripts/aur-prepare.sh` - updated PKGBUILD template with all helpers
- `scripts/upload-release-to-github.py` - new script for uploading releases to GitHub

## Dependencies

- ADR-031 approved
- GoReleaser multi-platform build working
