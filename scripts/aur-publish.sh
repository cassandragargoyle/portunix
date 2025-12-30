#!/bin/bash
#
# AUR Package Publication Script for Portunix
# Copies PKGBUILD and .SRCINFO from container and publishes to AUR
#
# Prerequisites:
# 1. Run aur-prepare.sh first to prepare the package
# 2. SSH key must be configured for AUR access
# 3. GitHub release must exist for the version
#

set -e

CONTAINER_NAME="portunix_aur"
AUR_WORKDIR="/aur-portunix"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
AUR_PACKAGE_DIR="$PROJECT_ROOT/aur-package"
AUR_REPO_DIR="$PROJECT_ROOT/aur-repo"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Usage information
usage() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Example: $0 v1.7.4"
    echo "         $0 1.7.4"
    echo ""
    echo "This script will:"
    echo "  1. Copy PKGBUILD and .SRCINFO from container"
    echo "  2. Clone/update AUR repository"
    echo "  3. Update PKGBUILD files"
    echo "  4. Update checksums (downloads from GitHub)"
    echo "  5. Generate new .SRCINFO"
    echo "  6. Show diff for review"
    echo "  7. Commit and push to AUR (with confirmation)"
    echo ""
    echo "Prerequisites:"
    echo "  - Run aur-prepare.sh first"
    echo "  - SSH access to AUR configured"
    echo "  - GitHub release must exist"
    exit 1
}

# Check if version parameter is provided
if [ -z "$1" ]; then
    error "Version parameter is required!"
    usage
fi

VERSION="$1"

# Add 'v' prefix if not present
if [[ ! "$VERSION" =~ ^v ]]; then
    VERSION="v${VERSION}"
fi

# Validate version format (vX.Y.Z)
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format: $VERSION"
    info "Expected format: vX.Y.Z (e.g., v1.7.4)"
    exit 1
fi

# Remove 'v' prefix for version number
VERSION_NUM="${VERSION#v}"

info "Publishing Portunix version: $VERSION to AUR"

# Check if portunix binary exists (for container management)
if [ ! -f "$PROJECT_ROOT/portunix" ]; then
    error "Portunix binary not found at $PROJECT_ROOT/portunix"
    info "Please build it first: make build"
    exit 1
fi

PORTUNIX="$PROJECT_ROOT/portunix"

# Check if container exists and is running
info "Checking container status..."
CONTAINER_EXISTS=$("$PORTUNIX" container list -a | grep -c "$CONTAINER_NAME" || true)

if [ "$CONTAINER_EXISTS" -eq 0 ]; then
    error "Container '$CONTAINER_NAME' not found!"
    info "Please run aur-prepare.sh first to prepare the package"
    exit 1
fi

CONTAINER_RUNNING=$("$PORTUNIX" container list | grep "$CONTAINER_NAME" | grep -c "Up" || true)

if [ "$CONTAINER_RUNNING" -eq 0 ]; then
    warning "Container is not running, starting it..."
    "$PORTUNIX" container start "$CONTAINER_NAME"
    sleep 2
fi

success "Container is running"

# Check if PKGBUILD exists in container
info "Verifying PKGBUILD exists in container..."
if ! "$PORTUNIX" container exec "$CONTAINER_NAME" test -f "$AUR_WORKDIR/PKGBUILD"; then
    error "PKGBUILD not found in container at $AUR_WORKDIR/PKGBUILD"
    info "Please run aur-prepare.sh first"
    exit 1
fi

success "PKGBUILD found in container"

# Step 1: Copy files from container
info "=== Step 1: Copying AUR files from container ==="
mkdir -p "$AUR_PACKAGE_DIR"

info "Copying PKGBUILD..."
"$PORTUNIX" container cp "$CONTAINER_NAME:$AUR_WORKDIR/PKGBUILD" "$AUR_PACKAGE_DIR/"

info "Copying .SRCINFO..."
"$PORTUNIX" container cp "$CONTAINER_NAME:$AUR_WORKDIR/.SRCINFO" "$AUR_PACKAGE_DIR/"

success "Files copied to: $AUR_PACKAGE_DIR"

# Display copied files
echo ""
info "Copied PKGBUILD preview:"
echo "----------------------------------------"
head -n 20 "$AUR_PACKAGE_DIR/PKGBUILD"
echo "..."
echo "----------------------------------------"
echo ""

# Step 2: Clone/Update AUR repository
info "=== Step 2: Setting up AUR repository ==="

if [ ! -d "$AUR_REPO_DIR" ]; then
    info "AUR repository not found locally, cloning..."

    # Check SSH access to AUR
    if ! ssh -T aur@aur.archlinux.org 2>&1 | grep -qE "(Hi|Welcome)"; then
        error "Cannot connect to AUR via SSH!"
        info "Please configure SSH access to AUR first:"
        info "  1. Generate SSH key: ssh-keygen -t ed25519"
        info "  2. Add to AUR account: https://aur.archlinux.org/account/"
        info "  3. Test connection: ssh -T aur@aur.archlinux.org"
        exit 1
    fi

    info "Cloning AUR repository..."
    git clone ssh://aur@aur.archlinux.org/portunix.git "$AUR_REPO_DIR"

    # AUR uses 'master' branch, ensure we're on it
    cd "$AUR_REPO_DIR"
    git checkout -B master 2>/dev/null || true

    success "AUR repository cloned"
