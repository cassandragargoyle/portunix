# Issue #207: QEMU Windows VM Clipboard Integration

> **Renumbered:** formerly internal issue #020. The number was
> renumbered because it collided with an existing GitHub issue or pull request.

## Overview
**Type**: Enhancement
**Priority**: Medium
**Status**: ✅ Implemented
**Created**: 2025-01-31
**Closed**: 2026-05-11
**Target Version**: 1.6.0

## Resolution

The scope of this issue was delivered across the broader QEMU/KVM virtualization
rework that landed after this ticket was filed. The original `portunix vm …` API
proposed here was replaced by the libvirt-based `portunix virt …` stack
implemented via the `ptx-virt` helper binary.

Coverage map:

- **SPICE display by default for new Windows VMs** — `virt-install` invocation
  uses `--graphics spice` and the QEMU command line uses `-vga qxl`
  (`src/app/vm.go`). Delivered as part of #017 / #049 / #055.
- **SPICE guest tools installation** — covered by the package registry entry
  `src/helpers/ptx-installer/assets/packages/spice-guest-tools.json`
  (Windows: `spice-guest-tools-latest.exe`; Linux: `spice-vdagent`).
- **SPICE server / client (host-side) installation** — tracked separately and
  implemented under #093 (`spice-server`, `virt-viewer`, `spice-guest-agent`
  packages with per-distro install methods).

The remaining proposals from this ticket (`portunix vm enhance --clipboard`,
`portunix vm migrate-display`, `portunix install --target vm:<name>`) were
written against the old direct-QEMU API and are not applicable to the current
libvirt-based architecture. If retrofitting clipboard support on existing
libvirt domains is needed in the future, it should be filed as a new issue
scoped to `portunix virt` / `ptx-virt`.

## Problem Description

Currently, Windows VMs created with QEMU through Portunix do not support clipboard sharing between host and guest. This significantly impacts productivity when working with virtual machines, as users cannot easily copy/paste text, commands, or data between the host system and Windows VM.

## Current Limitations

1. **No clipboard sharing**: Cannot copy/paste between host and VM
2. **Manual data transfer**: Users must type everything manually or use file sharing
3. **Poor user experience**: Increases friction when working with VMs
4. **No SPICE integration**: Missing SPICE protocol support for enhanced features

## Proposed Solution

### 1. SPICE Protocol Integration
Implement SPICE (Simple Protocol for Independent Computing Environments) support for both new and existing Windows VMs:

- **SPICE server**: Enable in QEMU for all Windows VMs
- **SPICE client tools**: Install in Windows guest (spice-guest-tools)
- **virt-viewer**: Recommend or integrate SPICE-compatible viewer
- **QXL graphics**: Use QXL video driver for better performance
- **Migration support**: Convert existing VMs from VNC/GTK to SPICE
- **Standalone installation**: Support SPICE tools installation on any Windows system

### 2. Required Components

#### Host-side Requirements
```bash
# Packages to add to install-packages.json
- spice-client-gtk     # SPICE GTK client
- virt-viewer          # Virtual machine viewer with SPICE support
- qemu-system-x86      # Already included, needs SPICE enabled
```

#### Guest-side Requirements
```
- spice-guest-tools-x.x.x.exe  # Windows SPICE guest tools
- virtio-win drivers            # Already handled in current implementation
```

### 3. Implementation Steps

#### Step 1: Add SPICE Support to Existing Windows VMs
Create new command to enable SPICE on already running Windows VMs:

```bash
$ portunix vm enhance win11-dev --clipboard
```

This command will:
1. Detect if VM is already using SPICE
2. If not, reconfigure VM to use SPICE display
3. Install SPICE guest tools inside running Windows

```go
func EnhanceVMWithClipboard(vmName string) error {
    // Check current VM configuration
    vmConfig, err := getVMConfig(vmName)
    if err != nil {
        return fmt.Errorf("failed to get VM config: %w", err)
    }
    
    // Check if SPICE already enabled
    if vmConfig.HasSPICE {
        fmt.Println("✓ SPICE already enabled for this VM")
    } else {
        // Stop VM if running
        if isVMRunning(vmName) {
            fmt.Println("Stopping VM to reconfigure...")
            stopVM(vmName)
        }
        
        // Update VM configuration to add SPICE
        updateVMConfigForSPICE(vmName)
        
        // Restart VM with new configuration
        fmt.Println("Starting VM with SPICE support...")
        startVMWithSPICE(vmName)
    }
    
    // Install SPICE tools in guest
    return installSPICEToolsInGuest(vmName)
}
```

#### Step 2: Guest Tools Installation for Existing VMs
Support both automatic and manual installation:

