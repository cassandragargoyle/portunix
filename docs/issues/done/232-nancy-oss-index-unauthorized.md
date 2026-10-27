# Issue #232: make security and the Security Scan job fail in nancy: OSS Index returns 401

**Status**: ✅ Implemented
**Priority**: High
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-11
**Closed**: 2026-10-25
**GitHub**: #232
**Related**: #224, #228, #240, `docs/testing/acceptance-232.md`

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

- Confirmed in CI on the v2.6.3 release PR #240 (2026-10-24, nancy 1.2.0): gosec passed in all
  23 modules, then the nancy step failed with the same `401 Unauthorized`
- `go list -deps ./...` covers only the root module, the same gap #228 closed for gosec.
  Helper modules with their own `go.mod` and their dependencies are not checked

### Reachable Vulnerabilities Found by govulncheck

A trial run of `govulncheck ./...` (v1.9.0, Go 1.26, root module only) on the v2.6.3 tree
exits 3 and reports 8 vulnerabilities that the code calls:

| ID | Module | Found | Fixed | Summary |
| -- | ------ | ----- | ----- | ------- |
| GO-2026-6617 | `golang.org/x/net` | v0.49.0 | v0.60.0 | HTTP/2 server crash due to HPACK encoder race |
| GO-2026-6612 | `golang.org/x/net` | v0.49.0 | v0.60.0 | Double flow control refund on HTTP/2 server streams |
| GO-2026-6611 | `golang.org/x/net` | v0.49.0 | v0.60.0 | Excessive CPU from repeated initial window changes |
| GO-2026-6603 | `golang.org/x/net` | v0.49.0 | v0.60.0 | HTTP/2 server memory exhaustion due to Trailer headers |
| GO-2026-6443 | `google.golang.org/grpc` | v1.78.0 | v1.82.2 | Server panic via missing authority or Host headers |
| GO-2026-6348 | `google.golang.org/grpc` | v1.78.0 | v1.83.1 | Heap exhaustion via HTTP/2 DATA frame fragmentation |
| GO-2026-6061 | `google.golang.org/grpc` | v1.78.0 | v1.82.1 | xDS RBAC authorization engine and HTTP/2 issues |
| GO-2026-4762 | `google.golang.org/grpc` | v1.78.0 | v1.79.3 | Authorization bypass via missing leading slash in `:path` |

It also lists 4 vulnerabilities in imported packages and 25 in required modules that the code
does not appear to call.

The old versions are pinned in `go.mod`, `src/app/go.mod`,
`src/helpers/ptx-plugin-registry/go.mod` and `src/helpers/ptx-trace/go.mod`. Other helper
modules were not scanned yet. gRPC carries the plugin protocol, so the gRPC findings affect
every plugin server and client.

## Proposed Approach

Decide between two options:

| Option | Pros | Cons |
| ------ | ---- | ---- |
| Keep nancy, add OSS Index credentials as repository secrets and as local configuration | Smallest change | Every contributor needs an account, `make security` fails without one |
| Replace nancy with `govulncheck` | Official Go tool, no account, reports only vulnerabilities reachable from the code | Different report format, new tool to install |

Whichever tool is used should run per module, the way `scripts/gosec-scan.py` iterates over
modules.

Recommendation after the trial: replace nancy with `govulncheck`. It needs no account and
already found the reachable vulnerabilities above.

Dependency updates:

1. Raise `golang.org/x/net` to v0.60.0 or newer and `google.golang.org/grpc` to v1.83.1 or newer
   in every module that requires them, then `go mod tidy` per module
2. Rebuild, run the unit tests and a plugin smoke test (install, start, call, stop), because
   the gRPC update touches the plugin protocol

## Acceptance Criteria

1. `make security` exits 0 on a clean tree without manual credential setup, or documents the
   required setup and fails with a clear message when it is missing
2. The Security Scan job in GitHub Actions is green
3. Dependencies of every Go module are checked, not only those of the root module
4. `docs/contributing/SECURITY-SCAN.md` describes the dependency scan
5. No module is affected by a vulnerability that the code calls (`govulncheck` reports none)
6. Plugins communicate over gRPC after the dependency update (install, start, call, stop)

## Testing

- `make security` locally, with and without the required configuration
- Security Scan job on a pull request
- A module with a deliberately old dependency that has a known vulnerability is reported
