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
  --path=<path>        Target installation path (for project generators like docusaurus)
  --dry-run            Preview installation without executing
  --force              Force reinstallation even if already installed
  --db-host=<host>     Override container DB HOST env (container variants that read it)
  --db-port=<port>     Override container DB PORT env
  --db-user=<user>     Override container DB USER env
  --db-password=<pwd>  Override container DB PASSWORD env
  -h, --help           Show this help message

Examples:
  portunix install python
  portunix install java --variant=21
  portunix install docusaurus --path ./my-docs
  portunix install nodejs --dry-run
  portunix install odoo --variant=container-external-db --db-host=my-pg

Use 'portunix package list' to see available packages
Use 'portunix package info <package>' for detailed package information

```

## Examples

```bash
  portunix install python
  portunix install java --variant=21
  portunix install docusaurus --path ./my-docs
  portunix install nodejs --dry-run
  portunix install odoo --variant=container-external-db --db-host=my-pg

```
