# Issue #138: PTX-Python Project-Local Virtual Environment Support

## Status

🚧 Partial (Phase 1 complete, Phase 2-3 pending)

## Priority

High

## Type

Enhancement

## Labels

enhancement, helper-binary, ptx-python, python, virtual-environment, developer-experience

## Summary

Extend `ptx-python` helper to support project-local virtual environments (`./.venv`) in addition to the existing centralized venv management. This enables users to replace verbose platform-specific setup scripts with simple cross-platform Portunix commands.

## Problem Statement

Currently, when setting up a Python project, users must write 50+ line platform-specific scripts to:

1. Check Python availability
2. Create virtual environment in project directory
3. Activate and install dependencies

Example of current approach (PowerShell):

```powershell
$VenvDir = Join-Path $ProjectDir "venv"
python -m venv $VenvDir
& $VenvDir\Scripts\Activate.ps1
pip install -r requirements.txt
```

This is verbose, error-prone, and different for each platform.

## Proposed Solution

### New Commands

```bash
# High-level project initialization
portunix python init                         # Auto-setup: create ./.venv, install deps
portunix python init --force                 # Recreate existing venv
portunix python init --python 3.11           # Specify Python version

# Explicit local venv management
portunix python venv create --local          # Creates ./.venv
portunix python venv create --path ./custom  # Custom location
portunix python venv info --local            # Show ./.venv details
portunix python venv delete --local          # Remove ./.venv

# Package management for local venv
portunix python pip install flask --local
portunix python pip install -r requirements.txt --local
portunix python pip list --local
portunix python pip freeze --local
```

### Behavior of `python init`

1. Check Python availability → suggest `portunix install python` if missing
2. Detect `requirements.txt` or `pyproject.toml`
3. If `./.venv` exists → prompt for recreation (or `--force`)
4. Create `./.venv` using system Python
5. **Upgrade pip** using `python -m pip install --upgrade pip`
6. Install dependencies from requirements file
7. Display platform-appropriate activation command

### Pip Self-Upgrade Problem

**Critical**: Direct pip calls fail when upgrading pip itself:

```text
ERROR: To modify pip, please run the following command:
.../.venv/Scripts/python.exe -m pip install --upgrade pip
```

**Solution**: All pip operations must use `python -m pip` pattern:

```go
// CORRECT
exec.Command(pythonExe, "-m", "pip", "install", "flask")

// WRONG - fails on pip self-upgrade
exec.Command(pipExe, "install", "flask")
```

## Phase 2: Script Generation

### Problem

Even with `portunix python init`, projects still need setup scripts for:

- CI/CD pipelines
- New developer onboarding
- Automated deployment
- Documentation (showing users how to set up the project)

Currently, AI assistants generate 40+ line platform-specific scripts manually. Example:

```powershell
# Setup virtual environment for Project X (Windows)
$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

Write-Host "Setting up virtual environment..." -ForegroundColor Cyan
Set-Location $ProjectDir

$VenvDir = Join-Path $ProjectDir ".venv"
if (Test-Path $VenvDir) {
    Write-Host "Virtual environment already exists" -ForegroundColor Yellow
    $response = Read-Host "Recreate? (y/N)"
    if ($response -eq 'y') {
        portunix python init --force --path .venv
    } else {
        portunix python pip install -r requirements.txt --path .venv
    }
} else {
    portunix python venv create --path .venv
    portunix python pip install -r requirements.txt --path .venv
}

Write-Host "Setup complete!" -ForegroundColor Green
```

This is repetitive - the logic is identical across projects.

### Proposed Solution

```bash
# Generate setup scripts during init
portunix python init --generate-scripts

# Generate scripts for existing project
portunix python generate-scripts
portunix python generate-scripts --dir ./tools       # Custom output directory
portunix python generate-scripts --project "My App"  # Custom project name in comments
```

**Note:** Script names are fixed (`setup`, `activate`, `start`) for consistency across projects.

### Project Name Detection

The `--project` parameter is optional. If not provided, project name is auto-detected:

```text
Detection priority:
1. --project "Name"           → Use explicit name
2. pyproject.toml [project]   → name = "my-project"
3. setup.py                   → name="my-project"
4. package.json               → "name": "my-project"
5. Directory name             → Current directory name as fallback
```

### Generated Files

| File | Platform | Description |
| ---- | -------- | ----------- |
| `scripts/setup.ps1` | Windows | PowerShell setup script |
| `scripts/setup.sh` | Linux/macOS | Bash setup script |
| `scripts/activate.ps1` | Windows | PowerShell venv activation wrapper |
| `scripts/activate.sh` | Linux/macOS | Bash venv activation wrapper (use with `source`) |
| `scripts/build.ps1` | Windows | PowerShell build script (PyInstaller/wheel) |
| `scripts/build.sh` | Linux/macOS | Bash build script (PyInstaller/wheel) |

