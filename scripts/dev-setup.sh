#!/bin/bash
# Portunix Development Environment Setup
# Installs all required development dependencies using portunix
#
# Usage: ./scripts/dev-setup.sh [--dry-run]
#
# Prerequisites: portunix binary must be available in PATH or current directory

set -euo pipefail

DRY_RUN=""
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN="--dry-run"
    echo "=== DRY RUN MODE ==="
fi

# Find portunix binary
if command -v portunix >/dev/null 2>&1; then
    PTX="portunix"
elif [[ -f "./portunix" ]]; then
    PTX="./portunix"
else
    echo "ERROR: portunix binary not found in PATH or current directory"
    echo "Install portunix first: https://github.com/cassandragargoyle/portunix"
    exit 1
fi

echo "============================================"
echo "  Portunix Development Environment Setup"
echo "============================================"
echo ""
echo "Using: $($PTX --version 2>/dev/null || echo "$PTX")"
echo ""

# Required packages for building and developing portunix
PACKAGES=(
    "go"          # Go compiler (primary language)
    "python"      # Python (release scripts, deploy scripts)
    "gh"          # GitHub CLI (release publishing)
    # "make"      # GNU Make (build automation) - not yet in registry
)

# Optional packages (install failures are non-fatal)
OPTIONAL_PACKAGES=(
    "hugo"        # Documentation site generator
)

echo "--- Installing required packages ---"
echo ""

FAILED=()
for pkg in "${PACKAGES[@]}"; do
    echo ">>> Installing: $pkg"
    if ! $PTX install $pkg $DRY_RUN; then
        FAILED+=("$pkg")
        echo "WARNING: Failed to install $pkg"
    fi
    echo ""
done

echo "--- Installing optional packages ---"
echo ""

for pkg in "${OPTIONAL_PACKAGES[@]}"; do
    echo ">>> Installing (optional): $pkg"
    if ! $PTX install $pkg $DRY_RUN; then
        echo "NOTE: Optional package $pkg failed to install (non-fatal)"
    fi
    echo ""
done

# Setup Python virtual environment if python is available
if command -v python3 >/dev/null 2>&1 || command -v python >/dev/null 2>&1; then
    if [[ ! -d ".venv" ]] && [[ -z "$DRY_RUN" ]]; then
        echo "--- Setting up Python virtual environment ---"
        if [[ -f "./scripts/setup-venv.sh" ]]; then
            ./scripts/setup-venv.sh
        else
            python3 -m venv .venv 2>/dev/null || python -m venv .venv
        fi
        echo ""
    fi
fi

# Install Node.js (provides npm) via portunix
echo "--- Installing Node.js for npm-based tools ---"
echo ">>> Installing: nodejs"
if ! $PTX install nodejs $DRY_RUN; then
    echo "WARNING: Failed to install nodejs"
fi
echo ""

# Install markdownlint-cli2 via npm
if command -v npm >/dev/null 2>&1; then
    echo "--- Installing markdownlint-cli2 ---"
    if [[ -z "$DRY_RUN" ]]; then
        npm install -g markdownlint-cli2 || echo "WARNING: Failed to install markdownlint-cli2"
    else
        echo "[dry-run] Would run: npm install -g markdownlint-cli2"
    fi
    echo ""
else
    echo "NOTE: npm not found after nodejs install, skipping markdownlint-cli2"
    echo "      Restart your terminal and re-run this script"
    echo ""
fi

echo "============================================"
echo "  Setup Summary"
echo "============================================"

if [[ ${#FAILED[@]} -gt 0 ]]; then
    echo ""
    echo "FAILED packages: ${FAILED[*]}"
    echo "Please install these manually or check error messages above."
    exit 1
else
    echo ""
    echo "All required packages installed successfully."
    echo ""
    echo "Next steps:"
    echo "  1. Restart your terminal (to pick up PATH changes)"
    echo "  2. Run: make build"
    echo ""
fi
