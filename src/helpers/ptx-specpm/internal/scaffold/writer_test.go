/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeKit builds a minimal kit tree at root resembling spec-kit-pm.
func fakeKit(t *testing.T, root string) {
	t.Helper()
	tpl := filepath.Join(root, "templates", "project")
	if err := os.MkdirAll(tpl, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tpl, "charter.md"), []byte("# Charter template\n"), 0o644); err != nil {
		t.Fatalf("write charter template: %v", err)
	}
	for _, sub := range []string{"agents", "workflows"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, sub, "stub.md"), []byte("stub\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWrite_BootstrapAndIdempotent(t *testing.T) {
	target := t.TempDir()
	kit := t.TempDir()
	fakeKit(t, kit)

	res, err := Write(Options{TargetDir: target, KitDir: kit, ProjectID: "demo"})
	if err != nil {
		t.Fatalf("first write: %v", err)
	}
	if res.SkippedNoOp {
		t.Fatalf("first write should not be a no-op")
	}

	charter := filepath.Join(target, "specs", "project", "demo", "project.md")
	if _, err := os.Stat(charter); err != nil {
		t.Fatalf("expected Charter placeholder at %s: %v", charter, err)
	}
	governance := filepath.Join(target, ".specpm", "memory", "governance.md")
	if _, err := os.Stat(governance); err != nil {
		t.Fatalf("expected governance memory at %s: %v", governance, err)
	}

	// Tamper with the user Charter — second run must NOT overwrite it.
	const userTouched = "# my own Charter — sacred\n"
	if err := os.WriteFile(charter, []byte(userTouched), 0o644); err != nil {
		t.Fatal(err)
	}

	res2, err := Write(Options{TargetDir: target, KitDir: kit, ProjectID: "demo"})
	if err != nil {
		t.Fatalf("second write: %v", err)
	}
	if !res2.SkippedNoOp {
		t.Fatalf("idempotent re-run expected to be no-op, got %#v", res2)
	}

	got, _ := os.ReadFile(charter)
	if string(got) != userTouched {
		t.Fatalf("user Charter was overwritten on idempotent re-run; got %q", string(got))
	}
}

