---
title: "install"
description: "Install packages and tools"
---

# install

Install packages and tools

## Usage

```bash
portunix install [options] [arguments]
```

## Full Help

```
Install software packages

Usage: portunix install <package> [options]

Options:
  --variant=<variant>  Select package variant (e.g., --variant=21 for Java 21)
  --dry-run            Preview installation without executing
  --force              Force reinstallation even if already installed
  -h, --help           Show this help message

Examples:
  portunix install python
  portunix install java --variant=21
  portunix install nodejs --dry-run

Use 'portunix package list' to see available packages
Use 'portunix package info <package>' for detailed package information

```

## Examples

```bash
  portunix install python
  portunix install java --variant=21
  portunix install nodejs --dry-run

```
