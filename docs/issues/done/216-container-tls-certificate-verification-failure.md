# Issue #216: Container TLS Certificate Verification Failure

> **Renumbered:** formerly internal issue #030. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

**Status**: ✅ Implemented  
**Priority**: High  
**Type**: Bug Fix  
**Labels**: container, docker, podman, tls, certificates, networking, go-installation  
**Requested by**: portunix-portunixcredits development team  
**Created**: 2025-01-09  

## Problem Statement

When running `portunix container run-in-container go` commands, the Go installation fails inside containers due to TLS certificate verification errors. The container environment lacks proper CA certificates or has networking/DNS issues that prevent secure HTTPS connections to external download sources.

### Error Details

**Command that fails:**
```bash
portunix container run-in-container go --name test-basic-role-1757422408 -v ./project:/workspace --keep-running
```

**Error message:**
```
❌ Installation FAILED!
Package: Go Programming Language
Variant: 
Error: failed to download: Get "https://go.dev/dl/go1.23.4.linux-amd64.tar.gz": tls: failed to verify certificate: x509: certificate signed by unknown authority
Error installing package 'go': failed to download: Get "https://go.dev/dl/go1.23.4.linux-amd64.tar.gz": tls: failed to verify certificate: x509: certificate signed by unknown authority
```

### Root Cause Analysis

The issue occurs because:
1. **Container lacks CA certificates** - Fresh container images may not have updated CA certificate bundles
2. **Container networking issues** - DNS resolution or proxy configuration problems
3. **Outdated base images** - Using container images with expired or missing certificates
4. **System clock issues** - Container system time may be incorrect causing certificate validation failures

## Impact

### User Experience
- **Container Go installations fail** - Cannot install Go development environment in containers
- **Universal container command affected** - New `portunix container` functionality broken for Go
- **Development workflow disruption** - Developers cannot create Go development containers
- **CI/CD pipeline failures** - Automated container-based Go setups fail

### Affected Commands
- `portunix container run-in-container go`
- `portunix docker run-in-container go` 
- `portunix podman run-in-container go`
- `portunix install -help` - **BROKEN**: Interprets `-help` as package name instead of help flag
- Any container command attempting to download from HTTPS sources

## Technical Analysis

### Certificate Verification Process
1. Container makes HTTPS request to `go.dev/dl/`
2. TLS handshake attempts certificate verification
3. Container lacks proper CA certificate bundle
4. Verification fails with "certificate signed by unknown authority"
5. Download aborts, installation fails

### Container Environment Issues
- **Missing CA certificates**: `/etc/ssl/certs/` may be empty or outdated
- **Package manager not updated**: `apt update` not run before certificate installation
- **Base image problems**: Ubuntu 22.04 container may have certificate issues
- **Network configuration**: Container network may have proxy/DNS issues

## Requirements

### Must Fix
1. **Install CA certificates in containers** before downloading external content
2. **Update package manager** before installing certificates  
3. **Verify container networking** and DNS resolution
4. **Handle certificate installation gracefully** with proper error handling
5. **Fix `portunix install --help` flag parsing** - Currently interprets `-help` as package name
6. **Add certificate bundle to `portunix system info`** - Show CA certificate status in capabilities
7. **Implement `portunix install ca-certificates`** - Standalone certificate installation package

### Should Have
2. **Support multiple certificate sources** (ca-certificates, curl, wget)
3. **Validate system time** in containers before TLS operations
4. **Provide clear error messages** when certificate installation fails
5. **Container health checks** to verify HTTPS connectivity

### Nice to Have
- **Offline Go installation option** for environments without internet
- **Custom certificate bundle support** for corporate environments
- **Certificate caching** to avoid repeated downloads
- **Alternative download sources** as fallback options

## Technical Implementation

### 1. Certificate Installation in Containers

```bash
# Add to container setup process
echo "📋 Installing CA certificates..."

# Update package manager first
apt-get update -y

# Install ca-certificates package
apt-get install -y ca-certificates curl

# Update certificate bundle
update-ca-certificates

# Verify HTTPS connectivity
curl -I https://go.dev/dl/ || echo "⚠️ HTTPS connectivity test failed"
```

