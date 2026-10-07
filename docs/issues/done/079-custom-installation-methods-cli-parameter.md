# Issue #79: Version and Method Override via CLI Parameters

**Title**: Override Preferred Installation Method and Version Selection
**Status**: ✅ Implemented
**Closed**: 2026-05-10
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, package-management, installation, method-override, cli

> **Note**: Implementation covers **Phase 1** of the original 3-phase scope
> (`--method` alias, `--list-variants`/`--list-methods`, improved error
> messages, optional `Preferred` field on `VariantSpec`). Phase 2
> (`--version=latest|prerelease`, GitHub Releases API integration) and
> Phase 3 (URL templating, version constraints) are intentionally deferred
> and should be tracked as separate issues if/when needed.
> Acceptance protocol: [`docs/testing/acceptance-079.md`](../../testing/acceptance-079.md)

## Problem Statement

Current Portunix installation system automatically selects the "preferred" installation method from `assets/install-packages.json`. However, packages often have multiple installation methods available, and users sometimes need to bypass the preferred method to access:

1. **Newer versions**: APT may have v1.0.1, but DEB download has v1.2.0
2. **Prerelease versions**: Stable repositories vs. beta/nightly downloads
3. **Different variants**: Minimal vs. full versions of the same package
4. **Compatibility**: Some methods work better on specific system configurations

### Specific Use Case - Hugo Installation

Hugo package definition may include:
- **Preferred**: APT installation (stable, older version v0.140.0)
- **Alternative**: Direct DEB download (latest version v1.150.1)
- **Alternative**: Snap installation (different update channel)

User needs ability to override the preferred APT method and choose newer version:

```bash
# Current behavior - uses preferred method (APT)
portunix install hugo

# Desired capability - get latest/prerelease version
portunix install hugo --version=latest
portunix install hugo --version=prerelease

# Alternative - override installation method
portunix install hugo --method=deb
portunix install hugo --method=snap
```

## Proposed Solution

### 1. CLI Parameters

Add new CLI parameters to `portunix install` command:

```bash
portunix install <package> [options]

New Options:
  --version <version>    Specify version: latest, prerelease, or specific version
  --method <method>      Override preferred method with specified alternative
  --list-methods         Show all available installation methods for package
  --dry-run              Show which method and version would be used
```

### 2. Version and Method Selection Logic

#### Current Behavior
```go
// Current: Always uses preferred method with default version
package.InstallMethods[0] // Preferred method only
```

#### Enhanced Behavior
```go
// Enhanced: Allow version and method override
selectedMethod := package.InstallMethods[0] // Start with preferred

// Override method if specified
if methodFlag != "" {
    selectedMethod = findMethodByName(package.InstallMethods, methodFlag)
}

// Handle version-specific logic
switch versionFlag {
case "latest":
    selectedMethod = findMethodWithLatestCapability(package.InstallMethods)
case "prerelease":
    selectedMethod = findMethodWithPrereleaseCapability(package.InstallMethods)
case "":
    // Use selected method as-is
default:
    // Specific version - find method that supports it
    selectedMethod = findMethodForVersion(package.InstallMethods, versionFlag)
}
```

### 3. Package Definition Structure

Enhanced package definitions to support version and method identification:

```json
{
  "hugo": {
    "description": "Static site generator",
    "install_methods": [
      {
        "id": "apt",
        "preferred": true,
        "platforms": ["linux"],
        "method": "apt",
        "package_name": "hugo",
        "version_support": ["stable"],
        "description": "Stable version from Ubuntu repositories"
      },
      {
        "id": "deb-latest",
        "platforms": ["linux"],
        "method": "deb",
        "github_repo": "gohugoio/hugo",
        "asset_pattern": "hugo_extended_*_linux-amd64.deb",
        "version_support": ["latest", "prerelease", "specific"],
        "url_template": "https://github.com/gohugoio/hugo/releases/download/v{version}/hugo_extended_{version}_linux-amd64.deb",
        "description": "Latest/prerelease version direct from GitHub"
      },
      {
        "id": "snap",
        "platforms": ["linux"],
        "method": "snap",
        "package_name": "hugo",
        "channel": "extended",
        "version_support": ["stable", "latest"],
        "description": "Extended version via Snap"
      }
    ]
  }
}

### 4. Implementation Approach

#### Phase 1: Core Method Selection
1. **CLI Parser Enhancement**
   - Add `--method` flag to install command parser
   - Add `--list-methods` flag for discovery
   - Add `--dry-run` flag for preview
   - Validate method exists in package definition

2. **Method Selection Logic**
   ```go
   type InstallMethodSelector struct {
       packageDef *PackageDefinition
       requestedMethod string
   }

   func (s *InstallMethodSelector) SelectMethod() (*InstallMethod, error) {
       if s.requestedMethod != "" {
           return s.findMethodByID(s.requestedMethod)
       }
       return s.getPreferredMethod()
   }

   func (s *InstallMethodSelector) findMethodByID(id string) (*InstallMethod, error) {
       for _, method := range s.packageDef.InstallMethods {
           if method.ID == id {
               return &method, nil
           }
       }
       return nil, fmt.Errorf("method '%s' not found", id)
   }
   ```

3. **Package Definition Parser Enhancement**
   - Support for multiple install methods with IDs
   - Method metadata (description, platform compatibility)
   - Preferred method marking

#### Phase 2: Enhanced User Experience
1. **Method Discovery**
   - `--list-methods` implementation
   - Show method descriptions and compatibility
   - Indicate preferred method

2. **Dry Run Functionality**
   - Show selected method without execution
   - Display method details and requirements
   - Preview installation steps

3. **Error Handling**
   - Clear error messages for invalid methods
   - Suggestions for available alternatives
   - Platform compatibility warnings

### 5. Usage Examples

#### Version-Based Installation
```bash
# List all available installation methods and version support
portunix install hugo --list-methods

