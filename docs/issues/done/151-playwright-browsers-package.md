# Issue #151: Playwright Browsers Package Definition

**Type**: Feature
**Priority**: Medium
**Status**: ✅ Implemented
**Created**: 2026-02-21
**Labels**: ptx-installer, package-management, playwright, browser-automation, cross-platform
**Origin**: ptx-agent team handoff (browser automation dependency)
**Related**: #149 (Vox ptx-install integration — same pattern), #086 (Package Registry Auto-Discovery), #152 (Plugin Package Dependency Resolution)

## Summary

Create a standard package definition `playwright-browsers.json` in `assets/packages/` that installs Chromium browser binaries required by the ptx-agent plugin for browser automation.

## Implementation

Package definition created at `src/helpers/ptx-installer/assets/packages/playwright-browsers.json` with:

- **Default variant**: Chromium only via `npx playwright install chromium` (~200 MB)
- **All variant**: All browsers (Chromium + Firefox + WebKit) via `npx playwright install` (~600 MB)
- **Linux**: Automatic system dependency installation via `npx playwright install-deps`
- **Dependency**: Node.js (for `npx` availability)
- **Verification**: `npx playwright --version`
- **Auto-discovery**: No index updates needed (package registry auto-discovery #086)

## Acceptance Criteria

- [x] `portunix install playwright-browsers` installs Chromium binary via `npx playwright install chromium`
- [x] `portunix install playwright-browsers --variant all` installs all browsers
- [x] Node.js dependency is resolved automatically (installed if missing)
- [x] Linux: system dependencies installed via `install-deps`
- [x] Windows: installation works without system deps
- [x] `portunix package list` shows playwright-browsers under `development/testing`
- [x] Reinstallation/update works when called again

---

**Created**: 2026-02-21
**Implemented**: 2026-02-21
**Author**: Architect
