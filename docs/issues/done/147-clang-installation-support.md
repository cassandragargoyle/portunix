# Issue #147: Clang/LLVM Installation Support

## Overview

Add Clang (LLVM C/C++/Objective-C compiler) as an installable package in ptx-installer. Clang is a production-quality compiler built on the LLVM infrastructure, known for expressive diagnostics, fast compilation, modular architecture, and a rich ecosystem of developer tools (clang-tidy, clang-format, clangd, sanitizers).

## Motivation

- Clang is the default compiler on macOS (Xcode), FreeBSD, Android NDK, and ChromeOS
- Growing adoption on Linux as primary or alternative compiler alongside GCC
- Rich tooling ecosystem: clang-tidy (400+ checks), clang-format, clangd (LSP), sanitizers (ASan, MSan, TSan, UBSan)
- Required for WebAssembly compilation via Emscripten
- Enterprise adoption: Apple, Google, Microsoft, Sony, AMD, Intel, ARM, IBM, Nvidia
- Complements existing C/C++ development documentation in Portunix (`docs/contributing/TESTING-CPP.md`, `CODE-STYLE-CPP.md`)
- Complements Ninja (#145) which is commonly used as Clang/LLVM build backend

## Requirements

### Functional Requirements

- [ ] FR-1: Install Clang via `portunix install clang` command
- [ ] FR-2: Support LLVM APT repository for Debian/Ubuntu (latest versions)
- [ ] FR-3: Support system package manager variants (apt, dnf, pacman, chocolatey)
- [ ] FR-4: Cross-platform support (Linux x64/ARM64, Windows x64)
- [ ] FR-5: Post-installation verification via `clang --version` and `clang++ --version`
- [ ] FR-6: Support version-specific installation (e.g. clang-21, clang-20)
- [ ] FR-7: Optional full LLVM toolchain variant including clang-tidy, clang-format, clangd, lld

### Non-Functional Requirements

- [ ] NFR-1: Follow existing ptx-installer package definition pattern
- [ ] NFR-2: Auto-discovery by registry (no manual index updates)
- [ ] NFR-3: Prefer LLVM APT repository on Debian/Ubuntu for latest version availability

## Technical Design

### Package Definition

Create file: `src/helpers/ptx-installer/assets/packages/clang.json`

Pattern: Similar to system package manager + script-based installation (LLVM APT script for Debian/Ubuntu)

### Installation Methods

| Platform | Method | Package/URL |
| -------- | ------ | ----------- |
| Linux (Debian/Ubuntu) | LLVM APT script (recommended) | `wget https://apt.llvm.org/llvm.sh && chmod +x llvm.sh && sudo ./llvm.sh 21` |
| Linux (Debian/Ubuntu) | APT (system) | `clang` (often outdated) |
| Linux (Fedora) | DNF | `clang clang-tools-extra` |
| Linux (Arch) | Pacman | `clang` |
| Linux | Pre-built binaries | `https://github.com/llvm/llvm-project/releases` |
| Windows | Chocolatey | `llvm` |
| Windows | WinGet | `LLVM.LLVM` |
| Windows | Pre-built installer | LLVM-{version}-win64.exe from GitHub releases |

### Download URLs

Source: `https://github.com/llvm/llvm-project/releases`

URL pattern:

```text
# LLVM APT setup script (Debian/Ubuntu - recommended)
https://apt.llvm.org/llvm.sh

# Pre-built binaries (GitHub releases)
https://github.com/llvm/llvm-project/releases/download/llvmorg-{version}/LLVM-{version}-Linux-X64.tar.xz
https://github.com/llvm/llvm-project/releases/download/llvmorg-{version}/LLVM-{version}-Linux-AArch64.tar.xz
https://github.com/llvm/llvm-project/releases/download/llvmorg-{version}/LLVM-{version}-win64.exe

# LLVM APT repository (versioned packages)
apt install clang-21 clang-tools-21 clang-format-21 clang-tidy-21 lld-21
```

### Key Technical Details

- **Current version**: 21.x (LLVM 21 series, stable as of 2025; LLVM 22 expected ~March 2026)
- **License**: Apache-2.0 with LLVM Exceptions
- **Binary size**: ~100MB (clang only), ~500MB+ (full LLVM toolchain with tools)
- **Dependencies**:
  - Linux: libc, libstdc++ (or libc++), zlib
  - Windows: MSVC runtime or MinGW-w64
- **LLVM APT versioned packages**: `clang-21`, `clang-format-21`, `clang-tidy-21`, `lld-21`, `lldb-21`
- **System packages (generic)**: `clang`, `clang-tools-extra` (varies by distro)

### Variant Support

| Variant | Description | Use Case |
| ------- | ----------- | -------- |
| `default` | LLVM APT script (Debian/Ubuntu) or system package | Latest stable version |
| `apt` | System APT package | Quick install, may be older version |
| `dnf` | Fedora/RHEL system package | Fedora managed install |
| `pacman` | Arch Linux system package | Arch managed install |
| `full` | Full LLVM toolchain (clang + clang-tidy + clang-format + clangd + lld + lldb) | Complete development setup |
| `chocolatey` / `winget` | Windows package manager | Windows managed install |

### Environment Variables

```bash
# Typical setup after LLVM APT installation
export CC=clang
export CXX=clang++

# If version-specific (e.g. clang-21)
export CC=clang-21
export CXX=clang++-21

# Optional: use lld linker for faster linking
export LDFLAGS="-fuse-ld=lld"
```

### Verification

```bash
clang --version
# Expected output: clang version 21.x.x (...)

clang++ --version
# Expected output: clang version 21.x.x (...)

# Full variant verification
clang-tidy --version
clang-format --version
```

## Architecture Considerations

### LLVM APT vs System Package vs Pre-built Binary

- **LLVM APT** (Debian/Ubuntu recommended): Provides latest version, versioned packages, official repository
- **System package**: Easy to install but often outdated (e.g. Ubuntu 22.04 ships Clang 14)
- **Pre-built binaries**: Platform-independent but large download, manual PATH setup
- **Recommendation**: Default to LLVM APT script on Debian/Ubuntu, system package on other distros

### Clang vs Full LLVM Toolchain

Two package granularity options:

1. **`clang`** package: Only the compiler (clang, clang++)
2. **`full` variant**: Complete toolchain (clang + clang-tidy + clang-format + clangd + lld + lldb)

This follows the pattern where the default is minimal and the `full` variant provides everything.

## Implementation Scope

### Files to Create

| File | Description |
| ---- | ----------- |
| `src/helpers/ptx-installer/assets/packages/clang.json` | Package definition JSON |

### Files to Modify

None required - registry auto-discovers packages from `assets/packages/` directory.

## Acceptance Criteria

1. `portunix install clang` successfully installs Clang on Debian/Ubuntu (via LLVM APT)
2. `portunix install clang --variant apt` installs via system APT on Debian/Ubuntu
3. `portunix install clang --variant dnf` installs via DNF on Fedora
4. `portunix install clang --variant full` installs full LLVM toolchain
5. `portunix install clang` successfully installs Clang on Windows (via Chocolatey or WinGet)
6. `clang --version` and `clang++ --version` return expected versions after installation
7. Package appears in `portunix install --list` output
8. Installation tested in clean container environment

## Testing

All installation testing MUST be performed in containers per project methodology:

```bash
# Linux testing (Ubuntu - LLVM APT)
portunix container run --image ubuntu:22.04
# inside container:
apt update && apt install -y wget lsb-release software-properties-common gnupg
./portunix install clang
clang --version
clang++ --version

# Linux testing (Fedora)
portunix container run --image fedora:latest
# inside container:
./portunix install clang
clang --version

# Variant testing - full toolchain
./portunix install clang --variant full
clang-tidy --version
clang-format --version

# Variant testing - system APT
./portunix install clang --variant apt
```

## Reference

- Manifest: `portunix-architecture/docs/manifests/manifest-clang.md`
- Clang website: <https://clang.llvm.org/>
- LLVM APT packages: <https://apt.llvm.org/>
- GitHub: <https://github.com/llvm/llvm-project>
- Clang documentation: <https://clang.llvm.org/docs/>
- Related issues: #145 (Ninja - common build backend for LLVM projects)

## Complexity

**Medium** - LLVM APT script-based installation on Debian/Ubuntu adds complexity beyond simple package manager calls. Version-specific package naming (clang-21 vs clang) and the `full` variant with multiple tools require careful JSON definition. Windows installation via pre-built installer or Chocolatey is straightforward.

---

**Status:** ✅ Implemented
**Created**: 2026-02-08
**Closed**: 2026-02-08
**Author**: Architect
**Priority**: Medium
**Type**: Enhancement
**Labels**: enhancement, package-management, clang, llvm, compiler, c-cpp, cross-platform, ptx-installer
