# Windows Setup Guide

This guide covers Windows-specific setup and configuration for Portunix development.

## PowerShell UTF-8 Configuration

Portunix uses emoji and UTF-8 characters in build output. To display these correctly in PowerShell, you need to configure UTF-8 encoding.

### Quick Fix (Current Session Only)

Run this command in your PowerShell session before using `make`:

```powershell
chcp 65001
```

### Permanent Fix (Recommended)

Add UTF-8 configuration to your PowerShell profile:

1. **Open PowerShell profile for editing:**
   ```powershell
   notepad $PROFILE
   ```

   If the file doesn't exist, create it first:
   ```powershell
   New-Item -Path $PROFILE -Type File -Force
   notepad $PROFILE
   ```

2. **Add these lines to the profile:**
   ```powershell
   # Set UTF-8 encoding for proper emoji display
   $OutputEncoding = [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
   [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
   chcp 65001 | Out-Null
   ```

3. **Save and reload:**
   ```powershell
   . $PROFILE
   ```

4. **Verify it works:**
   ```powershell
   make help
   ```

   You should now see emoji correctly displayed: 🔧 📖 💡

### Alternative: Windows Terminal

If you use [Windows Terminal](https://aka.ms/terminal), it has better UTF-8 support out of the box. You may only need to:

1. Set the font to one that supports emoji (e.g., "Cascadia Code", "Segoe UI Emoji")
2. Open Settings → Profiles → Defaults → Appearance → Font face

## Font Recommendations

For best emoji support in Windows, use one of these fonts:

- **Cascadia Code** (recommended, includes Powerline glyphs)
- **Segoe UI Emoji** (Windows default emoji font)
- **Noto Color Emoji**
- **JetBrains Mono** with emoji fallback

## Git Bash / MSYS2 Users

If you're using Git Bash or MSYS2:

1. Add to `~/.bashrc`:
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```

2. Restart your terminal

## Makefile Commands

After configuring UTF-8, all `make` commands should display emoji correctly:

```bash
make help      # 🔧 Shows all available commands
make build     # 🎉 Build project
make test      # 🧪 Run tests
make clean     # 🧹 Clean build artifacts
```

## Troubleshooting

### Emoji still not displaying?

1. **Check your PowerShell version:**
   ```powershell
   $PSVersionTable.PSVersion
   ```
   Recommended: PowerShell 7.x (cross-platform)

2. **Check current encoding:**
   ```powershell
   [Console]::OutputEncoding
   [Console]::InputEncoding
   ```
   Both should show UTF-8.

3. **Try Windows Terminal:**
   Modern replacement for CMD/PowerShell with better Unicode support.

### Make commands fail on Windows?

Ensure you have:
- **Make** installed (via Chocolatey: `choco install make`)
- **Git for Windows** with MSYS2 tools
- Or use **WSL2** for Linux-like environment

## Development Tools

### Recommended Chocolatey Packages

```powershell
# Package manager for Windows
choco install chocolatey

# Development essentials
choco install make
choco install git
choco install golang
choco install docker-desktop

# Optional but recommended
choco install windows-terminal
choco install vscode
```

### Portunix Installation Profiles

Portunix can install development tools automatically:

```bash
# Install default development environment
portunix install default

# Or choose a specific profile
portunix install minimal   # Python only
portunix install full      # Python + Java + Go + VS Code
```

## See Also

- [Contributing Guide](contributing/README.md)
- [Testing Methodology](contributing/TESTING_METHODOLOGY.md)
- [Features Overview](FEATURES_OVERVIEW.md)

---

**Last Updated:** 2025-01-04
**Maintainer:** Portunix Development Team
