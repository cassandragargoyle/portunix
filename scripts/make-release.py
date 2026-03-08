#!/usr/bin/env python3
"""
Make Release Script for Portunix
Part of ADR-035: Separate Platforms Distribution Strategy
(Supersedes ADR-031 partially)

Usage:
    python scripts/make-release.py v1.7.8

This script:
1. Validates version format
2. Updates version in source files
3. Builds cross-platform binaries using GoReleaser
4. Creates platform archives for cross-platform distribution
5. Creates separate platforms bundle (ADR-035)
6. Generates checksums and release notes

The script automatically uses the project's .venv if available.
"""

import hashlib
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
                capture: bool = False, check: bool = True,
                env: Optional[dict] = None) -> subprocess.CompletedProcess:
    """Run a command and optionally capture output"""
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=capture,
            text=True,
            check=check,
            env=env
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
        content = build_script.read_text(encoding='utf-8')
        content = re.sub(
            r'^VERSION=\$\{1:-v[0-9]+\.[0-9]+\.[0-9]+\}',
            f'VERSION=${{1:-{version}}}',
            content,
            flags=re.MULTILINE
        )
        build_script.write_text(content, encoding='utf-8')
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
    """Build cross-platform binaries using pure Python (Windows compatible)"""
    print_step("Building cross-platform binaries for ADR-031...")
    project_root = get_project_root()

    # Define target platforms
    platforms = [
        ("linux", "amd64"),
        ("linux", "arm64"),
        ("windows", "amd64"),
        ("darwin", "amd64"),
    ]

    # Define helper binaries to build
    helpers = [
        "ptx-container",
        "ptx-virt",
        "ptx-mcp",
        "ptx-ansible",
        "ptx-prompting",
        "ptx-python",
        "ptx-installer",
        "ptx-aiops",
        "ptx-make",
        "ptx-pft",
        "ptx-credential",
    ]

    dist_dir = project_root / "dist"
    platforms_dir = dist_dir / "platforms"

    # Find .syso files (Windows resource files) that break non-Windows builds
    syso_files = list(project_root.glob("*.syso"))

    # Build for each platform
    for goos, goarch in platforms:
        platform_name = f"{goos}-{goarch}"
        platform_dir = platforms_dir / platform_name
        platform_dir.mkdir(parents=True, exist_ok=True)

        ext = ".exe" if goos == "windows" else ""
        print_info(f"Building for {goos}/{goarch}...")

        # Set environment for cross-compilation
        env = os.environ.copy()
        env["GOOS"] = goos
        env["GOARCH"] = goarch
        env["CGO_ENABLED"] = "0"

        # Temporarily rename .syso files for non-Windows builds
        # (they contain Windows-specific relocations incompatible with other platforms)
        renamed_syso = []
        if goos != "windows":
            for syso in syso_files:
                backup = syso.with_suffix(".syso.bak")
                syso.rename(backup)
                renamed_syso.append((backup, syso))

        # Build main portunix binary
        output_path = platform_dir / f"portunix{ext}"
        try:
            run_command(
                ["go", "build", "-o", str(output_path), "."],
                cwd=project_root,
                env=env,
                capture=True
            )
        except subprocess.CalledProcessError as e:
            print_warning(f"Failed to build portunix for {platform_name}: {e}")
            # Restore .syso files before continuing
            for backup, original in renamed_syso:
                backup.rename(original)
            continue
        finally:
            # Restore .syso files
            for backup, original in renamed_syso:
                if backup.exists():
                    backup.rename(original)

        # Build helper binaries
        for helper in helpers:
            helper_dir = project_root / "src" / "helpers" / helper
            if not helper_dir.exists():
                continue

            helper_output = platform_dir / f"{helper}{ext}"
            try:
                run_command(
                    ["go", "build", "-o", str(helper_output), "."],
                    cwd=helper_dir,
                    env=env,
                    capture=True
                )
            except subprocess.CalledProcessError as e:
                print_warning(f"Failed to build {helper} for {platform_name}: {e}")

        print_info(f"  ✓ {platform_name} complete")

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


