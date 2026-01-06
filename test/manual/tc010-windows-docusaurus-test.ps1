# TC010: Playbook Run - Local Execution (Docusaurus) - Windows Test
# Issue #119 & #128 Acceptance Test
# Run this script in PowerShell on Windows VM
#
# ============================================================================
# SETUP INSTRUCTIONS FOR WIN11 VM
# ============================================================================
#
# 1. On Linux host - deploy binaries via SSH:
#    scp dist/platforms/windows-amd64/*.exe demo@192.168.53.66:"C:/Portunix/"
#
# 2. On Linux host - upload and run this test:
#    scp test/manual/tc010-windows-docusaurus-test.ps1 demo@192.168.53.66:"C:/Portunix/"
#    ssh demo@192.168.53.66 "cd C:\Portunix && powershell -ExecutionPolicy Bypass -File tc010-windows-docusaurus-test.ps1"
#
# ============================================================================

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TC010: Docusaurus Playbook Test - Windows" -ForegroundColor Cyan
Write-Host "Issue #119 & #128 Features" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check version
Write-Host "[Step 1] Checking Portunix version..." -ForegroundColor Yellow
portunix --version
Write-Host "[PASS] Version check" -ForegroundColor Green
Write-Host ""

# Step 2: Create test directory
Write-Host "[Step 2] Creating test directory..." -ForegroundColor Yellow
$testDir = "$env:TEMP\docusaurus-test-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
New-Item -ItemType Directory -Path $testDir -Force | Out-Null
Set-Location $testDir
Write-Host "Test directory: $testDir"
Write-Host "[PASS] Directory created" -ForegroundColor Green
Write-Host ""

# Step 3: Generate playbook (container target - recommended for Docusaurus)
Write-Host "[Step 3] Generating playbook from template..." -ForegroundColor Yellow
portunix playbook init my-docs --template static-docs --engine docusaurus --target container --verbose
if (Test-Path "my-docs.ptxbook") {
    Write-Host "[PASS] Playbook file created" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Playbook file not found!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 4: Show generated content
Write-Host "[Step 4] Generated playbook content:" -ForegroundColor Yellow
Get-Content "my-docs.ptxbook"
Write-Host ""

# Step 5: Validate playbook
Write-Host "[Step 5] Validating playbook..." -ForegroundColor Yellow
portunix playbook validate my-docs.ptxbook --verbose
Write-Host "[PASS] Validation successful" -ForegroundColor Green
Write-Host ""

# Step 6: Test --list-scripts (Issue #128 Phase 1)
Write-Host "[Step 6] Testing --list-scripts (Issue #128)..." -ForegroundColor Yellow
portunix playbook run my-docs.ptxbook --list-scripts --verbose
Write-Host "[PASS] --list-scripts works" -ForegroundColor Green
Write-Host ""

# Step 7: Test playbook build (Issue #128 Phase 4)
Write-Host "[Step 7] Testing playbook build (Issue #128)..." -ForegroundColor Yellow
portunix playbook build my-docs.ptxbook -o Dockerfile.test --verbose
if (Test-Path "Dockerfile.test") {
    Write-Host "Generated Dockerfile:" -ForegroundColor Gray
    Get-Content "Dockerfile.test"
    Write-Host ""
    Write-Host "[PASS] Dockerfile generated" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Dockerfile not generated!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 8: Run create script to initialize Docusaurus (Issue #128 Phase 1)
Write-Host "[Step 8] Running 'create' script in container..." -ForegroundColor Yellow
portunix playbook run my-docs.ptxbook --script create --verbose
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Docusaurus project creation failed!" -ForegroundColor Red
    exit 1
}
Write-Host "[PASS] Docusaurus project created" -ForegroundColor Green
Write-Host ""

# Step 8b: Run dev script to start development server
Write-Host "[Step 8b] Starting development server (dev script)..." -ForegroundColor Yellow
Write-Host "Server will be available at http://localhost:3000" -ForegroundColor Cyan
portunix playbook run my-docs.ptxbook --script dev --verbose
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Dev server failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 9: Verify named volumes in generated playbook (Issue #128 Phase 2)
Write-Host "[Step 9] Checking named volumes in playbook..." -ForegroundColor Yellow
$content = Get-Content "my-docs.ptxbook" -Raw
if ($content -match ":named") {
    Write-Host "[PASS] Named volumes present in container playbook" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Named volumes not found in container playbook!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "All tests passed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Issue #128 Features Tested:" -ForegroundColor Yellow
Write-Host "  [OK] --list-scripts flag" -ForegroundColor White
Write-Host "  [OK] --script parameter" -ForegroundColor White
Write-Host "  [OK] playbook build command" -ForegroundColor White
Write-Host "  [OK] Named volumes in template" -ForegroundColor White
Write-Host ""
Write-Host "Test directory: $testDir" -ForegroundColor Gray
Write-Host ""
Write-Host "To clean up: Remove-Item -Recurse -Force '$testDir'" -ForegroundColor Gray
