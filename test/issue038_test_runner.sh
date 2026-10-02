#!/bin/bash
# Test runner for Issue #038 - Container Run Command Shorthand Flag Parsing
# This script runs all test categories for issue 038 with proper reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
PROJECT_ROOT="$(dirname "$0")/.."
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"

# Create directories
mkdir -p "$TEST_RESULTS_DIR"
mkdir -p "$COVERAGE_DIR"

# Test execution functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for Issue #038 tests..."
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "Not in project root directory or go.mod not found"
        exit 1
    fi
    
    # Check for container runtime (for E2E tests)
    CONTAINER_RUNTIME=""
    if command -v podman &> /dev/null; then
        CONTAINER_RUNTIME="podman"
        log_info "Found Podman for E2E testing"
    elif command -v docker &> /dev/null; then
        CONTAINER_RUNTIME="docker"
        log_info "Found Docker for E2E testing"
    else
        log_warn "No container runtime found - E2E tests will be skipped"
    fi
    
    # Build portunix binary
    log_info "Building portunix binary..."
    cd "$PROJECT_ROOT"
    if ! go build -o portunix .; then
        log_error "Failed to build portunix binary"
        exit 1
    fi
    
    log_info "Prerequisites check completed"
}

# Run unit tests
run_unit_tests() {
    log_info "Running Unit Tests for Issue #038..."
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests with coverage
    if go test -v -race -coverprofile="$COVERAGE_DIR/unit_coverage.out" \
       -covermode=atomic \
       -run="TestContainerRun.*Flag|TestContainerRun.*Argument|TestContainerRun.*Validation|TestContainerRun.*OriginalIssue" \
       ./cmd > "$TEST_RESULTS_DIR/unit_tests.log" 2>&1; then
        log_info "✅ Unit tests PASSED"
        
        # Generate coverage report
        go tool cover -html="$COVERAGE_DIR/unit_coverage.out" -o "$COVERAGE_DIR/unit_coverage.html"
        
        # Show coverage stats
        UNIT_COVERAGE=$(go tool cover -func="$COVERAGE_DIR/unit_coverage.out" | grep "total:" | awk '{print $3}')
        log_info "Unit test coverage: $UNIT_COVERAGE"
        
        return 0
    else
        log_error "❌ Unit tests FAILED"
        cat "$TEST_RESULTS_DIR/unit_tests.log"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running Integration Tests for Issue #038..."
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests
    if go test -v -race \
       -run="TestContainerRun.*Runtime|TestContainerRun.*Options|TestContainerRun.*Error|TestContainerRun.*Configuration" \
       ./cmd > "$TEST_RESULTS_DIR/integration_tests.log" 2>&1; then
        log_info "✅ Integration tests PASSED"
        return 0
    else
        log_error "❌ Integration tests FAILED"
        cat "$TEST_RESULTS_DIR/integration_tests.log"
        return 1
    fi
}

# Run E2E tests
run_e2e_tests() {
    if [[ -z "$CONTAINER_RUNTIME" ]]; then
        log_warn "Skipping E2E tests - no container runtime available"
        return 0
    fi
    
    log_info "Running E2E Tests for Issue #038 (using $CONTAINER_RUNTIME)..."
    
    cd "$PROJECT_ROOT"
    
    # Set timeout for E2E tests
    export E2E_TEST_TIMEOUT="300s"
    
    # Run E2E tests
    if go test -v -timeout="$E2E_TEST_TIMEOUT" \
       -run="TestContainerRunE2E" \
       ./test/e2e > "$TEST_RESULTS_DIR/e2e_tests.log" 2>&1; then
        log_info "✅ E2E tests PASSED"
        return 0
    else
        log_error "❌ E2E tests FAILED"
        cat "$TEST_RESULTS_DIR/e2e_tests.log"
        return 1
    fi
}

# Run performance benchmarks
run_benchmarks() {
    log_info "Running Performance Benchmarks for Issue #038..."
    
    cd "$PROJECT_ROOT"
    
    # Run benchmarks
    if go test -bench="BenchmarkContainerRunFlagParsing" \
       -benchmem \
       ./test/e2e > "$TEST_RESULTS_DIR/benchmarks.log" 2>&1; then
        log_info "✅ Benchmarks completed"
        
        # Show benchmark results
        grep "BenchmarkContainerRunFlagParsing" "$TEST_RESULTS_DIR/benchmarks.log" || true
        
        return 0
    else
        log_warn "⚠️ Benchmarks failed (non-critical)"
        return 0  # Don't fail the whole suite for benchmark issues
    fi
}

