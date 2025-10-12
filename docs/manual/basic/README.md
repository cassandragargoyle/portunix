# Portunix Basic Manual

Welcome to the Portunix Basic Manual! This documentation is designed for new users and provides clear, simple guidance for common tasks.

## Getting Started

### Quick Start
1. [Installation](getting-started/installation.md) - Install Portunix on your system
2. [First Steps](getting-started/first-steps.md) - Your first commands and basic setup
3. [Common Tasks](getting-started/common-tasks.md) - Everyday operations made simple

### Essential Commands
- [`portunix install`](commands/install.md) - Install software packages
- [`portunix container`](commands/container.md) - Manage Docker/Podman containers
- [`portunix playbook`](commands/playbook.md) - Infrastructure as Code with Ansible
- [`portunix virt`](commands/virt.md) - Virtual machine management

## Tutorials

### By Use Case
- [Development Environment Setup](tutorials/dev-environment-setup.md)
- [Container-Based Development](tutorials/container-development.md)
- [Infrastructure Deployment](tutorials/infrastructure-deployment.md)
- [Team Collaboration Setup](tutorials/team-collaboration.md)

### By Technology
- [Working with Docker](tutorials/docker-basics.md)
- [Ansible Playbooks](tutorials/ansible-basics.md)
- [Virtual Machines](tutorials/vm-basics.md)
- [Package Management](tutorials/package-management.md)

## Common Tasks

### Package Management
- Installing development tools
- Setting up programming languages
- Managing dependencies
- Creating installation profiles

### Container Operations
- Creating development containers
- Managing container lifecycle
- Working with multiple containers
- Container networking basics

### Infrastructure Management
- Writing simple playbooks
- Deploying to different environments
- Managing secrets and configuration
- Basic automation workflows

## Troubleshooting

### Common Issues
- [Installation Problems](troubleshooting/installation-issues.md)
- [Container Issues](troubleshooting/container-issues.md)
- [Playbook Errors](troubleshooting/playbook-errors.md)
- [Permission Issues](troubleshooting/permission-issues.md)

### Quick Fixes
- Command not found
- Permission denied errors
- Container startup failures
- Network connectivity issues

## Reference

### Quick Command Reference
```bash
# Package management
portunix install default          # Install default development environment
portunix install nodejs           # Install Node.js

# Container management
portunix container run ubuntu     # Create Ubuntu container
portunix container list           # List containers
portunix container ssh mycontainer # SSH into container

# Infrastructure management
portunix playbook run deploy.ptxbook    # Run deployment playbook
portunix playbook validate config.ptxbook # Validate playbook

# Virtualization
portunix virt list                # List virtual machines
portunix virt create myvm         # Create new VM
```

### Help System
```bash
# Get help for any command
portunix command --help           # Basic help (this manual level)
portunix command --help-expert    # Expert-level help
portunix command --help-ai        # AI-optimized help
```

## Next Steps

### Ready for More?
- **Expert Features**: Check the [Expert Manual](../expert/README.md) for advanced configuration
- **Development**: Learn about [plugin development](../expert/development/README.md)
- **Integration**: Explore [AI assistant integration](../ai/README.md)

### Community Resources
- [Issue Tracking](../../issues/README.md) - Report bugs or request features
- [Contributing Guide](../../contributing/README.md) - Help improve Portunix
- [Architecture Overview](../expert/architecture/README.md) - Understand how Portunix works

---

**Manual Level**: Basic
**Target Audience**: New users, quick references, common tasks
**Last Updated**: 2025-09-24