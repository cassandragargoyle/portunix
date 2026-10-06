# Issue #56: Ansible Infrastructure as Code Integration

**Status**: 📋 Open
**Priority**: High
**Type**: Feature
**Created**: 2025-09-23
**Architecture Decision Record**: [ADR-016](../../adr/016-ansible-infrastructure-as-code-integration.md)

## Problem Statement

Portunix needs a robust Infrastructure as Code (IaC) solution to manage complex deployment scenarios across multiple environments (local machines, virtual machines, and containers). Currently, Portunix provides excellent package management and environment setup capabilities, but lacks orchestration capabilities for complex multi-node deployments and configuration management.

## Proposed Solution

Implement Ansible Infrastructure as Code integration following the architecture defined in [ADR-016](../../adr/016-ansible-infrastructure-as-code-integration.md).

### Key Components

#### 1. Helper Binary Architecture
- **`ptx-ansible`** - Dedicated Ansible integration helper binary
- **`portunix`** - Main dispatcher routing playbook commands

#### 2. Portunix Playbook Format (.ptxbook)
New YAML-based file format optimized for Portunix workflows with both Portunix commands and Ansible playbook references.

#### 3. Multi-Environment Support
- Local execution on current machine
- Container execution via Portunix container management
- Virtual machine execution via Portunix VM management

## Implementation Phases

### Phase 1: Foundation
**Goal**: Basic .ptxbook execution with Ansible integration
**Timeline**: Sprint 1-2

#### Tasks:
1. **Helper Binary Infrastructure**
   - [ ] Create `ptx-ansible` binary project structure
   - [ ] Implement basic binary skeleton with version and help commands
   - [ ] Set up build pipeline for `ptx-ansible` in main Makefile
   - [ ] Create integration tests for binary communication

2. **Playbook Parser and Validator**
   - [ ] Implement .ptxbook YAML parser in `ptx-ansible`
   - [ ] Add schema validation for .ptxbook format (apiVersion, kind, spec)
   - [ ] Validate `spec.requirements.ansible` section
   - [ ] Validate `spec.portunix.packages` section
   - [ ] Implement error reporting for malformed .ptxbook files

3. **Ansible Installation System**
   - [ ] Add ansible package definition to `assets/install-packages.json`
   - [ ] Include ansible-core, required collections, and dependencies
   - [ ] Test Ansible installation across Windows/Linux platforms
   - [ ] Implement version requirement checking

4. **Main Binary Dispatcher**
   - [ ] Add `playbook` command group to main `portunix` binary
   - [ ] Implement `portunix playbook run` → `ptx-ansible` delegation
   - [ ] Add `portunix playbook validate` → `ptx-ansible` delegation
   - [ ] Add `portunix playbook check` for helper availability
   - [ ] Implement error handling when `ptx-ansible` not found

5. **Local Environment Execution**
   - [ ] Implement local .ptxbook execution in `ptx-ansible`
   - [ ] Support Portunix-only .ptxbook files (no ansible section)
   - [ ] Support mixed .ptxbook files (portunix + ansible sections)
   - [ ] Implement `--dry-run` mode for validation

6. **Basic Workflows**
   - [ ] Test simple package installation via .ptxbook
   - [ ] Test Ansible playbook execution after package installation
   - [ ] Test error scenarios (missing Ansible, invalid playbooks)
   - [ ] Create integration tests for Phase 1 functionality

#### Success Criteria:
- [ ] `ptx-ansible` binary builds and executes independently
- [ ] .ptxbook files parse and validate correctly
- [ ] Portunix-only workflows execute without Ansible dependency
- [ ] Mixed workflows execute Portunix packages followed by Ansible playbooks
- [ ] Error handling provides clear user guidance

### Phase 2: Multi-Environment
**Goal**: Container and VM environment support
**Timeline**: Sprint 3-4

#### Tasks:
1. **Container Environment Support**
   - [ ] Extend `ptx-ansible` with `--env container` support
   - [ ] Implement container creation for .ptxbook execution
   - [ ] Add container SSH setup for Ansible connectivity
   - [ ] Support container image specification in .ptxbook
   - [ ] Test .ptxbook execution inside containers

