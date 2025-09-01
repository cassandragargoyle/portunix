# Portunix Installation Script for Windows
# This script uses the built-in install-self command of Portunix

param(
    [switch]$Silent,
    [string]$Path = "",
    [switch]$AddToPath,
    [switch]$CreateConfig,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

# Script version
$ScriptVersion = "1.0.0"

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
    -Help           Show this help message

Examples:
    # Interactive installation
    .\install.ps1
    
    # Silent installation with defaults
    .\install.ps1 -Silent
    
    # Install to C:\Portunix (recommended)
    .\install.ps1 -Path "C:\Portunix" -AddToPath
    
    # Install to Program Files (requires admin)
    .\install.ps1 -Path "C:\Program Files\Portunix" -AddToPath

Default installation paths:
  - C:\Portunix\portunix.exe (recommended, no admin required)
  - $env:PROGRAMFILES\Portunix\portunix.exe (requires admin)
"@
}

function Test-Administrator {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-PortunixBinary {
    # Check if portunix.exe exists in current directory
    $localPortunix = Join-Path $PSScriptRoot "portunix.exe"
    if (Test-Path $localPortunix) {
        return $localPortunix
    }
    
    # Check in parent directory (in case script is in subdirectory)
    $parentPortunix = Join-Path (Split-Path $PSScriptRoot -Parent) "portunix.exe"
    if (Test-Path $parentPortunix) {
        return $parentPortunix
    }
    
    Write-Error "portunix.exe not found. Please ensure it's in the same directory as this script."
    exit 1
}

# Main installation logic
function Install-Portunix {
    Write-Host ""
    Write-ColorOutput Cyan "=========================================="
    Write-ColorOutput Cyan "     Portunix Installation Script"
    Write-ColorOutput Cyan "=========================================="
    Write-Host ""
    
    # Find portunix binary
    $portunixPath = Get-PortunixBinary
    Write-Host "Found portunix at: $portunixPath"
    
    # Get version
    try {
        $version = & $portunixPath --version 2>&1
        Write-Host "Version: $version"
    } catch {
        Write-Warning "Could not determine version"
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
        Write-Warning "Running without administrator privileges."
        Write-Warning "Installation to Program Files may fail."
        Write-Host ""
        
        $response = Read-Host "Continue anyway? (Y/N)"
        if ($response -ne 'Y' -and $response -ne 'y') {
            Write-Host "Installation cancelled."
            exit 0
        }
    }
    
    # Run the installation
    try {
        Write-Host ""
        & $portunixPath $installArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-ColorOutput Green "========================================"
            Write-ColorOutput Green "  ✅ INSTALLATION SUCCESSFUL!"
            Write-ColorOutput Green "========================================"
            Write-Host ""
            
            # Show installation location
            Write-Host "Portunix has been installed successfully."
            if ($Path) {
                Write-Host "Installation location: $Path"
            }
            
            # Additional instructions for PATH
            if ($AddToPath) {
                Write-Host ""
                Write-ColorOutput Yellow "⚠️  PATH Configuration:"
                Write-Host "You may need to restart your terminal or log out"
                Write-Host "and log back in for PATH changes to take effect."
            }
            
            Write-Host ""
            Write-ColorOutput Cyan "Next steps:"
            Write-Host "1. Verify installation: portunix --version"
            Write-Host "2. Get help: portunix --help"
            Write-Host "3. Install development tools: portunix install default"
            
            Write-Host ""
            Write-ColorOutput Green "Installation completed successfully! 🎉"
            
        } else {
            Write-Host ""
            Write-ColorOutput Red "========================================"
            Write-ColorOutput Red "  ❌ INSTALLATION FAILED!"
            Write-ColorOutput Red "========================================"
            Write-Host ""
            Write-Error "Installation failed with exit code $LASTEXITCODE"
            
            Write-Host ""
            Write-Host "Press any key to exit..."
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
            exit $LASTEXITCODE
        }
    } catch {
        Write-Host ""
        Write-ColorOutput Red "========================================"
        Write-ColorOutput Red "  ❌ INSTALLATION FAILED!"
        Write-ColorOutput Red "========================================"
        Write-Host ""
        Write-Error "Installation failed: $_"
        
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 1
    }
    
    # Wait for user input before closing (if interactive)
    if (-not $Silent) {
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    }
}

# Handle help flag
if ($Help) {
    Show-Help
    exit 0
}

# Check for elevation request and suggest installation paths
if (-not $Silent -and -not (Test-Administrator) -and $Path -eq "") {
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
                        Write-Error "Failed to start elevated process: $_"
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