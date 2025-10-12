# ADR-014: Git-like Dispatcher with Python Distribution Model

**Status**: Proposed
**Date**: 2025-09-19
**Author**: Zdeněk Kurc

## Context

Portunix has grown into a complex system with multiple subsystems (Docker, Podman, MCP, plugins, sandbox, virtual machines, etc.). While the project already implements a gRPC-based plugin system for external extensions, the core functionality remains in a single binary. As the system continues to grow, we face several challenges:

1. **Binary size growth**: Each new core feature increases the main binary size
2. **Dependency conflicts**: Different subsystems may require incompatible dependencies
3. **Development complexity**: All developers must understand the entire core codebase
4. **Testing overhead**: Changes to core functionality require testing the entire system
5. **Extension variety**: While gRPC plugins work well for complex extensions, simple utilities would benefit from a lighter approach

Git provides an excellent architectural pattern: from the user's perspective, it appears as a single command (`git`), but internally it's a dispatcher that delegates to specialized helper binaries (`git-commit`, `git-push`, etc.). This architecture has proven successful for over 15 years.

Python provides an equally successful distribution pattern: all related binaries (`python`, `pip`, `python3.11`, `pip3.11`) are installed in the same directory, making them easily discoverable and manageable. This cross-platform approach eliminates complex directory structures and symlink dependencies while maintaining version coherence.

## Decision

We will adopt a Git-like dispatcher architecture combined with Python distribution model for Portunix:

### Binary Naming Convention
```
Main dispatcher:     portunix
Helper binaries:     ptx-{subsystem}
gRPC plugins:        ptx-plugin-{name}

Examples:
- portunix           # Main dispatcher
- ptx-container      # Helper: Container management (Docker, Podman, unified interface)
- ptx-mcp            # Helper: MCP server subsystem
- ptx-plugin-agile   # Plugin: Agile development plugin
- ptx-plugin-github  # Plugin: GitHub integration plugin
```

### Binary Type Distinction
- **Helpers**: Integral parts of Portunix, distributed with main installation, version-locked
- **Plugins**: External extensions, independent installation and versioning, managed via `portunix plugin` commands

### 1. Core Dispatcher
- Main `portunix` binary acts as a lightweight dispatcher
- Handles command parsing, help system, and routing
- Discovers and invokes helper binaries based on command structure
- Manages gRPC plugin lifecycle and communication
- Maintains backward compatibility with existing command interface

### 2. Binary Discovery
**Single-Version Directory Structure (Simple installations):**
```
Linux:   /usr/local/bin/
Windows: C:\Program Files\Portunix\bin\

Contents:
├── portunix[.exe]           # Main dispatcher
├── ptx-container[.exe]      # Helper: Container management (Docker and Podman)
├── ptx-mcp[.exe]            # Helper: MCP server
├── ptx-plugin-agile[.exe]   # gRPC plugins
└── ptx-plugin-github[.exe]
```

**Discovery Logic:**
1. Look for helper/plugin binaries in the same directory as main `portunix` binary
2. Fallback to system `$PATH` if not found in primary location
3. Validate version compatibility between main binary and helpers

**Benefits:**
- Cross-platform compatible (no symlinks needed)
- Simple installation and updates (all files in one location)
- Proven pattern used by Python, Node.js, Go and other successful tools
- Easy version management (update all binaries together)

### Version Management and Multiple Installations

**Version Compatibility Scenarios:**
```
Scenario 1: Clean Installations (Supported)
/usr/local/bin/ → portunix (v1.5.14) + matching helpers
~/bin/ → portunix (v1.6.0) + matching helpers
→ PATH priority determines active version

Scenario 2: Mixed Versions (Problematic - Detected & Warned)
/usr/local/bin/
├── portunix (v1.6.0)        # Main dispatcher
├── ptx-container (v1.5.14)  # ⚠️ Version mismatch
└── ptx-mcp (v1.5.14)        # ⚠️ Version mismatch
```

**Version Detection and Validation:**
- Main dispatcher validates helper binary versions on startup
- Warnings issued for version mismatches
- Graceful degradation when helpers are unavailable or incompatible
- `portunix system info` displays version information for all components

