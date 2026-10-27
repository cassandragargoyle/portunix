/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// scriptedPrompter answers every confirmation with a fixed value
type scriptedPrompter struct {
	interactive bool
	answer      bool
	asked       []string
}

func (p *scriptedPrompter) Interactive() bool { return p.interactive }
func (p *scriptedPrompter) Confirm(q string) bool {
	p.asked = append(p.asked, q)
	return p.answer
}
func (p *scriptedPrompter) ChooseSkills(available []ResolvedSkill) ([]string, error) {
	return []string{available[0].Name}, nil
}

// setupEnv isolates project dir, home and cache in temp directories
func setupEnv(t *testing.T) (project, home string) {
	t.Helper()
	root := t.TempDir()
	project = filepath.Join(root, "project")
	home = filepath.Join(root, "home")
	for _, d := range []string{project, home} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("PORTUNIX_SKILLS_CACHE", filepath.Join(root, "cache"))
	t.Setenv("PORTUNIX_SKILLS_TRUST", filepath.Join(root, "config", "trust.json"))
	t.Chdir(project)
	SetDetector(func(string) bool { return false })
	return project, home
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixtureSource creates a local skill repository with skills alpha and beta
func fixtureSource(t *testing.T) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "repo")
	writeFile(t, filepath.Join(src, "skills", "alpha", "SKILL.md"), "---\nname: alpha\ndescription: >\n  First\n  skill\n---\n# Alpha\n")
	writeFile(t, filepath.Join(src, "skills", "alpha", "refs", "notes.md"), "notes\n")
	writeFile(t, filepath.Join(src, "skills", "beta", "SKILL.md"), "---\nname: \"beta\"\ndescription: 'Second skill'\n---\n")
	writeFile(t, filepath.Join(src, "README.md"), "not a skill\n")
	return src
}

func addLocal(t *testing.T, src string, opts AddOptions) (*Result, error) {
	t.Helper()
	s, err := ParseSource(src, "")
	if err != nil {
		t.Fatal(err)
	}
	opts.Source = s
	if opts.Targets == nil {
		opts.Targets = []AgentTarget{Universal(), mustAgent(t, "claude")}
	}
	return Add(opts)
}

func mustAgent(t *testing.T, id string) AgentTarget {
	t.Helper()
	a, err := LookupAgent(id)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestParseSource(t *testing.T) {
	local := t.TempDir()
	tests := []struct {
		arg      string
		kind     SourceKind
		location string
		wantErr  bool
	}{
		{arg: "https://github.com/obra/superpowers", kind: SourceGit, location: "https://github.com/obra/superpowers.git"},
		{arg: "https://github.com/obra/superpowers.git/", kind: SourceGit, location: "https://github.com/obra/superpowers.git"},
		{arg: "obra/superpowers", kind: SourceGit, location: "https://github.com/obra/superpowers.git"},
		{arg: "git@github.com:obra/superpowers.git", kind: SourceGit, location: "git@github.com:obra/superpowers.git"},
		{arg: local, kind: SourceLocal, location: local},
		{arg: "not a url", wantErr: true},
		{arg: "https://github.com/onlyowner", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ParseSource(tt.arg, "")
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseSource(%q) expected error", tt.arg)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSource(%q): %v", tt.arg, err)
			continue
		}
		if got.Kind != tt.kind || got.Location != tt.location {
			t.Errorf("ParseSource(%q) = %+v, want kind %v location %q", tt.arg, got, tt.kind, tt.location)
		}
	}
}

func TestCacheDirLayout(t *testing.T) {
	t.Setenv("PORTUNIX_SKILLS_CACHE", "/cache")
	dir, err := CacheDir(Source{Kind: SourceGit, Location: "https://github.com/obra/superpowers.git", Ref: "v1/rc"})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/cache", "github.com", "obra", "superpowers", "v1_rc")
	if dir != want {
		t.Errorf("CacheDir = %q, want %q", dir, want)
	}
	dir, _ = CacheDir(Source{Kind: SourceGit, Location: "git@gitlab.com:group/sub/repo.git"})
	want = filepath.Join("/cache", "gitlab.com", "group_sub", "repo", "default")
	if dir != want {
		t.Errorf("CacheDir = %q, want %q", dir, want)
	}
}

