# ADR-018: Post-Release Documentation Automation and Static Site Generation

**Status**: Proposed
**Date**: 2025-09-26
**Author**: Architect

## Context

### Current Release Process
The Portunix project has an established release workflow:
1. `scripts/make-release.sh` creates cross-platform binaries using GoReleaser
2. Generates release notes and packages for GitHub
3. Manual upload to GitHub releases

### Documentation Challenges
Currently, comprehensive command documentation is scattered across:
- Individual command help texts (`portunix --help`)
- Feature documentation in `docs/`
- Plugin-specific documentation in separate repository
- Release notes with limited usage examples

### Business Need
We need automated generation of complete, searchable, web-hosted documentation that:
- Documents ALL commands in the Portunix ecosystem (core + plugins)
- Updates automatically with each release
- Provides searchable, cross-referenced command documentation
- Hosts as static site on GitHub Pages for easy access

## Decision

### Architecture Overview
```
Release Workflow Extension:
make-release.sh → post-release-docs.sh → GitHub Pages Deployment

Components:
1. Documentation Generator (post-release-docs.sh)
2. Command Discovery System
3. Static Site Builder (Hugo/Jekyll)
4. GitHub Pages Integration
```

### Implementation Strategy

#### 1. Documentation Generation Script (`scripts/post-release-docs.sh`)
```bash
# Called from make-release.sh after successful release
post-release-docs.sh:
  ├── discover-commands()      # Core + Plugin command discovery
  ├── generate-command-docs()  # Auto-generate from --help outputs
  ├── build-static-site()      # Hugo/Jekyll build
  └── deploy-github-pages()    # Git push to gh-pages branch
```

#### 2. Command Discovery System
**Two-phase discovery:**
- **Core Commands**: Parse `portunix --help` and subcommand trees
- **Plugin Commands**: Query installed plugins via gRPC API

**Output**: Structured JSON manifest of all available commands

#### 3. Static Site Architecture
**Generator**: Hugo (Go-based, fast, GitHub Pages compatible)
**Structure**:
```
docs-site/
├── hugo.toml              # Hugo configuration
├── content/
│   ├── commands/          # Auto-generated command docs
│   │   ├── core/          # portunix install, docker, etc.
│   │   └── plugins/       # portunix agile, etc.
│   ├── guides/            # Manual documentation
│   └── releases/          # Release notes archive
├── layouts/               # Hugo templates
└── static/                # CSS, images, etc.
```

#### 4. GitHub Integration
**Repository**: Use existing `cassandragargoyle/Portunix` repository
**Branch Strategy**: `gh-pages` branch for static site
**URL**: `https://cassandragargoyle.github.io/Portunix/`

### Workflow Integration
```
Current:  make-release.sh → [manual GitHub upload]
Proposed: make-release.sh → post-release-docs.sh → GitHub Pages
```

**Modified make-release.sh** calls documentation script after successful build:
```bash
# At end of make-release.sh
if [ "$AUTO_DOCS" != "false" ]; then
    print_step "Generating documentation site..."
    ./scripts/post-release-docs.sh "$VERSION"
fi
```

## Consequences

### Positive
1. **Automated Documentation**: Complete command reference updates with each release
2. **Discoverability**: Searchable web interface for all Portunix commands
3. **User Experience**: Professional documentation site improves adoption
4. **Maintenance**: Reduces manual documentation sync burden
5. **Plugin Integration**: Includes plugin commands automatically
6. **SEO**: Web-hosted docs improve search engine visibility

### Negative
1. **Build Complexity**: Adds dependency on Hugo and additional build steps
2. **GitHub Pages Limitation**: 1GB repository limit, public repository requirement
3. **Maintenance Overhead**: New script to maintain and debug
4. **Release Time**: Additional 2-3 minutes per release for doc generation
5. **Plugin Dependencies**: Requires plugins to be properly queryable

### Risk Mitigation
1. **Fallback Strategy**: Document generation failures don't block releases
2. **Size Management**: Archive old documentation versions
3. **Plugin Compatibility**: Graceful handling of non-responsive plugins
4. **Testing**: Local documentation generation for testing

## Technical Implementation

### Phase 1: Core Documentation Generator
```bash
scripts/post-release-docs.sh:
├── Command discovery (portunix --help parsing)
├── Basic Hugo site generation
└── Local testing capability
```

### Phase 2: Plugin Integration
```bash
├── Plugin command discovery via gRPC
├── Plugin-specific documentation inclusion
└── Cross-referencing between core and plugins
```

### Phase 3: GitHub Pages Automation
```bash
├── Automated gh-pages branch management
├── Site deployment pipeline
└── Archive management for old versions
```

### Dependencies
**Build Dependencies**:
- Hugo static site generator
- Git (for gh-pages deployment)
- jq (for JSON processing)

**Runtime Dependencies**:
- Functioning Portunix binary (for command discovery)
- Plugin system (for plugin discovery)

## Alternatives Considered

### Alternative 1: Manual Documentation
**Rejected**: Doesn't scale with plugin ecosystem growth

### Alternative 2: GitHub Wiki
**Rejected**: Not easily automated, limited formatting options

### Alternative 3: External Documentation Platform (GitBook, Notion)
**Rejected**: Adds external dependency, not integrated with repository

### Alternative 4: MkDocs
**Considered**: Python-based, but Hugo preferred for consistency with Go ecosystem

## References
- Current release workflow: `scripts/make-release.sh`
- Hugo documentation: https://gohugo.io/
- GitHub Pages documentation
- Portunix plugin architecture (ADR-016, gRPC system)

---

**Next Steps**:
1. Developer implements `scripts/post-release-docs.sh` Phase 1
2. Test with local Hugo installation
3. Integrate with `make-release.sh`
4. Deploy initial version to GitHub Pages