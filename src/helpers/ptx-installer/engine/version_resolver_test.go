/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// pythonListing mimics the python.org FTP index: an upcoming 3.15.0 directory
// without installers, the current 3.14.8 and older releases
const pythonListing = `<html><body>
<a href="../">../</a>
<a href="2.7.18/">2.7.18/</a>
<a href="3.9.25/">3.9.25/</a>
<a href="3.13.6/">3.13.6/</a>
<a href="3.14.8/">3.14.8/</a>
<a href="3.15.0/">3.15.0/</a>
<a href="3.14.10/">3.14.10/</a>
</body></html>`

// newPythonOrgServer serves the listing and answers 200 only for installers
// of the given published versions
func newPythonOrgServer(t *testing.T, published ...string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ftp/python/" {
			fmt.Fprint(w, pythonListing)
			return
		}
		for _, v := range published {
			if strings.HasPrefix(r.URL.Path, "/ftp/python/"+v+"/") {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestParseReleaseVersions_NewestFirst(t *testing.T) {
	got := strings.Join(parseReleaseVersions(pythonListing), " ")
	want := "3.15.0 3.14.10 3.14.8 3.13.6 3.9.25 2.7.18"
	if got != want {
		t.Errorf("parseReleaseVersions() = %q, want %q", got, want)
	}
}

// TestResolvePythonOrgVersion_SkipsUnpublished verifies that a listed version
// without an installer (upcoming 3.15.0) is skipped (issue #202)
func TestResolvePythonOrgVersion_SkipsUnpublished(t *testing.T) {
	server := newPythonOrgServer(t, "3.14.8", "3.13.6")

	version, err := resolvePythonOrgVersion(server.URL+"/ftp/python/", func(v string) string {
		return server.URL + "/ftp/python/" + v + "/python-" + v + "-amd64.exe"
	})
	if err != nil {
		t.Fatalf("resolvePythonOrgVersion failed: %v", err)
	}
	if version != "3.14.8" {
		t.Errorf("resolved version = %s, want 3.14.8", version)
	}
}

func TestResolvePythonOrgVersion_NoInstallerFails(t *testing.T) {
	server := newPythonOrgServer(t)

	_, err := resolvePythonOrgVersion(server.URL+"/ftp/python/", func(v string) string {
		return server.URL + "/ftp/python/" + v + "/python-" + v + "-amd64.exe"
	})
	if err == nil {
		t.Fatal("expected error when no candidate has an installer, got nil")
	}
}

// TestResolveVariantVersion_SubstitutesURLs covers the variant-level flow
func TestResolveVariantVersion_SubstitutesURLs(t *testing.T) {
	server := newPythonOrgServer(t, "3.14.10")
	original := pythonOrgFTP
	pythonOrgFTP = server.URL + "/ftp/python/"
	t.Cleanup(func() { pythonOrgFTP = original })

	base := server.URL + "/ftp/python/{version}/python-{version}"
	variant := registry.VariantSpec{
		Version:         "latest",
		VersionResolver: "python.org",
		URLs: map[string]string{
			"x64":   base + "-amd64.exe",
			"x86":   base + ".exe",
			"arm64": base + "-arm64.exe",
			"amd64": base + "-amd64.exe",
			"386":   base + ".exe",
		},
	}
	if err := resolveVariantVersion(&variant); err != nil {
		t.Fatalf("resolveVariantVersion failed: %v", err)
	}
	if variant.Version != "3.14.10" {
		t.Errorf("version = %s, want 3.14.10", variant.Version)
	}
	for arch, url := range variant.URLs {
		if strings.Contains(url, versionPlaceholder) || !strings.Contains(url, "/3.14.10/python-3.14.10") {
			t.Errorf("URL for %s not resolved: %s", arch, url)
		}
	}
}

func TestResolveVariantVersion_UnknownResolver(t *testing.T) {
	variant := registry.VariantSpec{VersionResolver: "nowhere"}
	if err := resolveVariantVersion(&variant); err == nil {
		t.Error("expected error for an unknown resolver, got nil")
	}
}

func TestResolveVariantVersion_NoResolverUnchanged(t *testing.T) {
	variant := registry.VariantSpec{Version: "3.14.8", URL: "https://example.invalid/{version}"}
	if err := resolveVariantVersion(&variant); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if variant.Version != "3.14.8" || variant.URL != "https://example.invalid/{version}" {
		t.Errorf("variant without resolver was modified: %+v", variant)
	}
}

// TestAutoDetectVariant_PreferredWins verifies that a preferred variant is
// chosen deterministically when no package manager variant matches
func TestAutoDetectVariant_PreferredWins(t *testing.T) {
	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"embeddable": {Version: "3.14.8", Preferred: true},
		"full":       {Version: "3.14.8"},
		"latest":     {Version: "latest"},
	}}
	installer := &Installer{}
	for run := 0; run < 20; run++ {
		if got := installer.autoDetectVariant(&platform); got != "embeddable" {
			t.Fatalf("autoDetectVariant() = %s, want embeddable", got)
		}
	}
}

