# Bug Reporting Guidelines

## Purpose
This document defines the standardized process for reporting, documenting, and tracking bugs across all CassandraGargoyle projects. It ensures consistent bug documentation and efficient resolution workflow.

## Bug Reporting Process

### 1. Identification and Initial Assessment

#### Before Reporting
1. **Reproduce the issue** - Ensure the bug is consistent and reproducible
2. **Check existing issues** - Search for similar reported bugs
3. **Verify expected behavior** - Confirm the behavior is actually incorrect
4. **Test on clean environment** - Verify the issue isn't environment-specific

#### Bug Severity Classification
- **Critical** - System crashes, data loss, security vulnerabilities
- **High** - Major feature failure, significant performance degradation
- **Medium** - Minor feature issues, usability problems
- **Low** - Cosmetic issues, enhancement suggestions

### 2. Bug Documentation Structure

#### Bug Report Naming Convention
```
BUG-[ISSUE-NUMBER]-[SEQUENCE]-[SHORT-DESCRIPTION].md
```

**Format Rules:**
- `ISSUE-NUMBER`: Three-digit issue number (e.g., 012)
- `SEQUENCE`: Three-digit sequence within issue (001, 002, etc.)
- `SHORT-DESCRIPTION`: Hyphenated lowercase description (max 30 characters)

**Examples:**
```
BUG-012-001-powershell-help-parsing.md
BUG-012-002-container-ssh-timeout.md
BUG-015-001-java-installation-failure.md
```

#### Directory Structure
```
project-root/
├── test/
│   ├── bug-reports/
│   │   ├── BUG-012-001-powershell-help-parsing.md
│   │   └── BUG-012-002-container-ssh-timeout.md
│   ├── reproductions/       # Scripts to reproduce bugs
│   └── regression/          # Regression test cases
```

### 3. Bug Report Template

```markdown
# 🐛 **Bug Report - Issue #[XXX] [Feature Name]**

---

## **Bug Summary**
Brief one-line description of the bug.

---

## **Bug Details**

| **Field** | **Value** |
|-----------|-----------|
| **Bug ID** | BUG-XXX-001 |
| **Reporter** | [Reporter name/role] |
| **Date** | [YYYY-MM-DD] |
| **Severity** | [Critical/High/Medium/Low] |
| **Priority** | [High/Medium/Low] |
| **Component** | [Component name] |
| **Affected Feature** | Issue #XXX - [Feature name] |
| **Branch** | [Branch name] |

---

## **Environment**
- **OS**: [Operating system]
- **Build**: [Build version/commit]
- **Test Phase**: [Unit/Integration/System/Acceptance]

---

## **Steps to Reproduce**
1. [First step]
2. [Second step]
3. [Third step]
4. [Observe result]

---

## **Expected Result**
[Detailed description of expected behavior, including code examples or output if relevant]
```
[Expected output/behavior examples]
```

---

## **Actual Result**
[Detailed description of actual behavior, including error messages and unexpected output]

---

## **Impact Assessment**
- **User Experience**: [Description of UX impact]
- **Documentation**: [Impact on documentation/help]
- **Functionality**: [Core functionality impact]
- **Testing**: [Impact on testing process]

---

## **Suggested Investigation Areas**
1. **[Area 1]**: [Investigation suggestion]
2. **[Area 2]**: [Investigation suggestion]
3. **[Area 3]**: [Investigation suggestion]
4. **[Specific File]**: [File-specific investigation]

---

## **Acceptance Criteria for Fix**
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]
- [ ] [Criterion 4]
- [ ] [Criterion 5]

---

## **Fix Implementation**
**Files Modified:**
- `[file1]`: [Description of changes]
- `[file2]`: [Description of changes]

**Solution:**
1. [Solution step 1]
2. [Solution step 2]
3. [Solution step 3]
4. [Solution step 4]

**Testing Results:**
- ✅ [Test result 1]
- ✅ [Test result 2]
- ✅ [Test result 3]
- ✅ [Test result 4]
- ✅ [Test result 5]
- ✅ [Test result 6]

---

## **Priority Justification**
**[Priority Level] Priority** because:
- ✅ ~~[Reason 1]~~ - **[STATUS]**
- ✅ ~~[Reason 2]~~ - **[STATUS]**
- ✅ ~~[Reason 3]~~ - **[STATUS]**
- ✅ ~~[Reason 4]~~ - **[STATUS]**

---

**Status**: ✅ **[FINAL STATUS]**
```

## Bug Report Creation Workflow

### Step 1: Create Bug Report File
```bash
# Navigate to project root
cd /path/to/project

# Create bug report directory if needed
mkdir -p test/bug-reports

# Create bug report file
touch test/bug-reports/BUG-XXX-001-description.md
```

### Step 2: Document the Bug
1. Fill out all sections of the template
2. Include exact error messages and logs
3. Provide clear reproduction steps
4. Classify severity and impact

