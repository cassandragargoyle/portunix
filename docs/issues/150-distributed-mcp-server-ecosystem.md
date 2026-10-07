# Issue #150: Distributed MCP Server Ecosystem for Plugins

## Summary

Evolve the Portunix MCP architecture from a single central server to a distributed ecosystem where plugins can extend the central MCP server, provide their own standalone MCP server, or both. Add automated registration/unregistration of plugin MCP servers in AI assistants (Claude Code, Gemini CLI, etc.).

**Architecture Decision**: [ADR-038](../adr/038-distributed-mcp-server-ecosystem.md)

## Current State

- Central MCP server (`ptx-mcp`) has hardcoded handlers only
- Plugin manifest `ai_integration.mcp_tools` field is unused metadata
- No mechanism for plugins to contribute tools to the central MCP server
- No support for registering plugin-provided MCP servers in AI assistants
- If a plugin has an MCP server, users must manually configure it

## Desired State

```bash
# Central MCP now includes proxied plugin tools
portunix mcp serve
# → tools/list returns core tools + enabled plugin tools (extend mode)
# → tools/call routes to plugin via gRPC or exec

# Plugin standalone MCP server
portunix plugin mcp-serve fulltext-search
# → Launches plugin's own MCP server via stdio

# Register everything into Claude Code
portunix mcp configure --with-plugins
# → Registers central MCP + all standalone plugin MCP servers

# Register single plugin MCP
portunix mcp configure --plugin fulltext-search

# List all MCP servers (central + plugins)
portunix mcp list

# Plugin install auto-offers MCP registration
portunix plugin install fulltext-search
# → "Plugin provides MCP server. Register with Claude Code? [Y/n]"

# Plugin uninstall offers MCP cleanup
portunix plugin uninstall fulltext-search
# → "Remove MCP server from Claude Code? [Y/n]"
```

## Requirements

### FR-1: Plugin Manifest Extension

Extend `ai_integration` section in `plugin.json`:

```json
{
  "ai_integration": {
    "mcp_mode": "extend|standalone|both",
    "mcp_tools": [
      {
        "name": "tool_name",
        "description": "Tool description",
        "input_schema": {
          "type": "object",
          "properties": { ... },
          "required": [...]
        }
      }
    ],
    "mcp_server": {
      "command": ["./binary", "mcp", "serve"],
      "name": "server-name",
      "description": "Server description",
      "mode": "stdio"
    }
  }
}
```

**Affected types:**

- `AIIntegrationConfig` — add `MCPMode string` and `MCPServer *MCPServerConfig`
- `MCPTool` — add `InputSchema json.RawMessage`
- New `MCPServerConfig` struct

**Files:** `src/app/plugins/types.go`, `src/app/plugins/manifest.go`

### FR-2: Central MCP Proxy (extend mode)

When `portunix mcp serve` starts:

1. Discover enabled plugins with `mcp_mode: "extend"` or `"both"`
2. Register proxy handlers for each plugin tool
3. On `tools/list` — return core + plugin tools
4. On `tools/call` for a plugin tool:
   - gRPC plugin: forward via `Execute()` with tool name and JSON params
   - Helper plugin: invoke binary with JSON on stdin, read JSON from stdout

**Files:** `src/app/mcp/server.go` (handler discovery, proxy handlers), `src/helpers/ptx-mcp/` (plugin bridge)

### FR-3: Standalone MCP Support

New command: `portunix plugin mcp-serve <plugin-name>`

- Resolve plugin path from registry
- Execute `mcp_server.command` from manifest
- Bridge stdin/stdout for AI assistant

**Files:** `src/cmd/plugin.go` (new subcommand), plugin manager

### FR-4: MCP Registration for Plugins

Extend `portunix mcp configure`:

- `--with-plugins` — register central + all standalone plugin MCP servers
- `--plugin <name>` — register specific plugin's standalone MCP server

Auto-registration:

- `portunix plugin install` — detect `mcp_server` in manifest, prompt for registration
- `portunix plugin uninstall` — detect registered MCP server, prompt for removal

**Files:** `src/helpers/ptx-mcp/configure.go`, `src/cmd/plugin.go`

### FR-5: MCP Listing and Management

Extend `portunix mcp list` to show:

- Central MCP server status
- Plugin MCP servers (extend mode — proxied tools)
- Plugin MCP servers (standalone — registered in AI assistant)

New: `portunix mcp remove --plugin <name>` — unregister plugin MCP from AI assistant

**Files:** `src/helpers/ptx-mcp/`

### FR-6: Tool Name Collision Prevention

- Namespace convention: plugin tools use `plugin_name.tool_name` format
- Validate uniqueness during registration
- Error on collision with core tools or other plugins

## Architecture Decision Points

### DP-1: How to bridge plugin tools in central MCP?