2. **Virtual Machine Environment Support**
   - [ ] Extend `ptx-ansible` with `--env virt` support
   - [ ] Integration with `portunix virt` system for VM targeting
   - [ ] Add VM SSH connectivity for Ansible execution
   - [ ] Support VM targeting via `--target` parameter
   - [ ] Test .ptxbook execution on VMs

3. **Inventory Auto-Generation**
   - [ ] Implement dynamic Ansible inventory generation
   - [ ] Support localhost targeting for local execution
   - [ ] Support container targeting with SSH details
   - [ ] Support VM targeting with SSH connectivity
   - [ ] Handle inventory templating and variable substitution

4. **SSH Key Management**
   - [ ] Implement SSH key generation for environments
   - [ ] Add SSH key distribution to containers/VMs
   - [ ] Handle SSH connectivity testing and validation
   - [ ] Support custom SSH key paths and configurations

5. **Enhanced Integration**
   - [ ] Improve error communication between main binary and helper
   - [ ] Add detailed logging for multi-environment operations
   - [ ] Implement progress reporting for long-running operations
   - [ ] Add timeout handling for environment setup

#### Success Criteria:
- [ ] .ptxbook execution works in containers with proper SSH connectivity
- [ ] .ptxbook execution works on VMs with proper targeting
- [ ] Inventory auto-generation produces valid Ansible inventories
- [ ] SSH connectivity is established and validated before execution

### Phase 3: Advanced Features
**Goal**: Production-ready features and integrations
**Timeline**: Sprint 5-6

#### Tasks:
1. **Conditional Execution**
   - [ ] Implement `when` conditions in .ptxbook format
   - [ ] Support variable-based conditional execution
   - [ ] Add environment-based conditionals (local/container/VM)
   - [ ] Test complex conditional workflows

2. **Variable Templating**
   - [ ] Implement Jinja2-style variable templating in .ptxbook
   - [ ] Support environment variables in templates
   - [ ] Add custom variable sources and overrides
   - [ ] Test variable resolution and substitution

3. **Error Handling and Rollback**
   - [ ] Implement transaction-like execution with rollback
   - [ ] Add rollback hooks for failed Portunix installations
   - [ ] Implement rollback hooks for failed Ansible executions
   - [ ] Test rollback scenarios and error recovery

4. **Integration with Portunix MCP Server**
   - [ ] Add MCP tools for .ptxbook creation and management
   - [ ] Implement MCP tools for playbook execution monitoring
   - [ ] Add AI-assisted .ptxbook generation capabilities
   - [ ] Test MCP integration with Claude Code

#### Success Criteria:
- [ ] Conditional execution works correctly across all environments
- [ ] Variable templating resolves correctly in complex scenarios
- [ ] Rollback mechanisms prevent system contamination on failures
- [ ] MCP integration provides AI-assisted workflow capabilities

### Phase 4: Enterprise Features
**Goal**: Enterprise-grade security and management
**Timeline**: Sprint 7-8

#### Tasks:
1. **Secrets Management Integration**
   - [ ] Implement secure secrets handling in .ptxbook files
   - [ ] Add integration with external secret stores
   - [ ] Support encrypted variables and sensitive data
   - [ ] Test secrets management across environments

2. **Audit Logging**
   - [ ] Implement comprehensive audit logging
   - [ ] Add execution tracking and change logging
   - [ ] Support log export and integration with monitoring systems
   - [ ] Test audit trail completeness

3. **Role-Based Access Control**
   - [ ] Implement user authentication and authorization
   - [ ] Add role-based .ptxbook execution permissions
   - [ ] Support environment-specific access controls
   - [ ] Test RBAC scenarios and edge cases

4. **CI/CD Pipeline Integration**
   - [ ] Add support for CI/CD environment execution
   - [ ] Implement headless/automated execution modes
   - [ ] Add integration with popular CI/CD systems
   - [ ] Test automated deployment scenarios

