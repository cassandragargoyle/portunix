---
title: Issue Management
description: Workflow for creating, tracking, and archiving issues in portunix — GitHub-first, where GitHub assigns the number and each issue keeps a long-form write-up in docs/issues/NNN-*.md (archived to docs/issues/done/ on completion). No overview tables.
applies_to:
  - docs/issues/**
type: contributing-guide
status: active
last_updated: 2026-10-07
---

# Issue Management Guidelines

## Purpose

This document defines how issues are created, tracked, and archived in Portunix.

> **GitHub is the single source of truth for the issue list.** Status, labels,
> and assignee live on GitHub. The repository keeps only the long-form write-up
> for each issue. There are **no overview tables** and no mapping files — a shared
> Markdown table edited by everyone caused constant merge conflicts and was
> removed.

- Open issues: <https://github.com/cassandragargoyle/portunix/issues>
- Closed issues: <https://github.com/cassandragargoyle/portunix/issues?q=is%3Aissue+is%3Aclosed>

## GitHub-First Workflow

The issue is **created on GitHub first**, so **GitHub assigns the number `N`**.
The GitHub issue carries only the basics (title, a short description, and a link
to the detail file). The full write-up lives in `docs/issues/NNN-name.md`, where
`NNN` is the GitHub issue number zero-padded to three digits.

The `create-issue` skill (`/create-issue`) automates this flow.

### 1. Create the GitHub issue

```bash
gh issue create --repo cassandragargoyle/portunix \
  --title "Short descriptive title" \
  --body "One-line summary.

Full details: docs/issues/NNN-name.md" \
  --label enhancement,ptx-installer
```

Take `N` from the URL that `gh` prints. Never invent a number.

### 2. Write the detail file

Create `docs/issues/NNN-name.md` with the `# Issue #N: Title` heading, problem
description, acceptance criteria, and testing notes (see [Issue Template](#issue-template)).

### 3. Link the detail file from GitHub

```bash
gh issue edit N --repo cassandragargoyle/portunix --body "One-line summary.

Full details: https://github.com/cassandragargoyle/portunix/blob/main/docs/issues/NNN-name.md"
```

The link starts working once the file reaches `main` on GitHub (next
`deploy-github` sync).

### 4. Reference the number

- Branch: `feature/issue-N-short-name`
- Commits: `feat(N): …`, `fix(N): …`, `refs #N`, `closes #N`

### 5. Archive on completion

When the issue reaches `✅ Implemented` or `❌ Closed`:

```bash
git mv docs/issues/NNN-name.md docs/issues/done/NNN-name.md
gh issue close N --repo cassandragargoyle/portunix --comment "Implemented in <version/commit>."
```

Update the GitHub body link to the `done/` path when archiving.

## Publication and Sensitive Content

`docs/issues/` is **published to GitHub** together with the rest of the repository.
Write every issue as public content:

- No customer names, business relationships, deals, or priorities tied to a customer
- No third-party or employer project names and internal identifiers
- No local paths such as `/home/<user>/…` — use repository-relative paths
- No internal hostnames, IP addresses, credentials, or tokens (placeholders like
  `<token>` or `example.com` are fine)

The GitHub preflight check (`scripts/github-01-preflight-check.sh`, patterns in
`scripts/github-private-files.json`) scans published files for sensitive
patterns. Content that must stay internal belongs in `docs/private/`, not in an
issue file.

## File Structure

```text
docs/issues/
├── NNN-name.md          # Active issue write-up (NNN = GitHub issue number)
└── done/                # Archived issues (✅ Implemented / ❌ Closed)
    ├── NNN-name.md
    └── NNN-*-acceptance.md, NNN-*-report.md   # Supporting files of issue NNN
```

## Numbering

- GitHub issues and pull requests share one counter, so numbers skip where a PR
  consumed one.
- Files use three-digit zero padding (`009-…`, `150-…`); GitHub references do not
  (`#9`, `#150`).

### Legacy numbering

Until 2026-10 issues were numbered internally (`docs/issues/internal/`) with a
separate `PUB-` mapping for GitHub. During the migration to GitHub-first:

- Internal numbers `035`–`203` were created on GitHub with identical numbers.
- `001`–`004`, `007`–`010`, `012`–`015`, `019`, and `022` already existed on GitHub
  with the same number (`#22` is the external Docusaurus issue).
- Internal numbers that collided with existing GitHub issues or pull requests were
  renumbered to the end of the range. Older documents (CHANGELOG, ADRs, notes,
  acceptance protocols) still use the legacy numbers:

| Legacy | GitHub | Title |
| ------ | ------ | ----- |
| #023 | #204 | Arch Linux Distribution Support Integration |
| #016 | #205 | Protocol Buffers Compiler (protoc) |
| #017 | #206 | QEMU/KVM Windows 11 Virtualization with Snapshot Support |
| #020 | #207 | QEMU Windows VM Clipboard Integration |
| #021 | #208 | GitHub Actions Local Testing Support with Act |
| #022 | #209 | Google Chrome Installation Implementation |
| #024 | #210 | Plugin Registration and Discovery System |
| #025 | #211 | GitHub Integration for Portunix Core |
| #026 | #212 | GitHub CLI (gh) Installation Support |
| #027 | #213 | Container Lifecycle Management with Cleanup Guarantees |
| #028 | #214 | Universal Container Parameters Support |
| #029 | #215 | Universal Container Command Implementation |
| #030 | #216 | Container TLS Certificate Verification Failure |
| #031 | #217 | Universal Container Exec Command |
| #032 | #218 | Universal Container Management Commands |
| #033 | #219 | MCP Server Plugin Development Guide for AI Agents |
| #034 | #220 | MCP Server Wizard — Advanced Features Completion |
| #133 (run forwarding) | #221 | Plugin Run Command Argument Forwarding |

A duplicate stale copy of `#103 PTX-Make Helper Implementation` (legacy `#102`
active file) was dropped; legacy `#102` now refers only to Compose Command
Implementation.

## Labels

Every issue gets **one type label** and **one or more component labels**:

- **Type**: `bug`, `enhancement`, `documentation`
- **Component** — the binary the issue concerns:
  - `portunix` — main binary (dispatcher, plugin system, update, system info,
    release and build tooling)
  - `ptx-<helper>` — one label per helper binary in `src/helpers/`:
    `ptx-aiops`, `ptx-ansible`, `ptx-container`, `ptx-credential`,
    `ptx-database`, `ptx-github`, `ptx-installer`, `ptx-make`, `ptx-mcp`,
    `ptx-pft`, `ptx-plugin-registry`, `ptx-prompting`, `ptx-proxmox`,
    `ptx-python`, `ptx-specpm`, `ptx-ssh`, `ptx-trace`, `ptx-virt`,
    `ptx-vocalio`, `ptx-wizard`

Filter per binary with `is:issue label:ptx-installer`. When a new helper is added,
create its label:

```bash
gh label create ptx-<name> --repo cassandragargoyle/portunix \
  --color 1d76db --description "Helper binary ptx-<name>"
```

## Issue Template

```markdown
# Issue #N: Title

**Status**: Open
**Priority**: Low | Medium | High | Critical
**Type**: Bug | Enhancement | Documentation
**Component**: portunix | ptx-<helper>
**Created**: YYYY-MM-DD
**GitHub**: #N
**Related**: #...

## Summary

## Problem Description

## Acceptance Criteria

1. ...

## Testing
```

### Formatting Rules

- **No icons or emoji** in issue files: not in headings (`## Priority`, not `## 🎯 Priority`),
  and not in the `Status` field (`Open`, not `📋 Open`)
- Older issues using emoji are legacy style — do not copy it into new issues
- Exception: quoted literal program output stays verbatim (e.g. `Status: ❌ Unhealthy`)

## Commit and Pull Request References

```bash
git commit -m "fix(42): container timeout (refs #42)"
git commit -m "feat(42): retry logic (closes #42)"
```

Pull requests on GitHub reference the issue with `Fixes #N` / `Refs #N`.
