@echo off
REM Portunix Installer Bootstrap for Windows
REM This batch file bypasses PowerShell execution policy restrictions
REM and launches the install.ps1 script with proper flags.
REM
REM Usage: Double-click this file or run from Command Prompt

echo.
echo  Portunix Installer
echo  ==================
echo.

REM Check if install.ps1 exists alongside this batch file
if exist "%~dp0install.ps1" (
    echo Starting installer from local script...
    powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0install.ps1" %*
) else (
    echo Downloading and running latest installer from GitHub...
    powershell -NoProfile -ExecutionPolicy Bypass -Command "& { $ProgressPreference = 'SilentlyContinue'; try { $script = Invoke-WebRequest -Uri 'https://github.com/cassandragargoyle/Portunix/releases/latest/download/install.ps1' -UseBasicParsing; $tempFile = Join-Path $env:TEMP 'portunix_install.ps1'; Set-Content -Path $tempFile -Value $script.Content -Encoding UTF8; & $tempFile %*; Remove-Item $tempFile -Force -ErrorAction SilentlyContinue } catch { Write-Host 'ERROR: Failed to download installer.' -ForegroundColor Red; Write-Host ''; Write-Host 'Please check your internet connection and try again.'; Write-Host 'Manual download: https://github.com/cassandragargoyle/Portunix/releases/latest' } }"
)

echo.
if "%1"=="-Silent" goto :eof
if "%1"=="--silent" goto :eof
pause
