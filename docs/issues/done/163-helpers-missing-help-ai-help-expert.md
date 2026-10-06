# Issue #163: Helpers Missing --help-ai and --help-expert Flags

**Date**: 2026-04-02
**Reporter**: preflight check
**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Labels**: helpers, help-system, consistency, preflight
**Affects**: portunix v2.1.0

## Problem

6 out of 12 helper binaries (ptx-*) are missing `--help-ai` and `--help-expert` flags.
These flags are required for consistent multi-level help system across all helpers.

The preflight check (`scripts/github-01-preflight-check.sh`) now enforces this and
blocks GitHub publication until all helpers implement both flags.

## Affected Helpers

| Helper | --help-ai | --help-expert |
| ------ | --------- | ------------- |
| ptx-aiops | OK | OK |
| ptx-ansible | OK | OK |
| ptx-container | OK | OK |
| ptx-credential | MISSING | MISSING |
| ptx-installer | OK | OK |
| ptx-make | MISSING | MISSING |
| ptx-mcp | MISSING | MISSING |
| ptx-pft | OK | OK |
| ptx-prompting | MISSING | MISSING |
| ptx-python | OK | OK |
| ptx-trace | MISSING | MISSING |
| ptx-virt | MISSING | MISSING |

## Required Changes

For each affected helper, implement `--help-ai` and `--help-expert` flags following the
pattern used in helpers that already have them (e.g. ptx-container, ptx-installer, ptx-python).

### Implementation Pattern

1. Add `--help-ai` flag that outputs AI-friendly help (structured, machine-parseable)
2. Add `--help-expert` flag that outputs detailed expert-level help
3. Follow existing implementation in a working helper as reference

### Helpers to Fix

1. **ptx-credential** - both flags missing
2. **ptx-make** - both flags missing
3. **ptx-mcp** - both flags missing
4. **ptx-prompting** - both flags missing
5. **ptx-trace** - both flags missing
6. **ptx-virt** - both flags missing

## Acceptance Criteria

- All 12 helpers pass `--help-ai` and `--help-expert` without error
- Preflight check passes: `./scripts/github-01-preflight-check.sh`
- Help output is meaningful (not empty or generic)

## Notes

- This issue was discovered during preflight check improvement
- Blocking GitHub publication until resolved