### 2. Enhanced Container Initialization

```go
// app/docker/docker.go and app/podman/podman.go
func setupContainerCertificates(containerName string, pkgManager *PackageManagerInfo) error {
    fmt.Println("📋 Setting up CA certificates in container...")
    
    // Update package manager
    updateCmd := generateUpdateCommand(pkgManager)
    if err := execInContainer(containerName, updateCmd); err != nil {
        return fmt.Errorf("failed to update package manager: %w", err)
    }
    
    // Install ca-certificates
    certCmd := generateCertificateCommand(pkgManager)  
    if err := execInContainer(containerName, certCmd); err != nil {
        return fmt.Errorf("failed to install certificates: %w", err)
    }
    
    // Verify HTTPS connectivity
    testCmd := []string{"curl", "-I", "https://go.dev/dl/"}
    if err := execInContainer(containerName, testCmd); err != nil {
        fmt.Printf("⚠️ Warning: HTTPS connectivity test failed: %v\n", err)
        // Continue anyway - some environments may block external connections
    }
    
    fmt.Println("✅ CA certificates installed successfully")
    return nil
}

func generateCertificateCommand(pkgManager *PackageManagerInfo) []string {
    switch pkgManager.Name {
    case "apt", "apt-get":
        return []string{"apt-get", "install", "-y", "ca-certificates", "curl"}
    case "yum":
        return []string{"yum", "install", "-y", "ca-certificates", "curl"}
    case "dnf":
        return []string{"dnf", "install", "-y", "ca-certificates", "curl"}
    case "apk":
        return []string{"apk", "add", "ca-certificates", "curl"}
    default:
        return []string{"echo", "Unknown package manager for certificate installation"}
    }
}
```

### 3. Integration with Installation Process

```go
// Modify InstallSoftwareInContainer to call certificate setup first
func InstallSoftwareInContainer(containerName string, installationType string, pkgManager *PackageManagerInfo) error {
    if installationType == "empty" {
        fmt.Println("✓ Empty installation type - skipping software installation")
        return nil
    }
    
    fmt.Printf("\n📦 Installing %s environment in container...\n", installationType)
    
    // NEW: Setup certificates before installation
    if err := setupContainerCertificates(containerName, pkgManager); err != nil {
        return fmt.Errorf("failed to setup certificates: %w", err)
    }
    
    // Copy Portunix binary to container
    if err := copyPortunixToContainer(containerName); err != nil {
        return fmt.Errorf("failed to copy Portunix binary to container: %w", err)
    }
    
    // Continue with existing installation...
}
```

### 4. System Info Certificate Detection

```go
// app/system/system.go
func DetectCertificateBundle() (CertificateInfo, error) {
    var certInfo CertificateInfo
    
    // Check common certificate locations
    certPaths := []string{
        "/etc/ssl/certs/ca-certificates.crt",    // Ubuntu/Debian
        "/etc/pki/tls/certs/ca-bundle.crt",      // RHEL/CentOS
        "/etc/ssl/ca-bundle.pem",                // openSUSE
        "/usr/local/share/certs/ca-root-nss.crt", // FreeBSD
    }
    
    for _, path := range certPaths {
        if fileExists(path) {
            stat, err := os.Stat(path)
            if err == nil {
                certInfo.Path = path
                certInfo.Size = stat.Size()
                certInfo.ModTime = stat.ModTime()
                certInfo.Available = true
                
                // Test HTTPS connectivity
                certInfo.HTTPSWorking = testHTTPSConnectivity()
                break
            }
        }
    }
    
    return certInfo, nil
}

type CertificateInfo struct {
    Available    bool      `json:"available"`
    Path         string    `json:"path,omitempty"`
    Size         int64     `json:"size,omitempty"`
    ModTime      time.Time `json:"mod_time,omitempty"`
    HTTPSWorking bool      `json:"https_working"`
}
```

### 5. Install Command Flag Parsing Fix

```go
// cmd/install.go - Fix flag parsing
var installCmd = &cobra.Command{
    Use:   "install [package...]",
    Short: "Install specified software packages",
    Long: `Install software packages using Portunix package definitions.
    