Output:
Available installation methods for 'hugo':
  apt (preferred) - Stable version from Ubuntu repositories [stable]
  deb-latest      - Latest/prerelease version direct from GitHub [latest, prerelease, specific]
  snap            - Extended version via Snap [stable, latest]

# Preview which method would be used
portunix install hugo --dry-run
Output: Would use method 'apt' (preferred) to install hugo stable version

portunix install hugo --version=latest --dry-run
Output: Would use method 'deb-latest' to install hugo latest version

portunix install hugo --version=prerelease --dry-run
Output: Would use method 'deb-latest' to install hugo prerelease version
```

#### Version and Method Examples
```bash
# Default installation (uses preferred method - APT stable)
portunix install hugo

# Get latest version (automatically selects best method)
portunix install hugo --version=latest

# Get prerelease version (automatically selects method with prerelease support)
portunix install hugo --version=prerelease

# Specific version
portunix install hugo --version=v1.150.1

# Override method explicitly
portunix install hugo --method=snap
```

#### Other Package Examples
```bash
# Node.js version examples
portunix install nodejs                    # Default (stable from repositories)
portunix install nodejs --version=latest   # Latest stable from nodejs.org
portunix install nodejs --version=v20.10.0 # Specific version

# VS Code version examples
portunix install vscode                    # Default (stable release)
portunix install vscode --version=latest   # Latest stable
portunix install vscode --version=prerelease # Insider builds
portunix install vscode --method=snap      # Force Snap installation
```

### 6. Configuration Options

#### Global Configuration
```yaml
# ~/.portunix/config.yaml
installation:
  default_method_preference: "preferred"  # preferred, latest, stable
  allow_prerelease: false                 # Allow prerelease methods
  confirm_method_override: true           # Confirm non-preferred methods
