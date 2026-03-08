# Test Plan - Issue #045: Node.js Installation Critical Fixes

## Overview
**Issue**: #045 - Node.js Installation Critical Fixes  
**Type**: Critical Bug Fix Testing  
**Created**: 2025-09-12  
**Testing Approach**: Container-based testing across all 9 officially supported Linux distributions  
**Based on**: ADR-009 - Officially Supported Linux Distributions

## Officially Supported Linux Distributions

Based on ADR-009, Portunix officially supports **9 Linux distributions** organized into **4 categories**:

### 1. APT-Based Distributions (Debian Family)

#### Ubuntu Distributions
1. **Ubuntu 22.04 LTS** (`ubuntu:22.04`)
   - **Support Level**: Full support with comprehensive testing
   - **Prerequisites**: `sudo wget curl lsb-release`
   - **Target Community**: General developers, startups, cloud deployments
   - **Priority**: **CRITICAL** - Primary supported platform

2. **Ubuntu 24.04 LTS** (`ubuntu:24.04`)
   - **Support Level**: Full support with comprehensive testing
   - **Prerequisites**: `sudo wget curl lsb-release`
   - **Target Community**: Modern developers, AI/ML practitioners, cloud-native development
   - **Priority**: **CRITICAL** - Latest LTS, primary development platform

#### Debian Distributions
3. **Debian 11 "Bullseye"** (`debian:11`)
   - **Support Level**: Full support with regular testing
   - **Prerequisites**: `sudo wget curl lsb-release`
   - **Target Community**: System administrators, enterprise environments
   - **Priority**: **HIGH** - Proven stability for production servers

4. **Debian 12 "Bookworm"** (`debian:12`)
   - **Support Level**: Full support with comprehensive testing
   - **Prerequisites**: `sudo wget curl lsb-release`
   - **Target Community**: Enterprise developers, security-conscious teams
   - **Priority**: **HIGH** - Current stable release with modern features

### 2. RPM-Based Distributions (Red Hat Family)

#### Fedora Distributions
5. **Fedora 39** (`fedora:39`)
   - **Support Level**: Standard support with regular testing
   - **Prerequisites**: `sudo curl`
   - **Target Community**: Red Hat ecosystem developers, upstream contributors
   - **Priority**: **MEDIUM** - Short lifecycle, good for Red Hat feature testing

6. **Fedora 40** (`fedora:40`)
   - **Support Level**: Standard support with regular testing
   - **Prerequisites**: `sudo curl`
   - **Target Community**: Modern Linux developers, container enthusiasts
   - **Priority**: **HIGH** - Current release with modern technologies

#### Enterprise Distributions
7. **Rocky Linux 9** (`rockylinux:9`)
   - **Support Level**: Standard support with regular testing
   - **Prerequisites**: `sudo curl`
   - **Target Community**: Enterprise developers, CentOS refugees
   - **Priority**: **HIGH** - Strong CentOS replacement with enterprise focus

### 3. Pacman-Based Distributions (Arch Family)

8. **Arch Linux** (`archlinux:latest`)
   - **Support Level**: Standard support with regular testing
   - **Prerequisites**: `sudo curl base-devel`
   - **Target Community**: Power users, DIY enthusiasts, minimalist developers
   - **Priority**: **MEDIUM** - Rolling release requires frequent testing

### 4. Universal Package Managers

9. **Snap-based Universal** (Base: `ubuntu:22.04`)
   - **Support Level**: Standard support for universal packages
   - **Prerequisites**: `sudo wget curl snapd`
   - **Target Community**: Cross-distribution users, CI/CD pipelines
   - **Priority**: **MEDIUM** - Valuable fallback solution

## Test Strategy

### Container-Based Testing Methodology
**MANDATORY**: All testing MUST be performed using Portunix native container commands as defined in `TESTING_METHODOLOGY.md`:

```bash
# Use Portunix container management (not direct Docker/Podman)
./portunix docker run-in-container nodejs --image [distribution]
./portunix podman run [distribution]
```

### Test Environment Setup
Each distribution will be tested in isolated containers with:
- Fresh container instance per test
- Required prerequisites pre-installed
- Clean environment verification
- Portunix binary mounted in container

## Critical Issues Test Cases

### Issue #1: Node.js Installation Download Failure

#### CRITICAL Priority Tests (Must Pass)

