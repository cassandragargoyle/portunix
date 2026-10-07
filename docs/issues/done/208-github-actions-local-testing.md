# Issue #208: GitHub Actions Local Testing Support with Act

> **Renumbered:** formerly internal issue #021. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

## Overview
**Type**: Feature  
**Priority**: Medium  
**Status**: ✅ Implemented  
**Created**: 2025-02-01  
**Target Version**: 1.6.0

## Problem Description

Developers need to test GitHub Actions workflows locally before pushing to GitHub to:
- Avoid breaking CI/CD pipeline
- Save time on debugging workflow issues
- Test changes without creating commits
- Reduce GitHub Actions usage minutes

Currently, there's no easy way to install and use `act` (GitHub Actions local runner) through Portunix.

## Proposed Solution

### 1. Add Act Tool to Package Definitions

Create new package definition for `act` in `install-packages.json`:
- Support for Linux, macOS, and Windows
- Binary download from GitHub releases
- Automatic version detection
- Docker dependency check

### 2. Create GitHub Actions Development Profile

Add new installation profile `github-actions` that includes:
- `act` - GitHub Actions local runner
- `docker` - Required dependency for act
- `gh` - GitHub CLI for workflow management
- `actionlint` - GitHub Actions workflow linter (optional)

### 3. Implementation Details

#### Package Definition Structure:
```json
{
  "act": {
    "name": "Act - GitHub Actions Local Runner",
    "description": "Run GitHub Actions locally using Docker containers",
    "category": "development",
    "platforms": {
      "linux": {
        "type": "direct_download",
        "variants": {
          "latest": {
            "version": "latest",
            "url": "https://api.github.com/repos/nektos/act/releases/latest",
            "download_pattern": "act_Linux_x86_64.tar.gz",
            "extract": true,
            "binary": "act",
            "install_path": "/usr/local/bin"
          }
        }
      },
      "windows": {
        "type": "direct_download",
        "variants": {
          "latest": {
            "version": "latest",
            "url": "https://api.github.com/repos/nektos/act/releases/latest",
            "download_pattern": "act_Windows_x86_64.zip",
            "extract": true,
            "binary": "act.exe",
            "install_path": "C:\\Program Files\\Act"
          }
        }
      },
      "darwin": {
        "type": "direct_download",
        "variants": {
          "latest": {
            "version": "latest",
            "url": "https://api.github.com/repos/nektos/act/releases/latest",
            "download_pattern": "act_Darwin_x86_64.tar.gz",
            "extract": true,
            "binary": "act",
            "install_path": "/usr/local/bin"
          }
        }
      }
    },
    "dependencies": ["docker"],
    "post_install": [
      "act --version",
      "echo 'Act installed successfully. Docker is required to run workflows.'"
    ]
  },
  "actionlint": {
    "name": "Actionlint - GitHub Actions Linter",
    "description": "Static checker for GitHub Actions workflow files",
    "category": "development",
    "platforms": {
      "linux": {
        "type": "direct_download",
        "variants": {
          "latest": {
            "version": "latest",
            "url": "https://api.github.com/repos/rhysd/actionlint/releases/latest",
            "download_pattern": "actionlint_.*_linux_amd64.tar.gz",
            "extract": true,
            "binary": "actionlint",
            "install_path": "/usr/local/bin"
          }
        }
      }
    }
  }
}
```

#### Installation Profile:
```json
{
  "profiles": {
    "github-actions": {
      "name": "GitHub Actions Development",
      "description": "Tools for local GitHub Actions development and testing",
      "packages": [
        {"name": "docker", "variant": "latest"},
        {"name": "act", "variant": "latest"},
        {"name": "gh", "variant": "latest"},
        {"name": "actionlint", "variant": "latest"}
      ]
    }
  }
}
```

### 4. Usage Examples

```bash
# Install GitHub Actions development tools
portunix install github-actions

# Or install just act
portunix install act

# After installation, users can:
act -l                    # List workflows
act                       # Run default push event
act -W .github/workflows/test.yml  # Run specific workflow
act -j build             # Run specific job
act --container-architecture linux/amd64  # Specify architecture
```

