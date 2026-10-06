# Issue #66: Double Commander Installation Support

## Summary
Add support for installing Double Commander through Portunix package management system. Double Commander is a popular open-source dual-panel file manager that provides advanced file operations and cross-platform compatibility.

## Current Behavior
- `portunix install double-commander` is not supported
- Users must manually download and install Double Commander
- No integration with Portunix package management system
- No version management or update capabilities through Portunix

## Expected Behavior
```bash
# Install latest stable version
portunix install double-commander

# Install specific variant
portunix install double-commander --variant full     # With plugins
portunix install double-commander --variant portable # Portable version (Windows)
portunix install double-commander --variant qt       # Qt version (Linux)
portunix install double-commander --variant gtk      # GTK version (Linux)

# Verify installation
portunix verify double-commander
doublecmd --version

# Update double commander
portunix update double-commander
```

## Expected Output
```
════════════════════════════════════════════════
📦 INSTALLING: Double Commander
════════════════════════════════════════════════
📄 Description: Cross-platform dual-panel file manager
🔧 Variant: latest (v1.1.15)
💻 Platform: windows
🏗️  Installation type: exe
🌐 Download URL: https://sourceforge.net/projects/doublecmd/files/DC%201.1.15/doublecmd-1.1.15.i386-win32.exe
════════════════════════════════════════════════
🔍 Checking if double-commander is already installed...
📋 double-commander is not installed, proceeding with installation...
🚀 Starting installation...
Downloading doublecmd-1.1.15.i386-win32.exe...
✅ Downloaded: .cache\doublecmd-1.1.15.i386-win32.exe
Running installer...
✅ Double Commander v1.1.15 installed successfully
Adding to PATH...
✅ PATH updated successfully
════════════════════════════════════════════════
🎉 Double Commander v1.1.15 installed successfully!
📋 Verify with: doublecmd --version
📋 Launch with: doublecmd
════════════════════════════════════════════════
```

## Implementation Requirements

### Package Definition
```json
{
  "double-commander": {
    "name": "Double Commander",
    "description": "Cross-platform dual-panel file manager",
    "category": "file-manager",
    "homepage": "https://doublecmd.sourceforge.io/",
    "license": "GPL-2.0+",
    "maintainer": "Alexander Koblov",
    "variants": {
      "latest": {
        "version": "1.1.15",
        "description": "Latest stable version"
      },
      "full": {
        "version": "1.1.15",
        "description": "Full version with all plugins",
        "includes": ["plugins", "additional_tools"]
      },
      "portable": {
        "version": "1.1.15",
        "description": "Portable version (Windows only)",
        "type": "portable"
      },
      "qt": {
        "version": "1.1.15",
        "description": "Qt version (Linux)",
        "ui_toolkit": "qt"
      },
      "gtk": {
        "version": "1.1.15",
        "description": "GTK version (Linux)",
        "ui_toolkit": "gtk"
      }
    },
    "platforms": {
      "windows": {
        "architecture": {
          "amd64": {
            "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.x86_64-win64.exe",
            "installer": "exe",
            "install_path": "{program_files}\\Double Commander",
            "binary": "doublecmd.exe",
            "silent_args": "/SILENT /NORESTART"
          },
          "386": {
            "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.i386-win32.exe",
            "installer": "exe",
            "install_path": "{program_files}\\Double Commander",
            "binary": "doublecmd.exe",
            "silent_args": "/SILENT /NORESTART"
          }
        },
        "portable": {
          "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.i386-win32.zip",
          "installer": "zip",
          "install_path": "{portunix_apps}\\doublecmd",
          "binary": "doublecmd.exe"
        }
      },
      "linux": {
        "architecture": {
          "amd64": {
            "debian": {
              "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd_{version}-1_amd64.deb",
              "installer": "deb",
              "binary": "doublecmd"
            },
            "rpm": {
              "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}-1.x86_64.rpm",
              "installer": "rpm",
              "binary": "doublecmd"
            },
            "tar": {
              "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.x86_64-linux-qt.tar.xz",
              "installer": "tar",
              "install_path": "/opt/doublecmd",
              "binary": "doublecmd"
            }
          }
        }
      },
      "darwin": {
        "architecture": {
          "amd64": {
            "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.x86_64-darwin.dmg",
            "installer": "dmg",
            "install_path": "/Applications",
            "binary": "Double Commander.app/Contents/MacOS/doublecmd"
          },
          "arm64": {
            "url": "https://sourceforge.net/projects/doublecmd/files/DC%20{version}/doublecmd-{version}.aarch64-darwin.dmg",
            "installer": "dmg",
            "install_path": "/Applications",
            "binary": "Double Commander.app/Contents/MacOS/doublecmd"
          }
        }
      }
    },
    "post_install": {
      "commands": [
        "doublecmd --version"
      ],
      "desktop_entry": {
        "name": "Double Commander",
        "comment": "Dual-panel file manager",
        "icon": "doublecmd",
        "categories": ["System", "FileManager"]
      }
    },
    "verification": {
      "command": "doublecmd --version",
      "expected_output": "Double Commander {version}"
    }
  }
}
```

