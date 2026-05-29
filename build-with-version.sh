#!/bin/bash

# Build script with version embedding and portunix.rc update for all binaries
#
# Usage: build-with-version.sh [VERSION] [CONTEXT]
#   VERSION  - semantic version (default: v2.2.3)
#   CONTEXT  - "github" (clean version only), "internal" (+dev.N required),
#              or "auto" (default — accepts both, warns on clean version
#              outside known GitHub release context)
#
# Version format rules (ADR-036):
#   - GitHub stable:   vX.Y.Z              (e.g. v1.9.2)
#   - Internal dev:    vX.Y.Z+dev.N        (e.g. v1.9.2+dev.3)
#   - Pre-release:     vX.Y.Z-rc.N         (e.g. v1.10.0-rc.1)
#   - Snapshot (test): vX.Y.Z-SNAPSHOT     (legacy test builds)
VERSION=${1:-v2.4.0+dev.1}
CONTEXT=${2:-${PORTUNIX_RELEASE_CONTEXT:-auto}}

# Version validation (ADR-036)
RE_STABLE='^v[0-9]+\.[0-9]+\.[0-9]+$'
RE_DEV='^v[0-9]+\.[0-9]+\.[0-9]+\+dev\.[0-9]+$'
RE_PRERELEASE='^v[0-9]+\.[0-9]+\.[0-9]+-(rc|alpha|beta)\.[0-9]+$'
RE_SNAPSHOT='^v[0-9]+\.[0-9]+\.[0-9]+-SNAPSHOT$'

if [[ "$VERSION" =~ $RE_STABLE ]]; then
    VERSION_KIND="stable"
elif [[ "$VERSION" =~ $RE_DEV ]]; then
    VERSION_KIND="dev"
elif [[ "$VERSION" =~ $RE_PRERELEASE ]]; then
    VERSION_KIND="prerelease"
elif [[ "$VERSION" =~ $RE_SNAPSHOT ]]; then
    VERSION_KIND="snapshot"
else
    echo "ERROR: Invalid version format: $VERSION"
    echo "       Expected one of:"
    echo "         vX.Y.Z          (GitHub stable)"
    echo "         vX.Y.Z+dev.N    (internal development)"
    echo "         vX.Y.Z-rc.N     (release candidate; alpha/beta also allowed)"
    echo "         vX.Y.Z-SNAPSHOT (test build)"
    exit 1
fi

case "$CONTEXT" in
    github)
        if [[ "$VERSION_KIND" != "stable" && "$VERSION_KIND" != "prerelease" ]]; then
            echo "ERROR: Context 'github' requires clean stable or pre-release version."
            echo "       Got '$VERSION' (kind: $VERSION_KIND)."
            echo "       Use 'internal' context for +dev.N versions."
            exit 1
        fi
        ;;
    internal)
        if [[ "$VERSION_KIND" != "dev" ]]; then
            echo "ERROR: Context 'internal' requires +dev.N suffix (ADR-036)."
            echo "       Got '$VERSION' (kind: $VERSION_KIND)."
            echo "       Use 'github' context for clean release versions."
            exit 1
        fi
        ;;
    auto)
        if [[ "$VERSION_KIND" == "stable" ]]; then
            echo "WARNING: Building clean stable version '$VERSION' in 'auto' context."
            echo "         According to ADR-036, stable versions are reserved for"
            echo "         GitHub releases. Pass 'github' as 2nd arg to silence,"
            echo "         or use vX.Y.Z+dev.N for internal builds."
        fi
        ;;
    *)
        echo "ERROR: Unknown context '$CONTEXT' (expected: github | internal | auto)"
        exit 1
        ;;
esac

