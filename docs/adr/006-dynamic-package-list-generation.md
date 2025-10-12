# ADR-006: Dynamic Package List Generation for Install Command Help

## Status
**Accepted** - Implemented as part of Issue #041

## Context
The `portunix install --help` command previously used a hardcoded list of available packages and their variants in `cmd/install.go`. This approach had several drawbacks:

### Problems with Hardcoded Package Lists:
1. **Maintenance Overhead**: Every new package addition required updating both `assets/install-packages.json` AND `cmd/install.go`
2. **Inconsistency Risk**: Package definitions and help text could become out of sync
3. **Scalability Issues**: As the package ecosystem grows, maintaining hardcoded lists becomes error-prone
4. **Developer Experience**: Contributors had to remember to update multiple files for package additions

### Business Impact:
- **User Confusion**: Help text showing unavailable packages or missing new packages
- **Development Friction**: Two-step process for package additions slowed development
- **Documentation Drift**: Risk of help text not reflecting actual available packages

## Decision

We replace the hardcoded package list generation with a **dynamic system** that reads from the embedded `assets/install-packages.json` configuration.

### Architecture Changes:

#### 1. New Dynamic Help Generation Functions
Add to `app/install/config.go`:
- `GeneratePackageListDescription()` - Creates formatted package list
- `GeneratePresetListDescription()` - Creates AI and other preset lists  
- `GenerateVariantListDescription()` - Creates variant information

#### 2. Modified Install Command Structure
In `cmd/install.go`:
- `generateInstallHelp()` function dynamically builds help content
- `init()` function calls `generateInstallHelp()` to set `Long` field
- Fallback error handling if configuration loading fails

#### 3. Configuration-Driven Content
The help system now automatically:
- **Discovers packages** from JSON configuration
- **Categorizes presets** (AI Assistant vs. Others)
- **Lists variants** for each platform
- **Maintains sorting** for consistent output

### Technical Implementation:

```go
// Dynamic help generation
func generateInstallHelp() string {
    config, err := install.LoadInstallConfig()
    if err != nil {
        return fallbackHelpText + err.Error()
    }
    
    var helpText strings.Builder
    helpText.WriteString(config.GeneratePackageListDescription())
    helpText.WriteString(config.GeneratePresetListDescription())
    helpText.WriteString(config.GenerateVariantListDescription())
    return helpText.String()
}

// Initialization in install command
func init() {
    installCmd.Long = generateInstallHelp() + staticExamplesText
    rootCmd.AddCommand(installCmd)
}
```

## Alternatives Considered

### 1. Keep Hardcoded Lists
**Pros**: Simple, no runtime overhead
**Cons**: High maintenance, consistency issues, doesn't scale

### 2. External Configuration File
**Pros**: Easy to edit without rebuilding
**Cons**: Deployment complexity, runtime file dependencies

### 3. Code Generation at Build Time
**Pros**: No runtime overhead, type safety
**Cons**: Complex build pipeline, limited flexibility

## Consequences

### Positive:
- ✅ **Single Source of Truth**: Package definitions only in `install-packages.json`
- ✅ **Automatic Consistency**: Help always matches available packages
- ✅ **Reduced Development Overhead**: Add package once, help updates automatically
- ✅ **Better Maintainability**: No manual synchronization required
- ✅ **Scalability**: System handles unlimited packages without code changes

### Potential Risks:
- ⚠️ **Runtime Dependency**: Help generation depends on embedded configuration
- ⚠️ **Error Handling**: Configuration loading errors affect help display
- ⚠️ **Performance**: Minor runtime overhead for help generation

### Mitigation Strategies:
- **Robust Error Handling**: Fallback help text if configuration fails
- **Embedded Assets**: Configuration bundled in binary via `go:embed`
- **Caching**: Help text generated once during initialization

## Implementation Notes

### Package Discovery Algorithm:
1. Load embedded `install-packages.json` configuration
2. Extract package names and descriptions
3. Sort alphabetically for consistent output
4. Format with consistent spacing and alignment

### Preset Categorization:
- **AI Presets**: Contains "claude", "gemini", "ai-", or "mcp" keywords
- **Other Presets**: All remaining presets

### Variant Detection:
- Only shows variants for packages with multiple options
- Indicates default variant in parentheses
- Platform-aware variant filtering

## Validation

### Success Criteria:
- [x] `portunix install --help` shows all packages from configuration
- [x] New packages automatically appear in help without code changes  
- [x] Preset categorization works correctly
- [x] Variant information displays properly
- [x] Error handling works when configuration loading fails

### Test Cases:
```bash
# Test dynamic package discovery
portunix install --help | grep nodejs  # Should show new Node.js package

# Test fallback behavior
# (Configuration loading failure simulation)
```

## Related ADRs
- **ADR-007**: Prerequisite Package Handling System (companion architecture)

---
**Author**: Claude Code Assistant  
**Date**: 2025-09-12  
**Issue**: #041 Node.js/npm Installation Support  
**Implementation**: `cmd/install.go`, `app/install/config.go`