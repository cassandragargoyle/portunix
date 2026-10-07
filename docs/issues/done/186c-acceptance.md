# Acceptance Protocol - Issue #186c

**Issue**: PTX-Installer Legacy Code Cleanup — sub-task 186c: extract
`LoadInstallConfig` into shared `src/pkg/installconfig/` package
**Branch**: `refactor/issue-186c-installconfig-extract` (commit `c8839db`)
**Tester**: Tester (generic role)
**Date**: 2026-05-19
**Testing OS**: Windows 11 Pro 10.0.26200 (host) — source-level / build-level
refactor; container provisioning behaviour unchanged so container testing not
required for this sub-task

## Test Summary

- Total test scenarios: 7
- Passed: 7
- Failed: 0
- Skipped: 0

## Test Environment

- Host OS: Windows 11 Pro 10.0.26200
- Go: project pinned to `go 1.24.0` (toolchain go1.24.2 in helper modules)
- Build system: `make build` (project rule, NOT `go build` directly)
- Commit under test: `c8839db refactor(186c): extract installconfig into shared package`

## Test Results

### TC01 — installconfig unit tests + coverage

- Command: `go test ./src/pkg/installconfig/ -v -cover`
- Result: **PASS** — 9 test functions / sub-tests pass; coverage **90.7 %**
  of statements (target ≥ 80 % per ADR-026)
- Covered:
  - `TestLoadDefaultConfig`
  - `TestMergeConfigs_UserOverridesDefault`
  - `TestLoadInstallConfig_NoUserOverlay`
  - `TestLoadInstallConfig_WithUserOverlay`
  - `TestLoadInstallConfig_MalformedUserConfigFallsBackToDefault`
  - `TestVariantConfig_GetDistributionsList` (4 sub-cases: nil, list, map, mixed)
  - `TestInstallConfig_JSONRoundTrip`

### TC02 — Full `make build` (main + helpers)

- Command: `make clean ; make build`
- Result: **PASS** — main `portunix.exe` and 19 helpers built
  (ptx-container, ptx-mcp, ptx-virt, ptx-ansible, ptx-prompting, ptx-python,
  ptx-installer, ptx-aiops, ptx-make, ptx-pft, ptx-credential, ptx-trace,
  ptx-ssh, ptx-plugin-registry, ptx-proxmox, ptx-specpm, ptx-database,
  ptx-github, ptx-wizard)
- Notes: cosmetic `del` errors in `make clean` on Git Bash are explicitly
  marked `(ignored)` in the Makefile; the build target itself completed
  cleanly.

### TC03 — Consumer imports switched

- Acceptance criterion (per issue):
  `src/cmd/docker_run_in_container.go` and
  `src/cmd/podman_run_in_container.go` no longer import
  `portunix.ai/app/install`.
- Result: **PASS**
  - `docker_run_in_container.go:12` imports
    `portunix.ai/portunix/src/pkg/installconfig`
  - `podman_run_in_container.go:14` imports
    `portunix.ai/portunix/src/pkg/installconfig`
  - Neither file imports `portunix.ai/app/install` (grep confirmed).

### TC04 — `LoadInstallConfig` behavioural equivalence

- Given/When/Then triplets verified via unit tests:
  1. **No overlay**: `LoadInstallConfig()` returns
     `*InstallConfig{Version:"1.0", Packages:map{}, Presets:map{}}` —
     `TestLoadInstallConfig_NoUserOverlay` passes.
  2. **With overlay** (`~/.portunix/install-config.json` present): user
     packages and presets are merged into default —
     `TestLoadInstallConfig_WithUserOverlay` passes.
  3. **Malformed overlay**: function falls back to defaults silently —
     `TestLoadInstallConfig_MalformedUserConfigFallsBackToDefault` passes.
- Source-level equivalence: function bodies of `LoadInstallConfig`,
  `loadDefaultConfig`, `loadUserConfig`, `mergeConfigs` in the new
  `src/pkg/installconfig/installconfig.go` are byte-identical to the
  pre-refactor versions in `src/app/install/config.go` on `main` (only a
  comment was reworded in `loadDefaultConfig`).
- Result: **PASS**

### TC05 — Backward-compatibility aliases

- Acceptance criterion (per issue): `src/app/install/config.go` keeps a thin
  wrapper that re-exports from the shared package until sub-tasks 186a/186d
  are merged.
