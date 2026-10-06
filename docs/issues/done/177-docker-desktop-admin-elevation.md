# Issue #177: `portunix install docker` Fails Without Admin Elevation on Windows

**Type:** Bug Fix
**Priority:** High
**Status:** ✅ Implemented
**Labels:** bug, installation, docker, ptx-installer, windows, user-experience, uac

## Description

When running `portunix install docker` on Windows from a non-elevated shell, the
Docker Desktop installer (`Docker Desktop Installer.exe`) fails because it
requires Administrator privileges. The failure happens after package download
and after the new `data-root` prompt (issue #176) — i.e. the interactive flow
completes normally and only then the install step aborts.

Legacy `src/app/docker/docker.go:installDockerDesktopWindows` was a stub that
only printed status messages and never actually executed the installer, so the
current `ptx-installer` path is the first real invocation and does not handle
elevation at all.

## Steps to Reproduce

1. Open a standard (non-elevated) PowerShell or cmd prompt on Windows
2. Run `portunix install docker`
3. Observe the interactive `data-root` prompt — works correctly
4. Installation proceeds to download Docker Desktop installer
5. Installer is launched and immediately fails:

```text
🔧 Installing Docker Desktop...
For security reasons C:\ProgramData\DockerDesktop must be owned by an elevated account

❌ Docker installation failed: Docker Desktop installation failed: exit status 0xfffffffb
```

`0xfffffffb` decodes to `-5` → `ERROR_ACCESS_DENIED`.

## Expected Behavior

Before downloading the installer, the ptx-installer should detect whether the
current process has Administrator privileges. If not, it should:

- Abort early with a clear, actionable message (e.g. "Docker Desktop requires
  Administrator privileges — please run Portunix from an elevated shell")
- Avoid downloading the ~500 MB installer file just to fail on launch

Optionally (follow-up): trigger UAC elevation automatically via
`Start-Process -Verb RunAs -Wait` and stream output back. Out of scope for the
initial fix because UAC handling is non-trivial (stdout/stderr redirection,
user experience, multiple elevation prompts).

## Actual Behavior

Installation proceeds through all preparatory steps, downloads the installer,
then fails at the final launch step with a cryptic Windows error code. The
temporary installer is left in `%TEMP%`.

## Environment

- **Command:** `portunix install docker`
- **Installer path:** `ptx-installer` helper binary
- **Platform:** Windows 10/11 (non-elevated shell)
- **Reporter environment:** Windows 11 Pro 10.0.26200

## Root Cause

In `src/helpers/ptx-installer/engine/docker.go` the Docker Desktop installer is
launched via plain `exec.Command(installerPath, args...)` without any
elevation check or `runas` wrapper:

```go
cmd := exec.Command(installerPath, args...)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
    return fmt.Errorf("Docker Desktop installation failed: %w", err)
}
```

Docker Desktop Installer requires membership in the Administrators group to
create `C:\ProgramData\DockerDesktop` with the correct ACLs.

## Acceptance Criteria

- [ ] `portunix install docker` detects missing Administrator privileges on
      Windows **before** downloading the installer
- [ ] Clear error message instructs the user to re-run from an elevated shell
- [ ] `--dry-run` flow is unaffected (should work from non-elevated shell)
- [ ] Linux and macOS paths are unaffected (they already handle elevation via
      `sudo` / `brew`)
- [ ] Unit test for the privilege-check helper function

## Related Issues

- #176 — `portunix install docker` Missing Installation Directory Prompt
  (discovered during manual testing of the #176 fix)
- #019 — Docker Installation Issues on Windows (historical context)
- #122 — Consolidate Docker/Podman Installation into ptx-installer

## Workaround

Run Portunix from an elevated PowerShell or cmd:

```powershell
# From an Administrator PowerShell
portunix install docker
```

## Notes

Follow-up work (separate issue): auto-elevation via UAC prompt. This would
require spawning `powershell.exe -Command "Start-Process -Verb RunAs -Wait ..."`
and handling the loss of stdout/stderr streaming during elevation.

---
**Created:** 2026-04-21
**Closed:** 2026-04-21
**Reporter:** @zdendaku
