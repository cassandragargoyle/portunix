# Acceptance Protocol - Issue #202 (Python 3.14.8, `latest` Variant and Variant-Level `installArgs`)

**Issue**: Python Package — Install Python 3.14.8 and Honour the Variant's `installArgs` on Windows
**Branch**: `feature/202-python-3-14-8-install-args` (first run: commit `5a3bff9`; re-test: commit `30dd711`)
**Tester**: Claude Code in the tester role (generic), on behalf of zdendaku
**Date**: 2026-10-06, re-test 2026-10-07
**Testing OS**:

- Windows 11 Pro 10.0.26200 (host) — unit tests, `go vet`, build, `--dry-run` comparisons only
- Windows 11 Enterprise 10.0.26100 AMD64 (Windows Sandbox, clean instance, user
  `WDAGUtilityAccount`, elevated) — all installation scenarios
- Ubuntu 22.04.5 LTS (container `ubuntu:22.04`, Docker via `portunix container`) —
  `--dry-run` comparison of default variants, nothing installed

## Scope Note

The branch changes four things: the installer switches of `exe` / `msi` variants
(`effectiveInstallArgs`), the Windows `python.json` (3.14.8, arm64, new default variant
`latest` resolved from python.org, `fallbackVariant`), the new `versionResolver` engine
feature, and `autoDetectVariant` (honours `preferred`, deterministic last resort).

The Python features pass. The first run found a regression in the `autoDetectVariant`
change, which applies to **every package on every platform** (D1). The developer fixed
it in `30dd711`; the re-test confirms the fix (see Re-test).

## Test Summary

- First run: 32 scenarios (20 unit + 10 sandbox + 2 default-variant comparisons),
  30 passed, 2 failed (both default-variant comparisons, defect D1)
- Re-test after the D1 fix: 6 new unit tests + 2 default-variant comparisons + 7 sandbox
  checks, all passed
- Open defects: 0

## Test Plan

| Level | What | Where |
| ----- | ---- | ----- |
| Unit | `effectiveInstallArgs`, python manifest guards, resolver (offline, `httptest`), fallback, `autoDetectVariant` | host, `go test` |
| E2E | dry-run of all three variants, fallback with python.org blocked, embeddable, default `latest`, explicit `full`, quoted verification (`jinja2`), uv on the combined build | Windows Sandbox |
| Diagnostics | Burn and MSI logs of the full installer, process sampling | Windows Sandbox |
| Regression | default variant of all 69 packages, `main` vs branch, `--dry-run` | host (Windows), container (Linux) |

## Test Cases

### Unit Tests (host)

All pass: `TestEffectiveInstallArgs`, `TestPythonManifest_WindowsFullVariant`,
`TestPythonManifest_WindowsDefaultIsLatest`, `TestParseReleaseVersions_NewestFirst`,
`TestResolvePythonOrgVersion_SkipsUnpublished`, `TestResolvePythonOrgVersion_NoInstallerFails`,
`TestResolveVariantVersion_SubstitutesURLs`, `TestResolveVariantVersion_UnknownResolver`,
`TestResolveVariantVersion_NoResolverUnchanged`, `TestResolveOrFallback_UsesFallbackVariant`,
`TestResolveOrFallback_NoFallbackFails`, `TestResolveOrFallback_ResolvingFallbackRejected`,
`TestAutoDetectVariant_PreferredWins`, `TestAutoDetectVariant_StableFallback`, plus the
six #201 tests and the full `ptx-installer` suite. `go vet` clean on Windows and Linux.

### E2E — Windows Sandbox

| ID | Given / When / Then | Result |
| -- | ------------------- | ------ |
| T202-01 | When `portunix install python --dry-run`, then variant `latest`, 3.14.8 resolved, 3.15.0 skipped, URL shown | PASS |
| T202-02 | When `--variant=full --dry-run`, then 3.14.8 and `python-3.14.8-amd64.exe` | PASS |
| T202-03 | When `--variant=embeddable --dry-run`, then 3.14.8 and `python-3.14.8-embed-amd64.zip` | PASS |
| T202-04 | Given `www.python.org` blocked, when `install python --dry-run`, then warning, fallback to `full` 3.14.8, exit 0 | PASS |
| T202-05 | When `--variant=embeddable`, then `C:\PortablePython\python.exe` is 3.14.8, `python -m pip` works, `python314._pth` enables `import site` | PASS (10 s) — see note |
| T202-06 | When `portunix install python` (no variant), then `latest`, installer runs with `/quiet InstallAllUsers=1 PrependPath=1 Include_test=0`, no `/S /silent` | PASS (1113 s) |
| T202-07 | Given T202-06, when `python --version` in a new terminal, then `Python 3.14.8` | PASS, `C:\Program Files\Python314\python.exe` |
| T202-08 | When `--variant=full` (already installed), then same switches, exit 0 | PASS (14 s) |
| T202-09 | When `portunix install jinja2` (verification `python -c "import jinja2; …"`), then exit 0 | PASS — quoted verification through the #201 shell helper |
| T202-10 | When `portunix install uv` on this build, then exit 0 | PASS |

