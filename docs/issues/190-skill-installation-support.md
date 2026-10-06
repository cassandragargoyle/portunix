# Issue #190: Skill Installation Support — `portunix skills`

**Priority**: High
**Type**: Feature
**Labels**: enhancement, helper-binary, ptx-installer, ai-integration, skills, multi-agent, dispatcher
**Assignee**: Developer
**Reporter**: Architect
**Created**: 2026-05-30
**Status**: 📋 Open
**Related ADR**: [ADR-043 — Skill Installation Architecture](../adr/043-skill-installation-architecture.md)

## Summary

Add the ability to install AI-agent **skills** (Markdown `SKILL.md` bundles distributed as
Git repositories) into a project or the user's home, mirroring the community `skills` CLI:

```text
npx skills add https://github.com/obra/superpowers --skill brainstorming
```

Portunix equivalent:

```text
portunix skills add https://github.com/obra/superpowers --skill brainstorming
```

Per ADR-043, the capability is delivered as a new `skills` verb inside the existing
**`ptx-installer`** helper (not a new helper), and installs each selected skill into a
**universal `.agents/skills/`** directory plus the per-agent directories of the selected
agents (full parity with the `skills` CLI multi-agent model).

## Motivation

- **No Node dependency**: developers with Portunix install skills without `npx`/Node.
- **Multi-agent**: one command provisions a skill for Claude Code, Cursor, Gemini CLI,
  Copilot, Zed, OpenCode, Amp, … via the shared `.agents/skills/` convention + per-agent
  drivers.
- **Provenance**: Portunix tracks what it installed (source repo + commit) so `list`,
  `remove`, and `update` work reliably.
