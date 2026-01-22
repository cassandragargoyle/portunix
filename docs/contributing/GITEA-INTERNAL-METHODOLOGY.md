# Gitea Internal Development Methodology

## Overview

This document defines the methodology for internal development using our self-hosted Gitea instance at `http://gitea.cassandragargoyle.cz:3000`. This complements the general Git workflow and addresses specific practices for internal repository management.

## Gitea Infrastructure

### Primary Development Environment
- **Internal Repository**: `http://gitea.cassandragargoyle.cz:3000/CassandraGargoyle/portunix`
- **Purpose**: Internal development, experimentation, and pre-publication work
- **Access**: Team members only, protected environment for sensitive development

### Public Mirror
- **GitHub Repository**: `https://github.com/cassandragargoyle/Portunix`
- **Purpose**: Public releases, community engagement, and external collaboration
- **Sync**: Selective publishing via `github-02-sync-publish.sh` script

## Branch Naming Conventions

### Primary Branch: `main` (Not `master`)

**The primary branch is named `main`, never `master`. This convention is mandatory across all CassandraGargoyle projects.**

#### Rationale for `main` over `master`:

1. **Industry Best Practice** (2020+)
   - GitHub, GitLab, and other major platforms adopted `main` as default
   - Reflects modern software development standards
   - Aligns with inclusive language initiatives

2. **Technical Clarity**
   - More descriptive name indicating the primary development branch
   - Reduces confusion with Git's master/slave terminology
   - Clearer intent for new team members and contributors

3. **Forward Compatibility**
   - New repositories on GitHub/GitLab default to `main`
   - Tools and CI/CD systems increasingly expect `main` as default
   - Future-proofs our development process

4. **Professional Standards**
   - Many enterprise environments mandate inclusive language policies
   - Demonstrates commitment to modern development practices
   - Facilitates collaboration with external organizations

#### Migration Notes:
- All existing repositories have been migrated from `master` to `main`
- GitHub publishing scripts are configured for `main` branch
- CI/CD pipelines reference `main` branch
- Documentation consistently refers to `main` branch

### Branch Hierarchy

```
main (Primary Branch)
├── feature/issue-xxx-description
├── fix/issue-xxx-description  
├── release/vX.X.X
└── hotfix/critical-fix-name
```

## Internal Development Workflow

### 1. Issue-Driven Development
- All work begins with an internal issue in `docs/issues/internal/`
- Issue numbers are sequential (001, 002, etc.)
- Branch names reference issue numbers: `feature/issue-022-google-chrome-installation`

### 2. Feature Development Process

```bash
# 1. Start from main branch
git checkout main
git pull origin main

# 2. Create feature branch with issue reference
git checkout -b feature/issue-XXX-short-description

# 3. Develop and commit regularly
git add .
git commit -m "feat: implement feature component"

# 4. Push to Gitea for backup and collaboration
git push -u origin feature/issue-XXX-short-description

# 5. Merge to main when complete
git checkout main
git merge feature/issue-XXX-short-description

# 6. Clean up feature branch
git branch -d feature/issue-XXX-short-description
git push origin --delete feature/issue-XXX-short-description
```

### 3. Internal vs Public Content Management

#### Internal-Only Content (Stay in Gitea):
- `CLAUDE.md` - AI assistant instructions
- `GEMINI.md` - Alternative AI instructions  
- `NOTES.md` - Development notes
- `docs/private/` - Internal documentation
- `config/dev/` - Development configurations
- Build scripts (`.bat`, `.sh` files)
- Sensitive development utilities

#### Public Content (Gets Published):
- Core application code
- Public documentation
- README files
- License files
- Release configurations
- User-facing features

## Publishing to GitHub

### Enhanced Workflow Script
Use `./scripts/github-02-sync-publish.sh` for controlled publishing:

```bash
# Publishes current state with privacy filtering
./scripts/github-02-sync-publish.sh

# The script automatically:
# 1. Creates staging directory
# 2. Filters out private content
# 3. Creates feature branch on GitHub
# 4. Pushes for PR review
```

### Content Filtering
The publishing script automatically removes:
- Private documentation files
- Development configurations
- AI assistant instruction files
- Internal build scripts
- Sensitive development notes

## Security and Access

### Gitea Access Control
- **Admin**: Full repository access, settings management
- **Write**: Push access to feature branches, issue management
- **Read**: Clone, pull, view issues (for external consultants)

### Branch Protection
- `main` branch is protected on both Gitea and GitHub
- Requires clean history (no force pushes)
- Direct pushes to `main` discouraged, use feature branches

## Issue Management Integration

### Internal Issue Tracking
- Issues stored in `docs/issues/internal/`
- Sequential numbering system (001, 002, 003...)
- Branch names must reference issue numbers
- Status tracked in `docs/issues/README.md`

### Public Issue Mapping
- Selected internal issues can be published to GitHub
- Mapping maintained in `docs/issues/public/mapping.json`
- Public issues get PUB-XXX identifiers

## Backup and Disaster Recovery

### Repository Backup
- Gitea instance is regularly backed up
- GitHub serves as secondary backup for public content
- Critical development data is preserved

### Recovery Procedures
1. **Gitea Failure**: Continue development from GitHub clone
2. **Complete Loss**: Restore from latest GitHub public state
3. **Branch Recovery**: Use local developer clones as source

## Best Practices

### Commit Messages
Follow conventional commit format:
```
type(scope): description

feat(install): add Chrome browser support
fix(docker): resolve Windows installation issues
docs(readme): update installation instructions
```

### Branch Naming
- Use issue numbers: `feature/issue-022-description`
- Keep descriptions short but clear
- Use hyphens, not underscores
- Include issue type: `feature/`, `fix/`, `docs/`

### Collaboration
- Push feature branches to Gitea for collaboration
- Use draft PRs on GitHub for public feature previews
- Communicate significant architectural changes in issues first

## Troubleshooting

### Common Issues

**Branch out of sync:**
```bash
git checkout main
git fetch origin
git reset --hard origin/main
```

**Gitea connection issues:**
```bash
# Check connectivity
curl -I http://gitea.cassandragargoyle.cz:3000

# Verify remote URL
git remote -v
```

**Publishing script failures:**
```bash
# Check script permissions
chmod +x scripts/github-02-sync-publish.sh

# Manual cleanup if needed
rm -rf ../portunix-github-sync
```

## Team Responsibilities

### Developers
- Follow issue-driven development process
- Create feature branches for all work
- Write clear commit messages
- Test changes before merging to main

### Project Maintainer
- Review merge requests
- Manage Gitea repository settings
- Coordinate GitHub publishing schedule
- Maintain documentation currency

### DevOps
- Monitor Gitea instance health
- Manage backup procedures
- Update publishing scripts as needed
- Coordinate with external Git services

---

**Document Version**: 1.0  
**Last Updated**: September 1, 2025  
**Maintainer**: CassandraGargoyle Development Team  
**Related Documents**: 
- [Git Workflow Guidelines](GIT-WORKFLOW.md)
- [Issue Management](ISSUE-MANAGEMENT.md)
- [Tools Recommendations](TOOLS-RECOMMENDATIONS.md)