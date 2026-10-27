# Acceptance Protocol - Issue #232 (Replace nancy with govulncheck)

**Issue**: make security and the Security Scan job fail in nancy: OSS Index returns 401
**Branch**: `fix/232-replace-nancy-with-govulncheck` (commit `542df3c`, D1/D2 fix retested on top)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-25
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-38, amd64 (host) — build, scans, unit and integration tests,
  failure injection, plugin lifecycle in a throwaway home directory; Go 1.26.8 and the
  Go 1.26.10 toolchain from the module cache (`GOTOOLCHAIN=go1.26.10`)

Developer tools of the project were installed on the host with `go install` into `~/go/bin`:
`govulncheck` v1.9.0 and `gosec`. No other software was installed. Plugin tests ran with
`HOME` set to a scratch directory, the real `~/.portunix` was not touched.

## Scope Note

The branch replaces `go list -json -deps ./... | nancy sleuth` with
`scripts/govulncheck-scan.py`, which runs `govulncheck` in every Go module and fails a module
on a vulnerability that the code calls. `make security`, `make deps` and the Security Scan job
use the new script, the job runs on the latest stable Go. Vulnerable dependencies were updated:
gRPC v1.78.0 → v1.84.0, `golang.org/x/net` → v0.61.0, `golang.org/x/crypto` → v0.58.0,
go-git v5.12.0 → v5.19.3, go-billy, circl. The fixed versions require Go 1.26, the `go`
directive of the touched modules was raised to `1.26.0`.

**Independence note**: the code on the branch was written by the same Claude Code session in
the developer role before testing.

**Not covered**:

- Security Scan job on GitHub — can only be confirmed on a pull request
- Windows — touched modules cross-compile (`GOOS=windows`), nothing was executed
- Plugin start, call and stop through the CLI — blocked by D3, which exists on `main` too;
  gRPC communication is covered by the `ptx-plugin-registry` lifecycle test instead
- `ptx-ssh` against a real SSH server — unit tests only

## Test Summary

