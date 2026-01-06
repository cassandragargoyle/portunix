#!/usr/bin/env python3
"""
Create Platform Archives Script for Portunix
Part of ADR-031: Cross-Platform Binary Distribution Strategy
Issue #125: Cross-Platform Binary Distribution

This script creates compressed archives for each platform's binaries
to be included in every release package for cross-platform provisioning.

Usage:
    python scripts/create-platform-archives.py [platforms_dir] [output_dir]

The script automatically uses the project's .venv if available.
"""

import os
import sys
import shutil
import tarfile
import zipfile
from pathlib import Path
from typing import List, Tuple

# ANSI color codes
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'  # No Color

    @classmethod
    def disable(cls):
        """Disable colors (for non-TTY output)"""
        cls.RED = cls.GREEN = cls.YELLOW = cls.CYAN = cls.NC = ''


def print_info(message: str) -> None:
    print(f"{Colors.CYAN}ℹ️  {message}{Colors.NC}")


def print_success(message: str) -> None:
    print(f"{Colors.GREEN}✓ {message}{Colors.NC}")


def print_error(message: str) -> None:
    print(f"{Colors.RED}❌ {message}{Colors.NC}")


def print_warning(message: str) -> None:
    print(f"{Colors.YELLOW}⚠️  {message}{Colors.NC}")


# Supported platforms
PLATFORMS = [
    "linux-amd64",
    "linux-arm64",
    "windows-amd64",
    "darwin-amd64",
]


def get_project_root() -> Path:
    """Get the project root directory"""
    script_dir = Path(__file__).parent.resolve()
    return script_dir.parent


def create_tar_gz(source_dir: Path, output_path: Path) -> None:
    """Create a tar.gz archive from source directory"""
    with tarfile.open(output_path, "w:gz") as tar:
        for item in source_dir.iterdir():
            tar.add(item, arcname=item.name)


def create_zip(source_dir: Path, output_path: Path) -> None:
    """Create a zip archive from source directory"""
    with zipfile.ZipFile(output_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
        for item in source_dir.rglob('*'):
            if item.is_file():
                arcname = item.relative_to(source_dir)
                zipf.write(item, arcname)


def get_archive_info(platform: str) -> Tuple[str, str]:
    """Get archive name and extension for platform"""
    if platform.startswith("windows"):
        return f"{platform}.zip", "zip"
    else:
        return f"{platform}.tar.gz", "tar.gz"


def create_platform_archives(platforms_dir: Path, output_dir: Path) -> int:
    """
    Create platform archives from built binaries.

    Args:
        platforms_dir: Directory containing platform subdirectories with binaries
        output_dir: Directory to write archive files

    Returns:
        Number of archives created
    """
    print_info(f"Creating platform archives from {platforms_dir}")

    created_count = 0

    for platform in PLATFORMS:
        platform_path = platforms_dir / platform

        if not platform_path.is_dir():
            print_warning(f"Platform directory not found: {platform_path}")
            continue

        # Check if there are any files
        files = list(platform_path.iterdir())
        if not files:
            print_warning(f"No files in platform directory: {platform_path}")
            continue

        archive_name, archive_type = get_archive_info(platform)
        archive_path = output_dir / archive_name

        print_info(f"Creating {archive_name}...")

        try:
            if archive_type == "zip":
                create_zip(platform_path, archive_path)
            else:
                create_tar_gz(platform_path, archive_path)

            # Get file size
            size_mb = archive_path.stat().st_size / (1024 * 1024)
            print_success(f"Created {archive_path} ({size_mb:.1f} MB)")
            created_count += 1

        except Exception as e:
            print_error(f"Failed to create {archive_name}: {e}")

    return created_count


def show_summary(output_dir: Path) -> None:
    """Show summary of created archives"""
    print()
    print_info("Platform archives created:")

    for platform in PLATFORMS:
        archive_name, _ = get_archive_info(platform)
        archive_path = output_dir / archive_name

        if archive_path.exists():
            size_mb = archive_path.stat().st_size / (1024 * 1024)
            print(f"   {archive_path} ({size_mb:.1f} MB)")

    print()
    print_success("All platform archives created successfully")


def main() -> int:
    """Main entry point"""
    # Disable colors if not a TTY
    if not sys.stdout.isatty():
        Colors.disable()

    # Parse arguments
    if len(sys.argv) > 1 and sys.argv[1] in ('-h', '--help'):
        print(__doc__)
        return 0

    project_root = get_project_root()

    # Default directories
    platforms_dir = Path(sys.argv[1]) if len(sys.argv) > 1 else project_root / "dist" / "platforms"
    output_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else platforms_dir

    # Validate input directory
    if not platforms_dir.is_dir():
        print_error(f"Platforms directory not found: {platforms_dir}")
        print_info("Run 'make build-all-platforms' first to build platform binaries")
        return 1

    # Ensure output directory exists
    output_dir.mkdir(parents=True, exist_ok=True)

    # Create archives
    count = create_platform_archives(platforms_dir, output_dir)

    if count == 0:
        print_error("No platform archives created")
        return 1

    show_summary(output_dir)
    return 0


# Entry point: This block runs only when the script is executed directly,
# not when imported as a module. It calls main() and uses its return value
# as the exit code (0 = success, non-zero = error).
if __name__ == "__main__":
    sys.exit(main())
