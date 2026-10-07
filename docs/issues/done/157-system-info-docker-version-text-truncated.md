# Issue #157: System Info Docker Version Text Truncated

**Type:** Bug Fix
**Priority:** High
**Status:** ✅ Implemented
**Labels:** bug, system-info, docker, text-parsing

## Description

The `portunix system info` command displays incorrect Docker version text. Instead of showing the full version string (e.g., "Docker: 27.5.1"), it outputs:

```text
Docker:       ersion (daemon not running)
```

The "V" from "Version" (or the beginning of the version string) is being cut off, suggesting an off-by-one error or incorrect string slicing in the Docker version parsing logic.

## Steps to Reproduce

1. Ensure Docker daemon is not running
2. Run `portunix system info`
3. Observe Docker line in output

## Expected Behavior

```text
Docker:       version 28.3.2 (daemon not running)
```

Properly formatted message with "version" prefix and version number, followed by daemon status.

## Actual Behavior

```text
Docker:       ersion (daemon not running)
```

The output is missing the leading character(s) from the version/status string.

## Root Cause Analysis

The fallback parser in `GetDockerVersion()` (`src/app/system/virtualization.go`) used a loop with `strings.HasPrefix(part, "v")` to find the version token. The word "version" also starts with "v", so it matched first. `TrimPrefix("version", "v")` then returned "ersion" instead of the actual version number.

Additionally, `formatContainerRuntime()` in `src/cmd/system.go` displayed the raw version number without a "version" prefix.

## Acceptance Criteria

- [ ] `portunix system info` displays correct Docker version text when daemon is not running
- [ ] `portunix system info` displays correct Docker version text when daemon is running
- [ ] No truncation or missing characters in Docker status output
