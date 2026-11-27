# Acceptance Protocol - Issue #263 (Plugin Health Exec-Bit Check for Interpreter-Run Plugins)

**Issue**: plugin health reports interpreter-run plugin files as not executable
**Branch**: `fix/263-plugin-health-interpreter-exec-bit` (commit `2c86b56`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-23
**Testing OS**: Unit tests, build and Windows cross-compilation on the Linux host —
Ubuntu 26.04 (amd64, physical). E2E in a fresh Podman container `ubuntu:24.04`
(Ubuntu 24.04.5 LTS) `r263` with Python 3.12.3, OpenJDK 21.0.13 and Node.js
v18.19.1 (installed with `apt-get` as test preconditions)

## Scope Note

The fix adds `requiresExecBit` to `src/app/plugins/manager/manager.go`: the
executable bit is checked only for files executed directly (`native` binaries and
Python wheel entry points). `java`, `node` and Python script plugins are run by
their interpreter and are no longer checked.

E2E used five fixture helper plugins, all installed with `0644` files:
`fx-python` (Python script), `fx-java` (runnable JAR built with `javac --release 17`),
`fx-node` (JavaScript), `fx-native` (shell script) and `fx-wheel` (Python wheel with
a `console_scripts` entry point, built with `uv build`). As a baseline, the same
scenarios were run with a binary built from `main` (commit `14a2aa1`).

## Test Summary

- Total test scenarios: 25 (8 unit + 4 build + 13 E2E)
- Passed: 25
- Failed: 0
- Skipped: 0

## Test Results

### Unit Tests and Build

| # | Scenario | Result |
| - | -------- | ------ |
| U1 | `TestRequiresExecBit` (empty, native, python wheel, python script, java, node) | ✅ PASS |
| U2 | `TestCheckHelperPluginHealth_ExecBitPerRuntime` — native `0644` → `not_executable` | ✅ PASS |
| U3 | Same — wheel entry point `0644` → `not_executable` | ✅ PASS |
| U4 | Same — python script `0644` → `ready` | ✅ PASS |
| U5 | Same — java JAR `0644` → `ready` | ✅ PASS |
| U6 | Same — node script `0644` → `ready` | ✅ PASS |
| U7 | `TestCheckHelperPluginHealth_NodeRuntime` (regression from #260) | ✅ PASS |
| U8 | Failure injection: old condition (`plugin.Runtime != "node"`) put back → U4 and U5 fail; file restored | ✅ PASS |
| B1 | `go test -count=1 portunix.ai/app/plugins/... portunix.ai/cmd` (full suites) | ✅ PASS |
| B2 | `go vet`, `gofmt -l` on `app/plugins/manager` | ✅ PASS |
| B3 | `make build` (main binary + helpers) | ✅ PASS |
| B4 | `GOOS=windows go build` (main binary), `GOOS=windows go vet` (`app/plugins/manager`) | ✅ PASS |

### E2E — `ubuntu:24.04` (AC1–AC4)

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| E1 | `plugin install` of all five fixtures | installed; installer keeps `0644`, pip creates the wheel entry point `0755` | ✅ PASS |
| E2 | `plugin health fx-java` (JAR `0644`, java present) — AC1 | `✅ Healthy` | ✅ PASS |
| E3 | `plugin health fx-python` (script `0644`, python3 present) — AC2 | `✅ Healthy` | ✅ PASS |
| E4 | `plugin health fx-native` (binary `0644`) — AC3 | `❌ Unhealthy`, `Binary is not executable` | ✅ PASS |
| E5 | `plugin health fx-wheel`, entry point `0755` | `✅ Healthy` | ✅ PASS |
| E6 | `plugin health fx-wheel`, entry point changed to `0644` — AC3 | `❌ Unhealthy`, `Binary is not executable` | ✅ PASS |
| E7 | `plugin health fx-native` after `chmod 755` | `✅ Healthy` | ✅ PASS |
| E8 | `plugin health fx-node` (script `0644`) — AC4 | `✅ Healthy` (unchanged) | ✅ PASS |
| E9 | `portunix fx-python\|fx-java\|fx-node\|fx-wheel hello x` | plugins run and receive the arguments | ✅ PASS |
| E10 | `portunix fx-native hello x` with `0644` | `permission denied` — confirms `not_executable` is correct for native | ✅ PASS |
| E11 | `plugin health` of python, java and node fixtures with the interpreters removed from `PATH` | `❌ Unhealthy`, `Runtime not found: python3 / java / node` with an install hint (#260 unchanged) | ✅ PASS |
| E12 | Baseline: `main` binary, `plugin health fx-python` and `fx-java` | `❌ Unhealthy`, `Binary is not executable` — the bug from the issue reproduced | ✅ PASS |
| E13 | `plugin list` | all fixtures `ready` | ✅ PASS |

## Observations (non-blocking)

- `portunix plugin install` expects the plugin directory; passing the path to
  `plugin.json` fails with `open .../plugin.json/plugin.json: not a directory`. This
  is existing behavior, outside the scope of this issue; accepting a manifest path
  (or a clearer error) is a possible usability improvement
- Windows is not affected by the change (the exec-bit check is POSIX-only) and was
  verified by cross-compilation and `go vet` only

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-11-23
**Tester signature**: Claude Code (tester role)
