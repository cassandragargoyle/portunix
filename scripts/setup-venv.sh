#!/bin/bash

# Setup Python Virtual Environment for Portunix (uv-based)
# Usage: ./scripts/setup-venv.sh [--with-tests]
#
# Provisions ./.venv/ via `uv sync` (or `uv sync --group test` with --with-tests).
# See ADR-039 for the rationale behind adopting uv.

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

show_help() {
    cat <<EOF
Usage: $0 [--with-tests]

Options:
  --with-tests    Also install test dependencies (pytest, pytest-xdist, ...)
  --help, -h      Show this help message

This script uses uv (https://docs.astral.sh/uv/) to provision the Python
environment. If uv is not installed, run one of:

  portunix install uv                                        # Portunix package
  curl -LsSf https://astral.sh/uv/install.sh | sh            # Linux / macOS
  powershell -c "irm https://astral.sh/uv/install.ps1 | iex" # Windows
EOF
}

check_uv() {
    if ! command -v uv >/dev/null 2>&1; then
        print_error "uv is not installed or not on PATH"
        echo
        echo "Install uv via one of:"
        echo "  portunix install uv"
        echo "  curl -LsSf https://astral.sh/uv/install.sh | sh"
        echo
        echo "See https://docs.astral.sh/uv/ for details."
        exit 1
    fi
    print_success "Found uv: $(uv --version)"
}

sync_env() {
    cd "$PROJECT_ROOT"

    if [ "$INSTALL_TESTS" = "true" ]; then
        print_step "Provisioning .venv/ with test dependencies (uv sync --group test)..."
        uv sync --group test
        print_success "Environment ready with test dependencies"
    else
        print_step "Provisioning .venv/ (uv sync)..."
        uv sync
        print_success "Environment ready"
    fi
}

show_instructions() {
    echo
    print_step "Setup complete!"
    echo
    echo "Run commands without activation:"
    echo "  uv run pytest test/"
    echo "  uv run scripts/file-server.py --help"
    echo
    echo "Or activate the virtual environment:"
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

main() {
    echo "================================"
    echo "Portunix Python Environment Setup"
    echo "================================"
    echo

    INSTALL_TESTS="false"

    for arg in "$@"; do
        case $arg in
            --with-tests)
                INSTALL_TESTS="true"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
        esac
    done

    check_uv
    sync_env
    show_instructions
}

main "$@"
