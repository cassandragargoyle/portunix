# Portunix Self-Update Script for Windows
# This script helps with manual updates when automatic update fails

param(
    [Parameter(Mandatory=$true)]
    [string]$NewBinaryPath,
    
    [Parameter(Mandatory=$true)]
    [string]$TargetPath,
    
    [switch]$Force
)

Write-Host "Portunix Update Assistant" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

# Verify new binary exists
if (-not (Test-Path $NewBinaryPath)) {
    Write-Error "New binary not found: $NewBinaryPath"
    exit 1
}

# Verify target directory exists
$targetDir = Split-Path $TargetPath -Parent
if (-not (Test-Path $targetDir)) {
    Write-Error "Target directory not found: $targetDir"
    exit 1
}

# Create backup
$backupPath = $TargetPath + ".backup"
if (Test-Path $TargetPath) {
    Write-Host "Creating backup..." -ForegroundColor Yellow
    try {
        Copy-Item $TargetPath $backupPath -Force
        Write-Host "✓ Backup created: $backupPath" -ForegroundColor Green
    }
    catch {
        Write-Error "Failed to create backup: $_"
        exit 1
    }
}

# Stop any running portunix processes (except this PowerShell)
Write-Host "Checking for running Portunix processes..." -ForegroundColor Yellow
$portunixProcesses = Get-Process -Name "portunix" -ErrorAction SilentlyContinue
if ($portunixProcesses) {
    Write-Host "Found $($portunixProcesses.Count) running Portunix process(es)" -ForegroundColor Yellow
    if ($Force) {
        Write-Host "Force flag set, terminating processes..." -ForegroundColor Yellow
        $portunixProcesses | Stop-Process -Force
        Start-Sleep -Seconds 2
    } else {
        Write-Host "Please close all Portunix processes and run with -Force flag" -ForegroundColor Red
        exit 1
    }
}

# Attempt update
Write-Host "Installing update..." -ForegroundColor Yellow
try {
    # Remove old binary if it exists
    if (Test-Path $TargetPath) {
        Remove-Item $TargetPath -Force
    }
    
    # Copy new binary
    Copy-Item $NewBinaryPath $TargetPath -Force
    
    # Verify the new binary works
    Write-Host "Verifying installation..." -ForegroundColor Yellow
    $version = & $TargetPath --version 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Update successful!" -ForegroundColor Green
        Write-Host "New version: $version" -ForegroundColor Green
        
        # Remove backup on success
        if (Test-Path $backupPath) {
            Remove-Item $backupPath -Force
            Write-Host "✓ Backup cleaned up" -ForegroundColor Green
        }
        
        # Remove temporary binary
        if (Test-Path $NewBinaryPath) {
            Remove-Item $NewBinaryPath -Force
        }
        
        Write-Host ""
        Write-Host "Update completed successfully! 🎉" -ForegroundColor Green
    } else {
        throw "New binary verification failed"
    }
}
catch {
    Write-Error "Update failed: $_"
    
    # Restore backup
    if (Test-Path $backupPath) {
        Write-Host "Restoring backup..." -ForegroundColor Yellow
        try {
            Copy-Item $backupPath $TargetPath -Force
            Write-Host "✓ Backup restored" -ForegroundColor Green
        }
        catch {
            Write-Error "Failed to restore backup: $_"
            Write-Host "Manual intervention required!" -ForegroundColor Red
        }
    }
    exit 1
}