# Validate test results
validate_results() {
    log_info "Validating test results against acceptance criteria..."
    
    local failed=0
    
    # Check unit test coverage
    if [[ -f "$COVERAGE_DIR/unit_coverage.out" ]]; then
        COVERAGE=$(go tool cover -func="$COVERAGE_DIR/unit_coverage.out" | grep "total:" | awk '{print $3}' | sed 's/%//')
        
        if (( $(echo "$COVERAGE >= 85" | bc -l) )); then
            log_info "✅ Coverage target met: ${COVERAGE}% (target: 85%)"
        else
            log_error "❌ Coverage target not met: ${COVERAGE}% (target: 85%)"
            failed=1
        fi
    fi
    
    # Check if all test categories passed
    local categories=("unit" "integration")
    if [[ -n "$CONTAINER_RUNTIME" ]]; then
        categories+=("e2e")
    fi
    
    for category in "${categories[@]}"; do
        if [[ -f "$TEST_RESULTS_DIR/${category}_tests.log" ]]; then
            if grep -q "FAIL" "$TEST_RESULTS_DIR/${category}_tests.log"; then
                log_error "❌ $category tests have failures"
                failed=1
            else
                log_info "✅ $category tests passed"
            fi
        fi
    done
    
    return $failed
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$TEST_RESULTS_DIR/issue038_test_report.md"
    
    cat > "$report_file" << EOF
# Issue #038 Test Results

**Test Date:** $(date)
**Container Runtime:** ${CONTAINER_RUNTIME:-"Not Available"}
**Go Version:** $(go version)

## Test Summary

EOF

    # Add unit test results
    if [[ -f "$TEST_RESULTS_DIR/unit_tests.log" ]]; then
        echo "### Unit Tests" >> "$report_file"
        if grep -q "PASS" "$TEST_RESULTS_DIR/unit_tests.log"; then
            echo "✅ **PASSED**" >> "$report_file"
        else
            echo "❌ **FAILED**" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # Add integration test results
    if [[ -f "$TEST_RESULTS_DIR/integration_tests.log" ]]; then
        echo "### Integration Tests" >> "$report_file"
        if grep -q "PASS" "$TEST_RESULTS_DIR/integration_tests.log"; then
            echo "✅ **PASSED**" >> "$report_file"
        else
            echo "❌ **FAILED**" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # Add E2E test results
    if [[ -f "$TEST_RESULTS_DIR/e2e_tests.log" ]]; then
        echo "### E2E Tests" >> "$report_file"
        if grep -q "PASS" "$TEST_RESULTS_DIR/e2e_tests.log"; then
            echo "✅ **PASSED**" >> "$report_file"
        else
            echo "❌ **FAILED**" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    # Add coverage information
    if [[ -f "$COVERAGE_DIR/unit_coverage.out" ]]; then
        COVERAGE=$(go tool cover -func="$COVERAGE_DIR/unit_coverage.out" | grep "total:" | awk '{print $3}')
        echo "### Coverage" >> "$report_file"
        echo "**Total Coverage:** $COVERAGE" >> "$report_file"
        echo "" >> "$report_file"
        echo "Coverage report: [unit_coverage.html](../coverage/unit_coverage.html)" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # Add file locations
    echo "### Test Artifacts" >> "$report_file"
    echo "- Unit test log: [unit_tests.log](unit_tests.log)" >> "$report_file"
    echo "- Integration test log: [integration_tests.log](integration_tests.log)" >> "$report_file"
    if [[ -f "$TEST_RESULTS_DIR/e2e_tests.log" ]]; then
        echo "- E2E test log: [e2e_tests.log](e2e_tests.log)" >> "$report_file"
    fi
    if [[ -f "$TEST_RESULTS_DIR/benchmarks.log" ]]; then
        echo "- Benchmarks: [benchmarks.log](benchmarks.log)" >> "$report_file"
    fi
    
    log_info "Test report generated: $report_file"
}

# Main execution
main() {
    log_info "🚀 Starting Issue #038 Test Suite"
    log_info "Testing: Container Run Command Shorthand Flag Parsing"
    echo
    
    local failed=0
    
    # Check prerequisites
    check_prerequisites || exit 1
    
    # Run test categories
    run_unit_tests || failed=1
    echo
    
    run_integration_tests || failed=1
    echo
    
    run_e2e_tests || failed=1
    echo
    
    run_benchmarks
    echo
    
    # Validate and report
    validate_results || failed=1
    generate_report
    
    if [[ $failed -eq 0 ]]; then
        log_info "🎉 All Issue #038 tests PASSED!"
        log_info "Test results available in: $TEST_RESULTS_DIR"
        exit 0
    else
        log_error "💥 Some Issue #038 tests FAILED!"
        log_error "Check test results in: $TEST_RESULTS_DIR"
        exit 1
    fi
}

# Script options
case "${1:-}" in
    "unit")
        check_prerequisites
        run_unit_tests
        ;;
    "integration")
        check_prerequisites
        run_integration_tests
        ;;
    "e2e")
        check_prerequisites
        run_e2e_tests
        ;;
    "benchmarks")
        check_prerequisites
        run_benchmarks
        ;;
    "report")
        generate_report
        ;;
    "--help"|"-h")
        echo "Usage: $0 [unit|integration|e2e|benchmarks|report]"
        echo "  unit        - Run only unit tests"
        echo "  integration - Run only integration tests"
        echo "  e2e         - Run only E2E tests"
        echo "  benchmarks  - Run only benchmarks"
        echo "  report      - Generate test report only"
        echo "  (no args)   - Run full test suite"
        exit 0
        ;;
    *)
        main
        ;;
esac