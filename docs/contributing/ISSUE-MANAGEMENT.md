# Issue Management Guidelines

## Purpose
This document defines the process for creating and managing issues in CassandraGargoyle projects that have GitHub repositories. It covers two primary workflows for issue creation and ensures proper synchronization between internal team discussions and public GitHub tracking.

## Issue Creation Models

### Model A: GitHub-First Issue Creation

**When to use**: For bugs, feature requests, or issues that can be discussed publicly from the start.

**Process**:
1. **Create GitHub Issue**
   - Go to the project's GitHub repository
   - Click "Issues" → "New issue"
   - Choose appropriate template (Bug Report, Feature Request, etc.)
   - Fill in all required fields:
     - Clear, descriptive title
     - Detailed description
     - Steps to reproduce (for bugs)
     - Expected vs actual behavior
     - Environment details
     - Labels (bug, enhancement, question, etc.)
     - Assign to milestone if applicable

2. **Reference in Team Communication**
   - Share GitHub issue link in team channels
   - Use issue number in commits: `Fix authentication bug (closes #42)`
   - Reference in pull requests: `Fixes #42`

3. **Track Progress**
   - Update issue status as work progresses
   - Add comments for significant updates
   - Close when resolved with summary

**Example GitHub Issue Creation**:
```
Title: PowerShell installation fails on Fedora 40

Description:
## Bug Report

**Environment:**
- OS: Fedora 40
- Portunix version: 1.2.3
- Installation method: portunix install powershell --variant fedora

**Steps to reproduce:**
1. Run `portunix install powershell --variant fedora`
2. Wait for installation to complete

**Expected behavior:**
PowerShell should install successfully and be available via `pwsh` command

**Actual behavior:**
Installation fails with error: "Package powershell-fedora not found"

**Additional context:**
This worked on Fedora 39 but fails on Fedora 40. May be related to repository changes.

Labels: bug, powershell, fedora
Milestone: v1.3.0
```

### Model B: Team-First Issue Creation

**When to use**: For sensitive issues, internal planning, or when initial discussion is needed before public visibility.

**Process**:
1. **Internal Team Discussion**
   - Discuss the issue in team channels (Slack, Discord, etc.)
   - Gather initial requirements and scope
   - Determine if issue should be public or remain internal
   - Assign team member initials for tracking

2. **Create GitHub Issue**
   - Once ready for implementation, create public GitHub issue
   - Reference internal discussion: "Based on team discussion from [date]"
   - Include refined requirements and scope
   - Add appropriate labels and assignments

3. **Synchronize Tracking**
   - Link GitHub issue number to internal tracking
   - Use consistent numbering: "Issue #42 from GitHub corresponds to internal SD-042"
   - Update both systems as work progresses

**Example Team-First Workflow**:
```
Internal Discussion:
Team Member 1: "We need better error handling for container failures"
Team Member 2: "Yes, and we should add retry logic"
Team Lead: "Let's create GitHub issue #43 for this enhancement"

GitHub Issue Creation:
Title: Improve container failure handling with retry logic

Description:
## Enhancement Request

Based on team analysis, we need to improve how Portunix handles container failures.

**Current behavior:**
- Container failures cause immediate termination
- No retry mechanism for transient failures
- Limited error information provided to user

**Proposed enhancement:**
- Add configurable retry logic (3 attempts by default)
- Implement exponential backoff between retries
- Provide detailed error messages with troubleshooting hints
- Log failure details for debugging

**Acceptance criteria:**
- [ ] Retry mechanism with configurable attempts
- [ ] Exponential backoff implementation
- [ ] Enhanced error messages
- [ ] Unit tests for retry logic
- [ ] Integration tests with container failures

Labels: enhancement, containers, error-handling
Milestone: v1.4.0
Assignee: developer-username
```

## Issue Numbering and References

### GitHub Issue Numbers
- Use GitHub's automatic numbering (#1, #2, #3, etc.)
- Reference in commits: `git commit -m "Fix container timeout (refs #42)"`
- Close with commits: `git commit -m "Add retry logic (closes #42)"`

