# Acceptance Protocol - Issue #244 (helpers missing --help-ai / --help-expert)

**Issue**: Helpers ptx-database, ptx-github, ptx-plugin-registry, ptx-proxmox missing --help-ai / --help-expert
**Branch**: `feature/244-helpers-help-ai-expert` (commit `0b28b07`; retest 1 on `c2b7d43`)
**Tester**: Claude Code in the tester role, on behalf of zdendaku
**Date**: 2026-10-26
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The change only adds help output; no software was installed and no host configuration was
changed, so the tests ran on the host.

## Scope Note

`ptx-database`, `ptx-github` and `ptx-plugin-registry` got a new `help.go` with
`handleMetaArg`, `showHelpAI` and `showHelpExpert`; `main()` calls `handleMetaArg` when the
flag is the only argument. `ptx-proxmox` got `showHelpExpert` and a `--help-expert` branch in its
existing `handleMetaArg`.

**Independence note**: the code and the unit tests were written by Claude Code in the developer
role in the same session. The scenarios below were designed separately in the tester role.

**Not covered**: Windows and macOS — the four modules vet for `GOOS=windows` and `GOOS=darwin`,
nothing was executed there.

## Test Summary

Initial run (`0b28b07`):

- Total test scenarios: 9
- Passed: 8
- Failed: 1 (TC07, see O1)
- Skipped: 0
- Defects found: 1 (O1)
- Observations: 1 (O2)

