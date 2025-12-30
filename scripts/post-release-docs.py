#!/usr/bin/env python3
"""
Post-Release Documentation Generation Script
Generates static documentation site from Portunix commands and deploys to GitHub Pages
Called automatically by make-release.sh or manually for testing
"""

import os
import sys
import json
import re
import shutil
import argparse
import subprocess
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional, Tuple

# Configuration
SCRIPT_DIR = Path(__file__).parent.absolute()
PROJECT_ROOT = SCRIPT_DIR.parent
DOCS_SITE_DIR = PROJECT_ROOT / "docs-site"
CONTENT_DIR = DOCS_SITE_DIR / "content"
DOCS_DIR = CONTENT_DIR / "docs"  # Hugo Book expects content in docs/
COMMAND_DOCS_DIR = DOCS_DIR / "commands"  # Will be docs/commands/ for Hugo Book
# Use .exe extension on Windows
if sys.platform == 'win32':
    PORTUNIX_BIN = PROJECT_ROOT / "portunix.exe"
else:
    PORTUNIX_BIN = PROJECT_ROOT / "portunix"
HUGO_CMD = 'hugo'  # Default Hugo command, may be updated during dependency check

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

def check_hugo_version(hugo_cmd: str) -> Tuple[bool, str, bool]:
    """
    Check Hugo version and if it's extended
    Returns: (version_ok, version_string, is_extended)
    """
    try:
        result = subprocess.run([hugo_cmd, 'version'], capture_output=True, text=True)
        if result.returncode != 0:
            return False, "", False

        version_output = result.stdout.strip()
        print_info(f"Hugo version output: {version_output}")

        # Parse version - Hugo outputs like: "hugo v0.150.1+extended linux/amd64 BuildDate=..."
        # or "hugo v0.150.1 linux/amd64 BuildDate=..." for non-extended
        is_extended = "+extended" in version_output.lower() or "extended" in version_output.lower()

        # Extract version number (e.g., "0.150.1" from "hugo v0.150.1+extended...")
        version_match = re.search(r'v?(\d+\.\d+\.\d+)', version_output)
        if not version_match:
            return False, version_output, is_extended

        version_str = version_match.group(1)

        # Check if version is >= 0.146.0
        def version_tuple(v):
            return tuple(map(int, v.split('.')))

        current_version = version_tuple(version_str)
        required_version = version_tuple("0.146.0")

        version_ok = current_version >= required_version

        return version_ok, version_str, is_extended

    except (FileNotFoundError, subprocess.SubprocessError) as e:
        return False, str(e), False

def check_dependencies() -> bool:
    """Check for required dependencies"""
    print_step("Checking dependencies...")

    # Check for Portunix binary first (needed to install Hugo)
    if not PORTUNIX_BIN.exists():
        print_error(f"Portunix binary not found at: {PORTUNIX_BIN}")
        print("Please build the project first: go build -o .")
        return False

    # Check for Hugo version and extended support
    global HUGO_CMD
    hugo_needs_install = False

    try:
        # First try to find existing Hugo
        result = subprocess.run([HUGO_CMD, 'version'], capture_output=True, text=True)
        if result.returncode == 0:
            version_ok, version_str, is_extended = check_hugo_version(HUGO_CMD)

            if version_ok and is_extended:
                print_success(f"Hugo Extended {version_str} found (compatible with Hugo Book theme)")
            elif version_ok and not is_extended:
                print_warning(f"Hugo {version_str} found but not Extended version")
                print_warning("Hugo Book theme requires Hugo Extended with SCSS/Sass support")
                hugo_needs_install = True
            elif not version_ok:
                print_warning(f"Hugo {version_str} found but version < 0.146.0 required")
                print_warning("Hugo Book theme requires Hugo Extended >= 0.146.0")
                hugo_needs_install = True
            else:
                hugo_needs_install = True
        else:
            raise FileNotFoundError
    except (FileNotFoundError, subprocess.SubprocessError):
        print_warning("Hugo is not installed")
        hugo_needs_install = True

    if hugo_needs_install:
        print_step("Installing Hugo Extended via Portunix...")

        # Try to install Hugo Extended using Portunix
        try:
            install_result = subprocess.run(
                [str(PORTUNIX_BIN), 'install', 'hugo-extended'],
                capture_output=True,
                text=True,
                timeout=120  # 2 minute timeout for installation
            )

            if install_result.returncode == 0:
                print_success("Hugo Extended installed successfully via Portunix")

                # Try multiple Hugo locations
                hugo_locations = [
                    'hugo',  # In PATH
                    '/usr/local/bin/hugo',  # Common Linux location
                    'C:\\Portunix\\bin\\hugo\\hugo.exe',  # Windows location
                    str(Path.home() / '.local' / 'bin' / 'hugo'),  # User local bin
                ]

                hugo_cmd = None
                for location in hugo_locations:
                    try:
                        version_ok, version_str, is_extended = check_hugo_version(location)
                        if version_ok and is_extended:
                            hugo_cmd = location
                            print_success(f"Hugo Extended {version_str} verified at {location}")
                            break
                        elif version_ok:
                            print_warning(f"Hugo {version_str} found at {location} but not Extended")
                        else:
                            print_warning(f"Hugo found at {location} but version < 0.146.0")
                    except (FileNotFoundError, subprocess.SubprocessError):
                        continue

                if not hugo_cmd:
                    print_error("Hugo Extended installed but not found or incompatible version")
                    print("Hugo Book theme requires Hugo Extended >= 0.146.0")
                    print("Please check installation or install manually:")
                    print("  https://gohugo.io/installation/")
                    return False

                # Store the working Hugo command for later use
                HUGO_CMD = hugo_cmd
            else:
                print_error("Failed to install Hugo Extended via Portunix")
                print("Error:", install_result.stderr)
                print("Please install Hugo Extended manually:")
                print("  https://gohugo.io/installation/")
                print("  Required: Hugo Extended >= 0.146.0")
                return False

        except subprocess.TimeoutExpired:
            print_error("Hugo Extended installation timed out")
            return False
        except subprocess.SubprocessError as e:
            print_error(f"Failed to run Portunix install: {e}")
            return False

    # Final verification
    version_ok, version_str, is_extended = check_hugo_version(HUGO_CMD)
    if not (version_ok and is_extended):
        print_error(f"Hugo verification failed: version={version_str}, extended={is_extended}")
        print("Hugo Book theme requires Hugo Extended >= 0.146.0")
        return False

    # Check for jq (optional)
    try:
        subprocess.run(['jq', '--version'], capture_output=True, text=True)
        print_success("jq found (optional)")
    except (FileNotFoundError, subprocess.SubprocessError):
        print_warning("jq is not installed (optional for JSON processing)")

    print_success("Dependencies checked and compatible")
    return True

