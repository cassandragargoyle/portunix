# Issue #202: Python Package — Install Python 3.14.8 and Honour the Variant's `installArgs` on Windows

## Priority

**MEDIUM** — `portunix install python` on Windows installs Python 3.13.6, while the
newest stable release is 3.14.8. The `full` variant also runs the Python installer
with generic switches instead of the ones the package defines, so the unattended
install is not the one intended and was seen to take about 18 minutes.

## Status

- **Created**: 2026-10-06
- **Status**: Implemented
- **Closed**: 2026-10-07
- **Assignee**: zdendaku
- **Branch**: feature/202-python-3-14-8-install-args (merged)
- **Related**:
  - Acceptance protocol: `docs/testing/acceptance-202.md` (PASS after re-test)
  - `src/helpers/ptx-installer/assets/packages/python.json`
  - `src/helpers/ptx-installer/engine/installer.go` — `installWindowsBinary`, `runExeInstaller`
  - Issues #200, #201 — found in the same run (LeadSonar issue 013, clean Windows in
    Windows Sandbox)

## Problem Description

### 1. The Windows version is 3.13.6

Both Windows variants of `python.json` are pinned to `3.13.6`:

| Variant | Version | URL |
| ------- | ------- | --- |
| `embeddable` | `3.13.6` | `python-3.13.6-embed-amd64.zip` |
| `full` | `3.13.6` | `python-3.13.6-amd64.exe` |

On 2026-10-06 the newest stable Python is **3.14.8**
(`https://www.python.org/ftp/python/3.14.8/python-3.14.8-amd64.exe` answers 200). The
`3.15.0` directory exists on python.org, but has no Windows installer yet (404), so it
is not a candidate.

### 2. The `full` variant's `installArgs` are ignored

`python.json` defines for `full`:

```json
"installArgs": ["/quiet", "InstallAllUsers=1", "PrependPath=1", "Include_test=0"]
```

`installWindowsBinary` passes `platform.InstallArgs` to `runMsiInstaller` /
`runExeInstaller`, not `variant.InstallArgs`. The platform level of `python.json` has
none, so `runExeInstaller` falls back to its defaults:

```text
🔧 Installing EXE package...
   Running: C:\Users\WDAGUtilityAccount\.portunix\cache\python-3.13.6-amd64.exe /S /silent /quiet
```

`/S` and `/silent` are NSIS / Inno Setup switches; the Python installer (WiX Burn)
knows `/quiet` and `/passive`. `PrependPath=1` is lost, so Python is not put on `PATH`
by the installer. In the observed run the installer started at 11:11 and returned at
11:29.

## Proposed Solution

1. Update both Windows variants of `python.json` to **3.14.8**:
   - `full`: `python-3.14.8-amd64.exe` / `python-3.14.8.exe`
   - `embeddable`: `python-3.14.8-embed-amd64.zip` / `-embed-win32.zip`, and the
     `postInstall` line that edits `python313._pth` must edit `python314._pth`
   - add an `arm64` URL (`python-3.14.8-arm64.exe`) if the variant format allows it
2. In `installWindowsBinary`, use `variant.InstallArgs` when set and fall back to
   `platform.InstallArgs`, the same precedence `installArgs` has elsewhere
   (`installer.go:841`)
3. Check the other `type: exe` / `msi` packages whose `installArgs` sit on the variant
   level, they are affected the same way
4. Add a Windows variant **`latest`** — the `full` installer, but of the newest stable
   Python, resolved at install time instead of being pinned in the manifest:
   - new variant field `versionResolver` (first resolver: `python.org`); the resolved
     version replaces `{version}` in `url` / `urls`
   - the `python.org` resolver reads `https://www.python.org/ftp/python/`, sorts the
     `X.Y.Z/` directories newest first and takes the first one whose installer for the
     current architecture answers 200 (skips upcoming releases such as `3.15.0`
     whose installers are not published yet)
   - `latest` is the **default** Windows variant (`"preferred": true`), so
     `portunix install python` without `--variant` installs it
   - new variant field `fallbackVariant`: when the resolver fails (python.org
     unreachable), the named pinned variant is installed instead with a warning;
     `latest` falls back to `full`
5. `autoDetectVariant` honours `preferred` and no longer picks a random variant from
   the map when neither `default` nor `standard` exists

## Acceptance Criteria

- [x] On a clean Windows, `portunix install python --variant=full` installs Python
      3.14.8 and `python --version` in a new terminal prints `Python 3.14.8`
- [x] The log shows the installer run with
      `/quiet InstallAllUsers=1 PrependPath=1 Include_test=0`
- [x] `portunix install python --variant=embeddable` installs 3.14.8 with a working `pip`
- [x] `portunix install python` (no variant) selects `latest`, resolves the newest
      stable version from python.org and installs it like `full`
- [x] `portunix install python --variant=latest --dry-run` prints the resolved version
      and download URL
- [x] A unit test covers variant-level `installArgs` taking precedence over the
      platform level
- [x] When python.org cannot be reached, `portunix install python` warns and installs
      the pinned `full` variant instead of failing
- [x] Unit tests cover the `python.org` resolver (newest first, unpublished versions
      skipped) without network access
