# ptx-make

Portunix Cross-Platform Makefile Utility - dedicated helper binary for cross-platform build operations.

## Overview

`ptx-make` is a helper binary that provides cross-platform replacements for common shell utilities used in Makefiles.
It eliminates the need for platform-specific conditionals by providing identical behavior on Windows, Linux, and macOS.

Typically invoked via `portunix make <command>` dispatcher or directly as `ptx-make <command>`.

## Commands

### File Operations

| Command | Description | Example |
| ------- | ----------- | ------- |
| `copy` | Copy files/directories with wildcard support | `ptx-make copy src/*.go dest/` |
| `mkdir` | Create directory tree (equivalent to `mkdir -p`) | `ptx-make mkdir build/output` |
| `rm` | Remove files/directories recursively with glob support | `ptx-make rm build/*.o` |
| `exists` | Check if path exists (exit code 0/1) | `ptx-make exists config.json` |
| `ls` | List directory contents with flags (-l, -a, -R, -t, -S) | `ptx-make ls -la src/` |
| `chmod` | Set file permissions (symbolic or octal) | `ptx-make chmod +x script.sh` |

### Build Metadata

| Command | Description | Example |
| ------- | ----------- | ------- |
| `version` | Get git version tag (`git describe --tags --always --dirty`) | `ptx-make version` |
| `commit` | Get short git commit hash | `ptx-make commit` |
| `timestamp` | Get current UTC timestamp (RFC3339) | `ptx-make timestamp` |

### Build Tools

| Command | Description | Example |
| ------- | ----------- | ------- |
| `gobuild` | Cross-platform Go build with env variables | `ptx-make gobuild GOOS=linux go build -o app` |
| `checksum` | Generate SHA256 checksums for files in a directory | `ptx-make checksum dist/ -o checksums.txt` |

### Utilities

| Command | Description | Example |
| ------- | ----------- | ------- |
| `env` | Export platform variables (OS, ARCH, EXE, SLASH, PATHSEP) | `ptx-make env` |
| `json` | Generate JSON from key-value pairs | `ptx-make json name=portunix version=2.1.0` |

## Usage in Makefile

```makefile
# Cross-platform directory creation
setup:
	ptx-make mkdir build/output

# Cross-platform copy with wildcards
install:
	ptx-make copy dist/*.tar.gz /usr/local/share/

# Cross-platform Go build (works on Windows too)
build:
	ptx-make gobuild GOOS=linux GOARCH=amd64 go build -o build/app

# Generate build metadata
metadata:
	ptx-make json \
		version=$(shell ptx-make version) \
		commit=$(shell ptx-make commit) \
		timestamp=$(shell ptx-make timestamp) > build/metadata.json

# SHA256 checksums
checksums:
	ptx-make checksum dist/ -o dist/checksums.txt
```

## Cross-Platform Behavior

| Feature | Linux/macOS | Windows |
| ------- | ----------- | ------- |
| `chmod` | Sets Unix permissions | No-op (no permission model) |
| `env` EXE | `` (empty) | `.exe` |
| `env` SLASH | `/` | `\` |
| `env` PATHSEP | `:` | `;` |
| `ls` | Uses native `ls` | Go emulation |
| `gobuild` | Direct execution | Parses VAR=value syntax |
