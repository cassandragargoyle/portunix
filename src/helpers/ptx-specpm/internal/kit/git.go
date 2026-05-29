/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package kit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GitProvider fetches the kit by shallow-cloning the upstream repository
// at a pinned ref, persisting the result under ~/.cache/portunix/specpm/<ref>/.
// Subsequent calls with the same ref hit the cache and never invoke git.
type GitProvider struct {
	URL string
	Ref string
}

// NewGitProvider constructs a Provider that uses git for kit retrieval.
func NewGitProvider(url, ref string) *GitProvider {
	return &GitProvider{URL: url, Ref: ref}
}

// Fetch returns the cache directory holding the kit, cloning if necessary.
// fromCache is true when the cache was already populated for the requested ref.
func (g *GitProvider) Fetch() (string, bool, error) {
	if g.URL == "" {
		return "", false, errors.New("git provider: empty upstream URL")
	}
	if g.Ref == "" {
		return "", false, errors.New("git provider: empty ref")
	}
	dir, err := cacheDirFor(g.Ref)
	if err != nil {
		return "", false, err
	}
	if isPopulatedKit(dir) {
		return dir, true, nil
	}
	if err := requireGit(); err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", false, fmt.Errorf("create cache parent: %w", err)
	}
	// Clone into a sibling .tmp directory and rename on success so a partial
	// failure cannot leave the cache half-populated.
	tmpDir := dir + ".tmp"
	_ = os.RemoveAll(tmpDir)
	args := []string{"clone", "--depth", "1", "--branch", g.Ref, g.URL, tmpDir}
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmpDir)
		// ADR-041 D4: locked remediation hint. Do not paraphrase.
		return "", false, fmt.Errorf(
			"--source git requires a local 'git' binary and network access on first run.\n"+
				"       Cache miss for ref %q at %s.\n"+
				"       For air-gapped operation: see \"Air-gapped workflow\" in\n"+
				"       docs/helpers/ptx-specpm.md, or pass --source <path> to a vendored mirror.\n"+
				"       (underlying error: %v)",
			g.Ref, dir, err)
	}
	// Drop .git from the cache to keep it lean — we never push from the cache.
	_ = os.RemoveAll(filepath.Join(tmpDir, ".git"))
	if err := os.Rename(tmpDir, dir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", false, fmt.Errorf("promote cache dir: %w", err)
	}
	return dir, false, nil
}

// CacheRoot returns the per-user cache root used by the git provider.
func CacheRoot() string {
	root, err := userCacheRoot()
	if err != nil {
		return ""
	}
	return root
}

func cacheDirFor(ref string) (string, error) {
	root, err := userCacheRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, sanitizeRefForFS(ref)), nil
}

func userCacheRoot() (string, error) {
	if env := os.Getenv("PORTUNIX_SPECPM_CACHE"); env != "" {
		return env, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	return filepath.Join(base, "portunix", "specpm"), nil
}

// isPopulatedKit performs a shallow check that the cache directory looks
// like a real spec-kit-pm checkout (templates + agents + workflows present).
func isPopulatedKit(dir string) bool {
	for _, want := range []string{"templates", "agents", "workflows"} {
		info, err := os.Stat(filepath.Join(dir, want))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

func requireGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		// ADR-041 D4: same locked remediation framing as the clone-failure path.
		hint := "install 'git' (e.g. 'portunix install git')"
		if runtime.GOOS == "windows" {
			hint = "install 'git' (Git for Windows)"
		}
		return fmt.Errorf(
			"--source git requires a local 'git' binary and network access on first run.\n"+
				"       'git' was not found on PATH.\n"+
				"       Either %s, or pass --source <path> to a vendored mirror\n"+
				"       (see \"Air-gapped workflow\" in docs/helpers/ptx-specpm.md).",
			hint)
	}
	return nil
}

// sanitizeRefForFS turns refs like "main" or "v1.2.3" into a directory-safe
// segment. Refs are validated upstream by git; this is a defensive pass.
func sanitizeRefForFS(ref string) string {
	out := make([]rune, 0, len(ref))
	for _, r := range ref {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '-', r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "default"
	}
	return string(out)
}
