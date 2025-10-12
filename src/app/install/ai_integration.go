package install

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIPackageManager handles AI-assisted package management
type AIPackageManager struct {
	registry *PackageRegistry
	client   *http.Client
}

// NewAIPackageManager creates a new AI package manager
func NewAIPackageManager(registry *PackageRegistry) *AIPackageManager {
	return &AIPackageManager{
		registry: registry,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VersionDiscoveryResult represents the result of version discovery
type VersionDiscoveryResult struct {
	PackageName     string            `json:"packageName"`
	CurrentVersion  string            `json:"currentVersion"`
	LatestVersion   string            `json:"latestVersion"`
	UpdateAvailable bool              `json:"updateAvailable"`
	ReleaseDate     time.Time         `json:"releaseDate,omitempty"`
	DownloadURLs    map[string]string `json:"downloadUrls,omitempty"`
	ChangelogURL    string            `json:"changelogUrl,omitempty"`
	ReleaseNotes    string            `json:"releaseNotes,omitempty"`
	Error           string            `json:"error,omitempty"`
}

// PackageUpdateContext contains context for AI-guided updates
type PackageUpdateContext struct {
	Package         *Package                `json:"package"`
	CurrentVersions map[string]string       `json:"currentVersions"` // variant -> version
	Sources         map[string]SourceSpec   `json:"sources"`
	AIPrompts       *AIPrompts              `json:"aiPrompts"`
	LastUpdate      time.Time               `json:"lastUpdate"`
}

// DiscoverLatestVersions uses AI prompts to discover latest versions for a package
func (ai *AIPackageManager) DiscoverLatestVersions(packageName string) (*VersionDiscoveryResult, error) {
	pkg, err := ai.registry.GetPackage(packageName)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}

	if pkg.Spec.AIPrompts == nil {
		return &VersionDiscoveryResult{
			PackageName: packageName,
			Error:       "no AI prompts configured for this package",
		}, nil
	}

	// Create update context
	context := &PackageUpdateContext{
		Package:         pkg,
		CurrentVersions: ai.extractCurrentVersions(pkg),
		Sources:         pkg.Spec.Sources,
		AIPrompts:       pkg.Spec.AIPrompts,
		LastUpdate:      time.Now(),
	}

	// Use different strategies based on package sources
	if len(pkg.Spec.Sources) > 0 {
		return ai.discoverVersionsFromSources(context)
	}

	return ai.discoverVersionsFromAIPrompts(context)
}

// extractCurrentVersions extracts current versions from all variants
func (ai *AIPackageManager) extractCurrentVersions(pkg *Package) map[string]string {
	versions := make(map[string]string)

	for platformName, platform := range pkg.Spec.Platforms {
		for variantName, variant := range platform.Variants {
			key := fmt.Sprintf("%s:%s", platformName, variantName)
			versions[key] = variant.Version
		}
	}

	return versions
}

// discoverVersionsFromSources discovers versions using configured sources
func (ai *AIPackageManager) discoverVersionsFromSources(context *PackageUpdateContext) (*VersionDiscoveryResult, error) {
	result := &VersionDiscoveryResult{
		PackageName:    context.Package.Metadata.Name,
		CurrentVersion: ai.findMostRecentVersion(context.CurrentVersions),
	}

	// Try each source until we find version information
	for _, source := range context.Sources {
		switch source.Type {
		case "github":
			if latestVersion, err := ai.discoverGitHubLatestVersion(source); err == nil {
				result.LatestVersion = latestVersion
				result.UpdateAvailable = ai.isUpdateAvailable(result.CurrentVersion, latestVersion)
				break
			}
		case "direct":
			if latestVersion, err := ai.discoverDirectLatestVersion(source, context.AIPrompts); err == nil {
				result.LatestVersion = latestVersion
				result.UpdateAvailable = ai.isUpdateAvailable(result.CurrentVersion, latestVersion)
				break
			}
		}
	}

	if result.LatestVersion == "" {
		result.Error = fmt.Sprintf("could not discover latest version from any source")
	}

	return result, nil
}

// discoverVersionsFromAIPrompts discovers versions using only AI prompts
func (ai *AIPackageManager) discoverVersionsFromAIPrompts(context *PackageUpdateContext) (*VersionDiscoveryResult, error) {
	result := &VersionDiscoveryResult{
		PackageName:    context.Package.Metadata.Name,
		CurrentVersion: ai.findMostRecentVersion(context.CurrentVersions),
		Error:          "AI prompt-based discovery not yet implemented (requires external AI service)",
	}

	// For now, return a placeholder result
	// In a full implementation, this would call an AI service with the prompts
	return result, nil
}

// discoverGitHubLatestVersion discovers latest version from GitHub releases
func (ai *AIPackageManager) discoverGitHubLatestVersion(source SourceSpec) (string, error) {
	if source.APIEndpoint == "" {
		return "", fmt.Errorf("no GitHub API endpoint configured")
	}

	resp, err := ai.client.Get(source.APIEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to fetch GitHub release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse GitHub release: %w", err)
	}

	// Clean up the version tag (remove 'v' prefix if present)
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// discoverDirectLatestVersion discovers version from direct sources
func (ai *AIPackageManager) discoverDirectLatestVersion(source SourceSpec, prompts *AIPrompts) (string, error) {
	// This would implement direct URL parsing based on AI prompts
	// For now, return an error to indicate it's not implemented
	return "", fmt.Errorf("direct version discovery not yet implemented")
}

// findMostRecentVersion finds the most recent version from a map of versions
func (ai *AIPackageManager) findMostRecentVersion(versions map[string]string) string {
	mostRecent := ""
	for _, version := range versions {
		if mostRecent == "" || ai.isVersionNewer(version, mostRecent) {
			mostRecent = version
		}
	}
	return mostRecent
}

// isUpdateAvailable checks if an update is available
func (ai *AIPackageManager) isUpdateAvailable(current, latest string) bool {
	if current == "" || latest == "" {
		return false
	}
	return current != latest && ai.isVersionNewer(latest, current)
}

// isVersionNewer compares two version strings (basic implementation)
func (ai *AIPackageManager) isVersionNewer(v1, v2 string) bool {
	// This is a simplified version comparison
	// A full implementation would use proper semantic versioning
	return strings.Compare(v1, v2) > 0
}

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a GitHub release asset
type GitHubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// CheckAllPackagesForUpdates checks all packages for available updates
func (ai *AIPackageManager) CheckAllPackagesForUpdates() ([]*VersionDiscoveryResult, error) {
	allPackages := ai.registry.GetAllPackages()
	results := make([]*VersionDiscoveryResult, 0, len(allPackages))

	for packageName := range allPackages {
		result, err := ai.DiscoverLatestVersions(packageName)
		if err != nil {
			result = &VersionDiscoveryResult{
				PackageName: packageName,
				Error:       err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateUpdateReport generates a human-readable update report
func (ai *AIPackageManager) GenerateUpdateReport() (string, error) {
	results, err := ai.CheckAllPackagesForUpdates()
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}

	var report strings.Builder
	report.WriteString("📦 PORTUNIX PACKAGE UPDATE REPORT\n")
	report.WriteString("=====================================\n\n")

	updatesAvailable := 0
	errorsEncountered := 0

	for _, result := range results {
		if result.Error != "" {
			errorsEncountered++
			report.WriteString(fmt.Sprintf("❌ %s: %s\n", result.PackageName, result.Error))
		} else if result.UpdateAvailable {
			updatesAvailable++
			report.WriteString(fmt.Sprintf("🔄 %s: %s → %s\n",
				result.PackageName, result.CurrentVersion, result.LatestVersion))
		} else {
			report.WriteString(fmt.Sprintf("✅ %s: %s (up to date)\n",
				result.PackageName, result.CurrentVersion))
		}
	}

	report.WriteString(fmt.Sprintf("\n📊 SUMMARY:\n"))
	report.WriteString(fmt.Sprintf("   Packages checked: %d\n", len(results)))
	report.WriteString(fmt.Sprintf("   Updates available: %d\n", updatesAvailable))
	report.WriteString(fmt.Sprintf("   Errors: %d\n", errorsEncountered))

	if updatesAvailable > 0 {
		report.WriteString(fmt.Sprintf("\n💡 Run 'portunix registry update' to apply available updates\n"))
	}

	return report.String(), nil
}