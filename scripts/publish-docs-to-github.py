#!/usr/bin/env python3
"""
Publish Documentation to GitHub Pages
Deploys generated static documentation site to GitHub Pages using gh CLI
"""

import os
import sys
import subprocess
import shutil
import argparse
from pathlib import Path
from datetime import datetime
from typing import Optional

# Configuration
SCRIPT_DIR = Path(__file__).parent.absolute()
PROJECT_ROOT = SCRIPT_DIR.parent
DOCS_SITE_DIR = PROJECT_ROOT / "docs-site"
PUBLIC_DIR = DOCS_SITE_DIR / "public"
PORTUNIX_BIN = PROJECT_ROOT / "portunix"

# Color codes for output
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'  # No Color

def print_step(message: str):
    """Print step message in blue"""
    print(f"{Colors.BLUE}==>{Colors.NC} {message}")

def print_success(message: str, *details):
    """Print success message in green"""
    msg = f"{Colors.GREEN}✓{Colors.NC} {message}"
    if details:
        msg += " " + " ".join(str(d) for d in details)
    print(msg)

def print_error(message: str):
    """Print error message in red"""
    print(f"{Colors.RED}✗{Colors.NC} {message}", file=sys.stderr)

def print_warning(message: str):
    """Print warning message in yellow"""
    print(f"{Colors.YELLOW}⚠{Colors.NC} {message}")

def print_info(message: str):
    """Print info message"""
    print(f"   {message}")

def run_command(cmd: list, cwd: Optional[Path] = None, capture: bool = True) -> tuple:
    """Run a command and return result"""
    try:
        if capture:
            result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
            return result.returncode, result.stdout, result.stderr
        else:
            result = subprocess.run(cmd, cwd=cwd)
            return result.returncode, "", ""
    except FileNotFoundError as e:
        return 1, "", f"Command not found: {cmd[0]}"
    except subprocess.SubprocessError as e:
        return 1, "", str(e)

def check_github_cli():
    """Check if GitHub CLI is installed and authenticated"""
    print_step("Checking GitHub CLI...")

    # Check if gh is installed
    returncode, stdout, stderr = run_command(['gh', '--version'])
    if returncode != 0:
        print_warning("GitHub CLI is not installed, attempting to install via Portunix...")

        # Check if Portunix binary exists
        if not PORTUNIX_BIN.exists():
            print_error(f"Portunix binary not found at: {PORTUNIX_BIN}")
            print_error("Please build the project first: go build -o .")
            return False

        # Install GitHub CLI using Portunix
        print_info("Installing GitHub CLI...")
        returncode, stdout, stderr = run_command([str(PORTUNIX_BIN), 'install', 'github-cli'])

        if returncode != 0:
            print_error("Failed to install GitHub CLI via Portunix")
            print_error(f"Error: {stderr}")
            print_info("Please install GitHub CLI manually: https://cli.github.com/")
            return False

        print_success("GitHub CLI installed successfully")

        # Verify installation
        returncode, stdout, stderr = run_command(['gh', '--version'])
        if returncode != 0:
            print_error("GitHub CLI installed but not available in PATH")
            print_info("You may need to restart your terminal or update PATH")
            return False

    # Display version
    version_line = stdout.split('\n')[0] if stdout else "unknown version"
    print_success(f"GitHub CLI found: {version_line}")

    # Check authentication status
    print_step("Checking GitHub authentication...")
    returncode, stdout, stderr = run_command(['gh', 'auth', 'status'])

    if returncode != 0:
        print_warning("GitHub CLI is not authenticated")
        print_info("Please authenticate with: gh auth login")
        print_info("Choose authentication method:")
        print_info("  1. Login via web browser (recommended)")
        print_info("  2. Paste authentication token")
        return False

    print_success("GitHub CLI is authenticated")
    return True

def check_repository():
    """Check if we're in a git repository with GitHub remote"""
    print_step("Checking repository status...")

    # Check if we're in a git repo
    returncode, stdout, stderr = run_command(['git', 'status'], cwd=PROJECT_ROOT)
    if returncode != 0:
        print_error("Not in a git repository")
        return False

    print_success("Git repository found")

    # Check for GitHub remote
    returncode, stdout, stderr = run_command(['git', 'remote', 'get-url', 'origin'], cwd=PROJECT_ROOT)
    if returncode != 0:
        # Try github remote
        returncode, stdout, stderr = run_command(['git', 'remote', 'get-url', 'github'], cwd=PROJECT_ROOT)
        if returncode != 0:
            print_error("No GitHub remote found")
            print_info("Add GitHub remote with:")
            print_info("  git remote add origin https://github.com/cassandragargoyle/Portunix.git")
            return False
        else:
            remote_name = 'github'
    else:
        remote_name = 'origin'

    remote_url = stdout.strip()
    if 'github.com' not in remote_url.lower():
        print_warning(f"Remote '{remote_name}' is not a GitHub URL: {remote_url}")
        print_info("This script is designed for GitHub Pages deployment")
        return False

    print_success(f"GitHub remote found: {remote_url}")
    return True

