/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// versionPlaceholder is replaced in a variant's url/urls by the resolved version
const versionPlaceholder = "{version}"

// pythonOrgFTP is the python.org release listing used by the "python.org" resolver
var pythonOrgFTP = "https://www.python.org/ftp/python/"

// gitForWindowsReleaseAPI is the GitHub API endpoint of the newest Git for
// Windows release used by the "git-for-windows" resolver
var gitForWindowsReleaseAPI = "https://api.github.com/repos/git-for-windows/git/releases/latest"

// gitForWindowsAssetPatterns match the installer asset of each architecture
// (e.g. Git-2.56.0.2-64-bit.exe); the captured group is the asset version
var gitForWindowsAssetPatterns = map[string]*regexp.Regexp{
	"x64":   regexp.MustCompile(`^Git-(\d+(?:\.\d+)+)-64-bit\.exe$`),
	"arm64": regexp.MustCompile(`^Git-(\d+(?:\.\d+)+)-arm64\.exe$`),
}

// maxResolverCandidates limits how many of the newest versions are probed
const maxResolverCandidates = 5

var resolverHTTPClient = &http.Client{Timeout: 30 * time.Second}

// releaseDirPattern matches a final release directory (e.g. "3.14.8/") in the
// python.org listing; pre-release names never form a plain X.Y.Z directory
var releaseDirPattern = regexp.MustCompile(`href="(\d+)\.(\d+)\.(\d+)/"`)

// resolveVariantVersion fills in the version of a variant that declares a
// versionResolver: it finds the newest version whose download for the current
// architecture exists and substitutes it into url/urls (issue #202)
func resolveVariantVersion(variant *registry.VariantSpec) error {
	if variant.VersionResolver == "" {
		return nil
	}

	switch variant.VersionResolver {
	case "python.org":
		version, err := resolvePythonOrgVersion(pythonOrgFTP, func(v string) string {
			return selectArchURL(substituteVersion(variant, v))
		})
		if err != nil {
			return err
		}
		resolved := substituteVersion(variant, version)
		*variant = resolved
		return nil
	case "git-for-windows":
		resolved, err := resolveGitForWindows(gitForWindowsReleaseAPI, *variant)
		if err != nil {
			return err
		}
		*variant = resolved
		return nil
	default:
		return fmt.Errorf("unknown version resolver: %s", variant.VersionResolver)
	}
}

// resolveOrFallback resolves the version of a variant with a versionResolver.
// When resolving fails and the variant names a fallbackVariant, that pinned
// variant is returned instead, so an unreachable release listing does not
// block the install
func resolveOrFallback(platform *registry.PlatformSpec, name string, variant registry.VariantSpec) (string, registry.VariantSpec, error) {
	err := resolveVariantVersion(&variant)
	if err == nil {
		return name, variant, nil
	}

	fallbackName := variant.FallbackVariant
	fallback, ok := platform.Variants[fallbackName]
	if fallbackName == "" || !ok || fallback.VersionResolver != "" {
		return name, variant, fmt.Errorf("failed to resolve latest version: %w", err)
	}

	fmt.Printf("⚠️  Could not resolve the latest version: %v\n", err)
	fmt.Printf("   Falling back to variant '%s' (version %s)\n", fallbackName, fallback.Version)
	return fallbackName, fallback, nil
}

// substituteVersion returns a copy of variant with {version} replaced in its
// download URLs and Version set to the resolved version
func substituteVersion(variant *registry.VariantSpec, version string) registry.VariantSpec {
	resolved := *variant
	resolved.Version = version
	resolved.URL = strings.ReplaceAll(variant.URL, versionPlaceholder, version)
	if len(variant.URLs) > 0 {
		resolved.URLs = make(map[string]string, len(variant.URLs))
		for arch, url := range variant.URLs {
			resolved.URLs[arch] = strings.ReplaceAll(url, versionPlaceholder, version)
		}
	}
	return resolved
}

// selectArchURL returns the variant's download URL for the current
// architecture, using the same x64/x86 mapping as installWindowsBinary
func selectArchURL(variant registry.VariantSpec) string {
	if variant.URL != "" {
		return variant.URL
	}
	return selectArchValue(variant.URLs)
}

// selectArchChecksum returns the variant's checksum for the current
// architecture; a single-URL variant may key its checksum by "default"
func selectArchChecksum(variant registry.VariantSpec) string {
	if sum := selectArchValue(variant.Checksum); sum != "" {
		return sum
	}
	return variant.Checksum["default"]
}

