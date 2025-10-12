package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	Build string
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (*Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Special case for development version
	if versionStr == "dev" {
		return &Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
			Pre:   "dev",
			Build: "",
		}, nil
	}

	// Regex to match semantic version pattern
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	matches := re.FindStringSubmatch(versionStr)

	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   matches[4],
		Build: matches[5],
	}, nil
}

// String returns the string representation of the version
func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		version += "-" + v.Pre
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// IsCompatible checks if this version is compatible with another version
// For now, we check major and minor version compatibility
func (v *Version) IsCompatible(other *Version) bool {
	// Development versions are always compatible during development
	if v.Pre == "dev" || other.Pre == "dev" {
		return true
	}

	// Same major and minor version are compatible
	return v.Major == other.Major && v.Minor == other.Minor
}

// Compare compares two versions and returns:
// -1 if v < other
//  0 if v == other
//  1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// For simplicity, consider pre-release versions equal if major.minor.patch are equal
	return 0
}