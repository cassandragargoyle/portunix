/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"testing"

	"github.com/google/go-github/v56/github"
)

func makeAssets(names ...string) []*github.ReleaseAsset {
	out := make([]*github.ReleaseAsset, 0, len(names))
	for _, n := range names {
		name := n
		out = append(out, &github.ReleaseAsset{Name: &name})
	}
	return out
}

func TestFindAssetByPlatform(t *testing.T) {
	assets := makeAssets(
		"portunix-linux-amd64.tar.gz",
		"portunix-linux-arm64.tar.gz",
		"portunix-windows-amd64.zip",
		"portunix-darwin-amd64.tar.gz",
		"checksums.txt",
	)

	cases := []struct {
		platform string
		want     string
		nil      bool
	}{
		{"linux-amd64", "portunix-linux-amd64.tar.gz", false},
		{"linux-x86_64", "portunix-linux-amd64.tar.gz", false},
		{"windows-amd64", "portunix-windows-amd64.zip", false},
		{"darwin-arm64", "", true},
		{"linux-arm64", "portunix-linux-arm64.tar.gz", false},
	}
	for _, c := range cases {
		got := findAssetByPlatform(assets, c.platform)
		if c.nil {
			if got != nil {
				t.Errorf("%s: expected nil, got %s", c.platform, got.GetName())
			}
			continue
		}
		if got == nil {
			t.Errorf("%s: expected match", c.platform)
			continue
		}
		if got.GetName() != c.want {
			t.Errorf("%s: got %s want %s", c.platform, got.GetName(), c.want)
		}
	}
}

func TestFindAssetByName(t *testing.T) {
	assets := makeAssets("a.txt", "b.txt")
	if findAssetByName(assets, "missing") != nil {
		t.Errorf("expected nil for missing")
	}
	got := findAssetByName(assets, "b.txt")
	if got == nil || got.GetName() != "b.txt" {
		t.Errorf("expected b.txt, got %v", got)
	}
}

func TestParseOwnerRepoFlexible(t *testing.T) {
	cases := []struct {
		in    string
		owner string
		repo  string
		err   bool
	}{
		{"foo/bar", "foo", "bar", false},
		{"foo/bar.git", "foo", "bar", false},
		{"https://github.com/foo/bar", "foo", "bar", false},
		{"https://github.com/foo/bar.git", "foo", "bar", false},
		{"git@github.com:foo/bar.git", "foo", "bar", false},
		{"only-one", "", "", true},
	}
	for _, c := range cases {
		o, r, err := parseOwnerRepoFlexible(c.in)
		if c.err {
			if err == nil {
				t.Errorf("%q: expected err", c.in)
			}
			continue
		}
		if err != nil || o != c.owner || r != c.repo {
			t.Errorf("%q: got %s/%s err=%v want %s/%s", c.in, o, r, err, c.owner, c.repo)
		}
	}
}