**Multi-Version Support (Python Launcher Pattern):**

Following Python's `py.exe` launcher approach for version-aware dispatching:

**Enhanced Main Binary Requirements:**
- **Registry reading capability** (Windows) for version discovery
- **Version selection** via `--use-version=X.Y.Z` parameter
- **Version enumeration** via `portunix versions` command
- **Default version management** with fallback logic

**Usage Examples:**
```bash
portunix --version                           # Current default version
portunix versions                            # All installed versions
portunix --use-version=1.5.14 container run # Specific version
```

### Multi-Version Installation Examples

**Linux Installation Structure:**
```bash
# System installation (via package manager)
/usr/local/portunix-1.5.14/bin/
├── portunix
├── ptx-container
└── ptx-mcp

/usr/local/portunix-1.6.0/bin/
├── portunix
├── ptx-container
└── ptx-mcp

# Main binary symlink (points to default version)
/usr/local/bin/portunix -> /usr/local/portunix-1.6.0/bin/portunix

# User can switch versions:
ln -sf /usr/local/portunix-1.5.14/bin/portunix /usr/local/bin/portunix

# Or use PATH preference:
export PATH="/usr/local/portunix-1.5.14/bin:$PATH"
```

**Windows Installation Structure:**
```cmd
# Multiple installations in Program Files
C:\Program Files\Portunix\1.5.14\bin\
├── portunix.exe
├── ptx-container.exe
└── ptx-mcp.exe

C:\Program Files\Portunix\1.6.0\bin\
├── portunix.exe
├── ptx-container.exe
└── ptx-mcp.exe

# Registry entries created by installer:
HKEY_LOCAL_MACHINE\SOFTWARE\Portunix\1.5.14
  InstallPath = "C:\Program Files\Portunix\1.5.14"
  Version = "1.5.14"

HKEY_LOCAL_MACHINE\SOFTWARE\Portunix\1.6.0
  InstallPath = "C:\Program Files\Portunix\1.6.0"
  Version = "1.6.0"
  Default = "1"  # Marks default version

# Main binary in PATH (latest installer wins)
C:\Program Files\Portunix\bin\portunix.exe -> copies from default version

# Version discovery and switching:
portunix versions                    # Reads registry, lists: 1.5.14, 1.6.0 (default)
portunix --use-version=1.5.14 container run  # Uses 1.5.14 binaries
```

**Installation Commands Examples:**
```bash
# Linux package manager approach
apt install portunix                # Latest version (single installation)
apt install portunix=1.5.14         # Specific version (versioned directory)
apt install portunix=1.6.0          # Additional version (alongside previous)
update-alternatives --config portunix  # System version switching

# Windows installer approach
.\portunix-installer.exe            # Latest version (single installation)
.\portunix-1.5.14-installer.exe     # Specific version (versioned + registry)
.\portunix-1.6.0-installer.exe      # Additional version (alongside previous)
```

### 3. Communication Protocol
- **Helper binaries**: stdio/argv interface (like Git) for straightforward command execution
- **gRPC plugins**: Existing architecture for complex external extensions
- **Shared configuration**: Via environment variables and config files

### 4. Hybrid Architecture
Complements the existing gRPC plugin system:
- **Helper binaries** (`ptx-{subsystem}`): stdio/argv interface for core subsystems
- **gRPC plugins** (`ptx-plugin-{name}`): Complex external extensions
- **Core functionality**: Remains in main binary for essential operations

### 5. Core vs Helper Binary Distribution

**What stays in main `portunix` binary:**
- Command dispatcher and routing logic
- Help system and command discovery
- Core configuration management
- Update mechanism (`portunix update`)
- Version information (`portunix version`, `portunix --version`)
- Basic system information (`portunix system info`)
- Shell completion generation (`portunix completion`)
- Plugin management (`portunix plugin` - gRPC plugin lifecycle)
- Sandbox management (`portunix sandbox` - core environment isolation)
- Package installation (`portunix install` - software package management)
- Virtual machine management (`portunix virt` - virtualization platform)
- Core library functions shared across helpers

