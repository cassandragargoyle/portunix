# Issue #239: Plugin prerequisites: duplicated "uv" in the uv runtime warning

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Bug
**Component**: portunix (plugin manager)
**Created**: 2026-10-24
**Closed**: 2026-10-24
**GitHub**: #239
**Related**: #234, `docs/testing/acceptance-234.md` (observations O1, O2)

## Summary

Two cosmetic problems in the output of `plugin install`, `plugin install --force` and
`plugin repair` for Python wheel plugins, found while testing #234:

1. The uv runtime warning prints the word `uv` twice
2. The pin recommendation follows the prerequisite table without an empty line

## Problem Description

### 1. Duplicated "uv"

When uv does not have a matching managed interpreter yet, `CheckUVPythonRuntime` in
`src/app/plugins/prerequisites.go` returns `Found: "uv"` and a message that also starts with
`uv`. `FormatHuman` prints both:

```text
  Runtime:            python               ⚠️ Found: uv uv will download Python 3.14 during setup
```

### 2. Missing empty line

The pin recommendation printed by `setupPythonWheelPlugin` in
`src/app/plugins/manager/manager.go` follows the table directly:

```text
  OS support:         linux,windows,darwin ✅ Current: linux
  ⚠️  Plugin demo does not set plugin.python_version; using Python 3.14 (override with PORTUNIX_PLUGIN_PYTHON_VERSION)
```

## Proposed Solution

1. Leave `Found` empty in the download case (the message already names uv), so the line reads
   `⚠️ uv will download Python 3.14 during setup`
2. Separate the prerequisite table from the following setup output with an empty line

## Acceptance Criteria

1. The uv download warning contains the word `uv` once
2. The JSON output of the runtime check (`found`) stays valid; `found` is omitted or empty in
   the download case
3. An empty line separates the prerequisite table from the setup output of `install`,
   `install --force` and `repair`
4. The unit test of `CheckUVPythonRuntime` checks the download-case message and `Found`

## Testing

- Unit: `go test portunix.ai/app/plugins/...`
- Container with uv and without a managed interpreter of the requested version:
  `portunix plugin install <wheel-plugin>`; verify the warning line and the empty line
