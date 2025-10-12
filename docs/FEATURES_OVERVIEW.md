# 🔧🚀 Portunix Features Overview

This document provides an overview of features available in the Portunix core and standard plugins. It serves as a reference for both the core project (`portunix`) and plugins project (`portunix-plugins`).

## Core Features (portunix)

### 🎯 Core System
- **Cross-platform support**: Windows, Linux, macOS
- **Intelligent OS detection**: Automatic platform detection and optimization
- **Configuration management**: Flexible configuration system

### 🔄 Self-Update & Version Management
- **Automatic updates**: Update Portunix to latest version from GitHub releases
- **Update commands**:
  - `portunix update` - Check and install updates
  - `portunix update --check` - Check for updates only
  - `portunix update --force` - Force reinstall current version
- **Security features**:
  - SHA256 checksum verification
  - Backup before update
  - Rollback capability on failure
  - HTTPS-only downloads
- **Version tracking**: Semantic versioning (SemVer) support
- **Plugin updates** (🚧 planned): Automatic plugin version management and updates

### 📦 Package Management
- **Universal installer**: Cross-platform package installation system
- **Package definitions**: JSON-based package definitions (`assets/install-packages.json`)
- **Supported package managers**:
  - Windows: Chocolatey, WinGet, MSI installers, PowerShell scripts
  - Linux: APT, YUM, DNF, Snap, direct downloads (tar.gz, deb)

#### Pre-configured Software Packages

**Programming Languages & Runtimes:**
- **Java (OpenJDK)**: Multiple LTS versions (8, 11, 17, 21) from Eclipse Adoptium
- **Python**: Version 3.13.6 (embeddable and full installations)
- **Go**: Latest version (1.23.4) from Google
- **PowerShell**: Cross-platform PowerShell for Linux systems

**Development Tools:**
- **Visual Studio Code**: Microsoft's source code editor (stable/insider builds)
- **Apache Maven**: Build automation tool for Java projects (3.9.9)
- **Claude Code**: Anthropic's official CLI for Claude AI assistant
- **GitHub CLI**: Official command line tool for GitHub - manage repos, issues, PRs from the terminal

**Package Managers:**
- **Chocolatey**: Windows package manager
- **WinGet**: Microsoft's official Windows Package Manager

**Web Browsers:**
- **Google Chrome**: Latest stable version

**Installation Profiles:**
- **`default`**: Python + Java 17 + VS Code (recommended for most developers)
- **`minimal`**: Python only (lightweight setup)
- **`full`**: Python + Java 17 + VS Code + Go (complete development environment)
- **`empty`**: Clean environment without any pre-installed packages

### 🐳 Container Management
- **Docker support**:
  - Intelligent Docker installation
  - Container lifecycle management
  - SSH-enabled containers
  - Multi-platform containers (Ubuntu, Alpine, CentOS, Debian)
  - Cache directory mounting
  - Package manager auto-detection
- **Podman support**:
  - Podman Desktop integration
  - Container management
  - Alternative to Docker for rootless containers

