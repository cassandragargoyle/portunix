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

# Setup log directory and file (per ADR-039: deps managed by uv)
LOG_DIR="$TEST_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/uv-sync-$(date +%Y%m%d-%H%M%S).log"

# Check uv is installed
if ! command -v uv &> /dev/null; then
    echo -e "${RED}❌ uv is not installed${NC}"
    echo "Install with: portunix install uv"
    echo "Or: curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
fi

# Provision .venv with test dependencies (idempotent)
echo -e "${BLUE}📦 Syncing Python test dependencies via uv...${NC}"
(cd "$PROJECT_ROOT" && uv sync --group test) >> "$LOG_FILE" 2>&1
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ uv sync failed. Check $LOG_FILE for details${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Test dependencies ready${NC}"

# Run the Python test runner via uv run (no activation needed)
echo -e "${GREEN}🚀 Running integration tests...${NC}"
(cd "$PROJECT_ROOT" && uv run python "$PYTHON_RUNNER" "$@")
exit_code=$?

exit $exit_code