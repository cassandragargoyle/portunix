# Issue #179: Docker Install UX Improvements on Windows

**Type:** Enhancement
**Priority:** High
**Status:** ✅ Implemented
**Labels:** enhancement, installation, docker, ptx-installer, windows, user-experience, wsl, hyper-v

## Description

Three related UX problems around `portunix install docker` on Windows,
discovered during manual verification of the #177 fix:

### Problem 1 — No visible install progress

Docker Desktop installer is launched with the `--quiet` flag, which suppresses
its built-in progress dialog. The user sees nothing for ~1–2 minutes except
the final success/failure line, and cannot tell whether the install is making
progress or has frozen.

Relevant code in `src/helpers/ptx-installer/engine/docker.go:141-156`:

```go
args := []string{
    "install",
    "--accept-license",
    "--quiet",       // <-- hides Docker's progress UI
}
```

### Problem 2 — Missing virtualization prerequisites (WSL2 / Hyper-V)

Docker Desktop install succeeds but Docker Desktop then fails to start with:

```text
Virtualization support not detected
Docker Desktop failed to start because virtualisation support wasn't detected.
Sign in to try restoring access to Docker features.
```

Docker Desktop requires one of:

- **WSL2** (default backend, recommended) — needs `wsl --install` + restart
- **Hyper-V** (legacy backend) — needs Windows feature enabled + restart
- **CPU virtualization** enabled in BIOS (VT-x / AMD-V)

Portunix does not detect any of these before running the installer, leading to
a broken post-install state.

### Problem 3 — No standalone `portunix install wsl`

Users should be able to install WSL2 on its own, independently of Docker:

```bash
portunix install wsl
```

This lets users:

- Prepare WSL2 in advance (optionally from an automation script), then
  later run `portunix install docker` without the prerequisites failure
- Install WSL2 on machines where Docker is not wanted
- Compose multi-step install flows (WSL2 → Docker → …) cleanly

Current state: no `wsl` package definition exists in
`src/helpers/ptx-installer/assets/packages/`.

## Expected Behavior

1. **Progress**: Docker Desktop installer's built-in progress dialog is
   visible by default; user can follow the install.
2. **Prerequisites — interactive resolution**: Before downloading the
   installer, portunix detects whether WSL2 or Hyper-V is available. If
   neither is, it does **not** just fail — it offers an **interactive menu**
   and runs the selected prerequisite install:

   ```text
   ❌ Docker Desktop requires WSL2 or Hyper-V, but neither is available.

   📁 Select prerequisite to install:
      1. WSL2 (recommended — requires restart)
      2. Hyper-V (requires restart)
      3. Cancel (install Docker later)
   Choice [1]: _
   ```

   - **Choice 1 (WSL2)**: runs the standalone `install wsl` flow
     (`wsl --install --no-distribution`), reports that a restart is required,
     and exits — Docker install resumes on next invocation after restart.
   - **Choice 2 (Hyper-V)**: enables the Windows feature via PowerShell
     (`Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All
     -NoRestart`), reports restart requirement, exits.
   - **Choice 3 (Cancel)**: aborts cleanly, preserving the
     `portunix install wsl` hint.
   - In non-interactive mode (`--yes`/`-y`): auto-pick WSL2 as the
     recommended default.
   - In `--dry-run`: print the planned choice without touching the system.
3. **Standalone WSL install**: `portunix install wsl` is a first-class command
   that installs WSL2 on Windows (skipping or warning on non-Windows). It
   handles:
   - `wsl --install` invocation (requires Administrator → same elevation
     check as #177)
   - Reporting that a restart is required before WSL is usable
   - `--dry-run` to preview the plan

## Scope

### Phase 1 — Minimal fix (this issue)

- [ ] Remove `--quiet` from Docker Desktop installer args
- [ ] Pre-flight WSL2 detection via `wsl --status` (exit code 0 = OK)
- [ ] Fallback Hyper-V detection via
      `Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All`
- [ ] Interactive prerequisite resolution (menu + run selected install):
  - WSL2 choice → reuse `WSLInstaller` (runs `wsl --install --no-distribution`)
  - Hyper-V choice → `enableWindowsHyperV()`
    (`Enable-WindowsOptionalFeature … -NoRestart`)
  - Cancel → abort with hint
  - `--yes`/`-y`: auto-pick WSL2
  - `--dry-run`: print plan, don't execute
- [ ] Add standalone `portunix install wsl` command:
  - Package definition in `src/helpers/ptx-installer/assets/packages/wsl.json`
  - Implementation reuses existing admin-elevation check from #177
  - On Windows: shell out to `wsl --install` and report restart requirement
  - On non-Windows: refuse with a clear message ("WSL is Windows-only")
  - `--dry-run` supported
- [ ] Unit tests for the prerequisite-prompt decision logic (fake stdin,
      not real `wsl` / PowerShell calls)

### Phase 2 — Follow-up (separate issue if needed)

- Stream install progress into portunix console instead of spawning Docker's
  own dialog
- Auto-continue Docker install after a prerequisite restart (persistent
  state / scheduled task / post-reboot resume)
- Detect BIOS-level virtualization disable
  (`HyperVRequirementVirtualizationFirmwareEnabled = false`) and tell the
  user to enable VT-x / AMD-V in BIOS before any prerequisite helps

## Acceptance Criteria

- [ ] `portunix install docker` from elevated shell shows Docker's progress
      dialog during install (no more silent ~2 min gap)
- [ ] On a fresh Windows where WSL and Hyper-V are both unavailable,
      `portunix install docker` offers an interactive menu (WSL2 / Hyper-V /
      Cancel) and runs the chosen install before exiting with the restart
      hint
- [ ] `--yes`/`-y` auto-picks WSL2 without prompting
- [ ] When WSL2 is present, install proceeds as today (no prompt)
- [ ] `portunix install wsl` works as a standalone command on Windows
      (requires Administrator, supports `--dry-run`)
- [ ] `portunix install wsl` on Linux/macOS refuses with a clear message
- [ ] `--dry-run` is unaffected for both `install docker` and `install wsl`
- [ ] Linux and macOS docker paths unaffected
- [ ] Unit tests for the prerequisite-prompt decision logic

## Related Issues

- #177 — Fails Without Admin Elevation on Windows (dependency — elevation
  check happens first)
- #178 — Auto-elevate Docker Desktop Install via UAC (complementary)
- #019 — Docker Installation Issues on Windows (historical context)

## Notes

- Docker Desktop installer CLI flags (official):
  - `install`, `uninstall`
  - `--accept-license`
  - `--quiet` / `--silent`
  - `--backend=<wsl-2|hyper-v|windows>`
  - `--installation-dir="..."`
  - `--wsl-default-data-root="..."` / `--hyper-v-default-data-root="..."`
- `wsl --status` returns exit 0 when WSL is installed; non-zero or no output
  means not installed or disabled.
- `Get-ComputerInfo -Property HyperVRequirementVirtualizationFirmwareEnabled`
  can detect BIOS-level virtualization support but is slow (~500 ms per call).

---
**Created:** 2026-04-21
**Closed:** 2026-04-21
**Reporter:** @zdendaku