Retest 1 (`c2b7d43`) — see [Retest 1](#retest-1):

- Total test scenarios: 11 (TC01–TC11)
- Passed: 11
- Failed: 0
- Skipped: 0
- O1 and O2 resolved

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Static | `make build`, `gofmt`, `go vet` (also `GOOS=windows` / `darwin`), `golangci-lint` | host |
| Unit | `help_test.go` in the four helpers | host |
| Manual | direct invocation, JSON structure, content, command resolution, regression | host |
| Integration | flags through the main dispatcher (`portunix <command> --help-ai`) | host |
| Gate | `scripts/github-01-preflight-check.sh` helper section | host |

## Test Cases

### TC01: --help-ai exit code (AC1, AC2)

- **Given** the binaries built by `make build`
- **When** `./ptx-<helper> --help-ai` runs for the four helpers
- **Then** each exits with 0

Result: **PASS** — 4/4 exit 0.

### TC02: --help-expert exit code (AC1, AC2)

- **When** `./ptx-<helper> --help-expert` runs for the four helpers
- **Then** each exits with 0

Result: **PASS** — 4/4 exit 0 (previously `Error: unknown flag`).

### TC03: --help-ai format matches the other helpers (AC3)

- **Given** `ptx-ssh --help-ai` as reference (`tool`, `version`, `description`,
  `commands[{name, description}]`)
- **Then** the four outputs are valid JSON with the same keys, the correct `tool`, a non-empty
  command list and a non-empty description for every command

Result: **PASS** — commands: database 13, github 12, plugin-registry 1, proxmox 23.

### TC04: --help-expert is detailed, not a copy of --help (AC4)

| Helper | Expert lines | Basic `--help` lines | Identical | Sections |
| ------ | ------------ | -------------------- | --------- | -------- |
| ptx-database | 55 | 21 | no | 9 |
| ptx-github | 57 | 25 | no | 9 |
| ptx-plugin-registry | 37 | 15 | no | 7 |
| ptx-proxmox | 63 | 15 | no | 9 |

Flags, environment variables and storage paths were compared with the subcommand `--help`
output and the source (`PORTUNIX_GITHUB_TOKEN`/`GITHUB_TOKEN`/`GH_TOKEN`,
`~/.portunix/github/auth.json`, `PORTUNIX_PROXMOX_CONFIG`, `XDG_RUNTIME_DIR`, socket mode 0600).

Result: **PASS**.

### TC05: every --help-ai command exists

- **When** `./ptx-<helper> <command> --help` runs for every name listed by `--help-ai`
- **Then** each resolves to a real subcommand

Result: **PASS** — 49/49 resolved.

### TC06: regression of existing behaviour

- `--help`, `--version`, `ptx-database database engines`, `ptx-proxmox --description` and
  `--list-commands` unchanged
- an unknown flag (`ptx-github --bogus`) still fails with exit 1
- subcommand `--help` still served by Cobra

Result: **PASS**.

### TC07: flags through the main dispatcher

- **When** `portunix <command> --help-ai` / `--help-expert` runs for the four commands
- **Then** the helper help is shown with exit 0, like `portunix ssh --help-ai` and
  `portunix container --help-ai`

| Command | Exit |
| ------- | ---- |
| `portunix database --help-ai` | 1 |
| `portunix github --help-expert` | 1 |
| `portunix plugin-registry --help-ai` | 1 |
| `portunix proxmox --help-ai` | 1 (same on installed v2.7.0+dev.1) |
| `portunix proxmox --help-expert` | 1 |
| `portunix ssh --help-ai` | 0 (reference) |
| `portunix container --help-ai` | 0 (reference) |

Result: **FAIL** — see O1.

### TC08: unit tests and static checks

- `go test ./...` in the four modules — PASS
- `gofmt -l` — clean; `go vet` for linux, windows, darwin — PASS
- `golangci-lint run ./...` (root, v1.64.8 built with Go 1.26.8) — PASS

Result: **PASS**.

### TC09: pre-flight gate (AC5)

- **When** `scripts/github-01-preflight-check.sh` runs on the branch
- **Then** the helper section reports `All 19 helpers have --help-ai and --help-expert`

Result: **PASS** — the only remaining failure is the expected `Branch: ... (not main)`.

## Findings

### O1: --help-ai / --help-expert do not work through the dispatcher (defect)

The dispatcher runs the helper as `ptx-<helper> <command> --help-ai`, so `os.Args` has three
elements and the `len(os.Args) == 2` guard in `main()` skips `handleMetaArg`; Cobra then rejects
the flag. The pattern was copied from `ptx-proxmox`, whose `--help-ai` already failed the same way
before this change. `ptx-ssh` scans all arguments for the two flags and works through the
dispatcher.

**Required**: also handle `<command> --help-ai` / `<command> --help-expert` (the command name
followed by the flag) in the four helpers, and cover it with a unit test.

### O2: version header prints "vv" (cosmetic, not blocking)

`showHelpExpert` prints `PTX-<NAME> v%s`, but release builds embed the version with the `v`
prefix (`build-with-version.sh`), so the header reads `PTX-GITHUB vv2.7.0`. The same pattern is
used by 12 existing helpers (e.g. `ptx-ssh`: `PTX-SSH vv2.7.0+dev.1`), so it is a project-wide
cosmetic issue, not part of this issue.

## Coverage

- Direct invocation, JSON contract, content accuracy and command tree consistency are covered
  by unit tests and manual cases
- Dispatcher invocation is not covered by unit tests (O1)

## CI Notes

- The new `help_test.go` files run in the per-module unit test job
- The pre-flight helper gate now passes for all 19 helpers

## Retest 1

**Commit**: `c2b7d43` — `fix(244): handle help flags through the dispatcher, fix version header`
**Date**: 2026-10-26
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

The fix adds `handleMetaArgs`, which accepts the flags directly and in the dispatcher form
`<command> --help-ai` (the command name is checked against the Cobra tree, aliases included),
and strips the `v` prefix from the version in the `--help-expert` header of all 13 helpers that
print one.

### TC07 (rerun): flags through the main dispatcher — PASS

| Command | Exit | Output |
| ------- | ---- | ------ |
| `portunix database --help-ai` | 0 | JSON, `ptx-database`, 13 commands |
| `portunix database --help-expert` | 0 | `PTX-DATABASE ...` |
| `portunix db --help-ai` (alias) | 0 | JSON |
| `portunix github --help-ai` / `--help-expert` | 0 / 0 | JSON, 12 commands / expert text |
| `portunix plugin-registry --help-ai` / `--help-expert` | 0 / 0 | JSON, 1 command / expert text |
| `portunix proxmox --help-ai` / `--help-expert` | 0 / 0 | JSON, 23 commands / expert text |

### TC10: version header with an embedded version (O2) — PASS

`--help-expert` built with `-X main.version=v2.7.0` (build-with-version.sh) and with
`-X main.version=2.7.0` (GoReleaser) prints `v2.7.0` in both cases for ptx-github,
ptx-database, ptx-proxmox, ptx-ssh and ptx-make. No `v%s` header without `TrimPrefix` remains
in `src/helpers/`.

### TC11: flags in other positions are not intercepted — PASS

- `portunix proxmox exec x -- echo --help-ai` runs `exec` (fails on the missing profile, as
  expected), the flag after `--` is not taken as a help request
- `ptx-github github clone --help-ai` — rejected by Cobra (`unknown flag`), as before
- `ptx-database unknown --help-ai` — `unknown command`, not intercepted

### Regression and static checks — PASS

- TC01–TC06 repeated: direct calls exit 0, JSON contract unchanged, subcommand `--help`,
  `--description`, `--list-commands`, unknown flag (exit 1), `portunix ssh|container --help-ai`
  unchanged
- `go vet` (linux, windows, darwin) and `go test ./...` pass in all 13 changed modules; new tests
  `TestHandleMetaArgs_DirectAndDispatcherForm` and `TestHandleMetaArgs_NotHandled` pass in the
  four helpers
- `gofmt -l src/helpers/` clean; `golangci-lint run ./...` (root) passes
- Pre-flight: `All 19 helpers have --help-ai and --help-expert`, golangci-lint and gosec pass;
  the only failure is the expected `Branch: ... (not main)`

### Findings status

- **O1**: resolved
- **O2**: resolved (all 13 helper headers)

## Final Decision

**STATUS**: PASS

Initial run: CONDITIONAL (O1). Retest 1 on `c2b7d43`: all acceptance criteria met, including
invocation through `portunix <command>`; O1 and O2 resolved, no regressions found.

**Approval for merge**: YES
**Date**: 2026-10-26
**Tester signature**: Claude Code (tester role)
