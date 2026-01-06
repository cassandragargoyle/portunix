package engine

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// DriveInfo represents information about a Windows drive
type DriveInfo struct {
	Letter     string
	FreeSpace  string
	TotalSpace string
}

// PartitionInfo represents information about a Linux partition
type PartitionInfo struct {
	Device     string
	MountPoint string
	FreeSpace  string
	TotalSpace string
	FileSystem string
}

// StorageAnalyzer provides storage analysis functionality
type StorageAnalyzer struct {
	minSpace int64 // Minimum required space in bytes
}

// NewStorageAnalyzer creates a new storage analyzer with specified minimum space
func NewStorageAnalyzer(minSpaceGB int64) *StorageAnalyzer {
	return &StorageAnalyzer{
		minSpace: minSpaceGB * 1024 * 1024 * 1024, // Convert GB to bytes
	}
}

// AnalyzeStorage analyzes available storage and returns recommended path
func (s *StorageAnalyzer) AnalyzeStorage() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return s.analyzeWindowsStorage()
	case "linux":
		return s.analyzeLinuxStorage()
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetWindowsDrives returns sorted list of Windows drives
func (s *StorageAnalyzer) GetWindowsDrives() ([]DriveInfo, error) {
	// Get available drives using PowerShell
	psCmd := `Get-WmiObject -Class Win32_LogicalDisk | Where-Object {$_.DriveType -eq 3} | Select-Object DeviceID, FreeSpace, Size`

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get drive information: %w", err)
	}

	var drives []DriveInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "DeviceID") || strings.HasPrefix(line, "---") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			driveLetter := strings.TrimSuffix(fields[0], ":")
			freeSpace := formatBytes(parseNumber(fields[1]))
			totalSpace := formatBytes(parseNumber(fields[2]))

			drives = append(drives, DriveInfo{
				Letter:     driveLetter,
				FreeSpace:  freeSpace,
				TotalSpace: totalSpace,
			})
		}
	}

	if len(drives) == 0 {
		return nil, fmt.Errorf("no accessible drives found")
	}

	// Sort drives: drives with sufficient space first, non-C drives preferred, then by free space
	sort.Slice(drives, func(i, j int) bool {
		spaceI := parseSpaceString(drives[i].FreeSpace)
		spaceJ := parseSpaceString(drives[j].FreeSpace)

		// Check if drives have sufficient space
		hasSufficientI := spaceI >= s.minSpace
		hasSufficientJ := spaceJ >= s.minSpace

		// Drives with insufficient space should be sorted last
		if hasSufficientI && !hasSufficientJ {
			return true
		}
		if !hasSufficientI && hasSufficientJ {
			return false
		}

		// Both have sufficient space: prioritize non-C drives
		if hasSufficientI && hasSufficientJ {
			if drives[i].Letter != "C" && drives[j].Letter == "C" {
				return true
			}
			if drives[i].Letter == "C" && drives[j].Letter != "C" {
				return false
			}
		}

		// Sort by free space (descending)
		return spaceI > spaceJ
	})

	return drives, nil
}

// GetLinuxPartitions returns sorted list of Linux partitions
func (s *StorageAnalyzer) GetLinuxPartitions() ([]PartitionInfo, error) {
	cmd := exec.Command("df", "-B1", "--output=source,target,avail,size,fstype")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get partition information: %w", err)
	}

	var partitions []PartitionInfo
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// Skip tmpfs, devtmpfs, etc.
			if strings.HasPrefix(fields[0], "tmpfs") || strings.HasPrefix(fields[0], "devtmpfs") {
				continue
			}

			partitions = append(partitions, PartitionInfo{
				Device:     fields[0],
				MountPoint: fields[1],
				FreeSpace:  formatBytes(parseNumber(fields[2])),
				TotalSpace: formatBytes(parseNumber(fields[3])),
				FileSystem: fields[4],
			})
		}
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("no accessible partitions found")
	}

	// Sort partitions: sufficient space first, then by free space
	sort.Slice(partitions, func(i, j int) bool {
		spaceI := parseSpaceString(partitions[i].FreeSpace)
		spaceJ := parseSpaceString(partitions[j].FreeSpace)

		hasSufficientI := spaceI >= s.minSpace
		hasSufficientJ := spaceJ >= s.minSpace

		if hasSufficientI && !hasSufficientJ {
			return true
		}
		if !hasSufficientI && hasSufficientJ {
			return false
		}

		return spaceI > spaceJ
	})

	return partitions, nil
}

func (s *StorageAnalyzer) analyzeWindowsStorage() (string, error) {
	drives, err := s.GetWindowsDrives()
	if err != nil {
		return "", err
	}

	if len(drives) == 0 {
		return "", fmt.Errorf("no drives with sufficient space (>= %d GB)", s.minSpace/(1024*1024*1024))
	}

	// Check if first drive has sufficient space
	firstDriveSpace := parseSpaceString(drives[0].FreeSpace)
	if firstDriveSpace < s.minSpace {
		return "", fmt.Errorf("no drives with sufficient space (>= %d GB). Best option: %s:\\ with %s",
			s.minSpace/(1024*1024*1024), drives[0].Letter, drives[0].FreeSpace)
	}

	return drives[0].Letter, nil
}

func (s *StorageAnalyzer) analyzeLinuxStorage() (string, error) {
	partitions, err := s.GetLinuxPartitions()
	if err != nil {
		return "", err
	}

	if len(partitions) == 0 {
		return "", fmt.Errorf("no partitions with sufficient space (>= %d GB)", s.minSpace/(1024*1024*1024))
	}

	// Check if first partition has sufficient space
	firstPartitionSpace := parseSpaceString(partitions[0].FreeSpace)
	if firstPartitionSpace < s.minSpace {
		return "", fmt.Errorf("no partitions with sufficient space (>= %d GB). Best option: %s with %s",
			s.minSpace/(1024*1024*1024), partitions[0].MountPoint, partitions[0].FreeSpace)
	}

	return partitions[0].MountPoint, nil
}

// parseSpaceString converts a human-readable space string to bytes
func parseSpaceString(spaceStr string) int64 {
	if spaceStr == "Unknown" || spaceStr == "" {
		return 0
	}

	// Remove spaces and convert to uppercase
	spaceStr = strings.ReplaceAll(strings.ToUpper(spaceStr), " ", "")

	// Extract number and unit
	var value float64
	var unit string
	if n, err := fmt.Sscanf(spaceStr, "%f%s", &value, &unit); n == 2 && err == nil {
		multiplier := int64(1)
		switch unit {
		case "B":
			multiplier = 1
		case "KB":
			multiplier = 1024
		case "MB":
			multiplier = 1024 * 1024
		case "GB":
			multiplier = 1024 * 1024 * 1024
		case "TB":
			multiplier = 1024 * 1024 * 1024 * 1024
		}
		return int64(value * float64(multiplier))
	}

	return 0
}

// parseNumber parses a number string to int64
func parseNumber(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// formatBytes formats bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
