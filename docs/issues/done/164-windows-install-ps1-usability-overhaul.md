# Issue #164: Windows install.ps1 Usability Overhaul

**Type**: Bug Fix / Enhancement
**Priority**: High
**Labels**: bug, enhancement, installation, windows, powershell, user-experience, github-distribution
**Status**: Implemented
**Created**: 2026-04-03

## Problem Description

The `scripts/install.ps1` script distributed on GitHub is practically unusable on standard Windows systems.
Users encounter multiple blocking issues before they can even run the installer.

## Current Issues

### 1. PowerShell Execution Policy Blocks Script

Windows default execution policy (`Restricted`) prevents running any `.ps1` scripts.
Users must first run `Set-ExecutionPolicy` or use `-ExecutionPolicy Bypass`, which is not
documented in the script or README.

**Impact**: Script fails immediately with cryptic error message.

### 2. SmartScreen / "Unknown Publisher" Warning

Downloaded `.ps1` files get NTFS Alternate Data Stream (ADS) Zone.Identifier mark.
Windows SmartScreen blocks execution with "Windows protected your PC" dialog.
Users must manually unblock via Properties or `Unblock-File`.

**Impact**: Even after bypassing execution policy, script may still be blocked.

### 3. Script Requires portunix.exe in Same Directory

The script expects `portunix.exe` to be already present next to the script.
This defeats the purpose of an installer - users must manually download the binary first.

**Impact**: Confusing UX - installer does not actually install anything by itself.

### 4. No Automatic Binary Download

Unlike `install.sh` which can download the binary from GitHub Releases,
`install.ps1` has no download capability. It's essentially a wrapper around
`portunix install-self`, not a standalone installer.

**Impact**: Users must manually find and download the correct binary for their platform.

### 5. UAC Elevation Prompt Friction

When user selects "Program Files" path, script tries `Start-Process -Verb RunAs`
which triggers UAC prompt. Combined with SmartScreen, this creates double-prompt friction.

### 6. No Error Recovery Guidance

When something fails, user gets generic "Installation failed" with no actionable guidance
on how to resolve the issue.

## Proposed Improvements

### P1 - Critical (Must Have)

1. **Add self-contained binary download**
   - Detect platform (amd64/arm64) automatically
   - Download latest release from GitHub Releases API
   - Verify SHA256 checksum after download
   - Make script work as true standalone installer

2. **Add execution policy bypass wrapper**
   - Provide one-liner for README: `irm https://raw.githubusercontent.com/.../install.ps1 | iex`
   - Or batch file wrapper: `install.bat` that calls PowerShell with `-ExecutionPolicy Bypass`
   - Document both methods prominently

3. **Handle Zone.Identifier / SmartScreen**
   - Auto-unblock downloaded files with `Unblock-File`
   - Or use `Invoke-WebRequest` + `Invoke-Expression` pattern that avoids file-based execution

### P2 - Important (Should Have)

4. **Provide install.bat bootstrap**
   - Simple `.bat` file that users can double-click
   - Internally calls PowerShell with correct flags
   - Avoids execution policy issues entirely
   - Example: `powershell -ExecutionPolicy Bypass -File "%~dp0install.ps1"`

5. **Improve error messages with resolution steps**
   - "Execution policy blocked" -> show exact command to fix
   - "Binary not found" -> show download URL
   - "Access denied" -> suggest running as admin or using different path

6. **Add --dry-run / --check mode**
   - Show what would be installed and where
   - Check prerequisites without making changes

### P3 - Nice to Have

7. **WinGet integration**
   - Publish Portunix to WinGet repository
   - Users can install via `winget install portunix`
   - Eliminates all script-related issues

8. **Scoop integration**
   - Create Scoop bucket for Portunix
   - `scoop install portunix`

9. **MSI installer option**
   - Proper Windows installer with digital signature
   - No script execution issues
   - Professional appearance

## Acceptance Criteria

- [ ] User can install Portunix on fresh Windows with single command/double-click
- [ ] No manual execution policy changes required
- [ ] No manual binary download required
- [ ] Script downloads correct binary for platform automatically
- [ ] Clear error messages with resolution steps
- [ ] Works without admin rights (default path)
- [ ] Checksum verification for downloaded binary

## Technical Notes

- Current `install.sh` (Linux) already has download capability - use as reference
- GitHub Releases API: `https://api.github.com/repos/cassandragargoyle/portunix/releases/latest`
- Consider using `Invoke-RestMethod` for API calls and `Invoke-WebRequest` for downloads
- SHA256 verification: `Get-FileHash -Algorithm SHA256`

## References

- Current script: `scripts/install.ps1`
- Linux installer: `scripts/install.sh`
- ADR-031: Cross-Platform Binary Distribution
- Related: #125 (Cross-Platform Binary Distribution)
