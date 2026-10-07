# Issue #140: Version Management Strategy Implementation (ADR-036)

**Status**: ✅ Implemented
**Closed**: 2026-05-19
**Priority**: High
**Type**: Enhancement
**Labels**: enhancement, versioning, workflow, contributor-experience, adr-implementation

## Summary

Implement the version management strategy defined in ADR-036 to establish clear versioning rules between internal development (Gitea) and public releases (GitHub).

## Related ADR

- [ADR-036: Version Management Strategy](../../adr/036-version-management-strategy.md)

## Background

Currently versions/tags are created independently on Gitea and GitHub, leading to confusion. ADR-036 defines:

- **GitHub** = single source of truth for stable versions (`v1.9.2`)
- **Internal repos** = development versions with suffix (`v1.9.2+dev.1`)
- Clear workflow for version progression

## Implementation Tasks

### Phase 1: Documentation (Priority: High)

- [x] Create `docs/contributing/VERSIONING.md` for contributors
- [ ] Update main `CONTRIBUTING.md` to reference versioning guide
- [ ] Add versioning section to README.md

### Phase 2: Skill Updates (Priority: High)

- [ ] Update `/cs:release-gitea` skill:
  - Add validation for `+dev.N` suffix
  - Add step to check last GitHub version
  - Prevent creating clean versions without suffix

- [ ] Update `/cs:deploy-github` skill:
  - Add validation for clean version (no suffix)
  - Add step to determine version bump type (major/minor/patch)
  - Add verification that version doesn't exist on GitHub

### Phase 3: Build Script Updates (Priority: Medium)

- [ ] Update `build-with-version.sh`:
  - Add version format validation
  - Add context parameter (github vs internal)
  - Warn when using clean version outside GitHub context

- [ ] Update `scripts/make-release.py`:
  - Add version format validation
  - Support `+dev.N` suffix for internal releases

### Phase 4: Validation Tooling (Priority: Medium)

- [ ] Create helper command `portunix version check`:
  - Show current local version
  - Show last GitHub version
  - Suggest next dev version number

- [ ] Add pre-commit hook (optional):
  - Validate version format in commits
  - Warn about version conflicts

### Phase 5: Migration (Priority: Low)

- [ ] Clean up existing Gitea tags that don't follow convention
- [ ] Document migration path for existing contributors
- [ ] Update any CI/CD scripts that reference versions

## Acceptance Criteria

1. ✅ `docs/contributing/VERSIONING.md` exists and is comprehensive
2. [ ] Skills enforce versioning rules automatically
3. [ ] Build scripts validate version format
4. [ ] No clean versions can be created on Gitea accidentally
5. [ ] Contributors can easily determine correct version to use

## Technical Notes

### Version Format Regex

```regex
# Stable version (GitHub only)
^v\d+\.\d+\.\d+$

# Development version (internal)
^v\d+\.\d+\.\d+\+dev\.\d+$

# Pre-release version
^v\d+\.\d+\.\d+-(rc|alpha|beta)\.\d+$
```

### Example Workflow

```bash
# Check GitHub version
git fetch github && git describe --tags github/main --abbrev=0
# v1.9.2

# Start development
./build-with-version.sh v1.9.2+dev.1

# Continue development
./build-with-version.sh v1.9.2+dev.2

# Ready to release to GitHub
./build-with-version.sh v1.10.0  # Only via deploy-github skill
```

## Dependencies

- ADR-036 (Accepted)
- Existing skill files in `.claude/commands/cs/`

## Estimated Effort

- Phase 1: Done
- Phase 2: Medium (skill modifications)
- Phase 3: Low (minor script changes)
- Phase 4: Medium (new tooling)
- Phase 5: Low (cleanup)

---

**Created**: 2026-01-23
**Author**: Zdeněk
