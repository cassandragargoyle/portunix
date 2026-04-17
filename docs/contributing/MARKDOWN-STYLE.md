# Markdown Style Guide

## Overview

This document defines formatting rules for Markdown files in the Portunix ecosystem.
All documentation — specifications, component READMEs, contributing guides, ADRs — MUST
follow these conventions so that documents are consistent and machine-parseable.

## Document Categories

Different document types use different header structures. Pick the matching template.

### Specification (`docs/architecture/specifications/`)

```markdown
# Specification: Short Title

## Status

**Draft** — one-line explanation of current status

## Metadata

- **Created**: YYYY-MM-DD
- **Author**: Name
- **Target**: repository or module path (if applicable)
- **Related**:
  - [Link 1](relative-path.md)
  - [Link 2](relative-path.md)

---

## 1. Overview

Body starts here with numbered top-level sections.
```

Status values: **Draft**, **Active**, **Deprecated**, **Superseded by [link]**

### Component README (`docs/architecture/components/*/README.md`)

```markdown
# Component Name

## Overview

**Repository**: `repo-name`
**Language**: Language + version
**Status**: Early | Active | Stable | Deprecated

Description paragraph.

## Architecture Components

### Sub-component A

...
```

### Contributing Guide (`docs/contributing/`)

```markdown
# Guide Title

## Overview

Description paragraph.

## Section Heading

Content.
```

### ADR (`docs/adr/`)

Follow the standard ADR format: `# Context`, `# Decision`, `# Consequences`.

## Heading Rules

- **H1** (`#`): Document title only, exactly one per file
- **H2** (`##`): Major sections — use numbered form (`## 1. Overview`) in specifications, unnumbered in guides and READMEs
- **H3** (`###`): Subsections — numbered (`### 1.1`) in specifications, named in guides
- **H4** (`####`): Rarely needed, avoid deeper nesting
- Blank line before and after every heading

## Text Formatting

- One blank line between paragraphs
- No trailing whitespace
- Lines should not exceed ~100 characters in source (soft wrap is acceptable)
- Use `**bold**` for key terms on first mention, field names in prose, and status labels
- Use `` `backticks` `` for code identifiers, file paths, CLI commands, field values
- Use `*italic*` sparingly for emphasis in prose

## Lists

- Use `-` (hyphen) for unordered lists, not `*` or `+`
- Use `1.` for ordered lists (auto-numbering)
- One blank line before and after a list block
- Nested items indented by 2 spaces
- No blank lines between list items within the same level

```markdown
- First item
- Second item
  - Nested item
  - Another nested item
- Third item
```

## Tables

- Header separator row MUST use at least 5 hyphens per column (`-----`), never the
  minimal `---`. This improves readability in source and avoids markdownlint warnings.
- Align columns with spaces for readability in source
- Keep cell content short — move details to footnotes or subsections

**Correct:**

```markdown
| Field | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `name` | string | yes | Plugin identifier |
| `version` | string | yes | Semantic version |
```

**Wrong** (do not use `---` separators):

```markdown
| Field | Type | Required | Description |
|---|---|---|---|
```

## Code Blocks

- Always specify the language after opening backticks: ` ```json `, ` ```go `, ` ```bash `
- **Never use bare ` ``` ` without a language tag.** When the content has no specific
  language (plain-text diagrams, command output, pseudo-code), use ` ```text `
- Keep code blocks focused — show only the relevant fragment
- Indent code inside list items by 2 spaces to maintain list continuity

## Links

- Use relative paths for internal links: `[Plugin Manifest](plugin-manifest.md)`
- Use full URLs for external resources
- Multi-line Related lists use nested bullet form:

```markdown
  - **Related**:
    - [Link A](a.md)
    - [Link B](b.md)
```

- Single Related reference can be inline: `- **Related**: [Link A](a.md)`

## Metadata Block

Specifications and formal documents use a Metadata section with dash-prefixed bold labels:

```markdown
## Metadata

- **Created**: 2026-03-29
- **Author**: Zdenek
- **Target**: `portunix-reco` repository
- **Priority**: Medium
- **Related**:
  - [Doc A](a.md)
```

Only include fields that are relevant. Common fields:

| Field | When to use |
| ----- | ----------- |
| Created | Always |
| Author | Always |
| Target | Specs targeting a specific repo or module |
| Priority | When triaging or scheduling work |
| Related | Cross-references to other specs, ADRs, components |

## Section Numbering

- Specifications MAY use numbered sections: `## 1. Overview`, `### 1.1 Sub-topic`
  - Recommended for short, stable specs (up to ~8 sections) where it aids navigation
  - Optional for larger or frequently changing specs — descriptive headings are more
    maintainable because inserting a section does not force renumbering
  - When used, numbering starts at 1, subsections use dot notation (1.1, 1.2, 2.1)
  - Be consistent within a single document — either all H2 sections are numbered or none
- Contributing guides, READMEs, and ADRs use descriptive headings without numbers

## File Naming

- Contributing guides: `SCREAMING-KEBAB-CASE.md` (e.g. `MARKDOWN-STYLE.md`, `CODE-STYLE-GO.md`)
- Specifications: `kebab-case.md` (e.g. `plugin-manifest.md`, `task-platform-grpc.md`)
- Component READMEs: `README.md` inside the component directory
- PlantUML diagrams: `kebab-case.puml` alongside the referencing document

## PlantUML Files

- Use `@startuml Title` with a descriptive title
- Include `title` directive for rendered output
- Place in the same directory as the referencing spec or component README
- Name with `kebab-case.puml` suffix

## Footer

Specifications and component READMEs end with a creation/update line at the bottom,
separated by `---`:

```markdown
---

**Created**: 2026-03-29
**Last Updated**: 2026-03-29
```

Contributing guides do not need a footer.

## Linting

All Markdown files MUST pass `markdownlint-cli2` with the project configuration.
The same rule set is used in VS Code (via the markdownlint extension) and in CI.

### Setup

```bash
npm install -g markdownlint-cli2
```

### Usage

```bash
# Lint all Markdown files
markdownlint-cli2 "**/*.md"

# Lint a single file
markdownlint-cli2 docs/architecture/specifications/plugin-manifest.md
```

### Configuration

Project configuration lives in `.markdownlint.jsonc` at the repository root.
This file is shared between VS Code extension and CLI — any rule change applies
to both environments automatically.

#### Default Config

When creating a new Portunix repository, copy the following into `.markdownlint.jsonc`
at the repository root:

```jsonc
{
  // Line length: 120 chars, ignore tables and code blocks
  "MD013": {
    "line_length": 120,
    "tables": false,
    "code_blocks": false
  },
  // Allow hard tabs inside code blocks (Java/Go code style uses tabs)
  "MD010": {
    "code_blocks": false
  }
}
```

#### Rule Rationale

| Rule | Setting | Why |
| ----- | ------- | --- |
| MD013 | 120 chars, tables + code excluded | 80 is too strict for prose with URLs; tables and code have their own constraints |
| MD010 | tabs allowed in code blocks | Java and Go code style uses tabs for indentation |

#### Adding to a New Repository

1. Create `.markdownlint.jsonc` at the repo root with the config above
2. Install VS Code extension `davidanson.vscode-markdownlint` — it picks up
   the config automatically
3. Add `make lint` target to Makefile:

```makefile
.PHONY: lint lint-fix

lint:
	markdownlint-cli2 "**/*.md"

lint-fix:
	markdownlint-cli2 --fix "**/*.md"
```

### AI Assistant Integration

When creating or editing Markdown files, AI assistants SHOULD run
`markdownlint-cli2` on the changed files and fix any reported violations
before considering the task complete.
