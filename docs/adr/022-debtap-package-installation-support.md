# ADR-022: Debtap Package Installation Support

**Status:** Proposed
**Date:** 2025-09-28
**Architect:** Claude (AI Assistant)

## Context

Currently, Portunix supports various package installation methods across different Linux distributions:
- **APT** (Debian/Ubuntu): Native .deb package management
- **YUM/DNF** (Red Hat/Fedora/CentOS): Native .rpm package management
- **Snap**: Universal package format
- **Direct downloads**: tar.gz, AppImage, and binary installations

However, there are scenarios where software is only available as .deb packages but users are running RPM-based distributions (Fedora, CentOS, openSUSE), or vice versa. This creates a compatibility gap that forces users to:

1. **Manual conversion**: Use external tools to convert packages
2. **Source compilation**: Build from source code
3. **Alternative versions**: Use potentially outdated versions from native repositories
4. **Container workarounds**: Run software in containers with compatible base images

### Problem Scenarios

#### Real-World Examples
- **Chrome on Fedora**: Google provides .deb but not .rpm for Chrome
- **Proprietary software**: Many vendors only provide .deb packages
- **Development tools**: Some tools are packaged only for Ubuntu/Debian
- **Legacy applications**: Older software with .deb-only distributions

#### Current Workarounds
```bash
# Manual debtap conversion (current user workflow)
sudo pacman -S debtap
sudo debtap -u
debtap google-chrome-stable_current_amd64.deb
sudo pacman -U google-chrome-stable-*.pkg.tar.xz

# Portunix current limitation
portunix install chrome  # ❌ Fails on Arch/Fedora if no native package
```

### Debtap Overview

**Debtap** (DEB To Arch Package) is a script that converts Debian packages (.deb) to Arch Linux packages, but can be adapted for other RPM-based systems:

- **Purpose**: Convert .deb packages to native format
- **Functionality**: Extracts metadata, dependencies, and files
- **Compatibility**: Works with most .deb packages
- **Dependencies**: Handles dependency mapping between package systems

## Decision

Implement **debtap integration** as a fallback installation method for cross-distribution package compatibility.

### 1. Debtap Integration Architecture

```
app/install/
├── debtap/
│   ├── converter.go           # Main debtap conversion logic
│   ├── dependency_mapper.go   # Map .deb dependencies to RPM equivalents
│   ├── package_validator.go   # Validate converted packages
│   └── cache_manager.go       # Cache converted packages
└── installer.go               # Main installer with debtap fallback
```

### 2. Installation Flow with Debtap Fallback

```
Package Installation Request
├── 1. Try Native Package Manager (APT/YUM/DNF)
├── 2. Try Universal Formats (Snap/Flatpak)
├── 3. Try Direct Download (tar.gz/AppImage)
└── 4. Try Debtap Conversion (NEW)
    ├── Download .deb package
    ├── Convert using debtap
    ├── Install converted package
    └── Cache for future use
```

### 3. Package Definition Enhancement

Extend package definitions to support debtap conversion:

```json
{
  "metadata": {
    "name": "chrome",
    "displayName": "Google Chrome"
  },
  "spec": {
    "platforms": {
      "linux": {
        "fallback": {
          "debtap": {
            "enabled": true,
            "source": {
              "url": "https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb",
              "checksum": "sha256:..."
            },
            "dependencyMapping": {
              "libgtk-3-0": "gtk3",
              "libgconf-2-4": "gconf",
              "libnss3": "nss"
            },
            "postInstall": [
              "sudo update-desktop-database",
              "sudo gtk-update-icon-cache /usr/share/icons/hicolor"
            ]
          }
        }
      }
    }
  }
}
```

### 4. Debtap Conversion Process

#### Step 1: Environment Setup
```go
func (d *DebtapConverter) Setup() error {
    // Install debtap if not present
    if !d.isDebtapInstalled() {
        return d.installDebtap()
    }

    // Update debtap database
    return d.updateDatabase()
}

func (d *DebtapConverter) installDebtap() error {
    switch d.detectDistribution() {
    case "arch":
        return d.runCommand("sudo", "pacman", "-S", "debtap")
    case "fedora":
        return d.installDebtapFromSource()
    default:
        return errors.New("debtap not supported on this distribution")
    }
}
```

#### Step 2: Package Conversion
```go
func (d *DebtapConverter) ConvertPackage(debPath string, options ConvertOptions) (*ConvertedPackage, error) {
    // Create temporary workspace
    workspace := d.createWorkspace()
    defer d.cleanupWorkspace(workspace)

    // Convert package
    convertedPath, err := d.runDebtapConversion(debPath, workspace, options)
    if err != nil {
        return nil, err
    }

    // Validate converted package
    pkg, err := d.validateConvertedPackage(convertedPath)
    if err != nil {
        return nil, err
    }

    // Cache converted package
    d.cachePackage(pkg)

    return pkg, nil
}

func (d *DebtapConverter) runDebtapConversion(debPath, workspace string, options ConvertOptions) (string, error) {
    args := []string{"-q"}  // quiet mode

    if options.SkipDependencyCheck {
        args = append(args, "-Q")
    }

    args = append(args, debPath)

    cmd := exec.Command("debtap", args...)
    cmd.Dir = workspace

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("debtap conversion failed: %w\nOutput: %s", err, output)
    }

    // Find converted package file
    return d.findConvertedPackage(workspace)
}
```

