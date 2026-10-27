---
title: "skills"
description: "Install AI-agent skills from git repositories"
---

# skills

Install AI-agent skills from git repositories

## Usage

```bash
portunix skills [options] [arguments]
```

## Full Help

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.

```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| [add](#add) | <repo-url|owner/repo|path>   Install skills from a source |
| [list](#list) | List installed skills (managed and unmanaged) |
| [remove](#remove) | <skill-name>              Remove a Portunix-installed skill |
| [update](#update) | [<skill-name>...]         Re-install managed skills from their recorded source |
| [sources](#sources) | Show install targets and agent detection state |
| [trust](#trust) | [list|remove <source>]     Show or revoke user-wide trusted sources |

### add

<repo-url|owner/repo|path>   Install skills from a source

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

### list

List installed skills (managed and unmanaged)

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

### remove

<skill-name>              Remove a Portunix-installed skill

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

### update

[<skill-name>...]         Re-install managed skills from their recorded source

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

### sources

Show install targets and agent detection state

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

### trust

[list|remove <source>]     Show or revoke user-wide trusted sources

```
Usage: portunix skills <command> [options]

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
not asked for again in other projects; 'skills trust remove' revokes that.
```

## Examples

```bash
  portunix skills add https://github.com/obra/superpowers --skill brainstorming
  portunix skills add obra/superpowers --all --agent claude --yes
  portunix skills add --source ./my-skills --skill reviewer --global
  portunix skills list --json
  portunix skills update
  portunix skills remove brainstorming

```
