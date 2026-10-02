#!/usr/bin/env python3
# /// script
# requires-python = ">=3.14"
# dependencies = []
# ///
"""
Deploy locally built Portunix binaries to system installation directory.
Automatically detects existing installation path and copies all binaries.

Use --verbose for step-by-step diagnostics (useful when the deploy appears to
hang: every step is printed and flushed, so the last line pinpoints the stall).
"""

import argparse
import os
import sys
import shutil
import subprocess
import time
from pathlib import Path

# True on Windows. Uses os.name instead of the platform module: platform.system()
# / platform.uname() can hang on some Windows machines (WMI query / gethostname),
# while os.name is an instant, side-effect-free constant.
IS_WINDOWS = os.name == "nt"

# Set from --verbose in main()
VERBOSE = False


def vlog(msg):
    """Print a diagnostic line only in verbose mode (flushed immediately)."""
    if VERBOSE:
        print(f"[verbose] {msg}", flush=True)


def find_install_dir(source_dir: Path):
    """Find existing Portunix installation directory (excluding source directory)."""
    vlog(f"find_install_dir: is_windows={IS_WINDOWS}, source_dir={source_dir}")

    probe = ["where", "portunix"] if IS_WINDOWS else ["which", "portunix"]
    try:
        vlog(f"find_install_dir: running {' '.join(probe)} ...")
        result = subprocess.run(
            probe,
            capture_output=True,
            text=True,
            check=True,
            timeout=15,
        )
        vlog(f"find_install_dir: output:\n{result.stdout.strip()}")
    except subprocess.CalledProcessError:
        vlog("find_install_dir: probe returned non-zero (portunix not found)")
        return None
    except subprocess.TimeoutExpired:
        vlog("find_install_dir: probe timed out")
        return None

    # Find first path that is NOT in the source/build directory
    for path_str in result.stdout.strip().split("\n"):
        path_str = path_str.strip()
        if not path_str:
            continue
        install_dir = Path(path_str).parent
        vlog(f"find_install_dir: candidate={install_dir}")
        if install_dir.resolve() != source_dir.resolve():
            vlog(f"find_install_dir: selected={install_dir}")
            return install_dir

    vlog("find_install_dir: no candidate outside source dir")
    return None


def get_binary_extension():
    """Get platform-specific binary extension."""
    return ".exe" if IS_WINDOWS else ""


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
    if IS_WINDOWS:
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

    print(f"Deploying {len(binaries)} binaries to {install_dir}", flush=True)

    for binary in binaries:
        dest = install_dir / binary.name
        print(f"  {binary.name} -> {dest}", flush=True)
        if VERBOSE:
            src_size = binary.stat().st_size if binary.exists() else -1
            dst_exists = dest.exists()
            vlog(f"copy start: {binary.name} ({src_size} bytes), dest exists={dst_exists}")
        started = time.monotonic()
        try:
            copy_with_sudo(binary, dest)
        except Exception as e:
            print(f"  Error copying {binary.name}: {e}", flush=True)
            return False
        vlog(f"copy done: {binary.name} in {time.monotonic() - started:.2f}s")

    return True


def run_install_script(source_dir: Path):
    """Run installation script for first-time install."""
    if IS_WINDOWS:
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


def parse_args():
    parser = argparse.ArgumentParser(
        description="Deploy locally built Portunix binaries to the system installation directory."
    )
    parser.add_argument(
        "source_dir",
        nargs="?",
        default=None,
        help="Source directory holding the built binaries (default: current directory)",
    )
    parser.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="Print detailed step-by-step diagnostics",
    )
    return parser.parse_args()


def main():
    global VERBOSE
    args = parse_args()
    VERBOSE = args.verbose

    print("=== deploy-local: deploying pre-built Portunix binaries ===", flush=True)

    # Determine source directory (where this script is run from, or explicit arg)
    source_dir = Path(args.source_dir) if args.source_dir else Path.cwd()
    vlog(f"main: source_dir={source_dir}")
    vlog(f"main: sys.platform={sys.platform}, python={sys.version.split()[0]}")

    print(f"Source directory: {source_dir}", flush=True)

    # Find existing installation (excluding source directory)
    install_dir = find_install_dir(source_dir)

    if install_dir is None:
        print("No existing Portunix installation found.")
        print("Running first-time installation...")
        if run_install_script(source_dir):
            # After install, find the new install directory
            install_dir = find_install_dir(source_dir)
            if install_dir:
                print(f"Installed to: {install_dir}")
            else:
                print("Installation completed but could not verify install path.")
        return

    print(f"Found existing installation: {install_dir}", flush=True)

    # Deploy binaries
    if deploy(source_dir, install_dir):
        print("Deployment successful!", flush=True)

        # Show version
        ext = get_binary_extension()
        portunix_path = install_dir / f"portunix{ext}"
        vlog(f"main: verifying version via {portunix_path} version")
        try:
            result = subprocess.run(
                [str(portunix_path), "version"],
                capture_output=True,
                text=True,
                timeout=15,
            )
            print(f"\nInstalled version:\n{result.stdout.strip()}", flush=True)
        except Exception as e:
            print(f"(version check skipped: {e})", flush=True)
    else:
        print("Deployment failed!")
        sys.exit(1)


if __name__ == "__main__":
    main()
