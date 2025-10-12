# Portunix Install Command

## Quick Start

The `install` command is one of the most frequently used commands in Portunix. It enables cross-platform installation of development tools, programming languages, and packages.

### Simplest Usage
```bash
portunix install nodejs
```

This command installs Node.js with default settings for your platform.

### Basic Syntax
```bash
portunix install <package-name> [options]
```

### Most Common Packages
- `python` - Python programming language
- `nodejs` - Node.js JavaScript runtime
- `java` - Java Development Kit (OpenJDK)
- `go` - Go programming language
- `vscode` - Visual Studio Code editor
- `git` - Git version control system
- `docker` - Docker container platform

## Intermediate Usage

### Package Variants

Many packages support different installation variants:

```bash
# Install Python with full toolset
portunix install python --variant full

# Minimal Python installation
portunix install python --variant minimal

# Default recommended configuration
portunix install python --variant default
```

### Dry Run - Preview Installation

Before actual installation, you can use `--dry-run` to preview what will be installed:

```bash
portunix install nodejs --dry-run
```

Output shows:
- Which package will be installed
- Installation method to be used
- Dependencies to be installed
- Installation location

### Force Reinstallation

To reinstall a package even if already installed:

```bash
portunix install vscode --force
```

### Installation Profiles

Portunix supports installation profiles for quick development environment setup:

```bash
# Default developer profile (Python + Java 17 + VS Code)
portunix install default

# Minimal profile (Python only)
portunix install minimal

# Complete profile (Python + Java 17 + VS Code + Go)
portunix install full

# Empty profile (no pre-installed packages)
portunix install empty
```

## Advanced Usage

### Installation System Architecture

The Portunix installation system uses a multi-layered architecture:

1. **Package Definition Layer** - JSON definitions in `assets/install-packages.json`
2. **Platform Detection Layer** - Automatic OS and architecture detection
3. **Package Manager Selection** - Optimal package manager selection
4. **Installation Executor** - Installation execution with rollback support
5. **Verification Layer** - Successful installation verification

### Supported Installation Methods

#### Windows
- **Chocolatey** - Preferred package manager
- **WinGet** - Microsoft Package Manager
- **MSI** - Direct MSI installer
- **PowerShell** - Installation scripts
- **Direct Download** - Download and extract

#### Linux
- **APT** (Debian/Ubuntu) - `apt-get install`
- **YUM/DNF** (RHEL/Fedora) - `yum install` / `dnf install`
- **Snap** - Universal Linux packages
- **Direct Download** - tar.gz, AppImage, binary files
- **Script** - Shell installation scripts

### Package Definition Structure

Each package is defined in JSON format:

```json
{
  "nodejs": {
    "name": "Node.js",
    "description": "JavaScript runtime built on Chrome's V8 engine",
    "category": "runtime",
    "prerequisites": [],
    "variants": {
      "latest": {
        "version": "22.12.0",
        "windows": {
          "chocolatey": {
            "package": "nodejs",
            "version": "22.12.0"
          },
          "winget": {
            "id": "OpenJS.NodeJS",
            "version": "22.12.0"
          },
          "msi": {
            "url": "https://nodejs.org/dist/v22.12.0/node-v22.12.0-x64.msi",
            "sha256": "..."
          }
        },
        "linux": {
          "script": {
            "commands": [
              "curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -",
              "sudo apt-get install -y nodejs"
            ]
          }
        }
      }
    }
  }
}
```

### Dependency Resolution

Portunix automatically resolves dependencies:

```bash
# Claude Code requires Node.js
portunix install claude-code
# Automatically installs Node.js as prerequisite
```

Dependency graph:
- `claude-code` → requires → `nodejs`
- `maven` → requires → `java`
- `python-full` → includes → `pip`, `venv`, `setuptools`

### Custom Package Definition

You can add custom packages to `assets/install-packages.json`:

```json
{
  "my-tool": {
    "name": "My Custom Tool",
    "description": "Custom development tool",
    "category": "development",
    "prerequisites": ["python"],
    "variants": {
      "latest": {
        "version": "1.0.0",
        "windows": {
          "script": {
            "commands": [
              "pip install my-tool"
            ]
          }
        },
        "linux": {
          "script": {
            "commands": [
              "pip3 install my-tool"
            ]
          }
        }
      }
    }
  }
}
```

