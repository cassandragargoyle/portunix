# Issue #50: Multi-Level Help System

**Status:** ✅ Implemented
**Priority:** Medium
**Type:** Enhancement
**Created:** 2025-09-17
**ADR:** [011-multi-level-help-system.md](../../adr/011-multi-level-help-system.md)

## Summary
Implement a three-tier help system to better serve different user types: regular users, experts, and AI assistants.

## Problem Description
The current help system has several limitations:
- For regular users, the help output is too detailed and overwhelming
- For experts, in-depth information about all options is missing
- For AI assistants (LLMs), the format is difficult to parse due to excessive formatting
- One help format cannot effectively serve all types of users

## Proposed Solution
Implement three separate help levels based on ADR 011:

### 1. Basic Help (`--help`)
- Only the most important commands
- Brief one-line description for each command
- Mandatory footer with available help levels
- Maximum 30-40 lines of output

### 2. Expert Help (`--help-expert`)
- Complete list of all commands and switches
- Detailed description of each function
- Usage examples
- Environment variables and configuration
- Advanced workflows and best practices

### 3. AI Help (`--help-ai`)
- Machine-readable format (structured JSON)
- Systematic command categorization
- Explicit parameters and their types
- Minimal formatting
- Consistent structure for easy parsing

## Implementation Requirements

### Mandatory Features
- **Help Levels Section**: Basic help must always display a "Help levels" section explaining all three options
- **Central Definition**: Commands defined centrally to maintain consistency
- **Format Generators**: Separate generators for each output type
- **Consistency Tests**: Automated tests to ensure all levels stay synchronized

### Example Output Format
```
Help levels:
  --help         This help - basic commands and usage (current)
  --help-expert  Extended help with all options, examples, and advanced features
  --help-ai      Machine-readable format optimized for AI/LLM parsing
```

## Technical Implementation

### Architecture
1. Create central command registry with structured definitions
2. Implement three output generators:
   - `generateBasicHelp()`
   - `generateExpertHelp()`
   - `generateAIHelp()`
3. Modify command parser to recognize new flags
4. Update `help <command>` to respect selected help level

### Data Structure
```go
type Command struct {
    Name        string
    Brief       string      // For basic help
    Description string      // For expert help
    Parameters  []Parameter // For all levels
    Examples    []string    // For expert help
    Category    string      // For AI help
}
```

### JSON Output Format (AI Help)
```json
{
  "commands": [
    {
      "name": "install",
      "description": "Install packages and tools",
      "parameters": [
        {"name": "package", "type": "string", "required": true},
        {"name": "variant", "type": "string", "required": false},
        {"name": "dry-run", "type": "boolean", "required": false}
      ],
      "examples": [
        "portunix install nodejs",
        "portunix install python --variant full"
      ]
    }
  ]
}
```

## Migration Plan

### Phase 1: Foundation ✅ Completed
- [x] Create central command registry
- [x] Implement basic help generator
- [x] Keep current help as expert level

### Phase 2: Simplification ✅ Completed
- [x] Simplify basic help to 22 lines
- [x] Add help levels section to all outputs
- [x] Remove docker from basic help
- [x] Remove duplicate help command syntax

### Phase 3: AI Format ✅ Completed
- [x] Implement JSON output generator
- [x] Add structured metadata to commands
- [x] Use full description in AI format

### Phase 4: Polish ✅ Completed
- [x] Add consistency tests
- [x] Update help text references
- [x] Test all three formats

## Acceptance Criteria
1. `--help` displays only essential commands in 30-40 lines
2. `--help` includes mandatory "Help levels" section
3. `--help-expert` shows complete documentation
4. `--help-ai` outputs valid JSON with all command metadata
5. All three formats stay synchronized when commands change
6. Tests verify consistency between all help levels
7. Performance: Help generation takes <100ms

## Testing Requirements
- Unit tests for each generator
- Integration tests for flag parsing
- Consistency tests between formats
- User acceptance testing for readability
- AI tool compatibility testing

## Documentation Updates
- Update main README with new help options
- Add developer guide for maintaining help system
- Create examples of all three formats
- Update Claude.md with AI help usage

## Related Issues
- None

## Implementation Details

### Completed Features
- Central command registry in `cmd/help_registry.go`
- Three help levels fully functional:
  - Basic help: 22 lines with essential commands
  - Expert help: 130 lines with full documentation
  - AI help: Valid JSON format
- Docker removed from basic help (only in expert)
- Duplicate `help <command>` syntax removed
- Only `<command> --help` syntax supported
- "Universal environment management tool" (removed "development")
- Full description in both expert and AI formats
- Integration tests in `test/integration/issue_050_multi_level_help_test.go`

### Branch
- `feature/issue-050-multi-level-help` (ready for merge)

## Notes
- Based on Architecture Decision Record (ADR) 011
- Prioritize backward compatibility - existing scripts using `--help` should work
- Consider adding `--help-format=json` as alternative to `--help-ai`