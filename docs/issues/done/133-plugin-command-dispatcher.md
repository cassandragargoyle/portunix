# Issue #133: Plugin Command Dispatcher

## Overview
**Title**: Plugin Command Dispatcher
**Status**: ✅ Implemented
**Closed**: 2026-11-20
**Priority**: High
**Type**: Feature
**Labels**: plugin, cli, dispatcher

## Problem Description

Portunix plugin system currently supports plugin installation, enabling, and lifecycle management, but lacks the ability to invoke plugin commands directly from the CLI.

**Current behavior:**
```bash
portunix text-extractor extract document.pdf
# Error: unknown command "text-extractor" for "portunix"
```

**Expected behavior:**
```bash
portunix text-extractor extract document.pdf
# Invokes the text-extractor plugin's extract command
```

## Requirements

### Functional Requirements

1. **Dynamic Command Registration**
   - Enabled plugins should register their commands as portunix subcommands
   - Commands should be available via `portunix <plugin-name> <command> [args]`

2. **Helper Plugin Invocation**
   - For helper plugins (mode: "helper"), invoke the binary/JAR on-demand
   - Pass arguments to the plugin binary
   - Return output to the user

3. **Service Plugin Invocation**
   - For service plugins (mode: "service"), communicate via gRPC
   - Plugin must be running (started) to accept commands

4. **Runtime Support**
   - Native binaries: direct execution
   - Java plugins: `java -jar <path> [args]`
   - Python plugins: `python3 <path> [args]`

### Technical Requirements

1. **Command Discovery**
   - Read enabled plugins from registry
   - Parse plugin.json for available commands
   - Register commands dynamically with Cobra

2. **Argument Passing**
   - Pass all arguments after plugin name to the plugin binary
   - Support flags and positional arguments

3. **Output Handling**
   - Stream stdout/stderr from plugin to user
   - Return exit code from plugin

## Implementation Plan

### Phase 1: Basic Dispatcher
1. Add plugin command discovery in `cmd/root.go` or `cmd/plugin.go`
2. Register enabled plugin names as dynamic subcommands
3. Implement basic execution for helper plugins

### Phase 2: Runtime Support
1. Implement Java runtime invocation (java -jar)
2. Implement Python runtime invocation (python3)
3. Handle JVM args from plugin.json

### Phase 3: Service Plugin Support

> Moved to Issue #258 — not implemented in this issue.

1. For running service plugins, use gRPC Execute method
2. Handle plugin not running errors gracefully

## Example Usage

```bash
# Helper plugin (invokes JAR directly)
portunix text-extractor extract text document.pdf
portunix text-extractor extract tables report.xlsx --format csv

# Service plugin (uses gRPC if running)
portunix agile board show
portunix agile task add "New task"
```

## Acceptance Criteria

- [x] `portunix <plugin-name> --help` shows plugin commands (forwarded to the plugin)
- [x] `portunix <plugin-name> <command>` executes plugin command
- [x] Java plugins are invoked with correct JVM args
- [x] Python plugins are invoked correctly
- [x] Native plugins are invoked correctly
- [x] Error handling for disabled/missing plugins
- [x] Output is streamed to user in real-time
- [ ] Service plugins via gRPC `Execute` — moved to Issue #258

## Related Issues

- Issue #132: Text Extractor Plugin Integration (implemented plugin mode)
- Issue #221: Plugin run command argument forwarding
- Issue #258: Service plugin command execution via gRPC (Phase 3 follow-up)

## Notes

- Plugin commands should only be available for enabled plugins
- Consider caching command discovery for performance
- May need to handle command name conflicts with core commands
