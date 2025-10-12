# Portunix Manual

Welcome to the comprehensive Portunix documentation system. This manual provides three different levels of documentation to serve different user needs and use cases.

## Documentation Levels

### 📚 [Basic Manual](basic/README.md)
**Target Audience**: New users, quick references, common tasks

Perfect for:
- Getting started with Portunix
- Learning basic commands and workflows
- Quick reference during daily usage
- Step-by-step tutorials for common tasks

### 🔧 [Expert Manual](expert/README.md)
**Target Audience**: Advanced users, system administrators, developers

Ideal for:
- Advanced feature configuration
- System integration and customization
- Troubleshooting complex issues
- Understanding Portunix architecture

### 🤖 [AI Assistant Manual](ai/README.md)
**Target Audience**: AI assistants, automation tools, integration systems

Designed for:
- Machine-readable command references
- Structured workflow patterns
- API integration guidance
- Automation and scripting contexts

## Quick Navigation

### Getting Started
- [Installation Guide](basic/getting-started/installation.md)
- [First Steps](basic/getting-started/first-steps.md)
- [Common Commands](basic/commands/README.md)

### Popular Topics
- [Package Management](basic/commands/install.md)
- [Container Management](basic/commands/container.md)
- [Infrastructure as Code](basic/commands/playbook.md)
- [Virtualization](basic/commands/virt.md)

### Advanced Topics
- [Architecture Overview](expert/architecture/README.md)
- [Plugin Development](expert/development/plugins.md)
- [Custom Configurations](expert/customization/README.md)

## Help System Integration

This manual is designed to complement Portunix's built-in help system:

```bash
# Basic help (matches Basic Manual)
portunix command --help

# Expert help (matches Expert Manual)
portunix command --help-expert

# AI-optimized help (matches AI Manual)
portunix command --help-ai
```

## Documentation Standards

All documentation follows the [Manual Creation Methodology](../contributing/MANUAL-CREATION-METHODOLOGY.md) to ensure:
- Consistency across all documentation levels
- Up-to-date information synchronized with the CLI
- Quality and accuracy standards
- Proper integration with the help system

## Contributing to Documentation

See the [Manual Creation Methodology](../contributing/MANUAL-CREATION-METHODOLOGY.md) for:
- How to create new documentation
- Templates and style guides
- Quality standards and validation
- Integration with development workflow

## Search and Navigation

### By Feature Category
- **Package Management**: Install, update, and manage software packages
- **Container Management**: Docker and Podman integration
- **Virtualization**: VM creation and management
- **Infrastructure as Code**: Ansible integration with .ptxbook files
- **AI Integration**: MCP server and AI assistant features
- **Plugin System**: Extend Portunix functionality

### By User Journey
- **New User**: Basic Manual → Installation → First Steps
- **Developer**: Expert Manual → Architecture → Development
- **AI Integration**: AI Manual → API Reference → Workflow Patterns

## Feedback and Support

- **Documentation Issues**: Report via [GitHub Issues](https://github.com/cassandragargoyle/Portunix/issues)
- **Feature Requests**: Use the [issue tracking system](../issues/README.md)
- **Community Support**: Join the discussion in project channels

---

**Last Updated**: 2025-09-24
**Version**: 1.0
**Methodology**: [Manual Creation Methodology](../contributing/MANUAL-CREATION-METHODOLOGY.md)