# Remove 'v' prefix if present for version numbers
VERSION_NUM=${VERSION#v}

# Strip build metadata (+dev.N) and pre-release (-rc.N, -SNAPSHOT) for the
# Windows FILEVERSION/PRODUCTVERSION numeric tuples (must be plain X.Y.Z).
# The string fields below keep the full VERSION_NUM with suffix.
VERSION_NUMERIC=$(echo "$VERSION_NUM" | sed -E 's/[-+].*$//')

# Detect Windows and set binary extension
EXT=""
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]] || [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
    EXT=".exe"
fi

echo "Building Portunix $VERSION (including helper binaries)..."

# Read update-signing public key (issue #165). Empty value disables update
# signature verification (development builds). Override via env var if needed.
UPDATE_PUBKEY="${PORTUNIX_UPDATE_PUBKEY:-}"
if [ -z "$UPDATE_PUBKEY" ] && [ -f "assets/update-signing-pubkey.txt" ]; then
    UPDATE_PUBKEY=$(tr -d '[:space:]' < assets/update-signing-pubkey.txt)
fi
if [ -n "$UPDATE_PUBKEY" ]; then
    echo "Embedding update-signing pubkey (${#UPDATE_PUBKEY} hex chars)"
fi
LDFLAGS_COMMON="-X main.version=$VERSION -X portunix.ai/app/update.Version=$VERSION -X portunix.ai/app/update.PublicKeyHex=$UPDATE_PUBKEY -s -w"

# Update portunix.rc file with new version
if [ -f "portunix.rc" ]; then
    echo "Updating portunix.rc with version $VERSION_NUM..."

    # Convert version to Windows format (e.g., 1.5.0 -> 1,5,0,0)
    # Use VERSION_NUMERIC — Windows FILEVERSION requires plain X.Y.Z integers,
    # so +dev.N / -rc.N suffixes must be stripped from the numeric tuple.
    WIN_VERSION=$(echo $VERSION_NUMERIC | sed 's/\./,/g'),0

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

# Update versioninfo.json (source for portunix.syso) and regenerate the .syso.
# portunix.syso is what go build actually embeds into portunix.exe (version
# metadata, application manifest, icon). Without this step, Windows file
# properties stay frozen at whatever .syso was last committed.
if [ -f "versioninfo.json" ]; then
    echo "Updating versioninfo.json with version $VERSION_NUM..."

    # Split semver into parts for FixedFileInfo (handles X.Y.Z, ignoring any
    # +dev.N / -rc.N suffix — FixedFileInfo only accepts integer fields).
    IFS='.' read -r V_MAJOR V_MINOR V_PATCH <<EOF
$VERSION_NUMERIC
EOF
    V_MAJOR=${V_MAJOR:-0}
    V_MINOR=${V_MINOR:-0}
    V_PATCH=${V_PATCH:-0}

    # Update FixedFileInfo numeric fields. Targets one occurrence of each
    # block by anchoring on the surrounding "FileVersion"/"ProductVersion"
    # heading and bumping Major/Minor/Patch lines that follow.
    python3 - <<PYEOF
import json, pathlib
p = pathlib.Path("versioninfo.json")
data = json.loads(p.read_text())
for k in ("FileVersion", "ProductVersion"):
    block = data["FixedFileInfo"].get(k, {})
    block["Major"] = $V_MAJOR
    block["Minor"] = $V_MINOR
    block["Patch"] = $V_PATCH
    block.setdefault("Build", 0)
    data["FixedFileInfo"][k] = block
data["StringFileInfo"]["FileVersion"] = "$VERSION_NUM"
data["StringFileInfo"]["ProductVersion"] = "$VERSION_NUM"
p.write_text(json.dumps(data, indent=4) + "\n")
PYEOF

    echo "versioninfo.json updated successfully"
fi

# Regenerate portunix.syso so the new version + manifest land in the binary.
GOVERSIONINFO_BIN="$(go env GOPATH)/bin/goversioninfo${EXT}"
if [ -x "$GOVERSIONINFO_BIN" ]; then
    echo "Regenerating portunix.syso..."
    "$GOVERSIONINFO_BIN" -o portunix.syso versioninfo.json
    if [ $? -ne 0 ]; then
        echo "portunix.syso generation failed!"
        exit 1
    fi
else
    echo "WARNING: goversioninfo not found at $GOVERSIONINFO_BIN"
    echo "         portunix.exe will be built with the existing (possibly stale) portunix.syso."
    echo "         Run: go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest"
fi

# Build main binary with ldflags to set version
echo "Building main binary: portunix..."
go build -ldflags "$LDFLAGS_COMMON" -o portunix${EXT} .

if [ $? -ne 0 ]; then
    echo "Main binary build failed!"
    exit 1
fi

# Build helper binaries with the same version
echo "Building helper binaries..."

# Build ptx-container
echo "Building ptx-container..."
cd src/helpers/ptx-container
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-container${EXT} .
CONTAINER_BUILD=$?
cd ../../..

# Build ptx-mcp
echo "Building ptx-mcp..."
cd src/helpers/ptx-mcp
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-mcp${EXT} .
MCP_BUILD=$?
cd ../../..

# Build ptx-virt
echo "Building ptx-virt..."
cd src/helpers/ptx-virt
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-virt${EXT} .
VIRT_BUILD=$?
cd ../../..

# Build ptx-ansible
echo "Building ptx-ansible..."
cd src/helpers/ptx-ansible
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-ansible${EXT} .
ANSIBLE_BUILD=$?
cd ../../..

# Build ptx-prompting
echo "Building ptx-prompting..."
cd src/helpers/ptx-prompting
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-prompting${EXT} .
PROMPTING_BUILD=$?
cd ../../..

# Build ptx-aiops
echo "Building ptx-aiops..."
cd src/helpers/ptx-aiops
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-aiops${EXT} .
AIOPS_BUILD=$?
cd ../../..

# Build ptx-make
echo "Building ptx-make..."
cd src/helpers/ptx-make
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-make${EXT} .
MAKE_BUILD=$?
cd ../../..

# Build ptx-pft
echo "Building ptx-pft..."
cd src/helpers/ptx-pft
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-pft${EXT} .
PFT_BUILD=$?
cd ../../..

# Build ptx-trace
echo "Building ptx-trace..."
cd src/helpers/ptx-trace
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-trace${EXT} .
TRACE_BUILD=$?
cd ../../..

# Build ptx-installer
echo "Building ptx-installer..."
cd src/helpers/ptx-installer
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-installer${EXT} .
INSTALLER_BUILD=$?
cd ../../..

# Build ptx-ssh
echo "Building ptx-ssh..."
cd src/helpers/ptx-ssh
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-ssh${EXT} .
SSH_BUILD=$?
cd ../../..

# Build ptx-plugin-registry
echo "Building ptx-plugin-registry..."
cd src/helpers/ptx-plugin-registry
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-plugin-registry${EXT} .
PLUGIN_REGISTRY_BUILD=$?
cd ../../..

# Build ptx-proxmox (issue #167)
echo "Building ptx-proxmox..."
cd src/helpers/ptx-proxmox
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-proxmox${EXT} .
PROXMOX_BUILD=$?
cd ../../..

# Build ptx-specpm (issue #183)
echo "Building ptx-specpm..."
cd src/helpers/ptx-specpm
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-specpm${EXT} .
SPECPM_BUILD=$?
cd ../../..

# Build ptx-database (issue #013)
echo "Building ptx-database..."
cd src/helpers/ptx-database
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-database${EXT} .
DATABASE_BUILD=$?
cd ../../..

# Build ptx-github (issue #025)
echo "Building ptx-github..."
cd src/helpers/ptx-github
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-github${EXT} .
GITHUB_BUILD=$?
cd ../../..

# Build ptx-wizard (issue #014)
echo "Building ptx-wizard..."
cd src/helpers/ptx-wizard
go build -ldflags "$LDFLAGS_COMMON" -o ../../../ptx-wizard${EXT} .
WIZARD_BUILD=$?
cd ../../..

# Check all builds
if [ $CONTAINER_BUILD -ne 0 ] || [ $MCP_BUILD -ne 0 ] || [ $VIRT_BUILD -ne 0 ] || [ $ANSIBLE_BUILD -ne 0 ] || [ $PROMPTING_BUILD -ne 0 ] || [ $AIOPS_BUILD -ne 0 ] || [ $MAKE_BUILD -ne 0 ] || [ $PFT_BUILD -ne 0 ] || [ $TRACE_BUILD -ne 0 ] || [ $INSTALLER_BUILD -ne 0 ] || [ $SSH_BUILD -ne 0 ] || [ $PLUGIN_REGISTRY_BUILD -ne 0 ] || [ $PROXMOX_BUILD -ne 0 ] || [ $SPECPM_BUILD -ne 0 ] || [ $DATABASE_BUILD -ne 0 ] || [ $GITHUB_BUILD -ne 0 ] || [ $WIZARD_BUILD -ne 0 ]; then
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
./ptx-ssh${EXT} --version
./ptx-plugin-registry${EXT} --version
./ptx-proxmox${EXT} --version
./ptx-specpm${EXT} --version
./ptx-database${EXT} --version
./ptx-github${EXT} --version
./ptx-wizard${EXT} --version