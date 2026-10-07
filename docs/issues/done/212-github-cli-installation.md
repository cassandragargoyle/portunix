# Issue #212: GitHub CLI (gh) Installation Support

> **Renumbered:** formerly internal issue #026. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Issue ID**: #026  
**Type**: Feature  
**Priority**: Medium  
**Status**: ✅ Implemented  
**Created**: 2025-01-09  
**Labels**: enhancement, package-management, github, cli, cross-platform

## Problem Statement

Client requires GitHub CLI (`gh`) installation through Portunix package management system to enable GitHub integration capabilities directly from command line.

## Background

GitHub CLI is the official command-line tool for GitHub, allowing users to:
- Create and manage repositories
- Handle pull requests and issues
- Run GitHub Actions workflows
- Manage GitHub releases
- Authenticate with GitHub services

Currently, Portunix doesn't include GitHub CLI in its package definitions, limiting users' ability to integrate GitHub workflows seamlessly.

## Requirements

### Functional Requirements
1. **Cross-platform Installation**: Support GitHub CLI installation on Windows and Linux
2. **Package Integration**: Add GitHub CLI to `assets/install-packages.json`
3. **Version Management**: Support latest stable version installation
4. **Installation Profiles**: Include GitHub CLI in relevant installation profiles (full, developer)

### Technical Requirements
1. **Windows Support**:
   - Use WinGet package manager: `winget install --id GitHub.cli`
   - Alternative: MSI installer from GitHub releases
   - Chocolatey fallback: `choco install gh`

2. **Linux Support**:
   - Ubuntu/Debian: Use official GitHub APT repository
   - CentOS/RHEL: Use official GitHub RPM repository
   - Generic: Direct binary download from GitHub releases

### Installation Profiles Integration
- **full**: Include GitHub CLI for complete development environment
- **developer** (new profile): GitHub CLI + Git + essential dev tools
- **minimal**: Exclude GitHub CLI
- **default**: Include GitHub CLI for standard development workflow

## Proposed Solution

### Package Definition Structure
```json
{
  "name": "github-cli",
  "displayName": "GitHub CLI",
  "description": "Official GitHub command line tool",
  "category": "Development Tools",
  "homepage": "https://cli.github.com/",
  "platforms": {
    "windows": {
      "installers": [
        {
          "type": "winget",
          "package": "GitHub.cli",
          "priority": 1
        },
        {
          "type": "chocolatey",
          "package": "gh",
          "priority": 2
        },
        {
          "type": "msi",
          "url": "https://github.com/cli/cli/releases/latest/download/gh_{version}_windows_amd64.msi",
          "priority": 3
        }
      ]
    },
    "linux": {
      "installers": [
        {
          "type": "apt",
          "commands": [
            "curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg",
            "echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null",
            "sudo apt update",
            "sudo apt install gh"
          ],
          "distros": ["ubuntu", "debian"],
          "priority": 1
        },
        {
          "type": "yum",
          "commands": [
            "sudo dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo",
            "sudo dnf install gh"
          ],
          "distros": ["fedora", "centos", "rhel"],
          "priority": 1
        },
        {
          "type": "tar.gz",
          "url": "https://github.com/cli/cli/releases/latest/download/gh_{version}_linux_amd64.tar.gz",
          "extract_to": "/usr/local/bin",
          "priority": 2
        }
      ]
    }
  },
  "verification": {
    "command": "gh --version",
    "expected_pattern": "gh version \\d+\\.\\d+\\.\\d+"
  },
  "profiles": ["full", "developer", "default"]
}
```

## Implementation Plan

### Phase 1: Basic Integration
1. Add GitHub CLI package definition to `assets/install-packages.json`
2. Implement basic Windows installation (WinGet)
3. Implement basic Linux installation (APT for Ubuntu/Debian)
4. Add verification command
5. Update installation profiles

### Phase 2: Advanced Features
1. Add support for more Linux distributions (RPM-based)
2. Implement fallback installation methods
3. Add post-installation configuration guidance
4. Integration with existing GitHub workflows

### Phase 3: Enhancement
1. Add GitHub authentication setup wizard
2. Integration with Portunix's potential GitHub features
3. Configuration management for GitHub CLI

## Acceptance Criteria

- [ ] GitHub CLI can be installed on Windows via `portunix install github-cli`
- [ ] GitHub CLI can be installed on Linux via `portunix install github-cli`
- [ ] Installation is included in `full` and `default` profiles
- [ ] Installation verification works correctly
- [ ] Package supports latest stable version
- [ ] Cross-platform compatibility maintained
- [ ] Documentation updated

## Testing Strategy

### Manual Testing
1. Test installation on Windows 10/11
2. Test installation on Ubuntu 20.04/22.04
3. Test installation through different package managers
4. Verify `gh --version` command works
5. Test integration with installation profiles

### Automated Testing
1. Add unit tests for package definition parsing
2. Add integration tests for installation verification
3. Test fallback mechanisms

## Documentation Updates

1. Update `docs/FEATURES_OVERVIEW.md` with GitHub CLI support
2. Add usage examples to installation documentation
3. Update package list in README files

## Dependencies

- No new dependencies required
- Uses existing Portunix package management system
- Requires internet connection for installation

## Related Issues

- #025: GitHub Integration for Portunix Core (complementary)
- #021: GitHub Actions Local Testing Support (potential integration)

## Notes

- GitHub CLI requires authentication for most operations
- Consider adding post-installation setup guidance
- May integrate with future GitHub-related Portunix features
- Installation should be optional but recommended for developers

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented:** 2025-01-09  
**Version:** v1.5.7+  

### Implementation Summary

GitHub CLI (gh) installation support has been successfully implemented:

#### ✅ Package Installation
- **`portunix install gh`** command available 
- **Cross-platform support** for Windows, Linux, macOS
- **Integration with package management system** via install-packages.json
- **Automatic download** from GitHub releases

#### ✅ Features Implemented
- **Universal installation command** across all supported platforms
- **Version management** with automatic latest version detection
- **Integration with existing CLI patterns** consistent with other Portunix packages
- **Installation verification** to ensure proper setup

### Verification Commands

```bash
# Install GitHub CLI
portunix install gh

# Verify installation  
gh --version

# Basic usage examples
gh auth login
gh repo list
gh issue list
gh pr create
```

### Usage Integration

GitHub CLI integrates seamlessly with Portunix development workflows:

```bash
# Install development environment with GitHub CLI
portunix install default gh

# Use in container environments
portunix docker run-in-container default
# Inside container: portunix install gh
```

### Resolution Status

**FULLY IMPLEMENTED** ✅ - GitHub CLI (gh) is now available through the Portunix package management system with full cross-platform support and seamless integration.

---

**Estimated Implementation Time**: 4-6 hours  
**Complexity**: Low-Medium  
**Client Impact**: High (direct requirement satisfaction)