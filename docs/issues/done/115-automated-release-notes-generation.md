# Issue #115: Automated Release Notes Generation System

| Field | Value |
|-------|-------|
| **ID** | #115 |
| **Title** | Automated Release Notes Generation System |
| **Status** | ✅ Implemented |
| **Priority** | High |
| **Type** | Enhancement |
| **Labels** | enhancement, release-process, automation, ai-integration, documentation |
| **Created** | 2025-12-25 |

## Problem Statement

Current release process (`make dist`) generates a generic `RELEASE_NOTES_vX.Y.Z.md` file with static content that doesn't reflect actual changes made in each version. Users and maintainers need accurate, version-specific release notes that describe what changed between versions.

## Requirements

1. **JSON-based Release Notes Storage**
   - Each version has a corresponding JSON file with structured release notes
   - JSON files stored in `release-notes/` directory (version without `v` prefix)
   - Schema validation for consistency

2. **AI-Assisted Generation**
   - Claude Code command `/cs:generate-release-notes` analyzes git history
   - Extracts changes from commits between tags
   - Categorizes changes (features, fixes, improvements, breaking changes)
   - Generates human-readable descriptions

3. **Integration with `make dist`**
   - Automatically checks for missing JSON files
   - Collects JSON data and generates Markdown files
   - Places RELEASE-NOTES.md in each distribution directory

4. **Backward Compatibility**
   - Works with existing tag structure
   - Can generate notes for historical versions retroactively

## Proposed Architecture

### Directory Structure

```
portunix/
├── release-notes/
│   ├── _schema.json              # JSON schema for validation
│   ├── 1.5.1.json               # Release notes for v1.5.1
│   ├── 1.6.0.json
│   ├── 1.7.0.json
│   ├── 1.8.0.json
│   └── ...
├── .claude/commands/cs/
│   └── generate-release-notes.md # Claude command for AI generation
├── scripts/release/
│   └── collect_release_notes.py  # Python collector script
└── Makefile                      # Updated dist target
```

### Component Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Release Notes Workflow                        │
└─────────────────────────────────────────────────────────────────┘

    ┌──────────────┐          ┌──────────────────┐
    │  Developer   │          │   make dist      │
    │  runs Claude │          │   command        │
    │  command     │          │                  │
    └──────┬───────┘          └────────┬─────────┘
           │                           │
           ▼                           ▼
    ┌──────────────┐          ┌──────────────────┐
    │ /cs:generate │          │ collect_release_ │
    │ -release-    │          │ notes.py         │
    │ notes        │          │ --check          │
    └──────┬───────┘          └────────┬─────────┘
           │                           │
           ▼                           ▼
    ┌──────────────┐          ┌──────────────────┐
    │ Analyze git  │          │ Check for        │
    │ commits      │          │ missing JSON     │
    │ between tags │          │ files            │
    └──────┬───────┘          └────────┬─────────┘
           │                           │
           ▼                           ▼
    ┌──────────────┐    YES   ┌──────────────────┐
    │ Generate     │◄─────────│ Missing files?   │
    │ JSON file    │          │                  │
    └──────┬───────┘          └────────┬─────────┘
           │                           │ NO
           ▼                           ▼
    ┌──────────────┐          ┌──────────────────┐
    │ Save to      │          │ Collect all JSON │
    │ release-     │          │ files and        │
    │ notes/X.Y.Z  │          │ generate MD      │
    │ .json        │          │                  │
    └──────────────┘          └────────┬─────────┘
                                       │
                                       ▼
                              ┌──────────────────┐
                              │ Write RELEASE-   │
                              │ NOTES.md to each │
                              │ dist/ directory  │
                              └──────────────────┘