#### Step 3: Dependency Mapping
```go
type DependencyMapper struct {
    mappings map[string]map[string]string  // [distribution][debian_package]target_package
}

func (dm *DependencyMapper) MapDependencies(debDeps []string, targetDistro string) ([]string, error) {
    var mapped []string

    for _, debDep := range debDeps {
        if targetDep, exists := dm.mappings[targetDistro][debDep]; exists {
            mapped = append(mapped, targetDep)
        } else {
            // Try fuzzy matching or suggest manual mapping
            suggested := dm.suggestMapping(debDep, targetDistro)
            if suggested != "" {
                mapped = append(mapped, suggested)
            } else {
                log.Warnf("No mapping found for dependency: %s", debDep)
            }
        }
    }

    return mapped, nil
}

func (dm *DependencyMapper) loadMappings() {
    // Load from embedded mappings or external configuration
    dm.mappings = map[string]map[string]string{
        "fedora": {
            "libgtk-3-0":    "gtk3",
            "libgconf-2-4":  "gconf",
            "libnss3":       "nss",
            "libxss1":       "libXScrnSaver",
            "libgdk-pixbuf2.0-0": "gdk-pixbuf2",
        },
        "arch": {
            "libgtk-3-0":    "gtk3",
            "libgconf-2-4":  "gconf",
            "libnss3":       "nss",
            "libxss1":       "libxss",
        },
    }
}
```

### 5. Cache Management

```go
type CacheManager struct {
    cacheDir    string
    maxSize     int64
    maxAge      time.Duration
}

func (cm *CacheManager) CachePackage(pkg *ConvertedPackage) error {
    cacheKey := fmt.Sprintf("%s_%s_%s", pkg.Name, pkg.Version, pkg.Architecture)
    cachePath := filepath.Join(cm.cacheDir, cacheKey+".pkg.tar.xz")

    // Copy converted package to cache
    if err := cm.copyFile(pkg.Path, cachePath); err != nil {
        return err
    }

    // Create metadata file
    metadata := CacheMetadata{
        OriginalURL:  pkg.OriginalURL,
        ConvertedAt:  time.Now(),
        Checksum:     pkg.Checksum,
        Dependencies: pkg.Dependencies,
    }

    return cm.saveMetadata(cacheKey, metadata)
}

func (cm *CacheManager) GetCachedPackage(name, version, arch string) (*ConvertedPackage, bool) {
    cacheKey := fmt.Sprintf("%s_%s_%s", name, version, arch)
    cachePath := filepath.Join(cm.cacheDir, cacheKey+".pkg.tar.xz")

    if !cm.fileExists(cachePath) {
        return nil, false
    }

    // Check if cache is still valid
    metadata, err := cm.loadMetadata(cacheKey)
    if err != nil || time.Since(metadata.ConvertedAt) > cm.maxAge {
        return nil, false
    }

    return &ConvertedPackage{
        Path:     cachePath,
        Metadata: metadata,
    }, true
}
```

### 6. Integration with Main Installer

```go
func (installer *Installer) InstallPackage(packageName string, options InstallOptions) error {
    // Try standard installation methods first
    if err := installer.tryNativeInstall(packageName, options); err == nil {
        return nil
    }

    if err := installer.tryUniversalFormats(packageName, options); err == nil {
        return nil
    }

    if err := installer.tryDirectDownload(packageName, options); err == nil {
        return nil
    }

    // Fallback to debtap conversion
    if installer.isDebtapFallbackEnabled(packageName) {
        log.Infof("Trying debtap conversion for %s", packageName)
        return installer.installViaDebtap(packageName, options)
    }

    return fmt.Errorf("no installation method available for package: %s", packageName)
}

func (installer *Installer) installViaDebtap(packageName string, options InstallOptions) error {
    pkg, err := installer.packageManager.GetPackage(packageName)
    if err != nil {
        return err
    }

    debtapConfig := pkg.Spec.Platforms.Linux.Fallback.Debtap
    if debtapConfig == nil || !debtapConfig.Enabled {
        return errors.New("debtap fallback not configured for this package")
    }

    // Check cache first
    if cached, found := installer.debtapCache.GetCachedPackage(packageName, pkg.Version, options.Architecture); found {
        return installer.installConvertedPackage(cached)
    }

    // Download .deb package
    debPath, err := installer.downloadDebPackage(debtapConfig.Source)
    if err != nil {
        return err
    }
    defer os.Remove(debPath)

    // Convert package
    converted, err := installer.debtapConverter.ConvertPackage(debPath, ConvertOptions{
        DependencyMapping: debtapConfig.DependencyMapping,
        PostInstall:      debtapConfig.PostInstall,
    })
    if err != nil {
        return err
    }

    // Install converted package
    return installer.installConvertedPackage(converted)
}
```

