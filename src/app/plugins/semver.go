/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semverPattern matches major.minor.patch with optional pre-release and build
// suffixes, per the shape required by the plugin-manifest schema (v1.1.0).
var semverPattern = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)

// semver is the minimal version representation used for supported_platforms
// range matching. Build metadata is intentionally discarded per SemVer §10.
type semver struct {
	major int
	minor int
	patch int
	pre   string
}

// parseSemver parses a SemVer string. Leading 'v' is stripped for leniency,
// matching the convention of shared.ParseVersion.
func parseSemver(s string) (semver, error) {
	s = strings.TrimPrefix(s, "v")
	m := semverPattern.FindStringSubmatch(s)
	if m == nil {
		return semver{}, fmt.Errorf("invalid semver: %q", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return semver{major: major, minor: minor, patch: patch, pre: m[4]}, nil
}

// compareSemver returns -1, 0, or 1 for a vs b. Pre-release ordering follows
// SemVer §11: a version with a pre-release suffix ranks below the same version
// without one. Within pre-releases, identifiers are compared field-by-field,
// numeric identifiers numerically and alphanumeric identifiers lexically.
func compareSemver(a, b semver) int {
	if c := cmpInt(a.major, b.major); c != 0 {
		return c
	}
	if c := cmpInt(a.minor, b.minor); c != 0 {
		return c
	}
	if c := cmpInt(a.patch, b.patch); c != 0 {
		return c
	}
	switch {
	case a.pre == "" && b.pre == "":
		return 0
	case a.pre == "":
		return 1
	case b.pre == "":
		return -1
	}
	return comparePre(a.pre, b.pre)
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func comparePre(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	n := len(aParts)
	if len(bParts) < n {
		n = len(bParts)
	}
	for i := 0; i < n; i++ {
		ai, aErr := strconv.Atoi(aParts[i])
		bi, bErr := strconv.Atoi(bParts[i])
		switch {
		case aErr == nil && bErr == nil:
			if c := cmpInt(ai, bi); c != 0 {
				return c
			}
		case aErr == nil:
			return -1
		case bErr == nil:
			return 1
		default:
			if c := strings.Compare(aParts[i], bParts[i]); c != 0 {
				return c
			}
		}
	}
	return cmpInt(len(aParts), len(bParts))
}

// inRange reports whether v satisfies the closed interval [min, max]. Empty
// minStr / maxStr means "no bound". Bounds are inclusive on both ends.
func inRange(v, minStr, maxStr string) (bool, error) {
	ver, err := parseSemver(v)
	if err != nil {
		return false, err
	}
	if minStr != "" {
		lo, err := parseSemver(minStr)
		if err != nil {
			return false, err
		}
		if compareSemver(ver, lo) < 0 {
			return false, nil
		}
	}
	if maxStr != "" {
		hi, err := parseSemver(maxStr)
		if err != nil {
			return false, err
		}
		if compareSemver(ver, hi) > 0 {
			return false, nil
		}
	}
	return true, nil
}

// MatchesVersionRange is the exported form of inRange used by the registry
// query path to filter supported_platforms entries by the caller's platform
// version. See inRange for semantics.
func MatchesVersionRange(v, minStr, maxStr string) (bool, error) {
	return inRange(v, minStr, maxStr)
}
