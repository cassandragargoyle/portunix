/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// cacheMetaFile records the resolved commit of a cached clone. Its presence
// marks the cache entry as complete (the clone is promoted atomically).
const cacheMetaFile = ".portunix-source.json"

// SourceKind distinguishes git repositories from local directories
type SourceKind int

const (
	SourceGit SourceKind = iota
	SourceLocal
)

// Source describes where skills are read from (ADR-043 D2)
type Source struct {
	Kind SourceKind
	// Location is the normalized git URL or the absolute local path. It is
	// the value recorded in the provenance manifest.
	Location string
	// Ref is the requested branch, tag or commit; empty means the default branch
	Ref string
}

// Materialized is a source checked out on disk
type Materialized struct {
	Dir       string
	Commit    string
	FromCache bool
}

// shorthandPattern matches the GitHub "owner/repo" shorthand
var shorthandPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// shaPattern matches an abbreviated or full commit SHA
var shaPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

// scpLikePattern matches git@host:owner/repo(.git)
var scpLikePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+@([A-Za-z0-9_.-]+):(.+)$`)

// ParseSource interprets the argument of `skills add`. An existing directory
// is a local source, anything else is treated as a git URL ("owner/repo" is
// expanded to GitHub).
func ParseSource(arg, ref string) (Source, error) {
	if arg == "" {
		return Source{}, errors.New("missing skill source (git URL or local path)")
	}
	if info, err := os.Stat(arg); err == nil && info.IsDir() {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return Source{}, fmt.Errorf("resolve source path: %w", err)
		}
		return Source{Kind: SourceLocal, Location: abs, Ref: ref}, nil
	}
	if shorthandPattern.MatchString(arg) && !strings.HasPrefix(arg, ".") {
		arg = "https://github.com/" + arg
	}
	if _, _, _, err := repoKey(arg); err != nil {
		return Source{}, fmt.Errorf("invalid skill source %q: not an existing directory nor a git URL", arg)
	}
	return Source{Kind: SourceGit, Location: normalizeGitURL(arg), Ref: ref}, nil
}

// normalizeGitURL strips trailing slashes and adds the .git suffix to
// http(s) URLs so equivalent spellings map to the same manifest source
func normalizeGitURL(raw string) string {
	u := strings.TrimRight(raw, "/")
	if (strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")) && !strings.HasSuffix(u, ".git") {
		u += ".git"
	}
	return u
}

// repoKey extracts host, owner and repo from a git URL for the cache layout
func repoKey(raw string) (host, owner, repo string, err error) {
	var path string
	if m := scpLikePattern.FindStringSubmatch(raw); m != nil {
		host, path = m[1], m[2]
	} else {
		u, perr := url.Parse(raw)
		if perr != nil || u.Host == "" || u.Scheme == "" {
			return "", "", "", errors.New("unsupported git URL")
		}
		host, path = u.Hostname(), u.Path
	}
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
		return "", "", "", errors.New("git URL must contain owner and repository")
	}
	owner = strings.Join(parts[:len(parts)-1], "_")
	repo = parts[len(parts)-1]
	return sanitizeSegment(host), sanitizeSegment(owner), sanitizeSegment(repo), nil
}

// Materialize makes the source available on disk. Git sources are served
// from the cache unless refresh is set (used by `update`).
func Materialize(src Source, refresh bool) (*Materialized, error) {
	if src.Kind == SourceLocal {
		return &Materialized{Dir: src.Location, Commit: localCommit(src.Location)}, nil
	}
	dir, err := CacheDir(src)
	if err != nil {
		return nil, err
	}
	if !refresh {
		if commit, ok := readCacheMeta(dir); ok {
			return &Materialized{Dir: dir, Commit: commit, FromCache: true}, nil
		}
	}
	if err := requireGit(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil { // #nosec G301 -- cache tree uses regular directory permissions
		return nil, fmt.Errorf("create cache parent: %w", err)
	}
	// Clone into a sibling .tmp directory and swap on success so a failed
	// clone never leaves a half-populated cache entry
	tmpDir := dir + ".tmp"
	_ = os.RemoveAll(tmpDir)
	if err := cloneInto(src, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("git clone of %s failed: %w\n"+
			"       Check the URL and network access, or pass --source <path> to a local checkout", src.Location, err)
	}
	commit, err := gitOutput(tmpDir, "rev-parse", "HEAD")
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("resolve cloned commit: %w", err)
	}
	// Drop .git to keep the cache lean, the commit is kept in the meta file
	_ = os.RemoveAll(filepath.Join(tmpDir, ".git"))
	if err := writeCacheMeta(tmpDir, commit); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}
	_ = os.RemoveAll(dir)
	if err := os.Rename(tmpDir, dir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("promote cache dir: %w", err)
	}
	return &Materialized{Dir: dir, Commit: commit}, nil
}

// cloneInto performs a shallow clone; commit SHAs need init+fetch because
// `git clone --branch` accepts only branches and tags
func cloneInto(src Source, dir string) error {
	if src.Ref != "" && shaPattern.MatchString(src.Ref) {
		if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 -- cache tree uses regular directory permissions
			return err
		}
		steps := [][]string{
			{"init", "--quiet"},
			{"remote", "add", "origin", src.Location},
			{"fetch", "--quiet", "--depth", "1", "origin", src.Ref},
			{"checkout", "--quiet", "FETCH_HEAD"},
		}
		for _, args := range steps {
			if err := runGit(dir, args...); err != nil {
				return err
			}
		}
		return nil
	}
	args := []string{"clone", "--quiet", "--depth", "1"}
	if src.Ref != "" {
		args = append(args, "--branch", src.Ref)
	}
	args = append(args, "--", src.Location, dir)
	return runGit("", args...)
}

// CacheDir returns ~/.cache/portunix/skills/<host>/<owner>/<repo>/<ref>/
func CacheDir(src Source) (string, error) {
	root, err := CacheRoot()
	if err != nil {
		return "", err
	}
	host, owner, repo, err := repoKey(src.Location)
	if err != nil {
		return "", err
	}
	ref := src.Ref
	if ref == "" {
		ref = "default"
	}
	return filepath.Join(root, host, owner, repo, sanitizeSegment(ref)), nil
}

// CacheRoot returns the per-user skills cache root (PORTUNIX_SKILLS_CACHE overrides)
func CacheRoot() (string, error) {
	if env := os.Getenv("PORTUNIX_SKILLS_CACHE"); env != "" {
		return env, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	return filepath.Join(base, "portunix", "skills"), nil
}

type cacheMeta struct {
	Commit string `json:"commit"`
}

func readCacheMeta(dir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(dir, cacheMetaFile)) // #nosec G304 -- path inside the Portunix cache
	if err != nil {
		return "", false
	}
	var m cacheMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return "", false
	}
	return m.Commit, true
}

func writeCacheMeta(dir, commit string) error {
	data, err := json.Marshal(cacheMeta{Commit: commit})
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, cacheMetaFile), data, 0o644); err != nil { // #nosec G306 -- cache metadata is not sensitive
		return fmt.Errorf("write cache metadata: %w", err)
	}
	return nil
}

// localCommit returns the HEAD commit of a local checkout, or "" when the
// directory is not a git work tree or git is unavailable
func localCommit(dir string) string {
	if _, err := exec.LookPath("git"); err != nil {
		return ""
	}
	commit, err := gitOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return commit
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...) // #nosec G204 -- fixed git subcommands, URL passed after "--" or as remote value
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%v: %s", err, msg)
		}
		return err
	}
	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) // #nosec G204 -- fixed git subcommands
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func requireGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		hint := "install it with 'portunix install git'"
		if runtime.GOOS == "windows" {
			hint = "install Git for Windows ('portunix install git')"
		}
		return fmt.Errorf("installing skills from a git URL requires a local 'git' binary, which was not found on PATH.\n"+
			"       Either %s, or pass a local checkout path / --source <path>", hint)
	}
	return nil
}

// sanitizeSegment turns an arbitrary string into a directory-safe segment
func sanitizeSegment(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '-', r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	seg := string(out)
	if seg == "" || seg == "." || seg == ".." {
		return "default"
	}
	return seg
}
