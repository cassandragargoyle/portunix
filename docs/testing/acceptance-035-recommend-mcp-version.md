# Acceptance Protocol - Issue #035 (Recommend AI / MCP Auto-Dependency / Version Selection)

**Issue**: AI Assistant Installation Support — `--recommend-ai`, MCP auto-dependency
hook, and per-assistant version selection
**Branch**: `feature/035-ai-recommend-mcp-version` (commit `9f07c60`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-05-29
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — Go 1.25.0 toolchain

## Scope Note

This protocol covers the third slice of issue #035:

- `portunix install --recommend-ai` — list installable AI assistants for the
  platform with install state and offer to install the missing ones
  (`--dry-run` / `-y`).
- `portunix mcp init --install-missing` + the interactive wizard's
  missing-assistant offer (auto-dependency hook).
- `portunix install <pkg> --version <v>` (resolution by variant name or by the
  variant's `version` field) and `portunix package update <pkg>` (Force
  reinstall of the preferred variant).

Earlier slices were accepted separately: AI assistant bundles
(`acceptance-035.md`), detection (`acceptance-035-detection.md`), macOS install
(`acceptance-035-macos.md`). Still open in the issue: full version
pinning/downgrade with installed-state tracking and uninstall.

## Test Summary

- Total test scenarios: 13
- Passed: 12
- Conditional: 1 (real install/update execution + cross-platform — see below)
- Failed: 0

## Test Environment

- Helpers `ptx-installer` and `ptx-mcp` built cleanly with Go 1.25.0
  (`go build`, exit 0). Main `portunix.exe` built via `make build` (exit 0).
- Registry: embedded assets, 69 packages loaded, 0 errors.
- **Build note (not a #035 defect):** `make build` aborts at
  `build-helpers` on `ptx-prompting` — its `go.mod` pins `toolchain go1.24.2`
  while the parent module requires `go 1.25.0`. This is a pre-existing,
  unrelated module-version mismatch. `ptx-mcp` (built before that line) and
  `ptx-installer` (built directly) both compile cleanly. Recommend a separate
  fix for `ptx-prompting/go.mod`.
- **Execution note (not a #035 defect):** `ptx-installer.exe` triggers the
  Windows Installer-Detection heuristic (binary name contains "install") → UAC
  elevation, so it cannot run non-interactively. The helper was built under a
  neutral name (`ptxh.exe`) for live testing; the dispatcher forwards identical
  arguments, so the exercised code path is unchanged. Matches the note in
  `acceptance-035-detection.md`.

## Test Results

### T1 — Build sanity

- [x] `ptx-installer` and `ptx-mcp` build at HEAD (`9f07c60`), exit 0
- [x] `make build` produces `portunix.exe` (helper stage blocked later by the
      unrelated `ptx-prompting` toolchain mismatch — see Test Environment)

### T2 — Unit: version resolution (`Installer.ResolveVersion`)

- [x] `TestResolveVersion_ByVariantName` PASS — selector `"21"` → variant `21`
- [x] `TestResolveVersion_ByVersionField` PASS — selector `"17.0.16_8"` →
      variant `17` (matched on the variant's `version` field)
- [x] `TestResolveVersion_Unknown` PASS — unmatched selector returns an error
      enumerating available versions
- [x] `TestResolveVersion_UnknownPackage` PASS — missing package reported as
      "not found"

### T3 — Unit: detection suite regression

- [x] `engine` package detection tests PASS (`DetectAIAssistants_EndToEnd`,
      `SelectAIAssistants_*`, `ParseVersionOutput`, `IsPathLookup`,
      `VerifyCommandForOS`) plus Docker admin/data-root, UAC, variant suites
- [x] `registry` package tests PASS (0.36s)

### T4 — Live: `install --help` surface

- [x] Lists `--version=<version>` (resolved to a variant by name or version)
- [x] Lists `--recommend-ai`
- [x] Examples include `install java --version 21` and `install --recommend-ai`

### T5 — Live: `install --recommend-ai --dry-run`

- [x] Given embedded registry on a host with Claude Code installed, When the
      command runs, Then it lists the 3 known assistants with display names:
      - `claude-code` → ✅ installed (2.1.156)
      - `claude-desktop` → ⬜ available
      - `gemini-cli` → ⬜ available
- [x] Reports `🔍 DRY RUN: would install claude-desktop, gemini-cli`
- [x] No installation performed; exit code 0

### T6 — Live: `install --recommend-ai` interactive decline

- [x] Given missing assistants, When the prompt
      `Install missing assistant(s) [claude-desktop, gemini-cli]? [Y/n]:` is
      answered `n`, Then it prints `Skipping installation.` and exits 0
- [x] No installation performed

### T7 — Live: `--version` by variant name

- [x] `install java --version 21 --dry-run` → resolves to `Variant: 21
      (version: 21.0.8_9)`; dry-run preview; exit 0

### T8 — Live: `--version` by version field

- [x] `install java --version 17.0.16_8 --dry-run` → resolves to `Variant: 17
      (version: 17.0.16_8)`; dry-run preview; exit 0

### T9 — Failure injection: invalid `--version`

- [x] `install java --version 99 --dry-run` → `❌ version "99" not available for
      java on windows` + `Available versions: 11 (11.0.28_6), 17 (17.0.16_8),
      21 (21.0.8_9), 8 (8u462b08)`
- [x] Exits non-zero (exit code 1); no installation attempted

### T10 — Precedence: `--variant` wins over `--version`

- [x] `install java --variant=21 --version 8 --dry-run` → `Variant: 21
      (version: 21.0.8_9)` (the `--version 8` selector is ignored when an
      explicit variant is given) — matches documented behavior

### T11 — Live: `package update`

- [x] `package --help` lists `update — Reinstall a package's latest available
      version`
- [x] `package update java --dry-run` → `🔄 Updating java to the latest
      available version...`, resolves the auto-detected/preferred variant
      (`Variant: 8`), dry-run preview, Force path; exit 0

### T12 — Live: `mcp init --install-missing` wiring

- [x] `mcp init --help` exposes `--install-missing  Auto-install missing AI
      assistants via ptx-installer`
- [x] Example present: `mcp init --assistant gemini-cli --install-missing`
- [x] Code wiring confirmed: `initCmd` reads the flag into
      `opts.installMissing` and passes it to `runInteractiveWizard` (offer
      branch) and the non-interactive `runNonInteractiveConfiguration`
      (auto-install branch via `installAssistant` → ptx-installer)

### T13 — Regression

- [x] `package list` still reports 69 packages
- [x] `package detect` (table + `--json`) unchanged: 1 of 3 installed,
      `claude-code` 2.1.156, others not found
- [x] No AI assistant package JSON definitions modified by this branch

## Conditional Items (required before full merge approval)

1. **Real install/update execution not exercised on host.** Per the project
   container-testing policy, the following must be validated in a clean
   container, not on the developer host:
   - `install --recommend-ai -y` actually installing claude-desktop/gemini-cli
   - `mcp init --install-missing` auto-installing a genuinely missing assistant
     (interactive offer + non-interactive path)
   - `package update <pkg>` performing a real Force reinstall
   On the host these were verified only via `--dry-run` and the interactive
   decline path (the new #035 logic — resolution, listing, offer, Force
   plumbing — is fully exercised up to the install exec, which reuses the
   already-accepted `Installer.Install`).
2. **Cross-platform coverage.** Verified on Windows only. Linux and macOS
   detection/recommendation/version paths remain to be run.

## Known Limitations / Notes

1. **`package update --help` shows package-level help** instead of the
   update-specific usage block added in this change. Root cause: `handlePackage`
   intercepts `--help`/`-h` for any argument before dispatching to the
   subcommand — the same pre-existing pattern already noted for
   `list`/`search`/`info`/`detect` in `acceptance-035-detection.md`. The
   update-specific help text exists in `handlePackageUpdate` but is shadowed.
   Not a blocker; recommend folding into the existing per-subcommand `--help`
   cleanup issue.
2. **`package update` updates the auto-detected/preferred variant**, not a
   user-named version. Combined with `--version` resolution this covers the
   accepted scope; full pinning/downgrade with installed-state tracking remains
   open in the issue.
3. **UAC elevation** of the production-named `ptx-installer.exe` — Windows
   installer-detection heuristic, not a code defect (see Test Environment).
4. **Pre-existing engine test isolation flakiness (unrelated to #035):**
   `TestInstallDownload_SingleFile` passes in isolation (0.03s) but hangs when
   run immediately after `TestInstallBundle_MissingMemberFails` (loopback HTTP
   state leakage); under the restricted network sandbox all loopback download
   tests hang. Both files are untouched by this branch and all #035 tests pass.
   Recommend a separate fix for HTTP test isolation.

## Recommendations

- Validate the three real install/update flows in a clean Ubuntu container
  (`portunix container` / `docker run-in-container`) before closing the issue.
- Run the same matrix on Linux; revisit macOS alongside the macOS install slice.
- Fix `ptx-prompting/go.mod` toolchain pin so `make build` completes end-to-end.
- Fold the shadowed per-subcommand `--help` (now also affecting `update`) into
  one cleanup pass.

## Final Decision

**STATUS**: CONDITIONAL

**Approval for merge**: CONDITIONAL — the Windows CLI/logic slice
(`--recommend-ai` listing/dry-run/decline, `--version` resolution + precedence,
`package update` plumbing, `mcp init --install-missing` exposure/wiring, unit
tests) PASSES with no functional defects. Full approval is pending container
validation of the real install/update execution paths and Linux/macOS coverage
(see Conditional Items). The `package update --help` shadowing is a non-blocking
known limitation.

**Date**: 2026-05-29
**Tester signature**: zdendaku (QA/Test Engineer — generic)