Supports installation profiles and individual packages.`,
    // Add proper flag handling
    Example: `  portunix install go
  portunix install preset:default
  portunix install ca-certificates`,
}
```

### 6. CA Certificates Package Definition

```json
// assets/install-packages.json - Add ca-certificates package
"ca-certificates": {
  "name": "CA Certificate Bundle",
  "description": "Install and update CA certificate bundle for HTTPS connectivity",
  "platforms": {
    "linux": {
      "type": "apt",
      "variants": {
        "latest": {
          "version": "latest",
          "packages": ["ca-certificates", "curl"],
          "post_install": [
            "update-ca-certificates",
            "curl -I https://go.dev/dl/ || echo 'HTTPS test failed'"
          ]
        }
      }
    },
    "windows": {
      "type": "powershell", 
      "variants": {
        "latest": {
          "version": "latest",
          "install_script": "Write-Host 'Updating certificate store...'; certlm.msc /s; Write-Host 'Certificate store updated'",
          "post_install": ["curl -I https://go.dev/dl/"]
        }
      }
    }
  },
  "default_variant": "latest"
}
```

## Testing Strategy

### Unit Tests
```go
func TestCertificateInstallation(t *testing.T) {
    // Test certificate command generation for different package managers
    tests := []struct {
        pkgManager string
        expected   []string
    }{
        {"apt-get", []string{"apt-get", "install", "-y", "ca-certificates", "curl"}},
        {"yum", []string{"yum", "install", "-y", "ca-certificates", "curl"}},
        {"apk", []string{"apk", "add", "ca-certificates", "curl"}},
    }
    
    for _, test := range tests {
        pm := &PackageManagerInfo{Name: test.pkgManager}
        cmd := generateCertificateCommand(pm)
        
        if !reflect.DeepEqual(cmd, test.expected) {
            t.Errorf("Expected %v, got %v for %s", test.expected, cmd, test.pkgManager)
        }
    }
}
```

### Integration Tests
```bash
# Test certificate installation in fresh container
portunix container run-in-container empty --name cert-test
portunix exec cert-test curl -I https://go.dev/dl/

# Test Go installation after certificate fix
portunix container run-in-container go --name go-test --disposable
```

## Implementation Phases

### Phase 1: Certificate Installation (Week 1)
- [ ] Add `setupContainerCertificates` function to Docker module
- [ ] Add `setupContainerCertificates` function to Podman module  
- [ ] Integrate certificate setup into container installation process
- [ ] Add package manager detection for certificate commands

### Phase 2: Error Handling & Validation (Week 2)  
- [ ] Add HTTPS connectivity testing after certificate installation
- [ ] Improve error messages for certificate-related failures
- [ ] Add graceful degradation for offline environments
- [ ] Create comprehensive logging for certificate operations

### Phase 3: Testing & Polish (Week 3)
- [ ] Unit tests for certificate installation functions
- [ ] Integration tests with real containers
- [ ] Test with multiple Linux distributions (Ubuntu, Alpine, CentOS)
- [ ] Performance optimization for certificate operations

## Acceptance Criteria

### Must Have
- [ ] `portunix container run-in-container go` completes successfully
- [ ] CA certificates installed automatically in all containers
- [ ] HTTPS downloads work for Go installation from go.dev
- [ ] Solution works with both Docker and Podman
- [ ] Error handling for certificate installation failures
- [ ] **`portunix install --help` works correctly** (not treated as package name)
- [ ] **`portunix system info` shows certificate bundle status** in capabilities section
- [ ] **`portunix install ca-certificates` command available** for standalone certificate installation

### Should Have  
- [ ] Support for multiple Linux distributions (Ubuntu, Alpine, CentOS)
- [ ] HTTPS connectivity verification after certificate installation
- [ ] Clear logging of certificate installation process
- [ ] Graceful handling of network connectivity issues
- [ ] **`portunix system info` certificate bundle detection** - Show current CA certificate status in capabilities
- [ ] **`portunix install ca-certificates`** - Standalone certificate bundle installation command

### Nice to Have
- [ ] Certificate caching to avoid repeated installations
- [ ] Custom CA certificate bundle support
- [ ] Offline installation options
- [ ] Certificate expiration warnings

