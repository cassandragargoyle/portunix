#!/bin/bash

# Make Release Script for Portunix
# Usage: ./scripts/make-release.sh v1.5.1
# Uses existing GoReleaser configuration and creates everything needed for GitHub release

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

VERSION=${1}

print_header() {
    echo -e "${BLUE}╔══════════════════════════════════════════╗"
    echo -e "║        🚀 PORTUNIX RELEASE MAKER         ║"
    echo -e "║     One-command release preparation      ║"
    echo -e "╚══════════════════════════════════════════╝${NC}"
    echo
}

print_step() {
    echo -e "${GREEN}📋 $1${NC}"
    echo
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_usage() {
    echo "Usage: $0 <version>"
    echo
    echo "Examples:"
    echo "  $0 v1.5.1"
    echo "  $0 v1.6.0"
    echo
    echo "This script will:"
    echo "  1. Validate version format"
    echo "  2. Update version in source files"
    echo "  3. Build cross-platform binaries using GoReleaser"
    echo "  4. Create packages with install scripts"
    echo "  5. Generate checksums"
    echo "  6. Prepare everything for GitHub release"
}

validate_version() {
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format. Use semantic versioning: v1.2.3"
        show_usage
        exit 1
    fi
}

check_dependencies() {
    print_step "Checking dependencies..."

    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check for goreleaser in common locations
    GORELEASER_CMD=""
    if command -v goreleaser >/dev/null 2>&1; then
        GORELEASER_CMD="goreleaser"
    elif [ -x "$(go env GOPATH)/bin/goreleaser" ]; then
        GORELEASER_CMD="$(go env GOPATH)/bin/goreleaser"
    elif [ -x "$HOME/go/bin/goreleaser" ]; then
        GORELEASER_CMD="$HOME/go/bin/goreleaser"
    else
        print_error "GoReleaser is not installed"
        echo "Install with: go install github.com/goreleaser/goreleaser@latest"
        exit 1
    fi
    
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    if [ ! -f ".goreleaser.yml" ]; then
        print_error ".goreleaser.yml not found"
        exit 1
    fi
    
    print_info "✓ Go: $(go version | cut -d' ' -f3)"
    print_info "✓ GoReleaser: $($GORELEASER_CMD --version | head -n1)"
    print_info "✓ Git repository detected"
    print_info "✓ GoReleaser config found"
    echo
}

update_version_files() {
    print_step "Updating version in source files..."
    
    # Update build-with-version.sh default version
    if [ -f "build-with-version.sh" ]; then
        sed -i "s/^VERSION=\${1:-v[0-9]\+\.[0-9]\+\.[0-9]\+}/VERSION=\${1:-$VERSION}/" build-with-version.sh
        print_info "✓ Updated build-with-version.sh default version"
    fi
    
    # Update portunix.rc using the build script
    print_info "Updating portunix.rc with version $VERSION..."
    ./build-with-version.sh "$VERSION" >/dev/null 2>&1 || {
        print_warning "Version update in build script had issues, but continuing..."
    }
    
    echo
}

run_goreleaser() {
    print_step "Running GoReleaser to create cross-platform release..."
    
    print_info "Cleaning previous builds..."
    rm -rf dist/
    
    print_info "Building release packages..."
    
    # Create temporary git tag for proper versioning
    print_info "Creating temporary git tag $VERSION for build..."
    git tag -d "$VERSION" 2>/dev/null || true  # Remove if exists
    git tag "$VERSION"
    
    # Run GoReleaser with the tag for proper version embedding
    if $GORELEASER_CMD release --clean --skip-validate --skip-publish; then
        print_info "✓ GoReleaser completed successfully"
        
        # Remove temporary tag after successful build
        print_info "Removing temporary git tag..."
        git tag -d "$VERSION"
    else
        # Remove temporary tag on failure too
        git tag -d "$VERSION" 2>/dev/null || true
        print_error "GoReleaser failed"
        exit 1
    fi
    
    echo
}

verify_outputs() {
    print_step "Verifying generated files..."
    
    if [ ! -d "dist" ]; then
        print_error "dist/ directory not found"
        exit 1
    fi
    
    # Count generated files
    archives=$(find dist/ -name "*.tar.gz" -o -name "*.zip" | wc -l)
    checksums=$(find dist/ -name "*checksums*" | wc -l)
    
    print_info "Generated files:"
    echo "   Archives: $archives"
    echo "   Checksum files: $checksums"
    
    if [ "$archives" -eq 0 ]; then
        print_error "No archive files generated"
        exit 1
    fi
    
    # Show generated files
    echo
    echo "📦 Generated release files:"
    ls -la dist/*.{tar.gz,zip,txt} 2>/dev/null | while read -r line; do
        echo "   $line"
    done || print_warning "Some file types not found"
    
    echo
}

create_release_notes() {
    print_step "Creating release notes..."
    
    cat > "dist/RELEASE_NOTES_${VERSION}.md" << EOF
# Portunix ${VERSION}

Universal development environment management tool.

## 🎉 What's New

This release includes the latest improvements and bug fixes for Portunix.

## ✨ Key Features

### 🔧 Development Infrastructure
- Modern linting configuration with CI/CD compatibility
- Dynamic version management with build-time injection
- Enhanced GitHub Actions CI/CD pipeline
- Cross-platform testing across Linux, Windows, and macOS

### 📦 Package Management
- Universal installer with cross-platform package installation
- Pre-configured software: Java, Python, Go, VS Code, PowerShell, and more
- Installation profiles: default, minimal, full, and empty
- Smart package detection with automatic package manager optimization

### 🐳 Container Management  
- Docker integration with intelligent installation and management
- SSH-enabled containers for development
- Multi-platform support: Ubuntu, Alpine, CentOS, Debian
- Cache optimization with efficient directory mounting

### 🔌 Plugin System
- gRPC-based architecture for high-performance communication
- Dynamic plugin loading and management
- Protocol Buffer support for structured API definitions

## 📋 Installation

Choose the appropriate package for your platform:

### Linux
\`\`\`bash
# AMD64
wget https://github.com/cassandragargoyle/Portunix/releases/download/${VERSION}/portunix_${VERSION#v}_linux_amd64.tar.gz
tar -xzf portunix_${VERSION#v}_linux_amd64.tar.gz
cd portunix_${VERSION#v}_linux_amd64
./install.sh
\`\`\`

### Windows
\`\`\`powershell
# Download and extract
# https://github.com/cassandragargoyle/Portunix/releases/download/${VERSION}/portunix_${VERSION#v}_windows_amd64.zip
# Then run:
.\\install.ps1
\`\`\`

### macOS
\`\`\`bash
# Intel Macs
wget https://github.com/cassandragargoyle/Portunix/releases/download/${VERSION}/portunix_${VERSION#v}_darwin_amd64.tar.gz
tar -xzf portunix_${VERSION#v}_darwin_amd64.tar.gz
cd portunix_${VERSION#v}_darwin_amd64  
./install.sh

# Apple Silicon (M1/M2)
wget https://github.com/cassandragargoyle/Portunix/releases/download/${VERSION}/portunix_${VERSION#v}_darwin_arm64.tar.gz
tar -xzf portunix_${VERSION#v}_darwin_arm64.tar.gz
cd portunix_${VERSION#v}_darwin_arm64
./install.sh
\`\`\`

## 🚀 Quick Start

\`\`\`bash
# Install development environment
portunix install default

# Manage containers
portunix container run ubuntu
portunix container ssh container-name

# Configure MCP server
portunix mcp configure
\`\`\`

## 🔗 Links

- **Repository**: https://github.com/cassandragargoyle/Portunix
- **Issues**: https://github.com/cassandragargoyle/Portunix/issues
- **Documentation**: Repository docs/ directory

## 🔐 Verification

Verify downloads using SHA256 checksums provided with the release.

---

**Build Information:**
- Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
- Go Version: $(go version | cut -d' ' -f3)
- Git Commit: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

EOF

    print_info "✓ Release notes created: dist/RELEASE_NOTES_${VERSION}.md"
    echo
}

show_summary() {
    print_step "🎉 Release preparation complete!"
    
    echo -e "${CYAN}📊 Summary:${NC}"
    echo -e "   Version: ${GREEN}$VERSION${NC}"
    echo -e "   Build directory: ${BLUE}dist/${NC}"
    echo
    
    echo "📦 Generated files:"
    find dist/ -type f | sort | while read -r file; do
        size=$(du -h "$file" | cut -f1)
        echo "   $(basename "$file") ($size)"
    done
    echo
    
    echo -e "${GREEN}🚀 Next Steps:${NC}"
    echo "   1. Review files in dist/ directory"
    echo "   2. Test installation on different platforms" 
    echo "   3. Verify version in generated binaries:"
    echo "      ./dist/portunix_*/portunix version"
    echo "   4. Create GitHub release:"
    echo "      - Tag: $VERSION"
    echo "      - Title: Portunix $VERSION"
    echo "      - Description: Use dist/RELEASE_NOTES_${VERSION}.md"
    echo "      - Upload all files from dist/"
    echo
    
    echo -e "${CYAN}📋 Files ready for upload:${NC}"
    ls -la dist/*.{tar.gz,zip,txt,md} 2>/dev/null | sed 's/^/   /'
    echo
}

# Main execution
main() {
    print_header
    
    # Validate input
    if [ -z "$VERSION" ]; then
        print_error "Version parameter is required"
        show_usage
        exit 1
    fi
    
    validate_version
    
    print_info "🎯 Creating release for version: $VERSION"
    echo
    
    check_dependencies
    update_version_files
    run_goreleaser
    verify_outputs
    create_release_notes
    show_summary
    
    echo -e "${GREEN}✅ Release $VERSION ready for publication!${NC}"

    # Generate documentation site (unless disabled)
    if [ "${AUTO_DOCS:-true}" != "false" ]; then
        echo
        print_step "Generating documentation site..."
        if [ -x "./scripts/post-release-docs.py" ]; then
            python3 ./scripts/post-release-docs.py "$VERSION" --build-only || {
                print_warning "Documentation generation failed (non-blocking)"
                echo "You can manually regenerate documentation with:"
                echo "  python3 ./scripts/post-release-docs.py $VERSION"
            }
        else
            print_warning "post-release-docs.py not found or not executable"
            echo "Install Hugo and run:"
            echo "  python3 ./scripts/post-release-docs.py $VERSION"
        fi
    fi
}

# Run main function
main "$@"