func TestMaterializeCacheHitNeedsNoGit(t *testing.T) {
	setupEnv(t)
	t.Setenv("PATH", "") // a cache hit must not invoke git
	src := Source{Kind: SourceGit, Location: "https://example.invalid/owner/repo.git"}
	dir, err := CacheDir(src)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "s", SkillFile), "---\nname: s\n---\n")
	if err := writeCacheMeta(dir, "abc123"); err != nil {
		t.Fatal(err)
	}
	mat, err := Materialize(src, false)
	if err != nil {
		t.Fatal(err)
	}
	if !mat.FromCache || mat.Commit != "abc123" || mat.Dir != dir {
		t.Errorf("unexpected materialization %+v", mat)
	}
	if _, err := Materialize(src, true); err == nil || !strings.Contains(err.Error(), "git") {
		t.Errorf("refresh without git should fail with a git hint, got %v", err)
	}
}

func TestDiscoverAndFrontMatter(t *testing.T) {
	src := fixtureSource(t)
	// invalid name and nested skill inside a skill are ignored
	writeFile(t, filepath.Join(src, "skills", "bad", "SKILL.md"), "---\nname: ../escape\n---\n")
	writeFile(t, filepath.Join(src, "skills", "alpha", "nested", "SKILL.md"), "---\nname: nested\n---\n")
	writeFile(t, filepath.Join(src, ".git", "x", "SKILL.md"), "---\nname: hidden\n---\n")

	found, warnings, err := Discover(src)
	if err != nil {
		t.Fatal(err)
	}
	if got := skillNames(found); !reflect.DeepEqual(got, []string{"alpha", "beta"}) {
		t.Fatalf("discovered %v", got)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "invalid name") {
		t.Errorf("warnings = %v", warnings)
	}
	if found[0].Description != "First skill" || found[1].Description != "Second skill" {
		t.Errorf("descriptions = %q, %q", found[0].Description, found[1].Description)
	}
	wantFiles := []string{"SKILL.md", "nested/SKILL.md", "refs/notes.md"}
	if !reflect.DeepEqual(found[0].Files, wantFiles) {
		t.Errorf("files = %v, want %v", found[0].Files, wantFiles)
	}

	if _, err := SelectSkills(found, []string{"gamma"}, false); err == nil {
		t.Error("selecting an unknown skill should fail")
	}
	sel, err := SelectSkills(found, []string{"beta", "beta"}, false)
	if err != nil || len(sel) != 1 {
		t.Errorf("SelectSkills dedupe: %v %v", sel, err)
	}
}

func TestManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m, err := LoadManifest(dir)
	if err != nil || len(m.Skills) != 0 {
		t.Fatalf("empty manifest: %+v %v", m, err)
	}
	m.Upsert(ManifestEntry{Name: "b", Source: "s1"})
	m.Upsert(ManifestEntry{Name: "a", Source: "s2"})
	m.Upsert(ManifestEntry{Name: "b", Source: "s3"})
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Skills) != 2 || loaded.Skills[0].Name != "a" || loaded.Find("b").Source != "s3" {
		t.Errorf("loaded %+v", loaded.Skills)
	}
	if !loaded.HasSource("s2") || loaded.HasSource("s1") {
		t.Error("HasSource mismatch")
	}
	loaded.Remove("a")
	loaded.Remove("b")
	if err := loaded.Save(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ManifestFile)); !os.IsNotExist(err) {
		t.Error("empty manifest should remove the file")
	}
}

func TestAddInstallsIntoTargetsAndIsIdempotent(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)

	res, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Yes: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".agents/skills/alpha/SKILL.md", ".agents/skills/alpha/refs/notes.md", ".claude/skills/alpha/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	if len(res.Outcomes) != 2 || res.Outcomes[0].Action != ActionInstall {
		t.Errorf("outcomes %+v", res.Outcomes)
	}
	m, _ := LoadManifest(filepath.Join(project, ".claude", "skills"))
	e := m.Find("alpha")
	if e == nil || e.Source != src || e.Checksum == "" || len(e.Files) != 2 {
		t.Fatalf("manifest entry %+v", e)
	}

	// second run: known source, no prompt, nothing written
	p := &scriptedPrompter{}
	res, err = addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Prompt: p})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range res.Outcomes {
		if o.Action != ActionUnchanged {
			t.Errorf("expected unchanged, got %+v", o)
		}
	}
	if len(p.asked) != 0 {
		t.Errorf("unexpected prompts %v", p.asked)
	}
}

