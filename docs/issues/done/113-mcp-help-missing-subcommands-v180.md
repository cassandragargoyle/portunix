# Issue #113: MCP Help Missing Subcommands in v1.8.0 Release

## Overview

| Field | Value |
|-------|-------|
| **Issue ID** | #113 |
| **Title** | MCP Help Missing Subcommands in v1.8.0 Release |
| **Status** | 📋 Open |
| **Priority** | High |
| **Type** | Bug Fix |
| **Labels** | bug, mcp, release, help-system, regression |
| **Created** | 2025-12-25 |
| **Affected Version** | 1.8.0 |

## Problem Description

The deployed version 1.8.0 of Portunix shows only the `serve` subcommand in `portunix mcp --help`, while the current development version includes all MCP subcommands.

### Expected Behavior (from current dev build)

```
Available Commands:
  config      Manage MCP server configuration
  configure   Configure Portunix MCP server integration with Claude Code
  init        Initialize MCP server configuration with interactive wizard
  reconfigure Reconfigure Portunix MCP server with new settings (e.g., different port)
  remove      Remove Portunix MCP server integration from Claude Code
  serve       Start MCP server for AI assistant integration
  start       Start MCP server for AI assistant integration
  status      Check status of Portunix MCP server integration with Claude Code
  stop        Stop running MCP server
  test        Test MCP server connection with AI assistants
```

### Actual Behavior (v1.8.0 release)

```
Available Commands:
  serve       Start MCP server for AI assistant integration
```

## Impact

- Users cannot configure MCP server integration using `portunix mcp configure`
- Users cannot check MCP status using `portunix mcp status`
- Users cannot remove MCP configuration using `portunix mcp remove`
- Essential MCP management functionality is missing

## Root Cause Analysis

Needs investigation:
1. Check if MCP subcommands were properly registered in cobra command tree for v1.8.0
2. Check if release build process excluded MCP command files
3. Compare dist binary with local build

## Reproduction Steps

1. Use deployed binary from `~/bin/portunix` (v1.8.0)
2. Run `portunix mcp --help`
3. Observe only `serve` subcommand is listed

## Verification Commands

```bash
# Deployed version (broken)
~/bin/portunix --version        # Shows: Portunix version 1.8.0
~/bin/portunix mcp --help       # Shows only: serve

# Local dev version (working)
./portunix mcp --help           # Shows all subcommands
```

## Acceptance Criteria

- [ ] All MCP subcommands visible in `portunix mcp --help`
- [ ] `configure`, `status`, `remove`, `reconfigure` commands functional
- [ ] Release process verified to include all command registrations
- [ ] Test added to verify MCP command availability

## Files to Investigate

- `cmd/mcp.go` - Main MCP command registration
- `cmd/mcp_*.go` - Individual subcommand files
- `.goreleaser.yml` - Release build configuration
- `dist/` - Built release artifacts

## Related Issues

- #004 - MCP Server for AI Assistant Integration (original implementation)
- #036 - Default stdio Mode for MCP
- #037 - Revert Default stdio Mode and Implement MCP Serve Command