- Total test scenarios: 16
- Passed: 13
- Failed: 0
- Skipped: 3 (CI job, plugin start/call/stop through the CLI, Windows execution)
- Defects found in the scope of #232: 2 (D1, D2, both fixed and retested)
- Defects found outside the scope: 1 (D3)
- Observations: 4 (O1–O4)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make security`, scan script on the branch tree, report files, module count | host |
| Static | Cross-compile for Windows and macOS, `go vet` of the touched modules | host |
| Failure injection | Called and uncalled old dependency, broken module, missing tool, wrong repo | host, fake repository |
| Unit | Root unit tests, tests of the touched modules, `src/app/plugins/...` | host |
| Integration | `ptx-plugin-registry` gRPC lifecycle | host |
| E2E | `portunix github clone` (go-git update), plugin create/install/enable/start | host, throwaway home |
| Regression | Plugin lifecycle compared against a `main` build | host, throwaway home |

## Test Cases

### TC01: make security passes on a clean tree

- **Given** the branch tree, gosec and govulncheck installed, Go 1.26.10
- **When** `make security` runs
- **Then** both scans report 23 of 23 modules and the target exits 0

### TC02: Every module is scanned

- **Given** 23 `go.mod` files outside `docs/` and `docs-site/`
- **When** `python3 scripts/govulncheck-scan.py` runs
- **Then** it prints one result line per module, 23 in total

### TC03: An old toolchain is reported with a clear hint

- **Given** the local Go 1.26.8
- **When** the scan runs
- **Then** modules fail only on `stdlib` findings and the script names the Go version that fixes them

### TC04: A missing tool fails with an install hint

- **Given** `govulncheck` is not in `PATH`
- **When** the script or `make security` runs
- **Then** it exits 2 and prints the `go install` command

### TC05: A called vulnerability in a nested module fails the scan

- **Given** a fake repository with a helper module that requires `golang.org/x/crypto` v0.47.0
  and calls `ssh.Dial`
- **When** the scan runs on that repository
- **Then** the helper is reported as `FAIL` with the OSV IDs and the fixed versions, exit 1

### TC06: An uncalled vulnerability does not fail the scan

- **Given** a helper module with the same old `x/crypto`, using only `sha3`
- **When** the scan runs
- **Then** the module is reported as `PASS`

### TC07: A module that cannot be loaded fails the scan

- **Given** a helper module that does not compile
- **When** the scan runs
- **Then** the module is reported as `FAIL - govulncheck could not run` with the build error

### TC08: Plugins communicate over gRPC after the update

- **Given** binaries built from the branch
- **When** the `ptx-plugin-registry` helper is started and a gRPC client calls it
- **Then** HealthCheck and ListPluginsForPlatform answer as expected

## Test Results

### Static Checks

| Check | Result |
| ----- | ------ |
| `GOTOOLCHAIN=go1.26.10 make security` | PASS — exit 0, `gosec: all 23 module(s) passed`, `govulncheck: all 23 module(s) passed` |
| Module count | PASS — 23 `go.mod` files, 23 result lines |
| `--report-dir` | PASS — 23 files `govulncheck-<module>.json`, root report parses into 277 JSON objects |
| Scan with local Go 1.26.8 | PASS — exit 1, 12 modules fail, all findings are `stdlib`, hint `fixed by building with Go 1.26.10 or newer` |
| `nancy` references | PASS for code and CI; D2 in `test/TESTING.md` |
| Security Scan job YAML | PASS — parses, steps: gosec, govulncheck, both report uploads, `go-version: stable` |
| Cross-compile `GOOS=windows` / `darwin` | PASS — root and the 7 touched helpers |
| `go vet` | PASS — root and the 7 touched helpers |

### Failure Injection

| Scenario | Result |
| -------- | ------ |
| Nested module calls `ssh.Dial` with `x/crypto` v0.47.0 | PASS — `FAIL src/helpers/ptx-old - 7 called vulnerability(ies)`, exit 1 |
| Nested module imports only `sha3` from the same `x/crypto` | PASS — `PASS` |
| Module that does not compile | PASS — `FAIL ... govulncheck could not run` with the compiler error |
| `govulncheck` not in `PATH` (script) | PASS — exit 2, install hint |
| `govulncheck` not in `PATH` (`make security`) | PASS — gosec passes, then exit 2 with the install hint |
| `REPO_DIR` without `go.mod` | PASS — `error: no go.mod in /nonexistent`, exit 2 |
| Branch state before the dependency update | PASS — 13 modules reported, including gRPC, `x/net`, go-git and `x/crypto` |

### Unit and Integration Tests

| Suite | Result |
| ----- | ------ |
| `go test -tags=unit ./...` (root) | PASS |
| `ptx-github`, `ptx-ssh`, `ptx-trace`, `ptx-plugin-registry`, `ptx-mcp`, `ptx-installer`, `ptx-prompting` | PASS |
| `src/app/plugins/...` inside the workspace | PASS |
| `TestPtxPluginRegistry_Lifecycle` (`-tags=integration`) | PASS — HealthCheck, ListPluginsForPlatform, empty platform rejected |

`go test ./...` in `src/app` alone fails with `no required module provides package
portunix.ai/portunix/src/pkg/platform`; the stub module does not resolve outside the workspace,
the same on `main`.

### E2E

| Scenario | Result |
| -------- | ------ |
| `portunix github clone cassandragargoyle/portunix --depth 1` | PASS — cloned through go-git v5.19.3, HEAD `31be081` |
| `plugin create t232-grpc`, built with gRPC v1.84.0 | PASS — the plugin builds; the template pinned gRPC v1.78.0, see D1 (fixed) |
| `plugin install` / `enable` / `list` / `info` | PASS |
| `plugin start` / `health` / `stop` | SKIPPED — `plugin ... is not enabled`, see D3 |

### Regression — Against `main`

The plugin lifecycle was repeated with a binary built from `main` in a separate worktree. The
result is the same: `install` and `enable` succeed, `start` fails with `plugin t232-grpc is not
enabled`. The branch introduces no regression in this area.

### Skipped

- Security Scan job on GitHub — needs a pull request
- Plugin start, call and stop through the CLI — D3
- Windows execution

## Defects

### D1: `plugin create` generates a template with gRPC v1.78.0 — fixed

`src/cmd/plugin.go:1230` writes `google.golang.org/grpc v1.78.0` into the `go.mod` of a new
plugin. Every plugin created from the template starts with the gRPC version that has the four
called vulnerabilities this issue fixes (GO-2026-4762, GO-2026-6061, GO-2026-6348,
GO-2026-6443). The template is not a module of the repository, so the scan does not see it.

Fixed: the template requires `google.golang.org/grpc v1.84.0` and
`golang.org/x/net v0.61.0 // indirect`. gRPC v1.84.0 alone selects `x/net` v0.57.0, which
still has the four HTTP/2 vulnerabilities, so `x/net` is pinned explicitly. Retest: a plugin
created by the rebuilt binary passes `go mod tidy` and `go build`, `govulncheck ./...` with
Go 1.26.10 reports `No vulnerabilities found` and exits 0; root unit tests pass.

