# Test Plan for Issue #012: PowerShell Linux Installation

## Plan
**Scope**: Test hybrid forward compatibility strategy for PowerShell installation across Linux distributions
**Target**: 8 distributions + version ranges + fallback mechanisms
**Timeline**: Multi-phase testing approach with real distribution VMs

## Test Cases 

### TC001: Primary Distribution Support (Given/When/Then)
**Given** Ubuntu 20.04, 22.04, 24.04 systems  
**When** `portunix install powershell` is executed  
**Then** PowerShell installs via Microsoft APT repository successfully  
**And** `pwsh --version` returns valid version string  

**Given** Debian 11, 12 systems  
**When** `portunix install powershell` is executed  
**Then** PowerShell installs via Microsoft APT repository successfully  

**Given** Fedora 38, 39, 40 systems  
**When** `portunix install powershell` is executed  
**Then** PowerShell installs via Microsoft DNF repository successfully  

**Given** Rocky Linux 8, 9 systems  
**When** `portunix install powershell` is executed  
**Then** PowerShell installs via Microsoft DNF repository successfully  

### TC002: Version Range Compatibility (Given/When/Then)
**Given** Ubuntu 25.04 system (experimental)  
**When** `portunix install powershell --variant ubuntu` is executed  
**Then** Version matcher classifies as "Compatible"  
**And** Installation attempts native method first  
**When** Native installation fails  
**Then** Fallback mechanism triggers with user prompt  

**Given** Ubuntu 30.00 system (future)  
**When** `portunix install powershell --variant ubuntu` is executed  
**Then** Version matcher classifies as "Experimental"  
**And** Warning message displays before installation  

### TC003: Fallback Chain Execution (Given/When/Then)
**Given** Ubuntu 25.04 with failed APT repository setup  
**When** Native installation fails  
**Then** System offers Snap fallback option  
**And** User confirms fallback installation  
**When** Snap installation proceeds  
**Then** PowerShell installs successfully via Snap  

**Given** Unsupported distribution (e.g., Arch Linux)  
**When** `portunix install powershell` is executed  
**Then** System displays manual installation guide  
**And** Provides official Microsoft documentation links  

### TC004: Variant Specification (Given/When/Then)
**Given** Any supported Linux distribution  
**When** `portunix install powershell --variant snap` is executed  
**Then** PowerShell installs via Snap package directly  
**And** Native repository methods are bypassed  

### TC005: Error Handling & User Experience (Given/When/Then)
**Given** Network connectivity issues  
**When** Repository download fails  
**Then** Clear error message displays with reason  
**And** Alternative installation methods are suggested  

**Given** Insufficient privileges  
**When** Installation requires sudo but not available  
**Then** Permission error explains requirement  
**And** Manual installation guide is provided  

## Coverage
### Unit Test Coverage Target: **85%**
- `version_matcher.go`: 90% (critical logic)
- `fallback.go`: 85% (user interaction flows)
- `installer.go`: 80% (integration points)

### Integration Test Coverage:
- **Distribution Matrix**: 8 distributions × 2-3 versions each = 20+ environments
- **Variant Testing**: All 7 variants (ubuntu, debian, fedora, rocky, mint, elementary, snap)
- **Fallback Scenarios**: 5 failure scenarios × 3 recovery paths = 15 paths

### E2E Test Coverage:
- **Happy Path**: 8 distributions successful installation
- **Compatibility Path**: 4 version-range scenarios  
- **Fallback Path**: 6 failure-to-recovery scenarios
- **Manual Path**: 2 unsupported distribution scenarios

## CI Notes
### Simulated Testing Strategy
```yaml
# CI Pipeline Structure
test:
  strategy:
    matrix:
      distribution: [ubuntu-20.04, ubuntu-22.04, ubuntu-24.04, ubuntu-25.04, debian-11, debian-12, fedora-38, fedora-39, fedora-40, rocky-8, rocky-9]
  
  container: 
    image: ${{ matrix.distribution }}
  
  steps:
    - name: Test PowerShell Installation
      run: ./portunix install powershell
    - name: Verify Installation  
      run: pwsh --version
    - name: Test Fallback (failure injection)
      run: ./test-fallback-scenarios.sh
```

## Automated Test Program

### Quick Start - One Command Testing
```bash
# Single command to run complete Issue #012 test suite
make test-issue-012

# Or directly:
./test/scripts/issue-012-test-runner.sh --full-suite
```

### Test Program Architecture
```
test/
├── scripts/
│   ├── issue-012-test-runner.sh       # Main test program (entry point)
│   ├── test-environment-setup.sh      # Environment preparation
│   ├── test-execution-engine.sh       # Test execution framework  
│   ├── test-report-generator.sh       # Results and reporting
│   └── test-cleanup.sh                # Environment cleanup
├── configs/
│   ├── distributions.json             # Test distribution definitions
│   ├── test-scenarios.json            # Test scenario configurations
│   └── failure-injection.json         # Failure simulation configs
├── fixtures/
│   └── expected-results/               # Expected test outcomes
└── results/
    ├── test-reports/                   # Generated test reports
    └── logs/                          # Detailed execution logs
```

