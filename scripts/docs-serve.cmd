@echo off
REM Serve Portunix documentation locally
REM Wrapper for docs-serve.ps1

setlocal

set "SCRIPT_DIR=%~dp0"

REM Pass all arguments to PowerShell script
powershell -ExecutionPolicy Bypass -File "%SCRIPT_DIR%docs-serve.ps1" %*
