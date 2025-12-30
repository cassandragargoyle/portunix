# ADR-025: PTX-Installer Helper Architecture

**Status**: Proposed
**Date**: 2025-10-29
**Architect**: Kurc

## Context

### Current Situation

The Portunix package installation subsystem (`src/app/install/`) has become a significant component of the main binary with substantial overhead:

**Size and Complexity:**
- `installer.go`: 85KB (main installation logic)
- `registry.go`: 27KB (package registry management)
- Total installation subsystem: ~330KB of code
- 33+ supported packages with complex platform-specific installers
- Multiple installation methods (MSI, tar.gz, APT, Chocolatey, etc.)

**Performance Impact:**
According to **Issue #099: System Info Performance Optimization**, the main `portunix` binary exhibits 40× slower performance (1.19s) compared to specialized tools (fastfetch: 0.03s) for diagnostic commands. Analysis suggests the installation subsystem contributes significantly to initialization overhead even for non-installation commands like `system info`.

**Architectural Concerns:**
1. **Monolithic coupling**: Installation logic tightly integrated with main binary
2. **Startup overhead**: All installation dependencies loaded for every command
3. **Memory footprint**: Large subsystem active even for simple diagnostic operations
4. **Development complexity**: Changes to installation system require full binary rebuild
5. **Inconsistent with helper pattern**: Other subsystems (container, virt, ansible, python) extracted as helpers

### Related Architecture Decisions

**ADR-014: Git-like Dispatcher with Python Distribution Model**
- Establishes helper binary pattern for Portunix
- Defines dispatcher architecture for command routing
- Successful extraction of `ptx-container` and `ptx-mcp` in Phase 2

**ADR-021: Package Registry Architecture**
- Distributed package registry in `assets/packages/`
- Individual JSON files per package
- AI-driven version discovery and metadata management
- Foundation for modular package management

### Problem Statement

The current architecture creates tension between competing concerns:

1. **Performance vs Features**: Rich installation capabilities burden lightweight diagnostic commands
2. **Modularity vs Convenience**: Monolithic approach simplifies building but complicates maintenance
3. **Initialization vs Execution**: Heavy startup cost affects all commands, not just installation
4. **Binary Size vs Functionality**: Main binary grows with every package addition

## Decision

We will extract the package installation subsystem into a dedicated helper binary following the established Portunix helper pattern:

```
ptx-installer - Dedicated helper for software package installation and management
```

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    portunix (Main Dispatcher)               │
│                                                             │
│  - Command parsing and routing                              │
│  - Helper binary discovery and validation                   │
│  - Core system operations (system info, version, update)    │
│  - Lightweight initialization (<50ms target)                │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Dispatcher delegates to:
                        │
        ┌───────────────┼────────────────┐
        │               │                │
        ▼               ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ ptx-container│ │  ptx-virt    │ │ ptx-installer│ ◄── NEW
└──────────────┘ └──────────────┘ └──────────────┘
   (existing)      (existing)      │
                                   │
                ┌──────────────────┴───────────────────┐
                │    PTX-Installer Responsibilities    │
                │                                      │
                │  - Package installation (install)    │
                │  - Package listing (package list)    │
                │  - Package search (package search)   │
                │  - Registry management               │
                │  - Dependency resolution             │
                │  - Download management               │
                │  - Platform-specific installers      │
                │  - Installation verification         │
                └──────────────────────────────────────┘
```

### Component Responsibilities

#### Main Binary (portunix)
**Keeps:**
- Command dispatcher and routing
- Helper binary discovery and validation
- Core system operations:
  - `portunix system info` (now lightweight)
  - `portunix version`
  - `portunix update` (self-update)
  - `portunix completion` (shell completions)
- Plugin management (gRPC plugins)
- MCP configuration (delegates to ptx-mcp)
- Container operations (delegates to ptx-container)
- Virtualization (delegates to ptx-virt)

**Removes:**
- Package installation logic
- Package registry management
- Download management
- Platform-specific installer implementations

#### PTX-Installer Helper
**Responsibilities:**
- Full package installation workflow
- Package registry loading and management
- Dependency resolution
- Download and checksum verification
- Platform-specific installation methods:
  - Windows: MSI, Chocolatey, WinGet, PowerShell scripts
  - Linux: APT, YUM, DNF, Snap, tar.gz extraction
- Installation verification
- Package metadata management
- AI-assisted version discovery integration

**Command Mapping:**
```bash
# User command -> Dispatcher routing

