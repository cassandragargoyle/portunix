#!/bin/bash
#
# AUR Package Preparation Script for Portunix
# Creates and manages Arch Linux container for AUR package development
# Compiles from source inside the container
#

set -e

CONTAINER_NAME="portunix_aur"
CONTAINER_IMAGE="archlinux:latest"
AUR_WORKDIR="/aur-portunix"
SOURCE_DIR="/portunix-src"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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
    echo "  1. Create/start Arch Linux container"
    echo "  2. Download source code from GitHub (tag must exist!)"
    echo "  3. Compile Portunix with specified version"
    echo "  4. Generate PKGBUILD and .SRCINFO"
    echo "  5. Test build the AUR package"
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

info "Building Portunix version: $VERSION"

# Check if portunix binary exists (for container management)
if [ ! -f "$PROJECT_ROOT/portunix" ]; then
    error "Portunix binary not found at $PROJECT_ROOT/portunix"
    info "Please build it first: make build"
    exit 1
fi

PORTUNIX="$PROJECT_ROOT/portunix"

# Check if container exists
info "Checking if container '$CONTAINER_NAME' exists..."
CONTAINER_EXISTS=$("$PORTUNIX" container list | grep -c "$CONTAINER_NAME" || true)

if [ "$CONTAINER_EXISTS" -eq 0 ]; then
    info "Container doesn't exist, creating new one..."

    # Create container with detached mode
    "$PORTUNIX" container run -d --name "$CONTAINER_NAME" "$CONTAINER_IMAGE" sleep infinity
    success "Container '$CONTAINER_NAME' created"
else
    info "Container '$CONTAINER_NAME' already exists"

    # Check if container is running
    CONTAINER_RUNNING=$("$PORTUNIX" container list | grep "$CONTAINER_NAME" | grep -c "Up" || true)

    if [ "$CONTAINER_RUNNING" -eq 0 ]; then
        info "Starting stopped container..."
        "$PORTUNIX" container start "$CONTAINER_NAME"
        success "Container started"
    else
        success "Container is already running"
    fi
fi

# Wait a moment for container to be fully ready
sleep 2

# Install required packages in container (including Go)
info "Installing build dependencies in container..."
"$PORTUNIX" container exec "$CONTAINER_NAME" pacman -Sy --noconfirm --needed base-devel git go

success "Development tools installed (base-devel, git, go)"

# Create working directories
info "Creating working directory structure..."
"$PORTUNIX" container exec "$CONTAINER_NAME" mkdir -p "$AUR_WORKDIR"
"$PORTUNIX" container exec "$CONTAINER_NAME" mkdir -p "$SOURCE_DIR"

success "Directory structure created"

# Download source code from GitHub
info "Downloading source code from GitHub..."
VERSION_NUM="${VERSION#v}"
GITHUB_URL="https://github.com/cassandragargoyle/Portunix/archive/refs/tags/${VERSION}.tar.gz"

info "Checking if GitHub tag ${VERSION} exists..."
# Check if tag exists on GitHub (follow redirects with -L)
HTTP_STATUS=$("$PORTUNIX" container exec "$CONTAINER_NAME" curl -I -L -s -o /dev/null -w "%{http_code}" "$GITHUB_URL")

if [ "$HTTP_STATUS" != "200" ]; then
    error "GitHub tag ${VERSION} not found!"
    error "URL: $GITHUB_URL"
    error "HTTP Status: $HTTP_STATUS"
    echo ""
    info "GitHub tag ${VERSION} does not exist yet!"
    echo ""
    info "Option 1: Create tag manually (quick)"
    info "  git tag -a ${VERSION} -m 'Release ${VERSION}'"
    info "  git push origin ${VERSION}"
    echo ""
    info "Option 2: Use make-release.sh + manual tag push"
    info "  ./scripts/make-release.sh ${VERSION}  # Builds binaries"
    info "  git tag -a ${VERSION} -m 'Release ${VERSION}'"
    info "  git push origin ${VERSION}"
    info "  # Then create GitHub release and upload dist/ files"
    echo ""
    info "Option 3: Create GitHub release (creates tag automatically)"
    info "  1. Go to: https://github.com/cassandragargoyle/Portunix/releases/new"
    info "  2. Create new tag: ${VERSION}"
    info "  3. Add release notes"
    info "  4. Publish release"
    echo ""
    info "After creating tag, run this script again"
    exit 1
fi

success "GitHub tag ${VERSION} found (HTTP $HTTP_STATUS)"

info "Downloading source tarball from GitHub..."
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "
    cd /tmp &&
    curl -L -o portunix-${VERSION_NUM}.tar.gz '$GITHUB_URL' &&
    tar -xzf portunix-${VERSION_NUM}.tar.gz &&
    rm -rf $SOURCE_DIR &&
    mv portunix-${VERSION_NUM} $SOURCE_DIR
