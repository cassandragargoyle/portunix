# Acceptance Protocol - Issues #236 and #237 (service start interruption, not-found message)

**Issues**:

- #236 service start: plugin process left running outside the state file when the CLI is
  killed during start
- #237 service start: "plugin X not found" repeated in the error message

**Branch**: `feature/issue-236-237-service-start` (commit `1238419`)
**Tester**: Claude Code in the tester role (Linux), on behalf of zdendaku
**Date**: 2026-10-24
**Testing OS**: Ubuntu 26.04, kernel 7.0.0-38, amd64 (host); Go 1.26.8

No software was installed. Two environments on the host:

- **Isolated home** (`HOME` set to a session-local scratch directory) with a fake native gRPC
  plugin `stubborn` — a shell script that ignores SIGINT and never serves gRPC
- **Real home** with the installed Python gRPC plugin `reco-studio`; the service state was
  empty before and after every case, so no other service instance was stopped

The binary was built with `make build` from commit `1238419` and run from the repository, not
from the deployed installation.

## Scope Note

The branch records a service instance as `starting` right after the spawn and before the
health wait, cancels the start on SIGTERM/SIGINT, stops leftover `starting` entries on the
next `service list`/`start`/`stop`, makes `killProcess` force-kill processes that are not
children of the CLI, and stops wrapping the registry "not found" error.

**Independence note**: the code was written in the developer role of the same Claude Code
session (the fix of `killProcess` during the developer self-check); the tests in this protocol
were designed and run in the tester role.

**Not covered**:

- Windows execution — the packages cross-compile for Windows and macOS, nothing was run
- A client that sends SIGTERM and then SIGKILL within the kill grace period
- SIGKILL of the CLI while it waits for the state lock
- The VS Code extension / Portunix Pilot as the calling client — simulated with `timeout`

## Test Summary

- Total test scenarios: 17 (B2 has 8 signal sub-cases, B4 has 2)
- Passed: 17
- Failed: 0
- Skipped: 3 (Windows, SIGTERM followed by SIGKILL, SIGKILL during the lock wait)
- Defects found in the scope: 0 (the `killProcess` defect was found and fixed during the
  developer self-check before this run, see Coverage)
- Observations: 5 (O1–O5)

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `app/service` (8 tests), `cmd` not-found test, all with `-race` | host |
| Static | `go vet`, `gofmt`, cross-compile for Windows and macOS | host |
| E2E | Normal start, join, info, release, stop with `reco-studio` | host, real home |
| Failure injection | SIGTERM, SIGINT, SIGKILL during start; plugin ignoring SIGINT; health timeout; start waiting for the lock | host, both homes |
| Regression | `portunix.ai/cmd` against `main` | host |

## Test Cases

### Suite A — `stubborn` plugin, isolated home

- **A1** Given a plugin that never becomes healthy and ignores SIGINT, when `service start`
  runs, then it fails with `health check timeout after 10s`, the process is gone and the
  state is empty
- **A2 / A3** Given the same plugin, when the CLI gets SIGTERM / SIGINT after 2 s, then it
  exits with `start interrupted`, the process is gone and the state is empty
- **A4** Given a CLI killed with SIGKILL during start, then the state holds a `starting`
  entry for the running process; when `service list` runs, then the process is stopped and
  the entry removed
- **A5 / A6** Given a CLI killed with SIGKILL during start, when `service stop --all --force`
  / `service stop stubborn --force` runs, then the process that ignores SIGINT is gone after
  the grace period
- **A7** Two starts killed with SIGKILL one after the other leave at most one recorded
  process and no unrecorded one
- **A8** `service start <unknown>` prints the reason once, in text and with `--output json`
- **A9** `plugin uninstall|enable|check|repair|info <unknown>` print the reason once

### Suite B — `reco-studio`, real home

- **B1** Start, join (second session), `service info --port`, release both sessions — the
  last release stops the instance
- **B2** SIGTERM and SIGINT after 0.05, 0.3, 0.6 and 0.9 s (start takes about 1 s), each from
  an empty state — no process and no entry remain
- **B4** Issue repro: SIGKILL after 0.3 s and 0.6 s, then a normal start — the leftover
  `starting` entry is stopped and the new instance gets port 50100 again (before the fix it
  got 50101)
- **B5** A second start (`--mode exclusive`) gets SIGTERM while the first start holds the
  lock — it exits with `start interrupted` once the lock is free, does not spawn, the first
  start succeeds
- **B6** `service start <unknown> --output json` in the real home prints the reason once

### Unit, static and regression

- **U1** `go test -race portunix.ai/app/service portunix.ai/app/plugins/...` and the `cmd`
  not-found test pass
- **U2** `go vet`, `gofmt -l` clean; `GOOS=windows` and `GOOS=darwin` builds pass
- **U3** `portunix.ai/cmd` on the branch and on `main`: the same tests fail, one new test
  passes

## Test Results

### Suite A

| Case | Result | Evidence |
| ---- | ------ | -------- |
| A1 | ✅ PASS | exit 1 after 14.0 s (10 s health timeout + 3 s grace + startup); no process, `{}` |
| A2 | ✅ PASS | exit 124 after 5.0 s (SIGTERM at 2 s + 3 s grace); no process, `{}` |
| A3 | ✅ PASS | exit 124; no process, `{}` |
| A4 | ✅ PASS | `{'stubborn:50100': 'starting'}` with a live process; `service list` 3.0 s, then no process, `{}` |
| A5 | ✅ PASS | `Stopped 1 instance(s)`; no process after the grace period |
| A6 | ✅ PASS | `Stopped 1 instance(s) of stubborn`; no process after the grace period |
| A7 | ✅ PASS | one recorded `starting` process, nothing unrecorded; port reuse not shown here because the second CLI was killed during the grace period of the first cleanup — covered by B4 |
| A8 | ✅ PASS | `Error: plugin nonexistent-plugin-xyz not found` (text and json) |
| A9 | ✅ PASS | e.g. `Error: failed to repair plugin: plugin nonexistent-plugin-xyz not found` — reason once for all five commands |

