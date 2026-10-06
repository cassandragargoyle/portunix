# Issue #192: Create Python plugin venv with uv using a pinned interpreter version

## Overview
**Title**: Create Python plugin venv with uv using a pinned interpreter version
**Status**: ✅ Implemented
**Closed**: 2026-07-04
**Priority**: High
**Type**: Enhancement / Architecture
**Labels**: plugin-system, python, distribution, uv, bytecode, cross-platform

## Problem Description

Python plugins are distributed as **bytecode-only wheels** (source `.py` stripped,
only `.pyc` shipped — see the scraper plugin's `scripts/build_dist.py bytecode`
mode, portunix-plugins). CPython bytecode is tied to a specific **minor** version:
a `.pyc` compiled with 3.13 only loads on 3.13.x, a 3.12 `.pyc` only on 3.12.x.

Portunix currently creates the plugin venv with **whatever `python3` happens to be
on the host**:

```go
// src/app/plugins/manager/manager.go  (setupPythonWheelPlugin, ~line 606–622)
pythonCmd := findPython()                     // system python3 / python
...
cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
...
pipPath := venvExecPath(venvPath, "pip")
cmd = exec.Command(pipPath, "install", wheelPath)
```

If the host's `python3` minor version differs from the version the wheel's
bytecode was compiled with, the plugin fails at runtime with:

```
ImportError: bad magic number in 'scraper.cli': b'\xf3\r\r\n'
```

**Observed case**: the scraper plugin wheel is built with Python 3.12 (the build
tool's default), but portunix created `~/.portunix/plugins/scraper/.venv` with the
host's system Python 3.13.7 → the plugin does not load.

The `py3-none-any` wheel tag advertises "any Python 3", which is **not true** for a
bytecode-only wheel. Pinning only the plugin build side does not solve it, because
the host side (this venv creation) still picks an arbitrary interpreter.

## Goal

Make Python plugin activation **resilient on every machine**: the venv must be
created with the exact interpreter the plugin declares, and that interpreter must
be provisioned even if the host does not already have it. This is done with **uv**,
which can download and manage a specific CPython version on demand.

Build side and runtime side then follow the **same pinned version** → build and
runtime bytecode always match.

## Requirements

### Functional Requirements

1. **Declared interpreter version**
   - A Python plugin declares the exact interpreter version it was built for.
   - Source of truth options (pick one, document it):
     - a new `plugin.python_version` field in `plugin.json` (e.g. `"3.13"`), **or**
     - a `.python-version` file shipped inside the plugin dist.
   - Distinct from the existing `python_min_version` (which stays as a
     compatibility floor); bytecode needs an **exact minor**, not a minimum.

2. **Create venv via uv with the pinned version**
   - When uv is available, create the plugin venv with:
     `uv venv --python <declared-version> <venvPath>`
   - uv **auto-provisions** the interpreter (downloads a managed CPython) when the
     host lacks it — this is what guarantees "the right Python is available".
   - Install the wheel into that venv with `uv pip install --python <venv> <wheel>`
     (and likewise for `extra_wheels`).

3. **Graceful fallback**
   - If uv is not installed, fall back to the current `findPython()` +
     `python3 -m venv` path, but **verify** the resulting interpreter's minor
     version matches the plugin's declared version. On mismatch, fail the install
     with a clear, actionable message (offer `portunix install uv`) instead of
     letting it break later at runtime with "bad magic number".

4. **No behavior change for plugins that do not declare a version**
   - Backward compatible: plugins without a pinned version keep the current flow.

### Technical Requirements

1. **Code touch points**
   - `src/app/plugins/manager/manager.go` → `setupPythonWheelPlugin()` (venv
     creation + pip install).
   - `findPython()` stays as the fallback locator.
   - `src/app/plugins/manifest.go` → add/parse the `python_version` field if that
     option is chosen (next to `PythonMinVersion`, line ~107).

2. **uv detection & install**
   - Detect uv on PATH; if missing, surface `portunix install uv` (or reuse the
     existing package-install path).

3. **Version verification helper**
   - Read the declared version, compare against the created venv's
     `python --version` (minor granularity) for the fallback path.

## Proposed Solution (sketch)

```go
func (m *Manager) setupPythonWheelPlugin(manifest *plugins.PluginManifest, pluginDir string) error {
    wheelPath := filepath.Join(pluginDir, manifest.Plugin.Wheel)
    venvPath := filepath.Join(pluginDir, ".venv")
    pyVersion := manifest.Plugin.PythonVersion // e.g. "3.13" (new, optional)

    if uv := findUV(); uv != "" && pyVersion != "" {
        // uv provisions the exact interpreter (downloads it if the host lacks it)
        run(uv, "venv", "--python", pyVersion, venvPath)
        run(uv, "pip", "install", "--python", venvPath, wheelPath)
        return nil
    }

    // Fallback: system python3, but fail loudly on a bytecode-incompatible minor
    pythonCmd := findPython()
    // ... existing python3 -m venv + pip install ...
    if pyVersion != "" && !minorMatches(pythonCmd, pyVersion) {
        return fmt.Errorf(
            "plugin %s needs Python %s but host python is %s; install uv "+
            "(portunix install uv) so the exact interpreter can be provisioned",
            manifest.Name, pyVersion, hostMinor(pythonCmd))
    }
    return nil
}
```

## Plugin-side counterpart (portunix-plugins, tracked here for context)

- The plugin build pins the same version (`.python-version`) so its bytecode wheel
  is compiled with that interpreter.
- The build declares the version to portunix via `plugin.python_version` /
  shipped `.python-version`.
- Optionally the build tags the wheel honestly (`cp313`) instead of `py3-none-any`
  so an accidental wrong-interpreter install fails at install time.

## Acceptance Criteria

- [ ] A bytecode-only Python plugin declaring `python_version: "3.13"` installs and
      runs on a host whose system `python3` is **not** 3.13 (uv provisions 3.13).
- [ ] `portunix plugin enable <name>` + `portunix <name> --version` works without
      "bad magic number".
- [ ] On a host without uv and without a matching python, install fails with a
      clear message (not a runtime crash).
- [ ] Plugins without a declared version behave exactly as today.
- [ ] Documented in plugin authoring docs (how to declare the version).

## References

- Bytecode mismatch reproduced with the `scraper` plugin (v1.4.0), portunix-plugins.
- Build tool: `plugins/scraper/scripts/build_dist.py` (bytecode mode) — portunix-plugins.
- Existing venv creation: `src/app/plugins/manager/manager.go:590` `setupPythonWheelPlugin`.
- Existing manifest field: `src/app/plugins/manifest.go:107` `python_min_version`.
- uv managed Python: <https://docs.astral.sh/uv/concepts/python-versions/>
