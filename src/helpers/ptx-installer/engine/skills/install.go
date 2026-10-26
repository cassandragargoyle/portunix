/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Prompter is the interactive side of the orchestration. A nil Prompter or
// Interactive() == false means no questions can be asked.
type Prompter interface {
	Interactive() bool
	Confirm(question string) bool
	// ChooseSkills lets the user pick from the discovered skills
	ChooseSkills(available []ResolvedSkill) ([]string, error)
}

// Action is what happens to one skill at one target
type Action string

const (
	ActionInstall   Action = "installed"
	ActionUpdate    Action = "updated"
	ActionUnchanged Action = "unchanged"
	ActionConflict  Action = "skipped (conflict)"
	ActionFailed    Action = "failed"
)

// Outcome reports the result for one skill in one skills directory
type Outcome struct {
	Skill string
	// Target lists the IDs of the targets sharing the directory
	Target string
	Dir    string
	Action Action
	Reason string
}

// Result summarizes an add or update run
type Result struct {
	Source    string
	Commit    string
	FromCache bool
	Warnings  []string
	Outcomes  []Outcome
}

// Conflicts returns the number of skipped conflicting writes
func (r *Result) Conflicts() int { return r.count(ActionConflict) }

// Failures returns the number of writes that failed
func (r *Result) Failures() int { return r.count(ActionFailed) }

// placed reports whether any skill was installed, updated or already present
func (r *Result) placed() bool {
	return r.count(ActionInstall)+r.count(ActionUpdate)+r.count(ActionUnchanged) > 0
}

func (r *Result) count(a Action) int {
	n := 0
	for _, o := range r.Outcomes {
		if o.Action == a {
			n++
		}
	}
	return n
}

// AddOptions configures `skills add`
type AddOptions struct {
	Source  Source
	Skills  []string
	All     bool
	Targets []AgentTarget
	// SelectTargets, when set, is called after skill selection to choose
	// the targets interactively; it replaces Targets
	SelectTargets func() ([]AgentTarget, error)
	Scope         Scope
	// Yes bypasses the trust gate for a previously unseen source (D6)
	Yes bool
	// Force overwrites skills that differ from the source without asking (D4)
	Force  bool
	Prompt Prompter
	// Notify receives progress messages (may be nil)
	Notify func(msg string)
}

// job is one skill to place into one skills directory
type job struct {
	dir   TargetDir
	skill ResolvedSkill
}

// plannedJob is a job with its resolved destination and decided action
type plannedJob struct {
	job
	dest     string
	checksum string
	action   Action
	reason   string
}

// Add installs skills from a source into the selected targets
func Add(opts AddOptions) (*Result, error) {
	notify := opts.Notify
	if notify == nil {
		notify = func(string) {}
	}
	notify("Source: " + opts.Source.Location)
	mat, err := Materialize(opts.Source, false)
	if err != nil {
		return nil, err
	}
	if opts.Source.Kind == SourceGit {
		if mat.FromCache {
			notify("Repository served from cache")
		} else {
			notify("Repository cloned")
		}
	}
	discovered, warnings, err := Discover(mat.Dir)
	if err != nil {
		return nil, err
	}
	if len(discovered) == 0 {
		return nil, fmt.Errorf("no skills (directories with %s) found in %s", SkillFile, opts.Source.Location)
	}
	notify(fmt.Sprintf("Found %d skill(s)", len(discovered)))

	names := opts.Skills
	if !opts.All && len(names) == 0 {
		if opts.Prompt == nil || !opts.Prompt.Interactive() {
			return nil, fmt.Errorf("no skill selected: pass --skill <name> or --all (available: %s)", strings.Join(skillNames(discovered), ", "))
		}
		names, err = opts.Prompt.ChooseSkills(discovered)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, errors.New("no skill selected")
		}
	}
	selected, err := SelectSkills(discovered, names, opts.All)
	if err != nil {
		return nil, err
	}
	notify(fmt.Sprintf("Selected %d skill(s): %s", len(selected), strings.Join(skillNames(selected), ", ")))

	targets := opts.Targets
	if opts.SelectTargets != nil {
		if targets, err = opts.SelectTargets(); err != nil {
			return nil, err
		}
	}
	dirs, err := GroupByDir(targets, opts.Scope)
	if err != nil {
		return nil, err
	}
	var jobs []job
	for _, s := range selected {
		for _, d := range dirs {
			jobs = append(jobs, job{dir: d, skill: s})
		}
	}
	plan, manifests, err := planJobs(opts.Source, jobs)
	if err != nil {
		return nil, err
	}
	if err := trustGate(opts, selected, plan); err != nil {
		return nil, err
	}
	res := &Result{Source: opts.Source.Location, Commit: mat.Commit, FromCache: mat.FromCache, Warnings: warnings}
	err = execute(opts.Source, mat.Commit, plan, manifests, opts.Scope, opts.Force, opts.Prompt, res)
	// The trust gate has passed: remember the source user-wide (O4) once
	// anything from it is in place
	if res.placed() {
		if terr := TrustSource(opts.Source.Location); terr != nil {
			res.Warnings = append(res.Warnings, "could not record the trusted source: "+terr.Error())
		}
	}
	return res, err
}

