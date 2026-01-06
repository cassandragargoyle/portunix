//go:build windows

package windows

import (
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Hardware virtualization detection using registry
// This replaces the PowerShell Get-ComputerInfo | HyperVisorPresent command

// Registry keys for virtualization detection
var (
	RegKeyHyperVHypervisor = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization`,
	}
	RegKeyDeviceGuard = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `SYSTEM\CurrentControlSet\Control\DeviceGuard`,
	}
)

// IsHypervisorPresent checks if a hypervisor is present using registry
// This is faster than calling PowerShell Get-ComputerInfo
func IsHypervisorPresent() bool {
	// Method 1: Check registry for Hyper-V
	if hypervisorEnabled() {
		return true
	}

	// Method 2: Check via CPUID (hypervisor vendor)
	if detectHypervisorCPUID() {
		return true
	}

	// Method 3: Check if running in a VM
	if IsVM() {
		return true
	}

	return false
}

// hypervisorEnabled checks registry for hypervisor enabled status
func hypervisorEnabled() bool {
	// Check Hyper-V virtualization key
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer k.Close()
		if val, _, err := k.GetIntegerValue("EnabledVirtualizationBasedSecurity"); err == nil && val != 0 {
			return true
		}
	}

	// Check Device Guard
	k2, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\DeviceGuard`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer k2.Close()
		if val, _, err := k2.GetIntegerValue("EnableVirtualizationBasedSecurity"); err == nil && val != 0 {
			return true
		}
	}

	return false
}

// detectHypervisorCPUID uses IsProcessorFeaturePresent to detect hypervisor
// This is a simplified check - full CPUID would require assembly
func detectHypervisorCPUID() bool {
	// PF_VIRT_FIRMWARE_ENABLED = 21
	// Checks if virtualization firmware is enabled
	const PF_VIRT_FIRMWARE_ENABLED = 21

	modkernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procIsProcessorFeaturePresent := modkernel32.NewProc("IsProcessorFeaturePresent")

	ret, _, _ := procIsProcessorFeaturePresent.Call(uintptr(PF_VIRT_FIRMWARE_ENABLED))
	return ret != 0
}

// Windows Sandbox detection

// IsSandbox checks if running in Windows Sandbox
// Uses multiple detection methods without external process calls
func IsSandbox() bool {
	// Method 1: Check username (fastest)
	if username := os.Getenv("USERNAME"); username == "WDAGUtilityAccount" {
		return true
	}

	// Method 2: Check for sandbox-specific environment
	if os.Getenv("LOCALAPPDATA") != "" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if strings.Contains(strings.ToLower(localAppData), "packages\\microsoft.windows.sandbox") {
			return true
		}
	}

	// Method 3: Check for sandbox process using native API
	if isSandboxProcessRunning() {
		return true
	}

	return false
}

// isSandboxProcessRunning checks if CExecSvc.exe is running using native API
// This replaces tasklist command
func isSandboxProcessRunning() bool {
	const TH32CS_SNAPPROCESS = 0x00000002

	modkernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procCreateToolhelp32Snapshot := modkernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW := modkernel32.NewProc("Process32FirstW")
	procProcess32NextW := modkernel32.NewProc("Process32NextW")

	snapshot, _, err := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(windows.InvalidHandle) {
		return false
	}
	defer windows.CloseHandle(windows.Handle(snapshot))

	var pe processEntry32W
	pe.Size = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		return false
	}

	for {
		exeName := windows.UTF16ToString(pe.ExeFile[:])
		if strings.EqualFold(exeName, "CExecSvc.exe") {
			return true
		}

		ret, _, err = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			break
		}
	}

	return false
}

// processEntry32W is the PROCESSENTRY32W structure
type processEntry32W struct {
	Size              uint32
	Usage             uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	Threads           uint32
	ParentProcessID   uint32
	PriClassBase      int32
	Flags             uint32
	ExeFile           [260]uint16 // MAX_PATH
}

// VirtualizationCapabilities contains virtualization detection results
type VirtualizationCapabilities struct {
	HypervisorPresent bool
	IsVM              bool
	VMType            VMType
	IsSandbox         bool
	VTxEnabled        bool
}

// DetectVirtualization performs comprehensive virtualization detection
func DetectVirtualization() *VirtualizationCapabilities {
	caps := &VirtualizationCapabilities{}

	// Detect VM
	caps.VMType = DetectVM()
	caps.IsVM = caps.VMType != VMTypeNone

	// Detect hypervisor
	caps.HypervisorPresent = IsHypervisorPresent()

	// Detect sandbox
	caps.IsSandbox = IsSandbox()

	// Detect VT-x/AMD-V
	caps.VTxEnabled = detectHypervisorCPUID()

	return caps
}

// GetVariantString returns environment variant string based on detection
func GetVariantString() string {
	// Check sandbox first (highest priority)
	if IsSandbox() {
		return "Sandbox"
	}

	// Check VM
	vmType := DetectVM()
	switch vmType {
	case VMTypeVirtualBox:
		return "VirtualBox VM"
	case VMTypeVMware:
		return "VMware VM"
	case VMTypeHyperV:
		return "Hyper-V VM"
	case VMTypeQEMU:
		return "QEMU VM"
	case VMTypeXen:
		return "Xen VM"
	case VMTypeKVM:
		return "KVM VM"
	case VMTypeParallels:
		return "Parallels VM"
	case VMTypeUnknown:
		return "Virtual Machine"
	}

	return "Physical"
}
