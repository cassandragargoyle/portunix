# Acceptance Protocol - Issue #228 (gosec Scan of Nested Modules)

**Issue**: Extend gosec Security Scan to modules with their own go.mod
**Branch**: `fix/228-gosec-nested-modules` (commit `337908f`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-11
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-34, amd64 (host) — build, scan, unit tests, `ptx-trace`
  end-to-end, failure injection, regression suites; Go 1.26.0, Podman 5.7.0
- Debian-based `golang:1.25` container — gosec scan through the preflight check

No software was installed on the host. gosec ran from a session-local scratch directory and in
a container started through `portunix container`.

## Scope Note

The branch replaces the single `gosec ./...` call with `scripts/gosec-scan.py`, which scans
every Go module of the repository, and cleans up the findings in the 20 modules that were
never scanned: 539 findings before the change, 0 after. The script is used by the Security
Scan job, `make security` and the preflight check. G702/G703 and generated code are excluded,
`docs/contributing/SECURITY-SCAN.md` documents the rules.

**Independence note**: the code on the branch, including the fix of defect D1, was written by
the same Claude Code session in the developer role before and during testing.

**Not covered**:

- Windows — all modules cross-compile (`GOOS=windows`), nothing was executed
- Security Scan job on GitHub — can only be confirmed once a release PR exists
- golangci-lint — could not be started locally or in a container, see O6
- Plugin installation from a zip archive, the only caller of the hardened `app.UnzipFile`
  besides the broken `unzip` command — covered by unit tests only
- Full regression suites ran on commit `367dbbb`; the later commit `337908f` (Go 1.24) was
  retested with build, `go vet`, unit tests of all modules and the scan

## Test Summary

- Total test scenarios: 30
- Passed: 25
- Failed: 0
- Skipped: 5 (Windows execution, CI job, golangci-lint, `unzip` command, plugin zip install)
- Defects found in the scope of #228: 1 (D1, fixed in `337908f` and retested)
- Defects found outside the scope: 2 (D2, D3)
- Observations: 7 (O1–O7)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | Scan of all modules, audit of suppressed and excluded findings, `#nosec` audit | host, `golang:1.25` container |
| Static | Cross-compile for Windows and macOS, `go vet` of all modules | host |
| Unit | New tests of the branch, unit tests of all modules compared against `main` | host |
| E2E | `ptx-trace` session, index, export and dashboard in a throwaway home directory | host |
| E2E | `portunix unzip` with regular and malicious archives | host |
| Failure injection | Findings, build error, new module, missing tool against the scan script | host |
| Regression | Root unit tests, `./test/integration`, `./test/e2e` compared against `main` | host + containers |

## Test Cases

### TC01: Every module is scanned

- **Given** the branch tree with 23 `go.mod` files outside `docs/` and `docs-site/`
- **When** `python3 scripts/gosec-scan.py` runs
- **Then** it prints one result line per module, 23 in total, and exits 0

### TC02: The three entry points use the same scan

- **Given** the Security Scan job, `make security` and the preflight check
- **When** each of them runs its gosec step
- **Then** all call `scripts/gosec-scan.py` and report the same 23 modules

### TC03: Suppressed HIGH findings are known

- **Given** the branch tree
- **When** gosec runs with `-nosec`, which ignores every annotation
- **Then** the HIGH findings are exactly the ones documented as allowed exceptions

### TC04: Every annotation names a rule and a reason

- **Given** all Go sources
- **When** every `#nosec` comment is matched against `#nosec G<nnn> -- <reason>`
- **Then** no comment fails the match

### TC05: A finding in any module fails the scan

- **Given** a temporary file with a finding in a helper module or in a satellite module
- **When** the scan runs
- **Then** the module is reported as `FAIL` with the rule and the scan exits 1

### TC06: A module that cannot be loaded fails the scan

- **Given** a temporary file that does not compile
- **When** the scan runs
- **Then** the module is reported as `FAIL` with the build error instead of being skipped

### TC07: New trace data is private

- **Given** an empty home directory
- **When** a trace session is recorded and exported
- **Then** data files have mode `0600` and directories `0750`

### TC08: Archive entries cannot leave the destination

- **Given** a zip archive with an entry named `../escaped.txt`
- **When** the archive is extracted
- **Then** extraction fails with `illegal file path` and nothing is written outside

## Test Results

### Static Checks

| Check | Result |
| ----- | ------ |
| `make build` | PASS — main binary and all helpers |
| `python3 scripts/gosec-scan.py` (local gosec) | PASS — exit 0, 23 of 23 modules, 23 JSON reports with `--report-dir` |
| gosec step of `scripts/github-01-preflight-check.sh` (`golang:1.25` container) | PASS — 23 of 23 modules |
| `make security` | gosec step PASS — 23 of 23 modules; the target then fails in `nancy`, see D3 |
| Module discovery | PASS — 23 `go.mod` files found by `find`, 23 modules scanned |
| gosec with `-nosec` and the rule exclusions | 151 findings, all annotated; HIGH: G101 ×3, G402 ×2 — the documented exceptions |
| gosec without rule exclusions | 689 findings: G204 ×459, G304 ×167, G702 ×43, G703 ×20 — nothing else, see O1 |
| `#nosec` audit | PASS — 151 annotations, all name a rule and a reason (G301 ×80, G306 ×40, G302 ×17, G107 ×6, G101 ×3, G305 ×3, G402 ×2) |
| `GOOS=windows` and `GOOS=darwin` build of all 23 modules | PASS |
| `go vet` of the root module and the satellite modules | PASS after the fix of D1 |

Suppressed HIGH findings found by the `-nosec` audit:

| Rule | Location | Reason given |
| ---- | -------- | ------------ |
| G101 | `src/helpers/ptx-credential/storage.go:21` | File name, not a credential |
| G101 | `src/helpers/ptx-ssh/credential.go:36` | Environment variable name, not a credential |
| G101 | `src/helpers/ptx-trace/main.go:1187` | Help text with a placeholder connection string |
| G402 | `src/helpers/ptx-proxmox/client.go:53` | Verification is on unless the profile opts out |
| G402 | `src/helpers/ptx-mcp/test.go:192` | Handshake probe of the locally started MCP server |

### Unit Tests

New tests of the branch — all PASS: `TestUnzipExtractsFiles`, `TestUnzipRejectsZipSlip`,
`TestCopyLimited` (`src/app`, `src/app/update`), `TestRestoreBackups`,
`TestRestoreBackupsReportsFailures`, `TestCopyDirectory`, `TestRandomFloat64Range`,
`TestShouldSampleWithRate`.

`go test ./...` in every helper and app module, branch against `main`:

| Tree | Packages passing | Packages failing |
| ---- | ---------------- | ---------------- |
| `main` | 27 | 3 |
| branch | 29 | 3 — the same three |

The three failing packages (`app/datastore`, `app/plugins/manager`, `app/plugins/server`)
fail with `setup failed`: `src/app` imports a package of the root module and does not resolve
on its own. This is unchanged by the branch.

### E2E — `ptx-trace` in a Throwaway Home Directory

| Step | Result |
| ---- | ------ |
| `trace start`, `trace exec`, `trace end` | PASS — session recorded |
| `trace stats`, `trace errors`, `trace query` | PASS — index is built on demand and queried |
| Modes of the created tree | PASS — directories `750`, `manifest.json` and `chunks/*.ndjson` `600` (`644` on `main`) |
| `trace export file -o`, `trace export ai -o` | PASS — both output files `600` |
| `trace serve` | PASS — dashboard returns HTTP 200, `/api/sessions` returns the session |

### E2E — `portunix unzip`

Skipped. The command does not extract anything on `main` or on the branch, see D2. The
traversal protection and the size cap were verified by the unit tests above.

### Failure Injection

| Injection | Result |
| --------- | ------ |
| MEDIUM finding in a helper module (`ptx-ssh`) | PASS — `FAIL  src/helpers/ptx-ssh`, exit 1 |
| HIGH finding in a satellite module (`src/cmd`) | PASS — `FAIL  src/cmd`, finding listed with file and line, exit 1 |
| File that does not compile | PASS — `FAIL` with `2 package(s) with build errors` |
| New module with a finding | PASS — discovered without configuration, 24 modules scanned |
| Finding in a file marked as generated | PASS — skipped as documented |
| gosec not on `PATH` | PASS — error message with the install command, exit 2 |
| Directory without `go.mod` | PASS — error message, exit 2 |
| Bare `#nosec` without a rule | Accepted by gosec, the scan passes — see O2 |

All injected files were removed after the run.

### Regression — Against `main`

Both trees were built and tested in separate worktrees, `main` at `8b041c5` and the branch at
`367dbbb`.

| Suite | `main` | Branch |
| ----- | ------ | ------ |
| Root unit tests (`.`, `portunix.ai/cmd`, satellites, `./src/...`, `./pkg/...`, `./internal/...`) | 7 failing tests | The same 7 |
| `go test ./test/integration/ -timeout 40m` | 26 failing tests | The same 26 |
| `go test ./test/e2e/...` | `TestContainerRunE2ETestSuite` with 7 sub-tests | The same |

The failing sets are identical, the branch adds one passing package (`app/sandbox`). The
shared failures are existing test debt (#225).

### Skipped

| Scenario | Reason |
| -------- | ------ |
| Execution on Windows | Linux tester; only cross-compiled |
| Security Scan job on a release PR | No release PR exists yet |
| golangci-lint | Tool could not be started, see O6 |
| `portunix unzip` end to end | Command is broken independently of the branch, see D2 |
| Plugin installation from a zip archive | Software installation, would need a container scenario with a plugin archive |

## Defects

### D1: `go vet` fails on `src/app/sandbox` — fixed

`copyDirectory` uses `os.OpenRoot`, which needs Go 1.24, while `src/app/sandbox/go.mod`
declared `go 1.23.3`. `go vet portunix.ai/app/sandbox/...` exited 1 on the branch and 0 on
`main`. The build was not affected and CI does not vet this module.

Fixed in `337908f`: the `go` directive of the ten modules below 1.24 was raised to `1.24.0`,
and four `fmt.Errorf` calls with a non-constant format string, which Go 1.24 `vet` reports,
were corrected. Retest: `go vet` of the root and satellite modules exits 0, unit tests of all
modules show the same results as before, the scan passes.

### D2: `portunix unzip` never extracts — outside #228

`src/cmd/unzip.go` passes `--path=<file>` to `app.ProcessArguments`, which matches the
`--path` prefix and takes the next argument as the value. The path ends up empty or holds the
`--destinationpath=...` argument. On `main` the command prints nothing and exits 0; on the
branch it reports `Error: open ...: no such file or directory` and exits 1, because the branch
no longer discards the error. A directory named `--destinationpath=` is created in both cases.

### D3: `make security` fails in `nancy` — outside #228

`nancy sleuth` ends with `401 Unauthorized` from OSS Index, the service needs credentials. The
gosec step before it passes.

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | gosec runs on every Go module in CI, `make security` and the preflight check | PASS — 23 of 23 modules in all three |
| 2 | All three places use the same invocation and exit 0 | PASS — all call `scripts/gosec-scan.py`; the CI run itself is open |
| 3 | HIGH findings are fixed in code, only the documented G101/G402 cases may be annotated | PASS — 5 annotated HIGH findings, all within the documented cases |
| 4 | Every `#nosec` names the rule and a reason | PASS — 151 of 151 |

Criterion 3 originally read "No HIGH finding is suppressed". It was changed during
implementation with the consent of the product owner: G101 false positives and G402 on explicit
user request may be annotated, the rule is documented in `SECURITY-SCAN.md` under *HIGH
Findings That May Be Annotated*. The issue file carries the new wording and the five
annotations match it.

## Coverage

- Scan script: discovery, workspace scan of satellite modules, load failure, report files and
  both entry points on the host were executed; the Windows path of the script was not
- Code fixes: archive traversal and size cap, backup restore, directory copy and the sampler
  are covered by new unit tests; the HTTP timeouts, the cookie attributes and the `int32`
  clamp have no test
- Permissions: verified end to end for `ptx-trace` only; the 71 tightened call sites in other
  modules were reviewed in the diff, not executed
- `os.Root` in the MCP template handler has no test

## CI Notes

- CI installs `gosec@latest`, a newer release may add rules and turn the job red without a
  code change — pinning the version is still open from #224
- The Security Scan job now uploads the `gosec-reports/` directory, also when the scan fails
- `go vet ./...` in CI covers only the root module; D1 was invisible to it
- `GO_VERSION` in `test.yml` is `1.21` while the root `go.mod` requires 1.25, the toolchain is
  downloaded on every run

## Observations

- **O1**: the global exclusion of G702/G703 hides 63 HIGH findings. Most build commands and
  paths from CLI arguments, but some sit in MCP handlers (`src/app/mcp/vm_handlers.go`), where
  the VM name comes from an MCP client and is joined into a path and a `bash` invocation
- **O2**: gosec accepts a bare `#nosec`; criterion 4 holds today but is not enforced by the scan
- **O3**: `index/sessions.db` of `ptx-trace` is still created with `644` by the SQLite
  driver, it is protected only by its `750` directory
- **O4** (outside #228): `trace exec -- sh -c '<cmd>'` records `exec: "-c": executable file
  not found`, the same on `main`
- **O5** (outside #228): the integration suite leaves containers behind; the seven created by
  this run were removed
- **O6** (outside #228): golangci-lint cannot run — the local binary and the `v1.64.8` image
  are built with Go 1.24 and refuse a module targeting 1.25, the `latest` image rejects the
  `.golangci.yml` format; the preflight check fails in this category on `main` too
- **O7**: `IndexSession` and `IndexEvent` in `ptx-trace/index` have no caller

## Recommendations

1. Confirm the Security Scan job on the next release PR and record the result here
2. Review the G702/G703 findings in `src/app/mcp` and validate names received from MCP clients
3. Let `scripts/gosec-scan.py` reject a `#nosec` without a rule and a reason
4. File issues for D2 and D3; O4 and O5 fit into #225
5. Run `go vet` per module in CI, the same way the scan iterates over modules
6. Run the scan and the `ptx-trace` scenario once on Windows

## Final Decision

**STATUS**: CONDITIONAL

Linux: all executed scenarios passed, the one defect in the scope of #228 is fixed and
retested, no regression against `main`. Open items are the Security Scan job, which cannot be
verified before a release PR exists, and execution on Windows.

**Approval for merge**: YES for the Linux scope; the Security Scan job is to be confirmed on
the next release PR
**Date**: 2026-10-11
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
