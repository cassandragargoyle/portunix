# Issue #186: PTX-Installer Legacy Code Cleanup (follow-up to #100)

**Status**: ✅ Implemented (all 4 sub-tasks merged to `main` 2026-05-19; final
`src/app/install/` directory deletion deferred to a follow-up cleanup task)
**Priority**: Medium
**Type**: Refactor / Architecture
**Created**: 2026-05-19
**Architecture Decision Records**:
- [ADR-025: PTX-Installer Helper Architecture](../../adr/025-ptx-installer-helper-architecture.md)
- [ADR-026: Shared Platform Utilities](../../adr/026-shared-platform-utilities.md)
**Parent Issue**: [#100 — PTX-Installer Helper Implementation](100-ptx-installer-helper-implementation.md) (closed 2026-05-19)
**Branch (per sub-task)**: `refactor/issue-186a-…`, `refactor/issue-186b-…`, etc.

## Summary

Issue #100 (PTX-Installer Helper Implementation) was closed in 2026-05-19 as
"substantially complete" — the helper binary is production-deployed in
Portunix 2.3.0 and the performance goal from #099 is achieved. However, the
legacy installation subsystem `src/app/install/` (~373 KB Go code, 18 files)
**cannot yet be removed** because 8 active consumers in the main binary and
in `ptx-mcp` / `ptx-plugin-registry` helpers still depend on it.

This issue tracks the four follow-up refactorings needed to fully retire
`src/app/install/`. They are split into independently merge-able sub-tasks
so each can be tested in isolation and rolled back if needed.

## Scope: removing the remaining dependencies on `src/app/install/`

Audit performed 2026-05-19 — consumers still importing `portunix.ai/app/install`:

| Consumer | Location | What it uses | Sub-task |
| -------- | -------- | ------------ | -------- |
| `portunix registry` CLI | `src/cmd/registry.go` (~800 lines) | `LoadPackageRegistry`, `NewAIPackageManager`, `Package`, `VersionDiscoveryResult` | **186a** |
| `ptx-mcp` helper | `src/helpers/ptx-mcp/init.go:464` | `install.InstallPackage("claude-code", "npm")` | **186b** |
| `docker run-in-container` | `src/cmd/docker_run_in_container.go:124` | `install.LoadInstallConfig()` | **186c** |
| `podman run-in-container` | `src/cmd/podman_run_in_container.go` | `install.LoadInstallConfig()` | **186c** |
| `portunix install apt source ...` | `src/cmd/install_apt.go` | `apt.NewAptManager()` (9 call sites) | **186d** |
| `portunix install iso ...` | `src/cmd/install_iso.go` | `install.ISOInstaller` | **186d** |
| `portunix install chocolatey ...` | `src/cmd/install_chocolatey.go` | `chocolatey.*` | **186d** |
| `ptx-plugin-registry` go.mod | `src/helpers/ptx-plugin-registry/go.mod:9` | `replace` declaration | follow-up (likely vestigial; verify) |

---

## Sub-task #186a — Migrate `portunix registry` command to `ptx-installer`

**Goal**: Move the `portunix registry` CLI (list / info / check-updates /
update-report / deps / etc.) out of the main binary into `ptx-installer`
(or split AI version discovery into a shared `src/pkg/registry/` package
if migrating into the helper proves too invasive).

**Affected files**:
- `src/cmd/registry.go` (~800 lines, uses `install.LoadPackageRegistry`,
  `install.NewAIPackageManager`, `install.Package`,
  `install.VersionDiscoveryResult`).
- `src/dispatcher/dispatcher.go` (add `"registry"` to `ptx-installer`'s
  `Commands` list, or leave for shared-pkg variant).
- `src/helpers/ptx-installer/main.go` (add `registry` subcommand handlers
  if migrating into the helper).
- `src/app/install/registry.go`, `src/app/install/ai_integration.go`,
  `src/app/install/version_matcher.go` — candidates for removal once
  consumers are migrated.

**Open design questions** (decide before implementation):
1. Migrate the command into `ptx-installer`, or extract AI version discovery
   into a shared package and keep `registry` in the main binary?
2. If migrating: `portunix registry` becomes dispatched (consistent with
   `install` / `package`).
3. If extracting: shared package mirrors the ADR-026 pattern (Phase 2.5 of
   issue #100 already created `src/pkg/platform/`).

**Acceptance criteria**:
- [x] `portunix registry list`, `info`, `check-updates`, `update-report`,
      `deps` behave identically to current implementation.
- [x] `src/cmd/registry.go` no longer imports `portunix.ai/app/install`.
- [x] No regression in AI-assisted version discovery (ADR-020).
- [x] Acceptance protocol in `docs/issues/done/186a-acceptance.md`. — PASS, 2026-05-19

**Resolution**: Chose **extract** path (shared `src/pkg/packageregistry/`)
over **migrate** to keep the dispatcher surface unchanged.

**Status**: ✅ Implemented (merged to `main` 2026-05-19)

---

## Sub-task #186b — Refactor `ptx-mcp` to call `ptx-installer` for claude-code install

**Goal**: Replace the direct `install.InstallPackage("claude-code", "npm")`
call in `ptx-mcp` with a subprocess invocation of `ptx-installer install
claude-code`. This is the only cross-helper code dependency on
`src/app/install/`.

**Affected files**:
- `src/helpers/ptx-mcp/init.go:464` — replace direct call with
  `exec.Command(ptxInstallerPath, "install", "claude-code")` (use helper
  discovery from `src/shared` package).
- `src/helpers/ptx-mcp/go.mod:10,15` — remove `replace portunix.ai/app/install`
  and the `require` line once unused.
- `src/helpers/ptx-mcp/init.go:25` — remove `import "portunix.ai/app/install"`.

**Design notes**:
- ptx-mcp must locate `ptx-installer` reliably (use existing
  `shared.HelperDiscovery` mechanism in `src/shared/`).
- Handle the case where `ptx-installer` is not available (fall back to the
  existing curl-based install on line 471 with a clear log message).
- Wire stdout/stderr forwarding so the user sees install progress.

**Acceptance criteria**:
- [x] `ptx-mcp` setup installs claude-code via subprocess when
      `ptx-installer` is present.
- [x] Fallback path (curl install) still works when `ptx-installer` is
      missing.
- [x] `src/helpers/ptx-mcp/` no longer imports `portunix.ai/app/install`.
- [x] Acceptance protocol in `docs/issues/done/186b-acceptance.md`. — PASS, 2026-05-19

**Status**: ✅ Implemented (merged to `main` 2026-05-19)

---

## Sub-task #186c — Extract `LoadInstallConfig` into a shared package

**Goal**: Container runtime commands (`docker run-in-container`,
`podman run-in-container`) need the parsed install config (package list,
variant defaults) to provision containers — they do NOT need the full
installation engine. Move the config loader into a small shared package.

**Affected files**:
- New: `src/pkg/installconfig/installconfig.go` — moves the
  `LoadInstallConfig()`, `loadDefaultConfig()`, `loadUserConfig()`,
  `mergeConfigs()` functions (plus the `InstallConfig`, `PackageConfig`,
  `PlatformConfig`, `VariantConfig` types) out of `src/app/install/config.go`.
- New: `src/pkg/installconfig/installconfig_test.go` (mirror the
  ADR-026 test pattern from `src/pkg/platform/`).
- `src/cmd/docker_run_in_container.go:12,124` — import shared package
  instead of `portunix.ai/app/install`.
- `src/cmd/podman_run_in_container.go:13` — same.
- `src/app/install/config.go` — keep a thin wrapper that re-exports
  from the shared package (backward compat for sub-tasks 186a/186d) until
  those are also merged.

**Acceptance criteria**:
- [x] `portunix docker run-in-container` and
      `portunix podman run-in-container` behave identically.
- [x] `src/pkg/installconfig/` has ≥ 80 % test coverage (per ADR-026
      pattern). — achieved **90.7 %** on 2026-05-19
- [x] `src/cmd/docker_run_in_container.go` and
      `src/cmd/podman_run_in_container.go` no longer import
      `portunix.ai/app/install`.
- [x] Acceptance protocol in `docs/issues/done/186c-acceptance.md`. — PASS, 2026-05-19

**Status**: ✅ Implemented (merged to `main` 2026-05-19)

---

## Sub-task #186d — Migrate `install apt source / iso / chocolatey` admin subcommands

**Goal**: These are administrative subcommands of `portunix install` (not
package-install operations themselves). They manage APT sources, ISO
files, and Chocolatey state. Migrate them into `ptx-installer` so the
main binary stops importing `app/install/apt`, `app/install/chocolatey`,
and `app/install/ISOInstaller`.

**Affected files**:
- `src/cmd/install_apt.go` (9 sites using `apt.NewAptManager()`) — replace
  with delegation to `ptx-installer` subcommands.
- `src/cmd/install_iso.go:45` (`install.ISOInstaller`) — same.
- `src/cmd/install_chocolatey.go` — same.
- `src/helpers/ptx-installer/main.go` — add `install apt source ...`,
  `install iso ...`, `install chocolatey ...` subcommand handlers.
- `src/app/install/apt/`, `src/app/install/chocolatey/`,
  `src/app/install/install_iso.go` — move to `src/helpers/ptx-installer/`.

**Design notes**:
- The dispatcher already routes `install` to `ptx-installer` (`src/dispatcher/dispatcher.go:97-101`),
  so once the subcommand handlers exist in the helper, the `src/cmd/install_*.go`
  files can be deleted entirely.
- Confirm none of the `apt.*` types are used outside the
  `install_apt.go` command (audit before deletion).

**Acceptance criteria**:
- [x] `portunix install apt source add/remove/list`, `portunix install iso ...`,
      `portunix install chocolatey ...` behave identically (source-level
      equivalence verified; containerised Linux-host runtime test deferred
      to follow-up tester-linux acceptance — see TC07 in 186d-acceptance.md).
- [x] `src/cmd/install_apt.go`, `install_iso.go`, `install_chocolatey.go`
      deleted; dispatch handles everything via helper.
- [x] Acceptance protocol in `docs/issues/done/186d-acceptance.md`. — PASS (TC07 conditional), 2026-05-19

**Status**: ✅ Implemented (merged to `main` 2026-05-19)

---

## Final cleanup (after 186a-d merged)

Once all four sub-tasks land, `src/app/install/` should have no remaining
consumers. Verify with:

```bash
grep -rn "portunix.ai/app/install" --include='*.go' --include='go.mod'
```

Then delete:
- `src/app/install/` (entire directory, ~373 KB)
- `src/helpers/ptx-plugin-registry/go.mod` line 9 (`replace
  portunix.ai/app/install => ../../app/install`) — verify it's vestigial.
- Root `go.mod` `replace` directive for `portunix.ai/app/install` (if present).

Expected outcome:
- Main `portunix` binary size drops below 20 MB (currently ~24 MB).
- Memory footprint for non-install commands drops further (already at
  acceptable levels in 2.3.0 from issue #099).

## Suggested order of execution

1. **186c** first — smallest, least invasive (pure extraction). De-risks
   shared package pattern for the others.
2. **186b** next — touches only one helper, decouples cross-helper deps.
3. **186a** — largest of the migrations, but isolated to the `registry`
   command.
4. **186d** — last, deletes the most code. Best done when the rest of
   `src/app/install/` is already shrinking.

Each sub-task is an independent PR/merge; the parent issue stays open
until all four are done and `src/app/install/` is deleted.

## Labels

- refactor
- architecture
- technical-debt
- helper-binary
- package-management
- follow-up-100