def check_public_dir():
    """Check if public directory exists with built documentation"""
    print_step("Checking documentation build...")

    if not PUBLIC_DIR.exists():
        print_error(f"Public directory not found: {PUBLIC_DIR}")
        print_info("Please build documentation first:")
        print_info("  python scripts/post-release-docs.py")
        return False

    # Check if directory has content
    files = list(PUBLIC_DIR.glob('*'))
    if not files:
        print_error("Public directory is empty")
        print_info("Please build documentation first:")
        print_info("  python scripts/post-release-docs.py")
        return False

    # Check for index.html
    if not (PUBLIC_DIR / 'index.html').exists():
        print_warning("index.html not found in public directory")
        print_info("Documentation may not be properly built")

    print_success(f"Documentation found: {len(files)} files/directories")
    return True

def create_gh_pages_branch():
    """Create or update gh-pages branch"""
    print_step("Managing gh-pages branch...")

    # Create temporary directory for gh-pages
    temp_dir = Path('/tmp/portunix-gh-pages')
    if temp_dir.exists():
        shutil.rmtree(temp_dir)
    temp_dir.mkdir(parents=True)

    # Check if gh-pages branch exists on remote
    returncode, stdout, stderr = run_command(
        ['git', 'ls-remote', '--heads', 'origin', 'gh-pages'],
        cwd=PROJECT_ROOT
    )

    gh_pages_exists = returncode == 0 and 'gh-pages' in stdout

    if gh_pages_exists:
        print_info("gh-pages branch exists, cloning...")
        # Clone only gh-pages branch
        returncode, stdout, stderr = run_command([
            'git', 'clone', '-b', 'gh-pages', '--single-branch',
            '.', str(temp_dir)
        ], cwd=PROJECT_ROOT)

        if returncode != 0:
            print_error(f"Failed to clone gh-pages branch: {stderr}")
            return None

        # Clean existing files (except .git)
        for item in temp_dir.iterdir():
            if item.name != '.git':
                if item.is_dir():
                    shutil.rmtree(item)
                else:
                    item.unlink()
    else:
        print_info("Creating new gh-pages branch...")
        # Initialize new repo
        returncode, stdout, stderr = run_command(['git', 'init'], cwd=temp_dir)
        if returncode != 0:
            print_error(f"Failed to initialize git repo: {stderr}")
            return None

        # Create orphan branch
        returncode, stdout, stderr = run_command(
            ['git', 'checkout', '-b', 'gh-pages'],
            cwd=temp_dir
        )
        if returncode != 0:
            print_error(f"Failed to create gh-pages branch: {stderr}")
            return None

        # Add remote
        returncode, stdout, stderr = run_command([
            'git', 'remote', 'add', 'origin',
            'https://github.com/cassandragargoyle/Portunix.git'
        ], cwd=temp_dir)

    print_success("gh-pages branch ready")
    return temp_dir

def copy_documentation(temp_dir: Path):
    """Copy built documentation to temp directory"""
    print_step("Copying documentation files...")

    # Copy all files from public to temp directory
    file_count = 0
    for item in PUBLIC_DIR.iterdir():
        dest = temp_dir / item.name
        if item.is_dir():
            shutil.copytree(item, dest)
        else:
            shutil.copy2(item, dest)
        file_count += 1

    # Create .nojekyll file to bypass Jekyll processing
    nojekyll = temp_dir / '.nojekyll'
    nojekyll.touch()

    # Create CNAME file if custom domain is configured
    # (Optional - only if using custom domain)

    print_success(f"Copied {file_count} items")
    return True

