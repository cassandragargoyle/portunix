#!/bin/bash
# Test Environment Setup Script for Portunix
# Prepares environment for running any Portunix tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORTUNIX_ROOT="$(dirname "$SCRIPT_DIR")"
PORTUNIX_BINARY="$PORTUNIX_ROOT/portunix"

# Test setup flags
SKIP_DOCKER_CHECK=false
SKIP_PYTHON_DEPS=false
SKIP_GO_DEPS=false
FORCE_REBUILD=false
VERBOSE=false

show_usage() {
    cat << EOF
🧪 Portunix Test Environment Setup

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    --skip-docker           Skip Docker/Podman availability check
    --skip-python           Skip Python dependencies installation
    --skip-go               Skip Go dependencies installation
    --force-rebuild         Force rebuild of Portunix binary
    --check-only            Only check prerequisites, don't install

EXAMPLES:
    $0                      # Full setup
    $0 --skip-docker        # Setup without Docker check
    $0 --check-only         # Only verify prerequisites
    $0 -v --force-rebuild   # Verbose setup with binary rebuild

This script prepares the test environment for running any Portunix tests (unit, integration, e2e).
EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                set -x
                shift
                ;;
            --skip-docker)
                SKIP_DOCKER_CHECK=true
                shift
                ;;
            --skip-python)
                SKIP_PYTHON_DEPS=true
                shift
                ;;
            --skip-go)
                SKIP_GO_DEPS=true
                shift
                ;;
            --force-rebuild)
                FORCE_REBUILD=true
                shift
                ;;
            --check-only)
                CHECK_ONLY=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for Portunix test suite..."
    
    local errors=0
    
    # Check Go
    if ! command_exists go; then
        log_error "Go is not installed or not in PATH"
        ((errors++))
    else
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_success "Go version: $go_version"
        
        # Check Go version (need 1.21+)
        if ! go version | grep -E "go1\.(2[1-9]|[3-9][0-9])" >/dev/null; then
            log_warning "Go version might be too old (need 1.21+)"
        fi
    fi
    
    # Check Python
    if ! command_exists python3; then
        log_error "Python 3 is not installed or not in PATH"
        ((errors++))
    else
        local python_version=$(python3 --version 2>&1 | awk '{print $2}')
        log_success "Python version: $python_version"
        
        # Check Python version (need 3.7+)
        if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 7) else 1)" 2>/dev/null; then
            log_error "Python version is too old (need 3.7+)"
            ((errors++))
        fi
    fi
    
    # Check pip
    if ! command_exists pip3; then
        log_warning "pip3 is not available, will try to install Python dependencies with python3 -m pip"
    else
        log_success "pip3 is available"
    fi
    
    # Check container runtime if not skipped
    if [[ "$SKIP_DOCKER_CHECK" != "true" ]]; then
        local container_runtime=""
        
        if command_exists docker; then
            if docker info >/dev/null 2>&1; then
                container_runtime="docker"
                local docker_version=$(docker --version | awk '{print $3}' | sed 's/,//')
                log_success "Docker version: $docker_version (daemon running)"
            else
                log_warning "Docker is installed but daemon is not running"
            fi
        elif command_exists podman; then
            container_runtime="podman"
            local podman_version=$(podman --version | awk '{print $3}')
            log_success "Podman version: $podman_version"
        else
            log_error "Neither Docker nor Podman is available"
            log_info "Install Docker or Podman for integration tests"
            ((errors++))
        fi
        
        if [[ -n "$container_runtime" ]]; then
            export CONTAINER_RUNTIME="$container_runtime"
        fi
    else
        log_info "Skipping container runtime check (--skip-docker)"
    fi
    
    # Check testcontainers-go dependencies
    if [[ "$SKIP_GO_DEPS" != "true" ]]; then
        log_info "Checking Go test dependencies..."
        
        if ! go list github.com/stretchr/testify/suite >/dev/null 2>&1; then
            log_warning "testify not found in go.mod, will install"
        else
            log_success "testify dependency available"
        fi
        
        if ! go list github.com/testcontainers/testcontainers-go >/dev/null 2>&1; then
            log_warning "testcontainers-go not found in go.mod, will install"
        else
            log_success "testcontainers-go dependency available"
        fi
    fi
    
    # Check project structure
    if [[ ! -f "$PORTUNIX_ROOT/go.mod" ]]; then
        log_error "go.mod not found in project root: $PORTUNIX_ROOT"
        ((errors++))
    else
        log_success "go.mod found"
    fi
    
    if [[ ! -d "$SCRIPT_DIR/unit" ]]; then
        log_error "Unit test directory not found: $SCRIPT_DIR/unit"
        ((errors++))
    else
        log_success "Unit test directory exists"
    fi
    
    if [[ ! -d "$SCRIPT_DIR/integration" ]]; then
        log_error "Integration test directory not found: $SCRIPT_DIR/integration"
        ((errors++))
    else
        log_success "Integration test directory exists"
    fi
    
    if [[ ! -d "$SCRIPT_DIR/e2e" ]]; then
        log_error "E2E test directory not found: $SCRIPT_DIR/e2e"
        ((errors++))
    else
        log_success "E2E test directory exists"
    fi
    
    if [[ $errors -gt 0 ]]; then
        log_error "Found $errors errors in prerequisites check"
        return 1
    else
        log_success "All prerequisites check passed"
        return 0
    fi
}

