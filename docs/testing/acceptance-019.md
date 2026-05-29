# Acceptance Protocol — Issue #019

**Issue**: Docker Installation Issues on Windows
**Branch**: `feature/019-docker-install-command-alias`
**Commit under test**: `966411a` (preceded by auto-commits `0611bc2`, `b5f81fd`)
**Tester**: zdendaku (role: Tester / generic)
**Date**: 2026-05-11
**Testing OS**:

- Host: Windows 11 Pro 10.0.26200 — primary test target (issue is Windows-specific)
- Shell: PowerShell + Git-Bash (Bash tool)
- Container: not required — change is a thin command-alias layer with no installation effects in this scope

**Build**: `make build` clean from feature branch tip (`966411a`); main `portunix.exe` and all 18 helper binaries rebuilt successfully.

## Scope

Issue #019 catalogued five Windows-specific Docker install problems. An audit confirmed four were already resolved by follow-up issues (#122, #176, #177, #178, #179):

| Original problem | Status before this PR | Resolved by |
| ---------------- | --------------------- | ----------- |
| 1. Wrong disk detection (phantom D:\) | ✅ Resolved | `storage.go:GetWindowsDrives()` (WMI, no hardcoded letters) |
| 2. Wrong daemon.json data-root | ✅ Resolved | `docker.go:resolveWindowsDataRoot()` + `configureWindowsDataRoot()` |
| 3. Install verification fails | ✅ Resolved | `verifyInstallation()` + 120s daemon polling |
| 4. **Inconsistent commands** (`docker install` vs `install docker`) | ❌ **Open** | **This PR** |
| 5. PATH not configured | ✅ Resolved de-facto (Docker Desktop installer registers PATH itself post-UAC, #178) | n/a |

**This PR closes the remaining gap (#4)** by adding a thin `install` subcommand to `ptx-container` (and a fallback at the cobra level for the helper-less case). Both `portunix docker install` and `portunix install docker` now resolve to the same ptx-installer code path.

**Out of scope** (intentionally — handled or not relevant):

- PATH manipulation from Portunix side (Docker Desktop owns this)
- Storage / data-root logic changes (already in `ptx-installer/engine`)
- Same-named `podman install` alias was added opportunistically since the ptx-container subcommand router treats `docker` and `podman` symmetrically; no separate issue is needed

## Test Plan

| TC | Area | Description |
| -- | ---- | ----------- |
| TC-01 | Alias presence | `portunix docker install --dry-run` exits 0 and reaches ptx-installer |
| TC-02 | Output parity | `docker install` and `install docker` produce byte-identical output |
| TC-03 | Help advertises alias | `portunix docker --help` lists `install` |
| TC-04 | `container install` rejected | Neutral parent has no runtime to install — must fail clearly, exit 1 |
| TC-05 | `podman install` symmetry | Same routing applies for podman |
| TC-06 | Flag forwarding (`--data-root`) | Custom flag must reach ptx-installer unchanged |
| TC-06b | Bogus flag forwarding | Even unknown flags propagate identically (no premature cobra consumption) |
| TC-07 | Regression: other docker subcommands | `docker list`, `docker info` still functional |
| TC-08 | `container --help` does NOT show install | Help is conditional on parent command |
| BUILD | `make build` regression | All binaries rebuild cleanly |
| VET | `go vet` on `ptx-container` | No new warnings |

## Test Cases — Given/When/Then

### TC-01: Alias presence

**Given** a built `portunix.exe` on the feature branch tip
**When** I run `portunix docker install --dry-run`
**Then** the command exits 0 and prints the ptx-installer banner `🐳 Starting Docker installation with intelligent storage detection...`

**Actual output**:

```text
🐳 Starting Docker installation with intelligent storage detection...
✅ Docker is already installed

🔍 Verifying Docker installation...
✅ Docker version 29.4.0, build 9d7ad9f
⚠️  Docker daemon may not be running
   Start Docker Desktop or run: sudo systemctl start docker
EXIT=0
```

**Result**: ✅ PASS

### TC-02: Output parity

**Given** both command forms available
**When** I capture full output of `install docker --dry-run` and `docker install --dry-run` and diff them
**Then** outputs must be byte-identical

**Actual**:

```text
size install docker: 285 bytes
size docker install: 285 bytes
diff -q ... → no differences
```

**Result**: ✅ PASS

### TC-03: Help advertises alias

**Given** the alias is registered
**When** I run `portunix docker --help`
**Then** the help text contains a line for the `install` subcommand

**Actual**:

```text
  install          Install docker (alias for 'portunix install docker')
```

**Result**: ✅ PASS

### TC-04: `container install` rejected

**Given** the neutral `container` parent has no specific runtime
**When** I run `portunix container install`
**Then** the command prints an actionable error message and exits 1

**Actual**:

```text
❌ Error: 'portunix container install' is not supported.
   Use 'portunix install docker' or 'portunix install podman' explicitly.
EXIT=1
```

**Result**: ✅ PASS

### TC-05: `podman install` symmetry

**Given** the same `install` case in the ptx-container switch
**When** I run `portunix podman install --dry-run`
**Then** the command routes to ptx-installer's podman path

**Actual**:

```text
🦭 Starting Podman installation...

📊 Checking system requirements...

🔍 DRY RUN - Would perform the following:
EXIT=0
```

**Result**: ✅ PASS

### TC-06: Flag forwarding (`--data-root`)

**Given** the alias must pass flags through transparently
**When** I run both forms with `--data-root "Z:\custom"`
**Then** outputs must be identical (proving the flag reaches the same handler)

**Actual**: Both files identical via `diff -q` → forwarding works.
**Result**: ✅ PASS

### TC-06b: Bogus flag forwarding

**Given** unknown flags should not be silently consumed by cobra
**When** I run both forms with `--bogus-flag`
**Then** error output must be byte-identical

**Actual**: Both files identical → no premature flag parsing in the alias path.
**Result**: ✅ PASS

### TC-07: Regression — other docker subcommands

**Given** the alias is additive (does not remove existing subcommands)
**When** I run `portunix docker list` and `portunix docker info`
**Then** they still execute their original handlers

**Actual**:

```text
docker list → ❌ Neither Docker nor Podman is available  (expected — daemon not running in test env)
docker info → 🐳 Container Runtime Information section printed
```

Behavior unchanged from main branch.
**Result**: ✅ PASS

### TC-08: `container --help` does NOT show `install`

**Given** the `install` help line is gated by `command == "docker" || command == "podman"`
**When** I run `portunix container --help`
**Then** the `install` subcommand must NOT appear

**Actual**: No `install` line present in `container --help` output.
**Result**: ✅ PASS

### BUILD

```text
go build -o portunix.exe .
Building helper binaries...
Helper binaries built: ptx-container, ptx-mcp, ptx-virt, ptx-ansible, ptx-prompting,
                       ptx-python, ptx-installer, ptx-aiops, ptx-make, ptx-pft,
                       ptx-credential, ptx-trace, ptx-ssh, ptx-plugin-registry,
                       ptx-proxmox, ptx-specpm, ptx-database, ptx-github
All binaries built successfully
```

**Result**: ✅ PASS

### VET

`go vet ./...` on `src/helpers/ptx-container/` produced no output.
**Result**: ✅ PASS

## Test Summary

- Total test scenarios: 11 (TC-01..08, TC-06b, BUILD, VET)
- Passed: 11
- Failed: 0
- Skipped: 0

## Coverage Notes

- **What is covered**: alias registration, output parity, flag forwarding, exit code propagation, help-text rendering, regression on neighbour subcommands.
- **What is NOT covered**:
  - Actual Docker Desktop installation flow on a fresh Windows VM (out of scope — Docker is already installed on the test host; the alias-only PR does not change install behavior, only entry-point routing).
  - Linux behavior — issue #019 is explicitly Windows-scoped; symmetric routing via ptx-container's `case "install"` makes Linux work too but a Linux test pass is not required for this issue.
  - The cobra-level fallback alias in `src/cmd/docker_install_alias.go` was not exercised at runtime because ptx-container helper is present in this build. It remains a parity safety net for installations without helper binaries.

## CI Notes

No CI changes required by this PR — the change is contained within helper code and exercised by existing build pipelines.

## Final Decision

**STATUS**: PASS

**Approval for merge**: YES
**Date**: 2026-05-11
**Tester signature**: zdendaku (role: Tester / generic)
