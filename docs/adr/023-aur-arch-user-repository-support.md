# ADR-023: AUR (Arch User Repository) Support

**Status:** Proposed
**Date:** 2025-09-28
**Architect:** Claude (AI Assistant)

## Context

Currently, Portunix supports various package installation methods across different Linux distributions, but lacks support for the **Arch User Repository (AUR)** - one of the most comprehensive and up-to-date software repositories in the Linux ecosystem.

### AUR Overview

The **Arch User Repository** is a community-driven repository containing package build scripts (PKGBUILDs) for software not included in official Arch repositories:

- **Package Count**: 85,000+ packages (significantly larger than most official repositories)
- **Update Frequency**: Community-maintained, often more current than official packages
- **Coverage**: Includes bleeding-edge software, development tools, proprietary applications
- **Build System**: Source-based compilation with automated dependency resolution

AUR je "App Store" pro Arch Linux, ale místo hotových aplikací obsahuje recepty, které si zkompilujete sami. Je to hlavní důvod, proč má Arch Linux tak obrovský ekosystém software - pokud něco
  existuje pro Linux, je to pravděpodobně v AUR.

### Current Limitations

#### Missing Software Categories
```bash
# Software commonly found only in AUR
portunix install discord          # ❌ Not in official repos
portunix install visual-studio-code-bin  # ❌ AUR-only binary version
portunix install google-chrome   # ❌ AUR has latest versions
portunix install zoom             # ❌ Proprietary software in AUR
portunix install spotify         # ❌ Music streaming clients
portunix install slack-desktop   # ❌ Communication tools
portunix install jetbrains-toolbox # ❌ Development IDEs
portunix install minecraft-launcher # ❌ Gaming applications
```

#### Current User Workflow vs Desired
```bash
# Current manual workflow (what users do now)
git clone https://aur.archlinux.org/discord.git
cd discord
makepkg -si
sudo pacman -U discord-*.pkg.tar.xz

# Desired Portunix workflow
portunix install discord         # ✅ Automatic AUR handling
```

### AUR Helpers Landscape

Popular AUR helpers that Portunix should integrate with or emulate:
- **yay**: Most popular, written in Go (good integration candidate)
- **paru**: Modern Rust-based helper with advanced features
- **pikaur**: Python-based with sophisticated dependency resolution
- **aurman**: Pacman-like syntax and behavior

### Real-World Use Cases

#### Development Environment Setup
```bash
# Comprehensive development setup with AUR packages
portunix install default-dev-aur  # New profile with AUR packages
# Could include: visual-studio-code-bin, jetbrains-toolbox, discord, slack-desktop
```

#### Enterprise Software Management
```bash
# Business applications often available only in AUR
portunix install zoom teams-for-linux slack-desktop
portunix install google-chrome-dev  # Development versions
```

#### Gaming and Entertainment
```bash
# Gaming software ecosystem
portunix install steam lutris minecraft-launcher discord
portunix install spotify youtube-music-desktop-app
```

## Decision

Implement **comprehensive AUR support** in Portunix with intelligent integration of existing AUR helpers and custom build logic.

### 1. AUR Integration Architecture

```
app/install/
├── aur/
│   ├── helper_manager.go      # Manage AUR helper installation and selection
│   ├── package_resolver.go    # Resolve AUR package dependencies
│   ├── build_manager.go       # Handle package building and compilation
│   ├── cache_manager.go       # Cache built packages and metadata
│   ├── security_validator.go  # Validate PKGBUILDs for security
│   └── aur_client.go         # AUR RPC API client
└── installer.go              # Main installer with AUR integration
```

### 2. AUR Helper Integration Strategy

