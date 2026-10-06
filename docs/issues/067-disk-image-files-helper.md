# Issue #67: Disk Image Files Helper for Multiple Formats

## Summary
Create a comprehensive helper system for working with various disk image file formats including VDI (VirtualBox), VMDK (VMware), VHD/VHDX (Hyper-V/Windows), HDD (Parallels), QCOW2 (QEMU), IMG, ISO, and other formats commonly used on Linux and cross-platform virtualization environments.

## Current Behavior
- No built-in support for disk image manipulation in Portunix
- Users must manually use various tools (VBoxManage, qemu-img, etc.)
- No unified interface for different image formats
- No integration with Portunix virtualization commands

## Expected Behavior
```bash
# General disk image operations
portunix disk-image create --format vdi --size 20G /path/to/disk.vdi
portunix disk-image convert /path/to/disk.vdi --to vmdk --output /path/to/disk.vmdk
portunix disk-image resize /path/to/disk.vdi --size 40G
portunix disk-image info /path/to/disk.vdi

# Advanced operations
portunix disk-image compact /path/to/disk.vdi
portunix disk-image clone /path/to/source.vdi /path/to/target.vdi
portunix disk-image snapshot create /path/to/disk.vdi snapshot-name
portunix disk-image mount /path/to/disk.vdi --mountpoint /mnt/disk

# Batch operations
portunix disk-image batch-convert /path/to/images/*.vdi --format vmdk --output-dir /path/to/output
portunix disk-image batch-compact /path/to/images/*.vdi

# List supported formats
portunix disk-image formats --list
portunix disk-image formats --supported-conversions
```

## Expected Output
```
════════════════════════════════════════════════
🖼️  DISK IMAGE: Create VDI
════════════════════════════════════════════════
📄 Format: VDI (VirtualBox Disk Image)
📏 Size: 20GB
📁 Output: /home/user/vm-disks/disk.vdi
🏗️  Type: Dynamic allocation
════════════════════════════════════════════════
🔍 Checking prerequisites...
✅ VBoxManage found: 7.0.12
🚀 Creating disk image...
Creating dynamic VDI image '/home/user/vm-disks/disk.vdi'...
Progress: 100% (20971520 KB)
✅ VDI disk image created successfully
════════════════════════════════════════════════
📊 IMAGE INFO:
Format: VDI
Virtual Size: 20 GB
Actual Size: 4 MB (dynamic)
UUID: {12345678-1234-1234-1234-123456789abc}
════════════════════════════════════════════════
```

## Implementation Requirements

### Core Architecture
```go
// app/diskimage/types.go
package diskimage

type ImageFormat string

const (
    FormatVDI    ImageFormat = "vdi"     // VirtualBox
    FormatVMDK   ImageFormat = "vmdk"    // VMware
    FormatVHD    ImageFormat = "vhd"     // Hyper-V/Windows
    FormatVHDX   ImageFormat = "vhdx"    // Hyper-V modern
    FormatHDD    ImageFormat = "hdd"     // Parallels
    FormatQCOW2  ImageFormat = "qcow2"   // QEMU
    FormatQCOW   ImageFormat = "qcow"    // QEMU legacy
    FormatRAW    ImageFormat = "raw"     // Raw disk image
    FormatIMG    ImageFormat = "img"     // Generic disk image
    FormatISO    ImageFormat = "iso"     // ISO 9660 (read-only)
)

type DiskImage struct {
    Path         string
    Format       ImageFormat
    VirtualSize  uint64
    ActualSize   uint64
    Compressed   bool
    Encrypted    bool
    Snapshots    []Snapshot
    UUID         string
    Properties   map[string]interface{}
}

type Snapshot struct {
    Name        string
    UUID        string
    Description string
    Created     time.Time
    Size        uint64
}

type ImageHandler interface {
    Create(path string, size uint64, options CreateOptions) error
    Convert(source, target string, targetFormat ImageFormat) error
    Resize(path string, newSize uint64) error
    Compact(path string) error
    Clone(source, target string) error
    GetInfo(path string) (*DiskImage, error)
    Mount(path, mountpoint string) error
    Unmount(path string) error
    CreateSnapshot(path, snapshotName, description string) error
    ListSnapshots(path string) ([]Snapshot, error)
    DeleteSnapshot(path, snapshotUUID string) error
}
```