### 5. Additional Features

#### Helper Commands
Create wrapper commands for common act operations:

```bash
portunix gh-actions test            # Run all tests locally
portunix gh-actions test --workflow build.yml  # Run specific workflow
portunix gh-actions validate        # Lint workflow files with actionlint
portunix gh-actions list           # List available workflows
```

#### Configuration Helper
Create `.actrc` configuration helper:

```bash
portunix gh-actions configure      # Interactive setup for act
```

This would create `.actrc` file with:
- Default runner image
- Environment variables
- Secrets management
- Cache configuration

## Success Criteria

- [ ] Act package successfully installs on Linux, Windows, macOS
- [ ] Docker dependency is checked and user is warned if missing
- [ ] Installation profile includes all necessary tools
- [ ] Version detection works for latest releases
- [ ] Post-install verification confirms act is working
- [ ] Documentation includes usage examples
- [ ] Helper commands simplify common operations

## Testing Requirements

1. **Installation Testing**
   - Test on Ubuntu, Windows 10/11, macOS
   - Verify binary is in PATH
   - Check version command works

2. **Functionality Testing**
   - Run simple workflow with act
   - Verify Docker integration
   - Test with different runner images

3. **Profile Testing**
   - Install full github-actions profile
   - Verify all tools work together

## Dependencies

- Docker must be installed and running
- Internet connection for downloading releases
- Sufficient disk space for Docker images

## Documentation Updates

1. Add act to supported packages list
2. Create guide for GitHub Actions local testing
3. Document common workflow testing scenarios
4. Add troubleshooting section for Docker issues

## Future Enhancements

1. **Workflow Templates**
   - Provide common workflow templates
   - Generate workflow files from templates

2. **Caching Support**
   - Cache runner images locally
   - Manage act cache directory

3. **Integration with IDE**
   - VS Code extension configuration
   - IntelliJ IDEA integration

4. **Monitoring**
   - Resource usage monitoring during workflow runs
   - Performance profiling

## References

- [Act GitHub Repository](https://github.com/nektos/act)
- [Act Documentation](https://github.com/nektos/act/blob/master/README.md)
- [Actionlint Repository](https://github.com/rhysd/actionlint)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)

## Notes

This feature will significantly improve developer productivity by allowing local testing of GitHub Actions workflows before pushing to repository. It's especially valuable for complex workflows and reduces the feedback loop during development.

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented:** 2025-01-09  
**Version:** v1.5.7+  

### Implementation Summary

GitHub Actions local testing support has been successfully implemented:

#### ✅ Act Installation Support
- **`portunix install act`** command available for Act installation
- **Cross-platform support** for Linux, Windows, macOS
- **Automatic binary detection** and version management
- **Integration with package management system** via install-packages.json

#### ✅ Actionlint Integration  
- **`portunix install actionlint`** command for workflow validation
- **Static checking** of GitHub Actions workflow files
- **Error detection** and best practices enforcement
- **CI/CD integration** capabilities

#### ✅ Package Definitions
Both Act and Actionlint are properly defined in the Portunix package system:
- Automated downloads from official GitHub releases
- Version management and updates
- Cross-platform binary selection
- Installation verification

### Verification Commands

```bash
# Install Act for local GitHub Actions testing
portunix install act

# Install Actionlint for workflow validation  
portunix install actionlint

# Verify installation
act --version
actionlint --version

# Test local GitHub Actions (in repository with .github/workflows/)
act
act -l  # List available workflows
act -j test  # Run specific job
```

### Usage Examples

```bash
# Local testing workflow
cd your-project/
portunix install act
act  # Run all workflows locally

# Validate workflows
portunix install actionlint  
actionlint .github/workflows/*.yml
```

### Resolution Status

**FULLY IMPLEMENTED** ✅ - Both Act and Actionlint are available through Portunix installation system, enabling comprehensive local GitHub Actions testing and workflow validation capabilities.