#!/usr/bin/env python3
"""
Make Release Script for Portunix
Part of ADR-031: Cross-Platform Binary Distribution Strategy
Issue #125: Cross-Platform Binary Distribution

Usage:
    python scripts/make-release.py v1.7.8

This script:
1. Validates version format
2. Updates version in source files
3. Builds cross-platform binaries using GoReleaser
4. Creates platform archives for cross-platform distribution
5. Injects platform archives into release packages
6. Generates checksums and release notes

The script automatically uses the project's .venv if available.
"""

import os
import re
import sys
import shutil
import subprocess
import tarfile
import zipfile
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional, List, Tuple

# ANSI color codes
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'

    @classmethod
    def disable(cls):
        cls.RED = cls.GREEN = cls.YELLOW = cls.BLUE = cls.CYAN = cls.NC = ''


def print_header() -> None:
    print(f"{Colors.BLUE}╔══════════════════════════════════════════╗")
    print(f"║        🚀 PORTUNIX RELEASE MAKER         ║")
    print(f"║     One-command release preparation      ║")
    print(f"╚══════════════════════════════════════════╝{Colors.NC}")
    print()


def print_step(message: str) -> None:
    print(f"{Colors.GREEN}📋 {message}{Colors.NC}")
    print()


def print_info(message: str) -> None:
    print(f"{Colors.CYAN}ℹ️  {message}{Colors.NC}")


def print_warning(message: str) -> None:
    print(f"{Colors.YELLOW}⚠️  {message}{Colors.NC}")


def print_error(message: str) -> None:
    print(f"{Colors.RED}❌ {message}{Colors.NC}")


def get_project_root() -> Path:
    """Get the project root directory"""
    script_dir = Path(__file__).parent.resolve()
    return script_dir.parent


def find_venv_python() -> Optional[Path]:
    """Find Python in project's virtual environment"""
    project_root = get_project_root()
    venv_dir = project_root / ".venv"

    if sys.platform == "win32":
        venv_python = venv_dir / "Scripts" / "python.exe"
    else:
        venv_python = venv_dir / "bin" / "python"

    if venv_python.exists():
        return venv_python
    return None


def run_command(cmd: List[str], cwd: Optional[Path] = None,
                capture: bool = False, check: bool = True) -> subprocess.CompletedProcess:
    """Run a command and optionally capture output"""
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=capture,
            text=True,
            check=check
        )
        return result
    except subprocess.CalledProcessError as e:
        if capture:
            print_error(f"Command failed: {' '.join(cmd)}")
            if e.stdout:
                print(e.stdout)
            if e.stderr:
                print(e.stderr)
        raise


def validate_version(version: str) -> bool:
    """Validate version format (vX.Y.Z or vX.Y.Z-SNAPSHOT)"""
    pattern = r'^v\d+\.\d+\.\d+(-SNAPSHOT)?$'
    return bool(re.match(pattern, version))


def check_dependencies() -> Tuple[bool, str]:
    """Check required dependencies and return GoReleaser command"""
    print_step("Checking dependencies...")

    # Check Go
    try:
        result = run_command(["go", "version"], capture=True)
        go_version = result.stdout.strip().split()[2] if result.stdout else "unknown"
        print_info(f"✓ Go: {go_version}")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print_error("Go is not installed or not in PATH")
        return False, ""

    # Check GoReleaser
    goreleaser_cmd = None
    for cmd in ["goreleaser", str(Path.home() / "go" / "bin" / "goreleaser")]:
        try:
            result = run_command([cmd, "--version"], capture=True, check=False)
            if result.returncode == 0:
                goreleaser_cmd = cmd
                version_line = result.stdout.strip().split('\n')[0] if result.stdout else "unknown"
                print_info(f"✓ GoReleaser: {version_line}")
                break
        except FileNotFoundError:
            continue

    if not goreleaser_cmd:
        print_error("GoReleaser is not installed")
        print_info("Install with: go install github.com/goreleaser/goreleaser@latest")
        return False, ""

    # Check git
    try:
        run_command(["git", "rev-parse", "--git-dir"], capture=True)
        print_info("✓ Git repository detected")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print_error("Not in a git repository")
        return False, ""

    # Check .goreleaser.yml
    project_root = get_project_root()
    if not (project_root / ".goreleaser.yml").exists():
        print_error(".goreleaser.yml not found")
        return False, ""
    print_info("✓ GoReleaser config found")

    print()
    return True, goreleaser_cmd


