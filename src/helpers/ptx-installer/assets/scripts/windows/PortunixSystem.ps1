# Portunix System Detection PowerShell Module
# This module provides easy-to-use PowerShell functions for system detection

# Get comprehensive system information
function Get-PortunixSystemInfo {
    [CmdletBinding()]
    param(
        [switch]$AsJson,
        [switch]$Short
    )
    
    if ($AsJson) {
        & portunix system info --json
    } elseif ($Short) {
        & portunix system info --short
    } else {
        & portunix system info
    }
}

# Check specific system conditions
function Test-PortunixSystem {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [ValidateSet('windows', 'linux', 'macos', 'sandbox', 'docker', 'wsl', 'vm', 'powershell', 'admin')]
        [string]$Condition
    )
    
    & portunix system check $Condition
    return $LASTEXITCODE -eq 0
}

# Convenient wrapper functions for common checks
function Test-IsWindows { Test-PortunixSystem -Condition 'windows' }
function Test-IsLinux { Test-PortunixSystem -Condition 'linux' }
function Test-IsMacOS { Test-PortunixSystem -Condition 'macos' }
function Test-IsSandbox { Test-PortunixSystem -Condition 'sandbox' }
function Test-IsDocker { Test-PortunixSystem -Condition 'docker' }
function Test-IsWSL { Test-PortunixSystem -Condition 'wsl' }
function Test-IsVM { Test-PortunixSystem -Condition 'vm' }
function Test-HasPowerShell { Test-PortunixSystem -Condition 'powershell' }
function Test-IsAdmin { Test-PortunixSystem -Condition 'admin' }

# Get system information as PowerShell object
function Get-PortunixSystemObject {
    [CmdletBinding()]
    param()
    
    $jsonOutput = & portunix system info --json
    if ($LASTEXITCODE -eq 0 -and $jsonOutput) {
        return $jsonOutput | ConvertFrom-Json
    }
    return $null
}

# Conditional execution based on OS
function Invoke-PortunixConditional {
    [CmdletBinding()]
    param(
        [scriptblock]$Windows,
        [scriptblock]$Linux,
        [scriptblock]$MacOS,
        [scriptblock]$Sandbox,
        [scriptblock]$Docker,
        [scriptblock]$WSL,
        [scriptblock]$Default
    )
    
    if ($Sandbox -and (Test-IsSandbox)) {
        & $Sandbox
    } elseif ($Windows -and (Test-IsWindows)) {
        & $Windows
    } elseif ($Linux -and (Test-IsLinux)) {
        & $Linux
    } elseif ($MacOS -and (Test-IsMacOS)) {
        & $MacOS
    } elseif ($Docker -and (Test-IsDocker)) {
        & $Docker
    } elseif ($WSL -and (Test-IsWSL)) {
        & $WSL
    } elseif ($Default) {
        & $Default
    }
}

# Example usage functions
function Show-PortunixSystemExamples {
    Write-Host "Portunix System Detection Examples:" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "Basic Detection:" -ForegroundColor Yellow
    Write-Host '  Test-IsWindows              # Returns $true if Windows'
    Write-Host '  Test-IsSandbox              # Returns $true if Windows Sandbox'
    Write-Host '  Test-IsAdmin                # Returns $true if running as admin'
    Write-Host ""
    
    Write-Host "System Information:" -ForegroundColor Yellow
    Write-Host '  Get-PortunixSystemInfo      # Display formatted info'
    Write-Host '  Get-PortunixSystemInfo -Short  # Short format: "Windows 11 Sandbox"'
    Write-Host '  Get-PortunixSystemObject    # Get as PowerShell object'
    Write-Host ""
    
    Write-Host "Conditional Execution:" -ForegroundColor Yellow
    Write-Host '  Invoke-PortunixConditional -Windows { Write-Host "Running on Windows" }'
    Write-Host '  Invoke-PortunixConditional -Sandbox { Write-Host "In sandbox!" } -Default { Write-Host "Not in sandbox" }'
    Write-Host ""
    
    Write-Host "Script Examples:" -ForegroundColor Yellow
    Write-Host '  if (Test-IsSandbox) {'
    Write-Host '      Write-Host "Configuring for sandbox environment"'
    Write-Host '      # Sandbox-specific configuration'
    Write-Host '  }'
    Write-Host ""
    Write-Host '  $sys = Get-PortunixSystemObject'
    Write-Host '  Write-Host "Running on $($sys.os) $($sys.version) ($($sys.variant))"'
}

# Functions are now available in current session (loaded via dot-sourcing)
Write-Host "Portunix System Detection PowerShell functions loaded!" -ForegroundColor Green
Write-Host "Available functions: Test-IsWindows, Test-IsSandbox, Test-IsAdmin, Get-PortunixSystemInfo, etc." -ForegroundColor Cyan
Write-Host "Type 'Show-PortunixSystemExamples' to see usage examples." -ForegroundColor Cyan