# Issue #224: CI Security Scan fails: clean up gosec findings (4 HIGH in ptx-pft)

**Status**: Implemented
**Priority**: High
**Type**: Bug
**Component**: portunix, ptx-pft
**Created**: 2026-10-07
**Closed**: 2026-10-11
**GitHub**: #224
**Related**: #225

## Summary

The GitHub Actions **Security Scan** job runs `gosec -fmt json -out gosec-report.json ./...`
and fails whenever gosec reports any finding. It has failed on every release PR so far
(PR #34 for v2.5.1, PR #223 for v2.6.1). Release v2.6.1 was merged with the job red; the
finding set was identical on GitHub `main` before and after the merge (no regression).

## Problem Description

gosec (latest, Go 1.25) on the published tree reports **250 findings in 76 files**
(scanning only the root module — helper modules with their own `go.mod` are not covered
by `./...`):

| Rule | Severity | Count | Meaning |
| ---- | -------- | ----- | ------- |
| G204 | MEDIUM | 75 | Subprocess launched with variable |
| G301 | MEDIUM | 44 | Poor directory permissions (mkdir > 0750) |
| G304 | MEDIUM | 41 | File inclusion via variable path |
| G104 | LOW | 39 | Unhandled errors |
| G306 | MEDIUM | 38 | Poor file permissions (write > 0600) |
| G703 | HIGH | 3 | Path traversal via taint analysis |
| G110 | MEDIUM | 3 | Decompression bomb |
| G107 | MEDIUM | 2 | HTTP request with variable URL |
| G305 | MEDIUM | 2 | File traversal when extracting archive |
| G302 | MEDIUM | 2 | Poor file permissions (chmod) |
| G115 | HIGH | 1 | Integer overflow conversion |

HIGH findings:

- `src/helpers/ptx-pft/sync.go:460` — G703 path traversal
- `src/helpers/ptx-pft/sync.go:888` — G703 path traversal
- `src/helpers/ptx-pft/main.go:2873` — G703 path traversal
- `src/helpers/ptx-pft/cache.go:140` — G115 integer overflow (rune → uint32)

## Proposed Approach

1. Fix all HIGH findings (validate / clean paths before writing; bound the rune conversion).
2. Fix the real issues among MEDIUM findings (G110 decompression bomb, G305 archive
   extraction traversal, directory/file permissions where the data is user-private).
3. For intentional patterns (G204 — Portunix runs external tools by design; G304 — reading
   user-specified files), add a project gosec configuration excluding them with a written
   rationale, or `#nosec` annotations with a justification at each site.
4. Decide whether helper modules (`src/helpers/*` with own `go.mod`) should be scanned too.

## Acceptance Criteria

1. `gosec ./...` exits 0 on the published tree (same command as CI)
2. No HIGH finding is suppressed — all are fixed in code
3. Every exclusion / `#nosec` has a documented reason
4. GitHub Actions Security Scan job is green on the next release PR

## Testing

- Local: `./scripts/github-01-preflight-check.sh` — the gosec step mirrors the CI job
  (runs gosec in a `golang` container through `portunix container` when gosec is not
  installed) and prints a per-rule summary plus HIGH findings
- CI: Security Scan job on the release PR
