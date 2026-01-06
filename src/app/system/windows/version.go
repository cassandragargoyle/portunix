//go:build windows

package windows

import (
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsVersion contains Windows version information
type WindowsVersion struct {
	Major        uint32
	Minor        uint32
	Build        uint32
	UBR          uint32 // Update Build Revision
	ProductName  string
	EditionID    string
	DisplayVer   string // e.g., "23H2"
	BuildLabEx   string
	IsWindows11  bool
	IsWindows10  bool
	IsServer     bool
}

// GetWindowsVersion returns comprehensive Windows version information
// using native APIs instead of external commands
func GetWindowsVersion() (*WindowsVersion, error) {
	major, minor, build, err := RtlGetVersion()
	if err != nil {
		return nil, err
	}

	ver := &WindowsVersion{
		Major:       major,
		Minor:       minor,
		Build:       build,
		UBR:         GetUBR(),
		ProductName: GetProductName(),
		EditionID:   GetEditionID(),
		DisplayVer:  GetDisplayVersion(),
		BuildLabEx:  GetBuildLabEx(),
	}

	// Determine Windows version
	// Windows 11 starts from build 22000
	if major == 10 && minor == 0 {
		if build >= 22000 {
			ver.IsWindows11 = true
		} else {
			ver.IsWindows10 = true
		}
	}

	// Check if server edition
	if ver.EditionID != "" {
		if containsAny(ver.EditionID, "Server", "Datacenter") {
			ver.IsServer = true
		}
	}

	return ver, nil
}

// GetVersionString returns human-readable version string (e.g., "11" or "10")
func (v *WindowsVersion) GetVersionString() string {
	if v.IsWindows11 {
		return "11"
	}
	if v.IsWindows10 {
		return "10"
	}
	return strconv.FormatUint(uint64(v.Major), 10)
}

// GetBuildString returns build number as string
func (v *WindowsVersion) GetBuildString() string {
	return strconv.FormatUint(uint64(v.Build), 10)
}

// GetFullBuildString returns full build string with UBR (e.g., "22631.4602")
func (v *WindowsVersion) GetFullBuildString() string {
	if v.UBR > 0 {
		return strconv.FormatUint(uint64(v.Build), 10) + "." + strconv.FormatUint(uint64(v.UBR), 10)
	}
	return strconv.FormatUint(uint64(v.Build), 10)
}

// Architecture detection using GetNativeSystemInfo

var (
	modkernel32            = windows.NewLazySystemDLL("kernel32.dll")
	procGetNativeSystemInfo = modkernel32.NewProc("GetNativeSystemInfo")
)

// SYSTEM_INFO structure
type systemInfo struct {
	ProcessorArchitecture     uint16
	Reserved                  uint16
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uintptr
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

// Processor architecture constants
const (
	PROCESSOR_ARCHITECTURE_AMD64   = 9
	PROCESSOR_ARCHITECTURE_ARM     = 5
	PROCESSOR_ARCHITECTURE_ARM64   = 12
	PROCESSOR_ARCHITECTURE_IA64    = 6
	PROCESSOR_ARCHITECTURE_INTEL   = 0
	PROCESSOR_ARCHITECTURE_UNKNOWN = 0xFFFF
)

// GetArchitecture returns the system architecture string
func GetArchitecture() string {
	var si systemInfo
	procGetNativeSystemInfo.Call(uintptr(unsafe.Pointer(&si)))

	switch si.ProcessorArchitecture {
	case PROCESSOR_ARCHITECTURE_AMD64:
		return "amd64"
	case PROCESSOR_ARCHITECTURE_ARM64:
		return "arm64"
	case PROCESSOR_ARCHITECTURE_ARM:
		return "arm"
	case PROCESSOR_ARCHITECTURE_IA64:
		return "ia64"
	case PROCESSOR_ARCHITECTURE_INTEL:
		return "386"
	default:
		return "unknown"
	}
}

// GetProcessorCount returns the number of processors
func GetProcessorCount() uint32 {
	var si systemInfo
	procGetNativeSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	return si.NumberOfProcessors
}

// Hostname using GetComputerNameExW

// ComputerNameFormat constants
const (
	ComputerNameNetBIOS                   = 0
	ComputerNameDnsHostname               = 1
	ComputerNameDnsDomain                 = 2
	ComputerNameDnsFullyQualified         = 3
	ComputerNamePhysicalNetBIOS           = 4
	ComputerNamePhysicalDnsHostname       = 5
	ComputerNamePhysicalDnsDomain         = 6
	ComputerNamePhysicalDnsFullyQualified = 7
)

var procGetComputerNameExW = modkernel32.NewProc("GetComputerNameExW")

// GetHostname returns the computer's DNS hostname
func GetHostname() string {
	var size uint32 = 256
	buf := make([]uint16, size)

	ret, _, _ := procGetComputerNameExW.Call(
		uintptr(ComputerNameDnsHostname),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return ""
	}

	return windows.UTF16ToString(buf[:size])
}

// GetFullyQualifiedHostname returns the FQDN
func GetFullyQualifiedHostname() string {
	var size uint32 = 256
	buf := make([]uint16, size)

	ret, _, _ := procGetComputerNameExW.Call(
		uintptr(ComputerNameDnsFullyQualified),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return ""
	}

	return windows.UTF16ToString(buf[:size])
}

// helper function
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
