# Issue #146: Rust Programming Language Installation Support

## Overview

Add Rust programming language as an installable package in ptx-installer. Rust is a systems programming language focused on safety, speed, and concurrency. The primary installation method is via **rustup** - the official Rust toolchain installer and version manager.

## Motivation

- Rust is one of the most popular systems programming languages (consistently #1 "most loved" in Stack Overflow surveys)
- Widely used for CLI tools, web backends, WebAssembly, blockchain, and embedded systems
- MCP handlers already detect Rust projects (Cargo.toml) and suggest installation, but ptx-installer cannot actually install it
- Rustup provides a standardized cross-platform installation mechanism
- Many Portunix-ecosystem tools are written in Rust (ripgrep, fd, bat, etc.)

## Requirements

### Functional Requirements

- [ ] FR-1: Install Rust stable via `portunix install rust` command
- [ ] FR-2: Support rustup-based installation (primary method for all platforms)
- [ ] FR-3: Support package manager variants (apt, dnf, pacman, chocolatey)
- [ ] FR-4: Cross-platform support (Linux x64/ARM64, Windows x64, macOS x64/ARM64)
- [ ] FR-5: Post-installation verification via `rustc --version` and `cargo --version`
- [ ] FR-6: Support variant selection: stable, beta, nightly
- [ ] FR-7: Include essential components: rustfmt, clippy, rust-analyzer

### Non-Functional Requirements

- [ ] NFR-1: Follow existing ptx-installer package definition pattern
- [ ] NFR-2: Auto-discovery by registry (no manual index updates)
- [ ] NFR-3: Respect user home directory for rustup/cargo installation paths

## Technical Design

### Package Definition

Create file: `src/helpers/ptx-installer/assets/packages/rust.json`

Pattern: Unique - rustup is a toolchain installer (similar concept to how Java uses Adoptium)

### Installation Methods

| Platform | Method | Package/URL |
| -------- | ------ | ----------- |
| Linux | rustup (recommended) | `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \| sh -s -- -y` |
| Linux | APT | `rustc` (often outdated - not recommended) |
| Linux | DNF/YUM | `rust cargo` |
| Linux | Pacman | `rust` |
| Windows | rustup-init.exe (recommended) | Direct download from `https://win.rustup.rs/x86_64` |
| Windows | Chocolatey | `rust` |
| Windows | WinGet | `Rustlang.Rustup` |
| macOS | rustup (recommended) | Same curl command as Linux |
| macOS | Homebrew | `rust` |

### Download URLs

Source: `https://rustup.rs/` and `https://github.com/rust-lang/rustup`

URL pattern:

```text
# Rustup installers
Linux/macOS: https://sh.rustup.rs (shell script)
Windows x64: https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-msvc/rustup-init.exe
Windows x86: https://static.rust-lang.org/rustup/dist/i686-pc-windows-msvc/rustup-init.exe

# Standalone toolchains (alternative - for offline/container use)
https://static.rust-lang.org/dist/rust-{version}-x86_64-unknown-linux-gnu.tar.gz
https://static.rust-lang.org/dist/rust-{version}-x86_64-pc-windows-msvc.tar.gz
https://static.rust-lang.org/dist/rust-{version}-aarch64-unknown-linux-gnu.tar.gz
```

### Key Technical Details

- **Current version**: 1.84.0 (as of January 2025)
- **License**: MIT OR Apache-2.0
- **Toolchain size**: ~400MB (full toolchain with std)
- **Dependencies**:
  - Linux: gcc/cc linker, optional pkg-config, libssl-dev for many crates
  - Windows: MSVC Build Tools or MinGW-w64
- **Default install path (Linux)**: `$HOME/.rustup` (toolchains), `$HOME/.cargo/bin` (binaries)
- **Default install path (Windows)**: `%USERPROFILE%\.rustup`, `%USERPROFILE%\.cargo\bin`
- **Non-interactive flag**: `rustup-init -y` (skip confirmation prompts)

### Variant Support

| Variant | Description | Use Case |
| ------- | ----------- | -------- |
| `stable` (default) | Latest stable release | Production development |
| `beta` | Next stable preview | Testing upcoming features |
| `nightly` | Latest experimental | Nightly-only features, proc macros |
| `apt` / `dnf` / `pacman` | System package manager | Quick install, may be outdated |
| `chocolatey` / `winget` | Windows package manager | Windows managed install |

### Environment Variables

```bash
# Set by rustup
RUSTUP_HOME=$HOME/.rustup
CARGO_HOME=$HOME/.cargo
PATH=$HOME/.cargo/bin:$PATH
```

### Post-Install Steps

```bash
# Verify installation
rustc --version
cargo --version

# Install essential components (if not included by default)
rustup component add rustfmt clippy rust-analyzer
```

### Verification

```bash
rustc --version
# Expected output: rustc 1.84.0 (9fc6b4312 2025-01-07)

cargo --version
# Expected output: cargo 1.84.0 (fb7e1f128 2025-01-07)
```

## Architecture Considerations

### Rustup vs Direct Binary

Rustup is strongly recommended over standalone binary distribution because:

1. **Version management**: Switch between stable/beta/nightly
2. **Cross-compilation targets**: `rustup target add wasm32-unknown-unknown`
3. **Component management**: Add/remove rustfmt, clippy, rust-analyzer
4. **Self-update**: `rustup update` keeps toolchain current
5. **Official recommendation**: Rust project recommends rustup as the only install method

### Container Testing Specifics

Rust installation in containers requires attention to:

- Linker availability (gcc/cc must be present)
- Disk space (~400MB for full toolchain)
- Non-interactive mode (`-y` flag for rustup)
- PATH setup within container session

## Implementation Scope

### Files to Create

| File | Description |
| ---- | ----------- |
| `src/helpers/ptx-installer/assets/packages/rust.json` | Package definition JSON |

### Files to Modify

None required - registry auto-discovers packages from `assets/packages/` directory.

## Acceptance Criteria

1. `portunix install rust` successfully installs Rust stable on Linux (via rustup)
2. `portunix install rust --variant apt` installs via APT on Debian/Ubuntu
3. `portunix install rust` successfully installs Rust on Windows (via rustup-init.exe)
4. `portunix install rust --variant chocolatey` installs via Chocolatey on Windows
5. `rustc --version` and `cargo --version` return expected versions after installation
6. Package appears in `portunix install --list` output
7. Installation tested in clean container environment
8. Variant selection (stable/beta/nightly) works correctly

## Testing

All installation testing MUST be performed in containers per project methodology:

```bash
# Linux testing (Ubuntu)
portunix container run --image ubuntu:22.04
# inside container:
apt update && apt install -y curl gcc
./portunix install rust
source $HOME/.cargo/env
rustc --version
cargo --version

# Linux testing (Fedora)
portunix container run --image fedora:latest
# inside container:
./portunix install rust
source $HOME/.cargo/env
rustc --version

# Variant testing
./portunix install rust --variant apt
./portunix install rust --variant nightly
```

## Reference

- Manifest: `portunix-architecture/docs/manifests/manifest-rust.md`
- Rust website: <https://www.rust-lang.org/>
- Rustup: <https://rustup.rs/>
- GitHub (Rust): <https://github.com/rust-lang/rust>
- GitHub (Rustup): <https://github.com/rust-lang/rustup>

## Complexity

**Medium** - Rustup-based installation pattern differs from simple binary extraction. Requires handling of shell script execution (Linux) and exe download (Windows), PATH configuration via cargo env, and variant support for toolchain channels.

---

**Created**: 2026-02-08
**Closed**: 2026-02-08
**Author**: Architect
**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, package-management, rust, programming-language, cross-platform, ptx-installer, rustup