// planJobs decides the action of each job against the installed state and
// the manifests of the involved skills directories
func planJobs(src Source, jobs []job) ([]plannedJob, map[string]*Manifest, error) {
	manifests := map[string]*Manifest{}
	var plan []plannedJob
	for _, j := range jobs {
		dir := j.dir.Dir
		m, ok := manifests[dir]
		if !ok {
			var err error
			if m, err = LoadManifest(dir); err != nil {
				return nil, nil, err
			}
			manifests[dir] = m
		}
		sum, err := TreeChecksum(j.skill.SourceDir, j.skill.Files)
		if err != nil {
			return nil, nil, err
		}
		p := plannedJob{job: j, dest: filepath.Join(dir, j.skill.Name), checksum: sum}
		p.action, p.reason, err = decide(src, p.dest, sum, m.Find(j.skill.Name))
		if err != nil {
			return nil, nil, err
		}
		plan = append(plan, p)
	}
	return plan, manifests, nil
}

// decide applies the idempotency rules of ADR-043 D4
func decide(src Source, dest, sourceSum string, entry *ManifestEntry) (Action, string, error) {
	if _, err := os.Lstat(dest); errors.Is(err, os.ErrNotExist) {
		return ActionInstall, "", nil
	} else if err != nil {
		return "", "", err
	}
	installedSum, err := DirChecksum(dest)
	if err != nil {
		return "", "", err
	}
	if installedSum == sourceSum {
		return ActionUnchanged, "", nil
	}
	switch {
	case entry == nil:
		return ActionConflict, "directory exists and is not managed by Portunix", nil
	case entry.Source != src.Location:
		return ActionConflict, "installed from a different source (" + entry.Source + ")", nil
	case entry.Checksum != installedSum:
		return ActionConflict, "modified locally since installation", nil
	}
	return ActionUpdate, "", nil
}