# Install Go dependencies
install_go_dependencies() {
    if [[ "$SKIP_GO_DEPS" == "true" ]]; then
        log_info "Skipping Go dependencies (--skip-go)"
        return 0
    fi
    
    log_info "Installing/updating Go dependencies..."
    
    cd "$PORTUNIX_ROOT"
    
    # Update go.mod
    if ! go mod tidy; then
        log_error "Failed to run go mod tidy"
        return 1
    fi
    log_success "go mod tidy completed"
    
    # Download dependencies
    if ! go mod download; then
        log_error "Failed to download Go dependencies"
        return 1
    fi
    log_success "Go dependencies downloaded"
    
    # Verify test dependencies
    local test_deps=(
        "github.com/stretchr/testify/suite"
        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "github.com/testcontainers/testcontainers-go"
        "github.com/testcontainers/testcontainers-go/wait"
    )
    
    for dep in "${test_deps[@]}"; do
        if go list "$dep" >/dev/null 2>&1; then
            log_success "✓ $dep"
        else
            log_warning "✗ $dep (not found)"
        fi
    done
    
    log_success "Go dependencies setup completed"
}

# Install Python dependencies
install_python_dependencies() {
    if [[ "$SKIP_PYTHON_DEPS" == "true" ]]; then
        log_info "Skipping Python dependencies (--skip-python)"
        return 0
    fi
    
    log_info "Installing Python dependencies..."
    
    # Check if requirements file exists (from main test suite)
    local requirements_file="$SCRIPT_DIR/requirements-test.txt"
    
    if [[ -f "$requirements_file" ]]; then
        log_info "Installing from $requirements_file"
        if command_exists pip3; then
            pip3 install --user -r "$requirements_file"
        else
            python3 -m pip install --user -r "$requirements_file"
        fi
    else
        log_info "Installing basic Python test dependencies"
        local python_deps=(
            "pytest>=7.0.0"
            "pytest-html>=3.1.0"
            "pytest-xdist>=3.0.0"
        )
        
        for dep in "${python_deps[@]}"; do
            if command_exists pip3; then
                pip3 install --user "$dep"
            else
                python3 -m pip install --user "$dep"
            fi
            log_success "Installed $dep"
        done
    fi
    
    log_success "Python dependencies installed"
}

# Build Portunix binary
build_portunix_binary() {
    log_info "Building Portunix binary..."
    
    cd "$PORTUNIX_ROOT"
    
    # Check if binary exists and is recent (unless force rebuild)
    if [[ "$FORCE_REBUILD" != "true" ]] && [[ -f "$PORTUNIX_BINARY" ]]; then
        # Check if binary is newer than main.go
        if [[ "$PORTUNIX_BINARY" -nt "main.go" ]]; then
            log_info "Portunix binary is up to date"
            return 0
        fi
    fi
    
    # Build binary
    log_info "Compiling Portunix..."
    if go build -o portunix .; then
        log_success "Portunix binary built successfully"
        
        # Make executable
        chmod +x "$PORTUNIX_BINARY"
        
        # Show version
        if [[ -x "$PORTUNIX_BINARY" ]]; then
            local version=$("$PORTUNIX_BINARY" version 2>/dev/null || echo "unknown")
            log_success "Binary version: $version"
        fi
    else
        log_error "Failed to build Portunix binary"
        return 1
    fi
}

# Setup test directories
setup_test_directories() {
    log_info "Setting up test directories..."
    
    local test_dirs=(
        "$SCRIPT_DIR/results"
        "$SCRIPT_DIR/tmp"
        "$SCRIPT_DIR/logs"
        "$SCRIPT_DIR/coverage"
    )
    
    for dir in "${test_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            log_success "Created directory: $(basename "$dir")"
        else
            log_info "Directory exists: $(basename "$dir")"
        fi
    done
    
    # Create test configuration
    cat > "$SCRIPT_DIR/test-config.env" << EOF
# Test Configuration for Portunix Test Suite
export PORTUNIX_BINARY="$PORTUNIX_BINARY"
export CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-docker}"
export TEST_RESULTS_DIR="$SCRIPT_DIR/results"
export TEST_LOGS_DIR="$SCRIPT_DIR/logs"
export TEST_TMP_DIR="$SCRIPT_DIR/tmp"
export GO_TEST_TIMEOUT="5m"
export PYTHON_TEST_TIMEOUT="300"
export INTEGRATION_TEST_PARALLEL="4"
EOF
    
    log_success "Test configuration created: test-config.env"
}

