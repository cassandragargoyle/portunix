# Issue #189: On-demand privilege elevation per command (no requireAdministrator in EXE manifest)

**Status:** 📋 Open
**Priority:** Medium
**Type:** Enhancement
**Labels:** enhancement, installer, ptx-installer, windows, uac, elevation, security
**Affects:** `portunix.exe.manifest`, `src/helpers/ptx-installer/engine/installer.go`, `src/helpers/ptx-installer/engine/uac.go`, `src/helpers/ptx-installer/engine/admin_windows.go`, `src/helpers/ptx-installer/engine/admin_other.go`, `build-with-version.sh`, `Makefile`
**Related:** [#127 — admin check + env var expansion](done/127-migrate-openssh-to-ptx-installer.md), [#177 — Docker Desktop admin elevation check](done/177-docker-desktop-admin-elevation.md), [#178 — auto-elevate Docker Desktop installer via UAC](done/178-docker-auto-elevation-uac.md)
**Created:** 2026-06-03

## Progress

- 2026-06-08 — Implementation merged to `main` (commit `4a31b67`): shared
  `decideElevation` + `reExecElevated` in `engine/elevation.go`, on-demand
  elevation in `Install()`, Docker installer refactored to reuse the shared
  path, plus `manifest_guard_test.go` regression guard. **Build/tests not yet
  verified** (go toolchain unavailable in the dev shell at merge time).
- **Next step: TESTING.** Run `make build` + `go test ./...` (incl. root
  `manifest_guard_test.go`) and `go test ./src/helpers/ptx-installer/engine/`,
  then the manual UAC/sudo scenarios below and produce
  `docs/testing/acceptance-189.md`. Keep this issue `Open` until tests pass.

## Summary

`portunix.exe` must always run as a normal user (Windows manifest
`asInvoker`, no `requireAdministrator`) and request privilege elevation
**only for the concrete command that genuinely needs it** — never globally
for the whole process. Today this is only partially true: the manifest is
already `asInvoker`, but on-demand elevation is implemented for the Docker
Desktop installer only. Generic packages flagged `requiresAdmin` simply abort
with an error telling the user to manually relaunch an elevated shell.

This issue generalizes the on-demand elevation pattern (already proven for
Docker in #178) to every command that requires admin/root, and adds a
regression guard so the manifest never silently flips back to
`requireAdministrator`.

## Motivation

- **Least privilege**: the main binary and the vast majority of subcommands
  (help, info, system info, user-scoped installs, dry-run, etc.) need no
  elevation. Forcing the whole process to admin would be a security and UX
  regression.
- **No accidental escalation**: explicit `asInvoker` in the manifest is what
  suppresses Windows' "installer-detection" heuristic that escalated the
  unmanifested binary (the symptom fixed in commit `653341b`). This must be
  protected against regressions on version bumps / `.syso` regeneration.
- **Better UX for admin-requiring installs**: instead of failing with "run as
  Administrator and try again", the command should raise a single UAC prompt
  (Windows) / re-exec under `sudo` (Linux) scoped to that operation, the way
  `portunix install docker` already does.

## Current state (baseline)

- `portunix.exe.manifest` — already declares
  `<requestedExecutionLevel level="asInvoker" uiAccess="false" />`. ✅ Keep.
- `engine/installer.go:300` — when `variantSpec.RequiresAdmin && !IsAdmin()`
  the install **errors out** ("requires Administrator privileges … run
  PowerShell as Administrator"). ❌ Should request elevation on-demand instead.
- `engine/uac.go` — `runWithUAC()` / `buildUACScript()` already provide a
  **generic, reusable** UAC re-launch helper (PowerShell `Start-Process -Verb
  RunAs -Wait`), with `errUACDeclined` handling. Not Docker-specific.
- `engine/docker.go` — `decideElevation()` + `runDockerInstaller()` show the
  target pattern: dry-run / already-admin → run directly; otherwise UAC prompt.
- `engine/admin_windows.go` / `engine/admin_other.go` — `IsAdmin()` already
  abstracts admin/root detection cross-platform.

## Scope

### 1. Manifest stays `asInvoker` + regression guard

- Confirm `portunix.exe.manifest` keeps `asInvoker` (no
  `requireAdministrator`, no `highestAvailable`).
- Ensure `build-with-version.sh` / `Makefile gen-resources` regenerate
  `portunix.syso` from `versioninfo.json` referencing this manifest on every
  build (already wired in `653341b` — verify, don't duplicate).
- Add a guard (test or preflight check) that fails the build if the embedded
  manifest does not request `asInvoker`.

### 2. Generic on-demand elevation in ptx-installer

- Replace the hard error at `engine/installer.go:300` with the
  `decideElevation` pattern:
  - dry-run or already elevated → proceed directly (current behaviour);
  - not elevated + real install of a `requiresAdmin` variant → re-launch the
    **same** `portunix install …` invocation elevated via `runWithUAC` on
    Windows, forwarding the original args; on UAC decline surface the existing
    actionable fallback message.
- Linux/macOS: when `requiresAdmin` and not root, re-exec the command under
  `sudo` (or print the exact `sudo portunix …` line) instead of a bare error.
- Reuse the existing `runWithUAC` / `IsAdmin` helpers — do **not** introduce a
  parallel elevation path. Consider extracting `decideElevation` out of
  `docker.go` into a shared spot so both Docker and generic installs share it.

### 3. Guard against double elevation / loops

- When the elevated child re-runs, it must detect it is already elevated and
  perform the install directly (no second UAC prompt → no infinite loop).
- Preserve `--dry-run` semantics: dry-run must always work from a
  non-elevated shell with no prompt.

## Acceptance Criteria

- [ ] `portunix.exe` embedded manifest requests `asInvoker`; launching any
      non-admin command (`portunix`, `portunix system info`, `--help`,
      `--dry-run`) produces **no** UAC prompt.
- [ ] Build fails (guard) if the manifest is changed to
      `requireAdministrator` / `highestAvailable`.
- [ ] Installing a `requiresAdmin` package from a non-elevated Windows shell
      raises exactly one UAC prompt scoped to that command and completes after
      consent; declining shows the actionable fallback message.
- [ ] Same install from an already-elevated shell runs directly with no extra
      prompt.
- [ ] On Linux, a `requiresAdmin` install from a non-root shell elevates via
      `sudo` (or prints the precise `sudo` command) instead of a bare error.
- [ ] No elevation loop: the elevated re-launch detects elevation and installs
      directly.
- [ ] Unit tests for the elevation decision (extend `engine/docker_admin_test.go`
      pattern) cover the generic install path.

## Out of Scope

- Changing which packages are flagged `requiresAdmin` (manifest data).
- Per-command elevation for non-installer helpers (track separately if needed).

## Testing

- Unit: `decideElevation` / generic install elevation decision (pure-function
  tests, no OS calls) — see `engine/docker_admin_test.go`.
- Manual (Windows, per TESTING_METHODOLOGY container/host notes): non-elevated
  vs elevated shell, dry-run, UAC accept/decline.
- Manual (Linux container): non-root `requiresAdmin` install → sudo path.
