# Issue #64: Visual Studio Code Installation Filename Issue

## Summary
The command `portunix install vscode` fails because the downloaded file is saved as `.cache\download` instead of `.cache\download.exe` or proper filename. This causes the installation to fail with error "executable file not found in %PATH%" when trying to execute the downloaded installer.

## Current Behavior
```
════════════════════════════════════════════════
📦 INSTALLING: Visual Studio Code
════════════════════════════════════════════════
📄 Description: Source code editor from Microsoft
🔧 Variant: user (vlatest)
💻 Platform: windows_sandbox
🏗️  Installation type: exe
🌐 Download URL: https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user
════════════════════════════════════════════════
🔍 Checking if vscode is already installed...
📋 vscode is not installed, proceeding with installation...
🚀 Starting installation...
Downloading download...
Downloading https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user...
✅ Downloaded: .cache\download                    ← WRONG: Missing .exe extension
Installing download...

❌ Installation FAILED!
Package: Visual Studio Code
Variant: user
Error: exec: ".cache\\download": executable file not found in %PATH%
Error installing package 'vscode': exec: ".cache\\download": executable file not found in %PATH%
```

**Downloaded file location**: `C:\portunix\portunix_1.7.2\.cache\download`
**Expected file location**: `C:\portunix\portunix_1.7.2\.cache\VSCodeUserSetup-x64-1.95.3.exe`

## Expected Behavior
```
════════════════════════════════════════════════
📦 INSTALLING: Visual Studio Code
════════════════════════════════════════════════
📄 Description: Source code editor from Microsoft
🔧 Variant: user (vlatest)
💻 Platform: windows
🏗️  Installation type: exe
🌐 Download URL: https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user
════════════════════════════════════════════════
🔍 Checking if vscode is already installed...
📋 vscode is not installed, proceeding with installation...
🚀 Starting installation...
Downloading VSCodeUserSetup-x64-1.95.3.exe...
Downloading https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user...
✅ Downloaded: .cache\VSCodeUserSetup-x64-1.95.3.exe
Installing VSCodeUserSetup-x64-1.95.3.exe...
✅ Visual Studio Code installed successfully
```

## Problem Analysis

### Issue 1: Incorrect Filename Resolution
- **Current**: Downloads as generic "download" filename
- **Root Cause**: URL redirects to actual file, but filename not extracted from final URL or Content-Disposition header
- **Impact**: Downloaded file has no extension and cannot be executed

### Issue 2: Missing Content-Disposition Header Handling
- **Problem**: Microsoft's download URL uses redirect with proper filename in final response
- **Missing**: HTTP header parsing for `Content-Disposition: attachment; filename="VSCodeUserSetup-x64-1.95.3.exe"`
- **Alternative**: Extract filename from final redirected URL

### Issue 3: Platform Detection Issue
- **Current**: Shows "windows_sandbox" instead of "windows"
- **Secondary Issue**: Related to platform detection system
- **Impact**: May affect package selection logic

## Technical Investigation Required

### Download URL Analysis
```bash
# Test the actual download URL
curl -I "https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user"

# Expected response headers:
# HTTP/1.1 302 Found
# Location: https://az764295.vo.msecnd.net/stable/.../VSCodeUserSetup-x64-1.95.3.exe
#
# Final URL response:
# HTTP/1.1 200 OK
# Content-Disposition: attachment; filename="VSCodeUserSetup-x64-1.95.3.exe"
# Content-Type: application/octet-stream
```

### Filename Resolution Methods
1. **Content-Disposition Header**: Parse `filename=` parameter
2. **Final URL Path**: Extract filename from redirected URL
3. **Predefined Mapping**: Map package to expected filename pattern
4. **File Type Detection**: Add .exe extension for Windows executables

## Affected Components
- `app/install/downloader.go` - HTTP download functionality
- `app/install/exe_installer.go` - Windows executable installer
- Download URL resolution and filename extraction
- File caching system in `.cache/` directory

## Reproduction Steps
1. Run `portunix install vscode` on Windows
2. Observe download URL and filename resolution
3. Check downloaded file in `.cache/` directory
4. Note missing .exe extension and installation failure

## Implementation Requirements

