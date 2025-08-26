# Issue #10: Self-Update Command - Automatic Binary Updates from GitHub

**Status:** ✅ Implemented  
**Priority:** High  
**Labels:** enhancement, self-update, cross-platform  
**Milestone:** v1.4.0  

## Feature Request: Self-Update Command & Installation System

### Overview
Implement a self-update mechanism that allows Portunix to automatically update itself to the latest version from GitHub releases. Additionally, provide installation scripts that utilize Portunix's built-in installation command for a seamless setup experience. Users should be able to run `portunix update` to fetch and install the latest release, ensuring they always have access to the newest features and bug fixes.

### Requirements

#### 1. Command Line Interface
**Update Commands:**
- `portunix update` - Check for updates and install if available
- `portunix update --check` - Only check for updates without installing
- `portunix update --force` - Force update even if on latest version
- `portunix --version` - Display current version information

**Installation Commands:**
- `portunix install` - Interactive installation wizard
- `portunix install --silent` - Silent installation with defaults
- `portunix install --path /usr/local/bin` - Install to specific location
- `portunix install --create-config` - Create configuration files
- `portunix install --add-to-path` - Add to system PATH

#### 2. Update Capabilities
**Version Management:**
- Embedded version in binary using `-ldflags`
- Semantic versioning (SemVer) support
- Version comparison logic

**Update Process:**
- Query GitHub API for latest release
- Compare current version with latest
- Download appropriate binary for OS/arch
- Verify integrity (SHA256 checksum)
- Atomic replacement of running binary
- Rollback capability on failure

**Security:**
- SHA256 checksum verification
- Optional GPG signature verification
- HTTPS-only downloads
- Backup current binary before update

#### 3. Implementation Details

**Build Configuration:**
```bash
go build -ldflags "-X main.version=v1.4.2 -s -w" -o portunix .
```

**Self-Update Libraries (candidates):**
- `github.com/rhymond/go-github-selfupdate` - GitHub releases integration
- `github.com/inconshreveable/go-update` - Low-level update mechanics
- `github.com/tj/go-update` - Alternative implementation

**Update Flow:**
1. Get current version from embedded variable
2. Query GitHub API for latest release (`GET /repos/cassandragargoyle/Portunix/releases/latest`)
3. Compare versions using SemVer
4. Download correct asset based on `runtime.GOOS` and `runtime.GOARCH`
5. Verify SHA256 checksum
6. Create backup of current binary
7. Perform atomic replacement
8. Verify new binary works
9. Clean up backup on success

#### 4. Command Examples

**Basic Update:**
```bash
$ portunix update
Current version: v1.3.0
Checking for updates...
✓ New version available: v1.4.2
✓ Downloading portunix-v1.4.2-linux-amd64...
✓ Verifying checksum...
✓ Creating backup...
✓ Installing update...
✓ Update completed successfully!

Portunix has been updated from v1.3.0 to v1.4.2
```

**Check Only:**
```bash
$ portunix update --check
Current version: v1.3.0
Latest version: v1.4.2
Update available! Run 'portunix update' to install.
```

**Already Up-to-date:**
```bash
$ portunix update
Current version: v1.4.2
✓ You are running the latest version!
```

**Force Update:**
```bash
$ portunix update --force
Current version: v1.4.2
⚠ Forcing reinstall of v1.4.2...
✓ Download completed
✓ Update completed successfully!
```

#### 5. Installation System

**Installation Scripts:**
Each release archive will include an installation script:
- **Windows:** `install.ps1` - PowerShell script
- **Linux/macOS:** `install.sh` - Shell script

**Installation Process:**
1. User downloads and extracts the archive
2. Runs the installation script (`./install.sh` or `.\install.ps1`)
3. Script executes `portunix install` with appropriate parameters
4. Portunix performs interactive or silent installation

**Interactive Installation Flow:**
```
$ ./install.sh
Welcome to Portunix Installation!
Version: v1.4.2

[1] Install to /usr/local/bin (recommended)
[2] Install to ~/bin
[3] Install to custom location
[4] Cancel installation

Please select [1-4]: 1

✓ Checking permissions...
✓ Creating backup of existing installation...
✓ Installing portunix to /usr/local/bin...
✓ Verifying installation...
✓ Adding to PATH...

Installation completed successfully!
Run 'portunix --version' to verify.
```

