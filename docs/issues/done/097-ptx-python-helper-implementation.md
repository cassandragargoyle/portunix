# Issue #97: PTX-Python Helper Implementation

**Status**: ✅ Implemented
**Priority**: High
**Type**: Feature
**Created**: 2025-10-23
**Updated**: 2026-05-19
**Closed**: 2026-05-19
**Architecture Decision Record**: TBD

## Implementation Status

- ✅ **Phase 1**: Complete (Virtual Environment Management)
  - Commit: `27d5889` (2025-10-23)
  - Tests: 18/18 passed
  - Acceptance: `docs/testing/internal/acceptance-097-phase1.md` ✅ PASS

- ✅ **Phase 2**: Complete (Build & Distribution Support)
  - Commit: `7b0cee2` (2025-10-23)
  - Tests: 15/15 passed (Integration: 10/10)
  - Acceptance: `docs/testing/internal/acceptance-097-phase2.md` ✅ PASS

- ✅ **Phase 3**: Complete (Code Quality & Development Tools)
  - Commit: `755ccf9` (2026-05-19)
  - Tests: 11/12 PASS + 1 CONDITIONAL (F-097-P3-001: PEP 668 UX, non-blocking)
  - Acceptance: `docs/testing/internal/acceptance-097-phase3.md` ✅ PASS

- ✅ **Phase 4**: Complete (Advanced Features & Multi-Version Management)
  - Commit: `755ccf9` (2026-05-19)
  - Tests: 14/14 PASS
  - Acceptance: `docs/testing/internal/acceptance-097-phase4.md` ✅ PASS

## Problem Statement

Portunix currently provides Python installation capabilities through its package management system, but lacks comprehensive Python development utilities that are essential for modern Python workflows. Developers need tools for:

- Virtual environment management (venv, virtualenv)
- Package installation and dependency management (pip, requirements.txt)
- Python code compilation to executable binaries (PyInstaller, cx_Freeze)
- Code quality checks (syntax validation, linting, formatting)
- Testing and type checking support
- Multi-version Python management

Currently, these operations must be performed manually using native Python tools, which doesn't align with Portunix's unified development environment philosophy.

## Proposed Solution

Implement a dedicated `ptx-python` helper binary following the established helper binary architecture pattern used by `ptx-virt`, `ptx-ansible`, and `ptx-container`.

### Key Components

#### 1. Helper Binary Architecture
- **`ptx-python`** - Dedicated Python utilities helper binary
- **`portunix`** - Main dispatcher routing python commands

#### 2. Core Functionality Groups
- **Virtual Environment Management** - Create, activate, manage venvs
- **Package Management** - pip operations, requirements.txt handling
- **Build & Distribution** - Compile to exe, package for distribution
- **Code Quality** - Syntax check, linting, formatting
- **Testing & Type Checking** - pytest integration, mypy support
- **Multi-Version Support** - Switch between installed Python versions

## Implementation Phases

### Phase 1: Foundation & Virtual Environment Management ✅ COMPLETE
**Goal**: Basic helper infrastructure with venv management
**Status**: ✅ Implemented (Commit: `27d5889`)
**Acceptance**: ✅ PASS (18/18 tests, docs/testing/internal/acceptance-097-phase1.md)

#### Tasks:
1. **Helper Binary Infrastructure**
   - [x] Create `ptx-python` binary project structure
   - [x] Implement basic binary skeleton with version and help commands
   - [x] Set up build pipeline for `ptx-python` in main Makefile
   - [x] Create integration tests for binary communication

2. **Main Binary Dispatcher**
   - [x] Add `python` command group to main `portunix` binary
   - [x] Implement `portunix python venv` → `ptx-python` delegation
   - [x] Implement `portunix python pip` → `ptx-python` delegation
   - [x] Add `portunix python check` for helper availability
   - [x] Implement error handling when `ptx-python` not found