def install_hugo_book_theme():
    """Install Hugo Book theme from GitHub"""
    theme_dir = DOCS_SITE_DIR / "themes"
    book_theme_dir = theme_dir / "hugo-book"

    # Create themes directory if it doesn't exist
    theme_dir.mkdir(parents=True, exist_ok=True)

    # Check if theme already exists and is not empty
    # (empty directory can exist from failed clone)
    if book_theme_dir.exists():
        theme_files = list(book_theme_dir.iterdir())
        if theme_files:
            print_info("Hugo Book theme already installed")
            return True
        else:
            # Empty directory - remove it and re-clone
            print_warning("Hugo Book theme directory is empty, removing and re-cloning...")
            shutil.rmtree(book_theme_dir)

    print_step("Installing Hugo Book theme...")

    try:
        # Clone Hugo Book theme with --depth 1 for faster download
        result = subprocess.run(
            ["git", "clone", "--depth", "1", "https://github.com/alex-shpak/hugo-book", str(book_theme_dir)],
            check=True,
            capture_output=True,
            text=True,
            timeout=120  # 2 minute timeout
        )
        print_success("Hugo Book theme installed successfully")
        return True
    except subprocess.TimeoutExpired:
        print_error("Hugo Book theme clone timed out")
        print_warning("Falling back to basic theme creation")
        create_basic_theme()
        return False
    except subprocess.CalledProcessError as e:
        print_error(f"Failed to install Hugo Book theme: {e}")
        if e.stderr:
            print_info(f"Git error: {e.stderr}")
        print_warning("Falling back to basic theme creation")
        create_basic_theme()
        return False

def create_basic_theme():
    """Create basic Hugo theme structure (fallback)"""
    theme_dir = DOCS_SITE_DIR / "themes" / "portunix-docs"

    # Create theme directories
    (theme_dir / "layouts" / "_default").mkdir(parents=True, exist_ok=True)
    (theme_dir / "layouts" / "partials").mkdir(parents=True, exist_ok=True)
    (theme_dir / "static" / "css").mkdir(parents=True, exist_ok=True)

    # Create base layout
    base_layout = '''<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }} | Portunix Documentation</title>
    <link rel="stylesheet" href="{{ "/css/style.css" | relURL }}">
</head>
<body>
    <header>
        <nav>
            <h1><a href="{{ .Site.BaseURL }}">Portunix Documentation</a></h1>
            <ul>
                {{ range .Site.Menus.main }}
                <li><a href="{{ .URL }}">{{ .Name }}</a></li>
                {{ end }}
            </ul>
        </nav>
    </header>
    <main>
        {{ block "main" . }}{{ end }}
    </main>
    <footer>
        <p>© 2025 Portunix | Version {{ .Site.Params.version }}</p>
    </footer>
</body>
</html>'''

    (theme_dir / "layouts" / "_default" / "baseof.html").write_text(base_layout, encoding='utf-8')

    # Create single page layout
    single_layout = '''{{ define "main" }}
<article>
    <h1>{{ .Title }}</h1>
    {{ .Content }}
</article>
{{ end }}'''

    (theme_dir / "layouts" / "_default" / "single.html").write_text(single_layout, encoding='utf-8')

    # Create list layout
    list_layout = '''{{ define "main" }}
<section>
    <h1>{{ .Title }}</h1>
    {{ .Content }}

    {{ if .Sections }}
    <h2>Sections</h2>
    <ul>
        {{ range .Sections }}
        <li>
            <a href="{{ .RelPermalink }}">{{ .Title }}</a>
            {{ if .Description }}<p>{{ .Description }}</p>{{ end }}
        </li>
        {{ end }}
    </ul>
    {{ end }}

    {{ if .RegularPages }}
    <h2>Pages</h2>
    <ul>
        {{ range .RegularPages }}
        <li>
            <a href="{{ .RelPermalink }}">{{ .Title }}</a>
            {{ if .Summary }}<p>{{ .Summary }}</p>{{ end }}
        </li>
        {{ end }}
    </ul>
    {{ end }}
</section>
{{ end }}'''

    (theme_dir / "layouts" / "_default" / "list.html").write_text(list_layout, encoding='utf-8')

    # Create basic CSS
    css_content = '''body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.6;
    color: #333;
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    border-bottom: 2px solid #0366d6;
    padding-bottom: 10px;
    margin-bottom: 30px;
}

nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

nav h1 {
    margin: 0;
}

nav h1 a {
    color: #0366d6;
    text-decoration: none;
}

nav ul {
    list-style: none;
    display: flex;
    gap: 20px;
    margin: 0;
    padding: 0;
}

nav a {
    color: #0366d6;
    text-decoration: none;
}

nav a:hover {
    text-decoration: underline;
}

pre {
    background: #f6f8fa;
    padding: 16px;
    overflow: auto;
    border-radius: 6px;
}

code {
    background: #f6f8fa;
    padding: 2px 4px;
    border-radius: 3px;
}

footer {
    margin-top: 50px;
    padding-top: 20px;
    border-top: 1px solid #e1e4e8;
    text-align: center;
    color: #586069;
}'''

    (theme_dir / "static" / "css" / "style.css").write_text(css_content, encoding='utf-8')