### Multi-Backend Support
```go
// app/diskimage/backends/
├── vbox.go       // VBoxManage integration
├── qemu.go       // qemu-img integration
├── hyperv.go     // Hyper-V PowerShell integration
├── parallels.go  // Parallels prl_disk_tool integration
└── native.go     // Native Go implementations where possible

// Backend detection and routing
type BackendManager struct {
    backends map[ImageFormat]ImageHandler
}

func (bm *BackendManager) DetectBestBackend(format ImageFormat) ImageHandler {
    // Priority order for each format
    switch format {
    case FormatVDI:
        if isVBoxAvailable() {
            return &VBoxHandler{}
        }
        return &QemuHandler{}
    case FormatVMDK:
        if isVMwareAvailable() {
            return &VMwareHandler{}
        } else if isVBoxAvailable() {
            return &VBoxHandler{}
        }
        return &QemuHandler{}
    case FormatVHD, FormatVHDX:
        if isWindowsHyperV() {
            return &HyperVHandler{}
        }
        return &QemuHandler{}
    // ... other formats
    }
}
```

### Supported Disk Image Formats

#### VirtualBox VDI Support
```go
// app/diskimage/backends/vbox.go
type VBoxHandler struct{}

func (v *VBoxHandler) Create(path string, size uint64, opts CreateOptions) error {
    cmd := exec.Command("VBoxManage", "createmedium", "disk",
        "--filename", path,
        "--size", fmt.Sprintf("%d", size/1024/1024), // MB
        "--format", "VDI")

    if opts.FixedSize {
        cmd.Args = append(cmd.Args, "--variant", "Fixed")
    } else {
        cmd.Args = append(cmd.Args, "--variant", "Standard")
    }

    return cmd.Run()
}

func (v *VBoxHandler) Convert(source, target string, targetFormat ImageFormat) error {
    cmd := exec.Command("VBoxManage", "clonemedium", "disk",
        source, target, "--format", string(targetFormat))
    return cmd.Run()
}

func (v *VBoxHandler) GetInfo(path string) (*DiskImage, error) {
    cmd := exec.Command("VBoxManage", "showmediuminfo", "disk", path)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    return parseVBoxInfo(string(output))
}
```

#### QEMU IMG Support (Universal Backend)
```go
// app/diskimage/backends/qemu.go
type QemuHandler struct{}

func (q *QemuHandler) Create(path string, size uint64, opts CreateOptions) error {
    args := []string{"create", "-f", string(opts.Format), path,
                     fmt.Sprintf("%d", size)}

    if opts.Format == FormatQCOW2 {
        args = append(args, "-o", "cluster_size=65536")
    }

    cmd := exec.Command("qemu-img", args...)
    return cmd.Run()
}

func (q *QemuHandler) Convert(source, target string, targetFormat ImageFormat) error {
    cmd := exec.Command("qemu-img", "convert", "-f", "auto",
                        "-O", string(targetFormat), source, target)
    return cmd.Run()
}

func (q *QemuHandler) GetInfo(path string) (*DiskImage, error) {
    cmd := exec.Command("qemu-img", "info", "--output=json", path)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    return parseQemuInfo(output)
}
```

#### Hyper-V VHD/VHDX Support (Windows)
```go
// app/diskimage/backends/hyperv.go
type HyperVHandler struct{}

func (h *HyperVHandler) Create(path string, size uint64, opts CreateOptions) error {
    script := fmt.Sprintf(`
        New-VHD -Path "%s" -SizeBytes %d -Dynamic
    `, path, size)

    cmd := exec.Command("powershell", "-Command", script)
    return cmd.Run()
}

func (h *HyperVHandler) Convert(source, target string, targetFormat ImageFormat) error {
    script := fmt.Sprintf(`
        Convert-VHD -Path "%s" -DestinationPath "%s" -VHDType %s
    `, source, target, getHyperVType(targetFormat))

    cmd := exec.Command("powershell", "-Command", script)
    return cmd.Run()
}
```

