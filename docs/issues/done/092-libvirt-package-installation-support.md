# Issue #92: Libvirt Package Installation Support

## 🎯 Priority
**HIGH** - Critical for virt-manager functionality and #091 fix

## 📋 Status
- **Created**: 2025-10-04
- **Status**: 🔴 Open
- **Assignee**: TBD
- **Branch**: `feature/issue-092-libvirt-package-installation`

## 📝 Problem Description

### Current Situation
Issue #091 implemented dependency detection for libvirt but uses non-standard installation approach:

```go
// WRONG - Custom installation logic in virt check command
installPortunix := exec.Command("portunix", "install", "libvirt")
if err := installPortunix.Run(); err != nil {
    // Fallback to native package manager
    nativeInstall := exec.Command("sudo", "apt", "install", "libvirt-daemon-system")
    // ...
}
```

**Problems:**
1. ❌ Package 'libvirt' not found in `assets/install-packages.json`
2. ❌ Custom package manager logic in virt command (violates separation of concerns)
3. ❌ Error: "Error installing package 'libvirt': package 'libvirt' not found"
4. ❌ Duplicates OS detection logic already present in install system

### Impact
- virt check --fix-libvirt fails to install missing dependencies
- Inconsistent package installation approach across Portunix
- Users cannot fix libvirt issues automatically

### Related Issues
- **#091**: Libvirt Dependency Detection (implemented, but installation broken)
- **#090**: Libvirt Daemon Detection (implemented)

## 🎯 Required Solution

### 1. Add libvirt to Package Registry

Add libvirt package definition to `assets/install-packages.json`:

```json
{
  "libvirt": {
    "name": "Libvirt Virtualization Library",
    "description": "Libvirt daemon for managing QEMU/KVM virtual machines",
    "category": "virtualization",
    "platforms": {
      "linux": {
        "debian": {
          "method": "apt",
          "package": "libvirt-daemon-system",
          "alternatives": [
            "libvirt-daemon",
            "libvirt-clients"
          ]
        },
        "ubuntu": {
          "method": "apt",
          "package": "libvirt-daemon-system",
          "alternatives": [
            "libvirt-daemon",
            "libvirt-clients",
            "qemu-kvm",
            "virtinst"
          ]
        },
        "fedora": {
          "method": "dnf",
          "package": "libvirt-daemon",
          "alternatives": [
            "libvirt-client",
            "virt-install"
          ]
        },
        "rhel": {
          "method": "dnf",
          "package": "libvirt-daemon",
          "alternatives": [
            "libvirt-client"
          ]
        },
        "centos": {
          "method": "dnf",
          "package": "libvirt-daemon",
          "alternatives": [
            "libvirt-client"
          ]
        },
        "arch": {
          "method": "pacman",
          "package": "libvirt",
          "alternatives": [
            "virt-manager",
            "qemu-desktop"
          ]
        }
      }
    },
    "verify": {
      "command": "virsh --version",
      "expected_output": ""
    },
    "post_install": [
      {
        "description": "Enable libvirtd service",
        "command": "sudo systemctl enable libvirtd",
        "platforms": ["linux"]
      },
      {
        "description": "Start libvirtd service",
        "command": "sudo systemctl start libvirtd",
        "platforms": ["linux"]
      }
    ]
  }
}
```

### 2. Simplify virt check --fix-libvirt

**Remove custom installation logic from `src/cmd/virt_exec.go`:**

```go
// BEFORE (WRONG):
func executeInstallMissingDependencies(deps []string, dryRun bool) {
    // ... 100+ lines of custom OS detection and package manager logic
    osInfo, err := system.GetSystemInfo()
    // ... switch for each distro
    installCmd := exec.Command("sudo", "apt", "install", packageName)
    // ...
}

// AFTER (CORRECT):
func executeInstallMissingDependencies(deps []string, dryRun bool) {
    fmt.Println("⚠️  Missing libvirt dependencies detected:")
    for _, dep := range deps {
        fmt.Printf("   • %s\n", dep)
    }
    fmt.Println()

    if dryRun {
        fmt.Println("[DRY RUN] Would execute: portunix install libvirt")
        return
    }

    // Ask user for confirmation
    fmt.Print("Install libvirt package? [y/N]: ")
    reader := bufio.NewReader(os.Stdin)
    response, _ := reader.ReadString('\n')
    response = strings.TrimSpace(strings.ToLower(response))

    if response != "y" && response != "yes" {
        fmt.Println("Installation cancelled.")
        fmt.Println("💡 To install manually: portunix install libvirt")
        return
    }

    // Use standard portunix install system
    fmt.Println("\n🔧 Installing libvirt...")
    installCmd := exec.Command("portunix", "install", "libvirt")
    installCmd.Stdout = os.Stdout
    installCmd.Stderr = os.Stderr
    installCmd.Stdin = os.Stdin

    if err := installCmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Installation failed: %v\n", err)
        fmt.Println("💡 Try manually: portunix install libvirt")
        return
    }

    fmt.Println("\n✅ Libvirt installed successfully")
    fmt.Println("💡 Now run: sudo portunix virt check --fix-libvirt")
}
```

