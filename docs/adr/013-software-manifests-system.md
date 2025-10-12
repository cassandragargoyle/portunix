# ADR 013: Software Manifests System

## Context
The current software package management system in Portunix relies on JSON-based definitions in `assets/install-packages.json`. While functional, this approach has several limitations:

1. **Limited Metadata**: Current package definitions lack comprehensive information about software (description, licensing, community metrics, etc.)
2. **Poor AI Integration**: JSON format is not optimal for AI assistants to understand software purpose, capabilities, and relationships
3. **Maintenance Complexity**: Single large JSON file becomes difficult to maintain as package list grows
4. **Missing Context**: No standardized way to document software usage patterns, performance characteristics, or integration details
5. **Inadequate Documentation**: Lack of structured information about software dependencies, alternatives, and best practices

Current alternatives considered:
1. Extend JSON with more fields - becomes unwieldy and still not AI-friendly
2. Use external package registries - lacks customization for Portunix use cases
3. Create comprehensive software manifest files - provides structured, detailed, AI-friendly documentation

The need for better AI integration is particularly important as Portunix aims to be AI-assistant friendly and provide rich context for automated decision-making.

## Decision
We will implement a **Software Manifests System** consisting of individual markdown files documenting each software package with comprehensive metadata and AI-optimized structure.

### Manifest Structure
Each manifest file (`docs/manifests/manifest-{software}.md`) will contain:

1. **Basic Information**
   - Name, version, category, description
   - Official website and documentation
   - Primary maintainer/organization

2. **Technical Details**
   - Supported platforms, installation methods
   - Dependencies and system requirements
   - Configuration options and defaults

3. **Installation & Usage**
   - Portunix installation commands
   - Common usage patterns and examples
   - Integration with other tools

4. **Metrics & Community**
   - GitHub statistics (stars, forks, contributors)
   - Usage statistics and adoption metrics
   - Community rating and popularity indicators

5. **Maintenance & Support**
   - Update frequency and release patterns
   - Support channels and documentation quality
   - Known issues and troubleshooting

6. **AI Integration Notes**
   - Structured tags for AI assistants
   - Common use cases and workflows
   - Integration patterns with development environments

### Implementation Plan
1. **Phase 1**: Create manifest directory structure and template
2. **Phase 2**: Develop first manifest files for core packages (Git, Node.js, Python, etc.)
3. **Phase 3**: Create validation tools and consistency checks
4. **Phase 4**: Integrate manifests with installation system for enhanced metadata
5. **Phase 5**: Develop AI-assistant tools for manifest querying and analysis

### Naming Convention
- Directory: `docs/manifests/`
- File pattern: `manifest-{software-name}.md`
- Software name should be lowercase, hyphenated (e.g., `manifest-visual-studio-code.md`)

## Consequences

### Positive
- **Enhanced AI Integration**: Structured, comprehensive information enables better AI assistant decision-making
- **Better Documentation**: Standardized format ensures consistent, complete software information
- **Improved Maintenance**: Individual files are easier to maintain than monolithic JSON
- **Rich Metadata**: Comprehensive information supports better package selection and troubleshooting
- **Community Insights**: Usage statistics and ratings help users make informed choices
- **Scalability**: System can grow organically with new packages and enhanced metadata

### Negative
- **Increased Complexity**: More files to maintain compared to single JSON file
- **Duplication Risk**: Some information may overlap with existing package definitions
- **Maintenance Overhead**: Requires keeping manifests up-to-date with software changes
- **Learning Curve**: Contributors need to understand manifest structure and requirements

### Implementation Details
- Manifests complement, not replace, existing `install-packages.json`
- JSON file remains authoritative for installation logic
- Manifests provide rich metadata and documentation layer
- Validation tools will ensure consistency between manifests and package definitions
- Template file will guide manifest creation
- CI/CD integration will validate manifest completeness and accuracy

### Migration Strategy
1. Start with most popular packages (Git, Node.js, Python, VS Code)
2. Gradually add manifests for all supported packages
3. Develop tools to auto-generate basic manifests from existing data
4. Implement validation to ensure manifest-package definition consistency
5. Create documentation and guidelines for contributors

### Success Metrics
- Number of packages with complete manifests
- AI assistant usage and effectiveness improvements
- Community contribution to manifest maintenance
- Reduced support requests due to better documentation
- Enhanced package discovery and selection user experience

### Future Enhancements
- Auto-generation of manifests from package managers and repositories
- Integration with package update notifications
- Community rating and review system
- Automated testing of manifest accuracy
- Export capabilities for external tools and AI systems

Created by: Zdeněk Kurc