##### TC001: Ubuntu 22.04 Node.js Installation
**Given**: Clean Ubuntu 22.04 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image ubuntu:22.04`  
**Then**: 
- Node.js downloads successfully
- Extraction completes without errors
- Installation completes successfully
- `node --version` returns valid version
- `npm --version` returns valid version
**Priority**: CRITICAL - Primary platform

##### TC002: Ubuntu 24.04 Node.js Installation  
**Given**: Clean Ubuntu 24.04 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image ubuntu:24.04`  
**Then**: Same success criteria as TC001
**Priority**: CRITICAL - Latest LTS platform

##### TC003: Debian 12 Node.js Installation
**Given**: Clean Debian 12 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image debian:12`  
**Then**: Same success criteria as TC001
**Priority**: CRITICAL - Current stable release

#### HIGH Priority Tests (Should Pass)

##### TC004: Debian 11 Node.js Installation
**Given**: Clean Debian 11 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image debian:11`  
**Then**: Same success criteria as TC001
**Priority**: HIGH - Production server stability

##### TC005: Fedora 40 Node.js Installation
**Given**: Clean Fedora 40 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image fedora:40`  
**Then**: Same success criteria as TC001
**Priority**: HIGH - Current Fedora release

##### TC006: Rocky Linux 9 Node.js Installation
**Given**: Clean Rocky Linux 9 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image rockylinux:9`  
**Then**: Same success criteria as TC001
**Priority**: HIGH - Enterprise CentOS replacement

#### MEDIUM Priority Tests (Nice to Have)

##### TC007: Fedora 39 Node.js Installation
**Given**: Clean Fedora 39 container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image fedora:39`  
**Then**: Same success criteria as TC001
**Priority**: MEDIUM - Short lifecycle, good for feature testing

##### TC008: Arch Linux Node.js Installation
**Given**: Clean Arch Linux container with prerequisites  
**When**: Execute `./portunix docker run-in-container nodejs --image archlinux:latest`  
**Then**: Same success criteria as TC001
**Priority**: MEDIUM - Rolling release, requires maintenance

##### TC009: Snap-based Node.js Installation
**Given**: Ubuntu 22.04 container with snapd prerequisites  
**When**: Execute Node.js installation with snap variant  
**Then**: Same success criteria as TC001 but with snap verification
**Priority**: MEDIUM - Fallback solution

#### SPECIAL Testing (Requires Different Approach)

##### TC010: Distroless Node.js Runtime Verification
**Given**: Application built for distroless deployment  
**When**: Deploy to `gcr.io/distroless/base` container  
**Then**: 
- Application starts successfully
- Node.js runtime is available
- Application responds to requests
- No shell access (security verification)
**Priority**: SPECIAL - Different testing methodology required
**Note**: This tests runtime compatibility, not installation process

### Issue #2: Container Exec Command Parsing

#### Shell Command Execution Tests

##### TC011: Shell Command Execution - Ubuntu 22.04
**Given**: Container with Node.js successfully installed  
**When**: Execute `./portunix container exec test-container sh -c "node --version"`  
**Then**:
- Command executes without "unknown flag" errors
- Returns valid Node.js version
- Exit code is 0
**Priority**: CRITICAL

##### TC012: Shell Command Execution - Ubuntu 24.04
**Given**: Container with Node.js successfully installed  
**When**: Execute `./portunix container exec test-container bash -c "npm --version"`  
**Then**: Same success criteria as TC011
**Priority**: CRITICAL

##### TC013: Shell Command Execution - Debian 12
**Given**: Container with Node.js successfully installed  
**When**: Execute `./portunix container exec test-container sh -c "node --version"`  
**Then**: Same success criteria as TC011
**Priority**: HIGH

##### TC014: Shell Command Execution - Rocky Linux 9
**Given**: Container with Node.js successfully installed  
**When**: Execute `./portunix container exec test-container bash -c "npm --version"`  
**Then**: Same success criteria as TC011
**Priority**: HIGH

##### TC015: Complex Shell Command Parsing - All Platforms
**Given**: Container with Node.js installed  
**When**: Execute `./portunix container exec test-container sh -c "node -e 'console.log(\"test\")'"` 
**Then**:
- Command parses correctly across all supported distributions
- Executes JavaScript successfully
- Returns "test" output
**Priority**: HIGH - Tests complex flag parsing

## Coverage Matrix

| Distribution | Installation Test | Exec Test | Prerequisites | Priority | Notes |
|--------------|------------------|-----------|---------------|----------|-------|
| **Ubuntu 22.04** | TC001 | TC011 | apt, sudo, wget, curl, lsb-release | CRITICAL | Primary platform |
| **Ubuntu 24.04** | TC002 | TC012 | apt, sudo, wget, curl, lsb-release | CRITICAL | Latest LTS |
| **Debian 11** | TC004 | TC013 | apt, sudo, wget, curl, lsb-release | HIGH | Production stable |
| **Debian 12** | TC003 | TC013 | apt, sudo, wget, curl, lsb-release | HIGH | Current stable |
| **Fedora 39** | TC007 | TC014 | dnf, sudo, curl | MEDIUM | Short lifecycle |
| **Fedora 40** | TC005 | TC014 | dnf, sudo, curl | HIGH | Current release |
| **Rocky Linux 9** | TC006 | TC014 | dnf, sudo, curl | HIGH | Enterprise focus |
| **Arch Linux** | TC008 | TC015 | pacman, sudo, curl, base-devel | MEDIUM | Rolling release |
| **Snap Universal** | TC009 | TC011 | snap, snapd | MEDIUM | Fallback solution |
| **Google Distroless** | TC010 | N/A | External build | SPECIAL | Runtime only |

## Test Execution Commands

### Phase 1: Prerequisites Setup
```bash
# APT-based distributions (Ubuntu, Debian)
./portunix docker run ubuntu:22.04
# Inside: apt-get update && apt-get install -y sudo wget curl lsb-release