def init_hugo_site(version: str) -> bool:
    """Initialize Hugo site structure"""
    print_step("Initializing Hugo site structure...")

    # Create docs-site directory if it doesn't exist
    if not DOCS_SITE_DIR.exists():
        DOCS_SITE_DIR.mkdir(parents=True)

        # Initialize Hugo site
        try:
            subprocess.run([HUGO_CMD, 'new', 'site', '.', '--force'],
                         cwd=DOCS_SITE_DIR, check=True, capture_output=True)
        except subprocess.CalledProcessError as e:
            print_error(f"Failed to initialize Hugo site: {e}")
            return False

    # Try to install Hugo Book theme first
    theme_installed = install_hugo_book_theme()
    theme_name = 'hugo-book' if theme_installed else 'portunix-docs'

    # Create Hugo configuration for Hugo Book
    if theme_installed:
        # Hugo Book configuration
        hugo_config = f'''baseURL = 'https://cassandragargoyle.github.io/Portunix/'
languageCode = 'en-us'
title = 'Portunix Documentation'
theme = 'hugo-book'

# (Optional) Theme color scheme
# Available values: light, dark, auto
# Default: auto
[params]
  BookTheme = 'light'

  # (Optional) Set source repository location.
  # Used for 'Edit this page' links
  BookRepo = 'https://github.com/cassandragargoyle/Portunix'

  # (Optional) Specify branch for repository above
  BookRepoBranch = 'main'

  # (Optional) Configure path to repository 'Edit' link
  BookRepoPath = ''

  # (Optional) Configure edit link prefix
  BookEditPath = 'edit/main/'

  # (Optional) Configure date format used in various places in the theme
  BookDateFormat = 'January 2, 2006'

  # (Optional) Enable search feature
  BookSearch = true

  # (Optional) Set Google Analytics if you use it to track your website.
  BookGoogleAnalytics = ''

  # (Optional) Provide logos for dark/light mode
  BookLogo = ''

  # (Optional) Set description in header
  description = 'Universal development environment management tool'
  version = '{version}'

  # (Optional) Enable comments template on pages
  # By default partials/docs/comments.html includes Disqus template
  # See https://gohugo.io/content-management/comments/#configure-disqus
  # Can be overriden by same param in page frontmatter
  BookComments = false

  # (Optional) Set path to section to render as menu
  # When not set, renders menu from data/_index.md files
  BookSection = 'docs'

  # (Optional) This value is duplicate of $BookSection
  # Use it when you want to use different path to menu data
  BookMenuBundle = '/menu'

  # (Optional) Collapse menu by default
  BookCollapseSection = false

  # (Optional) Hide table of contents on page
  BookToC = true

  # (Optional) If you have analytics, enable it
  # googleAnalytics = ''

[menu]
# [[menu.before]]
[[menu.after]]
  name = "GitHub"
  url = "https://github.com/cassandragargoyle/Portunix"
  weight = 10
[[menu.after]]
  name = "Releases"
  url = "https://github.com/cassandragargoyle/Portunix/releases"
  weight = 20
'''
    else:
        # Fallback to custom theme configuration
        hugo_config = f'''baseURL = 'https://cassandragargoyle.github.io/Portunix/'
languageCode = 'en-us'
title = 'Portunix Documentation'
theme = 'portunix-docs'

[params]
  description = 'Universal development environment management tool'
  version = '{version}'
  github_repo = 'https://github.com/cassandragargoyle/Portunix'

[menu]
  [[menu.main]]
    name = 'Commands'
    url = '/commands/'
    weight = 1
  [[menu.main]]
    name = 'Guides'
    url = '/guides/'
    weight = 2
  [[menu.main]]
    name = 'Releases'
    url = '/releases/'
    weight = 3
  [[menu.main]]
    name = 'GitHub'
    url = 'https://github.com/cassandragargoyle/Portunix'
    weight = 4
'''
        # If Hugo Book failed, create basic theme
        if not theme_installed:
            create_basic_theme()

    # Write the configuration
    (DOCS_SITE_DIR / "hugo.toml").write_text(hugo_config, encoding='utf-8')

    # Ensure content directories exist
    # For Hugo Book theme, structure content under docs/
    if theme_installed:
        DOCS_DIR.mkdir(parents=True, exist_ok=True)
        (COMMAND_DOCS_DIR / "core").mkdir(parents=True, exist_ok=True)
        (COMMAND_DOCS_DIR / "plugins").mkdir(parents=True, exist_ok=True)
        (DOCS_DIR / "guides").mkdir(parents=True, exist_ok=True)
        (DOCS_DIR / "releases").mkdir(parents=True, exist_ok=True)
    else:
        # For custom theme, use standard structure
        (CONTENT_DIR / "commands" / "core").mkdir(parents=True, exist_ok=True)
        (CONTENT_DIR / "commands" / "plugins").mkdir(parents=True, exist_ok=True)
        (CONTENT_DIR / "guides").mkdir(parents=True, exist_ok=True)
        (CONTENT_DIR / "releases").mkdir(parents=True, exist_ok=True)

    # Create main commands index page if it doesn't exist
    commands_index = COMMAND_DOCS_DIR / "_index.md"
    if not commands_index.exists():
        if theme_installed:
            # Hugo Book format
            commands_index_content = f'''---
title: "Commands"
weight: 2
bookFlatSection: true
---

# Portunix Commands

Comprehensive documentation for all Portunix commands.

## Core Commands

Built-in commands available in all Portunix installations.

[View Core Commands →]({{{{< relref "/docs/commands/core" >}}}})

## Plugin Commands

Commands provided by installed plugins.

[View Plugin Commands →]({{{{< relref "/docs/commands/plugins" >}}}})
'''
        else:
            # Custom theme format
            commands_index_content = f'''---
title: "Commands"
date: {datetime.now().isoformat()}
draft: false
---

# Portunix Commands

Comprehensive documentation for all Portunix commands.

## Core Commands

Built-in commands available in all Portunix installations.

[View Core Commands →](core/)

## Plugin Commands

Commands provided by installed plugins.

[View Plugin Commands →](plugins/)
'''
        commands_index.write_text(commands_index_content, encoding='utf-8')

    # Create guides index page placeholder
    guides_index = DOCS_DIR / "guides" / "_index.md" if theme_installed else CONTENT_DIR / "guides" / "_index.md"
    if not guides_index.exists():
        if theme_installed:
            # Hugo Book format
            guides_index_content = '''---
title: "Guides"
weight: 3
bookFlatSection: true
---

# Portunix Guides

Step-by-step tutorials and how-to guides for common tasks.

## Coming Soon

Comprehensive guides are being prepared and will be available in future releases.

For now, please refer to:
- [Commands Documentation]({{< relref "/docs/commands" >}})
- [GitHub Repository](https://github.com/cassandragargoyle/Portunix)
- [GitHub Discussions](https://github.com/cassandragargoyle/Portunix/discussions)
'''
        else:
            # Custom theme format
            guides_index_content = f'''---
title: "Guides"
date: {datetime.now().isoformat()}
draft: false
---

# Portunix Guides

Step-by-step tutorials and how-to guides for common tasks.

## Coming Soon

Comprehensive guides are being prepared and will be available in future releases.

For now, please refer to:
- [Commands Documentation](../commands/)
- [GitHub Repository](https://github.com/cassandragargoyle/Portunix)
- [GitHub Discussions](https://github.com/cassandragargoyle/Portunix/discussions)
'''
        guides_index.write_text(guides_index_content, encoding='utf-8')

    # Create releases index page placeholder
    releases_index = DOCS_DIR / "releases" / "_index.md" if theme_installed else CONTENT_DIR / "releases" / "_index.md"
    if not releases_index.exists():
        if theme_installed:
            # Hugo Book format
            releases_index_content = '''---
title: "Release Notes"
weight: 4
bookFlatSection: true
---

# Portunix Release Notes

Latest updates and version history.

## Latest Release

Release notes are automatically generated during the release process.

For the latest releases, please visit:
- [GitHub Releases](https://github.com/cassandragargoyle/Portunix/releases)
'''
        else:
            # Custom theme format
            releases_index_content = f'''---
title: "Release Notes"
date: {datetime.now().isoformat()}
draft: false
---

# Portunix Release Notes

Latest updates and version history.

## Latest Release

Release notes are automatically generated during the release process.

For the latest releases, please visit:
- [GitHub Releases](https://github.com/cassandragargoyle/Portunix/releases)
'''
        releases_index.write_text(releases_index_content, encoding='utf-8')

    # Create main index page
    if theme_installed:
        # For Hugo Book, create docs/_index.md
        main_index = DOCS_DIR / "_index.md"
    else:
        # For custom theme, create content/_index.md
        main_index = CONTENT_DIR / "_index.md"

    if not main_index.exists():
        if theme_installed:
            # Hugo Book format with weight for ordering
            main_index_content = f'''---
title: "Introduction"
type: docs
weight: 1
---

# Portunix Documentation

**Universal development environment management tool**

Portunix is a cross-platform CLI tool that simplifies the installation and management of development tools, packages, and environments.

## Quick Start

```bash
# Install Portunix
curl -sf https://raw.githubusercontent.com/cassandragargoyle/Portunix/main/scripts/install.sh | bash

# Install development environment
portunix install default

# Start working!
```

## Documentation Sections

### 📚 [Commands]({{{{< relref "/docs/commands" >}}}})
Complete reference for all Portunix commands - both core and plugin commands.

### 📖 [Guides]({{{{< relref "/docs/guides" >}}}})
Step-by-step tutorials and how-to guides for common tasks.

### 📋 [Release Notes]({{{{< relref "/docs/releases" >}}}})
Latest updates and version history.

---

## Features

- **Cross-platform**: Windows, Linux, macOS support
- **Package Management**: Universal installer for development tools
- **Container Integration**: Docker and Podman support
- **Plugin System**: Extensible with gRPC-based plugins
- **AI Integration**: MCP server for AI assistants
- **Self-updating**: Automatic updates from GitHub releases

## Links

- [GitHub Repository](https://github.com/cassandragargoyle/Portunix)
- [Issues & Bug Reports](https://github.com/cassandragargoyle/Portunix/issues)
- [Discussions](https://github.com/cassandragargoyle/Portunix/discussions)
'''
        else:
            # Custom theme format
            main_index_content = f'''---
title: "Portunix Documentation"
date: {datetime.now().isoformat()}
draft: false
---

# Portunix Documentation

**Universal development environment management tool**

Portunix is a cross-platform CLI tool that simplifies the installation and management of development tools, packages, and environments.

## Quick Start

```bash
# Install Portunix
curl -sf https://raw.githubusercontent.com/cassandragargoyle/Portunix/main/scripts/install.sh | bash

# Install development environment
portunix install default

# Start working!
```

## Documentation Sections

### 📚 [Commands](commands/)
Complete reference for all Portunix commands - both core and plugin commands.

### 📖 [Guides](guides/)
Step-by-step tutorials and how-to guides for common tasks.

### 📋 [Release Notes](releases/)
Latest updates and version history.

---

## Features

- **Cross-platform**: Windows, Linux, macOS support
- **Package Management**: Universal installer for development tools
- **Container Integration**: Docker and Podman support
- **Plugin System**: Extensible with gRPC-based plugins
- **AI Integration**: MCP server for AI assistants
- **Self-updating**: Automatic updates from GitHub releases

## Links

- [GitHub Repository](https://github.com/cassandragargoyle/Portunix)
- [Issues & Bug Reports](https://github.com/cassandragargoyle/Portunix/issues)
- [Discussions](https://github.com/cassandragargoyle/Portunix/discussions)
'''
        main_index.write_text(main_index_content, encoding='utf-8')

    # For Hugo Book theme, also create home page (content/_index.md) with redirect
    if theme_installed:
        home_page = CONTENT_DIR / "_index.md"
        if not home_page.exists():
            home_page_content = '''---
title: "Portunix Documentation"
type: docs
---

# Portunix Documentation

**Universal development environment management tool**

Portunix is a cross-platform CLI tool that simplifies the installation and management of development tools, packages, and environments.

[→ Go to Documentation](/Portunix/docs/)
'''
            home_page.write_text(home_page_content, encoding='utf-8')

    print_success("Hugo site structure initialized")
    return True

