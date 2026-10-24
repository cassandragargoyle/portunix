# Acceptance Protocol - Issue #231 (portunix unzip Does Not Extract Anything)

**Issue**: portunix unzip does not extract anything
**Branch**: `fix/231-unzip-command-does-not-extract` (round 1: commit `b0d2ff8`, round 2: commit `3c9528a`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host)

The command extracts archives only into the given directories and installs no software, so
all scenarios ran on the host in a session-local scratch directory, which was removed after
the run. No container was needed.

## Scope Note

The branch changes four things:

1. `portunix unzip` calls `app.UnzipFile(path, destinationPath)` directly instead of
   rebuilding `--path=<file>` arguments
2. A missing `--path` is reported as an error with exit 1 (it used to exit 0)
3. `unzip()` opens the archive before it creates the destination directory
4. `app.Unzip`, `app.ProcessArgumentsUnzip`, `isParameter` and `app.ProcessArguments` are
   removed

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Windows — the change is platform-independent, not executed
- Plugin installation from a zip asset (`extractZip` → `app.UnzipFile`) end to end — covered
  only by the `app` unit tests

## Test Summary

| Round | Commit | Scenarios | Passed | Failed | Skipped | Decision |
| ----- | ------ | --------- | ------ | ------ | ------- | -------- |
| 1 | `b0d2ff8` | 12 | 11 | 1 (TC11, D1) | 2 | FAIL |
| 2 | `3c9528a` | 12 | 12 | 0 | 2 | PASS |

- Skipped in both rounds: Windows, plugin zip install end to end
- Defects found: 1 (D1, blocking, fixed in `3c9528a`)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | New `TestRunUnzip_*` tests, `portunix.ai/app` archive tests | host |
| Unit | Mutation check: `runUnzip` passes the wrong arguments, the new tests must fail | host |
| Unit | Root package `portunix.ai/portunix` and the full `portunix.ai/cmd` suite against `main` | host, worktree of `main` |
| Static | `go vet`, `gofmt`, `make build` | host |
| E2E | `portunix unzip` with regular, malicious, invalid and missing archives | host, scratch directory |

## Test Cases

Test archives: `a.zip` (`a.txt`, `sub/b.txt`, `emptydir/`), `evil.zip` (`ok.txt`,
`../escaped.txt`), `abs.zip` (`/tmp/abs-escape-231.txt`), `a=b.zip` in a directory with a
space, and `notzip.zip` (plain text). Commands ran from a sibling directory `run/`.

### TC01: Extraction next to the archive (AC1)

- **When** `portunix unzip --path ../arch/a.zip`
- **Then** the entries are extracted into `arch/`, exit 0

### TC02: Relative destination (AC2)

- **When** `portunix unzip --path ../arch/a.zip --destinationpath out`
- **Then** the entries are extracted into `run/out`, exit 0

### TC03: Absolute destination (AC2)

- **When** `--destinationpath <absolute path>`
- **Then** the entries are extracted there, exit 0

### TC04: Existing destination

- **When** TC02 is repeated
- **Then** files are overwritten, exit 0

### TC05: `=` and spaces in paths

- **When** `--path "../sp ace/a=b.zip" --destinationpath "dest with=eq"`
- **Then** `dest with=eq/eq.txt` is extracted, exit 0 (the old parser split on `=`)

### TC06: Missing archive (AC3, AC4)

- **When** `--path ../arch/missing.zip --destinationpath out-missing`
- **Then** `Error: open …: no such file or directory`, exit 1, `out-missing` is not created

### TC07: ZipSlip entries (AC3)

- **When** `evil.zip` (`../escaped.txt`) and `abs.zip` (absolute entry) are extracted
- **Then** `Error: illegal file path: …`, exit 1, nothing is written outside the destination

### TC08: Invalid archive and missing `--path`

- **When** `notzip.zip`, no arguments, and only `--destinationpath`
- **Then** `zip: not a valid zip file` and `please specify the path to the ZIP file with --path`,
  exit 1 in all cases

### TC09: No flag-named directories (AC4)

- **When** the scratch tree is listed after TC01–TC08
- **Then** no entry name starts with `-`

### TC10: Unit tests and mutation check (AC5)

- **When** `go test -count=1 portunix.ai/cmd -run RunUnzip` and `go test -count=1
  portunix.ai/app`, then the same with `runUnzip` mutated to
  `app.UnzipFile(destinationPath, "")`
- **Then** all pass on the branch; with the mutation all three `TestRunUnzip_*` tests fail

### TC11: Regression of the remaining test suites

- **When** `go vet .` / `go test .` in the repository root and `go test -count=1
  portunix.ai/cmd`, compared with `main`
- **Then** no new compile errors or test failures

### TC12: Static checks

- **When** `go vet portunix.ai/app portunix.ai/cmd`, `gofmt -l src/app src/cmd`, `make build`
- **Then** no findings, and the build succeeds

## Test Results — Round 1 (`b0d2ff8`)

