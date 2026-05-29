---
title: "package"
description: "Package management and registry"
---

# package

Package management and registry

## Usage

```bash
portunix package [options] [arguments]
```

## Full Help

```
Package management and registry operations

Usage: portunix package <subcommand> [options]

Available subcommands:
  list     List all available packages
  search   Search for packages by name or description
  info     Show detailed information about a package
  detect   Detect installed AI assistants
  update   Reinstall a package's latest available version

Options:
  -h, --help   Show this help message

Examples:
  portunix package list
  portunix package list --category development/languages
  portunix package search python
  portunix package info nodejs
  portunix package detect
  portunix package detect --json

```

## Examples

```bash
  portunix package list
  portunix package list --category development/languages
  portunix package search python
  portunix package info nodejs
  portunix package detect
  portunix package detect --json

```
