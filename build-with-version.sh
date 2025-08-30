#!/bin/bash

# Build script with version embedding and portunix.rc update
VERSION=${1:-v1.5.2}

# Remove 'v' prefix if present for version numbers
VERSION_NUM=${VERSION#v}

echo "Building Portunix $VERSION..."

# Update portunix.rc file with new version
if [ -f "portunix.rc" ]; then
    echo "Updating portunix.rc with version $VERSION_NUM..."
    
    # Convert version to Windows format (e.g., 1.4.0 -> 1,4,0,0)
    WIN_VERSION=$(echo $VERSION_NUM | sed 's/\./,/g'),0
    
    # Update FILEVERSION
    sed -i "s/^FILEVERSION .*/FILEVERSION $WIN_VERSION/" portunix.rc
    
    # Update PRODUCTVERSION
    sed -i "s/^PRODUCTVERSION .*/PRODUCTVERSION $WIN_VERSION/" portunix.rc
    
    # Update FileVersion string
    sed -i "s/VALUE \"FileVersion\", \".*\"/VALUE \"FileVersion\", \"$VERSION_NUM\"/" portunix.rc
    
    # Update ProductVersion string
    sed -i "s/VALUE \"ProductVersion\", \".*\"/VALUE \"ProductVersion\", \"$VERSION_NUM\"/" portunix.rc
    
    echo "portunix.rc updated successfully"
fi

# Build with ldflags to set version (use full version with 'v' prefix)
go build -ldflags "-X main.version=$VERSION -s -w" -o portunix .

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Version check:"
    ./portunix version
else
    echo "Build failed!"
    exit 1
fi