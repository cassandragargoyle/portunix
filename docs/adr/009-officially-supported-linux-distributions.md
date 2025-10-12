# ADR-009: Officially Supported Linux Distributions

**Status**: Accepted  
**Date**: 2025-09-12  
**Issue**: Standardization of Portunix Linux distribution support

## Context

Portunix has evolved to support multiple Linux distributions through its installation system, container management, and cross-platform testing. However, there has been no formal documentation defining which Linux distributions are officially supported, tested, and recommended for production use.

**Current Situation:**
- Multiple distributions are supported in the codebase
- Testing infrastructure exists for various distributions  
- No formal policy on distribution support lifecycle
- Inconsistent documentation about supported platforms
- Need for clear guidance for users and developers

**Business Impact:**
- Users unclear about which distributions are supported
- Developers lack guidance for testing requirements
- Support team needs clear boundaries for assistance
- Quality assurance needs defined testing matrix

## Decision

We formally define **9 officially supported Linux distributions** organized into five categories: **APT-based**, **RPM-based**, **Pacman-based**, **Universal package managers**, and **Micro distributions**. Additionally, we maintain **Google Distroless** variants for language-specific build testing purposes.

### 1. APT-Based Distributions (Debian Family)

#### Ubuntu Distributions
- **Ubuntu 22.04 LTS** (`ubuntu:22.04`)
  - **Status**: Primary supported platform
  - **Package Manager**: APT
  - **Support Level**: Full support with comprehensive testing
  - **Prerequisites**: `sudo wget curl lsb-release`
  - **End of Support**: April 2027
  - **Target Community**: General developers, startups, cloud deployments
  - **Focus**: User-friendly development environment, wide software compatibility
  - **Potential**: ⭐⭐⭐⭐⭐ **Excellent** - Most popular Linux desktop, extensive community support

- **Ubuntu 24.04 LTS** (`ubuntu:24.04`)  
  - **Status**: Latest LTS, primary development platform
  - **Package Manager**: APT
  - **Support Level**: Full support with comprehensive testing
  - **Prerequisites**: `sudo wget curl lsb-release`
  - **End of Support**: April 2029
  - **Target Community**: Modern developers, AI/ML practitioners, cloud-native development
  - **Focus**: Latest development tools, modern kernel features, performance optimization
  - **Potential**: ⭐⭐⭐⭐⭐ **Excellent** - Latest LTS with extended support, ideal for long-term projects

#### Debian Distributions
- **Debian 11 "Bullseye"** (`debian:11`)
  - **Status**: Stable release support
  - **Package Manager**: APT
  - **Support Level**: Full support with regular testing
  - **Prerequisites**: `sudo wget curl lsb-release`
  - **End of Support**: August 2026
  - **Target Community**: System administrators, enterprise environments, stability-focused developers
  - **Focus**: Rock-solid stability, security, minimal system resource usage
  - **Potential**: ⭐⭐⭐⭐ **Very Good** - Proven stability for production servers and conservative environments

- **Debian 12 "Bookworm"** (`debian:12`)
  - **Status**: Current stable release
  - **Package Manager**: APT  
  - **Support Level**: Full support with comprehensive testing
  - **Prerequisites**: `sudo wget curl lsb-release`
  - **End of Support**: June 2028
  - **Target Community**: Enterprise developers, security-conscious teams, containerized applications
  - **Focus**: Modern stability with updated packages, security hardening, container optimization
  - **Potential**: ⭐⭐⭐⭐⭐ **Excellent** - Best of both worlds - stability with modern features

### 2. RPM-Based Distributions (Red Hat Family)

#### Fedora Distributions
- **Fedora 39** (`fedora:39`)
  - **Status**: Community-supported rolling release
  - **Package Manager**: DNF
  - **Support Level**: Standard support with regular testing
  - **Prerequisites**: `sudo curl`
  - **End of Support**: November 2024
  - **Target Community**: Red Hat ecosystem developers, upstream contributors, bleeding-edge enthusiasts
  - **Focus**: Latest upstream packages, Red Hat technology preview, innovation testing
  - **Potential**: ⭐⭐⭐ **Good** - Short lifecycle but excellent for testing new Red Hat features

