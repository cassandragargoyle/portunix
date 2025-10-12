# ADR-019: Package Metadata URL Tracking

## Context

Portunix installs various software packages defined in the `assets/install-packages.json` file. For each package, we need the ability to trace back to the original source of installation information and the method for determining the current version. These metadata are not necessary for installation functionality but are important for maintenance, updates, and documentation.

Currently, we lack a systematic way to track:
- Official installation documentation from the software author
- URL for determining the latest version of the package
- Sources from which installation procedures were derived

## Decision

For each package defined in `assets/install-packages.json`, we will track the following metadata URLs:

1. **`installationDocsUrl`** - URL to the official installation documentation from the software author (required field)
2. **`latestVersionUrl`** - URL for determining the latest version (required field)

Both URLs will be mandatory fields in each package definition. Even when both URLs point to the same location, they must be explicitly specified to maintain consistency and completeness of metadata.

### Example structure:

```json
{
  "packages": [
    {
      "name": "nodejs",
      "description": "JavaScript runtime built on Chrome's V8 JavaScript engine",
      "installationDocsUrl": "https://nodejs.org/en/download/package-manager",
      "latestVersionUrl": "https://nodejs.org/dist/latest/",
      "windows": {
        // ... existing installation config
      },
      "linux": {
        // ... existing installation config
      }
    },
    {
      "name": "claude-code",
      "description": "Anthropic's official CLI for Claude",
      "installationDocsUrl": "https://docs.anthropic.com/en/docs/claude-code",
      "latestVersionUrl": "https://docs.anthropic.com/en/docs/claude-code",
      "windows": {
        // ... existing installation config
      }
    }
  ]
}
```

## Consequences

### Positive:
- **Transparency** - clear overview of installation procedure origins
- **Maintenance** - easier verification and updates of installation procedures against official sources
- **Documentation** - ability to automatically generate links to official documentation
- **Verification** - ability to verify that we're using current installation procedures
- **Trustworthiness** - users can verify the origin of installation procedures

### Negative:
- **URL maintenance** - URLs may change and need regular verification
- **Mandatory fields** - all packages must have both URLs specified, even if identical
- **No automation** - this decision only tracks URLs, doesn't introduce automatic version checking
- **Migration effort** - existing packages need to be updated with both URLs

### Neutral:
- Change is only at metadata level, has no impact on installation functionality
- URL duplication when both point to the same location is acceptable for consistency