3. **Virtual Environment Management**
   - [x] Implement `venv create <name>` - create new virtual environment
   - [x] Implement `venv list` - list all virtual environments with Python versions
   - [x] Implement `venv list --group-by-version` - group venvs by Python version
   - [x] Implement `venv exists <name>` - check if specific venv exists (exit code 0/1)
   - [x] Implement `venv scan [path]` - discover all venvs in specified directory with versions (TODO marker for future enhancement)
   - [x] Implement `venv activate <name>` - activate specific venv (manual activation instructions)
   - [x] Implement `venv delete <name>` - remove virtual environment
   - [x] Implement `venv info <name>` - show venv details (Python version, packages)
   - [x] Support custom venv location via `--path` parameter (basic support)
   - [x] Cross-platform venv activation script generation (Windows/Linux)

4. **Basic Package Management**
   - [x] Implement `pip install <package>` - install package to active/specified venv
   - [x] Implement `pip uninstall <package>` - remove package (TODO marker)
   - [x] Implement `pip list` - list installed packages
   - [x] Implement `pip freeze` - generate requirements.txt (Phase 2)
   - [x] Support `--venv <name>` parameter for targeting specific venv
   - [x] Support `--global` flag for system Python operations (partial)

#### Success Criteria:
- [x] `ptx-python` binary builds and executes independently
- [x] Virtual environments can be created, listed, and deleted
- [x] Venv existence checking works correctly with appropriate exit codes
- [x] Venv scanning discovers all virtual environments in specified directories (basic impl)
- [x] Packages can be installed and managed in venvs
- [x] Cross-platform compatibility (Windows/Linux) - Linux tested, Windows code implemented
- [x] Error handling provides clear user guidance

### Phase 2: Build & Distribution Support ✅ COMPLETE
**Goal**: Python application compilation and packaging
**Status**: ✅ Implemented (Commit: `7b0cee2`)
**Acceptance**: ✅ PASS (15/15 manual tests, 10/10 integration tests, docs/testing/internal/acceptance-097-phase2.md)

#### Tasks:
1. **PyInstaller Integration**
   - [x] Implement `build exe <script.py>` - compile Python to standalone exe
   - [x] Support `--onefile` mode for single executable
   - [x] Support `--console` / `--windowed` modes
   - [x] Support `--icon <icon.ico>` for custom application icon
   - [x] Support `--name <name>` for custom output name
   - [x] Auto-detect and bundle dependencies (PyInstaller automatic)

2. **cx_Freeze Support (Alternative)**
   - [x] Implement `build freeze <script.py>` - alternative compiler (command structure)
   - [x] Support Windows installer generation (.msi) (basic support)
   - [x] Support Linux package generation (.deb, .rpm) (basic support)
   - [x] Cross-platform build configuration

3. **Requirements Management**
   - [x] Implement `pip install -r requirements.txt` - batch install
   - [x] Implement `pip freeze` - generate requirements.txt from current venv
   - [x] Support `requirements-dev.txt` for development dependencies (via pip -r)
   - [x] Implement dependency resolution and conflict detection (pip automatic)

4. **Distribution Tools**
   - [x] Implement `build wheel` - create wheel distribution
   - [x] Implement `build sdist` - create source distribution
   - [x] Support setup.py and pyproject.toml configurations (python -m build)
   - [x] Basic PyPI publishing preparation (wheel/sdist generation)

#### Success Criteria:
- [x] Python scripts successfully compile to standalone executables
- [x] Executables run on target platforms without Python installation
- [x] Requirements.txt workflow functions correctly
- [x] Distribution packages (wheel, sdist) build successfully

#### Implementation Details:
**New Files:**
- `src/helpers/ptx-python/build_manager.go` (362 lines) - Build operations manager
- `test/integration/issue_097_python_phase2_test.go` (289 lines) - Integration tests

**Modified Files:**
- `src/helpers/ptx-python/main.go` (+263 lines) - Build command handlers, requirements support
- `src/helpers/ptx-python/venv_manager.go` (+19 lines) - InstallRequirements method

**Test Results:**
- Manual Testing: 15/15 scenarios passed
- Integration Testing: 10/10 steps passed (5.46s)
- Built executable verified working on Linux

### Phase 3: Code Quality & Development Tools 🔄 CODE COMPLETE
**Goal**: Syntax checking, linting, formatting, testing
**Status**: 🔄 Code complete, awaiting acceptance protocol

