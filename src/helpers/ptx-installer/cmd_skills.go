/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"golang.org/x/term"
	"portunix.ai/portunix/src/helpers/ptx-installer/engine"
	"portunix.ai/portunix/src/helpers/ptx-installer/engine/skills"
	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// Issue #190 / ADR-043: `portunix skills` installs AI-agent skills (SKILL.md
// bundles) from git repositories or local paths into .agents/skills/ and the
// per-agent skills directories

func init() {
	skills.SetDetector(detectAgentPackage)
}

var (
	agentStatusOnce sync.Once
	agentStatus     map[string]bool
)

// detectAgentPackage reuses the Issue #035 AI assistant detection
func detectAgentPackage(pkg string) bool {
	agentStatusOnce.Do(func() {
		agentStatus = map[string]bool{}
		// The registry loader reports progress on stdout; keep it out of
		// skills output (e.g. `skills sources --json`)
		stdout := os.Stdout
		if devNull, err := os.Open(os.DevNull); err == nil {
			os.Stdout = devNull
			defer devNull.Close()
		}
		reg, err := registry.LoadPackageRegistry("./assets")
		os.Stdout = stdout
		if err != nil {
			return
		}
		for _, s := range engine.DetectAIAssistants(reg, engine.GetOperatingSystem()) {
			agentStatus[s.Name] = s.Installed
		}
	})
	return agentStatus[pkg]
}

