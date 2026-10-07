# Issue #139: Tea (Gitea CLI) Installation Support

## Status

✅ Implemented

## Priority

Medium

## Type

Enhancement

## Labels

enhancement, package-management, gitea, cli, developer-tools, cross-platform

## Summary

Add support for installing `tea`, the official Gitea CLI tool, via Portunix package management system. This enables automated release management and repository operations for Gitea-hosted projects.

## Problem Statement

Currently, creating releases on Gitea requires either:

1. Manual upload via web UI
2. Direct API calls via curl (requires token management)

Having `tea` CLI available would enable:

- Automated release creation with asset uploads
- Repository management from command line
- Issue and PR management
- Consistent workflow similar to GitHub's `gh` CLI

## Proposed Solution

Add `tea` package to Portunix registry with cross-platform installation support.

### Installation Methods

**Linux (amd64/arm64):**

- Download from Gitea releases: https://gitea.com/gitea/tea/releases
- Binary installation to `~/.local/bin` or system path

**Windows:**

- Download `.exe` from Gitea releases
- Optional: Chocolatey/Scoop if available

**macOS:**

- Homebrew: `brew install tea`
- Binary download as fallback

### Package Registry Entry

```json
{
  "name": "tea",
  "description": "Gitea CLI - command line tool for Gitea",
  "homepage": "https://gitea.com/gitea/tea",
  "category": "developer-tools",
  "platforms": {
    "linux": {
      "method": "binary",
      "url": "https://gitea.com/gitea/tea/releases/download/v{version}/tea-{version}-linux-amd64",
      "bin_name": "tea"
    },
    "windows": {
      "method": "binary",
      "url": "https://gitea.com/gitea/tea/releases/download/v{version}/tea-{version}-windows-amd64.exe",
      "bin_name": "tea.exe"
    },
    "darwin": {
      "method": "homebrew",
      "package": "tea"
    }
  }
}
```

## Acceptance Criteria

- [x] `portunix install tea` works on Linux (amd64, arm64)
- [ ] `portunix install tea` works on Windows
- [ ] `portunix install tea` works on macOS
- [x] Version detection works correctly
- [x] `tea --version` returns expected output after installation

## Use Cases

1. **Release automation**: `tea release create v1.9.1 --asset dist/*.tar.gz`
2. **Repository management**: `tea repo list`, `tea repo create`
3. **Issue management**: `tea issue list`, `tea issue create`
4. **CI/CD integration**: Automated releases from scripts

## References

- Tea repository: <https://gitea.com/gitea/tea>
- Tea releases: <https://gitea.com/gitea/tea/releases>
- Tea documentation: <https://gitea.com/gitea/tea/src/branch/main/README.md>
- Similar: Issue #026 (GitHub CLI installation)

## Implementation Notes

- Package definition: `src/helpers/ptx-installer/assets/packages/tea.json`
- Version: 0.11.1
- Linux: curl download to `/usr/local/bin/tea`
- Windows: exe download to `C:/Program Files/tea`
- macOS: Homebrew (`brew install tea`)
- Tested on: Linux amd64

---

**Created**: 2025-01-21
**Implemented**: 2025-01-21
**Author**: ZK