func TestTrustGate(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)

	_, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Prompt: &scriptedPrompter{}})
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("non-interactive first install must require --yes, got %v", err)
	}
	p := &scriptedPrompter{interactive: true, answer: false}
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Prompt: p}); err == nil {
		t.Fatal("declined trust prompt should cancel")
	}
	if len(p.asked) != 1 || !strings.Contains(p.asked[0], filepath.Join(project, ".agents", "skills", "alpha")) {
		t.Errorf("prompt should list target directories: %v", p.asked)
	}
	if _, err := os.Stat(filepath.Join(project, ".agents")); !os.IsNotExist(err) {
		t.Error("nothing may be written before trust is granted")
	}
}

func TestConflictsRequireForce(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Yes: true}); err != nil {
		t.Fatal(err)
	}

	// clean managed skill + changed source: update without prompt
	writeFile(t, filepath.Join(src, "skills", "alpha", "SKILL.md"), "---\nname: alpha\n---\nv2\n")
	res, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outcomes[0].Action != ActionUpdate {
		t.Errorf("expected update, got %+v", res.Outcomes[0])
	}

	// local modification: skipped without --force, overwritten with it
	local := filepath.Join(project, ".claude", "skills", "alpha", "SKILL.md")
	writeFile(t, local, "edited\n")
	writeFile(t, filepath.Join(src, "skills", "alpha", "SKILL.md"), "---\nname: alpha\n---\nv3\n")
	res, err = addLocal(t, src, AddOptions{Skills: []string{"alpha"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Conflicts() != 1 {
		t.Fatalf("expected one conflict, got %+v", res.Outcomes)
	}
	if data, _ := os.ReadFile(local); string(data) != "edited\n" {
		t.Error("conflicting skill must not be overwritten without --force")
	}
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Force: true}); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(local); !strings.Contains(string(data), "v3") {
		t.Error("--force should overwrite the modified skill")
	}

	// unmanaged directory with the same name is a conflict too
	writeFile(t, filepath.Join(project, ".agents", "skills", "beta", "SKILL.md"), "mine\n")
	res, _ = addLocal(t, src, AddOptions{Skills: []string{"beta"}, Targets: []AgentTarget{Universal()}})
	if res.Conflicts() != 1 || !strings.Contains(res.Outcomes[0].Reason, "not managed") {
		t.Errorf("unmanaged dir should conflict: %+v", res.Outcomes)
	}
}

func TestUserScopeAndList(t *testing.T) {
	project, home := setupEnv(t)
	src := fixtureSource(t)
	if _, err := addLocal(t, src, AddOptions{All: true, Scope: ScopeUser, Yes: true, Targets: []AgentTarget{Universal()}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "beta", SkillFile)); err != nil {
		t.Errorf("user scope install missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agents")); !os.IsNotExist(err) {
		t.Error("user scope must not write into the project")
	}
	writeFile(t, filepath.Join(home, ".agents", "skills", "manual", SkillFile), "hand placed\n")
	writeFile(t, filepath.Join(home, ".agents", "skills", "beta", "extra.md"), "x\n")

	entries, err := List(ScopeUser)
	if err != nil {
		t.Fatal(err)
	}
	status := map[string]string{}
	for _, e := range entries {
		status[e.Name] = e.Status
	}
	want := map[string]string{"alpha": "ok", "beta": "modified", "manual": "unmanaged"}
	if !reflect.DeepEqual(status, want) {
		t.Errorf("list status = %v, want %v", status, want)
	}
	if project, _ := List(ScopeProject); len(project) != 0 {
		t.Errorf("project scope should be empty: %+v", project)
	}
}

func TestRemoveOnlyManagedFiles(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)
	if _, err := addLocal(t, src, AddOptions{All: true, Yes: true}); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(project, ".agents", "skills", "alpha", "mine.md")
	writeFile(t, extra, "keep me\n")

	res, err := Remove("alpha", nil, ScopeProject)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Kept) != 1 || len(res.Removed) != 1 {
		t.Errorf("remove result %+v", res)
	}
	if _, err := os.Stat(extra); err != nil {
		t.Error("files not installed by Portunix must be kept")
	}
	if _, err := os.Stat(filepath.Join(project, ".claude", "skills", "alpha")); !os.IsNotExist(err) {
		t.Error("claude copy should be removed")
	}
	m, _ := LoadManifest(filepath.Join(project, ".agents", "skills"))
	if m.Find("alpha") != nil || m.Find("beta") == nil {
		t.Errorf("manifest after remove: %+v", m.Skills)
	}

	writeFile(t, filepath.Join(project, ".agents", "skills", "manual", SkillFile), "x\n")
	if _, err := Remove("manual", nil, ScopeProject); err == nil {
		t.Error("removing an unmanaged skill must fail")
	}
	if _, err := Remove("../x", nil, ScopeProject); err == nil {
		t.Error("invalid names must be rejected")
	}
}