### D2: `test/TESTING.md` still lists Nancy — fixed

`test/TESTING.md:198` lists **Nancy** under *Security Scan Integration*. It should name
govulncheck and point to `docs/contributing/SECURITY-SCAN.md`.

Fixed: the line names govulncheck and links `SECURITY-SCAN.md`. Retest: `git grep -i nancy`
finds only the history note in `scripts/govulncheck-scan.py`.

### D3: plugin enable does not survive the CLI process — outside #232

`portunix plugin enable` stores `Enabled` in the registry, but `Manager.StartPlugin`,
`StopPlugin` and `GetPluginHealth` look up the in-memory map `Manager.plugins`
(`src/app/plugins/manager/manager.go:398`), which a new CLI process does not fill from the
registry. Every following `start`, `health` or `stop` ends with `plugin <name> is not enabled`.
The same on `main`.

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | `make security` exits 0 without credential setup, or fails with a clear message | PASS — exit 0 with Go 1.26.10; clear messages for a missing tool and for an old toolchain |
| 2 | The Security Scan job in GitHub Actions is green | OPEN — to be confirmed on a pull request |
| 3 | Dependencies of every Go module are checked | PASS — 23 of 23 modules, nested modules verified by failure injection |
| 4 | `docs/contributing/SECURITY-SCAN.md` describes the dependency scan | PASS — section *Dependency Scan* with fixing and local usage |
| 5 | No module is affected by a called vulnerability | PASS — 23 of 23 with Go 1.26.10; a plugin generated from the template passes too (D1) |
| 6 | Plugins communicate over gRPC after the update | PASS (partial) — `ptx-plugin-registry` gRPC lifecycle; CLI start/call/stop blocked by D3 |

## Coverage

- Scan script: discovery, workspace and standalone modules, called and uncalled findings,
  stdlib hint, load failure, missing tool, missing repository and report files were executed;
  the Windows path of the script was not
- Dependency update: build, vet and unit tests of every touched module; go-git end to end
  through `github clone`; gRPC through the registry helper; `x/crypto` in `ptx-ssh` by unit
  tests only
- CI workflow: reviewed and parsed, not executed

## CI Notes

- The Security Scan job uses `go-version: stable` with `check-latest`, other jobs still use
  `GO_VERSION: '1.21'` and download the toolchain from the `go` directive (now 1.26.0)
- The govulncheck step runs with `if: always()`, so both scan results are visible when gosec fails
- `govulncheck@latest` and `gosec@latest` are not pinned, a new release can change the result
  without a code change
- A new entry in the Go vulnerability database can turn the job red without a code change —
  that is the intended behaviour

## Observations

- **O1**: the local Go 1.26.8 fails `make security` on stdlib findings; releases are built with
  the local Go, so the toolchain must be upgraded to 1.26.10 or newer before the next release
- **O2**: the error output of a module that cannot be loaded contains empty lines and the
  generic govulncheck hint about package patterns
- **O3**: `docs/plugin-development/languages/go/template/go.mod` requires gRPC v1.58.0 and
  `x/net` v0.12.0; `docs/` is excluded from both scans by design
- **O4**: the direct dependency update raised the `go` directive to 1.26.0 in 9 modules, the
  other modules stay at 1.24/1.25

## Recommendations

1. Confirm the Security Scan job on the pull request and record the result here
2. Upgrade the local Go to 1.26.10 or newer (O1)
3. File an issue for D3; after it is fixed, repeat the plugin start, call and stop scenario
4. Pin the versions of the scanners in CI (open since #224)

## Final Decision

**STATUS**: CONDITIONAL

Linux: all executed scenarios passed and no regression against `main`. The dependency scan
works without an account, covers every module and reports only called vulnerabilities.
D1 and D2 are fixed and retested. The open item is the Security Scan job, which can only be
confirmed on a pull request.

**Approval for merge**: YES for the Linux scope; the Security Scan job is to be confirmed on
the pull request
**Date**: 2026-10-25
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
