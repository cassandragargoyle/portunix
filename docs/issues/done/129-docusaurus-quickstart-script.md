# Issue #129: Docusaurus QuickStart Script for GitHub Release

## Summary

Create a production-ready PowerShell script that will be included in GitHub releases as a sample/quickstart for setting up a Docusaurus documentation environment using Portunix. The script should maximize user experience and Portunix product utilization.

## Problem Statement

Users who want to create static documentation sites using Docusaurus need to:

1. Install Portunix manually
2. Understand playbook system
3. Configure environment manually

This creates friction for new users and doesn't showcase Portunix capabilities effectively.

## Proposed Solution

Create `release-assets/quickstart/quickstart-docusaurus.ps1` that:

### 1. Self-Documenting Header

Script begins with clear instructions on how to download and run it on Windows:
```powershell
<#
.SYNOPSIS
    Portunix Docusaurus QuickStart - Create documentation site in minutes

.DESCRIPTION
    This script sets up a complete Docusaurus documentation environment using Portunix.

.QUICK START
    # Option 1: Direct execution (recommended)
    irm https://github.com/cassandragargoyle/portunix/releases/latest/download/quickstart-docusaurus.ps1 | iex

    # Option 2: Download and run
    Invoke-WebRequest -Uri "https://github.com/cassandragargoyle/portunix/releases/latest/download/quickstart-docusaurus.ps1" -OutFile "quickstart-docusaurus.ps1"
    .\quickstart-docusaurus.ps1

.NOTES
    Requires: Windows 10/11 (Docker/Podman will be installed automatically if needed)
    Author: CassandraGargoyle Team
    Version: 1.0.0
#>
```

### 2. Core Functionality

1. **Check/Install Portunix**
   - Detect if Portunix is installed
   - If not, offer to install it automatically
   - If installed, check for updates and offer upgrade

2. **Interactive Configuration**
   - Ask user for documentation project path (with sensible default like `./my-docs`)
   - Ask for project name
   - Ask for target environment (container recommended, local as fallback)

3. **Container Runtime Check**
   - Verify Docker or Podman is available
   - Install automatically if missing (via `portunix install docker` or `portunix install podman`)

4. **Create Playbook**
   - Use `portunix playbook init` with docusaurus template
   - Apply user's configuration

5. **Initialize Project**
   - Run `portunix playbook run --script create`
   - Show progress and helpful messages

6. **Post-Setup Instructions**
   - Display how to start development server
   - Show available commands
   - Link to documentation

### 3. User Experience Features

- Colored output with clear status indicators
- Progress messages during long operations
- Error handling with helpful recovery suggestions
- Summary of what was created at the end
- Optional: Open browser with dev server URL

## File Structure

```
release-assets/
└── quickstart/
    ├── quickstart-docusaurus.ps1    # Main script (this issue)
    └── quickstart-hugo.ps1          # Future: Hugo equivalent
```

## Reference Implementation

Based on `test/manual/tc010-windows-docusaurus-test.ps1` but production-ready:
- Remove test-specific logic
- Add interactive prompts
- Add error recovery
- Add update checking
- Polish output messages

## Acceptance Criteria

1. [ ] Script can be downloaded and executed directly from GitHub release
2. [ ] Script installs/updates Portunix if needed
3. [ ] Script asks for project path interactively
4. [ ] Script creates functional Docusaurus environment
5. [ ] Script provides clear next-step instructions
6. [ ] Script handles errors gracefully with recovery suggestions
7. [ ] Script works on Windows 10/11 with PowerShell 5.1+
8. [ ] Script is included in GitHub release assets

## Implementation Notes

- Use Portunix native commands wherever possible
- Avoid hardcoded paths - use Portunix defaults
- Consider adding `--non-interactive` flag for CI/CD usage
- Script should be idempotent (safe to run multiple times)

## Release Integration

Add to `scripts/make-release.py` or `.goreleaser.yml`:
- Include `scripts/quickstart-docusaurus.ps1` in release assets
- Ensure script version matches release version

## Priority

Medium - Nice to have for v1.9.x release

## Labels

enhancement, documentation, user-experience, quickstart, docusaurus, release-assets

## Related Issues

- #128 Docusaurus Container Performance Optimization (provides underlying functionality)
- #119 PTX-Ansible Standalone Help and Template Examples

## Notes

- Consider creating Linux/macOS equivalent (`quickstart-docusaurus.sh`) in future issue
- Could serve as template for other quickstart scripts (Hugo, Jekyll, etc.)