func TestRemoveRejectsTamperedManifest(t *testing.T) {
	project, _ := setupEnv(t)
	dir := filepath.Join(project, ".agents", "skills")
	victim := filepath.Join(project, "victim.txt")
	writeFile(t, victim, "x")
	m := &Manifest{Version: 1, Skills: []ManifestEntry{{Name: "evil", Files: []string{"../../victim.txt"}}}}
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := Remove("evil", []AgentTarget{Universal()}, ScopeProject); err == nil {
		t.Error("path escaping the skill directory must be rejected")
	}
	if _, err := os.Stat(victim); err != nil {
		t.Error("file outside the skill directory was deleted")
	}
}

func TestUpdateRefreshesManagedSkills(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)
	if _, err := addLocal(t, src, AddOptions{All: true, Yes: true, Targets: []AgentTarget{Universal()}}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(src, "skills", "beta", "SKILL.md"), "---\nname: beta\n---\nnew\n")
	if err := os.RemoveAll(filepath.Join(src, "skills", "alpha")); err != nil {
		t.Fatal(err)
	}

	results, err := Update(UpdateOptions{Scope: ScopeProject})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].Outcomes) != 1 || results[0].Outcomes[0].Action != ActionUpdate {
		t.Fatalf("update results %+v", results)
	}
	if len(results[0].Warnings) != 1 || !strings.Contains(results[0].Warnings[0], "alpha") {
		t.Errorf("missing-skill warning expected: %v", results[0].Warnings)
	}
	data, _ := os.ReadFile(filepath.Join(project, ".agents", "skills", "beta", SkillFile))
	if !strings.Contains(string(data), "new") {
		t.Error("beta not updated")
	}
	if _, err := Update(UpdateOptions{Names: []string{"gamma"}, Scope: ScopeProject}); err == nil {
		t.Error("updating an unmanaged skill must fail")
	}
}

func targetIDs(ts []AgentTarget) []string {
	var out []string
	for _, t := range ts {
		out = append(out, t.ID())
	}
	return out
}

func TestResolveTargets(t *testing.T) {
	setupEnv(t)
	SetDetector(func(pkg string) bool { return pkg == "claude-code" })
	ids := targetIDs
	got, _ := ResolveTargets(nil, false, false)
	if !reflect.DeepEqual(ids(got), []string{"universal", "claude"}) {
		t.Errorf("detected defaults = %v", ids(got))
	}
	SetDetector(func(string) bool { return false })
	got, _ = ResolveTargets(nil, false, false)
	if !reflect.DeepEqual(ids(got), []string{"universal"}) {
		t.Errorf("no detection = %v", ids(got))
	}
	if _, err := ResolveTargets(nil, false, true); err == nil {
		t.Error("--no-universal without agents must fail")
	}
	if _, err := ResolveTargets([]string{"nope"}, false, false); err == nil {
		t.Error("unknown agent must fail")
	}
	got, _ = ResolveTargets(nil, true, true)
	want := []string{"amp", "claude", "copilot", "cursor", "gemini", "opencode", "zed"}
	if !reflect.DeepEqual(ids(got), want) {
		t.Errorf("all agents = %v, want %v", ids(got), want)
	}
	got, err := ResolveTargets([]string{"claude-code", "claude", "gemini-cli", "github-copilot"}, false, true)
	if err != nil || !reflect.DeepEqual(ids(got), []string{"claude", "gemini", "copilot"}) {
		t.Errorf("aliases = %v, %v", ids(got), err)
	}
	got, err = ResolveTargets([]string{}, false, false)
	if err != nil || !reflect.DeepEqual(ids(got), []string{"universal"}) {
		t.Errorf("explicit empty selection must skip detection: %v, %v", ids(got), err)
	}
}

