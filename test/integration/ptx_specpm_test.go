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

// TestIssue183_PtxSpecpm_Phase1 exercises the end-to-end Phase 1 surface of
// ptx-specpm against a local kit checkout (--source <path>). It avoids
// network calls so the test is hermetic and can run in any environment.
//
// Run with:
//
//	go test ./test/integration/ptx_specpm_test.go -v -timeout 5m
func TestIssue183_PtxSpecpm_Phase1(t *testing.T) {
	tf := testframework.NewTestFramework("Issue183_PtxSpecpm_Phase1")
	tf.Start(t, "ptx-specpm Phase 1: init / check / version / integration list against a local kit")

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

	// 1. version (via dispatcher)
	tf.Step(t, "portunix specpm version  (dispatcher form)")
	out, err := runCmd(binary, "specpm", "version")
	if err != nil {
		tf.Error(t, "version subcommand failed", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, "portunix specpm version") {
		tf.Error(t, "expected 'portunix specpm version' brand", out)
		success = false
	}
	if !strings.Contains(out, "kit (default pinned ref)") {
		tf.Error(t, "version output missing kit pinned ref line", out)
		success = false
	}
	tf.Success(t, "version subcommand returned both helper and kit ref")

	// 2. integration list
	tf.Step(t, "portunix specpm integration list")
	out, err = runCmd(binary, "specpm", "integration", "list")
	if err != nil {
		tf.Error(t, "integration list failed", err.Error()+"\n"+out)
		success = false
		return
	}
	if !strings.Contains(out, "claude") {
		tf.Error(t, "integration list expected to include 'claude'", out)
		success = false
	}
	tf.Success(t, "integration list reports claude driver")

	tf.Separator()

	// 3. init in an empty target directory using --source <path>
	target, err := os.MkdirTemp("", "ptx-specpm-int-")
	if err != nil {
		tf.Error(t, "tempdir failed", err.Error())
		success = false
		return
	}
	t.Cleanup(func() { _ = os.RemoveAll(target) })

	tf.Step(t, "portunix specpm init <tmp> --integration claude --source <kitFixture>")
	out, err = runCmd(binary, "specpm", "init", target,
		"--integration", "claude",
		"--source", kitDir,
		"--ignore-agent-tools",
	)
	if err != nil {
		tf.Error(t, "init failed", err.Error()+"\n"+out)
		success = false
		return
	}
	tf.Success(t, "init completed")

	// 4. Scaffold layout assertions (ADR-040 D1).
	tf.Step(t, "Verify ADR-040 split scaffold layout")
	mustExist(t, tf, &success, target, ".specpm", "templates", "project", "charter.md")
	mustExist(t, tf, &success, target, ".specpm", "agents", "pm.md")
	mustExist(t, tf, &success, target, ".specpm", "workflows", "project-initiation.md")
	mustExist(t, tf, &success, target, ".specpm", "memory", "governance.md")
	mustExist(t, tf, &success, target, ".specpm", "kit.json")
	for _, verb := range []string{"init", "spec", "plan", "risks", "decisions", "review", "report", "check"} {
		mustExist(t, tf, &success, target, ".claude", "commands", "specpm."+verb+".md")
	}
	mustExist(t, tf, &success, target, "CLAUDE.md")
	// Project ID is derived from target dir basename — verify the artifact dir.
	pid := filepath.Base(target)
	mustExist(t, tf, &success, target, "specs", "project", pid, "project.md")
	mustExist(t, tf, &success, target, "specs", "project", pid, "execution")
	mustExist(t, tf, &success, target, "specs", "project", pid, "validation")

	tf.Separator()

	// 5. Idempotent re-run: user-edited Charter must NOT be overwritten.
	charter := filepath.Join(target, "specs", "project", pid, "project.md")
	const userTouched = "# user-touched Charter — sacred\n"
	if err := os.WriteFile(charter, []byte(userTouched), 0o644); err != nil {
		tf.Error(t, "could not tamper Charter for idempotency check", err.Error())
		success = false
		return
	}

	tf.Step(t, "Re-run init: user Charter must survive")
	out, err = runCmd(binary, "specpm", "init", target,
		"--integration", "claude",
		"--source", kitDir,
		"--ignore-agent-tools",
	)
	if err != nil {
		tf.Error(t, "second init failed", err.Error()+"\n"+out)
		success = false
		return
	}
	body, _ := os.ReadFile(charter)
	if string(body) != userTouched {
		tf.Error(t, "user Charter was overwritten on idempotent re-run", string(body))
		success = false
	} else {
		tf.Success(t, "user Charter preserved (artifacts are sacred)")
	}

	tf.Separator()

	// 6. check from inside the project.
	tf.Step(t, "portunix specpm check (from inside the project)")
	cmd := exec.Command(binary, "specpm", "check")
	cmd.Dir = target
	cmdOut, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		tf.Error(t, "check failed", cmdErr.Error()+"\n"+string(cmdOut))
		success = false
		return
	}
	checkOut := string(cmdOut)
	if !strings.Contains(checkOut, "Project          : initialised") {
		tf.Error(t, "check did not report initialised project", checkOut)
		success = false
	}
	if !strings.Contains(checkOut, "Integration      : claude") {
		tf.Error(t, "check did not report claude integration", checkOut)
		success = false
	}
	tf.Success(t, "check reports initialised state and wired integration")

	tf.Separator()

	// 7. Skills mode in a fresh target.
	skillsTarget, err := os.MkdirTemp("", "ptx-specpm-skills-")
	if err != nil {
		tf.Error(t, "tempdir failed", err.Error())
		success = false
		return
	}
	t.Cleanup(func() { _ = os.RemoveAll(skillsTarget) })

	tf.Step(t, "init with --integration-options=\"--skills\"")
	out, err = runCmd(binary, "specpm", "init", skillsTarget,
		"--integration", "claude",
		"--integration-options", "--skills",
		"--source", kitDir,
		"--ignore-agent-tools",
	)
	if err != nil {
		tf.Error(t, "skills-mode init failed", err.Error()+"\n"+out)
		success = false
		return
	}
	for _, verb := range []string{"init", "spec", "plan", "risks", "decisions", "review", "report", "check"} {
		mustExist(t, tf, &success, skillsTarget, ".claude", "skills", "specpm-"+verb, "SKILL.md")
	}
	tf.Success(t, "skills bundle materialised under .claude/skills/specpm-*/")
}

