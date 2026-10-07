# Issue #209: Google Chrome Installation Implementation

> **Renumbered:** formerly internal issue #022. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Status:** ✅ Implemented  
**Priority:** Medium  
**Type:** Feature Enhancement  
**Component:** Package Management  
**Created:** 2025-09-01  
**Target Version:** v1.5.5  

## Summary

Implement comprehensive Google Chrome installation support across all platforms (Windows, Linux, macOS) with intelligent variant selection and proper system integration.

## Current State

Google Chrome is currently listed in `assets/install-packages.json` but has limited implementation:
- Only Windows MSI installer is defined
- No Linux support (despite multiple distribution options)
- No macOS support
- Missing automated updates and proper verification

## Requirements

### Functional Requirements

1. **Cross-platform Support**
   - Windows: MSI installer (stable, beta, dev, canary channels)
   - Linux: Repository-based installation for all major distributions
   - macOS: DMG installer support

2. **Installation Variants**
   - `stable` - Stable release channel (default)
   - `beta` - Beta testing channel
   - `dev` - Developer channel
   - `canary` - Bleeding edge channel

3. **Linux Distribution Support**
   - Ubuntu/Debian: Official Google APT repository
   - Fedora/RHEL/Rocky: Official Google YUM/DNF repository
   - openSUSE: Zypper repository
   - Arch Linux: AUR support or direct download
   - Universal: Snap package fallback

4. **Post-installation Tasks**
   - Set as default browser (optional)
   - Import bookmarks from existing browsers (optional)
   - Configure enterprise policies (for managed environments)
   - Verify installation integrity

### Technical Requirements

1. **Repository Management**
   - Add Google's official repository keys
   - Configure repository sources
   - Handle GPG key verification

2. **Version Detection**
   - Check for existing Chrome installation
   - Determine installed version
   - Support side-by-side channel installations

3. **Error Handling**
   - Network connectivity issues
   - Repository availability
   - Disk space verification
   - Permission requirements

## Implementation Plan

### Phase 1: Linux Implementation
```json
"chrome": {
  "platforms": {
    "linux": {
      "type": "repository",
      "variants": {
        "stable": {
          "distributions": ["ubuntu", "debian", "mint", "elementary"],
          "repository_setup": [
            "wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -",
            "sudo sh -c 'echo \"deb [arch=amd64] https://dl.google.com/linux/chrome/deb/ stable main\" > /etc/apt/sources.list.d/google-chrome.list'",
            "sudo apt-get update"
          ],
          "packages": ["google-chrome-stable"]
        },
        "fedora-stable": {
          "distributions": ["fedora", "rocky", "centos"],
          "repository_setup": [
            "sudo dnf install -y fedora-workstation-repositories",
            "sudo dnf config-manager --set-enabled google-chrome",
            "sudo rpm --import https://dl.google.com/linux/linux_signing_key.pub"
          ],
          "packages": ["google-chrome-stable"]
        }
      }
    }
  }
}
```

### Phase 2: Windows Enhancement
- Add beta, dev, and canary channels
- Implement silent installation options
- Add registry configuration for enterprise

### Phase 3: macOS Support
- DMG download and mounting
- Application folder installation
- Dock integration

### Phase 4: Advanced Features
- Profile migration utilities
- Extension pre-installation
- Managed browser policies
- Chromium as open-source alternative

## Testing Requirements

1. **Installation Testing**
   - Test on Ubuntu 22.04, 24.04
   - Test on Fedora 39, 40
   - Test on Windows 10, 11
   - Test on macOS 13, 14

2. **Functionality Testing**
   - Verify browser launches correctly
   - Test basic browsing functionality
   - Verify automatic updates work
   - Test profile creation

3. **Integration Testing**
   - Test with existing browser installations
   - Verify coexistence with other browsers
   - Test system integration (file associations, protocols)

## Success Criteria

- [ ] Chrome installs successfully on all supported platforms
- [ ] Repository configuration persists across system updates
- [ ] Installation completes in under 2 minutes on broadband
- [ ] Verification command (`google-chrome --version`) works
- [ ] Proper error messages for unsupported systems
- [ ] Documentation updated with usage examples

## Related Issues

- #021 - GitHub Actions Local Testing (uses Chrome for testing)
- #012 - PowerShell Installation (similar repository pattern)

## Notes

- Google Chrome requires acceptance of Terms of Service
- Consider privacy implications and offer Chromium as alternative
- Enterprise deployment may require additional configuration
- Some Linux distributions may have trademark restrictions

## Resources

- [Google Chrome Linux Repository](https://www.google.com/linuxrepositories/)
- [Chrome Enterprise Documentation](https://chromeenterprise.google/)
- [Chromium Project](https://www.chromium.org/)

---

**Assigned to:** Development Team  
**Labels:** `enhancement`, `package-management`, `cross-platform`