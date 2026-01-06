//go:build windows

package system

import (
	winapi "portunix.ai/app/system/windows"
)

// init registers native Windows implementations
func init() {
	nativeGetWindowsInfo = getWindowsInfoNativeImpl
	nativeDetectWindowsEnvironment = detectWindowsEnvironmentNativeImpl
	nativeIsAdmin = winapi.IsAdmin
	nativeCheckHardwareVirt = winapi.IsHypervisorPresent
	nativeQueryRegistry = queryWindowsRegistryNativeImpl
	nativeIsVirtualBoxAvailable = winapi.IsVirtualBoxInstalled
}

// getWindowsInfoNativeImpl gets Windows-specific information using native APIs
func getWindowsInfoNativeImpl(info *SystemInfo) error {
	info.OS = "Windows"
	info.WindowsInfo = &WindowsInfo{}

	// Get Windows version using native RtlGetVersion API
	winVer, err := winapi.GetWindowsVersion()
	if err != nil {
		return err
	}

	// Set version info
	info.Version = winVer.GetVersionString()
	info.Build = winVer.GetFullBuildString()
	info.WindowsInfo.ProductName = winVer.ProductName
	info.WindowsInfo.Edition = winVer.EditionID
	info.WindowsInfo.BuildLab = winVer.BuildLabEx

	return nil
}

// detectWindowsEnvironmentNativeImpl detects Windows environments using native APIs
func detectWindowsEnvironmentNativeImpl(info *SystemInfo) {
	// Check for Windows Sandbox
	if winapi.IsSandbox() {
		info.Environment = append(info.Environment, "sandbox")
		info.Variant = "Sandbox"
		return
	}

	// Check for VM using SMBIOS data
	vmType := winapi.DetectVM()
	if vmType != winapi.VMTypeNone {
		info.Environment = append(info.Environment, "vm")
		info.Variant = winapi.GetVariantString()
		return
	}

	// Physical machine
	info.Variant = "Physical"
}

// queryWindowsRegistryNativeImpl queries Windows registry using native API
func queryWindowsRegistryNativeImpl(keyPath, valueName string) string {
	// Parse keyPath to extract root and subkey
	// Format: "HKEY_LOCAL_MACHINE\SOFTWARE\Oracle\VirtualBox"
	var key winapi.RegistryKey

	switch {
	case len(keyPath) > 19 && keyPath[:18] == "HKEY_LOCAL_MACHINE":
		key.Root = 0x80000002 // HKEY_LOCAL_MACHINE
		if len(keyPath) > 19 {
			key.SubKey = keyPath[19:]
		}
	case len(keyPath) > 18 && keyPath[:17] == "HKEY_CURRENT_USER":
		key.Root = 0x80000001 // HKEY_CURRENT_USER
		if len(keyPath) > 18 {
			key.SubKey = keyPath[18:]
		}
	default:
		return ""
	}

	val, err := winapi.ReadString(key, valueName)
	if err != nil {
		return ""
	}
	return val
}