## 🎯 Acceptance Criteria

### Package Registry
- [ ] libvirt package added to `assets/install-packages.json`
- [ ] Support for Ubuntu/Debian/Fedora/RHEL/CentOS/Arch
- [ ] Correct package names for each distribution
- [ ] Verification command: `virsh --version`
- [ ] Post-install: enable and start libvirtd service

### virt check Command
- [ ] Remove custom OS detection from executeInstallMissingDependencies
- [ ] Remove custom package manager logic
- [ ] Use standard `portunix install libvirt` command
- [ ] Simple user confirmation prompt
- [ ] Clear error messages with manual install hint

### Testing
- [ ] `portunix install libvirt` works on Ubuntu
- [ ] `portunix install libvirt` works on Debian
- [ ] `portunix install libvirt` works on Fedora
- [ ] `sudo portunix virt check --fix-libvirt` installs missing dependencies
- [ ] After install, virt-manager can connect

## 📊 Technical Details

### Files to Modify

**1. `assets/install-packages.json`**
- Add complete libvirt package definition
- Include all supported distributions
- Add post-install service enablement

**2. `src/cmd/virt_exec.go`**
- Simplify `executeInstallMissingDependencies()` function
- Remove OS detection logic (~100 lines)
- Remove package manager switch logic
- Use standard `portunix install` command

### Code Cleanup

**Remove these lines from virt_exec.go:**
```go
// DELETE:
osInfo, err := system.GetSystemInfo()
var packageName string
var installCmd string
var distro string
// ... entire OS detection switch block (lines 586-663)
```

**Replace with:**
```go
// Use standard install command
installCmd := exec.Command("portunix", "install", "libvirt")
```

## 🧪 Testing Strategy

### Test Cases

**TC001: Install libvirt on Ubuntu**
```bash
portunix install libvirt
# Expected: Installs libvirt-daemon-system
# Verify: virsh --version works
# Verify: systemctl status libvirtd shows running
```

**TC002: Install via virt check**
```bash
sudo portunix virt check --fix-libvirt
# Expected: Detects missing dependencies
# Expected: Prompts to install libvirt
# Expected: Runs portunix install libvirt
# Verify: virt-manager connects
```

**TC003: Dry-run mode**
```bash
portunix virt check --fix-libvirt --dry-run
# Expected: Shows would install libvirt
# Expected: No actual installation
```

### Manual Testing
```bash
# 1. Remove libvirt (for testing)
sudo apt remove --purge libvirt-daemon-system
sudo systemctl daemon-reload

# 2. Test detection
./portunix virt check
# Expected: Shows missing dependencies

# 3. Test fix
sudo ./portunix virt check --fix-libvirt
# Choose 'y' to install
# Expected: Installs via portunix install libvirt

# 4. Verify
virt-manager
# Expected: Connects successfully
```

## 📚 References

### Existing Package Examples
- `assets/install-packages.json:nodejs` - Multi-distro package
- `assets/install-packages.json:docker` - Service installation with post-install
- `assets/install-packages.json:hugo` - Package with variants

### Related Code
- `src/app/install/installer.go` - Main installation logic
- `src/cmd/install.go` - Install command implementation
- Issue #091 commit: `1f05300` - Current broken implementation

## 💡 Implementation Notes

### Package Manager Selection
Portunix install system already handles OS detection and package manager selection. Do NOT duplicate this logic in virt command.

### Post-Install Actions
Libvirt requires systemd service to be enabled and started. Use `post_install` section in package definition.

### Verification
After installation, verify with `virsh --version` to ensure libvirt is working.

## ⚠️ Known Issues from #091
- Current implementation tries to call `portunix install libvirt` but package doesn't exist
- Fallback to direct package manager violates Portunix architecture
- Custom OS detection duplicates existing system

## 🎯 Success Metrics
- [ ] `portunix install libvirt` works on all supported distributions
- [ ] virt check uses standard install system
- [ ] Code reduced by ~100 lines (removed duplicate logic)
- [ ] virt-manager connects after automatic fix
- [ ] Consistent with other Portunix package installations

## 📅 Estimated Effort
**Complexity**: Low-Medium
**Time Estimate**: 1-2 hours
- Package definition: 30 minutes
- Code cleanup: 30 minutes
- Testing: 30-60 minutes

## 🔗 Related Files
```
assets/install-packages.json
src/cmd/virt_exec.go (lines 577-658 to be simplified)
src/app/install/installer.go
docs/issues/done/091-libvirt-dependency-failed-fix.md
```

---

**Note for AI Assistant:**
This is a cleanup/refactoring task. The functionality exists (package installation), it just needs to be properly integrated. Focus on:
1. Adding the package definition (copy pattern from existing packages)
2. Simplifying the virt command (remove custom logic, use standard install)
3. Testing the integration

Do NOT create new installation logic - use what already exists in Portunix install system.