```

### JSON Schema (`release-notes/_schema.json`)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "date", "tag"],
  "properties": {
    "version": {
      "type": "string",
      "description": "Version number without v prefix (e.g., 1.8.0)"
    },
    "date": {
      "type": "string",
      "format": "date",
      "description": "Release date in YYYY-MM-DD format"
    },
    "tag": {
      "type": "string",
      "description": "Git tag with v prefix (e.g., v1.8.0)"
    },
    "previous_tag": {
      "type": "string",
      "description": "Previous git tag for diff range"
    },
    "summary": {
      "type": "string",
      "description": "Brief summary of the release (1-2 sentences)"
    },
    "highlights": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Key highlights of this release"
    },
    "changes": {
      "type": "object",
      "properties": {
        "features": {
          "type": "array",
          "items": { "$ref": "#/definitions/change" }
        },
        "improvements": {
          "type": "array",
          "items": { "$ref": "#/definitions/change" }
        },
        "fixes": {
          "type": "array",
          "items": { "$ref": "#/definitions/change" }
        },
        "breaking": {
          "type": "array",
          "items": { "$ref": "#/definitions/change" }
        },
        "security": {
          "type": "array",
          "items": { "$ref": "#/definitions/change" }
        }
      }
    },
    "components": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Affected components/helpers"
    }
  },
  "definitions": {
    "change": {
      "type": "object",
      "required": ["description"],
      "properties": {
        "description": { "type": "string" },
        "issue": { "type": "string" },
        "commit": { "type": "string" }
      }
    }
  }
}
```

### Claude Command (`.claude/commands/cs/generate-release-notes.md`)

```markdown
Workflow pro generování release notes JSON souboru pro konkrétní verzi.

@../../prompts/assistant-programming.md

## PARAMETRY

Argument `$ARGUMENTS`:
- Pokud je prázdný: vygeneruj pro VŠECHNY chybějící tagy
- Pokud obsahuje verzi (např. `1.8.0` nebo `v1.8.0`): vygeneruj pouze pro tuto verzi

## KROK 1: Zjištění tagů

1. Spusť `git tag -l --sort=-v:refname` pro seznam tagů
2. Zkontroluj existující JSON soubory v `release-notes/`
3. Identifikuj chybějící JSON soubory

Pokud nejsou žádné chybějící tagy, informuj uživatele a SKONČI.

## KROK 2: Pro každý chybějící tag

Pro každý tag (nebo zadaný tag):

### 2.1 Zjisti rozsah commitů

```bash
# Pro první tag (bez předchozího):
git log <tag> --oneline

# Pro další tagy (zjisti předchozí tag):
git tag -l --sort=v:refname | grep -B1 "^<current_tag>$" | head -1
git log <previous_tag>..<current_tag> --oneline
```

### 2.2 Analyzuj commity

1. Přečti commit messages v rozsahu
2. Identifikuj issue ID (formát #XXX z docs/issues)
3. Kategorizuj změny podle konvenčních commitů:
   - `feat` -> features
   - `fix` -> fixes
   - `refactor`, `perf`, `chore` -> improvements
   - `BREAKING` nebo `!:` -> breaking
   - `security` -> security

### 2.3 Vygeneruj JSON

Vytvoř JSON podle schématu v `release-notes/_schema.json`:

```json
{
  "version": "X.Y.Z",
  "date": "YYYY-MM-DD",
  "tag": "vX.Y.Z",
  "previous_tag": "vA.B.C",
  "summary": "Stručné shrnutí hlavních změn",
  "highlights": [
    "Hlavní změna 1",
    "Hlavní změna 2"
  ],
  "changes": {
    "features": [
      {"description": "Popis funkce", "issue": "#042"}
    ],
    "improvements": [...],
    "fixes": [...],
    "breaking": [...],
    "security": [...]
  },
  "components": ["portunix", "ptx-mcp", "ptx-container"]
}
```

### 2.4 Ulož JSON

Ulož do `release-notes/<version>.json` (bez prefixu `v`).

## KROK 3: Validace

Po vygenerování všech JSON:
1. Zkontroluj, že JSON je validní
2. Zobraz shrnutí: kolik JSON souborů bylo vytvořeno

## POZNÁMKY

- Používej anglické popisy změn (projekt je v angličtině)
- Pokud commit message není dostatečně informativní, podívej se na diff
- Issue ID extrahuj z commit message
- Datum zjisti z tagu: `git log -1 --format=%ai <tag>`
- Components zjisti z changed files v commits
```

### Python Collector Script (`scripts/release/collect_release_notes.py`)