def commit_and_push(temp_dir: Path, version: str, message: Optional[str] = None):
    """Commit changes and push to GitHub"""
    print_step("Committing changes...")

    # Configure git if needed
    returncode, stdout, stderr = run_command(
        ['git', 'config', 'user.email'],
        cwd=temp_dir
    )
    if returncode != 0:
        run_command([
            'git', 'config', 'user.email', 'noreply@github.com'
        ], cwd=temp_dir)
        run_command([
            'git', 'config', 'user.name', 'Portunix Documentation Bot'
        ], cwd=temp_dir)

    # Add all files
    returncode, stdout, stderr = run_command(['git', 'add', '-A'], cwd=temp_dir)
    if returncode != 0:
        print_error(f"Failed to add files: {stderr}")
        return False

    # Check if there are changes
    returncode, stdout, stderr = run_command(
        ['git', 'status', '--porcelain'],
        cwd=temp_dir
    )

    if not stdout.strip():
        print_warning("No changes to commit")
        return True

    # Commit changes
    commit_message = message or f"Update documentation for version {version} - {datetime.now().strftime('%Y-%m-%d %H:%M')}"
    returncode, stdout, stderr = run_command([
        'git', 'commit', '-m', commit_message
    ], cwd=temp_dir)

    if returncode != 0:
        print_error(f"Failed to commit: {stderr}")
        return False

    print_success("Changes committed")

    # Push to GitHub using gh CLI for better authentication handling
    print_step("Pushing to GitHub Pages...")

    returncode, stdout, stderr = run_command([
        'git', 'push', 'origin', 'gh-pages', '--force'
    ], cwd=temp_dir)

    if returncode != 0:
        print_warning("Direct push failed, trying with gh CLI...")
        # Try using gh CLI for push
        returncode, stdout, stderr = run_command([
            'gh', 'repo', 'sync', '--branch', 'gh-pages'
        ], cwd=temp_dir)

        if returncode != 0:
            print_error(f"Failed to push: {stderr}")
            print_info("Try pushing manually:")
            print_info(f"  cd {temp_dir}")
            print_info("  git push origin gh-pages --force")
            return False

    print_success("Documentation pushed to GitHub Pages")
    return True

def get_pages_url():
    """Get GitHub Pages URL using gh CLI"""
    print_step("Getting GitHub Pages URL...")

    # Get repository info using gh CLI
    returncode, stdout, stderr = run_command([
        'gh', 'api', 'repos/{owner}/{repo}',
        '--jq', '.homepage'
    ])

    if returncode == 0 and stdout.strip():
        pages_url = stdout.strip()
    else:
        # Fallback to default URL pattern
        returncode, stdout, stderr = run_command([
            'gh', 'repo', 'view', '--json', 'owner,name',
            '--jq', r'"\(.owner.login)/\(.name)"'
        ])

        if returncode == 0:
            repo_info = stdout.strip()
            owner, name = repo_info.split('/')
            pages_url = f"https://{owner.lower()}.github.io/{name}/"
        else:
            pages_url = "https://cassandragargoyle.github.io/Portunix/"

    return pages_url

def cleanup(temp_dir: Path):
    """Clean up temporary directory"""
    print_step("Cleaning up...")

    if temp_dir and temp_dir.exists():
        shutil.rmtree(temp_dir)
        print_success("Temporary files removed")

def main():
    """Main execution"""
    parser = argparse.ArgumentParser(
        description='Publish Portunix Documentation to GitHub Pages'
    )
    parser.add_argument('version', nargs='?', default='latest',
                       help='Version being published')
    parser.add_argument('-m', '--message', type=str,
                       help='Custom commit message')
    parser.add_argument('--skip-checks', action='store_true',
                       help='Skip preliminary checks (use with caution)')
    parser.add_argument('--dry-run', action='store_true',
                       help='Perform dry run without pushing')

    args = parser.parse_args()

    print("=" * 60)
    print("Portunix Documentation Publisher")
    print(f"Version: {args.version}")
    print("=" * 60)
    print()

    # Perform checks
    if not args.skip_checks:
        if not check_github_cli():
            print_error("GitHub CLI check failed")
            print_info("Install with: portunix install github-cli")
            print_info("Then authenticate: gh auth login")
            return 1

        if not check_repository():
            print_error("Repository check failed")
            return 1

        if not check_public_dir():
            print_error("Documentation check failed")
            return 1

    # Create/update gh-pages branch
    temp_dir = create_gh_pages_branch()
    if not temp_dir:
        print_error("Failed to prepare gh-pages branch")
        return 1

    try:
        # Copy documentation
        if not copy_documentation(temp_dir):
            print_error("Failed to copy documentation")
            return 1

        if args.dry_run:
            print_warning("DRY RUN - Skipping commit and push")
            print_info(f"Documentation prepared in: {temp_dir}")
            print_info("Review changes and push manually if needed")
            return 0

        # Commit and push
        if not commit_and_push(temp_dir, args.version, args.message):
            print_error("Failed to publish documentation")
            return 1

        # Get and display Pages URL
        pages_url = get_pages_url()

        print()
        print("=" * 60)
        print_success("Documentation published successfully!")
        print()
        print(f"  📚 View documentation at: {Colors.CYAN}{pages_url}{Colors.NC}")
        print(f"  ⏱️  Note: GitHub Pages may take a few minutes to update")
        print()
        print("=" * 60)

    finally:
        # Clean up temporary directory
        if not args.dry_run:
            cleanup(temp_dir)

    return 0

if __name__ == '__main__':
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print("\n\nOperation cancelled by user")
        sys.exit(130)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)