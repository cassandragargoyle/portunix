# Issue #176: `portunix install docker` Missing Installation Directory Prompt

**Type:** Bug Fix
**Priority:** High
**Status:** ✅ Implemented
**Closed:** 2026-04-21
**Labels:** bug, installation, docker, ptx-installer, user-experience, regression

## Description

When running `portunix install docker`, the installation starts immediately without
giving the user any option to choose the target installation directory (or data-root
storage location). Historically, Docker installation included intelligent storage
selection logic (see issues #019 and #122) that prefers non-system drives with
sufficient free space for `data-root`. After consolidation into `ptx-installer`
(#122), this interactive selection appears to be missing — the user is no longer
prompted at all.

## Steps to Reproduce

1. Run `portunix install docker` on a system with multiple drives (e.g. Windows with
   C:\ and D:\ drives, both with sufficient free space)
2. Observe that the installation proceeds immediately
3. No prompt is shown asking where to install Docker or where to place `data-root`

## Expected Behavior

Before starting the actual installation, the installer should:

- Enumerate available drives/partitions
- Filter out drives with insufficient free space (Docker requires ≥ 10 GB)
- Offer the user a choice of target directory (with a sensible default)
- Allow override via CLI flag (e.g. `--data-root <path>` or `--install-dir <path>`)
  for non-interactive / scripted runs
- Respect `--yes` / non-interactive mode by falling back to the default without
  prompting

## Actual Behavior

Installation starts immediately using hardcoded defaults (system drive / default
Docker location). The user is never asked and has no opportunity to intervene.

## Environment

- **Command:** `portunix install docker`
- **Installer path:** `ptx-installer` helper binary (post-#122 consolidation)

## Root Cause Hypothesis

During the migration of Docker installation logic from `app/docker/` into
`src/helpers/ptx-installer/` (issue #122), the interactive storage selection flow
(previously implemented as `analyzeWindowsStorage()` and companion logic) may have
been dropped or skipped. The package-manifest-driven installation path used by
`portunix install <package>` does not currently expose a prompt hook for packages
that need directory selection.

This is likely a regression relative to the behavior documented in issues #019 and
#122.

## Acceptance Criteria

- [ ] `portunix install docker` prompts the user to choose the installation /
      `data-root` directory before starting the actual install
- [ ] The prompt shows only drives with sufficient free space (≥ 10 GB)
- [ ] A sensible default is pre-selected (preferably the largest non-system drive
      with enough space, matching prior behavior)
- [ ] CLI flag support for non-interactive selection (e.g. `--data-root <path>`)
- [ ] Non-interactive mode (`--yes` / CI) falls back to default without hanging on
      the prompt
- [ ] Works on both Windows and Linux
- [ ] Behavior is consistent between `portunix install docker` and the legacy
      `portunix docker install` (if the latter is still supported)

## Related Issues

- #019 — Docker Installation Issues on Windows (original storage-selection logic)
- #122 — Consolidate Docker/Podman Installation into ptx-installer (migration that
  likely introduced this regression)
- #123 — Consolidate Installation Systems

## Notes

This issue blocks users who want Docker's `data-root` on a non-system drive
(common scenario on Windows developer machines where C:\ has limited space and
D:\ is dedicated to workloads/data).

---
**Created:** 2026-04-21
**Reporter:** @zdendaku