- **Reuse**: leans on existing `ptx-installer` machinery (git clone+cache, ADR-032 paths,
  Issue #035 agent detection).

## Reference behaviour (observed `skills@1.5.9` run)

```text
◇  Source: https://github.com/obra/superpowers.git
◇  Repository cloned
◇  Found 14 skills
●  Selected 1 skill: brainstorming
◆  Which agents do you want to install to?
   ── Universal (.agents/skills) ── always included ──
   • Amp • Antigravity • Cline • Codex • Cursor • Deep Agents • Dexto
   • Firebender • Gemini CLI • GitHub Copilot • Kimi Code CLI • OpenCode • Warp • Zed
```

Key takeaways encoded in ADR-043: clone a Git source → discover `SKILL.md` directories →
select skill(s) → write universal `.agents/skills/` (always) + selected per-agent dirs.

## Command Surface

```text
portunix skills add <repo-url|path> [--skill <name>]... [--all] \
        [--agent <id>]... [--all-agents] [--no-universal] \
        [--global] [--ref <tag-or-sha>] [--source <path>] [--yes] [--force]
portunix skills list [--global] [--json]
portunix skills remove <skill-name> [--agent <id>]... [--global]
portunix skills update [<skill-name>] [--global]
portunix skills sources                 # list agent targets + detection state
```

Dispatcher change (`src/dispatcher/dispatcher.go`):

```go
d.helpers["ptx-installer"] = &HelperConfig{
    Commands: []string{"install", "package", "skills"}, // + "skills"
    Binary:   "ptx-installer",
    Required: false,
}
```

## Technical Design (per ADR-043)

### New code in `src/helpers/ptx-installer/`

```text
cmd_skills.go            # cobra subcommands: add / list / remove / update / sources
engine/skills/
  source.go              # git shallow clone + ~/.cache/portunix/skills/<host>/<owner>/<repo>/<ref>/
                         # + --source <path> (local / air-gapped)
  discover.go            # find SKILL.md dirs, parse front-matter (name, description)
  target.go              # Scope, ResolvedSkill, AgentTarget interface, universal target
  agents.go              # driver registry (claude [P1]; cursor, gemini, copilot, zed, opencode, amp [P2])
  manifest.go            # .portunix-skills.json read/write
  install.go             # orchestration: resolve → discover → select → trust-gate → fan-out
```

Reused as-is: `engine/detection.go` (Issue #035 agent detection → `AgentTarget.Detect()`),
`engine/path.go` (ADR-032 cross-platform cache + user-scope paths).

### `AgentTarget` interface

```go
type Scope int
const ( ScopeProject Scope = iota; ScopeUser )

type ResolvedSkill struct {
    Name        string
    Description string
    SourceDir   string
    Files       []string
}

type AgentTarget interface {
    ID() string                                   // "claude", "cursor", ...
    DisplayName() string
    Detect() (installed bool, err error)
    SkillsDir(scope Scope) (string, error)
    Install(s ResolvedSkill, scope Scope, force bool) (written []string, err error)
}
```

- **Universal target**: `.agents/skills/<name>/` (project) / `~/.agents/skills/<name>/`
  (user). Always written unless `--no-universal`.
- **Claude driver (P1)**: `.claude/skills/<name>/SKILL.md` (project) /
  `~/.claude/skills/<name>/` (user).

### Provenance manifest `<skills-dir>/.portunix-skills.json`

```json
{
  "version": 1,
  "skills": [
    { "name": "brainstorming", "source": "https://github.com/obra/superpowers.git",
      "ref": "main", "commit": "<sha>", "installedAt": "<ts>",
      "files": ["SKILL.md", "..."], "checksum": "<sha256>" }
  ]
}
```

Portunix only ever modifies skills recorded in its manifest; hand-placed skills appear in
`list` as "unmanaged" and are never touched.

## Implementation Plan

### Phase 1 — Core + universal + Claude Code driver

- [ ] Dispatcher: add `skills` to `ptx-installer` commands
- [ ] `cmd_skills.go`: `add`, `list`, `remove`, `update`, `sources` skeletons + meta-flags
      (`--version`, `--description`, `--list-commands`, `--help`, `--help-ai`, `--help-expert`)
- [ ] `source.go`: git shallow clone + cache; `--source <path>`; `git`-missing remediation
- [ ] `discover.go`: discover `SKILL.md` dirs + parse front-matter; validate `--skill`
- [ ] `target.go`: `AgentTarget` interface + universal target
- [ ] `agents.go`: registry + `claude` driver
- [ ] `manifest.go`: read/write `.portunix-skills.json`
- [ ] `install.go`: orchestration + trust gate (D6) + idempotency (`--force`)
- [ ] Skill selection: `--skill` (repeatable), `--all`, interactive multi-select, no-TTY error
- [ ] Scope: project default + `--global`/`--user`
- [ ] `list` (table + `--json`), `remove`, `update` against the manifest

### Phase 2 — Additional agent drivers + parity

- [ ] Per-agent drivers: Cursor, Gemini CLI, GitHub Copilot, Zed, OpenCode, Amp
      (each additive, owns its `SkillsDir`/`Install`)
- [ ] Interactive agent multi-select pre-populated by detection
- [ ] `--all-agents`
- [ ] `sources` command output (targets + detection state)

### Phase 3 — Polish

- [ ] Progress + summary output (source@commit, targets written)
- [ ] `update` for all managed skills; checksum/commit refresh
- [ ] Documentation: `docs/helpers/ptx-installer.md` skills section + trust/safety note
- [ ] Shell-completion entries for `skills` verbs

## Acceptance Criteria

### Functional

- [ ] `portunix skills add <repo> --skill <name>` clones, discovers, and installs the
      named skill into `.agents/skills/<name>/` and the Claude driver dir (when selected)
- [ ] `--all` installs every discovered skill; interactive multi-select works on a TTY;
      no-TTY without `--skill`/`--all` errors clearly
- [ ] `--source <path>` installs from a local checkout with no network (air-gapped)
- [ ] Second `add`/`update` of the same `<repo>@<ref>` is served from cache (no network)
- [ ] `--global`/`--user` installs under the home directory; project is the default
- [ ] `list` shows managed skills (table + `--json`); `remove` deletes only manifest-recorded
      files; `update` re-resolves at the recorded ref
- [ ] Trust gate prompts before the first write from a new source; `--yes` bypasses
- [ ] Idempotent re-install is a no-op; content change prompts or requires `--force`

### Integration

- [ ] `skills` routed correctly: `portunix skills add` → `ptx-installer skills add`
- [ ] `install` / `package` grammar unchanged
- [ ] `Detect()` reuses Issue #035 detection for default agent selection
- [ ] Works on Windows and Linux (paths via ADR-032)
- [ ] `git`-missing path fails fast with remediation hint

### Quality

- [ ] Unit tests: source resolution (git/path/cache), discovery + front-matter parse,
      manifest read/write, idempotency, scope path resolution
- [ ] Integration tests in a container (per Container-Based Testing Policy) using
      `portunix container`, exercising `add` from a local `--source` fixture and from a
      cached clone
- [ ] User-friendly errors for: invalid repo, `--skill` not found, missing `git`,
      unwritable target

## Testing Strategy

1. **Unit**: source/cache, discovery, selection, manifest, scope paths, trust gate.
2. **Integration (container)**: `portunix container` Ubuntu image; install Portunix; run
   `portunix skills add --source <fixture>` and a cached git clone; verify universal +
   claude dirs and manifest; verify `list`/`remove`/`update`.
3. **Cross-platform**: Windows + Linux path resolution and `.exe` git invocation.
4. **Air-gapped**: `--source <path>` with network disabled.

## Dependencies

- ADR-043 (this issue's architecture) — **must be Active before Phase 1**
- ADR-025 / Issue #100 — `ptx-installer` (the helper being extended)
- ADR-032 — cross-platform path resolution
- ADR-040 / ADR-041 — shallow-clone + cache + pinned-ref precedent
- Issue #035 — AI assistant detection (reused for agent defaults)

## Out of Scope

- Authoring/publishing skills (Portunix is a consumer/installer only)
- Executing skill content (Portunix only places files where agents read them)
- A skills registry/search index (sources are explicit Git URLs / paths)

## Notes

- Maintain the single `portunix` binary + dispatcher pattern (no `portunix_skills.exe`).
- Follow `docs/contributing/HELPER-BINARY-DEVELOPMENT.md` for any helper-surface changes
  even though this extends an existing helper rather than adding a new one.
