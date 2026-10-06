# Issue #37: Revert Default stdio Mode and Implement MCP Serve Command

## Overview
Implement the architectural changes defined in ADR-005 to revert the default stdio mode for MCP and replace the legacy `mcp-server` command with a new `mcp serve` command structure.

## Background
Based on ADR-005, we need to:
1. Revert the behavior introduced in ADR-004 where Portunix would enter stdio mode by default
2. Return to displaying help when no parameters are provided
3. Replace `./portunix mcp-server` with `portunix mcp serve`
4. Update AI assistant installation processes

## Requirements

### Core Changes
- [ ] **Restore Help Display**: When executed without parameters, show help text instead of entering stdio mode
- [ ] **Replace mcp-server Command**: Remove legacy `mcp-server` command and implement `mcp serve` with enhanced parameters
- [ ] **Communication Mode Support**: Add support for different communication modes (stdio, TCP, unix socket)
- [ ] **Parameter Handling**: Implement proper parameter parsing for `mcp serve` command

### Command Structure
New command structure should support:
```bash
# Basic stdio mode (default)
portunix mcp serve

# Explicit stdio mode
portunix mcp serve --mode stdio

# TCP mode with port specification
portunix mcp serve --mode tcp --port 8080

# Unix socket mode
portunix mcp serve --mode unix --socket /tmp/portunix.sock
```

### AI Assistant Integration Updates
- [ ] **Update mcp configure Command**: Modify `portunix mcp configure` to use new `mcp serve` structure
- [ ] **Installation Script Updates**: Update all installation guides and scripts
- [ ] **Claude Code Integration**: Update Claude Code configuration examples
- [ ] **VS Code Extension**: Update VS Code extension integration if applicable

## Technical Implementation

### Files to Modify
1. **main.go**: Restore original argument parsing logic to show help by default
2. **cmd/mcp.go**: 
   - Remove `mcp serve` command
   - Implement `mcp serve` command with parameter support
   - Add communication mode handling (stdio, tcp, unix socket)
3. **cmd/root.go**: Ensure help is displayed by default when no arguments provided
4. **Documentation**: Update all references from `mcp-serve` to `mcp serve`

### Migration Considerations
- Ensure backward compatibility during transition period
- Update error messages to guide users from old to new command structure
- Consider deprecation warnings for any legacy command usage

## Testing Requirements
- [ ] **Unit Tests**: Test new `mcp serve` command with various parameters
- [ ] **Integration Tests**: Test AI assistant integration with new command structure
- [ ] **Migration Tests**: Verify smooth transition from old to new commands
- [ ] **Help Display Tests**: Ensure help is properly displayed when no arguments provided

## Documentation Updates
- [ ] **README.md**: Update MCP integration examples
- [ ] **AI Integration Guides**: Update all AI assistant setup instructions
- [ ] **Command Documentation**: Document new `mcp serve` command and parameters
- [ ] **Migration Guide**: Create migration guide for existing installations

## Success Criteria
1. Running `portunix` without parameters displays help (not stdio mode)
2. `portunix mcp serve` successfully enters stdio mode
3. All communication modes (stdio, tcp, unix socket) work correctly
4. AI assistants can be configured with new command structure
5. Legacy `mcp serve` command is completely removed
6. All documentation reflects new command structure

## Priority
**High** - This implements an architectural decision that affects user experience and AI assistant integration

## Labels
- enhancement
- mcp
- ai-integration  
- cli
- breaking-change
- command-restructure

## Related Issues
- Related to ADR-005: Revert Default stdio Mode for MCP
- Supersedes functionality from Issue #036 (Default stdio Mode for MCP)
- Connected to Issue #034 (MCP Server Command Restructuring)

## Implementation Notes
This change represents a significant architectural shift that prioritizes standard CLI conventions while maintaining flexible MCP server functionality. The implementation should be done carefully to ensure minimal disruption to existing AI assistant integrations.

## Acceptance Criteria
- [ ] Help is displayed when running `portunix` without parameters
- [ ] `portunix mcp serve` works in stdio mode
- [ ] Additional communication modes are implemented and functional
- [ ] All AI assistant integration documentation is updated
- [ ] Migration path is clear and documented
- [ ] No breaking changes for users who explicitly use MCP commands

## Created
2025-09-11

## Status
📋 Open