# Issue #259: Consistent help levels (--help-expert, --help-ai) on every command level

**Status**: Open
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix, ptx-container
**Created**: 2026-11-20
**GitHub**: #259
**Related**: #50, #163, #203, #244, #257

## Summary

The multi-level help (#50) defines three levels: `--help`, `--help-expert`
and `--help-ai` (JSON). Today only the root command, the top-level helper
commands and, after #257, the plugin commands answer `--help-ai` with JSON.
Everywhere else both flags fall back to the plain help, sometimes on stderr,
always with exit code 0. A client cannot tell that it got no JSON.

## Problem Description

Behavior measured on `fix/257-plugin-help-ai-json` (commit `0fbf444`):

| Command type | Example | `--help-expert` | `--help-ai` |
| ------------ | ------- | --------------- | ----------- |
| Root | `portunix` | Extended text | JSON |
| Helper, top level | `mcp`, `github` | Extended text from the helper | JSON from the helper |
| Helper, subcommand | `github clone` | Same as `--help` | Text, same as `--help` |
| Helper without help levels | `container`, `container run` | Same as `--help` | Text |
| Built-in command | `install`, `docker install` | Same as `--help` | Text |
| Built-in Cobra subcommand | `sandbox generate` | Usage on **stderr** | Usage on **stderr** |
| Plugin | `scraper`, `docgen` | Empty wrapper usage on **stderr** | JSON (#257) |

Causes:

- `multiLevelHelp` in `src/cmd/root.go` handles the levels only for the root
  command. For any other command it calls `cmd.Usage()`, which writes to
  stderr and leaves out the long description
- The help guard from #203 (`guardHelpLevelFlags` in `src/cmd/help_flags.go`)
  prevents running the command, but it can only show this usage
- Helper subcommands and `ptx-container` do not implement the help levels;
  the helper work in #163 and #244 covered only the top-level helper command
- For plugins, #257 added `--help-ai` (plugin first, registry fallback), but
  `--help-expert` still shows the usage of the Cobra wrapper, which has no
  information about the plugin

## Proposed Solution

1. Contract for every command: help output goes to stdout with exit code 0.
   `--help-ai` always prints JSON. `--help-expert` always prints text with at
   least the content of `--help`
2. Built-in Cobra commands: a generic generator from the command tree (name,
   short and long description, flags with type and default, subcommands,
   examples) for `--help-ai`; the same data as text for `--help-expert`.
   `multiLevelHelp` uses it for every non-root command instead of
   `cmd.Usage()`. The root JSON keeps its current structure
3. Helpers (dispatcher): `--help-expert` and `--help-ai` are forwarded to the
   helper also for subcommands. When the helper does not answer as required
   (no JSON for `--help-ai`), the generator from step 2 answers, by the same
   rule as for plugins in #257. `ptx-container` gets the help levels
4. Plugins: `--help-expert` uses the mechanism of #257 — ask the plugin first,
   otherwise print text built from the manifest (commands, parameters,
   examples)
5. A contract test walks all commands of `rootCmd` and checks stdout, exit
   code 0 and valid JSON for `--help-ai`

## Acceptance Criteria

1. `portunix <command> [subcommand] --help-ai` prints valid JSON to stdout with
   exit code 0 for every built-in command and subcommand
2. `portunix <command> [subcommand] --help-expert` prints text to stdout with
   exit code 0 and contains at least the content of `--help`
3. Helper subcommands (e.g. `github clone`, `container run`) follow criteria
   1 and 2
4. `portunix <plugin> --help-expert` prints text from the plugin or from the
   manifest, never the empty wrapper usage
5. The root `--help`, `--help-expert` and `--help-ai` output keeps its
   structure
6. The guard from #203 still prevents running a command for both flags
   (`docker install`, `sandbox generate`)
7. A contract test over the command tree covers criteria 1 and 2

## Testing

- Unit: generator output for a Cobra command with flags and subcommands
- Contract test: all commands of `rootCmd` with `--help-ai` and
  `--help-expert` (stdout, exit code 0, JSON valid)
- E2E: `github clone`, `container run`, `sandbox generate`, `docker install`
  and one plugin with both flags
- Regression: root help levels, `--help` everywhere unchanged, nothing runs
  with a help flag
