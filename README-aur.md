# AUR Package Preparation for Portunix

This document describes how to prepare and publish Portunix to the Arch User Repository (AUR).

## Prerequisites

- Working Portunix binary (build with `make build`)
- Podman or Docker installed
- SSH access to AUR (configured with your AUR account)

## Quick Start

```bash
# Prepare AUR package for specific version
./scripts/aur-prepare.sh v1.7.4

# Or without 'v' prefix (script adds it automatically)
./scripts/aur-prepare.sh 1.7.4
```

## What the Script Does

1. **Creates/starts Arch Linux container** (`portunix_aur`)
2. **Installs build dependencies** (base-devel, git, go)
3. **Copies source code** to container
4. **Compiles Portunix** with specified version using `build-with-version.sh`
5. **Generates PKGBUILD** for source-based AUR package
6. **Generates .SRCINFO** (AUR metadata)
7. **Optionally tests** package build from GitHub
8. **Provides instructions** for AUR publication

## PKGBUILD Details

The generated PKGBUILD:

- Downloads source from GitHub releases (tag archive)
- Compiles using `build-with-version.sh` script
- Installs main binary + all helper binaries (ptx-*)
- Generates and installs shell completions (bash/zsh/fish)
- Installs documentation and license

### Dependencies

- **makedepends**: `go`, `git`
- **depends**: none (static binary)

## AUR Publication Workflow

### 1. Prepare Package

```bash
./scripts/aur-prepare.sh v1.7.4
```

### 2. Create GitHub Release

Before publishing to AUR, create a GitHub release:

```bash
# Using GitHub CLI
gh release create v1.7.4 --title "Release v1.7.4" --notes "Release notes here"

# Or manually at:
# https://github.com/cassandragargoyle/Portunix/releases/new
```

**Important**: AUR package downloads source from GitHub tag, so the release must exist!

### 3. Copy AUR Files

```bash
mkdir -p aur-package
./portunix container cp portunix_aur:/aur-portunix/PKGBUILD aur-package/
./portunix container cp portunix_aur:/aur-portunix/.SRCINFO aur-package/
```

### 4. Setup AUR Repository

First time only:

```bash
# Clone AUR repository (requires SSH access)
git clone ssh://aur@aur.archlinux.org/portunix.git
cd portunix
```

For updates:

```bash
cd portunix
git pull
```

### 5. Update Package Files

```bash
cp ../aur-package/PKGBUILD .
cp ../aur-package/.SRCINFO .
```

### 6. Update Checksums

After GitHub release is created:

```bash
# Option 1: Automatic (requires makepkg)
updpkgsums
makepkg --printsrcinfo > .SRCINFO

# Option 2: Manual
# Download source, calculate sha256, update PKGBUILD
wget https://github.com/cassandragargoyle/Portunix/archive/refs/tags/v1.7.4.tar.gz
sha256sum v1.7.4.tar.gz
# Update sha256sums=('...') in PKGBUILD
makepkg --printsrcinfo > .SRCINFO
```

### 7. Test Build (Optional)

```bash
makepkg -si
```

### 8. Commit and Push

```bash
git add PKGBUILD .SRCINFO
git commit -m "Update to v1.7.4"
git push
```

## Container Management

```bash
# Enter container for debugging
./portunix container exec -it portunix_aur bash

# View container logs
./portunix container logs portunix_aur

# Stop container (keeps data)
./portunix container stop portunix_aur

# Remove container (deletes data)
./portunix container rm portunix_aur
```

## Testing the Build

### Test Inside Container

```bash
# Enter container
./portunix container exec -it portunix_aur bash

# Inside container
cd /aur-portunix
su - builder
cd /aur-portunix
makepkg -si
```

### Test Installed Package

```bash
# After successful makepkg -si
portunix version
portunix --help
```

## Troubleshooting

### Build Fails with "Tag not found"

**Problem**: GitHub tag doesn't exist yet

**Solution**: Create GitHub release first:

```bash
gh release create v1.7.4
```

### Checksum Mismatch

**Problem**: sha256sums doesn't match downloaded file

**Solution**: Update checksum after release:

```bash
updpkgsums
makepkg --printsrcinfo > .SRCINFO
```

### makepkg Fails with "Running as root"

**Problem**: makepkg refuses to run as root

**Solution**: Use builder user (script creates it automatically):

```bash
su - builder
cd /aur-portunix
makepkg
```

### Go Module Download Fails

**Problem**: Network issues or missing dependencies

**Solution**: Ensure container has network access:

```bash
./portunix container exec portunix_aur ping -c 3 proxy.golang.org
```

## Files Generated

### In Container

- `/aur-portunix/PKGBUILD` - AUR package definition
- `/aur-portunix/.SRCINFO` - AUR metadata
- `/portunix-src/portunix` - Compiled binary (with correct version)
- `/portunix-src/ptx-*` - Helper binaries

### Locally (after copying)

- `aur-package/PKGBUILD`
- `aur-package/.SRCINFO`

## Version Management

The script:

1. Accepts version as parameter (e.g., `v1.7.4` or `1.7.4`)
2. Validates format (must be vX.Y.Z)
3. Uses `build-with-version.sh` for compilation
4. Embeds version in binary via ldflags
5. Updates portunix.rc (Windows resource file)

## AUR Package Naming

- **Package name**: `portunix`
- **Provides**: `portunix`
- **Conflicts**: `portunix-bin` (if binary package exists)

## License

The AUR package uses MIT license (same as Portunix).

## Maintainer

- **Maintainer**: CassandraGargoyle <info@cassandragargoyle.cz>
- **AUR Package**: https://aur.archlinux.org/packages/portunix
- **Upstream**: https://github.com/cassandragargoyle/Portunix

## Additional Resources

- [AUR Submission Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
- [PKGBUILD Documentation](https://wiki.archlinux.org/title/PKGBUILD)
- [makepkg Manual](https://man.archlinux.org/man/makepkg.8)
- [AUR Account Registration](https://aur.archlinux.org/register/)

## Quick Reference

```bash
# Full workflow for new version
./scripts/aur-prepare.sh v1.7.5          # Prepare in container
gh release create v1.7.5                 # Create GitHub release
./portunix container cp portunix_aur:/aur-portunix/PKGBUILD aur-package/
./portunix container cp portunix_aur:/aur-portunix/.SRCINFO aur-package/
cd portunix-aur-repo                     # Your local AUR git clone
cp ../aur-package/* .
updpkgsums && makepkg --printsrcinfo > .SRCINFO
git add PKGBUILD .SRCINFO
git commit -m "Update to v1.7.5"
git push
```

---

**Last Updated**: 2025-10-12
**Script Version**: 1.0
