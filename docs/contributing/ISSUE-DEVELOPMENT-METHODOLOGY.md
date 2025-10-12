# Issue Development Methodology

## Overview
This document defines the mandatory workflow for issue development in the Portunix project. This methodology ensures quality, traceability, and proper testing before code integration into the main branch.

## Development Workflow

### Phase 1: Issue Creation & Planning
**Role: Architect / Product Owner**

1. **Issue Creation**
   - Create file `docs/issues/{number}-{name}.md`
   - Update issue table in `docs/issues/README.md`
   - Create corresponding GitHub issue with same content
   - Define acceptance criteria and requirements
   - Assign appropriate labels and priority

2. **Technical Planning**
   - Define technical requirements
   - Identify affected components
   - Estimate implementation complexity
   - Create feature branch naming convention: `feature/issue-{number}-{short-name}`

### Phase 2: Implementation
**Role: Developer**

1. **Branch Setup**
   - Create feature branch from main: `git checkout -b feature/issue-{number}-{short-name}`
   - Ensure branch is up-to-date with main before starting

2. **Development Process**
   - Follow existing code conventions and patterns
   - Implement minimal diffs with maximum impact
   - Add appropriate tests for new functionality
   - Document changes in code comments (English only)
   - Commit frequently with descriptive messages

3. **Pre-Testing Validation**
   - Run local tests: `go test ./...`
   - Build successfully: `go build -o .`
   - Verify no breaking changes to existing functionality
   - Self-review code changes

4. **Merge Restriction**
   - **CRITICAL**: Developer MUST NOT merge to main branch at this stage
   - Feature branch remains separate until tester approval
   - Push feature branch to repository for tester access

### Phase 3: Testing & Validation
**Role: Tester**

1. **Test Environment Setup**
   - Checkout feature branch: `git checkout feature/issue-{number}-{short-name}`
   - Build and deploy in test environment
   - Prepare test data and scenarios
   - **OS Validation**: Verify testing environment matches tester role (Linux tester → Linux tests, Windows tester → Windows tests)
   - For container/VM testing: Accept tests based on container/VM OS, not host OS

2. **Testing Execution**
   - Execute functional tests based on acceptance criteria
   - Perform regression testing on related components
   - Test cross-platform compatibility (Windows/Linux)
   - Validate edge cases and error handling

3. **Acceptance Protocol Creation**
   - Create acceptance protocol document: `docs/testing/acceptance-{issue-number}.md`
   - Document all test scenarios and results
   - Include screenshots/logs where applicable
   - Record any issues or recommendations
   - **MANDATORY**: Document testing OS (host OS for local tests, container/VM OS for containerized tests)

4. **Acceptance Decision**
   - **PASS**: Issue ready for merge - create approval document
   - **FAIL**: Document blocking issues, return to developer
   - **CONDITIONAL**: List required changes before approval

### Phase 4: Integration
**Role: Developer (with Tester approval)**

1. **Pre-Merge Validation**
   - Verify acceptance protocol exists and shows PASS status
   - Ensure all tester recommendations are addressed
   - Rebase feature branch on current main if needed

2. **Merge Process**
   - Create merge request/pull request with acceptance protocol reference
   - Merge feature branch to main: `git merge feature/issue-{number}-{short-name}`
   - Update issue status to ✅ Implemented in `docs/issues/README.md`
   - Close GitHub issue with reference to acceptance protocol

3. **Post-Merge Cleanup**
   - Delete feature branch after successful merge
   - Update any related documentation
   - Notify team of completed feature

## Quality Gates

### Mandatory Requirements
- ✅ Issue must have defined acceptance criteria
- ✅ Code must pass all automated tests
- ✅ Feature branch must build successfully
- ✅ Tester must create acceptance protocol
- ✅ Acceptance protocol must show PASS status
- ✅ No merge to main without tester approval

### Acceptance Protocol Template
```markdown
# Acceptance Protocol - Issue #{number}

**Issue**: {issue-title}
**Branch**: feature/issue-{number}-{short-name}
**Tester**: {tester-name}
**Date**: {test-date}
**Testing OS**: {operating-system} (host/container/VM)

## Test Summary
- Total test scenarios: {count}
- Passed: {count}
- Failed: {count}
- Skipped: {count}

## Test Results
### Functional Tests
- [x] Feature works as specified
- [x] Acceptance criteria met
- [ ] Edge cases handled

### Regression Tests  
- [x] Existing functionality unaffected
- [x] Cross-platform compatibility verified

## Final Decision
**STATUS**: [PASS/FAIL/CONDITIONAL]

**Approval for merge**: [YES/NO]
**Date**: {approval-date}
**Tester signature**: {tester-name}
```

## Software Installation Testing Guidelines

### Container-Based Testing Policy

**MANDATORY**: All software installation testing MUST be performed in isolated containers, never on the host development system.