### Enhanced Download Function
```go
// app/install/downloader.go
func DownloadFile(url string, destDir string) (string, error) {
    // Follow redirects and get final URL
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Extract filename from Content-Disposition or final URL
    filename := extractFilename(resp)
    if filename == "" {
        filename = generateFilename(url, resp.Header.Get("Content-Type"))
    }

    destPath := filepath.Join(destDir, filename)
    return destPath, saveFile(resp.Body, destPath)
}

func extractFilename(resp *http.Response) string {
    // Method 1: Content-Disposition header
    disposition := resp.Header.Get("Content-Disposition")
    if filename := parseContentDisposition(disposition); filename != "" {
        return filename
    }

    // Method 2: Final URL path
    if resp.Request != nil && resp.Request.URL != nil {
        path := resp.Request.URL.Path
        if filename := filepath.Base(path); filename != "" && filename != "." {
            return filename
        }
    }

    return ""
}

func parseContentDisposition(disposition string) string {
    // Parse: attachment; filename="VSCodeUserSetup-x64-1.95.3.exe"
    if !strings.Contains(disposition, "filename=") {
        return ""
    }

    parts := strings.Split(disposition, "filename=")
    if len(parts) < 2 {
        return ""
    }

    filename := strings.TrimSpace(parts[1])
    filename = strings.Trim(filename, `"`)
    return filename
}
```

### Package Definition Enhancement
```json
{
  "vscode": {
    "name": "Visual Studio Code",
    "description": "Source code editor from Microsoft",
    "category": "development",
    "variants": {
      "user": {
        "description": "User installation (no admin required)",
        "installer": "exe",
        "platforms": {
          "windows": {
            "url": "https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user",
            "filename_pattern": "VSCodeUserSetup-x64-*.exe",
            "silent_args": ["/VERYSILENT", "/MERGETASKS=!runcode"]
          }
        }
      },
      "system": {
        "description": "System installation (requires admin)",
        "installer": "exe",
        "platforms": {
          "windows": {
            "url": "https://code.visualstudio.com/sha/download?build=stable&os=win32-x64",
            "filename_pattern": "VSCodeSetup-x64-*.exe",
            "silent_args": ["/VERYSILENT", "/MERGETASKS=!runcode"]
          }
        }
      }
    }
  }
}
```

### Fallback Filename Generation
```go
func generateFilename(url string, contentType string) string {
    // Extract base filename from URL
    parsedURL, err := url.Parse(url)
    if err == nil {
        if base := filepath.Base(parsedURL.Path); base != "." && base != "" {
            return base
        }
    }

    // Generate filename based on content type
    switch contentType {
    case "application/octet-stream", "application/x-msdownload":
        return fmt.Sprintf("download_%d.exe", time.Now().Unix())
    case "application/zip":
        return fmt.Sprintf("download_%d.zip", time.Now().Unix())
    default:
        return fmt.Sprintf("download_%d", time.Now().Unix())
    }
}
```

## Error Handling Improvements

### Better Error Messages
```go
func (e *ExeInstaller) Install(packageName string, filePath string) error {
    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return fmt.Errorf("downloaded file not found: %s", filePath)
    }

    // Check if file is executable
    if !strings.HasSuffix(strings.ToLower(filePath), ".exe") {
        return fmt.Errorf("downloaded file is not a Windows executable: %s", filePath)
    }

    // Verify file is not empty
    info, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("cannot access downloaded file: %v", err)
    }
    if info.Size() == 0 {
        return fmt.Errorf("downloaded file is empty: %s", filePath)
    }

    // Execute installer
    cmd := exec.Command(filePath, e.getSilentArgs()...)
    return cmd.Run()
}
```

### Download Validation
```go
func validateDownload(filePath string, expectedPattern string) error {
    filename := filepath.Base(filePath)

    // Check filename pattern
    if expectedPattern != "" {
        matched, err := filepath.Match(expectedPattern, filename)
        if err != nil || !matched {
            return fmt.Errorf("downloaded filename %s doesn't match expected pattern %s", filename, expectedPattern)
        }
    }

    // Verify file signature for executables
    if strings.HasSuffix(strings.ToLower(filename), ".exe") {
        return validateExecutableSignature(filePath)
    }

    return nil
}

