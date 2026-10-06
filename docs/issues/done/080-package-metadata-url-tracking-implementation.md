# Issue #80: Package Metadata URL Tracking Implementation

**Type:** Enhancement
**Priority:** Medium
**Status:** ✅ Implemented
**Labels:** enhancement, package-management, metadata, documentation, maintenance
**Created:** 2025-09-27
**Closed:** 2026-05-11
**Estimated Effort:** 2 hours

## Description

Implement the package metadata URL tracking system as defined in ADR-019. This involves adding mandatory URL fields to all package definitions in `assets/install-packages.json` to track the official installation documentation and version check URLs for each package.

## Background

As per ADR-019, we need to maintain traceability of installation procedures by tracking:
- Official installation documentation from software authors
- URLs for determining the latest versions
- Sources from which our installation procedures were derived

## Requirements

### 1. JSON Schema Update
- Add two mandatory fields to each package definition:
  - `installationDocsUrl` - URL to official installation documentation (required)
  - `latestVersionUrl` - URL for determining latest version (required)
- Both fields must be present even if they point to the same location

### 2. Package Definitions Update
Update all existing packages in `assets/install-packages.json` with both URLs:

#### Priority 1 - Core Development Tools
- [ ] Java (OpenJDK) - all versions
- [ ] Python
- [ ] Go
- [ ] Node.js
- [ ] PowerShell

#### Priority 2 - Development Tools
- [ ] Visual Studio Code
- [ ] Apache Maven
- [ ] Claude Code
- [ ] GitHub CLI
- [ ] Hugo

#### Priority 3 - Package Managers
- [ ] Chocolatey
- [ ] WinGet
- [ ] pipx

#### Priority 4 - Browsers and Other Tools
- [ ] Google Chrome
- [ ] Ansible
- [ ] Act (GitHub Actions)

### 3. Validation
- Ensure all URLs are valid and accessible
- Verify that installation documentation URLs point to official sources
- Check that version URLs provide current version information

## Example Implementation

```json
{
  "name": "nodejs",
  "description": "JavaScript runtime built on Chrome's V8 JavaScript engine",
  "installationDocsUrl": "https://nodejs.org/en/download/package-manager",
  "latestVersionUrl": "https://nodejs.org/dist/latest/",
  "windows": {
    // ... existing configuration
  },
  "linux": {
    // ... existing configuration
  }
}
```

## Acceptance Criteria

1. [ ] All packages in `assets/install-packages.json` have both URL fields populated
2. [ ] URLs are valid and point to official sources
3. [ ] Documentation URLs lead to installation instructions
4. [ ] Version URLs provide way to determine latest version
5. [ ] No functional changes to installation process
6. [ ] JSON structure remains backward compatible (only adds new fields)
7. [ ] Lightweight CI guard (jq-based) ensures all future package manifests carry both URL fields — added during testing review to prevent regression. URL reachability checks are deliberately NOT in CI (rate-limit, anti-bot, flakiness) and remain an ad-hoc manual audit.

## Testing

1. Validate JSON structure after changes
2. Verify all URLs are accessible (manual check or simple script)
3. Ensure installation functionality remains unchanged
4. Test that Portunix can still parse and use the updated JSON

## Implementation Notes

- Start with high-priority packages first
- Use official documentation sites, not third-party sources
- For packages from GitHub, use releases API for version URL
- For packages from specific registries (npm, pip, etc.), use their API endpoints
- Document any packages where official URLs cannot be found

## References

- ADR-019: Package Metadata URL Tracking (`docs/adr/019-package-metadata-url-tracking.md`)
- Current package definitions: `assets/install-packages.json`

## Migration Path

1. Create backup of current `assets/install-packages.json`
2. Add URL fields to each package definition
3. Validate JSON structure
4. Test installation of a few packages to ensure no regression
5. Commit changes with reference to this issue and ADR-019