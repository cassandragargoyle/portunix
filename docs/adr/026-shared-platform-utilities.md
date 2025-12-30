# ADR-026: Shared Platform Utilities Package

**Status**: Accepted
**Date**: 2025-10-29
**Author**: Development Team
**Related Issues**: #100 (PTX-Installer Helper Implementation)
**Related ADRs**: ADR-025 (PTX-Installer Helper Architecture), ADR-014 (Git-like Dispatcher Pattern)

---

## Diagrams

### Architecture Overview
![Architecture Before and After](diagrams/026-architecture-before-after.puml)

### Platform API
![Platform Package API](diagrams/026-platform-api.puml)

### Dependency Flow
![Dependency Flow](diagrams/026-dependency-flow.puml)

### Installation Sequence
![Installation Sequence](diagrams/026-installation-sequence.puml)

### Deployment Architecture
![Deployment Architecture](diagrams/026-deployment.puml)

---

## Context

During implementation of Issue #100 (PTX-Installer Helper), we discovered code duplication in platform detection and architecture normalization utilities:

### Duplicated Code Locations

1. **src/app/install/config.go:246** - `GetArchitecture()`
   - Used by main portunix binary
   - Normalizes `amd64` → `x64`, `386` → `x86`, `arm64` → `arm64`
   - Part of legacy installation system

2. **src/helpers/ptx-installer/engine/platform.go:32** - `GetArchitecture()`
   - Used by ptx-installer helper binary
   - Identical logic to main binary
   - Created during Phase 2 implementation

### Additional Platform Utilities

Other platform-related functions exist across the codebase:
- OS detection (`GetOperatingSystem()`)
- Privilege checking (`IsRunningAsRoot()`, `IsSudoAvailable()`)
- Permission handling (`CanWriteToDirectory()`)
- Path normalization
- Environment detection

### Problem Statement

**Code Duplication Issues:**
- Same logic maintained in multiple places
- Risk of divergence as code evolves
- Increased maintenance burden
- Inconsistent behavior if one implementation changes

**Architecture Issues:**
- No clear ownership of platform utilities
- Helper binaries can't share code with main binary
- Violates DRY (Don't Repeat Yourself) principle
- Testing requires duplicate test suites

**Example of Duplication:**
```go
// src/app/install/config.go:246
func GetArchitecture() string {
    switch runtime.GOARCH {
    case "amd64":
        return "x64"
    case "386":
        return "x86"
    case "arm64":
        return "arm64"
    default:
        return "x64"
    }
}

// src/helpers/ptx-installer/engine/platform.go:32
func GetArchitecture() string {
    arch := runtime.GOARCH
    switch arch {
    case "amd64":
        return "x64"
    case "386":
        return "x86"
    case "arm64":
        return "arm64"
    default:
        return arch
    }
}
```

---

## Decision

Create a shared platform utilities package at `src/pkg/platform` that can be imported by both the main portunix binary and all helper binaries.

### Package Structure

```
src/pkg/platform/
├── platform.go         # OS and architecture detection
├── permissions.go      # Privilege and permission utilities
├── paths.go           # Path normalization and validation
└── platform_test.go   # Comprehensive test suite
```

### API Design

#### Core Functions

```go
package platform

// OS and Architecture Detection
func GetOS() string                    // Returns: "windows", "linux", "darwin"
func GetArchitecture() string          // Returns: "x64", "x86", "arm64"
func GetPlatform() string              // Returns: "linux-x64", "windows-x64", etc.
func IsWindows() bool
func IsLinux() bool
func IsDarwin() bool

// Privilege and Permissions
func IsRunningAsRoot() bool
func IsSudoAvailable() bool
func RequiresSudo(path string) bool
func CanWriteToDirectory(path string) bool
func GetSudoPrefix() string            // Returns: "sudo " or ""

// Path Utilities
func NormalizePath(path string) string
func IsUserDirectory(path string) bool
func GetUserHomeDir() (string, error)
func GetCacheDir() (string, error)
func GetConfigDir() (string, error)

// Environment Detection
func IsContainerized() bool
func IsSandboxEnvironment() bool
func GetEnvironmentType() string       // Returns: "native", "container", "vm", "sandbox"
```

### Implementation Approach

1. **Create Package**: `src/pkg/platform/`
2. **Move Shared Code**: Extract from existing locations
3. **Add Missing Functions**: Implement common patterns
4. **Comprehensive Tests**: Unit tests for all functions
5. **Refactor Consumers**: Update import paths
6. **Document API**: Clear documentation for all functions

### Import Example

```go
// Main binary
import "portunix.ai/portunix/src/pkg/platform"

arch := platform.GetArchitecture()  // "x64"
os := platform.GetOS()               // "linux"

// Helper binary
import "portunix.ai/portunix/src/pkg/platform"

if platform.RequiresSudo("/usr/local/bin") {
    cmd := platform.GetSudoPrefix() + "install"
}
```

---

## Rationale

### Benefits

1. **Code Reuse**
   - Single implementation shared across binaries
   - Reduced maintenance burden
   - Consistent behavior guaranteed

2. **Improved Testing**
   - One comprehensive test suite
   - Higher test coverage
   - Easier to test edge cases

3. **Better API**
   - Clear, documented interface
   - Consistent naming conventions
   - Easy to discover functions

4. **Scalability**
   - Easy to add new helper binaries
   - Platform utilities available to all components
   - Future-proof architecture

5. **Maintainability**
   - Changes in one place
   - Clear ownership
   - Reduced risk of bugs from divergence

### Alternatives Considered

#### Alternative 1: Keep Duplication
**Rejected** - Maintenance burden too high, risk of divergence

#### Alternative 2: Use Internal Package
**Rejected** - `internal/` packages can't be imported by external tools, limits future flexibility

#### Alternative 3: Vendor External Library
**Rejected** - No existing library matches our needs, adds dependency

#### Alternative 4: Use src/common/
**Considered** - Similar to `src/pkg/`, but `pkg/` is more standard Go convention

---

## Implementation Plan

### Phase 1: Package Creation (Immediate)
- [x] Create `src/pkg/platform/` directory
- [x] Implement core platform detection functions
- [x] Add comprehensive unit tests
- [x] Document API with examples

### Phase 2: Code Migration (Immediate)
- [x] Update `src/helpers/ptx-installer/engine/` to use shared package
- [x] Update `src/app/install/` to use shared package
- [x] Remove duplicated code
- [x] Verify all tests pass

### Phase 3: Enhancement (Phase 5)
- [ ] Add environment detection functions
- [ ] Add path normalization utilities
- [ ] Extend test coverage to edge cases
- [ ] Add benchmarks for performance-critical functions

---

## Consequences

### Positive

- **Reduced Code Duplication**: Single source of truth for platform utilities
- **Improved Maintainability**: Changes in one place affect all consumers
- **Better Testing**: Comprehensive test suite for platform functions
- **Consistent Behavior**: All binaries use same detection logic
- **Easier Development**: New helpers can import ready-made utilities

### Negative

- **Import Dependency**: All binaries now depend on `src/pkg/platform`
- **Breaking Change**: Existing code must be refactored
- **Build Complexity**: Slightly more complex dependency graph

### Mitigation

- Keep package small and focused
- Maintain backward compatibility where possible
- Document migration path clearly
- Test extensively before deployment

---

## Technical Details

### Architecture Normalization

**Mapping Table:**
| Go GOARCH | Normalized | Used In |
|-----------|------------|---------|
| `amd64`   | `x64`      | Package registry, installers |
| `386`     | `x86`      | Legacy 32-bit systems |
| `arm64`   | `arm64`    | ARM-based systems (Apple Silicon, Pi) |
| `arm`     | `arm`      | Older ARM systems |

**Fallback Strategy:**
- Unknown architectures return Go's native GOARCH
- Allows forward compatibility with new architectures

### OS Detection

**Environment Variables:**
- `PORTUNIX_SANDBOX=windows` → `windows_sandbox`
- Standard Go `runtime.GOOS` for native detection

**Special Cases:**
- Windows Sandbox: Detected via environment variable
- WSL: Detected as Linux (matches package availability)
- Container: Detected via `/proc/1/cgroup` or `/.dockerenv`

### Permission Detection

**Unix-like Systems:**
- Root: `os.Geteuid() == 0`
- Sudo: Check `/usr/bin/sudo` existence
- Write permission: Attempt to create test file

**Windows:**
- Administrator: Windows API call (future enhancement)
- Simplified check for Phase 2 implementation

---

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "portunix.ai/portunix/src/pkg/platform"
)