def parse_command_from_help(help_text: str) -> List[Tuple[str, str]]:
    """Parse commands from help output"""
    print_step("Parse commands from help output..")
    commands = []
    in_commands = False

    for line in help_text.split('\n'):
        if 'Commands:' in line or 'Available Commands:' in line or 'Common commands:' in line:
            in_commands = True
            continue

        if in_commands:
            # Stop at next section
            if line and line[0].isupper() and not line.startswith('  '):
                break

            # Parse command line (format: "  command    description")
            line = line.strip()
            if line and not line.startswith('-'):
                parts = line.split(None, 1)
                if len(parts) >= 1:
                    cmd = parts[0]
                    desc = parts[1] if len(parts) > 1 else ""
                    if cmd and cmd[0].islower():
                        commands.append((cmd, desc))

    return commands

def generate_command_doc(cmd_type: str, cmd: str, desc: str):
    """Generate documentation for a single command"""
    cmd_file = COMMAND_DOCS_DIR / cmd_type / f"{cmd}.md"

    # Get detailed help for the command
    try:
        result = subprocess.run([str(PORTUNIX_BIN), cmd, '--help'],
                              capture_output=True, text=True, timeout=5,
                              encoding='utf-8', errors='replace')
        # Some commands output help to stderr even on success
        cmd_help = result.stdout if result.stdout and result.stdout.strip() else (result.stderr or '')
        if not cmd_help.strip():
            print_warning(f"   No help output for '{cmd}', trying without --help")
            # Try without --help flag as some commands might have different behavior
            result = subprocess.run([str(PORTUNIX_BIN), cmd],
                                  capture_output=True, text=True, timeout=5,
                                  encoding='utf-8', errors='replace')
            cmd_help = result.stdout if result.stdout and result.stdout.strip() else (result.stderr or '')
    except (subprocess.TimeoutExpired, subprocess.SubprocessError) as e:
        cmd_help = f"Error getting help: {e}"
        print_error(f"   Error getting help for '{cmd}': {e}")

    # Check if we're using Hugo Book theme
    hugo_book_theme = (DOCS_SITE_DIR / "themes" / "hugo-book").exists()

    # Create markdown file for the command
    if hugo_book_theme:
        # Hugo Book format - don't use weight for individual command pages
        content = f'''---
title: "{cmd}"
description: "{desc}"
---

# {cmd}

{desc}

## Usage

```bash
portunix {cmd} [options] [arguments]
```

## Full Help

```
{cmd_help}
```

'''
    else:
        # Custom theme format
        content = f'''---
title: "{cmd}"
date: {datetime.now().isoformat()}
draft: false
description: "{desc}"
---

# {cmd}

{desc}

## Usage

```bash
portunix {cmd} [options] [arguments]
```

## Full Help

```
{cmd_help}
```

'''

    # Try to extract subcommands and get their detailed help
    subcommands = parse_command_from_help(cmd_help)
    if subcommands:
        content += "## Subcommands\n\n"
        content += "| Subcommand | Description |\n"
        content += "|------------|-------------|\n"
        for subcmd, subdesc in subcommands:
            content += f"| [{subcmd}](#{subcmd}) | {subdesc} |\n"
        content += "\n"

        # Generate detailed help for each subcommand
        for subcmd, subdesc in subcommands:
            content += f"### {subcmd}\n\n"
            content += f"{subdesc}\n\n"

            # Get subcommand help
            try:
                subcmd_result = subprocess.run(
                    [str(PORTUNIX_BIN), cmd, subcmd, '--help'],
                    capture_output=True, text=True, timeout=5,
                    encoding='utf-8', errors='replace'
                )
                subcmd_help = subcmd_result.stdout if subcmd_result.stdout and subcmd_result.stdout.strip() else (subcmd_result.stderr or '')

                # Check if output is valid help (not an error message)
                # Skip outputs that start with error indicators or are execution results
                is_valid_help = (
                    subcmd_help.strip() and
                    not subcmd_help.strip().startswith('❌') and
                    not subcmd_help.strip().startswith('Error:') and
                    ('Usage:' in subcmd_help or 'Options:' in subcmd_help or 'Examples:' in subcmd_help or '--help' in subcmd_help)
                )

                if is_valid_help:
                    content += f"```\n{subcmd_help}```\n\n"
                else:
                    # Output was not proper help, generate usage hint instead
                    content += f"```bash\nportunix {cmd} {subcmd} --help\n```\n\n"
                    print_info(f"   Subcommand '{cmd} {subcmd}' does not have standard --help output")

            except (subprocess.TimeoutExpired, subprocess.SubprocessError) as e:
                print_warning(f"   Could not get help for subcommand '{cmd} {subcmd}': {e}")
                content += f"```bash\nportunix {cmd} {subcmd} --help\n```\n\n"

    # Check for examples in help text
    if "Examples:" in cmd_help or "Example:" in cmd_help:
        lines = cmd_help.split('\n')
        in_examples = False
        examples = []

        for line in lines:
            if 'Example' in line:
                in_examples = True
                continue
            if in_examples:
                if line and line[0].isupper() and not line.startswith('  '):
                    break
                examples.append(line)

        if examples:
            content += "## Examples\n\n```bash\n"
            content += '\n'.join(examples)
            content += "\n```\n"

    cmd_file.write_text(content, encoding='utf-8')

