# Issue #40: Migrate Go Module Name from portunix.cz to portunix.ai

## Summary
Update all Go module references and imports from `portunix.cz` to `portunix.ai` to align with the project's actual domain and branding strategy.

## Problem Statement
The current Go project uses `portunix.cz` as the module name and in all import statements, but the official project domain is `portunix.ai`. This creates inconsistency between:

1. **Module declaration**: `module portunix.cz/portunix` in go.mod
2. **Import paths**: All internal imports use `portunix.cz/app/...`, `portunix.cz/cmd/...`, etc.
3. **Actual branding**: Project website and documentation reference `portunix.ai`
4. **User expectations**: Users expect imports to match the official domain

## Current State Analysis
```bash
# Find all portunix.cz references
grep -r "portunix.cz" . --include="*.go" --include="*.mod"
```

### Files Affected
- `go.mod` - Module declaration
- All `.go` files with internal imports
- Test files
- Documentation examples

### Import Examples Currently Used
```go
import (
    "portunix.cz/app/container"
    "portunix.cz/app/docker" 
    "portunix.cz/app/system"
    "portunix.cz/cmd"
)
```

## Proposed Solution

### 1. Update go.mod
```go
// Before
module portunix.cz/portunix

// After  
module portunix.ai/portunix
```

### 2. Global Find and Replace
Update all import statements across the codebase:
```bash
find . -name "*.go" -exec sed -i 's/portunix\.cz/portunix.ai/g' {} \;
```

### 3. Verification Steps
- Ensure all imports compile correctly
- Run full test suite
- Check for any hardcoded references
- Update any documentation examples

## Implementation Plan

### Phase 1: Module Declaration
1. Update `go.mod` module declaration
2. Run `go mod tidy` to update dependencies

### Phase 2: Import Updates
1. Perform global find/replace on all `.go` files
2. Update any string literals containing old domain
3. Check configuration files and scripts

### Phase 3: Testing & Verification
1. Build project: `go build -o .`
2. Run all tests: `go test ./...`
3. Run E2E tests if applicable
4. Test plugin system compatibility

### Phase 4: Documentation
1. Update any code examples in documentation
2. Check README files for import examples
3. Update installation scripts if needed

## Risks & Considerations

### Breaking Changes
- **External dependencies**: Any external code importing this module will break
- **Plugin compatibility**: Existing plugins may need updates
- **Build systems**: CI/CD pipelines may need adjustment

### Mitigation Strategies
- This is primarily an internal refactor since the project is not yet widely distributed
- Document the change clearly in release notes
- Consider maintaining compatibility imports temporarily if needed

## Technical Impact

### Positive Impacts
- ✅ Consistent branding across all project materials
- ✅ Correct domain alignment with official website
- ✅ Professional appearance for external contributors
- ✅ Better SEO and discoverability

### Minimal Negative Impact
- 🔄 One-time refactoring effort
- 🔄 Need to update any existing documentation

## Alternative Solutions Considered

1. **Keep portunix.cz**: Would maintain status quo but creates ongoing confusion
2. **Use github.com path**: Could use `github.com/cassandragargoyle/portunix` but doesn't align with custom domain
3. **Gradual migration**: Too complex for internal project at current stage

## Acceptance Criteria
- [ ] go.mod updated to use portunix.ai module name
- [ ] All internal imports updated from portunix.cz to portunix.ai  
- [ ] Project compiles successfully with new module name
- [ ] All existing tests pass
- [ ] No references to old domain remain in codebase
- [ ] Documentation examples updated
- [ ] Release notes document the change

## Testing Strategy
```bash
# Compile check
go build -o portunix

# Run tests
go test ./...

# Check for remaining old references
grep -r "portunix\.cz" . --include="*.go" --include="*.mod" --include="*.md"

# Verify plugin system still works
./portunix plugin list
```

## Priority
**Medium** - Important for branding consistency but not blocking current functionality

## Type
**Refactoring** - Internal code organization improvement

## Labels
- refactoring
- branding  
- module-management
- breaking-change
- internal

## Estimated Effort
**2-4 hours** - Mostly automated find/replace with verification testing

## Dependencies
- None - standalone refactoring task

## References
- Official domain: https://portunix.ai
- Current module structure in go.mod
- Import statements throughout codebase

---

**Created:** 2025-01-12  
**Status:** 📋 Open  
**Assignee:** TBD  
**Labels:** refactoring, branding, module-management
