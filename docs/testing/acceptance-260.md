# Acceptance Protocol - Issue #260 (Plugin Runtime node for JavaScript Plugins)

**Issue**: Plugin runtime node for JavaScript plugins
**Branch**: `feature/260-plugin-runtime-node` (commit `553e61b`, test round 2)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-23
**Testing OS**: Unit tests, template tests, `plugin validate` and Windows
cross-compilation on the Linux host — Ubuntu 26.04 (amd64, physical). All
installations in fresh Podman 5.7.0 containers — `ubuntu:24.04` (Ubuntu 24.04.5
LTS): `r260-node` with Node.js v18.19.1 / npm 9.2.0 installed by
`portunix install nodejs`, and `r260-nonode` without Node.js

## Scope Note

The branch adds the plugin runtime `node` (commit `177c149`) and fixes the findings
of test round 1 (commit `553e61b`):

1. The JavaScript template implements `portunix.plugins.PluginService` (proto copy
   of the Portunix contract, handlers for `Initialize`, `GetInfo`, `Health`,
   `Execute`, `ListCommands`, `Shutdown`, server bound to `127.0.0.1`); a guard test
   keeps the template proto equal to the contract
2. Version-aware Node.js upgrade hint (Windows `--variant=latest`, Linux points to
   nodejs.org / nvm because the distribution package cannot be pinned)
3. node/npm are checked before the prerequisite warning on install and reinstall
4. A missing plugin interpreter (`java`, `python3`, `node`) makes `plugin health`
   report the helper as unhealthy (`runtime_missing`) with an install hint
5. The note `existing installation left unchanged` comes before the npm output

Round 2 retested everything from round 1 on the new commit plus the fixes. E2E used
the JavaScript template and the fixture plugins of round 1: `fx-node-nodeps` (node,
no dependencies, prints args, `PORTUNIX_CWD` and working directory),
`fx-node-badver` (`runtime_version: >=99`), `fx-badruntime` (`runtime: nodejs`),
`fx-native` (shell script) and `fx-python` (Python script). node/npm were hidden or
replaced by a failing npm stub through `PATH` for the failure cases.

## Test Summary

- Total test scenarios: 70 (22 unit + 5 build + 3 validate + 9 service lifecycle +
  11 E2E with Node.js + 9 finding fixes + 5 missing Node.js + 1 install robustness +
  5 regression)
- Passed: 70
- Failed: 0
- Skipped: 0

## Test Results

### Unit Tests and Build

| # | Scenario | Result |
| - | -------- | ------ |
| U1 | `TestValidateManifest_Runtimes` (empty, native, java, python, node, node + `>=18`, jvm, nodejs) | ✅ PASS |
| U2 | `TestJavaScriptTemplateManifest` | ✅ PASS |
| U3 | `TestParseNodeVersion` | ✅ PASS |
| U4 | `TestCheckNodeRuntime` (missing, no constraint, satisfied, too old, unparsable) | ✅ PASS |
| U5 | `TestNodeDependenciesMissing` | ✅ PASS |
| U6 | `TestPlanNodeDependencies` | ✅ PASS |
| U7 | `TestSetupNodeDependencies` | ✅ PASS |
| U8 | `TestSetupNodeDependencies_Failure` | ✅ PASS |
| U9 | `TestCheckInstall_Node` | ✅ PASS |
| U10 | `TestBinaryPath_Node` | ✅ PASS |
| U11 | `TestBuildPluginCommand_AllRuntimes` (incl. node) | ✅ PASS |
| U12 | `TestBuildPluginCommand_Node` | ✅ PASS |
| U13 | `TestCheckInstall` (Python venv, regression) | ✅ PASS |
| U14 | `TestBuildPluginCommand_StubPluginReceives*` (regression) | ✅ PASS |
| U15 | `TestSuggestNodeUpgrade` (Windows, Linux) | ✅ PASS |
| U16 | `TestRuntimeInstallFix` (java, python, node, native, empty) | ✅ PASS |
| U17 | `TestInstallPlugin_NodeMissing` (hard error, no "will be installed" warning, nothing copied) | ✅ PASS |
| U18 | `TestForceInstallPlugin_NodeNPMFailureKeepsOld` (message order, old install kept) | ✅ PASS |
| U19 | `TestCheckHelperPluginHealth_NodeRuntime` (runtime missing / present) | ✅ PASS |
| U20 | `TestJavaScriptTemplateProtoMatchesContract` | ✅ PASS |
| U21 | Guard test with injected drift (`rpc Health` → `rpc HealthCheck` in the template proto) fails, file restored | ✅ PASS |
| U22 | Template tests `node --test test/` (8 tests, incl. Initialize, Execute, Health, Shutdown handlers) | ✅ PASS |
| B1 | `go test -count=1 portunix.ai/app/plugins/... portunix.ai/cmd` (full suites) | ✅ PASS |
| B2 | `go vet`, `gofmt -l` on changed packages | ✅ PASS |
| B3 | `make build` (main binary + helpers) | ✅ PASS |
| B4 | `GOOS=windows go build` (main binary) | ✅ PASS |
| B5 | `GOOS=windows go vet` (`app/plugins/manager`, `cmd`) | ✅ PASS |

