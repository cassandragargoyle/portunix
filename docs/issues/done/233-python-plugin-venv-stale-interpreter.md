# Issue #233: Python plugins break silently after a Python upgrade, and force reinstall is not atomic

**Status**: Implemented
**Priority**: High
**Type**: Bug
**Component**: portunix (plugin manager)
**Created**: 2026-10-21
**Closed**: 2026-10-22
**GitHub**: #233
**Related**: cassandragargoyle/portunix-vscode#42, #234 (follow-ups), `docs/testing/acceptance-233.md`

## Summary

Two defects in the plugin manager surfaced together while reinstalling a Python wheel plugin
after an OS upgrade:

1. A Python plugin whose venv points to a system interpreter that no longer exists cannot run,
   but `portunix plugin list` still shows it as `ready`. Nothing detects or repairs it.
2. `portunix plugin install --force` deletes the working installation before it knows that
   the new one will succeed. A failed install leaves no plugin on disk and a stale registry
   entry that `plugin list` still shows as `ready`.

## Problem Description

### 1. Venv bound to a removed interpreter

A Python wheel plugin installed without `plugin.python_version` gets its venv from the host
`python3` (`setupPythonWheelPluginSystem` in `src/app/plugins/manager/manager.go`). The venv
links to that interpreter, for example `/usr/bin/python3.13`. When an OS upgrade replaces the
system Python with 3.14, the link breaks:

```text
$ ~/.portunix/plugins/reco-worker/.venv/bin/python3 -c 'print("ok")'
bash: .../reco-worker/.venv/bin/python3: No such file or directory
```

Observed on one machine with four plugins at once: reco-watcher, reco-worker, graphlens, vox.
All four stay `ready` in `plugin list`. The failure shows only when the plugin is called.

Plugins installed with `plugin.python_version` get a uv-managed interpreter and are not
affected by system upgrades (cabinet, scraper).

### 2. Non-atomic force reinstall

`ForceInstallPlugin` removes the old `.venv` before the prerequisites check and the new venv
setup. When `setupPythonWheelPlugin` fails, the error path removes the whole plugin directory:

```go
if err := m.setupPythonWheelPlugin(manifest, pluginDir); err != nil {
    _ = os.RemoveAll(pluginDir)
    return fmt.Errorf("failed to setup Python wheel plugin: %w", err)
}
```

The registry is not updated, so the plugin is still listed as `ready` with an install path
that no longer exists. Observed output:

```text
Error: failed to install plugin: failed to setup Python wheel plugin:
failed to create virtual environment: exit status 1
The virtual environment was not created successfully because ensurepip is not available.
```

The host had Python 3.14 without the `venv` module; the plugin had no `python_version`, so
uv was not used.

## Proposed Solution

1. **Health check of Python venvs** — `plugin list` and `plugin check` verify that
   `.venv/bin/python3` resolves to an existing interpreter; a broken venv is shown with a
   status such as `broken` and a hint to repair it
2. **Repair** — `portunix plugin repair <name>` (or `install --force` from the installed
   manifest) recreates the venv; prefer uv with the interpreter version the wheel was built
   for, read from the wheel bytecode when the manifest has no `python_version`
3. **Atomic force reinstall** — build the new installation in a staging directory (or keep a
   backup of the old one) and swap only on success; on failure the old installation and its
   registry entry stay untouched
4. **Registry consistency** — when an install fails after the old files were touched, remove
   or mark the registry entry so `plugin list` never reports `ready` for a missing directory
5. **Pre-check** — before the system fallback, check `python3 -m venv --help` *and* `ensurepip`
   availability, and suggest `portunix install uv` before any destructive step

## Acceptance Criteria

- [x] `plugin list` does not show `ready` for a plugin whose venv interpreter is missing
- [x] A broken venv can be repaired with one Portunix command, without the plugin source
- [x] A failed `install --force` leaves the previous installation working and registered
- [x] The registry never points to a plugin directory that does not exist
- [x] Unit tests cover the broken-venv detection and the rollback of a failed force install

## Testing Notes

- Simulate a removed interpreter: create a venv with a temporary interpreter copy, then delete
  the copy
- Simulate a failed venv setup: put a `python3` without `ensurepip` first on `PATH` and hide uv
