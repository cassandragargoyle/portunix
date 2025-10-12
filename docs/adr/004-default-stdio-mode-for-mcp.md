# ADR-004: Default stdio Mode for MCP When No Parameters Provided

## Status
Superseded by ADR-005

## Context
Portunix currently displays help text when executed without parameters, which is standard CLI behavior. However, with the introduction of MCP (Model Context Protocol) server functionality, AI assistants need a way to interact with Portunix in stdio mode directly.

The current behavior requires explicit `mcp` command to enter stdio mode:
- `portunix` → shows help
- `portunix mcp` → enters stdio mode

This creates friction for AI assistants that need to interact with Portunix as an MCP server, as they must know to append the `mcp` parameter.

## Decision
Change the default behavior when Portunix is executed without parameters:
- When no arguments are provided, Portunix will automatically enter MCP stdio mode
- Help text will only be displayed when explicitly requested with `--help` or `-h` flags
- This aligns with MCP server convention where tools run in stdio mode by default

Implementation details:
1. Modify main.go to detect when no arguments are provided
2. In this case, directly execute MCP stdio mode
3. Keep explicit `portunix mcp` command for backwards compatibility
4. Ensure help is still accessible via standard flags

## Consequences

### Positive
- **Better AI integration**: AI assistants can use Portunix directly without knowing specific commands
- **MCP compliance**: Follows MCP server conventions for stdio mode
- **Simplified configuration**: AI tools configuration becomes simpler (just specify binary path)
- **Future-proof**: As AI integration becomes more important, this default makes more sense

### Negative
- **Breaking change for users**: Users who run `portunix` expecting help will get stdio mode instead
- **Less discoverable**: New users won't immediately see available commands
- **Potential confusion**: Users might think the tool is hanging when it enters stdio mode

### Mitigation
- Clear documentation about the change
- Add startup message in stdio mode explaining how to exit and get help
- Consider adding timeout or detection of interactive terminal vs pipe

## Implementation Notes
The change primarily affects:
- `main.go`: Argument parsing logic
- `cmd/root.go`: Root command behavior
- Documentation: Update README and user guides

## Date
2025-09-11

## Author
Software Architect (via Claude Code)