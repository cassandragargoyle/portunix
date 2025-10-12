# ADR-016: Ansible Infrastructure as Code Integration

**Status**: Proposed
**Date**: 2025-09-23
**Author**: Architect

## Context

The Portunix project needs a robust Infrastructure as Code (IaC) solution to manage complex deployment scenarios across multiple environments (local machines, virtual machines, and containers). Currently, Portunix provides excellent package management and environment setup capabilities, but lacks orchestration capabilities for complex multi-node deployments and configuration management.

Ansible has emerged as the industry standard for Infrastructure as Code with several advantages:
- Agentless architecture using SSH
- YAML-based playbooks that are human-readable
- Extensive module ecosystem
- Strong community support
- Integration with cloud providers and container platforms

However, Portunix has unique advantages in package management, cross-platform support, and unified CLI experience that should be preserved and enhanced rather than replaced.

## Decision

Portunix will adopt Ansible Infrastructure as Code principles and integrate with Ansible playbooks while maintaining its own strengths. The integration will follow these principles:

### 1. Infrastructure as Code (IaC) Principles
- Configuration should be declarative, not imperative
- Infrastructure state should be version-controlled
- Changes should be reproducible and idempotent
- Infrastructure should be self-documenting

### 2. Hybrid Architecture: Portunix + Ansible
- **Portunix handles**: Package installation, environment setup, container/VM management
- **Ansible handles**: Multi-node orchestration, configuration management, service deployment
- **Integration point**: Portunix-specific playbooks that leverage both tools

### 3. Portunix Playbook Format
Introduce a new file format: **`.ptxbook`** (Portunix Playbook)
- Based on YAML like Ansible, but optimized for Portunix workflows
- Includes both Portunix commands and Ansible playbook references
- Supports environment-specific variables and configurations

### 4. Multi-Environment Support
- **Local execution**: Direct execution on local machine
- **Virtual machine execution**: Deployment to VMs managed by `portunix virt`
- **Container execution**: Deployment to containers managed by `portunix container`

## Implementation Architecture

### 1. Helper Binary Architecture

Following the established Portunix pattern of helper binaries (`ptx-container`, `ptx-mcp`), the Ansible integration will be implemented as a separate helper binary:

**`ptx-ansible`** - Dedicated Ansible integration helper binary
- Contains all Ansible-specific logic and playbook execution
- Handles .ptxbook parsing and validation
- Manages Ansible installation and dependency resolution
- Implements environment-specific execution (local/container/VM)
- Provides inventory generation and SSH key management

**`portunix`** - Main dispatcher
- Routes `playbook` commands to `ptx-ansible`
- Maintains consistent CLI interface
- Handles configuration and logging integration
- Delegates to existing Portunix subsystems (install, virt, docker)

This architecture provides:
- **Separation of concerns** - Ansible logic isolated from core
- **Independent development** - Can be updated separately from main binary
- **Reduced complexity** - Main binary stays lightweight
- **Consistent pattern** - Follows established `ptx-container` and `ptx-mcp` approach
- **Optional dependency** - `ptx-ansible` only required when .ptxbook references Ansible playbooks
- **Ansible-free workflows** - .ptxbook files with only Portunix packages don't require Ansible

### 2. Portunix Playbook Structure (.ptxbook)

```yaml
# example.ptxbook
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "development-environment-setup"
  description: "Complete development environment with Java, Node.js, and databases"
  version: "1.0.0"
  author: "DevOps Team"

spec:
  # Ansible version requirements and configuration
  requirements:
    ansible:
      min_version: "2.15.0"
      max_version: "3.0.0"
      collections:
        - "community.general"
        - "ansible.posix"

  # Environment-specific variables
  variables:
    java_version: "17"
    nodejs_version: "20"
    database_type: "postgresql"

  # Portunix-specific setup phase
  portunix:
    # Ensure Ansible is installed via Portunix
    prerequisites:
      - name: "ansible"
        variant: "latest"
        min_version: "2.15.0"
        required: true

    # Portunix package installations
    packages:
      - name: "java"
        variant: "{{ java_version }}"
      - name: "nodejs"
        variant: "{{ nodejs_version }}"
      - name: "vscode"
        variant: "stable"

    # Environment setup
    environments:
      - type: "local"
        enabled: true
      - type: "container"
        image: "ubuntu:22.04"
        enabled: false
      - type: "virt"
        template: "ubuntu-22.04"
        enabled: false

  # Ansible playbooks to execute after Portunix setup
  ansible:
    playbooks:
      - path: "./ansible/database-setup.yml"
        when: "{{ database_type == 'postgresql' }}"
      - path: "./ansible/application-deployment.yml"
        vars:
          app_environment: "development"

    # Ansible inventory management
    inventory:
      auto_generate: true
      groups:
        - name: "development"
          hosts: "localhost"
```

