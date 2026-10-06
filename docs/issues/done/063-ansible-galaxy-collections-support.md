# Issue #63: Ansible Galaxy Collections Installation Support

**Created**: 2025-09-24
**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement

## Summary
Add support for installing Ansible collections through `ansible-galaxy` command integration within Portunix. This would allow users to install community collections like `community.general`, `ansible.posix`, and other essential collections that extend Ansible's functionality beyond the core modules.

## Current Behavior
- `portunix install ansible` installs ansible-core but doesn't provide collection management
- Users must manually run `ansible-galaxy collection install` commands after Ansible installation
- No integration with Portunix package management system for Ansible collections
- Collections are not tracked or managed through Portunix

## Expected Behavior
```bash
# Basic collection installation
portunix install ansible-collection community.general
portunix install ansible-collection ansible.posix
portunix install ansible-collection community.docker

# Multiple collections at once
portunix install ansible-collection community.general,ansible.posix,community.docker

# Collection installation with variants
portunix install ansible-collection community.general --variant latest
portunix install ansible-collection community.general --variant ">=3.0.0"

# Essential collections bundle
portunix install ansible --variant full-collections  # Includes common collections

# List installed collections
portunix list ansible-collections

# Update collections
portunix update ansible-collection community.general
portunix update ansible-collections  # Update all installed collections
```

## Proposed Implementation

### Command Structure
```bash
# New command category: ansible-collection
portunix install ansible-collection <collection-name>
portunix remove ansible-collection <collection-name>
portunix list ansible-collections
portunix update ansible-collection <collection-name>
portunix info ansible-collection <collection-name>
```

### Package Definition Enhancement
```json
{
  "ansible-collections": {
    "description": "Ansible Galaxy Collections",
    "category": "automation",
    "installer": "ansible-galaxy",
    "prerequisites": ["ansible"],
    "collections": {
      "community.general": {
        "description": "General community modules and plugins",
        "version": "latest",
        "essential": true
      },
      "ansible.posix": {
        "description": "POSIX utilities and modules",
        "version": "latest",
        "essential": true
      },
      "community.docker": {
        "description": "Docker container management modules",
        "version": "latest",
        "essential": false
      },
      "community.kubernetes": {
        "description": "Kubernetes management modules",
        "version": "latest",
        "essential": false
      },
      "community.mysql": {
        "description": "MySQL database management modules",
        "version": "latest",
        "essential": false
      },
      "community.postgresql": {
        "description": "PostgreSQL database management modules",
        "version": "latest",
        "essential": false
      },
      "community.aws": {
        "description": "Amazon Web Services modules",
        "version": "latest",
        "essential": false
      },
      "google.cloud": {
        "description": "Google Cloud Platform modules",
        "version": "latest",
        "essential": false
      },
      "azure.azcollection": {
        "description": "Microsoft Azure modules",
        "version": "latest",
        "essential": false
      }
    }
  }
}
```

### Ansible Installation Variants Enhancement
```json
{
  "ansible": {
    "variants": {
      "core": {
        "description": "Ansible core engine only",
        "packages": ["ansible-core"],
        "collections": []
      },
      "standard": {
        "description": "Ansible core with essential collections",
        "packages": ["ansible-core"],
        "collections": [
          "community.general",
          "ansible.posix"
        ]
      },
      "full": {
        "description": "Ansible with comprehensive collection set",
        "packages": ["ansible-core"],
        "collections": [
          "community.general",
          "ansible.posix",
          "community.docker",
          "community.kubernetes",
          "community.mysql",
          "community.postgresql"
        ]
      },
      "cloud": {
        "description": "Ansible with cloud provider collections",
        "packages": ["ansible-core"],
        "collections": [
          "community.general",
          "ansible.posix",
          "community.aws",
          "google.cloud",
          "azure.azcollection"
        ]
      }
    }
  }
}
```

## Technical Implementation Areas

### New Installer Type: ansible-galaxy
```go
// app/install/ansible_galaxy.go
type AnsibleGalaxyInstaller struct {
    AnsiblePath string
    GalaxyPath  string
}

func (a *AnsibleGalaxyInstaller) Install(collection string, version string) error {
    cmd := exec.Command(a.GalaxyPath, "collection", "install", collection)
    if version != "latest" && version != "" {
        cmd.Args = append(cmd.Args, "--version", version)
    }
    return cmd.Run()
}

func (a *AnsibleGalaxyInstaller) Remove(collection string) error {
    // ansible-galaxy doesn't have native remove, need custom implementation
    return a.removeCollectionManually(collection)
}

func (a *AnsibleGalaxyInstaller) List() ([]string, error) {
    cmd := exec.Command(a.GalaxyPath, "collection", "list")
    output, err := cmd.Output()
    return a.parseCollectionList(string(output)), err
}

func (a *AnsibleGalaxyInstaller) IsInstalled(collection string) bool {
    collections, err := a.List()
    if err != nil {
        return false
    }
    return contains(collections, collection)
}
```

