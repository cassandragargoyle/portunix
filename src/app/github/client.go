package github

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// GitHubClient provides basic GitHub API functionality for plugin system
type GitHubClient struct {
	client *github.Client
	ctx    context.Context
}

// Repository represents a GitHub repository with essential information
type Repository struct {
	Owner       string
	Name        string
	Description string
	Stars       int
	Forks       int
	UpdatedAt   string
}

// Release represents a GitHub release
type Release struct {
	TagName     string
	Name        string
	Body        string
	PublishedAt string
	Assets      []*Asset
}

// Asset represents a downloadable asset from a release
type Asset struct {
	Name        string
	Size        int64
	DownloadURL string
	ContentType string
}

// NewClient creates a new GitHub client
// Token is optional - if not provided, client works with rate limits
func NewClient() *GitHubClient {
	ctx := context.Background()
	var client *github.Client

	// Try to get token from environment variable first
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		// Use client without authentication (with rate limits)
		client = github.NewClient(nil)
	}

	return &GitHubClient{
		client: client,
		ctx:    ctx,
	}
}

// NewClientForPortunixPlugins creates a GitHub client specifically for cassandragargoyle/portunix-plugins
// Requires PORTUNIX_PLUGINS_GITHUB_TOKEN environment variable for authenticated access
func NewClientForPortunixPlugins() *GitHubClient {
	ctx := context.Background()

	// Try dedicated environment variable first, then fall back to generic GITHUB_TOKEN
	token := os.Getenv("PORTUNIX_PLUGINS_GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		// Fallback to unauthenticated client (with rate limits)
		client = github.NewClient(nil)
	}

	return &GitHubClient{
		client: client,
		ctx:    ctx,
	}
}

// GetRepository retrieves basic repository information
func (g *GitHubClient) GetRepository(owner, repo string) (*Repository, error) {
	githubRepo, _, err := g.client.Repositories.Get(g.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}

	return &Repository{
		Owner:       owner,
		Name:        repo,
		Description: githubRepo.GetDescription(),
		Stars:       githubRepo.GetStargazersCount(),
		Forks:       githubRepo.GetForksCount(),
		UpdatedAt:   githubRepo.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}

// ListReleases retrieves all releases for a repository
func (g *GitHubClient) ListReleases(owner, repo string) ([]*Release, error) {
	githubReleases, _, err := g.client.Repositories.ListReleases(g.ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for %s/%s: %w", owner, repo, err)
	}

	var releases []*Release
	for _, githubRelease := range githubReleases {
		var assets []*Asset
		for _, githubAsset := range githubRelease.Assets {
			assets = append(assets, &Asset{
				Name:        githubAsset.GetName(),
				Size:        int64(githubAsset.GetSize()),
				DownloadURL: githubAsset.GetBrowserDownloadURL(),
				ContentType: githubAsset.GetContentType(),
			})
		}

		releases = append(releases, &Release{
			TagName:     githubRelease.GetTagName(),
			Name:        githubRelease.GetName(),
			Body:        githubRelease.GetBody(),
			PublishedAt: githubRelease.GetPublishedAt().Format("2006-01-02T15:04:05Z"),
			Assets:      assets,
		})
	}

	return releases, nil
}

// GetRelease retrieves a specific release by tag
func (g *GitHubClient) GetRelease(owner, repo, tag string) (*Release, error) {
	githubRelease, _, err := g.client.Repositories.GetReleaseByTag(g.ctx, owner, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get release %s for %s/%s: %w", tag, owner, repo, err)
	}

	var assets []*Asset
	for _, githubAsset := range githubRelease.Assets {
		assets = append(assets, &Asset{
			Name:        githubAsset.GetName(),
			Size:        int64(githubAsset.GetSize()),
			DownloadURL: githubAsset.GetBrowserDownloadURL(),
			ContentType: githubAsset.GetContentType(),
		})
	}

	return &Release{
		TagName:     githubRelease.GetTagName(),
		Name:        githubRelease.GetName(),
		Body:        githubRelease.GetBody(),
		PublishedAt: githubRelease.GetPublishedAt().Format("2006-01-02T15:04:05Z"),
		Assets:      assets,
	}, nil
}

// FindAsset finds a specific asset in a release by name pattern
func (r *Release) FindAsset(pattern string) *Asset {
	for _, asset := range r.Assets {
		if asset.Name == pattern {
			return asset
		}
	}
	return nil
}

// FindAssetByPlatform finds asset for current platform
func (r *Release) FindAssetByPlatform(platform string) *Asset {
	for _, asset := range r.Assets {
		// Simple pattern matching for platform-specific assets
		switch platform {
		case "linux-amd64":
			if contains(asset.Name, "linux") && contains(asset.Name, "amd64") {
				return asset
			}
		case "windows-amd64":
			if contains(asset.Name, "windows") && contains(asset.Name, "amd64") {
				return asset
			}
		case "darwin-amd64":
			if contains(asset.Name, "darwin") && contains(asset.Name, "amd64") {
				return asset
			}
		}
	}
	return nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