**Silent Installation:**
```bash
# Linux/macOS
./install.sh --silent

# Windows
.\install.ps1 -Silent
```

#### 6. GitHub Release Structure

**Expected Release Assets:**
```
portunix_1.4.2_windows_amd64.zip
  ├── portunix.exe
  └── install.ps1
  
portunix_1.4.2_linux_amd64.tar.gz
  ├── portunix
  └── install.sh
  
portunix_1.4.2_darwin_amd64.tar.gz
  ├── portunix
  └── install.sh
  
checksums_1.4.2.txt
```

**Release Automation (GoReleaser):**
```yaml
# .goreleaser.yml
builds:
  - binary: portunix
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -X main.version={{.Version}}
      - -s -w

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - src: scripts/install.sh
        dst: install.sh
        info:
          mode: 0755
      - src: scripts/install.ps1
        dst: install.ps1

checksum:
  name_template: "checksums_{{ .Version }}.txt"
```

#### 6. Error Handling

**Network Issues:**
```bash
$ portunix update
Error: Unable to check for updates
  Failed to connect to GitHub API
  Please check your internet connection and try again
```

**Permission Issues:**
```bash
$ portunix update
Error: Permission denied
  Cannot write to /usr/local/bin/portunix
  Try running with sudo: sudo portunix update
```

**Verification Failed:**
```bash
$ portunix update
Error: Checksum verification failed
  Downloaded file does not match expected checksum
  This could indicate a corrupted download or security issue
  Update aborted for safety
```

#### 7. Implementation Structure

**Files to Create:**
```
cmd/
├── update.go          # Update command implementation
├── install.go         # Install command implementation
└── version.go         # Version command implementation

app/
├── update/
│   ├── update.go      # Core update logic
│   ├── github.go      # GitHub API integration
│   ├── verify.go      # Checksum verification
│   └── update_test.go # Tests
└── install/
    ├── install.go     # Core installation logic
    ├── interactive.go # Interactive installation wizard
    ├── path.go        # PATH management
    └── install_test.go # Tests

scripts/
├── install.sh         # Linux/macOS installation script
└── install.ps1        # Windows installation script
```

**Key Functions:**
```go
// app/update/update.go
func CheckForUpdate() (*ReleaseInfo, error)
func DownloadUpdate(release *ReleaseInfo) error
func VerifyChecksum(filepath, expectedSum string) error
func ApplyUpdate(newBinary string) error
func CreateBackup() (string, error)
func RestoreBackup(backupPath string) error

// app/update/github.go
func GetLatestRelease(owner, repo string) (*ReleaseInfo, error)
func DownloadAsset(url string, dest string) error

// app/install/install.go
func InstallInteractive() error
func InstallSilent(path string) error
func InstallToPath(sourcePath, destPath string) error
func VerifyInstallation(path string) error

// app/install/interactive.go
func PromptInstallLocation() (string, error)
func PromptAddToPath() (bool, error)
func ShowInstallationSummary(path string) error

// app/install/path.go
func AddToSystemPath(binPath string) error
func IsInPath(binPath string) bool
func GetDefaultInstallPath() string
```

### Benefits
- **User Convenience:** Single command to stay updated
- **Security:** Automated security patch distribution
- **Feature Access:** Quick access to new features
- **Cross-Platform:** Works on Windows, Linux, macOS
- **Atomic Updates:** Safe replacement with rollback capability

### Implementation Priority
1. **Phase 1:** Basic update mechanism with version checking
2. **Phase 2:** SHA256 verification and backup/restore
3. **Phase 3:** GoReleaser integration for automated releases
4. **Phase 4:** GPG signature verification (optional)

### Notes
- Consider using GitHub Actions for automated release creation
- Implement rate limiting for GitHub API calls
- Add update channel support (stable/beta/nightly)
- Consider delta updates for large binaries (future enhancement)

---
**Created:** 2025-01-22  
**Last Updated:** 2025-01-22  
**Assigned:** @CassandraGargoyle  
**Related Issues:** None