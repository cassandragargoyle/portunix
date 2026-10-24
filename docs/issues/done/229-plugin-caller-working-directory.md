# Issue #229: Plugin dispatcher does not pass the caller's working directory to plugins

**Status**: Implemented
**Priority**: High
**Type**: Bug
**Component**: portunix
**Created**: 2026-10-10
**Closed**: 2026-10-10
**GitHub**: #229
**Related**: #230 (Windows verification)

## Summary

`portunix <plugin> ...` starts the plugin process with its working directory set to the
plugin install directory and gives the plugin no way to learn where the caller was
standing. Relative paths passed on the command line are therefore resolved against the
plugin directory and fail. The dispatcher should export `PORTUNIX_CWD` (the absolute
working directory of the caller) into the plugin environment while leaving the plugin's
own working directory unchanged.

## Problem Description

Observed on Linux with `v2.5.1+dev.1`:

```bash
cd /tmp/books
portunix cabinet pdf prepare --dry-run book.pdf
# -> input file not found
```

`book.pdf` exists in `/tmp/books`, but the plugin looks for it in
`~/.portunix/plugins/cabinet`.

### Root cause

`executePlugin` in `src/cmd/plugin_dispatcher.go` sets

```go
cmd.Dir = freshPlugin.InstallPath
```

and does not set `cmd.Env`, so the plugin inherits the Portunix environment as is.
Nothing in that environment carries the directory in which `portunix` was invoked.

## Expected Behavior

When the plugin process is started, its environment contains

```text
PORTUNIX_CWD=<absolute path of the caller's working directory>
```

- The value is the directory in which the `portunix` command was run, captured before
  Portunix itself changes directory
- The working directory of the plugin process stays as it is today — plugins may depend
  on it (for example a uv project looks up `pyproject.toml` in the plugin directory)
- Applies to all plugins and all runtimes (native Go, Java, Python), not only to `cabinet`
- Applies to Linux and Windows
- All other environment variables passed to the plugin remain unchanged

The variable name is fixed: the `cabinet` plugin already reads `PORTUNIX_CWD` and resolves
relative input, output, and report paths against it. When the variable is missing, it
falls back to its own working directory.

## Implementation Notes

1. CLI dispatch — `executePlugin` in `src/cmd/plugin_dispatcher.go` builds the command
   for every runtime (`buildJavaCommand`, `buildPythonCommand`, `buildNativeCommand`) and
   is the place to add the variable
2. Other launch paths — `src/app/plugins/grpc_client.go` starts long-running plugin
   services (`p.cmd.Dir = p.config.WorkingDir`, environment from `os.Environ()` plus
   `p.config.Environment`). Check for further paths (`serve`, MCP). Decide and document
   whether `PORTUNIX_CWD` makes sense there: a server has no per-request caller directory
3. Add `PORTUNIX_CWD` to the environment of the plugin process
4. Add a test proving that a launched plugin receives the variable and that it holds the
   caller's directory, not the plugin directory
5. Document the variable in the plugin contract for plugin developers
   (`docs/plugin-development/`)

### Caution

When inspecting the environment of a running plugin process, read only `PORTUNIX_CWD`.
Do not dump the whole environment or every variable with the `PORTUNIX` prefix — it
includes `PORTUNIX_PLUGINS_GITHUB_TOKEN`.

## Acceptance Criteria

1. From an arbitrary directory, `portunix cabinet pdf prepare --dry-run book.pdf`
   succeeds when `book.pdf` is located in that directory
2. `portunix cabinet pdf prepare -o out.pdf book.pdf` writes `out.pdf` into the caller's
   directory, not into the plugin directory
3. Absolute paths work the same as before
4. A plugin that does not read the variable behaves the same as before
5. The test from implementation note 4 passes on Linux and on Windows
6. `PORTUNIX_CWD` is described in the plugin developer documentation, including the
   decision for long-running (service / MCP) launch paths

## Testing

- Unit / integration test for the dispatcher: launch a stub plugin from a temporary
  directory and assert that `PORTUNIX_CWD` equals that directory and differs from the
  plugin install path; assert the plugin working directory is still the install path
- Regression: a plugin ignoring the variable produces identical output before and after
- Manual verification with the `cabinet` plugin per acceptance criteria 1–3, on Linux
  and Windows
