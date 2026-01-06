# ptx-ansible

Portunix Ansible Infrastructure as Code Helper - dedicated helper binary for Ansible IaC integration.

## Overview

`ptx-ansible` is a helper binary that handles all Ansible Infrastructure as Code operations within the Portunix ecosystem. It provides `.ptxbook` file parsing, validation, and execution for unified infrastructure management across multiple environments.

**Architecture Decision Record:** [ADR-016](../../../docs/adr/016-ansible-infrastructure-as-code-integration.md)
**Issue:** [#056](../../../docs/issues/internal/056-ansible-infrastructure-as-code-integration.md)

## Features

### Phase 1: Foundation

- **Ptxbook Parser**: YAML-based `.ptxbook` file format (API version: `portunix.ai/v1`)
- **Validation**: Schema validation for playbook structure and content
- **Local Execution**: Direct execution on host system
- **Ansible Integration**: Execute Ansible playbooks after Portunix package installation
- **Dry-Run Mode**: Validate without making changes

### Phase 2: Multi-Environment

- **Container Execution**: Execute playbooks inside isolated containers
- **Virtual Machine Execution**: Execute playbooks on VMs via SSH
- **Inventory Auto-Generation**: Dynamic Ansible inventory creation
- **SSH Key Management**: Automatic SSH connectivity setup

### Phase 3: Advanced Features

- **Conditional Execution**: `when` conditions for packages and playbooks
- **Variable Templating**: Jinja2-style variable substitution
- **Rollback Protection**: Transaction-like execution with automatic rollback on failure
- **MCP Integration**: AI-assisted playbook generation and validation

### Phase 4: Enterprise Features

- **Secrets Management**: AES-256-GCM encrypted secret storage
- **Audit Logging**: Comprehensive JSON-based audit trail
- **Role-Based Access Control (RBAC)**: Multi-user permission management
- **CI/CD Integration**: GitHub Actions, GitLab CI, Jenkins pipeline support

## Architecture

```
src/helpers/ptx-ansible/
├── main.go          # Entry point and command handling
├── executor.go      # Playbook execution engine
├── ptxbook.go       # .ptxbook file parsing and validation
├── templating.go    # Jinja2-style template processing
├── rollback.go      # Rollback management system
├── mcp.go           # MCP tools for AI integration
├── secrets.go       # Secrets management (AES-256-GCM)
├── audit.go         # Audit logging system
├── rbac.go          # Role-based access control
├── cicd.go          # CI/CD pipeline integration
├── go.mod           # Go module definition
└── go.sum           # Dependency checksums
```

## .ptxbook File Format

```yaml
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "development-environment"
  description: "Complete development environment setup"

spec:
  variables:
    java_version: "17"
    nodejs_version: "20"

  requirements:
    ansible:
      min_version: "2.15.0"

  portunix:
    packages:
      - name: "java"
        variant: "{{ java_version }}"
      - name: "nodejs"
        variant: "{{ nodejs_version }}"
        when: "os == 'linux'"
      - name: "vscode"
        variant: "stable"

  ansible:
    playbooks:
      - path: "./ansible/database-setup.yml"
      - path: "./ansible/nginx-config.yml"
        when: "environment == 'production'"

  rollback:
    enabled: true
    preserve_logs: true
    timeout: "5m"
    on_failure:
      - type: "command"
        command: "echo 'Cleaning up...'"
        description: "Clean up on failure"
```

## Commands

The helper is invoked by the main `portunix` dispatcher:

```bash
# Execute playbook
portunix playbook run <playbook.ptxbook>

# Execute with environment
portunix playbook run <playbook.ptxbook> --env container --image ubuntu:22.04
portunix playbook run <playbook.ptxbook> --env virt --target my-vm

# Dry-run validation
portunix playbook run <playbook.ptxbook> --dry-run

# Validate syntax
portunix playbook validate <playbook.ptxbook>

# Check helper availability
portunix playbook check

# List playbooks
portunix playbook list

# Initialize new playbook
portunix playbook init <name> --template development
```

### MCP Commands

```bash
# Generate playbook from natural language
ptx-ansible mcp generate "Setup Java development environment with Docker"

# Validate with AI suggestions
ptx-ansible mcp validate deployment.ptxbook

# List playbooks with metadata
ptx-ansible mcp list ./playbooks

# Export MCP manifest
ptx-ansible mcp manifest
```

### Enterprise Commands

```bash
# Secrets management
ptx-ansible secrets

# Audit logging
ptx-ansible audit status
ptx-ansible audit stats

# RBAC management
ptx-ansible rbac status
ptx-ansible rbac roles

# CI/CD integration
ptx-ansible cicd status

# Security validation
ptx-ansible security validate

# Compliance reporting
ptx-ansible compliance report
```

## Environment Detection

The template engine automatically provides built-in variables:

| Variable | Description |
|----------|-------------|
| `os` | Operating system (linux, windows, darwin) |
| `arch` | Architecture (amd64, arm64) |
| `os_family` | OS family (unix, windows) |
| `user` | Current username |
| `home` | Home directory path |
| `pwd` | Current working directory |
| `hostname` | System hostname |
| `is_container` | Running inside container (bool) |
| `is_vm` | Running inside VM (bool) |
| `is_wsl` | Running inside WSL (bool) |

## Conditional Expressions

```yaml
packages:
  # Simple equality
  - name: "powershell"
    when: "os == 'linux'"

  # Inequality
  - name: "chocolatey"
    when: "os != 'linux'"

  # Boolean check
  - name: "docker"
    when: "is_container"
```

## Secret References

Secrets can be referenced in `.ptxbook` files using the format:

```yaml
variables:
  api_key: "{{ secret:vault:api_key }}"
  db_password: "{{ secret:db_password }}"  # Uses default store
```

Supported secret stores:
- `file` - File-based encrypted storage
- `env` - Environment variables (read-only)
- `vault` - HashiCorp Vault (planned)

## RBAC Roles

Default roles:

| Role | Description | Permissions |
|------|-------------|-------------|
| `admin` | Full system administrator | All permissions |
| `developer` | Standard developer access | Read/write playbooks, local/container execution |
| `operator` | Production operations | Read/execute playbooks, VM execution |
| `auditor` | Audit and compliance | Read playbooks, view audit logs |

## CI/CD Provider Support

| Provider | Config Generation | Webhook Support |
|----------|-------------------|-----------------|
| GitHub Actions | ✅ | ✅ |
| GitLab CI | ✅ | ✅ |
| Jenkins | ✅ | ✅ |
| Azure DevOps | ⏳ Planned | ⏳ Planned |

## Build

```bash
# Build from project root
make build-helpers

# Or build directly
cd src/helpers/ptx-ansible
go build -o ptx-ansible
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing

## Integration with Portunix

The main `portunix` binary delegates playbook commands to `ptx-ansible`:

1. User runs: `portunix playbook run setup.ptxbook`
2. Main binary checks for `ptx-ansible` helper
3. If `.ptxbook` contains no `ansible` section, main binary handles execution
4. If `ansible` section present, delegates to `ptx-ansible`
5. Error handling and logging coordinated between binaries

## Related Documentation

- [ADR-016: Ansible IaC Integration](../../../docs/adr/016-ansible-infrastructure-as-code-integration.md)
- [ADR-014: Git-like Dispatcher Architecture](../../../docs/adr/014-git-dispatcher-python-distribution-architecture.md)
- [Issue #056: Ansible Integration](../../../docs/issues/internal/056-ansible-infrastructure-as-code-integration.md)
- [Features Overview](../../../docs/FEATURES_OVERVIEW.md)

## License

Part of the Portunix project.
