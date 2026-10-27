# Issue #246: CI: Windows, Coverage Report and Performance jobs fail on pre-existing test debt

**Status**: Open
**Priority**: High
**Type**: Bug
**Component**: portunix (CI, tests)
**Created**: 2026-10-26
**GitHub**: #246
**Related**: #225, #232, #245

## Summary

The release PR #245 (v2.7.0) is the first one where the Unit Tests job passes. The jobs that
depend on it therefore ran for the first time, and three of them fail on problems that existed
long before v2.7.0: Cross Platform Tests (windows-latest), Coverage Report and Performance Tests.
Quality Gate fails as a consequence.

## Problem Description

### Cross Platform Tests (windows-latest)

| Test | Cause |
| ---- | ----- |
| `TestMainWithoutArguments`, `TestMCPServeCommand`, `TestMCPServerCommandRemoved` (`main_test.go`) | The test builds `portunix_test` and runs `./portunix_test`; on Windows `exec` needs the `.exe` suffix |
| `TestCreateFileWriter`, `TestCreateMCPLogger`, `TestFactoryWithMultipleOutputs` (`pkg/logging`) | `Factory` opens log files and never closes them; Windows cannot delete an open file, so the `t.TempDir()` cleanup fails |
| `TestNewSyncCache` (`ptx-pft`) | Expected path is a hard-coded unix string (`/tmp/test-project/.pft-cache.json`) |
| `TestValidateVolumePath_ValidPaths_Success/Valid_absolute_path` (`test/unit`) | Uses `/tmp`, which is not an absolute path on Windows |

### Coverage Report

The job builds only `go build -o portunix ./`, without the helper binaries, and runs
`go test ./...` without the `unit` build tag. Tests that go through the dispatcher
(`mcp serve --help` → ptx-mcp, `install --help` → ptx-installer) fail because the helpers are
missing: `TestMCPServeCommand`, `TestInstallCommandHelpFlag`, `TestInstallCommandShortHelpFlag`.

### Performance Tests

`go test -bench=. ./... > benchmark-results.txt` also runs every regular test (no `-run`
filter), without binaries and without the `unit` tag, and hides the output in the artifact file.

### Cache noise

Every job runs `actions/cache@v3` for `~/go/pkg/mod` although `actions/setup-go@v4` already
restores the same cache; the second restore reports tens of thousands of
`tar: ... Cannot open: File exists` errors.

## Acceptance Criteria

1. Cross Platform Tests pass on ubuntu-latest, windows-latest and macos-latest
2. Coverage Report and Performance Tests pass on a pull request
3. Quality Gate passes on the release PR
4. `Factory.Close()` in `pkg/logging` releases the log files opened by its loggers (also those
   of `CreateMCPLogger` / `CreateTestLogger`)
5. Root and `test/unit` tests that need helper binaries skip with a clear message when the
   helpers are not built, instead of failing
6. Benchmarks run with `-run='^$'` and their output is visible in the job log
7. No duplicate Go module cache restore

## Testing

- Locally: `make build`, then the CI commands of the Unit Tests, Coverage Report and Performance
  jobs (`go test -tags=unit -race ./...`, with `-coverprofile`, and
  `go test -tags=unit -run='^$' -bench=. ./...`)
- Locally without helper binaries: the dispatcher tests skip instead of failing
- `GOOS=windows go vet` for the changed packages
- On GitHub: all jobs of the release PR, Windows included
