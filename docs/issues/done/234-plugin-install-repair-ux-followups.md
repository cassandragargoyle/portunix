# Issue #234: Plugin install/repair: false Python runtime error with uv, unpinned uv interpreter, usage text after errors

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix (plugin manager)
**Created**: 2026-10-22
**Closed**: 2026-10-24
**GitHub**: #234
**Related**: #233, `docs/testing/acceptance-233.md` (observations O1–O3), `docs/testing/acceptance-234.md`

## Summary

Testing of #233 found three usability problems in the plugin manager. None of them blocks #233,
but each one misleads the user or makes future breakage likely:

1. The prerequisite check reports a missing Python runtime as an error even when uv is about
   to provide the interpreter
2. A Python wheel plugin without `plugin.python_version` gets the newest uv-managed CPython,
   which may be newer than the plugin's dependencies support
3. Runtime errors of plugin commands are followed by the full cobra usage text, which pushes
   the fix hint out of view

## Problem Description

### 1. False "Python runtime not found" with uv

`plugin install`, `plugin install --force` and `plugin repair` run
`plugins.CheckPrerequisites` before the venv setup. For `runtime: python` it checks the host
`python3` only. On a host without system Python but with uv, the output is:

```text
  Runtime:            python               ❌ Python runtime not found
                      Fix: portunix install python

⚠️  Prerequisites check found issues. Plugin will be installed but may not work correctly.

  Creating Python virtual environment with uv (Python latest)...
  Python wheel plugin setup complete
✅ Plugin repaired successfully!
```

The plugin works, because the venv uses a uv-managed interpreter. `plugin check <name>`
reports the same runtime error for an installed plugin whose venv is healthy, and exits with
code 1, so scripts treat a working plugin as broken.

For a wheel plugin the relevant runtime is the venv interpreter, not the host `python3`.

### 2. Unpinned plugins get the newest uv-managed Python

Since #233 the venv is created with `uv venv --managed-python`. The `--python` request is
`python_version`, else `>=python_min_version`, else nothing. Without either field uv picks the
newest managed CPython; in the #233 test that was 3.15. A plugin with native dependencies
(for example numpy, torch, onnxruntime) may have no wheels for a just-released Python, and
installation then fails or builds from source.

Several published plugins still do not declare `python_version`.

### 3. Usage text after runtime errors

When `plugin repair`, `plugin install` or the plugin dispatcher (`portunix <plugin> …`) fails
at runtime, cobra prints the error and then the full usage and flag list:

```text
Error: plugin 'demo' is broken: venv interpreter is missing: …/.venv/bin/python -> python3
Fix: portunix plugin repair demo
Usage:
  portunix demo [command] [args...] [flags]

Flags:
  -h, --help   help for demo
...
```

The usage text is meant for wrong arguments, not for runtime failures. The same happens on
`main` before #233.

## Proposed Solution

1. **Runtime check for wheel plugins**
   - Installed plugin: check the venv interpreter (`CheckInstall` from #233) instead of the
     host `python3`; report the venv Python version as `Found`
   - Install / repair: when uv is available, report the runtime as provided by uv (OK, or a
     warning when uv will download an interpreter) instead of an error
2. **Interpreter choice without `python_version`**
   - Default to a conservative minor (for example the newest minor that has been released for
     at least six months, or a version set in Portunix config) instead of the newest one
   - Print a warning that recommends adding `plugin.python_version` to the manifest
   - Document `python_version` as recommended for every wheel plugin in the plugin
     development guide
3. **Silence usage on runtime errors**
   - Set `SilenceUsage: true` on the `plugin` command tree and on the dynamically created
     plugin dispatcher commands, so usage is printed only for argument errors

## Acceptance Criteria

1. `plugin install`, `install --force` and `repair` of a wheel plugin print no runtime error
   when uv is available and no system Python is installed
2. `plugin check <name>` exits 0 for a healthy uv-backed wheel plugin on a host without
   system Python, and reports the venv interpreter version
3. A wheel plugin without `python_version` does not silently get the newest CPython; the
   chosen version and a recommendation to pin it are printed
4. Runtime errors of `plugin repair`, `plugin install` and `portunix <plugin> …` print the
   error and the fix hint without the usage text; argument errors still show usage
5. Unit tests cover the runtime check for wheel plugins and the interpreter choice

## Testing

- Container without system Python, with uv: install, repair and check a wheel plugin without
  `python_version`; verify output and exit codes
- Container with system Python and uv: verify the chosen interpreter and the pin warning
- `portunix plugin repair <missing-plugin>` and running a broken plugin: verify no usage text
- `portunix plugin repair` without an argument: verify usage is still shown
