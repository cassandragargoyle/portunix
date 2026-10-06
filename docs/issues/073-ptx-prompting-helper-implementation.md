# Issue #73: PTX-Prompting Helper Implementation

**Priority**: High
**Type**: Feature
**Labels**: enhancement, helper-system, ai-integration, template-system
**Assignee**: Developer
**Reporter**: Architect
**Created**: 2025-09-26

## Overview
Implement the `ptx-prompting` helper tool for template-based prompt generation as defined in ADR-017.

## Problem Statement
The development team needs a standardized tool for generating prompts for AI assistants (Claude, ChatGPT) based on templates with placeholders. Current manual prompt creation is inefficient and inconsistent.

## Solution
Create a new Portunix helper `ptx-prompting` that:
- Loads prompt templates from files (.md, .yaml, .txt)
- Automatically detects placeholders in `{placeholder}` format
- Supports both command-line arguments and interactive input
- Provides flexible output options (stdout, clipboard, file)

## Technical Requirements

### Core Functionality
1. **Template Loading**: Support for .md, .yaml, .txt template files
2. **Placeholder Detection**: Regex-based parsing of `{placeholder}` patterns
3. **Parameter Resolution**: CLI arguments + interactive fallback for missing values
4. **Output Modes**: stdout + clipboard (default), file output, configurable via flags

### Command Structure

### Direct Helper Usage
```bash
# Direct ptx-prompting usage
ptx-prompting build prompts/translate.md  # Output to stdout AND clipboard by default
ptx-prompting build prompts/translate.md --no-copy  # Only stdout, no clipboard
ptx-prompting build prompts/translate.md --quiet  # Only clipboard, no stdout
ptx-prompting build prompts/translate.md --source_file README.cs.md --target_file README.en.md --target_language English
ptx-prompting list
ptx-prompting create translation-prompt.md
```

### Main Portunix Integration
```bash
# Via main Portunix dispatcher (requires dispatcher modification)
portunix prompt build prompts/translate.md
portunix prompt build prompts/translate.md --source_file README.cs.md --target_file README.en.md --target_language English
portunix prompt build prompts/translate.md --copy
portunix prompt list
portunix prompt create translation-prompt.md
```

### Helper Integration
- Binary name: `ptx-prompting` (with `.exe` on Windows)
- Location: Same directory as main Portunix binary
- Helper discovery: Automatic via `HelperDiscovery` system
- Standard helper commands: `--version`, `--list-commands`, `--description`
- **Main Portunix Integration**: Requires dispatcher modification to support `portunix prompt` commands

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/atotto/clipboard` - Clipboard operations
- Standard Go libraries for file operations and regex

## Implementation Plan

### Phase 1: Core Structure
- [ ] Create directory structure
- [ ] Setup go.mod with dependencies
- [ ] Create main.go entry point with helper integration
- [ ] Add dispatcher integration for `portunix prompt` → `ptx-prompting` routing
- [ ] Implement root command structure
- [ ] Create basic CLI commands (build, list, create)

### Phase 2: Template System
- [ ] Implement template parser (placeholder detection)
- [ ] Create prompt builder logic
- [ ] Add template validation
- [ ] Support multiple template formats

### Phase 3: Interactive Features
- [ ] Add interactive parameter input
- [ ] Implement clipboard integration
- [ ] Add file output options
- [ ] Create template listing functionality

### Phase 4: Templates & Documentation
- [ ] Create default template examples
- [ ] Add multilingual template support (en/, cs/)
- [ ] Write usage documentation
- [ ] Add help text and examples

## Acceptance Criteria

### Functional Requirements
- [ ] Can load templates from files
- [ ] Automatically detects all placeholders
- [ ] Supports CLI argument parameter input
- [ ] Falls back to interactive mode for missing parameters
- [ ] Outputs to stdout by default
- [ ] Optional clipboard integration with --copy flag
- [ ] Lists available templates
- [ ] Creates new template files

### Integration Requirements
- [ ] Discovered automatically by Portunix helper system
- [ ] Responds to standard helper commands
- [ ] Helper routing works correctly: `portunix prompt build` → `ptx-prompting build`
- [ ] Works on Windows and Linux
- [ ] Follows Portunix coding conventions

### Quality Requirements
- [ ] Unit tests for core functionality
- [ ] Integration tests with Portunix
- [ ] Error handling for missing files/invalid templates
- [ ] User-friendly error messages

## Example Templates

### Translation Template (en/translate.md)
```markdown
# Translation Request

Please translate the file {source_file} from {source_language} to {target_language}.
Save the result as {target_file}.

Requirements:
- Preserve all formatting
- Keep technical terms consistent
- Target audience: {audience}
```

### Code Review Template (en/review.md)
```markdown
# Code Review Request

Please review the following code changes in {file_path}:

Focus areas:
- {focus_area_1}
- {focus_area_2}
- {focus_area_3}

Language: {programming_language}
Context: {context_description}
```

## Testing Strategy
1. **Unit Tests**: Template parsing, placeholder detection, parameter resolution
2. **Integration Tests**: Helper discovery, CLI command execution
3. **E2E Tests**: Full workflow from template to output
4. **Cross-platform Tests**: Windows and Linux compatibility

## References
- ADR-017: PTX-Prompting Helper for Template-Based Prompt Generation
- Portunix helper system: `src/shared/helper.go`
- Existing helpers: `ptx-container`, `ptx-mcp`, `ptx-ansible`

## Implementation Notes
- Follow existing Portunix helper patterns
- Use consistent error handling and logging
- Implement proper input validation
- Support both absolute and relative template paths
- Consider template search paths for convenience