// TestAutoDetectVariant_StableFallback verifies the last-resort choice no
// longer depends on random map iteration order
func TestAutoDetectVariant_StableFallback(t *testing.T) {
	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"zeta":  {Version: "1"},
		"alpha": {Version: "1"},
		"mid":   {Version: "1"},
	}}
	installer := &Installer{}
	for run := 0; run < 20; run++ {
		if got := installer.autoDetectVariant(&platform); got != "alpha" {
			t.Fatalf("autoDetectVariant() = %s, want alpha", got)
		}
	}
}

// unreachableListing points the python.org resolver at a server that fails
func unreachableListing(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	original := pythonOrgFTP
	pythonOrgFTP = server.URL + "/ftp/python/"
	t.Cleanup(func() { pythonOrgFTP = original })
}

// TestResolveOrFallback_UsesFallbackVariant verifies that an unreachable
// release listing installs the pinned fallback variant (issue #202)
func TestResolveOrFallback_UsesFallbackVariant(t *testing.T) {
	unreachableListing(t)
	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"full":   {Version: "3.14.8", URL: "https://example.invalid/3.14.8.exe"},
		"latest": {Version: "latest", VersionResolver: "python.org", FallbackVariant: "full", URL: "https://example.invalid/{version}.exe"},
	}}

	name, spec, err := resolveOrFallback(&platform, "latest", platform.Variants["latest"])
	if err != nil {
		t.Fatalf("expected fallback, got error: %v", err)
	}
	if name != "full" || spec.Version != "3.14.8" {
		t.Errorf("resolveOrFallback() = %s (%s), want full (3.14.8)", name, spec.Version)
	}
}

func TestResolveOrFallback_NoFallbackFails(t *testing.T) {
	unreachableListing(t)
	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"latest": {Version: "latest", VersionResolver: "python.org", URL: "https://example.invalid/{version}.exe"},
	}}

	if _, _, err := resolveOrFallback(&platform, "latest", platform.Variants["latest"]); err == nil {
		t.Fatal("expected error without a fallback variant, got nil")
	}
}

// TestResolveOrFallback_ResolvingFallbackRejected guards against a fallback
// that itself needs resolving (it would fail the same way)
func TestResolveOrFallback_ResolvingFallbackRejected(t *testing.T) {
	unreachableListing(t)
	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"other":  {Version: "latest", VersionResolver: "python.org"},
		"latest": {Version: "latest", VersionResolver: "python.org", FallbackVariant: "other"},
	}}

	if _, _, err := resolveOrFallback(&platform, "latest", platform.Variants["latest"]); err == nil {
		t.Fatal("expected error for a fallback that also resolves, got nil")
	}
}

// gitForWindowsRelease mimics the GitHub API answer for a Git for Windows
// release; the asset version differs from the tag version
const gitForWindowsRelease = `{
  "tag_name": "v2.56.0.windows.2",
  "assets": [
    {"name": "Git-2.56.0.2-32-bit.exe", "browser_download_url": "https://example.invalid/Git-2.56.0.2-32-bit.exe"},
    {"name": "Git-2.56.0.2-64-bit.exe", "browser_download_url": "https://example.invalid/Git-2.56.0.2-64-bit.exe", "digest": "sha256:aaaa"},
    {"name": "Git-2.56.0.2-64-bit.tar.bz2", "browser_download_url": "https://example.invalid/Git-2.56.0.2-64-bit.tar.bz2"},
    {"name": "Git-2.56.0.2-arm64.exe", "browser_download_url": "https://example.invalid/Git-2.56.0.2-arm64.exe", "digest": "sha256:bbbb"},
    {"name": "PortableGit-2.56.0.2-64-bit.7z.exe", "browser_download_url": "https://example.invalid/PortableGit-2.56.0.2-64-bit.7z.exe"}
  ]
}`

