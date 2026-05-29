/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import "testing"

func TestParseSemver_Valid(t *testing.T) {
	cases := []struct {
		in              string
		major, minor, p int
		pre             string
	}{
		{"1.0.0", 1, 0, 0, ""},
		{"v2.3.4", 2, 3, 4, ""},
		{"0.9.0-alpha.1", 0, 9, 0, "alpha.1"},
		{"1.2.3+build.5", 1, 2, 3, ""},
		{"10.20.30-rc.2+meta", 10, 20, 30, "rc.2"},
	}
	for _, c := range cases {
		got, err := parseSemver(c.in)
		if err != nil {
			t.Errorf("parseSemver(%q) returned error: %v", c.in, err)
			continue
		}
		if got.major != c.major || got.minor != c.minor || got.patch != c.p || got.pre != c.pre {
			t.Errorf("parseSemver(%q) = %+v, want major=%d minor=%d patch=%d pre=%q",
				c.in, got, c.major, c.minor, c.p, c.pre)
		}
	}
}

func TestParseSemver_Invalid(t *testing.T) {
	// Note: we intentionally accept leading zeros (`01.02.03`) — SemVer §2
	// disallows them, but the platform-range matching here only cares about
	// integer ordering, and rejecting them would complicate legitimate
	// manifests migrated from other tooling. Add the leading-zero case back
	// to this slice if we ever tighten the parser.
	cases := []string{"", "1", "1.2", "1.2.3.4", "abc", "1.2.x"}
	for _, s := range cases {
		if _, err := parseSemver(s); err == nil {
			t.Errorf("parseSemver(%q) expected error, got nil", s)
		}
	}
}

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"1.10.0", "1.2.0", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.10", "1.0.0-alpha.2", 1},
		{"1.0.0-rc.1", "1.0.0-rc.1", 0},
	}
	for _, c := range cases {
		a, _ := parseSemver(c.a)
		b, _ := parseSemver(c.b)
		got := compareSemver(a, b)
		if got != c.want {
			t.Errorf("compareSemver(%s, %s) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestMatchesVersionRange(t *testing.T) {
	cases := []struct {
		v, min, max string
		want        bool
	}{
		{"0.5.0", "0.3.0", "0.9.0", true},
		{"0.3.0", "0.3.0", "0.9.0", true}, // inclusive low
		{"0.9.0", "0.3.0", "0.9.0", true}, // inclusive high
		{"0.2.9", "0.3.0", "0.9.0", false},
		{"0.9.1", "0.3.0", "0.9.0", false},
		{"1.0.0", "", "2.0.0", true},  // unbounded low
		{"3.0.0", "", "2.0.0", false}, // unbounded low, above max
		{"1.0.0", "0.5.0", "", true},  // unbounded high
		{"0.4.0", "0.5.0", "", false}, // unbounded high, below min
		{"1.0.0", "", "", true},       // unbounded both
	}
	for _, c := range cases {
		got, err := MatchesVersionRange(c.v, c.min, c.max)
		if err != nil {
			t.Errorf("MatchesVersionRange(%s,%s,%s) error: %v", c.v, c.min, c.max, err)
			continue
		}
		if got != c.want {
			t.Errorf("MatchesVersionRange(%s,%s,%s) = %v, want %v", c.v, c.min, c.max, got, c.want)
		}
	}
}

func TestMatchesVersionRange_InvalidInputs(t *testing.T) {
	if _, err := MatchesVersionRange("not-a-version", "", ""); err == nil {
		t.Error("expected error for invalid version")
	}
	if _, err := MatchesVersionRange("1.0.0", "bad", ""); err == nil {
		t.Error("expected error for invalid min")
	}
	if _, err := MatchesVersionRange("1.0.0", "", "bad"); err == nil {
		t.Error("expected error for invalid max")
	}
}
