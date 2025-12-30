#!/bin/bash
#
# AUR Installation Test Script for Portunix
# Creates fresh Arch Linux container and tests AUR package installation
#
# This script:
# 1. Creates/recreates clean Arch Linux container
# 2. Sets up AUR helper (yay)
# 3. Installs portunix from AUR
# 4. Verifies installation and basic functionality
#

set -e

CONTAINER_NAME="portunix_aur_test"
CONTAINER_IMAGE="archlinux:latest"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

step() {
    echo -e "${CYAN}▶️  $1${NC}"
}

# Usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --keep-container    Don't remove container after test"
    echo "  --no-verify         Skip functionality verification"
    echo "  --help             Show this help message"
    echo ""
    echo "This script will:"
    echo "  1. Remove existing test container (if exists)"
    echo "  2. Create fresh Arch Linux container"
    echo "  3. Install base development tools"
    echo "  4. Install yay AUR helper"
    echo "  5. Install portunix from AUR using yay"
    echo "  6. Verify portunix installation and functionality"
    echo "  7. Clean up (unless --keep-container specified)"
    echo ""
    echo "Examples:"
    echo "  $0                      # Full test with cleanup"
    echo "  $0 --keep-container     # Keep container for debugging"
    echo "  $0 --no-verify          # Skip verification tests"
    exit 1
}

# Parse arguments
KEEP_CONTAINER=false
NO_VERIFY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --keep-container)
            KEEP_CONTAINER=true
            shift
            ;;
        --no-verify)
            NO_VERIFY=true
            shift
            ;;
        --help)
            usage
            ;;
        *)
            error "Unknown option: $1"
            usage
            ;;
    esac
done

info "=== Portunix AUR Installation Test ==="
echo ""

# Check if portunix binary exists (for container management)
if [ ! -f "$PROJECT_ROOT/portunix" ]; then
    error "Portunix binary not found at $PROJECT_ROOT/portunix"
    info "Please build it first: make build"
    exit 1
fi

PORTUNIX="$PROJECT_ROOT/portunix"

# Step 1: Clean up existing test container
step "Step 1: Cleaning up existing test container (if exists)..."

CONTAINER_EXISTS=$("$PORTUNIX" container list -a | grep -c "$CONTAINER_NAME" || true)

if [ "$CONTAINER_EXISTS" -gt 0 ]; then
    info "Removing existing container '$CONTAINER_NAME'..."

    # Stop if running
    "$PORTUNIX" container stop "$CONTAINER_NAME" 2>/dev/null || true

    # Remove container
    "$PORTUNIX" container rm -f "$CONTAINER_NAME" 2>/dev/null || true

    success "Old container removed"
else
    info "No existing container found"
fi

# Step 2: Create fresh container
step "Step 2: Creating fresh Arch Linux container..."

"$PORTUNIX" container run -d --name "$CONTAINER_NAME" "$CONTAINER_IMAGE" sleep infinity

success "Container '$CONTAINER_NAME' created"

# Wait for container to be ready
sleep 2

# Step 3: Install base development tools
step "Step 3: Installing base development tools..."

info "Updating package database..."
"$PORTUNIX" container exec "$CONTAINER_NAME" pacman -Sy --noconfirm

info "Installing base-devel and git..."
"$PORTUNIX" container exec "$CONTAINER_NAME" pacman -S --noconfirm --needed base-devel git

success "Base development tools installed"

# Step 4: Create non-root user for AUR operations
step "Step 4: Setting up non-root user for AUR..."

"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "
    # Create user if doesn't exist
    id auruser &>/dev/null || useradd -m -G wheel auruser

    # Allow passwordless sudo for wheel group
    echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel

    # Set proper permissions
    chmod 440 /etc/sudoers.d/wheel
"

success "User 'auruser' created with sudo access"

# Step 5: Install yay AUR helper
step "Step 5: Installing yay AUR helper..."

info "Cloning yay repository..."
"$PORTUNIX" container exec "$CONTAINER_NAME" su - auruser -c "
    cd ~
    git clone https://aur.archlinux.org/yay.git
    cd yay
    makepkg -si --noconfirm
"

if [ $? -ne 0 ]; then
    error "Failed to install yay!"
    exit 1
fi

success "yay AUR helper installed"

# Verify yay installation
info "Verifying yay installation..."
"$PORTUNIX" container exec "$CONTAINER_NAME" su - auruser -c "yay --version"

# Step 6: Install portunix from AUR
step "Step 6: Installing portunix from AUR..."

info "This will download and build portunix from AUR..."
warning "This may take a few minutes depending on your internet connection and CPU"

echo ""
info "Running: yay -S --noconfirm portunix"
echo ""