### Manifest Validation (AC1)

| # | Manifest | Expected | Result |
| - | -------- | -------- | ------ |
| V1 | JavaScript template | accepted, exit 0 | ✅ PASS |
| V2 | `runtime: nodejs` | rejected, exit 1, message lists `native, java, python, node` | ✅ PASS |
| V3 | Template in the container without Node.js | accepted | ✅ PASS |

### Template Service Lifecycle (finding 1) — `r260-node`

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| S1 | `plugin start js-plugin-template` | started, no `Unimplemented` error | ✅ PASS |
| S2 | Listening socket | `127.0.0.1:9001` only (`/proc/net/tcp` `0100007F:2329`) | ✅ PASS |
| S3 | `plugin health` | healthy, message `ok`, uptime reported | ✅ PASS |
| S4 | `plugin list` / `plugin info` | `running`; commands from `plugin.json` (`hello: Print a greeting`) | ✅ PASS |
| S5 | `portunix js-plugin-template hello Service` while the service runs | `Hello, Service!` | ✅ PASS |
| S6 | `plugin stop` | success, exit 0, no plugin process left, log shows cleanup + graceful shutdown | ✅ PASS |
| S7 | Second start/stop cycle | both succeed, status `stopped` | ✅ PASS |
| S8 | Failure injection: port 9001 held by another process, `plugin start` | `port 9001 is already in use by another process`, exit 1, no process left, status `failed` | ✅ PASS |
| S9 | Port freed, `plugin start` again | started and healthy, then stopped | ✅ PASS |

### E2E with Node.js — `r260-node` (AC2)

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| E1 | `plugin install` of the template (no `node_modules`, no lockfile) | `npm install`, `node_modules` + `package-lock.json` in install dir | ✅ PASS |
| E2 | `portunix js-plugin-template hello Portunix` from `/tmp` | `Hello, Portunix!`, exit 0 | ✅ PASS |
| E3 | `hello` without argument | `Hello, World!`, exit 0 | ✅ PASS |
| E4 | Unknown command `nope` | `Error: unknown command: nope`, exit 1 | ✅ PASS |
| E5 | `plugin check` | `node >=18 ✅ Found: 18.19.1`, installation ok | ✅ PASS |
| E6 | `fx-node-nodeps` install | npm not run, no `node_modules` | ✅ PASS |
| E7 | `fx-node-nodeps hello x` from `/tmp/caller` | args passed, `PORTUNIX_CWD=/tmp/caller`, working dir = install dir | ✅ PASS |
| E8 | Arguments `"a b"`, `"--x=1 2"` | received as single arguments | ✅ PASS |
| E9 | `node_modules` removed, run plugin | `is broken: Node.js dependencies not installed`, `plugin list` shows broken | ✅ PASS |
| E10 | `plugin repair` | `npm ci` (lockfile present), plugin runs again | ✅ PASS |
| E11 | `plugin uninstall -f` | install directory removed | ✅ PASS |