def update_version_files(version: str) -> None:
    """Update version in source files"""
    print_step("Updating version in source files...")
    project_root = get_project_root()

    # Update build-with-version.sh default version
    build_script = project_root / "build-with-version.sh"
    if build_script.exists():
        content = build_script.read_text()
        content = re.sub(
            r'^VERSION=\$\{1:-v[0-9]+\.[0-9]+\.[0-9]+\}',
            f'VERSION=${{1:-{version}}}',
            content,
            flags=re.MULTILINE
        )
        build_script.write_text(content)
        print_info("✓ Updated build-with-version.sh default version")

    # Run build-with-version.sh to update portunix.rc
    print_info(f"Updating portunix.rc with version {version}...")
    try:
        run_command(["./build-with-version.sh", version], cwd=project_root, capture=True, check=False)
    except Exception:
        print_warning("Version update in build script had issues, but continuing...")

    print()


def run_goreleaser(version: str, goreleaser_cmd: str) -> bool:
    """Run GoReleaser to create cross-platform release"""
    print_step("Running GoReleaser to create cross-platform release...")
    project_root = get_project_root()

    # Clean previous builds
    print_info("Cleaning previous builds...")
    dist_dir = project_root / "dist"
    if dist_dir.exists():
        shutil.rmtree(dist_dir)

    print_info("Building release packages...")

    # Create temporary git tag
    print_info(f"Creating temporary git tag {version} for build...")
    try:
        run_command(["git", "tag", "-d", version], cwd=project_root, capture=True, check=False)
    except Exception:
        pass
    run_command(["git", "tag", version], cwd=project_root)

    try:
        # Run GoReleaser
        run_command(
            [goreleaser_cmd, "release", "--clean", "--skip-validate", "--skip-publish"],
            cwd=project_root
        )
        print_info("✓ GoReleaser completed successfully")

        # Remove temporary tag
        print_info("Removing temporary git tag...")
        run_command(["git", "tag", "-d", version], cwd=project_root, capture=True)

    except subprocess.CalledProcessError:
        # Clean up tag on failure
        try:
            run_command(["git", "tag", "-d", version], cwd=project_root, capture=True, check=False)
        except Exception:
            pass
        print_error("GoReleaser failed")
        return False

    print()
    return True


def build_platform_archives() -> bool:
    """Build cross-platform binaries using Makefile"""
    print_step("Building cross-platform binaries for ADR-031...")
    project_root = get_project_root()

    print_info("Building binaries for all target platforms...")
    try:
        run_command(["make", "build-all-platforms"], cwd=project_root)
    except subprocess.CalledProcessError:
        print_warning("Platform builds had some issues, continuing...")

    # Create platform archives using Python script
    print_info("Creating platform archives...")
    try:
        python_cmd = sys.executable
        run_command(
            [python_cmd, "scripts/create-platform-archives.py"],
            cwd=project_root
        )
    except subprocess.CalledProcessError:
        print_warning("Archive creation had some issues, continuing...")

    print_info("✓ Cross-platform binaries built")
    print()
    return True