- **Fedora 40** (`fedora:40`)
  - **Status**: Latest community release
  - **Package Manager**: DNF
  - **Support Level**: Standard support with regular testing  
  - **Prerequisites**: `sudo curl`
  - **End of Support**: May 2025
  - **Target Community**: Modern Linux developers, container enthusiasts, RHEL-adjacent development
  - **Focus**: Cutting-edge technologies, container-first approach, systemd innovations
  - **Potential**: ⭐⭐⭐⭐ **Very Good** - Current release with modern technologies and active community

#### Enterprise Distributions  
- **Rocky Linux 9** (`rockylinux:9`)
  - **Status**: Enterprise-focused, RHEL-compatible
  - **Package Manager**: DNF
  - **Support Level**: Standard support with regular testing
  - **Prerequisites**: `sudo curl`
  - **End of Support**: May 2032
  - **Target Community**: Enterprise developers, CentOS refugees, RHEL-compatible environments
  - **Focus**: Enterprise stability, RHEL compatibility, long-term support, security compliance
  - **Potential**: ⭐⭐⭐⭐⭐ **Excellent** - Strong CentOS replacement with enterprise focus and long-term support

### 3. Pacman-Based Distributions (Arch Family)

- **Arch Linux** (`archlinux:latest`)
  - **Status**: Rolling release, bleeding-edge
  - **Package Manager**: Pacman
  - **Support Level**: Standard support with regular testing
  - **Prerequisites**: `sudo curl base-devel`
  - **End of Support**: Rolling (continuous updates)
  - **Special Notes**: Requires frequent testing due to rolling nature
  - **Target Community**: Power users, DIY enthusiasts, minimalist developers, AUR contributors
  - **Focus**: Bleeding-edge packages, minimalist base system, user control, latest software versions
  - **Potential**: ⭐⭐⭐⭐ **Very Good** - Excellent for developers wanting latest tools, but requires maintenance knowledge

### 4. Universal Package Managers

- **Snap-based Universal** (Base: `ubuntu:22.04`)
  - **Status**: Universal package system testing
  - **Package Manager**: Snap
  - **Support Level**: Standard support for universal packages
  - **Prerequisites**: `sudo wget curl snapd`
  - **Use Case**: Fallback for unsupported distributions
  - **Target Community**: Cross-distribution users, simplified deployment scenarios, CI/CD pipelines
  - **Focus**: Universal package compatibility, sandboxed applications, vendor-neutral deployment
  - **Potential**: ⭐⭐⭐ **Good** - Valuable fallback solution, but performance and size trade-offs limit primary use

### 5. Micro Distributions (Container Optimized)

- **Google Distroless** (`gcr.io/distroless/base`, `gcr.io/distroless/go`)
  - **Status**: Officially supported for build testing and production containers
  - **Package Manager**: None (runtime-only, no shell)
  - **Support Level**: Standard support for runtime-specific deployments and build testing
  - **Prerequisites**: External build environment required
  - **Use Case**: Production containers, security-hardened deployments, minimal attack surface, build testing for programming languages
  - **Target Community**: Security-focused developers, production deployments, microservices architects, build engineers
  - **Focus**: Minimal attack surface, runtime-only containers, language-specific build verification
  - **Potential**: ⭐⭐⭐⭐⭐ **Excellent** - Gold standard for production security, Google's proven approach
  - **Special Notes**: 
    - No package manager or shell - applications must be built externally
    - Runtime/Build variants: `gcr.io/distroless/java`, `gcr.io/distroless/python3`, `gcr.io/distroless/go`, `gcr.io/distroless/static`
    - Significant security benefits through minimal system components
    - Used specifically for testing builds in respective programming languages (Go variant for Go builds, Java variant for Java builds, etc.)
    - Requires specialized deployment strategies and tooling integration

#### Rationale for Distroless Support

**Security Benefits:**
- **Minimal Attack Surface**: Contains only runtime dependencies, no unnecessary system tools
- **No Shell Access**: Eliminates shell-based attacks and reduces container escape vectors
- **Reduced CVE Exposure**: Fewer packages mean fewer potential security vulnerabilities
- **Google's Security Expertise**: Leverages Google's container security best practices

**Production Advantages:**
- **Smaller Image Size**: Significantly smaller than traditional Linux distributions
- **Faster Startup**: Reduced image size leads to faster container initialization
- **Runtime Focus**: Optimized specifically for application runtime rather than development
- **Language-Specific Optimization**: Tailored images for specific programming languages

