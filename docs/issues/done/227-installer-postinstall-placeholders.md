# Issue #227: ptx-installer does not substitute ${sudo_prefix} and ${actual_extract_to} in postInstall commands

**Status**: Implemented
**Priority**: High
**Type**: Bug
**Component**: ptx-installer
**Created**: 2026-10-07
**Closed**: 2026-10-10
**GitHub**: #227
**Related**: #226, #186

## Summary

Package manifests use the placeholders `${sudo_prefix}` and `${actual_extract_to}` in
`postInstall` commands. They were substituted by the legacy installer in `src/app/install`,
which was removed in #186. ptx-installer only substitutes `${install_path}` /
`%install_path%`, so the shell expands the two unknown placeholders to empty strings.

## Problem Description

Found while testing #226 in an `ubuntu:24.04` container:

```bash
portunix install hugo --variant extended
```

```text
✅ Extracted 3 files
   ❌ Command failed: ${sudo_prefix}chmod +x ${actual_extract_to}/hugo (error: exit status 1)
❌ Installation failed: 1 post-install command(s) failed:
```

The command becomes `chmod +x /hugo`, which fails. Hugo itself is extracted correctly and
works, but `portunix install` exits with an error.

Affected manifests:

| Package | Command | Effect |
| ------- | ------- | ------ |
| hugo (`standard`, `extended`) | `${sudo_prefix}chmod +x ${actual_extract_to}/hugo` | Install reports failure |
| nodejs (3 variants) | `${sudo_prefix}ln -sf /usr/bin/node /usr/local/bin/node \|\| true` | Runs without sudo; silently does nothing for non-root users |
| python (apt / dnf variants) | `${sudo_prefix}update-alternatives --install … \|\| true` | Runs without sudo; silently does nothing for non-root users |

The same container run of #226 showed two related `postInstall` problems:

| Package | Command | Effect |
| ------- | ------- | ------ |
| go (`latest`) | `sudo tar -C /usr/local -xzf ${downloaded_file}` | `${downloaded_file}` is never substituted; hard-coded `sudo` fails with exit 127 where sudo is not installed (containers, root shells). The engine has already extracted the archive, so the command is redundant |
| go (`latest`) | `echo 'export PATH=…' \| sudo tee /etc/profile.d/go.sh` | Hard-coded `sudo` fails with exit 127 as root without sudo |
| protoc (`latest`) | `sudo chmod +x /usr/local/bin/protoc` | Hard-coded `sudo` fails with exit 127 as root without sudo |

In all three cases the package is installed and works, but `portunix install` reports
"Installation failed".

## Proposed Approach

1. Substitute in `postInstall` commands of `installArchive()` (and other install types that
   run `postInstall`):
   - `${actual_extract_to}` — the directory the archive was really extracted to (after the
     user-directory fallback), same value as `${install_path}`
   - `${sudo_prefix}` — `sudo ` when the process is not root and the command needs
     elevation, empty when running as root or on Windows
2. Fail loudly on any unknown `${…}` placeholder left in a `postInstall` command instead of
   letting the shell expand it to an empty string
3. Add a manifest check (unit test over `assets/packages/*.json`) that every placeholder
   used in a manifest is one the engine supports
4. Replace hard-coded `sudo` in manifests with `${sudo_prefix}` and drop the redundant
   `tar` re-extraction from `go.json`

## Acceptance Criteria

1. `portunix install hugo --variant extended` succeeds in a clean `ubuntu:24.04` container
   (as root and as a non-root user with sudo)
2. nodejs and python post-install steps run with sudo for non-root users
3. An unknown placeholder in `postInstall` produces a clear error naming the placeholder
4. `portunix install go` and `portunix install protoc` succeed as root in a container
   without sudo installed
5. Unit tests cover placeholder substitution and the manifest check

## Testing

Container only:

```bash
portunix container run-in-container hugo --image ubuntu:24.04
portunix install hugo --variant extended; echo "exit=$?"
```
