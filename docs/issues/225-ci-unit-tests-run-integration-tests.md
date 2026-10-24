# Issue #225: CI Unit Tests job runs integration/E2E tests and always fails

**Status**: Open
**Priority**: High
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-07
**GitHub**: #225
**Related**: #224, #195

## Summary

The GitHub Actions **Unit Tests** job runs `go test -tags=unit -race ./...`. The `unit`
build tag does not exclude anything, so the run also picks up container / E2E tests in
`test/integration` and tests in the root package that need a built binary or a container
runtime. 29 tests fail on every PR (PR #34 for v2.5.1, PR #223 for v2.6.1). Because the
Unit Tests job fails, the Coverage, Integration, Cross Platform and Performance jobs are
skipped.

## Problem Description

Failing packages on PR #223:

- `portunix.ai/portunix` — `TestMainWithoutArguments`
- `portunix.ai/portunix/test/integration` (437 s) — `TestClaudeCodeContainerInstallAndSetup`,
  `TestE2EContainerMCPIntegration`, `TestIssue037MCPServeImplementation`,
  `TestIssue041NodeJsNpmInstallation`, `TestIssue044ContainerCpCommand`,
  `TestIssue047*` (3), `TestIssue048*` (2), `TestIssue049*` (3), `TestIssue052*` (5),
  `TestIssue059PlaybookHelpCommand`, `TestIssue074*` (2), `TestIssue075HugoContainerInstall`,
  `TestIssue079*` (2), `TestIssue082*` (2), `TestIssue083HugoRegistryInstallation`,
  `TestPluginInstallation`

Passing packages: `pkg/docker`, `pkg/logging`, `src/helpers/ptx-make/cmd`,
`src/helpers/ptx-pft`, `src/pkg/installconfig`, `src/pkg/platform`, `test/unit`.

Cause of `TestMainWithoutArguments` (found while testing #231): the test expects cobra's
help sections `Available Commands:` and `Flags:`, but `portunix` without arguments prints
the custom help from `src/cmd/help_registry.go` (`Usage: portunix [command] [options]`,
`Common commands:`). The test is stale, not the binary; it fails locally as well.

Note: test file names still use legacy issue numbers (`issue_0xx_*`); see the legacy
numbering table in `docs/contributing/ISSUE-MANAGEMENT.md`.

## Proposed Approach

1. Add a build tag (e.g. `//go:build integration`) to `test/integration` tests and other
   tests that need a binary, network, or container runtime, so `-tags=unit` skips them
2. Run them in the Integration Tests job with `-tags=integration`, where a container
   runtime is available, or keep them local-only
3. Triage the remaining real failures (`TestMainWithoutArguments` in the root package)
4. Remove or update tests that are stale (see #195 for the same problem in `src/cmd`)

## Acceptance Criteria

1. `go test -tags=unit -race ./...` passes on a clean GitHub runner
2. Integration tests are excluded from the unit run by a build tag, not deleted
3. Integration tests that can run in CI run in the Integration Tests job
4. Unit Tests job is green on the next release PR

## Testing

- Local: `go test -tags=unit -race ./...` without a container runtime available
- CI: Unit Tests job on the release PR
