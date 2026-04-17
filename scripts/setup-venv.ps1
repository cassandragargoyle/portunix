# Setup Python Virtual Environment for Portunix (uv-based)
# Usage: .\scripts\setup-venv.ps1 [-WithTests]
#
# Provisions .\.venv\ via `uv sync` (or `uv sync --group test` with -WithTests).
# See ADR-039 for the rationale behind adopting uv.

param(
    [switch]$WithTests,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "[X] $Message" -ForegroundColor Red
}

function Show-Help {
    Write-Host "Usage: .\scripts\setup-venv.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -WithTests    Also install test dependencies (pytest, ...)"
    Write-Host "  -Help         Show this help message"
    Write-Host ""
    Write-Host "This script uses uv (https://docs.astral.sh/uv/) to provision the"
    Write-Host "Python environment. If uv is not installed, run one of:"
    Write-Host ""
    Write-Host "  portunix install uv"
    Write-Host "  powershell -c `"irm https://astral.sh/uv/install.ps1 | iex`""
    exit 0
}

function Test-UvInstalled {
    Write-Step "Detecting uv..."
    try {
        $uvVersion = & uv --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Found uv: $uvVersion"
            return $true
        }
    } catch {}

    Write-ErrorMsg "uv is not installed or not on PATH"
    Write-Host ""
    Write-Host "Install uv via one of:"
    Write-Host "  portunix install uv"
    Write-Host "  powershell -c `"irm https://astral.sh/uv/install.ps1 | iex`""
    Write-Host ""
    Write-Host "See https://docs.astral.sh/uv/ for details."
    return $false
}

function Invoke-UvSync {
    if ($WithTests) {
        Write-Step "Provisioning .venv\ with test dependencies (uv sync --group test)..."
        & uv sync --group test
    } else {
        Write-Step "Provisioning .venv\ (uv sync)..."
        & uv sync
    }
    if ($LASTEXITCODE -ne 0) {
        Write-ErrorMsg "uv sync failed"
        exit 1
    }
    Write-Success "Environment ready"
}

function Show-Instructions {
    Write-Host ""
    Write-Step "Setup complete!"
    Write-Host ""
    Write-Host "Run commands without activation:"
    Write-Host "  uv run pytest test/"
    Write-Host "  uv run scripts/file-server.py --help"
    Write-Host ""
    Write-Host "Or activate the virtual environment:"
    Write-Host ""
    Write-Host "  PowerShell:    .\.venv\Scripts\Activate.ps1"
    Write-Host "  CMD:           .\.venv\Scripts\activate.bat"
    Write-Host "  Git Bash:      source .venv/Scripts/activate"
    Write-Host ""
    Write-Host "To deactivate: deactivate"
    Write-Host ""
}

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
        if (-not (Test-UvInstalled)) {
            exit 1
        }
        Invoke-UvSync
        Show-Instructions
    } finally {
        Pop-Location
    }
}

Main