#### Rationale
- **Host Protection**: Prevents contamination of development environment with test software
- **Isolation**: Each test runs in clean, reproducible environment
- **Safety**: Eliminates risk of conflicts with existing host software
- **Consistency**: Standardized testing environment across all developers

#### Implementation Methods

##### Method 1: Portunix Docker Integration (Recommended)
```bash
# Create clean Ubuntu container for testing
portunix docker run ubuntu

# Inside container, install Portunix and test packages
./portunix install nodejs
./portunix install claude-code  # Should auto-install nodejs prerequisite
```

##### Method 2: Portunix Podman Integration
```bash
# Create clean container environment
portunix podman run ubuntu

# Test installation sequences
./portunix install java
./portunix install --dry-run python --variant full
```

##### Method 3: Manual Container Setup
```bash
# Only when Portunix container commands unavailable
docker run -it --rm ubuntu:latest bash
# Transfer binary and test in clean environment
```

#### Container Testing Workflow

1. **Environment Preparation**
   - Create fresh container instance
   - Copy Portunix binary to container
   - Verify clean starting state (no pre-installed software)

2. **Installation Testing**
   - Test primary package installation
   - Test prerequisite dependency resolution
   - Test variant selections and configurations
   - Test error handling and rollback scenarios

3. **Verification Testing**
   - Verify installed software works correctly
   - Test command availability and functionality
   - Validate PATH updates and environment setup
   - Check prerequisite satisfaction

4. **Cleanup Testing**
   - Test uninstallation if supported
   - Verify clean removal of dependencies
   - Test reinstallation after removal

#### Prohibited Practices

**❌ NEVER DO:**
- Install software packages directly on development host for testing
- Test package installation on primary development machine
- Use `sudo` commands during testing on host system
- Install dependencies manually on host to "help" testing

**✅ ALWAYS DO:**
- Use containers for all software installation tests
- Test in clean, isolated environments
- Document container setup procedures
- Include container commands in test scripts

#### Container Test Environment Standards

##### Base Images
- **Ubuntu**: `ubuntu:22.04` or `ubuntu:24.04` (primary testing)
- **Debian**: `debian:bookworm` (compatibility testing)
- **Fedora**: `fedora:latest` (RPM-based testing)
- **Alpine**: `alpine:latest` (minimal environment testing)

##### Container Configuration
```bash
# Standard container setup for testing
docker run -it --rm \
  --name portunix-test \
  --user root \
  ubuntu:22.04 bash

# Inside container
apt update && apt install -y curl wget
# Copy and test Portunix
```

#### Integration with Testing Phase

During **Phase 3: Testing & Validation**, testers MUST:

1. **Document Container Environment**
   ```markdown
   ## Test Environment
   - Container: ubuntu:22.04
   - Portunix Version: v1.5.12
   - Test Date: 2025-09-12
   - Container Setup: portunix docker run ubuntu
   ```

2. **Include Container Test Results**
   - Document container setup commands
   - Include installation output from clean environment
   - Verify no host system contamination
   - Test multiple container instances if needed

3. **Container-Specific Test Cases**
   - Fresh environment installation
   - Dependency resolution testing
   - Cross-platform compatibility
   - Clean environment verification

#### Developer Guidelines for AI Assistants

**For AI Development Assistants (Claude, etc.):**
- Always suggest container-based testing for software installation
- Never recommend installing test software on host system
- Provide container setup commands in testing recommendations
- Include dry-run testing before real installation attempts

**Example Workflow:**
```bash
# ✅ CORRECT: Container-based testing
portunix docker run ubuntu
./portunix install nodejs --dry-run    # Preview changes
./portunix install nodejs             # Test actual installation

# ❌ INCORRECT: Host system testing  
./portunix install nodejs             # Installs on development machine
```

## Process Enforcement

### Developer Responsibilities
- Follow all development phases in order
- Never bypass testing phase
- Never merge without tester approval
- Maintain feature branch until approval received

### Tester Responsibilities
- Test within reasonable timeframe
- Provide detailed acceptance protocol
- Clearly communicate pass/fail status
- Document all findings and recommendations

### Architect/Product Owner Responsibilities
- Provide clear acceptance criteria
- Review acceptance protocols when needed
- Make final decisions on conditional approvals

## Exception Handling

### Critical Hotfixes
- May bypass normal workflow with architect approval
- Must be tested immediately after deployment
- Retroactive acceptance protocol required within 24 hours

### Testing Delays
- If tester unavailable, architect may designate alternate tester
- Testing deadline: 48 hours for normal issues, 24 hours for high priority
- Escalation to architect if testing timeline exceeded

## Benefits of This Methodology

1. **Quality Assurance**: Every change is thoroughly tested before integration
2. **Traceability**: Clear documentation trail from issue to implementation
3. **Risk Reduction**: No untested code reaches main branch
4. **Team Coordination**: Clear roles and responsibilities
5. **Documentation**: Comprehensive testing records for future reference

---

**Last Updated**: 2025-09-11
**Version**: 1.0
**Status**: Active