### SourceForge Download Handler
```go
// app/install/sourceforge.go
package install

import (
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

type SourceForgeDownloader struct {
    BaseURL string
    Client  *http.Client
}

func NewSourceForgeDownloader() *SourceForgeDownloader {
    return &SourceForgeDownloader{
        BaseURL: "https://sourceforge.net",
        Client:  &http.Client{},
    }
}

func (sf *SourceForgeDownloader) GetDirectDownloadURL(projectURL string) (string, error) {
    // SourceForge redirects to mirrors, we need to handle this
    resp, err := sf.Client.Head(projectURL)
    if err != nil {
        return "", err
    }

    // Follow redirects to get actual download URL
    finalURL := resp.Request.URL.String()

    // Add download parameter to force download
    u, err := url.Parse(finalURL)
    if err != nil {
        return "", err
    }

    q := u.Query()
    q.Set("use_mirror", "autoselect")
    u.RawQuery = q.Encode()

    return u.String(), nil
}

func (sf *SourceForgeDownloader) DownloadFile(projectURL, destPath string) error {
    directURL, err := sf.GetDirectDownloadURL(projectURL)
    if err != nil {
        return err
    }

    return DownloadFile(directURL, destPath)
}
```

### Version Detection
```go
// app/install/double_commander.go
func GetInstalledDoubleCommanderVersion() (string, error) {
    cmd := exec.Command("doublecmd", "--version")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    versionOutput := string(output)

    // Parse version from output like "Double Commander 1.1.15"
    lines := strings.Split(versionOutput, "\n")
    for _, line := range lines {
        if strings.Contains(line, "Double Commander") {
            parts := strings.Fields(line)
            if len(parts) >= 3 {
                return parts[2], nil
            }
        }
    }

    return "", fmt.Errorf("could not parse version from output: %s", versionOutput)
}

func CheckDoubleCommanderInstallation() error {
    version, err := GetInstalledDoubleCommanderVersion()
    if err != nil {
        return err
    }

    fmt.Printf("Double Commander version %s is installed\n", version)
    return nil
}
```

### Desktop Integration (Linux)
```go
func CreateDoubleCommanderDesktopEntry(installPath string) error {
    desktopEntry := `[Desktop Entry]
Version=1.0
Type=Application
Name=Double Commander
Comment=Dual-panel file manager
Exec=%s/doublecmd
Icon=doublecmd
Terminal=false
Categories=System;FileManager;
MimeType=inode/directory;
`

    desktopContent := fmt.Sprintf(desktopEntry, installPath)

    // Write to applications directory
    appsDir := filepath.Join(os.Getenv("HOME"), ".local/share/applications")
    if err := os.MkdirAll(appsDir, 0755); err != nil {
        return err
    }

    desktopFile := filepath.Join(appsDir, "doublecmd.desktop")
    return ioutil.WriteFile(desktopFile, []byte(desktopContent), 0644)
}
```

