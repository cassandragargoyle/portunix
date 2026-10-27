# ptx-installer — `portunix install`, `portunix package`, `portunix skills`

`ptx-installer` is the helper binary behind package installation (`install`),
package registry operations (`package`) and AI-agent skill installation
(`skills`). This document currently covers the `skills` command; see
`portunix install --help` and `portunix package --help` for the others.

## Skills

`portunix skills` installs AI-agent **skills** — directories with a `SKILL.md`
file plus supporting files — from git repositories or local paths, without a
Node/`npx` toolchain. Delivered by
[Issue #190](../issues/done/190-skill-installation-support.md); architecture in
[ADR-043](../adr/043-skill-installation-architecture.md).

### Quick start

```bash
# Install one skill from a GitHub repository into the current project
portunix skills add https://github.com/obra/superpowers --skill brainstorming

# GitHub shorthand, every skill, also for Claude Code, without prompts
portunix skills add obra/superpowers --all --agent claude --yes

# Show what is installed, refresh it, remove it
portunix skills list
portunix skills update
portunix skills remove brainstorming
```

### Commands

| Command | Description |
| ------- | ----------- |
| `add <repo-url\|owner/repo\|path>` | Install skills from a source |
| `list [--json]` | List managed and unmanaged skills of every target |
| `remove <skill-name> [--agent <id>]` | Remove a Portunix-installed skill |
| `update [<skill-name>...] [--ref <ref>] [--force]` | Re-install managed skills from their recorded source |
| `sources [--json]` | Show install targets, their directories and agent detection |
| `trust [list [--json]\|remove <source>]` | Show or revoke user-wide trusted sources |

All commands accept `--global` (alias `--user`, `-g`) to work in the user scope.

### Options of `add`

| Option | Description |
| ------ | ----------- |
| `--skill`, `-s <name>` | Skill to install; repeatable or comma separated |
| `--all` | Install every discovered skill |
| `--agent`, `-a <id>` | Also install for this agent; repeatable or comma separated |
| `--all-agents` | Install for every supported agent |
| `--no-universal` | Skip the universal `.agents/skills/` directory |
| `--ref <ref>` | Branch, tag or commit SHA (default: repository default branch) |
| `--source <path>` | Install from a local checkout (air-gapped) |
| `--global`, `--user`, `-g` | Install under the home directory |
| `--yes`, `-y` | Trust a new source without asking |
| `--force`, `-f` | Overwrite skills that differ from the source |

Without `--skill` or `--all`, an interactive terminal offers a numbered
selection; a non-interactive run fails and lists the available skill names.

Without `--agent`, `--all-agents` or `--yes`, an interactive terminal then
offers the agents with the detected ones pre-selected (Enter keeps them,
`all`, `none` or numbers change the selection). A non-interactive run uses the
detected agents.

### Targets and scope

| Target | Agent | Project scope (default) | User scope (`--global`) |
| ------ | ----- | ----------------------- | ----------------------- |
| `universal` | shared convention | `./.agents/skills/` | `~/.agents/skills/` |
| `amp` | Amp | `./.agents/skills/` | `$XDG_CONFIG_HOME/agents/skills/` |
| `claude` | Claude Code | `./.claude/skills/` | `~/.claude/skills/` (`$CLAUDE_CONFIG_DIR/skills/`) |
| `copilot` | GitHub Copilot | `./.agents/skills/` | `~/.copilot/skills/` |
| `cursor` | Cursor | `./.agents/skills/` | `~/.cursor/skills/` |
| `gemini` | Gemini CLI | `./.agents/skills/` | `~/.gemini/skills/` |
| `opencode` | OpenCode | `./.agents/skills/` | `$XDG_CONFIG_HOME/opencode/skills/` |
| `zed` | Zed | `./.agents/skills/` | `~/.agents/skills/` |

`$XDG_CONFIG_HOME` defaults to `~/.config` on every platform. The layout
follows the agent table of the `skills` CLI (`skills@1.5.9`); the IDs
`claude-code`, `gemini-cli` and `github-copilot` are accepted as aliases.

The universal target is always written unless `--no-universal` is given.
Targets that resolve to the same directory are written once and listed as one
row (for example `universal,amp,copilot,cursor,gemini,opencode,zed` in a
project). An agent is detected when its registry package is installed (Issue #035
detection, `claude-code`, `gemini-cli`) or its configuration directory
exists (`~/.claude`, `~/.cursor`, `~/.copilot`, `~/.gemini`, and `amp`,
`opencode`, `zed` under `$XDG_CONFIG_HOME`). `portunix skills sources` shows
the directories and the detection state of every target.

### Sources and cache

- **Git URL** — `https://…`, `git@host:owner/repo.git` or the GitHub shorthand
  `owner/repo`. The repository is shallow-cloned with the local `git` binary
  into `<user-cache>/portunix/skills/<host>/<owner>/<repo>/<ref>/`
  (`~/.cache/…` on Linux, `%LocalAppData%\…` on Windows; override with
  `PORTUNIX_SKILLS_CACHE`). Later `add` runs for the same repository and ref are
  served from the cache without network access; `update` always re-fetches.
- **Local path** — an existing directory (positional or `--source <path>`) is
  used as-is. This is the air-gapped and local-development workflow.

A skill is any directory containing `SKILL.md`. Its name comes from the
`name` field of the YAML front-matter (or the directory name) and must be a
single safe path segment.

### Provenance and safety

Every skills directory gets a `.portunix-skills.json` manifest recording, per
skill, the source, ref, resolved commit, installed files and a SHA-256
checksum. Portunix updates and removes only skills recorded there;
hand-placed skills are shown by `list` as `unmanaged` and never touched.

- **Trust gate** — skills are instructions your AI agents will follow. The
  first install from a new source prints the source, the skills and the exact
  target directories and asks for confirmation. Non-interactive runs must pass
  `--yes`. A source counts as known when it is in the user-wide trust record
  or in a manifest of the current project or the user scope.
- **Trusted sources** — after a confirmed (or `--yes`) install, the source is
  recorded in `<user-config>/portunix/skills-trusted-sources.json`
  (`~/.config/…` on Linux, `%AppData%\…` on Windows; override with
  `PORTUNIX_SKILLS_TRUST`). Installing it into another project does not ask
  again. `portunix skills trust remove <source>` revokes the trust; projects
  whose manifests record the source still treat it as known.
  A corrupt trust file blocks installs from new sources (except with `--yes`)
  and `trust list` until it is fixed or deleted; deleting it only resets the
  record, so new sources ask for confirmation again. Review untrusted repositories before
  installing; Portunix never executes skill content, it only places files.
- **Idempotency** — identical content is a no-op. A managed skill that is
  unchanged locally is updated when the source changes. An unmanaged
  directory, a locally modified skill or a skill from a different source is a
  conflict: the command asks on a terminal, otherwise it skips the target and
  exits non-zero until `--force` is given. A target that cannot be written
  (for example a read-only directory) is reported as `failed`.
- **Removal** — only files listed in the manifest are deleted; a skill
  directory containing other files is kept.

### Shell completion

The `portunix completion` scripts complete the `skills` subcommands, their
flags, agent IDs for `--agent` and installed skill names for `remove` and
`update`.

### Requirements

Git sources need `git` on `PATH` (`portunix install git`). Local sources need
nothing else.
