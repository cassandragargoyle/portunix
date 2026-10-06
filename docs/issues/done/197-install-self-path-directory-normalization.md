# Issue #197: `install-self --path` Corrupts Install When Given a Directory

## 🎯 Priority

**HIGH** — Breaks first-time installation on Windows. `install.ps1` (the
GitHub distribution installer and the `make deploy-local` fallback) passes an
install **directory** to `portunix install-self --path`, but `install-self`
treats `--path` as a full **file** path. The mismatch either hard-fails
(directory already exists) or silently produces a corrupt layout.

## 📋 Status

- **Created**: 2026-07-09
- **Completed**: 2026-07-09
- **Status**: ✅ Implemented
- **Assignee**: -
- **Branch**: fix/issue-197-install-self-path-normalization (merged to main)
- **Related**:
  - Issue #164 — Windows install.ps1 Usability Overhaul (`internal/done/`)
  - `src/app/selfinstall/install.go`
  - `src/cmd/install-self.go`
  - `scripts/install.ps1`
  - `scripts/deploy-local.py`

## 📝 Problem Description

### Current Situation

`installAllBinaries` writes the main binary with `copyFile(sourcePath,
targetPath)` → `os.Create(targetPath)`, i.e. `--path` is used verbatim as the
destination **file**:

```go
// src/app/selfinstall/install.go
func installAllBinaries(sourcePath, targetPath string) error {
    if err := copyFile(sourcePath, targetPath); err != nil {   // os.Create(targetPath)
        return fmt.Errorf("failed to copy main binary: %w", err)
    }
    ...
}
```

But callers pass a **directory**:

- `scripts/install.ps1` — the interactive path picker sets `$Path = "C:\Portunix"`
  and forwards it as `--path $Path` (`install.ps1:413-416`, `:562`).
- The command help even documents directories: `--path /usr/local/bin`
  (`src/cmd/install-self.go:41,44`).

Yet `GetDefaultInstallPath()` returns a **file** path
(`…\Portunix\portunix.exe`), so the two contracts disagree.

### Reproduction

```powershell
# C:\Portunix exists as an (empty) directory, e.g. left over after undeploy
portunix install-self --silent --path C:\Portunix --add-to-path
# Error: failed to install binaries: failed to copy main binary:
#        open C:\Portunix: is a directory
# exit 1
```

Surfaced through `make deploy-local` after `make undeploy-local`:
`deploy-local.py` finds no existing install, runs `install.ps1`, which invokes
`install-self` with a directory `--path` → generic "INSTALLATION FAILED".

### Failure Modes

| `--path` value | `C:\Portunix` exists as dir? | Result |
| -------------- | ---------------------------- | ------ |
| `C:\Portunix` | yes | hard error `open …: is a directory`, exit 1 |
| `C:\Portunix` | no | silently creates a **file** literally named `C:\Portunix` (no `.exe`); helper binaries land in `C:\` root |
| `C:\Portunix\portunix.exe` | n/a | works (accidental correct usage) |

### Root Cause

`--path` has an ambiguous, undocumented contract. `install-self` assumes a
full binary file path; every real caller (`install.ps1`, the help examples)
supplies an install directory. There is no normalization step to reconcile the
two.

## ✅ Acceptance Criteria

1. `portunix install-self --silent --path <dir>` installs the main binary as
   `<dir>/portunix[.exe]` and helpers alongside it, whether or not `<dir>`
   already exists.
2. `portunix install-self --silent --path <dir>/portunix[.exe]` (full file
   path) continues to work unchanged (back-compat).
3. Interactive install and `GetDefaultInstallPath()` behavior are unchanged.
4. `--add-to-path` adds the install **directory** to PATH in every case.
5. Unit test covers directory input, full-file input, and existing-directory
   input on both Windows and non-Windows filename rules.
6. `make build` succeeds; `go test portunix.ai/app/selfinstall/...` passes.

## 💡 Implementation Notes

- Add `NormalizeTargetPath(path)` in `selfinstall`: if `filepath.Base(path)`
  equals the platform binary name (`portunix.exe` on Windows —
  case-insensitive; `portunix` elsewhere — case-sensitive) treat it as a full
  file path; otherwise treat it as the install directory and append the binary
  name via `filepath.Join`.
- Apply it at the top of `InstallSilent` and after `PromptInstallLocation()`
  in `InstallInteractive` so every entry point (CLI, `install.ps1`, direct API)
  is covered by one choke point.
- Secondary (out of scope, note for follow-up): `install.ps1` runs
  `install-self` **without** `--silent` in the `deploy-local` path, forcing an
  interactive prompt inside a non-interactive subprocess. Consider passing
  `--silent` when invoked non-interactively.

## 🧪 Test Instructions

1. `go test portunix.ai/app/selfinstall/...` — unit tests pass
2. Windows manual (container/VM per methodology):
   - `portunix install-self --silent --path C:\Portunix --add-to-path`
     with `C:\Portunix` pre-created → succeeds, `C:\Portunix\portunix.exe` exists
   - repeat with `C:\Portunix` absent → same result (no stray `C:\Portunix` file)
   - `portunix install-self --silent --path C:\Portunix\portunix.exe` → still works
3. Regression: `make deploy-local` after `make undeploy-local` completes the
   first-time install without error
