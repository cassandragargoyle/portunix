# Issue #161: Plugin Install Extra Wheels Support

**Date**: 2026-03-28
**Reporter**: reco-studio team
**Status**: ✅ Implemented
**Priority**: High
**Type**: Bug Fix / Enhancement
**Labels**: bug, plugin-system, python, wheel, installation
**Affects**: portunix v1.11.0
**Target version**: v2.0.0

## Problem

`portunix plugin install` ignores the `extra_wheels` field in `plugin.json`. When a plugin declares bundled dependency wheels via `extra_wheels`, they are not installed into the plugin's virtual environment before the main wheel.

This causes installation failure for any plugin whose main wheel declares pip dependencies that are only available as bundled wheels in the `dist/` directory.

## Reproduction

1. Build a plugin with bundled dependency wheels:

   ```text
   dist/
     plugin.json
     main_plugin-0.1.0-py3-none-any.whl       # main wheel
     dependency_lib-0.3.0-py3-none-any.whl     # bundled dependency
   ```

2. Verify `plugin.json` declares extra wheels:

   ```json
   {
     "plugin": {
       "wheel": "main_plugin-0.1.0-py3-none-any.whl",
       "extra_wheels": ["dependency_lib-*.whl"]
     }
   }
   ```

3. Install:

   ```bash
   portunix plugin install dist/
   ```

4. **Result**: Installation fails because pip cannot find the bundled dependency.

## Expected Behavior

`portunix plugin install` should:

1. Read `extra_wheels` from `plugin.json`
2. Resolve glob patterns (e.g. `dependency_lib-*.whl`) against the plugin dist directory
3. Install matched wheels into the plugin venv **before** the main wheel
4. Then install the main wheel (whose pip dependencies are now satisfied)

## Current Workaround

Manual installation:

```bash
mkdir -p ~/.portunix/plugins/my-plugin
python3 -m venv ~/.portunix/plugins/my-plugin/.venv

# Install dependency wheel first, then main wheel
~/.portunix/plugins/my-plugin/.venv/bin/pip install \
    dist/dependency_lib-0.3.0-py3-none-any.whl \
    dist/main_plugin-0.1.0-py3-none-any.whl

cp dist/plugin.json ~/.portunix/plugins/my-plugin/
```

This workaround does not register the plugin in `~/.portunix/plugins/registry.json`, so portunix does not recognize it.

## Additional Issue: No `plugin update` Command

There is no way to update an already installed plugin. The current flow requires:

```bash
portunix plugin uninstall my-plugin
portunix plugin install dist/
```

A `portunix plugin update <path>` or `portunix plugin install --force <path>` command would simplify the development workflow significantly.

## Suggested Implementation

### extra_wheels Support

In the Python wheel plugin installer (Go code), after creating the venv and before installing the main wheel:

1. Read `plugin.json` -> `plugin.extra_wheels`
2. For each glob pattern in `extra_wheels`:
   - Resolve against the plugin source directory
   - `pip install` matched `.whl` files
3. `pip install` main wheel (dependencies now available)

### plugin update Command

```text
portunix plugin update <path>   # reinstall wheel(s) into existing venv
portunix plugin install --force <path>  # alternative: overwrite existing
```

## Acceptance Criteria

- [ ] `extra_wheels` glob patterns in `plugin.json` are resolved and installed before the main wheel
- [ ] Plugin installation succeeds when all dependencies are provided as bundled wheels
- [ ] Error message when `extra_wheels` glob matches no files
- [ ] `plugin update` or `plugin install --force` command available for reinstallation
