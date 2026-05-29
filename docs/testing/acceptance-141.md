# Acceptance Protocol — Issue #141

**Issue**: PTX-TRACE Helper Implementation — Phase 5 Bash integration improvements
**Branch**: `feature/141-ptx-trace-bash-integration`
**Commit under test**: `71ba8a3`
**Tester**: zdendaku (role: Tester / generic)
**Date**: 2026-05-06
**Testing OS**: Linux 6.17 (host) — Ubuntu derivative; Bash 5.x; Go 1.x via `make build`
**Method**: Direct host execution against a sandboxed `$HOME` (no container isolation needed — pure CLI binary, no installer side-effects)

## Scope

The change adds the final missing checkbox in Issue #141 Phase 5: **Bash
integration improvements**. Specifically:

- `portunix trace pipe <op>` — wrap stdin → stdout pipe
- `portunix trace exec <op> -- <cmd>` — wrap external command
- Bash SDK (`sdk/bash/ptx-trace.sh`)
- Bash ETL example (`sdk/examples/bash/etl-pipeline.sh`)
- Helper-level `README.md`
- Bash SDK README (en + cs)
- SDK overview README updates

## Test Plan

| Area | Tests |
| ---- | ----- |
| Command registration & help | TC-01, TC-02 |
| `pipe` semantics | TC-03 (text), TC-04 (binary), TC-05 (ad-hoc) |
| `exec` semantics | TC-06 (success), TC-07 (failure + stderr capture), TC-08 (missing binary), TC-09 (ad-hoc), TC-10 (signal forward), TC-11 (stdin forward) |
| Session precedence | TC-12 |
| Bash SDK | TC-13 (functions), TC-14 (env override), TC-15 (session_id), TC-16 (fallback stubs from README) |
| End-to-end | TC-17 (ETL example) |
| Regression | TC-18 |
| Documentation | TC-19 |

## Test Cases (Given/When/Then)

### TC-01 — `pipe`/`exec` listed in `trace --help`

- **Given** the freshly built `ptx-trace` binary
- **When** running `portunix trace --help`
- **Then** both `pipe` and `exec` appear with one-line summaries; their own `--help` pages render flag tables
- **Result**: ✅ PASS

### TC-02 — Help schemas updated

- **Given** the new commands
- **When** running `--help-ai` (JSON) and `--help-expert`
- **Then** both modes mention `trace pipe` and `trace exec`
- **Result**: ✅ PASS

### TC-03 — `pipe` text mode

- **Given** an active session
- **When** piping `a\nbc\ndef\n` through `trace pipe stage1 --format text --tag t1`
- **Then** stdout is byte-perfect identical to stdin; recorded event has `bytes=9`, `lines=3`, `format=text`, tag `t1`, `status=success`, non-zero `duration_us`
- **Result**: ✅ PASS

### TC-04 — `pipe --binary`

- **Given** 1024 random bytes from `/dev/urandom`
- **When** piping with `--binary`
- **Then** SHA-256 of input equals SHA-256 of output; event has `bytes=1024` and **no** `lines` field
- **Result**: ✅ PASS

### TC-05 — `pipe` ad-hoc session

- **Given** no active session and `--session-name tc05-name`
- **When** piping data through
- **Then** session `tc05-name` is created, marked `completed`, has the event attached
- **Result**: ✅ PASS

### TC-06 — `exec` success

- **Given** an active session
- **When** running `trace exec ok_cmd -- bash -c "sleep 0.05; echo hi"`
- **Then** rc=0; stdout `hi` forwarded; event has `exit_code=0`, `duration_ms ≥ 50`, `status=success`, command captured in `input.command`
- **Result**: ✅ PASS

### TC-07 — `exec` failure with `--capture-stderr`

- **When** running `trace exec fail_cmd --capture-stderr -- bash -c "echo err1>&2; echo err2>&2; exit 13"`
- **Then** rc=13; stderr lines forwarded to terminal; event has `exit_code=13`, `error.code=E_EXEC_FAILED`, `error.severity=high`, `context.stderr_tail` contains both stderr lines
- **Result**: ✅ PASS

### TC-08 — `exec` non-existent binary

- **When** running `trace exec missing_cmd -- /no/such/binary-xyz`
- **Then** rc=127 (POSIX convention); event records `E_EXEC_FAILED` with the underlying `fork/exec ... no such file` message
- **Result**: ✅ PASS

### TC-09 — `exec` ad-hoc session

- **Given** no active session
- **When** running `trace exec adhoc_exec --tag oneshot -- date '+TC09-%Y'`
- **Then** an auto-named session is created and ended `completed`, 1 event attached
- **Result**: ✅ PASS

### TC-10 — Signal forwarding

- **When** wrapping `sleep 30` and sending SIGTERM to the wrapper after 0.3 s
- **Then** wrapper exits within ~0.3 s (proves signal reached the child); shell sees rc=143 (128+SIGTERM); event records `exit_code=143`
- **Result**: ✅ PASS (after follow-up fix to map signal-killed children to POSIX `128+signum`)

### TC-11 — Stdin forwarding

- **When** piping `line1\nline2\n` to `trace exec cat_test -- cat`
- **Then** wrapped `cat` receives stdin and prints both lines
- **Result**: ✅ PASS

### TC-12 — Active session precedence

- **Given** an explicitly started session and calls to `pipe`/`exec` with `--session-name ignored`
- **When** the calls run
- **Then** events attach to the active session; no extra session is created; `--session-name` is silently ignored
- **Result**: ✅ PASS

