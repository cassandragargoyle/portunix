# ADR-028: Package Installation Fallback Paths

## Status
Implemented

## Date
2025-12-03

## Context

When installing packages via `ptx-installer`, the default extraction paths often require administrator privileges:
- Windows: `C:\Program Files\<package>` requires elevated permissions
- Linux: `/usr/local/bin` requires sudo

This causes installation failures for users who:
- Don't have admin rights on their machine
- Prefer not to run installations as administrator
- Want quick, non-intrusive installations

Error example:
```
❌ Installation failed: failed to create extract directory: mkdir C:/Program Files/Hugo: Access is denied.
```

## Decision

Implement a fallback mechanism in `ptx-installer` that automatically falls back to user-writable directories when the primary (admin) path fails due to permission errors.

### Fallback Paths by Platform

**Windows:**
- Primary: Package-defined path (e.g., `C:\Program Files\Hugo`)
- Fallback: `%USERPROFILE%\AppData\Local\Programs\<package-name>`

**Linux/macOS:**
- Primary: Package-defined path (e.g., `/usr/local/bin`)
- Fallback: `~/.local/share/portunix/packages/<package-name>`

### Rationale for Windows Fallback Path

`AppData\Local\Programs` was chosen because:
1. **Standard location** - Microsoft recommends this for per-user application installations
2. **No admin rights needed** - Users always have write access
3. **Survives Windows updates** - Not affected by system updates
4. **Clean separation** - User binaries separate from system binaries
5. **Discoverable** - Standard path that other tools recognize

### Implementation

```go
// Determine fallback directory based on OS
var fallbackDir string
if runtime.GOOS == "windows" {
    // Windows: Use AppData\Local\Programs (standard user app location)
    fallbackDir = filepath.Join(homeDir, "AppData", "Local", "Programs", options.PackageName)
} else {
    // Linux/macOS: Use ~/.local/share/portunix/packages
    fallbackDir = filepath.Join(homeDir, ".local", "share", "portunix", "packages", options.PackageName)
}

// Try primary path first, fallback on permission error
if err := os.MkdirAll(extractTo, 0755); err != nil {
    if extractTo != fallbackDir {
        fmt.Printf("⚠️  Cannot create %s (permission denied), using user directory\n", extractTo)
        extractTo = fallbackDir
        // Retry with fallback
    }
}
```

## Consequences

### Positive
- **No more permission errors** - Installations succeed without admin rights
- **Better UX** - Users don't need to run as administrator
- **Transparent fallback** - Clear message when fallback is used
- **Consistent behavior** - Same pattern across all packages
- **Standard paths** - Uses OS-recommended user directories

### Negative
- **PATH configuration** - Users may need to add fallback path to their PATH
- **Multiple locations** - Same package could be in different locations on different systems
- **Symlink handling** - Need to update symlink creation for fallback paths

### Neutral
- Package definitions still specify preferred (admin) paths
- Fallback is automatic and doesn't require configuration
- Original behavior preserved when running with admin rights

## Alternatives Considered

1. **Always install to user directory** - Rejected because some users prefer system-wide installation
2. **Prompt user for location** - Rejected for non-interactive/scripted installations
3. **Require admin rights** - Rejected as it reduces usability
4. **Use C:\Portunix** - Rejected as it's non-standard and may still require admin rights

## Related

- Package registry format: `assets/packages/*.json`
- Installer engine: `src/helpers/ptx-installer/engine/installer.go`
- Similar pattern used by: Chocolatey (user mode), Scoop, WinGet