Key features:
- `--check` mode: Verify all tags have JSON files, warn if missing
- `--warn-only`: With --check, only warn (don't fail)
- `--list-missing`: List versions without JSON files
- `--version X.Y.Z`: Generate for specific version only
- Default mode: Generate RELEASE-NOTES.md for all dist/ directories

### Makefile Integration

Update `make dist` target to:

```makefile
dist: ## Create distribution release
	@echo "📋 Checking release notes..."
	@python3 scripts/release/collect_release_notes.py --check --warn-only || true
	@echo "📦 Creating distribution release..."
	# ... existing dist logic ...
	@echo "📝 Generating release notes for distribution..."
	@python3 scripts/release/collect_release_notes.py
```

## Implementation Phases

### Phase 1: Foundation
- [ ] Create `release-notes/` directory structure
- [ ] Create `_schema.json` with JSON schema
- [ ] Create Python collector script with basic functionality

### Phase 2: AI Integration
- [ ] Create `.claude/commands/cs/generate-release-notes.md`
- [ ] Test with a few sample versions

### Phase 3: Makefile Integration
- [ ] Update `make dist` to call collector
- [ ] Update `scripts/make-release.sh` to use collector output

### Phase 4: Historical Notes
- [ ] Generate JSON files for existing tags (v1.5.1 to v1.8.0)
- [ ] Validate all generated files

## Example Output

### JSON File (`release-notes/1.8.0.json`)

```json
{
  "version": "1.8.0",
  "date": "2025-12-20",
  "tag": "v1.8.0",
  "previous_tag": "v1.7.6",
  "summary": "Major release with PTX-PFT helper, MCP improvements, and container enhancements",
  "highlights": [
    "New PTX-PFT Product Feedback Tool helper",
    "MCP configure improvements with mode flags",
    "Enhanced container management commands"
  ],
  "changes": {
    "features": [
      {"description": "PTX-PFT Product Feedback Tool helper implementation", "issue": "#107"},
      {"description": "PTX-PFT email notifications for user actions", "issue": "#108"},
      {"description": "PTX-PFT ClearFlask provider", "issue": "#109"},
      {"description": "PTX-PFT Eververse provider", "issue": "#110"}
    ],
    "improvements": [
      {"description": "MCP configure default stdio mode", "issue": "#114"},
      {"description": "Container list command implementation", "issue": "#084"}
    ],
    "fixes": [
      {"description": "MCP help missing subcommands in v1.8.0", "issue": "#113"},
      {"description": "Container exec returns helper version fix", "issue": "#095"}
    ]
  },
  "components": ["portunix", "ptx-mcp", "ptx-pft", "ptx-container"]
}
```

### Generated RELEASE-NOTES.md

```markdown
# Portunix Release Notes

Generated: 2025-12-25 10:30:00

---

# Release Notes - 1.8.0

**Release Date:** 2025-12-20

Major release with PTX-PFT helper, MCP improvements, and container enhancements

## Highlights

- New PTX-PFT Product Feedback Tool helper
- MCP configure improvements with mode flags
- Enhanced container management commands

## New Features

- PTX-PFT Product Feedback Tool helper implementation (#107)
- PTX-PFT email notifications for user actions (#108)
- PTX-PFT ClearFlask provider (#109)
- PTX-PFT Eververse provider (#110)

## Improvements

- MCP configure default stdio mode (#114)
- Container list command implementation (#084)

## Bug Fixes

- MCP help missing subcommands in v1.8.0 (#113)
- Container exec returns helper version fix (#095)

## Affected Components

- portunix, ptx-mcp, ptx-pft, ptx-container

---

# Release Notes - 1.7.6
...
```

## Benefits

1. **Accurate Documentation**: Each release has precise change documentation
2. **AI-Assisted**: Claude analyzes commits and generates meaningful descriptions
3. **Structured Data**: JSON format enables tooling and automation
4. **Consistent Format**: Schema ensures all release notes follow same structure
5. **Retroactive Generation**: Can generate notes for historical releases
6. **Integration**: Seamless integration with existing release workflow

## References

- Inspired by: release notes system of an external project
- Related: `scripts/make-release.sh` (current release script)
- Related: `.claude/commands/cs/release.md` (current release workflow)

## Acceptance Criteria

- [ ] JSON schema validates all generated files
- [ ] Claude command generates accurate release notes from git history
- [ ] `make dist` warns about missing JSON files
- [ ] RELEASE-NOTES.md generated in each dist/ directory
- [ ] At least 3 historical versions have JSON files for testing
- [ ] Integration with existing release workflow (no breaking changes)
