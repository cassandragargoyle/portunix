//go:build integration
// +build integration

/*
 *  SPDX-License-Identifier: MIT
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

// Issue #190: `portunix skills` exercised end-to-end in a clean Ubuntu
// container managed through `portunix container`. Covers local and git
// sources, cache reuse, multi-agent targets, conflicts, failures and the
// interactive prompts on a pseudo-terminal (util-linux `script`).

const (
	issue190Container = "portunix-test-issue190"
	issue190Image     = "ubuntu:24.04"
	issue190Bin       = "/opt/ptx/portunix"
)

// issue190Fixture creates /work/fixture with skills alpha and beta
const issue190Fixture = `set -e
rm -rf /work && mkdir -p /work/fixture/skills/alpha/refs /work/fixture/skills/beta
printf -- '---\nname: alpha\ndescription: Alpha test skill\n---\n# Alpha\n' > /work/fixture/skills/alpha/SKILL.md
echo notes > /work/fixture/skills/alpha/refs/notes.md
printf -- '---\nname: beta\ndescription: Beta test skill\n---\n' > /work/fixture/skills/beta/SKILL.md
chmod -R a+rX /work`

type issue190Case struct {
	name   string
	script string
	// want are substrings expected in the combined output
	want []string
	// wantFail expects a non-zero exit of the script
	wantFail bool
	// network marks cases that need GitHub access; they are skipped on clone failure
	network bool
}

func TestIssue190_SkillsInContainer(t *testing.T) {
	tf := testframework.NewTestFramework("Issue190_Skills_Container")
	tf.Start(t, "portunix skills add/list/update/remove in a clean Ubuntu container")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	if runtime.GOOS != "linux" {
		t.Skip("container test builds Linux binaries in place")
	}
	binaryPath, ok := tf.VerifyPortunixBinary(t)
	if !ok {
		success = false
		return
	}
	helperPath := filepath.Join(filepath.Dir(binaryPath), "ptx-installer")
	if _, err := os.Stat(helperPath); err != nil {
		tf.Error(t, "ptx-installer helper not found next to portunix (run 'make build')", err.Error())
		success = false
		return
	}

	ptx := func(args ...string) (string, error) {
		out, err := exec.Command(binaryPath, args...).CombinedOutput()
		return string(out), err
	}
	inContainer := func(script string) (string, error) {
		return ptx("container", "exec", issue190Container, "bash", "-c", script)
	}

	tf.Step(t, "Start container "+issue190Image)
	_, _ = ptx("container", "rm", "-f", issue190Container)
	if out, err := ptx("container", "run", "-d", "--name", issue190Container, issue190Image, "sleep", "3600"); err != nil {
		tf.Warning(t, "No usable container runtime — skipping", out)
		t.Skip("container runtime unavailable")
		return
	}
	defer func() { _, _ = ptx("container", "rm", "-f", issue190Container) }()

	tf.Step(t, "Copy binaries and install git")
	if out, err := inContainer("mkdir -p /opt/ptx"); err != nil {
		tf.Error(t, "mkdir failed", out)
		success = false
		return
	}
	for _, f := range []string{binaryPath, helperPath} {
		if out, err := ptx("container", "cp", f, issue190Container+":/opt/ptx/"+filepath.Base(f)); err != nil {
			tf.Error(t, "copy failed", out)
			success = false
			return
		}
	}
	if out, err := inContainer("apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq git ca-certificates >/dev/null && git --version"); err != nil {
		tf.Error(t, "git installation in container failed", out)
		success = false
		return
	}
	if out, err := inContainer(issue190Fixture); err != nil {
		tf.Error(t, "fixture creation failed", out)
		success = false
		return
	}
	tf.Success(t, "Container ready")

	p := issue190Bin
	cases := []issue190Case{
		{
			name:   "TC01 routing and help",
			script: p + " skills --help",
			want:   []string{"Usage: portunix skills"},
		},
		{
			name:     "TC02 no TTY and no --skill fails",
			script:   "mkdir -p /work/p1 && cd /work/p1 && " + p + " skills add --source /work/fixture </dev/null",
			want:     []string{"pass --skill <name> or --all"},
			wantFail: true,
		},
		{
			name:     "TC03 trust gate without --yes",
			script:   "cd /work/p1 && " + p + " skills add --source /work/fixture --skill alpha </dev/null; rc=$?; test ! -e .agents && echo NOTHING_WRITTEN; exit $rc",
			want:     []string{"re-run with --yes", "NOTHING_WRITTEN"},
			wantFail: true,
		},
		{
			name: "TC04 add local source to universal + claude",
			script: "cd /work/p1 && " + p + " skills add --source /work/fixture --skill alpha --agent claude --yes </dev/null" +
				" && test -f .agents/skills/alpha/refs/notes.md && test -f .claude/skills/alpha/SKILL.md && test -f .agents/skills/.portunix-skills.json",
			want: []string{"installed"},
		},
		{
			name:   "TC05 re-add is a no-op",
			script: "cd /work/p1 && " + p + " skills add --source /work/fixture --skill alpha --agent claude </dev/null",
			want:   []string{"unchanged"},
		},
		{
			name: "TC06 local edit is a conflict, --force overwrites",
			script: "cd /work/p1 && echo edit >> .claude/skills/alpha/SKILL.md && echo v2 >> /work/fixture/skills/alpha/SKILL.md" +
				" && { " + p + " skills update </dev/null; test $? -ne 0; } && grep -q edit .claude/skills/alpha/SKILL.md" +
				" && " + p + " skills update --force </dev/null && ! grep -q edit .claude/skills/alpha/SKILL.md",
			want: []string{"modified locally", "updated"},
		},
		{
			name: "TC07 all agents share .agents/skills in a project",
			script: "mkdir -p /work/p2 && cd /work/p2 && " + p + " skills add --source /work/fixture --skill beta --all-agents --yes </dev/null" +
				" && " + p + " skills list --json </dev/null | grep -c '\"name\": \"beta\"' | grep -qx 2",
			want: []string{"universal,amp,copilot,cursor,gemini,opencode,zed", "claude"},
		},
		{
			name: "TC08 user scope per-agent directories",
			script: "cd /work/p2 && " + p + " skills add --source /work/fixture --skill beta --global --agent cursor,opencode,amp --yes </dev/null" +
				" && test -f /root/.agents/skills/beta/SKILL.md && test -f /root/.cursor/skills/beta/SKILL.md" +
				" && test -f /root/.config/opencode/skills/beta/SKILL.md && test -f /root/.config/agents/skills/beta/SKILL.md",
			want: []string{"installed"},
		},
		{
			name: "TC09 remove keeps foreign files and refuses unmanaged skills",
			script: "cd /work/p1 && echo mine > .agents/skills/alpha/mine.md && " + p + " skills remove alpha </dev/null" +
				" && test -f .agents/skills/alpha/mine.md && test ! -e .claude/skills/alpha" +
				" && mkdir -p .agents/skills/hand && echo x > .agents/skills/hand/SKILL.md" +
				" && { " + p + " skills remove hand </dev/null; test $? -ne 0; } && test -f .agents/skills/hand/SKILL.md",
			want: []string{"Kept", "not managed by Portunix"},
		},
		{
			name: "TC10 write error reported as failed (O1)",
			script: "mkdir -p /work/ro/.agents && chmod 555 /work/ro/.agents && cd /work/ro" +
				" && su nobody -s /bin/bash -c 'HOME=/tmp PORTUNIX_SKILLS_CACHE=/tmp/c " + p + " skills add --source /work/fixture --skill beta --yes' </dev/null",
			want:     []string{"failed", "permission denied"},
			wantFail: true,
		},
		{
			name: "TC11 interactive skill and agent selection with trust prompt (O2)",
			script: "cp -r /work/fixture /work/fixture-tty && mkdir -p /work/tty && cd /work/tty && printf '2\\n\\ny\\n' | script -qec '" + p +
				" skills add --source /work/fixture-tty' /dev/null && test -f .agents/skills/beta/SKILL.md && test ! -e .agents/skills/alpha",
			want: []string{"Available skills:", "Which agents do you want to install to?", "First install from a new source", "Proceed?", "installed"},
		},
		{
			name: "TC12 declined trust prompt writes nothing (O2)",
			script: "cp -r /work/fixture /work/fixture-copy && mkdir -p /work/tty2 && cd /work/tty2 && printf 'n\\n' | script -qec '" + p +
				" skills add /work/fixture-copy --skill alpha --agent claude' /dev/null; test ! -e .agents && test ! -e .claude && echo NOTHING_WRITTEN",
			want: []string{"Proceed?", "installation cancelled", "NOTHING_WRITTEN"},
		},
		{
			name: "TC13 interactive overwrite prompt (O2)",
			script: "cd /work/tty && echo edit >> .agents/skills/beta/SKILL.md && echo v3 >> /work/fixture-tty/skills/beta/SKILL.md" +
				" && printf 'y\\n' | script -qec '" + p + " skills update' /dev/null && grep -q v3 .agents/skills/beta/SKILL.md && ! grep -q edit .agents/skills/beta/SKILL.md",
			want: []string{"Overwrite?", "updated"},
		},
		{
			name: "TC14 git source cloned, then served from cache",
			script: "mkdir -p /work/git && cd /work/git && " + p + " skills add https://github.com/obra/superpowers --skill brainstorming --yes </dev/null" +
				" && https_proxy=http://127.0.0.1:9 " + p + " skills add obra/superpowers --skill writing-plans </dev/null" +
				" && test -f .agents/skills/writing-plans/SKILL.md",
			want:    []string{"Repository cloned", "Repository served from cache"},
			network: true,
		},
		{
			name: "TC16 trusted source needs no confirmation in another project (O4)",
			script: "cp -r /work/fixture /work/fixture-o4 && mkdir -p /work/o4a /work/o4b /work/o4c" +
				" && cd /work/o4a && " + p + " skills add /work/fixture-o4 --skill beta --no-universal --agent claude --yes </dev/null" +
				" && cd /work/o4b && " + p + " skills add /work/fixture-o4 --skill beta --no-universal --agent claude </dev/null" +
				" && test -f .claude/skills/beta/SKILL.md && " + p + " skills trust list </dev/null" +
				" && " + p + " skills trust remove /work/fixture-o4 </dev/null" +
				" && cd /work/o4c && { " + p + " skills add /work/fixture-o4 --skill beta --no-universal --agent claude </dev/null; test $? -ne 0; } && test ! -e .claude",
			want: []string{"/work/fixture-o4", "no longer trusted", "re-run with --yes"},
		},
		{
			name:     "TC15 missing git gives a remediation hint",
			script:   "mv /usr/bin/git /usr/bin/git.off; cd /work/git; " + p + " skills add https://github.com/other/repo --all --yes </dev/null; rc=$?; mv /usr/bin/git.off /usr/bin/git; exit $rc",
			want:     []string{"portunix install git"},
			wantFail: true,
		},
	}

	for _, c := range cases {
		tf.Separator()
		tf.Step(t, c.name)
		out, err := inContainer(c.script)
		tf.Output(t, out, 1500)
		if c.network && err != nil && strings.Contains(out, "git clone of") {
			tf.Warning(t, "GitHub not reachable from the container — case skipped")
			continue
		}
		failed := err != nil
		if failed != c.wantFail {
			tf.Error(t, "unexpected exit status", "wantFail="+boolString(c.wantFail), "err="+errString(err))
			success = false
			continue
		}
		missing := ""
		for _, w := range c.want {
			if !strings.Contains(out, w) {
				missing = w
				break
			}
		}
		if missing != "" {
			tf.Error(t, "expected output not found", missing)
			success = false
			continue
		}
		tf.Success(t, c.name)
	}
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func errString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
