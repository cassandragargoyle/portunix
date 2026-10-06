# Issue #188: `portunix plugin list` — Description Truncation with `...` and `--verbose` Full Output

## 🎯 Priority

**MEDIUM** — Readability fix for `portunix plugin list`. Current table output
already truncates description to 30 chars but the truncation is not
rune-safe, and there is no way to see the full description in the human
readable output.

## 📋 Status

- **Created**: 2026-05-20
- **Status**: ✅ Implemented
- **Closed**: 2026-05-20
- **Assignee**: -
- **Branch**: feature/issue-188-plugin-list-description-truncation (merged)
- **Related**:
  - #133 — Plugin Run Command Argument Forwarding (recent plugin CLI work)
  - #156 — Plugin Health Reports "Not Enabled" for All Helper Plugins (recent
    plugin list output work)

## 📝 Problem Description

### Current Situation

`portunix plugin list` prints a fixed-width table. Description is the last
column with a width of 30 chars. The existing `truncateString` helper
appends `...` when the description exceeds the column width:

`src/cmd/plugin.go:1007-1048`

```go
func outputPluginTable(pluginList []plugins.PluginInfo, showAll bool) error {
    fmt.Printf("%-20s %-10s %-12s %-15s %-30s\n",
        "NAME", "VERSION", "INTERFACE", "STATUS", "DESCRIPTION")
    ...
    for _, plugin := range pluginList {
        ...
        fmt.Printf("%-20s %-10s %-12s %-15s %-30s\n",
            plugin.Name,
            plugin.Version,
            iface,
            status,
            truncateString(plugin.Description, 30))
    }
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

Two concrete problems are visible today:

1. `truncateString` uses `len(s)` (bytes) and slices with `s[:maxLen-3]`
   (bytes). For UTF-8 input (e.g. Czech / accented descriptions or emoji),
   this can split inside a multi-byte rune and produce broken output.
2. There is no way to read the full description from the human-readable
   table. The user has to fall back to `--output json` / `--output yaml` or
   `portunix plugin info <name>`, which is awkward when scanning many
   plugins.

Real output from `portunix plugin list`:

```text
NAME                 VERSION    INTERFACE    STATUS          DESCRIPTION
----                 -------    ---------    ------          -----------
vim-parser           0.1.6      cli          ready           Layered Information Model p...
youtube              1.0.0      cli          ready           YouTube transcript download...
m365-extractor       1.1.0      cli          ready           Microsoft 365 integration p...
scraper              1.2.0      cli          ready           Universal web scraping plug...
text-extractor       1.1.3      cli          ready           Advanced text extraction pl...
```

The 30-char truncation works for ASCII but is not safe for UTF-8 input,
and the full text is not reachable from the table view.

### Expected Behavior

1. **Default (compact) view** — `portunix plugin list`:
   - Description column truncated with `…` (or `...`) when the description
     is longer than the column width.
   - Truncation operates on **runes**, not bytes — UTF-8 descriptions never
     get cut inside a multi-byte sequence.

2. **Verbose view** — `portunix plugin list --verbose` (alias `-v`):
   - Full description printed for every plugin, with line wrapping so the
     output is readable in a normal terminal (no horizontal overflow).
   - Wrapping should respect word boundaries where reasonable; long
     unbreakable tokens (URLs, identifiers) may be broken at the column
     boundary.
   - All other columns continue to show as in the compact view.

3. `--output json` / `--output yaml` always include the full description
   (no change in behavior — they already do).

### Note on Existing `--verbose` Flag

`pluginListCmd` already declares `--verbose` / `-v`, but it is currently
only honored in the platform-capability query path:

`src/cmd/plugin.go:83-93, 367`

```go
verbose, _ := cmd.Flags().GetBool("verbose")
if platform != "" {
    return listPluginsForPlatform(platform, platformVersion, features, outputFormat, verbose)
}
...
return listPlugins(showAll, outputFormat)   // verbose is dropped here
```

`pluginListCmd.Flags().BoolP("verbose", "v", false, "Include platform_payload in human-readable output")`

This issue extends the existing `--verbose` flag so that, in the default
(non-platform) listing path, it triggers the full-description output described
above. The flag's help text must be updated to describe both behaviors.

## 🎯 Root Cause

- `truncateString` is byte-based, not rune-based.
- `listPlugins` ignores the `verbose` flag — only `listPluginsForPlatform`
  reads it.
- There is no terminal-aware word-wrap helper for multi-line table cells.

## ✅ Acceptance Criteria

1. `portunix plugin list` truncates long descriptions with an ellipsis
   (`...` or `…`) and never splits a UTF-8 rune. A description containing
   Czech / accented characters or emoji renders correctly.
2. `portunix plugin list --verbose` (and `-v`) prints the full description
   for every plugin, wrapped so it fits the terminal width and remains
   readable.
3. Wrapping prefers word boundaries; lines do not exceed the available
   terminal width (fallback to 80 columns when width cannot be detected).
4. `portunix plugin list --help` documents the new `--verbose` behavior
   for the default listing path (alongside the existing platform-payload
   description).
5. `portunix plugin list --output json` and `--output yaml` continue to
   emit the full description unchanged.
6. Behavior is identical on Windows and Linux terminals (PowerShell,
   Windows Terminal, bash, etc.). On Windows, a plain `...` is acceptable
   if `…` does not render reliably in the default code page.
7. Integration test in `test/integration/` (using `testframework`) covers:
   - Compact mode truncation for an ASCII-only description longer than 30
     chars.
   - Compact mode truncation for a description containing multi-byte UTF-8
     characters (no broken runes in the output).
   - `--verbose` mode prints the full description and wraps long lines.
   - `--output json` is unchanged when `--verbose` is set.

## 🛠️ Suggested Fix

### Compact-mode truncation (rune-safe)

Replace the byte-based `truncateString` with a rune-aware version in
`src/cmd/plugin.go`:

```go
func truncateString(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    if maxLen <= 3 {
        return string(runes[:maxLen])
    }
    return string(runes[:maxLen-3]) + "..."
}
```

### Verbose mode

Plumb `verbose` into `listPlugins` and `outputPluginTable`:

```go
// pluginListCmd.RunE
return listPlugins(showAll, outputFormat, verbose)

