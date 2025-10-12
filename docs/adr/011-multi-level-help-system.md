# ADR 011: Multi-Level Help System

## Context
The current Portunix help system combines different levels of detail into a single output, causing the following issues:
- For regular users, the help is too detailed and overwhelming
- For experts, in-depth information about all options is missing
- For AI assistants (LLMs), the format is difficult to parse due to excessive formatting and unsystematic structure
- One help format cannot effectively serve all types of users

Alternatives considered:
1. Keep unified help with improved formatting
2. Use external documentation for experts
3. Split help into multiple levels based on target audience

## Decision
We will implement a three-tier help system with separate outputs:

1. **`portunix --help`** - Basic help
   - Only the most important commands
   - Brief one-line description for each command
   - **Mandatory footer with available help levels and their purpose**
   - Reference to extended help
   - Maximum 30-40 lines of output

2. **`portunix --help-expert`** - Extended help for experts
   - Complete list of all commands and switches
   - Detailed description of each function
   - Usage examples
   - Environment variables and configuration
   - Advanced workflows and best practices

3. **`portunix --help-ai`** - Structured help for LLMs
   - Machine-readable format (structured JSON)
   - Systematic command categorization
   - Explicit parameters and their types
   - Minimal formatting
   - Consistent structure for easy parsing

## Consequences

### Positive
- **Better UX for regular users**: Quick overview of basic commands without information overload
- **Full documentation for experts**: All details available without reading external documentation
- **AI optimization**: LLMs can efficiently parse and utilize system commands
- **Easier maintenance**: Each help type can be modified independently
- **Backward compatibility**: `--help` maintains basic functionality

### Negative
- **Increased code complexity**: Need to maintain three different help formats
- **Risk of inconsistency**: Changes must be propagated to all three levels
- **Larger codebase**: More functions for generating different formats

### Implementation Details
- Central command definitions in structured form (e.g., struct or JSON)
- Generators for each output type from central definition
- Tests to ensure consistency between all levels
- Command `portunix help <command>` will respect detail level based on used switch
- **Mandatory requirement**: Basic help (`--help`) must always display a "Help levels" section explaining all three available help options and their purposes
- The help level section must be clearly visible and formatted consistently across all commands

### Output Examples

**Basic (`--help`):**
```
Portunix - Universal development environment management tool

Usage: portunix [command] [options]

Common commands:
  install     Install packages and tools
  docker      Manage Docker containers
  plugin      Manage plugins
  update      Update Portunix

Help levels:
  --help         This help - basic commands and usage (current)
  --help-expert  Extended help with all options, examples, and advanced features
  --help-ai      Machine-readable format optimized for AI/LLM parsing

Use 'portunix help <command>' for command details
```

**Expert (`--help-expert`):**
```
[Complete documentation including all switches, env variables, examples...]
```

**AI (`--help-ai`):**
```markdown
## Commands
### install
- description: Install packages and tools
- parameters:
  - package: string, required
  - variant: string, optional
  - dry-run: boolean, optional
- examples: ["portunix install nodejs", "portunix install python --variant full"]
...
```

### Migration Plan
1. Phase 1: Implement basic and expert help (keep current as expert)
2. Phase 2: Simplify basic help
3. Phase 3: Add AI format
4. Phase 4: Optimization and testing with various LLMs