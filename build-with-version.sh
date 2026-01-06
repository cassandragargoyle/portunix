#!/bin/bash

# Build script with version embedding and portunix.rc update for all binaries
VERSION=${1:-v1.9.1}

# Remove 'v' prefix if present for version numbers
VERSION_NUM=${VERSION#v}

echo "Building Portunix $VERSION (including helper binaries)..."

# Update portunix.rc file with new version
if [ -f "portunix.rc" ]; then
    echo "Updating portunix.rc with version $VERSION_NUM..."

    # Convert version to Windows format (e.g., 1.5.0 -> 1,5,0,0)
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

# Build main binary with ldflags to set version
echo "Building main binary: portunix..."
go build -ldflags "-X main.version=$VERSION -s -w" -o portunix .

if [ $? -ne 0 ]; then
    echo "Main binary build failed!"
    exit 1
fi

# Build helper binaries with the same version
echo "Building helper binaries..."

# Build ptx-container
echo "Building ptx-container..."
cd src/helpers/ptx-container
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-container .
CONTAINER_BUILD=$?
cd ../../..

# Build ptx-mcp
echo "Building ptx-mcp..."
cd src/helpers/ptx-mcp
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-mcp .
MCP_BUILD=$?
cd ../../..

# Build ptx-virt
echo "Building ptx-virt..."
cd src/helpers/ptx-virt
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-virt .
VIRT_BUILD=$?
cd ../../..

# Build ptx-ansible
echo "Building ptx-ansible..."
cd src/helpers/ptx-ansible
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-ansible .
ANSIBLE_BUILD=$?
cd ../../..

# Build ptx-prompting
echo "Building ptx-prompting..."
cd src/helpers/ptx-prompting
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-prompting .
PROMPTING_BUILD=$?
cd ../../..

# Build ptx-aiops
echo "Building ptx-aiops..."
cd src/helpers/ptx-aiops
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-aiops .
AIOPS_BUILD=$?
cd ../../..

# Build ptx-make
echo "Building ptx-make..."
cd src/helpers/ptx-make
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-make .
MAKE_BUILD=$?
cd ../../..

# Build ptx-pft
echo "Building ptx-pft..."
cd src/helpers/ptx-pft
go build -ldflags "-X main.version=$VERSION -s -w" -o ../../../ptx-pft .
PFT_BUILD=$?
cd ../../..

# Check all builds
if [ $CONTAINER_BUILD -ne 0 ] || [ $MCP_BUILD -ne 0 ] || [ $VIRT_BUILD -ne 0 ] || [ $ANSIBLE_BUILD -ne 0 ] || [ $PROMPTING_BUILD -ne 0 ] || [ $AIOPS_BUILD -ne 0 ] || [ $MAKE_BUILD -ne 0 ] || [ $PFT_BUILD -ne 0 ]; then
    echo "Helper binary build failed!"
    exit 1
fi

echo "All builds successful!"
echo "Version checks:"
echo "Main binary:"
./portunix version
echo "Helper binaries:"
./ptx-container --version
./ptx-mcp --version
./ptx-virt --version
./ptx-ansible --version
./ptx-prompting --version
./ptx-aiops --version
./ptx-make --version
./ptx-pft --version