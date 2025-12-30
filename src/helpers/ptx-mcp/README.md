# ptx-mcp

MCP (Model Context Protocol) helper for Portunix. Provides AI assistant integration with Claude Code and other MCP-compatible clients.

## Claude Code Registration

### Automatic Registration

```bash
portunix mcp configure
```

This command automatically:
1. Detects Claude Code installation
2. Finds Portunix executable path
3. Registers Portunix as MCP server
4. Verifies the configuration

### Manual Registration

If automatic configuration fails, register manually:

```bash
claude mcp add portunix /path/to/portunix mcp serve
```

Example with common paths:

```bash
# Linux (user installation)
claude mcp add portunix ~/bin/portunix mcp serve

# Linux (system installation)
claude mcp add portunix /usr/local/bin/portunix mcp serve

# macOS
claude mcp add portunix /usr/local/bin/portunix mcp serve

# Windows (PowerShell)
claude mcp add portunix "$env:USERPROFILE\bin\portunix.exe" mcp serve
```

### Verify Registration

```bash
# List all MCP servers
claude mcp list

# Get Portunix MCP details
claude mcp get portunix
```

### Remove Registration

```bash
# Via Portunix
portunix mcp remove

# Or directly via Claude
claude mcp remove portunix
```

## Available Commands

| Command | Description |
|---------|-------------|
| `mcp serve` | Start MCP server (stdio mode by default) |
| `mcp configure` | Auto-configure with Claude Code |
| `mcp status` | Check MCP integration status |
| `mcp remove` | Remove MCP configuration |
| `mcp init` | Interactive configuration wizard |
| `mcp config` | View/edit MCP configuration |
| `mcp test` | Test MCP server connection |

## Server Modes

```bash
# stdio mode (default, recommended for Claude Code)
portunix mcp serve

# TCP mode
portunix mcp serve --mode tcp --port 3001

# Unix socket mode
portunix mcp serve --mode unix --socket /tmp/portunix.sock
```

## MCP Tools Provided

When registered, Portunix exposes these tools to AI assistants:

- `get_system_info` - OS detection and system information
- `detect_project_type` - Analyze project structure
- `list_packages` - Available packages for installation
- `install_package` - Install development tools
- `vm_*` - Virtual machine management
- `container_*` - Container operations
- And more...

## Troubleshooting

### Claude Code not found

```bash
# Check if Claude Code is installed
which claude

# Install Claude Code
curl -fsSL https://claude.ai/cli/install.sh | sh
```

### MCP server not responding

```bash
# Check status
portunix mcp status --verbose

# Test server manually
portunix mcp serve --mode stdio
```

### Permission issues

Ensure Portunix binary is executable:

```bash
chmod +x ~/bin/portunix
chmod +x ~/bin/ptx-mcp
```