// selectArchValue looks up the entry of an architecture keyed map for the
// current architecture, trying the Windows name (x64/x86) before the Go name
func selectArchValue(values map[string]string) string {
	arch := GetArchitecture()
	archKey := arch
	switch arch {
	case "amd64":
		archKey = "x64"
	case "386":
		archKey = "x86"
	}
	if value, ok := values[archKey]; ok {
		return value
	}
	return values[arch]
}

// resolvePythonOrgVersion reads the python.org release listing and returns the
// newest version for which downloadURL(version) exists. Directories of
// upcoming releases are listed before their installers are published, so
// each candidate is probed
func resolvePythonOrgVersion(listingURL string, downloadURL func(version string) string) (string, error) {
	fmt.Printf("🔎 Resolving latest version from %s ...\n", listingURL)

	resp, err := resolverHTTPClient.Get(listingURL)
	if err != nil {
		return "", fmt.Errorf("failed to read release listing: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to read release listing: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read release listing: %w", err)
	}

	versions := parseReleaseVersions(string(body))
	if len(versions) == 0 {
		return "", fmt.Errorf("no release versions found in %s", listingURL)
	}

	for idx, version := range versions {
		if idx >= maxResolverCandidates {
			break
		}
		url := downloadURL(version)
		if url == "" {
			return "", fmt.Errorf("no download URL for the current architecture")
		}
		if urlExists(url) {
			fmt.Printf("✅ Latest version: %s\n", version)
			return version, nil
		}
		fmt.Printf("   ⏭️  %s has no installer yet, skipping\n", version)
	}
	return "", fmt.Errorf("no installer found among the %d newest versions", maxResolverCandidates)
}

// githubRelease is the subset of the GitHub release API response the
// resolvers read
type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

// resolveGitForWindows reads the newest Git for Windows release from the
// GitHub API and returns a copy of variant with the installer URL and sha256
// checksum of every architecture filled in (issue #222). Asset names carry
// the release number (Git-2.56.0.2-64-bit.exe for tag v2.56.0.windows.2), so
// URLs are taken from the release instead of a {version} template
func resolveGitForWindows(apiURL string, variant registry.VariantSpec) (registry.VariantSpec, error) {
	fmt.Printf("🔎 Resolving latest Git for Windows release from %s ...\n", apiURL)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return variant, fmt.Errorf("failed to build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "portunix-installer")

	resp, err := resolverHTTPClient.Do(req)
	if err != nil {
		return variant, fmt.Errorf("failed to read release info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return variant, fmt.Errorf("failed to read release info: HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return variant, fmt.Errorf("failed to parse release info: %w", err)
	}

	resolved := variant
	resolved.URLs = make(map[string]string, len(gitForWindowsAssetPatterns))
	resolved.Checksum = make(map[string]string, len(gitForWindowsAssetPatterns))
	for _, asset := range release.Assets {
		for arch, pattern := range gitForWindowsAssetPatterns {
			m := pattern.FindStringSubmatch(asset.Name)
			if m == nil {
				continue
			}
			resolved.URLs[arch] = asset.BrowserDownloadURL
			if asset.Digest != "" {
				resolved.Checksum[arch] = asset.Digest
			}
			// All installers of one release share the asset version
			resolved.Version = m[1]
		}
	}

	if selectArchURL(resolved) == "" {
		return variant, fmt.Errorf("release %s has no installer for architecture %s", release.TagName, GetArchitecture())
	}
	fmt.Printf("✅ Latest version: %s (%s)\n", resolved.Version, release.TagName)
	return resolved, nil
}

// parseReleaseVersions extracts X.Y.Z release directories, newest first
func parseReleaseVersions(listing string) []string {
	type semver struct {
		text                string
		major, minor, patch int
	}
	seen := make(map[string]bool)
	var found []semver
	for _, m := range releaseDirPattern.FindAllStringSubmatch(listing, -1) {
		text := m[1] + "." + m[2] + "." + m[3]
		if seen[text] {
			continue
		}
		seen[text] = true
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		patch, _ := strconv.Atoi(m[3])
		found = append(found, semver{text, major, minor, patch})
	}
	sort.Slice(found, func(a, b int) bool {
		if found[a].major != found[b].major {
			return found[a].major > found[b].major
		}
		if found[a].minor != found[b].minor {
			return found[a].minor > found[b].minor
		}
		return found[a].patch > found[b].patch
	})

	versions := make([]string, len(found))
	for idx, v := range found {
		versions[idx] = v.text
	}
	return versions
}

// urlExists reports whether a HEAD request for url answers 200
func urlExists(url string) bool {
	resp, err := resolverHTTPClient.Head(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