### 🪟 Virtualization
- **Windows Sandbox integration**: Isolated development environments
- **VM management**: Virtual machine creation and management
- **Development disk**: Virtual development disk support (Issue #8)

### 🔌 Plugin System
- **gRPC-based architecture**: High-performance plugin communication
- **Plugin lifecycle management**: Install, enable, disable, start, stop, uninstall plugins
- **Plugin creation**: Generate new plugin templates with `plugin create`
- **Plugin validation**: Validate plugin integrity and compatibility
- **Plugin health monitoring**: Check plugin health status
- **Plugin information**: Display detailed plugin metadata and configuration
- **Plugin registry**: Centralized plugin registry
- **Plugin client**: Built-in plugin client for communication
- **Proto definitions**: Protocol Buffer support for plugin API

### 🤖 AI Integration
- **MCP Server**: Model Context Protocol server for AI assistants
- **MCP configuration**: Configure, reconfigure, remove MCP integrations
- **AI-friendly commands**: Designed for integration with AI tools

### 🛠️ Developer Tools
- **Shell completion**: Auto-completion scripts for bash, zsh, fish, and PowerShell
- **Sandbox environments**: Isolated development sandboxes
- **SSH integration**: SSH client and server support
- **Archive management**: ZIP/TAR archive handling
- **REST API**: Built-in REST API server
- **Preprocessor**: Command preprocessing support

### 📊 Data Management
- **Datastore system**: Pluggable datastore architecture (Issue #9)
- **File-based storage**: Default file datastore implementation
- **Plugin datastore**: Plugin-specific data storage support

### 🔐 Security
- **Authentication**: Login/logout system
- **Service management**: Linux service integration
- **Update verification**: Secure update mechanism

## Standard Plugins (portunix-plugins)

### 📋 Agile Software Development Plugin
- **Kanban board management**: Full Kanban methodology support
- **Project management**: Create and manage agile projects
- **Task tracking**: User stories and task management
- **Team collaboration**: Team management features
- **Flow metrics**: Analyze Kanban flow efficiency
- **AI integration**: MCP tools for agile workflows
  - `create_agile_project`
  - `manage_kanban_board`
  - `track_agile_tasks`
  - `analyze_flow_metrics`

## Architecture Overview

### Core Components
```
portunix/
├── app/              # Core application logic
│   ├── datastore/    # Datastore implementations
│   ├── docker/       # Docker integration
│   ├── install/      # Installation system
│   ├── mcp/          # MCP server
│   ├── plugins/      # Plugin system
│   ├── podman/       # Podman integration
│   ├── sandbox/      # Sandbox management
│   ├── system/       # System utilities
│   └── update/       # Update mechanism
├── cmd/              # CLI commands
├── assets/           # Package definitions and scripts
└── parser/           # Command parser
```

### Plugin Architecture
```
portunix-plugins/
├── plugins/          # Plugin implementations
│   └── agile-software-development/
└── registry/         # Plugin registry
```

## Integration Points

### For Core Development
- Extend commands in `cmd/` directory
- Add new installers in `app/install/`
- Implement new datastores in `app/datastore/`
- Add container support in `app/docker/` or `app/podman/`

### For Plugin Development
- Use gRPC protocol for communication
- Implement plugin interface from `app/plugins/proto/`
- Register plugin in `registry/plugin-index.json`
- Follow plugin template structure

## Command Categories

### Core Commands
- `portunix install` - Install packages and tools
- `portunix docker` - Docker management
- `portunix podman` - Podman management
- `portunix sandbox` - Sandbox environments
- `portunix mcp` - MCP server management
- `portunix plugin` - Full plugin lifecycle management (list, install, enable/disable, start/stop, create, validate)
- `portunix update` - Update Portunix
- `portunix system` - System information
- `portunix completion` - Generate shell completion scripts

### Plugin Commands
- `portunix agile` - Agile development tools (via plugin)

## Future Roadmap

### Planned Core Features
- Enhanced virtual development disk (Issue #8)
- Configurable datastore backends (Issue #9)
- Extended MCP capabilities (Issue #4)

### Planned Plugins
- CI/CD integration
- Code quality tools
- Database management
- Cloud provider integrations

## Usage Examples

### Core Usage
```bash
# Install development environment
portunix install default

# Manage Docker containers
portunix docker run ubuntu
portunix docker ssh container-name

# Configure MCP server
portunix mcp configure

# Generate shell completions
portunix completion bash > ~/.bash_completion.d/portunix
portunix completion zsh > ~/.zsh/completions/_portunix

# Plugin management
portunix plugin list                    # List all installed plugins
portunix plugin create my-plugin        # Create new plugin template
portunix plugin install ./my-plugin     # Install plugin from directory
portunix plugin enable my-plugin        # Enable plugin
portunix plugin start my-plugin         # Start plugin service
portunix plugin health my-plugin        # Check plugin health
```

### Plugin Usage
```bash
# Create agile project
portunix agile project create "WebApp Development"

# Manage Kanban board
portunix agile board show
portunix agile task add "Implement authentication"
```

## Documentation References

- Core documentation: `/README.md`, `/docs/`
- Plugin documentation: `../portunix-plugins/docs/`
- Issue tracking: `/docs/issues/`
- Architecture decisions: `/docs/adr/`