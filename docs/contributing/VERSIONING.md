# Versioning Guide for Contributors

This document explains Portunix versioning conventions for contributors.

## Overview

Portunix uses **GitHub as the single source of truth** for stable versions. Contributors working on internal systems (personal repos, Gitea, GitLab, etc.)
use development version suffixes to avoid conflicts.

## Version Formats

### Stable Versions (GitHub only)

```text
v{MAJOR}.{MINOR}.{PATCH}
```

**Examples:** `v1.9.2`, `v1.10.0`, `v2.0.0`

- Created ONLY when publishing to GitHub
- Follow [Semantic Versioning 2.0.0](https://semver.org/)
- Never created on internal/development repositories

### Development Versions (Your local repo)

```text
v{LAST_GITHUB_VERSION}+dev.{N}
```

**Examples:** `v1.9.2+dev.1`, `v1.9.2+dev.2`, `v1.9.2+dev.15`

- `{LAST_GITHUB_VERSION}` = last stable version from GitHub
- `{N}` = your sequential development iteration number
- The `+dev.N` suffix marks it as development version
- These versions are **never** published to GitHub

### Pre-release Versions (Optional)

```text
v{TARGET_VERSION}-rc.{N}
```

**Examples:** `v1.10.0-rc.1`, `v1.10.0-rc.2`

- Used when target version is decided and release is imminent
- `rc` = release candidate

## Quick Start Workflow

### 1. Check Current GitHub Version

Before starting development, check the latest GitHub version:

```bash
# If you have GitHub remote configured
git fetch github
git describe --tags github/main --abbrev=0
# Example output: v1.9.2
```

Or check [GitHub Releases](https://github.com/cassandragargoyle/portunix/releases).

### 2. Start Development

Use the GitHub version as your base with `+dev.1` suffix:

```bash
# Build with development version
./build-with-version.sh v1.9.2+dev.1

# Optional: tag your work
git tag v1.9.2+dev.1
```

### 3. Continue Development

Increment the dev number for each milestone:

```text
v1.9.2+dev.1  → First iteration
v1.9.2+dev.2  → After completing a feature
v1.9.2+dev.3  → After bug fixes
...
```

### 4. Submit Your Contribution

When ready to contribute:

1. Create a Pull Request to GitHub
2. **Do NOT include version tags** in your PR
3. Maintainers will determine the final version at release time

## Version Hierarchy

```text
v1.9.2           ← GitHub stable (source of truth)
    ↓
v1.9.2+dev.1     ← Your development
v1.9.2+dev.2
v1.9.2+dev.3
    ↓
v1.10.0-rc.1     ← Optional: release candidate
v1.10.0-rc.2
    ↓
v1.10.0          ← Next GitHub stable (decided at release)
```

## Examples

### Example 1: Adding a New Feature

```text
1. GitHub has v1.9.2
2. You start: v1.9.2+dev.1
3. Feature complete: v1.9.2+dev.3
4. Submit PR to GitHub
5. Maintainer releases as v1.10.0 (minor bump for new feature)
```

### Example 2: Bug Fix

```text
1. GitHub has v1.10.0
2. You start: v1.10.0+dev.1
3. Fix complete: v1.10.0+dev.1
4. Submit PR to GitHub
5. Maintainer releases as v1.10.1 (patch bump for bug fix)
```

### Example 3: Multiple Contributors

```text
Contributor A: v1.10.0+dev.1, v1.10.0+dev.2
Contributor B: v1.10.0+dev.1, v1.10.0+dev.2
Contributor C: v1.10.0+dev.1

All submit PRs → Maintainer merges and releases v1.11.0
```

No conflicts because each contributor's `+dev.N` is local.

## Important Rules

### DO

- Always base your development version on the latest GitHub version
- Use `+dev.N` suffix for all internal/development versions
- Increment dev number for significant milestones
- Check GitHub version before starting new development

### DON'T

- Never create "clean" versions (without suffix) in your local repo
- Never push development versions (`+dev.N`) to GitHub
- Never assume what the final version number will be

## Semantic Versioning

Final version numbers follow [Semantic Versioning](https://semver.org/):

| Change Type | Version Bump | Example |
| ----------- | ------------ | ------- |
| Breaking API changes | MAJOR | v1.9.2 → v2.0.0 |
| New features (backward compatible) | MINOR | v1.9.2 → v1.10.0 |
| Bug fixes only | PATCH | v1.9.2 → v1.9.3 |

**Note:** The version bump is decided by maintainers at release time, not during development.

## FAQ

### Q: What if GitHub releases a new version while I'm developing?

Rebase your work and update your base version:

```bash
# GitHub released v1.9.3 while you were on v1.9.2+dev.5
git fetch github
git rebase github/main
# Continue as v1.9.3+dev.1
```

### Q: Can I use my own versioning internally?

Yes, as long as you don't push those versions to GitHub. The `+dev.N` convention is recommended but your internal system is your choice.

### Q: Who decides the final version number?

Project maintainers decide the version number when creating a GitHub release, based on the accumulated changes.

### Q: What about hotfixes?

Use `+hotfix.N` suffix for hotfix branches:

```text
v1.10.0+hotfix.1 → Released as v1.10.1
```

## Related Documentation

- [Semantic Versioning 2.0.0](https://semver.org/)
- [ADR-036: Version Management Strategy](../adr/036-version-management-strategy.md)

---

Last updated: 2026-01-23
