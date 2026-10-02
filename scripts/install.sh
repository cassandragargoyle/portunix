#!/usr/bin/env bash

# Portunix Installation Script for Linux/macOS
# This script uses the built-in install-self command of Portunix

set -e

# Script version
SCRIPT_VERSION="1.0.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
SILENT=false
INSTALL_PATH=""
ADD_TO_PATH=false
CREATE_CONFIG=false

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Portunix Installation Script v${SCRIPT_VERSION}

Usage: $0 [options]

Options:
    --silent        Silent installation with defaults
    --path <path>   Custom installation path
    --add-to-path   Add to system PATH
    --create-config Create default configuration
    --help          Show this help message

Examples:
    # Interactive installation
    ./install.sh
    
    # Silent installation with defaults
    ./install.sh --silent
    
    # Custom installation
    ./install.sh --path /usr/local/bin/portunix --add-to-path

Default installation paths:
    Linux:   /usr/local/bin/portunix (with sudo)
             ~/bin/portunix (without sudo)
    macOS:   /usr/local/bin/portunix
EOF
}

# Function to find portunix binary
find_portunix() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Check in script directory
    if [ -f "$script_dir/portunix" ]; then
        echo "$script_dir/portunix"
        return 0
    fi

    # Check in parent directory
    local parent_dir="$(dirname "$script_dir")"
    if [ -f "$parent_dir/portunix" ]; then
        echo "$parent_dir/portunix"
        return 0
    fi

    # Check in current directory
    if [ -f "./portunix" ]; then
        echo "$(pwd)/portunix"
        return 0
    fi

    return 1
}

# Function to find platforms directory (ADR-031: Cross-Platform Binary Distribution)
find_platforms_dir() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Check in script directory
    if [ -d "$script_dir/platforms" ]; then
        echo "$script_dir/platforms"
        return 0
    fi

    # Check in parent directory
    local parent_dir="$(dirname "$script_dir")"
    if [ -d "$parent_dir/platforms" ]; then
        echo "$parent_dir/platforms"
        return 0
    fi

    # Check in current directory
    if [ -d "./platforms" ]; then
        echo "$(pwd)/platforms"
        return 0
    fi

    return 1
}

