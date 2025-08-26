package install

import (
	"fmt"
	"strconv"
	"strings"
)

// SupportLevel represents the level of support for a version
type SupportLevel int

const (
	Unsupported  SupportLevel = iota // Outside range, fallback required
	Experimental                     // Newer versions, with warnings
	Compatible                       // Within range, not tested but likely to work
	Supported                        // Explicitly tested versions
)

func (s SupportLevel) String() string {
	switch s {
	case Supported:
		return "Supported"
	case Compatible:
		return "Compatible"
	case Experimental:
		return "Experimental"
	case Unsupported:
		return "Unsupported"
	default:
		return "Unknown"
	}
}

// VersionRange represents a range of supported versions
type VersionRange struct {
	Min        string `json:"min"`
	Max        string `json:"max"`
	Type       string `json:"type"` // lts_and_interim, lts_only, etc.
	Confidence string `json:"confidence,omitempty"` // high, medium, low
}

// VersionMatcher handles version matching logic
type VersionMatcher struct{}

// NewVersionMatcher creates a new version matcher instance
func NewVersionMatcher() *VersionMatcher {
	return &VersionMatcher{}
}

// IsVersionSupported checks if a version is supported within given ranges
func (vm *VersionMatcher) IsVersionSupported(version string, ranges []VersionRange, explicitVersions []string) SupportLevel {
	// Check for explicit support first (highest priority)
	for _, explicitVersion := range explicitVersions {
		if explicitVersion == version {
			return Supported
		}
	}

	// Check version ranges
	for _, vRange := range ranges {
		if vm.isVersionInRange(version, vRange.Min, vRange.Max) {
			// Determine confidence level based on version characteristics
			if vm.isVersionExperimental(version, vRange.Max) {
				return Experimental
			}
			return Compatible
		}
	}

	return Unsupported
}

// isVersionInRange checks if version is within min-max range (inclusive)
func (vm *VersionMatcher) isVersionInRange(version, min, max string) bool {
	// Handle Ubuntu-style versions (e.g., "20.04", "22.04")
	if vm.isUbuntuStyleVersion(version) && vm.isUbuntuStyleVersion(min) && vm.isUbuntuStyleVersion(max) {
		return vm.compareUbuntuVersions(version, min) >= 0 && vm.compareUbuntuVersions(version, max) <= 0
	}

	// Handle simple numeric versions (e.g., "11", "12")
	if vm.isSimpleNumeric(version) && vm.isSimpleNumeric(min) && vm.isSimpleNumeric(max) {
		versionNum, _ := strconv.Atoi(version)
		minNum, _ := strconv.Atoi(min)
		maxNum, _ := strconv.Atoi(max)
		return versionNum >= minNum && versionNum <= maxNum
	}

	// Fallback to string comparison for other formats
	return version >= min && version <= max
}

// isVersionExperimental determines if a version should be marked as experimental
func (vm *VersionMatcher) isVersionExperimental(version, maxSupported string) bool {
	// For Ubuntu versions, consider anything above current LTS as experimental
	if vm.isUbuntuStyleVersion(version) && vm.isUbuntuStyleVersion(maxSupported) {
		versionComparison := vm.compareUbuntuVersions(version, maxSupported)
		// If version is significantly higher than max supported, it's experimental
		return versionComparison > 0
	}
	return false
}

// isUbuntuStyleVersion checks if version follows Ubuntu format (XX.YY)
func (vm *VersionMatcher) isUbuntuStyleVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return false
	}
	
	// Check if both parts are numeric
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

// compareUbuntuVersions compares two Ubuntu-style versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (vm *VersionMatcher) compareUbuntuVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	if len(parts1) != 2 || len(parts2) != 2 {
		// Fallback to string comparison
		if v1 < v2 {
			return -1
		} else if v1 > v2 {
			return 1
		}
		return 0
	}
	
	// Compare major version first
	major1, _ := strconv.Atoi(parts1[0])
	major2, _ := strconv.Atoi(parts2[0])
	
	if major1 != major2 {
		if major1 < major2 {
			return -1
		}
		return 1
	}
	
	// Compare minor version
	minor1, _ := strconv.Atoi(parts1[1])
	minor2, _ := strconv.Atoi(parts2[1])
	
	if minor1 < minor2 {
		return -1
	} else if minor1 > minor2 {
		return 1
	}
	return 0
}

// isSimpleNumeric checks if version is a simple numeric string
func (vm *VersionMatcher) isSimpleNumeric(version string) bool {
	_, err := strconv.Atoi(version)
	return err == nil
}

// GetVersionSupportMessage returns a user-friendly message about version support
func (vm *VersionMatcher) GetVersionSupportMessage(version string, supportLevel SupportLevel) string {
	switch supportLevel {
	case Supported:
		return fmt.Sprintf("✅ Version %s is fully supported and tested", version)
	case Compatible:
		return fmt.Sprintf("🟡 Version %s is within supported range but not explicitly tested", version)
	case Experimental:
		return fmt.Sprintf("⚠️  Version %s is experimental - newer than tested versions", version)
	case Unsupported:
		return fmt.Sprintf("❌ Version %s is not supported", version)
	default:
		return fmt.Sprintf("❓ Version %s support status unknown", version)
	}
}

// GetRecommendedAction returns recommended action based on support level
func (vm *VersionMatcher) GetRecommendedAction(supportLevel SupportLevel, fallbackVariants []string) string {
	switch supportLevel {
	case Supported:
		return "Proceed with installation"
	case Compatible:
		return "Proceed with installation (compatibility mode)"
	case Experimental:
		return "Proceed with caution - consider using fallback variant if available"
	case Unsupported:
		if len(fallbackVariants) > 0 {
			return fmt.Sprintf("Use fallback variant: %s", strings.Join(fallbackVariants, ", "))
		}
		return "Installation not recommended - version not supported"
	default:
		return "Cannot determine recommended action"
	}
}