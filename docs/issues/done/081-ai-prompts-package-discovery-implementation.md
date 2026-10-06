# Issue #81: AI Prompts for Package Discovery Implementation

**Status:** ✅ Implemented
**Implemented:** 2026-05-19

## Summary
Implement AI prompts in `assets/install-packages.json` to guide AI assistants in package research, version discovery, and maintenance tasks as defined in ADR-020.

## Description
Following ADR-020, we need to extend the package definition structure to include AI-specific prompts that help AI assistants effectively research and maintain packages. This will provide structured guidance for package discovery, version updates, and variant management.

## Implementation Requirements

### 1. JSON Schema Extension
Extend `assets/install-packages.json` structure with new `aiPrompts` fields:

```json
{
  "packages": {
    "package-name": {
      "aiPrompts": {
        "packageResearch": "string",
        "versionDiscovery": "string",
        "variantDiscovery": "string"
      },
      "platforms": {
        "windows": {
          "variants": {
            "variant-name": {
              "aiPrompts": {
                "variantResearch": "string"
              }
            }
          }
        }
      }
    }
  }
}
```

### 2. Critical Packages to Implement First
Implement AI prompts for these high-priority packages:

1. **Java (OpenJDK)** - Complex multi-variant package with LTS versions
2. **Node.js** - Popular runtime with multiple installation methods
3. **Python** - Core dependency with multiple variants
4. **Visual Studio Code** - Developer tool with multiple editions
5. **Claude Code** - AI tool requiring specific research approaches

### 3. Prompt Quality Requirements

Each prompt must include:
- **Specific sources**: Exact URLs and repositories to check
- **Version patterns**: How to identify and parse version numbers
- **URL patterns**: Download URL conventions and naming patterns
- **Architecture support**: What architectures are available
- **Installation paths**: Expected installation location patterns
- **Common gotchas**: Known issues or special considerations

### 4. Implementation Phases

#### Phase 1: Core Structure & Templates
- [ ] Define final JSON schema structure
- [ ] Create prompt writing guidelines and templates
- [ ] Document prompt quality standards

#### Phase 2: Critical Packages
- [ ] Implement AI prompts for Java (OpenJDK)
- [ ] Implement AI prompts for Node.js
- [ ] Implement AI prompts for Python
- [ ] Test prompts with AI assistant integration

#### Phase 3: Extended Coverage
- [ ] Implement AI prompts for Visual Studio Code
- [ ] Implement AI prompts for Claude Code
- [ ] Implement AI prompts for remaining packages in install-packages.json

#### Phase 4: Validation & Integration
- [ ] Test AI prompt effectiveness with real research scenarios
- [ ] Validate prompt accuracy against current package information
- [ ] Document best practices for writing new prompts

## Acceptance Criteria

### Functional Requirements
- [ ] `assets/install-packages.json` supports new `aiPrompts` structure
- [ ] All existing package installation functionality remains unchanged
- [ ] AI prompts are available at both package and variant levels
- [ ] Schema validation supports optional AI prompts (backward compatibility)

### Quality Requirements
- [ ] AI prompts provide actionable, specific guidance
- [ ] Prompts include all necessary source URLs and patterns
- [ ] Each prompt follows established writing guidelines
- [ ] Prompts are validated against current package ecosystems

### Testing Requirements
- [ ] JSON schema validation tests for new structure
- [ ] Backward compatibility tests (old packages without prompts)
- [ ] AI prompt effectiveness testing with real scenarios
- [ ] Documentation examples work as described

## Examples

### Java Package Prompts
```json
{
  "java": {
    "aiPrompts": {
      "packageResearch": "Research Eclipse Adoptium Temurin OpenJDK releases. Focus on LTS versions (8, 11, 17, 21). Check GitHub releases at adoptium/temurin*-binaries repositories. Look for MSI installers for Windows and appropriate packages for Linux distributions.",
      "versionDiscovery": "Check GitHub API for latest releases in adoptium/temurin{8,11,17,21}-binaries repositories. Parse release tags to extract version numbers. Verify download assets include both x64 and x86 MSI files.",
      "variantDiscovery": "Research available JDK versions from Eclipse Adoptium. Focus on LTS releases (8, 11, 17, 21). For each version, identify latest update release and verify installation paths follow Eclipse Adoptium conventions."
    },
    "platforms": {
      "windows": {
        "variants": {
          "21": {
            "aiPrompts": {
              "variantResearch": "Research OpenJDK 21 LTS from Eclipse Adoptium. Check adoptium/temurin21-binaries releases on GitHub. Look for latest JDK 21.x.x releases with MSI installers for x64 and x86. Verify download URLs follow pattern: OpenJDK21U-jdk_{arch}_windows_hotspot_{version}.msi."
            }
          }
        }
      }
    }
  }
}
```

## Technical Considerations

### Backward Compatibility
- AI prompts are optional fields
- Existing package installation logic unchanged
- Schema validation must support packages without AI prompts

### Maintenance Strategy
- AI prompts require updates when package ecosystems change
- Regular validation against current package information
- Version control tracking of prompt changes

### Performance Impact
- AI prompts are metadata only
- No runtime performance impact on installation
- Minimal storage overhead

## Dependencies
- No external dependencies
- Based on ADR-020 architectural decisions
- Compatible with existing package management system

## Estimated Effort
- **Phase 1**: 1-2 days (schema + guidelines)
- **Phase 2**: 3-4 days (critical packages implementation)
- **Phase 3**: 2-3 days (extended coverage)
- **Phase 4**: 1-2 days (validation + documentation)
- **Total**: 7-11 days

## Success Metrics
- AI assistants can effectively research packages using prompts
- Reduced time for package maintenance and updates
- Improved accuracy of new package additions
- Consistent quality across package definitions

## Related Issues
- Related to package management system improvements
- Supports AI-assisted development workflows
- Enhances maintainability of install-packages.json

## Implementation Notes
- Start with most complex packages (Java) to validate approach
- Test prompts with actual AI assistant interactions
- Document lessons learned for future prompt writing
- Consider automation for prompt validation against live sources

---

**Created**: 2025-01-27
**Implemented**: 2026-05-19
**Priority**: Medium
**Type**: Enhancement
**Labels**: enhancement, package-management, ai-integration, metadata, maintenance
**Estimated Effort**: 7-11 days

## Implementation Notes

Schema and structure evolved between issue creation and implementation:

- **Location changed**: monolithic `assets/install-packages.json` → per-package files
  in `src/helpers/ptx-installer/assets/packages/*.json`
- **Schema simplified**: original 4-key proposal (`packageResearch`,
  `versionDiscovery`, `variantDiscovery`, `variantResearch`) consolidated to
  the 3-key form actually used across all 66 packages:
  `versionDiscovery`, `urlResolution`, `updateGuidance` (under `spec.aiPrompts`)
- **Coverage**: by the time this issue was picked up, 63/66 packages already
  had `aiPrompts`. This implementation completed the remaining 3:
  `vox-deps`, `vox-model-tts-czech`, `vox-model-tts-english`
- **No code changes**: `aiPrompts` is metadata only; `ptx-installer package
  info` already surfaces "🤖 AI Integration: Available" automatically