func validateExecutableSignature(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Read first few bytes to check for MZ header
    header := make([]byte, 2)
    _, err = file.Read(header)
    if err != nil {
        return err
    }

    if string(header) != "MZ" {
        return fmt.Errorf("file is not a valid Windows executable")
    }

    return nil
}
```

## Testing Requirements

### Unit Tests
```go
func TestFilenameExtraction(t *testing.T) {
    tests := []struct {
        name                string
        contentDisposition  string
        finalURL           string
        expectedFilename   string
    }{
        {
            name:               "Content-Disposition with quotes",
            contentDisposition: `attachment; filename="VSCodeUserSetup-x64-1.95.3.exe"`,
            expectedFilename:   "VSCodeUserSetup-x64-1.95.3.exe",
        },
        {
            name:               "Content-Disposition without quotes",
            contentDisposition: "attachment; filename=VSCodeUserSetup-x64-1.95.3.exe",
            expectedFilename:   "VSCodeUserSetup-x64-1.95.3.exe",
        },
        {
            name:             "Final URL path",
            finalURL:         "https://example.com/path/installer.exe",
            expectedFilename: "installer.exe",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test filename extraction logic
        })
    }
}
```

### Integration Tests
```go
func TestVSCodeDownloadAndInstall(t *testing.T) {
    tf := testframework.NewTestFramework("VSCodeInstallation")
    tf.Start(t, "Test Visual Studio Code download and installation")

    success := true
    defer tf.Finish(t, success)

    // Test download
    tf.Step(t, "Download VS Code installer")
    url := "https://code.visualstudio.com/sha/download?build=stable&os=win32-x64-user"
    filename, err := downloadFile(url, ".cache")

    if err != nil {
        tf.Error(t, "Download failed", err.Error())
        success = false
        return
    }

    tf.Success(t, "Downloaded: " + filename)

    // Verify filename
    tf.Step(t, "Verify downloaded filename")
    if !strings.HasSuffix(filename, ".exe") {
        tf.Error(t, "Downloaded file missing .exe extension")
        success = false
        return
    }

    tf.Success(t, "Filename validation passed")
}
```

### Manual Testing Scenarios
1. **Fresh Installation**: Test on clean Windows VM
2. **Network Variations**: Test with different network conditions
3. **Redirect Handling**: Verify redirect following works correctly
4. **File Validation**: Test various file validation scenarios
5. **Error Recovery**: Test behavior when downloads fail or are corrupted

## Additional Issues to Address

### Related Platform Detection Issue
- Fix "windows_sandbox" vs "windows" platform detection
- Ensure consistent platform reporting across all commands

### Download Progress Indication
```go
// Enhanced progress reporting
func downloadWithProgress(url string, destPath string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Create progress reader
    size := resp.ContentLength
    counter := &ProgressCounter{Total: size}
    _, err = io.Copy(out, io.TeeReader(resp.Body, counter))

    return err
}

type ProgressCounter struct {
    Total     int64
    Downloaded int64
}

func (pc *ProgressCounter) Write(p []byte) (int, error) {
    n := len(p)
    pc.Downloaded += int64(n)
    pc.printProgress()
    return n, nil
}
```

## Acceptance Criteria
- [ ] VS Code downloads with correct filename (includes .exe extension)
- [ ] Content-Disposition header is parsed correctly
- [ ] Final redirect URL is followed and filename extracted
- [ ] Downloaded file is validated before installation attempt
- [ ] Installation executes successfully with proper silent arguments
- [ ] Error messages are clear and actionable
- [ ] Platform detection shows "windows" instead of "windows_sandbox"
- [ ] Progress indication works during download
- [ ] File validation prevents execution of corrupted downloads
- [ ] Tests cover all filename extraction scenarios
- [ ] Cross-platform compatibility maintained (other platforms still work)

## Priority
**High** - Critical functionality affecting one of the most popular development tools

## Labels
- critical
- bug
- installation
- download
- filename-resolution
- vscode
- windows
- exe-installer

## Related Issues
- Platform detection inconsistencies (windows_sandbox vs windows)
- General download system improvements
- HTTP redirect handling
- File validation system

## Estimated Effort
- **Analysis & Investigation**: 1-2 hours
- **Download System Enhancement**: 4-6 hours
- **Filename Resolution Logic**: 3-4 hours
- **File Validation**: 2-3 hours
- **Testing & Validation**: 3-4 hours
- **Error Handling**: 1-2 hours
- **Documentation Updates**: 1 hour
- **Total**: 15-22 hours

---

**Created**: 2025-09-24
**Status**: 📋 Open
**Priority**: High
**Type**: Bug Fix