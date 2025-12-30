# Activate Python Virtual Environment for Portunix
# Usage: .\scripts\activate-venv.ps1
#
# Note: This script needs to be dot-sourced to work correctly:
#   . .\scripts\activate-venv.ps1

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$VenvDir = Join-Path $ProjectRoot ".venv"
$ActivateScript = Join-Path $VenvDir "Scripts\Activate.ps1"

# Check if venv exists
if (-not (Test-Path $ActivateScript)) {
    Write-Host "Virtual environment not found at $VenvDir" -ForegroundColor Red
    Write-Host "Run '.\scripts\setup-venv.ps1' first to create it." -ForegroundColor Yellow
    return
}

# Activate
& $ActivateScript

Write-Host "Virtual environment activated: $VenvDir" -ForegroundColor Green
$pythonVersion = & python --version 2>&1
Write-Host "Python: $pythonVersion" -ForegroundColor Cyan
Write-Host "To deactivate, run: deactivate" -ForegroundColor Gray
