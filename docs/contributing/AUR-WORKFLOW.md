# AUR Publication Workflow for Portunix

This document describes the complete workflow for publishing Portunix to the Arch User Repository (AUR).

## Prerequisites

1. **SSH Access to AUR**
   ```bash
   # Generate SSH key (if you don't have one)
   ssh-keygen -t ed25519 -C "your_email@example.com"

   # Add public key to AUR account
   # https://aur.archlinux.org/account/

   # Test connection
   ssh -T aur@aur.archlinux.org
   # Should return: "Hi username, you've successfully authenticated..."
   ```

2. **GitHub Release Created**
   - Git tag must exist on GitHub
   - Release tarball must be downloadable
   - URL format: `https://github.com/cassandragargoyle/Portunix/archive/refs/tags/v1.7.5.tar.gz`

3. **Portunix Binary Built**
   ```bash
   make build
   ```

## Complete Workflow

### Step 1: Create GitHub Release

**Option A: Quick manual tag (recommended for testing)**
```bash
# Create and push tag
git tag -a v1.7.5 -m "Release v1.7.5"
git push origin v1.7.5

# Then create release on GitHub with this existing tag
```

**Option B: Use make-release.sh (for full release)**
```bash
# Build cross-platform binaries
./scripts/make-release.sh v1.7.5

# Script creates dist/ with all binaries but doesn't push tag!
# You must manually create and push the tag:
git tag -a v1.7.5 -m "Release v1.7.5"
git push origin v1.7.5

# Then create GitHub release and upload files from dist/
```

**Option C: GitHub web interface**
1. Go to: https://github.com/cassandragargoyle/Portunix/releases/new
2. Create new tag: `v1.7.5`
3. Add release title: "Release v1.7.5"
4. Add release notes (or use dist/RELEASE_NOTES_v1.7.5.md)
5. Upload binaries from dist/ (if using make-release.sh)
6. Publish release

### Step 2: Prepare AUR Package

```bash
# This script:
# - Creates/reuses Arch Linux container
# - Downloads source from GitHub (verifies tag exists!)
# - Compiles Portunix from downloaded source
# - Generates PKGBUILD and .SRCINFO
# - Optionally tests build

./scripts/aur-prepare.sh v1.7.5
```

**What happens:**
1. ✅ Checks if GitHub tag v1.7.5 exists (HTTP 200)
2. ✅ Downloads tarball: `Portunix/archive/refs/tags/v1.7.5.tar.gz`
3. ✅ Compiles inside Arch Linux container
4. ✅ Creates PKGBUILD with correct version
5. ✅ Generates .SRCINFO metadata
6. ⚠️ Asks if you want to test build (recommended!)

**Output:**
- Container: `portunix_aur` (reusable)
- Files inside container:
  - `/aur-portunix/PKGBUILD`
  - `/aur-portunix/.SRCINFO`
  - `/portunix-src/portunix` (compiled binary)

### Step 3: Publish to AUR

```bash
# This script:
# - Copies PKGBUILD and .SRCINFO from container
# - Clones/updates AUR repository
# - Updates checksums (IN CONTAINER - no host pollution!)
# - Commits and pushes to AUR (with confirmation)

./scripts/aur-publish.sh v1.7.5
```

**What happens:**
1. ✅ Copies files from container to `aur-package/`
2. ✅ Clones AUR repo to `aur-repo/` (or updates if exists)
3. ✅ Copies PKGBUILD files to AUR repo
4. ✅ Runs `updpkgsums` **in Arch container** - downloads tarball and calculates sha256
5. ✅ Regenerates .SRCINFO **in container**
6. ✅ Copies updated files back to host
7. ✅ Shows git diff for review
8. ⚠️ Asks for commit confirmation
9. ⚠️ Asks for push confirmation

**Important:** Script uses Arch Linux container for `updpkgsums` and `makepkg`, so it works on **any Linux distro** (Debian, Ubuntu, etc.)!

**Output:**
- `aur-package/` - Extracted files from container
- `aur-repo/` - AUR git repository (pushed to AUR)

### Step 4: Test Installation from AUR

```bash
# This script:
# - Creates FRESH Arch Linux container
# - Installs yay AUR helper
# - Installs portunix from AUR
# - Verifies installation works

./scripts/aur-test-install.sh

# Or keep container for debugging
./scripts/aur-test-install.sh --keep-container
```

**What happens:**
1. ✅ Creates clean `portunix_aur_test` container
2. ✅ Installs base-devel, git, go
3. ✅ Installs yay from AUR
4. ✅ Runs `yay -S portunix` (downloads from AUR, compiles, installs)
5. ✅ Tests `portunix version`, `portunix --help`
6. ✅ Shows package info with `pacman -Qi portunix`
7. ✅ Cleans up (or keeps with --keep-container)

