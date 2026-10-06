# Issue #131: OpenSSH Installation Refactoring

## Summary

Refactor OpenSSH installation to use standard paths, proper host key generation, and clear variant logic.

## Current Problems

1. **Non-standard path**: `C:\Portunix\OpenSSH` is not standard, causes confusion
2. **Missing host keys**: `ssh-keygen -A` not called, sshd fails with "no hostkeys available"
3. **Inconsistent reinstall**: Installing to different paths on reinstall
4. **Unused script**: `Install-PortableOpenSSH.ps1` exists but ptx-installer doesn't use it

## Proposed Solution

### Clear Variant Structure

| Variant | Purpose | Install Path | Admin | Service |
|---------|---------|--------------|-------|---------|
| `client` | SSH client only | `%LOCALAPPDATA%\Programs\OpenSSH` | No | No |
| `server` | Full SSH server | `C:\Program Files\OpenSSH-Win64` | Yes | Yes (sshd) |
| `portable` | No system changes | User-specified | No | No |

### Server Variant Requirements

1. Extract to `C:\Program Files\OpenSSH-Win64`
2. Create `C:\ProgramData\ssh` directory
3. Copy `sshd_config_default` to `C:\ProgramData\ssh\sshd_config`
4. **Generate host keys**: `ssh-keygen -A`
5. Fix ACL permissions (FixHostFilePermissions.ps1)
6. Create sshd service
7. Add firewall rule
8. Start service

### Implementation Options

**Option A**: Embed and use `Install-PortableOpenSSH.ps1` in ptx-installer

- Pros: Script already handles all edge cases
- Cons: Need to modify ptx-installer to support embedded scripts

**Option B**: Expand postInstall commands in openssh.json

- Pros: Simple, no code changes
- Cons: Complex JSON, hard to maintain

**Option C**: Create dedicated OpenSSH installer in ptx-installer Go code

- Pros: Full control, proper error handling
- Cons: More code to maintain

### Recommended: Option A

Modify ptx-installer to:

1. Support `installScript` field in package JSON
2. Embed `Install-PortableOpenSSH.ps1`
3. Extract and execute script during installation

## Updated openssh.json Structure

```json
{
  "server": {
    "version": "latest",
    "description": "SSH client and server (requires admin)",
    "urls": {
      "x64": "https://github.com/PowerShell/Win32-OpenSSH/releases/latest/download/OpenSSH-Win64.zip",
      "x86": "https://github.com/PowerShell/Win32-OpenSSH/releases/latest/download/OpenSSH-Win32.zip"
    },
    "extractTo": "C:/Program Files",
    "requiresAdmin": true,
    "installScript": "windows/Install-PortableOpenSSH.ps1",
    "installScriptArgs": "-OpenSSHPath \"${install_path}/OpenSSH-Win64\" -Verbose"
  }
}
```

Note: `extractTo` is set to `C:/Program Files` because the ZIP archive contains `OpenSSH-Win64/` folder.

## Implementation Status

### Implemented (Option A):

1. **registry.go**: Added `InstallScriptArgs` field to `VariantSpec` struct
2. **assets.go**: Already embeds `assets/scripts/windows/*.ps1` and `*.cmd`
3. **engine/installer.go**:
   - Added `EmbeddedScriptsFS` variable and `SetEmbeddedScripts()` function
   - Modified `installScript()` to detect embedded scripts vs inline commands
   - Added `executeEmbeddedScript()` function to extract and run embedded PowerShell scripts
   - Added `isEmbeddedScript()` helper to detect script file paths
   - Modified `installArchive()` to execute `installScript` after extraction
4. **main.go**: Calls `engine.SetEmbeddedScripts(embeddedScripts)` in init()
5. **openssh.json**: Updated server variant with new structure

### Script Execution Flow:

1. Download OpenSSH archive
2. Extract to `C:/Program Files/` (creates `C:/Program Files/OpenSSH-Win64/`)
3. Detect embedded script in `installScript` field
4. Extract `Install-PortableOpenSSH.ps1` from embedded assets to temp directory
5. Execute script with arguments: `-OpenSSHPath "C:/Program Files/OpenSSH-Win64" -Verbose`
6. Script handles: service creation, host key generation, firewall, permissions

## Acceptance Criteria

1. [x] Server variant installs to standard path `C:\Program Files\OpenSSH-Win64`
2. [x] Host keys are always generated (`ssh-keygen -A`) - handled by embedded script
3. [x] sshd service starts successfully after installation - handled by embedded script
4. [ ] Reinstallation works correctly (same path, regenerates keys if missing) - needs VM testing
5. [x] Clear error messages if installation fails - error handling in Go code
6. [ ] Documentation updated with correct paths

## Migration Notes

- Existing installations at `C:\Portunix\OpenSSH` should be detected and migrated or warned about
- Users may need to manually remove old installation

## Priority

High - SSH server installation is broken

## Labels

bug, openssh, windows, installation, refactoring

## Related Issues

- #127 Migrate Win32-OpenSSH Installation to ptx-installer
- #129 Docusaurus QuickStart Script (discovered during testing)
