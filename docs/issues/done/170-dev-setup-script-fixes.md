# Issue #170: Dev Setup Script Fixes

- **Type**: Bug Fix
- **Priority**: Medium
- **Status**: ✅ Implemented
- **Labels**: bug, dev-setup, package-management, json, developer-experience
- **Created**: 2026-04-08

## Description

The `scripts/dev-setup.sh` and `scripts/dev-setup.ps1` environment setup scripts have issues that prevent
successful installation of Node.js and related npm-based tools (markdownlint-cli2).

## Problems Found

### 1. Wrong Package Name for Node.js

The dev-setup scripts use `node` as the package name, but the ptx-installer registry defines it as `nodejs`.

**Error output:**

```text
[INFO] Installing node via portunix...
[ERROR] Failed to install node
Installing package: node
Installation failed: package not found: package 'node' not found
```

**Fix:** Change package name from `node` to `nodejs` in both `dev-setup.sh` and `dev-setup.ps1`.

### 2. Trailing Comma in hugo.json Causes Parse Error

The embedded package file `src/helpers/ptx-installer/assets/packages/hugo.json` contains trailing commas
in the `extractTo` fields (lines 27 and 35), which is invalid JSON.

**Error output:**

```text
Warning: Failed to load embedded package hugo.json: failed to parse embedded package:
invalid character '}' looking for beginning of object key string
Embedded package discovery complete: 53 packages loaded, 1 errors
```

**Location:** `src/helpers/ptx-installer/assets/packages/hugo.json`, lines 26-27 and 34-35:

```json
"extractTo": "C:/Portunix/bin/hugo",
```

The trailing comma before `}` is invalid in strict JSON.

**Fix:** Remove trailing commas from lines 27 and 35.

## Files to Modify

1. `scripts/dev-setup.sh` - Fix package name `node` -> `nodejs`
2. `scripts/dev-setup.ps1` - Fix package name `node` -> `nodejs`
3. `src/helpers/ptx-installer/assets/packages/hugo.json` - Remove trailing commas (lines 27, 35)

## Acceptance Criteria

- [ ] `portunix install nodejs` works correctly from dev-setup scripts
- [ ] hugo.json loads without parse errors (54 packages loaded, 0 errors)
- [ ] `scripts/dev-setup.sh --dry-run` completes without errors
- [ ] `scripts/dev-setup.ps1 -DryRun` completes without errors
- [ ] markdownlint-cli2 installs successfully after nodejs is available
