//go:build windows

package windows

import (
	"bytes"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SMBIOS table signature
const (
	RSMB = 0x52534D42 // 'RSMB' in little-endian
)

// SMBIOS structure types
const (
	SMBIOSTypeBIOS        = 0
	SMBIOSTypeSystem      = 1
	SMBIOSTypeBaseboard   = 2
	SMBIOSTypeChassis     = 3
	SMBIOSTypeProcessor   = 4
	SMBIOSTypeEndOfTable  = 127
)

var procGetSystemFirmwareTable = modkernel32.NewProc("GetSystemFirmwareTable")

// RawSMBIOSData represents the raw SMBIOS data header
type RawSMBIOSData struct {
	Used20CallingMethod byte
	SMBIOSMajorVersion  byte
	SMBIOSMinorVersion  byte
	DmiRevision         byte
	Length              uint32
	// SMBIOSTableData follows
}

// SMBIOSHeader represents a SMBIOS structure header
type SMBIOSHeader struct {
	Type   byte
	Length byte
	Handle uint16
}

// SystemInfo contains SMBIOS system information (Type 1)
type SystemInfo struct {
	Manufacturer string
	ProductName  string
	Version      string
	SerialNumber string
	UUID         string
	SKUNumber    string
	Family       string
}

// GetSMBIOSData retrieves raw SMBIOS data using GetSystemFirmwareTable
func GetSMBIOSData() ([]byte, error) {
	// First call to get required buffer size
	size, _, _ := procGetSystemFirmwareTable.Call(
		uintptr(RSMB),
		0,
		0,
		0,
	)

	if size == 0 {
		return nil, windows.GetLastError()
	}

	// Allocate buffer and retrieve data
	buffer := make([]byte, size)
	ret, _, err := procGetSystemFirmwareTable.Call(
		uintptr(RSMB),
		0,
		uintptr(unsafe.Pointer(&buffer[0])),
		size,
	)

	if ret == 0 {
		return nil, err
	}

	return buffer, nil
}

// GetSystemInfo retrieves SMBIOS system information (Type 1)
func GetSystemInfo() (*SystemInfo, error) {
	data, err := GetSMBIOSData()
	if err != nil {
		return nil, err
	}

	if len(data) < int(unsafe.Sizeof(RawSMBIOSData{})) {
		return nil, windows.ERROR_INVALID_DATA
	}

	// Parse header
	header := (*RawSMBIOSData)(unsafe.Pointer(&data[0]))
	tableData := data[8 : 8+header.Length]

	return parseSystemInfo(tableData)
}

// parseSystemInfo parses SMBIOS Type 1 (System Information) structure
func parseSystemInfo(data []byte) (*SystemInfo, error) {
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		header := SMBIOSHeader{
			Type:   data[offset],
			Length: data[offset+1],
			Handle: uint16(data[offset+2]) | uint16(data[offset+3])<<8,
		}

		if header.Type == SMBIOSTypeEndOfTable {
			break
		}

		// Get the formatted section
		formattedEnd := offset + int(header.Length)
		if formattedEnd > len(data) {
			break
		}

		// Find strings section (after double null terminator)
		stringsStart := formattedEnd
		stringsEnd := stringsStart

		// Find the end of strings section (double null terminator)
		for stringsEnd < len(data)-1 {
			if data[stringsEnd] == 0 && data[stringsEnd+1] == 0 {
				stringsEnd += 2
				break
			}
			stringsEnd++
		}
		if stringsEnd >= len(data) {
			stringsEnd = len(data)
		}

		strings := extractStrings(data[stringsStart:stringsEnd])

		if header.Type == SMBIOSTypeSystem && header.Length >= 8 {
			info := &SystemInfo{}

			// String indices are 1-based, 0 means no string
			if header.Length > 4 && data[offset+4] > 0 && int(data[offset+4]) <= len(strings) {
				info.Manufacturer = strings[data[offset+4]-1]
			}
			if header.Length > 5 && data[offset+5] > 0 && int(data[offset+5]) <= len(strings) {
				info.ProductName = strings[data[offset+5]-1]
			}
			if header.Length > 6 && data[offset+6] > 0 && int(data[offset+6]) <= len(strings) {
				info.Version = strings[data[offset+6]-1]
			}
			if header.Length > 7 && data[offset+7] > 0 && int(data[offset+7]) <= len(strings) {
				info.SerialNumber = strings[data[offset+7]-1]
			}
			// SKU and Family are at higher offsets (0x19 and 0x1A) in SMBIOS 2.4+
			if header.Length > 0x19 && data[offset+0x19] > 0 && int(data[offset+0x19]) <= len(strings) {
				info.SKUNumber = strings[data[offset+0x19]-1]
			}
			if header.Length > 0x1A && data[offset+0x1A] > 0 && int(data[offset+0x1A]) <= len(strings) {
				info.Family = strings[data[offset+0x1A]-1]
			}

			return info, nil
		}

		// Move to next structure
		offset = stringsEnd
	}

	return nil, windows.ERROR_NOT_FOUND
}

// extractStrings extracts null-terminated strings from SMBIOS data
func extractStrings(data []byte) []string {
	var result []string
	var current bytes.Buffer

	for _, b := range data {
		if b == 0 {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(b)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// VMType represents detected VM type
type VMType string

const (
	VMTypeNone       VMType = ""
	VMTypeVirtualBox VMType = "VirtualBox"
	VMTypeVMware     VMType = "VMware"
	VMTypeHyperV     VMType = "Hyper-V"
	VMTypeQEMU       VMType = "QEMU"
	VMTypeXen        VMType = "Xen"
	VMTypeKVM        VMType = "KVM"
	VMTypeParallels  VMType = "Parallels"
	VMTypeUnknown    VMType = "Unknown VM"
)

// DetectVM detects if running in a virtual machine using SMBIOS data
func DetectVM() VMType {
	info, err := GetSystemInfo()
	if err != nil {
		return VMTypeNone
	}

	// Check manufacturer and product name for VM indicators
	manufacturer := strings.ToLower(info.Manufacturer)
	productName := strings.ToLower(info.ProductName)
	combined := manufacturer + " " + productName

	switch {
	case strings.Contains(combined, "virtualbox"):
		return VMTypeVirtualBox
	case strings.Contains(combined, "vmware"):
		return VMTypeVMware
	case strings.Contains(combined, "hyper-v") || strings.Contains(combined, "microsoft virtual"):
		return VMTypeHyperV
	case strings.Contains(combined, "qemu"):
		return VMTypeQEMU
	case strings.Contains(combined, "xen"):
		return VMTypeXen
	case strings.Contains(combined, "kvm"):
		return VMTypeKVM
	case strings.Contains(combined, "parallels"):
		return VMTypeParallels
	case strings.Contains(combined, "virtual"):
		return VMTypeUnknown
	}

	return VMTypeNone
}

// IsVM returns true if running in a virtual machine
func IsVM() bool {
	return DetectVM() != VMTypeNone
}

// GetSMBIOSManufacturer returns the system manufacturer from SMBIOS
func GetSMBIOSManufacturer() string {
	info, err := GetSystemInfo()
	if err != nil {
		return ""
	}
	return info.Manufacturer
}

// GetSMBIOSProductName returns the system product name from SMBIOS
func GetSMBIOSProductName() string {
	info, err := GetSystemInfo()
	if err != nil {
		return ""
	}
	return info.ProductName
}
