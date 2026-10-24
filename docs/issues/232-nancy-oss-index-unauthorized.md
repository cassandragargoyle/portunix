# Issue #232: make security and the Security Scan job fail in nancy: OSS Index returns 401

**Status**: Open
**Priority**: Medium
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-11
**GitHub**: #232
**Related**: #224, #228

## Summary

The dependency vulnerability scan runs `nancy sleuth` without credentials and OSS Index
answers `401 Unauthorized`. `make security` therefore fails after the gosec step has passed.
The Security Scan job in GitHub Actions runs the same command and is expected to fail the same
way. The defect was found while testing #228 (`docs/testing/acceptance-228.md`, defect D3).

## Problem Description

`make security` and the job `security` in `.github/workflows/test.yml` both run:

```bash
go list -json -deps ./... | nancy sleuth
```

Observed with nancy 1.0.51:

```text
Error: An error occurred: [401 Unauthorized] error accessing OSS Index
```

OSS Index requires an account; nancy accepts the credentials through `--username` and
`--token`. Neither the `Makefile` nor the workflow provides them.

Two side notes:

- The nancy step has not been observed in CI yet. Until #224 and #228 the job stopped earlier,
  in the gosec step
- `go list -deps ./...` covers only the root module, the same gap #228 closed for gosec.
  Helper modules with their own `go.mod` and their dependencies are not checked

## Proposed Approach

Decide between two options:

| Option | Pros | Cons |
| ------ | ---- | ---- |
| Keep nancy, add OSS Index credentials as repository secrets and as local configuration | Smallest change | Every contributor needs an account, `make security` fails without one |
| Replace nancy with `govulncheck` | Official Go tool, no account, reports only vulnerabilities reachable from the code | Different report format, new tool to install |

Whichever tool is used should run per module, the way `scripts/gosec-scan.py` iterates over
modules.

## Acceptance Criteria

1. `make security` exits 0 on a clean tree without manual credential setup, or documents the
   required setup and fails with a clear message when it is missing
2. The Security Scan job in GitHub Actions is green
3. Dependencies of every Go module are checked, not only those of the root module
4. `docs/contributing/SECURITY-SCAN.md` describes the dependency scan

## Testing

- `make security` locally, with and without the required configuration
- Security Scan job on a pull request
- A module with a deliberately old dependency that has a known vulnerability is reported
