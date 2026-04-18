#!/bin/bash

# Build script with version embedding and portunix.rc update for all binaries
VERSION=${1:-v2.2.3}

# Remove 'v' prefix if present for version numbers
VERSION_NUM=${VERSION#v}

# Detect Windows and set binary extension
EXT=""
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]] || [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
    EXT=".exe"
fi

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
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o portunix${EXT} .

if [ $? -ne 0 ]; then
    echo "Main binary build failed!"
    exit 1
fi

# Build helper binaries with the same version
echo "Building helper binaries..."

# Build ptx-container
echo "Building ptx-container..."
cd src/helpers/ptx-container
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-container${EXT} .
CONTAINER_BUILD=$?
cd ../../..

# Build ptx-mcp
echo "Building ptx-mcp..."
cd src/helpers/ptx-mcp
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-mcp${EXT} .
MCP_BUILD=$?
cd ../../..

# Build ptx-virt
echo "Building ptx-virt..."
cd src/helpers/ptx-virt
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-virt${EXT} .
VIRT_BUILD=$?
cd ../../..

# Build ptx-ansible
echo "Building ptx-ansible..."
cd src/helpers/ptx-ansible
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-ansible${EXT} .
ANSIBLE_BUILD=$?
cd ../../..

# Build ptx-prompting
echo "Building ptx-prompting..."
cd src/helpers/ptx-prompting
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-prompting${EXT} .
PROMPTING_BUILD=$?
cd ../../..

# Build ptx-aiops
echo "Building ptx-aiops..."
cd src/helpers/ptx-aiops
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-aiops${EXT} .
AIOPS_BUILD=$?
cd ../../..

# Build ptx-make
echo "Building ptx-make..."
cd src/helpers/ptx-make
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-make${EXT} .
MAKE_BUILD=$?
cd ../../..

# Build ptx-pft
echo "Building ptx-pft..."
cd src/helpers/ptx-pft
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-pft${EXT} .
PFT_BUILD=$?
cd ../../..

# Build ptx-trace
echo "Building ptx-trace..."
cd src/helpers/ptx-trace
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-trace${EXT} .
TRACE_BUILD=$?
cd ../../..

# Build ptx-installer
echo "Building ptx-installer..."
cd src/helpers/ptx-installer
go build -ldflags "-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -s -w" -o ../../../ptx-installer${EXT} .
INSTALLER_BUILD=$?
cd ../../..

# Check all builds
if [ $CONTAINER_BUILD -ne 0 ] || [ $MCP_BUILD -ne 0 ] || [ $VIRT_BUILD -ne 0 ] || [ $ANSIBLE_BUILD -ne 0 ] || [ $PROMPTING_BUILD -ne 0 ] || [ $AIOPS_BUILD -ne 0 ] || [ $MAKE_BUILD -ne 0 ] || [ $PFT_BUILD -ne 0 ] || [ $TRACE_BUILD -ne 0 ] || [ $INSTALLER_BUILD -ne 0 ]; then
    echo "Helper binary build failed!"
    exit 1
fi

echo "All builds successful!"
echo "Version checks:"
echo "Main binary:"
./portunix${EXT} version
echo "Helper binaries:"
./ptx-container${EXT} --version
./ptx-mcp${EXT} --version
./ptx-virt${EXT} --version
./ptx-ansible${EXT} --version
./ptx-prompting${EXT} --version
./ptx-aiops${EXT} --version
./ptx-make${EXT} --version
./ptx-pft${EXT} --version
./ptx-trace${EXT} --version
./ptx-installer${EXT} --version