### Step 3: Create Reproduction Script
```bash
# Create reproduction script
mkdir -p test/reproductions
touch test/reproductions/reproduce-BUG-XXX-001.sh
```

### Step 4: Link to Issue Tracking
1. Create GitHub issue if using GitHub-First model
2. Update project issue tracker
3. Cross-reference bug report with issue number

### Step 5: Assign and Track
1. Assign to appropriate developer
2. Set priority based on impact assessment
3. Update status as work progresses

## Bug Investigation Process

### Initial Triage
1. **Severity Assessment** - Determine impact and urgency
2. **Resource Allocation** - Assign appropriate developer
3. **Reproduction Verification** - Confirm bug can be reproduced
4. **Impact Analysis** - Assess broader system implications

### Investigation Steps
1. **Environment Setup** - Replicate reported environment
2. **Code Analysis** - Review relevant code sections
3. **Debugging** - Use appropriate debugging tools
4. **Root Cause Identification** - Determine underlying cause

### Documentation Updates
1. **Investigation Notes** - Document findings and analysis
2. **Test Cases** - Create regression tests
3. **Fix Documentation** - Document solution approach
4. **Verification Plan** - Define acceptance criteria

## Bug Resolution Workflow

### Fix Implementation
1. Create feature branch: `bugfix/BUG-XXX-001-description`
2. Implement fix with appropriate tests
3. Update bug report with resolution details
4. Create pull request with bug report reference

### Verification Process
1. **Unit Tests** - Ensure fix doesn't break existing functionality
2. **Integration Tests** - Verify fix in broader context
3. **Manual Testing** - Follow original reproduction steps
4. **Regression Testing** - Run full test suite

### Closure Criteria
- [ ] Fix implemented and tested
- [ ] Original reporter verified fix
- [ ] Regression tests added
- [ ] Documentation updated
- [ ] No new issues introduced

## Quality Assurance

### Review Checklist
- [ ] Bug report follows template structure
- [ ] All required fields completed
- [ ] Reproduction steps are clear and actionable
- [ ] Error messages and logs included
- [ ] Impact assessment is realistic
- [ ] Resolution is properly documented

### Metrics Tracking
- Average time to resolution by severity
- Bug discovery rate vs fix rate
- Most common bug categories
- Regression bug percentage

## Best Practices

### For Bug Reporters
1. **Be Specific** - Provide exact steps and error messages
2. **Include Context** - Environment, configuration, system state
3. **Test Thoroughly** - Verify reproduction and workarounds
4. **Update Regularly** - Keep bug report current with new findings

### For Developers
1. **Respond Quickly** - Acknowledge bug reports promptly
2. **Communicate Status** - Regular updates on investigation progress
3. **Document Thoroughly** - Record analysis and solution rationale
4. **Prevent Regression** - Add tests to prevent recurrence

### For Project Managers
1. **Prioritize Effectively** - Balance impact vs effort
2. **Track Trends** - Monitor bug patterns and root causes
3. **Resource Planning** - Allocate time for bug fixing
4. **Process Improvement** - Refine procedures based on experience

## Example: PowerShell Help Parsing Bug

### Real Example from Portunix Project
```
test/bug-reports/BUG-012-001-powershell-help-parsing.md
```

**Scenario**: CLI argument parsing incorrectly handles `--help` flag when used with PowerShell package.

**Key Elements Used:**
- **Structured Header**: `# 🐛 **Bug Report - Issue #012 PowerShell Linux Installation**`
- **Tabular Bug Details**: Organized metadata in table format
- **Clear Reproduction Steps**: `./portunix install powershell --help`
- **Expected vs Actual**: Detailed comparison with code examples
- **Impact Assessment**: UX, Documentation, Functionality, Testing impacts
- **Investigation Areas**: Specific files and components to check
- **Acceptance Criteria**: Checkboxes for fix verification
- **Implementation Tracking**: Files modified, solution steps, testing results
- **Priority Justification**: High priority reasons with status updates
- **Live Status Updates**: ✅ checkmarks and strikethrough for resolved items

**Resolution Results:**
- **Files Modified**: `cmd/install.go`, `app/install/install.go`
- **Solution**: Enhanced CLI parsing + package-specific help system
- **Status**: ✅ **RESOLVED - READY FOR QA TESTING**

**Format Benefits:**
1. **Visual Organization**: Clear section separation with horizontal rules
2. **Status Tracking**: Live updates with checkboxes and strikethrough
3. **Comprehensive Documentation**: All phases from reporting to resolution
4. **Professional Presentation**: Structured format suitable for team review

---

**Note**: This process should be adapted based on project complexity and team size. Regular review ensures the procedure remains effective and relevant.

*Created: 2025-08-23*
*Last updated: 2025-08-23*