# Acceptance Protocol - Issue #035 (AI Assistant Detection)

**Issue**: AI Assistant Installation Support — enhanced AI assistant detection
(scope: `portunix package detect`)
**Branch**: `feature/035-ai-assistant-detection` (commit `a7dd6e8`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-05-27
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — PowerShell 5.1 + Go 1.25.0 toolchain

## Scope Note

This protocol covers only the **enhanced AI assistant detection** slice of
issue #035 (`portunix package detect`, table + `--json`). The AI assistant
bundles (`ai-assistant-basic`, `ai-assistant-full`, `mcp-ready`) were accepted
separately in `acceptance-035.md`. Still out of scope (remains open in the
issue): `--recommend-ai`, MCP `serve init` dependency hook, per-assistant
version management.

## Test Summary

- Total test scenarios: 11
- Passed: 11
- Failed: 0
- Conditional: 0

## Test Environment

- Build: `GOTOOLCHAIN=go1.25.0 go build` (helper module `ptx-installer`)
- Main `portunix.exe` built via `make build` (exit 0)
- Registry: embedded assets, 69 packages loaded, 0 errors
- Note on execution: `ptx-installer.exe` triggers the Windows Installer
  Detection heuristic (binary name contains "install") → UAC elevation, so it
  cannot run non-interactively. The helper was therefore built under a neutral
  name (`ptxdetect-test.exe`) for live testing; the dispatcher only forwards
  identical arguments, so this does not affect the exercised code path. Verdict:
  not a defect of this change (matches the note in `acceptance-035.md`).

## Test Results

### T1 — Build Sanity

- [x] Helper `ptx-installer` builds cleanly at HEAD (exit 0)
- [x] `go vet ./engine/ ./` clean (exit 0)
- [x] Commit `a7dd6e8` present in branch history

### T2 — Unit: package selection (`selectAIAssistants`)

- [x] `TestSelectAIAssistants_FiltersBundlesAndCategory` PASS — only
      `Kind:"Package"` in `development/ai-tools` selected; bundle and
      foreign-category package excluded; result sorted by name
- [x] `TestSelectAIAssistants_Empty` PASS — no matches handled gracefully

### T3 — Unit: version parsing (`parseVersionOutput`)

- [x] PASS (6 sub-cases) — 3-part `1.2.3`, 2-part `0.4`, prefixed `v20.11.1`,
      path (no version) → empty, empty input → empty, whitespace → empty

### T4 — Unit: path-lookup detection (`isPathLookup`)

- [x] PASS — `where`/`which` (any case, leading spaces) → true;
      `--version` commands → false; empty → false

### T5 — Unit: platform command selection (`verifyCommandForOS`)

- [x] PASS — windows command resolved; `windows_sandbox` → windows fallback;
      platform without verification → empty; unknown platform → empty

### T6 — Integration: end-to-end orchestration (`DetectAIAssistants`)

- [x] `TestDetectAIAssistants_EndToEnd` PASS — registry loaded from temp assets
      (2 packages + 1 bundle); returns exactly 2 statuses (bundle excluded), in
      name order, with verification command populated; verified for both
      `linux` and `windows`

### T7 — Live: `package detect` (table)

- [x] Given the embedded registry (69 packages) on a host with Claude Code
      installed, When `package detect` runs, Then it reports:
      - `claude-code` → ✅ installed (2.1.152)
      - `claude-desktop` → ⬜ not found
      - `gemini-cli` → ⬜ not found
      - summary `Detected 1 of 3 AI assistant(s) installed.`
- [x] Exactly 3 entries shown — bundles correctly excluded
- [x] Version `2.1.152` correctly parsed from `claude --version`
- [x] Exit code 0

### T8 — Live: `package detect --json`

- [x] Emits valid JSON array of 3 objects with fields `name`, `displayName`,
      `category`, `installed`, `version` (omitted when empty), `verifyCommand`
- [x] `claude-desktop` shows `"verifyCommand": "where claude-desktop"` and no
      `version` key (path-lookup → version intentionally omitted)
- [x] Exit code 0

### T9 — Live: help integration

- [x] `package --help` lists the new `detect` subcommand + examples
- [x] `package detect --help` shows package help (consistent with existing
      `list`/`search`/`info` — see Known Limitations)

### T10 — Failure injection: unknown subcommand

- [x] `package detectx` → `Unknown package subcommand: detectx` + guidance,
      no crash

### T11 — Regression

- [x] `package list` still reports `Total packages: 69`
- [x] `package info claude-code` unchanged (Name/Display Name correct)
- [x] `engine` + `registry` test suites PASS (network download tests excluded —
      pre-existing socket-dial timeout in the restricted sandbox, unrelated to
      this change; this change does not touch download logic)
- [x] No AI assistant package JSON definitions were modified

## Known Limitations / Notes

1. **`package detect --help` shows package-level help** rather than the
   detect-specific help block. This is because `handlePackage` intercepts
   `--help`/`-h` before dispatching to the subcommand — the *same* behavior
   already exhibited by `list`, `search` and `info`. The detect-specific help
   text exists in `handlePackageDetect` but is currently shadowed. Pre-existing
   inconsistency, not introduced by this change. Verdict: not a blocker;
   recommend a separate cleanup issue if per-subcommand help is desired.
2. **Detection runs each verification command** (e.g. `claude --version`). This
   is a read-only probe and safe, but detection time scales with the number of
   AI packages × command startup latency (observed sub-second for 3 packages).
3. **UAC elevation** of the production-named `ptx-installer.exe` (see Test
   Environment) — Windows installer-detection heuristic, not a code defect.

## Recommendations

- Optional follow-up: fix the shadowed per-subcommand `--help` across all
  `package` subcommands in one pass (cleanup issue), since it now affects
  `detect` too.
- When `--recommend-ai` / the MCP `serve init` hook are implemented, they should
  consume the `--json` output verified here as their detection source.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES (for the detection scope; remaining issue features
stay open)
**Date**: 2026-05-27
**Tester signature**: zdendaku (QA/Test Engineer — generic)