def discover_core_commands():
    """Discover and parse core commands"""
    print_step("Discovering core commands...")

    # Get main help output
    try:
        result = subprocess.run([str(PORTUNIX_BIN), '--help'],
                              capture_output=True, text=True,
                              encoding='utf-8', errors='replace')
        help_output = result.stdout or ''
    except subprocess.SubprocessError as e:
        print_error(f"Failed to get help output: {e}")
        return

    # Extract commands from help output
    commands = parse_command_from_help(help_output)


    # Create index file for core commands with command list
    commands_file = COMMAND_DOCS_DIR / "core" / "_index.md"

    # Check if we're using Hugo Book theme
    hugo_book_theme = (DOCS_SITE_DIR / "themes" / "hugo-book").exists()

    if hugo_book_theme:
        # Hugo Book format with weight
        index_content = '''---
title: "Core Commands"
weight: 1
bookCollapseSection: false
---

# Core Commands

These are the built-in commands available in Portunix.

## Available Commands

| Command | Description |
|---------|-------------|
'''
    else:
        # Custom theme format
        index_content = f'''---
title: "Core Commands"
date: {datetime.now().isoformat()}
draft: false
---

# Core Commands

These are the built-in commands available in Portunix.

## Available Commands

| Command | Description |
|---------|-------------|
'''

    # Add each command to the index and generate its documentation
    for cmd, desc in commands:
        print_info(f"Processing command: {cmd}")
        # Add to index table
        index_content += f"| [{cmd}]({cmd}/) | {desc} |\n"
        # Generate individual command documentation
        generate_command_doc("core", cmd, desc)

    # Write the complete index file
    commands_file.write_text(index_content, encoding='utf-8')

    print_success("Core commands discovered")