### Main Test Program
```bash
#!/bin/bash
# test/scripts/issue-012-test-runner.sh
# Complete automated test program for Issue #012 PowerShell Installation

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$TEST_ROOT")"
RESULTS_DIR="$TEST_ROOT/results"
REPORT_FILE="$RESULTS_DIR/issue-012-test-report-$(date +%Y%m%d-%H%M%S).html"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

usage() {
    echo "Issue #012 PowerShell Installation Test Suite"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --full-suite         Run complete test suite (all distributions)"
    echo "  --quick              Run quick test (Ubuntu 22.04 only)"
    echo "  --distribution NAME  Run specific distribution test"
    echo "  --with-cleanup       Clean up test environments after completion"
    echo "  --report-only        Generate report from existing results"
    echo "  --help               Show this help message"
    echo ""
    echo "Available distributions:"
    echo "  ubuntu-22, ubuntu-24, debian-11, debian-12"
    echo "  fedora-39, fedora-40, rocky-9, mint-21"
    echo ""
    echo "Examples:"
    echo "  $0 --full-suite --with-cleanup"
    echo "  $0 --quick"
    echo "  $0 --distribution ubuntu-22"
}

log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')] $1${NC}"
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
    ((PASSED_TESTS++))
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    ((FAILED_TESTS++))
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

skip() {
    echo -e "${YELLOW}[SKIPPED] $1${NC}"
    ((SKIPPED_TESTS++))
}

# Test execution framework
run_test_suite() {
    local suite_type="$1"
    
    log "Starting Issue #012 PowerShell Installation Test Suite"
    log "Suite Type: $suite_type"
    
    # Create results directory
    mkdir -p "$RESULTS_DIR/logs"
    
    # Initialize test environment
    source "$SCRIPT_DIR/test-environment-setup.sh"
    setup_test_environment "$suite_type"
    
    # Load test configurations
    local distributions_config="$TEST_ROOT/configs/distributions.json"
    local scenarios_config="$TEST_ROOT/configs/test-scenarios.json"
    
    # Execute tests based on suite type
    case "$suite_type" in
        "full-suite")
            run_full_distribution_tests "$distributions_config"
            run_version_compatibility_tests
            run_fallback_mechanism_tests
            run_failure_injection_tests
            ;;
        "quick")
            run_quick_test
            ;;
        "distribution")
            run_single_distribution_test "$2"
            ;;
    esac
    
    # Generate test report
    source "$SCRIPT_DIR/test-report-generator.sh"
    generate_test_report "$REPORT_FILE"
    
    # Show summary
    show_test_summary
}

run_full_distribution_tests() {
    local config_file="$1"
    
    log "Running full distribution test suite"
    
    # Read distribution configurations
    local distributions=(
        "ubuntu-22:ubuntu:22.04"
        "ubuntu-24:ubuntu:24.04"  
        "debian-11:debian:bullseye"
        "debian-12:debian:bookworm"
        "fedora-39:fedora:39"
        "fedora-40:fedora:40"
        "rocky-9:rockylinux:9"
        "mint-21:linuxmintd/mint21-amd64"
    )
    
    for distro_info in "${distributions[@]}"; do
        IFS=':' read -r name variant image <<< "$distro_info"
        ((TOTAL_TESTS++))
        
        log "Testing distribution: $name"
        
        if run_distribution_test "$name" "$variant" "$image"; then
            success "PowerShell installation succeeded on $name"
        else
            error "PowerShell installation failed on $name"
        fi
    done
}

run_distribution_test() {
    local name="$1"
    local variant="$2" 
    local image="$3"
    
    local container_name="test-$name"
    local log_file="$RESULTS_DIR/logs/$name.log"
    
    {
        echo "=== Distribution Test: $name ==="
        echo "Container: $container_name"
        echo "Image: $image"
        echo "Variant: $variant"
        echo "Started: $(date)"
        echo ""
        
        # Create and start container
        if ! create_test_container "$container_name" "$image"; then
            echo "Failed to create container"
            return 1
        fi
        
        # Copy portunix binary
        if ! copy_portunix_to_container "$container_name"; then
            echo "Failed to copy portunix binary"
            return 1
        fi
        
        # Execute installation test
        if ! test_powershell_installation "$container_name" "$variant"; then
            echo "PowerShell installation failed"
            
            # Try fallback
            echo "Attempting snap fallback..."
            if test_powershell_fallback "$container_name"; then
                echo "Fallback installation succeeded"
                return 0
            else
                echo "Fallback installation also failed"
                return 1
            fi
        fi
        
        # Verify installation
        if ! verify_powershell_installation "$container_name"; then
            echo "PowerShell verification failed"
            return 1
        fi
        
        echo "Test completed successfully"
        return 0
        
    } > "$log_file" 2>&1
}

run_version_compatibility_tests() {
    log "Running version compatibility tests"
    ((TOTAL_TESTS++))
    
    # Test Ubuntu 25.04 (experimental)
    if test_experimental_version "ubuntu-25" "ubuntu:25.04" "ubuntu"; then
        success "Ubuntu 25.04 compatibility test passed"
    else
        error "Ubuntu 25.04 compatibility test failed"
    fi
}

run_fallback_mechanism_tests() {
    log "Running fallback mechanism tests"
    ((TOTAL_TESTS++))
    
    # Test fallback chain: ubuntu → snap → manual
    if test_fallback_chain; then
        success "Fallback mechanism test passed"
    else
        error "Fallback mechanism test failed"
    fi
}

run_failure_injection_tests() {
    log "Running failure injection tests"
    
    local injection_scenarios=(
        "network-failure"
        "permission-failure"
        "repository-corruption"
        "disk-space-full"
    )
    
    for scenario in "${injection_scenarios[@]}"; do
        ((TOTAL_TESTS++))
        log "Testing failure scenario: $scenario"
        
        if test_failure_scenario "$scenario"; then
            success "Failure injection test passed: $scenario"
        else
            error "Failure injection test failed: $scenario"
        fi
    done
}

run_quick_test() {
    log "Running quick test (Ubuntu 22.04 only)"
    ((TOTAL_TESTS++))
    
    if run_distribution_test "ubuntu-22" "ubuntu" "ubuntu:22.04"; then
        success "Quick test passed"
    else
        error "Quick test failed"
    fi
}

show_test_summary() {
    echo ""
    echo "================================================"
    echo "           TEST SUMMARY"
    echo "================================================"
    echo -e "Total Tests:  ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
    echo -e "Skipped:      ${YELLOW}$SKIPPED_TESTS${NC}"
    echo ""
    echo -e "Success Rate: ${GREEN}$(( PASSED_TESTS * 100 / TOTAL_TESTS ))%${NC}"
    echo ""
    echo "Detailed report: $REPORT_FILE"
    echo "================================================"
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Parse command line arguments
SUITE_TYPE=""
WITH_CLEANUP=false
SPECIFIC_DISTRO=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --full-suite)
            SUITE_TYPE="full-suite"
            shift
            ;;
        --quick)
            SUITE_TYPE="quick"
            shift
            ;;
        --distribution)
            SUITE_TYPE="distribution"
            SPECIFIC_DISTRO="$2"
            shift 2
            ;;
        --with-cleanup)
            WITH_CLEANUP=true
            shift
            ;;
        --report-only)
            generate_test_report "$REPORT_FILE"
            exit 0
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate arguments
if [[ -z "$SUITE_TYPE" ]]; then
    echo "Error: Test suite type required"
    usage
    exit 1
fi

# Run the test suite
if [[ "$SUITE_TYPE" == "distribution" ]]; then
    run_test_suite "$SUITE_TYPE" "$SPECIFIC_DISTRO"
else
    run_test_suite "$SUITE_TYPE"
fi

# Cleanup if requested
if [[ "$WITH_CLEANUP" == true ]]; then
    source "$SCRIPT_DIR/test-cleanup.sh"
    cleanup_test_environment
fi
```

