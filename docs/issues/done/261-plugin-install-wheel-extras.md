# Issue #261: Plugin install: opt-in wheel extras (--extra) for Python wheel plugins

**Status**: ✅ Implemented
**Priority**: Medium
**Type**: Enhancement
**Component**: portunix (plugin manager)
**Created**: 2026-11-22
**Closed**: 2026-11-22
**GitHub**: #261
**Related**: #161, cassandragargoyle/portunix-reco#141, cassandragargoyle/portunix-reco#140

## Summary

A Python wheel plugin cannot install an extra (optional dependency group) of
its wheel. Add an opt-in at install time, `portunix plugin install <plugin>
--extra <name>`, so the default install stays small and a user who needs the
extra gets it through Portunix.

## Problem Description

`setupPythonWheelPluginUV` and `setupPythonWheelPluginSystem`
(`src/app/plugins/manager/manager.go`) install the plugin wheel with

```text
uv pip install --python <venv> <plugin dir>/<plugin.wheel>
<venv>/bin/pip install <plugin dir>/<plugin.wheel>
```

The manifest field `plugin.wheel` must end with `.whl`
(`src/app/plugins/manifest.go`), and there is no field for an extra, so
`<wheel>[train]` cannot be requested. Only the main dependencies of the wheel
are installed.

The plugin `reco` of portunix-reco needs the extra `train` (`torch`,
`transformers`, several GB) for `portunix reco model train`. Most users of the
plugin only collect and verify training samples and do not need it, so the
extra must not be installed by default.

## Proposed Solution

1. Manifest: the plugin declares the extras a user can select.

   ```json
   "plugin": {
     "wheel": "ptx_reco-0.4.1-py3-none-any.whl",
     "optional_extras": [
       {"name": "train", "description": "Model training (torch, transformers, several GB)"}
     ]
   }
   ```

   Validation: `optional_extras` requires `wheel`; a name matches
   `^[A-Za-z0-9][A-Za-z0-9._-]*$`; names are unique.
2. Install: `portunix plugin install <plugin> --extra train` (repeatable or
   comma separated) installs `"<wheel>[train]"` with uv and with the pip
   fallback. An extra that the manifest does not declare is an error before
   anything is installed.
3. The default install (no `--extra`) is unchanged.
4. Optional: add an extra to an installed plugin without a full reinstall, and
   show the declared and the installed extras in `portunix plugin info
   <plugin>`.

### Workaround Until Then

```text
uv pip install --python ~/.portunix/plugins/reco/.venv "~/.portunix/plugins/reco/ptx_reco-<version>-py3-none-any.whl[train]"
```

## Acceptance Criteria

1. `portunix plugin validate` accepts a manifest with `optional_extras` and
   rejects an invalid name or `optional_extras` without `wheel`
2. `portunix plugin install <plugin> --extra train` installs the wheel with the
   extra (uv and pip fallback)
3. `portunix plugin install <plugin>` without `--extra` installs only the main
   dependencies, as today
4. `--extra` with a name that the manifest does not declare stops with an error
   that lists the declared extras

## Testing

- Unit tests: manifest validation of `optional_extras`; the install arguments
  with and without extras for uv and for pip
- Integration: install a small test wheel with an extra that pulls one small
  package; check the package is in the venv only with `--extra`
