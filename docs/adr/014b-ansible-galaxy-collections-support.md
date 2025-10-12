# ADR-014: Ansible Galaxy Collections Support

**Status**: Accepted
**Date**: 2025-09-24
**Deciders**: Development Team

## Context

Ansible has evolved from a monolithic structure to a modular ecosystem where ansible-core provides the engine, and collections provide the modules and plugins. The current Portunix ansible installer only installs ansible-core, leaving users without essential collections like `community.general` and `ansible.posix`, which are required for most real-world Ansible usage.

### Problem Statement
- Installing `portunix install ansible` provides limited functionality without collections
- Users must manually manage collections via `ansible-galaxy` commands
- No integration with Portunix package management for Ansible collections
- Missing automation-friendly collection management for CI/CD and development workflows

### Current State
- Ansible installation variants: `core`, `full`, `latest`
- Post-install commands in install-packages.json manually install some collections
- No systematic approach to collection management
- No support for collection versioning or dependency resolution

## Decision

Implement comprehensive Ansible Galaxy Collections support within Portunix package management system with the following architecture:

### 1. New Package Category: `ansible-collection`
```bash
portunix install ansible-collection community.general
portunix install ansible-collection community.general --variant ">=7.0.0"
portunix list ansible-collections
portunix update ansible-collection community.general
portunix info ansible-collection community.general
```

### 2. Enhanced Ansible Installation Variants
Extend existing ansible variants to include collections:
- `core`: ansible-core only (current behavior)
- `standard`: ansible-core + essential collections (community.general, ansible.posix)
- `full`: ansible-core + comprehensive collection set
- `cloud`: ansible-core + cloud provider collections

### 3. New Installer Type: `ansible-galaxy`
Create dedicated installer that:
- Uses `ansible-galaxy collection install` commands
- Handles version constraints and dependency resolution
- Provides installation verification and status checking
- Integrates with Portunix prerequisites system

### 4. Collection Metadata Integration
Extend install-packages.json structure to define:
- Collection definitions with descriptions and categories
- Version constraints and compatibility requirements
- Essential vs optional collections classification
- Installation groupings for different use cases

## Implementation Architecture

### Core Components

1. **AnsibleGalaxyInstaller** (`src/app/install/ansible_galaxy/installer.go`)
   - Implements standard Portunix installer interface
   - Handles ansible-galaxy command execution
   - Manages collection verification and status
   - Provides error handling and retry logic

2. **Collection Management Commands** (`src/cmd/ansible_collection.go`)
   - New command category for collection-specific operations
   - Integrates with existing install/list/info/update commands
   - Provides collection-specific help and documentation

3. **Package Definition Extensions** (`assets/install-packages.json`)
   - ansible-collections package category
   - Enhanced ansible variants with collection bundles
   - Collection metadata with descriptions and requirements

4. **Prerequisites Integration**
   - Automatic ansible dependency checking
   - Version compatibility validation
   - System prerequisites (Python, pip) verification

### Collection Tiers

**Tier 1 (Essential)**: Included in `standard` variant
- community.general: General utilities, archive management
- ansible.posix: POSIX system utilities, ACL management

**Tier 2 (Popular)**: Included in `full` variant
- community.docker: Docker container management
- community.kubernetes: Kubernetes cluster management
- community.mysql, community.postgresql: Database management

**Tier 3 (Cloud)**: Included in `cloud` variant
- community.aws, google.cloud, azure.azcollection
- cloud-specific infrastructure management

**Tier 4 (Specialized)**: On-demand installation
- community.grafana, community.zabbix: Monitoring
- community.rabbitmq, community.mongodb: Specialized services

## Technical Considerations

### Advantages
1. **Unified Package Management**: Collections managed through Portunix ecosystem
2. **Dependency Resolution**: Automatic ansible prerequisite handling
3. **Version Management**: Support for version constraints and updates
4. **Automation Friendly**: Scriptable collection management for CI/CD
5. **User Experience**: Consistent interface with other Portunix packages
6. **AI Integration**: MCP tools for AI-assisted collection management

### Challenges Addressed
1. **ansible-galaxy Limitations**: No native uninstall command
   - Solution: Manual collection removal via filesystem operations
2. **Collection Discovery**: Finding appropriate collections for use cases
   - Solution: Curated collection tiers and descriptions
3. **Version Conflicts**: Managing collection dependencies
   - Solution: Validation and clear error reporting
4. **Network Dependencies**: Galaxy API availability
   - Solution: Retry logic and graceful degradation

### Performance Considerations
- Collection installation target: < 30 seconds per collection
- Parallel installation for multiple collections
- Caching of collection metadata
- Progress reporting for long-running operations

## Testing Strategy

### Unit Tests
- Collection name validation and parsing
- Version constraint handling
- ansible-galaxy command generation
- Error handling scenarios

### Integration Tests
- End-to-end collection installation in containers
- Cross-platform compatibility (Ubuntu, Fedora, Debian)
- Ansible workflow validation with installed collections
- Performance benchmarking

### Container-Based Testing
```bash
portunix container run-in-container ansible-collection-test --image ubuntu:22.04
```

## Security Considerations

### Supply Chain Security
- Collection signature verification (when available)
- Source validation from official Galaxy registry
- Dependency scanning for known vulnerabilities

### Permission Management
- User-level collection installation (avoid system-wide when possible)
- Clear permission requirements documentation
- Fallback strategies for restricted environments

## Migration Strategy

### Phase 1: Core Implementation (Week 1-2)
- ansible-galaxy installer implementation
- Basic collection install/list commands
- Enhanced ansible variants

### Phase 2: Advanced Features (Week 3)
- Collection info and update commands
- MCP integration for AI assistants
- Performance optimization

### Phase 3: Documentation & Testing (Week 4)
- Comprehensive testing across platforms
- User documentation and guides
- Integration with existing Portunix workflows

### Backward Compatibility
- Existing ansible installation commands unchanged
- Current post-install collection installations continue working
- Gradual migration to new collection management system

## Monitoring and Success Metrics

### Success Criteria
- 95% success rate for essential collection installations
- < 30 second installation time per collection
- Zero breaking changes to existing ansible workflows
- Positive user feedback on collection management experience

### Monitoring Points
- Collection installation success/failure rates
- Performance metrics for different collection types
- User adoption of new collection commands
- Error patterns and resolution effectiveness

## Alternative Approaches Considered

### Alternative 1: External ansible-galaxy wrapper
- **Rejected**: Poor integration with Portunix ecosystem
- **Reason**: Users expect unified package management experience

### Alternative 2: Collection bundling in ansible package
- **Rejected**: Inflexible and resource intensive
- **Reason**: Users need fine-grained collection control

### Alternative 3: Separate collection management tool
- **Rejected**: Increases complexity for users
- **Reason**: Conflicts with Portunix unified approach

## Future Considerations

### Extensibility
- Custom collection source support (private Galaxy, Git)
- Collection development workflow integration
- Advanced dependency resolution algorithms

### Integration Opportunities
- Ansible playbook template generation
- Collection requirement.yml auto-generation
- IDE integration for collection discovery

---

**Decision Record**: This ADR establishes the foundation for comprehensive Ansible Galaxy Collections support within Portunix, enabling users to manage collections through the unified package management interface while maintaining flexibility and performance.

**Implementation Priority**: High - Essential for practical Ansible usage
**Estimated Effort**: 20-29 hours across 4 weeks
**Dependencies**: Issue #062 (Ansible Installation Issues) - Completed