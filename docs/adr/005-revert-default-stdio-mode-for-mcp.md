# ADR-005: Revert Default stdio Mode for MCP - Return to Help Display

## Status
Accepted

## Context
ADR-004 introduced a change where Portunix would automatically enter MCP stdio mode when executed without parameters, instead of displaying help text. After implementation and evaluation, this approach has proven problematic:

1. **User Experience Issues**: Users expecting standard CLI behavior (help display) are confused when the tool enters stdio mode
2. **Discovery Problems**: New users cannot easily discover available commands
3. **Standard CLI Conventions**: Most CLI tools display help when run without arguments
4. **Terminal Hanging**: Users think the application is unresponsive in stdio mode

The original goal of simplifying AI assistant integration can be achieved through explicit command structure while maintaining standard CLI conventions.

## Decision
Revert the behavior introduced in ADR-004 and return to standard CLI conventions:
- When no arguments are provided, Portunix will display help text (original behavior)
- MCP server functionality remains available through explicit commands:
  - `portunix mcp serve` → enters stdio mode
  - `portunix mcp serve --mode stdio` → explicit stdio mode
  - `portunix mcp serve --mode tcp --port 8080` → TCP mode with port specification
  - Additional parameters can specify communication methods and configuration

Implementation details:
1. Modify main.go to restore original help display behavior
2. Replace `mcp serve` command with `mcp serve` command structure
3. Enhance `mcp serve` command with proper parameter handling
4. Add communication mode parameters (stdio, tcp, unix socket, etc.)
5. Update AI assistant installation process to use new command structure
6. Maintain clear command structure for AI assistant configuration

## Consequences

### Positive
- **Standard CLI Behavior**: Users get expected help display when running without parameters
- **Better Discoverability**: New users can immediately see available commands
- **Clear Command Structure**: Explicit `mcp serve` makes purpose obvious
- **Flexible Configuration**: Parameters allow different communication modes
- **No Breaking Changes**: Returns to familiar behavior for existing users

### Negative
- **Slightly More Complex AI Configuration**: AI assistants need to specify `mcp serve` instead of just binary path
- **Additional Command Length**: Longer command for AI assistant configuration

### Mitigation
- Clear documentation for AI assistant integration
- Simple examples in README showing proper MCP configuration
- Consider shell aliases or wrapper scripts for frequent usage

## Implementation Notes
The change affects:
- `main.go`: Restore original argument parsing logic
- `cmd/mcp.go`: Replace `mcp serve` with `mcp serve` command and enhance with communication parameters
- `cmd/root.go`: Ensure help is displayed by default
- AI assistant installation processes: Update from `./portunix mcp serve` to `portunix mcp serve`
- Documentation: Update integration guides for AI assistants with new command structure

## Rationale for Reversion
1. **User-First Design**: CLI tools should prioritize human user experience
2. **Convention Compliance**: Follow established CLI tool patterns
3. **Clear Intent**: `mcp serve` clearly indicates server mode activation
4. **Flexibility**: Parameters allow various communication modes beyond stdio

## Migration Path
For AI assistants and existing MCP integrations:

### From bare command:
- Change configuration from: `portunix`
- To: `portunix mcp serve`
- Additional options: `portunix mcp serve --mode stdio` (explicit)

### From legacy mcp serve:
- Change configuration from: `./portunix mcp serve`
- To: `portunix mcp serve`
- Update all AI assistant installation scripts and documentation

### AI Assistant Installation Process:
- Update `portunix mcp configure` command to use new `mcp serve` structure
- Modify installation guides to reflect new command syntax
- Update Claude Code, VS Code extensions, and other AI assistant integrations

## Date
2025-09-11

## Author
Software Architect (via Claude Code)

## Supersedes
This ADR supersedes and reverts ADR-004: Default stdio Mode for MCP When No Parameters Provided