**What moves to helper binaries (`ptx-*`) - Phase 1:**
- `ptx-container` - Container management subsystem (Docker, Podman, unified interface)
- `ptx-mcp` - MCP server subsystem

**Future extractions (subsequent ADRs):**
Additional subsystems may be extracted in future phases based on separate architectural decisions and implementation experience.

**Rationale:**
- **Core functions**: Essential operations remain in main binary
- **Container management**: Extracted in Phase 1 (well-defined boundaries)
- **MCP server**: Extracted in Phase 1 (clear isolation, independent operation)
- **Future extractions**: Evaluated in subsequent ADRs

### 6. Implementation Phases

**Phase 1: Dispatcher Infrastructure (Version 1.6)**
- Implement dispatcher logic in main binary
- Add helper binary discovery mechanism
- Create shared library for common functionality
- **Version bump to 1.6** to mark architectural transition

**Phase 2: Container System Extraction (Version 1.6.x)**
- Extract unified container management (`ptx-container`)
- Extract MCP server (`ptx-mcp`)
- Maintain compatibility facades in main binary
- Test deployment across VM/container/sandbox environments

**Phase 3: Documentation and Deployment (Version 1.6.x)**
- Update installation scripts and documentation
- Deploy version management features
- Complete migration impact testing

## Consequences

### Positive Consequences

1. **Modularity**: Each subsystem can be developed independently
2. **Smaller binaries**: Users only need helpers for features they use
3. **Parallel development**: Teams can work on different helpers without conflicts
4. **Language flexibility**: Helpers can be written in different languages if needed
5. **Easier testing**: Test individual helpers in isolation
6. **Simpler contributions**: Contributors can focus on specific subsystems
7. **Better performance**: Smaller memory footprint for main dispatcher
8. **Compatibility**: Maintains existing user interface and commands

### Negative Consequences

1. **Distribution complexity**: Multiple binaries to manage
2. **Initial development effort**: Refactoring existing code into helpers
3. **Version synchronization**: Helper compatibility requirements
4. **Installation complexity**: Package managers must handle multiple files


## Implementation Example

```go
// Main dispatcher (portunix)
func main() {
    // Validate helper versions on startup
    validateHelperVersions()

    if len(os.Args) < 2 {
        showHelp()
        return
    }

    command := os.Args[1]

    // Try to find helper binary
    helperName := "ptx-" + command
    helperPath := findHelper(helperName)

    if helperPath != "" {
        // Validate specific helper version before execution
        if !isHelperVersionCompatible(helperPath) {
            log.Warning("Version mismatch for %s, proceeding with caution", helperName)
        }

        // Delegate to helper binary
        cmd := exec.Command(helperPath, os.Args[2:]...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        os.Exit(cmd.ProcessState.ExitCode())
    }

    // Fallback to built-in commands or error
    handleBuiltinCommand(command, os.Args[2:])
}

func validateHelperVersions() {
    expectedVersion := getCurrentVersion()
    helpers := []string{"ptx-container", "ptx-mcp"}

    for _, helper := range helpers {
        if path := findHelper(helper); path != "" {
            if version := getHelperVersion(path); version != expectedVersion {
                log.Warning("Version mismatch: %s is %s, expected %s", helper, version, expectedVersion)
            }
        }
    }
}
```

## Migration Impact on Existing Functionality

**⚠️ CRITICAL MIGRATION REQUIREMENT:**

The adoption of this architecture requires significant updates to existing Portunix deployment functionality:

### Affected Systems
1. **VM Installation (`portunix virt`)**
   - Current: Copies single `portunix` binary to VM
   - Required: Copy complete binary set (dispatcher + helpers)
   - Update: VM provisioning scripts and procedures

2. **Container Integration (`portunix container`)**
   - Current: Mounts or copies single binary to containers
   - Required: Deploy complete binary set (dispatcher + helpers)
   - Update: Container image building and runtime mounting logic

3. **Sandbox Environment (`portunix sandbox`)**
   - Current: Provides single binary in isolated environment
   - Required: Ensure all helper binaries are available in sandbox
   - Update: Sandbox initialization and binary provisioning

### Implementation Requirements
The `portunix` dispatcher must provide helper methods for:
- Identifying all required helper binaries for deployment
- Bundling complete binary set for distribution
- Verifying binary completeness in target environments
- Providing deployment manifests for VM/container/sandbox setup