// listPlugins
func listPlugins(showAll bool, outputFormat string, verbose bool) error {
    ...
    return outputPluginTable(plugins, showAll, verbose)
}

// outputPluginTable
func outputPluginTable(pluginList []plugins.PluginInfo, showAll, verbose bool) error {
    ...
    for _, plugin := range pluginList {
        ...
        if verbose {
            // Print the row without DESCRIPTION
            fmt.Printf("%-20s %-10s %-12s %-15s\n",
                plugin.Name, plugin.Version, iface, status)
            // Then print the full description, indented and wrapped
            for _, line := range wrapForTerminal(plugin.Description, indent) {
                fmt.Println(line)
            }
        } else {
            fmt.Printf("%-20s %-10s %-12s %-15s %-30s\n",
                plugin.Name, plugin.Version, iface, status,
                truncateString(plugin.Description, 30))
        }
    }
    return nil
}
```

For terminal width detection prefer
`golang.org/x/term.GetSize(int(os.Stdout.Fd()))` with an 80-column
fallback when the call fails (non-TTY, redirected output). The wrapping
helper should be a small local function — no new third-party dependency
unless a clean stdlib-only implementation becomes too unwieldy.

### Help text update

Update `pluginListCmd.Flags().BoolP("verbose", ...)`:

```go
pluginListCmd.Flags().BoolP("verbose", "v", false,
    "Show full description (wrapped) in the default listing; "+
        "include platform_payload in platform-query mode")
```

And expand the `Long:` description on `pluginListCmd` to mention the new
default-listing behavior.

## 🧪 Testing Strategy

- Unit tests for the new rune-safe `truncateString` covering: ASCII shorter
  than max, ASCII at exact max, ASCII longer than max, UTF-8 input with
  multi-byte characters at the truncation boundary, empty string,
  `maxLen <= 3` edge case.
- Unit tests for the line-wrap helper: words shorter than width, words
  longer than width (must break), terminal width fallback when not a TTY.
- Integration test (`test/integration/issue_188_plugin_list_verbose_test.go`)
  using `testframework`:
  - Installs / uses an existing plugin with a long description (or a
    fixture plugin) and asserts:
    - Default `plugin list` truncates with `...` at column 30.
    - `plugin list --verbose` prints the full description.
    - UTF-8 description renders without broken runes.
    - `--output json` is unaffected.
- Run integration tests in a container per
  `docs/contributing/ISSUE-DEVELOPMENT-METHODOLOGY.md` (container-based
  testing policy).

## 📎 References

- `src/cmd/plugin.go:60-93` — `pluginListCmd` definition (RunE, flags)
- `src/cmd/plugin.go:367` — current `--verbose` flag declaration
- `src/cmd/plugin.go:416-435` — `listPlugins`
- `src/cmd/plugin.go:1007-1048` — `outputPluginTable` + `truncateString`
- `docs/issues/done/156-plugin-health-not-enabled-for-helper-plugins.md`
  — prior cleanup of `plugin list` status column
- `docs/issues/done/221-plugin-run-command-argument-forwarding.md`
  — prior plugin-CLI UX issue
