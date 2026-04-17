#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""
Upload Release to Gitea Script for Portunix

This script uploads release artifacts from dist/ directory to Gitea Releases.
Uses the 'tea' CLI (Gitea's official command-line tool).

Usage:
    python scripts/upload-release-to-gitea.py v1.10.7
    python scripts/upload-release-to-gitea.py v1.10.7 --draft
    python scripts/upload-release-to-gitea.py v1.10.7 --prerelease

Requirements:
    - tea CLI must be installed and configured with a login
    - Release artifacts must exist in dist/ directory
    - Run 'python scripts/make-release.py <version>' first
"""

import os
import re
import sys
import shutil
import subprocess
from pathlib import Path
from typing import List, Optional, Tuple

# Configuration
GITEA_REPO = "CassandraGargoyle/portunix"
GITEA_LOGIN = "gitea"
GITEA_URL = "http://gitea:3000"

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
    print(f"{Colors.BLUE}================================================")
    print(f"  GITEA RELEASE UPLOADER")
    print(f"  Upload Portunix release to Gitea")
    print(f"================================================{Colors.NC}")
    print()


def print_step(message: str) -> None:
    print(f"{Colors.GREEN}==> {message}{Colors.NC}")


def print_info(message: str) -> None:
    print(f"{Colors.CYAN}    {message}{Colors.NC}")


def print_warning(message: str) -> None:
    print(f"{Colors.YELLOW}    WARNING: {message}{Colors.NC}")


def print_error(message: str) -> None:
    print(f"{Colors.RED}    ERROR: {message}{Colors.NC}")


def print_success(message: str) -> None:
    print(f"{Colors.GREEN}    OK: {message}{Colors.NC}")


def get_project_root() -> Path:
    """Get the project root directory"""
    script_dir = Path(__file__).parent.resolve()
    return script_dir.parent


def validate_version(version: str) -> bool:
    """Validate version format (vX.Y.Z or vX.Y.Z-SNAPSHOT)"""
    pattern = r'^v\d+\.\d+\.\d+(-SNAPSHOT)?$'
    return bool(re.match(pattern, version))


def find_tea() -> Optional[str]:
    """Find tea CLI binary"""
    # Check PATH
    tea_path = shutil.which("tea")
    if tea_path:
        return tea_path

    # Check common locations on Windows
    common_paths = [
        os.path.expandvars(r"%USERPROFILE%\tea.exe"),
        r"D:\portunix\tea.exe",
        r"C:\portunix\tea.exe",
    ]
    for path in common_paths:
        if os.path.isfile(path):
            return path

    # Check common locations on Linux/macOS
    unix_paths = [
        "/usr/local/bin/tea",
        os.path.expanduser("~/bin/tea"),
        os.path.expanduser("~/.local/bin/tea"),
    ]
    for path in unix_paths:
        if os.path.isfile(path):
            return path

    return None


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


def check_tea_cli(tea_bin: str) -> bool:
    """Check if tea CLI is installed and has a login configured"""
    print_step("Checking tea CLI...")

    try:
        result = run_command([tea_bin, "--version"], capture=True)
        version_line = result.stdout.strip().split('\n')[0] if result.stdout else "unknown"
        print_info(f"tea version: {version_line}")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print_error("tea CLI is not installed")
        print_info("Download from: https://gitea.com/gitea/tea/releases")
        return False

    # Check logins
    try:
        result = run_command([tea_bin, "logins", "list"], capture=True)
        if GITEA_LOGIN not in (result.stdout or ""):
            print_warning(f"Login '{GITEA_LOGIN}' not found in tea configuration")
            print_info(f"Run: tea login add --name {GITEA_LOGIN} --url {GITEA_URL} --token <YOUR_TOKEN>")
            print_info("Generate token at: {}/user/settings/applications".format(GITEA_URL))
            return False
        print_success(f"Login '{GITEA_LOGIN}' configured")
    except subprocess.CalledProcessError:
        print_error("Failed to list tea logins")
        return False

    print()
    return True


def find_release_files(version: str, dist_dir: Path) -> Tuple[List[Path], Optional[Path], Optional[Path]]:
    """Find release files for the given version"""
    version_num = version.lstrip('v')

    # Find archive files
    archives = []
    for pattern in [f"portunix_{version_num}_*.tar.gz", f"portunix_{version_num}_*.zip"]:
        archives.extend(dist_dir.glob(pattern))

    # Find additional assets
    extra_assets = []
    quickstart = dist_dir / "quickstart-docusaurus.ps1"
    if quickstart.exists():
        extra_assets.append(quickstart)

    # Find checksums file
    checksums = list(dist_dir.glob(f"checksums_{version_num}.txt"))
    checksum_file = checksums[0] if checksums else None

    # Find release notes
    release_notes = list(dist_dir.glob(f"RELEASE_NOTES_{version}.md"))
    notes_file = release_notes[0] if release_notes else None

    return archives + extra_assets, checksum_file, notes_file


def check_existing_release(tea_bin: str, version: str) -> bool:
    """Check if a release with this tag already exists"""
    try:
        result = run_command(
            [tea_bin, "release", "list", "--login", GITEA_LOGIN, "--repo", GITEA_REPO],
            capture=True, check=False
        )
        if result.stdout and version in result.stdout:
            return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        pass
    return False


def delete_existing_release(tea_bin: str, version: str) -> bool:
    """Delete an existing release"""
    try:
        run_command(
            [tea_bin, "release", "delete", "--login", GITEA_LOGIN, "--repo", GITEA_REPO,
             "--tag", version, "--confirm"],
            capture=True
        )
        print_info(f"Deleted existing release {version}")
        return True
    except subprocess.CalledProcessError:
        print_error(f"Failed to delete existing release {version}")
        return False


def upload_release(tea_bin: str, version: str, archives: List[Path],
                   checksum_file: Optional[Path], notes_file: Optional[Path],
                   draft: bool = False, prerelease: bool = False) -> bool:
    """Upload release to Gitea"""
    print_step(f"Creating Gitea release {version}...")

    # Read release notes
    notes_content = f"Portunix {version} release"
    if notes_file and notes_file.exists():
        notes_content = notes_file.read_text(encoding="utf-8")

    # Build command
    cmd = [
        tea_bin, "release", "create",
        "--login", GITEA_LOGIN,
        "--repo", GITEA_REPO,
        "--tag", version,
        "--title", f"Portunix {version}",
        "--note", notes_content,
    ]

    if draft:
        cmd.append("--draft")
        print_info("Creating as DRAFT release")

    if prerelease or "-SNAPSHOT" in version:
        cmd.append("--prerelease")
        print_info("Creating as PRE-RELEASE")

    # Add asset files
    all_files = list(archives)
    if checksum_file and checksum_file.exists():
        all_files.append(checksum_file)

    for f in all_files:
        cmd.extend(["--asset", str(f)])

    # Show what we're uploading
    print_info("Files to upload:")
    for f in all_files:
        size_mb = f.stat().st_size / (1024 * 1024)
        print(f"      {f.name} ({size_mb:.1f} MB)")
    print()

    # Execute
    try:
        print_info("Uploading to Gitea...")
        run_command(cmd)
        return True
    except subprocess.CalledProcessError:
        print_error("Failed to create Gitea release")
        return False


def show_usage() -> None:
    """Show usage information"""
    print("Usage: python scripts/upload-release-to-gitea.py <version> [options]")
    print()
    print("Options:")
    print("  --draft       Create as draft release (not published)")
    print("  --prerelease  Mark as pre-release")
    print("  --force       Delete existing release and re-create")
    print()
    print("Examples:")
    print("  python scripts/upload-release-to-gitea.py v1.10.7")
    print("  python scripts/upload-release-to-gitea.py v1.10.7 --draft")
    print("  python scripts/upload-release-to-gitea.py v1.10.7 --force")
    print()
    print("Prerequisites:")
    print("  1. Run 'python scripts/make-release.py v1.10.7' first")
    print(f"  2. tea CLI must be installed and configured:")
    print(f"     tea login add --name {GITEA_LOGIN} --url {GITEA_URL} --token <TOKEN>")
    print(f"     Generate token at: {GITEA_URL}/user/settings/applications")


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
    force = '--force' in sys.argv

    # Validate version
    if not validate_version(version):
        print_error("Invalid version format. Use: vX.Y.Z or vX.Y.Z-SNAPSHOT")
        return 1

    print_info(f"Version: {version}")
    print()

    # Find tea binary
    tea_bin = find_tea()
    if not tea_bin:
        print_error("tea CLI not found")
        print_info("Download from: https://gitea.com/gitea/tea/releases")
        print_info("Or install with: portunix install tea")
        return 1

    # Check tea CLI
    if not check_tea_cli(tea_bin):
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

    print_info(f"Found {len(archives)} file(s)")
    if checksum_file:
        print_info("Found checksums file")
    if notes_file:
        print_info("Found release notes")
    print()

    # Check for existing release
    if check_existing_release(tea_bin, version):
        if force:
            print_warning(f"Release {version} already exists, deleting (--force)")
            if not delete_existing_release(tea_bin, version):
                return 1
        else:
            print_error(f"Release {version} already exists on Gitea")
            print_info("Use --force to delete and re-create")
            return 1

    # Upload
    if not upload_release(tea_bin, version, archives, checksum_file, notes_file, draft, prerelease):
        return 1

    print()
    print_success(f"Release {version} uploaded to Gitea!")
    print()
    print_info("View release at:")
    print(f"   {GITEA_URL}/{GITEA_REPO}/releases/tag/{version}")

    return 0


# Entry point
if __name__ == "__main__":
    sys.exit(main())
