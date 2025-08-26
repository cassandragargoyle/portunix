# Issue #12: PowerShell Installation Support for Linux Distributions

## Summary
Add automatic PowerShell installation capability to Portunix for multiple Linux distributions to support IT support workers who frequently need to run PowerShell scripts across different platforms.

## Problem Description
IT support workers often need to execute PowerShell scripts on both Windows and Linux systems. While Windows has PowerShell pre-installed, Linux distributions require manual installation, which is complex due to the variety of distributions, versions, and different installation methods.

## Target Distributions
The following distributions should be supported (based on frequent usage):
- Ubuntu / Kubuntu
- Fedora  
- Debian
- Mint
- Elementary OS
- Rocky Linux

## Proposed Solution
Integrate PowerShell installation into the existing Portunix package management system:

1. **Detection System**: Leverage existing OS detection to identify Linux distribution and version
2. **Installation Methods**: Use appropriate installation method per distribution:
   - APT-based (Ubuntu, Debian, Mint, Elementary): Microsoft APT repository
   - DNF/YUM-based (Fedora, Rocky): Microsoft RPM repository
   - Snap packages as fallback option
3. **Integration**: Extend `assets/install-packages.json` with PowerShell installation definitions
4. **Command Interface**: Add PowerShell to available packages in Portunix install commands

## Technical Requirements

### Package Definition Structure
```json
{
  "powershell": {
    "name": "PowerShell",
    "description": "Cross-platform PowerShell scripting environment",
    "category": "development",
    "platforms": {
      "linux": {
        "ubuntu": {
          "method": "apt",
          "repository": "https://packages.microsoft.com/repos/microsoft-ubuntu-22.04-prod",
          "key": "https://packages.microsoft.com/keys/microsoft.asc",
          "package": "powershell"
        },
        "debian": {
          "method": "apt", 
          "repository": "https://packages.microsoft.com/repos/microsoft-debian-11-prod",
          "key": "https://packages.microsoft.com/keys/microsoft.asc",
          "package": "powershell"
        },
        "fedora": {
          "method": "dnf",
          "repository": "https://packages.microsoft.com/rhel/8/prod/",
          "key": "https://packages.microsoft.com/keys/microsoft.asc",
          "package": "powershell"
        }
      }
    }
  }
}
```

### Command Examples
```bash
# Install PowerShell
portunix install powershell

# Check if PowerShell is available
portunix status powershell

# Remove PowerShell
portunix remove powershell
```

## Implementation Phases

### Phase 1: Core Distribution Support
- Ubuntu/Debian (APT-based systems)
- Repository setup and GPG key management
- Basic installation verification

### Phase 2: Additional Distributions  
- Fedora/Rocky (DNF/YUM-based systems)
- Mint, Elementary OS, Kubuntu variants
- Snap package fallback

### Phase 3: Advanced Features
- Version management (install specific PowerShell version)
- Automatic updates integration
- Configuration management

## Benefits
- **Simplified IT Support**: One command installs PowerShell across multiple Linux distributions
- **Consistency**: Standardized installation process regardless of underlying Linux distribution
- **Integration**: PowerShell becomes part of Portunix managed environment
- **Maintenance**: Automatic updates and management through Portunix system

## Acceptance Criteria
- [ ] PowerShell installs successfully on Ubuntu 20.04, 22.04, 24.04
- [ ] PowerShell installs successfully on Debian 11, 12
- [ ] PowerShell installs successfully on Fedora 38, 39, 40
- [ ] PowerShell installs successfully on Rocky Linux 8, 9
- [ ] Installation verification confirms PowerShell is executable
- [ ] Uninstall process removes PowerShell cleanly
- [ ] Error handling for unsupported distributions
- [ ] Integration with existing Portunix package management

## Priority: High
This addresses a frequent pain point for IT support workers and leverages existing Portunix capabilities.

## Labels
enhancement, powershell, linux, cross-platform, package-management, it-support

---
*Created: 2025-08-23*
*Status: 📋 Open*