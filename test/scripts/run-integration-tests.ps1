# Integration Test Runner for Portunix
# PowerShell wrapper for Python-based integration tests

param(
    [Alias("q")][switch]$quick,
    [Alias("f", "full")][switch]$full_suite,
    [Alias("d", "dist")][string]$distribution,
    [Alias("l", "list")][switch]$list_distributions,
    [Alias("p")][switch]$parallel,
    [Alias("c")][switch]$cleanup,
    [Alias("v")][switch]$verbose,
    [Alias("html")][string]$html_report,
    [Alias("h")][switch]$help
)

# Enable strict error handling
$ErrorActionPreference = "Stop"

# Get script directory and project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$TestDir = Join-Path $ProjectRoot "test"
$PythonRunner = Join-Path $TestDir "scripts" "test-integration.py"

# Function to show usage
function Show-Usage {
    Write-Host "Usage: run-integration-tests.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -quick, -q              Run quick test (Ubuntu 22.04 only)"
    Write-Host "  -full_suite, -f         Run complete test suite (all distributions)"
    Write-Host "  -distribution, -d NAME  Run specific distribution test"
    Write-Host "  -list_distributions, -l  List available distributions"
    Write-Host "  -parallel, -p           Run tests in parallel"
    Write-Host "  -cleanup, -c            Clean up test containers"
    Write-Host "  -verbose, -v            Verbose output"
    Write-Host "  -html_report FILE       Generate HTML report"
    Write-Host "  -help, -h               Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run-integration-tests.ps1 -quick"
    Write-Host "  .\run-integration-tests.ps1 -full_suite -parallel"
    Write-Host "  .\run-integration-tests.ps1 -d ubuntu-22"
    Write-Host "  .\run-integration-tests.ps1 -list_distributions"
    Write-Host ""
    Write-Host "Note: PowerShell uses single dash (-) for parameters, not double dash (--)"
}

# Check for help flag
if ($help) {
    Show-Usage
    exit 0
}

# Check if Python is available
$python = $null
foreach ($cmd in @("python3", "python")) {
    if (Get-Command $cmd -ErrorAction SilentlyContinue) {
        $pythonVersion = & $cmd --version 2>&1
        if ($pythonVersion -match "Python 3") {
            $python = $cmd
            break
        }
    }
}

if (-not $python) {
    Write-Host "❌ Python 3 is not installed" -ForegroundColor Red
    exit 1
}

# Check if test runner exists
if (-not (Test-Path $PythonRunner)) {
    Write-Host "❌ Python test runner not found: $PythonRunner" -ForegroundColor Red
    exit 1
}

# Check if no test mode specified
$modeCount = 0
if ($quick) { $modeCount++ }
if ($full_suite) { $modeCount++ }
if ($distribution) { $modeCount++ }
if ($list_distributions) { $modeCount++ }
if ($cleanup) { $modeCount++ }

if ($modeCount -eq 0) {
    Write-Host "⚠️  No test mode specified" -ForegroundColor Yellow
    Write-Host ""
    Show-Usage
    exit 1
}

# Setup log directory and file
$LogDir = Join-Path $TestDir "logs"
if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
}
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$LogFile = Join-Path $LogDir "pip-install-$timestamp.log"

# Setup virtual environment
$VenvDir = Join-Path $TestDir "venv"
$VenvActivate = if ($IsWindows -or $env:OS -eq "Windows_NT") {
    Join-Path $VenvDir "Scripts" "Activate.ps1"
} else {
    Join-Path $VenvDir "bin" "Activate.ps1"
}

# Create virtual environment if it doesn't exist
if (-not (Test-Path $VenvDir)) {
    Write-Host "📦 Creating Python virtual environment..." -ForegroundColor Blue
    & $python -m venv $VenvDir
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to create virtual environment" -ForegroundColor Red
        exit 1
    }
}

# Activate virtual environment
if (Test-Path $VenvActivate) {
    & $VenvActivate
} else {
    Write-Host "⚠️  Virtual environment activation script not found, using system Python" -ForegroundColor Yellow
}

# Check and install required packages
Write-Host "📦 Checking Python dependencies..." -ForegroundColor Blue
$RequiredPackages = @("pytest", "pytest-xdist", "pytest-html")
$PackagesToInstall = @()

foreach ($package in $RequiredPackages) {
    $installed = & pip show $package 2>$null
    if ($LASTEXITCODE -ne 0) {
        $PackagesToInstall += $package
    }
}

if ($PackagesToInstall.Count -gt 0) {
    Write-Host "📝 Installing missing packages (see $LogFile for details)..." -ForegroundColor Yellow
    Add-Content -Path $LogFile -Value "Installing packages: $($PackagesToInstall -join ' ')"
    & pip install $PackagesToInstall *>> $LogFile 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Packages installed successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Some packages failed to install. Check $LogFile for details" -ForegroundColor Red
        exit 1
    }
}

# Check if requirements-test.txt exists and install from it
$RequirementsFile = Join-Path $ProjectRoot "requirements-test.txt"
if (Test-Path $RequirementsFile) {
    Write-Host "📦 Installing from requirements-test.txt (see $LogFile for details)..." -ForegroundColor Blue
    & pip install -r $RequirementsFile *>> $LogFile 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Requirements installed successfully" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Some requirements failed to install. Check $LogFile for details" -ForegroundColor Yellow
    }
}

# Build command arguments
$args = @()

if ($quick) {
    $args += "--quick"
}
if ($full_suite) {
    $args += "--full-suite"
}
if ($distribution) {
    $args += "--distribution"
    $args += $distribution
}
if ($list_distributions) {
    $args += "--list-distributions"
}
if ($parallel) {
    $args += "--parallel"
}
if ($cleanup) {
    $args += "--cleanup"
}
if ($verbose) {
    $args += "--verbose"
}
if ($html_report) {
    $args += "--html-report"
    $args += $html_report
}

# Run the Python test runner
Write-Host "🚀 Running integration tests..." -ForegroundColor Green

# Build the complete command
$pythonArgs = @($PythonRunner) + $args

# Execute Python with arguments
& $python $pythonArgs
$exitCode = $LASTEXITCODE

# Exit with the same code as the Python runner
exit $exitCode