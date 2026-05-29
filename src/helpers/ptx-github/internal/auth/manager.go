/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// EnvTokenVars are environment variables consulted (in order) for a token
// when no explicit account is selected.
var EnvTokenVars = []string{"PORTUNIX_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"}

// TokenSource describes where a token was obtained from.
type TokenSource int

const (
	SourceNone TokenSource = iota
	SourceEnv
	SourceStore
)

func (s TokenSource) String() string {
	switch s {
	case SourceEnv:
		return "environment"
	case SourceStore:
		return "store"
	default:
		return "none"
	}
}

// ResolvedToken holds a token and metadata about its origin.
type ResolvedToken struct {
	Token       string
	Source      TokenSource
	EnvVarName  string
	AccountName string
	Username    string
}

// Manager coordinates token resolution between environment variables and
// the encrypted on-disk store.
type Manager struct {
	store *Store
}

// NewManager creates a Manager backed by the given store. Store may be nil
// for environment-only lookups.
func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

// Resolve returns the best token to use. If account is non-empty it must
// exist in the store. Otherwise environment variables win, falling back
// to the store's default account.
func (m *Manager) Resolve(account string) (*ResolvedToken, error) {
	if account != "" {
		if m.store == nil {
			return nil, fmt.Errorf("no store configured")
		}
		token, acc, err := m.store.GetToken(account)
		if err != nil {
			return nil, err
		}
		return &ResolvedToken{
			Token: token, Source: SourceStore,
			AccountName: acc.Name, Username: acc.Username,
		}, nil
	}

	for _, name := range EnvTokenVars {
		if v := os.Getenv(name); v != "" {
			return &ResolvedToken{Token: v, Source: SourceEnv, EnvVarName: name}, nil
		}
	}

	if m.store != nil {
		token, acc, err := m.store.GetToken("")
		if err == nil {
			return &ResolvedToken{
				Token: token, Source: SourceStore,
				AccountName: acc.Name, Username: acc.Username,
			}, nil
		}
	}

	return &ResolvedToken{Source: SourceNone}, nil
}

// NewGitHubClient builds an authenticated *github.Client using the resolved
// token, or an unauthenticated client if no token is available.
func (m *Manager) NewGitHubClient(ctx context.Context, account string) (*github.Client, *ResolvedToken, error) {
	r, err := m.Resolve(account)
	if err != nil {
		return nil, nil, err
	}
	if r.Token == "" {
		return github.NewClient(nil), r, nil
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: r.Token})
	return github.NewClient(oauth2.NewClient(ctx, ts)), r, nil
}

// Validate calls the GitHub /user endpoint to verify credentials.
// Returns the authenticated username on success.
func (m *Manager) Validate(ctx context.Context, account string) (username string, scopes string, err error) {
	client, resolved, err := m.NewGitHubClient(ctx, account)
	if err != nil {
		return "", "", err
	}
	if resolved.Token == "" {
		return "", "", fmt.Errorf("no token available to validate")
	}
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", "", fmt.Errorf("validate token: %w", err)
	}
	scopeHeader := ""
	if resp != nil && resp.Response != nil {
		scopeHeader = resp.Response.Header.Get("X-OAuth-Scopes")
	}
	return user.GetLogin(), scopeHeader, nil
}
