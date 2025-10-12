# ADR-017: PTX-Prompting Helper for Template-Based Prompt Generation

## Status
Active

## Context
The development team frequently needs to generate prompts for AI assistants (Claude, ChatGPT) with consistent structure and placeholders. Currently, this is done manually by copying and modifying existing prompts, which leads to:
- Inconsistent prompt formatting
- Time wasted on repetitive prompt creation
- Difficulty in maintaining prompt templates
- No standardized way to share common prompt patterns across the team

The team identified the need for a dedicated tool to streamline prompt generation based on templates with placeholders.

## Decision
We will create a new Portunix helper called `ptx-prompting` that will:

1. **Load prompt templates** from files (.md, .yaml, .txt) containing placeholders in `{placeholder}` format
2. **Automatically detect placeholders** and create a parameter list
3. **Support two input modes**:
   - Command-line arguments for automation
   - Interactive mode for missing parameters
4. **Provide flexible output options**:
   - Standard output for piping
   - Direct clipboard integration for immediate use
   - File output for persistence

### Architecture
The helper will follow the standard Portunix helper pattern:
- Binary name: `ptx-prompting` (with `.exe` extension on Windows)
- Location: Same directory as main Portunix binary
- Discovery: Automatic via `HelperDiscovery` system
- Integration: Responds to standard helper commands (`--version`, `--list-commands`, `--description`)

### Command Structure
```bash
# Basic usage
ptx-prompting build prompts/translate.md

# With arguments
ptx-prompting build prompts/translate.md --source_file README.cs.md --target_file README.en.md --target_language English

# With clipboard output
ptx-prompting build prompts/translate.md --copy

# List available templates
ptx-prompting list

# Create new template
ptx-prompting create translation-prompt.md
```

### Template Format
Templates will use simple placeholder syntax:
```markdown
# Translation Request

Please translate the file {source_file} from Czech to {target_language}.
Save the result as {target_file}.

Requirements:
- Preserve all formatting
- Keep technical terms consistent
- Target audience: {audience}
```

### Directory Structure
```
ptx-prompting/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # Root command
│   ├── build.go           # Build command
│   ├── list.go            # List templates
│   └── create.go          # Create template
├── internal/
│   ├── parser/
│   │   └── parser.go      # Template parsing
│   ├── prompt/
│   │   └── builder.go     # Prompt building
│   └── clipboard/
│       └── clipboard.go   # Clipboard integration
├── templates/             # Default templates
│   ├── en/
│   │   ├── translate.md
│   │   └── review.md
│   └── cs/
│       └── translate.md
└── go.mod
```

## Consequences

### Positive
- **Consistency**: All prompts follow standardized templates
- **Efficiency**: Rapid prompt generation saves time
- **Reusability**: Templates can be shared and versioned
- **Integration**: Works seamlessly with Portunix ecosystem
- **Multilingual support**: Templates can be organized by language
- **Automation-friendly**: Can be integrated into scripts and workflows
- **Extensible**: Easy to add new templates and placeholder types

### Negative
- **Additional binary**: Increases distribution size
- **Dependency management**: Requires clipboard library dependency
- **Template maintenance**: Templates need to be kept up-to-date
- **Learning curve**: Users need to learn template syntax

### Implementation Notes
1. Use `github.com/spf13/cobra` for CLI structure (consistent with Portunix)
2. Use `github.com/atotto/clipboard` for clipboard operations
3. Implement regex-based placeholder detection: `\{([^}]+)\}`
4. Support both positional and named arguments
5. Provide helpful error messages for missing placeholders
6. Include example templates in the binary or as separate files

## References
- Issue discussion with ChatGPT about prompt template system
- Portunix helper system documentation (src/shared/helper.go)
- Standard Portunix helper pattern (ptx-container, ptx-mcp, ptx-ansible)

## Author
Architect

## Date
2025-09-26