# Issue #53: Fix Module Path Naming Inconsistencies

**Issue ID:** #053
**Type:** Bug Fix / Refactoring
**Priority:** High
**Status:** 📋 Open
**Created:** 2025-09-22
**Labels:** refactoring, module-management, consistency, architecture

## Problem Description

During architecture review, architect discovered incorrect module path naming in the codebase that violates the established module naming conventions. Two modules have incorrect paths that need to be corrected for consistency and proper module organization.

### Current Incorrect Paths
1. **Virt Module**: `portunix.ai/portunix/app/virt` should be `portunix.ai/app/virt`
2. **Wizard Module**: `github.com/cassandragargoyle/portunix/app/wizard` should be `portunix.ai/app/wizard`

### Impact Analysis
- **Consistency**: Breaks established module naming conventions
- **Architecture**: Creates confusion in module hierarchy
- **Maintainability**: Makes codebase harder to navigate and understand
- **Future Development**: Incorrect patterns may be copied to new modules

## Root Cause

The inconsistencies appear to stem from:
1. **Legacy naming**: Some modules still use old `github.com/cassandragargoyle/portunix` prefix
2. **Redundant paths**: The `portunix.ai/portunix/app/*` pattern includes unnecessary `portunix` duplication
3. **Migration oversight**: Incomplete module path standardization during previous refactoring

## Expected Behavior

All app modules should follow the consistent pattern: `portunix.ai/app/{module-name}`

### Corrected Paths
1. `portunix.ai/app/virt` (remove redundant `/portunix`)
2. `portunix.ai/app/wizard` (update from legacy GitHub path and remove redundant `/portunix`)

## Affected Files

Based on initial analysis, the following files need updates:

### Virt Module (portunix.ai/portunix/app/virt → portunix.ai/app/virt)
- `go.mod` (main module file)
- `src/cmd/virt_*.go` files (11 files)
- `src/app/virt/go.mod`
- `src/app/virt/**/*.go` files
- `test/unit/virt_manager_test.go`

### Wizard Module (github.com/cassandragargoyle/portunix/app/wizard → portunix.ai/app/wizard)
- `go.mod` (main module file)
- `src/cmd/wizard.go`
- `src/app/wizard/go.mod`
- `src/app/wizard/**/*.go` files
- Test files in wizard module
- `.golangci.yml` (linter configuration)

## Acceptance Criteria

### ✅ Module Path Corrections
- [ ] Update all import statements for virt module to use `portunix.ai/app/virt`
- [ ] Update all import statements for wizard module to use `portunix.ai/app/wizard`
- [ ] Update go.mod files to reflect correct module paths
- [ ] Ensure no references to old paths remain in codebase

### ✅ Build & Test Verification
- [ ] All Go code compiles successfully with new paths
- [ ] All existing tests pass with updated imports
- [ ] No broken import references in any files
- [ ] Go mod tidy runs cleanly without errors

### ✅ Consistency Check
- [ ] All app modules follow the pattern `portunix.ai/app/{name}`
- [ ] No legacy `github.com/cassandragargoyle/portunix` references remain
- [ ] No redundant `portunix.ai/portunix/app` patterns exist
- [ ] Module hierarchy is consistent across the codebase

### ✅ Documentation Updates
- [ ] Update any documentation that references the old module paths
- [ ] Ensure architecture documentation reflects correct paths
- [ ] Update developer guides if they mention module structure

## Implementation Steps

1. **Audit Phase**
   - [ ] Complete scan of all files containing incorrect module paths
   - [ ] Create comprehensive list of files requiring updates
   - [ ] Identify any additional modules with similar issues

2. **Update Phase - Virt Module**
   - [ ] Update `src/app/virt/go.mod` module declaration
   - [ ] Update all import statements in virt-related command files
   - [ ] Update imports in virt module internal files
   - [ ] Update main go.mod dependencies

3. **Update Phase - Wizard Module**
   - [ ] Update `src/app/wizard/go.mod` module declaration
   - [ ] Update all import statements in wizard-related files
   - [ ] Update imports in wizard module internal files
   - [ ] Update main go.mod dependencies

4. **Verification Phase**
   - [ ] Run `go mod tidy` on all affected modules
   - [ ] Compile entire project to verify no broken imports
   - [ ] Run all tests to ensure functionality is preserved
   - [ ] Perform final scan for any missed references

5. **Documentation Phase**
   - [ ] Update architecture documentation
   - [ ] Update any developer guides mentioning module paths
   - [ ] Add note about module naming conventions

## Testing Strategy

### Unit Tests
- Run existing unit tests for both virt and wizard modules
- Verify all tests pass with updated import paths
- Add regression test to prevent future naming inconsistencies

### Integration Tests
- Test module interactions with corrected paths
- Verify CLI commands using these modules still function correctly
- Test cross-module dependencies work properly

### Build Tests
- Clean build from scratch with new paths
- Verify go mod download works correctly
- Test build on multiple platforms (Windows/Linux)

## Risk Assessment

### Low Risk
- **Pure refactoring**: No functional logic changes
- **Automated verification**: Build system will catch import errors
- **Reversible**: Changes can be rolled back if issues arise

### Mitigation Strategies
- Perform changes in feature branch
- Run comprehensive test suite before merge
- Have rollback plan ready if critical issues discovered
- Document all changes for future reference

## Definition of Done

- [ ] All module paths follow consistent `portunix.ai/app/{name}` pattern
- [ ] No legacy or incorrect module path references remain
- [ ] All code compiles and tests pass
- [ ] Documentation updated to reflect correct paths
- [ ] Changes reviewed and approved by architect
- [ ] Merge to main branch completed successfully

## Notes

This refactoring is critical for maintaining architectural consistency and preventing future confusion. While the change scope is significant, the risk is low since this is purely a naming/import update without functional changes.

The architect has identified this as a priority issue that should be addressed before further module development continues.

---

**Estimated Effort:** 4-6 hours
**Dependencies:** None
**Blocks:** Future module development
**Related Issues:** #040 (Previous module migration)