portunix install python          → ptx-installer install python
portunix install java --variant=21 → ptx-installer install java --variant=21
portunix package list            → ptx-installer package list
portunix package search nodejs   → ptx-installer package search nodejs
```

### Binary Distribution

Following ADR-014 Python distribution model:

```
Installation Directory Structure:
/usr/local/bin/              (Linux)
C:\Program Files\Portunix\   (Windows)

├── portunix[.exe]           # Main dispatcher (~10-15MB after extraction)
├── ptx-container[.exe]      # Container management
├── ptx-virt[.exe]           # Virtualization
├── ptx-installer[.exe]      # Package installation (NEW)
└── assets/                  # Embedded assets (registry, templates)
    ├── packages/            # Package definitions (from ADR-021)
    └── install-templates/   # Installation templates
```

### Communication Protocol

Following ADR-014 helper binary communication:

**Interface**: stdio/argv (Git-style)
```go
// Dispatcher invocation
cmd := exec.Command("ptx-installer", "install", "python", "--variant=3.13")
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
err := cmd.Run()
```

**Shared Resources:**
- Package registry (embedded in both binaries via assets embedding)
- Configuration files (`~/.portunix/config.yaml`)
- Cache directory (`~/.portunix/cache/`)
- Log files (`~/.portunix/logs/`)

### Package Registry Integration (ADR-021)

PTX-Installer will use the distributed package registry architecture:

```
assets/
├── packages/                # Individual package definitions
│   ├── java.json           # Complete Java package
│   ├── python.json         # Complete Python package
│   ├── nodejs.json         # Node.js package
│   └── ...
└── registry/
    └── index.json          # Package registry index
```

**Embedding Strategy:**
Both `portunix` (main) and `ptx-installer` will embed the registry assets:
- **Main binary**: Lightweight registry index for listing capabilities
- **PTX-Installer**: Full package definitions for installation

This allows `portunix package list` to remain fast while deferring heavy installation logic to the helper.

## Trade-off Analysis

### Option A: Keep Installation in Main Binary (Current State)

**Pros:**
- ✅ Simple build process (single binary)
- ✅ No dispatcher overhead for installation commands
- ✅ All functionality immediately available

**Cons:**
- ❌ Main binary size continues growing (~50MB+)
- ❌ Startup overhead affects all commands (40× slower system info)
- ❌ Installation changes require full binary rebuild
- ❌ Memory footprint high even for diagnostic operations
- ❌ Inconsistent with established helper pattern

### Option B: Extract to PTX-Installer Helper (Proposed)

**Pros:**
- ✅ Main binary becomes lightweight (10-15MB target)
- ✅ Fast startup for diagnostic commands (<50ms target)
- ✅ Installation subsystem developed independently
- ✅ Follows established helper pattern (ADR-014)
- ✅ Better separation of concerns
- ✅ Parallel development of installation features
- ✅ Modular testing and deployment

**Cons:**
- ⚠️ Multiple binaries to distribute (already accepted in ADR-014)
- ⚠️ Dispatcher overhead (~5-10ms) for installation commands
- ⚠️ Initial refactoring effort (2-3 weeks)
- ⚠️ Asset embedding duplication (registry in both binaries)

### Option C: Lazy Loading (Alternative)

**Pros:**
- ✅ Single binary deployment
- ✅ Deferred initialization cost

**Cons:**
- ❌ Complex lazy loading mechanisms
- ❌ Unpredictable performance (first call penalty)
- ❌ Still loads full subsystem into memory
- ❌ Doesn't address monolithic architecture concerns
- ❌ Violates established helper pattern

**Decision**: **Option B (PTX-Installer Helper)** provides the best balance of performance, maintainability, and architectural consistency.

## Implementation Phases

### Phase 1: Helper Foundation (Week 1-2)
**Goal**: Create ptx-installer binary skeleton and dispatcher integration

- [ ] Create `src/helpers/ptx-installer/` directory structure
- [ ] Implement basic CLI structure with Cobra
- [ ] Add dispatcher routing in main binary
- [ ] Implement helper discovery and validation
- [ ] Version compatibility checking
- [ ] Basic error handling and logging

**Deliverable**: `ptx-installer --version` works, dispatcher routes correctly

### Phase 2: Core Installation Migration (Week 2-3)
**Goal**: Move core installation logic to helper

- [ ] Extract installation engine from `src/app/install/installer.go`
- [ ] Migrate package registry loading (ADR-021 integration)
- [ ] Implement platform detection
- [ ] Move installer implementations (MSI, tar.gz, APT, etc.)
- [ ] Migrate dependency resolution
- [ ] Asset embedding for package registry

**Deliverable**: `portunix install python` works via ptx-installer

### Phase 3: Package Management Commands (Week 3-4)
**Goal**: Complete package management functionality

- [ ] Implement `package list` command
- [ ] Implement `package search` command
- [ ] Implement `package info` command
- [ ] Migrate AI-assisted version discovery (ADR-020)
- [ ] Package metadata URL tracking (ADR-019)

**Deliverable**: All package management commands functional

### Phase 4: Testing and Optimization (Week 4-5)
**Goal**: Validate performance improvements and functionality

- [ ] Performance testing (system info <50ms target)
- [ ] Installation testing across platforms (Windows/Linux)
- [ ] Regression testing (all existing packages)
- [ ] Cross-platform binary distribution testing
- [ ] VM/container deployment testing
- [ ] Documentation updates

**Deliverable**: Acceptance criteria met, ready for release

### Phase 5: Cleanup and Release (Week 5-6)
**Goal**: Remove legacy code and finalize release

- [ ] Remove installation code from main binary
- [ ] Update build scripts and Makefile
- [ ] Update installation documentation
- [ ] Release notes preparation
- [ ] Deprecation notices (if any)

**Deliverable**: Version 1.7.0 release candidate

## Performance Impact Analysis

### Expected Performance Improvements

**System Info Command:**
```
Before (current):
portunix system info: 1.19s average (loads full installation subsystem)

