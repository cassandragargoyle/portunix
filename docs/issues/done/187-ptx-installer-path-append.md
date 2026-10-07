# Issue #187: ptx-installer — Honor `environment.PATH_APPEND` After Install

**Status:** ✅ Implemented (Phase 1 + Phase 2; Phase 3 deferred)
**Priority:** High
**Type:** Bug Fix / Enhancement
**Labels:** bug, enhancement, ptx-installer, installation, path, windows, linux, user-experience
**Affects:** `src/helpers/ptx-installer/engine/installer.go`, `src/helpers/ptx-installer/engine/path.go`, `src/helpers/ptx-installer/assets/packages/*.json`
**Created:** 2026-05-20
**Closed:** 2026-05-22

## Summary

Many package definitions under `src/helpers/ptx-installer/assets/packages/`
declare an `environment.PATH_APPEND` entry (e.g. `${LOCALAPPDATA}\Programs\tea`,
`C:/Program Files/act`, `${USERPROFILE}/.cargo/bin`). The installer engine was
ignoring this field entirely, so packages installed into user-local directories
never landed on the user's `PATH` — `portunix install tea` placed `tea.exe`
correctly but `tea --version` from a fresh shell still produced "command not
found".

This issue tracks adding PATH manipulation to `ptx-installer` so the field
matches the documented behaviour for new installs.

## Motivation

- Make `portunix install <pkg>` "just work" — after the command returns, the
  binary is on `PATH` without a manual `setx` / `.bashrc` edit.
- Unblock 30+ existing packages that already declare `PATH_APPEND` but never
  had it applied (`act`, `actionlint`, `caddy`, `hugo`, `ninja`, `protoc`,
  `tea`, `terraform`, `rust`, …).
- Triggered concretely by the broken `portunix install tea` flow on Windows
  (related: rename-by-binary-field fix landed alongside this issue).

## Scope

### Phase 1 — Static path values (this commit)

- Read `platformSpec.Environment["PATH_APPEND"]` after a successful install.
- Expand env-var placeholders only — `${LOCALAPPDATA}`, `${USERPROFILE}`,
  `$HOME`, `%APPDATA%`, etc. (reuses `expandEnvVars`).
- Skip with a warning when value contains `${install_path}` or
  `${extract_to}` placeholders (those depend on per-installer-method state
  that isn't surfaced yet — Phase 2).
- Windows: write `HKCU\Environment\Path` via PowerShell
  `[Environment]::SetEnvironmentVariable("Path", …, "User")`. PowerShell
  also broadcasts `WM_SETTINGCHANGE` so newly-spawned shells see the change.
- Linux/macOS: append a marked `export PATH="…:$PATH"` block to the active
  shell's rc file (`.zshrc` / `.bashrc` / `.profile`), idempotent.
- PATH failure is non-fatal — install is reported successful; PATH issue
  surfaces as a `⚠️` line.

### Phase 2 — Path placeholders (follow-up)

- Track final install path inside `installArchive`, `installDownload`,
  `installWindowsBinary`, … and expand `${install_path}` /
  `${extract_to}` in `PATH_APPEND` against that value.
- Affects packages like `go.json` (`${install_path}/bin`), `maven.json`
  (`${extract_to}/apache-maven-3.9.9/bin`), `java.json` (`${extract_to}/bin`).
- Likely refactor: install methods return `(installPath string, err error)`.

### Phase 3 — Optional follow-ups

- Symmetric `RemoveFromUserPath` invoked during `portunix package uninstall`
  (uninstall flow doesn't exist yet — defer).
- `environment` map keys beyond `PATH_APPEND` — currently undefined; consider
  `ENV_SET` / `ENV_APPEND` if a real use case appears.

## Acceptance Criteria

1. `portunix install tea` on Windows leaves `tea.exe` in
   `%LOCALAPPDATA%\Programs\tea\` AND adds that directory to the User PATH;
   `tea --version` works from a new shell session.
2. Re-running `portunix install tea` does not add a duplicate PATH entry.
3. Installing a package whose `PATH_APPEND` uses `${install_path}` or
   `${extract_to}` prints a warning but otherwise completes successfully.
4. On Linux, installing a package with a static `PATH_APPEND` appends a
   single marked block to the appropriate shell rc.
5. PATH-update failures (permission, missing PowerShell, …) do not fail the
   install — they only emit a `⚠️` and the binary stays usable via its full
   path.

## Out of Scope

- System (machine-wide) PATH — User PATH only. System PATH would force
  elevation on every install.
- Path removal during uninstall (no uninstall flow today).
- Live `PATH` injection into the currently-running portunix shell — restart
  the terminal.

## Notes

- Code touches: new `src/helpers/ptx-installer/engine/path.go`, modified
  switch-tail in `src/helpers/ptx-installer/engine/installer.go::Install`.
- Test plan covered in acceptance protocol (`docs/testing/internal/acceptance-187.md` — TODO when verified).
- Companion fix in this branch: `installDownload` now honors `variant.Binary`
  as the saved filename, so packages like `tea` no longer need shell-based
  rename in `postInstall` (was the original symptom that surfaced the
  `PATH_APPEND` gap).
