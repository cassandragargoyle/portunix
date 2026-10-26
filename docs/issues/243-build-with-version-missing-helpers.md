# Issue #243: build-with-version.sh does not build ptx-credential and ptx-python

**Status**: Open
**Priority**: Low
**Type**: Bug
**Component**: portunix (build tooling), ptx-credential, ptx-python
**Created**: 2026-10-26
**GitHub**: #243

## Summary

`build-with-version.sh` builds the main binary and 17 helper binaries with the
embedded version, but skips `ptx-credential` and `ptx-python`. A versioned local
deploy (`build-with-version.sh vX.Y.Z` followed by `make deploy-local`) therefore
installs these two helpers without a version — they report `version dev`.

## Problem Description

### Current Situation

- `build-with-version.sh` contains one build block per helper
  (`ptx-container`, `ptx-mcp`, `ptx-virt`, …, `ptx-wizard`); there is no block
  for `ptx-credential` and `ptx-python`, and the final version check does not
  list them either
- `make build` builds both helpers (`Makefile`, helper build target), but without
  `-ldflags "-X main.version=..."`, so the binaries in the repository root carry
  the default `var version = "dev"`
- `make deploy-local` (`scripts/deploy-local.py`) copies every `ptx-*` binary
  from the repository root, including these two unversioned ones

Observed after deploying `v2.7.0+dev.1`:

```text
$ portunix version
Portunix version v2.7.0+dev.1
$ ptx-credential --version
ptx-credential version dev
$ ptx-python --version
ptx-python version dev
```

### Not Affected

Release packages built by `scripts/make-release.py` are correct: `.goreleaser.yml`
builds both helpers with `-X main.version={{ .Version }}`, and
`dist/ptx-credential_*/ptx-credential --version` reports the release version.

### Impact

- Locally deployed helpers cannot be identified by version, which makes
  debugging and acceptance testing harder
- Stale binaries of these helpers may be deployed after a source change,
  because the versioned build does not rebuild them

## Acceptance Criteria

1. `build-with-version.sh vX.Y.Z` builds `ptx-credential` and `ptx-python` with
   the same `LDFLAGS_COMMON` as the other helpers
2. The version check at the end of the script lists both helpers
3. After `build-with-version.sh vX.Y.Z` and `make deploy-local`,
   `ptx-credential --version` and `ptx-python --version` report `vX.Y.Z`
4. Windows builds (`.exe` suffix) are covered the same way as other helpers
5. Optional: the script derives the helper list from `src/helpers/` or a shared
   list, so a newly added helper cannot be missed again

## Testing

- Run `bash ./build-with-version.sh v0.0.0+dev.1` and check that the version
  check prints `v0.0.0+dev.1` for every helper in `src/helpers/` that produces
  a binary
- Run `./ptx-credential --version` and `./ptx-python --version` in the
  repository root
- Regression: `make build` and `python3 scripts/make-release.py vX.Y.Z-SNAPSHOT`
  still produce all binaries
