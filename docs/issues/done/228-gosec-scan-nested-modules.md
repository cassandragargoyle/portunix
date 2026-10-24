# Issue #228: Extend gosec Security Scan to modules with their own go.mod

**Status**: Implemented
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix
**Created**: 2026-10-07
**Closed**: 2026-10-11
**GitHub**: #228
**Related**: #224

## Summary

The GitHub Actions **Security Scan** job runs `gosec -exclude=G204,G304 ./...` from the
repository root. `./...` stops at nested module boundaries, so every directory with its
own `go.mod` is skipped. Issue #224 cleaned up the root module only.

## Problem Description

Modules that are not scanned today:

- `src/app`, `src/app/github`, `src/app/sandbox`, `src/app/virt`, `src/app/wizard`
- `src/cmd`, `src/parser`
- `src/helpers/ptx-ansible`, `ptx-container`, `ptx-credential`, `ptx-database`,
  `ptx-github`, `ptx-installer`, `ptx-mcp`, `ptx-plugin-registry`, `ptx-prompting`,
  `ptx-proxmox`, `ptx-python`, `ptx-specpm`, `ptx-ssh`, `ptx-trace`, `ptx-wizard`

These contain most of the code that downloads, extracts and executes things, so the scan
currently misses the parts where findings matter most. The number of findings is unknown.

## Proposed Approach

1. Run gosec per module (loop over every `go.mod`) to get a baseline count per rule.
2. Fix findings following [docs/contributing/SECURITY-SCAN.md](../contributing/SECURITY-SCAN.md):
   HIGH fixed in code (with the documented G101/G402 exceptions), intentional patterns
   annotated with `#nosec <rule> -- reason`.
3. Extend the CI job, `make security` and `scripts/github-01-preflight-check.sh` to scan
   all modules with the same invocation.
4. Update the **Scope** section of `SECURITY-SCAN.md`.

Splitting the cleanup per module (one PR per helper) is acceptable.

## Acceptance Criteria

1. gosec runs on every Go module in the repository in CI, `make security` and the preflight check
2. All three places use the same invocation and exit 0
3. HIGH findings are fixed in code; only the two cases listed in `SECURITY-SCAN.md` under
   *HIGH Findings That May Be Annotated* (G101 false positive, G402 on explicit user request)
   may carry a `#nosec`
4. Every `#nosec` names the rule and a reason

## Testing

- Local: `./scripts/github-01-preflight-check.sh` reports per-module results
- CI: Security Scan job green on the next release PR
