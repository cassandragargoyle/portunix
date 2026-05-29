/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package git

import "testing"

func TestNormalizeRepoSpec(t *testing.T) {
	cases := []struct {
		in   string
		want string
		err  bool
	}{
		{"cassandragargoyle/portunix", "https://github.com/cassandragargoyle/portunix.git", false},
		{"https://github.com/foo/bar.git", "https://github.com/foo/bar.git", false},
		{"git@github.com:foo/bar.git", "git@github.com:foo/bar.git", false},
		{"", "", true},
		{"only-owner", "", true},
	}
	for _, c := range cases {
		got, err := NormalizeRepoSpec(c.in)
		if c.err {
			if err == nil {
				t.Errorf("%q: expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q want %q", c.in, got, c.want)
		}
	}
}

func TestSplitOwnerRepo(t *testing.T) {
	cases := []struct {
		in    string
		owner string
		repo  string
		err   bool
	}{
		{"foo/bar", "foo", "bar", false},
		{"https://github.com/foo/bar.git", "foo", "bar", false},
		{"git@github.com:foo/bar.git", "foo", "bar", false},
		{"justone", "", "", true},
	}
	for _, c := range cases {
		o, r, err := SplitOwnerRepo(c.in)
		if c.err {
			if err == nil {
				t.Errorf("%q: expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected: %v", c.in, err)
			continue
		}
		if o != c.owner || r != c.repo {
			t.Errorf("%q: got %s/%s want %s/%s", c.in, o, r, c.owner, c.repo)
		}
	}
}
