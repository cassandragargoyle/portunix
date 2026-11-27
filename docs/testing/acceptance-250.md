# Acceptance Protocol - Issue #250 (Windows application icon)

**Issue**: Replace Windows application icon with the Portunix gear icon
**Branch**: `feature/250-replace-windows-icon-with-gear` (commit `d02c56f`; retest 1 on `a514efc`; retest 2 on `6d5750f`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-10-27
**Testing OS**: Ubuntu 26.04.1, kernel 7.0.0-38, amd64 (host); Go 1.26.8; GoReleaser (snapshot)

The change replaces resource files (`.ico`, `.syso`, `versioninfo.json`, `.rc`). No software
was installed and no host configuration was changed, so the tests ran on the host.

## Scope Note

**Independence note**: the change was made by Claude Code in the developer role in the same
session. The scenarios below were designed separately in the tester role.

**Not covered locally**: the visual check in Windows Explorer and the taskbar (AC1) needs a
Windows machine. The embedded resources were verified by extracting the `.rsrc` section of the
built executables with `7z`.

## Test Summary

- Total test scenarios: 8
- Passed: 7
- Failed: 1 (TC06 — AC4, GoReleaser build has no resources)
- Pending: AC1 (Windows visual check)

Retest 1 (`a514efc`) — see [Retest 1](#retest-1):

- Total test scenarios: 6 (R1–R6)
- Passed: 4
- Failed: 2 (R3 — Windows arm64 resources corrupted; R6 — pre-existing, same cause in `ptx-installer`)
- Observations: O2 (pre-existing, `ptx-installer` arm64 manifest)

Retest 2 (`6d5750f`) — see [Retest 2](#retest-2):

- Total test scenarios: 10 (S1–S10)
- Passed: 10
- Failed: 0
- R3 and O2 resolved, new observation O3 (not blocking)
- Pending: AC1 (Windows visual check, TC08)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `.ico` content, references to `home.ico`, `.syso` reproducibility | host |
| Build | `make gen-resources`, `make build`, `GOOS=windows go build .` | host |
| Release | GoReleaser snapshot build of the `portunix` build id, platform build path | host |
| Resources | `.rsrc` of the built `portunix.exe`: icons, version info | host |
| Visual | Explorer views, taskbar | Windows (pending) |

## Test Cases

### TC01: icon file content (AC2)

- **Given** `assets/brand/portunix.ico`
- **When** it is inspected with `magick identify` and its directory parsed
- **Then** it holds four images, 16, 32, 48 and 256 px, with an alpha channel

Result: **PASS** — 16/32/48 px as 32-bit BMP, 256 px as PNG; transparent background
(corner pixel `srgba(0,0,0,0)`); 22 910 bytes. `assets/brand/portunix.svg` is identical to the
brand master in the marketing repository.

### TC02: `home.ico` removed (AC3)

- **When** `git grep -n home.ico` runs outside `docs/issues/`
- **Then** nothing is found and the file does not exist

Result: **PASS** — no match; `home.ico` deleted. The only remaining mentions are in the issue
file itself, which describes the previous state.

### TC03: `make gen-resources` is reproducible (AC4)

- **When** `make gen-resources` runs on the committed tree
- **Then** it succeeds and `portunix.syso` is byte-identical to the committed one

Result: **PASS** — SHA-1 `6972346d…` before and after, working tree clean.

### TC04: `make build` (AC4)

Result: **PASS** — main binary and all 19 helpers built.

### TC05: Windows build from the package (AC2, AC5)

- **When** `GOOS=windows GOARCH=amd64 go build .` runs and the `.rsrc` section is extracted
- **Then** the icon group holds 16/32/48/256 px images and the version info is unchanged

Result: **PASS** — icons 16, 32, 48 (BMP) and 256 (PNG); `FileVersion` / `ProductVersion`
2.7.0, `ProductName` Portunix, `LegalCopyright` "CassandraGargoyle team. All rights
reserved." — same as before. Two icon groups (IDs 2 and 32512) are produced by goversioninfo,
the build from `main` has the same layout.

### TC06: GoReleaser Windows build (AC4)

- **Given** `.goreleaser.yml` with `dist` redirected to a temporary directory (the existing
  `dist/` holds release 2.7.0 artifacts and is not touched)
- **When** `goreleaser build --snapshot --single-target --id portunix` runs for
  `windows/amd64`
- **Then** `portunix.exe` carries the new icon

Result: **FAIL** — the build succeeds, but `portunix.exe` has **no `.rsrc` section at all**:
no icon, no version info, no manifest.

Cause: the `portunix` build in `.goreleaser.yml` uses `main: ./main.go`. When `go build` gets a
list of `.go` files instead of a package, it ignores `.syso` files in the directory. Verified:

| Command | `.rsrc` entries |
| ------- | --------------- |
| `GOOS=windows go build ./main.go` | 0 |
| `GOOS=windows go build .` | 12 |

The defect predates this issue: `portunix.exe` in `dist/portunix_2.7.0_windows_amd64.zip`
(release 2.7.0) has no resources either. It still blocks AC4, because this archive
(`portunix_<version>_windows_<arch>.zip`) is what `scripts/install.ps1` and `portunix update`
download, so users would never see the new icon.

Verified fix: with `main: ./` for the `portunix` build, a full snapshot build of all five
targets (linux amd64/arm64, windows amd64/arm64, darwin amd64) succeeds; both Windows
executables carry the 16/32/48/256 px icons and version info, the Linux and macOS binaries have
no Windows resources, and the Linux binary runs (`--version`). The helper builds already use
`main: ./`.

### TC07: platform archive build path (AC4)

- **When** the build commands of `scripts/make-release.py` (platform archives) and
  `make build-all-platforms` are inspected
- **Then** they build the package, not a file list

Result: **PASS** — both use `go build -o … .`; `dist/platforms/windows-amd64/portunix.exe` of
release 2.7.0 has resources (old icon), so this path picks up the new `.syso`.

### TC08: Explorer and taskbar (AC1)

Result: **PENDING** — needs a Windows machine: Explorer small/details/large/extra large views,
taskbar while `portunix.exe` runs, file properties → Details.

## Observations

- **O1 (informative)**: the icon is black on a transparent background, as chosen by the
  product owner. On a dark taskbar or dark Explorer theme (Windows 11 default) it has very low
  contrast. Worth checking during TC08.

## Required Changes

1. `.goreleaser.yml`, build `id: portunix`: change `main: ./main.go` to `main: ./` so the
   Windows build links `portunix.syso` (TC06)

## Retest 1

Commit `a514efc` changes `main: ./main.go` to `main: ./` for the `portunix` build in
`.goreleaser.yml` (required change 1). Only `.goreleaser.yml` changed since `bab3581`.

Resources were read with `pefile` (`uv run --with pefile`, not added to the project), because
`7z` misreads the resource tree of ARM64 executables. Icon data were compared byte by byte with
the images in `assets/brand/portunix.ico`.

### R1: configuration valid

`goreleaser check` — **PASS**, 1 configuration file validated.

### R2: full snapshot release

- **When** `goreleaser release --snapshot --clean --skip=publish,sign` runs with `dist`
  redirected to a temporary directory
- **Then** it succeeds and produces the Windows archives

Result: **PASS** — exit 0 in 13 s; `portunix_2.7.1-snapshot_windows_amd64.zip` and
`…_windows_arm64.zip` contain `portunix.exe`, all helpers and the install scripts.

### R3: `portunix.exe` resources in the Windows archives (AC2, AC4, AC5)

| Archive | Resource types | Icons equal to `portunix.ico` | Version info |
| ------- | -------------- | ----------------------------- | ------------ |
| `windows_amd64` | ICON ×8, GROUP_ICON 2 + 32512, VERSION, MANIFEST | 16, 32, 48, 256 (×2 each) | 2.7.0, Portunix, copyright unchanged |
| `windows_arm64` | same tree | **none** | **not readable** |

Result: **FAIL** for arm64, PASS for amd64.

Cause: `portunix.syso` is an **x86-64 COFF** object (goversioninfo default `-64`). Linked into
an ARM64 image, its relocations are not applied: every resource data entry of the arm64
`portunix.exe` points to the start of the `.rsrc` section (RVA `0x2f65000`, data begin with
`00000000…`) instead of to its own data. Icon, version info **and the manifest** are therefore
garbage on Windows ARM64.

Impact: before `a514efc` the arm64 `portunix.exe` had no resources at all; now it has a
resource tree with a corrupted manifest. Windows may refuse to start an executable whose
manifest does not parse (side-by-side configuration error). This could not be run here, but it
makes `a514efc` alone a possible regression for Windows ARM64.

Verified fix (temporary worktree on `a514efc`, removed after the run): replace `portunix.syso`
with per-architecture files using Go's file name build constraints:

```bash
goversioninfo -64      -o portunix_windows_amd64.syso versioninfo.json
goversioninfo -64 -arm -o portunix_windows_arm64.syso versioninfo.json
```

GoReleaser snapshot build of the `portunix` id then gives, for both `windows_amd64` and
`windows_arm64`, icons byte-identical to `portunix.ico` (16/32/48/256) and version info 2.7.0;
the Linux binary runs. The `_windows_<arch>` suffix also keeps the objects out of non-Windows
builds.

### R4: Linux binary from the release

`portunix --version` from `portunix_2.7.1-snapshot_linux_amd64.tar.gz` — **PASS**
(`Portunix version 2.7.1-snapshot`); Linux and macOS binaries carry no Windows resources.

### R5: regression of the earlier cases

TC01–TC05 and TC07 are unaffected by `a514efc` (only `.goreleaser.yml` changed) — **PASS**
(not repeated).

### R6: `ptx-installer.exe` on ARM64

`ptx-installer.syso` is also x86-64 COFF. Its manifest entry in the arm64 `ptx-installer.exe`
points to the section start as well — in the snapshot and in the released
`dist/portunix_2.7.0_windows_arm64.zip`. Result: **FAIL**, pre-existing, see O2.

### Observations (Retest 1)

- **O2 (pre-existing, out of scope)**: the `asInvoker` manifest of `ptx-installer.exe` for
  Windows ARM64 is corrupted since it got a `.syso` (released 2.7.0 affected). Same fix
  pattern as R3; better tracked as a separate issue.

### Required Changes (Retest 1)

1. Generate `portunix_windows_amd64.syso` and `portunix_windows_arm64.syso` instead of
   `portunix.syso`, and update every producer and consumer: `make gen-resources`,
   `build-with-version.sh`, the `.syso` handling in `scripts/make-release.py`, `.gitignore` if
   it lists the file, and the Makefile comments (R3)

## Retest 2

Commit `6d5750f` replaces `portunix.syso` and `ptx-installer.syso` with
`<name>_windows_amd64.syso` + `<name>_windows_arm64.syso` (required change of Retest 1, plus
O2), generates both in `make gen-resources` and `build-with-version.sh` (`gen_syso`), and adds
`TestSysoPerArchitecture` to both manifest guard tests.

All scenarios ran in a clean `git worktree` of `6d5750f` in a temporary directory (removed after
the run), so untracked files of the working copy could not influence the result. PE resources
were read with `pefile`.

### S1: tracked files

`git ls-files '*.syso'` lists exactly the four per-architecture files; `git check-ignore` matches
none of them (the `.gitignore` exception for `ptx-installer` follows the new name) — **PASS**.

### S2: guard tests

`TestSysoPerArchitecture` and the manifest tests pass for the root package and for
`ptx-installer` (`-count=1`); `go vet .` clean, also for `GOOS=windows GOARCH=arm64` — **PASS**.

### S3: full snapshot release, all Windows executables (AC2, AC4, AC5)

`goreleaser release --snapshot --clean --skip=publish,sign` (temporary `dist`) — exit 0, five
archives. All 40 executables of both Windows archives were inspected:

| Executable | Resource entries | Entries at section start | Manifest | Icons equal to `portunix.ico` |
| ---------- | ---------------- | ------------------------ | -------- | ----------------------------- |
| `windows_amd64/portunix.exe` | 12 | 0 | XML, asInvoker | 16, 32, 48, 256 |
| `windows_arm64/portunix.exe` | 12 | 0 | XML, asInvoker | 16, 32, 48, 256 |
| `windows_amd64/ptx-installer.exe` | 2 | 0 | XML, asInvoker | — |
| `windows_arm64/ptx-installer.exe` | 2 | 0 | XML, asInvoker | — |

The other 36 helpers have no resources, as before. Version info of both `portunix.exe`:
2.7.0, Portunix, copyright unchanged — **PASS**.

### S4: platform builds without `.syso` renaming

`CGO_ENABLED=0 go build .` with the `.syso` files present for linux/amd64, linux/arm64,
windows/amd64, windows/arm64, darwin/amd64 and windows/386 — all succeed; the Linux binary
runs. The file name suffix keeps the objects out of foreign targets, so the renaming in
`scripts/make-release.py` is no longer needed (harmless) — **PASS**.

### S5: `make gen-resources` reproducible

Regenerated files are byte-identical to the committed ones — **PASS**.

### S6: `goversioninfo` missing

`GOVERSIONINFO=/nonexistent/goversioninfo make gen-resources` prints the warning
"keeping existing portunix_windows_*.syso", exits 0, files untouched — **PASS**.

### S7: `build-with-version.sh` (no version argument)

Exit 0; regenerates both `portunix` and both `ptx-installer` objects; `portunix` objects
byte-identical; no tracked file changed — **PASS**.

### S8: `build-with-version.sh` with a generation failure

`IconPath` pointed to a missing file: the script stops with exit 1 and
"portunix .syso generation failed!" — **PASS**.

### S9: stale `portunix.syso` left in the tree

With an extra `portunix.syso`, the Windows build fails (`too many .rsrc sections`) and
`TestSysoPerArchitecture` fails with "portunix.syso must not exist …" — **PASS** (the failure
is loud and the test names the cause), see O3.

### S10: earlier cases

TC01–TC05, TC07 and R1–R4 cover files not changed by `6d5750f` or are repeated by S3–S5
— **PASS**.

### Observations (Retest 2)

- **O3 (informative)**: a stale `portunix.syso` (for example from an old local
  `versioninfo.cmd` run) breaks the Windows build with the linker message
  `too many .rsrc sections`, which does not name the file; `TestSysoPerArchitecture` does.
  Git removes the tracked file on checkout, so only locally regenerated leftovers are affected.

## Final Decision

**STATUS**: CONDITIONAL

All criteria that can be verified on Linux pass (AC2–AC5, R3, O2). Remaining condition:

- TC08 / AC1 — visual check on Windows: Explorer small/details/large/extra large views,
  taskbar while `portunix.exe` runs, file properties → Details; include the contrast of the
  black icon on a dark taskbar (O1)

**Approval for merge**: YES after TC08 passes on Windows
**Date**: 2026-10-27
**Tester signature**: Claude Code (tester role)

## Merge Note

**Date**: 2026-10-27

`feature/250-replace-windows-icon-with-gear` was merged into `main` on the decision of the
repository owner (zdendaku), as an exception to the "no merge without PASS" rule. Retest 2
(`6d5750f`) passed all criteria that can be verified on Linux; the Windows visual check (TC08 /
AC1, including the contrast of the black icon on a dark taskbar, O1) is still outstanding.
Issue #250 stays open until it passes; any defect found there is fixed on a new branch.

Update: on the repository owner's decision, issue #250 was closed as implemented on the same
day, before TC08 ran. TC08 remains outstanding; a defect found there is tracked as a new issue.
