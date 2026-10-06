# Issue #65: Terraform Installation Support

## Summary
Add support for installing HashiCorp Terraform through Portunix package management system. Terraform is a critical Infrastructure as Code (IaC) tool widely used for provisioning and managing cloud infrastructure across multiple providers.

## Current Behavior
- `portunix install terraform` is not supported
- Users must manually download and install Terraform from HashiCorp's website
- No integration with Portunix package management system
- No version management or update capabilities through Portunix

## Expected Behavior
```bash
# Install latest stable version
portunix install terraform

# Install specific version
portunix install terraform --variant 1.6.6
portunix install terraform --variant latest
portunix install terraform --variant 1.5-lts  # LTS version if available

# Verify installation
portunix verify terraform
terraform --version

# Update terraform
portunix update terraform

# List available versions
portunix list terraform --versions
```

## Expected Output
```
════════════════════════════════════════════════
📦 INSTALLING: Terraform
════════════════════════════════════════════════
📄 Description: Infrastructure as Code tool for multi-cloud provisioning
🔧 Variant: latest (v1.6.6)
💻 Platform: windows
🏗️  Installation type: zip
🌐 Download URL: https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_windows_amd64.zip
════════════════════════════════════════════════
🔍 Checking if terraform is already installed...
📋 terraform is not installed, proceeding with installation...
🚀 Starting installation...
Downloading terraform_1.6.6_windows_amd64.zip...
✅ Downloaded: .cache\terraform_1.6.6_windows_amd64.zip
Extracting terraform...
✅ Extracted to: C:\portunix\bin\terraform.exe
Adding to PATH...
✅ PATH updated successfully
════════════════════════════════════════════════
🎉 Terraform v1.6.6 installed successfully!
📋 Verify with: terraform --version
════════════════════════════════════════════════
```

## Implementation Requirements

### Package Definition
```json
{
  "terraform": {
    "name": "Terraform",
    "description": "Infrastructure as Code tool for multi-cloud provisioning",
    "category": "infrastructure",
    "homepage": "https://www.terraform.io/",
    "license": "MPL-2.0",
    "maintainer": "HashiCorp",
    "variants": {
      "latest": {
        "version": "1.6.6",
        "description": "Latest stable version"
      },
      "1.6": {
        "version": "1.6.6",
        "description": "1.6.x series (current)"
      },
      "1.5": {
        "version": "1.5.6",
        "description": "1.5.x series (previous stable)"
      },
      "1.4-lts": {
        "version": "1.4.7",
        "description": "Long-term support version"
      }
    },
    "platforms": {
      "windows": {
        "architecture": {
          "amd64": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_windows_amd64.zip",
            "installer": "zip",
            "binary": "terraform.exe",
            "install_path": "{portunix_bin}"
          },
          "386": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_windows_386.zip",
            "installer": "zip",
            "binary": "terraform.exe",
            "install_path": "{portunix_bin}"
          }
        }
      },
      "linux": {
        "architecture": {
          "amd64": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_linux_amd64.zip",
            "installer": "zip",
            "binary": "terraform",
            "install_path": "/usr/local/bin",
            "permissions": "755"
          },
          "arm64": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_linux_arm64.zip",
            "installer": "zip",
            "binary": "terraform",
            "install_path": "/usr/local/bin",
            "permissions": "755"
          }
        }
      },
      "darwin": {
        "architecture": {
          "amd64": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_darwin_amd64.zip",
            "installer": "zip",
            "binary": "terraform",
            "install_path": "/usr/local/bin",
            "permissions": "755"
          },
          "arm64": {
            "url": "https://releases.hashicorp.com/terraform/{version}/terraform_{version}_darwin_arm64.zip",
            "installer": "zip",
            "binary": "terraform",
            "install_path": "/usr/local/bin",
            "permissions": "755"
          }
        }
      }
    },
    "post_install": {
      "commands": [
        "terraform version"
      ],
      "environment": {
        "TF_CLI_CONFIG_FILE": "{home}/.terraformrc"
      }
    },
    "verification": {
      "command": "terraform version",
      "expected_output": "Terraform v{version}"
    }
  }
}
```

### HashiCorp Release API Integration
```go
// app/install/terraform.go
package install

import (
    "encoding/json"
    "fmt"
    "net/http"
    "runtime"
)

type TerraformRelease struct {
    Version string `json:"version"`
    Builds  []struct {
        OS   string `json:"os"`
        Arch string `json:"arch"`
        URL  string `json:"url"`
    } `json:"builds"`
}

func GetLatestTerraformVersion() (string, error) {
    resp, err := http.Get("https://api.releases.hashicorp.com/v1/releases/terraform")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var releases []TerraformRelease
    if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
        return "", err
    }

    if len(releases) > 0 {
        return releases[0].Version, nil
    }

    return "", fmt.Errorf("no terraform releases found")
}

func GetTerraformDownloadURL(version string) string {
    os := runtime.GOOS
    arch := runtime.GOARCH

    // Handle special cases
    if os == "darwin" && arch == "amd64" {
        arch = "amd64"
    } else if arch == "x86_64" {
        arch = "amd64"
    }

    return fmt.Sprintf(
        "https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip",
        version, version, os, arch,
    )
}
```

