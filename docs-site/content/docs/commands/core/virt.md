---
title: "virt"
description: "Virtual machine management"
---

# virt

Virtual machine management

## Usage

```bash
portunix virt [options] [arguments]
```

## Full Help

```
Usage:
  portunix virt [flags]
  portunix virt [command]

Available Commands:
  check       Check virtualization requirements
  copy        Copy files between host and virtual machine
  create      Create a new virtual machine
  delete      Delete a virtual machine
  exec        Execute a command in a virtual machine
  info        Show detailed information about a virtual machine
  iso         Manage ISO files for virtual machines
  list        List all virtual machines
  restart     Restart a virtual machine
  resume      Resume a suspended virtual machine
  snapshot    Manage virtual machine snapshots
  ssh         SSH into a virtual machine with smart boot waiting
  start       Start a virtual machine
  status      Show the status of a virtual machine
  stop        Stop a virtual machine
  suspend     Suspend a virtual machine
  template    Manage virtual machine templates

Flags:
  -h, --help   help for virt

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix virt [command] --help" for more information about a command.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [check](#check) | Check virtualization requirements |
| [copy](#copy) | Copy files between host and virtual machine |
| [create](#create) | Create a new virtual machine |
| [delete](#delete) | Delete a virtual machine |
| [exec](#exec) | Execute a command in a virtual machine |
| [info](#info) | Show detailed information about a virtual machine |
| [iso](#iso) | Manage ISO files for virtual machines |
| [list](#list) | List all virtual machines |
| [restart](#restart) | Restart a virtual machine |
| [resume](#resume) | Resume a suspended virtual machine |
| [snapshot](#snapshot) | Manage virtual machine snapshots |
| [ssh](#ssh) | SSH into a virtual machine with smart boot waiting |
| [start](#start) | Start a virtual machine |
| [status](#status) | Show the status of a virtual machine |
| [stop](#stop) | Stop a virtual machine |
| [suspend](#suspend) | Suspend a virtual machine |
| [template](#template) | Manage virtual machine templates |

### check

Check virtualization requirements

```
Usage:
  portunix virt check [flags]

Flags:
      --blacklist-kvm   Blacklist KVM permanently (use with --fix)
      --dry-run         Show what would be done without making changes
      --fix             Interactively fix detected conflicts
      --fix-libvirt     Fix libvirt daemon issues
  -h, --help            help for check
      --unload-kvm      Unload KVM modules (use with --fix)
      --use-kvm         Switch to KVM (use with --fix)

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### copy

Copy files between host and virtual machine

```
Usage:
  portunix virt copy [source] [destination] [flags]

Flags:
  -h, --help   help for copy

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### create

Create a new virtual machine

```
Usage:
  portunix virt create [vm-name] [flags]

Flags:
      --cpus int          Number of CPUs
      --disk string       Disk size (e.g., 40G, 50000M)
      --enable-ssh        Enable SSH access
  -h, --help              help for create
      --iso string        ISO file to mount
      --os-type string    OS type hint for optimization
      --ram string        RAM allocation (e.g., 4G, 2048M)
      --template string   VM template to use (ubuntu-24.04, windows11, etc.)

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### delete

Delete a virtual machine

```
Usage:
  portunix virt delete [vm-name] [flags]

Flags:
      --force       Skip confirmation prompt
  -h, --help        help for delete
      --keep-disk   Keep disk files when deleting

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### exec

Execute a command in a virtual machine

```
Usage:
  portunix virt exec [vm-name] [command] [flags]

Flags:
  -h, --help   help for exec

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### info

Show detailed information about a virtual machine

```
Usage:
  portunix virt info [vm-name] [flags]

Flags:
  -h, --help   help for info

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### iso

Manage ISO files for virtual machines

```
Usage:
  portunix virt iso [command]

Available Commands:
  clean       Clean up old or unused ISO files
  download    Download an ISO file
  info        Show detailed information about an ISO
  list        List available and downloaded ISOs
  verify      Verify the checksum of a downloaded ISO

Flags:
  -h, --help   help for iso

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix virt iso [command] --help" for more information about a command.
```

### list

List all virtual machines

```
Usage:
  portunix virt list [flags]

Aliases:
  list, ls

Flags:
  -h, --help   help for list

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### restart

Restart a virtual machine

```
Usage:
  portunix virt restart [vm-name] [flags]

Flags:
  -h, --help   help for restart

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### resume

Resume a suspended virtual machine

```
Usage:
  portunix virt resume [vm-name] [flags]

Flags:
  -h, --help   help for resume

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### snapshot

Manage virtual machine snapshots

```
Usage:
  portunix virt snapshot [command]

Available Commands:
  create      Create a new snapshot of a virtual machine
  delete      Delete a snapshot
  list        List all snapshots for a virtual machine
  revert      Revert a virtual machine to a snapshot

Flags:
  -h, --help   help for snapshot

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix virt snapshot [command] --help" for more information about a command.
```

### ssh

SSH into a virtual machine with smart boot waiting

```
Usage:
  portunix virt ssh [vm-name] [command] [flags]

Flags:
      --check                 Just check if SSH is ready (don't connect)
  -h, --help                  help for ssh
      --no-wait               Don't wait for SSH availability
      --start                 Automatically start/resume VM if needed
      --wait-timeout string   Maximum time to wait for SSH (e.g., 60s, 2m) (default "30s")

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### start

Start a virtual machine

```
Usage:
  portunix virt start [vm-name] [flags]

Flags:
      --force   Force restart if already running
  -h, --help    help for start

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### status

Show the status of a virtual machine

```
Usage:
  portunix virt status [vm-name] [flags]

Flags:
  -h, --help     help for status
      --simple   Output only the status value

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### stop

Stop a virtual machine

```
Usage:
  portunix virt stop [vm-name] [flags]

Flags:
      --force   Force immediate shutdown
  -h, --help    help for stop

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### suspend

Suspend a virtual machine

```
Usage:
  portunix virt suspend [vm-name] [flags]

Flags:
  -h, --help   help for suspend

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples
```

### template

Manage virtual machine templates

```
Usage:
  portunix virt template [command]

Available Commands:
  list        List all available VM templates
  show        Show detailed information about a template

Flags:
  -h, --help   help for template

Global Flags:
      --help-ai       Show machine-readable help in JSON format
      --help-expert   Show extended help with all options and examples

Use "portunix virt template [command] --help" for more information about a command.
```

