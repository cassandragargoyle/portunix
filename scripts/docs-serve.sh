#!/bin/bash
# Serve Portunix documentation locally
# Uses Hugo development server or Python fallback

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCS_SITE="$PROJECT_ROOT/docs-site"
PUBLIC_DIR="$DOCS_SITE/public"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║           Portunix Documentation Server                        ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Check for --static flag (serve only public/ without Hugo)
if [ "$1" == "--static" ] || [ "$1" == "-s" ]; then
    if [ ! -d "$PUBLIC_DIR" ]; then
        print_error "Directory $PUBLIC_DIR does not exist"
        print_info "Run 'python3 scripts/post-release-docs.py --build-only' first"
        exit 1
    fi

    print_info "Serving static files from: $PUBLIC_DIR"
    print_info "Server URL: http://localhost:8080"
    echo ""

    cd "$PUBLIC_DIR"
    python3 -m http.server 8080
    exit 0
fi

# Try Hugo first
if command -v hugo &> /dev/null; then
    print_success "Hugo found"
    print_info "Starting Hugo development server..."
    print_info "Server URL: http://localhost:1313"
    echo ""

    cd "$DOCS_SITE"
    hugo server
else
    print_warning "Hugo not found, using Python HTTP server"

    if [ ! -d "$PUBLIC_DIR" ]; then
        print_error "Directory $PUBLIC_DIR does not exist"
        print_info "Install Hugo: portunix install hugo"
        print_info "Or run: python3 scripts/post-release-docs.py --build-only"
        exit 1
    fi

    print_info "Serving static files from: $PUBLIC_DIR"
    print_info "Server URL: http://localhost:8080"
    echo ""

    cd "$PUBLIC_DIR"
    python3 -m http.server 8080
fi