#### Primary Strategy: yay Integration
```go
type AURHelperManager struct {
    preferredHelper string
    availableHelpers map[string]AURHelper
    fallbackEnabled bool
}

type AURHelper interface {
    Install(packageName string, options InstallOptions) error
    Search(query string) ([]AURPackage, error)
    Update(packageNames []string) error
    IsInstalled() bool
    InstallHelper() error
}

type YayHelper struct {
    binaryPath string
    version    string
}

func (y *YayHelper) Install(packageName string, options InstallOptions) error {
    args := []string{"-S", packageName}

    if options.NoConfirm {
        args = append(args, "--noconfirm")
    }

    if options.Needed {
        args = append(args, "--needed")
    }

    cmd := exec.Command("yay", args...)
    return cmd.Run()
}

func (y *YayHelper) InstallHelper() error {
    // Install yay if not present
    if y.IsInstalled() {
        return nil
    }

    return y.installYayFromGit()
}

func (y *YayHelper) installYayFromGit() error {
    workDir := "/tmp/yay-install"

    // Clone yay repository
    if err := exec.Command("git", "clone", "https://aur.archlinux.org/yay.git", workDir).Run(); err != nil {
        return err
    }

    defer os.RemoveAll(workDir)

    // Build and install yay
    cmd := exec.Command("makepkg", "-si", "--noconfirm")
    cmd.Dir = workDir

    return cmd.Run()
}
```

#### Fallback Strategy: Native Implementation
```go
type NativeAURBuilder struct {
    cacheDir    string
    buildDir    string
    aurClient   *AURClient
}

func (n *NativeAURBuilder) Install(packageName string, options InstallOptions) error {
    // 1. Query AUR for package metadata
    aurPkg, err := n.aurClient.GetPackage(packageName)
    if err != nil {
        return err
    }

    // 2. Download and extract package source
    buildPath, err := n.downloadAndExtractSource(aurPkg)
    if err != nil {
        return err
    }
    defer n.cleanup(buildPath)

    // 3. Validate PKGBUILD security
    if err := n.validatePKGBUILD(filepath.Join(buildPath, "PKGBUILD")); err != nil {
        return err
    }

    // 4. Resolve and install dependencies
    if err := n.resolveDependencies(aurPkg.Dependencies); err != nil {
        return err
    }

    // 5. Build package
    builtPackage, err := n.buildPackage(buildPath)
    if err != nil {
        return err
    }

    // 6. Install built package
    return n.installBuiltPackage(builtPackage)
}

func (n *NativeAURBuilder) buildPackage(buildPath string) (string, error) {
    cmd := exec.Command("makepkg", "-f", "--noconfirm")
    cmd.Dir = buildPath

    // Set build environment
    cmd.Env = append(os.Environ(),
        "PKGDEST="+n.cacheDir,
        "SRCDEST="+filepath.Join(n.cacheDir, "sources"),
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("makepkg failed: %w\nOutput: %s", err, output)
    }

    // Find built package file
    return n.findBuiltPackage(buildPath)
}
```

### 3. AUR RPC API Integration

```go
type AURClient struct {
    baseURL    string
    httpClient *http.Client
}

type AURPackage struct {
    Name          string    `json:"Name"`
    Version       string    `json:"Version"`
    Description   string    `json:"Description"`
    URL           string    `json:"URL"`
    URLPath       string    `json:"URLPath"`
    License       []string  `json:"License"`
    Dependencies  []string  `json:"Depends"`
    MakeDepends   []string  `json:"MakeDepends"`
    OptDepends    []string  `json:"OptDepends"`
    Maintainer    string    `json:"Maintainer"`
    FirstSubmitted int64    `json:"FirstSubmitted"`
    LastModified   int64    `json:"LastModified"`
    OutOfDate      int64     `json:"OutOfDate"`
    Popularity     float64   `json:"Popularity"`
    NumVotes       int       `json:"NumVotes"`
}

func (c *AURClient) SearchPackages(query string) ([]AURPackage, error) {
    url := fmt.Sprintf("%s/rpc/?v=5&type=search&arg=%s", c.baseURL, url.QueryEscape(query))

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Results []AURPackage `json:"results"`
        Type    string       `json:"type"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    return response.Results, nil
}

