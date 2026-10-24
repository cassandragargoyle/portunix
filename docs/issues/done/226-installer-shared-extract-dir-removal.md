# Issue #226: ptx-installer offers to delete shared extractTo directory (/usr/local/bin, /usr/local) on reinstall

**Status**: Implemented
**Priority**: Critical
**Type**: Bug
**Component**: ptx-installer
**Created**: 2026-10-07
**Closed**: 2026-10-07
**GitHub**: #226
**Related**: #227
**Acceptance**: [acceptance-226.md](../../testing/acceptance-226.md) — PASS (Linux)

## Implementation

Fixed in `ad612b3` (branch `fix/226-installer-shared-extract-dir`):

- `engine/shared_dirs.go` — `isSharedDirectory()`: well-known shared directories and
  filesystem roots are never removed, not even with `--force`
- `engine/installer.go` — `prepareSharedDirectory()`: files from the archive replace existing
  ones in place; only a top-level directory created by the archive (e.g. `/usr/local/go`) is
  offered for removal
- `src/pkg/archive` — tar extraction replaces existing files and symlinks instead of writing
  over them; new `ListTopLevelEntries()`
- `hugo.json` — removed the non-existent Windows arm64 extended URL

Not done here:

- Proposed Approach item 4 (package-owned directory plus symlink) — left for a separate issue
- Windows behaviour is compiled but was not executed; see Recommendations in the acceptance
  protocol
- False "Installation failed" from `postInstall` commands found during testing — tracked as #227

## Summary

Installing a `tar.gz` / `zip` package whose `extractTo` is a shared system directory
(`/usr/local/bin`, `/usr/local`) asks the user to **remove the whole directory** when it is
not empty. Answering `y` — or passing `--force`, which skips the question — runs
`os.RemoveAll` on the shared directory and deletes every unrelated tool installed there.

## Problem Description

Reproduction (Linux, `/usr/local/bin` already contains other tools):

```bash
portunix install hugo --variant extended
```

```text
⚠️  Target directory already exists and is not empty:
   /usr/local/bin
   Contains 10 item(s)

Do you want to remove the existing directory and reinstall? [y/N]:
```

Answering `y` would delete `act`, `docker-compose`, `tea`, `thrift`, the installed
`portunix` / `ptx-*` binaries and anything else in `/usr/local/bin`. With `--force` the
directory is removed without any question.

Root cause: `checkExistingDirectory()` in `src/helpers/ptx-installer/engine/installer.go`
treats `extractTo` as a directory owned by the package. Its "not empty → remove and
reinstall" logic is only correct for package-private directories (for example
`~/.local/share/portunix/packages/<name>`, `C:/Portunix/bin/<name>`).

Packages with a shared `extractTo` on Linux:

| Package | extractTo |
| ------- | --------- |
| actionlint, act, caddy, hugo, hugo-extended, ninja, terraform | `/usr/local/bin` |
| go, protoc | `/usr/local` |

For `go` / `protoc`, `--force` would wipe the whole `/usr/local` tree.

Secondary problem found at the same time: `hugo.json` variant `extended` (Windows)
references `hugo_extended_<version>_windows-arm64.zip`, which Hugo does not publish
(neither 0.150.1 nor 0.167.0) — the URL always returns 404.

## Proposed Approach

1. Never remove a directory that the package does not own:
   - refuse removal for well-known shared paths (`/`, `/usr`, `/usr/local`, `/usr/local/bin`,
     `/usr/bin`, `/opt`, `$HOME`, `~/.local/bin`, `C:/Windows`, `C:/Program Files`, …), and/or
   - only remove when the directory is a per-package directory (path ends with the package
     name or is under the Portunix packages root)
2. For shared directories, overwrite only the files contained in the archive (in place),
   optionally after backing up the previous binary
3. Apply the same rule to `--force` — it must never widen the removal scope
4. Consider extracting to a package-private directory plus a symlink in `/usr/local/bin`
   instead of extracting straight into a shared directory
5. Remove (or replace with a valid asset) the non-existent Windows arm64 extended Hugo URL

## Acceptance Criteria

1. Reinstalling hugo, act, actionlint, caddy, ninja, terraform, go, protoc never deletes
   files that are not part of the package archive — with and without `--force`
2. Reinstall over an existing installation in a shared directory upgrades the binary in place
3. Per-package directories keep the current "remove and reinstall" behaviour
4. Unit tests cover the shared-path guard (including `--force`)
5. `hugo.json` contains no URL that returns 404

## Testing

Container only (never on the host):

```bash
portunix container run-in-container hugo --image ubuntu:24.04
# inside: create dummy files in /usr/local/bin, then
portunix install hugo --variant extended          # reinstall, answer y
portunix install hugo --variant extended --force
# verify the dummy files still exist and hugo is upgraded
```