# Install with detailed output
if "$PORTUNIX" container exec "$CONTAINER_NAME" su - auruser -c "yay -S --noconfirm portunix"; then
    success "portunix installed successfully from AUR!"
else
    error "Failed to install portunix from AUR!"
    error "This could mean:"
    echo "  - Package not yet available on AUR"
    echo "  - Build failed due to missing dependencies"
    echo "  - Network issues downloading source"
    echo ""
    info "Container '$CONTAINER_NAME' kept for debugging"
    info "To investigate: $PORTUNIX container exec -it $CONTAINER_NAME bash"
    exit 1
fi

echo ""

# Step 7: Verify installation
if [ "$NO_VERIFY" = false ]; then
    step "Step 7: Verifying portunix installation..."

    # Check if binary exists
    info "Checking if portunix binary is installed..."
    if "$PORTUNIX" container exec "$CONTAINER_NAME" which portunix > /dev/null 2>&1; then
        success "portunix binary found in PATH"
    else
        error "portunix binary not found in PATH!"
        exit 1
    fi

    # Test version command
    info "Testing 'portunix version' command..."
    if VERSION_OUTPUT=$("$PORTUNIX" container exec "$CONTAINER_NAME" portunix version 2>&1); then
        echo "$VERSION_OUTPUT"

        if [[ "$VERSION_OUTPUT" == *"Portunix version"* ]] || [[ "$VERSION_OUTPUT" == *"v"* ]]; then
            success "Version command works correctly"
        else
            warning "Version output format unexpected but command succeeded"
        fi
    else
        error "Version command failed!"
        exit 1
    fi

    echo ""

    # Test help command
    info "Testing 'portunix --help' command..."
    if HELP_OUTPUT=$("$PORTUNIX" container exec "$CONTAINER_NAME" portunix --help 2>&1); then
        if [[ "$HELP_OUTPUT" == *"Portunix"* ]] || [[ "$HELP_OUTPUT" == *"Usage"* ]]; then
            success "Help command works correctly"
        else
            warning "Help output format unexpected but command succeeded"
        fi
    else
        error "Help command failed!"
        exit 1
    fi

    echo ""

    # Test install command dry-run
    info "Testing 'portunix install --help' command..."
    if "$PORTUNIX" container exec "$CONTAINER_NAME" portunix install --help > /dev/null 2>&1; then
        success "Install command is available"
    else
        warning "Install command test failed (may be expected)"
    fi

    echo ""

    # List installed files
    info "Listing installed files..."
    if "$PORTUNIX" container exec "$CONTAINER_NAME" pacman -Ql portunix 2>&1 | head -n 20; then
        success "Package files listed successfully"
    else
        error "Failed to list package files!"
        exit 1
    fi

    echo ""
    success "Installation verification completed"
else
    info "Skipping verification (--no-verify specified)"
fi

# Step 8: Show package information
echo ""
step "Package Information:"
if "$PORTUNIX" container exec "$CONTAINER_NAME" pacman -Qi portunix 2>&1; then
    success "Package information displayed"
else
    error "Failed to get package information!"
    exit 1
fi

# Step 9: Cleanup or keep container
echo ""
if [ "$KEEP_CONTAINER" = true ]; then
    warning "Container kept for debugging (--keep-container specified)"
    info "Container name: $CONTAINER_NAME"
    info "To access: $PORTUNIX container exec -it $CONTAINER_NAME bash"
    info "To remove: $PORTUNIX container rm -f $CONTAINER_NAME"
else
    step "Step 8: Cleaning up test container..."

    if [ -t 0 ]; then
        read -p "Remove test container? (Y/n) " -n 1 -r
        echo
        REMOVE_RESPONSE="${REPLY:-Y}"
    else
        info "Non-interactive mode, removing container"
        REMOVE_RESPONSE="Y"
    fi

    if [[ $REMOVE_RESPONSE =~ ^[Yy]$ ]]; then
        "$PORTUNIX" container stop "$CONTAINER_NAME"
        "$PORTUNIX" container rm "$CONTAINER_NAME"
        success "Test container removed"
    else
        info "Container kept: $CONTAINER_NAME"
        info "To remove later: $PORTUNIX container rm -f $CONTAINER_NAME"
    fi
fi

echo ""
success "=== AUR Installation Test Complete ==="
echo ""
info "Summary:"
echo "  ✅ Fresh Arch Linux container created"
echo "  ✅ yay AUR helper installed"
echo "  ✅ portunix installed from AUR"
if [ "$NO_VERIFY" = false ]; then
    echo "  ✅ Installation verified and tested"
fi
if [ "$KEEP_CONTAINER" = false ]; then
    echo "  ✅ Test container cleaned up"
else
    echo "  ⚠️  Test container kept for debugging"
fi
echo ""
success "Portunix is installable from AUR! 🎉"
