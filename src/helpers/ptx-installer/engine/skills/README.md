# engine/skills

Package `skills` is the engine behind `portunix skills`. It installs AI-agent
skills, which are directories containing a `SKILL.md` file, from git
repositories or local paths into the skills directories that AI coding agents
read. It records provenance so that `list`, `update` and `remove` only touch
skills Portunix installed.

Import path: `portunix.ai/portunix/src/helpers/ptx-installer/engine/skills`.
Architecture: [ADR-043](../../../../../docs/adr/043-skill-installation-architecture.md).
Issue: [#190](../../../../../docs/issues/done/190-skill-installation-support.md).
User documentation: [docs/helpers/ptx-installer.md](../../../../../docs/helpers/ptx-installer.md).

## Role in the project

```text
portunix skills …                     main binary, src/dispatcher routes "skills"
  └─ ptx-installer skills …           helper binary
       ├─ cmd_skills.go               flag parsing, terminal prompts, output
       └─ engine/skills (this package) sources, discovery, targets, manifests, trust
```

The package has no terminal I/O of its own. Questions go through the
`Prompter` interface and progress through the `Notify` callback, both supplied
by `cmd_skills.go`. Tests drive it with a scripted prompter.

`src/cmd/skills.go` in the main binary only provides shell completion for the
same command surface.

## Files

| File | Responsibility |
| ----- | ----- |
| `source.go` | Parse a source (git URL, `owner/repo`, local path), shallow clone into the cache, resolve the commit |
| `discover.go` | Find `SKILL.md` directories, parse front-matter, list skill files, select skills by name |
| `target.go` | `Scope`, `ResolvedSkill`, `AgentTarget`, the directory-based target, `GroupByDir`, atomic tree writes |
| `agents.go` | Universal target and agent driver registry, aliases, `ResolveTargets` |
| `manifest.go` | `.portunix-skills.json` provenance manifest and SHA-256 tree checksums |
| `trust.go` | User-wide record of trusted sources |
| `install.go` | Orchestration: `Add`, `Update`, `Remove`, `List`, idempotency rules, trust gate |
| `skills_test.go` | Unit tests; every test isolates cwd, home, cache and trust file |

## Flow of `Add`

```text
Materialize(source)        local dir as-is, or cache hit, or git clone into the cache
  → Discover(dir)          every directory with SKILL.md (not nested), sorted by name
  → select skills          --skill / --all / Prompter.ChooseSkills
  → SelectTargets          optional interactive agent choice (after the skills)
  → GroupByDir(targets)    targets sharing a directory become one write
  → planJobs + decide      install / update / unchanged / conflict per skill × directory
  → trustGate              new source → confirm or --yes, before anything is written
  → execute                write via the first target of each directory, upsert manifests
  → TrustSource            remember the source user-wide once something is in place
```

`Update` groups manifest entries by `source@ref`, always re-fetches git sources
(`Materialize(src, true)`) and reuses `planJobs` and `execute`. `Remove` deletes
only the files a manifest lists. `List` reports one row per skills directory,
including unmanaged skills that have a `SKILL.md` but no manifest entry.

## Targets

An `AgentTarget` resolves a skills directory per scope:

- **Project scope** (default): relative to the working directory.
- **User scope** (`--global`): below the home directory or `$XDG_CONFIG_HOME`
  (falling back to `~/.config`). `userEnv` lets an agent override its root,
  for example `CLAUDE_CONFIG_DIR` for Claude Code.

All agent drivers are `dirTarget` values in `agentRegistry`. Their paths follow
the agent table of the `skills` CLI (`skills@1.5.9`). Most agents read the
shared `.agents/skills` in a project, so `GroupByDir` merges them with the
universal target. The skill is written once, and outcomes and list rows carry
a label such as `universal,amp,copilot,cursor,gemini,opencode,zed`.

An agent counts as detected when its registry package is installed or any of
its `detectPaths` exists. The registry check is the Issue #035 detection,
injected through `SetDetector` so tests do not load the registry. Detection
picks the default agents when the user gives no `--agent`.

### Adding an agent

1. Add a `dirTarget` to `agentRegistry` in `agents.go` with `projectDir`,
   `userDir` and `detectPaths`. Add `detectPackage` when the agent is in the
   package registry.
2. Add an alias to `agentAliases` when the `skills` CLI uses a different name.
3. Extend the expectations in `TestAgentSkillsDirs` and `TestResolveTargets`.
4. Update `skillsAgentIDs` in `src/cmd/skills.go` (completion) and the target
   table in `docs/helpers/ptx-installer.md`.

## Files on disk

| File | Location | Content |
| ----- | ----- | ----- |
| Cache | `<user-cache>/portunix/skills/<host>/<owner>/<repo>/<ref>/` (`PORTUNIX_SKILLS_CACHE`) | Clone without `.git`, plus `.portunix-source.json` with the commit |
| Manifest | `<skills-dir>/.portunix-skills.json` | Per skill: source, ref, commit, files, checksum, install time |
| Trust record | `<user-config>/portunix/skills-trusted-sources.json` (`PORTUNIX_SKILLS_TRUST`), mode 0600 | Sources the user confirmed or installed with `--yes` |

A cache entry is promoted from a `.tmp` sibling only after a successful clone.
Manifests and the trust record are written to a `.tmp` file and renamed.
Skill directories are staged as `<name>.portunix-new` and swapped in.

## Rules

- **Idempotency (D4)** — identical content is `unchanged`. A managed skill that
  is unchanged locally is `updated`. An unmanaged directory, a locally modified
  skill or a skill from another source is a `conflict`: overwritten only with
  `force` or a confirmed prompt. A write error is `failed`.
- **Trust gate (D6, O4)** — a source is known when it is in the trust record
  or in a manifest of the current project or the user scope. Otherwise `Add`
  asks before writing, or requires `Yes` when there is no terminal. A declined
  or failed gate records nothing.
- **Safety** — skill names must be one safe path segment. Symlinks in sources
  are ignored. `Remove` rejects manifest paths that escape the skill
  directory. Portunix never deletes files it did not install.

## Tests

```bash
cd src/helpers/ptx-installer
go test ./engine/skills/            # unit tests
go test -race ./...                 # whole helper
cd ../../.. && go test -tags integration ./test/integration/ -run TestIssue190 -v
```

The integration test runs the full command surface in an `ubuntu:24.04`
container through `portunix container`. It covers prompts on a pseudo-terminal
via `script` and a real clone of `obra/superpowers`, which is skipped when
GitHub is not reachable.
