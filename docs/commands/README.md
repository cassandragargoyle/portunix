# Portunix Command Reference

Comprehensive documentation for all Portunix commands, organized by functional categories and structured from basic usage to expert-level features.

## 📖 Documentation Philosophy

Each command documentation follows a progressive structure:
- **Quick Start** - Essential usage for immediate productivity
- **Intermediate Usage** - Common scenarios and workflow integration
- **Advanced Usage** - Complex configurations and expert features
- **Expert Tips & Tricks** - Performance optimization and advanced patterns
- **Troubleshooting** - Common issues and diagnostic techniques

## 🗂️ Command Categories

### 🔧 [Core Commands](core/)
**Essential package management, system operations, and Portunix maintenance**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`install`](core/install.md) | Universal package installation | `portunix install nodejs` |
| [`update`](core/update.md) | Portunix self-update system | `portunix update` |
| [`system`](core/system.md) | System information & diagnostics | `portunix system info` |

*These commands form the foundation of every Portunix workflow. Master these first.*

---

### 🐳 [Container Commands](containers/)
**Universal container management with automatic runtime detection and enhanced Portunix integration**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`container`](containers/container.md) ⭐ | **Universal container interface** | `portunix container run ubuntu` |
| [`docker`](containers/docker.md) | Docker-specific operations | `portunix docker run-in-container nodejs` |
| `podman` | Podman-specific operations | `portunix podman run ubuntu --rootless` *(Coming Soon)* |

*The **`container`** command is the recommended interface that automatically detects and uses the best available runtime (Docker/Podman) while providing cross-runtime compatibility.*

---

### 🖥️ [Virtualization Commands](virtualization/)
**Cross-platform virtual machine management and automation**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`virt`](virtualization/virt.md) | Universal VM management | `portunix virt create myvm --iso ubuntu.iso` |
| `sandbox` | Windows Sandbox integration | `portunix sandbox create test-env` *(Coming Soon)* |

*Complete isolation and cross-platform compatibility for complex development and testing scenarios.*

---

### 🔌 [Plugin Commands](plugins/)
**Extensible plugin ecosystem with gRPC-based architecture**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`plugin`](plugins/plugin.md) | Plugin lifecycle management | `portunix plugin create my-plugin` |

*Transform Portunix from a tool into a platform with unlimited extensibility and customization.*

---

### 🤖 [Integration Commands](integration/)
**AI assistant integration and external system connectivity**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`mcp`](integration/mcp.md) | Model Context Protocol server | `portunix mcp configure` |
| `api` | REST API server | `portunix api serve` *(Coming Soon)* |

*Bridge the gap between human intent and system operations through AI assistants and external integrations.*

---

### 🛠️ [Utility Commands](utilities/)
**Productivity enhancements and user experience improvements**

| Command | Purpose | Quick Example |
|---------|---------|---------------|
| [`completion`](utilities/completion.md) | Shell completion system | `portunix completion bash` |
| `help` | Advanced help system | `portunix help --interactive` *(Coming Soon)* |
| `config` | Configuration management | `portunix config set theme dark` *(Coming Soon)* |
| `version` | Version information | `portunix version --detailed` *(Coming Soon)* |

*Essential productivity tools that make the command-line experience delightful and efficient.*

## 🚀 Quick Start Guides

### New User Onboarding
```bash
# 1. Set up shell completion for better experience
portunix completion bash > ~/.bash_completion.d/portunix
source ~/.bash_completion.d/portunix

# 2. Install development tools
portunix install default

# 3. Get system overview
portunix system info

# 4. (Optional) Enable AI integration
portunix mcp configure
```

### Development Environment Setup
```bash
# Complete development environment in minutes
portunix install default                           # Core dev tools
portunix docker run-in-container nodejs --name dev # Containerized environment
portunix plugin install agile-software-development # Project management
portunix mcp serve --daemon                       # AI assistant integration
```

### Container Development Workflow
```bash
# Create development container with pre-installed tools
portunix docker run-in-container "nodejs python git" \
  --name fullstack-dev \
  --volumes $(pwd):/workspace \
  --ports 3000:3000

# SSH into development environment
portunix docker ssh fullstack-dev

# Inside container: start development
cd /workspace && npm run dev
```

### Cross-Platform Testing
```bash
# Create VMs for different platforms
portunix virt create ubuntu-test --iso ubuntu-22.04.iso
portunix virt create centos-test --iso centos-8.iso

# Run tests across platforms
for vm in ubuntu-test centos-test; do
  portunix virt ssh $vm "portunix install myapp && myapp test"
done
```

