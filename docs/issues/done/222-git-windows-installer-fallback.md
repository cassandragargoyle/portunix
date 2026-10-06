# Issue #222: Git on Windows: Fall Back to Official Git for Windows Installer When winget Is Unavailable

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Component**: ptx-installer
**Created**: 2026-10-07
**Closed**: 2026-10-07
**GitHub**: #222

## Summary

On Windows, `portunix install git` should work even on machines without winget.
When winget is not available, the installer falls back to the official Git for
Windows installer published at <https://git-scm.com/install/windows> and runs it
silently.

## Problem Description

The package definition `src/helpers/ptx-installer/assets/packages/git.json`
declares only package-manager based variants for Windows:

- `chocolatey` (platform default type, declared first)
- `winget` (`Git.Git`)

`selectDefaultVariant()` in `src/helpers/ptx-installer/engine/installer.go` has
no mapping for Windows package managers, so the first declared variant is used
regardless of which package manager is actually present. On a clean Windows
machine (e.g. Windows Sandbox, LTSC / Server editions, locked-down corporate
images, fresh VMs) winget is often missing and chocolatey is not installed
either, so Git cannot be installed without manual steps.

Git for Windows provides an official standalone installer (Inno Setup `.exe`)
that supports unattended installation and requires no package manager.

## Proposed Solution

1. Add a direct-installer variant to `git.json` for Windows, e.g. `installer`:
   - `type: "exe"`, architecture-specific URLs (`x64`, `arm64`) pointing to the
     Git for Windows release assets referenced from
     <https://git-scm.com/install/windows>
     (`https://github.com/git-for-windows/git/releases/download/v{version}.windows.1/Git-{version}-64-bit.exe`)
   - silent install args: `/VERYSILENT /NORESTART /NOCANCEL /SP- /SUPPRESSMSGBOXES`
     (optionally `/COMPONENTS=...` and `/o:PathOption=Cmd` so `git` is on PATH)
   - `requiresAdmin: true` for the machine-wide installer
2. Version resolution: add a `versionResolver` for the latest Git for Windows
   release (GitHub releases of `git-for-windows/git`), with a pinned
   `fallbackVariant` when the release listing is unreachable — same pattern as
   the existing `python.org` resolver.
3. Variant selection on Windows: when no `--variant` is given, pick the variant
   based on available tooling:
   - winget available → `winget`
   - otherwise chocolatey available → `chocolatey`
   - otherwise → `installer` (official Git for Windows installer)
   This should be implemented generically in `selectDefaultVariant()` /
   `DetectPackageManager()` (map `winget` / `choco` to variant names) rather than
   hard-coded for Git, so other packages benefit as well.
4. Runtime fallback: if the selected package-manager variant fails because the
   package manager binary is not found, retry with the `installer` variant and
   print a clear message (`winget not found, falling back to official installer`).
5. After installation, refresh PATH for the current session or print a hint that
   a new shell is needed (`C:\Program Files\Git\cmd`).

## Acceptance Criteria

1. On Windows without winget and without chocolatey,
   `portunix install git` downloads the official Git for Windows installer and
   installs Git silently; `git --version` succeeds afterwards (new shell)
2. On Windows with winget available, `portunix install git` still uses winget
3. `portunix install git --variant installer` forces the official installer even
   when winget is available
4. `portunix install git --dry-run` shows the selected variant and the download
   URL that would be used
5. If the latest-version lookup fails, the pinned fallback version is installed
   and a warning is printed
6. Download is verified (checksum where available) and the temporary installer
   file is removed after installation
7. Both `x64` and `arm64` architectures are supported
8. Existing Linux variants (`apt`, `dnf`) are unaffected
9. Unit tests cover variant selection for: winget present, only chocolatey
   present, neither present, explicit `--variant`

## Testing

All installation tests must run in an isolated environment (Windows Sandbox or
Windows VM), never on the developer host.

- Windows Sandbox (no winget by default): `portunix install git` → official
  installer path, verify `git --version`
- Windows VM with winget: `portunix install git` → winget path
- Windows VM with winget: `portunix install git --variant installer` → official
  installer path
- Simulate unreachable release listing (offline / blocked GitHub API) → fallback
  version used
- `go test ./src/helpers/ptx-installer/...` — unit tests for variant selection
- Linux container (`ubuntu:22.04`): `portunix install git` still uses apt
