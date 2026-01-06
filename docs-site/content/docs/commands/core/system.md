---
title: "system"
description: "System information"
---

# system

System information

## Usage

```bash
portunix system [options] [arguments]
```

## Full Help

```
Usage:
  portunix system [command]

Available Commands:
  check       Check specific system conditions
  dispatcher  Display dispatcher and helper binary information
  info        Display detailed system information

Flags:
  -h, --help   help for system

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix system [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [check](#check) | Check specific system conditions |
| [dispatcher](#dispatcher) | Display dispatcher and helper binary information |
| [info](#info) | Display detailed system information |

### check

Check specific system conditions

```
Usage:
  portunix system check [condition] [flags]

Flags:
  -h, --help   help for check

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### dispatcher

Display dispatcher and helper binary information

```
Usage:
  portunix system dispatcher [flags]

Flags:
  -h, --help   help for dispatcher
  -j, --json   Output as JSON

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### info

Display detailed system information

```
Usage:
  portunix system info [flags]

Flags:
      --cpuprofile string   Write CPU profile to file
  -h, --help                help for info
  -j, --json                Output as JSON
      --memprofile string   Write memory profile to file
  -s, --short               Short output (OS Version Variant)
  -t, --time                Show execution time
      --trace string        Write execution trace to file

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