### Windows Registry Integration
```go
func AddDoubleCommanderToWindowsRegistry(installPath string) error {
    // Add to Windows registry for context menu integration
    key, err := registry.OpenKey(registry.CURRENT_USER,
        `SOFTWARE\Classes\Directory\shell\DoubleCMD`, registry.CREATE_SUB_KEY)
    if err != nil {
        return err
    }
    defer key.Close()

    err = key.SetStringValue("", "Open with Double Commander")
    if err != nil {
        return err
    }

    commandKey, err := registry.OpenKey(registry.CURRENT_USER,
        `SOFTWARE\Classes\Directory\shell\DoubleCMD\command`, registry.CREATE_SUB_KEY)
    if err != nil {
        return err
    }
    defer commandKey.Close()

    cmdValue := fmt.Sprintf(`"%s\doublecmd.exe" "%%1"`, installPath)
    return commandKey.SetStringValue("", cmdValue)
}
```

## Platform-Specific Handling

### Windows Installation
```go
func InstallDoubleCommanderWindows(variant string) error {
    switch variant {
    case "portable":
        return installDoubleCommanderPortable()
    default:
        return installDoubleCommanderEXE()
    }
}

func installDoubleCommanderEXE() error {
    // Download and run installer
    url := getDoubleCommanderURL("windows", "exe")
    tempFile := filepath.Join(os.TempDir(), "doublecmd-installer.exe")

    if err := DownloadFile(url, tempFile); err != nil {
        return err
    }
    defer os.Remove(tempFile)

    // Run silent installation
    cmd := exec.Command(tempFile, "/SILENT", "/NORESTART")
    return cmd.Run()
}

func installDoubleCommanderPortable() error {
    // Download and extract portable version
    url := getDoubleCommanderURL("windows", "zip")
    return ExtractAndInstall(url, GetPortableInstallPath("doublecmd"))
}
```

### Linux Package Management
```go
func InstallDoubleCommanderLinux() error {
    // Try distribution-specific package managers first
    if hasPackageManager("apt") {
        return installViaAPT()
    } else if hasPackageManager("dnf") {
        return installViaDNF()
    } else if hasPackageManager("pacman") {
        return installViaPacman()
    } else {
        // Fall back to binary installation
        return installDoubleCommanderBinary()
    }
}

func installViaAPT() error {
    // Try official repository first
    cmd := exec.Command("sudo", "apt", "install", "-y", "doublecmd-qt")
    if err := cmd.Run(); err == nil {
        return nil
    }

    // Fall back to manual installation
    return installDoubleCommanderBinary()
}
```

### macOS DMG Handling
```go
func InstallDoubleCommanderMacOS() error {
    url := getDoubleCommanderURL("darwin", "dmg")
    tempFile := filepath.Join(os.TempDir(), "doublecmd.dmg")

    // Download DMG
    if err := DownloadFile(url, tempFile); err != nil {
        return err
    }
    defer os.Remove(tempFile)

    // Mount DMG
    mountPoint, err := MountDMG(tempFile)
    if err != nil {
        return err
    }
    defer UnmountDMG(mountPoint)

    // Copy application
    srcApp := filepath.Join(mountPoint, "Double Commander.app")
    dstApp := "/Applications/Double Commander.app"

    return CopyDirectory(srcApp, dstApp)
}
```

## Testing Requirements

### Unit Tests
```go
func TestDoubleCommanderVersionParsing(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Double Commander 1.1.15", "1.1.15"},
        {"Double Commander 1.1.14\n", "1.1.14"},
        {"Version: Double Commander 1.1.13", "1.1.13"},
    }

    for _, test := range tests {
        result := parseDoubleCommanderVersion(test.input)
        if result != test.expected {
            t.Errorf("Expected %s, got %s", test.expected, result)
        }
    }
}
```

### Integration Tests
```go
func TestDoubleCommanderInstallation(t *testing.T) {
    tf := testframework.NewTestFramework("DoubleCommanderInstall")
    tf.Start(t, "Test Double Commander installation and verification")

    success := true
    defer tf.Finish(t, success)

    // Test installation
    tf.Step(t, "Install Double Commander")
    err := InstallDoubleCommander("latest")
    if err != nil {
        tf.Error(t, "Installation failed", err.Error())
        success = false
        return
    }

    // Verify installation
    tf.Step(t, "Verify Double Commander installation")
    version, err := GetInstalledDoubleCommanderVersion()
    if err != nil {
        tf.Error(t, "Version check failed", err.Error())
        success = false
        return
    }

    tf.Success(t, fmt.Sprintf("Double Commander %s installed successfully", version))

    // Test basic functionality
    tf.Step(t, "Test Double Commander help")
    cmd := exec.Command("doublecmd", "--help")
    if err := cmd.Run(); err != nil {
        tf.Error(t, "Double Commander help failed", err.Error())
        success = false
    }
}
```