Note on T202-05: the runner first reported FAIL because it looked for the exact line
`import site`. A follow-up sandbox run dumped the file: the line is `import site `
(trailing space written by `echo … >>` in `cmd`). Python accepts it and `pip` works, so
the scenario passes; the trailing space is cosmetic.

T202-06 log (excerpt):

```text
✅ Latest version: 3.14.8
🎯 Variant: latest (version: 3.14.8)
   Running: C:\Users\WDAGUtilityAccount\.portunix\cache\python-3.14.8-amd64.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0
✅ EXE installation completed
```

### Acceptance Criteria

- [x] **AC1** — `full` installs 3.14.8 and `python --version` prints `Python 3.14.8` in a
  new terminal. Verified through `latest` on a clean sandbox (same installer and switches,
  T202-06/07); the explicit `full` run (T202-08) ran on top of it
- [x] **AC2** — Log shows `/quiet InstallAllUsers=1 PrependPath=1 Include_test=0` (T202-06, T202-08)
- [x] **AC3** — `--variant=embeddable` installs 3.14.8 with a working `pip` (T202-05)
- [x] **AC4** — No variant selects `latest`, resolves the newest stable version and installs
  it like `full` (T202-01, T202-06)
- [x] **AC5** — `--variant=latest --dry-run` prints the resolved version and URL (T202-01)
- [x] **AC6** — Unit test for variant-level `installArgs` precedence
- [x] **AC7** — Unit tests for the resolver without network access
- [x] **AC8** — python.org unreachable: warning and fallback to `full` (T202-04, dry-run;
  a real install cannot fall back while python.org also hosts the download)

### Regression — Default Variant of Every Package

`--dry-run` of all 69 packages, branch vs `ptx-installer` from `main` (`main` sampled
4× on Windows, 10× on Linux because its last-resort choice was random).

Windows (17 packages differ), selection:

| Package | `main` | Branch |
| ------- | ------ | ------ |
| `chrome` | stable ×4 | beta |
| `rust` | stable / winget / beta | beta |
| `clang`, `make`, `ninja` | latest | chocolatey |
| `terraform` | latest ×3, 1.15 | 1.14 |
| `nodejs` | 18 / 20 / latest | 18 |
| `vscode` | user ×3, system | system |
| `claude-code`, `ai-assistant-*` | npm ×4 | curl |
| `mcp-ready` (bundle) | python embeddable | python latest (intended, see Observations) |

Linux (`ubuntu:22.04`, 9 packages differ), selection:

| Package | `main` (10 runs) | Branch |
| ------- | ---------------- | ------ |
| `chrome` | ubuntu ×8, snap ×2 | **fedora** |
| `vscode` | stable ×10 | snap |
| `terraform` | latest ×7, 1.15 ×2, 1.14 | 1.14 |
| `java` | 8 ×9, 17 | 11 |
| `claude-code` | npm ×8, curl ×2 | curl |

## Defects

**D1 (blocking, fixed in `30dd711`) — deterministic last-resort variant is often the wrong one.**
`autoDetectVariant` now returns the alphabetically first variant when no package-manager
variant, `preferred`, `default` or `standard` exists. The old code returned a random map
entry; Go's map order made that a de facto preference for the first-declared variant in
most runs. The branch turns occasional wrong picks into permanent ones: `chrome` on
Ubuntu selects `fedora`, `chrome` and `rust` on Windows select `beta`, `terraform` an
old pinned version, `clang` / `make` / `ninja` Chocolatey instead of the direct download.