### Suite B

| Case | Result | Evidence |
| ---- | ------ | -------- |
| B1 | ✅ PASS | start 1.0 s; 2 sessions on one instance; `Session … released, instance stopped (was last session)` |
| B2 | ✅ PASS | 8 cases, all exit 124 with `start interrupted: context canceled`, `{}` and no process (second and third run, see O1) |
| B4 | ✅ PASS | `('starting', 68964, 0)` → new start `port 50100 pid 69138`, leftover gone; same for 0.6 s |
| B5 | ✅ PASS | second CLI exit 124 after 0.8 s, no second process; first start exit 0, one session |
| B6 | ✅ PASS | `Error: plugin nonexistent-plugin-xyz not found` |

The first two runs of suite B had script errors (missing `--session`/`--port`, and a 1.5 s
SIGTERM that arrived after a successful start); those runs joined a running instance and are
not counted. Results above are from the corrected runs.

### Unit, static and regression

| Check | Result |
| ----- | ------ |
| U1 `-race` unit tests | ✅ `app/service`, `app/plugins`, `app/plugins/manager`, `app/plugins/server`, `cmd` (not-found test) ok |
| U2 vet, gofmt, cross-compile | ✅ clean; Windows and macOS build |
| U3 regression `portunix.ai/cmd` | ✅ 6 failing tests on both, all `TestContainer*` (#225); `TestResolvePluginBinary_UnknownPluginMessage` new and passing |

## Acceptance Criteria

### #236

| Criterion | Result |
| --------- | ------ |
| After `timeout 0.05 portunix service start <plugin>` no plugin process runs outside `service list` | ✅ B2 (0.05–0.9 s, SIGTERM and SIGINT), A2, A3; see O1 |
| `portunix service stop --all --force` stops every process that `service start` spawned | ✅ A5, A6, B4, also for a plugin that ignores SIGINT |
| A normal start and a failed health check behave as before | ✅ B1, A1 |

### #237

| Criterion | Result |
| --------- | ------ |
| `portunix service start <unknown>` prints the reason once | ✅ A8, B6 |
| `--output json` and text output use the same message | ✅ A8 |

## Coverage

- `app/service` statement coverage 40.3 %: `StartService` 64 %, `cleanStaleEntries` 71 %,
  `killProcess` 70 %, `waitForHealthy` 63 %, `StopAll` 100 %; `ReleaseSession` and
  `StopPlugin` have no unit tests and are covered by B1 and A6 only
- The unit test `TestKillProcess_ForceKillsDetachedProcessIgnoringInterrupt` fails on the
  previous `killProcess` (process survives) and passes on the branch — verified during the
  developer self-check
- Gaps: join path cancelled after the health check of an existing instance, Windows
  `killProcess` path (`Signal(os.Interrupt)` fails, `Kill` directly)

## CI Notes

- `TestKillProcess_…` and `TestWaitForHealthy_ReturnsOnCancel` take 0.4 s and 2 s; the kill
  test skips on Windows
- No new test needs network access or a real plugin

## Observations

- **O1**: in the first corrected run of B2 one `reco-studio` process (PID 53192) was running
  outside the state file through all cases and disappeared on its own before it could be
  inspected. It was not reproduced in two further runs of the 16 signal cases nor in a
  dedicated repro (3 × SIGTERM after 0.05 s, watched for 60 s). It may have been started by
  another client on the host (the deployed Portunix v2.6.2 does not contain this fix); the
  origin is not proven. **Update 2026-10-24**: the developer was working on `ptx-reco-studio`
  in parallel at that time; a process started outside `portunix service start` is never
  recorded in the state file, so O1 is most likely unrelated to #236
- **O2**: for a plugin that ignores SIGINT the CLI needs up to the 3 s grace period to exit
  after SIGTERM (A2: 5.0 s for SIGTERM at 2 s) and a failed health check takes about 14 s
  (A1); a client that escalates to SIGKILL inside that window leaves a `starting` entry, which
  the next service command cleans up
- **O3**: SIGTERM is acted on only after the state lock is acquired (B5); a start waiting for
  the lock waits until the other start finishes, up to its health timeout
- **O4**: `service list` takes up to the grace period when it has to stop a leftover
  `starting` process that ignores SIGINT (A4: 3.0 s)
- **O5**: runtime errors of `service start` are followed by the full usage text — the same
  issue as point 3 of #234

## Recommendations

1. Run suites A and B once on Windows (tester-windows role)
2. If O1 shows up again, record `ps -o pid,ppid,lstart,args` and the listening port at once
3. Add unit tests for `ReleaseSession`, `StopPlugin` and the cancelled join path
4. Consider SIGTERM instead of SIGINT plus a shorter grace period for processes that never
   became healthy (O2, O4)
5. Cover the service commands in #234 point 3 (O5)
6. After the merge, deploy and repeat B2 with the VS Code extension as the client

## Final Decision

**STATUS**: CONDITIONAL

Linux: all 17 executed scenarios passed, every acceptance criterion of #236 and #237 is met,
no regression against `main`. The open item is execution on Windows; the process in O1 was
not reproducible and most likely came from parallel work on `ptx-reco-studio`.

**Approval for merge**: YES for the Linux scope; Windows to be confirmed by the tester-windows role
**Date**: 2026-10-24
**Tester signature**: Claude Code (tester role), on behalf of zdendaku