### Cross-Platform Testing
```bash
# Windows testing
portunix test double-commander --platform windows --arch amd64
portunix test double-commander --platform windows --arch 386 --variant portable

# Linux testing
portunix test double-commander --platform linux --arch amd64 --variant qt
portunix test double-commander --platform linux --arch amd64 --variant gtk

# macOS testing
portunix test double-commander --platform darwin --arch amd64
portunix test double-commander --platform darwin --arch arm64
```

## Error Handling

### Common Error Scenarios
1. **Download Issues**: SourceForge mirror problems
2. **Permission Denied**: Installation requires admin rights
3. **Dependency Missing**: GTK+/Qt libraries not available (Linux)
4. **Archive Corruption**: Download corruption detection
5. **Path Issues**: Installation directory conflicts

### Error Messages
```go
var ErrMessages = map[string]string{
    "download_failed": "Failed to download Double Commander. Please check your internet connection and try again.",
    "sourceforge_mirror": "SourceForge mirror unavailable. Trying alternative download source.",
    "permission_denied": "Permission denied. Run as administrator or use --user-install flag.",
    "missing_dependencies": "Missing dependencies. Install GTK+ or Qt development libraries.",
    "version_not_found": "Double Commander version %s not found. Use 'portunix list double-commander --versions' to see available versions.",
    "binary_not_found": "Double Commander binary not found after installation. Check installation path.",
}
```

## Documentation Requirements

### User Documentation
- Installation guide for each platform
- Configuration and customization guide
- Plugin installation instructions
- Troubleshooting common issues

### Developer Documentation
- SourceForge download integration
- Cross-platform installation handling
- Desktop environment integration
- Testing procedures

## Security Considerations

### Download Verification
```go
func VerifyDoubleCommanderDownload(filePath string) error {
    // SourceForge doesn't provide checksums, but we can verify file signature
    // Check file size and basic integrity
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return err
    }

    // Basic file size validation
    if fileInfo.Size() < 1024*1024 { // Less than 1MB is suspicious
        return fmt.Errorf("downloaded file too small: %d bytes", fileInfo.Size())
    }

    // For executables, check PE/ELF signature
    return VerifyExecutableSignature(filePath)
}
```

## Acceptance Criteria
- [ ] `portunix install double-commander` successfully installs latest version
- [ ] Variant installation works (full, portable, qt, gtk)
- [ ] Double Commander binary is correctly accessible from command line
- [ ] Installation works on Windows, Linux, and macOS
- [ ] Version detection correctly identifies installed version
- [ ] Desktop integration works (shortcuts, context menu)
- [ ] Update functionality works when new versions are available
- [ ] Proper error messages for all failure scenarios
- [ ] Integration tests pass on all platforms
- [ ] Documentation is complete and accurate
- [ ] SourceForge download handling works reliably
- [ ] Cross-platform installation handling works correctly

## Priority
**Medium** - Double Commander is a useful file manager for power users and system administrators

## Labels
- enhancement
- package-management
- double-commander
- file-manager
- sourceforge
- cross-platform
- gui-application

## Related Issues
- General package installation framework improvements
- SourceForge download handler implementation
- Cross-platform GUI application support
- Desktop integration enhancements

## Estimated Effort
- **Package Definition**: 3-4 hours
- **SourceForge Integration**: 4-6 hours
- **Platform-specific Installation**: 6-8 hours
- **Desktop Integration**: 3-4 hours
- **Version Detection**: 2-3 hours
- **Testing**: 5-6 hours
- **Documentation**: 2-3 hours
- **Total**: 25-34 hours

---

**Created**: 2025-09-24
**Status**: ✅ Implemented
**Implemented**: 2026-05-09
**Priority**: Medium
**Type**: Enhancement