```go
func installSPICEToolsInGuest(vmName string) error {
    fmt.Println("Installing SPICE guest tools in Windows VM...")
    
    // Method 1: Automatic via QEMU guest agent
    if hasQEMUGuestAgent(vmName) {
        return installViaGuestAgent(vmName)
    }
    
    // Method 2: Mount ISO and trigger autorun
    spiceToolsISO := downloadSPICEToolsISO()
    mountISOToVM(vmName, spiceToolsISO)
    
    fmt.Println(`
SPICE Guest Tools ISO mounted as D: drive in VM.
Please install manually:
1. Open File Explorer in VM
2. Navigate to D: drive
3. Run spice-guest-tools-x.x.x.exe
4. Follow installation wizard
5. Restart VM when prompted
    `)
    
    return nil
}

func installViaGuestAgent(vmName string) error {
    // Download SPICE tools to VM via guest agent
    spiceToolsURL := "https://www.spice-space.org/download/windows/spice-guest-tools/spice-guest-tools-latest.exe"
    
    // Execute PowerShell in guest to download and install
    psScript := fmt.Sprintf(`
        $url = "%s"
        $output = "$env:TEMP\spice-guest-tools.exe"
        Invoke-WebRequest -Uri $url -OutFile $output
        Start-Process -FilePath $output -ArgumentList "/S" -Wait
    `, spiceToolsURL)
    
    return executeInGuest(vmName, "powershell", psScript)
}
```

#### Step 3: Modify VM Creation for New VMs
Update VM creation to include SPICE by default:

```go
func runQEMUWindows11(vmName, isoPath string, config VMConfig) error {
    // Add SPICE display and audio
    args = append(args,
        "-spice", "port=5900,addr=127.0.0.1,disable-ticketing=on",
        "-device", "virtio-serial-pci",
        "-device", "virtserialport,chardev=spicechannel0,name=com.redhat.spice.0",
        "-chardev", "spicevmc,id=spicechannel0,name=vdagent",
        "-display", "spice-app",  // or "gtk" with SPICE support
    )
    
    // Add QXL video for better performance
    args = append(args,
        "-device", "qxl-vga,vram_size=67108864,ram_size=67108864",
    )
}
```

#### Step 4: Standalone SPICE Tools Installation Command
Add command to install SPICE tools on any existing Windows system:

```bash
$ portunix install spice-guest-tools --target vm:win11-dev
# or for physical Windows machines accessed via SSH/RDP
$ portunix install spice-guest-tools --target windows-host
```

Implementation:

```go
func InstallSPICEGuestTools(target string) error {
    // Parse target (VM name or remote host)
    if strings.HasPrefix(target, "vm:") {
        vmName := strings.TrimPrefix(target, "vm:")
        return installSPICEToolsInVM(vmName)
    }
    
    // For physical Windows machines
    return installSPICEToolsRemote(target)
}

func installSPICEToolsInVM(vmName string) error {
    // Check if VM is running
    if !isVMRunning(vmName) {
        return fmt.Errorf("VM %s is not running", vmName)
    }
    
    // Download SPICE tools
    toolsPath := downloadSPICETools()
    
    // Method 1: Via QEMU Monitor
    if hasQEMUMonitor(vmName) {
        // Change CD-ROM to SPICE tools ISO
        changeCD(vmName, toolsPath)
        fmt.Println("SPICE tools ISO inserted. Please run setup from D: drive in VM.")
    }
    
    // Method 2: Via network share
    if hasNetworkShare(vmName) {
        copyToShare(vmName, toolsPath)
        fmt.Println("SPICE tools copied to VM network share.")
    }
    
    return nil
}
```

#### Step 5: Configuration Detection and Migration
Detect existing VM display configuration and offer migration:

```go
func MigrateVMToSPICE(vmName string) error {
    config := readVMConfig(vmName)
    
    if config.Display == "vnc" || config.Display == "gtk" {
        fmt.Printf("VM '%s' currently uses %s display.\n", vmName, config.Display)
        fmt.Println("Would you like to migrate to SPICE for clipboard support? (y/n)")
        
        if getUserConfirmation() {
            // Backup current config
            backupVMConfig(vmName)
            
            // Update to SPICE
            config.Display = "spice"
            config.Graphics = "qxl"
            config.AddDevice("virtio-serial-pci")
            config.AddDevice("virtserialport")
            
            saveVMConfig(vmName, config)
            
            fmt.Println("✓ VM configuration updated to use SPICE")
            fmt.Println("Please restart the VM for changes to take effect")
        }
    }
    
    return nil
}
```

#### Step 6: Connection Helper
Create helper function for connecting with clipboard support:

