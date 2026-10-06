# Issue #178: Auto-elevate Docker Desktop Install via UAC on Windows

**Type:** Enhancement
**Priority:** Medium
**Status:** ✅ Implemented
**Labels:** enhancement, installation, docker, ptx-installer, windows, user-experience, uac

## Description

Follow-up to issue #177. After #177, `portunix install docker` on Windows
fails fast with a clear message when the shell is not elevated. The next step
is to handle elevation automatically via a UAC prompt instead of asking the
user to re-open an Administrator shell manually.

Goal: from a standard (non-elevated) PowerShell/cmd, running
`portunix install docker` should trigger a UAC consent dialog, then complete
the installation without the user having to restart their shell.

## Motivation

- Reduces friction for the common case (most users install Docker Desktop
  once, from whatever shell they happen to have open)
- Matches behavior of native Windows installers (MSI, Chocolatey, WinGet)
  that escalate themselves when needed
- Keeps the #177 fail-fast check as a safety net when auto-elevation fails
  or is declined

## Design Options

### Option A — Re-exec `portunix` elevated

Restart the whole `portunix` process elevated via `ShellExecute` with the
`runas` verb, passing the same argv.

- **Pros:** Simple, one place to change (the installer entry point)
- **Cons:** Elevated child runs in a new console window; original console
  loses stdout/stderr. Need `-Wait` plus logging-to-file or an
  `--elevated-child` flag plus IPC (named pipe / shared log file) to stream
  output back.

### Option B — Elevate only `Docker Desktop Installer.exe`

Keep `portunix` non-elevated for download, prompts, and config. Launch just
the Docker Desktop installer via `runas` when it's time.

- **Pros:** Portunix state (progress, logs, post-install config) stays in the
  original console.
- **Cons:** Docker installer runs across a UAC boundary — we still cannot
  capture its stdout/stderr live. Need to poll exit code and surface it back.

### Option C — Hybrid

Detect elevation early (already done in #177). On first download-and-config
steps run non-elevated. Escalate for the actual installer launch (B) and fall
back to A if B turns out to fail for reasons other than ACCESS_DENIED.

## Proposed Approach

Start with **Option B** (elevate only the installer). Acceptable even without
live stdout streaming — Docker installer has its own progress UI anyway, and
exit code is enough to report success/failure back in the original console.
Re-evaluate after manual testing; if the UX is poor, fall back to Option A
with logging-to-file.

## Acceptance Criteria

- [ ] Running `portunix install docker` from a **non-elevated** PowerShell
      triggers a single UAC dialog before the installer launches
- [ ] On UAC approval: Docker Desktop installs successfully without the user
      needing to restart the shell
- [ ] On UAC decline: installer fails with a clear message pointing at the
      existing #177 fallback (run from an elevated shell, or re-try)
- [ ] Elevated-shell path unchanged (no duplicate UAC prompts)
- [ ] `--dry-run` unaffected (works from non-elevated shell, no UAC prompt)
- [ ] Linux and macOS paths unaffected
- [ ] Manual test on Windows 10 and Windows 11 (clean VM)
- [ ] Unit or integration test for the elevation decision logic (when to
      prompt, when to pass through, when to reject)

## Implementation Notes

- Go stdlib has no direct wrapper for `ShellExecuteEx` with `runas`. Use
  `golang.org/x/sys/windows` (already a dependency via `admin_windows.go`)
  or shell out to PowerShell: `Start-Process -Verb RunAs -Wait ...`
- UAC prompt is a user-visible action — must happen before any destructive
  or long-running work (download is fine, but any file system changes under
  `C:\ProgramData\DockerDesktop` must wait)
- Consider telemetry / logging: capture whether UAC was prompted, accepted,
  declined, or skipped (elevated shell) for future UX tuning

## Related Issues

- #177 — `portunix install docker` Fails Without Admin Elevation on Windows
  (dependency — #178 builds on the fail-fast check)
- #019 — Docker Installation Issues on Windows (historical context)

## Out of Scope

- Auto-elevation for other installers (VS Code, Java, etc.). If #178 lands
  well, a separate issue can generalize the mechanism to `ptx-installer`'s
  core and apply it to any package with `"requires_admin": true`.

---
**Created:** 2026-04-21
**Closed:** 2026-04-22
**Reporter:** @zdendaku
