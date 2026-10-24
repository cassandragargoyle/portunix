# Acceptance Protocol - Issue #226 (ptx-installer Shared extractTo Directory Removal)

**Issue**: ptx-installer offers to delete shared extractTo directory (/usr/local/bin, /usr/local) on reinstall
**Branch**: `fix/226-installer-shared-extract-dir` (commit `ad612b3`; branch head `c314306` adds only the #227 issue file)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-10-07
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-34 (host) — unit tests, `go vet`, `gofmt`, build only; nothing installed
- Ubuntu 24.04.3 LTS (container `ubuntu:24.04`, Podman 5.7.0 via `portunix container`, root) — all installation scenarios
- Debian GNU/Linux 12 bookworm (container `debian:bookworm`, Podman 5.7.0 via `portunix container`, root) — all installation scenarios

## Scope Note

The branch changes how `ptx-installer` treats the extraction directory of `tar.gz` / `zip`
packages: well-known shared directories are never removed (not even with `--force`), only a
top-level directory created by the archive is offered for removal, and tar extraction
replaces existing files and symlinks instead of writing over them. It also removes a
non-existent Hugo URL.

**Not covered**: Windows. The Windows list of shared directories and the `openssh` package
(`extractTo: C:/Program Files`) were only compiled (`GOOS=windows go vet`), not executed.
See Recommendations.

## Test Summary

- Total test scenarios: 45 (14 unit, 2 baseline on `main`, 13 container scenarios × 2
  distributions, 3 package-owned directory scenarios on `debian:bookworm`)
- Passed: 45 (TC08 on `debian:bookworm` passed on repetition, see O4)
- Failed: 0
- Skipped: 0
- Open defects in the scope of #226: 0
- Observations outside the scope: 4 (O1–O4)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `gofmt -s`, `go vet` (linux, windows), `go build` (darwin), `make build` | host |
| Unit | shared-directory guard, `checkExistingDirectory`, `prepareSharedDirectory`, tar overwrite, symlink replace, `ListTopLevelEntries` (`go test -race`) | host |
| Baseline | reproduction of the defect on a binary built from `main` (`b518142`) | container `ubuntu:24.04` |
| E2E | hugo (`/usr/local/bin`), go (`/usr/local`, archive root `go/`), protoc (`/usr/local`, archive roots `bin/`, `include/`), maven (`/opt/maven`, package-owned) | containers `ubuntu:24.04`, `debian:bookworm` |
| Failure injection | longer old binary, running binary, cancel at the prompt, non-root user | containers |

Container setup for every run (fresh container, removed afterwards):

```bash
portunix container run --rm -v "$PWD:/ptx:ro,Z" docker.io/library/ubuntu:24.04 bash /t226.sh
```

Inside: `ca-certificates` installed, binaries copied to `/opt/ptx`, marker files created in
`/usr/local/bin/other-tool`, `/usr/local/bin/act`, `/usr/local/unrelated`,
`/usr/local/share/keepdir/f`.

## Test Results

### Static Checks and Unit Tests

| Check | Result |
| ----- | ------ |
| `gofmt -s -l` on changed packages | PASS (no output) |
| `go vet ./src/pkg/archive/`, `go vet ./...` in `ptx-installer` | PASS |
| `GOOS=windows go vet ./engine/`, `GOOS=darwin go build ./...` | PASS |
| `go test -race ./src/pkg/archive/` (5 tests) | PASS |
| `go test -race ./...` in `ptx-installer` (9 new + all existing) | PASS |
| `make build` | PASS |

The two tar tests fail against the `main` version of `extract.go` (content
`"new-binary-content-that-is-longer"` instead of `"new"`; second extraction fails with
`symlink …: file exists`), so they do detect the original behaviour.

### Baseline on main (defect reproduction)

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| B1 | Given marker files in `/usr/local/bin`, when `portunix install hugo --variant extended` and answer `n`, then the installer offers to remove `/usr/local/bin` | Defect reproduced: "Do you want to remove the existing directory and reinstall?" for `/usr/local/bin` |
| B2 | Same, when `--force`, then `/usr/local/bin` is removed | Defect reproduced: `act` and `other-tool` deleted, directory contains only `LICENSE README.md hugo` |

### Functional Tests (fix branch)

Identical results on `ubuntu:24.04` and `debian:bookworm` unless noted.

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| TC01 | Given marker files in `/usr/local/bin`, when hugo is installed for the first time, then no prompt, markers kept, `hugo version` = 0.167 | PASS |
| TC02 | Given hugo installed, when reinstalled with `--force`, then `/usr/local/bin` is not removed and markers are kept | PASS |
| TC03 | Given hugo installed, when reinstalled with `y` on stdin, then no removal prompt appears and markers are kept | PASS |
| TC04 | Given `/usr/local/bin/hugo` replaced by a file 500 kB longer, when reinstalled, then the file has exactly the original size and runs | PASS (62382560 bytes) |
| TC05 | Given `hugo server` running from `/usr/local/bin/hugo`, when reinstalled with `--force`, then extraction succeeds (no "text file busy"), the server keeps running, new binary runs | PASS |
| TC06 | Given markers in `/usr/local`, when go is installed for the first time, then no prompt, `/usr/local/go/bin/go` works, markers kept | PASS |
| TC07 | Given `/usr/local/go/STALE-FILE`, when go is reinstalled and the answer is `n`, then the prompt names only `/usr/local/go`, installation is cancelled, nothing changes | PASS |
| TC08 | Same, when the answer is `y`, then only `/usr/local/go` is removed (stale file gone), hugo and markers kept | PASS (debian: see O4) |
| TC09 | Same, when `--force`, then only `/usr/local/go` is removed, hugo and markers kept, go works | PASS |
| TC10 | Given hugo, go and markers in `/usr/local`, when protoc (archive roots `bin/`, `include/`) is installed with `--force`, then `/usr/local/bin` is not removed | PASS |
| TC11 | Given protoc installed, when reinstalled with `--force`, then `/usr/local/bin` content is kept, `protoc --version` = 25.1 | PASS |
| FINAL | SHA-256 over all marker files before TC01 and after the last scenario | PASS (unchanged) |

### Regression Tests

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| TC12 | When maven `latest` is installed, then it extracts to the package-owned `/opt/maven` without a shared-directory message | PASS (debian) |
| TC13 | Given `/opt/maven/STALE-FILE`, when reinstalled and the answer is `n`, then the prompt names `/opt/maven`, installation is cancelled, file kept | PASS (debian) |
| TC14 | Same, when `--force`, then `/opt/maven` is removed and recreated (stale file gone), `/opt` and its other content kept | PASS (debian) |
| — | All existing `ptx-installer` unit tests | PASS |

TC12–TC14 were executed with the default variant in the first pass of both containers; the
default is `apt`, which does not exercise the archive path, so that pass is not counted. The
valid run (`--variant latest`) was done on `debian:bookworm`.

### Failure Injection

| ID | Scenario | Result |
| -- | -------- | ------ |
| TC04 | Longer previous binary (truncate check) | PASS |
| TC05 | Binary in use during reinstall | PASS |
| TC07 / TC13 | User cancels at the prompt | PASS — nothing modified |
| TC15 | Non-root user reinstalls hugo with `--force` | PASS for #226 — nothing removed, clear error (see O3) |

### Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | Reinstall never deletes files that are not part of the archive, with and without `--force` | PASS — executed for hugo, go, protoc; act, actionlint, caddy, ninja, terraform and hugo-extended use the same code path with the same `extractTo` and were not installed individually |
| 2 | Reinstall in a shared directory upgrades the binary in place | PASS (TC02–TC05) |
| 3 | Per-package directories keep the "remove and reinstall" behaviour | PASS (TC12–TC14) |
| 4 | Unit tests cover the shared-path guard including `--force` | PASS |
| 5 | `hugo.json` contains no URL that returns 404 | PASS — all 7 URLs of `hugo.json` and `hugo-extended.json` return 200 |

## Coverage

- Unit: `isSharedDirectory` (configured list, filesystem root, symlink, default Unix list),
  `checkExistingDirectory` (shared with and without force, owned with force),
  `prepareSharedDirectory` (archive directory removed, shared subdirectory kept),
  in-place upgrade, tar file / symlink replacement, `ListTopLevelEntries` (tar.gz, zip, unsupported)
- Not covered by unit tests: the interactive prompt path of `checkExistingDirectory`
  (reads `os.Stdin`; covered by E2E TC07, TC08, TC13), the Windows directory list
- E2E: 3 of 9 affected Linux packages, both shared directories (`/usr/local/bin`, `/usr/local`)

## Observations (outside the scope of #226)

- **O1** — `portunix install hugo|go|protoc` exits with "Installation failed" although the
  package is installed and works: `postInstall` commands use unsubstituted placeholders
  (`${sudo_prefix}`, `${actual_extract_to}`, `${downloaded_file}`) or a hard-coded `sudo`
  that is missing in the container. Same behaviour on `main`. Tracked as **#227**.
- **O2** — the Hugo tarball also drops `LICENSE` and `README.md` into `/usr/local/bin`.
  Same behaviour on `main`.
- **O3** — as a non-root user the reinstall fails with
  `extraction failed: remove /usr/local/bin/hugo: permission denied`; there is no fallback
  to the user directory because the fallback only triggers when the directory cannot be
  created. Nothing is removed. Not a regression of the safety fix, but the error message is
  new (previously the failure came from opening the file).
- **O4** — in the first `debian:bookworm` pass, TC08 produced no installer output and
  changed nothing (stale file and all markers still present). The step was repeated three
  times in a fresh `debian:bookworm` container with full output and passed every time
  (exit path identical to Ubuntu). The first-pass output was filtered and is not available;
  the most likely cause is a failed download before the directory handling. The failure
  mode was safe — no file was removed.

## CI Notes

- New tests are plain unit tests without network or container access and run under
  `go test ./...` in both modules
- `ptx-installer` is a separate Go module, so the root-module CI run (`go test ./...`)
  executes only the `src/pkg/archive` tests; the engine tests need a run inside
  `src/helpers/ptx-installer` (same gap as described in #194 / #225)
- Suggested follow-up: a manifest test that flags any `extractTo` equal to a shared directory
  whose archive has no top-level directory, to make such packages visible

## Recommendations

1. **Windows verification** (Windows tester, Windows Sandbox): reinstall `openssh`
   (`extractTo: C:/Program Files`) and one `C:/Program Files/<name>` package with and
   without `--force`; confirm `C:\Program Files` is never removed. Add a unit test for the
   Windows shared-directory list (runs only on Windows)
2. Fix #227 so that the affected installs stop reporting a false failure
3. Consider extracting into a package-owned directory with a symlink in `/usr/local/bin`
   (item 4 of the issue), which would also resolve O2 and O3

## Final Decision

**STATUS**: PASS (Linux)

**Approval for merge**: YES — the data-loss defect is fixed and verified on Linux; Windows
behaviour is compiled but not executed (Recommendation 1)
**Date**: 2026-10-07
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
