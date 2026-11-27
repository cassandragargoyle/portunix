# Acceptance Protocol - Issue #261 (Plugin Install: Opt-in Wheel Extras)

**Issue**: Plugin install: opt-in wheel extras (--extra) for Python wheel plugins
**Branch**: `feature/261-plugin-wheel-extras` (base `c60aab2`, tested from the
uncommitted working tree of the branch)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-11-22
**Testing OS**: Unit tests and `plugin validate` on the Linux host — Ubuntu
26.04 (amd64, physical). All installations in a Podman container —
`ubuntu:24.04` (Python 3.12.3 system, Python 3.14 uv-managed)

## Scope Note

The change adds `plugin.optional_extras` to the manifest and the flag `--extra`
(repeatable or comma separated) to `plugin install` and `plugin install-github`.
The main wheel is installed as `<wheel>[extra,...]` with uv or with the pip
fallback. The registry records the declared and the installed extras,
`plugin repair` keeps the installed extras, `plugin install --force` installs
only the given extras and warns about dropped ones, and `plugin info` lists the
extras.

E2E tests used a test wheel `ptx_extdemo-0.1.0` built for this protocol, with the
extras `small` (pulls `six`) and `other` (pulls `idna`) and an entry point that
reports whether `six` is importable. The presence of a package was checked with
`import` in the plugin venv, the recorded extras in `~/.portunix/plugins/registry.json`.

## Test Summary

- Total test scenarios: 41 (10 unit + 3 build + 4 validate + 14 E2E pip +
  7 E2E uv + 3 regression)
- Passed: 41
- Failed: 0
- Skipped: 0

## Test Results

### Unit Tests and Build

| # | Scenario | Result |
| - | -------- | ------ |
| U1 | `TestValidateManifest_OptionalExtras` (9 cases: valid, without wheel, empty name, leading dash, bracket, space, duplicate, normalized duplicate) | ✅ PASS |
| U2 | `TestResolveWheelExtras` (trim, normalized match, dedupe, sort, undeclared, none declared) | ✅ PASS |
| U3 | `TestWheelRequirement` | ✅ PASS |
| U4 | `TestDroppedExtras` | ✅ PASS |
| U5 | `TestInstallWithExtrasPip` (comma separated + repeated value) | ✅ PASS |
| U6 | `TestInstallWithoutExtrasPip` | ✅ PASS |
| U7 | `TestInstallWithExtrasUV` | ✅ PASS |
| U8 | `TestInstallUnknownExtraInstallsNothing` | ✅ PASS |
| U9 | `TestForceInstallUnknownExtraKeepsOldInstall` | ✅ PASS |
| U10 | `TestRepairKeepsInstalledExtras` | ✅ PASS |
| B1 | `go test portunix.ai/app/plugins/... portunix.ai/cmd/...` (full suites) | ✅ PASS |
| B2 | `go vet`, `gofmt -l` on changed packages | ✅ PASS |
| B3 | `make build` (main binary + helpers) | ✅ PASS |

### Manifest Validation (AC1)

| # | Manifest | Expected | Result |
| - | -------- | -------- | ------ |
| V1 | Valid `optional_extras` | accepted | ✅ PASS |
| V2 | Name `bad name` | rejected, regex in message | ✅ PASS |
| V3 | Names `small` + `Small` | rejected as duplicate | ✅ PASS |
| V4 | `optional_extras` without `wheel` | rejected | ✅ PASS |

### E2E — pip Fallback (no uv in the container)

| # | Command | Expected | Result |
| - | ------- | -------- | ------ |
| P1 | `install --extra cuda` | error lists `small, other`; no plugin dir, not registered (AC4) | ✅ PASS |
| P2 | `install` | `six`, `idna` absent; `installed_extras` empty (AC3) | ✅ PASS |
| P3 | `info` | both extras `[ ]`, `Installed: none` with hint | ✅ PASS |
| P4 | `extdemo run` | `six missing` | ✅ PASS |
| P5 | `install --force --extra small` | `six` present, `idna` absent; registry `[small]` (AC2) | ✅ PASS |
| P6 | `extdemo run` | `six available` | ✅ PASS |
| P7 | `info` | `[x] small`, `[ ] other` | ✅ PASS |
| P8 | `install --force --extra small,other` | both present; registry `[other, small]` | ✅ PASS |
| P9 | `install --force --extra other --extra small` | same as P8 | ✅ PASS |
| P10 | `install --force` (no `--extra`) | warning names dropped `other, small`; both absent | ✅ PASS |
| P11 | `install --force --extra small` | `six` present again | ✅ PASS |
| P12 | `repair extdemo` | reinstalls with `(extras: small)`; registry kept | ✅ PASS |
| P13 | `install --force --extra cuda` | error; old install with `six` and `[small]` intact | ✅ PASS |
| P14 | `info` | `[x] small`, `Installed: small` | ✅ PASS |

### E2E — uv Backend (uv installed with `portunix install uv`)

| # | Command | Expected | Result |
| - | ------- | -------- | ------ |
| X1 | `install --force --extra small` | uv venv, `six` present (AC2) | ✅ PASS |
| X2 | `extdemo run` | `six available` | ✅ PASS |
| X3 | `install --force --extra small,other` | both present | ✅ PASS |
| X4 | `repair extdemo` | keeps `other, small` | ✅ PASS |
| X5 | `info` | both `[x]` | ✅ PASS |
| X6 | `uninstall -f extdemo` | removed | ✅ PASS |
| X7 | `install` (fresh, no `--extra`) | both absent, no extras recorded (AC3) | ✅ PASS |

### Regression

| # | Scenario | Expected | Result |
| - | -------- | -------- | ------ |
| R1 | Plugin without `optional_extras`, `install --extra small` | error "declares no optional extras" | ✅ PASS |
| R2 | Same plugin, `install` | installs as before | ✅ PASS |
| R3 | `plugin install-github --help` | shows `--extra` | ✅ PASS |

## Observations (non-blocking)

- `plugin install --help` shows only the flag list, not the long description
  with the `--extra` example; this is the existing behavior of the basic help
  level for all commands
- The warning about a missing `plugin.python_version` in the uv runs comes from
  the test fixture, not from this change
- Adding an extra to an installed plugin without a reinstall (optional item 4 of
  the issue) is not implemented; `plugin install --force --extra <name>` covers it
- The branch must be committed before the merge; this protocol covers the
  working tree as tested

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES (after the tested changes are committed unchanged)
**Date**: 2026-11-22
**Tester signature**: Claude Code (tester role)