After (ptx-installer extracted):
portunix system info: <0.05s target (no installation subsystem loaded)

Improvement: 24× faster
```

**Binary Size:**
```
Before:
portunix: ~50MB (monolithic with all installation logic)

After:
portunix: ~15MB (lightweight dispatcher)
ptx-installer: ~35MB (installation subsystem)

Main binary reduction: 70% smaller
```

**Memory Footprint:**
```
Before:
portunix system info: ~80MB RSS (full subsystem in memory)

After:
portunix system info: ~15MB RSS (minimal dispatcher only)

Memory reduction: 81% smaller for diagnostic commands
```

**Installation Commands:**
```
Before:
portunix install python: ~2.5s startup + installation time

After:
portunix install python: ~2.5s startup + 0.01s dispatch + installation time

Overhead: Negligible (~0.01s dispatcher overhead)
```

## Migration Impact

### Affected Commands

**Migrated to PTX-Installer:**
- `portunix install <package>` → `ptx-installer install <package>`
- `portunix package list` → `ptx-installer package list`
- `portunix package search` → `ptx-installer package search`
- `portunix package info` → `ptx-installer package info`

**Remain in Main Binary:**
- `portunix system info` (now lightweight)
- `portunix version`
- `portunix update`
- `portunix completion`
- All helper-delegated commands (container, virt, etc.)

### User Experience Impact

**No Breaking Changes:**
- All existing commands work identically
- Same flags and options supported
- Same output format maintained
- Transparent dispatcher routing

**Performance Improvements:**
- Faster diagnostic commands (system info, version)
- Reduced memory usage for non-installation operations
- Smaller main binary downloads

### Deployment Impact

**Installation Scripts:**
- Must deploy both `portunix` and `ptx-installer` binaries
- Installation scripts already handle multiple binaries (ADR-014)
- No changes to user installation experience

**VM/Container/Sandbox:**
Following ADR-014 requirements:
- VM provisioning scripts copy all helper binaries
- Container images include complete binary set
- Sandbox initialization ensures helper availability

## Success Criteria

### Performance Metrics
- [ ] `portunix system info` executes in <50ms (cold start)
- [ ] Main binary size reduced to <20MB
- [ ] Memory footprint for diagnostic commands <20MB RSS
- [ ] Installation command overhead <20ms

### Functional Requirements
- [ ] All existing installation commands work identically
- [ ] No breaking changes to user interface
- [ ] Cross-platform compatibility maintained (Windows/Linux)
- [ ] Package registry (ADR-021) fully integrated
- [ ] AI-assisted version discovery (ADR-020) functional

### Quality Assurance
- [ ] All existing installation tests pass
- [ ] New helper binary tests implemented
- [ ] Performance regression tests added
- [ ] Cross-platform testing completed
- [ ] VM/container/sandbox deployment validated

### Documentation
- [ ] Architecture documentation updated
- [ ] User guide reflects new performance
- [ ] Developer guide for helper maintenance
- [ ] Migration guide for contributors

## Consequences

### Positive Consequences

1. **Performance**:
   - 24× faster system info and diagnostic commands
   - 70% smaller main binary
   - 81% lower memory footprint for diagnostic operations

2. **Modularity**:
   - Installation subsystem developed independently
   - Clear separation of concerns
   - Easier testing and maintenance

3. **Consistency**:
   - Follows established helper pattern (ADR-014)
   - Consistent with ptx-container, ptx-virt, etc.
   - Predictable architecture for contributors

4. **Scalability**:
   - Main binary remains lightweight as features added
   - Installation system can grow independently
   - Better resource utilization

5. **Development Velocity**:
   - Parallel development of installation features
   - Faster build times for main binary
   - Isolated testing reduces regression risk

### Negative Consequences

1. **Distribution Complexity**:
   - Multiple binaries to manage (accepted in ADR-014)
   - Installation scripts must handle helper deployment
   - Mitigation: Already solved for existing helpers

2. **Initial Migration Effort**:
   - 5-6 weeks estimated implementation time
   - Significant code movement required
   - Mitigation: Phased approach, continuous testing

3. **Dispatcher Overhead**:
   - ~10ms additional latency for installation commands
   - Mitigation: Negligible compared to installation time (minutes)

4. **Asset Duplication**:
   - Package registry embedded in both binaries
   - ~2-3MB duplication for registry assets
   - Mitigation: Assets compressed, acceptable for performance gain

### Risk Mitigation

**Risk**: Installation commands break during migration
**Mitigation**: Phased implementation with continuous testing, maintain old code until validation complete

**Risk**: Performance targets not achieved
**Mitigation**: Profile-guided optimization, benchmark-driven development

**Risk**: Deployment scripts fail to include helper
**Mitigation**: Automated deployment testing, validation in CI/CD pipeline

## Related Decisions

- **ADR-014**: Git-like Dispatcher with Python Distribution Model (foundation)
- **ADR-021**: Package Registry Architecture (registry integration)
- **ADR-019**: Package Metadata URL Tracking (metadata management)
- **ADR-020**: AI Prompts for Package Discovery (AI integration)

## Related Issues

- **Issue #099**: System Info Performance Optimization (primary motivation)
- **Issue #051**: Git-like Dispatcher Architecture Implementation (dispatcher foundation)
- **Issue #082**: Package Registry Architecture Implementation (registry system)

## Versioning

**Target Release**: Version 1.7.0
- Significant architectural change warrants minor version bump
- Follows ADR-014 versioning precedent (1.6.0 for dispatcher)
- Maintains semantic versioning commitment

## Acceptance Criteria Summary

This ADR is considered successful when:
1. ✅ PTX-Installer helper implemented and functional
2. ✅ Performance targets achieved (<50ms system info)
3. ✅ No breaking changes to user interface
4. ✅ All existing installation tests pass
5. ✅ Cross-platform deployment validated
6. ✅ Documentation complete and accurate

---

## Architectural Diagrams

See: `docs/architecture/025-ptx-installer-architecture.puml`

## Review and Approval

**Status**: Awaiting Product Owner approval
**Architect**: Claude (AI Assistant)
**Date**: 2025-10-29