### Linux-Specific Extensions
```bash
# Linux loop device support
portunix disk-image mount /path/to/disk.img --loop --partition 1
portunix disk-image mount /path/to/disk.iso --mountpoint /mnt/iso --readonly

# NBD (Network Block Device) support
portunix disk-image export /path/to/disk.qcow2 --nbd --port 10809
portunix disk-image import nbd://localhost:10809 --output /path/to/local.qcow2

# LVM integration
portunix disk-image scan-lvm /path/to/disk.img
portunix disk-image mount-lvm /path/to/disk.img --lv-name volume1 --mountpoint /mnt/lvm

# Filesystem utilities
portunix disk-image fsck /path/to/disk.img --partition 1 --filesystem ext4
portunix disk-image resize-fs /path/to/disk.img --partition 1 --filesystem ext4
```

### Advanced Features

#### Image Compression and Optimization
```go
func (handler *ImageHandler) CompactImage(path string) error {
    format, err := DetectImageFormat(path)
    if err != nil {
        return err
    }

    switch format {
    case FormatVDI:
        return handler.vboxCompact(path)
    case FormatVMDK:
        return handler.vmwareCompact(path)
    case FormatQCOW2:
        return handler.qemuCompact(path)
    default:
        return fmt.Errorf("compact not supported for format %s", format)
    }
}

func (handler *ImageHandler) OptimizeImage(path string, opts OptimizeOptions) error {
    // Zero out free space
    if opts.ZeroFreeSpace {
        if err := handler.zeroFreeSpace(path); err != nil {
            return err
        }
    }

    // Compress image
    if opts.Compress {
        if err := handler.CompactImage(path); err != nil {
            return err
        }
    }

    return nil
}
```

#### Snapshot Management
```go
func (handler *ImageHandler) CreateSnapshot(path, name, description string) error {
    format, err := DetectImageFormat(path)
    if err != nil {
        return err
    }

    switch format {
    case FormatQCOW2:
        return handler.qemuCreateSnapshot(path, name, description)
    case FormatVDI:
        return fmt.Errorf("VDI snapshots managed by VirtualBox, not directly")
    default:
        return fmt.Errorf("snapshots not supported for format %s", format)
    }
}

func (handler *QemuHandler) qemuCreateSnapshot(path, name, description string) error {
    cmd := exec.Command("qemu-img", "snapshot", "-c", name, path)
    return cmd.Run()
}
```

### Format Detection and Validation
```go
// app/diskimage/detection.go
func DetectImageFormat(path string) (ImageFormat, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Read file signature
    header := make([]byte, 512)
    _, err = file.Read(header)
    if err != nil {
        return "", err
    }

    // Check magic numbers
    if bytes.HasPrefix(header, []byte("<<< Oracle VM VirtualBox Disk Image >>>")) {
        return FormatVDI, nil
    } else if bytes.HasPrefix(header, []byte("VMDK")) {
        return FormatVMDK, nil
    } else if bytes.HasPrefix(header, []byte("conectix")) {
        return FormatVHD, nil
    } else if bytes.HasPrefix(header, []byte("vhdx")) {
        return FormatVHDX, nil
    } else if bytes.HasPrefix(header, []byte("QFI")) {
        return FormatQCOW2, nil
    } else if bytes.HasPrefix(header, []byte("CD001")) {
        return FormatISO, nil
    }

    // Use file extension as fallback
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".vdi":
        return FormatVDI, nil
    case ".vmdk":
        return FormatVMDK, nil
    case ".vhd":
        return FormatVHD, nil
    case ".vhdx":
        return FormatVHDX, nil
    case ".qcow2":
        return FormatQCOW2, nil
    case ".img":
        return FormatRAW, nil
    case ".iso":
        return FormatISO, nil
    }

    return FormatRAW, nil // Default to raw
}
```

### CLI Integration
```go
// cmd/disk_image.go
package cmd

var diskImageCmd = &cobra.Command{
    Use:   "disk-image",
    Short: "Manage disk image files (VDI, VMDK, VHD, QCOW2, etc.)",
    Long: `Comprehensive disk image management for various virtualization formats.