### 3. Command Structure

The main `portunix` binary provides the user interface while delegating to `ptx-ansible`:

```bash
# Execute Portunix playbook (delegates to ptx-ansible)
portunix playbook run <playbook.ptxbook>

# Execute on specific environment
portunix playbook run <playbook.ptxbook> --env container
portunix playbook run <playbook.ptxbook> --env virt --target my-vm

# Dry-run mode - validate without making changes
portunix playbook run <playbook.ptxbook> --dry-run
portunix playbook run <playbook.ptxbook> --env container --dry-run

# Validate playbook syntax
portunix playbook validate <playbook.ptxbook>

# List available playbooks
portunix playbook list

# Generate template playbook
portunix playbook init <name> --template development|production|minimal

# Check if ptx-ansible helper is available
portunix playbook check

# Install ptx-ansible helper binary
portunix install ptx-ansible
```

**Internal delegation:**
- `portunix playbook *` commands are routed to `ptx-ansible` helper only when needed
- If .ptxbook contains no `ansible` section, execution handled by main binary
- If .ptxbook contains `ansible` section and `ptx-ansible` not found, user is prompted to install it
- Error handling and logging are coordinated between main binary and helper

### 4. Execution Flow

1. **Parse .ptxbook file**: Validate syntax and extract configuration
2. **Dependency analysis**:
   - Check if .ptxbook contains `ansible` section
   - If no Ansible section present, skip Ansible-related phases
   - If Ansible section present, ensure `ptx-ansible` helper is available
3. **Version validation phase** (only if Ansible required):
   - Check Ansible version requirements from `spec.requirements.ansible`
   - Validate installed Ansible version meets `min_version` and `max_version` constraints
   - Install missing Ansible collections if specified
   - Fail early if version requirements cannot be met
4. **Portunix prerequisite phase**:
   - Install required packages via Portunix package system
   - Ensure Ansible is installed via `portunix install ansible` (only if needed)
   - Setup target environments (containers/VMs) if needed
5. **Environment preparation** (only if Ansible required):
   - Generate Ansible inventory based on target environment
   - Setup SSH keys and connectivity for remote environments
6. **Ansible execution phase** (only if Ansible section present):
   - Execute referenced Ansible playbooks
   - Pass variables from .ptxbook to Ansible
   - Support `--dry-run` mode for validation without actual changes
   - Handle error reporting and rollback if needed
7. **Post-execution validation**:
   - Verify deployment success (skipped in dry-run mode)
   - Generate execution report

### 5. Integration Points

#### 5.1 Ansible Installation Management
```bash
# Portunix ensures Ansible is properly installed
portunix install ansible
# Includes: ansible-core, ansible collections, required dependencies
```

#### 5.2 Environment-Specific Execution

**Local execution:**
```bash
portunix playbook run setup.ptxbook --env local
# Executes on current machine

portunix playbook run setup.ptxbook --env local --dry-run
# Validates configuration without making changes
```

**Container execution:**
```bash
portunix playbook run setup.ptxbook --env container --image ubuntu:22.04
# Creates container, executes setup, optionally commits result

portunix playbook run setup.ptxbook --env container --image ubuntu:22.04 --dry-run
# Shows what would be executed without creating container
```

**Virtual machine execution:**
```bash
portunix playbook run setup.ptxbook --env virt --target production-vm
# Executes on specified VM via SSH

portunix playbook run setup.ptxbook --env virt --target production-vm --dry-run
# Shows what would be executed without making changes to VM
```

#### 5.3 Integration with Existing Portunix Commands

**.ptxbook files can reference Portunix commands:**
```yaml
portunix:
  custom_commands:
    - command: "portunix docker run-in-container nodejs --script ./build.sh"
      when: "{{ environment == 'container' }}"
    - command: "portunix virt exec {{ vm_name }} -- systemctl start nginx"
      when: "{{ environment == 'virt' }}"
```

## File Extension: .ptxbook

**Rationale for .ptxbook extension:**
- **Distinctiveness**: Clearly different from Ansible's `.yml`/`.yaml`
- **Brand alignment**: Incorporates "ptx" (Portunix abbreviation)
- **Functional clarity**: "book" indicates it's a playbook-like configuration
- **Convention consistency**: Follows container industry patterns (Dockerfile, Containerfile)

