package github

import (
	"fmt"
)

// Service provides GitHub integration functionality for other Portunix components
type Service struct {
	client          *GitHubClient
	downloadManager *DownloadManager
}

// NewService creates a new GitHub service
func NewService() *Service {
	return &Service{
		client:          NewClient(),
		downloadManager: NewDownloadManager(),
	}
}

// NewServiceForPortunixPlugins creates a GitHub service with embedded token for portunix-plugins
func NewServiceForPortunixPlugins() *Service {
	return &Service{
		client:          NewClientForPortunixPlugins(),
		downloadManager: NewDownloadManager(),
	}
}

// GetClient returns the GitHub API client
func (s *Service) GetClient() *GitHubClient {
	return s.client
}

// GetDownloadManager returns the download manager
func (s *Service) GetDownloadManager() *DownloadManager {
	return s.downloadManager
}

// DownloadReleaseAsset downloads a specific asset from a repository release
func (s *Service) DownloadReleaseAsset(owner, repo, tag, assetName, dest string) error {
	// Get the release
	release, err := s.client.GetRelease(owner, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Find the asset
	asset := release.FindAsset(assetName)
	if asset == nil {
		return fmt.Errorf("asset %s not found in release %s", assetName, tag)
	}

	// Download the asset
	return s.downloadManager.DownloadAsset(asset, dest)
}

// DownloadReleaseAssetWithProgress downloads with progress reporting
func (s *Service) DownloadReleaseAssetWithProgress(owner, repo, tag, assetName, dest string, progress ProgressCallback) error {
	// Get the release
	release, err := s.client.GetRelease(owner, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Find the asset
	asset := release.FindAsset(assetName)
	if asset == nil {
		return fmt.Errorf("asset %s not found in release %s", assetName, tag)
	}

	// Download the asset with progress
	return s.downloadManager.DownloadAssetWithProgress(asset, dest, progress)
}

// ListRepositoryReleases lists all releases for a repository
func (s *Service) ListRepositoryReleases(owner, repo string) ([]*Release, error) {
	return s.client.ListReleases(owner, repo)
}

// GetRepositoryInfo gets basic repository information
func (s *Service) GetRepositoryInfo(owner, repo string) (*Repository, error) {
	return s.client.GetRepository(owner, repo)
}

// FindPlatformAsset finds the appropriate asset for the given platform
func (s *Service) FindPlatformAsset(owner, repo, tag, platform string) (*Asset, error) {
	release, err := s.client.GetRelease(owner, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	asset := release.FindAssetByPlatform(platform)
	if asset == nil {
		return nil, fmt.Errorf("no asset found for platform %s in release %s", platform, tag)
	}

	return asset, nil
}

// ValidateRepository checks if a repository exists and is accessible
func (s *Service) ValidateRepository(owner, repo string) error {
	_, err := s.client.GetRepository(owner, repo)
	return err
}

// GetLatestRelease gets the most recent release for a repository
func (s *Service) GetLatestRelease(owner, repo string) (*Release, error) {
	releases, err := s.client.ListReleases(owner, repo)
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for repository %s/%s", owner, repo)
	}

	// First release in list is the latest
	return releases[0], nil
}