### ZIP Extraction Support
```go
// app/install/zip_extractor.go
func ExtractTerraformFromZip(zipPath, destPath string) error {
    reader, err := zip.OpenReader(zipPath)
    if err != nil {
        return err
    }
    defer reader.Close()

    for _, file := range reader.File {
        if file.Name == "terraform" || file.Name == "terraform.exe" {
            src, err := file.Open()
            if err != nil {
                return err
            }
            defer src.Close()

            dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
            if err != nil {
                return err
            }
            defer dst.Close()

            _, err = io.Copy(dst, src)
            return err
        }
    }

    return fmt.Errorf("terraform binary not found in zip")
}
```

### PATH Management
```go
// app/install/path_manager.go
func AddTerraformToPath(binaryPath string) error {
    // Windows
    if runtime.GOOS == "windows" {
        return addToWindowsPath(binaryPath)
    }

    // Unix-like systems
    return addToUnixPath(binaryPath)
}

func addToWindowsPath(path string) error {
    // Get current PATH
    currentPath := os.Getenv("PATH")

    // Check if already in PATH
    if strings.Contains(currentPath, path) {
        return nil
    }

    // Add to user PATH
    key, err := registry.OpenKey(registry.CURRENT_USER,
        `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
    if err != nil {
        return err
    }
    defer key.Close()

    userPath, _, _ := key.GetStringValue("Path")
    newPath := userPath + ";" + path

    return key.SetStringValue("Path", newPath)
}
```

## Version Management

### Version Checking
```go
func GetInstalledTerraformVersion() (string, error) {
    cmd := exec.Command("terraform", "version", "-json")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    var versionInfo struct {
        TerraformVersion string `json:"terraform_version"`
    }

    if err := json.Unmarshal(output, &versionInfo); err != nil {
        // Fallback to regex parsing
        return parseTerraformVersionFromText(string(output))
    }

    return versionInfo.TerraformVersion, nil
}

func CheckForTerraformUpdates(currentVersion string) (string, bool, error) {
    latest, err := GetLatestTerraformVersion()
    if err != nil {
        return "", false, err
    }

    if compareVersions(latest, currentVersion) > 0 {
        return latest, true, nil
    }

    return currentVersion, false, nil
}
```

## Additional Features

### Provider Plugin Cache
```bash
# Configure Terraform plugin cache directory
portunix terraform configure --plugin-cache ~/.terraform.d/plugin-cache

# This creates/updates ~/.terraformrc:
plugin_cache_dir = "$HOME/.terraform.d/plugin-cache"
```

### Terraform Version Manager (tfenv-like functionality)
```bash
# List installed versions
portunix terraform list-versions

# Switch between versions
portunix terraform use 1.5.6
portunix terraform use latest

# Set project-specific version
echo "1.6.6" > .terraform-version
portunix terraform use  # Uses version from .terraform-version
```

### Integration with Cloud Providers
```bash
# Install with cloud provider CLIs
portunix install terraform --with aws-cli,azure-cli,gcloud

# Configure credentials helpers
portunix terraform configure-credentials aws
portunix terraform configure-credentials azure
portunix terraform configure-credentials gcp
```

## Testing Requirements

### Unit Tests
```go
func TestTerraformVersionParsing(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Terraform v1.6.6", "1.6.6"},
        {"Terraform v1.5.0-beta1", "1.5.0-beta1"},
        {"Terraform v1.4.7\n", "1.4.7"},
    }

    for _, test := range tests {
        result := parseTerraformVersion(test.input)
        if result != test.expected {
            t.Errorf("Expected %s, got %s", test.expected, result)
        }
    }
}
```

### Integration Tests
```go
func TestTerraformInstallation(t *testing.T) {
    tf := testframework.NewTestFramework("TerraformInstall")
    tf.Start(t, "Test Terraform installation and verification")

    success := true
    defer tf.Finish(t, success)

    // Test installation
    tf.Step(t, "Install Terraform")
    err := InstallTerraform("latest")
    if err != nil {
        tf.Error(t, "Installation failed", err.Error())
        success = false
        return
    }

    // Verify installation
    tf.Step(t, "Verify Terraform installation")
    version, err := GetInstalledTerraformVersion()
    if err != nil {
        tf.Error(t, "Version check failed", err.Error())
        success = false
        return
    }

    tf.Success(t, fmt.Sprintf("Terraform %s installed successfully", version))

    // Test basic functionality
    tf.Step(t, "Test Terraform init")
    cmd := exec.Command("terraform", "init", "-help")
    if err := cmd.Run(); err != nil {
        tf.Error(t, "Terraform init help failed", err.Error())
        success = false
    }
}
```

### Cross-Platform Testing
```bash
# Windows testing
portunix test terraform --platform windows --arch amd64
portunix test terraform --platform windows --arch 386

