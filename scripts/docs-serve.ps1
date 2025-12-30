# Serve Portunix documentation locally
# Uses Hugo development server or Python fallback

param(
    [Alias("s")]
    [switch]$Static
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$DocsSite = Join-Path $ProjectRoot "docs-site"
$PublicDir = Join-Path $DocsSite "public"

function Print-Info($message) {
    Write-Host "i  $message" -ForegroundColor Blue
}

function Print-Success($message) {
    Write-Host "[OK] $message" -ForegroundColor Green
}

function Print-Warning($message) {
    Write-Host "[!] $message" -ForegroundColor Yellow
}

function Print-Error($message) {
    Write-Host "[X] $message" -ForegroundColor Red
}

Write-Host ""
Write-Host "================================================================" -ForegroundColor Blue
Write-Host "           Portunix Documentation Server                        " -ForegroundColor Blue
Write-Host "================================================================" -ForegroundColor Blue
Write-Host ""

# Check for --static flag (serve only public/ without Hugo)
if ($Static) {
    if (-not (Test-Path $PublicDir)) {
        Print-Error "Directory $PublicDir does not exist"
        Print-Info "Run 'python scripts/post-release-docs.py --build-only' first"
        exit 1
    }

    Print-Info "Serving static files from: $PublicDir"
    Print-Info "Server URL: http://localhost:8080"
    Write-Host ""

    Set-Location $PublicDir
    python -m http.server 8080
    exit 0
}

# Try Hugo first
$hugoPath = Get-Command hugo -ErrorAction SilentlyContinue

if ($hugoPath) {
    Print-Success "Hugo found"
    Print-Info "Starting Hugo development server..."
    Print-Info "Server URL: http://localhost:1313"
    Write-Host ""

    Set-Location $DocsSite
    hugo server
} else {
    Print-Warning "Hugo not found, using Python HTTP server"

    if (-not (Test-Path $PublicDir)) {
        Print-Error "Directory $PublicDir does not exist"
        Print-Info "Install Hugo: portunix install hugo"
        Print-Info "Or run: python scripts/post-release-docs.py --build-only"
        exit 1
    }

    Print-Info "Serving static files from: $PublicDir"
    Print-Info "Server URL: http://localhost:8080"
    Write-Host ""

    Set-Location $PublicDir
    python -m http.server 8080
}
