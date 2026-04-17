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

# Check if uv is available (per ADR-039, Python deps managed via uv)
if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
    Write-Host "❌ uv is not installed" -ForegroundColor Red
    Write-Host "Install with: portunix install uv"
    Write-Host 'Or: powershell -c "irm https://astral.sh/uv/install.ps1 | iex"'
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

# Setup log directory and file (per ADR-039: deps managed by uv)
$LogDir = Join-Path $TestDir "logs"
if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
}
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$LogFile = Join-Path $LogDir "uv-sync-$timestamp.log"

# Provision .venv with test dependencies (idempotent)
Write-Host "📦 Syncing Python test dependencies via uv..." -ForegroundColor Blue
Push-Location $ProjectRoot
try {
    & uv sync --group test *>> $LogFile 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ uv sync failed. Check $LogFile for details" -ForegroundColor Red
        exit 1
    }
} finally {
    Pop-Location
}
Write-Host "✅ Test dependencies ready" -ForegroundColor Green

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

# Run the Python test runner via uv run (no venv activation needed)
Write-Host "🚀 Running integration tests..." -ForegroundColor Green

Push-Location $ProjectRoot
try {
    & uv run python $PythonRunner @args
    $exitCode = $LASTEXITCODE
} finally {
    Pop-Location
}

exit $exitCode