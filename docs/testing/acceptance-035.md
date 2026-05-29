# Acceptance Protocol - Issue #035

**Issue**: AI Assistant Installation Support — AI assistant bundles (scope: profiles)
**Branch**: `feature/035-ai-assistant-bundles` (commits `1d72376` + tests in `8e201ed`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-05-26
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — Git Bash + Go 1.25.0 toolchain

## Scope Note

This protocol covers only the **AI assistant bundle** slice agreed for this
iteration (`ai-assistant-basic`, `ai-assistant-full`, `mcp-ready`). The package
definitions for `claude-code`, `claude-desktop` and `gemini-cli` already existed
prior to this work. Out of scope (remains open in the issue): assistant
detection, `--recommend-ai`, MCP `serve init` hook, version management.

## Test Summary

- Total test scenarios: 9
- Passed: 9
- Failed: 0
- Conditional: 0

## Test Environment

- Build command: `GOTOOLCHAIN=go1.25.0 go build` (helper module `ptx-installer`)
- Binaries: `ptx-installer.exe` (≈13,6 MB), `portunix.exe` (≈23,0 MB)
- Registry: embedded assets, 69 packages loaded, 0 errors
- Note: host-level installation testing limited to `--dry-run` (no real software
  installed on host, per Testing Methodology). Real install of members should be
  verified in a container (see Recommendations).

## Test Results

### T1 — Build Sanity

- [x] Helper `ptx-installer` builds cleanly at HEAD (exit 0, no `-mod=mod` needed
      after go.mod bump to `go 1.25.0`)
- [x] Main `portunix.exe` builds cleanly (exit 0)
- [x] Bundle commit `1d72376` present in branch history

### T2 — Bundle Definitions Load

- [x] All three bundle JSON files validate and load (69 packages, 0 errors)
- [x] `Kind: "Bundle"` accepted by registry validation
- [x] Bundle metadata (displayName, description, category) rendered correctly

### T3 — `ai-assistant-basic` (dry-run)

- [x] Resolves bundle, expands to 2 members in order: `claude-code`, `gemini-cli`
- [x] Each member runs through normal install flow (variant auto-detect, dry-run)
- [x] Reports `✅ Bundle ai-assistant-basic complete (2 packages)`

### T4 — `ai-assistant-full` (dry-run)

- [x] Expands to 3 members in order: `claude-code`, `claude-desktop`, `gemini-cli`
- [x] `[1/3] [2/3] [3/3]` progress indicators correct
- [x] Reports completion (3 packages)

### T5 — `mcp-ready` (dry-run)

- [x] Expands to 3 members in order: `python`, `claude-code`, `gemini-cli`
- [x] Mixed install types resolved (python: zip, claude-code: script, gemini: npm)
- [x] Reports completion (3 packages)

### T6 — Failure Injection: non-existent bundle/package

- [x] `install nonexistent-bundle-xyz` fails gracefully
- [x] Error: `package not found: package 'nonexistent-bundle-xyz' not found`
- [x] Exit code 1 (no partial side effects)

### T7 — Failure Injection: bundle with missing member (stop-on-error)

- [x] Covered by `TestInstallBundle_MissingMemberFails` (engine)
- [x] Bundle iteration stops at missing member, error names `does-not-exist`
- [x] Wrapped error: `bundle <name>: failed to install <member>: ...`

### T8 — Unit / Validation Tests

- [x] `TestValidatePackage_Bundle` PASS — valid bundle, empty bundle rejected,
      `Kind: Package` still requires platforms, unknown `Kind` rejected
- [x] `registry` package suite PASS
- [x] `engine` package suite PASS (network download tests excluded — see below)

### T9 — Regression

- [x] Existing packages unaffected (regular `Kind: Package` flow unchanged)
- [x] No changes to download / archive / script install paths
- [x] `claude-code` / `claude-desktop` / `gemini-cli` standalone dry-runs unchanged

## Known Limitations / Notes

1. **Dispatcher elevation on Windows**: invoking `portunix install <bundle>`
   through the dispatcher triggers UAC elevation for the helper in the test
   environment. The bundle logic was therefore exercised directly via
   `ptx-installer.exe` (the dispatcher only forwards identical arguments).
   Verdict: not a defect of this change.
2. **Network download tests time out**: `TestInstallDownload_*` (pre-existing,
   `httptest`-based) hang on socket dial in the network-restricted sandbox.
   Unrelated to this change — bundle code does not touch download logic.

## Recommendations

- Real end-to-end install of bundle members (actual `npm install -g`,
  Claude Desktop installer, etc.) should be verified in a clean Ubuntu/Windows
  container per the Container-Based Testing Policy before relying on it in
  production profiles.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES (for the bundle scope; remaining issue features stay open)
**Date**: 2026-05-26
**Tester signature**: zdendaku (QA/Test Engineer — generic)
