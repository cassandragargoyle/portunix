# Issue #130: Docker Version Detection Bug in System Info

**Status:** ❌ Closed (Duplicate of #157)
**Closed:** 2026-05-10

## Resolution

This issue is a duplicate of [#157 — System Info Docker Version Text Truncated](157-system-info-docker-version-text-truncated.md), which was implemented on 2026-03-17 (commits `735d314` + `cf06add`).

### Fix already in place

- `src/app/system/virtualization.go` — fallback parser `parseDockerVersionOutput()` extracts version via regex `\d+\.\d+\.\d+`, eliminating the previous off-by-one bug that returned `"ersion"` from `TrimPrefix("version", "v")`.
- `src/app/system/virtualization_test.go` — 7 unit tests including the regression test `must not return ersion from word version` (verified passing).
- `src/cmd/system.go` — display layer adds explicit `version` prefix.

No further code changes required. Closing as duplicate.

## Summary

Fix incorrect Docker version parsing in `portunix system info --json` output. Currently returns `"docker_version": "ersion"` instead of actual version.

## Problem Statement

When running `portunix system info --json` on Windows with Docker Desktop installed, the `docker_version` field contains truncated value `"ersion"` instead of the full version string (e.g., `"27.5.1"`).

### Current Output (Incorrect)

```json
{
  "capabilities": {
    "docker": true,
    "docker_version": "ersion",
    "container_available": true
  }
}
```

### Expected Output

```json
{
  "capabilities": {
    "docker": true,
    "docker_version": "27.5.1",
    "container_available": true
  }
}
```

## Root Cause

Likely a string parsing issue when extracting Docker version from `docker --version` or `docker version` command output. The parsing probably incorrectly captures part of the word "Version" instead of the actual version number.

Typical `docker --version` output:
```
Docker version 27.5.1, build 9f9e405
```

The bug suggests the regex or string split is capturing "ersion" from "Version" instead of "27.5.1".

## Affected Component

- `app/system/` or similar - system info detection code
- Windows-specific Docker detection path

## Acceptance Criteria

1. [ ] `docker_version` returns correct version string (e.g., "27.5.1")
2. [ ] Works on Windows with Docker Desktop
3. [ ] Works on Linux with Docker Engine
4. [ ] Handles case when Docker is installed but version cannot be determined (return empty string or "unknown")

## Priority

Low - cosmetic issue, `container_available` field works correctly for detection

## Labels

bug, system-info, docker, windows

## Related Issues

- #129 Docusaurus QuickStart Script (discovered during testing)

## Notes

- Discovered during testing of quickstart-docusaurus.ps1 script on Windows 11 VM
- The `container_available: true` field works correctly and can be used for detection