**Industry Adoption:**
- **Google Cloud Standards**: Standard practice for Google Cloud deployments
- **Kubernetes Ecosystem**: Widely adopted in Kubernetes production environments
- **Enterprise Security**: Meets stringent enterprise security requirements
- **Cloud-Native Best Practices**: Aligned with cloud-native security principles

**Integration Strategy:**
- **Build-time Support**: Portunix will support building applications for distroless deployment
- **Runtime Verification**: Tools for verifying application compatibility with distroless constraints
- **Security Scanning**: Enhanced security scanning for distroless-compatible applications
- **Documentation**: Comprehensive guides for migrating to distroless deployments

## Support Level Definitions

### Full Support
- **Comprehensive automated testing** in CI/CD pipeline
- **Priority bug fixes** for distribution-specific issues
- **Performance optimization** for platform-specific requirements
- **Documentation** includes distribution-specific examples
- **Community support** with guaranteed response times

### Standard Support  
- **Regular automated testing** in CI/CD pipeline
- **Standard bug fixes** according to priority
- **Basic functionality** guaranteed to work
- **General documentation** applies to platform
- **Community support** with best-effort response

### Experimental Support
- **Limited testing** - manual or occasional automated
- **Community-driven** bug fixes and improvements
- **Basic functionality** expected but not guaranteed
- **Minimal documentation** - users expected to adapt
- **Community support** only through forums/issues

## Distribution Selection Criteria

### Primary Criteria (Must Have)
1. **Active LTS/Stable Release**: Distribution must have active long-term support
2. **Container Image Availability**: Official Docker/Podman images available
3. **Package Manager Compatibility**: Compatible with APT, DNF, or universal packages
4. **Security Updates**: Regular security update cycle maintained

### Secondary Criteria (Nice to Have)
5. **Enterprise Usage**: Significant enterprise adoption
6. **Developer Popularity**: High developer mindshare
7. **CI/CD Integration**: Easy integration with automated testing
8. **Resource Efficiency**: Reasonable container resource requirements

### Exclusion Criteria (Disqualifiers)
- **Alpine Linux**: Excluded due to musl libc compatibility issues
- **CentOS**: Excluded due to end-of-life announcement
- **OpenSUSE**: Excluded due to limited testing resources

## Container Testing Matrix

All officially supported distributions MUST pass the following test scenarios:

### Core Installation Tests
- **Package Installation**: All standard packages install successfully
- **Prerequisite Handling**: Prerequisites install automatically
- **Error Handling**: Clear error messages for failures
- **Version Detection**: Accurate OS version detection

### Container Integration Tests  
- **Container Creation**: Successful container creation with Portunix
- **SSH Integration**: SSH server setup and connectivity
- **File Operations**: File copying and permission handling
- **Command Execution**: Container exec command functionality

### Software Package Tests
- **Node.js Installation**: Successful installation and verification
- **PowerShell Installation**: Platform-specific variant selection
- **Java Installation**: OpenJDK installation across platforms
- **Python Installation**: Version management and PATH setup

## Implementation Architecture

### Detection Logic (`app/install/installer.go`)
```go
type SupportedDistribution struct {
    Name               string   `json:"name"`
    Images             []string `json:"images"`
    PackageManager     string   `json:"package_manager"`
    SupportLevel       string   `json:"support_level"`
    Prerequisites      []string `json:"prerequisites"`  
    EndOfSupport       string   `json:"end_of_support"`
    VerificationCmd    string   `json:"verification_cmd"`
}

var OfficiallySupported = []SupportedDistribution{
    // APT-based distributions
    {
        Name: "Ubuntu 22.04 LTS",
        Images: []string{"ubuntu:22.04"},
        PackageManager: "apt",
        SupportLevel: "full",
        Prerequisites: []string{"sudo", "wget", "curl", "lsb-release"},
        EndOfSupport: "2027-04-30",
        VerificationCmd: "lsb_release -a",
    },
    {
        Name: "Arch Linux",
        Images: []string{"archlinux:latest"},
        PackageManager: "pacman",
        SupportLevel: "standard",
        Prerequisites: []string{"sudo", "curl", "base-devel"},
        EndOfSupport: "rolling",
        VerificationCmd: "pacman --version",
    },
    {
        Name: "Google Distroless",
        Images: []string{"gcr.io/distroless/base", "gcr.io/distroless/go", "gcr.io/distroless/java", "gcr.io/distroless/python3", "gcr.io/distroless/static"},
        PackageManager: "none",
        SupportLevel: "standard",
        Prerequisites: []string{}, // No prerequisites - runtime only
        EndOfSupport: "continuous",
        VerificationCmd: "", // No shell available for verification
    },
    // ... additional distributions
}
```