def inject_platform_archives() -> None:
    """Inject platform archives into release packages"""
    print_step("Injecting platform archives into release packages...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"
    platforms_dir = dist_dir / "platforms"

    if not platforms_dir.exists():
        print_warning("Platforms directory not found, skipping injection")
        return

    # Find all release archives
    for archive_path in list(dist_dir.glob("portunix_*.tar.gz")) + list(dist_dir.glob("portunix_*.zip")):
        archive_name = archive_path.name
        print_info(f"Processing {archive_name}...")

        # Create temp directory
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)

            # Extract archive
            if archive_name.endswith('.zip'):
                with zipfile.ZipFile(archive_path, 'r') as zipf:
                    zipf.extractall(temp_path)
            else:
                with tarfile.open(archive_path, 'r:gz') as tar:
                    tar.extractall(temp_path)

            # Create platforms directory
            platforms_dest = temp_path / "platforms"
            platforms_dest.mkdir(exist_ok=True)

            # Copy all platform archives
            for platform_archive in list(platforms_dir.glob("*.tar.gz")) + list(platforms_dir.glob("*.zip")):
                shutil.copy2(platform_archive, platforms_dest)

            # Recreate archive
            archive_path.unlink()

            if archive_name.endswith('.zip'):
                with zipfile.ZipFile(archive_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
                    for item in temp_path.rglob('*'):
                        if item.is_file():
                            arcname = item.relative_to(temp_path)
                            zipf.write(item, arcname)
            else:
                with tarfile.open(archive_path, 'w:gz') as tar:
                    for item in temp_path.iterdir():
                        tar.add(item, arcname=item.name)

            print_info(f"  ✓ Added platforms/ to {archive_name}")

    print_info("✓ Platform archives injected into all release packages")
    print()


def verify_outputs() -> bool:
    """Verify generated files"""
    print_step("Verifying generated files...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    if not dist_dir.exists():
        print_error("dist/ directory not found")
        return False

    # Count files
    archives = list(dist_dir.glob("*.tar.gz")) + list(dist_dir.glob("*.zip"))
    checksums = list(dist_dir.glob("*checksums*"))
    platforms_dir = dist_dir / "platforms"
    platform_archives = []
    if platforms_dir.exists():
        platform_archives = list(platforms_dir.glob("*.tar.gz")) + list(platforms_dir.glob("*.zip"))

    print_info("Generated files:")
    print(f"   Release archives: {len(archives)}")
    print(f"   Platform archives: {len(platform_archives)}")
    print(f"   Checksum files: {len(checksums)}")

    if len(archives) == 0:
        print_error("No archive files generated")
        return False

    print()
    print("📦 Generated release files:")
    for f in sorted(archives + checksums):
        size_mb = f.stat().st_size / (1024 * 1024)
        print(f"   {f.name} ({size_mb:.1f} MB)")

    if platform_archives:
        print()
        print("📦 Platform archives (for cross-platform provisioning):")
        for f in sorted(platform_archives):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"   {f.name} ({size_mb:.1f} MB)")

    print()
    return True


def create_release_notes(version: str) -> None:
    """Create release notes file"""
    print_step("Creating release notes...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    # Get build info
    try:
        result = run_command(["go", "version"], capture=True)
        go_version = result.stdout.strip().split()[2] if result.stdout else "unknown"
    except Exception:
        go_version = "unknown"

    try:
        result = run_command(["git", "rev-parse", "--short", "HEAD"], capture=True)
        git_commit = result.stdout.strip() if result.stdout else "unknown"
    except Exception:
        git_commit = "unknown"

    build_date = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S UTC")
    version_num = version.lstrip('v')

    release_notes = f"""# Portunix {version}

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
```bash
# AMD64
wget https://github.com/cassandragargoyle/Portunix/releases/download/{version}/portunix_{version_num}_linux_amd64.tar.gz
tar -xzf portunix_{version_num}_linux_amd64.tar.gz
cd portunix_{version_num}_linux_amd64
./install.sh
```

### Windows
```powershell
# Download and extract
# https://github.com/cassandragargoyle/Portunix/releases/download/{version}/portunix_{version_num}_windows_amd64.zip
# Then run:
.\\install.ps1
```

### macOS
```bash
# Intel Macs
wget https://github.com/cassandragargoyle/Portunix/releases/download/{version}/portunix_{version_num}_darwin_amd64.tar.gz
tar -xzf portunix_{version_num}_darwin_amd64.tar.gz
cd portunix_{version_num}_darwin_amd64
./install.sh
```

## 🚀 Quick Start

```bash
# Install development environment
portunix install default

# Manage containers
portunix container run ubuntu
portunix container ssh container-name

# Configure MCP server
portunix mcp configure
```

## 🔗 Links

- **Repository**: https://github.com/cassandragargoyle/Portunix
- **Issues**: https://github.com/cassandragargoyle/Portunix/issues
- **Documentation**: Repository docs/ directory

## 🔐 Verification

Verify downloads using SHA256 checksums provided with the release.

---

**Build Information:**
- Build Date: {build_date}
- Go Version: {go_version}
- Git Commit: {git_commit}

"""

    notes_path = dist_dir / f"RELEASE_NOTES_{version}.md"
    notes_path.write_text(release_notes)
    print_info(f"✓ Release notes created: {notes_path}")
    print()


def show_summary(version: str) -> None:
    """Show release summary"""
    print_step("🎉 Release preparation complete!")
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    print(f"{Colors.CYAN}📊 Summary:{Colors.NC}")
    print(f"   Version: {Colors.GREEN}{version}{Colors.NC}")
    print(f"   Build directory: {Colors.BLUE}dist/{Colors.NC}")
    print()

    print("📦 Generated files:")
    for f in sorted(dist_dir.iterdir()):
        if f.is_file():
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"   {f.name} ({size_mb:.1f} MB)")
    print()

    print(f"{Colors.GREEN}🚀 Next Steps:{Colors.NC}")
    print("   1. Review files in dist/ directory")
    print("   2. Test installation on different platforms")
    print("   3. Verify version in generated binaries:")
    print("      ./dist/portunix_*/portunix version")
    print("   4. Create GitHub release:")
    print(f"      - Tag: {version}")
    print(f"      - Title: Portunix {version}")
    print(f"      - Description: Use dist/RELEASE_NOTES_{version}.md")
    print("      - Upload all files from dist/")
    print()


def show_usage() -> None:
    """Show usage information"""
    print("Usage: python scripts/make-release.py <version>")
    print()
    print("Examples:")
    print("  python scripts/make-release.py v1.5.1")
    print("  python scripts/make-release.py v1.6.0")
    print()
    print("This script will:")
    print("  1. Validate version format")
    print("  2. Update version in source files")
    print("  3. Build cross-platform binaries using GoReleaser")
    print("  4. Create packages with install scripts")
    print("  5. Generate checksums")
    print("  6. Prepare everything for GitHub release")


def main() -> int:
    """Main entry point"""
    # Disable colors if not a TTY
    if not sys.stdout.isatty():
        Colors.disable()

    print_header()

    # Check arguments
    if len(sys.argv) < 2:
        print_error("Version parameter is required")
        show_usage()
        return 1

    if sys.argv[1] in ('-h', '--help'):
        show_usage()
        return 0

    version = sys.argv[1]

    # Validate version
    if not validate_version(version):
        print_error("Invalid version format. Use semantic versioning: v1.2.3")
        show_usage()
        return 1

    print_info(f"🎯 Creating release for version: {version}")
    print()

    # Check dependencies
    ok, goreleaser_cmd = check_dependencies()
    if not ok:
        return 1

    # Update version files
    update_version_files(version)

    # Run GoReleaser
    if not run_goreleaser(version, goreleaser_cmd):
        return 1

    # Build platform archives
    build_platform_archives()

    # Inject platform archives
    inject_platform_archives()

    # Verify outputs
    if not verify_outputs():
        return 1

    # Create release notes
    create_release_notes(version)

    # Show summary
    show_summary(version)

    print(f"{Colors.GREEN}✅ Release {version} ready for publication!{Colors.NC}")

    # Optional: Generate documentation
    project_root = get_project_root()
    docs_script = project_root / "scripts" / "post-release-docs.py"
    if docs_script.exists():
        print()
        print_step("Generating documentation site...")
        try:
            run_command([sys.executable, str(docs_script), version, "--build-only"], cwd=project_root)
        except subprocess.CalledProcessError:
            print_warning("Documentation generation failed (non-blocking)")

    return 0


# Entry point: This block runs only when the script is executed directly,
# not when imported as a module. It calls main() and uses its return value
# as the exit code (0 = success, non-zero = error).
if __name__ == "__main__":
    sys.exit(main())
