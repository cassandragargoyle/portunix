# Issue #201: Windows Inline `installScript` Is Not Executed, Yet Reported as Success (`portunix install uv`)

## Priority

**HIGH** — `portunix install uv` on Windows installs nothing and prints
`✅ Installation completed successfully!`. Every package of type `script` with an
inline command containing double quotes is affected the same way, and the false
success hides it from callers that rely on the exit code.

## Status

- **Created**: 2026-10-06
- **Status**: Implemented
- **Closed**: 2026-10-07
- **Assignee**: zdendaku
- **Branch**: fix/201-windows-inline-install-script (merged)
- **Related**:
  - Acceptance protocol: `docs/testing/acceptance-201.md` (PASS)
  - `src/helpers/ptx-installer/engine/installer.go` — `installScript`, `Install` (type dispatch)
  - `src/helpers/ptx-installer/assets/packages/uv.json`
  - Issue #200 — found in the same run (LeadSonar issue 013, clean Windows in Windows Sandbox)

## Problem Description

### Current Situation

`uv.json` defines the Windows variant as an inline script:

```json
"installScript": "powershell -ExecutionPolicy ByPass -c \"irm https://astral.sh/uv/install.ps1 | iex\""
```

`installScript` runs inline commands on Windows as
`exec.Command("cmd", "/c", expandedScript)`. Go builds the command line with its
`EscapeArg` rules, which turn every `"` inside the argument into `\"`. `cmd.exe` does
not understand `\"`, so PowerShell receives the script text as a **string literal**
and only prints it. The installer of uv never runs.

Afterwards `Install` returns success: neither `verification` (`uv --version`,
exit code 0) nor `postInstall` of the package is executed, so nothing notices that
`uv` does not exist.

### Reproduction

Clean Windows 11 (Windows Sandbox), Portunix 2.2.2 with all helpers:

```text
> portunix install uv
🚀 Starting installation (type: script)...
📝 Running install script (target: ./site)...
   ✅ Running: powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 |...
irm https://astral.sh/uv/install.ps1 | iex
✅ Script installation completed
✅ Installation completed successfully!

> Test-Path "$env:USERPROFILE\.local\bin\uv.exe"
False
```

The printed line `irm https://astral.sh/uv/install.ps1 | iex` is PowerShell echoing
the string. The same command typed into PowerShell by hand works:

```text
> powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
downloading uv 0.12.23 (x86_64-pc-windows-msvc)
installing to C:\Users\WDAGUtilityAccount\.local\bin
  uv.exe
  uvx.exe
  uvw.exe
everything's installed!
```

Also visible: the target is reported as `./site`, the default of `installScript`,
which means nothing for uv.

### Impact

- `portunix install uv` on Windows never installs uv, and says it did
- Any other inline Windows `installScript` with quoted arguments is broken the same way
- Callers (scripts, `ptx-make`, LeadSonar `scripts/install.ps1`) cannot trust the exit
  code and need their own check and fallback

## Proposed Solution

1. **Run the inline command verbatim on Windows.** Set the raw command line instead of
   letting Go escape it, e.g.
   `cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "cmd /c " + expandedScript}`
   (build-tagged for Windows), or run PowerShell scripts through `powershell -Command`
   with the script passed so that no re-quoting happens
2. **Verify after every install.** Run the package's `verification` command after
   `installScript` (and the other install types) and fail with a non-zero exit code
   when it fails. Refresh `PATH` from the registry for the check, so a tool installed
   into a directory just added to the user `PATH` (`%USERPROFILE%\.local\bin`) is found
3. **Run `postInstall`** as the package defines it, or remove it from the manifests
   where it is not meant to run
4. Do not print `target: ./site` for inline scripts that do not use `${INSTALL_PATH}`

## Acceptance Criteria

- [x] On a clean Windows, `portunix install uv` installs `uv.exe` into
      `%USERPROFILE%\.local\bin` and `uv --version` works in a new terminal
- [x] When an install script does nothing (verification fails), `portunix install`
      exits non-zero and says verification failed
- [x] A unit test covers an inline Windows script with double-quoted arguments
      reaching the shell unchanged
- [x] Other `type: script` packages with Windows inline scripts are checked for the
      same problem
