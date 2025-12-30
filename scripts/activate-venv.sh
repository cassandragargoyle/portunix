#!/bin/bash
# Activate Python Virtual Environment for Portunix
# Usage: source scripts/activate-venv.sh
#
# This script must be sourced, not executed!

# Detect script location
if [ -n "${BASH_SOURCE[0]}" ]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
    # Fallback for other shells
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
fi

PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VENV_DIR="$PROJECT_ROOT/.venv"

# Determine activation script based on OS
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    ACTIVATE_SCRIPT="$VENV_DIR/Scripts/activate"
else
    ACTIVATE_SCRIPT="$VENV_DIR/bin/activate"
fi

# Check if venv exists
if [ ! -f "$ACTIVATE_SCRIPT" ]; then
    echo "Virtual environment not found at $VENV_DIR"
    echo "Run './scripts/setup-venv.sh' first to create it."
    return 1 2>/dev/null || exit 1
fi

# Activate
# shellcheck disable=SC1090
source "$ACTIVATE_SCRIPT"

echo "Virtual environment activated: $VENV_DIR"
echo "Python: $(python --version 2>&1)"
echo "To deactivate, run: deactivate"
