# Core Commands

Essential commands for package management, system operations, and Portunix maintenance.

## Commands in this Category

### [`install`](install.md) - Package Installation System
Universal cross-platform package installation with dependency resolution and multiple package manager support.

**Quick Examples:**
```bash
portunix install nodejs                    # Install Node.js
portunix install python --variant full     # Full Python environment
portunix install default                   # Default development profile
```

**Key Features:**
- Cross-platform package management (Windows, Linux, macOS)
- Automatic dependency resolution
- Installation profiles (default, minimal, full, empty)
- Rollback mechanism with automatic recovery
- Checksum verification and security
- Offline installation support

**Common Use Cases:**
- Development environment setup
- CI/CD pipeline software installation
- Quick prototype environment creation
- Cross-platform deployment preparation

---

### [`update`](update.md) - Portunix Self-Update
Secure automatic update system with verification and rollback capabilities.

**Quick Examples:**
```bash
portunix update                    # Check and install updates
portunix update --check           # Check for updates only
portunix update --force           # Force reinstall current version
```

**Key Features:**
- SHA256 checksum verification
- Automatic rollback on failure
- Update channels (stable, beta, nightly)
- Background update scheduling
- GPG signature verification
- Update history and audit trail

**Common Use Cases:**
- Keeping Portunix up-to-date
- Testing beta features
- Automated CI/CD updates
- Security patch deployment

---

### [`system`](system.md) - System Information & Diagnostics
Comprehensive system information gathering, monitoring, and compatibility checking.

**Quick Examples:**
```bash
portunix system info              # Complete system overview
portunix system hardware          # Hardware specifications
portunix system requirements nodejs # Check package requirements
```

**Key Features:**
- Hardware and software inventory
- Real-time performance monitoring
- Compatibility checking for packages
- Environment variable analysis
- Multiple output formats (JSON, YAML, CSV, Markdown)
- Container and virtualization detection

**Common Use Cases:**
- System diagnostics and troubleshooting
- Environment preparation verification
- Performance monitoring and analysis
- Compatibility validation before installation

## Category Overview

The **Core** category contains the fundamental commands that every Portunix user will interact with regularly. These commands form the foundation of the Portunix ecosystem and are essential for:

### Daily Operations
- Installing and managing software packages
- Keeping Portunix updated with latest features
- Gathering system information for troubleshooting

### Development Workflows
- Setting up development environments quickly
- Ensuring compatibility across different systems
- Automating software installation in pipelines

### System Administration
- Monitoring system health and performance
- Validating system requirements
- Managing software dependencies

## Integration with Other Categories

Core commands integrate seamlessly with other Portunix categories:

### With Containers
```bash
# Install software in containers
portunix docker run ubuntu
portunix docker exec my-container portunix install nodejs
```

### With Plugins
```bash
# Install prerequisites for plugins
portunix install nodejs  # Required for many plugins
portunix plugin install agile-software-development
```

### With Virtualization
```bash
# Setup VM with software
portunix virt create dev-vm --iso ubuntu.iso
portunix virt ssh dev-vm "portunix install default"
```

### With MCP Integration
```bash
# Install MCP tools
portunix install claude-code
portunix mcp configure
```

## Quick Start Workflows

### New Developer Setup
```bash
# 1. Get system information
portunix system info

# 2. Install development profile
portunix install default

# 3. Verify installations
portunix system software --category development

# 4. Keep updated
portunix update
```

### CI/CD Integration
```bash
# 1. Check system compatibility
portunix system requirements --profile full

# 2. Install required tools
portunix install nodejs python java --parallel

# 3. Verify successful installation
portunix system software --check nodejs,python,java

# 4. Update Portunix if needed
portunix update --check && portunix update --auto-confirm
```

### Cross-Platform Development
```bash
# 1. Check platform-specific features
portunix system capabilities

# 2. Install with platform detection
portunix install docker --platform auto

# 3. Export system profile
portunix system profile export > dev-profile.json

# 4. Replicate on other systems
portunix system profile import dev-profile.json
```

## Best Practices

### Package Installation
- Use `--dry-run` to preview installations
- Always verify with `portunix system requirements` first
- Use installation profiles for consistent environments
- Keep backups with rollback points

### Update Management
- Enable automatic update checking
- Test updates in development environments first
- Use specific channels (stable/beta) based on needs
- Monitor update logs for issues

### System Monitoring
- Regular system information gathering
- Set up monitoring for critical systems
- Use JSON output for automated processing
- Archive system profiles for compliance

## Related Categories

- **[Plugins](../plugins/)** - Extend core functionality
- **[Containers](../containers/)** - Containerized environments
- **[Integration](../integration/)** - AI and external system integration
- **[Utilities](../utilities/)** - Supporting tools and enhancements

## Support and Troubleshooting

For issues with core commands:
1. Check `portunix system info` for environment details
2. Use `--debug` flag for detailed output
3. Consult troubleshooting sections in individual command docs
4. Check [GitHub Issues](https://github.com/cassandragargoyle/portunix/issues)

---

*Core commands form the foundation of the Portunix ecosystem - master these first for the best experience.*