### E2E — Host

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC01 | ✅ PASS | `Extracting: <S>/arch/a.txt`, `…/sub/b.txt`, `…/emptydir`; rc=0 |
| TC02 | ✅ PASS | `Extracting: out/a.txt` …; `run/out/a.txt`, `run/out/sub/b.txt` present; rc=0 |
| TC03 | ✅ PASS | `Extracting: <S>/absout/a.txt` …; rc=0 |
| TC04 | ✅ PASS | Second run into `out` rc=0 |
| TC05 | ✅ PASS | `Extracting: dest with=eq/eq.txt`; rc=0 |
| TC06 | ✅ PASS | `Error: open ../arch/missing.zip: no such file or directory`; rc=1; no `out-missing` |
| TC07 | ✅ PASS | `Error: illegal file path: ../escaped.txt` rc=1; `Error: illegal file path: /tmp/abs-escape-231.txt` rc=1; neither `escaped.txt` nor `/tmp/abs-escape-231.txt` exists |
| TC08 | ✅ PASS | `Error: zip: not a valid zip file` rc=1; `Error: please specify the path to the ZIP file with --path` rc=1 (twice) |
| TC09 | ✅ PASS | `find` lists `absout`, `run/{absevil,dest with=eq,evilout,out}`; no entry starting with `-` |

### Host — Tests and Static Checks

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC10 | ✅ PASS | `portunix.ai/app` ok, `TestRunUnzip_*` ok; mutation: `--- FAIL` for all three tests; file restored with `git checkout` |
| TC11 | ❌ FAIL | Root package does not compile, see D1. `portunix.ai/cmd`: 6 container tests fail identically on `main` and on the branch (pre-existing, #225) |
| TC12 | ✅ PASS | `go vet` and `gofmt -l` clean; `make build` succeeds |

### Skipped

- Windows execution
- Plugin installation from a zip asset end to end

## Defects

### D1: Root `main_test.go` still calls the removed functions (blocking, fixed)

`main_test.go` in the repository root has `TestProcessArgumentsUnzip` and `TestUnzip`, which
call `app.ProcessArgumentsUnzip` and `app.Unzip`. The branch removed both functions, so the
`portunix.ai/portunix` test package no longer compiles:

```text
./main_test.go:21:19: undefined: app.ProcessArgumentsUnzip
./main_test.go:29:18: undefined: app.ProcessArgumentsUnzip
./main_test.go:74:12: undefined: app.Unzip
FAIL	portunix.ai/portunix [build failed]
```

On `main`, `go vet .` is clean. Because the package does not build, the unrelated
`TestMainWithoutArguments`, `TestMCPServeCommand` and `TestMCPServerCommandRemoved` cannot
run either. The search for remaining callers was limited to `src/`.

**Required fix**: remove `TestProcessArgumentsUnzip` and `TestUnzip` (and the
`createTestZip` helper, plus the `archive/zip` and `path/filepath` imports if nothing else
uses them) from `main_test.go`. The new `src/cmd/unzip_test.go` covers the same behaviour.

**Fix**: commit `3c9528a` removes both tests, `createTestZip` and the unused imports, and
leaves a comment pointing to `src/cmd/unzip_test.go` and `src/app/archive_test.go`.

## Test Results — Round 2 (`3c9528a`)

`b0d2ff8..3c9528a` changes only `main_test.go` (3 insertions, 98 deletions), so the product
code is the same as in round 1. TC11 was re-run in full, the unit and static checks were
repeated, and a fresh build got an E2E smoke run.

| Case | Result | Evidence |
| ---- | ------ | -------- |
| TC11 | ✅ PASS | `go vet .` clean; root package compiles. Root tests: branch `TestMainWithoutArguments` FAIL, `TestMCPServeCommand`, `TestMCPServerCommandRemoved`, `TestManifestRequestsAsInvoker` PASS; `main` additionally runs the removed `TestProcessArgumentsUnzip`/`TestUnzip` (PASS) and fails `TestMainWithoutArguments` and `TestMCPServeCommand`. `portunix.ai/cmd`: the same 6 container tests fail on `main` and on the branch. No new failures |
| TC10 | ✅ PASS | `portunix.ai/app` ok; `TestRunUnzip_NextToArchive`, `_Destination`, `_Errors` PASS |
| TC12 | ✅ PASS | `go vet . portunix.ai/app portunix.ai/cmd` clean; `gofmt -l` clean; `make build` succeeds |
| Smoke (TC01, TC02, TC06, TC07) | ✅ PASS | next to archive rc=0; `--destinationpath out` rc=0; missing archive rc=1 and no `nm` directory; `../escaped.txt` rejected rc=1 and not written |
| Leftover references | ✅ PASS | `grep` over the whole repository: `ProcessArgumentsUnzip`, `app.Unzip(`, `ProcessArguments(`, `isParameter(` only in the explanatory comment in `main_test.go` |

`TestMCPServeCommand` fails only in the `main` worktree, where the helper binaries were not
built; it is not related to #231.

## Observations (Not Blocking)

- An archive that contains a ZipSlip entry after a valid one is extracted partially:
  `evilout/ok.txt` is written before the error. This behaviour existed before the branch and
  is outside the scope of #231
- The 6 failing container tests in `portunix.ai/cmd` are pre-existing test debt (#225)
- `TestMainWithoutArguments` fails on `main` and on the branch: it expects cobra's
  `Available Commands:` and `Flags:`, but Portunix prints its own help. Not related to #231
  and not tracked by an issue yet

## Final Decision

**STATUS**: PASS (round 2, commit `3c9528a`; round 1 FAIL because of D1)

**Approval for merge**: YES
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role, Linux) on behalf of zdendaku
