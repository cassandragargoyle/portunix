#!/usr/bin/env python3
"""
Remove Portunix binaries from system installation directory.
Automatically detects existing installation path.
"""

import os
import sys
import subprocess
import platform
from pathlib import Path


def find_install_dir():
    """Find existing Portunix installation directory (excluding current directory)."""
    system = platform.system()
    current_dir = Path.cwd()

    if system == "Windows":
        try:
            result = subprocess.run(
                ["where", "portunix"],
                capture_output=True,
                text=True,
                check=True
            )
            # Find first path that is NOT in current directory
            for path_str in result.stdout.strip().split('\n'):
                path = Path(path_str.strip())
                install_dir = path.parent
                if install_dir.resolve() != current_dir.resolve():
                    return install_dir
            return None
        except subprocess.CalledProcessError:
            return None
    else:
        try:
            result = subprocess.run(
                ["which", "portunix"],
                capture_output=True,
                text=True,
                check=True
            )
            install_dir = Path(result.stdout.strip()).parent
            if install_dir.resolve() != current_dir.resolve():
                return install_dir
            return None
        except subprocess.CalledProcessError:
            return None


def get_binary_extension():
    """Get platform-specific binary extension."""
    return ".exe" if platform.system() == "Windows" else ""


def remove_with_sudo(path: Path):
    """Remove file, using sudo if needed on Unix."""
    system = platform.system()

    if not path.exists():
        return True

    if system == "Windows":
        path.unlink()
    else:
        if os.access(path.parent, os.W_OK):
            path.unlink()
        else:
            subprocess.run(["sudo", "rm", "-f", str(path)], check=True)

    return True


def undeploy(install_dir: Path):
    """Remove binaries from installation directory."""
    ext = get_binary_extension()
    removed = []

    # Main binary
    main_binary = install_dir / f"portunix{ext}"
    if main_binary.exists():
        print(f"  Removing: {main_binary}")
        try:
            remove_with_sudo(main_binary)
            removed.append(main_binary.name)
        except Exception as e:
            print(f"  Error removing {main_binary.name}: {e}")

    # Helper binaries (ptx-*)
    for helper in install_dir.glob(f"ptx-*{ext}"):
        if helper.is_file():
            print(f"  Removing: {helper}")
            try:
                remove_with_sudo(helper)
                removed.append(helper.name)
            except Exception as e:
                print(f"  Error removing {helper.name}: {e}")

    return removed


def main():
    install_dir = find_install_dir()

    if install_dir is None:
        print("Portunix is not installed.")
        return

    print(f"Found installation: {install_dir}")
    print("Removing binaries...")

    removed = undeploy(install_dir)

    if removed:
        print(f"\nRemoved {len(removed)} binaries from {install_dir}")
    else:
        print("No binaries found to remove.")


if __name__ == "__main__":
    main()