# RPM-based distributions (Fedora, Rocky)
./portunix docker run fedora:40
# Inside: dnf update -y && dnf install -y sudo curl

# Arch Linux
./portunix docker run archlinux:latest
# Inside: pacman -Syu && pacman -S sudo curl base-devel

# Snap-based
./portunix docker run ubuntu:22.04
# Inside: apt-get update && apt-get install -y sudo wget curl snapd
```

### Phase 2: Installation Testing

#### Critical Priority (Must Execute First)
```bash
# Primary platforms - MUST PASS
./portunix docker run-in-container nodejs --image ubuntu:22.04
./portunix docker run-in-container nodejs --image ubuntu:24.04
./portunix docker run-in-container nodejs --image debian:12
```

#### High Priority
```bash
# Important platforms - SHOULD PASS
./portunix docker run-in-container nodejs --image debian:11
./portunix docker run-in-container nodejs --image fedora:40
./portunix docker run-in-container nodejs --image rockylinux:9
```

#### Medium Priority
```bash
# Additional platforms - NICE TO PASS
./portunix docker run-in-container nodejs --image fedora:39
./portunix docker run-in-container nodejs --image archlinux:latest
# Snap variant testing with specific commands
```

#### Special Testing
```bash
# Distroless runtime verification (different approach)
# Build application externally, then deploy to distroless container
```

### Phase 3: Container Exec Testing
```bash
# Test exec command parsing for each successful installation
./portunix container exec [container-name] sh -c "node --version"
./portunix container exec [container-name] bash -c "npm --version"
./portunix container exec [container-name] sh -c "node -e 'console.log(\"Hello World\")'"
```

## Success Criteria by Priority

### CRITICAL (Blocking - Must Pass)
**Platforms**: Ubuntu 22.04, Ubuntu 24.04, Debian 12
- ✅ Node.js installation completes successfully
- ✅ `node --version` returns version (e.g., "v18.x.x" or similar)
- ✅ `npm --version` returns version (e.g., "9.x.x" or similar)
- ✅ Container exec commands work without flag parsing errors
- ✅ No "Download or extraction failed" errors
- ✅ No "unknown shorthand flag" errors

### HIGH (Important - Should Pass)
**Platforms**: Debian 11, Fedora 40, Rocky Linux 9
- ✅ Same criteria as CRITICAL priority
- ⚠️  May have platform-specific workarounds documented
- ✅ Installation completes in reasonable time (< 8 minutes)

### MEDIUM (Nice to Have - Can Have Issues)
**Platforms**: Fedora 39, Arch Linux, Snap Universal
- ✅ Installation works or has clear fallback procedure
- ⚠️  May require additional configuration or manual steps
- ⚠️  Platform-specific limitations documented
- ✅ Error messages are clear and actionable

### SPECIAL (Different Approach)
**Platform**: Google Distroless
- ✅ Runtime verification works for pre-built applications
- ✅ No shell access maintained (security feature)
- ✅ Application can be deployed and runs successfully
- ⚠️  Installation process is external to container

## Failure Handling Strategy

### Critical Failures (Issue must remain open)
- Node.js installation failure on Ubuntu 22.04, Ubuntu 24.04, or Debian 12
- Container exec command parsing failure on any CRITICAL platform
- Missing node/npm executables after successful installation on CRITICAL platforms

### High Priority Failures (Fix if possible, document if not)
- Installation failures on Debian 11, Fedora 40, Rocky Linux 9
- Performance issues (> 10 minutes installation time)
- Platform-specific dependency issues

### Medium Priority Failures (Document workarounds)
- Installation failures on Fedora 39, Arch Linux, Snap
- Provide alternative installation methods
- Document platform-specific limitations

### Special Case Handling
- Distroless: Focus on runtime compatibility rather than installation process
- Document build-time requirements and deployment strategies

## Performance Requirements

### Installation Performance by Priority
- **CRITICAL Platforms**: < 5 minutes installation time
- **HIGH Platforms**: < 8 minutes installation time  
- **MEDIUM Platforms**: < 15 minutes (with documented reasons for delays)
- **All Platforms**: Clear progress indicators, no hanging processes

### Container Exec Performance
- Command parsing: < 2 seconds
- Command execution: < 30 seconds for simple commands
- No significant delays in flag parsing logic

## Documentation Requirements

### Test Results Documentation
- **Pass/Fail Status**: For each distribution with detailed error logs
- **Performance Metrics**: Installation time and resource usage
- **Platform-Specific Issues**: Workarounds and limitations
- **Regression Testing**: Compare with baseline from Issue #041

### Platform-Specific Documentation
- **Ubuntu/Debian**: Standard APT-based installation procedures
- **Fedora/Rocky**: DNF-based installation and RPM considerations  
- **Arch Linux**: AUR packages and rolling release considerations
- **Snap Universal**: Cross-distribution compatibility notes
- **Distroless**: Build-time setup and runtime deployment guides

## CI/CD Integration

### Automated Testing Strategy
```bash
# Run tests in priority order - fail fast on CRITICAL
go test ./test/integration/issue_045_nodejs_installation_test.go -v -timeout 45m

