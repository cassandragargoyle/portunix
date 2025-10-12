# Windows-only Release Builder for Portunix
# https://github.com/cassandragargoyle/Portunix/releases/new
$ErrorActionPreference = 'Stop'

# parse command line arguments
$AppName = "portunix"
$Version = "1.6.3"

if ($args.Length -ge 1) { $AppName = $args[0] }
if ($args.Length -ge 2) { $Version = $args[1] }

Write-Host "=== Portunix Windows Release Builder ===" -ForegroundColor Cyan
Write-Host "Building version: $Version" -ForegroundColor Green
Write-Host "Target: Windows AMD64" -ForegroundColor Green

# temporarily set environment variables for the current session
$env:GOOS  = 'windows'
$env:GOARCH = 'amd64'

# ensure output directory
New-Item -ItemType Directory -Force -Path 'bin' | Out-Null

# Build main binary
Write-Host "Building main binary: ${AppName}.exe" -ForegroundColor Yellow
go build -ldflags "-X main.version=v$Version -s -w" -o "./bin/${AppName}.exe" .
if (-not $?) { throw "Failed to build main binary" }

# Build helper binaries
$helpers = @(
    @{name="ptx-container"; path="./src/helpers/ptx-container"},
    @{name="ptx-mcp"; path="./src/helpers/ptx-mcp"},
    @{name="ptx-virt"; path="./src/helpers/ptx-virt"}
)

foreach ($helper in $helpers) {
    if (Test-Path $helper.path) {
        Write-Host "Building helper: $($helper.name).exe" -ForegroundColor Yellow
        # Change to helper directory and build
        Push-Location $helper.path
        go build -ldflags "-X main.version=v$Version -s -w" -o "../../../bin/$($helper.name).exe" .
        $buildSuccess = $?
        Pop-Location
        if (-not $buildSuccess) { throw "Failed to build $($helper.name)" }
    } else {
        Write-Host "Helper not found: $($helper.path) - skipping" -ForegroundColor Red
    }
}

Write-Host "Build completed successfully!" -ForegroundColor Green

# Package all binaries
Write-Host "Creating distribution package..." -ForegroundColor Yellow

# ensure dist directory exists
New-Item -ItemType Directory -Force -Path 'dist' | Out-Null

$zip = "dist\${AppName}_${Version}_windows_amd64.zip"
if (Test-Path $zip) {
    Write-Host "Removing existing package: $zip" -ForegroundColor Yellow
    Remove-Item $zip
}

# collect all built binaries
$binaries = Get-ChildItem "bin\*.exe"
if ($binaries.Count -eq 0) {
    Write-Error "No binaries found in bin\ directory"
    exit 1
}

Write-Host "Packaging binaries:" -ForegroundColor Green
foreach ($binary in $binaries) {
    Write-Host "  - $($binary.Name)" -ForegroundColor White
}

# create package with all binaries
Compress-Archive -Path "bin\*.exe" -DestinationPath $zip

Write-Host "Package created: $zip" -ForegroundColor Green

# show package contents
Write-Host "Package contents:" -ForegroundColor Green
Add-Type -AssemblyName System.IO.Compression.FileSystem
$archive = [System.IO.Compression.ZipFile]::OpenRead((Resolve-Path $zip).Path)
foreach ($entry in $archive.Entries) {
    $size = [math]::Round($entry.Length / 1KB, 2)
    Write-Host "  - $($entry.Name) (${size} KB)" -ForegroundColor White
}
$archive.Dispose()

Write-Host "Windows release build completed successfully!" -ForegroundColor Green