### Makefile Integration
```makefile
# Add to existing Makefile
.PHONY: test-issue-012 test-issue-012-quick test-issue-012-clean

# Run complete Issue #012 test suite
test-issue-012:
	@echo "Running Issue #012 PowerShell Installation Test Suite"
	@./test/scripts/issue-012-test-runner.sh --full-suite --with-cleanup

# Run quick test (Ubuntu 22.04 only)
test-issue-012-quick:
	@echo "Running Issue #012 Quick Test"
	@./test/scripts/issue-012-test-runner.sh --quick

# Clean up test environments
test-issue-012-clean:
	@echo "Cleaning up Issue #012 test environments"
	@./test/scripts/test-cleanup.sh
```

### Usage Examples
```bash
# Complete test suite (recommended for CI/CD)
make test-issue-012

# Quick development testing
make test-issue-012-quick

# Test specific distribution
./test/scripts/issue-012-test-runner.sh --distribution ubuntu-22

# Generate HTML report
./test/scripts/issue-012-test-runner.sh --report-only

# Manual cleanup
make test-issue-012-clean
```

### Failure Injection Ideas:
- **Network Failures**: Block Microsoft repository URLs
- **Permission Failures**: Remove sudo access temporarily  
- **Package Conflicts**: Pre-install conflicting PowerShell versions
- **Repository Corruption**: Modify GPG keys to cause verification failures
- **Version Mismatches**: Mock OS detection to return unsupported versions
- **Disk Space**: Fill disk to trigger installation space errors

---
*Created: 2025-08-23*
*Issue: #012*
*ADR: ADR-003*