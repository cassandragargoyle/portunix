//go:build windows

package windows

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// RegistryKey represents a registry key path
type RegistryKey struct {
	Root    registry.Key
	SubKey  string
}

// Common registry paths
var (
	RegKeyWindowsNTCurrentVersion = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
	}
	RegKeyCentralProcessor0 = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
	}
	RegKeyVirtualBox = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `SOFTWARE\Oracle\VirtualBox`,
	}
	RegKeyVirtualBoxWow64 = RegistryKey{
		Root:   registry.LOCAL_MACHINE,
		SubKey: `SOFTWARE\WOW6432Node\Oracle\VirtualBox`,
	}
)

// ReadString reads a string value from registry
func ReadString(key RegistryKey, valueName string) (string, error) {
	k, err := registry.OpenKey(key.Root, key.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetStringValue(valueName)
	if err != nil {
		return "", fmt.Errorf("failed to read value %s: %w", valueName, err)
	}

	return val, nil
}

// ReadUint32 reads a DWORD value from registry
func ReadUint32(key RegistryKey, valueName string) (uint32, error) {
	k, err := registry.OpenKey(key.Root, key.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("failed to read value %s: %w", valueName, err)
	}

	return uint32(val), nil
}

// KeyExists checks if a registry key exists
func KeyExists(key RegistryKey) bool {
	k, err := registry.OpenKey(key.Root, key.SubKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

// ReadStringWithDefault reads a string value with a default fallback
func ReadStringWithDefault(key RegistryKey, valueName, defaultValue string) string {
	val, err := ReadString(key, valueName)
	if err != nil {
		return defaultValue
	}
	return val
}

// ReadUint32WithDefault reads a DWORD value with a default fallback
func ReadUint32WithDefault(key RegistryKey, valueName string, defaultValue uint32) uint32 {
	val, err := ReadUint32(key, valueName)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetProductName returns Windows product name from registry
func GetProductName() string {
	return ReadStringWithDefault(RegKeyWindowsNTCurrentVersion, "ProductName", "")
}

// GetDisplayVersion returns Windows display version (e.g., "23H2")
func GetDisplayVersion() string {
	return ReadStringWithDefault(RegKeyWindowsNTCurrentVersion, "DisplayVersion", "")
}

// GetCurrentBuild returns Windows current build number from registry
func GetCurrentBuild() string {
	return ReadStringWithDefault(RegKeyWindowsNTCurrentVersion, "CurrentBuild", "")
}

// GetUBR returns Windows Update Build Revision
func GetUBR() uint32 {
	return ReadUint32WithDefault(RegKeyWindowsNTCurrentVersion, "UBR", 0)
}

// GetBuildLabEx returns detailed build information
func GetBuildLabEx() string {
	return ReadStringWithDefault(RegKeyWindowsNTCurrentVersion, "BuildLabEx", "")
}

// GetEditionID returns Windows edition ID (e.g., "Professional", "Enterprise")
func GetEditionID() string {
	return ReadStringWithDefault(RegKeyWindowsNTCurrentVersion, "EditionID", "")
}

// GetInstallDate returns Windows installation timestamp
func GetInstallDate() uint32 {
	return ReadUint32WithDefault(RegKeyWindowsNTCurrentVersion, "InstallDate", 0)
}

// GetVirtualBoxInstallDir returns VirtualBox installation directory if installed
func GetVirtualBoxInstallDir() string {
	// Try regular path first
	if path := ReadStringWithDefault(RegKeyVirtualBox, "InstallDir", ""); path != "" {
		return path
	}
	// Try WOW64 path
	return ReadStringWithDefault(RegKeyVirtualBoxWow64, "InstallDir", "")
}

// IsVirtualBoxInstalled checks if VirtualBox is installed via registry
func IsVirtualBoxInstalled() bool {
	return KeyExists(RegKeyVirtualBox) || KeyExists(RegKeyVirtualBoxWow64)
}

// GetProcessorNameString returns CPU name from registry
func GetProcessorNameString() string {
	return ReadStringWithDefault(RegKeyCentralProcessor0, "ProcessorNameString", "")
}

// GetProcessorMHz returns CPU frequency in MHz from registry
func GetProcessorMHz() uint32 {
	return ReadUint32WithDefault(RegKeyCentralProcessor0, "~MHz", 0)
}

// RtlGetVersion wrapper for getting accurate OS version
// This bypasses the compatibility shim that GetVersionEx uses

var (
	modntdll           = windows.NewLazySystemDLL("ntdll.dll")
	procRtlGetVersion  = modntdll.NewProc("RtlGetVersion")
)

// RTL_OSVERSIONINFOW structure
type rtlOSVersionInfoW struct {
	OSVersionInfoSize uint32
	MajorVersion      uint32
	MinorVersion      uint32
	BuildNumber       uint32
	PlatformId        uint32
	CSDVersion        [128]uint16
}

// RtlGetVersion returns accurate Windows version information
// This bypasses compatibility shims that affect GetVersionEx
func RtlGetVersion() (major, minor, build uint32, err error) {
	osvi := rtlOSVersionInfoW{
		OSVersionInfoSize: uint32(unsafe.Sizeof(rtlOSVersionInfoW{})),
	}

	ret, _, _ := procRtlGetVersion.Call(uintptr(unsafe.Pointer(&osvi)))
	if ret != 0 {
		return 0, 0, 0, fmt.Errorf("RtlGetVersion failed with status %d", ret)
	}

	return osvi.MajorVersion, osvi.MinorVersion, osvi.BuildNumber, nil
}

// GetFullVersion returns complete version string (e.g., "10.0.22631.4602")
func GetFullVersion() string {
	major, minor, build, err := RtlGetVersion()
	if err != nil {
		return ""
	}

	ubr := GetUBR()
	if ubr > 0 {
		return fmt.Sprintf("%d.%d.%d.%d", major, minor, build, ubr)
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, build)
}