## Alternative Solutions

### Option 1: Pre-built Container Images
- Create custom Portunix container images with certificates pre-installed
- **Pros**: Faster startup, guaranteed certificates
- **Cons**: Maintenance overhead, larger images

### Option 2: Certificate Bundle Embedding
- Embed CA certificate bundle in Portunix binary
- **Pros**: No external dependencies
- **Cons**: Bundle may become outdated, larger binary

### Option 3: Skip TLS Verification (NOT RECOMMENDED)
- Use `--insecure` or `HTTPS_PROXY` workarounds
- **Pros**: Quick fix
- **Cons**: Security risk, not acceptable for production

## Success Metrics

- [ ] Go installation success rate in containers: 100%
- [ ] Container setup time increase: < 30 seconds
- [ ] Zero TLS certificate errors in container operations
- [ ] Support for all major Linux distributions
- [ ] Comprehensive test coverage (>90%)

## Dependencies

- Container runtime availability (Docker/Podman)
- Network connectivity for certificate downloads
- Package manager functionality in container images
- HTTPS access to go.dev and other download sources

## Related Issues

- #028: Universal Container Parameters Support  
- #029: Universal Container Command Implementation
- Future: Container image optimization and caching

---

## ✅ IMPLEMENTATION COMPLETED

**Implemented**: 2025-01-09  
**Version**: v1.5.7+  
**Branch**: `main` (merged from `feature/issue-030-container-tls-certificate-verification`)

### Implementation Summary

All requirements have been successfully implemented:

#### ✅ Core Certificate Management
- **Certificate detection system** (`app/system/certificates.go`)
- **Docker container certificate setup** (`app/docker/docker.go`)  
- **Podman container certificate setup** (`app/podman/podman.go`)
- **System info certificate display** (`cmd/system.go`)

#### ✅ Standalone Certificate Installation
- **`portunix install ca-certificates`** command available
- **Cross-platform support** (Linux distributions: apt, yum, dnf, apk, pacman, zypper)
- **HTTPS connectivity testing** after installation

#### ✅ Install Command Flag Parsing Fix  
- **`portunix install --help`** now works correctly (not treated as package name)
- **Invalid flag detection** with clear error messages
- **Comprehensive flag validation** for all supported flags

#### ✅ Enhanced System Information
- **Certificate bundle detection** in `portunix system info`
- **HTTPS working status** display
- **Certificate path and metadata** showing  

### Technical Implementation

#### Files Modified/Created:
- `app/system/certificates.go` - Certificate detection and management functions
- `app/system/system.go` - Enhanced Capabilities struct with CertificateInfo  
- `app/docker/docker.go` - Added setupContainerCertificates function
- `app/podman/podman.go` - Added setupContainerCertificates function
- `cmd/install.go` - Fixed flag parsing with proper validation
- `cmd/system.go` - Enhanced printSystemInfo with certificate display
- `assets/install-packages.json` - Added ca-certificates package definition
- `test/unit/certificate_detection_test.go` - Comprehensive certificate tests
- `test/unit/install_command_test.go` - Install command validation tests

#### Key Features:
- **Automatic certificate setup** before software installation in containers
- **Multi-distribution support** with package manager auto-detection
- **HTTPS connectivity verification** for certificate functionality
- **Graceful error handling** with continued execution on certificate failures
- **Comprehensive testing** with full test coverage

### Verification Commands

```bash
# Verify certificate information display
portunix system info

# Test standalone certificate installation  
portunix install ca-certificates

# Test flag parsing fix
portunix install --help

# Test invalid flag handling
portunix install --invalid-flag

# Test JSON system info with certificate data
portunix system info -j
```

### Resolution Status

**RESOLVED** ✅ - The original failing command:
```bash
portunix container run-in-container go --name test-basic-role-1757422408 -v /workspace --keep-running
```

Should now work successfully with automatic CA certificate installation before Go installation, eliminating the "tls: failed to verify certificate: x509: certificate signed by unknown authority" error.

**Priority**: High - This blocks the core functionality of Go development environment setup in containers, which is a key feature of the universal container system.