### Environment Variables

The installation process respects these environment variables:

- `PORTUNIX_INSTALL_PATH` - Target directory for installations
- `PORTUNIX_CACHE_DIR` - Cache directory for downloaded files
- `PORTUNIX_NO_CACHE` - Disable cache usage
- `PORTUNIX_PACKAGE_MANAGER` - Force specific package manager
- `PORTUNIX_INSTALL_TIMEOUT` - Installation timeout (default: 30min)

### Installation Hooks

Pre and post installation hooks:

```bash
# Pre-install hook
export PORTUNIX_PRE_INSTALL_HOOK="echo 'Starting installation...'"

# Post-install hook
export PORTUNIX_POST_INSTALL_HOOK="echo 'Installation completed!'"

portunix install nodejs
```

### Rollback Mechanism

Portunix automatically creates rollback points:

1. **Pre-installation snapshot** - Save system state
2. **Installation** - Perform installation
3. **Verification** - Verify success
4. **Rollback on failure** - Restore system to original state

```bash
# Manual rollback
portunix install nodejs --rollback-on-error

# Show rollback history
portunix install --show-rollback-history
```

### Batch Installation

Installing multiple packages at once:

```bash
# Via command line
portunix install nodejs python java go

# Via configuration file
portunix install --from-file packages.txt

# packages.txt contains:
# nodejs
# python --variant full
# java --variant 17
```

### Platform-Specific Installation

```bash
# Install only on Windows
portunix install vscode --platform windows

# Install only on Linux
portunix install docker --platform linux

# Install only on specific distribution
portunix install package --platform ubuntu:22.04
```

### Cache Management

```bash
# Show cache
portunix install --show-cache

# Clean cache
portunix install --clean-cache

# Install without cache
portunix install nodejs --no-cache

# Pre-download to cache
portunix install nodejs --download-only
```

## Expert Tips & Tricks

### 1. CI/CD Installation Optimization

```bash
# Parallel installation
portunix install --parallel nodejs python java

# Silent mode for CI
portunix install nodejs --silent --accept-licenses

# JSON output for parsing
portunix install nodejs --output json
```

### 2. Installation with Proxy

```bash
# HTTP proxy
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
portunix install nodejs

# SOCKS proxy
export ALL_PROXY=socks5://proxy.company.com:1080
portunix install nodejs
```

### 3. Offline Installation

```bash
# Download packages for offline use
portunix install nodejs --download-only --target-dir ./offline-packages

# Install from offline cache
portunix install nodejs --offline --cache-dir ./offline-packages
```

### 4. Custom Installation Scripts

```bash
# Run custom script after installation
portunix install nodejs --post-script "npm install -g yarn pnpm"

# Run verification script
portunix install nodejs --verify-script "node --version && npm --version"
```

### 5. Specific Version Installation

```bash
# Install specific version
portunix install nodejs --version 20.11.0

# Install version range
portunix install nodejs --version ">=20.0.0 <21.0.0"

# Install latest/stable/beta
portunix install nodejs --channel beta
```

## Troubleshooting

### Common Problems and Solutions

#### 1. Permission Denied
```bash
# Linux/Mac
sudo portunix install nodejs

# Windows - run as Administrator
```

#### 2. Package Manager Not Available
```bash
# Install Chocolatey on Windows
portunix install chocolatey

# Install Homebrew on Mac
portunix install homebrew
```

#### 3. Version Conflict
```bash
# Show installed versions
portunix install nodejs --check-installed

# Uninstall before reinstallation
portunix uninstall nodejs
portunix install nodejs
```

#### 4. Network Timeouts
```bash
# Increase timeout
portunix install nodejs --timeout 60m

# Use mirror
portunix install nodejs --mirror https://mirror.company.com
```

### Debug Mode

```bash
# Verbose logging
portunix install nodejs -v

# Debug logging
portunix install nodejs --debug

# Trace logging (very detailed)
portunix install nodejs --trace
```

## Integration Examples

### Docker Container Setup
```bash
# Create container with Node.js
portunix docker run ubuntu
portunix docker exec my-container portunix install nodejs
```

### Virtual Machine Setup
```bash
# Set up VM with development environment
portunix virt create dev-vm --iso ubuntu.iso
portunix virt exec dev-vm "portunix install default"
```