### Command Implementation
```go
// cmd/install_ansible_collection.go
func installAnsibleCollection(collection string, variant string) error {
    // Verify ansible is installed
    if !isAnsibleInstalled() {
        return fmt.Errorf("ansible must be installed first. Run: portunix install ansible")
    }

    installer := &AnsibleGalaxyInstaller{
        AnsiblePath: findAnsiblePath(),
        GalaxyPath:  findAnsibleGalaxyPath(),
    }

    // Parse version from variant
    version := parseVersionFromVariant(variant)

    // Install collection
    return installer.Install(collection, version)
}
```

### Prerequisites Integration
```go
// Enhance existing ansible installation to include collections
func installAnsible(variant string) error {
    // Install ansible-core first
    err := installAnsibleCore()
    if err != nil {
        return err
    }

    // Install collections based on variant
    collections := getCollectionsForVariant(variant)
    for _, collection := range collections {
        err = installAnsibleCollection(collection, "latest")
        if err != nil {
            log.Warnf("Failed to install collection %s: %v", collection, err)
        }
    }

    return nil
}
```

## User Experience Enhancements

### Installation Progress Display
```
════════════════════════════════════════════════
📦 INSTALLING: Ansible Collections
════════════════════════════════════════════════
📄 Description: Essential Ansible Galaxy Collections
🔧 Variant: standard
💻 Platform: windows_vm
🏗️  Installation type: ansible-galaxy
📋 Collections: community.general, ansible.posix
════════════════════════════════════════════════
🔍 Checking if ansible is installed...
✅ ansible-core found at /usr/local/bin/ansible
🔍 Checking collection requirements...
📋 Installing community.general...
✅ community.general v7.5.0 installed successfully
📋 Installing ansible.posix...
✅ ansible.posix v1.5.4 installed successfully
════════════════════════════════════════════════
🎉 Installation completed successfully!
📋 Installed collections:
   • community.general (v7.5.0)
   • ansible.posix (v1.5.4)
════════════════════════════════════════════════
```

### Collection Information Display
```bash
# portunix info ansible-collection community.general
════════════════════════════════════════════════
📦 COLLECTION INFO: community.general
════════════════════════════════════════════════
📄 Name: community.general
🏷️  Version: 7.5.0 (installed), 8.0.2 (latest)
📝 Description: General community modules and plugins
🏗️  Namespace: community
👥 Authors: Ansible Community
📅 Last Updated: 2024-09-15
🔗 Galaxy URL: https://galaxy.ansible.com/community/general
📊 Downloads: 50M+
⭐ Rating: 4.8/5 (2,341 reviews)
════════════════════════════════════════════════
📋 Included Modules:
   • archive, unarchive - Archive management
   • ini_file, xml - Configuration file management
   • htpasswd - HTTP authentication
   • slack, mail - Notification modules
   • docker_* - Docker management (170+ modules)
📋 Dependencies: None
🔧 Compatible with: ansible-core >=2.11.0
════════════════════════════════════════════════
```

## Essential Collections to Support

### Tier 1 (Essential - included in 'standard' variant)
- **community.general**: General utilities, archive, notification modules
- **ansible.posix**: POSIX system utilities, ACL, firewall management

### Tier 2 (Popular - included in 'full' variant)
- **community.docker**: Docker container and image management
- **community.kubernetes**: Kubernetes cluster management
- **community.mysql**: MySQL database administration
- **community.postgresql**: PostgreSQL database administration

### Tier 3 (Cloud - included in 'cloud' variant)
- **community.aws**: Amazon Web Services management
- **google.cloud**: Google Cloud Platform resources
- **azure.azcollection**: Microsoft Azure services
- **community.vmware**: VMware infrastructure management

### Tier 4 (Specialized - install on demand)
- **community.grafana**: Grafana monitoring setup
- **community.zabbix**: Zabbix monitoring configuration
- **community.rabbitmq**: RabbitMQ message broker management
- **community.mongodb**: MongoDB database operations

