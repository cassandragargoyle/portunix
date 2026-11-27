# Issue #251: portunix plugin list: sort plugins alphabetically and widen DESCRIPTION column

**Status**: ✅ Implemented
**Priority**: Low
**Type**: Enhancement
**Component**: portunix
**Created**: 2026-11-06
**Completed**: 2026-11-06
**GitHub**: #251
**Related**: #188

## Summary

`portunix plugin list` should print plugins sorted alphabetically by name, so
the output is stable and easy to scan. The DESCRIPTION column in the compact
table view should also be about 1.5x wider than today (30 -> 45 characters).

## Problem Description

`Registry.ListPlugins()` in `src/app/plugins/manager/registry.go` builds the
result by ranging over the `r.data.Plugins` map. Go map iteration order is
randomized, so the order of plugins in `portunix plugin list` changes between
runs. The same applies to `Registry.ListPluginsForPlatform()`, which backs
`portunix plugin list --platform ...` and the `ListPluginsForPlatform` gRPC
RPC.

Consequences:

- The table output is hard to scan with many plugins installed
- JSON / YAML output is non-deterministic, which makes diffs and scripted
  comparisons unreliable

In addition, the compact table view (`outputPluginTable()` in
`src/cmd/plugin.go`) truncates the DESCRIPTION column to 30 characters
(`truncateString(plugin.Description, 30)`). This is too short for most plugin
descriptions, so they are cut off even on wide terminals.

Finally, JSON output is only reachable through `--output json` / `-o json`.
Most Portunix commands, including the sibling `portunix plugin check --json`,
`portunix skills list --json` and `portunix system info --json`, use a boolean
`--json` flag, so `portunix plugin list --json` is the form users expect.
When no plugins are installed, `--output json` / `--output yaml` currently
print the human-readable `No plugins installed.` message instead of an empty
list, which breaks scripted consumers.

The `--output` value is not validated either: an unknown format such as
`--output xml` silently falls back to the table view, and the value is
case-sensitive, so `--output JSON` also prints the table. Scripts that
mistype the format get human-readable output instead of an error.

## Acceptance Criteria

1. `portunix plugin list` (table output) lists plugins sorted alphabetically
   by plugin name (case-insensitive, ties broken by the exact name)
2. `--output json` and `--output yaml` return plugins in the same order
3. The platform-filtered query (`--platform`) returns matches in the same order,
   both in the CLI and in the `ListPluginsForPlatform` gRPC response
4. The order is stable across repeated runs
5. Sorting is done at the registry level (`ListPlugins` /
   `ListPluginsForPlatform`), so all callers get a deterministic order
6. In the compact table view the DESCRIPTION column is widened from 30 to 45
   characters (header, separator and truncation limit); descriptions longer
   than 45 characters are still truncated with `...`, `--verbose` keeps
   printing the full text (behavior from #188)
7. The width is defined as a single named constant instead of repeated
   literals
8. `portunix plugin list --json` is supported as a shorthand for
   `--output json` (boolean flag, no short letter — same as
   `portunix plugin check --json`); it works together with `--all` and
   `--platform` / `--platform-version` / `--feature`
9. Combining `--json` with an explicit non-JSON `--output` value
   (e.g. `--json --output yaml`) returns an error; `--json --output json`
   is accepted
10. With no plugins installed, JSON output prints `[]` and YAML output prints
    `[]`, not the human-readable message; the table view keeps
    `No plugins installed.`
11. Command help (`portunix plugin list --help`) documents `--json`
12. `--output` accepts only `table`, `json` and `yaml`, case-insensitively
    (`--output JSON` equals `--output json`); any other value returns an
    error listing the allowed formats, in both the default and the
    `--platform` mode
13. Unit tests cover sorting of both registry functions, the new truncation
    width, and the `--json` / `--output` resolution including validation

## Testing

- Unit test: register several plugins in non-alphabetical order (including
  mixed case) and assert the order returned by `ListPlugins()` and
  `ListPluginsForPlatform()`
- Manual: install at least three plugins, run `portunix plugin list` several
  times and verify identical, alphabetical output in table, JSON and YAML
  formats
- Manual: verify a plugin with a description longer than 45 characters is
  truncated to 45 characters ending with `...` in the compact view and
  printed in full with `--verbose`
- Manual: `portunix plugin list --json` produces the same output as
  `portunix plugin list --output json`; `--json --output yaml` fails with a
  clear error; with an empty registry `--json` prints `[]`
- Manual: `--output xml` fails with an error listing `table, json, yaml`;
  `--output JSON` and `--json --output JSON` print JSON