### Migration Tasks
- [ ] Update VM provisioning to deploy full binary set
- [ ] Modify container integration to handle multiple binaries
- [ ] Adapt sandbox initialization for helper binary availability
- [ ] Create deployment assistance commands in main dispatcher
- [ ] Update installation documentation and scripts
- [ ] Test binary deployment across all target environments

## Code Structure Reorganization

**Current Issue**: Code and non-code files are mixed in the root directory, making the project structure unclear.

**Proposed Structure:**
```
portunix/
├── src/                    # All source code
│   ├── cmd/               # CLI commands (main dispatcher)
│   ├── app/               # Application logic
│   ├── internal/          # Internal packages
│   ├── pkg/               # Public packages
│   └── helpers/           # Helper binaries source
│       ├── container/     # ptx-container source
│       └── mcp/           # ptx-mcp source
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
├── test/                  # Tests
├── assets/                # Static assets
├── go.mod                 # Go module definition
├── main.go -> src/cmd/main.go  # Symlink for Go convention
└── README.md              # Project documentation
```

**Benefits:**
- Clear separation between source code and project files
- Better organization for helper binaries development
- Easier navigation for contributors
- Standard Go project structure

**Migration Task:**
- [ ] Reorganize code structure with `src/` directory before Phase 2 implementation

### Versioning Strategy

This architectural change represents a significant milestone requiring a version bump:

**Current Version**: 1.5.x
**Target Version**: 1.6.0

**Version Significance:**
- **1.6.0**: Introduces dispatcher architecture and helper binary system
- **1.6.x**: Incremental releases with helper binary extractions
- **Backward Compatibility**: Maintained throughout 1.6.x series
- **Breaking Changes**: Reserved for future major versions (2.0+)

## Migration Strategy

1. **Maintain backward compatibility** throughout migration
2. **Reorganize code structure** with `src/` directory
3. **Start with new features** as helper binaries
4. **Gradually extract existing subsystems** based on stability and isolation
5. **Keep critical core features** in main binary (update, basic help)
6. **Document helper binary API** for external developers

## Success Criteria

- [ ] Main binary size reduced by at least 50%
- [ ] Helper binaries can be developed independently
- [ ] No breaking changes to user interface
- [ ] Installation process remains simple for end users
- [ ] Performance equal or better than monolithic version
- [ ] Clear documentation for helper binary development

## Long-term Maintenance Requirements

**Pattern Compatibility Monitoring:**
The Portunix development team must periodically monitor and adapt to changes in the foundational patterns this architecture follows:

1. **Git Dispatcher Pattern Evolution**
   - Monitor Git project architecture changes and improvements
   - Evaluate new dispatcher patterns and helper binary management approaches
   - Adapt Portunix dispatcher logic to incorporate proven Git innovations
   - Maintain compatibility with Git-like command delegation patterns

2. **Python Distribution Model Changes**
   - Track Python/pip distribution model evolution and best practices
   - Monitor cross-platform binary distribution patterns in Python ecosystem
   - Adapt binary discovery and installation procedures to match Python standards
   - Ensure compatibility with Python toolchain distribution methods

3. **Review Schedule**
   - Quarterly review of Git dispatcher pattern changes
   - Annual review of Python distribution model evolution
   - Ad-hoc reviews when major changes occur in either foundational pattern
   - Document any architectural adjustments in subsequent ADRs

**Compatibility Commitment:**
Portunix must remain compatible with established patterns while incorporating beneficial innovations from both Git and Python ecosystems.

## References

- Issue #007: Plugin System with gRPC Architecture
- Issue #024: Basic Plugin System
- Git source code architecture: https://github.com/git/git
- Python distribution documentation: https://packaging.python.org/
- Unix philosophy: "Do one thing and do it well"

## Notes

This architecture follows two proven patterns: Git's dispatcher model for command delegation and Python's distribution model for binary management. Both have successfully managed complexity and cross-platform compatibility for decades. By adopting and maintaining compatibility with these patterns, Portunix can scale to support many more features while maintaining simplicity and performance.