Suggested direction (developer's choice):

1. Keep `preferred` (needed for python `latest`)
2. Before the alphabetical fallback, prefer well-known names: `latest`, `stable`, then a
   variant named after the detected distribution (`ubuntu`, `fedora`, …)
3. Mark `preferred` in manifests where none of the names fit (e.g. `java`, `nodejs`)
4. Add a regression test that loads the real manifests and pins the default variant of
   the packages above per OS

## Re-test After the D1 Fix (2026-10-07)

Fix: `PlatformSpec` records the variant declaration order when the manifest is decoded
(`VariantOrder`); the last resort of `selectDefaultVariant` is the first **declared**
variant instead of the alphabetically first one. Sampling the old code showed that its
random map pick most often landed on the first declared variant, so this restores the
behaviour manifest authors relied on, now deterministically.

- [x] Unit: `TestPlatformSpec_VariantOrder`, `TestPlatformSpec_NoVariants`,
  `TestPlatformSpec_InvalidVariants`, `TestSelectDefaultVariant_DeclaredOrder`,
  `TestSelectDefaultVariant_PackageManagerFirst`, `TestDefaultVariants_RealManifests`
  (pins chrome, rust, clang, make, ninja, terraform, vscode, claude-code, java, python on
  Windows and Linux); full `ptx-installer` suite passes, `go vet` clean on Windows and Linux
- [x] Windows, all 69 packages, `--dry-run`, branch vs majority of 6 samples of `main`:
  1 difference — `mcp-ready` (bundle) now installs python `latest`, the intended change
- [x] Linux (`ubuntu:22.04` container), all 69 packages, same comparison: 0 differences
- [x] Windows Sandbox (clean), suite `202dry`: python default `latest` resolved to 3.14.8,
  fallback to `full` with python.org blocked, defaults chrome `stable`, rust `stable`,
  terraform `latest`, vscode `user`, java `8` — 7/7 PASS

The full Python installation scenarios (T202-05 … T202-10) were not repeated: the fix
touches only default variant selection, and the Windows python default (`latest`) is
unchanged and re-verified by R202-01.

## Coverage

| Area | Covered by |
| ---- | ---------- |
| `effectiveInstallArgs` | unit + sandbox (python) — other `exe`/`msi` packages (rust, vscode, double-commander, clang) not installed |
| Resolver + fallback | unit (offline) + sandbox (live python.org, blocked python.org) |
| `python.json` variants | sandbox, all three |
| `autoDetectVariant` / `selectDefaultVariant` | unit + full-manifest comparison on Windows and Linux (found D1, confirmed the fix) + regression test over real manifests |
| arm64 / x86 URLs | URLs answer HTTP 200 (checked during development); not installed |

## Failure Injection

- `www.python.org` blocked through `hosts` → fallback (T202-04)
- Unpublished release directory (`3.15.0`) in the live listing → skipped (T202-01)
- Unreachable listing in unit tests (`httptest` returning 503)

## CI Notes

- Resolver tests are offline; no network needed in CI
- The default-variant comparison is cheap (`--dry-run` only) and caught D1;
  `TestDefaultVariants_RealManifests` now guards the affected packages in CI

## Observations / Recommendations (non-blocking)

1. **Why the Python install takes ~18 minutes in Windows Sandbox.** Diagnostic run with
   `/log`: the bundle installs 9 MSI packages (core, exe, dev, lib, doc, tcltk, launcher,
   pip, path). For **each** of them Windows Installer spends 120–121 s in
   `SOFTWARE RESTRICTION POLICY: Verifying package … has a digital signature`, while the
   actual install of the package takes 1–6 s. 9 × ~121 s ≈ 1090 s, matching the 1113 s of
   T202-06. Defender real-time protection is off in the sandbox, so it is not the cause.
   The constant 2-minute wait points to a certificate revocation lookup timing out
   (hypothesis, not proven). The three `python-3.14.8-amd64.exe` processes visible
   during install are the Burn bootstrapper (original, clean-room copy, elevated engine),
   not Python. The installer switches are not the cause either, so the "18 minutes"
   in the issue description should not be attributed to them.
   If shorter installs matter in sandboxes, every MSI left out saves ~2 minutes there:
   `Include_doc=0`, `Include_tcltk=0` (also drops IDLE), `Include_launcher=0`
2. `PATH_APPEND: C:/PortablePython` is defined on the Windows platform, so it is also added
   after `full` / `latest` installs (T202-06, T202-08), where no such directory belongs.
   It should apply to `embeddable` only
3. With `latest` as the default, the `mcp-ready` bundle now installs the full Python
   (admin, ~1 minute on a normal machine, ~18 minutes in Windows Sandbox) instead of
   the embeddable one
4. The resolver waits up to 30 s before falling back; a shorter timeout (e.g. 10 s)
   would make the fallback less painful
5. `python314._pth` gets `import site ` with a trailing space (`echo` in `cmd`)

## Final Decision

**STATUS**: PASS (re-test after the D1 fix)

**Approval for merge**: YES
**Date**: 2026-10-07
**Tester signature**: Claude Code (tester role) — to be confirmed by zdendaku