- Aliases present:
  - `type InstallConfig = installconfig.InstallConfig`
  - `type PackageConfig = installconfig.PackageConfig`
  - `type PlatformConfig = installconfig.PlatformConfig`
  - `type VariantConfig = installconfig.VariantConfig`
  - `type VerificationConfig = installconfig.VerificationConfig`
  - `type PresetConfig = installconfig.PresetConfig`
  - `type PresetPackageConfig = installconfig.PresetPackageConfig`
  - `type VersionSupportPolicy = installconfig.VersionSupportPolicy`
  - `type FallbackStrategy = installconfig.FallbackStrategy` (in
    `fallback.go`) + 4 const aliases (`FallbackAuto`, `FallbackAutoConfirm`,
    `FallbackManual`, `FallbackDisabled`)
  - `type VersionRange = installconfig.VersionRange` (in `version_matcher.go`)
  - `var LoadInstallConfig = installconfig.LoadInstallConfig`
- Verified by successful build of the entire root module (all existing
  call sites inside `app/install` — `installer.go`, `install.go`,
  `registry.go`, `ai_integration.go`, `fallback.go`, `version_matcher.go`,
  `integration.go` — still compile against the aliases).
- Result: **PASS**

### TC06 — `go vet` regression

- Command: `go vet ./...`
- Result: **PASS** — the only reported issue
  (`test/integration/issue_037_mcp_serve_test.go:32 syscall.SysProcAttr.Setpgid`)
  is **pre-existing on `main`** (verified by checking out `main` and running
  the same command). No new vet errors introduced by #186c.

### TC07 — Smoke test main binary

- Commands: `./portunix.exe --version`, `./portunix.exe system info`
- Result: **PASS** — binary runs, prints version (`Portunix version dev`)
  and system info correctly. Dispatcher-routed helpers respond.

## Functional Tests Checklist

- [x] Feature (extracted shared package) works as specified — package
      compiles standalone, has tests, ≥ 80 % coverage
- [x] Acceptance criteria from issue met (TC03, TC05, TC07, ≥ 80 % coverage)
- [x] Edge cases handled (malformed overlay, nil distributions, map vs
      list distributions, JSON round-trip)

## Regression Tests Checklist

- [x] Existing functionality unaffected (TC02 full build, TC05 aliases
      maintain backward compat for `app/install` internal call sites)
- [x] Cross-platform compatibility verified — refactor is pure source-level,
      `installconfig` package uses only stdlib (`encoding/json`, `fmt`,
      `os`, `path/filepath`); the `withTempHome` test helper handles both
      `HOME` (Unix) and `USERPROFILE` (Windows) env vars

## Helper-Module Side Effects

- `src/helpers/ptx-mcp/go.mod` and `src/helpers/ptx-plugin-registry/go.mod`
  now contain `replace portunix.ai/portunix => ../../..` and a transitive
  `require portunix.ai/portunix` line.
- Reason: `app/install` (consumed by `ptx-mcp` line 464 — issue #186b
  target) now imports `installconfig` from the parent module; helper
  modules need the `replace` direction to resolve it. This is a known and
  acceptable transitional state — sub-task **186b** will remove the
  `app/install` import from `ptx-mcp` and these `replace` directives can
  then be dropped.
- `go mod tidy` was run for both helpers; `ptx-plugin-registry` confirmed
  no direct use of `portunix.ai/portunix` (require line dropped after tidy)
  — replace remains as defensive guard.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES

**Date**: 2026-05-19

**Tester signature**: Tester (generic role) — sub-task 186c acceptance
testing performed on Windows 11 host. All 7 test cases passed. Coverage of
new shared package is 90.7 % (target ≥ 80 %). No regression in existing
functionality. Helper-module `go.mod` changes are an expected and minimal
transitional side-effect that 186b will clean up.

## Recommendations for Developer / Architect

1. **Proceed to merge** `refactor/issue-186c-installconfig-extract` into
   `main` and archive sub-task progress per project workflow.
2. **Note for #186b implementer**: once `ptx-mcp` no longer imports
   `portunix.ai/app/install`, the `replace portunix.ai/portunix => ../../..`
   in `src/helpers/ptx-mcp/go.mod` and the corresponding `replace
   portunix.ai/app/install` line become removable. Same for
   `ptx-plugin-registry` (already flagged as vestigial in #186).
3. **Pre-existing `go vet` issue**: file a separate issue for the
   Linux-syscall test on Windows in
   `test/integration/issue_037_mcp_serve_test.go` — unrelated to #186c.