// trustGate asks before the first write from a source no manifest has
// recorded yet (ADR-043 D6)
func trustGate(opts AddOptions, selected []ResolvedSkill, plan []plannedJob) error {
	if opts.Yes || !anyWrite(plan) {
		return nil
	}
	known, err := sourceKnown(opts.Source.Location)
	if err != nil {
		return err
	}
	if known {
		return nil
	}
	dirs := map[string]bool{}
	for _, p := range plan {
		dirs[p.dest] = true
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n⚠️  First install from a new source: %s\n", opts.Source.Location)
	b.WriteString("   Skills are instructions your AI agents will follow. Review the source before trusting it.\n")
	fmt.Fprintf(&b, "   Skills: %s\n", strings.Join(skillNames(selected), ", "))
	b.WriteString("   Will write:\n")
	for _, d := range sortedKeys(dirs) {
		fmt.Fprintf(&b, "     %s\n", d)
	}
	if opts.Prompt == nil || !opts.Prompt.Interactive() {
		return fmt.Errorf("%s\nrefusing to install from an untrusted source without confirmation; re-run with --yes", b.String())
	}
	if !opts.Prompt.Confirm(b.String() + "\nProceed?") {
		return errors.New("installation cancelled")
	}
	return nil
}

// sourceKnown reports whether the source is in the user-wide trust record
// or recorded by a manifest of the current project or the user scope
func sourceKnown(source string) (bool, error) {
	if trusted, err := IsTrusted(source); err != nil {
		return false, err
	} else if trusted {
		return true, nil
	}
	for _, scope := range []Scope{ScopeProject, ScopeUser} {
		dirs, err := GroupByDir(AllTargets(), scope)
		if err != nil {
			continue
		}
		for _, d := range dirs {
			m, err := LoadManifest(d.Dir)
			if err != nil {
				continue
			}
			if m.HasSource(source) {
				return true, nil
			}
		}
	}
	return false, nil
}

func anyWrite(plan []plannedJob) bool {
	for _, p := range plan {
		if p.action != ActionUnchanged {
			return true
		}
	}
	return false
}

// execute performs the planned writes and records provenance
func execute(src Source, commit string, plan []plannedJob, manifests map[string]*Manifest, scope Scope, force bool, prompt Prompter, res *Result) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var firstErr error
	for _, p := range plan {
		out := Outcome{Skill: p.skill.Name, Target: p.dir.Label(), Dir: p.dest, Action: p.action}
		if p.action == ActionConflict {
			overwrite := force
			if !overwrite && prompt != nil && prompt.Interactive() {
				overwrite = prompt.Confirm(fmt.Sprintf("%s: %s. Overwrite?", p.dest, p.reason))
			}
			if !overwrite {
				out.Reason = p.reason + " (use --force to overwrite)"
				res.Outcomes = append(res.Outcomes, out)
				continue
			}
			out.Action = ActionUpdate
		}
		if out.Action != ActionUnchanged {
			if _, err := p.dir.Targets[0].Install(p.skill, scope, true); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				out.Action, out.Reason = ActionFailed, err.Error()
				res.Outcomes = append(res.Outcomes, out)
				continue
			}
		}
		manifests[p.dir.Dir].Upsert(ManifestEntry{
			Name:        p.skill.Name,
			Source:      src.Location,
			Ref:         src.Ref,
			Commit:      commit,
			InstalledAt: now,
			Files:       p.skill.Files,
			Checksum:    p.checksum,
		})
		res.Outcomes = append(res.Outcomes, out)
	}
	for _, dir := range sortedKeys(boolKeys(manifests)) {
		if err := manifests[dir].Save(dir); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// UpdateOptions configures `skills update`
type UpdateOptions struct {
	// Names limits the update to these skills; empty means all managed skills
	Names  []string
	Scope  Scope
	Ref    string
	Force  bool
	Prompt Prompter
	Notify func(msg string)
}

// Update re-resolves managed skills at their recorded source and ref
func Update(opts UpdateOptions) ([]*Result, error) {
	notify := opts.Notify
	if notify == nil {
		notify = func(string) {}
	}
	type group struct {
		src     Source
		entries map[string][]TargetDir // skill name -> skills directories
	}
	groups := map[string]*group{}
	wanted := map[string]bool{}
	for _, n := range opts.Names {
		wanted[n] = true
	}
	found := map[string]bool{}
	dirs, err := GroupByDir(AllTargets(), opts.Scope)
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		m, err := LoadManifest(d.Dir)
		if err != nil {
			return nil, err
		}
		for _, e := range m.Skills {
			if len(wanted) > 0 && !wanted[e.Name] {
				continue
			}
			found[e.Name] = true
			ref := e.Ref
			if opts.Ref != "" {
				ref = opts.Ref
			}
			key := e.Source + "@" + ref
			g, ok := groups[key]
			if !ok {
				g = &group{src: manifestSource(e.Source, ref), entries: map[string][]TargetDir{}}
				groups[key] = g
			}
			g.entries[e.Name] = append(g.entries[e.Name], d)
		}
	}
	for n := range wanted {
		if !found[n] {
			return nil, fmt.Errorf("skill %q is not managed by Portunix in %s scope", n, opts.Scope)
		}
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("no Portunix-managed skills in %s scope", opts.Scope)
	}

	var results []*Result
	var firstErr error
	for _, key := range sortedKeys(boolKeys(groups)) {
		g := groups[key]
		notify("Updating from " + g.src.Location)
		mat, err := Materialize(g.src, true)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		discovered, warnings, err := Discover(mat.Dir)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		res := &Result{Source: g.src.Location, Commit: mat.Commit, Warnings: warnings}
		byName := map[string]ResolvedSkill{}
		for _, s := range discovered {
			byName[s.Name] = s
		}
		var jobs []job
		for _, name := range sortedKeys(boolKeys(g.entries)) {
			s, ok := byName[name]
			if !ok {
				res.Warnings = append(res.Warnings, fmt.Sprintf("skill %q no longer exists in %s, left as is", name, g.src.Location))
				continue
			}
			for _, d := range g.entries[name] {
				jobs = append(jobs, job{dir: d, skill: s})
			}
		}
		plan, manifests, err := planJobs(g.src, jobs)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := execute(g.src, mat.Commit, plan, manifests, opts.Scope, opts.Force, opts.Prompt, res); err != nil && firstErr == nil {
			firstErr = err
		}
		results = append(results, res)
	}
	return results, firstErr
}

