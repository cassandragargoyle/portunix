# Acceptance Protocol - Issue #035 (AI Assistant macOS Install)

**Issue**: AI Assistant Installation Support — cross-platform install
(scope: macOS/`darwin` install for `claude-desktop` and `gemini-cli`)
**Branch**: `feature/035-ai-assistant-macos-install` (commit `b00c62b`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-05-27
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — PowerShell 5.1 + Go toolchain

## Scope Note

This protocol covers only the **macOS (`darwin`) install** slice of issue #035:
adding the `darwin` platform to the existing `claude-desktop.json` and
`gemini-cli.json` package definitions. The bundles and the detection slice were
accepted separately (`acceptance-035.md`, `acceptance-035-detection.md`). Still
open in the issue: `--recommend-ai`, MCP `serve init` dependency hook,
per-assistant version management.

## Critical Testing Constraint

The change is **macOS-only functionality**. Per the project testing
methodology, the testing OS must match the functionality under test. **No macOS
host or VM was available** for this session — testing ran on a Windows host.
Therefore:

- **Structural / static verification** (JSON schema, registry validation,
  install-script review) was performed and is reported below.
- **Functional install verification on macOS** (actual download, DMG mount,
  `/Applications` copy, `npm -g` install, version probe) **could NOT be
  executed** and is recorded as **BLOCKED — pending macOS tester**.

## Test Summary

- Total test scenarios: 8
- Passed (structural): 5
- Blocked (require macOS host): 3
- Failed: 0

## Test Environment

- JSON parse: PowerShell `ConvertFrom-Json`
- Registry load/validation: `go test` against on-disk assets
  (`LoadPackageRegistry("../assets")`) — 69 packages loaded, 0 errors
- Note: full `make build` not exercised — fails at `gen-resources` on this
  Windows host (`goversioninfo` / shell incompatibility), unrelated to this
  change which contains no Go code.

## Test Results

### T1 — JSON validity (structural) — PASS

- [x] `gemini-cli.json` parses; platforms = `[darwin, linux, windows]`;
      `darwin.type = npm`; `darwin.verification.command = gemini --version`
- [x] `claude-desktop.json` parses; platforms = `[darwin, linux, windows]`;
      `darwin.type = script`;
      `darwin.verification.command = test -d '/Applications/Claude.app'`

### T2 — Registry validation (structural) — PASS

- [x] `LoadPackageRegistry` loads all 69 packages with **0 errors** — the new
      `darwin` platforms pass schema validation (valid platform name `darwin`,
      valid types `npm`/`script`, each variant has an install method, each
      platform has a verification command)
- [x] `claude-desktop` → `darwin.type == "script"`, has variants, has
      verification command
- [x] `gemini-cli` → `darwin.type == "npm"`, has variants, has verification
      command

### T3 — Static review: `gemini-cli` darwin (structural) — PASS

- [x] Install method `npm install -g @google/gemini-cli` is identical to the
      already-shipping `linux`/`windows` variants — npm is cross-platform, so
      risk is low
- [x] `preInstall` guards on `which node`; dependency `nodejs` declared
- [x] Verification `gemini --version` consistent with other platforms

### T4 — Static review: `claude-desktop` darwin (structural) — PASS (with notes)

- [x] DMG flow mirrors the proven `double-commander.json` darwin pattern:
      `curl -fL` (fail-on-error + follow-redirect) from `https://claude.ai/download/mac`
      → `hdiutil attach -nobrowse -plist` → mount-point parsed via `sed` →
      `find … -maxdepth 1 -name '*.app'` (robust, not hard-coded `Claude.app`
      on the volume) → `sudo cp -R` to `/Applications/` → `hdiutil detach` →
      `rm` temp
- [x] Verification target `/Applications/Claude.app` matches the `cp`
      destination — internally consistent
- Notes / observations (not blockers, flagged for the macOS tester):
  1. `sudo cp -R` requires elevation → **interactive password prompt**; will
     block fully non-interactive automation. Manual install is fine.
  2. `mktemp -t claude-desktop` creates an empty temp file; the script then
     writes the DMG to `"$TMP".dmg` (a different path) and only `rm "$TMP"` the
     `.dmg` — same harmless orphaned-tempfile behavior as the reference pattern.
  3. The exact volume/app name produced by the Claude DMG is **unverified**
     (no macOS host); the `find *.app` approach is designed to tolerate it but
     must be confirmed on real hardware.

### T5 — Regression (structural) — PASS

- [x] Only two package JSON files changed (`git show --stat` → 2 files,
      33 insertions, 0 deletions)
- [x] No Go code modified; no bundle definitions modified
- [x] Existing `windows`/`linux` platforms of both packages unchanged

### T6 — Functional: `portunix install gemini-cli` on macOS — BLOCKED

- [ ] Requires macOS host with Node.js — verify `gemini --version` after install
- Reason: no macOS environment available this session

### T7 — Functional: `portunix install claude-desktop` on macOS — BLOCKED

- [ ] Requires macOS host — verify DMG download, mount, copy to `/Applications`,
      `Claude.app` present, clean detach/cleanup, idempotent re-install
- Reason: no macOS environment available this session

### T8 — Functional: `portunix package detect` on macOS — BLOCKED

- [ ] On macOS, detection should run the `darwin` verification commands and
      report install state for both assistants
- Reason: detection resolves the platform from `runtime.GOOS`; the `darwin`
  branch cannot be exercised from a Windows host

## Recommendations

- Route this branch to a **macOS tester** (or macOS VM/runner) to complete
  T6–T8 before merge to `main`.
- Suggested macOS verification steps:
  1. `portunix install gemini-cli` → expect `gemini --version` to succeed
  2. `portunix install claude-desktop` → expect `/Applications/Claude.app` to
     exist; confirm no leftover `/Volumes/...` mount and temp `.dmg` removed
  3. `portunix package detect` → both assistants reported with correct state
  4. Re-run `claude-desktop` install to confirm `sudo rm -rf` makes it idempotent
- Confirm `https://claude.ai/download/mac` serves a DMG (HTTP 200 after
  redirects) and note the actual `.app` bundle name observed.

## Final Decision

**STATUS**: CONDITIONAL

Structural and static verification PASS. Functional macOS install verification
is BLOCKED pending a macOS test environment.

**Approval for merge to `main`**: NO — withheld until T6–T8 pass on a macOS host.
**Date**: 2026-05-27
**Tester signature**: zdendaku (QA/Test Engineer — generic)
