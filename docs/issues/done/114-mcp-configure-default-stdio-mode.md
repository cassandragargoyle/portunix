# Issue #114: MCP Configure Should Default to stdio Mode

## Overview

| Field | Value |
|-------|-------|
| **Issue ID** | #114 |
| **Title** | MCP Configure Should Default to stdio Mode |
| **Status** | ✅ Implemented |
| **Priority** | Medium |
| **Type** | Enhancement |
| **Labels** | mcp, configuration, ux |
| **Created** | 2025-12-25 |

## Problem Description

The `portunix mcp configure` command currently configures MCP server in HTTP/TCP mode. However, Claude Code's recommended integration method is **stdio mode**, which is simpler and doesn't require port management.

### Current Behavior

```bash
portunix mcp configure
# Registers: portunix mcp serve --mode tcp --port 3001
```

### Expected Behavior

```bash
# Default: stdio mode (recommended)
portunix mcp configure
# Registers: portunix mcp serve

# Optional: HTTP/TCP mode with explicit parameter
portunix mcp configure --mode tcp --port 3001
# Registers: portunix mcp serve --mode tcp --port 3001
```

## Rationale

1. **stdio is Claude Code's preferred mode** - Direct process communication without network overhead
2. **No port conflicts** - stdio doesn't require managing port availability
3. **Simpler setup** - No firewall or network configuration needed
4. **Security** - No network exposure, communication only via stdin/stdout

## Proposed Changes

### 1. Update `configure.go`

Change `addMCPServerToClaudeCode()` to use stdio by default:

```go
// Current
args := []string{"mcp", "add", "portunix", portunixPath, "mcp", "serve", "--port", strconv.Itoa(port)}

// Proposed (default stdio)
args := []string{"mcp", "add", "portunix", portunixPath, "mcp", "serve"}

// Proposed (explicit TCP mode)
if mode == "tcp" {
    args = append(args, "--mode", "tcp", "--port", strconv.Itoa(port))
}
```

### 2. Add `--mode` flag to configure command

```go
configureCmd.Flags().StringP("mode", "m", "stdio", "Server mode: stdio (default), tcp, unix")
```

### 3. Add `--scope` flag for local vs global configuration

```go
configureCmd.Flags().StringP("scope", "s", "local", "Configuration scope: local (default, .mcp.json), user (global), project")
```

### 4. Update help text

Clarify that stdio is the recommended mode for Claude Code integration.

## Acceptance Criteria

- [x] `portunix mcp configure` uses stdio mode by default
- [x] `portunix mcp configure --mode tcp --port 3001` enables TCP mode
- [x] `portunix mcp configure --mode unix` enables Unix socket mode
- [x] `portunix mcp configure --scope local` saves to local .mcp.json (default)
- [x] `portunix mcp configure --scope user` saves to global/user configuration
- [x] Help text updated to reflect default behavior
- [x] `portunix mcp reconfigure` also supports --mode and --scope flags
- [ ] README.md updated with new usage examples

## Files Modified

- `src/helpers/ptx-mcp/configure.go` - Added --mode and --scope flags, updated logic
- `src/helpers/ptx-mcp/reconfigure.go` - Added --mode and --scope flags, updated logic
- `src/helpers/ptx-mcp/init.go` - Updated wizard to use stdio mode and local scope

## Related Issues

- #036 - Default stdio Mode for MCP
- #037 - Revert Default stdio Mode and Implement MCP Serve Command
- #113 - MCP Help Missing Subcommands in v1.8.0
