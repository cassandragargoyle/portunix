# ADR-007: Prerequisite Package Handling System

## Status
**Accepted** - Implemented as part of Issue #041

## Context

Modern development tools often have complex dependency chains. For example, `claude-code` (Anthropic's CLI) requires Node.js and npm to function. Previously, users had to manually install these dependencies, leading to:

### Problems with Manual Prerequisite Management:
1. **Poor User Experience**: Users encountered cryptic error messages like "npm: command not found"
2. **Installation Failures**: Packages would appear to install successfully but fail at runtime
3. **Documentation Overhead**: Each package needed detailed prerequisite documentation
4. **Support Burden**: Increased user support requests for dependency issues
5. **Inconsistent Behavior**: Some packages handled dependencies, others didn't

### Business Impact:
- **Reduced Adoption**: Complex setup processes deterred users
- **Support Costs**: Increased help requests and troubleshooting
- **User Frustration**: Failed installations without clear guidance
- **Development Friction**: Package maintainers had to implement custom dependency logic

## Decision

We implement an **automated prerequisite handling system** that:
1. **Declaratively defines** package dependencies in `install-packages.json`
2. **Automatically checks** for prerequisite availability before installation
3. **Recursively installs** missing prerequisites with user feedback
4. **Provides clear messaging** about dependency installation progress

### Architecture Design

#### 1. Configuration Schema Extension
Add `prerequisites` field to `PackageConfig`:

```go
type PackageConfig struct {
    Name           string                    `json:"name"`
    Description    string                    `json:"description"`  
    Prerequisites  []string                  `json:"prerequisites,omitempty"` // NEW
    Platforms      map[string]PlatformConfig `json:"platforms"`
    DefaultVariant string                    `json:"default_variant"`
}
```

#### 2. Prerequisite Processing Pipeline
```
Package Installation Request
         ↓
   Load Package Config
         ↓
   Check Prerequisites List
         ↓
   For Each Prerequisite:
     - Check if installed
     - If missing: Install recursively
     - Report progress to user
         ↓
   All Prerequisites Satisfied
         ↓
   Proceed with Main Installation
```

#### 3. Installation Flow Integration
Integrate prerequisite handling into `InstallPackageWithDryRun()`:

```go
func InstallPackageWithDryRun(packageName, variant string, dryRun bool) error {
    // 1. Load configuration
    config, err := LoadInstallConfig()
    
    // 2. Get package info  
    pkg, platform, variantConfig, err := config.GetPackageInfo(packageName, variant)
    
    // 3. Handle prerequisites (NEW)
    if err := handlePrerequisites(config, pkg, dryRun); err != nil {
        return fmt.Errorf("failed to handle prerequisites: %w", err)
    }
    
    // 4. Continue with main installation...
}
```

## Technical Implementation

### Core Functions

#### `handlePrerequisites()`
- **Purpose**: Main orchestration function for prerequisite processing
- **Logic**: Iterates through prerequisites, checks installation status, installs if missing
- **Output**: Progress messages with emoji indicators for user feedback

#### `isPackageInstalled()`  
- **Purpose**: Determines if a prerequisite is already satisfied
- **Method**: Uses package verification commands from configuration
- **Reliability**: Leverages existing verification infrastructure

### User Experience Design

#### Progress Messaging
```
🔍 Checking prerequisites for Claude Code...
   📋 Checking prerequisite: nodejs
   ✅ nodejs is already installed
🎉 All prerequisites satisfied!
```

#### Installation Flow
```
🔍 Checking prerequisites for Claude Code...
   📋 Checking prerequisite: nodejs
   🔄 Installing prerequisite: nodejs
   ✅ nodejs installed successfully  
🎉 All prerequisites satisfied!
```

#### Dry-Run Support
```
🔍 Checking prerequisites for Claude Code...
   📋 Checking prerequisite: nodejs  
   🔄 [DRY-RUN] Would install prerequisite: nodejs
```

## Alternatives Considered

### 1. Package Manager Dependencies
**Pros**: Leverages existing package manager dependency systems
**Cons**: Platform-specific, limited cross-platform support, no unified experience

### 2. Installation Scripts
**Pros**: Maximum flexibility, custom logic per package
**Cons**: Maintenance overhead, platform-specific code, no reusability

### 3. User Documentation Only
**Pros**: No implementation complexity
**Cons**: Poor UX, high support burden, manual process

### 4. External Dependency Manager
**Pros**: Proven solutions exist (e.g., package managers)
**Cons**: Additional dependency, platform limitations, learning curve

## Decision Rationale

### Why Automated Prerequisite Handling:
1. **Unified Experience**: Same behavior across all platforms and packages
2. **Declarative Configuration**: Dependencies defined in JSON, not code
3. **Recursive Resolution**: Handles nested dependencies automatically
4. **Transparent Process**: Clear user feedback at each step
5. **Dry-Run Support**: Users can preview changes before execution

### Why Configuration-Driven:
- **Maintainability**: Dependencies defined alongside package definitions
- **Consistency**: Same format and validation as package configurations
- **Extensibility**: Easy to add new dependency relationships
- **Visibility**: Dependencies clearly documented and version-controlled

## Consequences

### Positive Impacts:
- ✅ **Improved UX**: One-command installation with all dependencies
- ✅ **Reduced Support**: Fewer user issues related to missing dependencies  
- ✅ **Faster Onboarding**: Users can get started immediately
- ✅ **Consistent Behavior**: All packages follow same dependency model
- ✅ **Better Documentation**: Dependencies are self-documenting in JSON

### Potential Risks:
- ⚠️ **Circular Dependencies**: Risk of infinite loops in dependency resolution
- ⚠️ **Installation Failures**: Prerequisite installation failures block main package
- ⚠️ **Performance Impact**: Additional checks and installations increase time
- ⚠️ **Complexity**: More moving parts in installation system

### Risk Mitigation:
- **Dependency Validation**: Detect circular dependencies at configuration load time
- **Graceful Degradation**: Clear error messages for prerequisite failures
- **Caching**: Verify prerequisites only once per session
- **Comprehensive Testing**: Integration tests for dependency chains

## Implementation Details

### Configuration Example
```json
{
  "nodejs": {
    "name": "Node.js JavaScript Runtime",
    "description": "JavaScript runtime built on Chrome's V8 engine with npm package manager",
    "platforms": { /* ... */ }
  },
  "claude-code": {
    "name": "Claude Code", 
    "description": "Anthropic's official CLI for Claude AI assistant",
    "prerequisites": ["nodejs"],  // <-- Declares dependency
    "platforms": { /* ... */ }
  }
}
```

### Error Handling Strategy
```go
// Prerequisite installation failure
if err := InstallPackage(prereq, ""); err != nil {
    return fmt.Errorf("failed to install prerequisite '%s': %w", prereq, err)
}
```

### Verification Integration
Reuses existing package verification infrastructure:
- Each package defines verification commands in configuration
- `isPackageInstalled()` leverages these commands
- No duplicate verification logic needed

## Validation & Testing

### Success Criteria:
- [x] Prerequisites automatically installed before main package
- [x] Clear progress messages during prerequisite installation
- [x] Dry-run mode shows prerequisite installation plan
- [x] Already-installed prerequisites are detected and skipped
- [x] Installation fails gracefully if prerequisites cannot be satisfied

### Test Scenarios:
```bash
# Test prerequisite detection
portunix install claude-code --dry-run  # Shows nodejs prerequisite

# Test prerequisite installation (in clean environment)
portunix install claude-code  # Should install nodejs first

# Test prerequisite skip (when already installed)  
portunix install claude-code  # Should detect existing nodejs
```

### Integration Points:
- **E2E Tests**: Container-based testing without pre-installed dependencies
- **Unit Tests**: Mock prerequisite installation and verification
- **Cross-Platform Tests**: Verify behavior on Windows, Linux, macOS

## Future Enhancements

### Planned Features:
1. **Version Constraints**: Specify minimum prerequisite versions
2. **Optional Dependencies**: Mark some prerequisites as optional
3. **Conflict Detection**: Handle packages that conflict with each other  
4. **Dependency Graphs**: Visualize package dependency relationships
5. **Batch Installation**: Optimize multiple package installations

### Configuration Schema Evolution:
```json
{
  "prerequisites": [
    {
      "package": "nodejs",
      "version": ">=18.0.0",
      "optional": false
    }
  ]
}
```

## Related ADRs
- **ADR-006**: Dynamic Package List Generation (companion feature)

---
**Author**: Claude Code Assistant  
**Date**: 2025-09-12  
**Issue**: #041 Node.js/npm Installation Support  
**Implementation**: `app/install/installer.go`, `app/install/config.go`