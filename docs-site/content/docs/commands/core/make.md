---
title: "make"
description: "Cross-platform Makefile utilities"
---

# make

Cross-platform Makefile utilities

## Usage

```bash
portunix make [options] [arguments]
```

## Full Help

```
Usage: portunix make <command> [arguments]

Cross-platform Makefile utilities

File Operations:
  copy <src> <dst>         - Copy files/directories with wildcard support
  mkdir <path>             - Create directory tree (like mkdir -p)
  rm <path>                - Remove files/directories recursively
  exists <path>            - Check path existence (exit code 0/1)
  ls [options] [path]      - List directory contents (cross-platform)

Build Metadata:
  version                  - Git version tag (git describe)
  commit                   - Short git commit hash
  timestamp                - UTC timestamp in ISO 8601 format

Build Tools:
  gobuild [VAR=val]... cmd - Cross-platform Go compilation with env vars

Utilities:
  checksum <dir> [output]  - Generate SHA256 checksums
  chmod <mode> <file>      - Set file permissions (no-op on Windows)
  json <k=v>...            - Generate JSON from key-value pairs
  env                      - Export platform variables for Makefile

Examples:
  portunix make mkdir dist/bin
  portunix make copy src/*.go dist/
  portunix make rm build/
  portunix make ls -lah
  portunix make version
  portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .
  portunix make json version=1.0.0 platform=linux-x64

```

## Examples

```bash
  portunix make mkdir dist/bin
  portunix make copy src/*.go dist/
  portunix make rm build/
  portunix make ls -lah
  portunix make version
  portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .
  portunix make json version=1.0.0 platform=linux-x64

```
