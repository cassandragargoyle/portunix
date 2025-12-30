#!/bin/bash

# Setup Python Virtual Environment for Portunix
# Usage: ./scripts/setup-venv.sh [--with-tests]
#
# Creates a virtual environment in .venv/ directory
# Optionally installs test dependencies with --with-tests flag

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VENV_DIR="$PROJECT_ROOT/.venv"

print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Detect Python command (python3 on Linux/macOS, python on Windows)
detect_python() {
    PYTHON_CMD=""

    # Try python3 first (Linux/macOS)
    if command -v python3 >/dev/null 2>&1; then
        # Verify it actually works (not Windows stub)
        if python3 --version >/dev/null 2>&1; then
            PYTHON_CMD="python3"
        fi
    fi

    # Try python if python3 didn't work (Windows)
    if [ -z "$PYTHON_CMD" ] && command -v python >/dev/null 2>&1; then
        # Check if it's Python 3 and actually works
        if python --version 2>&1 | grep -q "Python 3"; then
            PYTHON_CMD="python"
        fi
    fi

    # Check if we found a working Python
    if [ -z "$PYTHON_CMD" ]; then
        print_error "Python 3 is not installed or not working"
        echo "Install Python 3.8+ or use: portunix install python"
        exit 1
    fi

    print_success "Found Python: $($PYTHON_CMD --version 2>&1)"
}

# Create virtual environment
create_venv() {
    print_step "Creating virtual environment in $VENV_DIR..."

    if [ -d "$VENV_DIR" ]; then
        print_warning "Virtual environment already exists"
        read -p "Do you want to recreate it? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$VENV_DIR"
        else
            print_step "Using existing virtual environment"
            return 0
        fi
    fi

    $PYTHON_CMD -m venv "$VENV_DIR"
    print_success "Virtual environment created"
}

# Activate virtual environment and install dependencies
install_dependencies() {
    print_step "Installing dependencies..."

    # Determine activation script based on OS
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
        ACTIVATE_SCRIPT="$VENV_DIR/Scripts/activate"
    else
        ACTIVATE_SCRIPT="$VENV_DIR/bin/activate"
    fi

    # shellcheck disable=SC1090
    source "$ACTIVATE_SCRIPT"

    # Upgrade pip
    pip install --upgrade pip >/dev/null 2>&1
    print_success "pip upgraded"

    # Install main requirements if exists and not empty
    if [ -f "$PROJECT_ROOT/requirements.txt" ]; then
        # Check if file has actual dependencies (not just comments)
        if grep -v '^#' "$PROJECT_ROOT/requirements.txt" | grep -v '^$' | grep -v '^\s*$' >/dev/null 2>&1; then
            pip install -r "$PROJECT_ROOT/requirements.txt"
            print_success "Main dependencies installed"
        else
            print_success "No main dependencies to install (requirements.txt has only comments)"
        fi
    fi

    # Install test requirements if requested
    if [ "$INSTALL_TESTS" = "true" ] && [ -f "$PROJECT_ROOT/test/requirements-test.txt" ]; then
        pip install -r "$PROJECT_ROOT/test/requirements-test.txt"
        print_success "Test dependencies installed"
    fi
}

# Show activation instructions
show_instructions() {
    echo
    print_step "Setup complete!"
    echo
    echo "To activate the virtual environment:"
    echo
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
        echo "  Windows (Git Bash/MSYS):  source .venv/Scripts/activate"
        echo "  Windows (PowerShell):     .venv\\Scripts\\Activate.ps1"
        echo "  Windows (CMD):            .venv\\Scripts\\activate.bat"
    else
        echo "  source .venv/bin/activate"
    fi
    echo
    echo "To deactivate: deactivate"
    echo
}

# Main
main() {
    echo "================================"
    echo "Portunix Python Environment Setup"
    echo "================================"
    echo

    INSTALL_TESTS="false"

    # Parse arguments
    for arg in "$@"; do
        case $arg in
            --with-tests)
                INSTALL_TESTS="true"
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [--with-tests]"
                echo
                echo "Options:"
                echo "  --with-tests    Also install test dependencies"
                echo "  --help, -h      Show this help message"
                exit 0
                ;;
        esac
    done

    cd "$PROJECT_ROOT"

    detect_python
    create_venv
    install_dependencies
    show_instructions
}

main "$@"