### Testing Infrastructure (`test/integration/`)
```go
func GetOfficiallySupported() []SupportedDistribution {
    // Returns complete list of officially supported distributions
    // Used by integration tests and CI/CD pipeline
}

func RunDistributionTests(dist SupportedDistribution) TestResult {
    // Standardized testing for each distribution
    // Ensures consistent test coverage across platforms
}
```

## Quality Assurance Requirements

### CI/CD Integration
- **Automated Testing**: All officially supported distributions tested in CI
- **Nightly Builds**: Extended testing with latest package versions
- **Release Testing**: Complete test suite before major releases
- **Performance Benchmarks**: Installation time and resource usage tracking

### Test Coverage Requirements
- **Core Functionality**: 100% test coverage for installation system
- **Distribution-Specific**: 90% test coverage for platform-specific features  
- **Container Integration**: 95% test coverage for container operations
- **Error Scenarios**: 80% test coverage for failure modes

### Documentation Requirements
- **Installation Guides**: Distribution-specific installation instructions
- **Troubleshooting**: Common issues and solutions per platform
- **Performance Notes**: Platform-specific performance characteristics
- **Migration Guides**: Upgrade paths between distribution versions

## Support Lifecycle Management

### New Distribution Evaluation
1. **Community Request**: User/developer requests support for new distribution
2. **Technical Assessment**: Evaluate against selection criteria
3. **Pilot Testing**: Limited testing in experimental support category
4. **Full Integration**: Move to standard or full support if successful
5. **Documentation**: Create distribution-specific documentation

### End of Life Management  
1. **Early Warning**: 12 months notice before dropping support
2. **Migration Guidance**: Provide clear migration path to supported distribution
3. **Deprecation Period**: 6 months deprecation warning in releases
4. **Final Removal**: Remove from officially supported list
5. **Community Handoff**: Transfer to community maintenance if requested

### Version Updates
- **LTS Releases**: Automatic inclusion when new LTS versions released  
- **Interim Releases**: Evaluation based on adoption and stability
- **Rolling Releases**: Version number updates quarterly
- **Security Updates**: Immediate testing when security updates affect packages

## Consequences

### ✅ **Positive**
- **Clear Expectations**: Users and developers know what's supported
- **Quality Assurance**: Standardized testing ensures reliability
- **Resource Focus**: Development effort concentrated on supported platforms
- **Professional Image**: Clear support matrix demonstrates maturity
- **Easier Troubleshooting**: Limited scope makes support more effective

### ⚠️ **Trade-offs**
- **Limited Scope**: Some users may want unsupported distributions  
- **Maintenance Overhead**: Requires ongoing testing and maintenance
- **Resource Requirements**: Testing infrastructure needs significant resources
- **Update Coordination**: Must coordinate with upstream distribution releases

### 🔄 **Migration Strategy**
1. **Phase 1**: Document current state and create formal support matrix
2. **Phase 2**: Implement standardized testing for all supported distributions  
3. **Phase 3**: Update documentation and user-facing materials
4. **Phase 4**: Begin end-of-life process for any unsupported distributions

## Success Metrics

### Quality Metrics
- **Test Coverage**: >95% success rate across all supported distributions
- **Installation Success**: >98% successful installations in clean containers
- **Issue Resolution**: <72 hours average resolution time for supported platforms
- **User Satisfaction**: >4.5/5 rating for installation experience

### Support Metrics  
- **Distribution Coverage**: 10 officially supported distributions maintained
- **Testing Frequency**: 100% of supported distributions tested per release
- **Documentation Quality**: <5% of issues due to documentation gaps
- **Community Engagement**: >80% of distribution-specific issues resolved by community

