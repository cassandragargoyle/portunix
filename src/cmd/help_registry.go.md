# help_registry.go

Documentation generated from code — `src/cmd/help_registry.go`.

## Overview

This file is the **single source of truth for Portunix CLI help output**. It
declares every core command together with its metadata (brief text, full
description, category, parameters, examples, sub-commands) in one central
registry, then renders that registry into three help formats:

| Format | Function | Audience |
| ------ | -------- | -------- |
| Basic help | `GenerateBasicHelp()` | End users — essential commands only (`--help`) |
| Expert help | `GenerateExpertHelp()` | Power users — full reference grouped by category (`--help-expert`) |
| AI help | `GenerateAIHelp()` | AI/LLM tools — machine-readable JSON (`--help-ai`) |

Installed plugins are discovered at runtime and merged into all outputs, so the
help stays accurate without hard-coding plugin entries.

Package: `cmd`.
Import of note: `portunix.ai/app/plugins/manager` (plugin registry access).

## Data Model

### `CommandInfo`

The metadata record for a command (or sub-command). It is annotated with JSON
tags so it can be serialized directly for the AI help output.

| Field | Type | Purpose |
| ----- | ---- | ------- |
| `Name` | `string` | Command name (e.g. `install`) |
| `Brief` | `string` | Short one-line text shown in basic help |
| `Description` | `string` | Long text shown in expert help |
| `Category` | `string` | Grouping key for expert help (`core`, `container`, …) |
| `Parameters` | `[]ParameterInfo` | Command flags/arguments (optional) |
| `Examples` | `[]string` | Usage examples (optional) |
| `SubCommands` | `[]CommandInfo` | Nested commands (optional, recursive) |

### `ParameterInfo`

Describes a single command parameter.

| Field | Type | Purpose |
| ----- | ---- | ------- |
| `Name` | `string` | Parameter name |
| `Type` | `string` | Value type (`string`, `boolean`, …) |
| `Required` | `bool` | Whether the parameter is mandatory |
| `Description` | `string` | Human-readable explanation |
| `Default` | `string` | Default value (optional) |
| `Choices` | `[]string` | Allowed values (optional) |

## Registry

### `CommandRegistry` (package-level `var`)

A static slice of `CommandInfo` holding all built-in commands. This is the
authoritative catalog — editing this slice updates every help format at once.
It is split loosely into two blocks:

- **Core / commonly used commands** — `install`, `update`, `docker`, `plugin`,
  `mcp`, `container`, `virt`, `system`, `make`, `package`, `pft`, `specpm`,
  `playbook`, `python`, `aiops`, `credential`.
- **Additional commands surfaced at expert level** — `podman`, `sandbox`,
  `completion`, `cache`, `config`, `datastore`, `guid` (the latter two of which
  demonstrate `SubCommands`).

## Functions

### `GetBasicCommands() []CommandInfo`

Filters `CommandRegistry` down to a curated `essentials` allow-list of command
names and returns them **in registry order** (the allow-list controls
membership, not ordering). This is what the basic help renders so the first-time
user is not overwhelmed by the full command set.

### `GetInstalledPlugins() []CommandInfo`

Discovers installed plugins at runtime and adapts them into `CommandInfo`
records so they can be rendered alongside built-in commands.

Flow:

1. Resolve the user's home directory; return `nil` on failure.
2. Look for `~/.portunix/plugins/registry.json`; return `nil` if it is missing.
3. Open the registry via `manager.NewRegistry`; return `nil` on error.
4. List plugins; for each, fetch full data with `GetPluginRegistryData`.
5. Build a `CommandInfo` per plugin, embedding the enabled/disabled status into
   the `Brief` text (`"<description> [enabled|disabled]"`) and tagging the
   `Category` as `"plugin"`.

The function is deliberately **fail-soft**: any error at any step yields `nil`
(no plugins) rather than breaking help output.

### `GenerateBasicHelp() string`

Builds the `--help` text with `strings.Builder`:

1. Header + usage line.
2. Fetch core commands (`GetBasicCommands`) and installed plugins
   (`GetInstalledPlugins`).
3. Compute `maxLen` — the longest name across **both** commands and plugins —
   for column alignment.
4. Print each command as an aligned `name  brief` row.
5. If any plugins exist, print an `Installed plugins:` section using the same
   alignment.
6. Append the mandatory **Help levels** section and footer hints.

### `GenerateExpertHelp() string`

Builds the `--help-expert` reference:

1. Header + extended description banner.
2. Group all `CommandRegistry` entries into a `map[category][]CommandInfo`
   (empty category becomes `"other"`).
3. Merge installed plugins in under the `"plugin"` category.
4. Render categories in a fixed `categoryOrder` so output is deterministic:
   `core → development → container → virtualization → integration → utility → plugin → other`.
5. For each command print name, description, parameters (with required flag,
   default, choices) and examples.
6. Append static **Environment variables**, **Configuration files**, and
   **Help levels** sections plus a project link.

### `GenerateAIHelp() (string, error)`

Serializes the entire registry as indented JSON for AI/LLM consumption. Defines
an anonymous `AIHelpOutput` struct (tool name, version, description, full
`CommandRegistry`, and environment variables) and marshals it with
`json.MarshalIndent`. Returns the JSON string, or the marshaling error.

### `GetCommandInfo(commandName string) *CommandInfo`

Linear lookup over `CommandRegistry` returning a pointer to the matching
top-level command, or `nil` if not found. (Matches top-level names only, not
sub-commands.)

## Extending

- **Add a command**: append a `CommandInfo` to `CommandRegistry`. To make it
  appear in basic help, also add its name to the `essentials` list in
  `GetBasicCommands`. Assign a `Category` that exists in `categoryOrder` so it
  renders in expert help.
- **Add a category**: extend `categoryOrder` in `GenerateExpertHelp`; otherwise
  commands in that category are silently skipped by the expert renderer.
- Plugins require no code changes — they are picked up from the plugin registry
  at runtime.

## Notes

- Alignment across commands and plugins is intentional so both sections share a
  single column width in basic help.
- The three generators are pure with respect to the static registry; only plugin
  discovery touches the filesystem, and it degrades gracefully to no plugins.
