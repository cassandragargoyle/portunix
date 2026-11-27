# Issue #257: Plugin commands print text help for --help-ai instead of JSON

**Status**: ✅ Implemented
**Closed**: 2026-11-20
**Priority**: Medium
**Type**: Bug
**Component**: portunix
**Created**: 2026-11-20
**GitHub**: #257
**Related**: #203, #244, #259, #260
**Acceptance**: [acceptance-257.md](../../testing/acceptance-257.md)

## Summary

`portunix <plugin> --help-ai` prints the text usage of the plugin wrapper
command, with exit code 0, instead of machine-readable JSON. Clients that ask a
plugin for its commands get no data. The plugin manifests also have no clear
rule for `commands`, so most of them list a placeholder entry (the binary name)
instead of the real subcommands.

Found in portunix-vscode issue 145 (Functions panel showed `(0)` functions for
every plugin) and in its issue 092 (supported file extensions from
`portunix text-extractor convert --help-ai` always fall back to a fixed list).

## Problem Description

### `--help-ai` on a plugin command

Each plugin is a Cobra command with `DisableFlagParsing: true`
(`createPluginCommand` in `src/cmd/plugin_dispatcher.go`). The guard from #203
(`guardHelpLevelFlags` in `src/cmd/help_flags.go`) wraps its `RunE`, sees
`--help-ai` in the raw arguments and calls `cmd.Help()`. `multiLevelHelp` in
`src/cmd/root.go` returns JSON only for the root command, so a plugin gets the
default Cobra usage:

```text
$ portunix scraper --help-ai
Usage:
  portunix scraper [command] [args...] [flags]

Flags:
  -h, --help   help for scraper
...
```

The same happens for `portunix <plugin> <command> --help-ai`. Forwarding the
flag to the plugin would not fix it: most plugins do not know `--help-ai`
(`portunix plugin run scraper -- --help-ai` ends with
`No such option: --help-ai`).

The guard from #203 is correct for helpers, where `--help-ai` must not run the
command. For plugins it must answer with JSON.

### Placeholder commands in manifests

`PluginCommand` (`src/app/plugins/types.go`) has `name`, `description`,
`subcommands`, `parameters` and `examples`. There is no documented rule that
`commands` lists the subcommands a user calls (`portunix <plugin> <name>`).
Of 26 installed plugins, 18 list one entry named after the plugin or its binary,
for example `ptx-scraper` for `scraper` or `m365` for `m365-extractor`. Some of
them put the real commands into `subcommands` of that entry, others have none.
The plugin templates in `docs/plugin-development/languages/*/template/plugin.json`
list a `help` command, which is also not a real command.

`portunix plugin validate` does not report this.

## Proposed Solution

1. For a plugin command, `--help-ai` (also `<plugin> <command> --help-ai`)
   asks the plugin first: the plugin is started with the same arguments,
   stdin closed, `PORTUNIX_HELP_AI=1` in its environment, and its output
   captured
   - the plugin exits with code 0 and prints valid JSON to stdout → this JSON
     is printed unchanged
   - the plugin exits on its own without valid JSON (unknown option, text
     output, non-zero exit code) or cannot be started → registry fallback
     (step 2)
   - the plugin does not exit within the timeout (10 s) → the process is
     killed, an error `plugin '<name>' did not answer --help-ai within 10s`
     is printed to stderr and the exit code is non-zero; there is **no**
     registry fallback in this case, because a hanging plugin is a fault the
     user must see
   - a disabled plugin, or a plugin whose installed files are unusable, is not
     started → registry fallback
2. Registry fallback: JSON built from the registry entry of the plugin: name,
   version, description, mode, runtime and `commands` with their fields, plus
   `"source": "registry"`. The structure follows the root `--help-ai` output
   where it fits. Exit code 0
3. Document the rule for `commands` in the plugin development guide: one entry
   per subcommand that a user calls, no entry for the binary itself, an empty
   list for a plugin without subcommands. Document also that a plugin must
   either answer `--help-ai` with JSON or reject it as an unknown option —
   it must never ignore it and run the command
4. Fix the templates to follow the rule
5. `portunix plugin validate` warns about a placeholder entry: the only command
   is named like the plugin, `ptx-<plugin>`, the plugin binary, or a prefix of
   the plugin name

The plugin manifests are fixed in the plugin repositories, in separate issues.

## Acceptance Criteria

1. `portunix scraper --help-ai` (plugin does not know `--help-ai`) prints valid
   JSON with the commands from the manifest, exit code 0
2. A plugin that answers `--help-ai` with valid JSON gets its own JSON printed
   unchanged
3. `portunix <plugin> <command> --help-ai` prints valid JSON for that plugin
   by the same rules
4. A plugin that does not exit within the timeout is killed and the command
   fails with a timeout error and a non-zero exit code, without fallback
5. `--help-ai` on helper commands works as before (#203, #244)
6. The plugin development guide describes the rule for `commands`, and the
   templates follow it
7. `portunix plugin validate` warns about a placeholder command
8. Unit tests cover the plugin JSON answer, the registry fallback, the
   timeout error and the validation warning

## Testing

- Unit tests for the plugin help JSON with a fake registry entry
- Unit tests with stub plugins: one printing JSON, one rejecting `--help-ai`,
  one hanging (short test timeout)
- `portunix <plugin> --help-ai | python3 -m json.tool` for each installed plugin
- `portunix plugin validate` on a manifest with a placeholder command
- Regression: `portunix docker install --help-ai` and
  `portunix sandbox generate --help-ai` still show help and run nothing
