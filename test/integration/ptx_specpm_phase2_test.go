/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue184_PtxSpecpm_Phase2_AirGapped exercises Phase 2 (#184): the
// upgrade subcommand against a local kit fixture, with PATH stripped of the
// host's `git` binary so the air-gapped invariant from ADR-041 D5 is asserted
// hermetically: the helper must complete init + upgrade + check without any
// git invocation, and user content under specs/project/ must survive.
//
// Note on D5 deviation: ADR-041 D5 prescribes `portunix container
// run-in-container <image>` against alpine:latest. That subcommand is
// specifically for installation-package testing (it does not support generic
// volume mounts for arbitrary executables); a true containerised end-to-end
// test of upgrade requires `portunix container run` with bind-mounts and is
// scoped as a follow-up. This hermetic host test asserts the same invariant
// (no git invocation, sentinel survives upgrade, --source <path> only) by
// stripping `git` from PATH for the duration of the test.
//
// Run with:
//
//	go test ./test/integration/ptx_specpm_phase2_test.go -v -timeout 5m
func TestIssue184_PtxSpecpm_Phase2_AirGapped(t *testing.T) {
	tf := testframework.NewTestFramework("Issue184_PtxSpecpm_Phase2_AirGapped")
	tf.Start(t, "ptx-specpm Phase 2: upgrade air-gapped flow with PATH stripped of git")

	success := true
	defer func() { tf.Finish(t, success) }()

	root := repoRoot(t, tf)
	binary := filepath.Join(root, "portunix")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if _, err := os.Stat(binary); err != nil {
		tf.Error(t, "portunix binary missing — run 'make build' first", err.Error())
		t.Skipf("portunix binary not built at %s", binary)
		return
	}

	tf.Step(t, "Build a local kit fixture")
	kitDir := buildKitFixture(t, tf)

	tf.Separator()

	// Strip 'git' from PATH for every subprocess invocation. We construct a
	// PATH that contains only directories where the real toolchain lives but
	// that excludes any directory holding a 'git' binary.
	tf.Step(t, "Build PATH without 'git' (air-gapped invariant)")
	cleanPath := pathWithoutGit(t, tf)
	tf.Info(t, "Stripped PATH (head): "+truncate(cleanPath, 100))

	// Sanity: confirm git really is unreachable under the stripped PATH.
	if locator, err := exec.LookPath("git"); err == nil {
		// LookPath uses os.Getenv("PATH"); we override per-process below, but
		// log the host's git location so the test report is unambiguous.
		tf.Info(t, "Host git resolved to "+locator+" (will be hidden via stripped PATH below)")
	}
	if _, found := lookPathIn(cleanPath, "git"); found {
		tf.Error(t, "PATH stripping failed — 'git' still resolvable under cleanPath", cleanPath)
		success = false
		return
	}
	tf.Success(t, "'git' is not resolvable under the stripped PATH")
	tf.Separator()

	target := t.TempDir()
	tf.Info(t, "Project target: "+target)

	// 1. init --here --source <path>: Phase 1 surface, but no git on PATH.
	tf.Step(t, "init --here --integration claude --source <kit> (no git on PATH)")
	out, err := runCmdWith(target, cleanPath, binary,
		"specpm", "init", "--here",
		"--integration", "claude",
		"--source", kitDir,
	)
	if err != nil {
		tf.Error(t, "init failed under stripped PATH", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, "init complete") {
		tf.Error(t, "init missing 'init complete' line", out)
		success = false
	}
	tf.Success(t, "init succeeded with no git on PATH")
	tf.Separator()

	// 2. Plant sentinel artifacts that MUST survive upgrade.
	tf.Step(t, "Plant sentinels: specs/project/<id>/USER.md and .specpm/extensions/my-ext/")
	projectID := filepath.Base(target)
	// init derives the project ID from the dir name; sanitize the same way
	// (lowercase, with non-[a-z0-9] mapped to '-' or stripped). For tempdirs
	// like ptx-184-Air... the writer's sanitiseProjectID will produce a
	// predictable path. We don't assume — read kit.json to confirm.
	kitJSONPath := filepath.Join(target, ".specpm", "kit.json")
	if _, err := os.Stat(kitJSONPath); err != nil {
		tf.Error(t, "kit.json missing after init", err.Error())
		success = false
		return
	}
	// Find the actual project directory under specs/project/.
	projectsDir := filepath.Join(target, "specs", "project")
	entries, err := os.ReadDir(projectsDir)
	if err != nil || len(entries) == 0 {
		tf.Error(t, "specs/project/ is empty after init", "")
		success = false
		return
	}
	projectID = entries[0].Name()
	userArtifact := filepath.Join(projectsDir, projectID, "USER.md")
	if err := os.WriteFile(userArtifact, []byte("user-sentinel\n"), 0o644); err != nil {
		tf.Error(t, "plant USER.md failed", err.Error())
		success = false
		return
	}
	extDir := filepath.Join(target, ".specpm", "extensions", "my-ext")
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		tf.Error(t, "mkdir extensions sentinel", err.Error())
		success = false
		return
	}
	extMarker := filepath.Join(extDir, "marker.txt")
	if err := os.WriteFile(extMarker, []byte("ext-marker\n"), 0o644); err != nil {
		tf.Error(t, "plant ext marker", err.Error())
		success = false
		return
	}
	presetDir := filepath.Join(target, ".specpm", "presets", "my-preset")
	if err := os.MkdirAll(presetDir, 0o755); err != nil {
		tf.Error(t, "mkdir presets sentinel", err.Error())
		success = false
		return
	}
	presetMarker := filepath.Join(presetDir, "marker.txt")
	if err := os.WriteFile(presetMarker, []byte("preset-marker\n"), 0o644); err != nil {
		tf.Error(t, "plant preset marker", err.Error())
		success = false
		return
	}
	tf.Success(t, "Sentinels planted (artifact + ext + preset)")
	tf.Separator()

	// 3. upgrade --source <path>: should refresh kit cache, preserve sentinels,
	//    succeed without git on PATH.
	tf.Step(t, "upgrade --source <kit>  (still no git on PATH)")
	out, err = runCmdWith(target, cleanPath, binary,
		"specpm", "upgrade",
		"--source", kitDir,
	)
	if err != nil {
		tf.Error(t, "upgrade failed under stripped PATH", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, "upgrade complete") {
		tf.Error(t, "upgrade missing 'upgrade complete' line", out)
		success = false
	}
	if !strings.Contains(out, "Refreshed subtrees") {
		tf.Error(t, "upgrade summary missing refreshed-subtrees line", out)
		success = false
	}
	if !strings.Contains(out, "specs/project/      : not touched") {
		tf.Error(t, "upgrade summary missing specs/project/ guarantee line", out)
		success = false
	}
	tf.Success(t, "upgrade succeeded with no git on PATH")
	tf.Separator()

	// 4. Verify sentinels survived the refresh.
	tf.Step(t, "Verify sentinels survived upgrade (carve-outs + sacred artifacts)")
	mustSurvive := func(label, path, expected string) {
		got, err := os.ReadFile(path)
		if err != nil {
			tf.Error(t, label+" missing after upgrade", err.Error())
			success = false
			return
		}
		if string(got) != expected {
			tf.Error(t, label+" content drifted after upgrade", string(got))
			success = false
			return
		}
		tf.Info(t, "  ✓ "+label+" preserved")
	}
	mustSurvive("specs/project/<id>/USER.md", userArtifact, "user-sentinel\n")
	mustSurvive(".specpm/extensions/my-ext/marker.txt", extMarker, "ext-marker\n")
	mustSurvive(".specpm/presets/my-preset/marker.txt", presetMarker, "preset-marker\n")
	tf.Success(t, "All sentinels survived (artifact + ext + preset)")
	tf.Separator()

	// 5. check: should report initialised and (since project pin is "(local
	//    path)" not DefaultPinnedRef) drift line should NOT appear (drift is
	//    git-mode-only per check.go).
	tf.Step(t, "check  (no git on PATH; project on --source <path>)")
	out, err = runCmdWith(target, cleanPath, binary, "specpm", "check")
	if err != nil {
		tf.Error(t, "check failed under stripped PATH", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, "Project          : initialised") {
		tf.Error(t, "check missing 'initialised' line", out)
		success = false
	}
	if strings.Contains(out, "Drift            :") {
		tf.Error(t, "check should NOT show drift on --source <path> (drift is git-mode-only)", out)
		success = false
	}
	tf.Success(t, "check reports initialised + no drift on local-path source")
	tf.Separator()

	// 6. upgrade --to-default WITH git stripped AND a cold cache must fail
	//    with the locked remediation hint from ADR-041 D4. We force a cold
	//    cache via PORTUNIX_SPECPM_CACHE pointing at a fresh tempdir, so the
	//    test is independent of any host cache the developer may have warmed.
	tf.Step(t, "upgrade --to-default with cold cache (must fail with locked remediation hint)")
	coldCache := t.TempDir()
	out, err = runCmdWithEnv(target, cleanPath, binary,
		[]string{"PORTUNIX_SPECPM_CACHE=" + coldCache},
		"specpm", "upgrade", "--to-default",
	)
	if err == nil {
		tf.Error(t, "upgrade --to-default unexpectedly succeeded with cold cache and no git", out)
		success = false
	} else if !strings.Contains(out, "Air-gapped workflow") || !strings.Contains(out, "vendored mirror") {
		tf.Error(t, "remediation hint missing or paraphrased — ADR-041 D4 wording must be verbatim", out)
		success = false
	} else {
		tf.Success(t, "Locked remediation hint present in error output")
	}
	tf.Separator()
}

// runCmdWith runs a binary with a custom working dir and PATH.
func runCmdWith(dir, path, bin string, args ...string) (string, error) {
	return runCmdWithEnv(dir, path, bin, nil, args...)
}

// runCmdWithEnv runs a binary with a custom working dir, PATH, and a list of
// extra "KEY=VALUE" entries (each strips any matching key from inherited env).
func runCmdWithEnv(dir, path, bin string, extra []string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	env := os.Environ()
	// Strip PATH and any keys overridden in extra, then append our values.
	overrideKeys := map[string]struct{}{"PATH": {}}
	for _, kv := range extra {
		if i := strings.IndexByte(kv, '='); i > 0 {
			overrideKeys[kv[:i]] = struct{}{}
		}
	}
	stripped := make([]string, 0, len(env)+len(extra)+1)
	for _, kv := range env {
		if i := strings.IndexByte(kv, '='); i > 0 {
			if _, ok := overrideKeys[kv[:i]]; ok {
				continue
			}
		}
		stripped = append(stripped, kv)
	}
	stripped = append(stripped, "PATH="+path)
	stripped = append(stripped, extra...)
	cmd.Env = stripped
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// pathWithoutGit returns a colon-separated PATH built from the host's PATH
// minus any directory that contains a 'git' executable.
func pathWithoutGit(t *testing.T, tf *testframework.TestFramework) string {
	t.Helper()
	host := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	out := make([]string, 0)
	for _, dir := range strings.Split(host, sep) {
		if dir == "" {
			continue
		}
		gitBin := filepath.Join(dir, "git")
		if runtime.GOOS == "windows" {
			gitBin += ".exe"
		}
		if _, err := os.Stat(gitBin); err == nil {
			tf.Info(t, "  - dropping "+dir+" (contains git)")
			continue
		}
		out = append(out, dir)
	}
	if len(out) == 0 {
		// Defensive: all PATH dirs had git? Just give a single dir that
		// definitely has core tools (the ptx-specpm test does not need
		// anything beyond the binary it invokes).
		out = []string{"/usr/bin"}
	}
	return strings.Join(out, sep)
}

// lookPathIn searches for `name` across the colon-separated `path`. Returns
// the absolute path and true if found.
func lookPathIn(path, name string) (string, bool) {
	sep := string(os.PathListSeparator)
	for _, dir := range strings.Split(path, sep) {
		if dir == "" {
			continue
		}
		full := filepath.Join(dir, name)
		if runtime.GOOS == "windows" && !strings.HasSuffix(full, ".exe") {
			full += ".exe"
		}
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return full, true
		}
	}
	return "", false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