"

if [ $? -ne 0 ]; then
    error "Failed to download or extract source from GitHub!"
    exit 1
fi

success "Source code downloaded from GitHub"

# Make build script executable
"$PORTUNIX" container exec "$CONTAINER_NAME" chmod +x "$SOURCE_DIR/build-with-version.sh"

# Compile Portunix inside container
info "Compiling Portunix $VERSION inside container..."
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "cd $SOURCE_DIR && ./build-with-version.sh $VERSION"

if [ $? -ne 0 ]; then
    error "Compilation failed!"
    exit 1
fi

success "Compilation successful"

# Verify binary
info "Verifying compiled binary..."
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "$SOURCE_DIR/portunix version"

success "Binary verified"

# Get pacman version hash (for AUR compatibility check)
info "Getting pacman version hash..."
PACMAN_HASH=$("$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "LC_ALL=C pacman -V|sed -r 's#[0-9]+#ad9#g'|md5sum|cut -c1-6")
info "Pacman hash: $PACMAN_HASH"

# Create PKGBUILD file
info "Creating PKGBUILD for source-based build..."

# Remove 'v' prefix for pkgver
VERSION_NUM="${VERSION#v}"

PKGBUILD_CONTENT="# Maintainer: CassandraGargoyle <cassandragargoyle@gmail.com>
pkgname=portunix
pkgver=${VERSION_NUM}
pkgrel=1
pkgdesc='Portunix CLI – intelligent developer environment automation toolkit'
arch=('x86_64')
url='https://github.com/cassandragargoyle/Portunix'
license=('MIT')
depends=()
makedepends=('go' 'git')
provides=('portunix')
conflicts=('portunix-bin')
source=(\"portunix-\${pkgver}.tar.gz::https://github.com/cassandragargoyle/Portunix/archive/refs/tags/v\${pkgver}.tar.gz\")
sha256sums=('SKIP')

build() {
  cd \"portunix-\${pkgver}\"

  # Build using the project's build script
  chmod +x build-with-version.sh
  ./build-with-version.sh \"v\${pkgver}\"
}

package() {
  cd \"portunix-\${pkgver}\"

  # Install main binary
  install -Dm755 \"portunix\" \"\$pkgdir/usr/bin/portunix\"

  # Install helper binaries
  install -Dm755 \"ptx-container\" \"\$pkgdir/usr/bin/ptx-container\"
  install -Dm755 \"ptx-mcp\" \"\$pkgdir/usr/bin/ptx-mcp\"
  install -Dm755 \"ptx-virt\" \"\$pkgdir/usr/bin/ptx-virt\"
  install -Dm755 \"ptx-ansible\" \"\$pkgdir/usr/bin/ptx-ansible\"
  install -Dm755 \"ptx-prompting\" \"\$pkgdir/usr/bin/ptx-prompting\"

  # Install shell completions
  \$pkgdir/usr/bin/portunix completion bash > portunix.bash
  \$pkgdir/usr/bin/portunix completion zsh > _portunix
  \$pkgdir/usr/bin/portunix completion fish > portunix.fish

  install -Dm644 \"portunix.bash\" \"\$pkgdir/usr/share/bash-completion/completions/portunix\"
  install -Dm644 \"_portunix\" \"\$pkgdir/usr/share/zsh/site-functions/_portunix\"
  install -Dm644 \"portunix.fish\" \"\$pkgdir/usr/share/fish/vendor_completions.d/portunix.fish\"

  # Install documentation
  install -Dm644 \"README.md\" \"\$pkgdir/usr/share/doc/portunix/README.md\"

  # Install license
  install -Dm644 \"LICENSE\" \"\$pkgdir/usr/share/licenses/portunix/LICENSE\"
}
"

# Write PKGBUILD to container using heredoc
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "cat > $AUR_WORKDIR/PKGBUILD << 'PKGBUILD_EOF'
$PKGBUILD_CONTENT
PKGBUILD_EOF
"

success "PKGBUILD created"

# Create a non-root user for makepkg operations (makepkg refuses to run as root)
info "Setting up build user for makepkg..."
"$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "
    id builder &>/dev/null || useradd -m -G wheel builder
    echo 'builder ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/builder
    chown -R builder:builder $AUR_WORKDIR
"

# Create .SRCINFO (required for AUR) - must run as non-root user
info "Generating .SRCINFO..."
"$PORTUNIX" container exec "$CONTAINER_NAME" su - builder -c "cd $AUR_WORKDIR && makepkg --printsrcinfo > .SRCINFO"

success ".SRCINFO generated"

# Display PKGBUILD
echo ""
info "=== Generated PKGBUILD ==="
"$PORTUNIX" container exec "$CONTAINER_NAME" cat "$AUR_WORKDIR/PKGBUILD"
echo ""

# Display .SRCINFO
info "=== Generated .SRCINFO ==="
"$PORTUNIX" container exec "$CONTAINER_NAME" cat "$AUR_WORKDIR/.SRCINFO"
echo ""

# Test build (optional)
echo ""
# Check if we're in interactive mode
if [ -t 0 ]; then
    read -p "Do you want to test build the AUR package now? (y/N) " -n 1 -r
    echo
    TEST_BUILD_RESPONSE="$REPLY"
else
    info "Non-interactive mode detected, skipping test build"
    TEST_BUILD_RESPONSE="n"
fi

if [[ $TEST_BUILD_RESPONSE =~ ^[Yy]$ ]]; then
    info "Testing AUR package build from GitHub source..."
    warning "This will download source from GitHub (tag v${VERSION_NUM} must exist!)"

    # Run makepkg as builder user
    if "$PORTUNIX" container exec "$CONTAINER_NAME" su - builder -c "cd $AUR_WORKDIR && makepkg -f"; then
        success "AUR package built successfully!"

        # List generated packages
        info "Generated packages:"
        "$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "ls -lh $AUR_WORKDIR/*.pkg.tar.zst 2>/dev/null || echo '  No packages found'"

        echo ""
        if [ -t 0 ]; then
            read -p "Do you want to test install the package? (y/N) " -n 1 -r
            echo
            TEST_INSTALL_RESPONSE="$REPLY"
        else
            TEST_INSTALL_RESPONSE="n"
        fi

        if [[ $TEST_INSTALL_RESPONSE =~ ^[Yy]$ ]]; then
            info "Installing package..."
            "$PORTUNIX" container exec "$CONTAINER_NAME" bash -c "cd $AUR_WORKDIR && pacman -U --noconfirm *.pkg.tar.zst"

            # Test installed binary
            info "Testing installed binary..."
            "$PORTUNIX" container exec "$CONTAINER_NAME" portunix version

            success "Package installation test successful!"
        fi
    else
        warning "Package build failed. Check the output above for errors."
        warning "Make sure GitHub tag v${VERSION_NUM} exists!"
        info "Create release: https://github.com/cassandragargoyle/Portunix/releases/new"
    fi
fi

echo ""
success "=== AUR preparation complete ==="
info "Container: $CONTAINER_NAME"
info "AUR working directory: $AUR_WORKDIR"
info "Source directory: $SOURCE_DIR (downloaded from GitHub)"
info "Version: $VERSION"
info "Pacman hash: $PACMAN_HASH"
echo ""
info "Generated files in container:"
echo "  - $AUR_WORKDIR/PKGBUILD (AUR package definition)"
echo "  - $AUR_WORKDIR/.SRCINFO (AUR metadata)"
echo "  - $SOURCE_DIR/portunix (compiled binary with correct version)"
echo ""
info "✅ Source downloaded from GitHub: $GITHUB_URL"
echo ""
info "=== Next steps for AUR publication ==="
echo ""
echo "1️⃣  GitHub release is already verified (tag v${VERSION_NUM} exists)"
echo ""
echo "2️⃣  Copy AUR files from container:"
echo "   mkdir -p aur-package"
echo "   $PORTUNIX container cp $CONTAINER_NAME:$AUR_WORKDIR/PKGBUILD aur-package/"
echo "   $PORTUNIX container cp $CONTAINER_NAME:$AUR_WORKDIR/.SRCINFO aur-package/"
echo ""
echo "3️⃣  Setup AUR repository (if first time):"
echo "   git clone ssh://aur@aur.archlinux.org/portunix.git"
echo "   cd portunix"
echo ""
echo "4️⃣  Update AUR files:"
echo "   cp ../aur-package/PKGBUILD ."
echo "   cp ../aur-package/.SRCINFO ."
echo ""
echo "5️⃣  Update checksum:"
echo "   updpkgsums  # Downloads tarball and updates sha256sums"
echo "   makepkg --printsrcinfo > .SRCINFO"
echo ""
echo "6️⃣  Commit and push to AUR:"
echo "   git add PKGBUILD .SRCINFO"
echo "   git commit -m 'Update to v${VERSION_NUM}'"
echo "   git push"
echo ""
info "=== Container management ==="
echo "  Enter container:  $PORTUNIX container exec -it $CONTAINER_NAME bash"
echo "  Stop container:   $PORTUNIX container stop $CONTAINER_NAME"
echo "  Remove container: $PORTUNIX container rm $CONTAINER_NAME"
echo ""
success "AUR package ready for publication!"
