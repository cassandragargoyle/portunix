#!/bin/bash
# test/scripts/issue-012-test-runner.sh
# Complete automated test program for Issue #012 PowerShell Installation
# Tests PowerShell installation across multiple Linux distributions using containers

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$TEST_ROOT")"
RESULTS_DIR="$TEST_ROOT/results"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
REPORT_FILE="$RESULTS_DIR/issue-012-test-report-${TIMESTAMP}.html"
LOG_DIR="$RESULTS_DIR/logs-${TIMESTAMP}"

# Portunix binary path
PORTUNIX_BIN="${PROJECT_ROOT}/portunix"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
START_TIME=$(date +%s)

# Test configuration
DISTRIBUTIONS=(
    "ubuntu-22:ubuntu:ubuntu:22.04"
    "ubuntu-24:ubuntu:ubuntu:24.04"
    "debian-11:debian:debian:bullseye"
    "debian-12:debian:debian:bookworm"
    "fedora-39:fedora:fedora:39"
    "fedora-40:fedora:fedora:40"
    "rocky-9:rocky:rockylinux:9"
    "mint-21:mint:linuxmintd/mint21-amd64:latest"
)

# Container prefix for test containers
CONTAINER_PREFIX="portunix-test-ps"

usage() {
    cat << EOF
Issue #012 PowerShell Installation Test Suite
Container Engine: Podman (rootless mode)

Usage: $0 [OPTIONS]

Options:
  --full-suite         Run complete test suite (all distributions)
  --quick              Run quick test (Ubuntu 22.04 only)
  --distribution NAME  Run specific distribution test
  --list-distributions List available distributions
  --with-cleanup       Clean up test environments after completion
  --keep-failed        Keep failed test containers for debugging
  --report-only        Generate report from existing results
  --verbose            Enable verbose output
  --help               Show this help message

Available distributions:
  ubuntu-22, ubuntu-24, debian-11, debian-12
  fedora-39, fedora-40, rocky-9, mint-21

Examples:
  $0 --full-suite --with-cleanup
  $0 --quick
  $0 --distribution ubuntu-22
  $0 --full-suite --keep-failed --verbose

Note: This test uses Podman in rootless mode for better security
      and compatibility with the Portunix container management.

EOF
}

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

skip() {
    echo -e "${YELLOW}[SKIPPED]${NC} $1"
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
}

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

debug() {
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${MAGENTA}[DEBUG]${NC} $1"
    fi
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if portunix binary exists
    if [[ ! -f "$PORTUNIX_BIN" ]]; then
        echo -e "${RED}[ERROR]${NC} Portunix binary not found at: $PORTUNIX_BIN"
        echo -e "${RED}[ERROR]${NC} Please build portunix first: go build -o portunix"
        exit 1
    fi
    
    # Check if Podman is installed
    if ! command -v podman &> /dev/null; then
        warning "Podman is not installed. Attempting to install via portunix..."
        if ! "$PORTUNIX_BIN" install podman -y; then
            echo -e "${RED}[ERROR]${NC} Failed to install Podman"
            exit 1
        fi
    fi
    
    # Check if Podman is working (rootless mode)
    if ! podman info &> /dev/null; then
        echo -e "${RED}[ERROR]${NC} Podman is not working properly"
        echo -e "${RED}[ERROR]${NC} Please check your Podman installation"
        exit 1
    fi
    
    info "Prerequisites check passed"
}

# Create results directory structure
setup_directories() {
    log "Setting up test directories..."
    mkdir -p "$RESULTS_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$TEST_ROOT/configs"
    mkdir -p "$TEST_ROOT/fixtures"
    info "Test directories created"
}