else
    info "AUR repository exists, updating..."
    cd "$AUR_REPO_DIR"

    # Ensure we're on master branch (AUR requirement)
    CURRENT_BRANCH=$(git branch --show-current)
    if [ "$CURRENT_BRANCH" != "master" ]; then
        info "Switching to master branch (AUR uses master, not main)"
        git checkout -B master 2>/dev/null || git branch -m "$CURRENT_BRANCH" master
    fi

    # Check if repository has any commits
    if git rev-parse --verify HEAD >/dev/null 2>&1; then
        # Repository has commits, check for uncommitted changes
        if ! git diff-index --quiet HEAD -- 2>/dev/null; then
            warning "Uncommitted changes detected in AUR repository!"
            git status --short
            echo ""

            if [ -t 0 ]; then
                read -p "Do you want to continue? This will overwrite local changes. (y/N) " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    error "Aborted by user"
                    exit 1
                fi
            else
                error "Cannot continue in non-interactive mode with uncommitted changes"
                exit 1
            fi

            git reset --hard HEAD
        fi

        # Pull latest changes
        git pull
        success "AUR repository updated"
    else
        # Repository is empty (first time setup)
        info "AUR repository is empty (first publication)"
        # Check for any uncommitted files
        if [ -n "$(git status --porcelain)" ]; then
            warning "Uncommitted files detected in empty repository!"
            git status --short
            echo ""
            info "These will be overwritten with new PKGBUILD files"
        fi
    fi
fi

cd "$AUR_REPO_DIR"

# Step 3: Update PKGBUILD files
info "=== Step 3: Updating PKGBUILD files ==="

cp "$AUR_PACKAGE_DIR/PKGBUILD" .
cp "$AUR_PACKAGE_DIR/.SRCINFO" .

success "PKGBUILD files updated"

# Step 4: Update checksums
info "=== Step 4: Updating checksums from GitHub ==="

info "Calculating checksums using container..."

# Copy PKGBUILD to container
"$PORTUNIX" container exec "$CONTAINER_NAME" mkdir -p /tmp/aur-checksum
"$PORTUNIX" container cp "$AUR_REPO_DIR/PKGBUILD" "$CONTAINER_NAME:/tmp/aur-checksum/"

# Install pacman-contrib in container if needed
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "
    if ! command -v updpkgsums &> /dev/null; then
        pacman -S --noconfirm --needed pacman-contrib
    fi
"

# Change ownership to builder user
"$PORTUNIX" container exec "$CONTAINER_NAME" chown -R builder:builder /tmp/aur-checksum

# Run updpkgsums in container as builder user
info "Downloading tarball and calculating checksums in container..."
if "$PORTUNIX" container exec "$CONTAINER_NAME" su - builder -c "cd /tmp/aur-checksum && updpkgsums"; then
    # Copy updated PKGBUILD back
    "$PORTUNIX" container cp "$CONTAINER_NAME:/tmp/aur-checksum/PKGBUILD" "$AUR_REPO_DIR/"
    success "Checksums updated"
else
    error "Failed to update checksums!"
    info "Please check if GitHub release exists: https://github.com/cassandragargoyle/Portunix/releases/tag/${VERSION}"
    exit 1
fi

# Step 5: Generate new .SRCINFO
info "=== Step 5: Generating .SRCINFO ==="

info "Generating .SRCINFO using container..."

# Generate .SRCINFO in container as builder user
if "$PORTUNIX" container exec "$CONTAINER_NAME" su - builder -c "cd /tmp/aur-checksum && makepkg --printsrcinfo > .SRCINFO"; then
    # Copy .SRCINFO back
    "$PORTUNIX" container cp "$CONTAINER_NAME:/tmp/aur-checksum/.SRCINFO" "$AUR_REPO_DIR/"
    success ".SRCINFO generated"
else
    error "Failed to generate .SRCINFO!"
    exit 1
fi

# Step 6: Show diff for review
info "=== Step 6: Review changes ==="
echo ""
info "Git status:"
git status
echo ""
info "Changes in PKGBUILD:"
git diff PKGBUILD || true
echo ""

# Step 7: Commit and push (with confirmation)
info "=== Step 7: Commit and push to AUR ==="
echo ""

if [ -t 0 ]; then
    read -p "Do you want to commit and push these changes to AUR? (y/N) " -n 1 -r
    echo
    PUSH_RESPONSE="$REPLY"
else
    warning "Non-interactive mode, skipping push"
    PUSH_RESPONSE="n"
fi

if [[ $PUSH_RESPONSE =~ ^[Yy]$ ]]; then
    info "Committing changes..."

    git add PKGBUILD .SRCINFO
    git commit -m "Update to v${VERSION_NUM}"

    success "Changes committed"

    echo ""
    read -p "Push to AUR now? (y/N) " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "Pushing to AUR..."
        git push origin master

        success "Successfully published to AUR!"
        echo ""
        info "Package URL: https://aur.archlinux.org/packages/portunix"
        info "Users can now install with: yay -S portunix"
    else
        warning "Push cancelled"
        info "To push later, run from $AUR_REPO_DIR:"
        info "  git push"
    fi
else
    warning "Commit cancelled"
    info "Files are ready in: $AUR_REPO_DIR"
    info "To commit and push manually:"
    info "  cd $AUR_REPO_DIR"
    info "  git add PKGBUILD .SRCINFO"
    info "  git commit -m 'Update to v${VERSION_NUM}'"
    info "  git push"
fi

echo ""
success "=== AUR publication process complete ==="
info "Local files:"
info "  - Package files: $AUR_PACKAGE_DIR"
info "  - AUR repository: $AUR_REPO_DIR"
info "  - Container: $CONTAINER_NAME"
echo ""
info "Container management:"
info "  Stop:   $PORTUNIX container stop $CONTAINER_NAME"
info "  Remove: $PORTUNIX container rm $CONTAINER_NAME"
echo ""