```go
func ConnectToVMWithClipboard(vmName string) error {
    // Launch virt-viewer or remote-viewer with SPICE
    cmd := exec.Command("remote-viewer", 
        fmt.Sprintf("spice://localhost:5900"))
    return cmd.Start()
}
```

## Alternative Solutions Considered

### VNC with clipboard extension
- ❌ Limited clipboard support
- ❌ Requires additional VNC server configuration
- ❌ Performance overhead

### RDP (Remote Desktop Protocol)
- ✅ Native Windows support
- ❌ Requires Windows Pro/Enterprise
- ❌ More complex network configuration

### QEMU Guest Agent
- ✅ Lightweight
- ❌ Limited clipboard functionality
- ❌ Primarily for VM management, not user interaction

## Success Criteria

- [ ] New Windows VMs created with SPICE support by default
- [ ] Existing Windows VMs can be enhanced with SPICE via `vm enhance` command
- [ ] SPICE guest tools installable on any Windows system (VM or physical)
- [ ] Configuration migration from VNC/GTK to SPICE works smoothly
- [ ] Bidirectional clipboard works (text, images)
- [ ] Clear documentation and user guidance
- [ ] Fallback to VNC if SPICE unavailable
- [ ] Performance comparable to or better than current implementation

## Testing Requirements

1. **Clipboard functionality**
   - Copy text from host to VM
   - Copy text from VM to host
   - Copy formatted text (RTF)
   - Copy images (optional)

2. **Performance testing**
   - Display responsiveness
   - CPU usage comparison
   - Memory overhead

3. **Compatibility testing**
   - Windows 10
   - Windows 11
   - Different Linux distributions as host

## User Interface Changes

### For New VMs
```
$ portunix vm create win11-dev --os windows11
Creating Windows 11 VM with enhanced features...
✓ TPM 2.0 enabled
✓ UEFI/Secure Boot configured
✓ SPICE display with clipboard support enabled
✓ Installing SPICE guest tools...

To connect with clipboard support:
$ portunix vm connect win11-dev

Or use: virt-viewer --connect spice://localhost:5900
```

### For Existing VMs
```
$ portunix vm enhance win11-existing --clipboard
Analyzing VM 'win11-existing'...
✓ VM is currently using VNC display
✓ Migrating to SPICE for clipboard support...
✓ Stopping VM...
✓ Updating configuration...
✓ Starting VM with SPICE...
✓ Installing SPICE guest tools...

Clipboard support enabled! Connect with:
$ portunix vm connect win11-existing
```

### Standalone Installation
```
$ portunix install spice-guest-tools --target vm:win11-dev
Downloading SPICE Guest Tools...
✓ Downloaded: spice-guest-tools-0.141.exe
✓ Mounting ISO to VM...
✓ Please run D:\spice-guest-tools-0.141.exe in the VM

$ portunix install spice-guest-tools --target 192.168.1.100
Connecting to Windows host...
✓ Uploading installer...
✓ Running silent installation...
✓ SPICE Guest Tools installed successfully
```

### Configuration Migration
```
$ portunix vm migrate-display win11-old --to spice
Current display: VNC
Target display: SPICE with clipboard support

This will:
- Enable bidirectional clipboard
- Improve graphics performance
- Add audio support

Continue? (y/n): y
✓ Configuration backed up
✓ Display migrated to SPICE
✓ Please restart VM for changes to take effect
```

## Documentation Updates

1. Update VM documentation with clipboard features
2. Add troubleshooting guide for SPICE issues
3. Document manual SPICE tools installation
4. Add performance tuning tips

## Dependencies

- SPICE protocol libraries
- virt-viewer or compatible SPICE client
- spice-guest-tools for Windows
- QXL video driver (included in virtio-win)

## Risks and Mitigation

1. **SPICE availability**: Not all systems have SPICE support
   - Mitigation: Detect and fallback to VNC

2. **Guest tools installation**: May fail or be blocked
   - Mitigation: Provide manual installation instructions

3. **Performance impact**: SPICE may use more resources
   - Mitigation: Make it optional, allow disabling

## References

- [SPICE Protocol](https://www.spice-space.org/)
- [SPICE Guest Tools](https://www.spice-space.org/download/windows/spice-guest-tools/)
- [QEMU SPICE Documentation](https://www.qemu.org/docs/master/system/invocation.html#hxtool-4)
- [virt-viewer Documentation](https://virt-manager.org/)
- ChatGPT recommendation for clipboard integration

## Notes

This enhancement significantly improves the Windows VM user experience by enabling seamless clipboard integration, making Portunix VMs more practical for daily development work.