func handleSkills(args []string) {
	if len(args) == 0 {
		showSkillsHelp()
		return
	}
	sub, rest := args[0], args[1:]
	if sub != "help" && hasFlag(rest, "--help", "-h") {
		showSkillsHelp()
		return
	}
	var err error
	switch sub {
	case "add":
		err = handleSkillsAdd(rest)
	case "list", "ls":
		err = handleSkillsList(rest)
	case "remove", "rm":
		err = handleSkillsRemove(rest)
	case "update":
		err = handleSkillsUpdate(rest)
	case "sources":
		err = handleSkillsSources(rest)
	case "trust":
		err = handleSkillsTrust(rest)
	case "help", "--help", "-h":
		showSkillsHelp()
	case "--help-ai":
		showSkillsHelpAI()
	case "--help-expert":
		showSkillsHelpExpert()
	case "--description":
		fmt.Println("Install AI-agent skills (SKILL.md bundles) from git repositories")
	case "--list-commands":
		fmt.Println("add\nlist\nremove\nupdate\nsources\ntrust")
	default:
		err = fmt.Errorf("unknown skills command: %s (use 'portunix skills --help')", sub)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

// skillsFlags are the parsed options shared by the skills subcommands
type skillsFlags struct {
	positional  []string
	skills      []string
	agents      []string
	all         bool
	allAgents   bool
	noUniversal bool
	global      bool
	ref         string
	source      string
	yes         bool
	force       bool
	json        bool
}

// parseSkillsFlags accepts "--flag value" and "--flag=value" spellings
func parseSkillsFlags(args []string) (*skillsFlags, error) {
	f := &skillsFlags{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, value, hasValue := strings.Cut(arg, "=")
		takeValue := func() (string, error) {
			if hasValue {
				return value, nil
			}
			if i+1 >= len(args) {
				return "", fmt.Errorf("flag %s requires a value", name)
			}
			i++
			return args[i], nil
		}
		var err error
		switch name {
		case "--skill", "-s":
			var v string
			if v, err = takeValue(); err == nil {
				f.skills = append(f.skills, splitList(v)...)
			}
		case "--agent", "-a":
			var v string
			if v, err = takeValue(); err == nil {
				f.agents = append(f.agents, splitList(v)...)
			}
		case "--ref":
			f.ref, err = takeValue()
		case "--source":
			f.source, err = takeValue()
		case "--all":
			f.all = true
		case "--all-agents":
			f.allAgents = true
		case "--no-universal":
			f.noUniversal = true
		case "--global", "--user", "-g":
			f.global = true
		case "--yes", "-y":
			f.yes = true
		case "--force", "-f":
			f.force = true
		case "--json":
			f.json = true
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			f.positional = append(f.positional, arg)
		}
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (f *skillsFlags) scope() skills.Scope {
	if f.global {
		return skills.ScopeUser
	}
	return skills.ScopeProject
}

func splitList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasFlag(args []string, names ...string) bool {
	for _, a := range args {
		for _, n := range names {
			if a == n {
				return true
			}
		}
	}
	return false
}

func handleSkillsAdd(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	location := f.source
	switch {
	case location != "" && len(f.positional) > 0:
		return fmt.Errorf("pass either a source argument or --source, not both")
	case location == "" && len(f.positional) != 1:
		return fmt.Errorf("usage: portunix skills add <repo-url|path> [--skill <name>]... [--all]")
	case location == "":
		location = f.positional[0]
	}
	src, err := skills.ParseSource(location, f.ref)
	if err != nil {
		return err
	}
	if f.source != "" && src.Kind != skills.SourceLocal {
		return fmt.Errorf("--source %s is not an existing directory", f.source)
	}
	if f.all && len(f.skills) > 0 {
		return fmt.Errorf("--all and --skill are mutually exclusive")
	}
	prompter := stdinPrompter{}
	opts := skills.AddOptions{
		Source: src,
		Skills: f.skills,
		All:    f.all,
		Scope:  f.scope(),
		Yes:    f.yes,
		Force:  f.force,
		Prompt: prompter,
		Notify: func(msg string) { fmt.Println("◇  " + msg) },
	}
	if f.agents == nil && !f.allAgents && !f.yes && prompter.Interactive() {
		// agents are chosen after the skills, like the `skills` CLI does
		opts.SelectTargets = func() ([]skills.AgentTarget, error) {
			ids, err := chooseAgents(f.scope())
			if err != nil {
				return nil, err
			}
			return skills.ResolveTargets(ids, false, f.noUniversal)
		}
	} else if opts.Targets, err = skills.ResolveTargets(f.agents, f.allAgents, f.noUniversal); err != nil {
		return err
	}
	res, err := skills.Add(opts)
	if res != nil {
		printSkillsResult(res)
	}
	if err != nil {
		return err
	}
	return outcomeError(res.Conflicts(), res.Failures())
}

func handleSkillsUpdate(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	results, err := skills.Update(skills.UpdateOptions{
		Names:  f.positional,
		Scope:  f.scope(),
		Ref:    f.ref,
		Force:  f.force,
		Prompt: stdinPrompter{},
		Notify: func(msg string) { fmt.Println("◇  " + msg) },
	})
	conflicts, failures := 0, 0
	for _, r := range results {
		printSkillsResult(r)
		conflicts += r.Conflicts()
		failures += r.Failures()
	}
	if err != nil {
		return err
	}
	return outcomeError(conflicts, failures)
}

// outcomeError turns skipped or failed targets into a non-zero exit
func outcomeError(conflicts, failures int) error {
	var parts []string
	if failures > 0 {
		parts = append(parts, fmt.Sprintf("%d skill target(s) failed", failures))
	}
	if conflicts > 0 {
		parts = append(parts, fmt.Sprintf("%d skill target(s) skipped due to conflicts", conflicts))
	}
	if len(parts) == 0 {
		return nil
	}
	return errors.New(strings.Join(parts, ", "))
}

func printSkillsResult(res *skills.Result) {
	for _, w := range res.Warnings {
		fmt.Printf("⚠️  %s\n", w)
	}
	source := res.Source
	if res.Commit != "" {
		source += "@" + shortCommit(res.Commit)
	}
	fmt.Printf("\nSource: %s\n", source)
	for _, o := range res.Outcomes {
		icon := "✅"
		switch o.Action {
		case skills.ActionUnchanged:
			icon = "➖"
		case skills.ActionConflict:
			icon = "⚠️ "
		case skills.ActionFailed:
			icon = "❌"
		}
		line := fmt.Sprintf("  %s %-24s %-10s %s  %s", icon, o.Skill, o.Target, o.Action, o.Dir)
		if o.Reason != "" {
			line += "\n       " + o.Reason
		}
		fmt.Println(line)
	}
}

func shortCommit(c string) string {
	if len(c) > 12 {
		return c[:12]
	}
	return c
}

func handleSkillsList(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	entries, err := skills.List(f.scope())
	if err != nil {
		return err
	}
	if f.json {
		if entries == nil {
			entries = []skills.ListEntry{}
		}
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	if len(entries) == 0 {
		fmt.Printf("No skills installed in %s scope.\n", f.scope())
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTARGET\tSTATUS\tSOURCE\tCOMMIT")
	for _, e := range entries {
		source := e.Source
		if source == "" {
			source = "-"
		}
		commit := shortCommit(e.Commit)
		if commit == "" {
			commit = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.Name, e.Target, e.Status, source, commit)
	}
	return w.Flush()
}

func handleSkillsRemove(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	if len(f.positional) != 1 {
		return fmt.Errorf("usage: portunix skills remove <skill-name> [--agent <id>]... [--global]")
	}
	var targets []skills.AgentTarget
	for _, id := range f.agents {
		if id == skills.UniversalID {
			targets = append(targets, skills.Universal())
			continue
		}
		t, err := skills.LookupAgent(id)
		if err != nil {
			return err
		}
		targets = append(targets, t)
	}
	res, err := skills.Remove(f.positional[0], targets, f.scope())
	if err != nil {
		return err
	}
	for _, d := range res.Removed {
		fmt.Printf("🗑️  Removed %s\n", d)
	}
	for _, d := range res.Kept {
		fmt.Printf("⚠️  Kept %s (contains files not installed by Portunix)\n", d)
	}
	return nil
}

func handleSkillsSources(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	type row struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Detected    bool   `json:"detected"`
		ProjectDir  string `json:"projectDir"`
		UserDir     string `json:"userDir"`
	}
	var rows []row
	for _, t := range skills.AllTargets() {
		detected, _ := t.Detect()
		pd, _ := t.SkillsDir(skills.ScopeProject)
		ud, _ := t.SkillsDir(skills.ScopeUser)
		rows = append(rows, row{ID: t.ID(), DisplayName: t.DisplayName(), Detected: detected, ProjectDir: pd, UserDir: ud})
	}
	if f.json {
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDETECTED\tPROJECT DIR\tUSER DIR")
	for _, r := range rows {
		det := "no"
		if r.ID == skills.UniversalID {
			det = "always"
		} else if r.Detected {
			det = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.ID, r.DisplayName, det, r.ProjectDir, r.UserDir)
	}
	return w.Flush()
}

// handleSkillsTrust lists or revokes user-wide trusted sources (O4)
func handleSkillsTrust(args []string) error {
	f, err := parseSkillsFlags(args)
	if err != nil {
		return err
	}
	action := "list"
	if len(f.positional) > 0 {
		action = f.positional[0]
	}
	switch action {
	case "list", "ls":
		sources, err := skills.TrustedSources()
		if err != nil {
			return err
		}
		if f.json {
			if sources == nil {
				sources = []skills.TrustedSource{}
			}
			data, err := json.MarshalIndent(sources, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}
		path, _ := skills.TrustFilePath()
		if len(sources) == 0 {
			fmt.Printf("No trusted skill sources (%s).\n", path)
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SOURCE\tTRUSTED AT")
		for _, s := range sources {
			fmt.Fprintf(w, "%s\t%s\n", s.Source, s.TrustedAt)
		}
		return w.Flush()
	case "remove", "rm", "revoke":
		if len(f.positional) != 2 {
			return fmt.Errorf("usage: portunix skills trust remove <source>")
		}
		arg := f.positional[1]
		candidates := []string{arg}
		if src, err := skills.ParseSource(arg, ""); err == nil && src.Location != arg {
			candidates = append(candidates, src.Location)
		}
		for _, c := range candidates {
			removed, err := skills.UntrustSource(c)
			if err != nil {
				return err
			}
			if removed {
				fmt.Printf("🔒 Source no longer trusted: %s\n", c)
				fmt.Println("   Projects whose manifests record it still treat it as known.")
				return nil
			}
		}
		return fmt.Errorf("source %q is not in the trusted sources (see 'portunix skills trust list')", arg)
	default:
		return fmt.Errorf("unknown trust command: %s (use list or remove)", action)
	}
}

// stdinPrompter asks questions on the terminal; it is non-interactive when
// stdin is not a terminal (pipes, /dev/null, CI)
type stdinPrompter struct{}

var stdinReader = bufio.NewReader(os.Stdin)

func (stdinPrompter) Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) // #nosec G115 -- file descriptors fit in int
}

func (stdinPrompter) Confirm(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	ans, _ := stdinReader.ReadString('\n')
	ans = strings.ToLower(strings.TrimSpace(ans))
	return ans == "y" || ans == "yes"
}

func (stdinPrompter) ChooseSkills(available []skills.ResolvedSkill) ([]string, error) {
	fmt.Println("\nAvailable skills:")
	for i, s := range available {
		fmt.Printf("  %2d) %-28s %s\n", i+1, s.Name, truncate(s.Description, 70))
	}
	fmt.Print("\nSelect skills (e.g. 1,3 or 'all'): ")
	ans, _ := stdinReader.ReadString('\n')
	picked, err := parseSelection(ans, len(available), nil)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(picked))
	for i, n := range picked {
		names[i] = available[n].Name
	}
	return names, nil
}

// chooseAgents offers the agent drivers on the terminal, pre-selecting the
// detected ones (Enter keeps the pre-selection). The universal target is
// controlled by --no-universal and not offered here.
func chooseAgents(scope skills.Scope) ([]string, error) {
	agents := skills.Agents()
	var detected []int
	fmt.Println("\nWhich agents do you want to install to?")
	fmt.Println("  ── Universal (.agents/skills) ── always included ──")
	for i, a := range agents {
		mark := " "
		if ok, _ := a.Detect(); ok {
			mark = "x"
			detected = append(detected, i)
		}
		dir, _ := a.SkillsDir(scope)
		fmt.Printf("  [%s] %2d) %-16s %s\n", mark, i+1, a.DisplayName(), dir)
	}
	fmt.Print("\nSelect agents (Enter = marked, numbers e.g. 1,3, 'all' or 'none'): ")
	ans, _ := stdinReader.ReadString('\n')
	picked, err := parseSelection(ans, len(agents), detected)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(picked))
	for _, n := range picked {
		ids = append(ids, agents[n].ID())
	}
	return ids, nil
}

// parseSelection parses a numbered selection ("1,3", "2 4", "all", "none")
// into zero-based indexes; an empty answer returns defaults
func parseSelection(answer string, count int, defaults []int) ([]int, error) {
	answer = strings.TrimSpace(answer)
	switch strings.ToLower(answer) {
	case "":
		return append([]int{}, defaults...), nil
	case "all":
		all := make([]int, count)
		for i := range all {
			all[i] = i
		}
		return all, nil
	case "none":
		return []int{}, nil
	}
	var picked []int
	seen := map[int]bool{}
	for _, p := range splitList(strings.ReplaceAll(answer, " ", ",")) {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > count {
			return nil, fmt.Errorf("invalid selection %q (expected 1-%d)", p, count)
		}
		if !seen[n] {
			seen[n] = true
			picked = append(picked, n-1)
		}
	}
	return picked, nil
}

// truncate shortens s to at most max runes
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

func showSkillsHelp() {
	fmt.Println(`Usage: portunix skills <command> [options]

Install AI-agent skills (directories with SKILL.md) from git repositories or
local paths. Each skill is written to the universal .agents/skills/ directory
and to the skills directory of the selected agents (e.g. .claude/skills/).

Commands:
  add <repo-url|owner/repo|path>   Install skills from a source
  list                             List installed skills (managed and unmanaged)
  remove <skill-name>              Remove a Portunix-installed skill
  update [<skill-name>...]         Re-install managed skills from their recorded source
  sources                          Show install targets and agent detection state
  trust [list|remove <source>]     Show or revoke user-wide trusted sources

Options for add:
  --skill, -s <name>    Skill to install (repeatable, comma separated)
  --all                 Install every discovered skill
  --agent, -a <id>      Also install for this agent (repeatable); without it a terminal
                        offers a selection pre-filled with the detected agents, a
                        non-interactive run uses the detected agents
  --all-agents          Install for every supported agent
  --no-universal        Skip the universal .agents/skills/ directory
  --ref <ref>           Branch, tag or commit SHA (default: repository default branch)
  --source <path>       Install from a local checkout (air-gapped)
  --global, --user, -g  Install under the home directory instead of the project
  --yes, -y             Trust a new source without asking
  --force, -f           Overwrite skills that differ from the source

Agents (portunix skills sources shows the directories and detection):
  amp, claude (claude-code), copilot (github-copilot), cursor, gemini (gemini-cli),
  opencode, zed. In a project all of them except claude read .agents/skills, so
  they share the universal directory; their user-scope directories differ.

Options for list / remove / update:
  --global, -g          Use the user scope
  --json                JSON output (list, sources)
  --agent <id>          Limit remove to this target (universal or an agent ID)
  --ref <ref>           Update to a different ref
  --force               Overwrite locally modified skills on update

Examples:
  portunix skills add https://github.com/obra/superpowers --skill brainstorming
  portunix skills add obra/superpowers --all --agent claude --yes
  portunix skills add --source ./my-skills --skill reviewer --global
  portunix skills list --json
  portunix skills update
  portunix skills remove brainstorming

Skills are instructions your AI agents will follow. The first install from a
new source asks for confirmation; review untrusted repositories before use.
Confirmed sources (or sources installed with --yes) are trusted user-wide and
not asked for again in other projects; 'skills trust remove' revokes that.`)
}

func showSkillsHelpAI() {
	fmt.Println(`portunix skills — AI-agent skill installer (ptx-installer, ADR-043)

Non-interactive usage (always pass --yes and --skill/--all, stdin is not a TTY):
  portunix skills add <git-url|owner/repo|path> --skill <name> --yes [--agent claude] [--global]
  portunix skills list --json [--global]
  portunix skills update [<name>] [--global] [--force]
  portunix skills remove <name> [--agent <id>] [--global]
  portunix skills sources --json

Targets: universal (.agents/skills, always unless --no-universal) and agents
amp, claude, copilot, cursor, gemini, opencode, zed (--agent <id>, --all-agents).
Without --agent and without a TTY the detected agents are used. Targets sharing a
directory are written once. Run 'portunix skills sources --json' for the paths.
Scope: project = current directory (default), user = home directory (--global).
Provenance: <skills-dir>/.portunix-skills.json; only recorded skills are updated/removed.
Trust: sources installed once are trusted user-wide (portunix skills trust list --json).
Exit code is non-zero when any target was skipped due to a conflict or failed.`)
}

func showSkillsHelpExpert() {
	showSkillsHelp()
	fmt.Println(`
Expert details:
  Cache:        <user-cache>/portunix/skills/<host>/<owner>/<repo>/<ref>/
                (override with PORTUNIX_SKILLS_CACHE); add reuses the cache,
                update always re-fetches
  Discovery:    every directory containing SKILL.md; front-matter name/description
  Manifest:     .portunix-skills.json per skills directory (source, ref, commit,
                files, sha256 checksum)
  Idempotency:  identical content is a no-op; a changed managed skill is updated;
                unmanaged, locally modified or foreign-source skills need
                confirmation or --force
  Trust gate:   first write from a source that is neither in the user-wide trust
                record nor in a manifest of the project / user scope asks for
                confirmation; --yes bypasses it. After a successful install the
                source is added to <user-config>/portunix/skills-trusted-sources.json
                (override with PORTUNIX_SKILLS_TRUST)`)
}
