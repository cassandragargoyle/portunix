#!/bin/bash
# Integration Test Runner for Portunix
# Shell wrapper for Python-based integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$PROJECT_ROOT/test"
PYTHON_RUNNER="$TEST_DIR/scripts/test-integration.py"

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Python 3 is not installed${NC}"
    exit 1
fi

# Check if test runner exists
if [ ! -f "$PYTHON_RUNNER" ]; then
    echo -e "${RED}❌ Python test runner not found: $PYTHON_RUNNER${NC}"
    exit 1
fi

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -q, --quick              Run quick test (Ubuntu 22.04 only)"
    echo "  -f, --full-suite         Run complete test suite (all distributions)"
    echo "  -d, --distribution NAME  Run specific distribution test"
    echo "  -l, --list               List available distributions"
    echo "  -p, --parallel           Run tests in parallel"
    echo "  -c, --cleanup            Clean up test containers"
    echo "  -v, --verbose            Verbose output"
    echo "      --html-report FILE   Generate HTML report"
    echo "  -h, --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --quick"
    echo "  $0 -q"
    echo "  $0 --full-suite --parallel"
    echo "  $0 -f -p"
    echo "  $0 --distribution ubuntu-22"
    echo "  $0 -d ubuntu-22"
}

# Check for help flag
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    show_usage
    exit 0
fi

# Check if no arguments provided
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}⚠️  No test mode specified${NC}"
    echo ""
    show_usage
    exit 1
fi

# Setup log directory and file
LOG_DIR="$TEST_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/pip-install-$(date +%Y%m%d-%H%M%S).log"

# Ensure virtual environment exists and is activated
VENV_DIR="$TEST_DIR/venv"
if [ ! -d "$VENV_DIR" ]; then
    echo -e "${BLUE}📦 Creating Python virtual environment...${NC}"
    python3 -m venv "$VENV_DIR"
fi

# Activate virtual environment
source "$VENV_DIR/bin/activate"

# Check and install required packages
echo -e "${BLUE}📦 Checking Python dependencies...${NC}"
REQUIRED_PACKAGES="pytest pytest-xdist pytest-html"
PACKAGES_TO_INSTALL=""
for package in $REQUIRED_PACKAGES; do
    if ! pip show "$package" > /dev/null 2>&1; then
        PACKAGES_TO_INSTALL="$PACKAGES_TO_INSTALL $package"
    fi
done

if [ -n "$PACKAGES_TO_INSTALL" ]; then
    echo -e "${YELLOW}📝 Installing missing packages (see $LOG_FILE for details)...${NC}"
    echo "Installing packages:$PACKAGES_TO_INSTALL" >> "$LOG_FILE"
    pip install $PACKAGES_TO_INSTALL >> "$LOG_FILE" 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Packages installed successfully${NC}"
    else
        echo -e "${RED}❌ Some packages failed to install. Check $LOG_FILE for details${NC}"
    fi
fi

# Check if requirements-test.txt exists and install from it
REQUIREMENTS_FILE="$PROJECT_ROOT/requirements-test.txt"
if [ -f "$REQUIREMENTS_FILE" ]; then
    echo -e "${BLUE}📦 Installing from requirements-test.txt (see $LOG_FILE for details)...${NC}"
    pip install -r "$REQUIREMENTS_FILE" >> "$LOG_FILE" 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Requirements installed successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Some requirements failed to install. Check $LOG_FILE for details${NC}"
    fi
fi

# Run the Python test runner with all arguments
echo -e "${GREEN}🚀 Running integration tests...${NC}"
python3 "$PYTHON_RUNNER" "$@"
exit_code=$?

# Deactivate virtual environment
deactivate

# Exit with the same code as the Python runner
exit $exit_code