// repoRoot finds the git repository root from the current test binary's CWD.
func repoRoot(t *testing.T, tf *testframework.TestFramework) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("could not locate repo root from %s", dir)
	return ""
}

// buildKitFixture writes a minimal-but-valid spec-kit-pm-shaped tree under
// a tempdir. We avoid pulling the real upstream so the test is hermetic and
// covers ADR-040 D3's --source <path> path explicitly.
func buildKitFixture(t *testing.T, tf *testframework.TestFramework) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "specpm-kit-")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	mk := func(rel, body string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mk("templates/project/charter.md", "# Charter template (fixture)\n")
	mk("templates/roadmap/gantt.md", "# Gantt template (fixture)\n")
	mk("agents/pm.md", "# PM agent (fixture)\n")
	mk("agents/reviewer.md", "# Reviewer (fixture)\n")
	mk("agents/risk-analyst.md", "# Risk analyst (fixture)\n")
	mk("agents/reporter.md", "# Reporter (fixture)\n")
	mk("workflows/project-initiation.md", "# Initiation workflow (fixture)\n")
	mk("workflows/spec-to-plan.md", "# Spec to plan (fixture)\n")
	mk("workflows/execution-loop.md", "# Execution loop (fixture)\n")
	mk("profiles/example-company/identity.yml", "name: example\n")

	tf.Info(t, "Kit fixture rooted at "+dir)
	return dir
}

func runCmd(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustExist(t *testing.T, tf *testframework.TestFramework, success *bool, parts ...string) {
	t.Helper()
	p := filepath.Join(parts...)
	if _, err := os.Stat(p); err != nil {
		tf.Error(t, "expected path missing", p)
		*success = false
		return
	}
	tf.Info(t, "  ✓ "+p)
}
