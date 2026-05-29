---
title: Markdown Frontmatter Convention
description: YAML frontmatter schema for project `.md` files. AI assistants (Claude Code, Gemini) read frontmatter first to score relevance, decide load strategy, and route to the right methodology. Apply to any new or substantially edited `.md` file.
category: style-guide
ai_load: eager
status: active
language: en
created: 2026-05-07
last_updated: 2026-05-07
related:
  - MARKDOWN-STYLE.md
  - README.md
---

# Markdown Frontmatter Convention

> **AI assistant load scope (Claude Code, Gemini):**
>
> **Load eagerly when:**
>
> - Creating a new `.md` file in this repository (you must emit valid frontmatter)
> - Editing an existing `.md` file that has frontmatter (preserve / update fields like `last_updated`)
> - Backfilling frontmatter into a legacy `.md` file
> - Validating that a doc's frontmatter matches its content (`status`, `category`, `ai_load`)
>
> **Do NOT load for:**
>
> - Reading `.md` content unrelated to its frontmatter (the body has its own scope rules)
> - Files exempt from frontmatter — root `README.md` of the repo, `CLAUDE.md` files, third-party `.md` (vendored, generated)
> - Single-question Q&A about the project
>
> **Authoritative scope:** the YAML schema, allowed values, lifecycle rules, and the
> migration procedure for legacy files. Body formatting (headings, tables, code blocks) lives in [MARKDOWN-STYLE.md](MARKDOWN-STYLE.md).

## Overview

YAML frontmatter sits at the very top of every project `.md` file. Supported AI
assistants — Claude Code and Gemini — read it before anything else to decide:

1. Whether the document is relevant to the current task (`description`, `category`)
2. Whether to load it eagerly, scoped, or only on demand (`ai_load`)
3. How it relates to other documents (`translation_of`, `superseded_by`, `related`)

Frontmatter **complements** — does NOT replace — the AI assistant load-scope
blockquote that follows the H1. Frontmatter is the machine-readable summary; the
blockquote is the human-readable detail with rationale and examples. When the two
disagree, the blockquote wins.

## Supported AI assistants

| Assistant | Surface | Reads frontmatter | Reads load-scope blockquote |
| --------- | ------- | ----------------- | --------------------------- |
| Claude Code | CLI / IDE (Anthropic) | yes | yes |
| Gemini Code Assist | CLI / IDE (Google) | yes | yes |