def discover_plugin_commands():
    """Discover plugin commands (Phase 2 - basic implementation)"""
    print_step("Discovering plugin commands...")

    # Check if we're using Hugo Book theme
    hugo_book_theme = (DOCS_SITE_DIR / "themes" / "hugo-book").exists()

    # Create index file for plugin commands
    if hugo_book_theme:
        # Hugo Book format with weight
        index_content = '''---
title: "Plugin Commands"
weight: 2
bookCollapseSection: false
---

# Plugin Commands

These are the commands provided by installed Portunix plugins.

'''
    else:
        # Custom theme format
        index_content = f'''---
title: "Plugin Commands"
date: {datetime.now().isoformat()}
draft: false
---

# Plugin Commands

These are the commands provided by installed Portunix plugins.

'''

    plugin_list = []

    # Check if plugin system is available
    try:
        result = subprocess.run([str(PORTUNIX_BIN), 'plugin', 'list'],
                              capture_output=True, text=True,
                              encoding='utf-8', errors='replace')

        if result.returncode == 0:
            # Try to parse JSON output if jq is available
            try:
                result_json = subprocess.run(
                    [str(PORTUNIX_BIN), 'plugin', 'list', '--format', 'json'],
                    capture_output=True, text=True,
                    encoding='utf-8', errors='replace'
                )
                if result_json.returncode == 0:
                    plugins = json.loads(result_json.stdout)
                    if 'plugins' in plugins:
                        # Add plugin table if we have plugins
                        if plugins['plugins']:
                            index_content += '''## Available Plugins

| Plugin | Description |
|--------|-------------|
'''
                        for plugin in plugins['plugins']:
                            plugin_name = plugin.get('name', 'unknown')
                            plugin_desc = plugin.get('description', 'Plugin for Portunix')
                            print_info(f"Processing plugin: {plugin_name}")

                            # Add to plugin list for index
                            plugin_list.append((plugin_name, plugin_desc))

                            # TODO: Query plugin for its commands via gRPC
                            # For now, create placeholder
                            plugin_file = COMMAND_DOCS_DIR / "plugins" / f"{plugin_name}.md"
                            plugin_content = f'''---
title: "{plugin_name}"
date: {datetime.now().isoformat()}
draft: false
---

# {plugin_name} Plugin

{plugin_desc}

## Commands

Documentation for {plugin_name} plugin commands.

*Note: Plugin command discovery will be implemented in Phase 2.*
'''
                            plugin_file.write_text(plugin_content, encoding='utf-8')

                            # Add plugin to index table
                            index_content += f"| [{plugin_name}]({plugin_name}/) | {plugin_desc} |\n"
                    else:
                        index_content += "\n*No plugins currently installed.*\n"
            except (json.JSONDecodeError, subprocess.SubprocessError):
                print_warning("Could not parse plugin list as JSON")
                index_content += "\n*Plugin information not available.*\n"
        else:
            print_warning("Plugin system not available or no plugins installed")
            index_content += "\n*Plugin system not available or no plugins installed.*\n"
    except subprocess.SubprocessError:
        print_warning("Plugin system not available")
        index_content += "\n*Plugin system not available.*\n"

    # Write the plugin index file
    index_file = COMMAND_DOCS_DIR / "plugins" / "_index.md"
    index_file.write_text(index_content, encoding='utf-8')

    print_success("Plugin command discovery completed")

