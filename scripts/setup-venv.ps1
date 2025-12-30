# Setup Python Virtual Environment for Portunix
# Usage: .\scripts\setup-venv.ps1 [-WithTests]
#
# Creates a virtual environment in .venv\ directory
# Optionally installs test dependencies with -WithTests flag

param(
    [switch]$WithTests,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$VenvDir = Join-Path $ProjectRoot ".venv"

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[!] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[X] $Message" -ForegroundColor Red
}

function Show-Help {
    Write-Host "Usage: .\scripts\setup-venv.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -WithTests    Also install test dependencies"
    Write-Host "  -Help         Show this help message"
    exit 0
}

function Find-Python {
    Write-Step "Detecting Python..."

    # Try py launcher first (Windows Python Launcher)
    try {
        $pyVersion = & py -3 --version 2>&1
        if ($LASTEXITCODE -eq 0 -and $pyVersion -match "Python 3") {
            $script:PythonCmd = "py"
            $script:PythonArgs = @("-3")
            Write-Success "Found Python via py launcher: $pyVersion"
            return $true
        }
    } catch {}

    # Try python command
    try {
        $pythonVersion = & python --version 2>&1
        if ($LASTEXITCODE -eq 0 -and $pythonVersion -match "Python 3") {
            $script:PythonCmd = "python"
            $script:PythonArgs = @()
            Write-Success "Found Python: $pythonVersion"
            return $true
        }
    } catch {}

    # Try python3 command
    try {
        $python3Version = & python3 --version 2>&1
        if ($LASTEXITCODE -eq 0 -and $python3Version -match "Python 3") {
            $script:PythonCmd = "python3"
            $script:PythonArgs = @()
            Write-Success "Found Python: $python3Version"
            return $true
        }
    } catch {}

    Write-Error "Python 3 is not installed"
    Write-Host "Install Python 3.8+ from https://python.org or use: portunix install python"
    return $false
}

function New-Venv {
    Write-Step "Creating virtual environment in $VenvDir..."

    if (Test-Path $VenvDir) {
        Write-Warning "Virtual environment already exists"
        $response = Read-Host "Do you want to recreate it? [y/N]"
        if ($response -match "^[Yy]$") {
            Remove-Item -Recurse -Force $VenvDir
        } else {
            Write-Step "Using existing virtual environment"
            return
        }
    }

    & $PythonCmd @PythonArgs -m venv $VenvDir
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to create virtual environment"
        exit 1
    }
    Write-Success "Virtual environment created"
}

function Install-Dependencies {
    Write-Step "Installing dependencies..."

    $VenvPython = Join-Path $VenvDir "Scripts\python.exe"

    # Upgrade pip
    & $VenvPython -m pip install --upgrade pip 2>&1 | Out-Null
    Write-Success "pip upgraded"

    # Install main requirements if exists and has content
    $RequirementsFile = Join-Path $ProjectRoot "requirements.txt"
    if (Test-Path $RequirementsFile) {
        $content = Get-Content $RequirementsFile | Where-Object { $_ -notmatch "^\s*#" -and $_ -notmatch "^\s*$" }
        if ($content) {
            & $VenvPython -m pip install -r $RequirementsFile
            Write-Success "Main dependencies installed"
        } else {
            Write-Success "No main dependencies to install (requirements.txt has only comments)"
        }
    }

    # Install test requirements if requested
    $TestRequirementsFile = Join-Path $ProjectRoot "test\requirements-test.txt"
    if ($WithTests -and (Test-Path $TestRequirementsFile)) {
        & $VenvPython -m pip install -r $TestRequirementsFile
        Write-Success "Test dependencies installed"
    }
}

function Show-Instructions {
    Write-Host ""
    Write-Step "Setup complete!"
    Write-Host ""
    Write-Host "To activate the virtual environment:"
    Write-Host ""
    Write-Host "  PowerShell:    .\.venv\Scripts\Activate.ps1"
    Write-Host "  CMD:           .\.venv\Scripts\activate.bat"
    Write-Host "  Git Bash:      source .venv/Scripts/activate"
    Write-Host ""
    Write-Host "Or use the helper scripts:"
    Write-Host ""
    Write-Host "  PowerShell:    .\scripts\activate-venv.ps1"
    Write-Host "  CMD:           scripts\activate-venv.cmd"
    Write-Host "  Bash:          source scripts/activate-venv.sh"
    Write-Host ""
    Write-Host "To deactivate: deactivate"
    Write-Host ""
}

# Main
function Main {
    Write-Host "================================"
    Write-Host "Portunix Python Environment Setup"
    Write-Host "================================"
    Write-Host ""

    if ($Help) {
        Show-Help
    }

    Push-Location $ProjectRoot
    try {
        if (-not (Find-Python)) {
            exit 1
        }
        New-Venv
        Install-Dependencies
        Show-Instructions
    } finally {
        Pop-Location
    }
}

Main
