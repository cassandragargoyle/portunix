# Acceptance Protocol - Issue #163

**Issue**: Helpers Missing --help-ai and --help-expert Flags
**Branch**: feature/163-helpers-help-ai-help-expert
**Tester**: Claude (QA/Test Engineer)
**Date**: 2026-04-02
**Testing OS**: Linux 6.17.0-19-generic (host)

## Test Summary

- Total test scenarios: 5
- Passed: 5
- Failed: 0
- Skipped: 0

## Test Results

### TC001: All 12 helpers --help-ai exit code 0

**Given**: All 12 built helper binaries (ptx-*)
**When**: Each is invoked with `--help-ai`
**Then**: Exit code is 0

**Result**: PASS - All 12 helpers return exit code 0

### TC002: All 12 helpers --help-expert exit code 0

**Given**: All 12 built helper binaries (ptx-*)
**When**: Each is invoked with `--help-expert`
**Then**: Exit code is 0

**Result**: PASS - All 12 helpers return exit code 0

### TC003: --help-ai outputs valid JSON with meaningful content

**Given**: 6 fixed helpers (ptx-credential, ptx-make, ptx-mcp, ptx-prompting, ptx-trace, ptx-virt)
**When**: Each is invoked with `--help-ai`
**Then**: Output is valid JSON containing tool name, version, description, and non-empty commands array

**Result**: PASS

| Helper | tool field | commands count |
| ------ | ---------- | -------------- |
| ptx-credential | ptx-credential | 11 |
| ptx-make | ptx-make | 13 |
| ptx-mcp | ptx-mcp | 1 |
| ptx-prompting | ptx-prompting | 3 |
| ptx-trace | ptx-trace | 20 |
| ptx-virt | ptx-virt | 12 |

### TC004: --help-expert outputs meaningful structured text

**Given**: 6 fixed helpers
**When**: Each is invoked with `--help-expert`
**Then**: Output contains DESCRIPTION section, command listings, and is longer than 10 lines

**Result**: PASS - All helpers produce structured help with domain-specific sections:

- ptx-credential: 43 lines (COMMANDS section)
- ptx-make: 41 lines (FILE OPERATIONS, BUILD METADATA, BUILD TOOLS, UTILITIES)
- ptx-mcp: 27 lines (COMMANDS section)
- ptx-prompting: 24 lines (COMMANDS section)
- ptx-trace: 58 lines (SESSION MANAGEMENT, EVENT RECORDING, ANALYSIS, EXPORT, ALERTS)
- ptx-virt: 43 lines (COMMANDS section)

### TC005: Preflight check passes

**Given**: All 12 helpers built in project root
**When**: `scripts/github-01-preflight-check.sh` is executed
**Then**: Helper check section reports "All 12 helpers have --help-ai and --help-expert"

**Result**: PASS

### Regression Tests

- [x] Existing functionality unaffected (build succeeds, no other test failures)
- [x] Helpers still respond correctly to --help, --version flags
- [x] Preflight check helper section passes

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-04-02
**Tester signature**: Claude (QA/Test Engineer)
