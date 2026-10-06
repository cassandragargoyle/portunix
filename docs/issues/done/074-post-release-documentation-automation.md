# Issue #74: Post-Release Documentation Automation and Static Site Generation

**Status**: 📋 Open
**Priority**: High
**Type**: Feature
**Labels**: enhancement, documentation, automation, release-process, github-pages, static-site
**Created**: 2025-09-26
**Updated**: 2025-09-26

## Problem Statement

Current documentation for Portunix ecosystem is scattered and manually maintained:
- Command documentation exists only in `--help` outputs
- No centralized, searchable command reference
- Manual documentation sync after releases
- Plugin commands not documented in central location
- No professional web presence for documentation

## Solution Overview

Implement automated static documentation site generation as part of the release process:

1. **Post-release script** (`scripts/post-release-docs.sh`)
2. **Command discovery system** (core + plugins)
3. **Static site generation** (Hugo-based)
4. **GitHub Pages deployment** (automated)

## Architecture Reference

Based on **ADR-018: Post-Release Documentation Automation and Static Site Generation**

### Workflow Integration
```
Current:  make-release.sh → [manual GitHub upload]
Proposed: make-release.sh → post-release-docs.sh → GitHub Pages
```

### Technical Components
```
post-release-docs.sh:
  ├── discover-commands()      # Core + Plugin command discovery
  ├── generate-command-docs()  # Auto-generate from --help outputs
  ├── build-static-site()      # Hugo/Jekyll build
  └── deploy-github-pages()    # Git push to gh-pages branch
```

## Acceptance Criteria

### Phase 1: Core Documentation Generator
- [ ] Create `scripts/post-release-docs.sh` script
- [ ] Implement command discovery from `portunix --help` parsing
- [ ] Generate structured command documentation (JSON/Markdown)
- [ ] Basic Hugo site generation with command docs
- [ ] Local testing capability (`hugo server`)

### Phase 2: Plugin Integration
- [ ] Plugin command discovery via gRPC API
- [ ] Plugin-specific documentation inclusion
- [ ] Cross-referencing between core and plugins
- [ ] Graceful handling of non-responsive plugins

### Phase 3: GitHub Pages Automation
- [ ] Automated `gh-pages` branch management
- [ ] Site deployment pipeline integration
- [ ] Archive management for old documentation versions
- [ ] Integration with `make-release.sh` workflow

### Quality Requirements
- [ ] Documentation generation doesn't block releases (fallback handling)
- [ ] Site builds locally for testing before deployment
- [ ] Proper error handling and logging
- [ ] Performance: Documentation generation adds <3 minutes to release process

## Technical Implementation

### Dependencies
**Build Dependencies**:
- Hugo static site generator (`hugo` command)
- Git (for gh-pages deployment)
- jq (for JSON processing)

**Runtime Dependencies**:
- Functioning Portunix binary (for command discovery)
- Plugin system active (for plugin discovery)

### Site Structure
```
docs-site/
├── hugo.toml              # Hugo configuration
├── content/
│   ├── commands/          # Auto-generated command docs
│   │   ├── core/          # portunix install, container, etc.
│   │   └── plugins/       # portunix agile, etc.
│   ├── guides/            # Manual documentation
│   └── releases/          # Release notes archive
├── layouts/               # Hugo templates
└── static/                # CSS, images, etc.
```

### GitHub Integration
**Repository**: Use existing `cassandragargoyle/portunix` repository
**Branch Strategy**: `gh-pages` branch for static site
**URL**: `https://cassandragargoyle.github.io/portunix/`

### Modified Release Process
Update `scripts/make-release.sh` to call documentation script:
```bash
# At end of make-release.sh
if [ "$AUTO_DOCS" != "false" ]; then
    print_step "Generating documentation site..."
    ./scripts/post-release-docs.sh "$VERSION"
fi
```

## Test Cases

### Unit Testing
- [ ] Command discovery parsing accuracy
- [ ] Plugin API communication handling
- [ ] Hugo site generation from templates
- [ ] Git operations for gh-pages deployment

### Integration Testing
- [ ] Full end-to-end documentation generation
- [ ] Release process integration (with/without AUTO_DOCS)
- [ ] GitHub Pages deployment verification
- [ ] Multi-platform Hugo compatibility

### Edge Cases
- [ ] Plugin system unavailable/unresponsive
- [ ] Git repository issues (permissions, network)
- [ ] Hugo build failures
- [ ] GitHub Pages size limits

## Resources

### Documentation
- Hugo documentation: https://gohugo.io/
- GitHub Pages setup guide
- Current release workflow: `scripts/make-release.sh`

### Files to Modify/Create
- `scripts/post-release-docs.sh` (new)
- `scripts/make-release.sh` (modify to call doc script)
- `docs-site/` directory structure (new)
- Hugo configuration and templates (new)

## Success Metrics

1. **Automation**: Documentation updates automatically with each release
2. **Completeness**: All core and plugin commands documented
3. **Accessibility**: Professional web interface at GitHub Pages URL
4. **Performance**: <3 minutes additional time per release
5. **Reliability**: No release failures due to documentation generation

## Risk Analysis

### High Risk
- **Build Complexity**: Hugo dependency and additional build steps
- **GitHub Pages Limits**: 1GB repository size, public repository requirement

### Medium Risk
- **Plugin Dependencies**: Requires plugins to be properly queryable
- **Maintenance Overhead**: New script to maintain and debug

### Mitigation Strategies
- Fallback: Documentation generation failures don't block releases
- Testing: Local documentation generation for validation
- Size Management: Archive old documentation versions
- Plugin Compatibility: Graceful handling of non-responsive plugins

## Related Issues
- Related to release workflow improvements
- Connects to plugin system architecture (#007)
- Supports MCP integration documentation needs (#004)

---

**Implementation Notes**:
- This issue implements the architecture defined in ADR-018
- Developer should follow three-phase implementation approach
- Priority on Phase 1 (core functionality) first