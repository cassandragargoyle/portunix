# Issue #220: MCP Server Wizard — Advanced Features Completion

> **Renumbered:** formerly internal issue #034. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Type**: Enhancement
**Priority**: Medium
**Status**: ✅ Implemented
**Created**: 2025-09-11
**Updated**: 2026-05-12
**Closed**: 2026-05-12
**Labels**: mcp, ai-integration, wizard, user-experience, cross-platform
**Related Issues**:

- [#004](004-mcp-server-ai-integration.md) — MCP Server for AI Assistant Integration (✅ Implemented)
- [#014](014-wizard-framework.md) — Wizard Framework for Interactive CLI (non-blocking)
- [#035](../035-ai-assistant-installation-support.md) — AI Assistant Installation Support (blocking for full install replacement)

## Summary

Complete the remaining advanced features of the MCP server setup wizard implemented
in `ptx-mcp`. The original scope of this issue covered creating a multi-assistant
interactive wizard, smart configuration management, security profiles and lifecycle
commands. The bulk of that work is already implemented in
`src/helpers/ptx-mcp/`. This issue is now narrowed down to the remaining
advanced functionality.

## Implementation Status (as of 2026-05-08)

The following capabilities are already shipped in `ptx-mcp`:

- `portunix mcp init` — interactive wizard with assistant detection, security profile
  selection and `--force` mode (`init.go`)
- `portunix mcp config` — smart configuration management with `--edit`, `--force`,
  `--add`, `--json` (`config.go`)
- `portunix mcp configure` / `reconfigure` / `remove` — Claude Code integration via
  `claude mcp add` / `remove`
- `portunix mcp serve` — stdio / tcp / unix communication modes
- `portunix mcp start` / `stop` / `status` — server lifecycle with PID tracking
- `portunix mcp test` — placeholder command (no actual handshake yet)
- Detection of Claude Code, Claude Desktop, Gemini CLI (`common.go`)
- Security profiles: `development`, `standard`, `restricted`
- Persistent configuration in `~/.portunix/mcp-server.json`
  (`MCPConfiguration` schema)

## Remaining Scope

### 1. Real Connection Testing

`testAssistantConnection()` in `src/helpers/ptx-mcp/init.go:459` is currently a
stub. Replace it with a real MCP handshake:

- `portunix mcp test [--assistant <name>]` must perform a JSON-RPC `initialize`
  request against the configured server (stdio for CLI assistants, network
  endpoint for remote)
- Report success / failure with the protocol version returned by the server and
  a human-readable diagnostic line per assistant
- Hook the same routine into the final step of `mcp init` so the wizard verifies
  the configuration before exiting

### 2. Advanced Claude Code Options for `mcp init`

Extend the non-interactive mode and the wizard prompts with Claude Code CLI
options that map directly to `claude mcp add`:

- `--scope local|project|user` — passes through to `claude mcp add --scope`;
  default remains `local`
- `--env KEY=VALUE,…` — environment variables propagated to the spawned MCP
  server process
- `--timeout <duration>` — startup timeout (e.g. `30s`, `2m`); persist in
  `MCPConfiguration` and pass to `claude mcp add` when supported
- `--from-json <file>` — import a complete configuration document, validate it
  against the `MCPConfiguration` schema, and apply it (skipping interactive
  prompts)

The interactive wizard should expose the same options behind plain prompts when
Claude Code is the selected assistant.

### 3. Auto-configure Claude Desktop

`configureClaudeDesktop()` (`init.go:434`) currently prints manual instructions.
Replace it with a writer that:

- Resolves the platform-specific path via `getClaudeDesktopConfigPath()`
- Creates the directory and `mcp_servers.json` if missing
- Merges the `portunix` entry into the existing JSON without clobbering other
  servers (load → modify → write back, preserving formatting where reasonable)
- Verifies the result by re-reading the file

Cross-platform paths to support:

- macOS: `~/Library/Application Support/Claude/mcp_servers.json`
- Windows: `%APPDATA%/Claude/mcp_servers.json`
- Linux: `~/.config/claude/mcp_servers.json`

### 4. Network Configuration for Remote Servers

The wizard branch for `serverType == "remote"` currently asks only for port and
protocol via free-form input (`init.go:241-255`). Extend it with:

- Protocol selection: `http`, `https`, `ws`, `wss` (validated, not free-form)
- Bind address: `localhost`, `0.0.0.0`, or a specific IP — with validation
- TLS configuration when `https` / `wss` is selected:
  - Path to certificate (`--tls-cert`) and key (`--tls-key`)
  - Optional self-signed generation helper for local development
- Persist the new fields in `MCPConfiguration` (`bind_address`, `tls_cert`,
  `tls_key`)
- Reuse `findAvailablePorts()` (`common.go:188`) to suggest free ports when the
  default is taken

### 5. Resolve Dependency on #035

`installClaudeCode()` (`init.go:350`) currently falls back to `npm install -g`
and then a `curl | sh` script. Once #035 ships:

- Replace the body of `installClaudeCode()` with a call to
  `portunix install claude-code`
- Add `portunix install claude-desktop` and `portunix install gemini-cli` calls
  to the corresponding wizard branches when those packages become available
- Remove the manual instructions in `configureClaudeDesktop()` once an
  installable package exists

Until #035 lands, the existing fallbacks remain in place.

## Out of Scope

- **Command restructuring under `mcp serve` subtree** (`mcp serve init`,
  `mcp serve config`, …) — the original proposal called for this, but the
  current flat structure is already in production use and integrated with
  Claude Code via `claude mcp add`. Renaming would be a breaking change for
  existing users without offsetting benefit.
- **Wizard Framework integration** — tracked under #014 and applied to all
  helpers at once, not per-helper.
- **Full Gemini CLI MCP integration** — remains experimental. The wizard
  detects Gemini CLI but only writes a placeholder configuration. A complete
  implementation requires upstream documentation that does not yet exist.

## Success Criteria

- [ ] `portunix mcp test` performs a real JSON-RPC `initialize` handshake and
      reports protocol version per assistant
- [ ] `portunix mcp init --assistant claude-code --scope project` works
      end-to-end and writes the correct `claude mcp add` configuration
- [ ] `portunix mcp init --assistant claude-code --env KEY=VAL` propagates env
      vars to the MCP server process
- [ ] `portunix mcp init --assistant claude-code --timeout 60s` is honored
- [ ] `portunix mcp init --from-json <file>` imports a validated configuration
      without interactive prompts
- [ ] `portunix mcp init --assistant claude-desktop` writes
      `mcp_servers.json` automatically on macOS / Windows / Linux without
      clobbering pre-existing entries
- [ ] Remote-server wizard offers validated protocol / bind address / TLS
      options and persists them in `MCPConfiguration`
- [ ] After #035 lands, `installClaudeCode()` and friends call
      `portunix install …` instead of `npm` / `curl`

## Technical Notes

- All changes live inside `src/helpers/ptx-mcp/`; no main `portunix` binary
  changes required (dispatcher already routes `mcp` to `ptx-mcp`)
- `MCPConfiguration` (`common.go:202`) must be extended with new fields in a
  backward-compatible way — existing `~/.portunix/mcp-server.json` files
  written by current users must continue to load
- Real handshake tests should use the same JSON-RPC client logic as
  `portunix mcp serve` to avoid duplicating protocol code

## Original Proposal

The original scope of this issue (multi-assistant wizard, smart config
management, lifecycle commands, security profiles, command restructuring
under `mcp serve`) is preserved in Git history. See the previous revision
of this file via `git log --follow -p -- docs/issues/done/220-mcp-server-installation-wizard.md`.