#### Tasks:
1. **Syntax Validation**
   - [x] Implement `check syntax <file/directory>` - Python syntax validation
   - [x] Support recursive directory scanning
   - [x] Detailed error reporting with line numbers (via ast.parse)
   - [x] Integration with Python AST parser

2. **Linting Support**
   - [x] Implement `lint <file/directory>` - code quality analysis
   - [x] Support multiple linters (pylint, flake8, ruff) via `--linter` flag
   - [x] Configurable rule sets and ignore patterns (passed through to linter)
   - [x] HTML/JSON report generation via `--format` / `--output`
   - [ ] Auto-fix suggestions — deferred (use linter's native `--fix` via extra args)

3. **Code Formatting**
   - [x] Implement `format <file/directory>` - automatic code formatting
   - [x] Support black formatter (primary, default)
   - [x] Support autopep8 (alternative) via `--formatter autopep8`
   - [x] Support `--check` mode for CI/CD (no changes)
   - [x] Support `pyproject.toml` configuration (honored by black/autopep8 natively)

4. **Type Checking**
   - [x] Implement `typecheck <file/directory>` - mypy integration
   - [x] Support type hints validation
   - [x] Incremental type checking for performance (mypy default)
   - [ ] Integration with IDE workflows — deferred to separate issue

5. **Testing Support**
   - [x] Implement `test` - pytest wrapper
   - [x] Support test discovery and filtering
   - [x] Coverage report generation (`--coverage` via pytest-cov)
   - [x] Support `--watch` mode for continuous testing (via pytest-watch)
   - [x] Integration with common testing frameworks (pytest standard)

#### Success Criteria:
- [x] Syntax errors are detected and reported accurately
- [x] Linting identifies code quality issues
- [x] Code formatting consistently formats Python code
- [x] Type checking validates type hints correctly
- [x] Tests execute with detailed reporting

#### Implementation Details:
**New Files:**
- `src/helpers/ptx-python/quality_manager.go` (~290 lines) — QualityManager
  with CheckSyntax / Lint / Format / Typecheck / Test methods, auto-install
  of pylint/flake8/ruff/black/autopep8/mypy/pytest/pytest-cov/pytest-watch.

**Modified Files:**
- `src/helpers/ptx-python/main.go` — handlers `handleLintCommand`,
  `handleFormatCommand`, `handleTypecheckCommand`, `handleTestCommand`,
  extended `handleCheckCommand` for `check syntax`, plus
  `parseQualityArgs` helper for `--local`/`--path`/`--venv` parsing.

### Phase 4: Advanced Features & Multi-Version Management 🔄 CODE COMPLETE
**Goal**: Python version management and advanced workflows
**Status**: 🔄 Code complete, awaiting acceptance protocol

#### Tasks:
1. **Python Version Switching**
   - [x] Implement `version list` - show all installed Python versions
   - [x] Implement `version use <version>` - pin via `.python-version`
         (pyenv/uv/hatch compatible — chosen over PATH mutation for safety)
   - [x] Implement `version detect` - detect project Python requirements
   - [x] Support per-directory Python version configuration (`.python-version`)
   - [x] Integration with Portunix package system (scans `~/.portunix/python/installs/`)

2. **Environment Variables & Path Management**
   - [x] Implement `env show` - display Python environment configuration
   - [x] Implement `env set <var> <value>` - emit shell snippet for env var
   - [ ] Automatic PATH manipulation for active venv — by design, helper
         binaries cannot mutate the parent shell; venv `activate` script is
         the standard mechanism
   - [ ] Support `.env` file integration — deferred to separate issue

3. **Project Scaffolding**
   - [x] Implement `init project <name>` - new Python project structure
         (subcommand under `init` to keep `init` = "init venv" backward compat)
   - [x] Support templates (console, web, library, package)
   - [x] Auto-generate pyproject.toml, README.md (no setup.py — modern PEP 621)
   - [x] Create basic test structure with pytest (tests/test_smoke.py)
   - [ ] Initialize git repository — deferred; `.gitignore` is generated

4. **Dependency Scanning & Security**
   - [x] Implement `audit` - scan dependencies for known vulnerabilities
   - [x] Integration with pip-audit (primary), safety (fallback)
   - [x] Generate security reports (pip-audit native JSON via extra args)
   - [ ] Suggest package updates — handled by pip-audit native output

5. **IDE Integration**
   - [ ] Generate VS Code configuration — **deferred** to separate issue
         (`.vscode/settings.json` generation is ~30 lines; low priority)
   - [ ] Generate PyCharm configuration — **deferred** to separate issue
         (`.idea/` is complex XML, high risk for limited value)
   - [ ] Jupyter notebook workflows — **deferred** to separate issue
   - [ ] MCP server integration — tracked in #073 and #004

#### Success Criteria:
- [x] Multiple Python versions can be listed and selected (`.python-version`)
- [x] Project initialization creates complete working structure
- [x] Security audits identify vulnerable dependencies
- [ ] IDE configurations work out-of-box — explicitly deferred above

#### Implementation Details:
**New Files:**
- `src/helpers/ptx-python/version_manager.go` (~140 lines) — VersionManager
  with ListInstalls / DetectProjectVersion / WriteProjectVersion.
  Discovers `python3.x` in PATH plus `~/.portunix/python/installs/`.
  Skips non-runnable stubs (e.g. Windows Store python.exe).
- `src/helpers/ptx-python/env_manager.go` (~340 lines) — EnvManager with
  ShowEnv / SetEnv / InitProject (4 templates) / Audit (pip-audit, safety).

**Modified Files:**
- `src/helpers/ptx-python/main.go` — handlers `handleVersionCommand`,
  `handleEnvCommand`, `handleAuditCommand`, `handleInitProject`, plus the
  `init project <name>` branch in `handleInitCommand`.

## Command Structure

```bash
# Virtual Environment Management
portunix python venv create myproject
portunix python venv create myproject --python 3.11
portunix python venv list                      # List all venvs with Python versions
portunix python venv list --group-by-version   # Group venvs by Python version
portunix python venv exists myproject          # Check if venv exists (exit 0 if exists)
portunix python venv scan                      # Discover venvs in current directory
portunix python venv scan /path/to/dir         # Discover venvs in specific directory
portunix python venv activate myproject
portunix python venv delete myproject
portunix python venv info myproject            # Show detailed info including Python version

# Package Management
portunix python pip install requests
portunix python pip install -r requirements.txt
portunix python pip install --venv myproject numpy
portunix python pip uninstall package-name
portunix python pip list
portunix python pip freeze > requirements.txt
portunix python pip install --global pipx  # System-wide installation

# Build & Distribution
portunix python build exe main.py
portunix python build exe main.py --onefile --icon app.ico
portunix python build exe main.py --windowed --name MyApp
portunix python build freeze main.py --target-version 3.11
portunix python build wheel
portunix python build sdist

# Code Quality
portunix python check syntax .
portunix python lint src/
portunix python lint . --format json --output report.json
portunix python format .
portunix python format src/ --check  # CI/CD mode, no changes
portunix python typecheck src/
portunix python test
portunix python test --coverage
portunix python test --watch

# Version Management
portunix python version list
portunix python version use 3.11
portunix python version detect  # From pyproject.toml or .python-version

# Environment & Configuration
portunix python env show
portunix python env set PYTHONPATH /custom/path
portunix python init myproject --template web
portunix python audit  # Security scan

# Check helper availability
portunix python check
```

## Example Workflows

### Workflow 1: Create New Project with Venv
```bash
# Create project structure
portunix python init mywebapp --template web

# Create virtual environment
portunix python venv create mywebapp

# Install dependencies
portunix python pip install -r requirements.txt --venv mywebapp

# Run tests
portunix python test --venv mywebapp
```

### Workflow 2: Discover and Manage Existing Virtual Environments
```bash
# List all managed venvs with Python versions
portunix python venv list

# Output:
# Virtual Environments in ~/.portunix/python/venvs/:
#   myproject    (Python 3.11.5, 23 packages, 185 MB)
#   webapp       (Python 3.11.5, 67 packages, 340 MB)
#   legacy-tool  (Python 3.8.10, 12 packages, 95 MB)
#   data-science (Python 3.10.8, 89 packages, 512 MB)

# Group venvs by Python version
portunix python venv list --group-by-version

# Output:
# Virtual Environments grouped by Python version:
#
# Python 3.11.5 (2 environments):
#   - myproject    (23 packages, 185 MB)
#   - webapp       (67 packages, 340 MB)
#
# Python 3.10.8 (1 environment):
#   - data-science (89 packages, 512 MB)
#
# Python 3.8.10 (1 environment):
#   - legacy-tool  (12 packages, 95 MB)

# Scan for all venvs in projects directory
portunix python venv scan ~/projects

# Output:
# Found 3 virtual environments in /home/user/projects:
#   - project1/.venv (Python 3.11.5, 45 packages)
#   - project2/venv (Python 3.10.8, 12 packages)
#   - legacy-app/.virtualenv (Python 3.8.10, 67 packages)

# Check if specific venv exists before operations
if portunix python venv exists myproject; then
  echo "Virtual environment exists, installing packages..."
  portunix python pip install requests --venv myproject
else
  echo "Creating new virtual environment..."
  portunix python venv create myproject
fi

# Get detailed info about specific venv
portunix python venv info project1/.venv

# Output:
# Virtual Environment: project1/.venv
# Python Version: 3.11.5
# Created: 2025-09-15 14:23:00
# Location: /home/user/projects/project1/.venv
# Packages: 45 installed
# Size: 245 MB
```

### Workflow 3: Build Standalone Executable
```bash
# Create venv and install dependencies
portunix python venv create myapp
portunix python pip install -r requirements.txt --venv myapp

# Build standalone exe
portunix python build exe src/main.py --onefile --name MyApp --icon assets/icon.ico

# Result: dist/MyApp.exe (Windows) or dist/MyApp (Linux)
```

### Workflow 4: Code Quality Pipeline
```bash
# Check syntax
portunix python check syntax src/

# Format code
portunix python format src/

# Run linter
portunix python lint src/ --format json --output lint-report.json

# Type checking
portunix python typecheck src/

# Run tests with coverage
portunix python test --coverage --output coverage-report.html
```

### Workflow 5: Security Audit
```bash
# Scan for vulnerabilities
portunix python audit

# Output:
# ⚠️  Found 2 vulnerabilities:
#   - requests 2.25.0 (CVE-2021-XXXXX) → Update to 2.31.0
#   - urllib3 1.26.5 (CVE-2021-XXXXX) → Update to 2.0.7
```

## Benefits

### Positive Consequences
- **Unified Python Workflow**: All Python operations through consistent Portunix interface
- **Cross-Platform Consistency**: Same commands work on Windows and Linux
- **Simplified Distribution**: Easy Python to exe compilation without complex PyInstaller setup
- **Code Quality Integration**: Built-in linting, formatting, and testing support
- **Beginner Friendly**: Simplifies Python development setup for new developers
- **CI/CD Ready**: Commands designed for automation and pipeline integration

### Challenges to Address
- **Python Version Compatibility**: Must handle Python 2.7 through 3.12+
- **Platform Differences**: Windows/Linux have different venv activation mechanisms
- **Dependency Complexity**: PyInstaller/cx_Freeze have complex dependency detection
- **Tool Installation**: Helper must install/manage linting and build tools

## Technical Considerations

### Python Detection
- Detect installed Python versions via Portunix package registry
- Support custom Python installations outside Portunix
- Handle multiple Python versions (3.8, 3.9, 3.10, 3.11, 3.12+)

### Virtual Environment Storage
- Default location: `~/.portunix/python/venvs/`
- Support custom locations via `--path` parameter
- Store venv metadata (Python version, creation date, packages)

### Build Tool Installation
- PyInstaller installation: `portunix python install-tool pyinstaller`
- Auto-install build dependencies when first used
- Cache build tools in `~/.portunix/python/tools/`

### Integration with Portunix Ecosystem
- Use Portunix logging system for consistent output
- Integrate with Portunix MCP server for AI assistance
- Follow Portunix helper binary architecture patterns

## Success Criteria

### Phase 1 (Foundation)
- [ ] `ptx-python` helper binary successfully builds and deploys
- [ ] Virtual environments can be created, activated, and managed
- [ ] Pip package installation works correctly in venvs
- [ ] Cross-platform compatibility verified (Windows/Linux)

### Phase 2 (Build & Distribution)
- [ ] Python scripts compile to standalone executables via PyInstaller
- [ ] Executables run without Python runtime on target systems
- [ ] Requirements.txt workflow functions end-to-end
- [ ] Distribution packages (wheel/sdist) build successfully

### Phase 3 (Code Quality)
- [ ] Syntax checking accurately detects Python errors
- [ ] Linting provides actionable code quality feedback
- [ ] Code formatting consistently formats Python codebases
- [ ] Type checking validates type hints with mypy
- [ ] Testing support executes pytest with coverage reports

### Phase 4 (Advanced Features)
- [ ] Multiple Python versions managed and switched seamlessly
- [ ] Project initialization creates complete working projects
- [ ] Security audits identify vulnerable dependencies
- [ ] IDE configuration generation works for VS Code and PyCharm

## Dependencies

### Core Dependencies
- Python installation via Portunix package system (already implemented)
- Helper binary architecture from [ADR-014](../../adr/014-git-dispatcher-with-python-distribution.md)
- Portunix logging system from [#052](052-logging-system-implementation.md)

### Python Tool Dependencies (Installed by ptx-python)
- **PyInstaller** - Python to exe compilation
- **cx_Freeze** - Alternative compiler (optional)
- **pylint** - Code linting
- **flake8** - Alternative linter
- **ruff** - Modern fast linter
- **black** - Code formatter
- **autopep8** - Alternative formatter
- **mypy** - Type checking
- **pytest** - Testing framework
- **pytest-cov** - Coverage reporting
- **pip-audit** - Security vulnerability scanning
- **safety** - Alternative security scanner

## Related Issues

### Core Infrastructure
- [#051](../051-git-dispatcher-python-distribution-architecture.md) - Git-like Dispatcher Architecture
- [#052](052-logging-system-implementation.md) - Logging System Implementation

### Package Management
- Python installation (already implemented in Portunix package system)

### Helper Binary Examples
- [#056](056-ansible-infrastructure-as-code-integration.md) - Ansible Helper (reference architecture)
- [#068](../068-main-binary-ptx-virt-helper-integration.md) - Virt Helper Integration (reference)

### Potential Future Integration
- [#073](../073-ptx-prompting-helper-implementation.md) - PTX-Prompting Helper (AI-assisted coding)
- [#004](004-mcp-server-ai-integration.md) - MCP Server AI Integration

## Labels

`enhancement`, `helper-binary`, `python`, `development-tools`, `build-automation`, `code-quality`

## Notes

### Why Not Direct Python Tool Usage?
While developers can use native Python tools directly (venv, pip, PyInstaller), the `ptx-python` helper provides:

1. **Unified Interface**: Consistent with Portunix command structure
2. **Cross-Platform Abstraction**: Handles Windows/Linux differences automatically
3. **Simplified Workflows**: Reduces command complexity for common operations
4. **Integration**: Works with Portunix package system and MCP server
5. **Beginner Friendly**: Lowers barrier to entry for Python development
6. **Automation Ready**: Designed for CI/CD and scripting

### Python Installation Remains in Portunix Core
The installation of Python itself (e.g., `portunix install python`) remains in the main Portunix package system. The `ptx-python` helper focuses on **working with** installed Python versions, not installing Python itself.

### Comparison with Other Helper Binaries

| Helper | Purpose | Complexity | Integration Level |
|--------|---------|------------|-------------------|
| `ptx-virt` | VM management | High | Core functionality |
| `ptx-ansible` | Infrastructure as Code | High | Optional workflow |
| `ptx-container` | Container operations | Medium | Core functionality |
| `ptx-python` | Python dev tools | Medium | Development workflow |

### Target Audience
- **Beginner Developers**: Simplified Python workflow without complex tool setup
- **DevOps Engineers**: Consistent Python operations in automation scripts
- **CI/CD Pipelines**: Standardized commands for build and test automation
- **Cross-Platform Teams**: Unified commands work on Windows and Linux

### Phased Rollout Strategy
The phased implementation allows teams to:
1. Start with basic venv management (Phase 1)
2. Add build capabilities as needed (Phase 2)
3. Integrate code quality tools when ready (Phase 3)
4. Adopt advanced features progressively (Phase 4)

This approach prevents overwhelming users while delivering immediate value.