## 📊 Command Usage Matrix

### By User Experience Level

| Experience | Start Here | Next Steps | Advanced |
|------------|------------|------------|----------|
| **Beginner** | [`install`](core/install.md), [`system`](core/system.md) | [`completion`](utilities/completion.md), [`docker`](containers/docker.md) | [`virt`](virtualization/virt.md) |
| **Intermediate** | [`docker`](containers/docker.md), [`plugin`](plugins/plugin.md) | [`mcp`](integration/mcp.md), [`virt`](virtualization/virt.md) | Custom plugins, automation |
| **Expert** | [`mcp`](integration/mcp.md), plugin development | Advanced automation, CI/CD | Platform integration |

### By Use Case

| Use Case | Primary Commands | Supporting Commands |
|----------|------------------|-------------------|
| **Package Management** | [`install`](core/install.md), [`update`](core/update.md) | [`system`](core/system.md) |
| **Container Development** | [`docker`](containers/docker.md) | [`install`](core/install.md), [`completion`](utilities/completion.md) |
| **VM Management** | [`virt`](virtualization/virt.md) | [`system`](core/system.md), [`install`](core/install.md) |
| **AI Integration** | [`mcp`](integration/mcp.md) | [`install`](core/install.md), [`plugin`](plugins/plugin.md) |
| **Plugin Development** | [`plugin`](plugins/plugin.md) | [`docker`](containers/docker.md), [`completion`](utilities/completion.md) |
| **System Administration** | [`system`](core/system.md), [`update`](core/update.md) | [`mcp`](integration/mcp.md) |

### By Technology Stack

| Technology | Primary Commands | Integration Notes |
|------------|------------------|-------------------|
| **Node.js** | [`install`](core/install.md), [`docker`](containers/docker.md) | Pre-configured containers available |
| **Python** | [`install`](core/install.md), [`virt`](virtualization/virt.md) | Multiple variants (full, minimal) |
| **Java** | [`install`](core/install.md), [`system`](core/system.md) | Multiple JDK versions supported |
| **Docker** | [`docker`](containers/docker.md), [`install`](core/install.md) | Enhanced with SSH and tools |
| **Kubernetes** | [`virt`](virtualization/virt.md), [`docker`](containers/docker.md) | VM-based K8s clusters |
| **AI/ML** | [`mcp`](integration/mcp.md), [`python`](core/install.md) | Claude integration |

## 🔄 Command Integration Patterns

### Core → Container → AI Workflow
```bash
# 1. Install development tools
portunix install "nodejs python docker"

# 2. Create development container
portunix docker run-in-container nodejs --name ai-dev

# 3. Enable AI integration
portunix mcp configure

# 4. AI can now manage your development environment
# Example: "Set up a React project with TypeScript"
```

### Plugin → Integration → Automation
```bash
# 1. Create custom plugin
portunix plugin create deployment-manager

# 2. Expose via MCP for AI access
portunix mcp tools register deployment-manager

# 3. AI-assisted deployments
# Example: "Deploy version 2.1.0 to staging environment"
```

### System → VM → Container Hierarchy
```bash
# 1. Analyze system capabilities
portunix system info --capabilities

# 2. Create development VM
portunix virt create dev-vm --template ubuntu-dev

# 3. Run containers inside VM
portunix virt ssh dev-vm "portunix docker run nginx"
```

## 🎯 Workflow Examples

### Daily Development Routine
```bash
# Morning setup
portunix system info --quick                  # Check system health
portunix update --check                      # Check for updates
portunix docker start dev-env                # Start development container
portunix mcp serve --daemon                  # Enable AI assistant

# During development
portunix docker ssh dev-env                  # Enter development environment
# AI: "Install the latest testing framework"
# AI: "Create a new feature branch and set up boilerplate"

# End of day
portunix docker snapshot dev-env --name daily-backup
portunix system monitor --duration 1h --report summary
```

### Project Onboarding Workflow
```bash
# New team member setup
portunix completion install --shell auto      # Shell completion
portunix install default                     # Standard development tools
portunix plugin install agile-software-development # Team tools

# Project-specific environment
portunix docker run-in-container "nodejs python postgresql" \
  --name $(basename $PWD)-dev \
  --volumes $(pwd):/workspace

# Enable AI assistance for project
portunix mcp configure --project-context
```

