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
ptx-virt is a helper binary for Portunix that handles all virtualization operations.
It provides unified interface for VirtualBox, QEMU/KVM, VMware, and Hyper-V management.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.

Supported virtualization backends:
- VirtualBox (cross-platform)
- QEMU/KVM (Linux)
- VMware (cross-platform)
- Hyper-V (Windows)

Usage:
  ptx-virt [flags]

Flags:
      --description     Show description
  -h, --help            help for ptx-virt
      --list-commands   List available commands
  -v, --version         Show version

```

