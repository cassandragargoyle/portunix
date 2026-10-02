#!/bin/bash
#
sudo apt update
if ! command -v go >/dev/null 2>&1; then
    sudo apt install -y golang
fi

if ! dpkg -s python3.12-venv >/dev/null 2>&1; then
    sudo apt install -y python3.12-venv
fi

PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VENV_DIR="$PROJECT_ROOT/.venv"

# Determine activation script based on OS
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    ACTIVATE_SCRIPT="$VENV_DIR/Scripts/activate"
else
    ACTIVATE_SCRIPT="$VENV_DIR/bin/activate"
fi

if [[ ! -f "$ACTIVATE_SCRIPT" ]]; then
    ./scripts/setup-venv.sh
fi
source .venv/bin/activate