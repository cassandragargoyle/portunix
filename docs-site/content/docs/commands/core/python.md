---
title: "python"
description: "Python development tools"
---

# python

Python development tools

## Usage

```bash
portunix python [options] [arguments]
```

## Full Help

```
Usage: portunix python [subcommand]

Compatibility: Python 3.x only (Python 2 is not supported)

Python Development Commands:

Project Setup:
  init                         - Initialize project: create ./.venv, install deps
  init --force                 - Recreate existing venv
  init --python <version>      - Specify Python version (e.g., 3.11)

Virtual Environment Management:
  venv create <name>           - Create centralized venv (~/.portunix/python/venvs/)
  venv create --local          - Create project-local venv (./.venv)
  venv create --path <dir>     - Create venv at custom location
  venv list                    - List all virtual environments
  venv list --group-by-version - Group venvs by Python version
  venv exists <name>           - Check if venv exists (exit code 0/1)
  venv info                    - Show ./.venv details (auto-detect)
  venv info --verbose          - Include component versions (pip, setuptools)
  venv info --json             - Output in JSON format (implies --verbose)
  venv delete <name>           - Remove virtual environment
  venv delete --local          - Remove ./.venv
  venv activate <name>         - Show activation command
  venv scan [path]             - Discover all venvs in directory

Package Management:
  pip install <package>        - Install package (auto-detects ./.venv)
  pip install -r requirements.txt - Install from requirements file
  pip install <pkg> --local    - Install to ./.venv explicitly
  pip install <pkg> --venv <n> - Install to centralized venv
  pip uninstall <package>      - Remove package
  pip list                     - List installed packages
  pip freeze                   - Generate requirements.txt

Build & Distribution:
  build exe <script.py>        - Build standalone executable with PyInstaller
  build freeze <script.py>     - Build with cx_Freeze
  build wheel                  - Build wheel distribution package
  build sdist                  - Build source distribution package

Options:
  --local                      - Use project-local venv (./.venv)
  --path <path>                - Use venv at custom location
  --venv <name>                - Use centralized venv by name
  --global                     - Operate on system Python

```

