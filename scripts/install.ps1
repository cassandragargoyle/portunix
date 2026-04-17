# Portunix Installation Script for Windows
# Standalone installer - downloads binary from GitHub Releases automatically

param(
    [switch]$Silent,
    [string]$Path = "",
    [switch]$AddToPath,
    [switch]$CreateConfig,
    [switch]$DryRun,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Script version
$ScriptVersion = "2.0.0"

# GitHub repository info
$GitHubOwner = "cassandragargoyle"
$GitHubRepo = "Portunix"
$GitHubApiUrl = "https://api.github.com/repos/$GitHubOwner/$GitHubRepo/releases/latest"
$GitHubDownloadBase = "https://github.com/$GitHubOwner/$GitHubRepo/releases/download"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Show-Help {
    Write-Host @"
Portunix Installation Script v$ScriptVersion

Usage: .\install.ps1 [options]

Options:
    -Silent         Silent installation with defaults
    -Path <path>    Custom installation path
    -AddToPath      Add to system PATH (requires admin for system-wide)
    -CreateConfig   Create default configuration
    -DryRun         Show what would be installed without making changes
    -Help           Show this help message

Examples:
    # Interactive installation (recommended)
    .\install.ps1

    # One-liner (paste into PowerShell)
    irm https://github.com/$GitHubOwner/$GitHubRepo/releases/latest/download/install.ps1 | iex

    # Silent installation with defaults
    .\install.ps1 -Silent

    # Install to C:\Portunix (recommended)
    .\install.ps1 -Path "C:\Portunix" -AddToPath

    # Install to Program Files (requires admin)
    .\install.ps1 -Path "C:\Program Files\Portunix" -AddToPath

    # Preview installation without changes
    .\install.ps1 -DryRun

Default installation paths:
  - C:\Portunix\portunix.exe (recommended, no admin required)
  - `$env:PROGRAMFILES\Portunix\portunix.exe (requires admin)
"@
}

function Test-Administrator {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-PlatformArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default {
            # Fallback for older PowerShell
            if ([Environment]::Is64BitOperatingSystem) {
                return "amd64"
            }
            Write-ColorOutput Red "ERROR: Unsupported architecture: $arch"
            Write-Host ""
            Write-Host "Portunix supports Windows amd64 and arm64 architectures."
            Write-Host "Your system architecture '$arch' is not supported."
            exit 1
        }
    }
}

function Get-LatestReleaseInfo {
    Write-Host "Checking latest Portunix release..."

    try {
        $headers = @{ "Accept" = "application/vnd.github+json" }
        $release = Invoke-RestMethod -Uri $GitHubApiUrl -Headers $headers -TimeoutSec 30
        return $release
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-ColorOutput Red "ERROR: Failed to check latest release"
        Write-Host ""

        if ($statusCode -eq 403) {
            Write-Host "GitHub API rate limit exceeded. Try again in a few minutes."
            Write-Host "Alternatively, download manually from:"
            Write-Host "  https://github.com/$GitHubOwner/$GitHubRepo/releases/latest"
        } elseif ($statusCode -eq 404) {
            Write-Host "No releases found for Portunix."
            Write-Host "Check the repository: https://github.com/$GitHubOwner/$GitHubRepo/releases"
        } else {
            Write-Host "Network error: $($_.Exception.Message)"
            Write-Host ""
            Write-Host "Possible causes:"
            Write-Host "  - No internet connection"
            Write-Host "  - Firewall blocking github.com"
            Write-Host "  - Proxy configuration required"
            Write-Host ""
            Write-Host "To configure proxy:"
            Write-Host '  $env:HTTPS_PROXY = "http://proxy:port"'
        }
        exit 1
    }
}

function Get-PortunixBinary {
    param(
        [string]$Arch,
        [string]$Version,
        [string]$DownloadDir
    )

    # First check if binary already exists locally (e.g. in release archive)
    $localPortunix = Join-Path $PSScriptRoot "portunix.exe"
    if (Test-Path $localPortunix) {
        Write-Host "Found local portunix.exe - using existing binary"
        return $localPortunix
    }

    $parentPortunix = Join-Path (Split-Path $PSScriptRoot -Parent) "portunix.exe"
    if (Test-Path $parentPortunix) {
        Write-Host "Found portunix.exe in parent directory - using existing binary"
        return $parentPortunix
    }

    # Download from GitHub Releases
    $archiveName = "portunix_${Version}_windows_${Arch}.zip"
    $downloadUrl = "$GitHubDownloadBase/v${Version}/$archiveName"
    $checksumName = "checksums_${Version}.txt"
    $checksumUrl = "$GitHubDownloadBase/v${Version}/$checksumName"
    $archivePath = Join-Path $DownloadDir $archiveName
    $checksumPath = Join-Path $DownloadDir $checksumName

    Write-Host ""
    Write-ColorOutput Cyan "Downloading Portunix v$Version for Windows $Arch..."
    Write-Host "  URL: $downloadUrl"

    if ($DryRun) {
        Write-ColorOutput Yellow "[DRY RUN] Would download: $downloadUrl"
        Write-ColorOutput Yellow "[DRY RUN] Would verify checksum from: $checksumUrl"
        return $null
    }

    # Download archive
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -TimeoutSec 120
        $ProgressPreference = 'Continue'
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-ColorOutput Red "ERROR: Failed to download Portunix"
        Write-Host ""

        if ($statusCode -eq 404) {
            Write-Host "Release archive not found: $archiveName"
            Write-Host ""
            Write-Host "This may mean:"
            Write-Host "  - Version v$Version does not have a Windows $Arch build"
            Write-Host "  - The release is still being published"
            Write-Host ""
            Write-Host "Check available assets at:"
            Write-Host "  https://github.com/$GitHubOwner/$GitHubRepo/releases/tag/v$Version"
        } else {
            Write-Host "Download failed: $($_.Exception.Message)"
            Write-Host ""
            Write-Host "Try downloading manually:"
            Write-Host "  $downloadUrl"
        }
        exit 1
    }

    # Download checksums
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -TimeoutSec 30
        $ProgressPreference = 'Continue'
    } catch {
        Write-ColorOutput Yellow "WARNING: Could not download checksum file - skipping verification"
        Write-Host "  This is not critical, but the download integrity cannot be verified."
    }

    # Verify checksum
    if (Test-Path $checksumPath) {
        Write-Host "Verifying SHA256 checksum..."
        $actualHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()
        $checksumContent = Get-Content $checksumPath

        $expectedLine = $checksumContent | Where-Object { $_ -match $archiveName }
        if ($expectedLine) {
            $expectedHash = ($expectedLine -split '\s+')[0].ToLower()
            if ($actualHash -eq $expectedHash) {
                Write-ColorOutput Green "  Checksum verified OK"
            } else {
                Write-ColorOutput Red "ERROR: Checksum verification FAILED"
                Write-Host ""
                Write-Host "  Expected: $expectedHash"
                Write-Host "  Actual:   $actualHash"
                Write-Host ""
                Write-Host "The downloaded file may be corrupted or tampered with."
                Write-Host "Please try downloading again or download manually from:"
                Write-Host "  https://github.com/$GitHubOwner/$GitHubRepo/releases/tag/v$Version"

                # Clean up
                Remove-Item -Path $archivePath -Force -ErrorAction SilentlyContinue
                exit 1
            }
        } else {
            Write-ColorOutput Yellow "  Checksum entry for $archiveName not found in checksums file"
        }
    }

    # Unblock downloaded file (remove Zone.Identifier ADS)
    try {
        Unblock-File -Path $archivePath -ErrorAction SilentlyContinue
    } catch {
        # Unblock-File may not be available on all systems
    }

    # Extract archive
    Write-Host "Extracting archive..."
    $extractDir = Join-Path $DownloadDir "portunix_extracted"
    try {
        Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
    } catch {
        Write-ColorOutput Red "ERROR: Failed to extract archive"
        Write-Host ""
        Write-Host "  $($_.Exception.Message)"
        Write-Host ""
        Write-Host "The archive may be corrupted. Try downloading again."
        exit 1
    }

    # Find portunix.exe in extracted files
    $extractedBinary = Get-ChildItem -Path $extractDir -Filter "portunix.exe" -Recurse | Select-Object -First 1
    if (-not $extractedBinary) {
        Write-ColorOutput Red "ERROR: portunix.exe not found in downloaded archive"
        Write-Host ""
        Write-Host "The release archive does not contain portunix.exe."
        Write-Host "This may be a packaging issue. Please report at:"
        Write-Host "  https://github.com/$GitHubOwner/$GitHubRepo/issues"
        exit 1
    }

    # Unblock extracted binary
    try {
        Unblock-File -Path $extractedBinary.FullName -ErrorAction SilentlyContinue
    } catch {
        # Unblock-File may not be available on all systems
    }

    Write-ColorOutput Green "  Download and extraction successful"
    return $extractedBinary.FullName
}

# ADR-031: Cross-Platform Binary Distribution
function Get-PlatformsDirectory {
    # Check if platforms directory exists in script directory
    $localPlatforms = Join-Path $PSScriptRoot "platforms"
    if (Test-Path $localPlatforms) {
        return $localPlatforms
    }

    # Check in parent directory
    $parentPlatforms = Join-Path (Split-Path $PSScriptRoot -Parent) "platforms"
    if (Test-Path $parentPlatforms) {
        return $parentPlatforms
    }

    return $null
}

# ADR-031: Install cross-platform binary archives
function Install-PlatformArchives {
    param(
        [string]$InstallDir
    )

    $platformsSource = Get-PlatformsDirectory
    if (-not $platformsSource) {
        Write-Host "No platforms directory found (cross-platform binaries not included)"
        return
    }

    Write-Host "Installing cross-platform binaries (ADR-031)..."

    # Create platforms directory in install location
    $destPlatforms = Join-Path $InstallDir "platforms"
    try {
        if (-not (Test-Path $destPlatforms)) {
            New-Item -ItemType Directory -Path $destPlatforms -Force | Out-Null
        }
    } catch {
        Write-Warning "Could not create platforms directory: $_"
        return
    }

    # Copy platform archives
    $count = 0
    $archives = Get-ChildItem -Path $platformsSource -Include "*.tar.gz", "*.zip" -File

    foreach ($archive in $archives) {
        try {
            Copy-Item -Path $archive.FullName -Destination $destPlatforms -Force
            $count++
            if (-not $Silent) {
                Write-ColorOutput Green "  Installed $($archive.Name)"
            }
        } catch {
            Write-Warning "Failed to copy $($archive.Name): $_"
        }
    }

    if ($count -gt 0) {
        Write-ColorOutput Green "Installed $count platform archive(s) for cross-platform provisioning"
    } else {
        Write-Host "No platform archives found to install"
    }
}

# Main installation logic
function Install-Portunix {
    Write-Host ""
    Write-ColorOutput Cyan "=========================================="
    Write-ColorOutput Cyan "     Portunix Installation Script v$ScriptVersion"
    Write-ColorOutput Cyan "=========================================="
    Write-Host ""

    # Detect architecture
    $arch = Get-PlatformArch
    Write-Host "Platform: Windows $arch"

    # Get latest release info
    $release = Get-LatestReleaseInfo
    $version = $release.tag_name -replace '^v', ''
    Write-Host "Latest version: v$version"

    # Create temp directory for download
    $downloadDir = Join-Path $env:TEMP "portunix_install_$([System.IO.Path]::GetRandomFileName())"
    New-Item -ItemType Directory -Path $downloadDir -Force | Out-Null

    try {
        # Get or download portunix binary
        $portunixPath = Get-PortunixBinary -Arch $arch -Version $version -DownloadDir $downloadDir

        if ($DryRun) {
            Write-Host ""
            Write-ColorOutput Yellow "[DRY RUN] Installation preview:"
            Write-ColorOutput Yellow "  Version: v$version"
            Write-ColorOutput Yellow "  Architecture: $arch"
            if ($Path) {
                Write-ColorOutput Yellow "  Install path: $Path"
            } else {
                Write-ColorOutput Yellow "  Install path: C:\Portunix (default, no admin required)"
            }
            Write-ColorOutput Yellow "  Add to PATH: $AddToPath"
            Write-Host ""
            Write-Host "Run without -DryRun to perform the actual installation."
            return
        }

        if (-not $portunixPath) {
            Write-ColorOutput Red "ERROR: Could not locate portunix binary"
            exit 1
        }

        Write-Host ""

        # Get version from binary
        try {
            $binaryVersion = & $portunixPath --version 2>&1
            Write-Host "Binary version: $binaryVersion"
        } catch {
            Write-Warning "Could not determine binary version"
        }

        Write-Host ""

        # Build command arguments
        $installArgs = @("install-self")

        if ($Silent) {
            $installArgs += "--silent"
            Write-Host "Running silent installation..."
        } else {
            Write-Host "Starting interactive installation..."
        }

        if ($Path) {
            $installArgs += "--path"
            $installArgs += $Path
        }

        if ($AddToPath) {
            $installArgs += "--add-to-path"
        }

        if ($CreateConfig) {
            $installArgs += "--create-config"
        }

        # Check if we need admin rights
        if (-not $Silent -and -not (Test-Administrator)) {
            Write-ColorOutput Yellow "Note: Running without administrator privileges."
            Write-Host "  Installation to 'Program Files' requires admin rights."
            Write-Host "  C:\Portunix is recommended (no admin required)."
            Write-Host ""
        }

        # Run the installation
        try {
            & $portunixPath $installArgs

            if ($LASTEXITCODE -eq 0) {
                Write-Host ""
                Write-ColorOutput Green "========================================"
                Write-ColorOutput Green "  INSTALLATION SUCCESSFUL"
                Write-ColorOutput Green "========================================"
                Write-Host ""

                Write-Host "Portunix has been installed successfully."
                if ($Path) {
                    Write-Host "Installation location: $Path"
                }

                # Determine install directory for platform archives
                if ($Path) {
                    $installDir = Split-Path $Path -Parent
                    if (-not $installDir) {
                        $installDir = $Path
                    }
                } else {
                    $installDir = "C:\Portunix"
                }

                # ADR-031: Install cross-platform binary archives
                Write-Host ""
                Install-PlatformArchives -InstallDir $installDir

                # PATH instructions
                if ($AddToPath) {
                    Write-Host ""
                    Write-ColorOutput Yellow "PATH Configuration:"
                    Write-Host "You may need to restart your terminal or log out"
                    Write-Host "and log back in for PATH changes to take effect."
                }

                Write-Host ""
                Write-ColorOutput Cyan "Next steps:"
                Write-Host "1. Open a new terminal window"
                Write-Host "2. Verify installation: portunix --version"
                Write-Host "3. Get help: portunix --help"
                Write-Host "4. Install development tools: portunix install default"

            } else {
                Show-InstallFailure $LASTEXITCODE
                exit $LASTEXITCODE
            }
        } catch {
            Write-Host ""
            Write-ColorOutput Red "========================================"
            Write-ColorOutput Red "  INSTALLATION FAILED"
            Write-ColorOutput Red "========================================"
            Write-Host ""
            Write-ColorOutput Red "Error: $_"
            Write-Host ""
            Write-Host "Troubleshooting steps:"
            Write-Host "  1. Try running as Administrator (right-click -> Run as administrator)"
            Write-Host "  2. Try a different install path: .\install.ps1 -Path C:\Portunix"
            Write-Host "  3. Check antivirus is not blocking the installation"
            Write-Host "  4. Report issue: https://github.com/$GitHubOwner/$GitHubRepo/issues"
            exit 1
        }
    } finally {
        # Clean up temp directory
        if (Test-Path $downloadDir) {
            Remove-Item -Path $downloadDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }

    # Wait for user input before closing (if interactive)
    if (-not $Silent) {
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
}

function Show-InstallFailure {
    param([int]$ExitCode)

    Write-Host ""
    Write-ColorOutput Red "========================================"
    Write-ColorOutput Red "  INSTALLATION FAILED"
    Write-ColorOutput Red "========================================"
    Write-Host ""
    Write-Host "Exit code: $ExitCode"
    Write-Host ""
    Write-Host "Common solutions:"

    if (-not (Test-Administrator)) {
        Write-Host "  - Run as Administrator for system-wide installation"
        Write-Host "    Right-click PowerShell -> 'Run as administrator'"
    }

    Write-Host "  - Use a user-writable path: .\install.ps1 -Path C:\Portunix"
    Write-Host "  - Check if antivirus is blocking the installation"
    Write-Host "  - Try silent mode: .\install.ps1 -Silent -Path C:\Portunix"
    Write-Host ""
    Write-Host "If the problem persists, please report at:"
    Write-Host "  https://github.com/$GitHubOwner/$GitHubRepo/issues"
}

# Handle help flag
if ($Help) {
    Show-Help
    exit 0
}

# Handle DryRun early display
if ($DryRun -and -not $Silent) {
    Write-ColorOutput Yellow "[DRY RUN MODE] No changes will be made"
    Write-Host ""
}

# Path selection (interactive)
if (-not $Silent -and -not $DryRun -and -not (Test-Administrator) -and $Path -eq "") {
    Write-Host "Installation Path Options:"
    Write-Host "  1. C:\Portunix (recommended, no admin required)"
    Write-Host "  2. $env:PROGRAMFILES\Portunix (requires admin)"
    Write-Host "  3. Custom path"
    Write-Host ""

    do {
        $choice = Read-Host "Choose installation path (1-3)"

        switch ($choice) {
            "1" {
                $Path = "C:\Portunix"
                break
            }
            "2" {
                Write-Host "This installation requires administrator privileges."
                $response = Read-Host "Would you like to restart with elevated privileges? (Y/N)"

                if ($response -eq 'Y' -or $response -eq 'y') {
                    # Restart script with elevation
                    $scriptPath = $MyInvocation.MyCommand.Path
                    if (-not $scriptPath) {
                        # Script run via iex (piped), save to temp and re-run
                        Write-ColorOutput Yellow "Cannot elevate when running via 'irm | iex'."
                        Write-Host "Please download and run the script directly:"
                        Write-Host '  Invoke-WebRequest -Uri "https://github.com/$GitHubOwner/$GitHubRepo/releases/latest/download/install.ps1" -OutFile install.ps1'
                        Write-Host '  .\install.ps1'
                        exit 1
                    }

                    $arguments = @()
                    if ($Silent) { $arguments += "-Silent" }
                    $arguments += "-Path"; $arguments += "`"$env:PROGRAMFILES\Portunix`""
                    if ($AddToPath) { $arguments += "-AddToPath" }
                    if ($CreateConfig) { $arguments += "-CreateConfig" }
                    $argString = $arguments -join " "

                    try {
                        Start-Process PowerShell -Verb RunAs -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`" $argString" -Wait
                        Write-Host ""
                        Write-ColorOutput Green "Installation process completed."
                        Write-Host "Check the elevated window for results."
                        Write-Host ""
                        Write-Host "Press any key to exit..."
                        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
                        exit 0
                    } catch {
                        Write-ColorOutput Red "ERROR: Failed to start elevated process"
                        Write-Host "  $($_.Exception.Message)"
                        Write-Host ""
                        Write-Host "Try running PowerShell as Administrator manually:"
                        Write-Host "  Right-click PowerShell -> 'Run as administrator'"
                        exit 1
                    }
                } else {
                    Write-Host "Installation cancelled."
                    Write-Host ""
                    Write-Host "Press any key to exit..."
                    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
                    exit 0
                }
            }
            "3" {
                $Path = Read-Host "Enter custom installation path"
                break
            }
            default {
                Write-Host "Invalid choice. Please select 1, 2, or 3."
            }
        }
    } while ($choice -notin @("1", "2", "3"))
}

# Run installation
Install-Portunix