### TC-13 — Bash SDK file integrity

- **When** running `bash -n` on `ptx-trace.sh` and sourcing it
- **Then** syntax OK; all 10 documented functions are defined
- **Result**: ✅ PASS

### TC-14 — `PTX_TRACE_BIN` override + missing-binary error

- **Given** `PTX_TRACE_BIN` set to an explicit path
- **When** sourcing the SDK and starting a session
- **Then** the binary is invoked; with a non-existent path, rc=127 and a clear `portunix binary not found` message
- **Result**: ✅ PASS

### TC-15 — `ptx_trace_session_id` helper

- **Given** an active session named `tc15`
- **When** calling `ptx_trace_session_id`
- **Then** `ses_2026-05-06_tc15` is printed (rc=0); when no session is active, rc=1 with empty stdout
- **Result**: ✅ PASS (after follow-up fix in awk extractor)

### TC-16 — README fallback-stubs pattern

- **Given** the "no portunix on PATH" fallback shown in `bash/README.md`
- **When** the stubs are pasted into a script and exercised
- **Then** they behave as drop-in no-ops; `ptx_trace_exec` correctly strips operation/flags/`--` and runs the command
- **Result**: ✅ PASS (fix applied during testing to both `README.md` and `README.cs.md`)

### TC-17 — Full ETL example end-to-end

- **When** executing `sdk/examples/bash/etl-pipeline.sh`
- **Then** session `customer-import-bash` is created with 5 events (`extract`, `normalize`, `validate-emails`, `validate-summary`, `load-to-db`); `trace stats` aggregates them with correct tags and per-operation timing
- **Result**: ✅ PASS

### TC-18 — Regression on existing CLI

- **When** running `start`/`event`/`end`/`sessions`/`export ai`/`alerts rules` on an unrelated session
- **Then** all behave as before
- **Result**: ✅ PASS

### TC-19 — Documentation links

- **When** all relative links in the new README files are resolved
- **Then** every linked file exists in the repository
- **Result**: ✅ PASS

## Test Summary

- Total scenarios: **19**
- Passed: **19** (all 3 issues found during testing have been fixed and re-verified)
- Failed: **0**

## Coverage

| Layer | Coverage notes |
| ----- | -------------- |
| CLI commands | `pipe` (text/binary, tag, format hint, ad-hoc, active-session) and `exec` (success, failure, missing binary, signal, stdin, stderr-capture) all exercised |
| Bash SDK | All 10 helper functions sourced; `start`/`end`/`event`/`pipe`/`exec`/`session_id` exercised end-to-end; env override exercised |
| Documentation | English & Czech Bash README, helper README, SDK index README — link-checked |
| Regression | `start`/`event`/`end`/`sessions`/`export ai`/`alerts rules` |
| Not covered (out of scope for #141) | `trace browse`, `trace replay`, `trace debug`, `trace reproduce`, `trace storage` (none implemented yet) |

No automated tests were added (the helper has no `*_test.go` files yet — pre-existing condition; tracked outside this issue).

## CI Notes

- All tests are reproducible via shell commands documented above.
- No new dependencies; nothing to add to CI.
- Recommend adding `go test ./src/helpers/ptx-trace/...` smoke-test job once unit tests exist.
- Recommend running `bash -n` on `sdk/bash/*.sh` and `sdk/examples/bash/*.sh` in CI as a quick lint.

## Findings

### Resolved follow-ups

1. **TC-15: `ptx_trace_session_id` rc** — ✅ **FIXED**
   - Previous behavior: rc=0 + empty stdout when no active session (contradicted docstring)
   - Fix: awk extractor now tracks a `found` flag and exits non-zero on no match
   - Verified: rc=1 + empty stdout when no session; rc=0 + SID when session active

2. **TC-10: Wrapper exit code on signal kill** — ✅ **FIXED**
   - Previous behavior: SIGTERM → wrapper returns rc=255 to the shell (Go's `os.Exit(-1)` truncates)
   - Fix: in `runWrappedCommand`, when `ProcessState.Signaled()` is true, return `128 + signum` (POSIX convention)
   - Verified: SIGTERM → rc=143 to shell; event also records `exit_code=143`

3. **TC-16: README fallback for `ptx_trace_exec` stub** — ✅ **FIXED**
   - Previous behavior: original stub did not strip the `--` argument, causing `--: command not found`
   - Fix: stub now consumes operation name + flags + `--` before exec'ing the rest
   - Updated in both `README.md` and `README.cs.md`

## Failure Injection Ideas (for future hardening)

- Disk full while writing NDJSON chunks during `pipe` (currently no graceful handling — should at least not hang the pipe)
- Very large stdin (>1 GiB) through `pipe` — verify constant memory (current impl uses 32 KiB buffer in binary mode and line-by-line `bufio.ReadBytes` in text mode, expected O(line-length) memory)
- Child process that closes its stdin early in `exec`
- Signal storm (SIGINT spam) during a long `exec` — verify only one is forwarded and no goroutine leak

## Final Decision

**STATUS**: **PASS** (with 2 CONDITIONAL items recorded above as non-blocking follow-ups; 1 issue fixed during testing)

**Approval for merge**: **YES** — all primary acceptance criteria of Phase 5 / Bash integration met; the two CONDITIONAL items are minor and can be tracked as separate small follow-ups inside Issue #141 or a successor ticket.

**Date**: 2026-05-06
**Tester signature**: zdendaku (Tester role)