# Linux testing
portunix test terraform --platform linux --arch amd64
portunix test terraform --platform linux --arch arm64

# macOS testing
portunix test terraform --platform darwin --arch amd64
portunix test terraform --platform darwin --arch arm64
```

## Error Handling

### Common Error Scenarios
1. **Network Issues**: Retry with exponential backoff
2. **Checksum Mismatch**: Verify download integrity
3. **Permission Denied**: Elevate privileges or use user directory
4. **PATH Update Failed**: Provide manual instructions
5. **Existing Installation**: Offer upgrade/downgrade options

### Error Messages
```go
var ErrMessages = map[string]string{
    "download_failed": "Failed to download Terraform. Please check your internet connection.",
    "extraction_failed": "Failed to extract Terraform. The download may be corrupted.",
    "permission_denied": "Permission denied. Try running with administrator privileges.",
    "path_update_failed": "Failed to update PATH. Please add %s to your PATH manually.",
    "version_not_found": "Terraform version %s not found. Use 'portunix list terraform --versions' to see available versions.",
}
```

## Documentation Requirements

### User Documentation
- Installation guide for each platform
- Version management documentation
- Troubleshooting guide
- Integration with cloud providers

### Developer Documentation
- Package definition format
- Adding new HashiCorp tools
- Version update process
- Testing procedures

## Security Considerations

### Checksum Verification
```go
func VerifyTerraformChecksum(filePath, version string) error {
    // Download checksum file
    checksumURL := fmt.Sprintf(
        "https://releases.hashicorp.com/terraform/%s/terraform_%s_SHA256SUMS",
        version, version,
    )

    checksums, err := downloadChecksums(checksumURL)
    if err != nil {
        return err
    }

    // Calculate file checksum
    fileChecksum, err := calculateSHA256(filePath)
    if err != nil {
        return err
    }

    // Verify checksum
    expectedChecksum := getChecksumForFile(checksums, filepath.Base(filePath))
    if fileChecksum != expectedChecksum {
        return fmt.Errorf("checksum mismatch: expected %s, got %s",
            expectedChecksum, fileChecksum)
    }

    return nil
}
```

### GPG Signature Verification (Optional)
```go
func VerifyTerraformSignature(filePath, version string) error {
    // Download signature file
    sigURL := fmt.Sprintf(
        "https://releases.hashicorp.com/terraform/%s/terraform_%s_SHA256SUMS.sig",
        version, version,
    )

    // Verify with HashiCorp's public key
    return verifyGPGSignature(filePath, sigURL, hashicorpPublicKey)
}
```

## Acceptance Criteria
- [ ] `portunix install terraform` successfully installs latest version
- [ ] Specific version installation works with `--variant` flag
- [ ] Terraform binary is correctly placed in PATH
- [ ] Installation works on Windows, Linux, and macOS
- [ ] Version detection correctly identifies installed version
- [ ] Update functionality works when new versions are available
- [ ] Checksum verification prevents corrupted downloads
- [ ] Proper error messages for all failure scenarios
- [ ] Integration tests pass on all platforms
- [ ] Documentation is complete and accurate
- [ ] ZIP extraction handles all Terraform archive formats
- [ ] PATH management works without requiring restart

## Priority
**High** - Terraform is a critical Infrastructure as Code tool widely used in DevOps and cloud infrastructure management

## Labels
- enhancement
- package-management
- terraform
- hashicorp
- infrastructure-as-code
- multi-platform
- devops

## Related Issues
- General package installation framework
- HashiCorp tools support (Vault, Consul, Nomad, Packer)
- ZIP archive extraction support
- PATH management improvements

## Estimated Effort
- **Package Definition**: 2-3 hours
- **Installation Logic**: 4-6 hours
- **Version Management**: 3-4 hours
- **Testing**: 4-5 hours
- **Documentation**: 2 hours
- **Total**: 15-20 hours

---

**Created**: 2025-09-24
**Status**: ✅ Implemented
**Closed**: 2026-05-09
**Priority**: High
**Type**: Enhancement