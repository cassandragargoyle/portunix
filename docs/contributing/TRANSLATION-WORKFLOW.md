# Translation Workflow

## Purpose
This document defines the standardized workflow for translating project documentation into team members' native languages using Claude Code.

## Translation Structure

### Directory Layout
All translations are stored in the `.translated/` directory with the following structure:

```
.translated/
├── [language-code]/
│   └── [original-path]/
│       └── [filename].md
```

### Examples
```
.translated/
├── cs/                           # Czech
│   └── docs/
│       └── contributing/
│           ├── README.md
│           ├── AI-ASSISTANTS.md
│           └── TODO-GUIDELINES.md
├── de/                           # German
│   └── docs/
│       └── contributing/
│           └── README.md
├── fr/                           # French
├── es/                           # Spanish
└── pl/                           # Polish
```

## Language Codes

Use ISO 639-1 two-letter language codes:

| Language | Code | Example Path |
|----------|------|-------------|
| Czech | `cs` | `.translated/cs/docs/contributing/README.md` |
| German | `de` | `.translated/de/docs/contributing/README.md` |
| French | `fr` | `.translated/fr/docs/contributing/README.md` |
| Spanish | `es` | `.translated/es/docs/contributing/README.md` |
| Polish | `pl` | `.translated/pl/docs/contributing/README.md` |
| Italian | `it` | `.translated/it/docs/contributing/README.md` |
| Portuguese | `pt` | `.translated/pt/docs/contributing/README.md` |

## Claude Code Instructions

### Translation Commands
When you receive a translation request, follow this workflow:

#### Command Patterns
- `"Translate docs/contributing/README.md to Czech"`
- `"Translate to German: docs/contributing/AI-ASSISTANTS.md"`
- `"Translate docs/contributing/ to Czech"` (directory translation)
- `"Přelož mi docs/contributing/README.md do češtiny"`

#### Translation Process
1. **Identify source path**: Extract the file/directory path from the request
2. **Identify target language**: Determine language code from request
3. **Check for directory translation**: If path ends with `/`, ask user for scope:
   - "Do you want to translate all files (including existing translations) or only new/missing translations?"
4. **Create directory structure**: Ensure `.translated/[language-code]/[path]/` exists
5. **Translate content**: Translate while preserving:
   - Markdown formatting
   - Code blocks and examples
   - Links structure
   - File references
6. **Save translated file**: Use same filename in translated directory

#### Example Workflow
Request: `"Translate docs/contributing/README.md to Czech"`

1. Source: `docs/contributing/README.md`
2. Target language: Czech (`cs`)
3. Create: `.translated/cs/docs/contributing/`
4. Translate content preserving formatting
5. Save as: `.translated/cs/docs/contributing/README.md`

### Translation Guidelines

#### What to Translate
- ✅ All text content
- ✅ Headings and titles
- ✅ Descriptions and explanations
- ✅ Examples and use cases
- ✅ Error messages and warnings

#### What NOT to Translate
- ❌ Code examples and syntax
- ❌ File paths and URLs
- ❌ Command line examples
- ❌ Environment variables
- ❌ Technical terms (API, CLI, etc.)
- ❌ Proper names (GitHub, Claude Code, etc.)

#### Preserve Structure
- Keep all markdown formatting intact
- Maintain heading hierarchy
- Preserve code block languages
- Keep link destinations unchanged
- Maintain table structure

#### Quality Standards
- Use natural, professional language
- Maintain technical accuracy
- Be consistent with terminology
- Adapt cultural references appropriately
- Keep the same tone and style

## File Management

### Git Integration
The `.translated/` directory should be ignored by Git:

```gitignore
# Translation files (temporary, generated on demand)
.translated/
```

### Refresh Strategy
- Translations are temporary and can be regenerated
- Delete `.translated/` folder when source documents change significantly
- Regenerate translations as needed by team members

### AI Context Exclusion
- Claude Code should NOT read from `.translated/` directories for project context
- Only use original English documentation for understanding project structure
- `.translated/` is output-only directory for translation tasks

## Usage Examples

### Individual File Translation
```
User: "Translate docs/contributing/AI-ASSISTANTS.md to Czech"
Claude Code: 
1. Reads docs/contributing/AI-ASSISTANTS.md
2. Creates .translated/cs/docs/contributing/ if needed
3. Translates content to Czech
4. Saves as .translated/cs/docs/contributing/AI-ASSISTANTS.md
```

### Directory Translation
```
User: "Translate docs/contributing/ to Czech"
Claude Code:
1. Asks: "Do you want to translate all files (including existing translations) or only new/missing translations?"
2. Based on user choice:
   - All files: Translates every .md file in directory and subdirectories
   - New/missing: Only translates files not yet in .translated/cs/docs/contributing/
3. Creates complete directory structure in .translated/cs/
4. Maintains relative paths and subdirectories
```

### Batch Translation
```
User: "Create Czech translations for the entire docs/ directory"
Claude Code:
1. Recursively finds all .md files in docs/
2. Maintains directory structure in .translated/cs/
3. Translates each file individually
```

## Quality Assurance

### Post-Translation Checks
- Verify all links still work (relative paths)
- Ensure code examples remain functional
- Check that technical terms are consistent
- Validate markdown rendering

### Team Feedback
- Team members can provide feedback on translation quality
- Update this workflow based on common translation issues
- Maintain glossary of preferred technical term translations

## Maintenance

### Regular Updates
- Review translation workflow quarterly
- Update language codes list as team grows
- Refresh translation guidelines based on usage patterns
- Archive unused language translations

### Documentation Updates
When source documentation changes:
1. Update English original first
2. Delete affected translations from `.translated/`
3. Regenerate translations on demand
4. No need to maintain translation versioning

---

**Note**: This translation system is designed for internal team communication only. All official project documentation remains in English.

*Created: 2025-08-23*
*Last updated: 2025-08-23*