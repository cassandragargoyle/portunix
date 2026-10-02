# Acceptance Protocol - Issue #199 (Plugin Binary Path Missing `.exe` Suffix on Windows)

**Issue**: Plugin Binary Path Missing `.exe` Suffix on Windows
**Branch**: `fix/199-plugin-binary-exe-suffix` (commit `378fd002`)
**Tester**: zdendaku (QA/Test Engineer — generic)
**Date**: 2026-09-28
**Testing OS**:

- Windows 11 Pro 10.0.26200 (host) — Go 1.25.0 windows/amd64
- Ubuntu 22.04 (WSL2) — Linux-compiled test binaries and `portunix` (`GOOS=linux`)

## Scope Note

The fix changes how the plugin binary path is resolved (`platform.ExecutableName`,
`RegistryPlugin.BinaryPath()`) and is used by the health check, the gRPC plugin
config and the plugin dispatcher. No software is installed, so the container
policy does not apply. Linux behavior was verified by cross-compiling test
binaries and `portunix` for Linux and running them in WSL with a temporary
`HOME`. No Go toolchain was installed in WSL.

Out of scope per the issue: `dispatcher.binSuffix`, `selfinstall.mainBinaryName()`.

## Test Summary

- Total test scenarios: 12 (3 unit + 6 acceptance criteria + 3 regression)
- Passed: 12
- Failed: 0
- Conditional: 0

## Test Environment

- `make build` from branch tip: main binary + 19 helpers built (exit 0)
- `go vet` clean for `portunix.ai/app/plugins/manager`, `portunix.ai/cmd`,
  `src/pkg/platform` on `GOOS=windows` and `GOOS=linux`
- `gofmt` clean (after stripping CRLF of the Windows working copy)
- Baseline for comparison: deployed `portunix.exe` v2.5.0+dev.1 (pre-fix)
- Installed plugins on host cover every runtime path:

| Plugin | Runtime | Mode |
| ------ | ------- | ---- |
| `modeler`, `neuralink`, `crawler` | native | helper |
| `docgen` | native (`./ptx-docgen`) | service |
| `fulltext`, `text-extractor` | java | helper |
| `scraper` | python (wheel) | helper |

## Test Results

### Unit Tests

Windows (`go test`) and Linux (WSL, `GOOS=linux go test -c`):

- [x] `TestExecutableName` (4 cases) — pass on Windows and Linux
- [x] `TestRegistryPluginBinaryPath` (6 cases: native, empty runtime, explicit
  `.exe`, java, python script, python wheel) — pass on Windows and Linux
- [x] Full `manager` and `platform` test suites pass on Windows and Linux

### Functional — Acceptance Criteria

- [x] **AC1** — Given native plugin `modeler` installed on Windows, when
  `portunix plugin health modeler` runs, then status is `Healthy`.
  Baseline: `Unhealthy — Binary not found: ...\modeler\ptx-modeler`.
  Same fix confirmed for `neuralink` and `crawler`.
- [x] **AC2** — Given native plugin `fake` (`ptx-fake`) in a temporary `HOME` on
  Linux, when health runs, then status is `Healthy` and the path has no suffix.
  Non-executable binary → `Binary is not executable`, missing binary →
  `Binary not found` (both unchanged).
- [x] **AC3** — Java (`fulltext`, `text-extractor`) and Python wheel (`scraper`)
  plugins report `Healthy` exactly as the baseline; unit tests confirm no `.exe`
  on `.jar` / `.py` paths.
- [x] **AC4** — Given a manifest with `binary_name: ptx-fake.exe` on Windows,
  health reports `Healthy` and dispatch executes the binary (no `.exe.exe`).
- [x] **AC5** — Unit tests for `platform.ExecutableName` and
  `RegistryPlugin.BinaryPath()` exist and pass.
- [x] **AC6** — Dispatcher execution works: on Windows `portunix <plugin> --help`
  exits 0 for `modeler`, `neuralink`, `crawler`, `scraper`, `fulltext`,
  `text-extractor`; on Linux `portunix fake hello` runs the plugin with arguments.

### Regression

- [x] `make build` succeeds (main + 19 helpers, including `ptx-plugin-registry`
  with the new `replace portunix.ai/portunix => ../../..`)
- [x] Service-mode plugin `docgen` behaves identically to baseline (`not enabled`)
- [x] `scraper --help` exit code 120 only when output is truncated by a pipe;
  identical on baseline, exit 0 without a pipe — not a regression

## Observations / Recommendations (non-blocking)

1. `/d/portunix/` contains a Linux binary `portunix` next to `portunix.exe`.
   Git Bash resolves `portunix` to the non-executable Linux file, so the command
   fails silently there. Unrelated to #199; recommend removing the Linux binaries
   from the Windows deploy directory.
2. `ptx-plugin-registry/go.mod` now declares `go 1.25.0` (required by the root
   module) and pulls `golang.org/x/sys` v0.44.0. Consistent with other helpers.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-09-28
**Tester signature**: zdendaku
