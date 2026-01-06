#!/usr/bin/env python3
"""
Upload Release to GitHub Script for Portunix

This script uploads release artifacts from dist/ directory to GitHub Releases.

Usage:
    python scripts/upload-release-to-github.py v1.9.2
    python scripts/upload-release-to-github.py v1.9.2 --draft
    python scripts/upload-release-to-github.py v1.9.2 --prerelease

Requirements:
    - GitHub CLI (gh) must be installed and authenticated
    - Release artifacts must exist in dist/ directory
"""

import os
import re
import sys
import subprocess
from pathlib import Path
from typing import List, Optional, Tuple

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
    print(f"║      🚀 GITHUB RELEASE UPLOADER          ║")
    print(f"║   Upload Portunix release to GitHub      ║")
    print(f"╚══════════════════════════════════════════╝{Colors.NC}")
    print()


def print_step(message: str) -> None:
    print(f"{Colors.GREEN}📋 {message}{Colors.NC}")


def print_info(message: str) -> None:
    print(f"{Colors.CYAN}ℹ️  {message}{Colors.NC}")


def print_warning(message: str) -> None:
    print(f"{Colors.YELLOW}⚠️  {message}{Colors.NC}")


def print_error(message: str) -> None:
    print(f"{Colors.RED}❌ {message}{Colors.NC}")


def print_success(message: str) -> None:
    print(f"{Colors.GREEN}✓ {message}{Colors.NC}")


def get_project_root() -> Path:
    """Get the project root directory"""
    script_dir = Path(__file__).parent.resolve()
    return script_dir.parent


def validate_version(version: str) -> bool:
    """Validate version format (vX.Y.Z or vX.Y.Z-SNAPSHOT)"""
    pattern = r'^v\d+\.\d+\.\d+(-SNAPSHOT)?$'
    return bool(re.match(pattern, version))


def run_command(cmd: List[str], capture: bool = False, check: bool = True) -> subprocess.CompletedProcess:
    """Run a command and optionally capture output"""
    try:
        result = subprocess.run(
            cmd,
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


def check_gh_cli() -> bool:
    """Check if GitHub CLI is installed and authenticated"""
    print_step("Checking GitHub CLI...")

    try:
        result = run_command(["gh", "--version"], capture=True)
        version_line = result.stdout.strip().split('\n')[0] if result.stdout else "unknown"
        print_info(f"gh version: {version_line}")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print_error("GitHub CLI (gh) is not installed")
        print_info("Install with: https://cli.github.com/")
        return False

    # Check authentication
    try:
        result = run_command(["gh", "auth", "status"], capture=True, check=False)
        if result.returncode != 0:
            print_error("GitHub CLI is not authenticated")
            print_info("Run: gh auth login")
            return False
        print_success("GitHub CLI authenticated")
    except FileNotFoundError:
        return False

    print()
    return True


def find_release_files(version: str, dist_dir: Path) -> Tuple[List[Path], Optional[Path]]:
    """Find release files for the given version"""
    version_num = version.lstrip('v')

    # Find archive files
    archives = []
    for pattern in [f"portunix_{version_num}_*.tar.gz", f"portunix_{version_num}_*.zip"]:
        archives.extend(dist_dir.glob(pattern))

    # Find checksums file
    checksums = list(dist_dir.glob(f"checksums_{version_num}.txt"))
    checksum_file = checksums[0] if checksums else None

    # Find release notes
    release_notes = list(dist_dir.glob(f"RELEASE_NOTES_{version}.md"))
    notes_file = release_notes[0] if release_notes else None

    return archives, checksum_file, notes_file


def upload_release(version: str, archives: List[Path], checksum_file: Optional[Path],
                   notes_file: Optional[Path], draft: bool = False,
                   prerelease: bool = False) -> bool:
    """Upload release to GitHub"""
    print_step(f"Creating GitHub release {version}...")

    # Build command
    cmd = ["gh", "release", "create", version]

    # Add options
    cmd.extend(["--title", f"Portunix {version}"])

    if notes_file and notes_file.exists():
        cmd.extend(["--notes-file", str(notes_file)])
    else:
        cmd.extend(["--notes", f"Portunix {version} release"])

    if draft:
        cmd.append("--draft")
        print_info("Creating as DRAFT release")

    if prerelease or "-SNAPSHOT" in version:
        cmd.append("--prerelease")
        print_info("Creating as PRE-RELEASE")

    # Add files
    for archive in archives:
        cmd.append(str(archive))

    if checksum_file and checksum_file.exists():
        cmd.append(str(checksum_file))

    # Show what we're uploading
    print_info("Files to upload:")
    for archive in archives:
        size_mb = archive.stat().st_size / (1024 * 1024)
        print(f"   {archive.name} ({size_mb:.1f} MB)")
    if checksum_file:
        print(f"   {checksum_file.name}")
    print()

    # Execute
    try:
        print_info("Uploading to GitHub...")
        run_command(cmd)
        return True
    except subprocess.CalledProcessError:
        print_error("Failed to create GitHub release")
        return False


def show_usage() -> None:
    """Show usage information"""
    print("Usage: python scripts/upload-release-to-github.py <version> [options]")
    print()
    print("Options:")
    print("  --draft       Create as draft release (not published)")
    print("  --prerelease  Mark as pre-release")
    print()
    print("Examples:")
    print("  python scripts/upload-release-to-github.py v1.9.2")
    print("  python scripts/upload-release-to-github.py v1.9.2 --draft")
    print("  python scripts/upload-release-to-github.py v1.9.2-SNAPSHOT --prerelease")
    print()
    print("Prerequisites:")
    print("  1. Run 'python scripts/make-release.py v1.9.2' first")
    print("  2. GitHub CLI must be installed and authenticated (gh auth login)")


def main() -> int:
    """Main entry point"""
    # Disable colors if not a TTY
    if not sys.stdout.isatty():
        Colors.disable()

    print_header()

    # Parse arguments
    if len(sys.argv) < 2 or sys.argv[1] in ('-h', '--help'):
        show_usage()
        return 0 if len(sys.argv) > 1 else 1

    version = sys.argv[1]
    draft = '--draft' in sys.argv
    prerelease = '--prerelease' in sys.argv

    # Validate version
    if not validate_version(version):
        print_error("Invalid version format. Use: vX.Y.Z or vX.Y.Z-SNAPSHOT")
        return 1

    print_info(f"Version: {version}")
    print()

    # Check GitHub CLI
    if not check_gh_cli():
        return 1

    # Check dist directory
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    if not dist_dir.exists():
        print_error("dist/ directory not found")
        print_info("Run 'python scripts/make-release.py' first")
        return 1

    # Find release files
    print_step("Finding release files...")
    archives, checksum_file, notes_file = find_release_files(version, dist_dir)

    if not archives:
        print_error(f"No release archives found for version {version}")
        print_info("Run 'python scripts/make-release.py' first")
        return 1

    print_info(f"Found {len(archives)} archive(s)")
    if checksum_file:
        print_info("Found checksums file")
    if notes_file:
        print_info("Found release notes")
    print()

    # Upload
    if not upload_release(version, archives, checksum_file, notes_file, draft, prerelease):
        return 1

    print()
    print_success(f"Release {version} uploaded to GitHub!")
    print()
    print_info("View release at:")
    print(f"   https://github.com/cassandragargoyle/Portunix/releases/tag/{version}")

    return 0


# Entry point
if __name__ == "__main__":
    sys.exit(main())