# Container management functions
create_test_container() {
    local name="$1"
    local image="$2"
    local container_name="${CONTAINER_PREFIX}-${name}"
    
    debug "Creating container: $container_name from image: $image"
    
    # Remove existing container if it exists
    if podman ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
        debug "Removing existing container: $container_name"
        podman rm -f "$container_name" &> /dev/null
    fi
    
    # Create and start container (Podman rootless mode)
    local container_output
    container_output=$(podman run -d \
        --name "$container_name" \
        --hostname "$name" \
        -e "DEBIAN_FRONTEND=noninteractive" \
        "$image" \
        /bin/sh -c "tail -f /dev/null" 2>&1)
    
    if [[ $? -eq 0 ]]; then
        # Install basic dependencies (without debug output to avoid conflicts)
        podman exec "$container_name" sh -c "
            if command -v apt-get &> /dev/null; then
                apt-get update && apt-get install -y sudo curl wget ca-certificates lsb-release software-properties-common
            elif command -v dnf &> /dev/null; then
                dnf install -y sudo curl wget ca-certificates redhat-lsb-core
            elif command -v yum &> /dev/null; then
                yum install -y sudo curl wget ca-certificates redhat-lsb-core
            fi
        " &> /dev/null
        
        return 0
    else
        return 1
    fi
}

copy_portunix_to_container() {
    local container_name="$1"
    
    debug "Copying portunix binary to container: $container_name"
    
    if podman cp "$PORTUNIX_BIN" "${container_name}:/usr/local/bin/portunix" &> /dev/null; then
        podman exec "$container_name" chmod +x /usr/local/bin/portunix &> /dev/null
        return 0
    else
        return 1
    fi
}

test_powershell_installation() {
    local container_name="$1"
    local variant="$2"
    
    debug "Testing PowerShell installation on $container_name with variant: $variant"
    
    # Execute installation and capture output
    local install_output
    install_output=$(podman exec "$container_name" /usr/local/bin/portunix install powershell --variant "$variant" 2>&1)
    local install_result=$?
    
    # Log the installation output for debugging  
    if [[ "$VERBOSE" == true ]] || [[ $install_result -ne 0 ]]; then
        echo "PowerShell installation output:"
        echo "$install_output"
    fi
    
    return $install_result
}