## Integration with Existing Systems

### MCP Integration for AI Assistants
```json
{
  "tools": [
    {
      "name": "install_ansible_collection",
      "description": "Install Ansible Galaxy collection",
      "parameters": {
        "collection": "Collection name (e.g., community.general)",
        "version": "Version constraint (optional)"
      }
    },
    {
      "name": "list_ansible_collections",
      "description": "List installed Ansible collections"
    }
  ]
}
```

### Playbook Template Integration
```yaml
# Automatically generate requirements.yml when collections are installed
collections:
  - name: community.general
    version: ">=7.0.0"
  - name: ansible.posix
    version: ">=1.5.0"
```

## Error Handling & Edge Cases

### Common Error Scenarios
1. **Ansible not installed**: Clear error message with installation guidance
2. **Network connectivity issues**: Retry logic with exponential backoff
3. **Version conflicts**: Dependency resolution with clear conflict reporting
4. **Permission issues**: Guidance on running with appropriate permissions
5. **Galaxy API failures**: Fallback to cached information when possible

### Validation Requirements
```go
func validateCollectionName(collection string) error {
    // Validate namespace.collection format
    parts := strings.Split(collection, ".")
    if len(parts) != 2 {
        return fmt.Errorf("invalid collection format, expected 'namespace.collection'")
    }

    // Validate namespace and collection name
    for _, part := range parts {
        if !isValidCollectionNamePart(part) {
            return fmt.Errorf("invalid characters in collection name: %s", part)
        }
    }

    return nil
}
```

## Testing Requirements

### Unit Tests
- Collection name validation
- Version parsing and constraint handling
- ansible-galaxy command generation
- Collection list parsing

### Integration Tests
```go
func TestAnsibleCollectionInstallation(t *testing.T) {
    // Test collection installation in container
    tf := testframework.NewTestFramework("AnsibleCollectionInstall")
    tf.Start(t, "Test Ansible collection installation workflow")

    success := true
    defer tf.Finish(t, success)

    // Install ansible first
    tf.Step(t, "Install Ansible core")
    // ... test ansible installation

    // Install collection
    tf.Step(t, "Install community.general collection")
    // ... test collection installation

    // Verify installation
    tf.Step(t, "Verify collection availability")
    // ... verify collection is usable
}
```

### Container-Based Testing
```bash
# Test in clean containers across platforms
portunix docker run-in-container test-ansible-collections --image ubuntu:22.04
portunix docker run-in-container test-ansible-collections --image fedora:latest
portunix docker run-in-container test-ansible-collections --image debian:bookworm
```

## Documentation Requirements

### User Documentation
- Collection installation guide
- Popular collections reference
- Troubleshooting guide for collection issues
- Best practices for collection management

### Developer Documentation
- ansible-galaxy installer API
- Collection metadata format
- Custom collection support
- Testing procedures

## Acceptance Criteria
- [ ] `portunix install ansible-collection community.general` works successfully
- [ ] Collections are properly validated before installation
- [ ] ansible-galaxy prerequisites are checked and installed if needed
- [ ] Multiple collections can be installed in a single command
- [ ] Collection information is displayed with `portunix info ansible-collection`
- [ ] Installed collections are listed with `portunix list ansible-collections`
- [ ] Collection updates work with `portunix update ansible-collection`
- [ ] Enhanced ansible variants include appropriate collections automatically
- [ ] Error messages are clear and actionable
- [ ] Cross-platform compatibility (Windows, Linux, macOS)
- [ ] Container-based testing passes on all platforms
- [ ] MCP integration allows AI assistants to manage collections
- [ ] Performance is acceptable for common operations (< 30 seconds per collection)

## Priority
**High** - Essential functionality for practical Ansible usage, especially given that ansible-core alone is limited without collections

## Labels
- enhancement
- ansible
- galaxy
- collections
- automation
- package-management
- infrastructure-as-code

## Related Issues
- #062: Ansible Installation Issues (prerequisite issue)
- #056: Ansible Infrastructure as Code Integration (related feature)
- General package management system enhancements

## Estimated Effort
- **Analysis & Design**: 2-3 hours
- **ansible-galaxy Installer Implementation**: 6-8 hours
- **Command Interface**: 4-6 hours
- **Package Definition Updates**: 2-3 hours
- **Testing & Validation**: 4-6 hours
- **Documentation**: 2-3 hours
- **Total**: 20-29 hours
