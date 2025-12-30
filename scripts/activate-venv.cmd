@echo off
REM Activate Python Virtual Environment for Portunix
REM Usage: scripts\activate-venv.cmd
REM
REM Note: This must be called without 'call' to properly activate in current session

set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "VENV_DIR=%PROJECT_ROOT%\.venv"
set "ACTIVATE_SCRIPT=%VENV_DIR%\Scripts\activate.bat"

REM Check if venv exists
if not exist "%ACTIVATE_SCRIPT%" (
    echo Virtual environment not found at %VENV_DIR%
    echo Run 'scripts\setup-venv.cmd' first to create it.
    exit /b 1
)

REM Activate
call "%ACTIVATE_SCRIPT%"

echo Virtual environment activated: %VENV_DIR%
python --version
echo To deactivate, run: deactivate