## Quick Reference

### Full workflow (one version, all steps)
```bash
# 1. Create and push tag
git tag -a v1.7.5 -m "Release v1.7.5"
git push origin v1.7.5

# 2. Prepare AUR package (downloads from GitHub, compiles, creates PKGBUILD)
./scripts/aur-prepare.sh v1.7.5

# 3. Publish to AUR (updates checksums, commits, pushes)
./scripts/aur-publish.sh v1.7.5

# 4. Test installation (fresh container, installs from AUR)
./scripts/aur-test-install.sh
```

### Container management
```bash
# List containers
./portunix container list -a

# Access container
./portunix container exec -it portunix_aur bash
./portunix container exec -it portunix_aur_test bash

# Remove containers
./portunix container rm -f portunix_aur
./portunix container rm -f portunix_aur_test

# Remove all test artifacts
rm -rf aur-package/ aur-repo/
```

## Troubleshooting

### "GitHub tag not found" (HTTP 404)
**Problem:** Tag doesn't exist on GitHub yet

**Solution:**
```bash
# Create and push tag
git tag -a v1.7.5 -m "Release v1.7.5"
git push origin v1.7.5

# Verify tag exists
curl -I -L https://github.com/cassandragargoyle/Portunix/archive/refs/tags/v1.7.5.tar.gz
# Should return: HTTP/2 200
```

### "Cannot connect to AUR via SSH"
**Problem:** SSH key not configured for AUR

**Solution:**
```bash
# Test connection - should see "Welcome" or "Hi"
ssh -T aur@aur.archlinux.org
# Expected: "Welcome to AUR, username!"

# If fails, add SSH key:
cat ~/.ssh/id_ed25519.pub
# Copy and paste to: https://aur.archlinux.org/account/
```

### "updpkgsums failed"
**Problem:** Can't download tarball from GitHub

**Solution:**
Script now runs `updpkgsums` in Arch container automatically, so this should work on any distro!

If it still fails:
```bash
# Check if GitHub release exists
curl -I -L https://github.com/cassandragargoyle/Portunix/archive/refs/tags/v1.7.5.tar.gz
# Should return: HTTP/2 200

# If tag doesn't exist, create it first (see above)
```

### "yay -S portunix failed"
**Problem:** Package not yet on AUR or build failure

**Solution:**
```bash
# Check if package is published
yay -Ss portunix

# Check AUR page
https://aur.archlinux.org/packages/portunix

# Debug in test container
./scripts/aur-test-install.sh --keep-container
./portunix container exec -it portunix_aur_test bash
su - auruser
cd /tmp
git clone https://aur.archlinux.org/portunix.git
cd portunix
makepkg -f  # Manual build for debugging
```

## Important Notes

### Why aur-prepare.sh downloads from GitHub?
AUR packages MUST be buildable from public sources. The script ensures:
1. ✅ Source is downloaded from GitHub (public URL)
2. ✅ Same source that AUR users will download
3. ✅ Build process is reproducible
4. ✅ No dependency on local development files

### What's the difference between containers?
- **portunix_aur** - Build container for PKGBUILD preparation (reusable)
- **portunix_aur_test** - Fresh test container for installation verification (recreated each time)

### Why test after publishing?
To ensure users can actually install from AUR! The test:
1. Uses fresh Arch Linux environment (no cached files)
2. Downloads PKGBUILD from AUR
3. Downloads source from GitHub
4. Compiles and installs like a real user would

### Cross-platform support
**Works on any Linux distro!** Scripts use Arch Linux container for:
- ✅ Building packages (`aur-prepare.sh`)
- ✅ Calculating checksums (`aur-publish.sh` - runs `updpkgsums` in container)
- ✅ Generating .SRCINFO (`aur-publish.sh` - runs `makepkg` in container)
- ✅ Testing installation (`aur-test-install.sh`)

You can develop on Debian/Ubuntu/Fedora and publish to AUR without any Arch-specific tools on your host!

## References

- **AUR Documentation**: https://wiki.archlinux.org/title/AUR
- **PKGBUILD Guide**: https://wiki.archlinux.org/title/PKGBUILD
- **AUR Submission Guidelines**: https://wiki.archlinux.org/title/AUR_submission_guidelines
- **Portunix on AUR**: https://aur.archlinux.org/packages/portunix

---

**Last Updated:** 2025-10-12
**Portunix Version:** v1.7.5+
