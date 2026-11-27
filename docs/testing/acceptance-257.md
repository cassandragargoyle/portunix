# Acceptance Protocol - Issue #257 (Plugin Commands Print Text Help for --help-ai Instead of JSON)

**Issue**: Plugin commands print text help for --help-ai instead of JSON
**Branch**: `fix/257-plugin-help-ai-json` (commit `41f7658`, re-tested at
`0fbf444` — see [Re-test](#re-test-of-0fbf444))
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-20
**Testing OS**: Linux host — Ubuntu 26.04 (amd64, physical). Nothing was
installed: `--help-ai` only reads the registry and starts already installed
plugins

## Scope Note

The fix answers `portunix <plugin> [command] --help-ai` by asking the plugin
first (same arguments, stdin closed, `PORTUNIX_HELP_AI=1`). A JSON object on
stdout with exit code 0 is printed unchanged. Otherwise Portunix prints JSON
built from the registry entry (`"source": "registry"`). A plugin that does not
exit within 10 s is killed and the call fails without fallback. In addition,
`plugin validate` warns about placeholder commands, and the templates and the
plugin guide follow the new `commands` rule.

Failure injection ran against stub plugins in an isolated registry
(`HOME` pointed to a scratch directory), so the real plugin registry was not
touched. Real plugins were tested with the real registry.

## Test Summary

- Total test scenarios: 43 (10 unit + 3 build + 19 E2E stubs + 11 E2E real)
- Passed: 43
- Failed: 0
- Skipped: 1 (Windows, see W1)
- Findings: 1 in scope (nit), 1 known limitation, 4 outside the issue scope

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | JSON fallback, plugin answer, fallback cases, not started, timeout, guard, placeholder detection | host, `go test -race -count=1` |
| Build | `gofmt`, `go vet`, `make build` | host |
| E2E — stubs | Plugin answers, failure injection, timeout, normal execution | host, isolated registry |
| E2E — real | All 26 installed plugins, command level, helper regressions, `plugin validate` | host, real registry |

## Test Cases

### Unit and Build

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| U1 | `TestGeneratePluginAIHelp`, `TestGeneratePluginAIHelp_NoCommands` — fallback JSON fields, parameters, `[]` instead of `null` | ✅ PASS |
| U2 | `TestRunPluginHelpAI_PluginAnswers` — plugin JSON printed, args and `PORTUNIX_HELP_AI=1` passed | ✅ PASS |
| U3 | `TestRunPluginHelpAI_Fallback` — unknown option, text, JSON with exit 1, JSON array, no output → registry | ✅ PASS |
| U4 | `TestRunPluginHelpAI_NotStarted` — disabled and missing plugin are not started | ✅ PASS |
| U5 | `TestRunPluginHelpAI_Timeout` — error, no output, process killed | ✅ PASS |
| U6 | `TestGuardHelpLevelFlags_SelfHandledHelpAI`, `TestPluginCommandHandlesHelpAI` | ✅ PASS |
| U7 | `TestCommandWarnings_Placeholder`, `TestCreateDefaultManifest_NoPlaceholderCommand` | ✅ PASS |
| U8 | Existing guard tests #203 (`TestGuardHelpLevelFlags_*`) | ✅ PASS |
| U9 | Full packages `portunix.ai/cmd`, `portunix.ai/app/plugins/...` with `-race -count=1` | ✅ PASS |
| U10 | No data race reported | ✅ PASS |
| B1 | `gofmt -l src/cmd src/app/plugins` — no files | ✅ PASS |
| B2 | `go vet portunix.ai/cmd/ portunix.ai/app/plugins/...` | ✅ PASS |
| B3 | `make build` — main binary and all helpers | ✅ PASS |

### E2E — Stub Plugins (Isolated Registry)

| ID | Command | Observed | Result |
| -- | ------- | -------- | ------ |
| P1 | `json --help-ai` (stub prints JSON) | Plugin JSON, `PORTUNIX_HELP_AI=1`, `PORTUNIX_CWD` set, rc 0 | ✅ PASS (AC2) |
| P2 | `json run --verbose --help-ai` | Plugin got `run --verbose --help-ai`, its JSON printed, rc 0 | ✅ PASS (AC3) |
| P3 | `json --help-ai=true` | Treated as `--help-ai`, plugin JSON | ✅ PASS |
| P4 | `noisy --help-ai` (JSON + stderr warning) | JSON printed, plugin stderr not shown | ✅ PASS |
| P5 | `stdin --help-ai` (stub reads stdin) | stdin is EOF, no hang, JSON printed | ✅ PASS |
| P6 | `ws --help-ai` (blank lines around JSON) | JSON accepted, trimmed | ✅ PASS |
| P7 | `bigjson --help-ai` (2 MB JSON) | 2 000 033 B printed unchanged, 50 ms | ✅ PASS |
| F1 | `reject --help-ai` (`No such option`, exit 2) | Registry JSON, rc 0, no plugin stderr | ✅ PASS (AC1) |
| F2 | `killed --help-ai` (`kill -9` itself) | Registry JSON, rc 0 | ✅ PASS |
| F3 | `disabled --help-ai` | Registry JSON, plugin not started (no marker file) | ✅ PASS |
| F4 | `missing --help-ai` (install dir missing) | Registry JSON, rc 0 | ✅ PASS |
| F5 | `nocmds --help-ai` (`commands: null`) | `"commands": []` | ✅ PASS |
| T1 | `hang --help-ai` (`exec sleep 60`) | After 10.0 s: `Error: plugin 'hang' did not answer --help-ai within 10s`, rc 1, no stdout | ✅ PASS (AC4) |
| T2 | `hangwrap --help-ai` (Python `.sh` wrapper, child `sleep 61`) | After 11.0 s same error, rc 1; the child `sleep` stays running (L1) | ✅ PASS (AC4) |
| N1 | `runner run x` | Plugin runs normally: `RAN: run x` | ✅ PASS |
| N2 | `runner --help-ai=false` | Plugin runs normally with the argument | ✅ PASS |
| N3 | `runner -- --help-ai` | Not intercepted, plugin gets `-- --help-ai` | ✅ PASS |
| N4 | `json --help-expert` | Text usage, plugin not started | ✅ PASS (F4) |
| N5 | `json --help-expert --help-ai` | `--help-ai` wins, plugin JSON (same priority as root help) | ✅ PASS |

### E2E — Real Plugins and Regressions

| ID | Command | Observed | Result |
| -- | ------- | -------- | ------ |
| E1 | `portunix <plugin> --help-ai` for all 26 installed plugins | 26/26 rc 0 and a JSON object; 5 plugin answers (docgen, fulltext, modeler, reco-studio, text-extractor), 21 registry fallbacks incl. `scraper`; no stderr; max 822 ms | ✅ PASS (AC1, AC2) |
| E2 | `docgen --help-ai` vs. `ptx-docgen --help-ai` run directly | Byte-identical (1749 B) | ✅ PASS (AC2) |
| E3 | `text-extractor convert`, `query search`, `modeler catalog`, `docgen build` with `--help-ai` | Valid JSON each (modeler: plugin, others: registry) | ✅ PASS (AC3) |
| R1 | `portunix --help-ai` | Root JSON unchanged | ✅ PASS |
| R2 | `portunix docker install --help-ai` | Help shown, nothing installed | ✅ PASS (AC5) |
| R3 | `portunix sandbox generate --help-ai` in an empty directory | Help shown, no `.wsb` created | ✅ PASS (AC5) |
| R4 | `portunix container`, `mcp`, `github` with `--help-ai` | Output identical to the deployed `v2.7.0+dev.3` (`container` md5 match) | ✅ PASS (AC5) |
| V1 | `plugin validate` — `m365-extractor` with only command `m365` | Validation succeeds with placeholder warning | ✅ PASS (AC7) |
| V2 | `plugin validate` — `scraper` with `fetch`; with `[]` | No warning | ✅ PASS (AC7) |
| V3 | `plugin validate` on the 4 language templates | No placeholder warning (java, javascript fail for another reason, F3) | ✅ PASS (AC6) |
| V4 | `plugin validate` on all installed plugin directories | 18 warnings — matches the 18 placeholder manifests named in the issue | ✅ PASS (AC7) |

## Acceptance Criteria

- [x] AC1 — `portunix scraper --help-ai` prints valid JSON from the manifest,
  exit code 0 (E1, F1)
- [x] AC2 — a plugin answering with JSON gets it printed unchanged (P1, E2)
- [x] AC3 — `<plugin> <command> --help-ai` prints valid JSON by the same rules
  (P2, E3)
- [x] AC4 — a hanging plugin is killed, timeout error, non-zero exit, no
  fallback (T1, T2, U5)
- [x] AC5 — `--help-ai` on helper commands unchanged (R2–R4, U8)
- [x] AC6 — plugin guide describes the `commands` rule and the `--help-ai`
  contract; templates follow it (V3, review of `getting-started.md`)
- [x] AC7 — `plugin validate` warns about a placeholder command (V1, V2, V4)
- [x] AC8 — unit tests cover plugin answer, fallback, timeout and validation
  warning (U1–U7)

## Findings

### In Scope

- **N1 (nit)** — the fallback JSON escapes `<` and `>` in `usage` as
  `<` / `>` (`json.MarshalIndent` HTML escaping). The JSON is valid
  and decodes correctly, and the root `--help-ai` behaves the same. Optional:
  encode with `SetEscapeHTML(false)` for readable raw output

### Known Limitation

- **L1** — after a timeout only the direct plugin process is killed. With a
  wrapper (Python `.sh`), the child process keeps running (T2: `sleep 61`
  reparented and had to be killed by hand). Output is not blocked
  (`WaitDelay` 1 s). Documented by the developer; a process-group kill would
  need OS-specific code. Recommended as a follow-up issue

### Outside the Issue Scope

- **F3 (pre-existing)** — `plugin validate` fails on the Java and JavaScript
  templates: `unsupported runtime: jvm` / `node`. The `runtime` values are
  unchanged by this branch (`git show main:` confirms `"runtime": "jvm"`)
- **F4 (pre-existing)** — `<plugin> --help-expert` prints the default Cobra
  usage to stderr instead of an expert help on stdout (`multiLevelHelp` uses
  `cmd.Usage()` for subcommands). Not part of #257
- **F5 (plugin side)** — `text-extractor convert --help-ai` gets the registry
  fallback: the plugin answers `--help-ai` only without a subcommand. The list
  of supported extensions (portunix-vscode issue 092) needs a plugin change
- **F6 (plugin side)** — 18 installed manifests have placeholder commands,
  among them docgen, modeler and reco-studio, which answer `--help-ai`
  themselves. To be fixed in the plugin repositories, as stated in the issue

### Skipped

- **W1** — Windows not tested (generic tester on a Linux host). The PowerShell
  stub branches of the unit tests and `Process.Kill` on Windows were not run.
  Recommended: a `tester-windows` run of `go test portunix.ai/cmd/ -run HelpAI`
  and E1 on a Windows host before the next release

## Re-test of 0fbf444

Commit `0fbf444` fixes N1, L1 and the Java part of F3: the plugin runs in its
own process group (Unix) and the whole group is killed on timeout, Ctrl+C or
SIGTERM (`taskkill /T` on Windows), the fallback JSON is encoded without HTML
escaping, and the Java template uses `runtime: java`. Same host, same
isolated stub registry, three new stubs (`trapterm`, `escape`, `escjson`).

| ID | Given / When / Then | Observed | Result |
| -- | ------------------- | -------- | ------ |
| RT1 | Unit and build: `gofmt`, `go vet` (Linux and `GOOS=windows`), `go test -race -count=1` for `cmd` and `plugins/...`, timeout tests `-count=3`, `make build`, Windows cross-build of `portunix.ai/cmd` | All green, incl. new `TestRunPluginHelpAI_TimeoutKillsChildren` | ✅ PASS |
| RT2 | `hang --help-ai` (`exec sleep 60`) | Error after 10.0 s, rc 1, nothing left | ✅ PASS (AC4) |
| RT3 | `hangwrap --help-ai` (wrapper with child `sleep 61`) | Error after 10.0 s, rc 1, **child killed** (was left running before, L1) | ✅ PASS (AC4, L1) |
| RT4 | `trapterm --help-ai` (ignores SIGTERM/SIGINT, has a child) | Error after 10.0 s, nothing left (SIGKILL to the group) | ✅ PASS |
| RT5 | `escape --help-ai` (child started with `setsid`) | Error after 10.0 s; the `setsid` child survives — it left the process group on purpose (L2) | ✅ PASS |
| RT6 | SIGINT to portunix during `hang` / `hangwrap` / `trapterm` | `Error: plugin '<name>' --help-ai interrupted by interrupt`, rc 1, nothing left | ✅ PASS |
| RT7 | SIGTERM to portunix during `hangwrap` | `... interrupted by terminated`, rc 1, nothing left | ✅ PASS |
| RT8 | SIGINT to the whole foreground group (terminal Ctrl+C) during `hangwrap`, 5 runs | Each run: portunix exits after 25–53 ms, child gone, nothing left | ✅ PASS |
| RT9 | `reject --help-ai` (registry fallback) | `"usage": "portunix reject <command> [args...]"`, no `<` | ✅ PASS (N1) |
| RT10 | `escjson --help-ai` (plugin JSON contains `<`) | Printed byte for byte, not re-encoded | ✅ PASS (AC2) |
| RT11 | `json run --help-ai`, `runner run x` | Plugin answer and normal execution unchanged | ✅ PASS |
| RT12 | E1 again: all 26 installed plugins | 26/26 JSON, rc 0, 5 plugin answers; no `<` in any fallback | ✅ PASS (AC1) |
| RT13 | E2 again: `docgen --help-ai` vs. `ptx-docgen --help-ai` | Byte-identical | ✅ PASS |
| RT14 | `plugin validate` on the templates | go, java, rust pass; javascript fails with `unsupported runtime: node` (tracked in #260) | ✅ PASS (F3 Java) |
| RT15 | Regressions R1–R4: root `--help-ai`, `docker install`, `sandbox generate` (no `.wsb`), `mcp --help-ai` | Unchanged | ✅ PASS (AC5) |

Notes:

- **L2 (accepted)** — a process that deliberately leaves the process group
  (`setsid`, daemonizing) is not killed on timeout. This is expected for a
  process-group kill and out of the reach of the CLI
- In the first group-SIGINT run a process seemed to remain. Timing showed it
  was a measurement artifact (`setsid` without `--wait` returns at once, the
  check ran before the signal was sent); RT8 measured the real timeline

### Finding Status After Re-test

| Finding | Status |
| ------- | ------ |
| N1 — HTML escaping in fallback JSON | Fixed in `0fbf444` (RT9) |
| L1 — wrapper children survive timeout | Fixed in `0fbf444` (RT3, RT6–RT8) |
| F3 — Java/JS templates fail validation | Java fixed in `0fbf444` (RT14); JavaScript needs runtime `node` → #260 |
| F4 — `--help-expert` usage on stderr | Moved to #259 (consistent help levels) |
| F5, F6 — plugin side | Unchanged, plugin repositories |
| W1 — Windows not tested | Still open; `taskkill /T` path compiled and vetted only |

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES (re-confirmed for `0fbf444`)
**Date**: 2026-11-20
**Tester signature**: Claude Code (tester role) on behalf of zdendaku