def generate_release_notes(version: str):
    """Generate release notes page"""
    print_step("Generating release notes page...")

    # Check if we're using Hugo Book theme
    hugo_book_theme = (DOCS_SITE_DIR / "themes" / "hugo-book").exists()

    if hugo_book_theme:
        release_file = DOCS_DIR / "releases" / "_index.md"
        # Hugo Book format with weight
        content = f'''---
title: "Release Notes"
weight: 4
bookFlatSection: true
---

# Portunix Release Notes

Latest updates and version history.

## Current Version: {version}

Released: {datetime.now().strftime("%Y-%m-%d")}

For detailed release notes, please visit the [GitHub Releases](https://github.com/cassandragargoyle/Portunix/releases) page.

## Previous Releases'''
    else:
        release_file = CONTENT_DIR / "releases" / "_index.md"
        # Custom theme format
        content = f'''---
title: "Release Notes"
date: {datetime.now().isoformat()}
draft: false
---

# Release Notes

Current version: **{version}**

## Latest Release

### Version {version}

Released: {datetime.now().strftime("%Y-%m-%d")}

For detailed release notes, please visit the [GitHub Releases](https://github.com/cassandragargoyle/Portunix/releases) page.

## Previous Releases

Visit our [GitHub Releases](https://github.com/cassandragargoyle/Portunix/releases) page for the complete release history.
'''

    release_file.write_text(content, encoding='utf-8')
    print_success("Release notes generated")

def build_hugo_site() -> bool:
    """Build Hugo site"""
    print_step("Building Hugo site...")

    try:
        subprocess.run([HUGO_CMD, '--minify'], cwd=DOCS_SITE_DIR, check=True)
        print_success("Hugo site built successfully")
        print(f"Site generated at: {DOCS_SITE_DIR / 'public/'}")
        return True
    except subprocess.CalledProcessError as e:
        print_error(f"Failed to build Hugo site: {e}")
        return False


def get_package_details(package_name: str) -> Dict:
    """Get detailed package information using 'portunix package info' command"""
    details = {
        'homepage': '',
        'documentation': '',
        'license': '',
        'maintainer': '',
        'dependencies': []
    }

    try:
        result = subprocess.run(
            [str(PORTUNIX_BIN), 'package', 'info', package_name],
            capture_output=True, text=True, timeout=10, encoding='utf-8', errors='replace'
        )
        output = result.stdout if result.stdout and result.stdout.strip() else (result.stderr or '')

        # Parse the output for metadata
        for line in output.split('\n'):
            stripped = line.strip()
            if stripped.startswith('Homepage:'):
                details['homepage'] = stripped.replace('Homepage:', '').strip()
            elif stripped.startswith('Documentation:'):
                details['documentation'] = stripped.replace('Documentation:', '').strip()
            elif stripped.startswith('License:'):
                details['license'] = stripped.replace('License:', '').strip()
            elif stripped.startswith('Maintainer:'):
                details['maintainer'] = stripped.replace('Maintainer:', '').strip()
            elif stripped.startswith('- ') and 'Dependencies:' in output:
                # Parse dependencies (lines starting with "- " after Dependencies section)
                dep = stripped[2:].strip()
                if dep:
                    details['dependencies'].append(dep)

    except (subprocess.TimeoutExpired, subprocess.SubprocessError) as e:
        print_warning(f"   Could not get details for package '{package_name}': {e}")

    return details