# Function to install platform archives (ADR-031: Cross-Platform Binary Distribution)
install_platform_archives() {
    local install_dir="$1"

    PLATFORMS_DIR=$(find_platforms_dir)
    if [ $? -ne 0 ]; then
        print_color "$BLUE" "No platforms directory found (cross-platform binaries not included)"
        return 0
    fi

    print_color "$BLUE" "Installing cross-platform binaries (ADR-031)..."

    # Create platforms directory in install location
    local dest_platforms="$install_dir/platforms"
    mkdir -p "$dest_platforms" 2>/dev/null || {
        print_color "$YELLOW" "Note: Could not create platforms directory (may need sudo)"
        return 1
    }

    # Copy platform archives
    local count=0
    for archive in "$PLATFORMS_DIR"/*.tar.gz "$PLATFORMS_DIR"/*.zip; do
        if [ -f "$archive" ]; then
            cp "$archive" "$dest_platforms/" && {
                count=$((count + 1))
                if [ "$SILENT" = false ]; then
                    print_color "$GREEN" "  ✓ Installed $(basename "$archive")"
                fi
            }
        fi
    done

    if [ $count -gt 0 ]; then
        print_color "$GREEN" "Installed $count platform archive(s) for cross-platform provisioning"
    else
        print_color "$BLUE" "No platform archives found to install"
    fi

    return 0
}

# Function to check if running with sudo
is_sudo() {
    [ "$EUID" -eq 0 ]
}

# Function to detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "macos";;
        *)          echo "unknown";;
    esac
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --silent)
            SILENT=true
            shift
            ;;
        --path)
            INSTALL_PATH="$2"
            shift 2
            ;;
        --add-to-path)
            ADD_TO_PATH=true
            shift
            ;;
        --create-config)
            CREATE_CONFIG=true
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            print_color "$RED" "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main installation logic
main() {
    echo ""
    print_color "$CYAN" "=========================================="
    print_color "$CYAN" "     Portunix Installation Script"
    print_color "$CYAN" "=========================================="
    echo ""
    
    # Detect OS
    OS=$(detect_os)
    print_color "$BLUE" "Detected OS: $OS"
    
    # Find portunix binary
    PORTUNIX_BIN=$(find_portunix)
    if [ $? -ne 0 ]; then
        print_color "$RED" "Error: portunix binary not found"
        print_color "$RED" "Please ensure portunix is in the same directory as this script"
        exit 1
    fi
    
    print_color "$GREEN" "Found portunix at: $PORTUNIX_BIN"
    
    # Make sure it's executable
    chmod +x "$PORTUNIX_BIN" 2>/dev/null || true
    
    # Get version
    VERSION=$("$PORTUNIX_BIN" --version 2>/dev/null || echo "unknown")
    print_color "$BLUE" "Version: $VERSION"
    echo ""
    
    # Build installation command
    INSTALL_CMD=("$PORTUNIX_BIN" "install-self")
    
    if [ "$SILENT" = true ]; then
        INSTALL_CMD+=("--silent")
        print_color "$BLUE" "Running silent installation..."
    else
        print_color "$BLUE" "Starting interactive installation..."
    fi
    
    if [ -n "$INSTALL_PATH" ]; then
        INSTALL_CMD+=("--path" "$INSTALL_PATH")
    fi
    
    if [ "$ADD_TO_PATH" = true ]; then
        INSTALL_CMD+=("--add-to-path")
    fi
    
    if [ "$CREATE_CONFIG" = true ]; then
        INSTALL_CMD+=("--create-config")
    fi
    
    # Check if we need sudo for default paths
    if [ "$SILENT" = false ] && [ -z "$INSTALL_PATH" ]; then
        if [ "$OS" = "linux" ] || [ "$OS" = "macos" ]; then
            if ! is_sudo && [ -w "/usr/local/bin" ]; then
                print_color "$YELLOW" "Note: /usr/local/bin is writable, sudo not required"
            elif ! is_sudo; then
                print_color "$YELLOW" "Note: Installation to /usr/local/bin requires sudo"
                echo ""
                read -p "Would you like to run with sudo? (y/N): " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    print_color "$BLUE" "Restarting with sudo..."
                    exec sudo "$0" "$@"
                fi
            fi
        fi
    fi
    
    # Run the installation
    echo ""
    if "${INSTALL_CMD[@]}"; then
        echo ""
        print_color "$GREEN" "Installation completed successfully!"

        # Determine install directory for platform archives
        # Default is /usr/local/portunix or ~/portunix based on permissions
        if [ -n "$INSTALL_PATH" ]; then
            INSTALL_DIR="$(dirname "$INSTALL_PATH")"
        elif is_sudo || [ -w "/usr/local" ]; then
            INSTALL_DIR="/usr/local/portunix"
        else
            INSTALL_DIR="$HOME/portunix"
        fi

        # ADR-031: Install cross-platform binary archives
        echo ""
        install_platform_archives "$INSTALL_DIR"

        # Additional instructions
        if [ "$ADD_TO_PATH" = true ]; then
            echo ""
            print_color "$YELLOW" "Note: You may need to restart your terminal or run:"
            print_color "$YELLOW" "  source ~/.bashrc  (or ~/.zshrc for Zsh)"
            print_color "$YELLOW" "for PATH changes to take effect."
        fi

        echo ""
        print_color "$BLUE" "You can verify the installation by running:"
        print_color "$YELLOW" "  portunix --version"
    else
        print_color "$RED" "Installation failed"
        exit 1
    fi
}

# Run main function
main "$@"