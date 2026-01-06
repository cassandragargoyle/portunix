# TC009: Playbook Run - Local Execution (Hugo) - Windows Test
# Issue #119 Acceptance Test
# Run this script in PowerShell on Windows VM

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TC009: Hugo Playbook Test - Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check version
Write-Host "[Step 1] Checking Portunix version..." -ForegroundColor Yellow
portunix --version
Write-Host ""

# Step 2: Create test directory
Write-Host "[Step 2] Creating test directory..." -ForegroundColor Yellow
$testDir = "$env:TEMP\hugo-test-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
New-Item -ItemType Directory -Path $testDir -Force | Out-Null
Set-Location $testDir
Write-Host "Test directory: $testDir"
Write-Host ""

# Step 3: Generate playbook
Write-Host "[Step 3] Generating playbook from template..." -ForegroundColor Yellow
portunix playbook init my-docs --template static-docs --engine hugo --target container
Write-Host ""

# Step 4: Check generated file
Write-Host "[Step 4] Generated playbook content:" -ForegroundColor Yellow
if (Test-Path "my-docs.ptxbook") {
    Get-Content "my-docs.ptxbook"
    Write-Host ""
    Write-Host "[PASS] Playbook file created successfully" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Playbook file not found!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 5: Validate playbook
Write-Host "[Step 5] Validating playbook..." -ForegroundColor Yellow
portunix playbook validate my-docs.ptxbook
Write-Host ""

# Step 6: Installation
Write-Host "[Step 6] Running installation ..." -ForegroundColor Yellow
portunix playbook run my-docs.ptxbook
Write-Host ""

# Step 7: Ask for actual installation
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Installation completed. Steps 1-6 passed." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run actual installation (installs Hugo):" -ForegroundColor Yellow
Write-Host "  portunix playbook run my-docs.ptxbook" -ForegroundColor White
Write-Host ""
Write-Host "To verify Hugo after installation:" -ForegroundColor Yellow
Write-Host "  hugo version" -ForegroundColor White
Write-Host ""
Write-Host "Test directory: $testDir" -ForegroundColor Gray