### Testing and Deployment Pipeline
```bash
# Create test environments
portunix virt create test-ubuntu --template test-base
portunix virt create test-centos --template test-base

# Automated testing
for vm in test-*; do
  portunix virt ssh $vm "
    portunix install myapp --version $CI_VERSION
    myapp test --suite integration
    myapp benchmark --output $vm-results.json
  "
done

# AI-assisted analysis
# AI: "Analyze test results and recommend deployment strategy"
```

## 🔍 Command Discovery

### Finding the Right Command

#### By Task
- **Install software** → [`install`](core/install.md)
- **Create development environment** → [`docker`](containers/docker.md) or [`virt`](virtualization/virt.md)
- **Extend functionality** → [`plugin`](plugins/plugin.md)
- **AI integration** → [`mcp`](integration/mcp.md)
- **Improve productivity** → [`completion`](utilities/completion.md)
- **System diagnostics** → [`system`](core/system.md)

#### By Problem
- **Software won't install** → [`system requirements`](core/system.md), [`install --debug`](core/install.md)
- **Container issues** → [`docker diagnose`](containers/docker.md)
- **VM problems** → [`virt health`](virtualization/virt.md)
- **Plugin errors** → [`plugin health`](plugins/plugin.md)
- **Slow completion** → [`completion optimize`](utilities/completion.md)

#### By Output Format
- **Human-readable** → Default format for all commands
- **JSON** → `--output json` (most commands)
- **YAML** → `--output yaml` (system, plugin info)
- **CSV** → `--output csv` (system software, metrics)
- **Markdown** → `--output markdown` (reports, documentation)

## 🆘 Getting Help

### Built-in Help System
```bash
# Quick help
portunix <command> --help

# Detailed help with examples
portunix <command> --help-expert

# AI-friendly format
portunix <command> --help-ai

# Interactive help (coming soon)
portunix help --interactive
```

### Documentation Navigation
- **Category Overview** → Each category README provides comprehensive overview
- **Command Deep-Dive** → Individual command documentation with examples
- **Cross-References** → Links between related commands and categories
- **Integration Examples** → Real-world usage patterns and workflows

### Community Resources
- **[GitHub Repository](https://github.com/cassandragargoyle/portunix)** - Source code and development
- **[GitHub Issues](https://github.com/cassandragargoyle/portunix/issues)** - Bug reports and feature requests
- **[Contributing Guidelines](../contributing/)** - How to contribute to Portunix
- **[Architecture Documentation](../adr/)** - Technical decision records

## 🔮 Future Roadmap

### Short-term Enhancements (v1.6.x)
- **Universal Container Interface** - Auto-detect Docker/Podman
- **Advanced Help System** - Interactive tutorials and context-aware assistance
- **Configuration Management** - Global settings and team configurations
- **Enhanced MCP Tools** - More AI assistant integrations

### Medium-term Features (v1.7.x)
- **Web Dashboard** - Browser-based management interface
- **Mobile App** - System monitoring and basic controls
- **Enhanced Plugins** - Marketplace and advanced plugin features
- **Cloud Integration** - Native cloud platform support

### Long-term Vision (v2.0+)
- **Visual Interface** - GUI for complex operations
- **Advanced AI Features** - Predictive assistance and automated optimization
- **Enterprise Features** - Advanced security, compliance, and management
- **Platform Ecosystem** - Third-party integrations and partnerships

## 📏 Documentation Standards

### Command Documentation Structure
1. **Quick Start** - Essential examples for immediate use
2. **Intermediate Usage** - Common patterns and configurations
3. **Advanced Usage** - Complex scenarios and expert features
4. **Expert Tips & Tricks** - Performance and advanced workflows
5. **Troubleshooting** - Common issues and solutions
6. **Command Reference** - Complete parameter documentation

### Cross-Reference Conventions
- `[command](category/command.md)` - Links to other commands
- `command:feature` - References to specific functionality
- **Bold** for categories and important concepts
- `code` for commands, parameters, and technical terms

### Example Standards
- Prefer real-world scenarios over abstract examples
- Include expected output where helpful
- Show both basic and advanced usage patterns
- Demonstrate integration with other commands

---

**Navigation**: [🏠 Main Documentation](../README.md) | [🤝 Contributing](../contributing/README.md) | [🏗️ Architecture](../adr/README.md)

*Last updated: Generated dynamically with Portunix v1.5.14*

---

*Portunix transforms system administration from reactive to proactive, enabling developers to focus on what matters most - building great software.*