# Issue #221: Plugin Run Command Argument Forwarding

> **Renumbered:** formerly internal issue #133. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

## Summary

Add `portunix plugin run <plugin-name> [args...]` command that forwards all arguments to the plugin executable, making it equivalent to `portunix <plugin-name> [args...]`.

## Background

AI assistants (like Claude) discover available plugins by calling `portunix plugin list`. After discovering a plugin, the natural next step is to explore its capabilities. Currently, the assistant might try:

```bash
portunix plugin run text-extractor --help
```

This should be equivalent to:

```bash
portunix text-extractor --help
```

## Current Behavior

- `portunix plugin list` - works, shows available plugins
- `portunix plugin run <plugin> [args]` - command may not exist or may not forward arguments correctly
- `portunix <plugin> [args]` - works if plugin is registered as top-level command

## Expected Behavior

```bash
# These should produce identical output:
portunix plugin run text-extractor --help
portunix text-extractor --help

# All arguments after plugin name should be forwarded:
portunix plugin run text-extractor extract --file document.pdf
portunix text-extractor extract --file document.pdf
```

## Use Cases

### 1. AI Assistant Discovery Flow
```bash
# Step 1: Discover plugins
$ portunix plugin list
NAME             VERSION   STATUS   DESCRIPTION
text-extractor   1.0.0     ready    Extract text from documents

# Step 2: Explore plugin capabilities (natural next step)
$ portunix plugin run text-extractor --help
Usage: text-extractor [command] [options]
...
```

### 2. Explicit Plugin Invocation
When users want to be explicit about running a plugin (vs a built-in command):
```bash
portunix plugin run my-plugin some-command --flag value
```

## Implementation Notes

1. Add `run` subcommand to `plugin` command
2. Parse plugin name as first argument
3. Forward all remaining arguments to plugin executable
4. Handle plugin not found / not enabled errors gracefully
5. Support both helper-mode and service-mode plugins

## Acceptance Criteria

- [x] `portunix plugin run <plugin> --help` shows plugin help
- [x] `portunix plugin run <plugin> [args]` forwards all arguments to plugin
- [x] Error message when plugin not found or not enabled
- [x] Works with helper plugins (executable)
- [x] Works with service plugins (gRPC) if applicable

## Priority

Medium - Improves AI assistant integration and discoverability

## Labels

enhancement, plugin-system, cli, ai-integration, user-experience
