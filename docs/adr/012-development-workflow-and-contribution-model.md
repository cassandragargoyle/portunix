# ADR-012: Development Workflow and Contribution Model

**Date**: 2025-09-18
**Status**: Accepted
**Author**: Zdeněk Kurc

## Context

As the Portunix project continues to grow, we need to establish a clear development workflow and contribution model. This decision is informed by studying established open-source projects, particularly Git's development model, and considering our project's specific needs and target audience.

Our analysis considered:
- The Git project's email-based mailing list workflow with GitGitGadget bridge
- Modern GitHub-centric workflows preferred by our target audience
- The project's current scale (individual/small team)
- Our cross-platform user base (Windows, Linux, macOS)
- The nature of Portunix as a development environment management tool

### Key Observations from Research

1. **Git Project Model**: Git uses a sophisticated hybrid model where:
   - Primary development occurs via email patches on mailing lists
   - GitHub serves as a read-only mirror
   - GitGitGadget bridges GitHub PRs to email patches
   - This model preserves independence from proprietary platforms

2. **Mailing List Benefits**:
   - Deep, threaded technical discussions
   - Platform independence
   - Superior archival and searchability
   - Offline workflow support
   - Higher quality, more thoughtful reviews

3. **GitHub Workflow Benefits**:
   - Lower barrier to entry for contributors
   - Visual code review interface
   - Integrated CI/CD pipelines
   - Familiar to modern developers
   - Better for our target audience (DevOps/tooling community)

## Decision

We will maintain a **GitHub-centric workflow** with structured quality gates, while remaining open to evolution as the project matures.

### Current Phase (1-10 contributors)
- **Primary Platform**: GitHub (issues, PRs, discussions)
- **Code Review**: GitHub pull requests with required reviews
- **Communication**: GitHub discussions for architectural decisions
- **Documentation**: Markdown files in repository
- **Quality Gates**: Automated CI/CD checks, PR templates, branch protection

### Future Evolution Path
We establish clear triggers for workflow evolution:

1. **Phase 1** (Current): GitHub-only workflow
2. **Phase 2** (10+ contributors): Consider structured RFC process
3. **Phase 3** (Enterprise adoption): Evaluate hybrid model needs
4. **Phase 4** (Critical infrastructure status): Consider mailing list adoption

## Consequences

### Positive
- **Low contribution barrier**: Developers can contribute using familiar tools
- **Modern tooling integration**: Full CI/CD, automated testing, visual reviews
- **Target audience alignment**: DevOps/tooling community expects GitHub
- **Rapid iteration**: Faster feedback loops for small team
- **Tool flexibility**: Developers can use their preferred IDEs and tools

### Negative
- **Platform dependency**: Reliance on Microsoft/GitHub infrastructure
- **Less structured discussion**: GitHub comments less suited for complex technical debates
- **Archival concerns**: Dependence on GitHub for long-term history
- **Quality risk**: Lower barrier may attract lower-quality contributions

### Mitigations
To address the negative consequences:

1. **Regular backups**: Implement repository mirroring to local infrastructure
2. **Structured templates**: Enforce PR and issue templates for quality
3. **Contributing guidelines**: Clear documentation of standards and processes
4. **Branch protection**: Automated quality checks before merge
5. **Discussion structure**: Use GitHub Discussions for architectural decisions

## Implementation Guidelines

### Required Infrastructure
```yaml
# .github/PULL_REQUEST_TEMPLATE.md
# .github/ISSUE_TEMPLATE/
# .github/workflows/
# CONTRIBUTING.md
# CODE_OF_CONDUCT.md
```

### Quality Standards
- All PRs require review
- CI/CD must pass (tests, linting, security)
- Documentation updates for features
- Semantic commit messages
- Issue tracking with clear acceptance criteria

### Evolution Triggers
Monitor these metrics quarterly:
- Active contributor count
- PR volume and complexity
- Architectural decision frequency
- Enterprise adoption indicators
- Community size and engagement

## Alternatives Considered

### Alternative 1: Immediate Mailing List Adoption
- **Rejected**: Would create unnecessary barrier for current scale
- **Rationale**: Project not mature enough to justify complexity

### Alternative 2: Hybrid Model (GitHub + Mailing List)
- **Deferred**: May adopt in future phases
- **Rationale**: Adds complexity without clear benefit at current scale

### Alternative 3: Alternative Platforms (GitLab, Gitea)
- **Rejected**: GitHub has largest developer community
- **Rationale**: Network effects more valuable than platform independence

## References

- Git project contribution model: https://git-scm.com/docs/SubmittingPatches
- GitGitGadget documentation: https://gitgitgadget.github.io/
- Linux kernel development process: https://www.kernel.org/doc/html/latest/process/
- GitHub best practices: https://docs.github.com/en/communities

## Review and Approval

This ADR acknowledges that development workflow is not prescriptive but adaptive. Like the Git project itself, we prioritize:
- Developer freedom in tool choice
- Quality over process dogma
- Evolution based on actual needs
- Community-driven decisions

The workflow will be reviewed quarterly and adjusted based on project growth and community feedback.