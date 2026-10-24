# Acceptance Protocol - Issue #224 (gosec Findings Cleanup)

**Issue**: CI Security Scan fails: clean up gosec findings (4 HIGH in ptx-pft)
**Branch**: `fix/224-gosec-findings-cleanup` (commit `42dbf86`, includes merge of `main` at `2b5b707`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-11
**Testing OS**:

- Ubuntu 26.04, kernel 7.0.0-34, amd64 (host) — build, unit tests, `pft` end-to-end,
  failure injection, integration suite; Go 1.26.0
- Debian-based `golang:1.25` container — gosec scan (same Go line as `go.mod`)
- Ubuntu 22.04.5 LTS container — package installation from archives

No software was installed on the host. gosec and the test packages ran only in containers
started through `portunix container`.

## Scope Note

The branch makes `gosec -exclude=G204,G304 ./...` exit 0 on the root module: the four HIGH
findings in `ptx-pft` are fixed in code, archive extraction gets a per-file size cap,
credential-bearing files and logs get `0600`/`0750`, intentional patterns carry a
`#nosec <rule> -- <reason>` annotation, and G204/G304 are excluded in CI, `make security`
and the preflight check (documented in `docs/contributing/SECURITY-SCAN.md`).

**Independence note**: the last commit on the branch (`42dbf86`, preflight fix) was written
by the same Claude Code session in the developer role before switching to the tester role.

**Not covered**:

- Windows — only compiled (`GOOS=windows go vet`, `GOOS=windows go build`), not executed
- Acceptance criterion 4 (Security Scan green on the next release PR) — can only be
  confirmed on GitHub once a release PR exists
- Helper modules with their own `go.mod` — out of scope, tracked as #228
- `pft` sync/push against a live Fider instance and `pft deploy` (compose files) — no
  feedback provider was deployed; the rewritten write paths were exercised through
  `assign`, `unassign` and `link`, and through unit tests

## Test Summary

- Total test scenarios: 26
- Passed: 24
- Failed: 0
- Skipped: 2 (Windows execution, CI job on a release PR)
- Open defects in the scope of #224: 0
- Observations: 5 (O1–O5)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | gosec with and without exclusions, `#nosec` audit, Windows cross-compile | `golang:1.25` container, host |
| Unit | `ptx-pft` file helpers, `hashItem`, archive extraction, logging | host |
| E2E | `pft` configure / add / category / assign / link / unassign on a throwaway project | host |
| E2E | Package installation from tar.gz and zip archives | Ubuntu 22.04 container |
| Failure injection | Symlink escape, read-only file, archive path traversal, size limit | host |
| Regression | `./test/integration` suite compared against `main` | host + containers |

## Test Cases

### TC01: Security scan passes

- **Given** the branch tree and gosec latest in a `golang:1.25` container
- **When** `gosec -exclude=G204,G304 ./...` runs
- **Then** it exits 0 with no findings

### TC02: No HIGH finding is suppressed

- **Given** the same tree
- **When** gosec runs without any exclusion
- **Then** only G204 and G304 are reported and no `#nosec` names G703 or G115

### TC03: Feedback file rewrite stays inside its directory

- **Given** a feedback item whose file is replaced by a symlink to a file outside the project
- **When** `pft assign` or `pft link` rewrites the item
- **Then** the command reports an error and the outside file is unchanged

### TC04: Credential-bearing files are private

- **Given** a new PFT project
- **When** `pft configure` writes `.pft-config.json`
- **Then** the file mode is `0600`

### TC05: Archive entries cannot leave the destination

- **Given** a zip and a tar.gz archive with an entry named `../evil.txt`
- **When** the archive is extracted
- **Then** extraction fails with `invalid file path` and nothing is written outside

### TC06: Archive installation still works

- **Given** a clean Ubuntu 22.04 container with the branch binaries
- **When** `portunix install actionlint` (tar.gz) and `portunix install terraform` (zip) run
- **Then** both install and the installed binaries execute

## Test Results

### Static Checks

| Check | Result |
| ----- | ------ |
| `gosec -exclude=G204,G304 ./...` | PASS — exit 0, 0 findings, 81 files, 30412 lines, 23 `#nosec` |
| `gosec ./...` (no exclusions) | 117 findings: G204 ×75, G304 ×42 — no HIGH, nothing else |
| `#nosec` audit | PASS — all 23 name a rule and a reason (G301 ×16, G306 ×3, G107 ×2, G305 ×2); none on a HIGH rule |
| gosec step of `scripts/github-01-preflight-check.sh` | PASS — `gosec passed (no findings)` |
| `make build` | PASS — main binary and all helpers |
| `GOOS=windows go vet` (changed packages), `GOOS=windows go build .` | PASS |

### Unit Tests

`go test ./src/helpers/ptx-pft/ ./src/pkg/archive/ ./pkg/logging/... -count=1` — all PASS,
including `TestRewriteExistingFile`, `TestRewriteExistingFile_MissingFile`,
`TestRewriteExistingFile_Symlink`, `TestWritePrivateFile_NarrowsExistingMode`,
`TestHashItem_NonASCIIStable`, `TestCopyLimited`.

Root-module unit tests outside `test/`: one failure, `TestMainWithoutArguments`
(expects `Available Commands:` in the help output). It fails identically on `main`.

### E2E — `pft` on a Throwaway Project

| Step | Result |
| ---- | ------ |
| `pft project create demo` | PASS — project tree created |
| `pft configure --name demo --path <dir>` | PASS — `.pft-config.json` mode `600` |
| `pft add --area voc --title "Dark mode"` | PASS — item `P01`, file mode `644` |
| `pft category add A --area voc` | PASS — `categories.json` mode `644` |
| `pft assign P01 --category A` | PASS — `categories:` front matter added |
| `pft link P01 '#224'` | PASS — `linked_issue: #224` added |
| File mode preserved on rewrite (set to `664` beforehand) | PASS — still `664` |
| `pft unassign P01 --all` | PASS — categories removed, `linked_issue` kept |

### E2E — Installation From Archives (Ubuntu 22.04 Container)

| Package | Archive | Result |
| ------- | ------- | ------ |
| `actionlint` 1.7.4 | tar.gz | PASS — 11 files extracted, `actionlint -version` prints `1.7.4` |
| `terraform` 1.15.2 | zip | PASS — extracted, `terraform version` prints `Terraform v1.15.2` |

The container was started with `portunix container run` and the branch binaries
(`portunix`, `ptx-installer`) were copied in; see O1 for why `run-in-container` was not used.

### Failure Injection

| Injection | Result |
| --------- | ------ |
| Feedback file replaced by a symlink to a file outside the project, then `pft assign` / `pft link` | PASS — `refusing to write P01-dark-mode.md: not a regular file`, outside file byte-identical |
| Feedback file made read-only (`444`), then `pft assign` | PASS — `permission denied` reported, no crash |
| Zip entry `../evil.txt` | PASS — `invalid file path`, nothing outside the destination |
| Tar.gz entry `../evil.txt` | PASS — `invalid file path`, nothing outside the destination |
| Entry one byte over the size limit (`TestCopyLimited`) | PASS — extraction fails |
| Regular zip round trip after the changes | PASS — content identical |

The traversal tests were temporary test functions, removed after the run (see
Recommendation 2).

### Regression — Integration Suite Against `main`

`go test ./test/integration/ -timeout 40m`, run once on the branch and once on `main`
(`2b5b707`) in a separate worktree:

| Tree | Failing tests |
| ---- | ------------- |
| `main` | 26 |
| branch | 27 — the same 26 plus `TestIssue052MCPSTDIOSeparation` |

`TestIssue052MCPSTDIOSeparation` is not a regression of the branch. The test runs
`go build -o ../../portunix` from the repository root, so the output lands two levels
above the checkout. In the regular checkout that path is the repository directory itself
and the build stops with `build output "../../portunix/portunix" already exists and is a
directory`; in the scratch worktree used for the `main` baseline the path was free. The
branch code passes the test in a worktree without that clash, and `test/integration/` is
unchanged by the branch.

The 26 shared failures are existing test debt (#225).

### Skipped

| Scenario | Reason |
| -------- | ------ |
| Execution on Windows | Linux tester; only cross-compiled |
| Security Scan job on a release PR (criterion 4) | No release PR exists yet |

## Acceptance Criteria

| # | Criterion | Result |
| - | --------- | ------ |
| 1 | `gosec ./...` exits 0 on the published tree (same command as CI) | PASS with the documented `-exclude=G204,G304`, applied identically in CI, `make security` and preflight |
| 2 | No HIGH finding is suppressed — all are fixed in code | PASS |
| 3 | Every exclusion / `#nosec` has a documented reason | PASS |
| 4 | Security Scan job is green on the next release PR | OPEN — verifiable only on GitHub |

## Coverage

- All four HIGH sites are covered by unit tests (`rewriteExistingFile`, `hashItem`) and
  three of the four by E2E commands; `UpdateFileWithFiderID` (`sync.go`) shares the
  helper but was not driven end-to-end (needs a Fider instance)
- Archive extraction: size cap by unit test, traversal by temporary injection tests,
  normal operation by two real installations
- Compose files written with `0600` (`pft deploy`) were not exercised

## CI Notes

- CI, `make security` and the preflight check must keep the same gosec command; the
  list of places is in `docs/contributing/SECURITY-SCAN.md`
- CI installs `gosec@latest`, so a newer gosec release may introduce new rules and turn
  the job red without a code change
- The scan was run on the Gitea tree; criterion 1 speaks of the published tree, which
  omits private files — the next preflight run before publishing covers that

## Observations

- **O1** (outside #224): `portunix container run-in-container <package>` copies only
  `portunix` into the container, the installation then fails with `ptx-installer helper
  not found` and the command still exits 0
- **O2** (outside #224): `pft assign` / `pft link` do not find an item by its
  `legacy_id` (`UC001`), although the help text uses `UC001` as an example
- **O3** (behaviour change, intended): a feedback file that is a symlink is no longer
  rewritten — the command refuses it
- **O4** (outside #224): `TestIssue052MCPSTDIOSeparation` builds into `../../portunix`
  relative to the repository root, so its result depends on the checkout location and it
  can drop a binary outside the repository
- **O5** (outside #224): the integration suite leaves running containers behind
  (`portunix-claude-code-test`, `portunix-nodejs-*`, `portunix-hugo-test`,
  `portunix-cp-test-*`); the ones created by this test run were removed

## Recommendations

1. Confirm criterion 4 on the next release PR and record the result here
2. Add permanent traversal tests for `ExtractZip` and `ExtractTarGz` — both sites carry
   `#nosec G305` and currently have no automated test
3. Pin the gosec version in CI and in the preflight check
4. File issues for O1, O2 and O4 (O4 and O5 fit into #225)
5. Run the `pft` scenarios once on Windows (`os.Root`, file modes)

## Final Decision

**STATUS**: CONDITIONAL

Linux: all executed scenarios passed, no defects found in the scope of #224, no
regression against `main`. Open items are criterion 4, which cannot be verified before a
release PR exists, and execution on Windows.

**Approval for merge**: YES for the Linux scope; criterion 4 is to be confirmed on the
next release PR
**Date**: 2026-10-11
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
