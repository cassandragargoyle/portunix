#!/usr/bin/env python3
"""
Deploy locally built Portunix binaries to system installation directory.
Automatically detects existing installation path and copies all binaries.
"""

import os
import sys
import shutil
import subprocess
import platform
from pathlib import Path


def find_install_dir(source_dir: Path):
    """Find existing Portunix installation directory (excluding source directory)."""
    system = platform.system()

    if system == "Windows":
        # Use 'where' command on Windows
        try:
            result = subprocess.run(
                ["where", "portunix"],
                capture_output=True,
                text=True,
                check=True
            )
            # Find first path that is NOT in source directory
            for path_str in result.stdout.strip().split('\n'):
                path = Path(path_str.strip())
                install_dir = path.parent
                # Skip if it's the source/build directory
                if install_dir.resolve() != source_dir.resolve():
                    return install_dir
            return None
        except subprocess.CalledProcessError:
            return None
    else:
        # Use 'which' command on Unix
        try:
            result = subprocess.run(
                ["which", "portunix"],
                capture_output=True,
                text=True,
                check=True
            )
            install_dir = Path(result.stdout.strip()).parent
            # Skip if it's the source/build directory
            if install_dir.resolve() != source_dir.resolve():
                return install_dir
            return None
        except subprocess.CalledProcessError:
            return None


def get_binary_extension():
    """Get platform-specific binary extension."""
    return ".exe" if platform.system() == "Windows" else ""


def get_binaries(source_dir: Path):
    """Get list of binaries to deploy."""
    ext = get_binary_extension()
    binaries = []

    # Main binary
    main_binary = source_dir / f"portunix{ext}"
    if main_binary.exists():
        binaries.append(main_binary)

    # Helper binaries (ptx-*)
    for helper in source_dir.glob(f"ptx-*{ext}"):
        if helper.is_file():
            binaries.append(helper)

    return binaries


def copy_with_sudo(src: Path, dest: Path):
    """Copy file, using sudo if needed on Unix."""
    system = platform.system()

    if system == "Windows":
        shutil.copy2(src, dest)
    else:
        # Check if we have write permission
        if os.access(dest.parent, os.W_OK):
            shutil.copy2(src, dest)
        else:
            subprocess.run(["sudo", "cp", str(src), str(dest)], check=True)


def deploy(source_dir: Path, install_dir: Path):
    """Deploy binaries to installation directory."""
    binaries = get_binaries(source_dir)

    if not binaries:
        print("Error: No binaries found to deploy.")
        print(f"Expected binaries in: {source_dir}")
        return False

    print(f"Deploying {len(binaries)} binaries to {install_dir}")

    for binary in binaries:
        dest = install_dir / binary.name
        print(f"  {binary.name} -> {dest}")
        try:
            copy_with_sudo(binary, dest)
        except Exception as e:
            print(f"  Error copying {binary.name}: {e}")
            return False

    return True


def run_install_script(source_dir: Path):
    """Run installation script for first-time install."""
    system = platform.system()

    if system == "Windows":
        script = source_dir / "scripts" / "install.ps1"
        if script.exists():
            print(f"Running install script: {script}")
            subprocess.run(["powershell", "-ExecutionPolicy", "Bypass", "-File", str(script)], check=True)
        else:
            print(f"Error: Install script not found: {script}")
            return False
    else:
        script = source_dir / "scripts" / "install.sh"
        if script.exists():
            print(f"Running install script: {script}")
            subprocess.run(["bash", str(script)], check=True)
        else:
            print(f"Error: Install script not found: {script}")
            return False

    return True


def main():
    # Determine source directory (where this script is run from)
    source_dir = Path.cwd()

    # Allow override via command line argument
    if len(sys.argv) > 1:
        source_dir = Path(sys.argv[1])

    print(f"Source directory: {source_dir}")

    # Find existing installation (excluding source directory)
    install_dir = find_install_dir(source_dir)

    if install_dir is None:
        print("No existing Portunix installation found.")
        print("Running first-time installation...")
        if run_install_script(source_dir):
            # After install, find the new install directory
            install_dir = find_install_dir()
            if install_dir:
                print(f"Installed to: {install_dir}")
            else:
                print("Installation completed but could not verify install path.")
        return

    print(f"Found existing installation: {install_dir}")

    # Deploy binaries
    if deploy(source_dir, install_dir):
        print("Deployment successful!")

        # Show version
        ext = get_binary_extension()
        portunix_path = install_dir / f"portunix{ext}"
        try:
            result = subprocess.run(
                [str(portunix_path), "version"],
                capture_output=True,
                text=True
            )
            print(f"\nInstalled version:\n{result.stdout.strip()}")
        except Exception:
            pass
    else:
        print("Deployment failed!")
        sys.exit(1)


if __name__ == "__main__":
    main()