### Business Metrics
- **Adoption Rate**: Track adoption across different distributions
- **Support Burden**: Measure support ticket volume per distribution
- **Development Velocity**: Measure impact on development speed
- **User Retention**: Track user retention by primary distribution

## Future Evolution

### Planned Additions (Next 12 months)
- **Ubuntu 26.04 LTS**: When released, evaluate for inclusion
- **Debian 13**: When stable, evaluate for standard support
- **RHEL 9**: Evaluate if enterprise demand justifies resources
- **Alternative Architectures**: Consider ARM64 support for existing distributions

### Future Enterprise & Cloud-Native Distributions (Under Evaluation)

#### Cloud Provider Optimized
- **Amazon Linux 2023** (`amazonlinux:latest`)
  - **Status**: Under evaluation for cloud deployment scenarios
  - **Target Community**: AWS-focused developers, cloud-native applications
  - **Focus**: AWS service integration, optimized performance, security hardening
  - **Package Manager**: YUM/DNF
  - **Potential**: ⭐⭐⭐⭐ **Very Good** - Essential for AWS ecosystems, Amazon's strategic push
  - **Evaluation Criteria**: AWS adoption rates, enterprise demand, container optimization

#### Container & Cloud Specialized
- **VMware Photon OS** (`photon:latest`)
  - **Status**: Under evaluation for container-first environments
  - **Target Community**: VMware ecosystem, container orchestration, minimal footprint deployments
  - **Focus**: Minimal attack surface, container optimization, security-by-design
  - **Package Manager**: TDNF (Tiny DNF)
  - **Potential**: ⭐⭐⭐⭐ **Very Good** - Excellent for production containers, growing VMware adoption
  - **Evaluation Criteria**: Container performance benchmarks, enterprise VMware usage

#### Performance Optimized
- **Intel Clear Linux** (`clearlinux/base:latest`)
  - **Status**: Under evaluation for performance-critical applications
  - **Target Community**: Performance-focused developers, HPC, machine learning workloads
  - **Focus**: Intel CPU optimization, performance tuning, stateless design
  - **Package Manager**: SWuPd (Software Updater)
  - **Potential**: ⭐⭐⭐ **Good** - Niche but valuable for specific performance scenarios
  - **Evaluation Criteria**: Performance benchmarks, Intel hardware prevalence, community adoption

#### Evaluation Timeline
- **Q2 2025**: Initial compatibility testing and community feedback collection
- **Q3 2025**: Performance benchmarking and resource impact assessment
- **Q4 2025**: Decision on promotion to experimental or standard support
- **2026**: Full integration if evaluation criteria are met

#### Inclusion Criteria for Future Distributions
1. **Enterprise Demand**: Significant user requests and business justification
2. **Container Performance**: Superior performance in containerized environments
3. **Cloud Integration**: Strong cloud provider support and optimization
4. **Maintenance Feasibility**: Available testing resources and CI/CD integration
5. **Community Health**: Active development and security update lifecycle
6. **Differentiated Value**: Unique benefits not provided by existing distributions

#### Strategic Considerations
- **Cloud-Native Focus**: Prioritize distributions optimized for container/cloud deployment
- **Performance Benefits**: Evaluate distributions offering measurable performance improvements
- **Enterprise Adoption**: Monitor enterprise adoption trends and customer requests
- **Resource Allocation**: Balance new distribution support with existing platform quality
- **Vendor Relationships**: Consider strategic partnerships with distribution maintainers

### Technology Evolution
- **Container Runtime**: Support for additional container runtimes (containerd, cri-o)
- **Package Managers**: Support for additional package managers (flatpak, appimage)
- **Cloud Integration**: Cloud-specific distribution variants (Amazon Linux 2)
- **Embedded Systems**: Lightweight distributions for IoT scenarios

### Community Expansion
- **Community Testing**: Enable community to run tests for experimental distributions
- **Plugin System**: Allow community plugins for unsupported distributions  
- **Documentation Contributions**: Community-driven documentation for additional platforms
- **Feedback Integration**: Regular surveys to evaluate support priorities

---

**Decision Owner**: Software Architect  
**Reviewers**: Development Team, QA Team, DevOps Team  
**Implementation Target**: Immediate (documentation), Q1 2025 (testing infrastructure)  
**Review Schedule**: Quarterly review of supported distributions list