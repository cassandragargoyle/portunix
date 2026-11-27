# Issue #203: Windows `installScript` Written in PowerShell Is Run by `cmd.exe` (`portunix install claude-desktop`)

## Priority

**HIGH** — `portunix install claude-desktop` fails on every Windows computer. The
package's install script is PowerShell, but Portunix hands it to `cmd.exe`, which
stops at the first `$variable`. `claude-code` has a PowerShell script too.

## Status

- **Created**: 2026-10-06
- **Status**: ✅ Implemented
- **Completed**: 2026-11-08
- **Assignee**: zdendaku
- **Branch**: `fix/203-windows-powershell-install-scripts`
- **Related**:
  - Acceptance protocol: `docs/testing/acceptance-203.md` (retest 2: PASS)
  - `src/cmd/root.go`, `src/cmd/sandbox_generate.go` — scope extension items 9, 10
  - `src/helpers/ptx-installer/engine/installer.go` — `installScript`
  - `src/helpers/ptx-installer/assets/packages/claude-desktop.json`
  - `src/helpers/ptx-installer/assets/packages/claude-code.json`
  - Issue #201 — the same function; quoting of inline commands and the missing
    verification (fixing one without the other leaves the packages broken)

## Problem Description

### Current Situation

`claude-desktop.json`, Windows variant `latest`:

```json
"installScript": "$url = 'https://claude.ai/download/win'; $output = Join-Path $env:TEMP 'claude-desktop-installer.exe'; Invoke-WebRequest -Uri $url -OutFile $output; Start-Process -FilePath $output -ArgumentList '/S' -Wait; Remove-Item $output"
```

`installScript` runs every inline Windows command as
`exec.Command("cmd", "/c", expandedScript)`. A package cannot say which interpreter
its script needs, and this one is PowerShell:

```text
> portunix install claude-desktop
🚀 Starting installation (type: script)...
📝 Running install script (target: ./site)...
'$url' is not recognized as an internal or external command,
operable program or batch file.
   ❌ Running: $url = 'https://claude.ai/download/win'; $output = Join-Path $env:TEMP 'claud...

❌ Installation failed: script failed: exit status 1
```

`claude-code.json` (variant `npm`) is the same kind:
`Write-Host 'Installing Claude Code via npm...'; npm install -g @anthropic-ai/claude-code; ...`
— under `cmd.exe`, `Write-Host` fails before `npm` runs.

Other packages with Windows inline scripts are to be checked; a search for PowerShell
syntax (`$name`, `Write-Host`, `Invoke-`, `Start-Process`) in `installScript` also
matches `powershell`, `qemu`, `spice-guest-agent`, `spice-guest-tools`, `virt-viewer`,
`virt`, `virtualbox`, `wireguard` and `winget` — some of those may be Linux scripts.

## Proposed Solution

1. **Let a package name its interpreter.** Add a field to the variant, e.g.
   `"scriptShell": "powershell" | "cmd" | "sh"`, defaulting to `cmd` on Windows and `sh`
   elsewhere as today. For `powershell`, run
   `powershell -NoProfile -ExecutionPolicy Bypass -Command <script>` (or write the script
   to a temporary `.ps1` and run it with `-File`, which avoids any re-quoting — see #201)
2. Set `"scriptShell": "powershell"` in `claude-desktop.json`, `claude-code.json` and every
   other Windows package whose script is PowerShell
3. **Check the Claude Desktop install itself** once the script runs: that
   `https://claude.ai/download/win` still delivers the installer, and which switch makes
   it unattended (the script passes `/S`); then verify with the package's
   `verification` (#201)
4. A registry test that fails when a Windows `installScript` uses PowerShell syntax
   without `scriptShell: powershell`

## Acceptance Criteria

- [ ] On a clean Windows, `portunix install claude-desktop` installs Claude Desktop
      without a window waiting for input, and it starts afterwards
- [ ] `portunix install claude-code` runs its npm install on Windows
- [ ] A package can choose `powershell`, `cmd` or `sh` for its inline script
- [ ] The registry test catches a PowerShell script declared without
      `scriptShell: powershell`

## Scope Extension (2026-11-07)

The first acceptance run in Windows Sandbox (`docs/testing/acceptance-203.md`) confirmed
the `scriptShell` mechanism but found two blocking defects and four further problems.
All of them are fixed on this branch.

### Blocking Findings

5. **claude-desktop download source** — `https://claude.ai/download/win` answers a
   script with a browser challenge (HTTP 403), so no script can download the installer.
   Claude Desktop is now distributed as MSIX with versioned URLs; install it through
   winget (`Anthropic.Claude`)
6. **claude-code `curl` variant reports a false failure** — `install.ps1` installs into
   `%USERPROFILE%\.local\bin`, which is not in `PATH`, so the verification fails.
   Declare `environment.PATH_APPEND` and make the `script` verification see that
   directory (it is otherwise added only after the verification has run)

### Additional Fixes

7. **qemu on Windows reports success without installing anything** — the Windows variant
   only prints a download link and the platform has no `verification`. Install QEMU via
   winget (`SoftwareFreedomConservancy.QEMU`) and add `PATH_APPEND` and `verification`
8. **Dependencies are not installed** — `Install` only prints
   `Dependency <name> should be installed first` (e.g. `nodejs` for `claude-code`) and
   continues. Install every missing dependency first (a dependency whose verification
   succeeds is skipped), then refresh `PATH`
9. **`--help-expert` / `--help-ai` on a subcommand run the command** — both are root
   persistent flags handled only for the root command, so e.g.
   `portunix sandbox generate --help-expert` generates a `.wsb` file instead of showing
   help. A subcommand given either flag must show its help and do nothing else
10. **`portunix sandbox generate --map-folder` rejects drive-letter paths** — the value is
    split on every `:`, so `C:\Host:C:\Sandbox:true` is invalid. Parse the drive-letter
    colons correctly

### Additional Acceptance Criteria

- [ ] `portunix install claude-code --variant curl` passes its verification on a clean
      Windows
- [ ] `portunix install qemu` on Windows installs QEMU or fails; it never reports success
      without QEMU
- [ ] `portunix install claude-code` on a clean Windows installs `nodejs` first
- [ ] `portunix <subcommand> --help-expert` and `--help-ai` show help without executing
      the subcommand
- [ ] `portunix sandbox generate --map-folder "C:\Host:C:\Sandbox:true"` maps the folder