#### Success Criteria:
- [ ] Secrets are handled securely without exposure
- [ ] Audit logs provide complete execution traceability
- [ ] RBAC prevents unauthorized .ptxbook execution
- [ ] CI/CD integration enables automated infrastructure deployment

## Command Structure

```bash
# Execute Portunix playbook (delegates to ptx-ansible)
portunix playbook run <playbook.ptxbook>

# Execute on specific environment
portunix playbook run <playbook.ptxbook> --env container
portunix playbook run <playbook.ptxbook> --env virt --target my-vm

# Dry-run mode - validate without making changes
portunix playbook run <playbook.ptxbook> --dry-run

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

## Example .ptxbook File

### Ansible-Free Development Environment
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

## Benefits

### Positive Consequences
- **Unified Workflow**: Single tool for package management + infrastructure orchestration
- **Best of Both Worlds**: Leverage Portunix's package management with Ansible's orchestration
- **Environment Consistency**: Same playbook works across local/container/VM environments
- **Progressive Adoption**: Teams can start with simple Portunix commands and gradually add Ansible complexity

### Challenges to Address
- **Learning Curve**: Teams need to understand both Portunix and Ansible concepts
- **Additional Dependency**: Ansible becomes optional dependency for IaC features
- **Complexity**: New file format and execution model increases cognitive load

## Success Criteria

- [ ] `ptx-ansible` helper binary successfully parses .ptxbook files
- [ ] Basic Portunix-only .ptxbook execution works without Ansible
- [ ] Ansible playbook execution works when Ansible section is present
- [ ] Multi-environment execution (local/container/VM) functions correctly
- [ ] Dry-run mode validates without making changes
- [ ] Integration with existing Portunix package system
- [ ] Comprehensive testing across Phase 1 functionality

## Dependencies

- Ansible package integration in Portunix package system
- Helper binary architecture from [ADR-014](../../adr/014-git-dispatcher-with-python-distribution.md)
- Container and VM management capabilities
- YAML parsing and validation libraries

## Related Issues

### Core Infrastructure Dependencies
- [#049](049-qemu-full-support-implementation.md) - QEMU Full Support Implementation
- [#051](../051-git-dispatcher-python-distribution-architecture.md) - Git-like Dispatcher Architecture

### Package Management System (Critical for Ansible Integration)
- [#041](041-nodejs-npm-installation-support.md) - Node.js/npm Installation Support
- [#012](012-powershell-linux-installation.md) - PowerShell Installation Support for Linux
- [#022](209-google-chrome-installation.md) - Google Chrome Installation Implementation
- [#023](../204-arch-linux-distribution-support.md) - Arch Linux Distribution Support
- [#026](212-github-cli-installation.md) - GitHub CLI (gh) Installation Support
- [#016](205-protoc-plugin-development-dependency.md) - Protocol Buffers Compiler (protoc)
- [#035](../035-ai-assistant-installation-support.md) - AI Assistant Installation Support

### Package Management Architecture & Design
- [ADR-006](../../adr/006-dynamic-package-list-generation.md) - Dynamic Package List Generation
- [ADR-002](../../adr/002-powershell-linux-architecture.md) - PowerShell Linux Architecture
- [ADR-007](../../adr/007-prerequisite-package-handling-system.md) - Prerequisite Package Handling System
- [ADR-003](../../adr/003-linux-distribution-version-support-strategy.md) - Linux Distribution Version Support Strategy
- [ADR-008](../../adr/008-dynamic-sudo-handling-post-install-commands.md) - Dynamic Sudo Handling Post-Install Commands

### Container & VM Integration
- [#030](216-container-tls-certificate-verification-failure.md) - Container TLS Certificate Verification
- [#028](214-universal-container-parameters-support.md) - Universal Container Parameters Support
- [#029](215-universal-container-command.md) - Universal Container Command Implementation

## Labels

`enhancement`, `infrastructure-as-code`, `ansible`, `helper-binary`, `multi-environment`