func TestAgentDetectionByConfigDir(t *testing.T) {
	_, home := setupEnv(t)
	if err := os.MkdirAll(filepath.Join(home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".config", "zed"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, _ := ResolveTargets(nil, false, false)
	if want := []string{"universal", "cursor", "zed"}; !reflect.DeepEqual(targetIDs(got), want) {
		t.Errorf("detected = %v, want %v", targetIDs(got), want)
	}
	SetDetector(func(pkg string) bool { return pkg == "gemini-cli" })
	gemini, _ := LookupAgent("gemini")
	if ok, _ := gemini.Detect(); !ok {
		t.Error("gemini should be detected through the registry package")
	}
}

func TestAgentSkillsDirs(t *testing.T) {
	project, home := setupEnv(t)
	cfg := filepath.Join(home, ".config")
	want := map[string][2]string{
		"universal": {filepath.Join(project, ".agents", "skills"), filepath.Join(home, ".agents", "skills")},
		"amp":       {filepath.Join(project, ".agents", "skills"), filepath.Join(cfg, "agents", "skills")},
		"claude":    {filepath.Join(project, ".claude", "skills"), filepath.Join(home, ".claude", "skills")},
		"copilot":   {filepath.Join(project, ".agents", "skills"), filepath.Join(home, ".copilot", "skills")},
		"cursor":    {filepath.Join(project, ".agents", "skills"), filepath.Join(home, ".cursor", "skills")},
		"gemini":    {filepath.Join(project, ".agents", "skills"), filepath.Join(home, ".gemini", "skills")},
		"opencode":  {filepath.Join(project, ".agents", "skills"), filepath.Join(cfg, "opencode", "skills")},
		"zed":       {filepath.Join(project, ".agents", "skills"), filepath.Join(home, ".agents", "skills")},
	}
	for _, target := range AllTargets() {
		w, ok := want[target.ID()]
		if !ok {
			t.Errorf("unexpected target %s", target.ID())
			continue
		}
		pd, _ := target.SkillsDir(ScopeProject)
		ud, _ := target.SkillsDir(ScopeUser)
		if pd != w[0] || ud != w[1] {
			t.Errorf("%s dirs = %s | %s, want %s | %s", target.ID(), pd, ud, w[0], w[1])
		}
	}

	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, "claude-alt"))
	opencode, _ := LookupAgent("opencode")
	claude, _ := LookupAgent("claude")
	if d, _ := opencode.SkillsDir(ScopeUser); d != filepath.Join(xdg, "opencode", "skills") {
		t.Errorf("XDG_CONFIG_HOME not honoured: %s", d)
	}
	if d, _ := claude.SkillsDir(ScopeUser); d != filepath.Join(home, "claude-alt", "skills") {
		t.Errorf("CLAUDE_CONFIG_DIR not honoured: %s", d)
	}
}

func TestSharedDirectoryWrittenOnce(t *testing.T) {
	project, home := setupEnv(t)
	src := fixtureSource(t)
	targets, err := ResolveTargets(nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	res, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Yes: true, Targets: targets})
	if err != nil {
		t.Fatal(err)
	}
	// project scope: .agents/skills (universal + 6 agents) and .claude/skills
	if len(res.Outcomes) != 2 {
		t.Fatalf("outcomes %+v", res.Outcomes)
	}
	if res.Outcomes[0].Target != "universal,amp,copilot,cursor,gemini,opencode,zed" || res.Outcomes[1].Target != "claude" {
		t.Errorf("targets = %q, %q", res.Outcomes[0].Target, res.Outcomes[1].Target)
	}
	entries, _ := List(ScopeProject)
	if len(entries) != 2 || len(entries[0].Targets) != 7 {
		t.Errorf("list should show one row per directory: %+v", entries)
	}

	// user scope: universal and zed share ~/.agents/skills, the others differ
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"beta"}, Scope: ScopeUser, Targets: targets}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".agents/skills", ".config/agents/skills", ".claude/skills", ".copilot/skills",
		".cursor/skills", ".gemini/skills", ".config/opencode/skills"} {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(rel), "beta", SkillFile)); err != nil {
			t.Errorf("missing user-scope install in %s", rel)
		}
	}
	entries, _ = List(ScopeUser)
	if len(entries) != 7 {
		t.Errorf("expected 7 user-scope directories, got %d", len(entries))
	}
	if _, err := Remove("beta", []AgentTarget{mustAgent(t, "zed")}, ScopeUser); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "beta")); !os.IsNotExist(err) {
		t.Error("removing via zed must clear the shared ~/.agents/skills copy")
	}
	if _, err := os.Stat(filepath.Join(project, ".agents", "skills", "alpha", SkillFile)); err != nil {
		t.Error("project install must be untouched")
	}
}