### Script Template Logic

Both generated scripts follow simplified logic (portunix handles venv existence checks internally):

1. Determine script and project directories
2. Change to project root
3. Call `portunix python init` (handles venv creation + dependency installation)
4. Display success message and activation instructions

**Note:** Portunix internally handles:

- Checking if `.venv` exists → returns error with hint to use `--force`
- Creating venv and installing dependencies in one command
- Pip upgrade after venv creation

### Generated Script Features

- **Portunix-native**: Uses `portunix python` commands (not direct python/pip)
- **Error handling**: Stops on first error (`set -e` / `$ErrorActionPreference`)
- **Minimal logic**: Delegates complexity to portunix
- **Colored output**: Visual feedback for progress
- **Cross-platform**: Generates both Windows and Unix scripts
- **Relative paths**: Scripts work from any location

### Example Generated Bash Script

```bash
#!/bin/bash
# Setup virtual environment for {{PROJECT_NAME}}
# Generated by: portunix python generate-scripts
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo -e "\033[36mSetting up virtual environment for {{PROJECT_NAME}}...\033[0m"
cd "$PROJECT_DIR"

# portunix handles: venv creation, pip upgrade, dependency installation
# Use --force to recreate existing venv
portunix python init --path .venv

echo ""
echo -e "\033[32mSetup complete!\033[0m"
echo -e "\033[36mTo activate: source .venv/bin/activate\033[0m"
```

### Example Generated PowerShell Script

```powershell
# Setup virtual environment for {{PROJECT_NAME}}
# Generated by: portunix python generate-scripts
$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

Write-Host "Setting up virtual environment for {{PROJECT_NAME}}..." -ForegroundColor Cyan
Set-Location $ProjectDir

# portunix handles: venv creation, pip upgrade, dependency installation
# Use --force to recreate existing venv
portunix python init --path .venv

Write-Host ""
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "To activate: source scripts/activate.sh (or scripts/activate.ps1)" -ForegroundColor Cyan
```

### Example Generated Activate Scripts

**Bash** (`scripts/activate.sh`):
```bash
#!/bin/bash
# Activate virtual environment for {{PROJECT_NAME}}
# Usage: source scripts/activate.sh
#
# Generated by: portunix python generate-scripts

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

source "$PROJECT_DIR/.venv/bin/activate"
```

**PowerShell** (`scripts/activate.ps1`):

```powershell
# Activate virtual environment for {{PROJECT_NAME}}
# Usage: . scripts/activate.ps1
#
# Generated by: portunix python generate-scripts

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

& "$ProjectDir\.venv\Scripts\Activate.ps1"
```

**Benefits of activate wrappers:**