func main() {
    // Get platform information
    fmt.Printf("OS: %s\n", platform.GetOS())
    fmt.Printf("Arch: %s\n", platform.GetArchitecture())
    fmt.Printf("Platform: %s\n", platform.GetPlatform())

    // Check permissions
    if platform.IsRunningAsRoot() {
        fmt.Println("Running with elevated privileges")
    }

    if platform.RequiresSudo("/usr/local/bin") {
        prefix := platform.GetSudoPrefix()
        fmt.Printf("Use sudo: %s\n", prefix)
    }
}
```

### Installation Use Case

```go
func installToDirectory(targetDir string) error {
    // Check if we can write to directory
    if !platform.CanWriteToDirectory(targetDir) {
        if platform.IsSudoAvailable() {
            return installWithSudo(targetDir)
        }
        // Fallback to user directory
        homeDir, _ := platform.GetUserHomeDir()
        targetDir = filepath.Join(homeDir, ".local", "bin")
    }

    return performInstallation(targetDir)
}
```

---

## Testing Strategy

### Unit Tests
```go
func TestGetArchitecture(t *testing.T) {
    arch := platform.GetArchitecture()
    validArchs := []string{"x64", "x86", "arm64", "arm"}
    assert.Contains(t, validArchs, arch)
}

func TestGetOS(t *testing.T) {
    os := platform.GetOS()
    validOS := []string{"windows", "linux", "darwin"}
    assert.Contains(t, validOS, os)
}
```

### Integration Tests
- Test in Docker containers (Linux)
- Test in Windows Sandbox
- Test with different architectures (via QEMU)

---

## Migration Checklist

- [x] Create ADR-026 document
- [x] Create `src/pkg/platform/` package
- [x] Implement core functions
- [x] Add unit tests
- [x] Update ptx-installer to use shared package
- [x] Update main binary to use shared package
- [x] Remove duplicated code
- [x] Update Issue #100 with ADR reference
- [x] Test all functionality
- [x] Document API
- [x] Commit changes

---

## References

- **Issue #100**: PTX-Installer Helper Implementation
- **ADR-025**: PTX-Installer Helper Architecture
- **ADR-014**: Git-like Dispatcher Pattern
- **Go Project Layout**: https://github.com/golang-standards/project-layout

---

## Approval

**Status**: ✅ Accepted
**Decision Date**: 2025-10-29
**Implementation**: Immediate (Phase 2 of Issue #100)
**Review**: No additional review required (internal refactoring)

---

## Revision History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-10-29 | 1.0 | Initial version | Development Team |