The schema is authored for these two assistants. Other agents (Cursor's built-in
agent, Aider, Cody, Continue, etc.) are welcome to consume it but are not gating
clients. When adding a new assistant, append a row here and document its behavior
in [Per-assistant overrides](#per-assistant-overrides) if it differs from the default.

## Schema

### Required fields

| Field | Type | Allowed values | Purpose |
| ----- | ---- | -------------- | ------- |
| `title` | string | any | Document title; should match the H1 |
| `description` | string | 1-2 sentences | Relevance signal for AI assistants — lead with *when to apply*, not *what it is about* |
| `category` | enum | `style-guide`, `methodology`, `workflow`, `ai-guidance`, `reference`, `index`, `adr`, `specification` | Coarse classification |
| `ai_load` | enum | `eager`, `scoped`, `router`, `on-demand`, `never` | When AI assistants should load this file |
| `status` | enum | `draft`, `active`, `deprecated`, `superseded` | Lifecycle state |

### Optional fields

| Field | Type | Purpose |
| ----- | ---- | ------- |
| `language` | `en` \| `cs` | Defaults to `en`; **required** for `.cs.md` files |
| `translation_of` | relative path | For `.cs.md` files, points to the canonical English original — **required** when `language: cs` |
| `superseded_by` | relative path | **Required** when `status: superseded` |
| `created` | `YYYY-MM-DD` | Creation date |
| `last_updated` | `YYYY-MM-DD` | Last meaningful update |
| `applies_to` | list of globs | For methodologies, the file globs they govern (e.g. `["**/*.bpmn"]`) |
| `related` | list of paths | Other docs to consult alongside this one |
| `tags` | list of strings | Free-form tags |
| `ai_load_overrides` | map | Per-assistant `ai_load` override (see [Per-assistant overrides](#per-assistant-overrides)) |

## `ai_load` values

| Value | When AI assistants load it | Typical category |
| ----- | -------------------------- | ---------------- |
| `eager` | On any contextually-relevant work touching `.md` files (style, naming, frontmatter, terminology) | `style-guide` |
| `scoped` | Only when the document's load-scope blockquote matches the current task | `methodology` |
| `router` | When navigating from a higher-level question; the file mostly links elsewhere | `index` |
| `on-demand` | Only when explicitly referenced by the user, by another doc, or by a task description | `reference`, `adr` |
| `never` | Never auto-load (archived, internal-only, work-in-progress) | `deprecated`, `draft` |

If `ai_load` and the load-scope blockquote disagree, the **blockquote wins** — it
carries more nuance than a single enum can express.

### Per-assistant overrides

`ai_load` applies uniformly to all supported assistants. In rare cases an assistant
has different context-window or tooling constraints that justify a different load
directive. Use the optional `ai_load_overrides` map:

```yaml
ai_load: scoped
ai_load_overrides:
  claude_code: scoped   # explicit, same as default
  gemini: on-demand     # tighter context budget on this assistant
```

Canonical assistant keys:

| Key | Refers to |
| --- | --------- |
| `claude_code` | Anthropic Claude Code (CLI / IDE) |
| `gemini` | Google Gemini Code Assist |

Use overrides sparingly. If most files override for a given assistant, fix the base
`ai_load` value or update the assistant's row in [Supported AI assistants](#supported-ai-assistants)
instead — overrides are escape hatches, not the primary lever.

## Rules

### Placement and format

- Frontmatter MUST be the very first content of the file (no leading blank lines, no BOM)
- Open and close with `---` on their own lines
- YAML 1.2 syntax; double-quote strings only when needed (leading colon, special characters)
- One blank line between the closing `---` and the H1 title

### Content

- Every project `.md` SHOULD have frontmatter. **Exempt:** root `README.md` of the
  repository, `CLAUDE.md` files (those use a different load mechanism), generated docs,
  and vendored third-party `.md`.
- All required fields MUST be present, even if `description` is short
- `description` is the most load-bearing field — AI assistants use it to decide
  whether to read further. **Lead with the use case** ("Apply when…", "Use when…"),
  not the topic ("About X")
- For `.cs.md` files: `language: cs` AND `translation_of: <original>.md` are both required
- Dates use ISO `YYYY-MM-DD`; never natural-language dates

### Lifecycle

- New documents start as `status: active`, or `status: draft` if explicitly work-in-progress
- Deprecated documents keep their content but flip to `status: deprecated` with a deprecation note at the top of the body
- Superseded documents add `superseded_by: <new-file>.md` and a redirect note in the body

### Synchronizing translations

When updating an English `.md` that has a `.cs.md` sibling (or vice versa), update both
within the same PR. If only one side is updated, mark the other with `tags: [out-of-sync]`
and open an issue — do not let translations silently diverge. The English version is
authoritative when content disagrees.

### Load-scope blockquote convention

The blockquote that follows the H1 should address AI assistants generically:

```markdown
> **AI assistant load scope (Claude Code, Gemini):**
>
> **Load when:** ...
>
> **Do NOT load for:** ...
>
> **Authoritative scope:** ...
```

If a specific assistant should behave differently, add a sub-paragraph inside the
blockquote (`> **Gemini-specific:** …`) rather than splitting the file's guidance
across multiple top-level callouts.

## Examples

### Methodology with explicit scope

```yaml
---
title: BPMN Methodology
description: Modeling rules, naming conventions, and Camunda 8 integration for BPMN/DMN files. Apply when creating, editing, or reviewing .bpmn / .dmn files.
category: methodology
ai_load: scoped
status: active
language: en
created: 2026-03-15
last_updated: 2026-05-07
applies_to:
  - "**/*.bpmn"
  - "**/*.dmn"
related:
  - PROCESS-ANALYSIS-METHODOLOGY-ICT.cs.md
  - USE-CASE-METHODOLOGY.md
---
```

### Czech translation

```yaml
---
title: Capability Registry — Metodika
description: Český překlad metodiky Capability Registry. Načti při onboardingu nebo úpravě konvencí; pro tvorbu konkrétních YAML souborů použij capability-registry/INDEX.md.
category: methodology
ai_load: scoped
status: active
language: cs
translation_of: CAPABILITY-REGISTRY-METHODOLOGY.md
created: 2026-04-10
last_updated: 2026-05-07
---
```

### Index / router

```yaml
---
title: Contributing
description: Router for contributing guides, methodologies, and conventions. Load when locating which guide applies to a task.
category: index
ai_load: router
status: active
language: en
created: 2026-05-07
last_updated: 2026-05-07
---
```

### Style guide (eager-load)

```yaml
---
title: Markdown Style Guide
description: Markdown formatting rules — headings, tables, code blocks, links, linting. Apply to any .md edit.
category: style-guide
ai_load: eager
status: active
language: en
created: 2026-03-29
last_updated: 2026-05-07
---
```

### Per-assistant override

```yaml
---
title: Long Reference Catalog
description: Catalog of every external system; load when answering integration or vendor questions.
category: reference
ai_load: on-demand
ai_load_overrides:
  gemini: never   # too large for default Gemini context budget
status: active
language: en
created: 2026-04-15
last_updated: 2026-05-07
---
```

### Deprecated document

```yaml
---
title: Old Issue Format
description: Historical issue format used before INT-XXX numbering was introduced. Kept for reference; do not apply to new issues.
category: workflow
ai_load: never
status: deprecated
language: en
created: 2025-08-12
last_updated: 2026-02-20
---
```

### Superseded document

```yaml
---
title: Inbound Mail Triage (combined)
description: Original combined inbound-mail process. Superseded by per-channel processes after 2026-05-06 1:1 with J. Khýrová.
category: specification
ai_load: never
status: superseded
superseded_by: ../pft-infinitecare/processes/as-is/correspondence/paper-mail.md
language: cs
created: 2026-04-30
last_updated: 2026-05-06
---
```

## Migration

Files without frontmatter get it added on their next non-trivial edit. Mass migration
is not required; opportunistic backfill keeps churn low and reviews focused.

Backfill procedure:

1. Read the file and identify `category` from location and content
2. Determine `ai_load` from the existing load-scope blockquote (if any) or from the
   file's purpose:
   - Style guide / convention → `eager`
   - Methodology with stated scope → `scoped`
   - Index / hub / table-of-contents → `router`
   - Reference, ADR, archived → `on-demand` or `never`
3. Pull `created` from git: `git log --diff-filter=A --follow --format=%ad --date=short -- <path>`
4. Set `last_updated` to today
5. Add frontmatter; **do not touch other content** in the same commit (one concern per PR)

If a legacy file's load-scope blockquote names only one assistant (e.g. "Claude Code —
when to load…"), generalize it to "AI assistant load scope (Claude Code, Gemini)"
in the same PR — see the
[Load-scope blockquote convention](#load-scope-blockquote-convention).

## Versioning the schema

Schema changes happen in this document. Rules:

- **Additive change** (new optional field, new assistant key) — update this doc;
  existing files stay valid
- **New required field** — update this doc, add a migration note in this section,
  backfill all files in one or more follow-up PRs
- **Renamed or removed field** — bump `last_updated` here and migrate all files in a
  single PR (no intermediate state where some files are valid and others are not)

When updating the schema, also re-check [README.md](README.md) and
[MARKDOWN-STYLE.md](MARKDOWN-STYLE.md) for cross-references that may need to follow.

### Schema history

| Date | Change | Migration |
| ---- | ------ | --------- |
| 2026-05-07 | Initial schema published | — |
| 2026-05-07 | `claude_load` → `ai_load` (generalized for Claude Code + Gemini) | Renamed field; allowed values unchanged. Files still using `claude_load` should be migrated on next edit. |

## Related

- [MARKDOWN-STYLE.md](MARKDOWN-STYLE.md) — body formatting rules (headings, tables, code blocks, links, linting)
- [README.md](README.md) — contributing index and document categorization
- [YAML 1.2 specification](https://yaml.org/spec/1.2.2/) — syntax reference

---

**Created**: 2026-05-07
**Last Updated**: 2026-05-07
