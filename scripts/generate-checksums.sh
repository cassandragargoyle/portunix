#!/bin/bash

# Generate individual SHA256 checksum files for release binaries
# Usage: ./generate-checksums.sh [version]

VERSION=${1:-v1.4.0}
DIST_DIR="dist"

echo "Generating checksums for version $VERSION..."

# Create dist directory if it doesn't exist
mkdir -p $DIST_DIR

# Generate checksums for each platform
for OS in linux windows darwin; do
    for ARCH in amd64 arm64; do
        # Skip darwin/arm64 for now
        if [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
            continue
        fi
        
        BINARY_NAME="portunix-$VERSION-$OS-$ARCH"
        if [ "$OS" = "windows" ]; then
            BINARY_NAME="${BINARY_NAME}.exe"
        fi
        
        if [ -f "$DIST_DIR/$BINARY_NAME" ]; then
            echo "Generating checksum for $BINARY_NAME..."
            sha256sum "$DIST_DIR/$BINARY_NAME" | awk '{print $1}' > "$DIST_DIR/${BINARY_NAME}.sha256"
            echo "  ✓ ${BINARY_NAME}.sha256"
        fi
    done
done

echo "Checksums generated successfully!"