# Run specific priority level
go test ./test/integration/issue_045_nodejs_installation_test.go -v -run "TC001|TC002|TC003" # CRITICAL
go test ./test/integration/issue_045_nodejs_installation_test.go -v -run "TC004|TC005|TC006" # HIGH  
go test ./test/integration/issue_045_nodejs_installation_test.go -v -run "TC007|TC008|TC009" # MEDIUM
```

### Test Report Generation
- **Priority-based Results**: Separate pass/fail rates by priority level
- **Distribution Matrix**: Visual matrix showing support status
- **Performance Dashboard**: Installation times across distributions
- **Regression Analysis**: Comparison with previous test runs

## Risk Mitigation

### High-Risk Scenarios
1. **Network Issues**: Test with different network configurations and proxies
2. **Resource Constraints**: Test with limited CPU/memory in containers
3. **Package Repository Issues**: Test fallback mechanisms and mirrors
4. **Version Conflicts**: Test with different Node.js versions and prerequisites

### Platform-Specific Risks
- **Ubuntu/Debian**: APT repository connectivity and GPG key issues
- **Fedora**: Short release lifecycle, repository changes
- **Rocky Linux**: RHEL compatibility and enterprise environment constraints
- **Arch Linux**: Rolling release instability, AUR dependencies
- **Distroless**: Build complexity and runtime limitations

## Success Metrics

### Overall Success Criteria
- **CRITICAL Platforms**: 100% pass rate (3/3 platforms)
- **HIGH Platforms**: 90%+ pass rate (2-3/3 platforms)  
- **MEDIUM Platforms**: 70%+ pass rate (2+/3 platforms)
- **Overall Coverage**: 80%+ of all 10 distributions working

### Quality Metrics
- **Zero Critical Bugs**: No blocking issues on Ubuntu 22.04, 24.04, Debian 12
- **Performance Targets**: Installation time targets met for each priority level
- **Error Handling**: Clear, actionable error messages for all failure modes
- **Documentation Quality**: Complete platform-specific documentation

---

**Execution Time**: Estimated 6-8 hours for complete test suite across all 10 distributions  
**Prerequisites**: Docker or Podman available, sufficient disk space for all distribution images (≈5GB)  
**Dependencies**: ADR-009 compliance, Issue #041 baseline results, container system functionality  
**Review Schedule**: Test plan review after each priority level completion