### Finding Fixes 2–5 — `r260-node`

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| F2 | `fx-node-badver` (`>=99`), `plugin check` | `18.19.1 does not satisfy >=99`; Linux hint to nodejs.org / nvm with the distribution note | ✅ PASS |
| F3 | Template copy installed with node hidden | only the hard error `Node.js runtime not found ... portunix install nodejs`, no "will be installed" warning, nothing left | ✅ PASS |
| F3b | `install --force` with npm hidden | `failed to setup Node.js plugin (existing installation left unchanged): npm not found ...`; old install still runs | ✅ PASS |
| F4 | `plugin health` of a node helper with node hidden | `❌ Unhealthy`, `Runtime not found: node. Fix: portunix install nodejs` | ✅ PASS |
| F4b | Same with node present | healthy | ✅ PASS |
| F4c | Python script helper with python3 hidden | `❌ Unhealthy`, `Runtime not found: python3. Fix: portunix install python` | ✅ PASS |
| F4d | Same with python3 present | healthy (helper mode) | ✅ PASS |
| F4e | Native helper with interpreters hidden | healthy (unaffected) | ✅ PASS |
| F5 | `install --force` with a failing npm stub | note before npm output, npm output in the error, install dir identical, no staging/backup leftovers, plugin still runs | ✅ PASS |

### Missing Node.js (AC3)

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| M1 | `r260-nonode`: install template | only the hard error with `portunix install nodejs`, exit 1, nothing left | ✅ PASS |
| M2 | `r260-nonode`: install `fx-node-nodeps`, run | installs with the prerequisite warning; run fails `needs the Node.js runtime ... portunix install nodejs` | ✅ PASS |
| M3 | `r260-nonode`: `plugin check` | `❌ Node.js runtime not found`, fix `portunix install nodejs` | ✅ PASS |
| M4 | `r260-nonode`: `plugin health` | `❌ Unhealthy`, `Runtime not found: node` | ✅ PASS |
| M5 | `r260-node`: node removed from `PATH` after install, run plugin | clear runtime error with install hint, exit 1 | ✅ PASS |

### Install Robustness

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| I1 | Template with shipped `node_modules`, npm hidden | installs without npm, `hello Vendored` works | ✅ PASS |

### Regression (AC4)

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| R1 | `fx-native` run | args, `PORTUNIX_CWD`, working dir as before | ✅ PASS |
| R2 | `fx-python` run + check | args, `PORTUNIX_CWD`, working dir as before; Python 3.12.3 found | ✅ PASS |
| R3 | Java template `plugin validate` | accepted | ✅ PASS |
| R4 | Java / Python / wheel unit suites (B1) | pass | ✅ PASS |
| R5 | `plugin list` with mixed runtimes | fixtures `ready`, template `stopped` | ✅ PASS |

## Round 1 Summary

Round 1 (2026-11-22, commit `177c149`) passed 47 scenarios with five non-blocking
findings: template service mode failed with `Unimplemented .../Initialize`, the
upgrade hint repeated the same install, a "will be installed" warning preceded the
hard install error, `plugin health` reported a helper as healthy without its
interpreter, and the "left unchanged" note followed the npm output. All five are
fixed in `553e61b` and verified above (S1–S9, F2–F5).

## Observations (non-blocking)

- **Out of scope, existing behavior**: `plugin health` reports a Python script
  helper (`runtime: python` without `wheel`) as `Binary is not executable` when the
  script has no executable bit, although it is run by `python3`. The check was
  already in place before this branch (`642814c`); #260 exempts only `node`
  scripts. Candidate for a separate fix (also for `java` JARs)
- The Go, Java and Rust templates have the same contract defect as the old
  JavaScript template; tracked in #262
- Windows was verified by cross-compilation and `go vet` only; an E2E run on
  Windows (`npm.cmd` resolution, `--variant=latest` hint) is left to the Windows tester

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-11-23
**Tester signature**: Claude Code (tester role)
