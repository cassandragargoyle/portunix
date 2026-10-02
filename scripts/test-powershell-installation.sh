#!/bin/bash

# PowerShell Installation Test Script
# Tests PowerShell installation across all supported Linux distributions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-"20m"}
VERBOSE=${VERBOSE:-"false"}
PARALLEL=${PARALLEL:-"false"}
DISTRIBUTIONS=${DISTRIBUTIONS:-"all"}

echo -e "${BLUE}==== Portunix PowerShell Installation Tests ====${NC}"
echo "Test timeout: $TEST_TIMEOUT"
echo "Verbose mode: $VERBOSE"
echo "Parallel execution: $PARALLEL"
echo "Target distributions: $DISTRIBUTIONS"
echo ""

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}Checking prerequisites...${NC}"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: Docker is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        echo -e "${RED}Error: Docker daemon is not running${NC}"
        exit 1
    fi
    
    # Check if portunix binary exists
    if [[ ! -f "./portunix" ]]; then
        echo -e "${YELLOW}Warning: portunix binary not found, attempting to build...${NC}"
        if ! go build -o portunix; then
            echo -e "${RED}Error: Failed to build portunix binary${NC}"
            exit 1
        fi
    fi
    
    echo -e "${GREEN}✓ Prerequisites check passed${NC}"
}

# Run integration tests
run_integration_tests() {
    echo -e "${BLUE}Running PowerShell integration tests...${NC}"
    
    local test_args="-tags integration"
    local test_pattern="./app/install"
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -v"
    fi
    
    if [[ "$PARALLEL" == "true" ]]; then
        test_args="$test_args -parallel 4"
    fi
    
    # Set test timeout
    test_args="$test_args -timeout $TEST_TIMEOUT"
    
    # Filter specific tests based on distributions
    case "$DISTRIBUTIONS" in
        "all")
            test_args="$test_args -run TestPowerShell.*"
            ;;
        "ubuntu")
            test_args="$test_args -run TestPowerShell.*Ubuntu"
            ;;
        "debian")
            test_args="$test_args -run TestPowerShell.*Debian"
            ;;
        "fedora")
            test_args="$test_args -run TestPowerShell.*Fedora"
            ;;
        "rocky")
            test_args="$test_args -run TestPowerShell.*Rocky"
            ;;
        "ssh")
            test_args="$test_args -run TestPowerShellSSH.*"
            ;;
        *)
            test_args="$test_args -run $DISTRIBUTIONS"
            ;;
    esac
    
    echo "Running: go test $test_args $test_pattern"
    echo ""
    
    if go test $test_args $test_pattern; then
        echo -e "${GREEN}✓ Integration tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Integration tests failed${NC}"
        return 1
    fi
}

# Run SSH-specific tests
run_ssh_tests() {
    echo -e "${BLUE}Running PowerShell SSH tests...${NC}"
    
    local test_args="-tags integration -v -timeout $TEST_TIMEOUT"
    local test_pattern="./app/install -run TestPowerShellSSH.*"
    
    echo "Running: go test $test_args $test_pattern"
    echo ""
    
    if go test $test_args $test_pattern; then
        echo -e "${GREEN}✓ SSH tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ SSH tests failed${NC}"
        return 1
    fi
}

# Run performance tests
run_performance_tests() {
    echo -e "${BLUE}Running PowerShell performance tests...${NC}"
    
    local test_args="-tags integration -v -timeout $TEST_TIMEOUT"
    local test_pattern="./app/install -run TestInstallationPerformance.*"
    
    echo "Running: go test $test_args $test_pattern"
    echo ""
    
    if go test $test_args $test_pattern; then
        echo -e "${GREEN}✓ Performance tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Performance tests failed${NC}"
        return 1
    fi
}

# Generate test report
generate_test_report() {
    echo -e "${BLUE}Generating test report...${NC}"
    
    local report_file="powershell_test_report.txt"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    cat > $report_file << EOF
PowerShell Installation Test Report
Generated: $timestamp
Test timeout: $TEST_TIMEOUT
Target distributions: $DISTRIBUTIONS

=== Test Results ===
EOF

    # Run tests with JSON output for report generation
    go test -tags integration -json -timeout $TEST_TIMEOUT ./app/install -run TestPowerShell.* >> $report_file 2>&1 || true
    
    echo -e "${GREEN}✓ Test report generated: $report_file${NC}"
}

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test containers...${NC}"
    
    # Stop and remove any test containers
    docker ps -q --filter "label=org.testcontainers" | xargs -r docker stop
    docker ps -aq --filter "label=org.testcontainers" | xargs -r docker rm
    
    # Clean up test images (optional)
    if [[ "${CLEANUP_IMAGES:-false}" == "true" ]]; then
        docker images -q --filter "dangling=true" | xargs -r docker rmi
    fi
    
    echo -e "${GREEN}✓ Cleanup completed${NC}"
}

# Signal handler for cleanup
trap cleanup EXIT INT TERM

# Main execution
main() {
    local test_type=${1:-"integration"}
    local exit_code=0
    
    check_prerequisites
    
    case "$test_type" in
        "integration")
            run_integration_tests || exit_code=1
            ;;
        "ssh")
            run_ssh_tests || exit_code=1
            ;;
        "performance")
            run_performance_tests || exit_code=1
            ;;
        "all")
            run_integration_tests || exit_code=1
            run_ssh_tests || exit_code=1
            run_performance_tests || exit_code=1
            ;;
        "report")
            generate_test_report || exit_code=1
            ;;
        *)
            echo "Usage: $0 [integration|ssh|performance|all|report]"
            echo ""
            echo "Options:"
            echo "  integration  - Run container-based integration tests (default)"
            echo "  ssh         - Run SSH-based integration tests"
            echo "  performance - Run performance benchmarks"
            echo "  all         - Run all test suites"
            echo "  report      - Generate detailed test report"
            echo ""
            echo "Environment variables:"
            echo "  TEST_TIMEOUT    - Test timeout (default: 20m)"
            echo "  VERBOSE         - Enable verbose output (default: false)"
            echo "  PARALLEL        - Enable parallel execution (default: false)"
            echo "  DISTRIBUTIONS   - Target distributions (default: all)"
            echo "  CLEANUP_IMAGES  - Clean up Docker images after tests (default: false)"
            echo ""
            echo "Examples:"
            echo "  $0 integration"
            echo "  DISTRIBUTIONS=ubuntu $0 integration"
            echo "  VERBOSE=true $0 ssh"
            echo "  TEST_TIMEOUT=30m $0 all"
            exit 1
            ;;
    esac
    
    if [[ $exit_code -eq 0 ]]; then
        echo -e "${GREEN}==== All tests completed successfully ====${NC}"
    else
        echo -e "${RED}==== Some tests failed ====${NC}"
    fi
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"