func TestWrite_ForceRefreshesKitButNotArtifacts(t *testing.T) {
	target := t.TempDir()
	kit := t.TempDir()
	fakeKit(t, kit)

	if _, err := Write(Options{TargetDir: target, KitDir: kit, ProjectID: "demo"}); err != nil {
		t.Fatal(err)
	}

	// User edits the Charter.
	charter := filepath.Join(target, "specs", "project", "demo", "project.md")
	const userTouched = "# user content\n"
	if err := os.WriteFile(charter, []byte(userTouched), 0o644); err != nil {
		t.Fatal(err)
	}

	// Bump the kit's Charter template; with --force the kit cache refreshes...
	tpl := filepath.Join(kit, "templates", "project", "charter.md")
	if err := os.WriteFile(tpl, []byte("# Charter template v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Write(Options{TargetDir: target, KitDir: kit, ProjectID: "demo", Force: true}); err != nil {
		t.Fatal(err)
	}

	// Kit cache template was refreshed.
	cached := filepath.Join(target, ".specpm", "templates", "project", "charter.md")
	got, _ := os.ReadFile(cached)
	if string(got) != "# Charter template v2\n" {
		t.Fatalf("kit cache not refreshed under --force; got %q", string(got))
	}

	// User Charter is still the user's.
	got, _ = os.ReadFile(charter)
	if string(got) != userTouched {
		t.Fatalf("user Charter overwritten under --force; got %q", string(got))
	}
}

func TestWrite_Refresh_WipesKitButPreservesCarveOutsAndArtifacts(t *testing.T) {
	target := t.TempDir()
	kit := t.TempDir()
	fakeKit(t, kit)

	// Initial bootstrap (init mode).
	if _, err := Write(Options{TargetDir: target, KitDir: kit, ProjectID: "demo"}); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	specpm := filepath.Join(target, ".specpm")

	// Plant a stale kit file to ensure refresh removes it.
	stale := filepath.Join(specpm, "templates", "stale-removed-upstream.md")
	if err := os.WriteFile(stale, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Plant carve-out content (extensions/, presets/) that MUST survive.
	extPath := filepath.Join(specpm, "extensions", "my-ext", "marker.txt")
	if err := os.MkdirAll(filepath.Dir(extPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extPath, []byte("ext-marker\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	presetPath := filepath.Join(specpm, "presets", "my-preset", "marker.txt")
	if err := os.MkdirAll(filepath.Dir(presetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(presetPath, []byte("preset-marker\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Plant user artifact under specs/project/<id>/ that MUST survive.
	userArtifact := filepath.Join(target, "specs", "project", "demo", "USER.md")
	if err := os.WriteFile(userArtifact, []byte("user-sentinel\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Bump the upstream Charter template so the refresh has something to do.
	tpl := filepath.Join(kit, "templates", "project", "charter.md")
	if err := os.WriteFile(tpl, []byte("# Charter v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Write(Options{TargetDir: target, KitDir: kit, Refresh: true})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(res.WipedSubtrees) == 0 {
		t.Fatalf("refresh expected to wipe subtrees, got %#v", res)
	}

	// Stale kit file is gone.
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale kit file should have been wiped; stat err=%v", err)
	}

	// Refreshed Charter template has v2 content.
	cached := filepath.Join(specpm, "templates", "project", "charter.md")
	got, _ := os.ReadFile(cached)
	if string(got) != "# Charter v2\n" {
		t.Fatalf("Charter template not refreshed; got %q", string(got))
	}

	// Carve-outs survived.
	if got, _ := os.ReadFile(extPath); string(got) != "ext-marker\n" {
		t.Fatalf("extensions/ carve-out lost on refresh; got %q", string(got))
	}
	if got, _ := os.ReadFile(presetPath); string(got) != "preset-marker\n" {
		t.Fatalf("presets/ carve-out lost on refresh; got %q", string(got))
	}

	// User artifact under specs/project/ survived.
	if got, _ := os.ReadFile(userArtifact); string(got) != "user-sentinel\n" {
		t.Fatalf("user artifact under specs/project/ overwritten on refresh; got %q", string(got))
	}
}

func TestAugmentClaudeMD_Idempotent(t *testing.T) {
	target := t.TempDir()
	const include = "@.specpm/memory/governance.md"

	path, changed, err := AugmentClaudeMD(target)
	if err != nil {
		t.Fatalf("first augment: %v", err)
	}
	if !changed {
		t.Fatalf("first augment expected to create CLAUDE.md")
	}
	body, _ := os.ReadFile(path)
	if !contains(string(body), include) {
		t.Fatalf("CLAUDE.md missing include line; got %q", string(body))
	}

	// Re-run is a no-op.
	_, changed2, err := AugmentClaudeMD(target)
	if err != nil {
		t.Fatalf("second augment: %v", err)
	}
	if changed2 {
		t.Fatalf("second augment should be no-op")
	}

	// Pre-existing CLAUDE.md without the include gets the line appended.
	other := filepath.Join(t.TempDir(), "")
	if err := os.WriteFile(filepath.Join(other, "CLAUDE.md"), []byte("# user content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, changed3, err := AugmentClaudeMD(other)
	if err != nil {
		t.Fatalf("append augment: %v", err)
	}
	if !changed3 {
		t.Fatalf("append augment expected to write")
	}
	body, _ = os.ReadFile(filepath.Join(other, "CLAUDE.md"))
	if !contains(string(body), "# user content") {
		t.Fatalf("user content lost; got %q", string(body))
	}
	if !contains(string(body), include) {
		t.Fatalf("include line missing after append; got %q", string(body))
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
