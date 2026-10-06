# Issue #195: Fix Stale Container Command Tests in `src/cmd` (Surfaced by #194)

## 🎯 Priority

**MEDIUM** — Test debt. Since #194, `make test` and `make test-cli` actually
run the `src/cmd` tests (previously silently skipped) and 6 container-related
test functions fail because their expectations drifted from the current
command behavior. Until fixed, `make test` cannot pass, weakening the quality
gate #194 restored.

## 📋 Status

- **Created**: 2026-07-08
- **Status**: 📋 Open
- **Assignee**: -
- **Branch**: fix/issue-195-stale-src-cmd-container-tests
- **Related**:
  - Issue #194 — `make test` multi-module coverage (revealed these failures)
  - `src/cmd/container_exec_test.go`
  - `src/cmd/container_management_test.go`
  - `src/cmd/container_run_test.go`
  - `src/cmd/container_run_integration_test.go`

## 📝 Problem Description

### Current Situation

The `src/cmd` module tests were never executed by `make test` before #194
(the module was outside the root `./...` boundary). Running them now via
`go test portunix.ai/cmd/...` yields 6 failing test functions (10 subtests),
all in the container command area:

| Test | Failing subtests |
| ---- | ---------------- |
| `TestContainerExecCommand` | `no_arguments`, `only_container_name` |
| `TestContainerCommandErrorMessages` | `Stop_command_error_message_format`, `All_commands_mention_runtime_delegation` |
| `TestContainerRunCommand_ErrorHandling` | `TC-038-I005: Runtime unavailable error`, `TC-038-I006: Container creation error` |
| `TestContainerRunCommand_ArgumentParsing` | `TC-038-U008: Image with complex command` |
| `TestContainerRunCommand_FlagValidation` | `TC-038-U012: Missing flag value` |
| `TestContainerRunCommand_OriginalIssueRepro` | `TC-038-U013: Original issue command should work` |

The failures are assertion mismatches, not build errors — the tests assert
error-message wording, argument-validation behavior, and runtime-delegation
hints that no longer match the current `portunix container` implementation
(dispatcher pattern, `ptx-container` helper delegation).

### Root Cause

Tests were written against an older in-process implementation of the
container commands and were never run after the commands migrated to the
dispatcher/helper architecture, so they drifted without anyone noticing.

## ✅ Acceptance Criteria

1. `go test portunix.ai/cmd/...` passes (0 failures)
2. Each failing test is either updated to assert the current intended
   behavior, or removed with justification if it tests behavior that no
   longer exists (e.g., superseded by `ptx-container` helper tests)
3. No production code behavior changes purely to satisfy stale assertions —
   if a test reveals a genuine defect in the current behavior, fix the
   defect instead and note it in the acceptance protocol
4. `make test-cli` exits 0

## 💡 Implementation Notes

- Compare expected error messages in tests with actual output of
  `portunix container exec|stop|run` (dispatcher delegates to
  `ptx-container`; `handleContainerRun` treats `args[0]` as image)
- `TC-038-*` cases originate from Issue #038 — check
  `docs/issues/internal/done/` for original intent before rewriting
- Argument-validation subtests (`no_arguments`, `only_container_name`,
  `Missing flag value`) may just need updated cobra `Args` expectations

## 🧪 Test Instructions

1. `make test-cli` — must pass
2. `make test GO_MODULES=.` — root module iteration includes satellite
   packages (`portunix.ai/cmd/...`); container tests must pass
3. Regression: `portunix container run ubuntu:22.04 --dry-run` and error
   paths (`container exec` without args) still behave as documented