### 7. User Experience

#### Command Line Interface
```bash
# Standard installation (tries debtap if needed)
portunix install chrome                    # ✅ Works on Fedora via debtap

# Explicit debtap conversion
portunix install chrome --force-debtap     # Force debtap even if native exists

# Cache management
portunix cache list --debtap               # List cached converted packages
portunix cache clean --debtap              # Clean debtap cache

# Conversion status
portunix install chrome --dry-run          # Shows: "Will use debtap conversion"
```

#### Progress Output
```
📦 Installing chrome...
├── ❌ Native package manager (yum): Package not found
├── ❌ Universal formats: Not available
├── ❌ Direct download: Not configured
└── 🔄 Debtap conversion:
    ├── ⬇️  Downloading .deb package (45.2 MB)
    ├── 🔧 Converting package using debtap
    ├── 📋 Mapping dependencies (3 found, 2 mapped)
    ├── ✅ Package converted successfully
    └── 📦 Installing converted package
✅ Chrome installed successfully via debtap conversion
```

## Consequences

### Positive

1. **Cross-Distribution Compatibility**: Enables installation of .deb packages on RPM-based systems
2. **Expanded Software Availability**: Access to Debian/Ubuntu exclusive packages
3. **Fallback Reliability**: Provides additional installation method when others fail
4. **Cache Efficiency**: Converted packages are cached for reuse
5. **Automatic Process**: Users don't need to manually handle conversions

### Negative

1. **Additional Complexity**: More code to maintain and test
2. **Dependency Mapping**: Manual mapping required for complex dependencies
3. **Conversion Reliability**: Not all .deb packages convert cleanly
4. **Performance Impact**: Conversion process adds installation time
5. **Storage Requirements**: Cache uses additional disk space

### Risks and Mitigation

#### Risk 1: Conversion Failures
- **Risk**: Some .deb packages may not convert properly
- **Mitigation**: Extensive testing, fallback error handling
- **Detection**: Pre-validation of package structure

#### Risk 2: Dependency Conflicts
- **Risk**: Mapped dependencies may conflict with system packages
- **Mitigation**: Comprehensive dependency mapping database
- **Detection**: Pre-installation dependency analysis

#### Risk 3: Security Concerns
- **Risk**: Converted packages bypass native security policies
- **Mitigation**: Checksum verification, trusted source validation
- **Detection**: Package signature checking where available

### Implementation Considerations

#### Phase 1: Basic Conversion (Weeks 1-2)
- Implement core debtap integration
- Add conversion for 3 high-priority packages (Chrome, VS Code, Discord)
- Basic cache management

#### Phase 2: Dependency Mapping (Weeks 3-4)
- Comprehensive dependency mapping database
- Automatic dependency resolution
- Conversion validation

#### Phase 3: Advanced Features (Weeks 5-6)
- Cache optimization and cleanup
- Conversion progress reporting
- Error recovery and troubleshooting

#### Phase 4: Testing and Optimization (Weeks 7-8)
- Cross-distribution testing
- Performance optimization
- Documentation and user guides

## Related ADRs

- ADR-009: Officially Supported Linux Distributions
- ADR-021: Package Registry Architecture
- ADR-007: Prerequisite Package Handling System

## Success Metrics

1. **Compatibility**: 90% of .deb packages convert successfully
2. **Performance**: Conversion adds <30 seconds to installation time
3. **Reliability**: <5% conversion failure rate in production
4. **User Adoption**: 20% of RPM-based users utilize debtap fallback

---

## Product Owner Decision

**Status: [PENDING REVIEW]**
**Date:** 2025-09-28
**Product Owner:** [To be assigned]

### Business Value Assessment

This proposal addresses a real **cross-distribution compatibility gap** that affects enterprise and individual users deploying Portunix across heterogeneous Linux environments.

### Key Business Benefits:

#### 1. **Market Expansion** 🌍
- Enables software installation on previously incompatible distributions
- **Business Impact**: Broader Portunix adoption across Linux ecosystem

#### 2. **User Experience** 👥
- Eliminates manual package conversion workflows
- **Business Impact**: Reduced support requests, higher user satisfaction

#### 3. **Competitive Advantage** ⚡
- Unique cross-distribution package installation capability
- **Business Impact**: Differentiation from other package managers

### Risk Assessment:

#### Technical Risks ⚠️
- **Conversion failures**: Some packages may not convert properly
- **Security implications**: Bypassing native package validation
- **Maintenance overhead**: Additional code complexity

#### Business Risks 💼
- **Support complexity**: More failure modes to troubleshoot
- **User expectations**: May expect 100% conversion success rate

### Recommendation:

**[TO BE COMPLETED BY PRODUCT OWNER]**

---

**Decision Status:** Pending Product Owner Review