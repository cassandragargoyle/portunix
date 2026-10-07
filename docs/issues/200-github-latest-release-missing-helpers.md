# Issue #200: GitHub `latest` Release v2.2.2 Installs Without `ptx-installer` on Windows

## Priority

**HIGH** — Breaks every `portunix install …` on a fresh Windows install from GitHub.
`releases/latest` on GitHub is still `v2.2.2`, whose `install-self` copies only a
hardcoded subset of helpers. `ptx-installer.exe` is not among them, so `portunix
install python` (and any other package) fails right after a successful install.
The fix exists in the code since `40d1375` (in `v2.2.3` … `v2.4.1`), but no release
containing it has been published to GitHub.

## Status

- **Created**: 2026-10-06
- **Status**: Open
- **Assignee**: -
- **Branch**: -
- **Related**:
  - Commit `40d1375` — `fix(selfinstall): install all ptx-* helpers via glob, not a hardcoded list`
  - Issue #197 — `install-self --path` normalization (also not on GitHub yet)
  - `src/app/selfinstall/install.go` — `installAllBinaries`
  - `.claude/commands/cs/deploy-github.md` — the GitHub publication workflow
  - Reported from LeadSonar issue 013 (Windows bootstrap script that installs
    Portunix from `releases/latest`)

## Problem Description

### Current Situation

GitHub releases of `cassandragargoyle/portunix` (2026-10-06):

| Release | Date | Note |
| ------- | ---- | ---- |
| `v2.2.2` | 2026-04-17 | Marked **Latest** |
| `v1.10.7` | 2026-03-08 | |

Local tags go up to `v2.4.1` (2026-05-29). Every tag from `v2.2.3` contains
`40d1375`; `v2.2.2` does not.

The `v2.2.2` Windows archive (`portunix_2.2.2_windows_amd64.zip`) ships 12 helpers
(`ptx-installer`, `ptx-python`, `ptx-make`, `ptx-credential`, `ptx-aiops`, `ptx-pft`,
`ptx-trace`, `ptx-prompting`, `ptx-container`, `ptx-mcp`, `ptx-virt`, `ptx-ansible`),
but its `install-self` installs only four of them:

```text
> portunix.exe install-self --silent --path <dir>\portunix.exe
✓ Main binary (portunix) installed
✓ Helper binary (ptx-container.exe) installed
✓ Helper binary (ptx-mcp.exe) installed
✓ Helper binary (ptx-virt.exe) installed
✓ Helper binary (ptx-ansible.exe) installed
```

The release's own `install.ps1` calls the same `install-self`, so it has the same
result.

### Reproduction

On a clean Windows 11 (Windows Sandbox), as LeadSonar's bootstrap script does:

1. Download `portunix_2.2.2_windows_amd64.zip` from `releases/latest`, verify it
   against `checksums_2.2.2.txt`, unpack
2. `portunix.exe install-self --silent --add-to-path`  installs to
   `C:\Program Files\Portunix` with the four helpers above
3. `portunix install python --variant=full`:

```text
❌ Error: ptx-installer helper not found

Expected location: C:\Program Files\Portunix\ptx-installer.exe

The install command requires the ptx-installer helper binary.
Please ensure ptx-installer is in the same directory as portunix.
```

### Impact

- Anyone installing Portunix on Windows from GitHub cannot install any package
- `portunix update` from GitHub stays on `v2.2.2`, so existing installs do not heal
- Downstream scripts that resolve `releases/latest` (LeadSonar `scripts/install.ps1`)
  have to work around it by copying the missing `ptx-*.exe` from the archive

## Proposed Solution

1. Publish the current release (`v2.4.1` or newer) to GitHub through the
   `deploy-github` workflow, so `releases/latest` contains `40d1375` and #197
2. Before publishing, verify on a clean Windows (Windows Sandbox) that
   `install-self` from the release archive installs **every** `ptx-*.exe` the
   archive contains, and that `portunix install python --variant=full` works
3. Add that check to the release process, so a release whose installed helper set
   differs from the archive's cannot be published again

## Acceptance Criteria

- [ ] GitHub `releases/latest` is a version containing `40d1375`
- [ ] On a clean Windows, `install-self` from the latest GitHub archive installs all
      `ptx-*.exe` helpers of the archive
- [ ] `portunix install python --variant=full` works right after that install
- [ ] `portunix update` on a `v2.2.2` install reaches the new release and leaves all
      helpers in place
- [ ] The release workflow checks the installed helpers against the archive