def generate_package_catalog():
    """Generate a page listing all available packages with rich metadata"""
    print_step("Generating package catalog...")

    # Check if we're using Hugo Book theme
    hugo_book_theme = (DOCS_SITE_DIR / "themes" / "hugo-book").exists()

    # Get package list output
    try:
        result = subprocess.run(
            [str(PORTUNIX_BIN), 'package', 'list'],
            capture_output=True, text=True, timeout=30, encoding='utf-8', errors='replace'
        )
        package_output = result.stdout if result.stdout and result.stdout.strip() else (result.stderr or '')

        # Parse the output to extract package information
        # Format is:
        # package-name         Short Title
        #                      Long description
        #                      Category: category/path
        packages = []
        current_package = None
        lines = package_output.split('\n')

        for line in lines:
            # Skip empty lines and header lines
            stripped = line.strip()
            if not stripped:
                continue
            if stripped.startswith('Embedded') or stripped.startswith('Package registry') or stripped.startswith('═') or stripped.startswith('📦 Available'):
                continue

            # Check if this is a new package line (starts at column 0, not indented)
            if line and not line.startswith(' '):
                # Skip footer lines like "Total packages: 34"
                if stripped.startswith('Total packages:') or stripped.startswith('Total:'):
                    continue

                # Save previous package
                if current_package:
                    packages.append(current_package)

                # Parse: "package-name         Short Title"
                parts = stripped.split(None, 1)
                if parts and len(parts) >= 1:
                    current_package = {
                        'name': parts[0],
                        'short_title': parts[1] if len(parts) > 1 else '',
                        'description': '',
                        'category': ''
                    }
            elif current_package and stripped.startswith('Category:'):
                # Category line
                current_package['category'] = stripped.replace('Category:', '').strip()
            elif current_package and stripped:
                # Description line (indented, not category)
                if current_package['description']:
                    current_package['description'] += ' ' + stripped
                else:
                    current_package['description'] = stripped

        if current_package:
            packages.append(current_package)

        # Fetch detailed metadata for each package
        print_info(f"Fetching detailed metadata for {len(packages)} packages...")
        for pkg in packages:
            pkg_name = pkg.get('name', '')
            if pkg_name:
                print_info(f"   Getting details for: {pkg_name}")
                details = get_package_details(pkg_name)
                pkg['homepage'] = details['homepage']
                pkg['documentation'] = details['documentation']
                pkg['license'] = details['license']
                pkg['maintainer'] = details['maintainer']
                pkg['dependencies'] = details['dependencies']

    except (subprocess.TimeoutExpired, subprocess.SubprocessError) as e:
        print_error(f"Failed to get package list: {e}")
        packages = []
        package_output = f"Error getting package list: {e}"

    # Create the package catalog page
    catalog_file = DOCS_DIR / "packages.md"

    if hugo_book_theme:
        content = '''---
title: "Available Packages"
weight: 5
---

# Available Packages

Complete catalog of software packages available for installation via Portunix.

## Quick Install

```bash
portunix install <package-name>
portunix install <package-name> --variant <variant>
portunix install <package-name> --dry-run
```

## Package Catalog

'''
    else:
        content = f'''---
title: "Available Packages"
date: {datetime.now().isoformat()}
draft: false
---

# Available Packages

Complete catalog of software packages available for installation via Portunix.

## Quick Install

```bash
portunix install <package-name>
portunix install <package-name> --variant <variant>
portunix install <package-name> --dry-run
```

## Package Catalog

'''

    # Add packages grouped by category
    if packages:
        # Group packages by category
        categories = {}
        for pkg in packages:
            cat = pkg.get('category', 'other')
            if not cat:
                cat = 'other'
            if cat not in categories:
                categories[cat] = []
            categories[cat].append(pkg)

        # Sort categories alphabetically
        sorted_categories = sorted(categories.keys())

        # Category display names and icons (using monochrome Unicode symbols)
        category_info = {
            'development/languages': ('Programming Languages', '◈'),
            'development/tools': ('Development Tools', '⚙'),
            'development/editors': ('Editors & IDEs', '▣'),
            'development/build-tools': ('Build Tools', '⛭'),
            'development/ai-tools': ('AI Tools', '◉'),
            'development/shells': ('Shells', '▶'),
            'development/libraries': ('Libraries', '▤'),
            'infrastructure/automation': ('Infrastructure Automation', '⚡'),
            'infrastructure/virtualization': ('Virtualization', '▢'),
            'infrastructure/web-servers': ('Web Servers', '◎'),
            'system/package-managers': ('Package Managers', '▦'),
            'system/browsers': ('Web Browsers', '◌'),
            'security/vpn': ('VPN & Security', '▧'),
            'security/firewall': ('Firewall', '◆'),
            'security/certificates': ('Certificates', '▨'),
            'security/intrusion-prevention': ('Security Tools', '◇'),
            'other': ('Other', '○'),
        }

        for cat in sorted_categories:
            cat_packages = categories[cat]
            display_name, icon = category_info.get(cat, (cat.replace('/', ' / ').title(), '○'))

            content += f"### {icon} {display_name}\n\n"

            for pkg in cat_packages:
                name = pkg.get('name', '')
                title = pkg.get('short_title', '')
                desc = pkg.get('description', '')
                homepage = pkg.get('homepage', '')
                documentation = pkg.get('documentation', '')
                license_info = pkg.get('license', '')
                maintainer = pkg.get('maintainer', '')

                # Use title if available, otherwise use description
                display_desc = title if title else desc

                # Package name as header
                content += f"#### `{name}`\n\n"
                content += f"{display_desc}\n\n"

                # Build links line
                links = []
                if homepage:
                    links.append(f"[Homepage]({homepage})")
                if documentation:
                    links.append(f"[Documentation]({documentation})")
                if license_info:
                    links.append(f"License: {license_info}")
                if maintainer:
                    links.append(f"Maintainer: {maintainer}")

                if links:
                    content += " | ".join(links) + "\n\n"

                content += "---\n\n"

        content += f"\n**Total packages available: {len(packages)}**\n"
    else:
        # Fallback to raw output
        content += "```\n"
        content += package_output
        content += "\n```\n"

    content += '''
## Getting Package Details

To see detailed information about a specific package:

```bash
portunix package info <package-name>
```

## Searching Packages

To search for packages by name or description:

```bash
portunix package search <query>
```
'''

    catalog_file.write_text(content, encoding='utf-8')
    print_success(f"Package catalog generated with {len(packages)} packages")


def run_local_server():
    """Run local Hugo server"""
    print_step("Starting local Hugo server...")
    print("Server will be available at: http://localhost:1313")
    print("Press Ctrl+C to stop")

    try:
        subprocess.run([HUGO_CMD, 'server', '-D'], cwd=DOCS_SITE_DIR)
    except KeyboardInterrupt:
        print("\nServer stopped")

def main():
    """Main execution"""
    parser = argparse.ArgumentParser(description='Portunix Documentation Generator')
    parser.add_argument('version', nargs='?', default='latest',
                       help='Version to generate documentation for')
    parser.add_argument('--serve', action='store_true',
                       help='Run local Hugo server')
    parser.add_argument('--build-only', action='store_true',
                       help='Build site without deploying')

    args = parser.parse_args()
    version = args.version

    print("=" * 48)
    print("Portunix Documentation Generator")
    print(f"Version: {version}")
    print("=" * 48)
    print()

    # Check dependencies first
    if not check_dependencies():
        return 1

    if args.serve:
        if not init_hugo_site(version):
            return 1
        discover_core_commands()
        discover_plugin_commands()
        generate_package_catalog()
        generate_release_notes(version)
        run_local_server()
    else:
        # Default or build-only mode
        if not init_hugo_site(version):
            return 1
        discover_core_commands()
        discover_plugin_commands()
        generate_package_catalog()
        generate_release_notes(version)

        if not args.build_only or not build_hugo_site():
            return 1

    print()
    print_success("Documentation generation completed successfully!")
    return 0

if __name__ == '__main__':
    # Ensure UTF-8 encoding on Windows
    if sys.platform == 'win32':
        import io
        sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
        sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')

    try:
        sys.exit(main())
    except Exception as e:
        print_error(f"Error occurred: {e}")
        sys.exit(1)