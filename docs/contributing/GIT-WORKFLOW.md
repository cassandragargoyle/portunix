# Git Workflow Guidelines

## Purpose
This document defines the Git workflow and conventions for CassandraGargoyle projects, ensuring consistent version control practices across all team projects.

## Branch Strategy

### Main Branches
- **`main`** - Primary branch containing production-ready code (not `master`)
- **`develop`** - Integration branch for ongoing development (optional, for git-flow)

### Supporting Branches

#### Feature Branches
For developing new features:
- **Naming**: `feature/short-description`
- **Source**: Branch from `main` or `develop`
- **Merge target**: `main` or `develop`
- **Lifetime**: Until feature is complete

**Just-in-Time Branching Strategy:**
- Create feature branches only when you actively start working on the issue
- Don't create branches for issues you plan to work on "someday"
- Clean up unused branches regularly to maintain a clear workspace
- Rely on issue tracking systems (GitHub/Gitea issues) rather than branch names for long-term planning

```bash
# Create feature branch
git checkout main
git pull origin main
git checkout -b feature/github-integration

# Work on feature
git add .
git commit -m \"Add GitHub API integration\"

# Push feature branch
git push -u origin feature/github-integration
```

#### Bug Fix Branches
For fixing bugs:
- **Naming**: `fix/short-description`
- **Source**: Branch from `main`
- **Merge target**: `main` (and `develop` if exists)
- **Lifetime**: Until bug is fixed

```bash
# Create bug fix branch
git checkout main
git pull origin main
git checkout -b fix/docker-installation-timeout

# Fix the bug
git add .
git commit -m \"Fix Docker installation timeout issue\"

# Push fix branch
git push -u origin fix/docker-installation-timeout
```

#### Hotfix Branches
For critical production fixes:
- **Naming**: `hotfix/critical-description`
- **Source**: Branch from `main`
- **Merge target**: `main` and `develop` (if exists)
- **Lifetime**: Until hotfix is deployed

```bash
# Create hotfix branch
git checkout main
git pull origin main
git checkout -b hotfix/security-vulnerability

# Apply critical fix
git add .
git commit -m \"Fix critical security vulnerability in auth\"

# Push hotfix branch
git push -u origin hotfix/security-vulnerability
```

## Branch Naming Conventions

### Standard Prefixes
- `feature/` - New features and enhancements
- `fix/` - Bug fixes and corrections
- `hotfix/` - Critical production fixes
- `refactor/` - Code restructuring without functional changes
- `docs/` - Documentation updates
- `test/` - Test additions or improvements
- `chore/` - Maintenance tasks and tooling
- `release/` - Release preparation

### Naming Rules
1. **Use lowercase letters only**
2. **Use hyphens (-) to separate words**
3. **No spaces, underscores, or special characters**
4. **Maximum length: 35 characters**
5. **Be descriptive but concise**
6. **Use present tense verbs**

### Good Examples
```
✅ feature/github-integration
✅ feature/user-authentication  
✅ feature/package-manager-detection
✅ fix/docker-installation-timeout
✅ fix/memory-leak-in-installer
✅ fix/config-parsing-issue
✅ hotfix/security-vulnerability
✅ hotfix/data-corruption-fix
✅ refactor/config-loader-structure
✅ docs/api-documentation-update
✅ test/integration-test-suite
✅ chore/update-dependencies
✅ release/v1.2.0
```

### Bad Examples
```
❌ feature/NewFeature              # Use lowercase
❌ fix_bug                         # Use hyphens, not underscores
❌ Feature/GitHub Integration       # No spaces, use lowercase
❌ very-long-branch-name-with-details     # Too long (>35 chars)
❌ feature/added_authentication    # Use present tense (add, not added)
❌ bugfix                          # Use proper prefix (fix/)
❌ temp                            # Not descriptive
❌ feature/issue-123               # Don't use issue numbers as description
```

## Commit Message Guidelines

### Format
```
<type>: <description>

[optional body]

[optional footer]
```

