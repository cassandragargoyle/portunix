# Architecture Decision Record Diagrams

This directory contains PlantUML diagrams for Architecture Decision Records (ADRs).

## Viewing Diagrams

### Option 1: PlantUML Online Server
Visit [PlantUML Web Server](http://www.plantuml.com/plantuml/uml/) and paste the diagram code.

### Option 2: VS Code Extension
Install the [PlantUML extension](https://marketplace.visualstudio.com/items?itemName=jebbs.plantuml) and preview diagrams directly in VS Code.

### Option 3: Local PlantUML
```bash
# Install PlantUML
sudo apt install plantuml  # Debian/Ubuntu
brew install plantuml      # macOS

# Generate PNG
plantuml diagram.puml

# Generate SVG
plantuml -tsvg diagram.puml
```

## ADR-026: Shared Platform Utilities

### 026-architecture-before-after.puml
**Component Diagram** showing the architecture before and after refactoring:
- **Before**: Duplicated platform code in main binary and helper
- **After**: Shared `src/pkg/platform/` package used by all binaries

**Key Points**:
- Eliminates ~100 lines of duplicate code
- Single source of truth for platform utilities
- Comprehensive test coverage (12 unit tests)

### 026-platform-api.puml
**Class Diagram** showing the Platform Package API structure:
- `platform.go` - OS and architecture detection functions
- `permissions.go` - Privilege and permission utilities
- `platform_test.go` - Unit tests for all public functions

**Public API**:
- OS Detection: `GetOS()`, `GetArchitecture()`, `GetPlatform()`
- Boolean Checks: `IsWindows()`, `IsLinux()`, `IsDarwin()`, `IsWindowsSandbox()`
- Permission Checks: `IsRunningAsRoot()`, `IsSudoAvailable()`, `GetSudoPrefix()`
- Directory Utils: `CanWriteToDirectory()`, `IsUserDirectory()`

### 026-dependency-flow.puml
**Package Diagram** showing dependencies between components:
- Go Runtime → Platform Package → Helper Binaries
- Shows how platform package depends on Go standard library
- Displays backward compatibility wrappers
- Illustrates future helper binary usage

**Relationships**:
- Platform package uses Go runtime (`runtime.GOOS`, `runtime.GOARCH`)
- PTX-Installer delegates to platform package
- Future helpers can directly import platform package
- Legacy code can be migrated in Phase 5

### 026-installation-sequence.puml
**Sequence Diagram** showing installation flow using platform utilities:
- User executes `portunix install hugo`
- Dispatcher routes to ptx-installer helper
- Helper queries platform package for OS/architecture
- Permission checks determine sudo requirement
- Installation proceeds with appropriate privileges

**Demonstrates**:
- Platform detection during installation
- Architecture normalization (amd64 → x64)
- Permission checking workflow
- Sudo prefix determination

### 026-deployment.puml
**Deployment Diagram** showing binary distribution:
- Development: Shared `src/pkg/platform/` source code
- Build: Platform utilities compiled into binaries
- Deployment: Self-contained binaries with embedded platform code

**Key Aspects**:
- Platform utilities embedded in helper binaries
- No separate library needed at runtime
- Main binary remains lightweight
- Helper contains all necessary utilities

## Diagram Conventions

### Colors
- **Light Blue**: Main binary components
- **Light Green**: Helper binary components (after refactoring)
- **Light Coral**: Helper binary components (before refactoring)
- **Light Yellow/Gold**: Shared packages
- **Red dashed**: Code that should be migrated
- **Green dotted**: Future usage patterns

### Symbols
- `→` Solid arrow: Direct dependency/compilation
- `..>` Dotted arrow: Usage/delegation
- `<<shared>>` Stereotype: Shared package
- `<<duplicate>>` Stereotype: Duplicated code (problematic)
- `✓` Checkmark: Completed/implemented
- `[ ]` Square brackets: Component or module

## Generating All Diagrams

```bash
# Generate all PNG diagrams
for file in *.puml; do
    plantuml "$file"
done

# Generate all SVG diagrams
for file in *.puml; do
    plantuml -tsvg "$file"
done
```

## Updating Diagrams

When updating diagrams:
1. Edit the `.puml` source file
2. Regenerate the output (PNG/SVG)
3. Update references in ADR document if needed
4. Commit both source and generated files

## Tools and Resources

- **PlantUML Official**: https://plantuml.com/
- **PlantUML Guide**: https://plantuml.com/guide
- **Themes**: https://plantuml.com/theme
- **VS Code Extension**: https://marketplace.visualstudio.com/items?itemName=jebbs.plantuml
- **Online Editor**: http://www.plantuml.com/plantuml/uml/

---

**Last Updated**: 2025-10-29
**Diagrams**: 5 (ADR-026)
**Format**: PlantUML (.puml)
