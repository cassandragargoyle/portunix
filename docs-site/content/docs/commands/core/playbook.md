---
title: "playbook"
description: "Infrastructure as Code management"
---

# playbook

Infrastructure as Code management

## Usage

```bash
portunix playbook [options] [arguments]
```

## Full Help

```
portunix playbook - Infrastructure as Code Management

USAGE:
  portunix playbook [subcommand] [flags]

DESCRIPTION:
  Manage Ansible Infrastructure as Code using .ptxbook files.
  Supports multi-environment deployments with enterprise features:
  - Secrets management with AES-256-GCM encryption
  - Audit logging with JSON-based tracking
  - Role-based access control (RBAC)
  - CI/CD pipeline integration

SUBCOMMANDS:
  run         Execute a .ptxbook file
  build       Generate production Dockerfile from playbook
  validate    Validate a .ptxbook file syntax and dependencies
  check       Check if ptx-ansible helper is available and working
  list        List available playbooks in current directory
  init        Generate playbook from template
  template    Manage playbook templates
  help        Show this help message

EXAMPLES:
  # Execute a playbook
  portunix playbook run deployment.ptxbook

  # Run specific scripts only
  portunix playbook run my-docs.ptxbook --script dev
  portunix playbook run my-docs.ptxbook --script create,build

  # List available scripts in playbook
  portunix playbook run my-docs.ptxbook --list-scripts

  # Generate production Dockerfile
  portunix playbook build my-docs.ptxbook

  # Validate playbook without execution
  portunix playbook run deployment.ptxbook --dry-run

  # Run in container environment
  portunix playbook run deployment.ptxbook --env container

  # List available templates
  portunix playbook template list

  # Initialize playbook from template
  portunix playbook init my-docs --template static-docs --engine hugo

  # List available playbooks
  portunix playbook list

ENVIRONMENTS:
  local       Execute directly on host system (default)
  container   Execute inside isolated container
  virt        Execute inside virtual machine

ENTERPRISE FEATURES:
  - Encrypted secrets storage and management
  - Complete audit trail with JSON logging
  - Role-based access control for team environments
  - GitHub Actions, GitLab CI, Jenkins integration

For more information about specific subcommands, run:
  portunix playbook [subcommand] --help

```

