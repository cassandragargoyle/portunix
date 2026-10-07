# Issue #117: PTX-PFT list command incompatible with QFD structure

| Field | Value |
|-------|-------|
| **ID** | #117 |
| **Title** | PTX-PFT list command incompatible with QFD structure |
| **Status** | ✅ Implemented |
| **Closed** | 2026-05-19 |
| **Priority** | Medium |
| **Type** | Bug |
| **Labels** | bug, helper-binary, ptx-pft, qfd |
| **Created** | 2025-12-26 |
| **Related** | #116 |

## Problem Statement

The `portunix pft list` command does not recognize items in projects created with the QFD template (`--template qfd`). The command searches for lowercase directories (`voc/`, `vos/`) but QFD template creates PascalCase directories (`VoC/`, `VoS/`, `VoB/`, `VoE/`).

## Steps to Reproduce

1. Create a QFD project:
   ```bash
   portunix pft project create "Test Project" --template qfd
   ```

2. Add content to VoS/verbatims/ or VoS/needs/

3. Run list command:
   ```bash
   portunix pft list
   ```

4. **Expected:** Items from VoS are listed
5. **Actual:** "Total: 0 items" - no items found

## Root Cause

In `main.go`, the list command hardcodes lowercase paths:

```go
vocDir := filepath.Join(projectDir, "voc")  // Should also check "VoC"
vosDir := filepath.Join(projectDir, "vos")  // Should also check "VoS"
```

## Proposed Solution

### Option A: Case-insensitive directory detection

Check both variants and use whichever exists:

```go
func getVoiceDir(projectDir, voice string) string {
    // Try PascalCase first (QFD template)
    pascalCase := filepath.Join(projectDir, voice)  // "VoC"
    if _, err := os.Stat(pascalCase); err == nil {
        return pascalCase
    }
    // Fall back to lowercase (basic template)
    return filepath.Join(projectDir, strings.ToLower(voice))
}
```

### Option B: Detect template type from config

Store template type in `.pft-config.json` and use appropriate paths.

## Affected Commands

- `pft list`
- `pft show`
- `pft sync`
- `pft pull`
- `pft push`
- `pft report`
- `pft export`
- `pft role init` - creates duplicate lowercase directories instead of using existing PascalCase
- `pft user` commands

## Acceptance Criteria

- [x] `pft list` works with QFD template (PascalCase directories)
- [x] `pft list` works with basic template (lowercase directories)
- [x] All affected commands updated
- [x] Backward compatibility maintained

## Technical Notes

- QFD template uses: `VoC/`, `VoB/`, `VoE/`, `VoS/`, `requirements/`, `matrices/`
- Basic template uses: `voc/`, `vos/`, `vob/`, `voe/`
- QFD also has subdirectories: `verbatims/`, `needs/`, `constraints/`

## Implementation Details

### Recursive Directory Scanning (2025-12-26)

Updated `ScanFeedbackDirectory()` in `sync.go` to recursively scan subdirectories:

```go
// ScanFeedbackDirectory scans a directory for feedback markdown files
// It recursively scans subdirectories (e.g., needs/, verbatims/) for QFD structure compatibility
func ScanFeedbackDirectory(dir string, feedbackType string) ([]*FeedbackItem, error) {
    for _, entry := range entries {
        if entry.IsDir() {
            // Recursively scan subdirectories (QFD structure: needs/, verbatims/, etc.)
            subItems, err := ScanFeedbackDirectory(entryPath, feedbackType)
            items = append(items, subItems...)
            continue
        }
        // ... parse .md files
    }
}
```

This enables `pft list` to find items in QFD subdirectory structure like:
- `VoS/needs/VS-P02-search-results.md`
- `VoC/verbatims/customer-interview-01.md`

## Category ID Normalization

Category IDs support case-insensitive input with uppercase normalization:
- Input: Can be entered in lowercase or uppercase (e.g., `user-auth` or `USER-AUTH`)
- Storage: Always stored in UPPERCASE (e.g., `USER-AUTH`)
- QFD compatibility: Categories A-I for VoC and VoS areas follow ISO 16355 naming

## Additional Requirement: Add Command (2025-12-26)

### Problem

The `pft add` command is missing. Users cannot add new feedback items/requirements via CLI.

### Proposed Solution

Implement `pft add` command for adding new feedback items:

```bash
portunix pft add --area vos \
  --title "Search Results Summarization" \
  --description "AI summarizes hundreds of results into an overview" \
  --category A \
  --author "Petr Vancel" \
  --source "Email Petr Vancel (2025-12-16)"
```

### Command Options

| Option | Required | Description |
|--------|----------|-------------|
| `--area` | Yes | Target area (voc, vos, vob, voe) |
| `--title` | Yes | Item title |
| `--description` | No | Item description |
| `--category` | No | Category ID to assign |
| `--author` | No | Author name |
| `--source` | No | Source of the requirement |
| `--status` | No | Initial status (default: pending) |
| `--path` | No | Path to PFT project |

### File Generation

The command should:
1. Generate unique ID (e.g., P01, P02, ...)
2. Create markdown file in appropriate area directory
3. Use QFD-compatible structure (needs/ subdirectory for requirements)
4. Apply category assignment if specified

### Acceptance Criteria

- [x] `pft add` command implemented
- [x] Generates valid markdown file with frontmatter
- [x] Supports QFD directory structure
- [x] Auto-generates sequential IDs
- [x] Works with `--path` parameter for specifying project location

### Additional Requirement: Automatic Role Assignment (2025-12-27)

When `--author` is specified in `pft add`, the command should automatically populate the author's role from the user registry.

**Behavior:**
1. Look up author name in `users.json` registry
2. If found, add `author_role` field to frontmatter
3. If not found, leave role empty (or warn user)

**Example:**
```bash
portunix pft add --area vos --title "Feature X" --author "Petr Vancel"
```

Should generate:
```yaml
---
author: Petr Vancel
author_role: product-owner
---
```

**Acceptance Criteria:**
- [x] Author role auto-populated from user registry
- [x] Works with both VoC and VoS user registries
- [x] Graceful handling when author not in registry