- Consistent path from project root (`source scripts/activate.sh`)
- Works regardless of current directory
- Self-documenting (shows which project's venv is being activated)

### Example Generated Build Scripts

Build scripts leverage existing `portunix python build` functionality (PyInstaller, cx_Freeze, wheel).

```bash
# Generate build scripts
portunix python generate-scripts --build
portunix python generate-scripts --build --build-type exe      # PyInstaller (default)
portunix python generate-scripts --build --build-type wheel    # Wheel distribution
```

**Bash** (`scripts/build.sh`):

```bash
#!/bin/bash
# Build {{PROJECT_NAME}} executable
# Generated by: portunix python generate-scripts --build
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo -e "\033[36mBuilding {{PROJECT_NAME}}...\033[0m"

# portunix handles: PyInstaller installation, build configuration
portunix python build exe {{ENTRY_POINT}} --onefile --path .venv

echo ""
echo -e "\033[32mBuild complete!\033[0m"
echo -e "\033[36mOutput: dist/\033[0m"
```

**PowerShell** (`scripts/build.ps1`):

```powershell
# Build {{PROJECT_NAME}} executable
# Generated by: portunix python generate-scripts --build
$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

Set-Location $ProjectDir

Write-Host "Building {{PROJECT_NAME}}..." -ForegroundColor Cyan

# portunix handles: PyInstaller installation, build configuration
portunix python build exe {{ENTRY_POINT}} --onefile --path .venv

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host "Output: dist/" -ForegroundColor Cyan
```

**Build types supported:**

- `exe` - Standalone executable via PyInstaller (default)
- `wheel` - Python wheel distribution (.whl)
- `sdist` - Source distribution (.tar.gz)

## Phase 3: Start Scripts + Run Command

### Problem

Projects also need start scripts to run the application. Currently, AI assistants generate 50+ line scripts like:

```powershell
# Start Application (Windows)
param(
    [string]$Config = "",
    [string]$Port = ""
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
$VenvDir = Join-Path $ProjectDir ".venv"

# Check if virtual environment exists
if (-not (Test-Path $VenvDir)) {
    Write-Host "ERROR: Virtual environment not found." -ForegroundColor Red
    Write-Host "Please run setup.ps1 first" -ForegroundColor Yellow
    exit 1
}

$PythonExe = Join-Path $VenvDir "Scripts\python.exe"

# Build command arguments
$Args = @()
if ($Config) { $Args += "--config"; $Args += $Config }
if ($Port) { $Args += "--port"; $Args += $Port }

# Run the application
Set-Location $ProjectDir
& $PythonExe -m src $Args
```

This is repetitive and requires knowing the venv path, Python executable location, etc.

### Proposed Solution

#### New Command: `portunix python run`

```bash
# Run Python application using local venv
portunix python run                              # Auto-detect entry point
portunix python run -m src                       # Run as module
portunix python run main.py                      # Run script
portunix python run -m src -- --port 8080        # Forward args to application
portunix python run --path .venv -m myapp        # Explicit venv path
```

**Behavior:**

1. Auto-detect `.venv` in current directory (or use `--path`)
2. Validate venv exists → clear error with hint to run setup script
3. Resolve Python executable path
4. Execute with forwarded arguments

#### Entry Point Detection

```text
Detection priority:
1. Explicit: -m <module> or <script.py>
2. pyproject.toml [project.scripts]     → First console_scripts entry
3. pyproject.toml [tool.poetry.scripts] → First entry
4. setup.py entry_points                → console_scripts
5. main.py in project root              → Fallback
6. Error with hint to specify entry point
```

#### Generate Start Scripts

```bash
# Generate start scripts along with setup scripts
portunix python generate-scripts --start
portunix python generate-scripts --start --entry-point "-m src"
portunix python generate-scripts --start --entry-point "main.py"
```

### Generated Start Scripts

| File | Platform | Description |
| ---- | -------- | ----------- |
| `scripts/start.ps1` | Windows | PowerShell start script |
| `scripts/start.sh` | Linux/macOS | Bash start script |

### Example Generated Start Scripts

**Bash:**

```bash
#!/bin/bash
# Start {{PROJECT_NAME}}
# Generated by: portunix python generate-scripts --start
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# portunix handles: venv detection, Python path resolution, argument forwarding
portunix python run {{ENTRY_POINT}} "$@"
```

**PowerShell:**

```powershell
# Start {{PROJECT_NAME}}
# Generated by: portunix python generate-scripts --start
$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir

Set-Location $ProjectDir

# portunix handles: venv detection, Python path resolution, argument forwarding
portunix python run {{ENTRY_POINT}} $args
```

### Benefits of `portunix python run`

1. **Auto-detect venv** - No need to specify `.venv/Scripts/python.exe`
2. **Cross-platform** - Same command works on Windows and Linux
3. **Validation** - Clear error if venv missing with setup hint
4. **Entry point detection** - Can auto-detect from pyproject.toml
5. **Argument forwarding** - Clean separation with `--`

## Architecture Decision

See [ADR-033](../../adr/033-ptx-python-project-local-venv-support.md) for detailed design.

## Acceptance Criteria

### Functional Requirements

- [ ] `portunix python init` creates `./.venv` and installs from `requirements.txt`
- [ ] `portunix python init --force` recreates existing venv without prompting
- [ ] `portunix python init` automatically upgrades pip after venv creation
- [ ] `--local` flag works for: `venv create`, `venv info`, `venv delete`, `pip install`, `pip list`, `pip freeze`
- [ ] `--path <custom>` allows arbitrary venv location
- [ ] Existing `--venv <name>` behavior unchanged (centralized venvs)
- [ ] Cross-platform: Works on Windows, Linux, macOS
- [ ] Auto-detection: When `./.venv` exists, pip commands use it implicitly
- [ ] **All pip operations use `python -m pip` pattern** (not direct pip executable)

### Error Handling

- [ ] Clear error if Python not found (with installation hint)
- [ ] Warning if `./.venv` exists and `--force` not specified
- [ ] Error if `requirements.txt` not found during `init`
- [ ] Graceful handling of permission issues

### User Experience

- [ ] Display activation command appropriate for detected shell/OS
- [ ] Progress indicators during venv creation and package installation
- [ ] Clear success/failure messages

### Help Documentation

- [ ] `portunix python --help` lists all subcommands including new ones
- [ ] `portunix python init --help` documents all init flags
- [ ] `portunix python run --help` documents run flags and entry point detection
- [ ] `portunix python generate-scripts --help` documents all script generation options
- [ ] Help includes usage examples for common scenarios

### Phase 2: Script Generation

- [ ] `portunix python init --generate-scripts` creates setup + activate scripts
- [ ] `portunix python generate-scripts` works standalone for existing projects
- [ ] `--dir <path>` flag allows custom output directory
- [ ] Generated PowerShell scripts are syntactically correct and executable
- [ ] Generated Bash scripts are syntactically correct and executable
- [ ] Scripts use `portunix python` commands (not direct python/pip)
- [ ] Scripts delegate venv existence handling to portunix (no manual if/else logic)
- [ ] Scripts include colored output for user feedback
- [ ] Scripts work with relative paths (can be run from any location)
- [ ] `--project <name>` flag sets project name in generated scripts
- [ ] Auto-detection of project name from pyproject.toml, setup.py, package.json
- [ ] Fallback to directory name when no project metadata found
- [ ] Generated `activate.sh` works with `source scripts/activate.sh`
- [ ] Generated `activate.ps1` works with `. scripts/activate.ps1`
- [ ] `--build` flag generates build.ps1 and build.sh
- [ ] `--build-type exe` generates PyInstaller build script (default)
- [ ] `--build-type wheel` generates wheel build script
- [ ] Generated build scripts use `portunix python build` command

### Phase 3: Start Scripts + Run Command

- [ ] `portunix python run` executes Python application using local venv
- [ ] `portunix python run -m <module>` runs specified module
- [ ] `portunix python run <script.py>` runs specified script
- [ ] `portunix python run ... -- <args>` forwards arguments to application
- [ ] `portunix python run` auto-detects entry point from pyproject.toml
- [ ] Clear error if venv not found with hint to run setup
- [ ] `--start` flag for generate-scripts creates start.ps1 and start.sh
- [ ] `--entry-point` flag specifies entry point for generated start scripts
- [ ] Generated start scripts use `portunix python run` command
- [ ] Generated start scripts forward all arguments to application

## Implementation Notes

### Files to Modify

| File | Changes |
| ---- | ------- |
| `src/helpers/ptx-python/main.go` | Add `init`, `run`, `generate-scripts` command handlers, **update --help** |
| `src/helpers/ptx-python/venv_manager.go` | Add `CreateLocalVenv()`, `ResolveVenvPath()` methods |
| `src/helpers/ptx-python/script_generator.go` | **NEW** - Setup/start/build script generation logic |
| `src/helpers/ptx-python/runner.go` | **NEW** - Python application runner |
| `src/helpers/ptx-python/entrypoint.go` | **NEW** - Entry point detection from pyproject.toml |
| `src/helpers/ptx-python/templates/` | **NEW** - Embedded script templates (setup, activate, start, build) |

### Help Text Updates

**IMPORTANT**: Update `--help` output for all new commands and flags:

```text
portunix python --help                    # Main help - list all subcommands
portunix python init --help               # Init command options
portunix python run --help                # Run command options
portunix python generate-scripts --help   # Script generation options
portunix python build --help              # Build command options (existing)
```

Each help should document:

- Command purpose and usage
- All available flags with descriptions
- Examples of common use cases

### Key Functions to Add

```go
// venv_manager.go
func (vm *VenvManager) CreateLocalVenv(force bool, pythonVersion string) error
func (vm *VenvManager) ResolveVenvPath(local bool, path string, venvName string) (string, error)
func (vm *VenvManager) DetectRequirementsFile() (string, error)

// main.go
func handleInitCommand(args []string)
func handleGenerateScriptsCommand(args []string)

// script_generator.go (Phase 2)
type ScriptGenerator struct {
    OutputDir   string
    ScriptName  string
    VenvPath    string
    ProjectName string
}

func NewScriptGenerator(outputDir, scriptName, venvPath, projectName string) *ScriptGenerator
func (sg *ScriptGenerator) Generate() error
func (sg *ScriptGenerator) generatePowerShell() error
func (sg *ScriptGenerator) generateBash() error
func (sg *ScriptGenerator) detectProjectName() string  // Auto-detect from pyproject.toml, etc.

// runner.go (Phase 3)
type Runner struct {
    VenvPath   string
    EntryPoint string
}

func NewRunner(venvPath string) *Runner
func (r *Runner) Run(args []string) error
func (r *Runner) DetectEntryPoint() (string, error)

// entrypoint.go (Phase 3)
func DetectEntryPointFromPyproject(path string) (string, error)
func DetectEntryPointFromSetupPy(path string) (string, error)
```

### Flag Precedence

```text
1. --path <explicit>  → Use explicit path
2. --local            → Use ./.venv
3. --venv <name>      → Use ~/.portunix/python/venvs/<name>
4. Auto-detect        → If ./.venv exists, use it (pip commands only)
5. No target          → Error with usage hint
```

## Testing Strategy

### Container-Based Testing (Mandatory)

```bash
# Test in clean Ubuntu container
portunix container run ubuntu
cd /tmp && mkdir testproject && cd testproject

# Test init
echo "flask==3.0.0" > requirements.txt
./portunix python init
test -d ./.venv && echo "PASS: venv created"

# Test pip operations
./portunix python pip list --local | grep -i flask && echo "PASS: flask installed"
./portunix python pip freeze --local > frozen.txt
```

### Test Cases

| Test Case | Description |
| --------- | ----------- |
| TC-001 | `python init` with requirements.txt |
| TC-002 | `python init --force` with existing venv |
| TC-003 | `python init` without requirements.txt (error) |
| TC-004 | `venv create --local` in empty directory |
| TC-005 | `venv create --path ./custom` |
| TC-006 | `pip install --local` package |
| TC-007 | `pip list --local` shows installed packages |
| TC-008 | Auto-detection of ./.venv for pip commands |
| TC-009 | Windows compatibility (activation command) |
| TC-010 | Linux compatibility (activation command) |
| TC-011 | `python init` upgrades pip without error (pip self-upgrade) |
| TC-012 | `pip install --upgrade pip --local` works correctly |
| TC-013 | `python generate-scripts` creates setup + activate scripts |
| TC-014 | `python generate-scripts --dir ./tools` uses custom directory |
| TC-015 | `python init --generate-scripts` creates venv AND scripts |
| TC-015a | `source scripts/activate.sh` activates venv correctly (Linux) |
| TC-015b | `. scripts/activate.ps1` activates venv correctly (Windows) |
| TC-016 | Generated .sh script passes `shellcheck` validation |
| TC-017 | Generated .ps1 script is valid PowerShell syntax |
| TC-018 | Generated scripts work when run from different directory |
| TC-019 | Generated scripts handle missing requirements.txt gracefully |
| TC-020 | `--project "My App"` sets project name in script comments |
| TC-021 | Project name auto-detected from pyproject.toml |
| TC-022 | Project name fallback to directory name |
| TC-023 | `python run -m src` executes module in local venv |
| TC-024 | `python run main.py` executes script in local venv |
| TC-025 | `python run -m src -- --port 8080` forwards args to application |
| TC-026 | `python run` auto-detects entry point from pyproject.toml |
| TC-027 | `python run` with missing venv shows clear error with setup hint |
| TC-028 | `generate-scripts --start` creates start.ps1 and start.sh |
| TC-029 | `generate-scripts --start --entry-point "-m src"` uses specified entry point |
| TC-030 | Generated start scripts forward arguments correctly |
| TC-031 | Generated start scripts work on Windows |
| TC-032 | Generated start scripts work on Linux |
| TC-033 | `generate-scripts --build` creates build.ps1 and build.sh |
| TC-034 | `generate-scripts --build --build-type exe` generates PyInstaller script |
| TC-035 | `generate-scripts --build --build-type wheel` generates wheel build script |
| TC-036 | Generated build scripts create executable successfully |
| TC-037 | Generated build scripts work on Windows |
| TC-038 | Generated build scripts work on Linux |
| TC-039 | `python --help` lists init, run, generate-scripts commands |
| TC-040 | `python init --help` shows all flags with descriptions |
| TC-041 | `python run --help` documents entry point detection |
| TC-042 | `python generate-scripts --help` documents --start, --build flags |

## Estimated Complexity

- **Phase 1 (Local Venv)**: Medium - Extends existing VenvManager with new methods
- **Phase 2 (Script Generation)**: Low - Template-based generation, no complex logic
- **Phase 3 (Run Command + Start Scripts)**: Medium - Entry point detection, argument forwarding

## Dependencies

- Issue #097 (PTX-Python Helper Implementation) - Phase 2 must be complete

## References

- ADR: [ADR-033](../../adr/033-ptx-python-project-local-venv-support.md)
- Related: Issue #097 (PTX-Python Helper Implementation)
- Python venv docs: https://docs.python.org/3/library/venv.html

---

**Created**: 2026-01-21
**Updated**: 2026-01-22 (Added Phase 2: Script Generation, Phase 3: Start Scripts + Run Command)
**Author**: ZK
