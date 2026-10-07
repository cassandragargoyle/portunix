# Issue #78: GitHub CLI Installation Support

## Status
✅ Implemented

## Priority
Medium

## Component
Installation System

## Description
Add support for installing GitHub CLI (gh) tool through Portunix package management system. The tool should be available under a more descriptive name than just "gh" to improve discoverability and clarity for users.

## Background
GitHub CLI (gh) is an essential tool for developers working with GitHub repositories. It provides command-line access to GitHub features including:
- Repository management
- Pull request operations
- Issue management
- GitHub Actions workflow control
- Release management

Currently, Portunix doesn't support installation of GitHub CLI, requiring users to install it manually through other means.

## Requirements

### Functional Requirements
1. Add GitHub CLI to `assets/install-packages.json` with proper configuration
2. Use descriptive name like `github-cli` or `gh-cli` instead of just `gh`
3. Support installation on multiple platforms:
   - Windows (via WinGet, Chocolatey, or direct download)
   - Linux (via package managers or direct download)
   - macOS (via Homebrew or direct download)
4. Ensure proper PATH configuration after installation
5. Verify installation with `gh --version`

### Package Naming
- Primary name in Portunix: `github-cli`
- Alternative acceptable names: `gh-cli`, `github`
- Binary name remains: `gh` (as per official tool)

### Installation Methods

#### Windows
- **Primary**: WinGet (`winget install GitHub.cli`)
- **Secondary**: Chocolatey (`choco install gh`)
- **Fallback**: Direct download from GitHub releases

#### Linux
- **Debian/Ubuntu**: APT repository or .deb package
- **RHEL/Fedora**: YUM/DNF repository or .rpm package
- **Universal**: Direct binary download from GitHub releases

#### macOS
- **Primary**: Homebrew (`brew install gh`)
- **Fallback**: Direct download from GitHub releases

## Implementation Plan

### Phase 1: Package Definition
1. Add entry to `assets/install-packages.json`
2. Define installation methods for each platform
3. Set up version detection and verification

### Phase 2: Testing
1. Test installation on Windows
2. Test installation on Linux (Ubuntu/Debian)
3. Test installation on Linux (RHEL/Fedora)
4. Verify PATH configuration
5. Test `gh auth login` functionality

### Phase 3: Documentation
1. Update package list documentation
2. Add usage examples
3. Document authentication setup

## Example Package Configuration

```json
{
  "github-cli": {
    "description": "GitHub CLI - Official command line tool for GitHub",
    "version": "latest",
    "binary": "gh",
    "windows": {
      "primary": {
        "type": "winget",
        "package": "GitHub.cli"
      },
      "fallback": {
        "type": "chocolatey",
        "package": "gh"
      }
    },
    "linux": {
      "debian": {
        "type": "apt",
        "repository": "https://cli.github.com/packages",
        "package": "gh"
      },
      "rhel": {
        "type": "yum",
        "repository": "https://cli.github.com/packages/rpm",
        "package": "gh"
      },
      "universal": {
        "type": "binary",
        "url": "https://github.com/cli/cli/releases/latest",
        "extract": true
      }
    }
  }
}
```

## Testing Scenarios

### Installation Test
```bash
# Install GitHub CLI
portunix install github-cli

# Verify installation
gh --version

# Test basic functionality
gh auth status
```

### Cross-Platform Test
- Windows 10/11 with WinGet
- Windows with Chocolatey only
- Ubuntu 22.04/24.04
- Fedora latest
- Alpine Linux (container)

## Success Criteria
- [ ] GitHub CLI can be installed via `portunix install github-cli`
- [ ] Installation works on Windows, Linux, and macOS
- [ ] `gh` command is available in PATH after installation
- [ ] Version verification works correctly
- [ ] Installation is documented in package list

## Notes
- Consider adding common gh extensions as separate packages in future
- May need to handle authentication setup instructions post-install
- Should coordinate with MCP integration for AI assistants using gh

## Related Issues
- None

## References
- [GitHub CLI Official Site](https://cli.github.com/)
- [GitHub CLI Installation Docs](https://github.com/cli/cli#installation)
- [GitHub CLI Releases](https://github.com/cli/cli/releases)

---

**Created**: 2025-09-27
**Author**: Development Team
**Assigned**: TBD