Supports: VDI, VMDK, VHD/VHDX, QCOW2, IMG, ISO and more.
Works with VirtualBox, VMware, Hyper-V, QEMU, and Parallels.`,
}

var createDiskImageCmd = &cobra.Command{
    Use:   "create [path]",
    Short: "Create a new disk image",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        path := args[0]
        format, _ := cmd.Flags().GetString("format")
        size, _ := cmd.Flags().GetString("size")

        handler := diskimage.NewManager()
        err := handler.Create(path, parseSize(size), diskimage.CreateOptions{
            Format: diskimage.ImageFormat(format),
        })

        if err != nil {
            fmt.Printf("❌ Failed to create disk image: %v\n", err)
            os.Exit(1)
        }

        fmt.Printf("✅ Created %s disk image: %s\n", format, path)
    },
}

func init() {
    createDiskImageCmd.Flags().StringP("format", "f", "vdi",
        "Image format (vdi, vmdk, vhd, vhdx, qcow2, raw)")
    createDiskImageCmd.Flags().StringP("size", "s", "10G",
        "Image size (e.g., 10G, 500M, 1T)")
    createDiskImageCmd.Flags().Bool("fixed", false,
        "Create fixed-size image (default: dynamic)")

    diskImageCmd.AddCommand(createDiskImageCmd)
    rootCmd.AddCommand(diskImageCmd)
}
```

### Linux Mount Integration
```go
// app/diskimage/mount_linux.go
//go:build linux

func (m *MountManager) MountImage(imagePath, mountPoint string, opts MountOptions) error {
    format, err := DetectImageFormat(imagePath)
    if err != nil {
        return err
    }

    switch format {
    case FormatISO:
        return m.mountISO(imagePath, mountPoint, opts)
    case FormatRAW, FormatIMG:
        return m.mountLoop(imagePath, mountPoint, opts)
    default:
        // Convert to raw format temporarily for mounting
        return m.mountWithConversion(imagePath, mountPoint, opts)
    }
}

func (m *MountManager) mountLoop(imagePath, mountPoint string, opts MountOptions) error {
    // Find available loop device
    loopDevice, err := m.findAvailableLoop()
    if err != nil {
        return err
    }

    // Associate image with loop device
    cmd := exec.Command("losetup", loopDevice, imagePath)
    if opts.Partition > 0 {
        cmd.Args = append(cmd.Args, "-P") // Enable partition support
    }

    if err := cmd.Run(); err != nil {
        return err
    }

    // Mount the loop device
    mountDevice := loopDevice
    if opts.Partition > 0 {
        mountDevice = fmt.Sprintf("%sp%d", loopDevice, opts.Partition)
    }

    cmd = exec.Command("mount", mountDevice, mountPoint)
    if opts.ReadOnly {
        cmd.Args = append(cmd.Args, "-o", "ro")
    }

    return cmd.Run()
}
```

### Cross-Platform Compatibility
```go
// app/diskimage/platform.go
//go:build windows
func init() {
    // Register Windows-specific handlers
    RegisterHandler(FormatVHD, &HyperVHandler{})
    RegisterHandler(FormatVHDX, &HyperVHandler{})
}

//go:build linux
func init() {
    // Register Linux-specific handlers
    RegisterHandler(FormatRAW, &LoopDeviceHandler{})
    RegisterHandler(FormatISO, &ISO9660Handler{})
}

//go:build darwin
func init() {
    // Register macOS-specific handlers
    RegisterHandler(FormatDMG, &DiskImageHandler{})
}
```

## Testing Requirements

### Unit Tests
```go
func TestImageFormatDetection(t *testing.T) {
    tests := []struct {
        filename string
        expected ImageFormat
    }{
        {"disk.vdi", FormatVDI},
        {"disk.vmdk", FormatVMDK},
        {"disk.vhd", FormatVHD},
        {"disk.qcow2", FormatQCOW2},
        {"disk.img", FormatRAW},
        {"disk.iso", FormatISO},
    }

    for _, test := range tests {
        result := detectFormatFromExtension(test.filename)
        if result != test.expected {
            t.Errorf("Expected %s for %s, got %s",
                test.expected, test.filename, result)
        }
    }
}
```

