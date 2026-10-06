# Issue #123: Consolidate Installation Systems - Remove Duplicate Assets

## Summary

Currently there are two parallel installation systems with duplicated assets:

1. **Old system** (`src/app/install/`) using root `assets/`
2. **New system** (`ptx-installer`) using `src/helpers/ptx-installer/assets/`

This causes code duplication, maintenance burden, and unnecessary binary size increase.

## Problem Statement

1. **Duplicate assets**: Package definitions exist in both `assets/packages/` and `src/helpers/ptx-installer/assets/packages/`
2. **Two registry systems**: `src/app/install/registry.go` and `src/helpers/ptx-installer/registry/registry.go`
3. **Embedded twice**: Main portunix embeds `assets/` (336KB), ptx-installer embeds its own (284KB)
4. **Maintenance burden**: Changes need to be applied in multiple places
5. **Inconsistent state**: Some packages only in one location (clearflask, eververse, fider in root; docker, podman in ptx-installer)

## Current State

```text
assets/                              # Embedded in main portunix
├── packages/                        # 35 package definitions
│   ├── chrome.json
│   ├── java.json
│   ├── clearflask.json             # Only here
│   ├── eververse.json              # Only here
│   ├── fider.json                  # Only here
│   └── ...
├── registry/
├── scripts/
└── templates/

src/helpers/ptx-installer/assets/    # Embedded in ptx-installer
├── packages/                        # 34 package definitions
│   ├── chrome.json                  # Duplicate
│   ├── java.json                    # Duplicate
│   ├── docker.json                  # Only here
│   ├── podman.json                  # Only here
│   └── ...
├── registry/
├── scripts/
└── templates/

src/app/install/
├── installer.go                     # Old installation logic
├── registry.go                      # Old registry system
└── ...
```

## Target State

```text
src/helpers/ptx-installer/assets/    # Single source of truth
├── packages/                        # All package definitions
├── registry/
├── scripts/
└── templates/

src/app/install/                     # Delegates to ptx-installer
├── delegator.go                     # Thin wrapper calling ptx-installer
└── ...

assets/                              # REMOVED or minimal (only non-package assets)
```

## Requirements

### Functional Requirements

1. `portunix install <package>` works exactly as before
2. All existing packages remain available
3. Package definitions consolidated in one location
4. No duplicate embedded assets

### Technical Requirements

1. Move missing packages to ptx-installer:
   - clearflask.json
   - eververse.json
   - fider.json
2. Update `src/app/install/` to delegate to ptx-installer binary
3. Remove old registry system from `src/app/install/registry.go`
4. Remove or minimize root `assets/` directory
5. Update `main.go` embed directive
6. Keep `assets/templates/edge/` if used by edge module

## Migration Plan

### Phase 1: Consolidate Package Definitions

- [ ] Copy clearflask.json, eververse.json, fider.json to ptx-installer/assets/packages/
- [ ] Verify all packages are in ptx-installer
- [ ] Sync any other differences

### Phase 2: Update Main Portunix

- [ ] Change `portunix install` to delegate to `ptx-installer` binary
- [ ] Remove old `src/app/install/registry.go`
- [ ] Remove old installation logic from `src/app/install/installer.go`
- [ ] Keep only delegation code

### Phase 3: Cleanup

- [ ] Remove `assets/packages/` from root
- [ ] Remove `assets/registry/` from root
- [ ] Update `main.go` embed directive (only scripts/templates if needed)
- [ ] Update tests
- [ ] Update documentation

## Acceptance Criteria

- [ ] Single source of package definitions in ptx-installer
- [ ] `portunix install chrome` works via delegation
- [ ] `portunix install docker` works (builtin type)
- [ ] No duplicate package JSON files
- [ ] Binary size reduced (no double embedding)
- [ ] All existing packages still installable

## Benefits

1. **Reduced binary size**: No duplicate embedded assets
2. **Single source of truth**: One place for package definitions
3. **Easier maintenance**: Changes in one location
4. **Cleaner architecture**: Clear separation of concerns

## Related Issues

- #122 - Consolidate Docker/Podman Installation (completed, added docker/podman to ptx-installer)
- #87 - Assets Embedding Architecture

## Priority

Medium

## Labels

refactoring, architecture, ptx-installer, assets

---
**Created**: 2026-01-03
**Status**: Implemented
**Implemented**: 2026-01-03