func (c *AURClient) GetPackage(name string) (*AURPackage, error) {
    url := fmt.Sprintf("%s/rpc/?v=5&type=info&arg[]=%s", c.baseURL, name)

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Results []AURPackage `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    if len(response.Results) == 0 {
        return nil, fmt.Errorf("package not found: %s", name)
    }

    return &response.Results[0], nil
}
```

### 4. Security and Validation

#### PKGBUILD Security Scanning
```go
type SecurityValidator struct {
    rules []SecurityRule
}

type SecurityRule struct {
    Pattern     *regexp.Regexp
    Severity    string
    Description string
}

func (sv *SecurityValidator) ValidatePKGBUILD(pkgbuildPath string) error {
    content, err := ioutil.ReadFile(pkgbuildPath)
    if err != nil {
        return err
    }

    var warnings []string
    var errors []string

    for _, rule := range sv.rules {
        if rule.Pattern.Match(content) {
            message := fmt.Sprintf("Security %s: %s", rule.Severity, rule.Description)

            switch rule.Severity {
            case "ERROR":
                errors = append(errors, message)
            case "WARNING":
                warnings = append(warnings, message)
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("security errors in PKGBUILD:\n%s", strings.Join(errors, "\n"))
    }

    if len(warnings) > 0 {
        log.Warnf("Security warnings in PKGBUILD:\n%s", strings.Join(warnings, "\n"))
    }

    return nil
}

func (sv *SecurityValidator) loadSecurityRules() {
    sv.rules = []SecurityRule{
        {
            Pattern:     regexp.MustCompile(`rm\s+-rf\s+/`),
            Severity:    "ERROR",
            Description: "Dangerous recursive deletion of root filesystem",
        },
        {
            Pattern:     regexp.MustCompile(`sudo\s+`),
            Severity:    "WARNING",
            Description: "PKGBUILD contains sudo usage (should not be needed)",
        },
        {
            Pattern:     regexp.MustCompile(`curl.*http://`),
            Severity:    "WARNING",
            Description: "Insecure HTTP download detected",
        },
        {
            Pattern:     regexp.MustCompile(`\|\s*sh`),
            Severity:    "ERROR",
            Description: "Piping to shell detected (potential security risk)",
        },
    }
}
```

### 5. Package Definition Enhancement

#### AUR-Enabled Package Definitions
```json
{
  "metadata": {
    "name": "discord",
    "displayName": "Discord",
    "description": "All-in-one voice and text chat for gamers"
  },
  "spec": {
    "platforms": {
      "linux": {
        "arch": {
          "aur": {
            "enabled": true,
            "packageName": "discord",
            "buildType": "binary",
            "trusted": true,
            "maintainer": "package-maintainer",
            "security": {
              "allowNetworkAccess": true,
              "allowSudo": false,
              "scanPKGBUILD": true
            },
            "alternatives": [
              "discord-ptb",
              "discord-canary"
            ]
          }
        }
      }
    }
  }
}
```

#### AUR Package Profiles
```json
{
  "profiles": {
    "arch-developer": {
      "description": "Arch Linux developer environment with AUR packages",
      "packages": [
        "visual-studio-code-bin",
        "jetbrains-toolbox",
        "discord",
        "slack-desktop",
        "google-chrome",
        "spotify",
        "docker-desktop"
      ]
    },
    "arch-gaming": {
      "description": "Gaming setup with AUR game clients",
      "packages": [
        "steam",
        "lutris",
        "minecraft-launcher",
        "discord",
        "teamspeak3",
        "obs-studio-git"
      ]
    }
  }
}
```

### 6. Installation Flow with AUR

```
Package Installation on Arch Linux
├── 1. Try Official Repositories (pacman)
├── 2. Try AUR Packages (NEW)
│   ├── 2a. Query AUR API for package
│   ├── 2b. Install AUR helper if needed (yay/paru)
│   ├── 2c. Validate PKGBUILD security
│   ├── 2d. Build and install package
│   └── 2e. Cache built package
├── 3. Try Universal Formats (Snap/Flatpak)
└── 4. Try Direct Download (tar.gz/AppImage)
```

### 7. User Experience

#### Command Line Interface
```bash
# Standard installation (includes AUR on Arch)
portunix install discord                  # ✅ Finds in AUR automatically

# AUR-specific commands
portunix aur search discord               # Search AUR packages
portunix aur info discord                # Show AUR package info
portunix aur install discord              # Force AUR installation
portunix aur update                       # Update all AUR packages

# Profile installation with AUR packages
portunix install arch-developer           # Install curated AUR profile

# Cache management
portunix cache list --aur                 # List cached AUR packages
portunix cache clean --aur                # Clean AUR build cache
```

#### Progress Output
```
📦 Installing discord...
├── ❌ Official repositories: Package not found
└── 🔄 AUR (Arch User Repository):
    ├── 🔍 Querying AUR API
    ├── ✅ Package found: discord 0.0.29 (maintained by package-maintainer)
    ├── 🔧 Installing AUR helper (yay)
    ├── 📋 Resolving dependencies (3 found)
    ├── 🔒 Validating PKGBUILD security
    ├── ⬇️  Downloading source files (2.3 MB)
    ├── 🏗️  Building package (estimated 2-3 minutes)
    ├── 📦 Installing built package
    └── 🗂️  Caching for future use