# Verify test suite
verify_test_suite() {
    log_info "Verifying test suite setup..."
    
    # Check if we can compile unit tests
    if find "$SCRIPT_DIR/unit" -name "*.go" -type f | head -1 >/dev/null 2>&1; then
        log_info "Checking unit tests compilation..."
        if go test -tags unit -c ./test/unit/... -o /dev/null 2>/dev/null; then
            log_success "Unit tests compile successfully"
        else
            log_warning "Unit tests compilation failed or no unit tests found"
        fi
    else
        log_info "No unit tests found to compile"
    fi
    
    # Check if we can compile integration tests  
    if find "$SCRIPT_DIR/integration" -name "*.go" -type f | head -1 >/dev/null 2>&1; then
        log_info "Checking integration tests compilation..."
        if go test -tags integration -c ./test/integration/... -o /dev/null 2>/dev/null; then
            log_success "Integration tests compile successfully"
        else
            log_warning "Integration tests compilation failed or no integration tests found"
        fi
    else
        log_info "No integration tests found to compile"
    fi
    
    # Check Python test syntax if they exist
    if find "$SCRIPT_DIR/e2e" -name "*.py" -type f | head -1 >/dev/null 2>&1; then
        log_info "Checking Python test syntax..."
        local python_errors=0
        for python_file in "$SCRIPT_DIR"/e2e/*.py; do
            if [[ -f "$python_file" ]]; then
                if python3 -m py_compile "$python_file" 2>/dev/null; then
                    log_success "✓ $(basename "$python_file")"
                else
                    log_warning "✗ $(basename "$python_file") has syntax errors"
                    ((python_errors++))
                fi
            fi
        done
        
        if [[ $python_errors -eq 0 ]]; then
            log_success "All Python tests have valid syntax"
        else
            log_warning "$python_errors Python test files have syntax errors"
        fi
    else
        log_info "No Python tests found to check"
    fi
    
    # Check if Portunix binary is executable
    if [[ -x "$PORTUNIX_BINARY" ]]; then
        log_success "Portunix binary is executable"
        
        # Test basic functionality
        if "$PORTUNIX_BINARY" --help >/dev/null 2>&1; then
            log_success "Portunix binary responds to --help"
        else
            log_warning "Portunix binary doesn't respond to --help"
        fi
    else
        log_error "Portunix binary is not executable: $PORTUNIX_BINARY"
        return 1
    fi
    
    log_success "Test suite verification completed"
}

# Show summary and next steps
show_summary() {
    log_info "Test Environment Setup Summary"
    echo
    echo "📁 Test Structure:"
    echo "   • Unit Tests:        $SCRIPT_DIR/unit/"
    echo "   • Integration Tests: $SCRIPT_DIR/integration/"  
    echo "   • E2E Python Tests:  $SCRIPT_DIR/e2e/"
    echo "   • Results Directory: $SCRIPT_DIR/results/"
    echo "   • Configuration:     $SCRIPT_DIR/test-config.env"
    echo
    echo "🚀 Next Steps - Running Tests:"
    echo
    echo "   # Load test configuration"
    echo "   source $SCRIPT_DIR/test-config.env"
    echo
    echo "   # Run unit tests"
    echo "   go test -tags unit ./test/unit/..."
    echo
    echo "   # Run integration tests (requires Docker/Podman)"
    echo "   go test -tags integration ./test/integration/..."
    echo
    echo "   # Run E2E Python tests"
    echo "   cd test/e2e && python3 python_integration_test.py"
    echo
    echo "   # Run all tests with coverage"
    echo "   go test -tags unit -coverprofile=test/coverage/unit.out ./test/unit/..."
    echo "   go test -tags integration -coverprofile=test/coverage/integration.out ./test/integration/..."
    echo
    echo "📝 Test Suite Status:"
    echo "   • Unit tests: Available for all components"
    echo "   • Integration tests: Require Docker/Podman for container features" 
    echo "   • E2E Python tests: Test real-world workflows and automation"
    echo
    echo "✅ Environment setup completed successfully!"
    echo
}

# Main execution
main() {
    log_info "🧪 Setting up test environment for Portunix"
    echo
    
    # Parse arguments
    parse_arguments "$@"
    
    # Check if this is check-only mode
    if [[ "${CHECK_ONLY:-false}" == "true" ]]; then
        check_prerequisites
        exit $?
    fi
    
    # Run setup steps
    if ! check_prerequisites; then
        log_error "Prerequisites check failed. Please fix the issues above."
        exit 1
    fi
    
    if ! install_go_dependencies; then
        log_error "Go dependencies installation failed"
        exit 1
    fi
    
    if ! install_python_dependencies; then
        log_error "Python dependencies installation failed"
        exit 1
    fi
    
    if ! build_portunix_binary; then
        log_error "Portunix binary build failed"
        exit 1
    fi
    
    if ! setup_test_directories; then
        log_error "Test directories setup failed"
        exit 1
    fi
    
    if ! verify_test_suite; then
        log_error "Test suite verification failed"
        exit 1
    fi
    
    show_summary
}

# Run main function with all arguments
main "$@"