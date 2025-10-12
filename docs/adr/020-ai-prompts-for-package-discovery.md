# ADR-020: AI Prompts for Package Discovery and Maintenance

## Context

The `assets/install-packages.json` file contains package definitions with installation procedures, but currently lacks structured information for AI assistants to help with package discovery, version updates, and variant management. When developers need to add new packages or update existing ones, AI assistants require specific context about:

- How to research and find information about specific package variants
- Where to look for the latest version information for a particular package
- How to understand the package ecosystem and available installation methods
- What sources to consult when adding new variants or versions

Currently, while we have `installationDocsUrl` and `latestVersionUrl` fields (ADR-019), these are designed for human consumption and don't provide the structured guidance that AI assistants need to effectively research and maintain package definitions.

## Decision

We will extend the `assets/install-packages.json` structure to include AI-specific prompts that guide AI assistants in package research and maintenance tasks. These prompts will be added at both the package level and variant level.

### New Fields Structure:

```json
{
  "packages": {
    "package-name": {
      "name": "Package Display Name",
      "description": "Package description",
      "installationDocsUrl": "...",
      "latestVersionUrl": "...",
      "aiPrompts": {
        "packageResearch": "Detailed prompt for AI to research this package, find latest versions, understand installation methods, and identify available variants",
        "versionDiscovery": "Specific instructions for AI on how to find and verify the latest version information for this package",
        "variantDiscovery": "Guidelines for AI on how to discover and research new variants of this package (different versions, architectures, etc.)"
      },
      "platforms": {
        "windows": {
          "variants": {
            "variant-name": {
              "aiPrompts": {
                "variantResearch": "Specific prompt for AI to research this particular variant, including how to find download URLs, version info, and installation specifics"
              }
            }
          }
        }
      }
    }
  }
}
```

### Example Implementation:

```json
{
  "packages": {
    "java": {
      "name": "Java (OpenJDK)",
      "description": "Java Development Kit from Eclipse Adoptium",
      "installationDocsUrl": "https://adoptium.net/temurin/releases/",
      "latestVersionUrl": "https://api.github.com/repos/adoptium/temurin21-binaries/releases/latest",
      "aiPrompts": {
        "packageResearch": "Research Eclipse Adoptium Temurin OpenJDK releases. Focus on LTS versions (8, 11, 17, 21). Check GitHub releases at adoptium/temurin*-binaries repositories. Look for MSI installers for Windows and appropriate packages for Linux distributions. Verify version numbering format and download URL patterns.",
        "versionDiscovery": "Check GitHub API for latest releases in adoptium/temurin{8,11,17,21}-binaries repositories. Parse release tags to extract version numbers. Verify that download assets include both x64 and x86 MSI files for Windows, and appropriate Linux packages.",
        "variantDiscovery": "Research available JDK versions from Eclipse Adoptium. Focus on LTS releases (8, 11, 17, 21). For each version, identify the latest update release, check download availability for multiple architectures (x64, x86, arm64), and verify installation paths follow Eclipse Adoptium conventions."
      },
      "platforms": {
        "windows": {
          "variants": {
            "21": {
              "aiPrompts": {
                "variantResearch": "Research OpenJDK 21 LTS from Eclipse Adoptium. Check adoptium/temurin21-binaries releases on GitHub. Look for latest JDK 21.x.x releases with MSI installers for x64 and x86. Verify download URLs follow pattern: OpenJDK21U-jdk_{arch}_windows_hotspot_{version}.msi. Check installation path format: jdk-21.x.x.x-hotspot."
              }
            }
          }
        }
      }
    },
    "nodejs": {
      "name": "Node.js",
      "description": "JavaScript runtime built on Chrome's V8 JavaScript engine",
      "installationDocsUrl": "https://nodejs.org/en/download/package-manager",
      "latestVersionUrl": "https://nodejs.org/dist/latest/",
      "aiPrompts": {
        "packageResearch": "Research Node.js installation methods. Check official Node.js download page and package manager documentation. Focus on LTS versions, verify installation methods for Windows (MSI, Chocolatey, winget) and Linux (package managers, direct downloads). Check version naming conventions and release patterns.",
        "versionDiscovery": "Check Node.js official releases at https://nodejs.org/dist/ and GitHub nodejs/node releases. Identify current LTS version and latest stable version. Verify download availability for different platforms and architectures.",
        "variantDiscovery": "Research Node.js release lines. Focus on LTS versions (even-numbered major versions). For each LTS line, identify the latest patch version. Check availability of MSI installers, ZIP packages, and package manager support."
      }
    }
  }
}
```

## Implementation Guidelines

### Prompt Writing Principles:
1. **Specific and Actionable** - Prompts should provide clear, step-by-step guidance
2. **Source-Focused** - Always specify where to look for information
3. **Context-Aware** - Include package-specific knowledge and conventions
4. **Maintenance-Oriented** - Focus on what AI needs to know to maintain and update packages
5. **Discovery-Enabled** - Help AI understand how to find new variants and versions

### Prompt Categories:

- **`packageResearch`** - High-level guidance for understanding the entire package ecosystem
- **`versionDiscovery`** - Specific instructions for finding latest version information
- **`variantDiscovery`** - Guidelines for discovering new variants (versions, architectures, installation methods)
- **`variantResearch`** - Detailed instructions for researching specific variants

### Required Information in Prompts:
- Official sources and repositories to consult
- URL patterns and naming conventions
- Version numbering schemes
- Architecture and platform considerations
- Installation path conventions
- Common gotchas and special considerations

## Consequences

### Positive:
- **Enhanced AI Assistance** - AI assistants can more effectively help with package maintenance
- **Reduced Research Time** - Structured guidance reduces time spent figuring out where to look for information
- **Consistency** - Standardized approach to package research across all packages
- **Knowledge Preservation** - Captures institutional knowledge about package ecosystems
- **Scalability** - New team members (and AI) can quickly understand package maintenance procedures
- **Quality Improvement** - Better research leads to more accurate and up-to-date package definitions

### Negative:
- **Maintenance Overhead** - Prompts need to be updated when package ecosystems change
- **Initial Effort** - Existing packages need to be updated with AI prompts
- **Prompt Accuracy** - Poorly written prompts could mislead AI assistants
- **Storage Overhead** - Increased file size due to additional metadata

### Neutral:
- **Optional Implementation** - Prompts are metadata only, don't affect installation functionality
- **Gradual Adoption** - Can be implemented progressively for different packages
- **Flexible Format** - Prompt structure can evolve based on practical experience

## Migration Strategy

1. **Phase 1** - Define prompt structure and create templates
2. **Phase 2** - Add prompts to most critical packages (Java, Node.js, Python)
3. **Phase 3** - Gradually extend to all existing packages
4. **Phase 4** - Make AI prompts mandatory for new package additions

## Success Metrics

- Reduced time for AI-assisted package research and updates
- Improved accuracy of new package additions
- Faster resolution of package maintenance issues
- Better consistency in package definition quality