### Internal Tracking Integration
- Format: `SD-XXX` for team tracking, `#XXX` for GitHub
- Example: "Internal issue SD-042 corresponds to GitHub issue #42"
- Document mapping in project tracking systems

## Issue Templates

### Bug Report Template
```markdown
## Bug Report

**Environment:**
- OS: [e.g., Ubuntu 22.04, Windows 11]
- Project version: [e.g., v1.2.3]
- Installation method: [e.g., package manager, source]

**Steps to reproduce:**
1. Step one
2. Step two
3. Step three

**Expected behavior:**
Clear description of what should happen

**Actual behavior:**
Clear description of what actually happens

**Additional context:**
Any other relevant information, logs, screenshots
```

### Feature Request Template
```markdown
## Feature Request

**Problem description:**
Clear description of the problem this feature would solve

**Proposed solution:**
Detailed description of the proposed feature

**Alternative solutions:**
Any alternative approaches considered

**Additional context:**
Any other relevant information or examples
```

## Labels and Classification

### Standard Labels
- **Type**: `bug`, `enhancement`, `question`, `documentation`
- **Priority**: `low`, `medium`, `high`, `critical`
- **Component**: `installation`, `docker`, `ssh`, `testing`, `ui`
- **Status**: `needs-triage`, `in-progress`, `blocked`, `ready-for-review`
- **Platform**: `windows`, `linux`, `macos`, `cross-platform`

### Label Usage Guidelines
1. **Always assign type label** (bug, enhancement, etc.)
2. **Add component labels** for easier filtering
3. **Use priority labels** for important issues
4. **Add platform labels** when platform-specific

## Milestones and Planning

### Milestone Creation
- Create milestones for major releases: `v1.3.0`, `v1.4.0`
- Use milestones for sprint planning: `Sprint 2024-Q1`
- Include target dates and release notes

### Issue Assignment
- Assign issues to specific milestones during planning
- Move issues between milestones as priorities change
- Close milestone when all issues are resolved

## Communication Guidelines

### Issue Comments
- **Be constructive** and professional
- **Provide context** for decisions and changes
- **Tag relevant team members** with @mentions
- **Update status** when significant progress is made

### Cross-References
- Reference related issues: "Related to #42"
- Link pull requests: "PR #123 addresses this issue"
- Reference commits: "Fixed in commit abc1234"

## Integration with Development Workflow

### Commit Messages
```bash
# Reference issue
git commit -m "Add logging for container operations (refs #42)"

# Close issue
git commit -m "Fix authentication timeout (closes #42)"

# Multiple issues
git commit -m "Refactor error handling (refs #42, #43, closes #44)"
```

### Pull Request Integration
```markdown
## Pull Request

**Related Issues:** Fixes #42, refs #43

**Description:**
Brief description of changes made

**Testing:**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

**Checklist:**
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated if needed
```

## Best Practices

### Issue Creation
1. **Use descriptive titles** that summarize the issue
2. **Provide complete context** in the description
3. **Add appropriate labels** and assignments immediately
4. **Include reproduction steps** for bugs
5. **Define acceptance criteria** for features

### Issue Management
1. **Triage new issues** within 24-48 hours
2. **Update issue status** regularly
3. **Close resolved issues** promptly
4. **Archive or label** old/stale issues
5. **Review and update** issue templates periodically

### Team Coordination
1. **Discuss complex issues** in team channels first
2. **Document decisions** in issue comments
3. **Coordinate assignments** to avoid duplication
4. **Review progress** in team meetings
5. **Celebrate completions** and acknowledge contributors

## Tools and Automation

### GitHub Features
- **Issue templates** for consistent reporting
- **Labels and milestones** for organization
- **Projects** for kanban-style tracking
- **Actions** for automated workflows

### Integration Options
- **Slack/Discord bots** for notifications
- **Project management tools** (Jira, Trello, etc.)
- **CI/CD integration** for automatic issue updates
- **Time tracking** tools if needed

---

**Note**: These guidelines should be adapted based on specific project requirements and team preferences. Regular review and updates ensure the process remains effective and relevant.

*Created: 2025-08-23*
*Last updated: 2025-08-23*