---
title: "plugin"
description: "Manage plugins"
---

# plugin

Manage plugins

## Usage

```bash
portunix plugin [options] [arguments]
```

## Full Help

```
Usage:
  portunix plugin [command]

Available Commands:
  create         Create a new plugin template
  disable        Disable a plugin
  enable         Enable a plugin
  health         Check plugin health
  info           Show plugin information
  install        Install a plugin
  install-github Install a plugin from GitHub
  list           List installed plugins
  list-available List available plugins from the official repository
  start          Start a plugin
  stop           Stop a plugin
  uninstall      Uninstall a plugin
  validate       Validate a plugin

Flags:
  -h, --help   help for plugin

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix plugin [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [create](#create) | Create a new plugin template |
| [disable](#disable) | Disable a plugin |
| [enable](#enable) | Enable a plugin |
| [health](#health) | Check plugin health |
| [info](#info) | Show plugin information |
| [install](#install) | Install a plugin |
| [install-github](#install-github) | Install a plugin from GitHub |
| [list](#list) | List installed plugins |
| [list-available](#list-available) | List available plugins from the official repository |
| [start](#start) | Start a plugin |
| [stop](#stop) | Stop a plugin |
| [uninstall](#uninstall) | Uninstall a plugin |
| [validate](#validate) | Validate a plugin |

### create

Create a new plugin template

```
Usage:
  portunix plugin create <plugin-name> [flags]

Flags:
  -a, --author string        Plugin author name
  -d, --description string   Plugin description
  -h, --help                 help for create
  -o, --output string        Output directory for plugin template (default ".")

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### disable

Disable a plugin

```
Usage:
  portunix plugin disable <plugin-name> [flags]

Flags:
  -h, --help   help for disable

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### enable

Enable a plugin

```
Usage:
  portunix plugin enable <plugin-name> [flags]

Flags:
  -h, --help   help for enable

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### health

Check plugin health

```
Usage:
  portunix plugin health <plugin-name> [flags]

Flags:
  -h, --help   help for health

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### info

Show plugin information

```
Usage:
  portunix plugin info <plugin-name> [flags]

Flags:
  -h, --help   help for info

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### install

Install a plugin

```
Usage:
  portunix plugin install <plugin-path> [flags]

Flags:
  -h, --help   help for install

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### install-github

Install a plugin from GitHub

```
Usage:
  portunix plugin install-github <plugin-name> [flags]

Flags:
  -f, --force            Force install even if plugin already exists
  -h, --help             help for install-github
  -v, --version string   Specific version to install (default: latest)

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### list

List installed plugins

```
Usage:
  portunix plugin list [flags]

Flags:
  -a, --all             Show all plugins (including disabled)
  -h, --help            help for list
  -o, --output string   Output format: table, json, yaml (default "table")

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### list-available

List available plugins from the official repository

```
Usage:
  portunix plugin list-available [flags]

Flags:
  -h, --help   help for list-available

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### start

Start a plugin

```
Usage:
  portunix plugin start <plugin-name> [flags]

Flags:
  -h, --help   help for start

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### stop

Stop a plugin

```
Usage:
  portunix plugin stop <plugin-name> [flags]

Flags:
  -h, --help   help for stop

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### uninstall

Uninstall a plugin

```
Usage:
  portunix plugin uninstall <plugin-name> [flags]

Flags:
  -f, --force   Force uninstall without confirmation
  -h, --help    help for uninstall

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### validate

Validate a plugin

```
Usage:
  portunix plugin validate <plugin-path> [flags]

Flags:
  -h, --help   help for validate

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

