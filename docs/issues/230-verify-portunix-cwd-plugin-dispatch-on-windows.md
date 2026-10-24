# Issue #230: Verify PORTUNIX_CWD plugin dispatch on Windows

**Status**: Open
**Priority**: Medium
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-10
**GitHub**: #230
**Related**: #229

## Summary

Issue #229 made the plugin dispatcher export `PORTUNIX_CWD` (the absolute working
directory of the caller) into the plugin process environment. The change was tested on
Linux only. On Windows the code and its tests were compiled but never executed, so
acceptance criterion 5 of #229 is still open and the acceptance protocol
`docs/testing/acceptance-229.md` has the status CONDITIONAL.

## Problem Description

The following was not exercised on Windows:

- `TestPluginEnvironment`, `TestBuildPluginCommand_AllRuntimes` and
  `TestBuildPluginCommand_StubPluginReceivesCallerDir` in
  `src/cmd/plugin_dispatcher_test.go` — the last one launches a `.ps1` stub plugin through
  `powershell.exe`, a path that has never run
- The Windows-only branch of `pluginEnvironment` in `src/cmd/plugin_dispatcher.go`, which
  drops an inherited `PORTUNIX_CWD` with a case-insensitive key comparison
- Manual verification with the `cabinet` plugin (acceptance criteria 1–3 of #229)
- Caller directories with spaces and non-ASCII characters, and drive-letter paths

Only `GOOS=windows go vet portunix.ai/cmd` and `GOOS=windows go test -c portunix.ai/cmd`
were run, both successfully.

## Acceptance Criteria

1. On Windows, the dispatcher tests pass:

   ```powershell
   go test portunix.ai/cmd -run "TestPluginEnvironment|TestBuildPluginCommand" -v
   ```

2. From an arbitrary directory, `portunix cabinet pdf prepare --dry-run book.pdf`
   succeeds when `book.pdf` is located in that directory
3. `portunix cabinet pdf prepare -o out.pdf book.pdf` writes `out.pdf` into the caller's
   directory, not into the plugin directory
4. Absolute paths (including a path on another drive, if available) work as before
5. A caller directory containing a space and a non-ASCII character is passed intact
6. With `PORTUNIX_CWD` (and a differently cased variant such as `Portunix_Cwd`) already
   set in the calling shell, the plugin sees exactly one value — the real caller directory
7. A plugin that does not read the variable produces the same output as with the
   previous Portunix version
8. `docs/testing/acceptance-229.md` is updated with the Windows results and its status
   changed from CONDITIONAL to PASS; any defect found is fixed under this issue

## Testing

- Use the `tester-windows` role; record the Windows version and build in the protocol
- Build with `make build` on the Windows machine or VM, run criterion 1 from the
  repository root (the `src/cmd` module is testable only by import path)
- For criteria 5–7 a stub plugin that prints only `PORTUNIX_CWD` and its working directory
  is sufficient; register it in an isolated user profile so the real plugin registry is
  not modified
- When inspecting the environment of a plugin process, read only `PORTUNIX_CWD` — never
  dump the whole environment
