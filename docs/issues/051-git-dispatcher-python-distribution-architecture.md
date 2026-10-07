# Issue #51: Git-like Dispatcher with Python Distribution Architecture

## Summary

Implement Git-like dispatcher architecture combined with Python distribution model to improve Portunix modularity and maintainability, as defined in ADR-014.

## Problem Statement

Portunix has grown into a monolithic binary with multiple subsystems. The current architecture faces:

- Binary size growth with each new core feature
- Development complexity requiring understanding of entire codebase
- Testing overhead for changes affecting the entire system
- Limited modularity for independent subsystem development

## Proposed Solution

Implement a dispatcher architecture following Git's proven pattern combined with Python's distribution model:

1. **Main Dispatcher**: Lightweight `portunix` binary for command routing
2. **Helper Binaries**: Extract subsystems as `ptx-{subsystem}` binaries
3. **Python Distribution Model**: All binaries in same directory for cross-platform compatibility
4. **Multi-Version Support**: Python launcher pattern for version management

## Architecture Reference

This implementation is based on **ADR-014: Git-like Dispatcher with Python Distribution Model**

- **ADR Location**: `docs/adr/014-git-dispatcher-with-python-distribution.md`
- **Status**: Proposed
- **Author**: Zdeněk
- **Date**: 2025-09-19

## Implementation Phases

### Phase 1: Dispatcher Infrastructure (Version 1.6.0)

**Prerequisites:**

- [ ] Code structure reorganization with `src/` directory
- [ ] Version bump to 1.6.0

**Core Implementation:**

- [ ] Implement dispatcher logic in main binary
- [ ] Add helper binary discovery mechanism
- [ ] Create shared library for common functionality
- [ ] Implement version validation between main binary and helpers
- [ ] Add multi-version support with registry reading (Windows)

### Phase 2: Container System Extraction (Version 1.6.1) ✅ COMPLETED

**Helper Binary Development:**

- [x] Extract unified container management (`ptx-container`)
  - Handles `portunix container`, `portunix docker`, `portunix podman`
  - Implements unified interface with backend detection
- [x] Extract MCP server (`ptx-mcp`)
  - Standalone MCP server implementation
  - Independent operation with clear isolation boundaries

**Integration:**

- [x] Maintain compatibility facades in main binary
- [x] Test deployment across VM/container/sandbox environments
- [x] Update deployment scripts for multiple binaries

**Build System Updates:**

- [x] Modify `build-with-version.sh` to build all binaries (main + helpers)
- [x] Update build system for multi-binary output
- [x] Add Makefile targets for helper binaries (`make build-helpers`, `make build-all`)
- [x] Ensure all binaries get same version number
- [x] Helper binaries placed under `src/helpers/` structure

**Implementation Status (2025-09-19):**

- ✅ Helper binaries: `ptx-container`, `ptx-mcp` implemented
- ✅ Dispatcher delegation: Commands correctly routed to helpers
- ✅ Version validation: "dev" versions supported during development
- ✅ Build system: Complete multi-binary build support
- ✅ Testing: Basic functionality verified
- ⏳ **Next**: Comprehensive testing by QA team

### Phase 3: Documentation and Deployment (Version 1.6.2)

- [ ] Update installation scripts and documentation
- [ ] Deploy version management features
- [ ] Complete migration impact testing
- [ ] Update build and release processes

## Success Criteria

- [ ] Main binary size reduced by at least 50%
- [ ] Helper binaries can be developed independently
- [ ] No breaking changes to user interface
- [ ] Installation process remains simple for end users
- [ ] Performance equal or better than monolithic version
- [ ] Clear documentation for helper binary development

## Migration Impact

**Critical Systems Requiring Updates:**

1. **VM Installation** (`portunix virt`): Deploy complete binary set
2. **Container Integration** (`portunix container`): Handle multiple binaries
3. **Sandbox Environment** (`portunix sandbox`): Ensure helper availability

## Technical Specifications

### Binary Naming Convention

- **Main dispatcher**: `portunix`
- **Helper binaries**: `ptx-{subsystem}` (e.g., `ptx-container`, `ptx-mcp`)
- **gRPC plugins**: `ptx-plugin-{name}` (unchanged)

### Communication Protocol

- **Helper binaries**: stdio/argv interface (Git-style)
- **gRPC plugins**: Existing architecture (unchanged)
- **Shared configuration**: Environment variables and config files

### Directory Structure (Single-Version)

```text
Linux:   /usr/local/bin/
Windows: C:\Program Files\Portunix\bin\

Contents:
├── portunix[.exe]           # Main dispatcher
├── ptx-container[.exe]      # Helper: Container management
├── ptx-mcp[.exe]            # Helper: MCP server
├── ptx-plugin-agile[.exe]   # gRPC plugins
└── ptx-plugin-github[.exe]
```

## Acceptance Criteria

1. **Functional Compatibility**: All existing commands work identically
2. **Performance**: No degradation in command execution speed
3. **Installation**: Simple single-package installation maintained
4. **Multi-Version**: Support for side-by-side version installations
5. **Cross-Platform**: Works on Windows and Linux without changes
6. **Documentation**: Complete developer guide for helper binary creation

## Dependencies

- ADR-014 approval and finalization
- Go 1.21+ for implementation
- Updated build scripts and CI/CD pipelines
- Cross-platform testing infrastructure

## Risk Assessment

**Low Risk:**

- Dispatcher implementation (proven Git pattern)
- Helper binary extraction (clear boundaries)

**Medium Risk:**

- Multi-version support complexity
- Installation script updates

**High Risk:**

- Migration impact on VM/container/sandbox deployment
- Backward compatibility during transition

## Implementation Notes & Lessons Learned

### Development Structure Consistency (2025-09-19)

**Issue**: During Phase 2 implementation, helper binaries were incorrectly placed in root `helpers/` directory instead of respecting Phase 1 `src/` reorganization.

**Correction**: All helper binaries must be placed under `src/helpers/` to maintain consistency with Phase 1 architecture:

- ✅ Correct: `src/helpers/ptx-container/`
- ✅ Correct: `src/helpers/ptx-mcp/`
- ❌ Wrong: `helpers/ptx-container/` (root level)

**Lesson**: Always respect established architecture patterns from previous phases.

## Timeline Estimate

- **Phase 1**: 3-4 weeks (dispatcher infrastructure)
- **Phase 2**: 2-3 weeks (helper extraction)
- **Phase 3**: 1-2 weeks (documentation and deployment)
- **Total**: 6-9 weeks

## Related Issues

- Issue #007: Plugin System with gRPC Architecture
- Issue #024: Basic Plugin System
- Issue #029: Universal Container Command
- Issue #032: Universal Container Management Commands

---

**Status**: Proposed
**Priority**: High
**Complexity**: High
**Estimated Effort**: 6-9 weeks
**Assignee**: TBD
**Created**: 2025-09-19
**Last Updated**: 2025-09-19