## Consequences

### Positive Consequences

1. **Unified Workflow**: Single tool for package management + infrastructure orchestration
2. **Best of Both Worlds**: Leverage Portunix's package management with Ansible's orchestration
3. **Environment Consistency**: Same playbook works across local/container/VM environments
4. **Reduced Complexity**: No need to learn separate Ansible inventory management
5. **Version Control Friendly**: YAML-based configuration fits well in Git workflows
6. **Progressive Adoption**: Teams can start with simple Portunix commands and gradually add Ansible complexity

### Negative Consequences

1. **Learning Curve**: Teams need to understand both Portunix and Ansible concepts
2. **Additional Dependency**: Ansible becomes a required dependency for IaC features
3. **Complexity**: New file format and execution model increases cognitive load
4. **Debugging Difficulty**: Errors could occur in either Portunix or Ansible phases
5. **Performance Overhead**: Additional layer between user and Ansible execution

### Risk Mitigation

1. **Documentation**: Comprehensive examples and tutorials for .ptxbook format
2. **Validation**: Strong syntax validation and helpful error messages
3. **Fallback**: Users can still use pure Ansible if they prefer
4. **Incremental Rollout**: Start with simple use cases and expand gradually

## Implementation Phases

### Phase 1: Foundation 
- Create `ptx-ansible` helper binary structure
- Basic .ptxbook parser and validator in `ptx-ansible`
- Ansible installation via Portunix package system
- Dispatcher logic in main `portunix` binary for `playbook` commands
- Local environment execution only
- Simple package + playbook workflows

### Phase 2: Multi-Environment 
- Extend `ptx-ansible` with container environment support
- Virtual machine environment support in `ptx-ansible`
- Inventory auto-generation
- SSH key management
- Enhanced integration between main binary and helper

### Phase 3: Advanced Features 
- Conditional execution
- Variable templating
- Error handling and rollback
- Integration with Portunix MCP server

### Phase 4: Enterprise Features 
- Secrets management integration
- Audit logging
- Role-based access control
- CI/CD pipeline integration

## Examples

### Ansible-Free Development Environment Setup
```yaml
# simple-dev-setup.ptxbook
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "simple-development-environment"
  description: "Basic development setup without Ansible"

spec:
  variables:
    java_version: "17"

  portunix:
    packages:
      - {name: "java", variant: "{{ java_version }}"}
      - {name: "nodejs", variant: "20"}
      - {name: "vscode", variant: "stable"}
      - {name: "docker", variant: "latest"}

  # No ansible section - ptx-ansible not required
```

### Full-Stack Development Environment with Ansible
```yaml
# dev-setup.ptxbook
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "full-stack-development"

spec:
  requirements:
    ansible:
      min_version: "2.15.0"

  portunix:
    packages:
      - {name: "java", variant: "17"}
      - {name: "nodejs", variant: "20"}
      - {name: "docker", variant: "latest"}

  ansible:
    playbooks:
      - path: "./ansible/postgres-setup.yml"
      - path: "./ansible/redis-setup.yml"
      - path: "./ansible/nginx-proxy.yml"
```

### Production Deployment
```yaml
# production.ptxbook
apiVersion: portunix.ai/v1
kind: Playbook
metadata:
  name: "production-deployment"

spec:
  variables:
    app_version: "{{ env.APP_VERSION | default('latest') }}"

  portunix:
    prerequisites:
      - {name: "ansible", variant: "latest"}

  ansible:
    playbooks:
      - path: "./ansible/security-hardening.yml"
      - path: "./ansible/application-deployment.yml"
        vars:
          version: "{{ app_version }}"
      - path: "./ansible/monitoring-setup.yml"
```

## References

- [Ansible Documentation](https://docs.ansible.com/)
- [Infrastructure as Code Principles](https://www.hashicorp.com/resources/what-is-infrastructure-as-code)
- [Portunix Package Management System](../FEATURES_OVERVIEW.md)
- [Issue #049: QEMU Full Support Implementation](../issues/internal/049-qemu-full-support-implementation.md)
- [ADR-014: Git-like Dispatcher with Python Distribution Model](014-git-dispatcher-python-distribution-architecture.md)

---

**Next Steps**: This ADR requires implementation planning and resource allocation. Consider creating a dedicated issue for tracking the implementation phases.