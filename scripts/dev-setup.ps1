# Portunix Development Environment Setup
# Installs all required development dependencies using portunix
#
# Usage: .\scripts\dev-setup.ps1 [-DryRun]
#
# Prerequisites: portunix binary must be available in PATH or current directory

param(
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

# Find portunix binary
$ptx = $null
if (Get-Command "portunix" -ErrorAction SilentlyContinue) {
    $ptx = "portunix"
} elseif (Test-Path ".\portunix.exe") {
    $ptx = ".\portunix.exe"
} else {
    Write-Error "portunix binary not found in PATH or current directory. Install portunix first: https://github.com/cassandragargoyle/portunix"
    exit 1
}

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Portunix Development Environment Setup" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$version = & $ptx --version 2>$null
if ($version) { Write-Host "Using: $version" }
if ($DryRun) { Write-Host "=== DRY RUN MODE ===" -ForegroundColor Yellow }
Write-Host ""

# Required packages for building and developing portunix
$packages = @(
    @{ Name = "go";     Description = "Go compiler (primary language)" }
    @{ Name = "python"; Description = "Python (release scripts, deploy scripts)" }
    @{ Name = "gh";     Description = "GitHub CLI (release publishing)" }
    # @{ Name = "make"; Description = "GNU Make (build automation) - not yet in registry" }
)

# Optional packages (install failures are non-fatal)
$optionalPackages = @(
    @{ Name = "hugo"; Description = "Documentation site generator" }
)

Write-Host "--- Installing required packages ---" -ForegroundColor Green
Write-Host ""

$failed = @()
foreach ($pkg in $packages) {
    Write-Host ">>> Installing: $($pkg.Name) - $($pkg.Description)" -ForegroundColor White
    $args = @("install", $pkg.Name)
    if ($DryRun) { $args += "--dry-run" }

    & $ptx @args
    if ($LASTEXITCODE -ne 0) {
        $failed += $pkg.Name
        Write-Host "WARNING: Failed to install $($pkg.Name)" -ForegroundColor Yellow
    }
    Write-Host ""
}

Write-Host "--- Installing optional packages ---" -ForegroundColor Green
Write-Host ""

foreach ($pkg in $optionalPackages) {
    Write-Host ">>> Installing (optional): $($pkg.Name) - $($pkg.Description)" -ForegroundColor White
    $args = @("install", $pkg.Name)
    if ($DryRun) { $args += "--dry-run" }

    & $ptx @args
    if ($LASTEXITCODE -ne 0) {
        Write-Host "NOTE: Optional package $($pkg.Name) failed to install (non-fatal)" -ForegroundColor DarkYellow
    }
    Write-Host ""
}

# Setup Python virtual environment
if (-not (Test-Path ".venv") -and -not $DryRun) {
    $pythonCmd = Get-Command "python" -ErrorAction SilentlyContinue
    if ($pythonCmd) {
        Write-Host "--- Setting up Python virtual environment ---" -ForegroundColor Green
        if (Test-Path ".\scripts\setup-venv.cmd") {
            & .\scripts\setup-venv.cmd
        } else {
            python -m venv .venv
        }
        Write-Host ""
    }
}

# Install Node.js (provides npm) via portunix
Write-Host "--- Installing Node.js for npm-based tools ---" -ForegroundColor Green
Write-Host ">>> Installing: nodejs - Node.js JavaScript Runtime" -ForegroundColor White
$installArgs = @("install", "nodejs")
if ($DryRun) { $installArgs += "--dry-run" }
& $ptx @installArgs
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Failed to install nodejs" -ForegroundColor Yellow
}
Write-Host ""

# Install markdownlint-cli2 via npm
$npmCmd = Get-Command "npm" -ErrorAction SilentlyContinue
if ($npmCmd) {
    Write-Host "--- Installing markdownlint-cli2 ---" -ForegroundColor Green
    if (-not $DryRun) {
        npm install -g markdownlint-cli2
        if ($LASTEXITCODE -ne 0) {
            Write-Host "WARNING: Failed to install markdownlint-cli2" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[dry-run] Would run: npm install -g markdownlint-cli2"
    }
    Write-Host ""
} else {
    Write-Host "NOTE: npm not found after nodejs install, skipping markdownlint-cli2" -ForegroundColor DarkYellow
    Write-Host "      Restart your terminal and re-run this script"
    Write-Host ""
}

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Setup Summary" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

if ($failed.Count -gt 0) {
    Write-Host ""
    Write-Host "FAILED packages: $($failed -join ', ')" -ForegroundColor Red
    Write-Host "Please install these manually or check error messages above."
    exit 1
} else {
    Write-Host ""
    Write-Host "All required packages installed successfully." -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  1. Restart your terminal (to pick up PATH changes)"
    Write-Host "  2. Run: make build"
    Write-Host ""
}