✅ Discord installed successfully from AUR
```

#### Interactive Security Prompts
```
⚠️  PKGBUILD Security Review Required
Package: some-package-git
Maintainer: unknown-user (first-time contributor)

Security Warnings:
├── Network access during build (downloads from GitHub)
├── Compiles from latest git commit (not stable release)
└── New maintainer (package created 2 days ago)

Actions:
[V] View PKGBUILD source
[R] Review dependencies
[C] Continue installation
[A] Abort installation

Your choice: _
```

### 8. Advanced Features

#### Automatic AUR Helper Selection
```go
func (am *AURHelperManager) SelectBestHelper() AURHelper {
    // Priority order for AUR helpers
    helpers := []string{"yay", "paru", "pikaur", "aurman"}

    for _, helper := range helpers {
        if am.isHelperInstalled(helper) {
            return am.availableHelpers[helper]
        }
    }

    // Install yay as default if none available
    yay := &YayHelper{}
    if err := yay.InstallHelper(); err == nil {
        return yay
    }

    // Fallback to native implementation
    return &NativeAURBuilder{
        cacheDir:  "/var/cache/portunix/aur",
        buildDir:  "/tmp/portunix-aur-build",
        aurClient: NewAURClient(),
    }
}
```

#### Build Optimization
```go
type BuildOptimizer struct {
    maxConcurrentBuilds int
    buildCache         map[string]CachedBuild
    dependencyResolver *DependencyResolver
}

func (bo *BuildOptimizer) OptimizeBuildOrder(packages []string) ([]BuildJob, error) {
    // Create dependency graph
    graph, err := bo.dependencyResolver.CreateDependencyGraph(packages)
    if err != nil {
        return nil, err
    }

    // Topological sort for build order
    buildOrder := graph.TopologicalSort()

    // Group independent packages for parallel builds
    return bo.createParallelBuildJobs(buildOrder), nil
}

func (bo *BuildOptimizer) createParallelBuildJobs(buildOrder []string) []BuildJob {
    var jobs []BuildJob

    for i := 0; i < len(buildOrder); i += bo.maxConcurrentBuilds {
        end := i + bo.maxConcurrentBuilds
        if end > len(buildOrder) {
            end = len(buildOrder)
        }

        job := BuildJob{
            Packages: buildOrder[i:end],
            Parallel: true,
        }
        jobs = append(jobs, job)
    }

    return jobs
}
```

## Consequences

### Positive

1. **Comprehensive Software Access**: Access to 85,000+ AUR packages
2. **Arch Linux Excellence**: First-class support for Arch's package ecosystem
3. **Developer Experience**: Easy access to cutting-edge development tools
4. **Community Integration**: Leverages existing AUR helper ecosystem
5. **Security Features**: PKGBUILD validation and review mechanisms
6. **Performance**: Caching and build optimization

### Negative

1. **Arch-Specific**: Only beneficial for Arch Linux users
2. **Build Complexity**: Source compilation adds complexity and time
3. **Security Risks**: AUR packages are community-maintained, less vetted
4. **Disk Usage**: Build cache and dependencies require significant space
5. **Network Dependencies**: Requires internet for source downloads
6. **Maintenance Overhead**: Additional code complexity for Arch-specific features

### Risks and Mitigation

#### Risk 1: Malicious PKGBUILDs
- **Risk**: Community packages could contain malicious code
- **Mitigation**: Automated security scanning, trusted maintainer lists
- **Detection**: PKGBUILD pattern analysis, reputation scoring

#### Risk 2: Build Failures
- **Risk**: Source compilation may fail due to missing dependencies
- **Mitigation**: Comprehensive dependency resolution, error recovery
- **Detection**: Build environment validation, dependency checking

#### Risk 3: Maintenance Burden
- **Risk**: AUR packages may become unmaintained or outdated
- **Mitigation**: Package health monitoring, alternative suggestions
- **Detection**: Last-updated tracking, build failure monitoring

### Platform-Specific Considerations

#### Arch Linux Only
This feature is specifically designed for Arch Linux and derivatives:
- **Manjaro**: Full compatibility expected
- **EndeavourOS**: Full compatibility expected
- **ArcoLinux**: Full compatibility expected
- **Other distributions**: Feature disabled, graceful degradation

#### Integration with Existing Systems
```go
func (installer *Installer) isAURSupported() bool {
    // Only enable on Arch-based distributions
    distro := installer.systemInfo.DetectDistribution()
    return distro.Family == "arch"
}