verify_powershell_installation() {
    local container_name="$1"
    
    debug "Verifying PowerShell installation on $container_name"
    
    # Check if pwsh command exists and works
    local verify_output
    verify_output=$(podman exec "$container_name" sh -c "
        echo 'Checking PowerShell binary location:'
        command -v pwsh
        echo ''
        echo 'PowerShell version:'
        pwsh --version 2>&1
        echo ''
        echo 'PowerShell test command:'
        pwsh -c 'Write-Host \"PowerShell is working: \$PSVersionTable.PSVersion\"' 2>&1
    " 2>&1)
    
    local verify_result=$?
    
    if [[ "$VERBOSE" == true ]] || [[ $verify_result -ne 0 ]]; then
        echo "PowerShell verification output:"
        echo "$verify_output"
    fi
    
    # Check if pwsh command exists and returns version
    if echo "$verify_output" | grep -q "PowerShell" && [[ $verify_result -eq 0 ]]; then
        return 0
    else
        return 1
    fi
}

test_powershell_fallback() {
    local container_name="$1"
    
    debug "Testing PowerShell fallback installation (snap) on $container_name"
    
    # Try snap installation as fallback
    if podman exec "$container_name" /usr/local/bin/portunix install powershell --variant snap &> /dev/null; then
        return verify_powershell_installation "$container_name"
    else
        return 1
    fi
}

# Main test execution functions
run_distribution_test() {
    local name="$1"
    local variant="$2"
    local image="$3"
    
    local container_name="${CONTAINER_PREFIX}-${name}"
    local log_file="${LOG_DIR}/${name}.log"
    
    info "Testing distribution: $name"
    
    {
        echo "════════════════════════════════════════════════════════════════"
        echo "🧪 DISTRIBUTION TEST: $name"
        echo "════════════════════════════════════════════════════════════════"
        echo "📦 Container: $container_name"
        echo "🐧 Image: $image"  
        echo "🔧 Variant: $variant"
        echo "⏰ Started: $(date)"
        echo "🌍 Test Environment: $(whoami)@$(hostname)"
        echo "════════════════════════════════════════════════════════════════"
        echo ""
        
        # Create and start container
        echo "🔨 STEP 1/5: Creating container..."
        echo "Command: podman run -d --name $container_name $image"
        if ! create_test_container "$name" "$image"; then
            echo "❌ Failed to create container"
            echo "Container creation failed at $(date)"
            return 1
        fi
        echo "✅ Container created successfully"
        
        # Show container info
        echo ""
        echo "📊 Container Information:"
        podman inspect "$container_name" --format "ID: {{.Id}}" 2>/dev/null || echo "Could not get container ID"
        podman inspect "$container_name" --format "Status: {{.State.Status}}" 2>/dev/null || echo "Could not get container status"  
        podman exec "$container_name" cat /etc/os-release 2>/dev/null | head -5 || echo "Could not get OS info"
        
        echo ""
        echo "🔨 STEP 2/5: Copying portunix binary..."
        echo "Source: $PORTUNIX_BIN"
        echo "Target: $container_name:/usr/local/bin/portunix"
        if ! copy_portunix_to_container "$container_name"; then
            echo "❌ Failed to copy portunix binary"
            return 1
        fi
        
        # Verify binary
        local binary_info
        binary_info=$(podman exec "$container_name" ls -la /usr/local/bin/portunix 2>&1)
        echo "✅ Binary copied successfully: $binary_info"
        
        echo ""
        echo "🔨 STEP 3/5: Installing PowerShell via $variant variant..."
        echo "Command: /usr/local/bin/portunix install powershell --variant $variant"
        echo "Installation started at: $(date)"
        echo "--- Installation Output ---"
        
        if ! test_powershell_installation "$container_name" "$variant"; then
            echo "--- End Installation Output ---"
            echo "❌ PowerShell installation failed via $variant variant"
            
            # Try fallback
            echo ""
            echo "🔄 FALLBACK: Attempting snap installation..."
            echo "Command: /usr/local/bin/portunix install powershell --variant snap"
            echo "--- Fallback Installation Output ---"
            
            if test_powershell_fallback "$container_name"; then
                echo "--- End Fallback Installation Output ---"
                echo "✅ Fallback installation succeeded"
            else
                echo "--- End Fallback Installation Output ---"
                echo "❌ Fallback installation also failed"
                echo "Test failed at: $(date)"
                return 1
            fi
        else
            echo "--- End Installation Output ---"
            echo "✅ PowerShell installation succeeded via $variant variant"
        fi
        
        echo ""
        echo "🔨 STEP 4/5: Verifying PowerShell installation..."
        echo "Verification started at: $(date)"
        echo "--- Verification Output ---"
        
        if ! verify_powershell_installation "$container_name"; then
            echo "--- End Verification Output ---"
            echo "❌ PowerShell verification failed"
            echo "Test failed at: $(date)"
            return 1
        fi
        
        echo "--- End Verification Output ---"
        echo "✅ PowerShell verification successful"
        
        echo ""
        echo "🔨 STEP 5/5: Final system state check..."
        echo "Checking installed packages and system state..."
        
        # Show final state
        podman exec "$container_name" sh -c "
            echo 'Installed PowerShell packages:'
            if command -v dpkg &> /dev/null; then
                dpkg -l | grep -i powershell || echo 'No PowerShell packages found via dpkg'
            elif command -v rpm &> /dev/null; then
                rpm -qa | grep -i powershell || echo 'No PowerShell packages found via rpm'
            fi
            echo ''
            echo 'PowerShell executable:'
            which pwsh || echo 'pwsh not found in PATH'
            echo ''
            echo 'Available PowerShell commands:'
            pwsh -c 'Get-Command | Select-Object -First 5 Name' 2>/dev/null || echo 'Could not list PowerShell commands'
        " 2>&1
        
        echo ""
        echo "════════════════════════════════════════════════════════════════"
        echo "✅ TEST COMPLETED SUCCESSFULLY"
        echo "🎉 Distribution: $name"
        echo "⏰ Completed: $(date)"
        echo "⏱️  Duration: $(($(date +%s) - $(date -d \"$start_time\" +%s 2>/dev/null || echo 0))) seconds"
        echo "════════════════════════════════════════════════════════════════"
        
    } > "$log_file" 2>&1
    
    local result=$?
    
    # Clean up container if successful and cleanup requested
    if [[ $result -eq 0 && "$WITH_CLEANUP" == true ]]; then
        debug "Cleaning up successful container: $container_name"
        podman rm -f "$container_name" &> /dev/null
    elif [[ $result -ne 0 && "$KEEP_FAILED" != true && "$WITH_CLEANUP" == true ]]; then
        debug "Cleaning up failed container: $container_name"
        podman rm -f "$container_name" &> /dev/null
    fi
    
    return $result
}

run_full_distribution_tests() {
    log "Running full distribution test suite"
    
    for distro_info in "${DISTRIBUTIONS[@]}"; do
        IFS=':' read -r name variant base_image version <<< "$distro_info"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        local image="${base_image}:${version}"
        
        if run_distribution_test "$name" "$variant" "$image"; then
            success "PowerShell installation succeeded on $name"
        else
            error "PowerShell installation failed on $name"
        fi
    done
}

run_quick_test() {
    log "Running quick test (Ubuntu 22.04 only)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if run_distribution_test "ubuntu-22" "ubuntu" "ubuntu:22.04"; then
        success "Quick test passed"
    else
        error "Quick test failed"
    fi
}

run_specific_distribution_test() {
    local dist_name="$1"
    local found=false
    
    for distro_info in "${DISTRIBUTIONS[@]}"; do
        IFS=':' read -r name variant base_image version <<< "$distro_info"
        
        if [[ "$name" == "$dist_name" ]]; then
            found=true
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
            log "Running test for distribution: $name"
            
            local image="${base_image}:${version}"
            
            if run_distribution_test "$name" "$variant" "$image"; then
                success "PowerShell installation succeeded on $name"
            else
                error "PowerShell installation failed on $name"
            fi
            break
        fi
    done
    
    if [[ "$found" == false ]]; then
        error "Distribution '$dist_name' not found"
        error "Use --list-distributions to see available options"
        exit 1
    fi
}

list_distributions() {
    echo "Available distributions for testing:"
    echo ""
    printf "%-15s %-15s %-30s\n" "Name" "Variant" "Image"
    echo "--------------------------------------------------------"
    
    for distro_info in "${DISTRIBUTIONS[@]}"; do
        IFS=':' read -r name variant base_image version <<< "$distro_info"
        printf "%-15s %-15s %-30s\n" "$name" "$variant" "${base_image}:${version}"
    done
}

cleanup_all_containers() {
    log "Cleaning up all test containers..."
    
    local containers=$(podman ps -a --format '{{.Names}}' | grep "^${CONTAINER_PREFIX}-" || true)
    
    if [[ -n "$containers" ]]; then
        echo "$containers" | while read -r container; do
            debug "Removing container: $container"
            podman rm -f "$container" &> /dev/null
        done
        info "All test containers removed"
    else
        info "No test containers found to clean up"
    fi
}

generate_html_report() {
    local report_file="$1"
    
    log "Generating HTML test report..."
    
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local minutes=$((duration / 60))
    local seconds=$((duration % 60))
    
    local pass_rate=0
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Issue #012 Test Report - ${TIMESTAMP}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 30px;
        }
        h1 {
            margin: 0;
            font-size: 2em;
        }
        .subtitle {
            opacity: 0.9;
            margin-top: 10px;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .card h3 {
            margin-top: 0;
            color: #667eea;
        }
        .stat {
            font-size: 2em;
            font-weight: bold;
        }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .total { color: #007bff; }
        .rate {
            font-size: 3em;
            font-weight: bold;
            text-align: center;
            margin: 20px 0;
        }
        .details {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background: #f8f9fa;
            font-weight: 600;
        }
        .status-pass {
            background: #d4edda;
            color: #155724;
            padding: 4px 8px;
            border-radius: 4px;
            display: inline-block;
        }
        .status-fail {
            background: #f8d7da;
            color: #721c24;
            padding: 4px 8px;
            border-radius: 4px;
            display: inline-block;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🧪 Issue #012: PowerShell Linux Installation</h1>
        <div class="subtitle">Test Report - ${TIMESTAMP}</div>
        <div class="subtitle">Duration: ${minutes}m ${seconds}s</div>
    </div>
    
    <div class="summary">
        <div class="card">
            <h3>Total Tests</h3>
            <div class="stat total">${TOTAL_TESTS}</div>
        </div>
        <div class="card">
            <h3>Passed</h3>
            <div class="stat passed">${PASSED_TESTS}</div>
        </div>
        <div class="card">
            <h3>Failed</h3>
            <div class="stat failed">${FAILED_TESTS}</div>
        </div>
        <div class="card">
            <h3>Pass Rate</h3>
            <div class="rate $([ $pass_rate -ge 80 ] && echo 'passed' || echo 'failed')">${pass_rate}%</div>
        </div>
    </div>
    
    <div class="details">
        <h2>Test Results by Distribution</h2>
        <table>
            <thead>
                <tr>
                    <th>Distribution</th>
                    <th>Variant</th>
                    <th>Image</th>
                    <th>Status</th>
                    <th>Log File</th>
                </tr>
            </thead>
            <tbody>
EOF
    
    # Add test results to table
    for distro_info in "${DISTRIBUTIONS[@]}"; do
        IFS=':' read -r name variant base_image version <<< "$distro_info"
        local log_file="${LOG_DIR}/${name}.log"
        local status="Not Run"
        local status_class="status-fail"
        
        if [[ -f "$log_file" ]]; then
            if grep -q "Test completed successfully" "$log_file"; then
                status="Passed"
                status_class="status-pass"
            else
                status="Failed"
                status_class="status-fail"
            fi
        fi
        
        echo "<tr>" >> "$report_file"
        echo "    <td>$name</td>" >> "$report_file"
        echo "    <td>$variant</td>" >> "$report_file"
        echo "    <td>${base_image}:${version}</td>" >> "$report_file"
        echo "    <td><span class='$status_class'>$status</span></td>" >> "$report_file"
        echo "    <td><a href='logs-${TIMESTAMP}/${name}.log'>${name}.log</a></td>" >> "$report_file"
        echo "</tr>" >> "$report_file"
    done
    
    cat >> "$report_file" << EOF
            </tbody>
        </table>
    </div>
    
    <div class="footer">
        <p>Generated by Portunix Test Suite</p>
        <p>Report created at $(date)</p>
    </div>
</body>
</html>
EOF
    
    info "HTML report generated: $report_file"
}

show_test_summary() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local minutes=$((duration / 60))
    local seconds=$((duration % 60))
    
    echo ""
    echo "================================================"
    echo "           TEST EXECUTION SUMMARY"
    echo "================================================"
    echo -e "Total Tests:    ${BLUE}${TOTAL_TESTS}${NC}"
    echo -e "Passed:         ${GREEN}${PASSED_TESTS}${NC}"
    echo -e "Failed:         ${RED}${FAILED_TESTS}${NC}"
    echo -e "Skipped:        ${YELLOW}${SKIPPED_TESTS}${NC}"
    echo ""
    
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        local pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo -e "Success Rate:   ${GREEN}${pass_rate}%${NC}"
    fi
    
    echo -e "Duration:       ${CYAN}${minutes}m ${seconds}s${NC}"
    echo ""
    echo "Detailed report: file://${REPORT_FILE}"
    echo "Logs directory:  ${LOG_DIR}"
    echo "================================================"
}

# Main execution
main() {
    # Parse command line arguments
    SUITE_TYPE=""
    WITH_CLEANUP=false
    KEEP_FAILED=false
    VERBOSE=false
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
            --list-distributions)
                list_distributions
                exit 0
                ;;
            --with-cleanup)
                WITH_CLEANUP=true
                shift
                ;;
            --keep-failed)
                KEEP_FAILED=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --report-only)
                if [[ -d "$RESULTS_DIR" ]]; then
                    LATEST_LOG_DIR=$(ls -dt "$RESULTS_DIR"/logs-* 2>/dev/null | head -1)
                    if [[ -n "$LATEST_LOG_DIR" ]]; then
                        LOG_DIR="$LATEST_LOG_DIR"
                        TIMESTAMP=$(basename "$LATEST_LOG_DIR" | sed 's/logs-//')
                        REPORT_FILE="$RESULTS_DIR/issue-012-test-report-${TIMESTAMP}.html"
                        generate_html_report "$REPORT_FILE"
                        echo "Report generated: file://${REPORT_FILE}"
                    else
                        error "No test results found"
                    fi
                else
                    error "No results directory found"
                fi
                exit 0
                ;;
            --help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Validate arguments
    if [[ -z "$SUITE_TYPE" ]]; then
        echo ""
        echo -e "${RED}ERROR: No test suite type specified!${NC}"
        echo ""
        echo -e "${YELLOW}Available test modes:${NC}"
        echo ""
        echo -e "  ${GREEN}Quick Test:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --quick"
        echo "    Tests only Ubuntu 22.04 - fast verification"
        echo ""
        echo -e "  ${GREEN}Full Suite:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --full-suite"
        echo "    Tests all 8 Linux distributions"
        echo ""
        echo -e "  ${GREEN}Specific Distribution:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --distribution ubuntu-22"
        echo "    Tests a single specified distribution"
        echo ""
        echo -e "  ${GREEN}Full Suite with Cleanup:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --full-suite --with-cleanup"
        echo "    Tests all distributions and removes containers after completion"
        echo ""
        echo -e "  ${GREEN}Verbose Mode with Failed Container Preservation:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --full-suite --keep-failed --verbose"
        echo "    Detailed output, keeps failed containers for debugging"
        echo ""
        echo -e "${CYAN}For complete help and all options:${NC}"
        echo "    ./test/scripts/issue-012-test-runner.sh --help"
        echo ""
        exit 1
    fi
    
    # Setup environment
    check_prerequisites
    setup_directories
    
    # Print test header
    echo ""
    echo "================================================"
    echo "     Issue #012 PowerShell Installation"
    echo "           Test Suite Execution"
    echo "================================================"
    echo -e "Container Engine: ${CYAN}Podman (rootless)${NC}"
    echo -e "Suite Type:     ${CYAN}${SUITE_TYPE}${NC}"
    echo -e "Timestamp:      ${CYAN}${TIMESTAMP}${NC}"
    echo -e "Cleanup:        ${CYAN}${WITH_CLEANUP}${NC}"
    echo -e "Keep Failed:    ${CYAN}${KEEP_FAILED}${NC}"
    echo -e "Verbose:        ${CYAN}${VERBOSE}${NC}"
    echo "================================================"
    echo ""
    
    # Run tests based on suite type
    case "$SUITE_TYPE" in
        "full-suite")
            run_full_distribution_tests
            ;;
        "quick")
            run_quick_test
            ;;
        "distribution")
            run_specific_distribution_test "$SPECIFIC_DISTRO"
            ;;
    esac
    
    # Generate report
    generate_html_report "$REPORT_FILE"
    
    # Cleanup if requested
    if [[ "$WITH_CLEANUP" == true ]]; then
        cleanup_all_containers
    fi
    
    # Show summary
    show_test_summary
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"