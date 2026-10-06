# Issue #135: PFT Assign/Show Commands Ignore --path Parameter

## Status
✅ Implemented

## Priority
High

## Type
Bug Fix

## Labels
bug, helper-binary, ptx-pft, cli, path-parameter, regression

## Description

The `portunix pft assign` and `portunix pft show` commands do not properly handle the `--path` parameter, while other PFT commands like `list` and `category list` work correctly with the same parameter.

## Environment

- **Project**: External project with PFT configuration
- **Config location**: `{project}/.pft-config.json`
- **Portunix version**: Current

## Problem Details

### Working Commands

```bash
# These work correctly with --path parameter
portunix pft list --path docs/pft-example-product
portunix pft category list --path docs/pft-example-product
```

### Failing Commands

```bash
# All these variants fail:
portunix pft assign P01 --category A --path docs/pft-example-product
# Error: feedback item 'P01' not found

portunix pft assign P01-sumarizace-výsledků-hledání --category A --path docs/pft-example-product
# Error: feedback item 'P01-sumarizace-výsledků-hledání' not found

portunix pft show P01-sumarizace-výsledků-hledání --path docs/pft-example-product
# Error: No configuration found. Run 'portunix pft configure' first.
```

## Root Cause Analysis

Likely causes:
1. `assign` and `show` commands may not be passing `--path` to the item lookup function
2. The commands may be looking in the default config location instead of the specified path
3. Item ID resolution (both short ID like `P01` and slug-based) may not be using the correct config context

## Expected Behavior

1. `pft assign` should accept `--path` parameter and resolve items from the specified project
2. `pft show` should accept `--path` parameter and display items from the specified project
3. Item lookup should work with both:
   - Short ID from YAML frontmatter (`id: P01`)
   - Full slug filename (`P01-sumarizace-výsledků-hledání`)

## Affected Files (User's Project)

### VoC needs
- `VoC/needs/P01-sumarizace-výsledků-hledání.md`
- `VoC/needs/P02-extrakce-entit-a-profilů.md`
- `VoC/needs/P03-hledání-stejných-entit.md`
- `VoC/needs/P04-křížové-vyhledávání-entit.md`
- `VoC/needs/P05-časová-osa-událostí.md`

### VoS needs
- `VoS/needs/P28-uživatelské-role-a-profily-pro-llm.md`
- `VoS/needs/P29-charakteristika-projektu-pro-llm.md`
- `VoS/needs/P30-charakteristika-zdrojů-per-oddělení.md`
- `VoS/needs/P31-dotaz-bez-kontextu.md`
- `VoS/needs/P32-ladění-promptů-s-uživateli.md`
- `VoS/needs/P33-...` (truncated in report)

## Additional Issues

### Issue A: Category counting from YAML frontmatter
User also reported that `portunix pft category list` does not count items that have category assigned in YAML frontmatter (e.g., `category: A`). The ITEMS column shows 0 even when items have categories assigned.

**Status**: ✅ Fixed - `ParseMarkdownFile` now parses YAML frontmatter `category:` field.

### Issue B: Item lookup in subdirectories
Item lookup functions (`findFeedbackItem`, `findFeedbackItemFile`) were not searching recursively in subdirectories like `needs/`, `verbatims/`. Files in `VoC/needs/P01-xxx.md` were not found.

**Status**: ✅ Fixed - Functions now use recursive `ScanFeedbackDirectory`.

### Issue C: Category assignment writes to wrong location
`pft assign` command writes categories to `## Categories` markdown section instead of YAML frontmatter. This causes duplicate category definitions:
- YAML frontmatter: `category: E` (original)
- Markdown section: `## Categories` with `E, F` (after assign)

**Status**: ✅ Fixed - Categories now written to YAML frontmatter as `categories:` array. Old `category:` field and `## Categories` markdown section are automatically removed.

### Issue D: Category replacement mode
Added `--set` flag to `pft assign` command to replace all existing categories instead of adding.

**Status**: ✅ Fixed - Use `pft assign P01 --category A --set` to replace all categories with A.

## Acceptance Criteria

- [x] `pft assign` command respects `--path` parameter
- [x] `pft show` command respects `--path` parameter
- [x] `pft unassign` command respects `--path` parameter
- [x] Item lookup works with short ID (e.g., `P01`)
- [x] Item lookup works with full slug (e.g., `P01-sumarizace-výsledků-hledání`)
- [x] Item lookup works in subdirectories (e.g., `VoC/needs/`)
- [x] `pft category list` correctly counts items with category in frontmatter
- [x] All commands behave consistently with `--path` parameter
- [x] `pft assign` writes categories to YAML frontmatter (Issue C)
- [x] `pft assign --set` replaces all categories with new one (Issue D)

## Implementation Notes

### Modified Files

```
src/helpers/ptx-pft/main.go   # handleShowCommand, handleAssignCommand, handleUnassignCommand,
                              # findFeedbackItem, findItemInDirectory, findFeedbackItemFile,
                              # showShowHelp, showAssignHelp, showUnassignHelp
src/helpers/ptx-pft/sync.go   # ParseMarkdownFile (YAML frontmatter parsing),
                              # UpdateFileCategories (write to YAML frontmatter),
                              # SetCategoryToFile (new function)
```

### Fix Summary

1. Added `--path` parameter parsing to `handleShowCommand`, `handleAssignCommand`, `handleUnassignCommand`
2. Changed config loading from `LoadConfigWithFilePath()` / `getProjectDir()` to `loadOrCreateConfig(configPath)`
3. Made `findFeedbackItem` and `findFeedbackItemFile` recursive using `ScanFeedbackDirectory`
4. Added `getVoiceDir()` usage for proper case handling (VoC vs voc)
5. Extended `ParseMarkdownFile` to parse YAML frontmatter `category:` field
6. Rewrote `UpdateFileCategories` to write to YAML frontmatter as `categories:` array
7. Added `SetCategoryToFile` function for `--set` mode
8. Added `--set` flag to `handleAssignCommand` for replacing all categories
9. Updated help texts with `--path` and `--set` documentation

## Workaround

~~Currently no known workaround. Users must navigate to the project directory and run commands without `--path`.~~

**Fixed** - All commands now support `--path` parameter.

## Related Issues

- #107 - PTX-PFT Product Feedback Tool Helper Implementation
- #112 - PTX-PFT Category Management for UC and Requirements
- #134 - PFT Config Cross-Platform Path Support

---

**Reported**: 2026-01-15
**Reporter**: External Portunix user
**Assigned**: Unassigned