### MCP Integration
```bash
# Install tools for AI assistants
portunix install claude-code
portunix mcp configure
```

## API Integration

### REST API
```bash
# Trigger installation via REST API
curl -X POST http://localhost:8080/api/install \
  -H "Content-Type: application/json" \
  -d '{"package": "nodejs", "variant": "latest"}'
```

### gRPC API
```go
// Go client example
client := portunix.NewClient()
result, err := client.Install(context.Background(), &InstallRequest{
    Package: "nodejs",
    Variant: "latest",
    DryRun:  false,
})
```

## Performance Optimization

### Caching Strategy
- **Level 1**: Memory cache (in-process)
- **Level 2**: Disk cache (persistent)
- **Level 3**: Network cache (shared)

### Parallelization
- File downloads: up to 5 concurrent connections
- Independent package installation: parallel execution
- Verification: asynchronous validation

### Resource Limits
```bash
# Limit CPU usage
portunix install nodejs --cpu-limit 50%

# Limit memory
portunix install nodejs --memory-limit 1GB

# Limit disk I/O
portunix install nodejs --io-limit 10MB/s
```

## Security Considerations

### Checksum Verification
All downloaded files are verified using SHA256:

```bash
# Show checksum
portunix install nodejs --show-checksum

# Skip checksum (not recommended)
portunix install nodejs --skip-checksum
```

### Signature Verification
```bash
# Verify GPG signature
portunix install nodejs --verify-signature

# Import GPG key
portunix install --import-key https://nodejs.org/key.asc
```

### Sandboxed Installation
```bash
# Install in sandbox environment
portunix sandbox create test-env
portunix sandbox exec test-env portunix install nodejs
```

## Compatibility Matrix

| Package | Windows | Linux | macOS | Min Version | Max Version |
|---------|---------|-------|-------|-------------|-------------|
| nodejs | ✅ | ✅ | ✅ | 18.0.0 | latest |
| python | ✅ | ✅ | ✅ | 3.8 | 3.13 |
| java | ✅ | ✅ | ✅ | 8 | 21 |
| go | ✅ | ✅ | ✅ | 1.20 | latest |
| docker | ⚠️ | ✅ | ✅ | 20.10 | latest |
| vscode | ✅ | ✅ | ✅ | stable | insider |

## Related Commands

- [`update`](update.md) - Update Portunix itself
- [`plugin`](plugin.md) - Plugin management
- [`system`](system.md) - System information
- [`config`](config.md) - Portunix configuration

## Command Reference

### Complete Parameter List

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `--variant` | string | `latest` | Package variant to install |
| `--dry-run` | boolean | `false` | Preview what would be installed |
| `--force` | boolean | `false` | Force reinstallation |
| `--version` | string | - | Specific version to install |
| `--platform` | string | auto | Target platform |
| `--timeout` | duration | `30m` | Installation timeout |
| `--cache-dir` | string | `~/.portunix/cache` | Cache directory |
| `--no-cache` | boolean | `false` | Don't use cache |
| `--parallel` | boolean | `false` | Parallel installation |
| `--silent` | boolean | `false` | Silent mode |
| `--output` | string | `text` | Output format (text/json/yaml) |
| `--rollback-on-error` | boolean | `true` | Rollback on error |
| `--verify` | boolean | `true` | Verify installation |
| `--proxy` | string | - | Proxy server |
| `--mirror` | string | - | Mirror URL |
| `--cpu-limit` | string | - | CPU limit |
| `--memory-limit` | string | - | Memory limit |
| `--io-limit` | string | - | I/O limit |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successful installation |
| 1 | General error |
| 2 | Package not found |
| 3 | Already installed (without --force) |
| 4 | Dependency error |
| 5 | Network error |
| 6 | Permission denied |
| 7 | Checksum mismatch |
| 8 | Timeout |
| 9 | Rollback performed |
| 10 | Platform not supported |

## Version History

- **v1.5.0** - Added support for package variants
- **v1.4.0** - Implemented dependency resolver
- **v1.3.0** - Added offline installation support
- **v1.2.0** - Implemented rollback mechanism
- **v1.1.0** - Added batch installation support
- **v1.0.0** - Basic installation functionality