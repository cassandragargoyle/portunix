# Issue #168: GNU Make Installation Support

**Status:** ✅ Implemented

## Summary

Add GNU Make installation support to Portunix with direct binary download on Windows (without dependency on Chocolatey/WinGet) and native package manager on Linux.

## Problem Description

GNU Make is a fundamental build tool used across virtually all software projects. Portunix currently cannot install it:

```bash
portunix install make
# ERROR: Package 'make' not found
```

Portunix itself uses Make (`Makefile` in project root), yet cannot provision it on clean systems.

## Expected Behavior

```bash
# Install GNU Make
portunix install make

# Verify
make --version
```

## Current State

- **53 packages** available in ptx-installer registry
- `make` is **not** among them
- Portunix's own build process depends on `make` (`make build`, `make build-helpers`, etc.)

## Impact

- Clean Windows development environments lack `make`
- Onboarding new developers requires manual installation
- Container-based testing environments need manual setup
- Portunix cannot bootstrap its own build dependency

## Proposed Solution

### Windows: Direct Binary Download (ezwinports)

**Why not Chocolatey/WinGet:**

- Chocolatey requires admin privileges and its own installation first
- WinGet's `GnuWin32.Make` installs ancient version 3.81
- Both add unnecessary dependency on external package managers

**Recommended: ezwinports standalone binary**

- **Source**: [ezwinports on SourceForge](https://sourceforge.net/projects/ezwinports/files/)
- **Version**: 4.4.1 (latest, actively maintained)
- **Archive**: `make-4.4.1-without-guile-w32-bin.zip` (~393 KB)
- **Key advantage**: Fully self-contained with all DLLs, extract and run
- **No admin required**: Can install to user-writable directory

Installation flow:

1. Download ZIP from ezwinports
2. Extract `bin/make.exe` (+ bundled DLLs) to install directory
3. Add to PATH
4. Verify with `make --version`

### Linux: Native Package Manager

Straightforward via system package managers:

- **APT** (Debian/Ubuntu): `apt-get install make`
- **DNF** (Fedora/RHEL): `dnf install make`
- **Pacman** (Arch): `pacman -S make`
- **APK** (Alpine): `apk add make`

### Package Definition Template

```json
{
  "apiVersion": "v1",
  "kind": "Package",
  "metadata": {
    "name": "make",
    "displayName": "GNU Make",
    "description": "GNU Make - build automation tool that controls generation of executables and other files",
    "category": "development/build-tools",
    "homepage": "https://www.gnu.org/software/make/",
    "license": "GPL-3.0-or-later",
    "maintainer": "GNU Project"
  },
  "spec": {
    "platforms": {
      "windows": {
        "type": "zip",
        "variants": {
          "latest": {
            "version": "4.4.1",
            "urls": {
              "x64": "https://sourceforge.net/projects/ezwinports/files/make-4.4.1-without-guile-w32-bin.zip/download"
            }
          }
        },
        "verification": {
          "command": "make --version",
          "expectedExitCode": 0
        }
      },
      "linux": {
        "type": "apt",
        "variants": {
          "apt": { "packages": ["make"] },
          "dnf": { "packages": ["make"] },
          "pacman": { "packages": ["make"] },
          "apk": { "packages": ["make"] }
        },
        "verification": {
          "command": "make --version",
          "expectedExitCode": 0
        }
      }
    },
    "aiPrompts": {
      "versionDiscovery": "Check ezwinports SourceForge page for latest make release builds",
      "updateGuidance": "Monitor GNU Make releases at gnu.org and ezwinports for Windows builds"
    }
  }
}
```

## Note on ezwinports Reliability

The ezwinports project is maintained by **Eli Zaretskii**, co-maintainer of GNU Emacs and a 30+ year contributor to the GNU ecosystem. He received the [FSF Free Software Award in 2023](https://www.fsf.org/news/free-software-awards-winners-announced-eli-zaretskii-tad-skewedzeppelin-gnu-jami). The project provides MinGW-compiled ports of GNU/Unix tools for Windows, created specifically because existing ports (like GnuWin32) were outdated or broken. This is a trustworthy, well-maintained source for Windows binaries.

## Alternative Windows Sources Evaluated

| Source | Version | Status | Notes |
| ------ | ------- | ------ | ----- |
| **ezwinports** | 4.4.1 | **Recommended** | Self-contained, actively maintained, ~393 KB |
| GnuWin32 | 3.81 | Not recommended | Unmaintained since ~2010, very old version |
| Chocolatey | latest | Fallback only | Requires choco pre-installed + admin |
| WinGet | 3.81 | Not recommended | Installs outdated GnuWin32 version |
| MSYS2/MinGW | latest | Overkill | Pulls entire MSYS2 environment |

## Acceptance Criteria

- [ ] `portunix install make` works on Windows (direct download, no choco/winget dependency)
- [ ] `portunix install make` works on Linux (apt, dnf, pacman, apk)
- [ ] `make --version` succeeds after installation
- [ ] PATH is updated correctly on Windows
- [ ] Package definition follows current registry architecture
- [ ] Container-based installation testing passes

## Testing Plan

```bash
# Windows - direct test
portunix install make --dry-run
portunix install make
make --version

# Linux - container test
portunix container run ubuntu:22.04
./portunix install make
make --version
```

## Labels

enhancement, package-management, build-tools, cross-platform, direct-download

## Priority

Medium

## Type

Enhancement
