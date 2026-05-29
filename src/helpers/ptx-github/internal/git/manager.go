/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	httpauth "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Manager wraps go-git operations used by ptx-github.
type Manager struct {
	// Token is sent as HTTP Basic auth password to GitHub when set.
	Token string
	// Username sent with Token. Defaults to "x-access-token" which works for any PAT.
	Username string
	// Progress receives Git progress lines (set to os.Stderr to stream output).
	Progress io.Writer
}

// NewManager returns a Manager. If progress is nil, progress output is suppressed.
func NewManager(token string, progress io.Writer) *Manager {
	return &Manager{Token: token, Username: "x-access-token", Progress: progress}
}

func (m *Manager) auth() *httpauth.BasicAuth {
	if m.Token == "" {
		return nil
	}
	user := m.Username
	if user == "" {
		user = "x-access-token"
	}
	return &httpauth.BasicAuth{Username: user, Password: m.Token}
}

// CloneOptions controls the Clone behaviour.
type CloneOptions struct {
	URL        string
	Dest       string
	Branch     string
	Depth      int
	Bare       bool
	Recurse    bool
	NoCheckout bool
}

// Clone clones a repository to dest.
func (m *Manager) Clone(opts CloneOptions) error {
	if opts.URL == "" {
		return errors.New("clone: URL is required")
	}
	if opts.Dest == "" {
		return errors.New("clone: destination is required")
	}
	if _, err := os.Stat(opts.Dest); err == nil {
		entries, _ := os.ReadDir(opts.Dest)
		if len(entries) > 0 {
			return fmt.Errorf("destination %q already exists and is not empty", opts.Dest)
		}
	}

	gitOpts := &gogit.CloneOptions{
		URL:      opts.URL,
		Progress: m.Progress,
		Auth:     m.auth(),
	}
	if opts.Branch != "" {
		gitOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		gitOpts.SingleBranch = true
	}
	if opts.Depth > 0 {
		gitOpts.Depth = opts.Depth
	}
	if opts.Recurse {
		gitOpts.RecurseSubmodules = gogit.DefaultSubmoduleRecursionDepth
	}
	if opts.NoCheckout {
		gitOpts.NoCheckout = true
	}

	if opts.Bare {
		_, err := gogit.PlainClone(opts.Dest, true, gitOpts)
		return err
	}
	_, err := gogit.PlainClone(opts.Dest, false, gitOpts)
	return err
}

// Checkout switches the working tree of repoDir to ref.
// ref may be a branch, tag, or commit hash.
func (m *Manager) Checkout(repoDir, ref string) error {
	if ref == "" {
		return errors.New("checkout: ref is required")
	}
	repo, err := gogit.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	// Try branch first, then tag, then commit hash.
	if br, err := repo.Reference(plumbing.NewBranchReferenceName(ref), true); err == nil {
		return w.Checkout(&gogit.CheckoutOptions{Branch: br.Name()})
	}
	if tag, err := repo.Reference(plumbing.NewTagReferenceName(ref), true); err == nil {
		return w.Checkout(&gogit.CheckoutOptions{Hash: tag.Hash()})
	}
	hash := plumbing.NewHash(ref)
	if hash.IsZero() {
		return fmt.Errorf("ref %q not found (not a branch, tag, or commit hash)", ref)
	}
	return w.Checkout(&gogit.CheckoutOptions{Hash: hash})
}

// Tags returns the sorted list of tag names from repoDir.
func (m *Manager) Tags(repoDir string) ([]string, error) {
	repo, err := gogit.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	iter, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer iter.Close()

	var tags []string
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref.Name().Short())
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(tags)
	return tags, nil
}

// RemoteTags lists tags from a remote URL without cloning.
func (m *Manager) RemoteTags(url string) ([]string, error) {
	rem := gogit.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	refs, err := rem.List(&gogit.ListOptions{Auth: m.auth()})
	if err != nil {
		return nil, fmt.Errorf("list remote refs: %w", err)
	}
	var tags []string
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tags = append(tags, ref.Name().Short())
		}
	}
	sort.Strings(tags)
	return tags, nil
}

// Status describes the working tree of a repository.
type Status struct {
	Path        string
	Branch      string
	Head        string
	Modified    int
	Untracked   int
	Added       int
	Deleted     int
	Clean       bool
	RemoteURL   string
	Description string
}

// Status returns a summary of the repository at repoDir.
func (m *Manager) Status(repoDir string) (*Status, error) {
	repo, err := gogit.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	st := &Status{Path: repoDir}

	head, err := repo.Head()
	if err == nil {
		st.Head = head.Hash().String()
		if head.Name().IsBranch() {
			st.Branch = head.Name().Short()
		} else {
			st.Branch = "(detached)"
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree: %w", err)
	}
	wstatus, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	st.Clean = wstatus.IsClean()
	for _, fs := range wstatus {
		switch fs.Worktree {
		case gogit.Modified:
			st.Modified++
		case gogit.Untracked:
			st.Untracked++
		case gogit.Added:
			st.Added++
		case gogit.Deleted:
			st.Deleted++
		}
	}

	if remote, err := repo.Remote("origin"); err == nil {
		urls := remote.Config().URLs
		if len(urls) > 0 {
			st.RemoteURL = urls[0]
		}
	}
	return st, nil
}

// FindRepoRoot walks up from start looking for a .git directory.
func FindRepoRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(abs, ".git")); err == nil && info != nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("not a git repository (or any parent)")
		}
		abs = parent
	}
}

// NormalizeRepoSpec converts owner/repo or full URL into an HTTPS clone URL.
func NormalizeRepoSpec(spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", errors.New("empty repo spec")
	}
	if strings.Contains(spec, "://") || strings.HasPrefix(spec, "git@") {
		return spec, nil
	}
	parts := strings.SplitN(spec, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repo spec %q (expected owner/repo)", spec)
	}
	return fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1]), nil
}

// SplitOwnerRepo extracts owner/repo from a spec or URL.
func SplitOwnerRepo(spec string) (owner, repo string, err error) {
	spec = strings.TrimSpace(spec)
	if strings.Contains(spec, "github.com") {
		// strip protocol/host
		idx := strings.Index(spec, "github.com")
		spec = strings.TrimLeft(spec[idx+len("github.com"):], "/:")
	}
	spec = strings.TrimSuffix(spec, ".git")
	parts := strings.Split(spec, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot extract owner/repo from %q", spec)
	}
	return parts[0], parts[1], nil
}
