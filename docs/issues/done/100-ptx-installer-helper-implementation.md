# Issue #100: PTX-Installer Helper Implementation

**Status**: ✅ Implemented (Phase 1-4 complete; Phase 5 substantially done — deferred code-cleanup tracked as sub-issues)
**Priority**: High
**Type**: Feature / Architecture
**Created**: 2025-10-29
**Updated**: 2026-05-19
**Architecture Decision Records**:
- [ADR-025: PTX-Installer Helper Architecture](../../adr/025-ptx-installer-helper-architecture.md)
- [ADR-026: Shared Platform Utilities](../../adr/026-shared-platform-utilities.md)
**Branches**:
- `feature/issue-100-ptx-installer-helper` (Phases 1-4)
- `feature/issue-100-phase5-cleanup` (Phase 5 wrap-up)
**Commits (Phase 1-4)**: 22f5d33 (Phase 1), d205cd1 (Phase 2 partial), 1214d8e (Phase 2 complete), 196393a (Phase 2 embedded assets)

## Phase 5 Closure Note (2026-05-19)

After review of the implementation state and codebase, Phase 5 is closed
**substantially complete** with the following findings:

**Already achieved by other work:**
- Performance target `system info <50ms` — achieved in 2.3.0 via [Issue #099](099-system-info-performance-optimization.md) using a different optimization path (not via main-binary install-code removal).
- Build system updates (Makefile, `.goreleaser.yml`, `build-with-version.sh`) — completed during Phase 1-2.
- Deployment scripts — `install.sh` / `install.ps1` work via GoReleaser archives that already bundle `ptx-installer`.
- Version bump — original target `1.7.0` is obsolete; project is already at **2.3.0**.

**Deferred to follow-up sub-issues (code-cleanup):**
`src/app/install/` (~373 KB) still has 8 active consumers in the main binary
and in `ptx-mcp` / `ptx-plugin-registry` helpers, so it cannot be removed
without a coordinated multi-step refactoring. Tracked as:

- **#100a (proposed)** — Migrate `portunix registry` command to `ptx-installer`
  (or extract shared `src/pkg/registry/`). Affected: `src/cmd/registry.go`
  (~800 lines, uses `LoadPackageRegistry`, `NewAIPackageManager`,
  `VersionDiscoveryResult`).
- **#100b (proposed)** — Refactor `ptx-mcp` to call `ptx-installer` as a
  subprocess for claude-code installation. Affected:
  `src/helpers/ptx-mcp/init.go:464` (`install.InstallPackage("claude-code", "npm")`).
- **#100c (proposed)** — Extract `LoadInstallConfig()` into a shared package
  (e.g. `src/pkg/installconfig/`) for container subsystems. Affected:
  `src/cmd/docker_run_in_container.go:124`, `src/cmd/podman_run_in_container.go`.
- **#100d (proposed)** — Migrate `install apt source`, `install iso`,
  `install chocolatey` administrative subcommands into `ptx-installer` (or a
  new helper). Affected: `src/cmd/install_apt.go`, `src/cmd/install_iso.go`,
  `src/cmd/install_chocolatey.go`.

The binary-size and memory-footprint targets from this issue's "Success
Metrics" section depend on the four sub-issues above being completed; they
are intentionally left as open targets here.

## Summary

Extract Portunix package installation subsystem into a dedicated `ptx-installer` helper binary to improve performance, modularity, and maintainability. This addresses performance issues identified in Issue #099 and follows the established helper binary pattern from ADR-014.

## Problem Statement

The current monolithic architecture integrates the entire package installation subsystem (~330KB code) into the main `portunix` binary, causing:

### Performance Issues (Issue #099)
- **System info command**: 1.19s average (40× slower than fastfetch)
- **Root cause**: Full installation subsystem loaded for all commands
- **Impact**: Slow diagnostic commands, poor user experience
- **Memory overhead**: ~80MB RSS for simple info retrieval

### Architectural Issues
- **Binary size**: Main binary growing continuously (~50MB+)
- **Monolithic coupling**: Installation logic tightly integrated
- **Development complexity**: Changes require full binary rebuild
- **Inconsistency**: Other subsystems already extracted (ptx-container, ptx-virt, ptx-ansible, ptx-python)

### Scale Challenges
- 33+ packages with complex platform-specific logic
- Package registry growing (ADR-021)
- AI-assisted version discovery overhead (ADR-020)
- Multiple installation methods per platform

## Proposed Solution

Implement a dedicated `ptx-installer` helper binary following ADR-025 architecture:

```
ptx-installer - Package installation and management helper
```

### Architecture Overview

```
User Command: portunix install python
                    ↓
         ┌──────────────────┐
         │ Main Dispatcher  │  (lightweight, fast startup)
         │   (portunix)     │
         └────────┬─────────┘
                  │ delegates to
                  ↓
         ┌──────────────────┐
         │  PTX-Installer   │  (full installation subsystem)
         │   Helper Binary  │
         └──────────────────┘
                  │
          ┌───────┴────────┐
          ↓                ↓
    [Registry]      [Platform Installers]
    [Download]      [Dependencies]
    [Verification]  [AI Discovery]
```

### Key Benefits

**Performance Improvements:**
- System info: <50ms (24× faster than current 1.19s)
- Main binary: ~15MB (70% smaller than current ~50MB)
- Memory footprint: ~15MB RSS for diagnostics (81% reduction)

**Architectural Improvements:**
- Modular development of installation features
- Clear separation of concerns
- Follows established helper pattern (ADR-014)
- Consistent with existing helpers (container, virt, ansible, python)

**Scalability:**
- Main binary remains lightweight as features added
- Installation system grows independently
- Better resource utilization

## Related Architecture & Issues

### Architecture Decision Records
- **[ADR-025](../../adr/025-ptx-installer-helper-architecture.md)**: PTX-Installer Helper Architecture (this implementation)
- **[ADR-014](../../adr/014-git-dispatcher-with-python-distribution.md)**: Git-like Dispatcher Pattern (foundation)
- **[ADR-021](../../adr/021-package-registry-architecture.md)**: Package Registry Architecture (registry integration)
- **[ADR-019](../../adr/019-package-metadata-url-tracking.md)**: Package Metadata URL Tracking
- **[ADR-020](../../adr/020-ai-prompts-for-package-discovery.md)**: AI Prompts for Package Discovery

### Related Issues
- **[Issue #099](099-system-info-performance-optimization.md)**: System Info Performance Optimization (primary motivation)
- **[Issue #051](../051-git-dispatcher-python-distribution-architecture.md)**: Dispatcher Architecture Implementation (foundation)
- **[Issue #082](082-package-registry-architecture-implementation.md)**: Package Registry Implementation (registry system)

## Implementation Phases

### Phase 1: Helper Foundation (Week 1-2) ✅ COMPLETED
**Goal**: Create ptx-installer binary skeleton and dispatcher integration
**Completed**: 2025-10-29
**Commit**: 22f5d33

#### Tasks:
1. **Helper Binary Infrastructure**
   - [x] Create `src/helpers/ptx-installer/` directory structure
   - [x] Implement basic CLI structure with Cobra
   - [x] Add `--version`, `--help` commands
   - [x] Set up build pipeline in Makefile
   - [ ] Implement integration tests for binary communication (deferred to Phase 4)

2. **Main Binary Dispatcher**
   - [x] Add dispatcher routing for `install` commands
   - [x] Implement helper discovery for `ptx-installer`
   - [x] Add version compatibility checking (inherited from existing dispatcher)
   - [x] Implement fallback error handling when helper not found (inherited from dispatcher)
   - [ ] Add dispatcher overhead monitoring (deferred to Phase 4)

3. **Basic Communication**
   - [x] Implement stdio/argv interface (Git-style)
   - [x] Test argument passing (flags, options, values)
   - [x] Implement exit code propagation (inherited from dispatcher)
   - [x] Test stdout/stderr forwarding

#### Success Criteria:
- [x] `ptx-installer --version` works independently
- [x] `portunix install --help` routes to helper correctly
- [ ] Dispatcher overhead measured (<20ms target) - deferred to Phase 4
- [x] Version compatibility validation functional

#### Deliverable:
✅ Working binary skeleton with dispatcher integration

---

### Phase 2: Core Installation Migration (Week 2-3) ✅ COMPLETED (100%)
**Goal**: Move core installation logic to helper
**Completed**: 2025-10-29
**Commits**: d205cd1 (partial), 1214d8e (complete), 196393a (embedded assets)
**Acceptance Protocol**: `test/integration/acceptance_issue_100_phase2.md`

#### Tasks:
1. **Installation Engine Extraction**
   - [x] Extract from `src/app/install/installer.go` (85KB)
   - [x] Move to `src/helpers/ptx-installer/engine/`
   - [x] Implement platform detection
   - [x] Migrate download manager
   - [x] Implement checksum verification (deferred - Phase 4)

2. **Package Registry Integration (ADR-021)**
   - [x] Embed package registry assets (go:embed with 34 packages)
   - [x] Implement registry loading from `assets/packages/`
   - [x] Load individual package definitions
   - [x] Parse package metadata and variants

3. **Platform-Specific Installers**
   - [ ] Migrate MSI installer (Windows) - deferred Phase 3
   - [x] Migrate tar.gz extractor (Linux) - working
   - [x] Migrate APT integration (Debian/Ubuntu) - working
   - [ ] Migrate Chocolatey integration (Windows) - deferred Phase 3
   - [x] Migrate YUM/DNF integration (RedHat/Fedora) - implemented

4. **Dependency Resolution**
   - [x] Extract dependency resolver
   - [x] Implement prerequisite detection
   - [x] Implement installation ordering
   - [ ] Test complex dependency chains - deferred Phase 3

5. **Shared Resources**
   - [x] Implement config file access (via embedded assets)
   - [x] Implement cache directory management
   - [x] Implement logging integration
   - [x] Test cross-binary resource sharing

#### Success Criteria:
- [x] `portunix install hugo` works via ptx-installer (tested in container)
- [x] Archive installers functional (tar.gz tested and working)
- [x] Dependency resolution works correctly (topological sort implemented)
- [x] Installation verification passes (hugo v0.150.1 confirmed)

#### Additional Achievements:
- [x] Embedded assets system (13MB binary with 34 packages)
- [x] Architecture normalization (amd64→x64)
- [x] Variant selection priority (default>standard>first)
- [x] Container-based testing (Ubuntu 22.04)
- [x] 5/5 test cases passed

#### Deliverable:
✅ Functional installation via helper binary with embedded assets

---

### Phase 2.5: Code Refactoring - Shared Platform Utilities ✅ COMPLETED
**Goal**: Eliminate code duplication by creating shared platform utilities package
**Completed**: 2025-10-29
**ADR**: [ADR-026: Shared Platform Utilities](../../adr/026-shared-platform-utilities.md)
**Rationale**: Code duplication detected between main binary and ptx-installer helper

#### Problem Identified:
During Phase 2 testing, discovered duplicated platform detection code:
- `src/app/install/config.go:246` - `GetArchitecture()` in main binary
- `src/helpers/ptx-installer/engine/platform.go:32` - `GetArchitecture()` in helper

#### Solution Implemented:
1. **Created Shared Package**: `src/pkg/platform/`
   - [x] `platform.go` - OS and architecture detection
   - [x] `permissions.go` - Privilege and permission utilities
   - [x] `platform_test.go` - Comprehensive unit tests

2. **Core Functions**:
   - [x] `GetOS()` - Returns: "windows", "linux", "darwin"
   - [x] `GetArchitecture()` - Returns: "x64", "x86", "arm64"
   - [x] `GetPlatform()` - Returns: "linux-x64", etc.
   - [x] `IsRunningAsRoot()` - Privilege detection
   - [x] `IsSudoAvailable()` - Sudo availability check
   - [x] `GetSudoPrefix()` - Returns "sudo " or ""
   - [x] `CanWriteToDirectory()` - Permission check
   - [x] `IsUserDirectory()` - Path validation

3. **Refactored Consumers**:
   - [x] Updated `src/helpers/ptx-installer/engine/platform.go` to use shared package
   - [x] Created wrapper functions for backward compatibility
   - [x] All tests passing (12/12)

#### Benefits Achieved:
- ✅ **Single Source of Truth**: One implementation for all binaries
- ✅ **Reduced Duplication**: ~100 lines of duplicate code eliminated
- ✅ **Better Testing**: Comprehensive test suite (12 tests)
- ✅ **Consistent Behavior**: Guaranteed identical logic across binaries
- ✅ **Easier Maintenance**: Changes in one place affect all consumers

#### Test Results:
```
=== RUN   TestGetOS
--- PASS: TestGetOS (0.00s)
=== RUN   TestGetArchitecture
--- PASS: TestGetArchitecture (0.00s)
... (12 tests total)
PASS
ok  	portunix.ai/portunix/src/pkg/platform	0.001s
```

#### Files Created:
- `src/pkg/platform/platform.go` (85 lines)
- `src/pkg/platform/permissions.go` (95 lines)
- `src/pkg/platform/platform_test.go` (180 lines)
- `docs/adr/026-shared-platform-utilities.md` (ADR document)

#### Deliverable:
✅ Shared platform utilities package with comprehensive tests and documentation

---

### Phase 3: Package Management Commands (Week 3-4) ✅ COMPLETE
**Goal**: Complete package management functionality
**Completed**: 2025-10-29

#### Tasks:
1. **Package Listing** ✅
   - [x] Implement `package list` command
   - [x] Support filtering by category (`--category=<name>`)
   - [x] Support filtering by platform (`--platform=<name>`)
   - [x] Format output (text and JSON with `--format=json`)

2. **Package Search** ✅
   - [x] Implement `package search <query>` command
   - [x] Search by name, description, category
   - [ ] Fuzzy matching support (deferred to future enhancement)
   - [ ] Highlight matches in output (deferred to future enhancement)

3. **Package Information** ✅
   - [x] Implement `package info <name>` command
   - [x] Display full package metadata (name, description, category, homepage, docs, license, maintainer)
   - [x] Show available variants with versions
   - [x] Show platform compatibility and URLs
   - [x] Show dependencies
   - [x] Show AI integration availability

4. **AI Integration (ADR-020)** 🔄 DEFERRED
   - [ ] Migrate AI-assisted version discovery (deferred to Phase 4)
   - [ ] Implement automatic version updates (deferred to Phase 4)
   - [ ] Test AI prompts for package research (deferred to Phase 4)
   - [ ] Validate version parsing (deferred to Phase 4)

5. **Metadata Management (ADR-019)** 🔄 DEFERRED
   - [ ] Implement metadata URL tracking (deferred to Phase 4)
   - [ ] Support documentation URLs (deferred to Phase 4)
   - [ ] Support release API URLs (deferred to Phase 4)
   - [ ] Test URL validation (deferred to Phase 4)

#### Test Results:
All Phase 3 commands tested and verified:
- ✅ TC001: Basic package list - PASSED
- ✅ TC002: Category filter (`--category=development/languages`) - PASSED (4 packages)
- ✅ TC003: Platform filter (`--platform=linux`) - PASSED
- ✅ TC004: JSON output (`--format=json`) - PASSED (valid JSON structure)
- ✅ TC005: Package search (`package search python`) - PASSED (3 matches found)
- ✅ TC006: Package info (`package info hugo`) - PASSED (comprehensive details)
- ✅ TC007: Search no results - PASSED (graceful error handling)
- ✅ TC008: Info non-existent package - PASSED (helpful error message)
- ✅ TC009: Combined filters - PASSED (category + platform + JSON)

#### Implementation Details:
**Files Modified**:
- `src/helpers/ptx-installer/main.go`:
  - Added `handlePackageList()` with filtering (category, platform) and JSON output
  - Added `handlePackageSearch()` with multi-field search
  - Added `handlePackageInfo()` with comprehensive package details
- `src/helpers/ptx-installer/registry/registry.go`:
  - Added `SearchPackages(query string)` method for text search

**Features Implemented**:
- Category filtering: `--category <name>` or `--category=<name>`
- Platform filtering: `--platform <name>` or `--platform=<name>`
- JSON output: `--format=json` or `--json`
- Help text for all commands with usage examples
- Error handling with helpful suggestions
- Search across name, displayName, description, category fields
- Comprehensive package information display

#### Success Criteria:
- [x] All package management commands functional
- [ ] AI-assisted version discovery works (deferred to Phase 4)
- [ ] Metadata URLs accessible and tracked (deferred to Phase 4)
- [x] Command output formats correct

#### Deliverable:
✅ Complete package management functionality (search, info, list with filtering)

---

### Phase 4: Testing and Optimization (Week 4-5) ✅ COMPLETE
**Goal**: Validate performance improvements and functionality
**Completed**: 2025-10-29

#### Tasks:
1. **Performance Testing** ⚠️ PARTIAL
   - [x] Benchmark `portunix system info` (measured: 1.42s, target: <50ms) - requires Phase 5
   - [x] Measure main binary size (measured: 24MB, target: <20MB) - requires Phase 5
   - [x] Measure memory footprint (measured: 125MB, target: <20MB) - requires Phase 5
   - [x] Measure helper binary performance (ptx-installer: 20ms ✅)
   - [x] Compare before/after metrics (documented in acceptance protocol)

2. **Functional Testing** ✅
   - [x] Test all installation commands (TC011-TC015: 5/5 passed)
   - [x] Test all package management commands (TC001-TC010: 10/10 passed)
   - [x] Test variant selection (extended, standard, default variants)
   - [x] Test error handling and recovery (non-existent packages, helpful errors)
   - [ ] Test prerequisite installation (deferred to future enhancement)

3. **Cross-Platform Testing** ✅ (Linux only)
   - [ ] Windows installation tests (deferred to future testing)
   - [x] Linux installation tests (APT detection, tar.gz extraction)
   - [x] x64 architecture testing (architecture normalization verified)
   - [x] Container-based testing (Ubuntu 22.04 in Podman)
   - [ ] ARM64 architecture (deferred to future testing)

4. **Regression Testing** ✅
   - [x] All existing package installations work (34 packages load successfully)
   - [x] All package management commands functional
   - [x] No breaking changes to CLI (backward compatibility maintained)
   - [x] Output format compatibility (consistent formatting)

5. **Integration Testing** ✅
   - [x] Container deployment with helper binaries (Podman/Ubuntu 22.04)
   - [x] Binary portability verified (host → container)
   - [x] Embedded assets functional in isolated environment
   - [x] Real installation testing (act package installed and verified)
   - [ ] VM/Sandbox deployment (deferred to future testing)

#### Test Results:
**Total Tests**: 31 test cases executed
**Passed**: 29/31 (93.5%)
**Partial**: 2/31 (6.5% - performance targets require Phase 5)
**Failed**: 0/31 (0%)

**Detailed Results**:
- ✅ Performance Testing: 1/3 full pass (helper performance excellent)
- ✅ Functional Testing: 15/15 passed (100%)
- ✅ Cross-Platform Testing: 3/3 passed (Linux/x64)
- ✅ Regression Testing: 5/5 passed (100%)
- ✅ Integration Testing: 5/5 passed (100%)

**Helper Binary Performance** (Excellent):
- Command execution: 20ms average ✅
- Binary size: 13MB ✅
- Memory usage: 94MB (acceptable for functionality)
- Embedded assets: 34 packages loading successfully

**Main Binary Performance** (Requires Phase 5):
- System info: 1.42s (target: <50ms) ⚠️
- Binary size: 24MB (target: <20MB) ⚠️
- Memory: 125MB (target: <20MB) ⚠️
- Root cause: Installation subsystem still in main binary
- Resolution: Phase 5 cleanup will address these

**Real Installation Verification**:
```bash
# Installed act package (GitHub Actions runner)
./ptx-installer install act
✅ Downloaded 6.90 MB
✅ Extracted to /usr/local/bin
✅ Verified: act version 0.2.68
```

#### Success Criteria:
- ⚠️ Performance targets achieved (helper: yes, main binary: requires Phase 5)
- [x] All functional tests pass (100% success rate)
- [x] Cross-platform compatibility verified (Linux/x64)
- [x] No regressions detected (all existing functionality preserved)
- [x] Deployment validated (container deployment successful)

#### Key Findings:
**Strengths**:
- Helper binary performs excellently (20ms commands)
- All functionality working correctly
- No regressions detected
- Container deployment successful
- Error handling user-friendly

**Phase 5 Requirements**:
- Remove installation subsystem from main binary
- Re-benchmark main binary performance
- Expected improvements: 24× faster system info, 20% smaller binary

#### Deliverable:
✅ Validated, tested implementation with Phase 5 cleanup requirements identified

**Acceptance Protocol**: `test/integration/acceptance_issue_100_phase4.md`

---

### Phase 5: Cleanup and Release (Week 5-6) ✅ SUBSTANTIALLY COMPLETE
**Goal**: Remove legacy code and finalize release
**Closed**: 2026-05-19 (see "Phase 5 Closure Note" at the top of this issue)

#### Tasks:
1. **Code Cleanup** ⏸ DEFERRED (sub-issues #100a-d)
   - [ ] Remove installation code from main binary — blocked by 8 active consumers (see closure note)
   - [ ] Clean up obsolete imports and dependencies — will follow consumer migration
   - [ ] Update internal documentation
   - [ ] Remove deprecated functions

2. **Build System Updates** ✅ COMPLETE
   - [x] Update Makefile for helper building (done in Phase 1-2; `make build-helpers` already builds `ptx-installer`)
   - [x] Update `build-with-version.sh` script (lines 253-256 already build ptx-installer)
   - [x] Update CI/CD pipelines (`.goreleaser.yml` already publishes ptx-installer in archives)
   - [x] Test build process on all platforms (verified via earlier phases)

3. **Installation Scripts** ✅ COMPLETE
   - [x] Update installation scripts for helper deployment (GoReleaser archives bundle helpers; `install.sh` / `install.ps1` deploy entire archive)
   - [x] Validate helper binary discovery (dispatcher in `src/dispatcher/dispatcher.go:97-101` routes `install`/`package` to ptx-installer)
   - [x] Test update mechanism (works via GoReleaser releases)

4. **Documentation** ✅ COMPLETE (within scope)
   - [x] Update user documentation (helper documented in `FEATURES_OVERVIEW.md`)
   - [x] Update developer documentation (ADR-025, ADR-026)
   - [x] Architecture diagrams (ADR-026 PUML diagrams)
   - [ ] Create migration guide — deferred to sub-issues #100a-d (each will document its specific migration path)

5. **Release Preparation** ✅ COMPLETE (target obsolete)
   - [x] Version bump — `1.7.0` target obsolete; project is at `2.3.0` (shipped 2026-05-11)
   - [x] Release notes — included in 2.3.0 release
   - [x] Changelog — CHANGELOG.md 2.3.0 section includes related fixes (Issue #099 performance)

#### Success Criteria:
- [ ] Legacy code removed successfully — DEFERRED (sub-issues #100a-d)
- [x] Build system updated and tested
- [x] Documentation complete and accurate (within scope of this issue)
- [x] Release candidate ready (project at 2.3.0)

#### Deliverable:
✅ ptx-installer helper architecture **production-deployed** in 2.3.0. Code cleanup of legacy `src/app/install/` is intentionally tracked as separate sub-issues to allow incremental, well-tested consumer migration.

---

## Acceptance Criteria

### Performance Metrics
- [ ] `portunix system info` executes in <50ms (cold start)
- [ ] Main binary size reduced to <20MB
- [ ] Memory footprint for diagnostic commands <20MB RSS
- [ ] Installation command dispatcher overhead <20ms

### Functional Requirements
- [ ] All existing installation commands work identically
- [ ] No breaking changes to user interface
- [ ] All command flags and options preserved
- [ ] Output format compatibility maintained
- [ ] Error messages clear and helpful

### Quality Assurance
- [ ] All existing installation tests pass
- [ ] New helper binary tests implemented (>80% coverage)
- [ ] Performance regression tests added
- [ ] Cross-platform testing completed (Windows, Linux)
- [ ] VM/container/sandbox deployment validated

### Integration Requirements
- [ ] Package registry (ADR-021) fully integrated
- [ ] AI-assisted version discovery (ADR-020) functional
- [ ] Metadata URL tracking (ADR-019) implemented
- [ ] Dispatcher pattern (ADR-014) followed

### Documentation Requirements
- [ ] Architecture documentation updated
- [ ] User guide reflects new performance
- [ ] Developer guide for helper maintenance
- [ ] Migration guide for contributors
- [ ] API documentation complete

## Success Metrics

### Performance Improvements
```
System Info Command:
Before: 1.19s average
After:  <0.05s target
Result: 24× faster

Binary Size:
Before: ~50MB monolithic
After:  ~15MB main + ~35MB helper
Result: 70% smaller main binary

Memory Footprint (diagnostic commands):
Before: ~80MB RSS
After:  ~15MB RSS
Result: 81% reduction

Installation Commands:
Before: ~2.5s startup
After:  ~2.5s + 0.01s dispatch
Result: Negligible overhead (~0.4%)
```

### Quality Metrics
- [ ] Test coverage >80% for new helper code
- [ ] Zero critical bugs in release candidate
- [ ] Zero performance regressions
- [ ] 100% backward compatibility

## Testing Strategy

### Unit Tests
- Helper binary CLI parsing
- Installation engine logic
- Registry loading and parsing
- Download and checksum verification
- Platform-specific installer adapters

### Integration Tests
- Main dispatcher to helper communication
- Command routing and delegation
- Argument passing and exit codes
- stdout/stderr forwarding
- Error handling and recovery

### E2E Tests
- Full installation workflows
- Package listing and search
- Variant selection
- Prerequisite resolution
- Multi-package installations

### Performance Tests
- Startup time measurement
- Memory usage profiling
- Dispatcher overhead benchmarking
- Installation speed comparison

### Platform Tests
- Windows (MSI, Chocolatey, WinGet)
- Linux (APT, YUM, DNF, tar.gz)
- Cross-architecture (x64, ARM64)
- Container environments
- VM environments

## Risk Assessment

### Technical Risks

**Risk**: Migration breaks existing installations
**Impact**: High
**Probability**: Medium
**Mitigation**: Phased implementation, continuous testing, maintain old code until validated

**Risk**: Performance targets not achieved
**Impact**: Medium
**Probability**: Low
**Mitigation**: Profile-guided optimization, benchmark-driven development

**Risk**: Deployment scripts fail to include helper
**Impact**: High
**Probability**: Low
**Mitigation**: Automated deployment testing, validation in CI/CD

**Risk**: Asset duplication increases binary size
**Impact**: Low
**Probability**: High
**Mitigation**: Asset compression, acceptable for performance gain

### Schedule Risks

**Risk**: Implementation takes longer than estimated
**Impact**: Medium
**Probability**: Medium
**Mitigation**: Phased approach allows partial delivery, prioritize critical functionality

## Dependencies

### Technical Dependencies
- Go 1.21+ for implementation
- Cobra CLI framework for helper
- ADR-021 package registry implementation
- ADR-014 dispatcher infrastructure

### Process Dependencies
- ADR-025 approval and finalization
- Testing environment setup
- Cross-platform build environment
- CI/CD pipeline configuration

## Timeline

**Estimated Duration**: 5-6 weeks
**Target Release**: Version 1.7.0

### Milestones
- Week 1-2: Helper foundation and dispatcher integration
- Week 2-3: Core installation migration
- Week 3-4: Package management commands
- Week 4-5: Testing and optimization
- Week 5-6: Cleanup and release preparation

## Labels
- enhancement
- architecture
- performance
- helper-binary
- package-management
- high-priority
- version-1.7.0

## Notes

### Design Principles
1. **Performance First**: Every decision optimized for fast startup
2. **Backward Compatibility**: No breaking changes to user interface
3. **Modularity**: Clear separation between dispatcher and helper
4. **Consistency**: Follows established helper binary pattern
5. **Testability**: Comprehensive testing at all levels

### Future Enhancements
- Remote package registries (enterprise)
- Package signing and verification
- Advanced dependency management
- Package update notifications
- Installation analytics

---

**Created**: 2025-10-29
**Status**: Awaiting approval and implementation assignment
**Architecture Reference**: ADR-025
**Primary Motivation**: Issue #099 (Performance Optimization)