| Option | Description | Pros | Cons |
| ------ | ----------- | ---- | ---- |
| **A: gRPC Execute()** (recommended for service plugins) | Use existing `Execute()` gRPC method | Reuses existing infrastructure, plugin already running | Requires running gRPC service |
| **B: Exec with JSON stdin/stdout** (recommended for helper plugins) | Invoke plugin binary, pass JSON on stdin | Works with any plugin type, no running service needed | Process startup overhead per call |
| **C: Shared library / plugin loading** | Load plugin as Go module | Zero overhead | Go plugin system unreliable, platform-dependent |

**Decision**: Option A for gRPC plugins, Option B for helper plugins — select based on `plugin.type` field.

### DP-2: Where to store plugin MCP server registrations?

| Option | Description |
| ------ | ----------- |
| **A: AI assistant config directly** (recommended) | Use `claude mcp add` / `claude mcp remove` — same as current central MCP |
| **B: Portunix registry** | Track in `~/.portunix/mcp/plugins.json`, sync on demand |

**Decision**: Option A — direct AI assistant configuration, consistent with current `portunix mcp configure`.

### DP-3: Backward compatibility of `mcp_tools` field

Existing plugins have `mcp_tools` with only `name` and `description`. New schema adds `input_schema`.

| Option | Description |
| ------ | ----------- |
| **A: mcp_mode required** (recommended) | `mcp_tools` remain metadata-only unless `mcp_mode` is explicitly set |
| **B: Auto-detect** | If `input_schema` present, auto-enable extend mode |

**Decision**: Option A — explicit opt-in via `mcp_mode` prevents breaking existing plugins.

## Implementation Phases

### Phase 1: Manifest Extension

- [ ] Add `MCPMode`, `MCPServerConfig` to `types.go`
- [ ] Add `InputSchema` to `MCPTool` struct
- [ ] Update `ValidateManifest()` in `manifest.go`
- [ ] Update `CreateDefaultManifest()` and template generator
- [ ] Unit tests for manifest validation

### Phase 2: Central MCP Proxy

- [ ] Implement plugin discovery in MCP server startup
- [ ] Implement gRPC proxy handler (for service plugins)
- [ ] Implement exec proxy handler (for helper plugins)
- [ ] Extend `tools/list` handler to include plugin tools
- [ ] Extend `tools/call` handler to route plugin calls
- [ ] Integration tests with mock plugin

### Phase 3: Standalone MCP Support

- [ ] Implement `plugin mcp-serve <name>` command
- [ ] Resolve plugin path and execute MCP server command
- [ ] Bridge stdio correctly
- [ ] Integration tests

### Phase 4: MCP Registration

- [ ] Extend `mcp configure` with `--with-plugins` flag
- [ ] Extend `mcp configure` with `--plugin <name>` flag
- [ ] Add MCP registration prompt to `plugin install`
- [ ] Add MCP unregistration prompt to `plugin uninstall`
- [ ] Implement `mcp remove --plugin <name>`
- [ ] Extend `mcp list` with plugin servers

### Phase 5: Testing and Documentation

- [ ] End-to-end test: plugin extend mode with Claude Code
- [ ] End-to-end test: plugin standalone mode with Claude Code
- [ ] Update plugin development documentation
- [ ] Update MCP documentation
- [ ] Update FEATURES_OVERVIEW.md

## Acceptance Criteria

- [ ] Plugin with `mcp_mode: "extend"` has its tools accessible via central `portunix mcp serve`
- [ ] Plugin with `mcp_mode: "standalone"` can be launched via `portunix plugin mcp-serve <name>`
- [ ] Plugin with `mcp_mode: "both"` supports both modes simultaneously
- [ ] `portunix mcp configure --with-plugins` registers central + all plugin MCP servers in Claude Code
- [ ] `portunix mcp configure --plugin <name>` registers a specific plugin's MCP server
- [ ] `portunix plugin install` prompts for MCP registration when plugin has `mcp_server`
- [ ] `portunix plugin uninstall` prompts for MCP unregistration
- [ ] `portunix mcp list` shows both central and plugin MCP servers
- [ ] Existing plugins without `mcp_mode` work unchanged (backward compatible)
- [ ] Tool name collisions are detected and prevented

## Related

- **ADR-038**: Distributed MCP Server Ecosystem for Plugins
- **#004**: MCP Server for AI Assistant Integration
- **#007**: Plugin System with gRPC Architecture
- **#132**: Text Extractor Plugin Integration (potential first consumer)
- **#141**: PTX-TRACE Universal Tracing Helper (potential standalone MCP)

---

**Status**: 📋 Open
**Priority**: High
**Type**: Architecture / Feature
**Labels**: mcp, plugin-system, ai-integration, architecture, distributed
**Assignee**: -
