# Issue #126: GitHub CLI Installation Package Manager Detection Bug

## Status
✅ Implemented

## Priority
High

## Type
Bug Fix

## Labels
bug, package-management, github-cli, arch-linux, detection

## Description

When installing GitHub CLI (`gh`) on Arch Linux, the installation system incorrectly selects `apt` as the package manager instead of `pacman`, causing installation failure.

### Observed Behavior

```
./portunix install gh

🔧 Installing package: gh
📦 Package: GitHub CLI
💻 Platform: linux
🎯 Variant: pacman (version: latest)

🚀 Starting installation (type: apt)...   <-- WRONG! Should be pacman
📦 Installing via APT: [github-cli]
❌ Installation failed: apt-get install failed: exit status 100
```

The system correctly identifies the variant as `pacman` but then uses `apt` for installation type.

### Expected Behavior

On Arch Linux systems:
- Variant should be detected as `pacman`
- Installation type should be `pacman`
- Package should be installed via `pacman -S github-cli`

### Root Cause Analysis

The issue is likely in the ptx-installer helper or the package registry logic where:
1. Variant detection works correctly (shows "pacman")
2. Installation type selection is incorrect (uses "apt" instead of "pacman")

### Affected Components

- `helpers/ptx-installer/` - Package installation logic
- Package registry variant selection

## Acceptance Criteria

- [ ] gh installation works correctly on Arch Linux using pacman
- [ ] Variant detection matches installation type
- [ ] No regression on other distributions (Ubuntu, Fedora)

## Technical Notes

- Related to issue #046 (Node.js Fedora detection) and #047 (Node.js Arch Linux detection)
- May require review of variant-to-installation-type mapping logic

## Test Plan

1. Test on Arch Linux: `./portunix install gh`
2. Verify correct package manager is used
3. Verify gh is installed and functional: `gh --version`

---

**Created**: 2026-01-04
**Closed**: 2026-05-10
**Branch**: feature/126-gh-installation-package-manager-detection
**Parent Branch**: feature/119-ptx-ansible-standalone-help
**Implemented in**: 55b3eac (`fix(#126): correct package manager detection for gh installation`)