// manifestSource rebuilds a Source from a manifest entry
func manifestSource(location, ref string) Source {
	if filepath.IsAbs(location) {
		return Source{Kind: SourceLocal, Location: location, Ref: ref}
	}
	return Source{Kind: SourceGit, Location: location, Ref: ref}
}

// RemoveResult lists the skill directories a removal touched
type RemoveResult struct {
	Removed []string
	// Kept are directories left in place because they contain files
	// Portunix did not install
	Kept []string
}

// Remove deletes a managed skill from the given targets (all when empty).
// Only files recorded in the manifest are deleted (ADR-043 D5).
func Remove(name string, targets []AgentTarget, scope Scope) (*RemoveResult, error) {
	if !skillNamePattern.MatchString(name) {
		return nil, fmt.Errorf("invalid skill name %q", name)
	}
	if len(targets) == 0 {
		targets = AllTargets()
	}
	dirs, err := GroupByDir(targets, scope)
	if err != nil {
		return nil, err
	}
	res := &RemoveResult{}
	managed := false
	for _, d := range dirs {
		m, err := LoadManifest(d.Dir)
		if err != nil {
			return res, err
		}
		entry := m.Find(name)
		if entry == nil {
			continue
		}
		managed = true
		dest := filepath.Join(d.Dir, name)
		for _, rel := range entry.Files {
			p, err := containedPath(dest, rel)
			if err != nil {
				return res, err
			}
			if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
				return res, fmt.Errorf("remove %s: %w", p, err)
			}
		}
		pruneEmptyDirs(dest)
		if _, err := os.Lstat(dest); err == nil {
			res.Kept = append(res.Kept, dest)
		} else {
			res.Removed = append(res.Removed, dest)
		}
		m.Remove(name)
		if err := m.Save(d.Dir); err != nil {
			return res, err
		}
	}
	if !managed {
		return nil, fmt.Errorf("skill %q is not managed by Portunix in %s scope (Portunix never removes skills it did not install)", name, scope)
	}
	return res, nil
}

// containedPath joins rel to base and rejects paths escaping base
func containedPath(base, rel string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("manifest lists unsafe path %q", rel)
	}
	return filepath.Join(base, clean), nil
}

// pruneEmptyDirs removes empty directories bottom-up, root included
func pruneEmptyDirs(root string) {
	var dirs []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = os.Remove(dirs[i]) // fails harmlessly on non-empty directories
	}
}

// ListEntry is one row of `skills list`
type ListEntry struct {
	Name string `json:"name"`
	// Target is the comma separated list of targets sharing the directory
	Target      string   `json:"target"`
	Targets     []string `json:"targets"`
	Scope       string   `json:"scope"`
	Path        string   `json:"path"`
	Managed     bool     `json:"managed"`
	Status      string   `json:"status"`
	Source      string   `json:"source,omitempty"`
	Ref         string   `json:"ref,omitempty"`
	Commit      string   `json:"commit,omitempty"`
	InstalledAt string   `json:"installedAt,omitempty"`
}

// List reports managed and unmanaged skills of every skills directory in a scope
func List(scope Scope) ([]ListEntry, error) {
	dirs, err := GroupByDir(AllTargets(), scope)
	if err != nil {
		return nil, err
	}
	var out []ListEntry
	for _, d := range dirs {
		m, err := LoadManifest(d.Dir)
		if err != nil {
			return nil, err
		}
		for _, e := range m.Skills {
			dest := filepath.Join(d.Dir, e.Name)
			status := "ok"
			if _, err := os.Lstat(dest); err != nil {
				status = "missing"
			} else if sum, err := DirChecksum(dest); err != nil || sum != e.Checksum {
				status = "modified"
			}
			out = append(out, ListEntry{
				Name: e.Name, Target: d.Label(), Targets: d.IDs(), Scope: scope.String(), Path: dest, Managed: true, Status: status,
				Source: e.Source, Ref: e.Ref, Commit: e.Commit, InstalledAt: e.InstalledAt,
			})
		}
		entries, err := os.ReadDir(d.Dir)
		if err != nil {
			continue
		}
		for _, de := range entries {
			if !de.IsDir() || m.Find(de.Name()) != nil {
				continue
			}
			if _, err := os.Stat(filepath.Join(d.Dir, de.Name(), SkillFile)); err != nil {
				continue
			}
			out = append(out, ListEntry{Name: de.Name(), Target: d.Label(), Targets: d.IDs(), Scope: scope.String(), Path: filepath.Join(d.Dir, de.Name()), Status: "unmanaged"})
		}
	}
	return out, nil
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func boolKeys[V any](m map[string]V) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}