func (installer *Installer) installPackageWithAUR(packageName string) error {
    if !installer.isAURSupported() {
        return installer.installPackageStandard(packageName)
    }

    // AUR-enabled installation flow
    return installer.installWithAURFallback(packageName)
}
```

## Implementation Timeline

### Phase 1: Core AUR Integration (Weeks 1-3)
- AUR API client implementation
- Basic yay helper integration
- Simple package installation (top 10 packages)
- Security validation framework

### Phase 2: Advanced Features (Weeks 4-6)
- Native build system fallback
- Dependency resolution optimization
- Build caching and optimization
- Interactive security reviews

### Phase 3: Package Profiles (Weeks 7-8)
- AUR package profiles (developer, gaming, etc.)
- Comprehensive package definitions
- Performance optimization
- Documentation and testing

### Phase 4: Production Readiness (Weeks 9-10)
- Cross-distribution testing
- Error handling and recovery
- Security hardening
- User experience refinement

## Related ADRs

- ADR-009: Officially Supported Linux Distributions
- ADR-021: Package Registry Architecture
- ADR-022: Debtap Package Installation Support
- ADR-007: Prerequisite Package Handling System

## Success Metrics

1. **Package Coverage**: Support for top 100 most popular AUR packages
2. **Build Success Rate**: >95% successful builds for supported packages
3. **Security**: 0 malicious packages installed through automated validation
4. **Performance**: AUR package installation completes within 5 minutes average
5. **User Adoption**: 60% of Arch Linux users utilize AUR features

---

## Product Owner Decision

**Status: [PENDING REVIEW]**
**Date:** 2025-09-28
**Product Owner:** [To be assigned]

### Business Value Assessment

This proposal targets the **Arch Linux ecosystem** - a significant and growing segment of advanced Linux users who value cutting-edge software and community-driven development.

### Key Business Benefits:

#### 1. **Arch Linux Market Leadership** 🎯
- Positions Portunix as the premier package manager for Arch Linux
- **Business Impact**: Market penetration in developer and enthusiast segments

#### 2. **Software Ecosystem Completeness** 📦
- Access to 85,000+ packages vs ~13,000 in official repos
- **Business Impact**: Eliminates need for multiple package management tools

#### 3. **Developer Community Appeal** 👨‍💻
- Arch users are often developers, early adopters, influencers
- **Business Impact**: Community growth, word-of-mouth marketing

#### 4. **Technical Differentiation** ⚡
- No other universal package manager offers comprehensive AUR integration
- **Business Impact**: Unique competitive advantage

### Market Analysis:

#### Target User Base 📊
- **Arch Linux users**: ~2-3% of Linux desktop market
- **Developer density**: High (60%+ developers/advanced users)
- **Influence factor**: High (community leaders, technical bloggers)
- **Growth trend**: Increasing adoption in cloud/container environments

#### Competitive Landscape 🏆
- **AUR Helpers**: yay, paru, pikaur (single-purpose tools)
- **Universal Managers**: None offer AUR integration
- **Opportunity**: First universal manager with native AUR support

### Risk Assessment:

#### Technical Risks ⚠️
- **Build complexity**: Source compilation increases failure potential
- **Security exposure**: Community packages less vetted than official repos
- **Platform limitation**: Arch-only feature (smaller user base)

#### Business Risks 💼
- **Development cost**: Significant engineering investment for subset of users
- **Support complexity**: More complex troubleshooting scenarios
- **Security liability**: Potential security incidents from AUR packages

### Strategic Recommendation:

**[TO BE COMPLETED BY PRODUCT OWNER]**

Considerations for decision:
1. **Resource allocation**: Does benefit justify development cost?
2. **Strategic focus**: Align with multi-platform vs Arch-specific strategy
3. **Security tolerance**: Acceptable risk level for community packages
4. **Timeline priority**: Urgency compared to other feature requests

---

**Decision Status:** Pending Product Owner Review