func TestWriteFailureReportedAsFailed(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permission checks do not apply to root")
	}
	project, _ := setupEnv(t)
	src := fixtureSource(t)
	ro := filepath.Join(project, ".agents")
	if err := os.MkdirAll(ro, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o755) })
	res, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Yes: true, Targets: []AgentTarget{Universal()}})
	if err == nil {
		t.Fatal("expected a write error")
	}
	if res == nil || res.Failures() != 1 || res.Conflicts() != 0 || res.Outcomes[0].Action != ActionFailed {
		t.Fatalf("write error must be reported as failed: %+v", res)
	}
}

func TestInteractiveTargetSelection(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)
	called := false
	_, err := addLocal(t, src, AddOptions{
		Prompt: &scriptedPrompter{interactive: true, answer: true},
		SelectTargets: func() ([]AgentTarget, error) {
			called = true
			return []AgentTarget{mustAgent(t, "claude")}, nil
		},
		Targets: []AgentTarget{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("SelectTargets not called")
	}
	// scriptedPrompter picks the first skill, SelectTargets only claude
	if _, err := os.Stat(filepath.Join(project, ".claude", "skills", "alpha", SkillFile)); err != nil {
		t.Error("interactively selected target not written")
	}
	if _, err := os.Stat(filepath.Join(project, ".agents")); !os.IsNotExist(err) {
		t.Error("unselected targets must not be written")
	}
}

func TestTrustRecordedAcrossProjects(t *testing.T) {
	project, _ := setupEnv(t)
	src := fixtureSource(t)

	// declined prompt records nothing
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Prompt: &scriptedPrompter{interactive: true}}); err == nil {
		t.Fatal("declined trust must cancel")
	}
	if ok, _ := IsTrusted(src); ok {
		t.Fatal("declined source must not be trusted")
	}

	if _, err := addLocal(t, src, AddOptions{Skills: []string{"alpha"}, Yes: true}); err != nil {
		t.Fatal(err)
	}
	if ok, err := IsTrusted(src); !ok || err != nil {
		t.Fatalf("source not trusted after install: %v", err)
	}

	// another project: no prompt, no --yes needed
	other := filepath.Join(filepath.Dir(project), "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(other)
	p := &scriptedPrompter{}
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"beta"}, Prompt: p}); err != nil {
		t.Fatalf("trusted source must install without confirmation: %v", err)
	}
	if len(p.asked) != 0 {
		t.Errorf("unexpected prompts %v", p.asked)
	}

	// revoked trust asks again in a project without a manifest record
	if removed, err := UntrustSource(src); !removed || err != nil {
		t.Fatalf("untrust: %v %v", removed, err)
	}
	third := filepath.Join(filepath.Dir(project), "third")
	if err := os.MkdirAll(third, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(third)
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"beta"}}); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Errorf("revoked source must hit the trust gate, got %v", err)
	}
	// the project that installed it keeps knowing it through its manifest
	t.Chdir(other)
	if _, err := addLocal(t, src, AddOptions{Skills: []string{"beta"}}); err != nil {
		t.Errorf("manifest-recorded source must stay known: %v", err)
	}
}

func TestTrustStore(t *testing.T) {
	setupEnv(t)
	for i := 0; i < 2; i++ {
		if err := TrustSource("https://example.com/a/b.git"); err != nil {
			t.Fatal(err)
		}
	}
	if err := TrustSource("/local/path"); err != nil {
		t.Fatal(err)
	}
	list, err := TrustedSources()
	if err != nil || len(list) != 2 || list[0].Source != "/local/path" || list[1].TrustedAt == "" {
		t.Fatalf("trusted sources %+v %v", list, err)
	}
	if removed, _ := UntrustSource("nope"); removed {
		t.Error("removing an unknown source must report false")
	}
	path, _ := TrustFilePath()
	if info, err := os.Stat(path); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o600) {
		t.Errorf("trust file mode: %v %v", info, err)
	}
	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := IsTrusted("x"); err == nil {
		t.Error("a corrupt trust file must be reported")
	} else if !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "delete the file to reset") {
		t.Errorf("corrupt trust file error must name the file and the recovery: %v", err)
	}
}