```

#### Package Definition Migration
Current format:
```json
{
  "hugo": {
    "description": "Static site generator",
    "platforms": ["linux", "windows"],
    "linux": {
      "method": "apt",
      "package": "hugo"
    },
    "windows": {
      "method": "chocolatey",
      "package": "hugo"
    }
  }
}
```

Enhanced format:
```json
{
  "hugo": {
    "description": "Static site generator",
    "install_methods": [
      {
        "id": "apt",
        "preferred": true,
        "platforms": ["linux"],
        "method": "apt",
        "package_name": "hugo"
      },
      {
        "id": "deb",
        "platforms": ["linux"],
        "method": "deb",
        "url": "https://github.com/gohugoio/hugo/releases/latest/download/hugo_extended_*_linux-amd64.deb"
      }
    ]
  }
}
```

## Technical Considerations

### 1. Backward Compatibility
- **Existing Package Definitions**: Current format must continue working
- **Migration Strategy**: Gradual transition from old to new format
- **Default Behavior**: No changes to current installation behavior
- **Parser Enhancement**: Support both old and new JSON structures

### 2. Method Resolution
- **Platform Filtering**: Only show methods compatible with current platform
- **Dependency Handling**: Some methods may have different prerequisites
- **Error Handling**: Clear error messages for invalid method selections
- **Validation**: Verify method exists before attempting installation

### 3. User Experience
- **Discovery**: Easy way to find available installation methods
- **Defaults**: Sensible defaults that don't change current behavior
- **Feedback**: Clear indication of which method is being used
- **Documentation**: Help text and examples for new flags

### 4. Testing Strategy
- **Unit Tests**: Test method selection logic
- **Integration Tests**: Test method override functionality
- **Container Tests**: Test different installation methods in isolation
- **Backward Compatibility Tests**: Ensure existing functionality works

## Implementation Phases

### Phase 1: Core Method Selection (v1.6.0)
- [ ] CLI parameter `--method` for basic method override
- [ ] Enhanced package definition parser (multiple methods)
- [ ] Method selection logic with fallback to preferred
- [ ] `--list-methods` and `--dry-run` functionality

### Phase 2: Version Intelligence (v1.7.0)
- [ ] `--version=latest` and `--version=prerelease` implementation
- [ ] GitHub Releases API integration for version discovery
- [ ] Intelligent latest version detection
- [ ] Configuration system for version sources
- [ ] URL template resolution with version substitution

### Phase 3: Advanced Version Resolution (v1.8.0)
- [ ] Multiple repository source support
- [ ] Version constraint handling (`>=1.0.0`, `~1.2.0`)
- [ ] Fallback chain for failed prerelease downloads
- [ ] Caching of version metadata for performance

## Acceptance Criteria

### Functional Requirements
1. **Method Override**: User can override preferred method with `--method=<id>`
2. **Method Discovery**: User can list available methods with `--list-methods`
3. **Version Support**: `--version=latest` and `--version=prerelease` intelligently find versions
4. **Dry Run**: User can preview method selection with `--dry-run`
5. **Backward Compatibility**: Existing installation workflow unchanged

### Non-Functional Requirements
1. **Performance**: Method selection adds <100ms overhead
2. **Reliability**: Graceful fallback when prerelease detection fails
3. **Usability**: Clear method descriptions and error messages
4. **Compatibility**: Works with existing package definitions
5. **Intelligence**: Automatic version detection from GitHub API

### Testing Requirements
1. **Unit Tests**: Method selection logic and GitHub API integration
2. **Integration Tests**: End-to-end method override workflow
3. **Container Tests**: Prerelease installation in clean environments
4. **API Tests**: GitHub Releases API interaction and error handling

### Version Intelligence Requirements
1. **GitHub Integration**: Query GitHub Releases API for latest/prerelease versions
2. **URL Resolution**: Substitute version placeholders in download URLs
3. **Fallback Handling**: Graceful degradation when API unavailable
4. **Configuration**: Package-specific version source configuration

## Version Intelligence Design

### GitHub API Integration
```go
type VersionResolver struct {
    apiClient *github.Client
    cache     map[string]*VersionInfo
}

type VersionInfo struct {
    Version     string
    DownloadURL string
    Published   time.Time
    Prerelease  bool
}

func (vr *VersionResolver) GetVersion(repo string, versionType string, pattern string) (*VersionInfo, error) {
    // 1. Query GitHub Releases API
    releases, err := vr.apiClient.Repositories.ListReleases(ctx, owner, repo, opts)

    // 2. Filter by version type (latest stable, prerelease, specific)
    // 3. Filter by asset pattern (e.g., "linux-amd64.deb")
    // 4. Return version info with download URL
}
```

### Configuration Example
```json
{
  "hugo": {
    "install_methods": [
      {
        "id": "deb-latest",
        "method": "deb",
        "github_repo": "gohugoio/hugo",
        "asset_pattern": "hugo_extended_*_linux-amd64.deb",
        "version_support": ["latest", "prerelease", "specific"],
        "url_template": "https://github.com/gohugoio/hugo/releases/download/v{version}/hugo_extended_{version}_linux-amd64.deb"
      }
    ]
  }
}
```

## Related Issues

- [#075](075-implement-hugo-installation-support.md) - Hugo Installation Support (base implementation)
- [#051](../051-git-dispatcher-python-distribution-architecture.md) - Dispatcher Architecture (extension point)

## Risk Assessment

### Medium Risk
- **API Rate Limits**: GitHub API has rate limits for unauthenticated requests
- **Network Dependencies**: Requires internet connection for version detection
- **URL Changes**: Upstream may change download URL patterns

### Mitigation Strategies
- **Caching**: Cache version information to reduce API calls
- **Fallback**: Fall back to preferred method if prerelease detection fails
- **Configuration**: Allow manual URL override in package definitions
- **Rate Limiting**: Implement proper API rate limiting and authentication

---

**Created**: 2025-09-27
**Assignee**: Development Team
**Milestone**: v1.6.0
**Related Components**: installer, cli, security