def create_platforms_bundle(version: str) -> None:
    """Create separate platforms bundle (ADR-035)"""
    print_step("Creating separate platforms bundle (ADR-035)...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"
    platforms_dir = dist_dir / "platforms"

    if not platforms_dir.exists():
        print_warning("Platforms directory not found, skipping bundle creation")
        return

    # Get version number without 'v' prefix
    version_num = version.lstrip('v')

    # Create bundle: portunix-platforms_X.Y.Z.tar.gz
    bundle_name = f"portunix-platforms_{version_num}.tar.gz"
    bundle_path = dist_dir / bundle_name

    print_info(f"Creating {bundle_name}...")

    # Create temp directory for bundle structure
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create platforms/ subdirectory in bundle
        platforms_dest = temp_path / "platforms"
        platforms_dest.mkdir(exist_ok=True)

        # Copy all platform archives
        for platform_archive in list(platforms_dir.glob("*.tar.gz")) + list(platforms_dir.glob("*.zip")):
            shutil.copy2(platform_archive, platforms_dest)
            print_info(f"  Added {platform_archive.name}")

        # Create the bundle archive
        with tarfile.open(bundle_path, 'w:gz') as tar:
            tar.add(platforms_dest, arcname="platforms")

    size_mb = bundle_path.stat().st_size / (1024 * 1024)
    print_info(f"✓ Created {bundle_name} ({size_mb:.1f} MB)")
    print()


def inject_platform_archives() -> None:
    """Legacy function - now a no-op per ADR-035"""
    # ADR-035: Platform archives are now distributed as separate bundle
    # This function is kept for backwards compatibility but does nothing
    print_info("Skipping platform injection (ADR-035: separate bundle)")
    print()


def copy_quickstart_scripts() -> None:
    """Copy quickstart scripts to dist and update checksums"""
    print_step("Copying quickstart scripts...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"
    quickstart_dir = project_root / "release-assets" / "quickstart"

    if not quickstart_dir.exists():
        print_warning("Quickstart directory not found, skipping")
        return

    # Copy all quickstart scripts
    copied_files = []
    for script in quickstart_dir.glob("*"):
        if script.is_file():
            dest = dist_dir / script.name
            shutil.copy2(script, dest)
            copied_files.append(dest)
            print_info(f"✓ Copied {script.name}")

    # Update checksums file
    if copied_files:
        # Find checksums file
        checksums_files = list(dist_dir.glob("checksums_*.txt"))
        if checksums_files:
            checksums_file = checksums_files[0]
            with open(checksums_file, 'a') as f:
                for file_path in copied_files:
                    # Calculate SHA256
                    sha256_hash = hashlib.sha256()
                    with open(file_path, 'rb') as file:
                        for chunk in iter(lambda: file.read(4096), b''):
                            sha256_hash.update(chunk)
                    checksum = sha256_hash.hexdigest()
                    f.write(f"{checksum}  {file_path.name}\n")
            print_info(f"✓ Updated checksums in {checksums_file.name}")

    print()


def verify_outputs() -> bool:
    """Verify generated files"""
    print_step("Verifying generated files...")
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    if not dist_dir.exists():
        print_error("dist/ directory not found")
        return False

    # Count files - separate platform-specific archives from platforms bundle
    all_archives = list(dist_dir.glob("*.tar.gz")) + list(dist_dir.glob("*.zip"))
    release_archives = [f for f in all_archives if not f.name.startswith("portunix-platforms")]
    platforms_bundle = [f for f in all_archives if f.name.startswith("portunix-platforms")]
    checksums = list(dist_dir.glob("*checksums*"))

    print_info("Generated files (ADR-035 structure):")
    print(f"   Platform-specific archives: {len(release_archives)}")
    print(f"   Platforms bundle: {len(platforms_bundle)}")
    print(f"   Checksum files: {len(checksums)}")

    if len(release_archives) == 0:
        print_error("No release archive files generated")
        return False

    print()
    print("📦 Platform-specific release files (slim, no bundled platforms/):")
    for f in sorted(release_archives):
        size_mb = f.stat().st_size / (1024 * 1024)
        print(f"   {f.name} ({size_mb:.1f} MB)")

    if platforms_bundle:
        print()
        print("📦 Platforms bundle (separate download for cross-platform users):")
        for f in sorted(platforms_bundle):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"   {f.name} ({size_mb:.1f} MB)")

    if checksums:
        print()
        print("🔐 Checksum files:")
        for f in sorted(checksums):
            print(f"   {f.name}")

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

### Cross-Platform Provisioning (Optional)

If you need to provision containers or VMs with different platforms (e.g., deploy Linux
binaries from Windows), download the separate platforms bundle:

```bash
# Download platforms bundle (contains all platform binaries)
wget https://github.com/cassandragargoyle/Portunix/releases/download/{version}/portunix-platforms_{version_num}.tar.gz

# Extract to Portunix installation directory
tar -xzf portunix-platforms_{version_num}.tar.gz -C /usr/local/portunix/
```

This enables cross-platform workflows like `ptx-ansible` provisioning Linux containers from Windows.

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
    notes_path.write_text(release_notes, encoding='utf-8')
    print_info(f"✓ Release notes created: {notes_path}")
    print()


def show_summary(version: str) -> None:
    """Show release summary"""
    print_step("🎉 Release preparation complete!")
    project_root = get_project_root()
    dist_dir = project_root / "dist"
    version_num = version.lstrip('v')

    print(f"{Colors.CYAN}📊 Summary (ADR-035 Structure):{Colors.NC}")
    print(f"   Version: {Colors.GREEN}{version}{Colors.NC}")
    print(f"   Build directory: {Colors.BLUE}dist/{Colors.NC}")
    print()

    # Separate files by type
    all_files = sorted([f for f in dist_dir.iterdir() if f.is_file()])
    release_archives = [f for f in all_files if f.name.startswith("portunix_")]
    platforms_bundle = [f for f in all_files if f.name.startswith("portunix-platforms")]
    other_files = [f for f in all_files if not f.name.startswith("portunix")]

    print("📦 Platform-specific packages (for users):")
    for f in release_archives:
        size_mb = f.stat().st_size / (1024 * 1024)
        print(f"   {f.name} ({size_mb:.1f} MB)")

    if platforms_bundle:
        print()
        print("📦 Cross-platform bundle (optional, for provisioning):")
        for f in platforms_bundle:
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"   {f.name} ({size_mb:.1f} MB)")

    if other_files:
        print()
        print("📄 Other files:")
        for f in other_files:
            size_mb = f.stat().st_size / (1024 * 1024)
            if size_mb > 0.01:
                print(f"   {f.name} ({size_mb:.1f} MB)")
            else:
                print(f"   {f.name}")

    print()
    print(f"{Colors.GREEN}🚀 Next Steps:{Colors.NC}")
    print("   1. Review files in dist/ directory")
    print("   2. Test installation on different platforms")
    print("   3. Verify version in generated binaries:")
    print("      ./dist/portunix_*/portunix version")
    print("   4. Create release and upload assets:")
    print(f"      - Tag: {version}")
    print(f"      - Title: Portunix {version}")
    print(f"      - Upload: portunix_*.tar.gz, portunix_*.zip")
    print(f"      - Upload: portunix-platforms_{version_num}.tar.gz (cross-platform bundle)")
    print(f"      - Upload: checksums, RELEASE_NOTES_{version}.md")
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

    # Create separate platforms bundle (ADR-035)
    create_platforms_bundle(version)

    # Copy quickstart scripts
    copy_quickstart_scripts()

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