### Commit Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, missing semicolons, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### Commit Message Rules
1. **Use imperative mood** (\"Add feature\" not \"Added feature\")
2. **Capitalize first letter** of the description
3. **Keep first line under 50 characters**
4. **No period at the end** of the first line
5. **Separate subject from body** with blank line
6. **Wrap body at 72 characters**

### Examples
```bash
# Good commit messages
git commit -m \"feat: Add Docker container support\"
git commit -m \"fix: Resolve package installation timeout\"
git commit -m \"docs: Update installation instructions\"
git commit -m \"refactor: Simplify configuration loading logic\"

# With body
git commit -m \"feat: Add GitHub API integration

Implements OAuth authentication and repository management
features. Includes error handling for rate limiting and
network connectivity issues.

Closes #123\"
```

## Workflow Examples

### Feature Development Workflow
```bash
# 1. Start from updated main branch
git checkout main
git pull origin main

# 2. Create feature branch
git checkout -b feature/user-dashboard

# 3. Work on feature with regular commits
git add .
git commit -m \"feat: Add user dashboard layout\"
git add .
git commit -m \"feat: Implement dashboard data fetching\"
git add .
git commit -m \"test: Add dashboard component tests\"

# 4. Keep branch updated (optional, for long-running features)
git checkout main
git pull origin main
git checkout feature/user-dashboard
git merge main  # or git rebase main

# 5. Push branch
git push -u origin feature/user-dashboard

# 6. Create Pull Request via GitHub/Gitea interface

# 7. After PR approval and merge, clean up
git checkout main
git pull origin main
git branch -d feature/user-dashboard
git push origin --delete feature/user-dashboard
```

### Bug Fix Workflow
```bash
# 1. Start from main branch
git checkout main
git pull origin main

# 2. Create fix branch
git checkout -b fix/login-validation-error

# 3. Fix the bug
git add .
git commit -m \"fix: Validate email format in login form\"

# 4. Push and create PR
git push -u origin fix/login-validation-error

# 5. After merge, clean up
git checkout main
git pull origin main
git branch -d fix/login-validation-error
```

### Hotfix Workflow
```bash
# 1. Create hotfix from main
git checkout main
git pull origin main
git checkout -b hotfix/security-patch

# 2. Apply critical fix
git add .
git commit -m \"hotfix: Fix SQL injection vulnerability\"

# 3. Push immediately
git push -u origin hotfix/security-patch

# 4. Create PR for immediate review and merge

# 5. Tag release after merge
git checkout main
git pull origin main
git tag -a v1.2.1 -m \"Hotfix release v1.2.1\"
git push origin v1.2.1
```

## Pull Request Guidelines

### PR Title Format
Use the same format as commit messages:
```
feat: Add user authentication system
fix: Resolve Docker installation issue
docs: Update API documentation
```

### PR Description Template
```markdown
## Summary
Brief description of changes.

## Changes
- List of specific changes made
- Another change
- Third change

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No merge conflicts
```

### PR Best Practices
1. **Keep PRs small and focused** (< 400 lines of changes)
2. **Use descriptive titles and descriptions**
3. **Link related issues** using keywords (\"Closes #123\")
4. **Request appropriate reviewers**
5. **Resolve all review comments** before merge
6. **Squash commits** if they're not meaningful individually

## Branch Management Best Practices

### Just-in-Time Branching Workflow

**✅ Recommended Approach:**
```bash
# 1. Check issue tracker for work to do
# 2. Only when you're ready to start coding, create the branch
git checkout main
git pull origin main
git checkout -b feature/implement-user-auth

# 3. Work on the feature
git add .
git commit -m "feat: Add user authentication system"
git push -u origin feature/implement-user-auth

# 4. When done, merge and immediately delete
git checkout main
git pull origin main
git merge feature/implement-user-auth
git branch -d feature/implement-user-auth
git push origin --delete feature/implement-user-auth
```

**❌ Avoid This Approach:**
```bash
# Don't create branches for all planned work
git checkout -b feature/future-feature-1
git checkout -b feature/future-feature-2  
git checkout -b feature/maybe-someday-feature
# ... creates branch clutter without active development
```

### Branch Cleanup

**Regular cleanup of unused branches:**
```bash
# List all branches
git branch -a

# Delete local branch
git branch -d branch-name

# Delete remote branch
git push origin --delete branch-name

# Cleanup remote tracking branches
git remote prune origin
```

**Automated cleanup script:**
```bash
#!/bin/bash
# cleanup-branches.sh
echo "Deleting merged local branches..."
git branch --merged main | grep -v main | xargs -n 1 git branch -d
echo "Pruning remote tracking branches..."
git remote prune origin
```

## Release Management

### Release Branches
For preparing releases:
```bash
# Create release branch
git checkout main
git pull origin main
git checkout -b release/v1.2.0

# Prepare release (update version, changelog, etc.)
git add .
git commit -m \"chore: Prepare release v1.2.0\"

# Push release branch
git push -u origin release/v1.2.0

# After testing and approval, merge to main
# Then tag the release
git checkout main
git pull origin main
git tag -a v1.2.0 -m \"Release version 1.2.0\"
git push origin v1.2.0
```

### Semantic Versioning
Use semantic versioning (MAJOR.MINOR.PATCH):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

```bash
# Examples
git tag -a v1.0.0 -m \"Initial release\"
git tag -a v1.1.0 -m \"Add new features\"
git tag -a v1.1.1 -m \"Bug fix release\"
git tag -a v2.0.0 -m \"Breaking changes\"
```

## Git Configuration

### Required Git Settings
```bash
# Set user information
git config --global user.name \"Your Name\"
git config --global user.email \"your.email@example.com\"

# Set default branch name
git config --global init.defaultBranch main

# Enable auto-correction
git config --global help.autocorrect 1

# Set default merge behavior
git config --global pull.rebase false
```

### Recommended Git Aliases
```bash
# Useful aliases
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.cm commit
git config --global alias.lg \"log --oneline --graph --decorate\"
git config --global alias.amend \"commit --amend --no-edit\"
```

## Best Practices

### Before Starting Work
1. Always pull the latest changes from main
2. Create a new branch for each feature/fix
3. Use descriptive branch names
4. Check that your branch name follows conventions

### While Working
1. Make small, focused commits
2. Write clear commit messages
3. Commit frequently (don't let changes pile up)
4. Test your changes before committing

### Before Creating PR
1. Rebase or merge latest main into your branch
2. Run tests and ensure they pass
3. Review your own changes
4. Update documentation if needed

### After PR Merge
1. Delete the feature branch locally and remotely
2. Pull latest main branch
3. Clean up any stale branches

## Common Git Commands

### Branch Management
```bash
# List all branches
git branch -a

# Create and switch to new branch
git checkout -b feature/new-feature

# Switch branches
git checkout main

# Delete local branch
git branch -d feature/old-feature

# Delete remote branch
git push origin --delete feature/old-feature

# Rename current branch
git branch -m new-branch-name
```

### Synchronization
```bash
# Update local main with remote changes
git checkout main
git pull origin main

# Update feature branch with latest main
git checkout feature/my-feature
git merge main
# OR
git rebase main

# Push local branch to remote
git push -u origin feature/my-feature

# Force push after rebase (use carefully)
git push --force-with-lease origin feature/my-feature
```

### Cleanup
```bash
# Remove branches that no longer exist on remote
git remote prune origin

# List merged branches
git branch --merged main

# Delete all merged branches except main
git branch --merged main | grep -v main | xargs -n 1 git branch -d
```

## Troubleshooting

### Common Issues

**Merge Conflicts**:
```bash
# When merge conflicts occur
git status  # See conflicted files
# Edit files to resolve conflicts
git add .   # Stage resolved files
git commit  # Complete the merge
```

**Accidentally Committed to Main**:
```bash
# Create branch from current state
git branch feature/accidental-work

# Reset main to origin
git reset --hard origin/main

# Switch to new branch
git checkout feature/accidental-work
```

**Wrong Commit Message**:
```bash
# Amend last commit message
git commit --amend -m \"New commit message\"

# For pushed commits, you'll need to force push
git push --force-with-lease origin branch-name
```

## Security Considerations

1. **Never commit secrets** (API keys, passwords, certificates)
2. **Use .gitignore** to exclude sensitive files
3. **Review changes** before committing
4. **Use signed commits** when required by project policy
5. **Be careful with force push** - use `--force-with-lease`

---

**Note**: These guidelines should be consistently applied across all CassandraGargoyle projects. Team members should review and follow these practices to maintain clean and manageable Git history.

*Created: 2025-08-23*
*Last updated: 2025-08-23*