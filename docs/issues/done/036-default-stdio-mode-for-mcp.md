# Issue #36: Default stdio Mode for MCP When No Parameters Provided

## Type
Enhancement

## Priority
High

## Status
✅ Implemented

## Labels
enhancement, mcp, ai-integration, cli, breaking-change, stdio, protocol

## Summary
Implement the architecture decision defined in ADR-004 to change Portunix's default behavior when executed without parameters. Instead of showing help text, Portunix should automatically enter MCP stdio mode to improve AI assistant integration and enable direct MCP protocol communication.

## Background
Based on ADR-004 (docs/adr/004-default-stdio-mode-for-mcp.md), this change will make Portunix more AI-friendly by defaulting to stdio mode when no arguments are provided. This aligns with MCP server conventions and simplifies AI tool configuration.

## Motivation
- **AI Assistant Integration**: When AI assistants (like Claude Code) start Portunix with MCP protocol, it should immediately be ready for stdio communication
- **Simplified Workflow**: No need to separately start MCP server - direct protocol communication
- **Performance**: Faster startup for AI-driven operations
- **Standards Compliance**: Following MCP specification for stdio transport

## Current State Analysis
Current MCP implementation:
- ✅ `portunix mcp` - Management commands (configure, remove, status)
- ✅ `portunix mcp serve` - Explicit server startup with daemon/remote modes
- ❌ **Missing**: Direct stdio mode when called without parameters
- 🔄 **Conflict**: Need to reorganize command structure for consistency

## Implementation Requirements

### 1. Default Behavior Change
- Modify `main.go` to detect when no arguments are provided
- When `os.Args` has length 1 (only program name), automatically enter MCP stdio mode
- Skip all command-line output (no welcome messages, prompts, etc.)
- Implement MCP JSON-RPC communication protocol immediately

### 2. Command Structure Reorganization
```bash
# New behavior:
portunix                            # Direct MCP stdio mode (NEW DEFAULT)
portunix --help / -h                # Show help text
portunix help                       # Show help text

# Existing commands remain:
portunix mcp configure              # MCP configuration management
portunix mcp remove                 # Remove MCP configuration
portunix mcp status                 # Show MCP status
portunix mcp serve                  # Explicit server modes (daemon/remote)
```

### 3. MCP stdio Protocol Implementation
When in stdio mode, Portunix should:
1. Read JSON-RPC messages from stdin
2. Write JSON-RPC responses to stdout
3. Write logs/errors to stderr (optional, based on config)
4. Implement MCP protocol:
   - initialize/initialized handshake
   - tools/list, tools/call
   - resources/list, resources/read
   - prompts/list, prompts/get

### 4. Terminal Detection
- Detect if running in interactive terminal vs pipe
- If interactive terminal detected AND no MCP client detected:
  - Show brief message to stderr: "MCP stdio mode active. Press Ctrl+C to exit. Use 'portunix --help' for help."
- If pipe/non-interactive OR MCP client detected:
  - Enter stdio mode silently

### 5. Available Tools
Provide same tools as mcp serve over stdio transport:
```json
{
  "tools": [
    "system_info",
    "package_install",
    "project_detect",
    "container_manage",
    "vm_manage",
    "environment_setup"
  ]
}
```

### 6. Security Considerations
- Same security profiles as mcp serve (development, standard, restricted)
- Configuration from file (~/.portunix/mcp server.json)
- Default to restricted mode for stdio (safer for AI assistants)

## Technical Implementation Details

### File Changes Required

1. **main.go**
   - Add argument checking before `cmd.Execute()`
   - If no args, directly call MCP stdio handler
   - Add terminal detection logic

2. **cmd/root.go**
   - Ensure root command doesn't interfere with new default behavior
   - Update command descriptions

3. **cmd/mcp.go**
   - Keep existing management subcommands (configure, remove, status)
   - Extract stdio mode logic into reusable function

4. **app/mcp/stdio.go** (new file)
   - Implement MCP stdio handler
   - Reuse existing MCP tools from mcp-server implementation
   - Add protocol handshake (initialize/initialized)

### Example Implementation Approach
```go
// In main.go
func main() {
    // ... existing initialization code ...
    
    // Check if no arguments provided
    if len(os.Args) == 1 {
        // Optional: Check if terminal is interactive
        if isInteractiveTerm() && !isMCPClient() {
            fmt.Fprintf(os.Stderr, "MCP stdio mode active. Press Ctrl+C to exit. Use 'portunix --help' for help.\n")
        }
        
        // Enter MCP stdio mode directly
        mcp.RunStdioMode()
        return
    }
    
    // Normal command execution
    cmd.Execute()
}
```

### Configuration Integration
Use existing MCP configuration with stdio-specific options:
```json
{
  "server_type": "stdio",
  "security_profile": "development",
  "stdio_mode": {
    "default_timeout": "30s",
    "log_level": "error",
    "tools_enabled": ["system_info", "package_install"]
  }
}
```

## Testing Requirements
1. Test default behavior (no args) → enters stdio mode
2. Test explicit help flags still work
3. Test existing `portunix mcp` subcommands still work
4. Test MCP functionality in stdio mode
5. Test with AI assistants (Claude Code, Cursor, Gemini CLI)
6. Test terminal detection logic
7. Test JSON-RPC protocol compliance
8. Performance test: startup time < 1 second

## Migration Considerations
- This is a **breaking change** for users who expect help when running `portunix`
- Update all documentation to reflect new behavior
- Add migration note in release notes
- Update installation scripts if they test for help output
- Consider deprecation warnings for old workflows

## Documentation Updates Required
1. README.md - Update usage section
2. MCP documentation - Explain new default behavior
3. Release notes - Highlight breaking change
4. AI integration guides - Simplify configuration instructions
5. Command help texts - Update all descriptions

## Claude Code Integration Example
```json
{
  "portunix": {
    "command": "portunix"  // No args needed - stdio mode is default
  }
}
```

## Acceptance Criteria
- [ ] Running `portunix` without arguments enters MCP stdio mode
- [ ] Running `portunix --help` or `portunix -h` shows help
- [ ] Running `portunix mcp configure/remove/status` still works
- [ ] MCP stdio mode implements full JSON-RPC protocol
- [ ] Same tool availability as mcp serve
- [ ] Integration works with Claude Code, Claude Desktop, Gemini CLI
- [ ] Performance: startup time < 1 second
- [ ] Security: respects configuration profiles
- [ ] Error handling: graceful protocol failures
- [ ] Documentation is updated to reflect new behavior
- [ ] Tests pass for all scenarios

## References
- ADR-004: `/docs/adr/004-default-stdio-mode-for-mcp.md`
- Related Issue: #004 (MCP Server for AI Assistant Integration)
- Related Issue: #034 (MCP Server Command Restructuring)
- MCP Documentation: `/docs/mcp/`
- MCP Specification: https://modelcontextprotocol.io/

## Notes for Developers
- Carefully review ADR-004 before implementation
- Ensure no code duplication between default and explicit MCP entry
- Leverage existing mcp serve tools and security profiles
- Consider user experience for both CLI users and AI assistants
- Test thoroughly with actual AI tools before merging
- Protocol compliance is critical - follow MCP specification exactly

## Priority
**High** - Critical for seamless AI assistant integration and MCP protocol compliance.

---
*Created: 2025-09-11*
*Last updated: 2025-09-11*