// newGitForWindowsServer serves body with the given status as the release API
func newGitForWindowsServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

// TestResolveGitForWindows_FillsURLsAndChecksums verifies that the installer
// assets of every architecture are picked from the release (issue #222)
func TestResolveGitForWindows_FillsURLsAndChecksums(t *testing.T) {
	server := newGitForWindowsServer(t, http.StatusOK, gitForWindowsRelease)
	variant := registry.VariantSpec{Type: "exe", Version: "latest", VersionResolver: "git-for-windows"}

	resolved, err := resolveGitForWindows(server.URL, variant)
	if err != nil {
		t.Fatalf("resolveGitForWindows failed: %v", err)
	}
	if resolved.Version != "2.56.0.2" {
		t.Errorf("Version = %s, want 2.56.0.2", resolved.Version)
	}
	wantURLs := map[string]string{
		"x64":   "https://example.invalid/Git-2.56.0.2-64-bit.exe",
		"arm64": "https://example.invalid/Git-2.56.0.2-arm64.exe",
	}
	wantChecksums := map[string]string{"x64": "sha256:aaaa", "arm64": "sha256:bbbb"}
	for arch, url := range wantURLs {
		if resolved.URLs[arch] != url {
			t.Errorf("URLs[%s] = %q, want %q", arch, resolved.URLs[arch], url)
		}
		if resolved.Checksum[arch] != wantChecksums[arch] {
			t.Errorf("Checksum[%s] = %q, want %q", arch, resolved.Checksum[arch], wantChecksums[arch])
		}
	}
	if len(resolved.URLs) != 2 {
		t.Errorf("expected only the x64 and arm64 installers, got %v", resolved.URLs)
	}
	if variant.URLs != nil || variant.Version != "latest" {
		t.Errorf("original variant must stay unchanged: %+v", variant)
	}
}

func TestResolveGitForWindows_Failures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"http error", http.StatusForbidden, `{"message": "API rate limit exceeded"}`},
		{"invalid json", http.StatusOK, `not json`},
		{"no installer assets", http.StatusOK, `{"tag_name": "v2.56.0.windows.2", "assets": []}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newGitForWindowsServer(t, tt.status, tt.body)
			if _, err := resolveGitForWindows(server.URL, registry.VariantSpec{VersionResolver: "git-for-windows"}); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestResolveOrFallback_GitForWindowsPinned verifies that an unreachable
// release API installs the pinned Git for Windows variant (issue #222)
func TestResolveOrFallback_GitForWindowsPinned(t *testing.T) {
	server := newGitForWindowsServer(t, http.StatusServiceUnavailable, "")
	original := gitForWindowsReleaseAPI
	gitForWindowsReleaseAPI = server.URL
	t.Cleanup(func() { gitForWindowsReleaseAPI = original })

	platform := registry.PlatformSpec{Variants: map[string]registry.VariantSpec{
		"installer":        {Type: "exe", Version: "latest", VersionResolver: "git-for-windows", FallbackVariant: "installer-pinned"},
		"installer-pinned": {Type: "exe", Version: "2.56.0.2", URLs: map[string]string{"x64": "https://example.invalid/x64.exe", "arm64": "https://example.invalid/arm64.exe"}},
	}}

	name, spec, err := resolveOrFallback(&platform, "installer", platform.Variants["installer"])
	if err != nil {
		t.Fatalf("expected fallback, got error: %v", err)
	}
	if name != "installer-pinned" || spec.Version != "2.56.0.2" {
		t.Errorf("resolveOrFallback() = %s (%s), want installer-pinned (2.56.0.2)", name, spec.Version)
	}
}