### Integration Tests
```go
func TestDiskImageOperations(t *testing.T) {
    tf := testframework.NewTestFramework("DiskImageOperations")
    tf.Start(t, "Test comprehensive disk image operations")

    success := true
    defer tf.Finish(t, success)

    tmpDir, err := ioutil.TempDir("", "diskimage-test")
    if err != nil {
        tf.Error(t, "Failed to create temp directory", err.Error())
        success = false
        return
    }
    defer os.RemoveAll(tmpDir)

    imagePath := filepath.Join(tmpDir, "test.vdi")

    // Test creation
    tf.Step(t, "Create VDI disk image")
    handler := diskimage.NewManager()
    err = handler.Create(imagePath, 100*1024*1024, diskimage.CreateOptions{
        Format: diskimage.FormatVDI,
    })
    if err != nil {
        tf.Error(t, "Failed to create disk image", err.Error())
        success = false
        return
    }

    // Test info
    tf.Step(t, "Get disk image information")
    info, err := handler.GetInfo(imagePath)
    if err != nil {
        tf.Error(t, "Failed to get image info", err.Error())
        success = false
        return
    }

    tf.Success(t, fmt.Sprintf("Created %s image, size: %d bytes",
        info.Format, info.VirtualSize))
}
```

## Error Handling

### Common Error Scenarios
1. **Missing Tools**: VBoxManage, qemu-img not installed
2. **Permission Issues**: Cannot access image files or mount points
3. **Disk Space**: Insufficient space for operations
4. **Format Compatibility**: Unsupported format conversions
5. **Corrupted Images**: Image file corruption detection

### Error Recovery
```go
type DiskImageError struct {
    Operation string
    Path      string
    Format    ImageFormat
    Cause     error
    Suggestion string
}

func (e *DiskImageError) Error() string {
    return fmt.Sprintf("%s failed for %s (%s): %v. %s",
        e.Operation, e.Path, e.Format, e.Cause, e.Suggestion)
}

func RecoverFromError(err error, operation string) error {
    switch {
    case strings.Contains(err.Error(), "VBoxManage: command not found"):
        return &DiskImageError{
            Operation: operation,
            Cause: err,
            Suggestion: "Install VirtualBox or use 'portunix install virtualbox'",
        }
    case strings.Contains(err.Error(), "qemu-img: command not found"):
        return &DiskImageError{
            Operation: operation,
            Cause: err,
            Suggestion: "Install QEMU or use 'portunix install qemu'",
        }
    default:
        return err
    }
}
```

## Documentation Requirements

### User Documentation
- Comprehensive format comparison guide
- Platform-specific installation requirements
- Common use cases and examples
- Troubleshooting guide

### Developer Documentation
- Backend handler interface
- Adding new format support
- Cross-platform considerations
- Testing procedures

## Acceptance Criteria
- [ ] Support for VDI, VMDK, VHD/VHDX, QCOW2, IMG, ISO formats
- [ ] Create, convert, resize, compact, clone operations work
- [ ] Cross-platform compatibility (Windows, Linux, macOS)
- [ ] Integration with existing virtualization tools
- [ ] Proper format detection and validation
- [ ] Mount/unmount support on Linux
- [ ] Snapshot management for supporting formats
- [ ] Batch operations for multiple files
- [ ] Comprehensive error handling and recovery
- [ ] Performance optimization for large images
- [ ] Complete documentation and examples

## Priority
**High** - Essential for advanced virtualization workflows and disk management

## Labels
- enhancement
- virtualization
- disk-management
- cross-platform
- vdi
- vmdk
- vhd
- qcow2
- image-processing

## Related Issues
- VM management system (#055)
- VirtualBox integration improvements
- QEMU/KVM support implementation
- Cross-platform file system utilities

## Estimated Effort
- **Core Architecture**: 8-10 hours
- **Backend Handlers**: 16-20 hours
- **Linux Mount Integration**: 6-8 hours
- **Cross-platform Support**: 8-10 hours
- **CLI Integration**: 4-6 hours
- **Testing**: 8-10 hours
- **Documentation**: 4-6 hours
- **Total**: 54-70 hours

---

**Created**: 2025-09-24
**Status**: 📋 Open
**Priority**: High
**Type**: Enhancement