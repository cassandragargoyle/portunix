# Issue #231: portunix unzip does not extract anything

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-11
**Closed**: 2026-10-24
**GitHub**: #231
**Related**: #228

## Summary

`portunix unzip --path <file> [--destinationpath <dir>]` never extracts the archive. The
arguments are handed to a home-grown parser in a form it does not understand, so the archive
path is lost before extraction starts. The defect was found while testing #228
(`docs/testing/acceptance-228.md`, defect D2).

## Problem Description

`src/cmd/unzip.go` reads the cobra flags and rebuilds them as strings:

```go
arguments = append(arguments, fmt.Sprintf("--path=%s", path))
arguments = append(arguments, fmt.Sprintf("--destinationpath=%s", destinationPath))
app.Unzip(arguments)
```

`app.ProcessArguments` in `src/app/tools.go` accepts `path=<value>` or `--path <value>`, but
not `--path=<value>`. The string `--path=a.zip` matches the `--path` prefix branch and the
parser takes the **next** argument as the value:

| Command line | Parsed `path` | Result |
| ------------ | ------------- | ------ |
| `--path a.zip` | empty | `open : no such file or directory` |
| `--path a.zip --destinationpath out` | `--destinationpath=out` | `open --destinationpath=out: no such file or directory` |

In the second case the destination is derived from the bogus path, so a stray directory tree
named `--destinationpath=...` is created under the current directory.

Until #228 the command discarded the error and exited 0 without output, which hid the defect.
Since #228 it prints the error and exits 1.

The hardened extraction itself (`app.UnzipFile`: ZipSlip check, size cap) works and is covered
by unit tests in `src/app/archive_test.go`; plugin installation from a zip archive uses it
directly and is not affected.

## Proposed Approach

1. Call `app.UnzipFile(path, destinationPath)` from the cobra command instead of rebuilding an
   argument list
2. Check whether `app.Unzip` and `app.ProcessArgumentsUnzip` still have a caller and remove
   them if not
3. Review the other callers of `app.ProcessArguments` for the same `--name=value` form

## Acceptance Criteria

1. `portunix unzip --path a.zip` extracts next to the archive and exits 0
2. `portunix unzip --path a.zip --destinationpath out` extracts into `out` and exits 0
3. A missing archive and an archive with an entry outside the destination report an error and
   exit 1
4. No directory named after a flag is created
5. The command has an automated test covering criteria 1 to 3

## Testing

- Unit or integration test of the `unzip` command with a regular archive, a missing file and
  an archive containing `../escaped.txt`
- Manual run in a temporary directory, with a relative and an absolute destination
