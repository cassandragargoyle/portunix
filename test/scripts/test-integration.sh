#!/bin/bash
# test/scripts/test-integration.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Running Integration Test Suite${NC}"

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-"30m"}
VERBOSE=${VERBOSE:-"false"}
PARALLEL=${PARALLEL:-"false"}
RUN_PATTERN=${RUN_PATTERN:-""}

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -v, --verbose       Enable verbose output"
    echo "  -p, --parallel      Run tests in parallel"
    echo "  -t, --timeout TIME  Set test timeout (default: 30m)"
    echo "  -r, --run PATTERN   Run specific test pattern"
    echo ""
    echo "Environment variables:"
    echo "  TEST_TIMEOUT    Test timeout (default: 30m)"
    echo "  VERBOSE         Enable verbose output (default: false)"
    echo "  PARALLEL        Enable parallel execution (default: false)"
    echo "  RUN_PATTERN     Test pattern to run (default: all)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run all integration tests"
    echo "  $0 -r TestIssue012                   # Run Issue #012 tests only"
    echo "  $0 -v -t 45m                         # Verbose mode with 45m timeout"
    echo "  PARALLEL=true $0                     # Run with parallel execution"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE="true"
            shift
            ;;
        -p|--parallel)
            PARALLEL="true"
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        -r|--run)
            RUN_PATTERN="$2"
            shift 2
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    
    # Check if we're in the correct directory
    if [[ ! -f "../../go.mod" ]]; then
        echo -e "${RED}Error: Must be run from project root or test/scripts directory${NC}"
        exit 1
    fi
    
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
    if [[ ! -f "../../portunix" ]]; then
        echo -e "${YELLOW}Warning: portunix binary not found, attempting to build...${NC}"
        cd ../.. && go build -o portunix
        if [[ $? -ne 0 ]]; then
            echo -e "${RED}Error: Failed to build portunix binary${NC}"
            exit 1
        fi
        cd test/scripts
    fi
    
    echo -e "${GREEN}✓ Prerequisites check passed${NC}"
}

# Run integration tests
run_integration_tests() {
    echo -e "${YELLOW}Running integration tests...${NC}"
    
    cd ../.. # Go to project root
    
    # Build test arguments
    local test_args="-tags integration"
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -v"
    fi
    
    if [[ "$PARALLEL" == "true" ]]; then
        test_args="$test_args -parallel 4"
    fi
    
    # Set test timeout
    test_args="$test_args -timeout $TEST_TIMEOUT"
    
    # Add run pattern if specified
    if [[ -n "$RUN_PATTERN" ]]; then
        test_args="$test_args -run $RUN_PATTERN"
    fi
    
    # Run tests
    local test_path="./test/integration"
    echo "Executing: go test $test_args $test_path"
    echo ""
    
    if go test $test_args $test_path; then
        echo -e "${GREEN}✓ Integration tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Integration tests failed${NC}"
        return 1
    fi
}

# Generate test report
generate_test_report() {
    echo -e "${YELLOW}Generating test report...${NC}"
    
    local report_file="integration_test_report.txt"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    cd ../.. # Go to project root
    
    cat > $report_file << EOF
Integration Test Report
Generated: $timestamp
Test timeout: $TEST_TIMEOUT
Test pattern: ${RUN_PATTERN:-"all"}
Parallel execution: $PARALLEL

=== Test Results ===
EOF

    # Run tests with JSON output for report generation
    go test -tags integration -json -timeout $TEST_TIMEOUT ./test/integration >> $report_file 2>&1 || true
    
    echo -e "${GREEN}✓ Test report generated: $report_file${NC}"
}

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test containers...${NC}"
    
    # Stop and remove any test containers
    docker ps -q --filter "label=org.testcontainers" | xargs -r docker stop 2>/dev/null || true
    docker ps -aq --filter "label=org.testcontainers" | xargs -r docker rm 2>/dev/null || true
    
    # Clean up test images if requested
    if [[ "${CLEANUP_IMAGES:-false}" == "true" ]]; then
        docker images -q --filter "dangling=true" | xargs -r docker rmi 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓ Cleanup completed${NC}"
}

# Signal handler for cleanup
trap cleanup EXIT INT TERM

# Main execution
main() {
    local action=${1:-"test"}
    local exit_code=0
    
    check_prerequisites
    
    case "$action" in
        "test")
            run_integration_tests || exit_code=1
            ;;
        "report")
            generate_test_report || exit_code=1
            ;;
        *)
            run_integration_tests || exit_code=1
            ;;
    esac
    
    if [[ $exit_code -eq 0 ]]; then
        echo -e "${GREEN}==== Integration tests completed successfully ====${NC}"
    else
        echo -e "${RED}==== Integration tests failed ====${NC}"
    fi
    
    exit $exit_code
}

# Run main function
main "$@"