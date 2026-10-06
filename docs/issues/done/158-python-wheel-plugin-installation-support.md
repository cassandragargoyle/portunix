# Issue #158: Python Wheel Plugin Installation Support

**Status**: ✅ Implemented
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, plugin-system, python, wheel, venv, installation

## Context

Portunix plugin system currently supports Go-based helper binaries dispatched via the main binary. A new category of plugins is emerging — **Python-based plugins** distributed as Python wheel (`.whl`) files. The plugin loader needs to understand `plugin.json` metadata that declares a Python runtime and wheel dependency, and handle the full lifecycle: venv creation, wheel installation via pip, and execution through the venv's `bin/` directory.

### plugin.json Schema Extension

```json
{
  "plugin": {
    "binary": "ptx-scraper",
    "runtime": "python",
    "python_min_version": "3.11",
    "wheel": "ptx_scraper-1.1.0-py3-none-any.whl"
  }
}
```

**Field semantics:**

| Field | Description |
|-------|-------------|
| `binary` | Name of the executable entry point inside `.venv/bin/` |
| `runtime` | Declares runtime type — `"python"` triggers venv+wheel workflow |
| `python_min_version` | Minimum required Python version (semver-compatible check) |
| `wheel` | Filename of the `.whl` file to install via pip |

## Requirements

### R1: Plugin Install (`portunix plugin install dist/`)

When the plugin loader processes a directory containing `plugin.json` with `runtime: "python"` and a `wheel` field:

1. **Copy distribution** — copy `dist/` contents to `~/.portunix/plugins/<plugin-name>/`
2. **Detect runtime** — read `runtime` field; if `"python"`, enter Python wheel workflow
3. **Validate Python version** — check that available `python3` meets `python_min_version` requirement
4. **Create virtual environment** — `python3 -m venv ~/.portunix/plugins/<plugin-name>/.venv`
5. **Install wheel** — `~/.portunix/plugins/<plugin-name>/.venv/bin/pip install ~/.portunix/plugins/<plugin-name>/<wheel-filename>`

### R2: Plugin Execution (`portunix <plugin-command> <args>`)

When dispatching a command to a Python-runtime plugin:

1. **Resolve binary path** — `~/.portunix/plugins/<plugin-name>/.venv/bin/<binary>`
2. **Execute** — run the resolved binary with all `<args>` forwarded

### R3: Error Handling

- Python not found or version too low → clear error message with required version
- Wheel file missing in plugin directory → error with expected filename
- pip install failure → show pip output, suggest manual troubleshooting
- venv creation failure → check Python venv module availability

### R4: Plugin Uninstall

- Remove entire `~/.portunix/plugins/<plugin-name>/` directory (includes `.venv`)

## Acceptance Criteria

- [ ] `portunix plugin install dist/` with Python wheel plugin creates venv and installs wheel
- [ ] `portunix <plugin-command> <args>` dispatches to `.venv/bin/<binary>` correctly
- [ ] Python version validation works (rejects too-old Python)
- [ ] Missing wheel file produces clear error
- [ ] Plugin uninstall removes venv and all plugin files
- [ ] Existing Go-based plugin workflow is unaffected

## Technical Notes

- The `binary` field in `plugin.json` serves dual purpose: it names both the dispatcher command and the entry point script that pip creates in `.venv/bin/`
- Virtual environment isolation ensures no system-wide pip pollution
- Consider future extension: `requirements` field for additional pip dependencies beyond the wheel

## Related Issues

- #051 — Git-like Dispatcher Architecture (dispatcher pattern this builds on)
- #155 — Plugin Prerequisites Validation (Python version check is a prerequisite)
- #097 — PTX-Python Helper (provides Python environment management utilities)

---

**Created**: 2026-03-20
**Author**: Architect
