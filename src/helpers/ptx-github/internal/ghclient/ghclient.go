/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// Package ghclient provides a thin wrapper around go-github used by the
// ptx-github CLI commands. It owns the *github.Client lifetime and adds a
// few helpers (find asset by name/platform, list releases, latest release).
package ghclient

import (
	"context"
	"fmt"

	"github.com/google/go-github/v56/github"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/auth"
)

// Client is a high-level GitHub client used by ptx-github commands.
type Client struct {
	gh   *github.Client
	auth *auth.ResolvedToken
	ctx  context.Context
}

// New builds an authenticated client using the auth.Manager.
// account selects a specific stored account; empty string uses defaults.
func New(ctx context.Context, mgr *auth.Manager, account string) (*Client, error) {
	gh, resolved, err := mgr.NewGitHubClient(ctx, account)
	if err != nil {
		return nil, err
	}
	return &Client{gh: gh, auth: resolved, ctx: ctx}, nil
}

// Token returns the active token (may be empty for unauthenticated clients).
func (c *Client) Token() string {
	if c.auth == nil {
		return ""
	}
	return c.auth.Token
}

// AuthInfo returns the resolved-token metadata.
func (c *Client) AuthInfo() *auth.ResolvedToken {
	return c.auth
}

// Repository fetches basic metadata for owner/repo.
func (c *Client) Repository(owner, repo string) (*github.Repository, error) {
	r, _, err := c.gh.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", owner, repo, err)
	}
	return r, nil
}

// ListReleases returns all releases for owner/repo.
func (c *Client) ListReleases(owner, repo string) ([]*github.RepositoryRelease, error) {
	rels, _, err := c.gh.Repositories.ListReleases(c.ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("list releases for %s/%s: %w", owner, repo, err)
	}
	return rels, nil
}

// LatestRelease returns the latest non-prerelease release.
func (c *Client) LatestRelease(owner, repo string) (*github.RepositoryRelease, error) {
	rel, _, err := c.gh.Repositories.GetLatestRelease(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("latest release for %s/%s: %w", owner, repo, err)
	}
	return rel, nil
}

// ReleaseByTag returns a specific release by tag name.
func (c *Client) ReleaseByTag(owner, repo, tag string) (*github.RepositoryRelease, error) {
	rel, _, err := c.gh.Repositories.GetReleaseByTag(c.ctx, owner, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("release %s for %s/%s: %w", tag, owner, repo, err)
	}
	return rel, nil
}

// RateLimit returns the current API rate limit.
func (c *Client) RateLimit() (*github.RateLimits, error) {
	rl, _, err := c.gh.RateLimits(c.ctx)
	if err != nil {
		return nil, err
	}
	return rl, nil
}
